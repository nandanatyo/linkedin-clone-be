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
}

type s3StorageService struct {
	s3Client *s3.S3
	uploader *s3manager.Uploader
	bucket   string
	region   string
}

func NewS3StorageService(accessKey, secretKey, region, bucket string) StorageService {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	}))

	return &s3StorageService{
		s3Client: s3.New(sess),
		uploader: s3manager.NewUploader(sess),
		bucket:   bucket,
		region:   region,
	}
}

func (s *s3StorageService) UploadImage(ctx context.Context, file *multipart.FileHeader, folder string) (string, error) {
	if !isImageFile(file.Filename) {
		return "", fmt.Errorf("invalid image file type")
	}
	return s.uploadFile(ctx, file, folder)
}

func (s *s3StorageService) UploadFile(ctx context.Context, file *multipart.FileHeader, folder string) (string, error) {
	return s.uploadFile(ctx, file, folder)
}

func (s *s3StorageService) uploadFile(ctx context.Context, file *multipart.FileHeader, folder string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s/%s_%d%s", folder, uuid.New().String(), time.Now().Unix(), ext)

	contentType := getContentType(ext)

	uploadInput := &s3manager.UploadInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(filename),
		Body:        src,
		ContentType: aws.String(contentType),
	}

	_, err = s.uploader.UploadWithContext(ctx, uploadInput)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return filename, nil
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
		return "", fmt.Errorf("file does not exist in S3 bucket %s with key %s: %w", s.bucket, key, err)
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

	_, err := s.s3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
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
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".txt":
		return "text/plain"
	case ".csv":
		return "text/csv"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".xls":
		return "application/vnd.ms-excel"
	default:
		return "application/octet-stream"
	}
}

func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	for _, validExt := range imageExts {
		if ext == validExt {
			return true
		}
	}
	return false
}
