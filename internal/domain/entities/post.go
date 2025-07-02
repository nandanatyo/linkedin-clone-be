package entities

import (
	"gorm.io/gorm"
	"time"
)

type Post struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	ImageURL  string         `json:"image_url,omitempty"`
	LikeCount int            `gorm:"default:0" json:"like_count"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User     User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Likes    []Like    `gorm:"foreignKey:PostID" json:"likes,omitempty"`
	Comments []Comment `gorm:"foreignKey:PostID" json:"comments,omitempty"`
}

type Like struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	PostID    uint           `gorm:"not null" json:"post_id"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Post Post `gorm:"foreignKey:PostID" json:"post,omitempty"`
}

type Comment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	PostID    uint           `gorm:"not null" json:"post_id"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Post Post `gorm:"foreignKey:PostID" json:"post,omitempty"`
}
