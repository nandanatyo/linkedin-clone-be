package entities

import (
	"gorm.io/gorm"
	"time"
)

type ApplicationStatus string

const (
	ApplicationPending  ApplicationStatus = "pending"
	ApplicationReviewed ApplicationStatus = "reviewed"
	ApplicationAccepted ApplicationStatus = "accepted"
	ApplicationRejected ApplicationStatus = "rejected"
)

type Application struct {
	ID          uint              `gorm:"primaryKey" json:"id"`
	UserID      uint              `gorm:"not null" json:"user_id"`
	JobID       uint              `gorm:"not null" json:"job_id"`
	CoverLetter string            `gorm:"type:text" json:"cover_letter"`
	ResumeURL   string            `json:"resume_url,omitempty"`
	Status      ApplicationStatus `gorm:"default:'pending'" json:"status"`
	AppliedAt   time.Time         `json:"applied_at"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	DeletedAt   gorm.DeletedAt    `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Job  Job  `gorm:"foreignKey:JobID" json:"job,omitempty"`
}
