package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
)

type StorageService interface {
	UploadImage(ctx context.Context, file *multipart.FileHeader, folder string) (string, error)
	UploadFile(ctx context.Context, file *multipart.FileHeader, folder string) (string, error)
	DeleteFile(ctx context.Context, url string) error
	GeneratePresignedURL(fileUrl string, expiry time.Duration) (string, error)
	TestConnection() error
}

type s3StorageService struct {
	s3Client *s3.S3
	uploader *s3manager.Uploader
	bucket   string
	region   string
	session  *session.Session
}

func NewS3StorageService(accessKey, secretKey, region, bucket string) StorageService {

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),

		LogLevel: aws.LogLevel(aws.LogDebugWithHTTPBody),
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create AWS session: %v", err))
	}

	return &s3StorageService{
		s3Client: s3.New(sess),
		uploader: s3manager.NewUploader(sess),
		bucket:   bucket,
		region:   region,
		session:  sess,
	}
}

func (s *s3StorageService) TestConnection() error {

	input := &s3.GetBucketLocationInput{
		Bucket: aws.String(s.bucket),
	}

	_, err := s.s3Client.GetBucketLocation(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				return fmt.Errorf("bucket %s does not exist: %v", s.bucket, aerr)
			case "AccessDenied":
				return fmt.Errorf("access denied to bucket %s: %v", s.bucket, aerr)
			case "InvalidAccessKeyId":
				return fmt.Errorf("invalid access key: %v", aerr)
			case "SignatureDoesNotMatch":
				return fmt.Errorf("signature mismatch - check secret key: %v", aerr)
			default:
				return fmt.Errorf("AWS error: %v", aerr)
			}
		}
		return fmt.Errorf("failed to test S3 connection: %v", err)
	}

	return nil
}

func (s *s3StorageService) UploadImage(ctx context.Context, file *multipart.FileHeader, folder string) (string, error) {
	if !isImageFile(file.Filename) {
		return "", fmt.Errorf("invalid image file type: %s", file.Filename)
	}
	return s.uploadFile(ctx, file, folder)
}

func (s *s3StorageService) UploadFile(ctx context.Context, file *multipart.FileHeader, folder string) (string, error) {
	return s.uploadFile(ctx, file, folder)
}

func (s *s3StorageService) uploadFile(ctx context.Context, file *multipart.FileHeader, folder string) (string, error) {

	if file == nil {
		return "", fmt.Errorf("file is nil")
	}

	if file.Size == 0 {
		return "", fmt.Errorf("file is empty")
	}

	if file.Size > 100*1024*1024 {
		return "", fmt.Errorf("file too large: %d bytes (max: 100MB)", file.Size)
	}

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", file.Filename, err)
	}
	defer src.Close()

	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".bin"
	}

	folder = strings.Trim(folder, "/")
	if folder == "" {
		folder = "uploads"
	}

	filename := fmt.Sprintf("%s/%s_%d%s", folder, uuid.New().String(), time.Now().Unix(), ext)
	contentType := getContentType(ext)

	uploadInput := &s3manager.UploadInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(filename),
		Body:        src,
		ContentType: aws.String(contentType),
		ACL:         aws.String("private"),
		Metadata: map[string]*string{
			"original-filename": aws.String(file.Filename),
			"uploaded-at":       aws.String(time.Now().UTC().Format(time.RFC3339)),
		},
	}

	result, err := s.uploader.UploadWithContext(ctx, uploadInput)
	if err != nil {

		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "AccessDenied":
				return "", fmt.Errorf("access denied to bucket %s: %v", s.bucket, aerr)
			case "InvalidAccessKeyId":
				return "", fmt.Errorf("invalid access key ID: %v", aerr)
			case "SignatureDoesNotMatch":
				return "", fmt.Errorf("signature mismatch - check AWS credentials: %v", aerr)
			case "NoSuchBucket":
				return "", fmt.Errorf("bucket %s does not exist: %v", s.bucket, aerr)
			case "EntityTooLarge":
				return "", fmt.Errorf("file too large: %v", aerr)
			case "InvalidRequest":
				return "", fmt.Errorf("invalid request - check file content: %v", aerr)
			default:
				return "", fmt.Errorf("AWS error [%s]: %v", aerr.Code(), aerr)
			}
		}
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	if err := s.verifyUpload(filename); err != nil {
		return "", fmt.Errorf("upload verification failed: %w", err)
	}

	if result.Location != "" {
		return filename, nil
	}

	return filename, nil
}

func (s *s3StorageService) verifyUpload(key string) error {

	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err := s.s3Client.HeadObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == "NotFound" {
				return fmt.Errorf("uploaded file not found in S3")
			}
		}
		return fmt.Errorf("failed to verify upload: %v", err)
	}

	return nil
}

func (s *s3StorageService) GeneratePresignedURL(fileKey string, expiry time.Duration) (string, error) {
	if fileKey == "" {
		return "", nil
	}

	key := extractKeyFromS3Url(fileKey)

	_, err := s.s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound":
				return "", fmt.Errorf("file not found in S3: %s", key)
			case "AccessDenied":
				return "", fmt.Errorf("access denied to file: %s", key)
			default:
				return "", fmt.Errorf("error checking file existence: %v", aerr)
			}
		}
		return "", fmt.Errorf("failed to check file existence: %v", err)
	}

	req, _ := s.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	urlStr, err := req.Presign(expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL for key %s: %w", key, err)
	}

	return urlStr, nil
}

func (s *s3StorageService) DeleteFile(ctx context.Context, fileKey string) error {
	if fileKey == "" {
		return nil
	}

	key := extractKeyFromS3Url(fileKey)

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err := s.s3Client.DeleteObjectWithContext(ctx, input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "AccessDenied":
				return fmt.Errorf("access denied when deleting file: %s", key)
			default:
				return fmt.Errorf("AWS error deleting file: %v", aerr)
			}
		}
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}

func extractKeyFromS3Url(fileUrl string) string {

	if !strings.HasPrefix(fileUrl, "http") {

		return fileUrl
	}

	parsedURL, err := url.Parse(fileUrl)
	if err != nil {

		if strings.Contains(fileUrl, ".com/") {
			parts := strings.Split(fileUrl, ".com/")
			if len(parts) > 1 {
				return parts[1]
			}
		}
		return fileUrl
	}

	key := strings.TrimPrefix(parsedURL.Path, "/")

	if strings.Contains(parsedURL.Host, "s3") {
		if strings.HasPrefix(parsedURL.Host, "s3.") || strings.HasPrefix(parsedURL.Host, "s3-") {

			pathParts := strings.SplitN(key, "/", 2)
			if len(pathParts) > 1 {
				return pathParts[1]
			}
		}

		return key
	}

	return key
}

func getContentType(ext string) string {
	contentTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".txt":  "text/plain",
		".csv":  "text/csv",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".xls":  "application/vnd.ms-excel",
		".mp4":  "video/mp4",
		".mp3":  "audio/mpeg",
	}

	ext = strings.ToLower(ext)
	if contentType, exists := contentTypes[ext]; exists {
		return contentType
	}

	return "application/octet-stream"
}

func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".tiff"}

	for _, validExt := range imageExts {
		if ext == validExt {
			return true
		}
	}
	return false
}

func ValidateAWSCredentials(accessKey, secretKey, region, bucket string) error {
	if accessKey == "" {
		return fmt.Errorf("AWS access key is required")
	}

	if secretKey == "" {
		return fmt.Errorf("AWS secret key is required")
	}

	if region == "" {
		return fmt.Errorf("AWS region is required")
	}

	if bucket == "" {
		return fmt.Errorf("S3 bucket name is required")
	}

	if len(accessKey) < 16 || len(accessKey) > 128 {
		return fmt.Errorf("invalid access key format")
	}

	if len(secretKey) < 8 || len(secretKey) > 128 {
		return fmt.Errorf("invalid secret key format")
	}

	return nil
}
