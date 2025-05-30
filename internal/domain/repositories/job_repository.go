package repositories

import (
	"context"
	"linked-clone/internal/domain/entities"
)

type JobRepository interface {
	Create(ctx context.Context, job *entities.Job) error
	GetByID(ctx context.Context, id uint) (*entities.Job, error)
	GetByUserID(ctx context.Context, userID uint, limit, offset int) ([]*entities.Job, error)
	GetAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*entities.Job, error)
	Update(ctx context.Context, job *entities.Job) error
	Delete(ctx context.Context, id uint) error
	IncrementApplicationCount(ctx context.Context, jobID uint) error
	Search(ctx context.Context, query string, filters map[string]interface{}, limit, offset int) ([]*entities.Job, error)
}

type ApplicationRepository interface {
	Create(ctx context.Context, application *entities.Application) error
	GetByID(ctx context.Context, id uint) (*entities.Application, error)
	GetByUserID(ctx context.Context, userID uint, limit, offset int) ([]*entities.Application, error)
	GetByJobID(ctx context.Context, jobID uint, limit, offset int) ([]*entities.Application, error)
	Update(ctx context.Context, application *entities.Application) error
	Delete(ctx context.Context, id uint) error
	FindByUserAndJob(ctx context.Context, userID, jobID uint) (*entities.Application, error)
}
