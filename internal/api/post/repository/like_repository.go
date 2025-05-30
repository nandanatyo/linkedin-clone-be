package repository

import (
	"context"
	"linked-clone/internal/domain/entities"
	"linked-clone/internal/domain/repositories"

	"gorm.io/gorm"
)

type likeRepository struct {
	db *gorm.DB
}

func NewLikeRepository(db *gorm.DB) repositories.LikeRepository {
	return &likeRepository{db: db}
}

func (r *likeRepository) Create(ctx context.Context, like *entities.Like) error {
	return r.db.WithContext(ctx).Create(like).Error
}

func (r *likeRepository) Delete(ctx context.Context, userID, postID uint) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Delete(&entities.Like{}).Error
}

func (r *likeRepository) FindByUserAndPost(ctx context.Context, userID, postID uint) (*entities.Like, error) {
	var like entities.Like
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND post_id = ?", userID, postID).
		First(&like).Error
	if err != nil {
		return nil, err
	}
	return &like, nil
}

func (r *likeRepository) GetPostLikes(ctx context.Context, postID uint, limit, offset int) ([]*entities.Like, error) {
	var likes []*entities.Like
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("post_id = ?", postID).
		Limit(limit).
		Offset(offset).
		Find(&likes).Error
	return likes, err
}
