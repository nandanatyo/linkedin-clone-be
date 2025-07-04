package dto

import (
	"linked-clone/internal/domain/entities"
	"time"
)

type UpdateProfileRequest struct {
	FullName string `json:"full_name" validate:"omitempty,min=2,max=100"`
	Bio      string `json:"bio" validate:"omitempty,max=500"`
	Location string `json:"location" validate:"omitempty,max=100"`
	Website  string `json:"website" validate:"omitempty,url"`
}

type UserProfileResponse struct {
	ID             uint      `json:"id"`
	Email          string    `json:"email"`
	Username       string    `json:"username"`
	FullName       string    `json:"full_name"`
	ProfilePicture string    `json:"profile_picture,omitempty"`
	Bio            string    `json:"bio,omitempty"`
	Location       string    `json:"location,omitempty"`
	Website        string    `json:"website,omitempty"`
	IsVerified     bool      `json:"is_verified"`
	IsPremium      bool      `json:"is_premium"`
	CreatedAt      time.Time `json:"created_at"`
}

type UserResponse struct {
	ID             uint   `json:"id"`
	Username       string `json:"username"`
	FullName       string `json:"full_name"`
	ProfilePicture string `json:"profile_picture,omitempty"`
	Bio            string `json:"bio,omitempty"`
	Location       string `json:"location,omitempty"`
	Website        string `json:"website,omitempty"`
	IsVerified     bool   `json:"is_verified"`
	IsPremium      bool   `json:"is_premium"`
}

type UploadResponse struct {
	URL string `json:"url"`
}

type ConnectionRequest struct {
	UserID uint `json:"user_id" validate:"required"`
}

type ConnectionResponse struct {
	ID          uint                      `json:"id"`
	RequesterID uint                      `json:"requester_id"`
	AddresseeID uint                      `json:"addressee_id"`
	Status      entities.ConnectionStatus `json:"status"`
	RequestedAt time.Time                 `json:"requested_at"`
	AcceptedAt  *time.Time                `json:"accepted_at,omitempty"`
	Requester   *UserResponse             `json:"requester,omitempty"`
	Addressee   *UserResponse             `json:"addressee,omitempty"`
	CreatedAt   time.Time                 `json:"created_at"`
}

type ConnectionStatusUpdate struct {
	Status entities.ConnectionStatus `json:"status" validate:"required,oneof=accepted blocked"`
}
