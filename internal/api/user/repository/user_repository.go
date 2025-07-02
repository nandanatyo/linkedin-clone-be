package repository

import (
	"context"
	"linked-clone/internal/domain/entities"
	"linked-clone/internal/domain/repositories"
	"time"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repositories.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entities.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*entities.User, error) {
	var user entities.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	var user entities.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	var user entities.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *entities.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entities.User{}, id).Error
}

func (r *userRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.User, error) {
	var users []*entities.User
	err := r.db.WithContext(ctx).
		Where("full_name ILIKE ? OR username ILIKE ? OR email ILIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%").
		Limit(limit).
		Offset(offset).
		Find(&users).Error
	return users, err
}

func (r *userRepository) VerifyEmail(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&entities.User{}).
		Where("id = ?", userID).
		Update("is_verified", true).Error
}

func (r *userRepository) UpdatePremiumStatus(ctx context.Context, userID uint, isPremium bool, premiumUntil *time.Time) error {
	updates := map[string]interface{}{
		"is_premium": isPremium,
	}

	if premiumUntil != nil {
		updates["premium_until"] = premiumUntil
	}

	return r.db.WithContext(ctx).Model(&entities.User{}).
		Where("id = ?", userID).
		Updates(updates).Error
}
