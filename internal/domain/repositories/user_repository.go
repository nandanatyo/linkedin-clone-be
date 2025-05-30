package repositories

import (
	"context"
	"linked-clone/internal/domain/entities"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id uint) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	GetByUsername(ctx context.Context, username string) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id uint) error
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.User, error)
	VerifyEmail(ctx context.Context, userID uint) error
	UpdatePremiumStatus(ctx context.Context, userID uint, isPremium bool, premiumUntil *time.Time) error
}
