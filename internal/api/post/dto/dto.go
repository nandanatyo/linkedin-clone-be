package dto

import "time"

type CreatePostRequest struct {
	Content string `form:"content" validate:"required,min=1,max=2000"`
}

type UpdatePostRequest struct {
	Content string `json:"content" validate:"required,min=1,max=2000"`
}

type PostResponse struct {
	ID        uint      `json:"id"`
	Content   string    `json:"content"`
	ImageURL  string    `json:"image_url,omitempty"`
	LikeCount int       `json:"like_count"`
	User      *UserInfo `json:"user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserInfo struct {
	ID             uint   `json:"id"`
	Username       string `json:"username"`
	FullName       string `json:"full_name"`
	ProfilePicture string `json:"profile_picture,omitempty"`
}

type LikeResponse struct {
	ID     uint `json:"id"`
	UserID uint `json:"user_id"`
	PostID uint `json:"post_id"`
}

type AddCommentRequest struct {
	Content string `json:"content" validate:"required,min=1,max=500"`
}

type CommentResponse struct {
	ID        uint      `json:"id"`
	Content   string    `json:"content"`
	User      *UserInfo `json:"user"`
	CreatedAt time.Time `json:"created_at"`
}
