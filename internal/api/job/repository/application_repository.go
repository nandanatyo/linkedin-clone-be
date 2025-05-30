package repository

import (
	"context"
	"linked-clone/internal/domain/entities"
	"linked-clone/internal/domain/repositories"

	"gorm.io/gorm"
)

type applicationRepository struct {
	db *gorm.DB
}

func NewApplicationRepository(db *gorm.DB) repositories.ApplicationRepository {
	return &applicationRepository{db: db}
}

func (r *applicationRepository) Create(ctx context.Context, application *entities.Application) error {
	return r.db.WithContext(ctx).Create(application).Error
}

func (r *applicationRepository) GetByID(ctx context.Context, id uint) (*entities.Application, error) {
	var application entities.Application
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Job").
		First(&application, id).Error
	if err != nil {
		return nil, err
	}
	return &application, nil
}

func (r *applicationRepository) GetByUserID(ctx context.Context, userID uint, limit, offset int) ([]*entities.Application, error) {
	var applications []*entities.Application
	err := r.db.WithContext(ctx).
		Preload("Job").
		Preload("Job.User").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&applications).Error
	return applications, err
}

func (r *applicationRepository) GetByJobID(ctx context.Context, jobID uint, limit, offset int) ([]*entities.Application, error) {
	var applications []*entities.Application
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("job_id = ?", jobID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&applications).Error
	return applications, err
}

func (r *applicationRepository) Update(ctx context.Context, application *entities.Application) error {
	return r.db.WithContext(ctx).Save(application).Error
}

func (r *applicationRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entities.Application{}, id).Error
}

func (r *applicationRepository) FindByUserAndJob(ctx context.Context, userID, jobID uint) (*entities.Application, error) {
	var application entities.Application
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND job_id = ?", userID, jobID).
		First(&application).Error
	if err != nil {
		return nil, err
	}
	return &application, nil
}
