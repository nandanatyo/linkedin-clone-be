package service

import (
	"context"
	"errors"
	"linked-clone/internal/api/user/dto"
	"linked-clone/internal/domain/repositories"
	"linked-clone/pkg/logger"
	"linked-clone/pkg/storage"
	"mime/multipart"
	"time"

	"gorm.io/gorm"
)

type UserService interface {
	GetProfile(ctx context.Context, userID uint) (*dto.UserProfileResponse, error)
	UpdateProfile(ctx context.Context, userID uint, req *dto.UpdateProfileRequest) (*dto.UserProfileResponse, error)
	UploadProfilePicture(ctx context.Context, userID uint, file *multipart.FileHeader) (*dto.UploadResponse, error)
	SearchUsers(ctx context.Context, query string, limit, offset int) ([]*dto.UserResponse, error)
	GetUserByID(ctx context.Context, id uint) (*dto.UserResponse, error)
}

type userService struct {
	userRepo       repositories.UserRepository
	storageService storage.StorageService
	logger         logger.Logger
}

func NewUserService(
	userRepo repositories.UserRepository,
	storageService storage.StorageService,
	logger logger.Logger,
) UserService {
	return &userService{
		userRepo:       userRepo,
		storageService: storageService,
		logger:         logger,
	}
}

func (s *userService) GetProfile(ctx context.Context, userID uint) (*dto.UserProfileResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		s.logger.Error("Failed to get user", "error", err)
		return nil, errors.New("failed to get user profile")
	}

	// Generate presigned URL for profile picture
	profilePictureURL := ""
	if user.ProfilePicture != "" {
		presignedURL, err := s.storageService.GeneratePresignedURL(user.ProfilePicture, 24*time.Hour)
		if err != nil {
			s.logger.Error("Failed to generate presigned URL for profile picture",
				"user_id", userID,
				"profile_picture_key", user.ProfilePicture,
				"error", err.Error())
			// Don't fail the entire request, just log the error and return empty URL
		} else {
			profilePictureURL = presignedURL
		}
	}

	return &dto.UserProfileResponse{
		ID:             user.ID,
		Email:          user.Email,
		Username:       user.Username,
		FullName:       user.FullName,
		ProfilePicture: profilePictureURL,
		Bio:            user.Bio,
		Location:       user.Location,
		Website:        user.Website,
		IsVerified:     user.IsVerified,
		IsPremium:      user.IsPremium,
		CreatedAt:      user.CreatedAt,
	}, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID uint, req *dto.UpdateProfileRequest) (*dto.UserProfileResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		s.logger.Error("Failed to get user", "error", err)
		return nil, errors.New("failed to get user")
	}

	// Update fields only if they are provided
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}
	if req.Location != "" {
		user.Location = req.Location
	}
	if req.Website != "" {
		user.Website = req.Website
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update user", "error", err)
		return nil, errors.New("failed to update profile")
	}

	return s.GetProfile(ctx, userID)
}

func (s *userService) UploadProfilePicture(ctx context.Context, userID uint, file *multipart.FileHeader) (*dto.UploadResponse, error) {
	// Upload new profile picture
	fileKey, err := s.storageService.UploadImage(ctx, file, "profile-pictures")
	if err != nil {
		s.logger.Error("Failed to upload image", "error", err)
		return nil, errors.New("failed to upload image")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", "error", err)
		return nil, errors.New("failed to get user")
	}

	// Delete old profile picture if exists
	if user.ProfilePicture != "" {
		go func() {
			if err := s.storageService.DeleteFile(context.Background(), user.ProfilePicture); err != nil {
				s.logger.Error("Failed to delete old profile picture", "error", err)
			}
		}()
	}

	// Update user with new profile picture key
	user.ProfilePicture = fileKey
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update user profile picture", "error", err)
		return nil, errors.New("failed to update profile picture")
	}

	// Generate presigned URL for immediate use
	presignedURL, err := s.storageService.GeneratePresignedURL(fileKey, 24*time.Hour)
	if err != nil {
		s.logger.Error("Failed to generate presigned URL after upload",
			"file_key", fileKey,
			"error", err.Error())
		return nil, errors.New("failed to generate access URL for uploaded image")
	}

	return &dto.UploadResponse{
		URL: presignedURL,
	}, nil
}

func (s *userService) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*dto.UserResponse, error) {
	users, err := s.userRepo.Search(ctx, query, limit, offset)
	if err != nil {
		s.logger.Error("Failed to search users", "error", err)
		return nil, errors.New("failed to search users")
	}

	var responses []*dto.UserResponse
	for _, user := range users {
		profilePictureURL := ""
		if user.ProfilePicture != "" {
			if presignedURL, err := s.storageService.GeneratePresignedURL(user.ProfilePicture, 24*time.Hour); err == nil {
				profilePictureURL = presignedURL
			} else {
				s.logger.Error("Failed to generate presigned URL for user in search",
					"user_id", user.ID,
					"error", err.Error())
			}
		}

		responses = append(responses, &dto.UserResponse{
			ID:             user.ID,
			Username:       user.Username,
			FullName:       user.FullName,
			ProfilePicture: profilePictureURL,
			Bio:            user.Bio,
			Location:       user.Location,
			Website:        user.Website,
			IsVerified:     user.IsVerified,
			IsPremium:      user.IsPremium,
		})
	}

	return responses, nil
}

func (s *userService) GetUserByID(ctx context.Context, id uint) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		s.logger.Error("Failed to get user", "error", err)
		return nil, errors.New("failed to get user")
	}

	profilePictureURL := ""
	if user.ProfilePicture != "" {
		if presignedURL, err := s.storageService.GeneratePresignedURL(user.ProfilePicture, 24*time.Hour); err == nil {
			profilePictureURL = presignedURL
		} else {
			s.logger.Error("Failed to generate presigned URL for user by ID",
				"user_id", id,
				"error", err.Error())
		}
	}

	return &dto.UserResponse{
		ID:             user.ID,
		Username:       user.Username,
		FullName:       user.FullName,
		ProfilePicture: profilePictureURL,
		Bio:            user.Bio,
		Location:       user.Location,
		Website:        user.Website,
		IsVerified:     user.IsVerified,
		IsPremium:      user.IsPremium,
	}, nil
}
