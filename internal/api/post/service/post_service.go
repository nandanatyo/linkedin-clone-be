package service

import (
	"context"
	"errors"
	"linked-clone/internal/api/post/dto"
	"linked-clone/internal/domain/entities"
	"linked-clone/internal/domain/repositories"
	"linked-clone/pkg/logger"
	"linked-clone/pkg/storage"
	"time"

	"mime/multipart"

	"gorm.io/gorm"
)

type PostService interface {
	CreatePost(ctx context.Context, userID uint, req *dto.CreatePostRequest, file *multipart.FileHeader) (*dto.PostResponse, error)
	GetPost(ctx context.Context, id uint) (*dto.PostResponse, error)
	GetUserPosts(ctx context.Context, userID uint, limit, offset int) ([]*dto.PostResponse, error)
	GetFeed(ctx context.Context, userID uint, limit, offset int) ([]*dto.PostResponse, error)
	UpdatePost(ctx context.Context, userID, postID uint, req *dto.UpdatePostRequest) (*dto.PostResponse, error)
	DeletePost(ctx context.Context, userID, postID uint) error

	LikePost(ctx context.Context, userID, postID uint) (*dto.LikeResponse, error)
	UnlikePost(ctx context.Context, userID, postID uint) error
	GetPostLikes(ctx context.Context, postID uint, limit, offset int) ([]*dto.LikeResponse, error)

	AddComment(ctx context.Context, userID, postID uint, req *dto.AddCommentRequest) (*dto.CommentResponse, error)
	GetComments(ctx context.Context, postID uint, limit, offset int) ([]*dto.CommentResponse, error)
	UpdateComment(ctx context.Context, userID, commentID uint, content string) (*dto.CommentResponse, error)
	DeleteComment(ctx context.Context, userID, commentID uint) error
}

type postService struct {
	postRepo       repositories.PostRepository
	userRepo       repositories.UserRepository
	likeRepo       repositories.LikeRepository
	commentRepo    repositories.CommentRepository
	storageService storage.StorageService
	logger         logger.Logger
}

func NewPostService(
	postRepo repositories.PostRepository,
	userRepo repositories.UserRepository,
	likeRepo repositories.LikeRepository,
	commentRepo repositories.CommentRepository,
	storageService storage.StorageService,
	logger logger.Logger,
) PostService {
	return &postService{
		postRepo:       postRepo,
		userRepo:       userRepo,
		likeRepo:       likeRepo,
		commentRepo:    commentRepo,
		storageService: storageService,
		logger:         logger,
	}
}

func (s *postService) CreatePost(ctx context.Context, userID uint, req *dto.CreatePostRequest, file *multipart.FileHeader) (*dto.PostResponse, error) {
	var imageURL string

	if file != nil {
		url, err := s.storageService.UploadImage(ctx, file, "posts")
		if err != nil {
			s.logger.Error("Failed to upload image", "error", err)
			return nil, errors.New("failed to upload image")
		}
		imageURL = url
	}

	post := &entities.Post{
		UserID:   userID,
		Content:  req.Content,
		ImageURL: imageURL,
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		s.logger.Error("Failed to create post", "error", err)
		return nil, errors.New("failed to create post")
	}

	return s.GetPost(ctx, post.ID)
}

func (s *postService) GetPost(ctx context.Context, id uint) (*dto.PostResponse, error) {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("post not found")
		}
		s.logger.Error("Failed to get post", "error", err)
		return nil, errors.New("failed to get post")
	}

	var imageURL string
	if post.ImageURL != "" {
		signed, err := s.storageService.GeneratePresignedURL(post.ImageURL, 15*time.Minute)
		if err != nil {
			s.logger.Error("Failed to generate image presigned URL", "error", err)
		} else {
			imageURL = signed
		}
	}

	var profilePicture string
	if post.User.ProfilePicture != "" {
		signed, err := s.storageService.GeneratePresignedURL(post.User.ProfilePicture, 15*time.Minute)
		if err != nil {
			s.logger.Error("Failed to generate profile picture presigned URL", "error", err)
		} else {
			profilePicture = signed
		}
	}

	return &dto.PostResponse{
		ID:        post.ID,
		Content:   post.Content,
		ImageURL:  imageURL,
		LikeCount: post.LikeCount,
		User: &dto.UserInfo{
			ID:             post.User.ID,
			Username:       post.User.Username,
			FullName:       post.User.FullName,
			ProfilePicture: profilePicture,
		},
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}, nil
}

func (s *postService) GetUserPosts(ctx context.Context, userID uint, limit, offset int) ([]*dto.PostResponse, error) {
	posts, err := s.postRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get user posts", "error", err)
		return nil, errors.New("failed to get posts")
	}

	var responses []*dto.PostResponse
	for _, post := range posts {
		var imageURL, profilePicture string
		if post.ImageURL != "" {
			signed, err := s.storageService.GeneratePresignedURL(post.ImageURL, 15*time.Minute)
			if err != nil {
				s.logger.Error("Failed to generate image presigned URL", "error", err)
			} else {
				imageURL = signed
			}
		}
		if post.User.ProfilePicture != "" {
			signed, err := s.storageService.GeneratePresignedURL(post.User.ProfilePicture, 15*time.Minute)
			if err != nil {
				s.logger.Error("Failed to generate profile picture presigned URL", "error", err)
			} else {
				profilePicture = signed
			}
		}
		responses = append(responses, &dto.PostResponse{
			ID:        post.ID,
			Content:   post.Content,
			ImageURL:  imageURL,
			LikeCount: post.LikeCount,
			User: &dto.UserInfo{
				ID:             post.User.ID,
				Username:       post.User.Username,
				FullName:       post.User.FullName,
				ProfilePicture: profilePicture,
			},
			CreatedAt: post.CreatedAt,
			UpdatedAt: post.UpdatedAt,
		})
	}
	return responses, nil
}

func (s *postService) GetFeed(ctx context.Context, userID uint, limit, offset int) ([]*dto.PostResponse, error) {
	posts, err := s.postRepo.GetFeed(ctx, userID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get feed", "error", err)
		return nil, errors.New("failed to get feed")
	}

	var responses []*dto.PostResponse
	for _, post := range posts {
		var imageURL, profilePicture string
		if post.ImageURL != "" {
			signed, err := s.storageService.GeneratePresignedURL(post.ImageURL, 15*time.Minute)
			if err != nil {
				s.logger.Error("Failed to generate image presigned URL", "error", err)
			} else {
				imageURL = signed
			}
		}
		if post.User.ProfilePicture != "" {
			signed, err := s.storageService.GeneratePresignedURL(post.User.ProfilePicture, 15*time.Minute)
			if err != nil {
				s.logger.Error("Failed to generate profile picture presigned URL", "error", err)
			} else {
				profilePicture = signed
			}
		}
		responses = append(responses, &dto.PostResponse{
			ID:        post.ID,
			Content:   post.Content,
			ImageURL:  imageURL,
			LikeCount: post.LikeCount,
			User: &dto.UserInfo{
				ID:             post.User.ID,
				Username:       post.User.Username,
				FullName:       post.User.FullName,
				ProfilePicture: profilePicture,
			},
			CreatedAt: post.CreatedAt,
			UpdatedAt: post.UpdatedAt,
		})
	}
	return responses, nil
}

func (s *postService) UpdatePost(ctx context.Context, userID, postID uint, req *dto.UpdatePostRequest) (*dto.PostResponse, error) {
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("post not found")
		}
		s.logger.Error("Failed to get post", "error", err)
		return nil, errors.New("failed to get post")
	}

	if post.UserID != userID {
		return nil, errors.New("unauthorized to update this post")
	}

	if req.Content != "" {
		post.Content = req.Content
	}

	if err := s.postRepo.Update(ctx, post); err != nil {
		s.logger.Error("Failed to update post", "error", err)
		return nil, errors.New("failed to update post")
	}

	return s.GetPost(ctx, postID)
}

func (s *postService) DeletePost(ctx context.Context, userID, postID uint) error {
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("post not found")
		}
		s.logger.Error("Failed to get post", "error", err)
		return errors.New("failed to get post")
	}

	if post.UserID != userID {
		return errors.New("unauthorized to delete this post")
	}

	if post.ImageURL != "" {
		go func() {
			if err := s.storageService.DeleteFile(ctx, post.ImageURL); err != nil {
				s.logger.Error("Failed to delete post image", "error", err)
			}
		}()
	}

	if err := s.postRepo.Delete(ctx, postID); err != nil {
		s.logger.Error("Failed to delete post", "error", err)
		return errors.New("failed to delete post")
	}

	return nil
}

func (s *postService) LikePost(ctx context.Context, userID, postID uint) (*dto.LikeResponse, error) {

	existingLike, _ := s.likeRepo.FindByUserAndPost(ctx, userID, postID)
	if existingLike != nil {
		return nil, errors.New("post already liked")
	}

	like := &entities.Like{
		UserID: userID,
		PostID: postID,
	}

	if err := s.likeRepo.Create(ctx, like); err != nil {
		s.logger.Error("Failed to create like", "error", err)
		return nil, errors.New("failed to like post")
	}

	if err := s.postRepo.IncrementLikeCount(ctx, postID); err != nil {
		s.logger.Error("Failed to increment like count", "error", err)
	}

	return &dto.LikeResponse{
		ID:     like.ID,
		UserID: userID,
		PostID: postID,
	}, nil
}

func (s *postService) UnlikePost(ctx context.Context, userID, postID uint) error {
	if err := s.likeRepo.Delete(ctx, userID, postID); err != nil {
		s.logger.Error("Failed to delete like", "error", err)
		return errors.New("failed to unlike post")
	}

	if err := s.postRepo.DecrementLikeCount(ctx, postID); err != nil {
		s.logger.Error("Failed to decrement like count", "error", err)
	}

	return nil
}

func (s *postService) AddComment(ctx context.Context, userID, postID uint, req *dto.AddCommentRequest) (*dto.CommentResponse, error) {

	_, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("post not found")
		}
		s.logger.Error("Failed to get post", "error", err)
		return nil, errors.New("failed to get post")
	}

	comment := &entities.Comment{
		UserID:  userID,
		PostID:  postID,
		Content: req.Content,
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		s.logger.Error("Failed to create comment", "error", err)
		return nil, errors.New("failed to add comment")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", "error", err)
		return nil, errors.New("failed to get user details")
	}

	profilePictureURL := ""
	if user.ProfilePicture != "" {
		if presignedURL, err := s.storageService.GeneratePresignedURL(user.ProfilePicture, 24*time.Hour); err == nil {
			profilePictureURL = presignedURL
		} else {
			s.logger.Error("Failed to generate profile picture presigned URL", "error", err)
		}
	}

	return &dto.CommentResponse{
		ID:      comment.ID,
		Content: comment.Content,
		User: &dto.UserInfo{
			ID:             user.ID,
			Username:       user.Username,
			FullName:       user.FullName,
			ProfilePicture: profilePictureURL,
		},
		CreatedAt: comment.CreatedAt,
	}, nil
}

func (s *postService) GetComments(ctx context.Context, postID uint, limit, offset int) ([]*dto.CommentResponse, error) {

	_, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("post not found")
		}
		return nil, errors.New("failed to get post")
	}

	comments, err := s.commentRepo.GetByPostID(ctx, postID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get comments", "error", err)
		return nil, errors.New("failed to get comments")
	}

	var responses []*dto.CommentResponse
	for _, comment := range comments {
		profilePictureURL := ""
		if comment.User.ProfilePicture != "" {
			if presignedURL, err := s.storageService.GeneratePresignedURL(comment.User.ProfilePicture, 24*time.Hour); err == nil {
				profilePictureURL = presignedURL
			} else {
				s.logger.Error("Failed to generate profile picture presigned URL", "error", err)
			}
		}

		responses = append(responses, &dto.CommentResponse{
			ID:      comment.ID,
			Content: comment.Content,
			User: &dto.UserInfo{
				ID:             comment.User.ID,
				Username:       comment.User.Username,
				FullName:       comment.User.FullName,
				ProfilePicture: profilePictureURL,
			},
			CreatedAt: comment.CreatedAt,
		})
	}

	return responses, nil
}

func (s *postService) UpdateComment(ctx context.Context, userID, commentID uint, content string) (*dto.CommentResponse, error) {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("comment not found")
		}
		return nil, errors.New("failed to get comment")
	}

	if comment.UserID != userID {
		return nil, errors.New("unauthorized to update this comment")
	}

	comment.Content = content
	if err := s.commentRepo.Update(ctx, comment); err != nil {
		s.logger.Error("Failed to update comment", "error", err)
		return nil, errors.New("failed to update comment")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", "error", err)
		return nil, errors.New("failed to get user details")
	}

	profilePictureURL := ""
	if user.ProfilePicture != "" {
		if presignedURL, err := s.storageService.GeneratePresignedURL(user.ProfilePicture, 24*time.Hour); err == nil {
			profilePictureURL = presignedURL
		} else {
			s.logger.Error("Failed to generate profile picture presigned URL", "error", err)
		}
	}

	return &dto.CommentResponse{
		ID:      comment.ID,
		Content: comment.Content,
		User: &dto.UserInfo{
			ID:             user.ID,
			Username:       user.Username,
			FullName:       user.FullName,
			ProfilePicture: profilePictureURL,
		},
		CreatedAt: comment.CreatedAt,
	}, nil
}

func (s *postService) DeleteComment(ctx context.Context, userID, commentID uint) error {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("comment not found")
		}
		return errors.New("failed to get comment")
	}

	if comment.UserID != userID {
		return errors.New("unauthorized to delete this comment")
	}

	if err := s.commentRepo.Delete(ctx, commentID); err != nil {
		s.logger.Error("Failed to delete comment", "error", err)
		return errors.New("failed to delete comment")
	}

	return nil
}
