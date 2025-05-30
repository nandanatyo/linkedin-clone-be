package repository

import (
	"context"
	"linked-clone/internal/domain/entities"
	"linked-clone/internal/domain/repositories"

	"gorm.io/gorm"
)

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) repositories.CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment *entities.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *commentRepository) GetByID(ctx context.Context, id uint) (*entities.Comment, error) {
	var comment entities.Comment
	err := r.db.WithContext(ctx).
		Preload("User").
		First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *commentRepository) GetByPostID(ctx context.Context, postID uint, limit, offset int) ([]*entities.Comment, error) {
	var comments []*entities.Comment
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("post_id = ?", postID).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&comments).Error
	return comments, err
}

func (r *commentRepository) Update(ctx context.Context, comment *entities.Comment) error {
	return r.db.WithContext(ctx).Save(comment).Error
}

func (r *commentRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entities.Comment{}, id).Error
}
