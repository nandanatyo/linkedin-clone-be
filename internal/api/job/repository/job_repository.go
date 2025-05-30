package repository

import (
	"context"
	"linked-clone/internal/domain/entities"
	"linked-clone/internal/domain/repositories"

	"gorm.io/gorm"
)

type jobRepository struct {
	db *gorm.DB
}

func NewJobRepository(db *gorm.DB) repositories.JobRepository {
	return &jobRepository{db: db}
}

func (r *jobRepository) Create(ctx context.Context, job *entities.Job) error {
	return r.db.WithContext(ctx).Create(job).Error
}

func (r *jobRepository) GetByID(ctx context.Context, id uint) (*entities.Job, error) {
	var job entities.Job
	err := r.db.WithContext(ctx).
		Preload("User").
		First(&job, id).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *jobRepository) GetByUserID(ctx context.Context, userID uint, limit, offset int) ([]*entities.Job, error) {
	var jobs []*entities.Job
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&jobs).Error
	return jobs, err
}

func (r *jobRepository) GetAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*entities.Job, error) {
	var jobs []*entities.Job
	query := r.db.WithContext(ctx).Preload("User").Where("is_active = true")

	for key, value := range filters {
		switch key {
		case "job_type":
			query = query.Where("job_type = ?", value)
		case "experience_level":
			query = query.Where("experience_level = ?", value)
		case "location":
			query = query.Where("location I LIKE ?", "%"+value.(string)+"%")
		}
	}

	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&jobs).Error
	return jobs, err
}

func (r *jobRepository) Update(ctx context.Context, job *entities.Job) error {
	return r.db.WithContext(ctx).Save(job).Error
}

func (r *jobRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entities.Job{}, id).Error
}

func (r *jobRepository) IncrementApplicationCount(ctx context.Context, jobID uint) error {
	return r.db.WithContext(ctx).Model(&entities.Job{}).
		Where("id = ?", jobID).
		Update("application_count", gorm.Expr("application_count + 1")).Error
}

func (r *jobRepository) Search(ctx context.Context, query string, filters map[string]interface{}, limit, offset int) ([]*entities.Job, error) {
	var jobs []*entities.Job
	dbQuery := r.db.WithContext(ctx).
		Preload("User").
		Where("is_active = true").
		Where("title I LIKE ? OR company I LIKE ? OR description I LIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%")

	for key, value := range filters {
		switch key {
		case "job_type":
			dbQuery = dbQuery.Where("job_type = ?", value)
		case "experience_level":
			dbQuery = dbQuery.Where("experience_level = ?", value)
		case "location":
			dbQuery = dbQuery.Where("location I LIKE ?", "%"+value.(string)+"%")
		}
	}

	err := dbQuery.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&jobs).Error
	return jobs, err
}
