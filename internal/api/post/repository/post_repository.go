package repository

import (
	"context"
	"linked-clone/internal/domain/entities"
	"linked-clone/internal/domain/repositories"

	"gorm.io/gorm"
)

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) repositories.PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(ctx context.Context, post *entities.Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *postRepository) GetByID(ctx context.Context, id uint) (*entities.Post, error) {
	var post entities.Post
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Comments").
		Preload("Comments.User").
		First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) GetByUserID(ctx context.Context, userID uint, limit, offset int) ([]*entities.Post, error) {
	var posts []*entities.Post
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error
	return posts, err
}

func (r *postRepository) GetFeed(ctx context.Context, userID uint, limit, offset int) ([]*entities.Post, error) {
	var posts []*entities.Post
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error
	return posts, err
}

func (r *postRepository) Update(ctx context.Context, post *entities.Post) error {
	return r.db.WithContext(ctx).Save(post).Error
}

func (r *postRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entities.Post{}, id).Error
}

func (r *postRepository) IncrementLikeCount(ctx context.Context, postID uint) error {
	return r.db.WithContext(ctx).Model(&entities.Post{}).
		Where("id = ?", postID).
		Update("like_count", gorm.Expr("like_count + 1")).Error
}

func (r *postRepository) DecrementLikeCount(ctx context.Context, postID uint) error {
	return r.db.WithContext(ctx).Model(&entities.Post{}).
		Where("id = ?", postID).
		Update("like_count", gorm.Expr("like_count - 1")).Error
}
