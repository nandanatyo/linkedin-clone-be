package service

import (
	"context"
	"errors"
	"linked-clone/internal/api/job/dto"
	"linked-clone/internal/domain/entities"
	"linked-clone/internal/domain/repositories"
	"linked-clone/pkg/logger"

	"linked-clone/pkg/storage"
	"mime/multipart"
	"time"

	"gorm.io/gorm"
)

type JobService interface {
	CreateJob(ctx context.Context, userID uint, req *dto.CreateJobRequest) (*dto.JobResponse, error)
	GetJob(ctx context.Context, id uint) (*dto.JobResponse, error)
	GetUserJobs(ctx context.Context, userID uint, limit, offset int) ([]*dto.JobResponse, error)
	GetAllJobs(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*dto.JobResponse, error)
	UpdateJob(ctx context.Context, userID, jobID uint, req *dto.UpdateJobRequest) (*dto.JobResponse, error)
	DeleteJob(ctx context.Context, userID, jobID uint) error
	SearchJobs(ctx context.Context, query string, filters map[string]interface{}, limit, offset int) ([]*dto.JobResponse, error)
	ApplyJob(ctx context.Context, userID, jobID uint, req *dto.ApplyJobRequest, resume *multipart.FileHeader) (*dto.ApplicationResponse, error)
	GetUserApplications(ctx context.Context, userID uint, limit, offset int) ([]*dto.ApplicationResponse, error)
	GetJobApplications(ctx context.Context, userID, jobID uint, limit, offset int) ([]*dto.ApplicationResponse, error)
}

type jobService struct {
	jobRepo         repositories.JobRepository
	applicationRepo repositories.ApplicationRepository
	userRepo        repositories.UserRepository
	storageService  storage.StorageService
	logger          logger.Logger
}

func NewJobService(
	jobRepo repositories.JobRepository,
	applicationRepo repositories.ApplicationRepository,
	userRepo repositories.UserRepository,
	storageService storage.StorageService,
	logger logger.Logger,
) JobService {
	return &jobService{
		jobRepo:         jobRepo,
		applicationRepo: applicationRepo,
		userRepo:        userRepo,
		storageService:  storageService,
		logger:          logger,
	}
}

func (s *jobService) CreateJob(ctx context.Context, userID uint, req *dto.CreateJobRequest) (*dto.JobResponse, error) {
	job := &entities.Job{
		UserID:          userID,
		Title:           req.Title,
		Company:         req.Company,
		Location:        req.Location,
		Description:     req.Description,
		Requirements:    req.Requirements,
		JobType:         req.JobType,
		ExperienceLevel: req.ExperienceLevel,
		SalaryMin:       req.SalaryMin,
		SalaryMax:       req.SalaryMax,
		IsActive:        true,
	}

	if err := s.jobRepo.Create(ctx, job); err != nil {
		s.logger.Error("Failed to create job", "error", err)
		return nil, errors.New("failed to create job")
	}

	return s.GetJob(ctx, job.ID)
}

func (s *jobService) GetJob(ctx context.Context, id uint) (*dto.JobResponse, error) {
	job, err := s.jobRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("job not found")
		}
		return nil, errors.New("failed to get job")
	}

	return s.mapJobToResponse(job), nil
}

func (s *jobService) GetUserJobs(ctx context.Context, userID uint, limit, offset int) ([]*dto.JobResponse, error) {
	jobs, err := s.jobRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, errors.New("failed to get user jobs")
	}

	var responses []*dto.JobResponse
	for _, job := range jobs {
		responses = append(responses, s.mapJobToResponse(job))
	}

	return responses, nil
}

func (s *jobService) GetAllJobs(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*dto.JobResponse, error) {
	jobs, err := s.jobRepo.GetAll(ctx, filters, limit, offset)
	if err != nil {
		return nil, errors.New("failed to get jobs")
	}

	var responses []*dto.JobResponse
	for _, job := range jobs {
		responses = append(responses, s.mapJobToResponse(job))
	}

	return responses, nil
}

func (s *jobService) UpdateJob(ctx context.Context, userID, jobID uint, req *dto.UpdateJobRequest) (*dto.JobResponse, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("job not found")
		}
		return nil, errors.New("failed to get job")
	}

	if job.UserID != userID {
		return nil, errors.New("unauthorized to update this job")
	}

	if req.Title != "" {
		job.Title = req.Title
	}
	if req.Company != "" {
		job.Company = req.Company
	}
	if req.Location != "" {
		job.Location = req.Location
	}
	if req.Description != "" {
		job.Description = req.Description
	}
	if req.Requirements != "" {
		job.Requirements = req.Requirements
	}
	if req.JobType != nil {
		job.JobType = *req.JobType
	}
	if req.ExperienceLevel != nil {
		job.ExperienceLevel = *req.ExperienceLevel
	}
	if req.SalaryMin != nil {
		job.SalaryMin = req.SalaryMin
	}
	if req.SalaryMax != nil {
		job.SalaryMax = req.SalaryMax
	}
	if req.IsActive != nil {
		job.IsActive = *req.IsActive
	}

	if err := s.jobRepo.Update(ctx, job); err != nil {
		return nil, errors.New("failed to update job")
	}

	return s.GetJob(ctx, jobID)
}

func (s *jobService) DeleteJob(ctx context.Context, userID, jobID uint) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("job not found")
		}
		return errors.New("failed to get job")
	}

	if job.UserID != userID {
		return errors.New("unauthorized to delete this job")
	}

	if err := s.jobRepo.Delete(ctx, jobID); err != nil {
		return errors.New("failed to delete job")
	}

	return nil
}

func (s *jobService) SearchJobs(ctx context.Context, query string, filters map[string]interface{}, limit, offset int) ([]*dto.JobResponse, error) {
	jobs, err := s.jobRepo.Search(ctx, query, filters, limit, offset)
	if err != nil {
		return nil, errors.New("failed to search jobs")
	}

	var responses []*dto.JobResponse
	for _, job := range jobs {
		responses = append(responses, s.mapJobToResponse(job))
	}

	return responses, nil
}

func (s *jobService) ApplyJob(ctx context.Context, userID, jobID uint, req *dto.ApplyJobRequest, resume *multipart.FileHeader) (*dto.ApplicationResponse, error) {

	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("job not found")
		}
		return nil, errors.New("failed to get job")
	}

	if !job.IsActive {
		return nil, errors.New("job is not active")
	}

	existingApp, _ := s.applicationRepo.FindByUserAndJob(ctx, userID, jobID)
	if existingApp != nil {
		return nil, errors.New("already applied to this job")
	}

	var resumeURL string
	if resume != nil {
		url, err := s.storageService.UploadFile(ctx, resume, "resumes")
		if err != nil {
			return nil, errors.New("failed to upload resume")
		}
		resumeURL = url
	}

	application := &entities.Application{
		UserID:      userID,
		JobID:       jobID,
		CoverLetter: req.CoverLetter,
		ResumeURL:   resumeURL,
		Status:      entities.ApplicationPending,
		AppliedAt:   time.Now(),
	}

	if err := s.applicationRepo.Create(ctx, application); err != nil {
		return nil, errors.New("failed to create application")
	}

	s.jobRepo.IncrementApplicationCount(ctx, jobID)

	fullApp, err := s.applicationRepo.GetByID(ctx, application.ID)
	if err != nil {
		return nil, errors.New("failed to get application")
	}

	return s.mapApplicationToResponse(fullApp), nil
}

func (s *jobService) GetUserApplications(ctx context.Context, userID uint, limit, offset int) ([]*dto.ApplicationResponse, error) {
	applications, err := s.applicationRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, errors.New("failed to get applications")
	}

	var responses []*dto.ApplicationResponse
	for _, app := range applications {
		responses = append(responses, s.mapApplicationToResponse(app))
	}

	return responses, nil
}

func (s *jobService) GetJobApplications(ctx context.Context, userID, jobID uint, limit, offset int) ([]*dto.ApplicationResponse, error) {

	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, errors.New("job not found")
	}

	if job.UserID != userID {
		return nil, errors.New("unauthorized to view applications")
	}

	applications, err := s.applicationRepo.GetByJobID(ctx, jobID, limit, offset)
	if err != nil {
		return nil, errors.New("failed to get applications")
	}

	var responses []*dto.ApplicationResponse
	for _, app := range applications {
		responses = append(responses, s.mapApplicationToResponse(app))
	}

	return responses, nil
}

func (s *jobService) mapJobToResponse(job *entities.Job) *dto.JobResponse {
	response := &dto.JobResponse{
		ID:               job.ID,
		Title:            job.Title,
		Company:          job.Company,
		Location:         job.Location,
		Description:      job.Description,
		Requirements:     job.Requirements,
		JobType:          job.JobType,
		ExperienceLevel:  job.ExperienceLevel,
		SalaryMin:        job.SalaryMin,
		SalaryMax:        job.SalaryMax,
		IsActive:         job.IsActive,
		ApplicationCount: job.ApplicationCount,
		CreatedAt:        job.CreatedAt,
		UpdatedAt:        job.UpdatedAt,
	}

	if job.User.ID != 0 {
		response.User = &dto.UserInfo{
			ID:             job.User.ID,
			Username:       job.User.Username,
			FullName:       job.User.FullName,
			ProfilePicture: job.User.ProfilePicture,
		}
	}

	return response
}

func (s *jobService) mapApplicationToResponse(app *entities.Application) *dto.ApplicationResponse {
	response := &dto.ApplicationResponse{
		ID:          app.ID,
		JobID:       app.JobID,
		CoverLetter: app.CoverLetter,
		Status:      app.Status,
		AppliedAt:   app.AppliedAt,
		CreatedAt:   app.CreatedAt,
	}

	if app.ResumeURL != "" {
		url, err := s.storageService.GeneratePresignedURL(app.ResumeURL, 15*time.Minute)
		if err != nil {
			s.logger.Error("Failed to generate resume presigned URL", "error", err)
		} else {
			response.ResumeURL = url
		}
	}

	if app.Job.ID != 0 {
		response.Job = &dto.JobInfo{
			ID:      app.Job.ID,
			Title:   app.Job.Title,
			Company: app.Job.Company,
		}
	}

	if app.User.ID != 0 {
		profilePicture := app.User.ProfilePicture
		if profilePicture != "" {
			url, err := s.storageService.GeneratePresignedURL(profilePicture, 15*time.Minute)
			if err != nil {
				s.logger.Error("Failed to generate profile picture presigned URL", "error", err)
			} else {
				profilePicture = url
			}
		}

		response.User = &dto.UserInfo{
			ID:             app.User.ID,
			Username:       app.User.Username,
			FullName:       app.User.FullName,
			ProfilePicture: profilePicture,
		}
	}

	return response
}
