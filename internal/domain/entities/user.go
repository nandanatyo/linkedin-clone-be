package entities

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Email          string         `gorm:"unique;not null" json:"email"`
	Username       string         `gorm:"unique;not null" json:"username"`
	FullName       string         `gorm:"not null" json:"full_name"`
	Password       string         `gorm:"not null" json:"-"`
	ProfilePicture string         `json:"profile_picture,omitempty"`
	Bio            string         `json:"bio,omitempty"`
	Location       string         `json:"location,omitempty"`
	Website        string         `json:"website,omitempty"`
	IsVerified     bool           `gorm:"default:false" json:"is_verified"`
	IsPremium      bool           `gorm:"default:false" json:"is_premium"`
	PremiumUntil   *time.Time     `json:"premium_until,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Posts        []Post        `gorm:"foreignKey:UserID" json:"posts,omitempty"`
	Jobs         []Job         `gorm:"foreignKey:UserID" json:"jobs,omitempty"`
	Applications []Application `gorm:"foreignKey:UserID" json:"applications,omitempty"`
	Likes        []Like        `gorm:"foreignKey:UserID" json:"-"`
	Comments     []Comment     `gorm:"foreignKey:UserID" json:"-"`
}
