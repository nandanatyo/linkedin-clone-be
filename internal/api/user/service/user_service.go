package service

import (
	"context"
	"errors"
	"linked-clone/internal/api/user/dto"
	"linked-clone/internal/domain/entities"
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

	profilePictureURL := ""
	if user.ProfilePicture != "" {
		presignedURL, err := s.storageService.GeneratePresignedURL(user.ProfilePicture, 24*time.Hour)
		if err != nil {
			s.logger.Error("Failed to generate presigned URL for profile picture",
				"user_id", userID,
				"profile_picture_key", user.ProfilePicture,
				"error", err.Error())
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

	if user.ProfilePicture != "" {
		go func() {
			if err := s.storageService.DeleteFile(context.Background(), user.ProfilePicture); err != nil {
				s.logger.Error("Failed to delete old profile picture", "error", err)
			}
		}()
	}

	user.ProfilePicture = fileKey
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update user profile picture", "error", err)
		return nil, errors.New("failed to update profile picture")
	}

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

type ConnectionService interface {
	SendConnectionRequest(ctx context.Context, requesterID, addresseeID uint) (*dto.ConnectionResponse, error)
	AcceptConnectionRequest(ctx context.Context, userID, connectionID uint) (*dto.ConnectionResponse, error)
	RejectConnectionRequest(ctx context.Context, userID, connectionID uint) error
	RemoveConnection(ctx context.Context, userID, connectionID uint) error
	GetUserConnections(ctx context.Context, userID uint, limit, offset int) ([]*dto.ConnectionResponse, error)
	GetConnectionRequests(ctx context.Context, userID uint, limit, offset int) ([]*dto.ConnectionResponse, error)
	GetSentRequests(ctx context.Context, userID uint, limit, offset int) ([]*dto.ConnectionResponse, error)
	GetConnectionStatus(ctx context.Context, userID1, userID2 uint) (*dto.ConnectionResponse, error)
	GetMutualConnections(ctx context.Context, userID1, userID2 uint, limit, offset int) ([]*dto.ConnectionResponse, error)
	BlockUser(ctx context.Context, userID, targetUserID uint) error
	UnblockUser(ctx context.Context, userID, targetUserID uint) error
}

type connectionService struct {
	connectionRepo repositories.ConnectionRepository
	userRepo       repositories.UserRepository
	storageService storage.StorageService
	logger         logger.Logger
}

func NewConnectionService(
	connectionRepo repositories.ConnectionRepository,
	userRepo repositories.UserRepository,
	storageService storage.StorageService,
	logger logger.Logger,
) ConnectionService {
	return &connectionService{
		connectionRepo: connectionRepo,
		userRepo:       userRepo,
		storageService: storageService,
		logger:         logger,
	}
}

func (s *connectionService) SendConnectionRequest(ctx context.Context, requesterID, addresseeID uint) (*dto.ConnectionResponse, error) {

	requester, err := s.userRepo.GetByID(ctx, requesterID)
	if err != nil {
		return nil, errors.New("requester not found")
	}

	addressee, err := s.userRepo.GetByID(ctx, addresseeID)
	if err != nil {
		return nil, errors.New("addressee not found")
	}

	if requesterID == addresseeID {
		return nil, errors.New("cannot send connection request to yourself")
	}

	existingConnection, _ := s.connectionRepo.FindConnection(ctx, requesterID, addresseeID)
	if existingConnection != nil {
		switch existingConnection.Status {
		case entities.ConnectionPending:
			return nil, errors.New("connection request already sent")
		case entities.ConnectionAccepted:
			return nil, errors.New("users are already connected")
		case entities.ConnectionBlocked:
			return nil, errors.New("cannot send connection request")
		}
	}

	connection := &entities.Connection{
		RequesterID: requesterID,
		AddresseeID: addresseeID,
		Status:      entities.ConnectionPending,
		RequestedAt: time.Now(),
	}

	if err := s.connectionRepo.Create(ctx, connection); err != nil {
		s.logger.Error("Failed to create connection request", "error", err)
		return nil, errors.New("failed to send connection request")
	}

	return s.mapConnectionToResponse(connection, requester, addressee), nil
}

func (s *connectionService) AcceptConnectionRequest(ctx context.Context, userID, connectionID uint) (*dto.ConnectionResponse, error) {
	connection, err := s.connectionRepo.GetByID(ctx, connectionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("connection request not found")
		}
		return nil, errors.New("failed to get connection request")
	}

	if connection.AddresseeID != userID {
		return nil, errors.New("unauthorized to accept this connection request")
	}

	if connection.Status != entities.ConnectionPending {
		return nil, errors.New("connection request is not pending")
	}

	now := time.Now()
	connection.Status = entities.ConnectionAccepted
	connection.AcceptedAt = &now

	if err := s.connectionRepo.Update(ctx, connection); err != nil {
		s.logger.Error("Failed to accept connection request", "error", err)
		return nil, errors.New("failed to accept connection request")
	}

	updatedConnection, err := s.connectionRepo.GetByID(ctx, connectionID)
	if err != nil {
		return nil, errors.New("failed to get updated connection")
	}

	return s.mapConnectionToResponse(updatedConnection, &updatedConnection.Requester, &updatedConnection.Addressee), nil
}

func (s *connectionService) RejectConnectionRequest(ctx context.Context, userID, connectionID uint) error {
	connection, err := s.connectionRepo.GetByID(ctx, connectionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("connection request not found")
		}
		return errors.New("failed to get connection request")
	}

	if connection.AddresseeID != userID {
		return errors.New("unauthorized to reject this connection request")
	}

	if connection.Status != entities.ConnectionPending {
		return errors.New("connection request is not pending")
	}

	if err := s.connectionRepo.Delete(ctx, connectionID); err != nil {
		s.logger.Error("Failed to reject connection request", "error", err)
		return errors.New("failed to reject connection request")
	}

	return nil
}

func (s *connectionService) RemoveConnection(ctx context.Context, userID, connectionID uint) error {
	connection, err := s.connectionRepo.GetByID(ctx, connectionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("connection not found")
		}
		return errors.New("failed to get connection")
	}

	if connection.RequesterID != userID && connection.AddresseeID != userID {
		return errors.New("unauthorized to remove this connection")
	}

	if err := s.connectionRepo.Delete(ctx, connectionID); err != nil {
		s.logger.Error("Failed to remove connection", "error", err)
		return errors.New("failed to remove connection")
	}

	return nil
}

func (s *connectionService) GetUserConnections(ctx context.Context, userID uint, limit, offset int) ([]*dto.ConnectionResponse, error) {
	connections, err := s.connectionRepo.GetUserConnections(ctx, userID, entities.ConnectionAccepted, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get user connections", "error", err)
		return nil, errors.New("failed to get connections")
	}

	var responses []*dto.ConnectionResponse
	for _, connection := range connections {
		responses = append(responses, s.mapConnectionToResponse(connection, &connection.Requester, &connection.Addressee))
	}

	return responses, nil
}

func (s *connectionService) GetConnectionRequests(ctx context.Context, userID uint, limit, offset int) ([]*dto.ConnectionResponse, error) {
	connections, err := s.connectionRepo.GetConnectionRequests(ctx, userID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get connection requests", "error", err)
		return nil, errors.New("failed to get connection requests")
	}

	var responses []*dto.ConnectionResponse
	for _, connection := range connections {
		responses = append(responses, s.mapConnectionToResponse(connection, &connection.Requester, &connection.Addressee))
	}

	return responses, nil
}

func (s *connectionService) GetSentRequests(ctx context.Context, userID uint, limit, offset int) ([]*dto.ConnectionResponse, error) {
	connections, err := s.connectionRepo.GetSentRequests(ctx, userID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get sent requests", "error", err)
		return nil, errors.New("failed to get sent requests")
	}

	var responses []*dto.ConnectionResponse
	for _, connection := range connections {
		responses = append(responses, s.mapConnectionToResponse(connection, &connection.Requester, &connection.Addressee))
	}

	return responses, nil
}

func (s *connectionService) GetConnectionStatus(ctx context.Context, userID1, userID2 uint) (*dto.ConnectionResponse, error) {
	connection, err := s.connectionRepo.FindConnection(ctx, userID1, userID2)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no connection found")
		}
		return nil, errors.New("failed to get connection status")
	}

	requester, err := s.userRepo.GetByID(ctx, connection.RequesterID)
	if err != nil {
		return nil, errors.New("failed to get requester details")
	}

	addressee, err := s.userRepo.GetByID(ctx, connection.AddresseeID)
	if err != nil {
		return nil, errors.New("failed to get addressee details")
	}

	return s.mapConnectionToResponse(connection, requester, addressee), nil
}

func (s *connectionService) GetMutualConnections(ctx context.Context, userID1, userID2 uint, limit, offset int) ([]*dto.ConnectionResponse, error) {
	connections, err := s.connectionRepo.GetMutualConnections(ctx, userID1, userID2, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get mutual connections", "error", err)
		return nil, errors.New("failed to get mutual connections")
	}

	var responses []*dto.ConnectionResponse
	for _, connection := range connections {
		responses = append(responses, s.mapConnectionToResponse(connection, &connection.Requester, &connection.Addressee))
	}

	return responses, nil
}

func (s *connectionService) BlockUser(ctx context.Context, userID, targetUserID uint) error {
	if userID == targetUserID {
		return errors.New("cannot block yourself")
	}

	_, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return errors.New("target user not found")
	}

	existingConnection, _ := s.connectionRepo.FindConnection(ctx, userID, targetUserID)

	if existingConnection != nil {

		existingConnection.Status = entities.ConnectionBlocked
		if err := s.connectionRepo.Update(ctx, existingConnection); err != nil {
			s.logger.Error("Failed to update connection to blocked", "error", err)
			return errors.New("failed to block user")
		}
	} else {

		connection := &entities.Connection{
			RequesterID: userID,
			AddresseeID: targetUserID,
			Status:      entities.ConnectionBlocked,
			RequestedAt: time.Now(),
		}

		if err := s.connectionRepo.Create(ctx, connection); err != nil {
			s.logger.Error("Failed to create blocked connection", "error", err)
			return errors.New("failed to block user")
		}
	}

	return nil
}

func (s *connectionService) UnblockUser(ctx context.Context, userID, targetUserID uint) error {
	connection, err := s.connectionRepo.FindConnection(ctx, userID, targetUserID)
	if err != nil {
		return errors.New("connection not found")
	}

	if connection.Status != entities.ConnectionBlocked {
		return errors.New("user is not blocked")
	}

	if err := s.connectionRepo.Delete(ctx, connection.ID); err != nil {
		s.logger.Error("Failed to unblock user", "error", err)
		return errors.New("failed to unblock user")
	}

	return nil
}

func (s *connectionService) mapConnectionToResponse(connection *entities.Connection, requester *entities.User, addressee *entities.User) *dto.ConnectionResponse {
	response := &dto.ConnectionResponse{
		ID:          connection.ID,
		RequesterID: connection.RequesterID,
		AddresseeID: connection.AddresseeID,
		Status:      connection.Status,
		RequestedAt: connection.RequestedAt,
		AcceptedAt:  connection.AcceptedAt,
		CreatedAt:   connection.CreatedAt,
	}

	if requester != nil {
		requesterProfilePicture := ""
		if requester.ProfilePicture != "" {
			if presignedURL, err := s.storageService.GeneratePresignedURL(requester.ProfilePicture, 24*time.Hour); err == nil {
				requesterProfilePicture = presignedURL
			}
		}

		response.Requester = &dto.UserResponse{
			ID:             requester.ID,
			Username:       requester.Username,
			FullName:       requester.FullName,
			ProfilePicture: requesterProfilePicture,
			Bio:            requester.Bio,
			Location:       requester.Location,
			Website:        requester.Website,
			IsVerified:     requester.IsVerified,
			IsPremium:      requester.IsPremium,
		}
	}

	if addressee != nil {
		addresseeProfilePicture := ""
		if addressee.ProfilePicture != "" {
			if presignedURL, err := s.storageService.GeneratePresignedURL(addressee.ProfilePicture, 24*time.Hour); err == nil {
				addresseeProfilePicture = presignedURL
			}
		}

		response.Addressee = &dto.UserResponse{
			ID:             addressee.ID,
			Username:       addressee.Username,
			FullName:       addressee.FullName,
			ProfilePicture: addresseeProfilePicture,
			Bio:            addressee.Bio,
			Location:       addressee.Location,
			Website:        addressee.Website,
			IsVerified:     addressee.IsVerified,
			IsPremium:      addressee.IsPremium,
		}
	}

	return response
}
