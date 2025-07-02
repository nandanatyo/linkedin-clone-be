package dto

import "time"

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=30,alphanum"`
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type VerifyEmailRequest struct {
	UserID uint   `json:"-"`
	Code   string `json:"code" validate:"required,len=6"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Code        string `json:"code" validate:"required,len=6"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type AuthResponse struct {
	User             *UserResponse `json:"user"`
	AccessToken      string        `json:"access_token"`
	RefreshToken     string        `json:"refresh_token"`
	ExpiresAt        time.Time     `json:"expires_at"`
	RefreshExpiresAt time.Time     `json:"refresh_expires_at"`
}

type TokenResponse struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	ExpiresAt        time.Time `json:"expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type UserResponse struct {
	ID             uint   `json:"id"`
	Email          string `json:"email"`
	Username       string `json:"username"`
	FullName       string `json:"full_name"`
	ProfilePicture string `json:"profile_picture,omitempty"`
	Bio            string `json:"bio,omitempty"`
	IsVerified     bool   `json:"is_verified"`
	IsPremium      bool   `json:"is_premium"`
}
