package dto

import (
	"linked-clone/internal/domain/entities"
	"time"
)

type CreateJobRequest struct {
	Title           string                   `json:"title" validate:"required,min=5,max=200"`
	Company         string                   `json:"company" validate:"required,min=2,max=100"`
	Location        string                   `json:"location" validate:"required,min=2,max=100"`
	Description     string                   `json:"description" validate:"required,min=50,max=5000"`
	Requirements    string                   `json:"requirements" validate:"omitempty,max=3000"`
	JobType         entities.JobType         `json:"job_type" validate:"required,oneof=full_time part_time contract internship"`
	ExperienceLevel entities.ExperienceLevel `json:"experience_level" validate:"required,oneof=entry mid senior executive"`
	SalaryMin       *int                     `json:"salary_min" validate:"omitempty,min=0"`
	SalaryMax       *int                     `json:"salary_max" validate:"omitempty,min=0"`
}

type UpdateJobRequest struct {
	Title           string                    `json:"title" validate:"omitempty,min=5,max=200"`
	Company         string                    `json:"company" validate:"omitempty,min=2,max=100"`
	Location        string                    `json:"location" validate:"omitempty,min=2,max=100"`
	Description     string                    `json:"description" validate:"omitempty,min=50,max=5000"`
	Requirements    string                    `json:"requirements" validate:"omitempty,max=3000"`
	JobType         *entities.JobType         `json:"job_type" validate:"omitempty,oneof=full_time part_time contract internship"`
	ExperienceLevel *entities.ExperienceLevel `json:"experience_level" validate:"omitempty,oneof=entry mid senior executive"`
	SalaryMin       *int                      `json:"salary_min" validate:"omitempty,min=0"`
	SalaryMax       *int                      `json:"salary_max" validate:"omitempty,min=0"`
	IsActive        *bool                     `json:"is_active"`
}

type JobResponse struct {
	ID               uint                     `json:"id"`
	Title            string                   `json:"title"`
	Company          string                   `json:"company"`
	Location         string                   `json:"location"`
	Description      string                   `json:"description"`
	Requirements     string                   `json:"requirements"`
	JobType          entities.JobType         `json:"job_type"`
	ExperienceLevel  entities.ExperienceLevel `json:"experience_level"`
	SalaryMin        *int                     `json:"salary_min,omitempty"`
	SalaryMax        *int                     `json:"salary_max,omitempty"`
	IsActive         bool                     `json:"is_active"`
	ApplicationCount int                      `json:"application_count"`
	User             *UserInfo                `json:"user"`
	CreatedAt        time.Time                `json:"created_at"`
	UpdatedAt        time.Time                `json:"updated_at"`
}

type ApplyJobRequest struct {
	CoverLetter string `form:"cover_letter" validate:"omitempty,max=2000"`
}

type ApplicationResponse struct {
	ID          uint                       `json:"id"`
	JobID       uint                       `json:"job_id"`
	CoverLetter string                     `json:"cover_letter"`
	ResumeURL   string                     `json:"resume_url,omitempty"`
	Status      entities.ApplicationStatus `json:"status"`
	Job         *JobInfo                   `json:"job,omitempty"`
	User        *UserInfo                  `json:"user,omitempty"`
	AppliedAt   time.Time                  `json:"applied_at"`
	CreatedAt   time.Time                  `json:"created_at"`
}

type JobInfo struct {
	ID      uint   `json:"id"`
	Title   string `json:"title"`
	Company string `json:"company"`
}

type UserInfo struct {
	ID             uint   `json:"id"`
	Username       string `json:"username"`
	FullName       string `json:"full_name"`
	ProfilePicture string `json:"profile_picture,omitempty"`
}
