package entities

import (
	"gorm.io/gorm"
	"time"
)

type SessionStatus string

const (
	SessionActive  SessionStatus = "active"
	SessionRevoked SessionStatus = "revoked"
	SessionExpired SessionStatus = "expired"
)

type Session struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"not null" json:"user_id"`
	RefreshToken string         `gorm:"not null;unique" json:"refresh_token"`
	TokenHash    string         `gorm:"not null" json:"-"`
	Status       SessionStatus  `gorm:"default:'active'" json:"status"`
	UserAgent    *string        `json:"user_agent,omitempty"`
	IPAddress    *string        `gorm:"type:inet" json:"ip_address,omitempty"`
	ExpiresAt    time.Time      `gorm:"not null" json:"expires_at"`
	LastUsedAt   *time.Time     `json:"last_used_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Session) TableName() string {
	return "sessions"
}

func (s *Session) SetIPAddress(ip string) {
	if ip != "" && ip != "::1" && ip != "127.0.0.1" {
		s.IPAddress = &ip
	} else if ip == "::1" {

		localhost := "127.0.0.1"
		s.IPAddress = &localhost
	} else if ip == "127.0.0.1" {
		s.IPAddress = &ip
	}

}

func (s *Session) SetUserAgent(ua string) {
	if ua != "" {
		s.UserAgent = &ua
	}

}
