package entities

import (
	"gorm.io/gorm"
	"time"
)

type JobType string
type ExperienceLevel string

const (
	JobTypeFullTime   JobType = "full_time"
	JobTypePartTime   JobType = "part_time"
	JobTypeContract   JobType = "contract"
	JobTypeInternship JobType = "internship"

	ExperienceEntry     ExperienceLevel = "entry"
	ExperienceMid       ExperienceLevel = "mid"
	ExperienceSenior    ExperienceLevel = "senior"
	ExperienceExecutive ExperienceLevel = "executive"
)

type Job struct {
	ID               uint            `gorm:"primaryKey" json:"id"`
	UserID           uint            `gorm:"not null" json:"user_id"`
	Title            string          `gorm:"not null" json:"title"`
	Company          string          `gorm:"not null" json:"company"`
	Location         string          `gorm:"not null" json:"location"`
	Description      string          `gorm:"type:text;not null" json:"description"`
	Requirements     string          `gorm:"type:text" json:"requirements"`
	JobType          JobType         `gorm:"not null" json:"job_type"`
	ExperienceLevel  ExperienceLevel `gorm:"not null" json:"experience_level"`
	SalaryMin        *int            `json:"salary_min,omitempty"`
	SalaryMax        *int            `json:"salary_max,omitempty"`
	IsActive         bool            `gorm:"default:true" json:"is_active"`
	ApplicationCount int             `gorm:"default:0" json:"application_count"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	DeletedAt        gorm.DeletedAt  `gorm:"index" json:"-"`

	User         User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Applications []Application `gorm:"foreignKey:JobID" json:"applications,omitempty"`
}
