package repositories

import (
	"context"
	"linked-clone/internal/domain/entities"
)

type PostRepository interface {
	Create(ctx context.Context, post *entities.Post) error
	GetByID(ctx context.Context, id uint) (*entities.Post, error)
	GetByUserID(ctx context.Context, userID uint, limit, offset int) ([]*entities.Post, error)
	GetFeed(ctx context.Context, userID uint, limit, offset int) ([]*entities.Post, error)
	Update(ctx context.Context, post *entities.Post) error
	Delete(ctx context.Context, id uint) error
	IncrementLikeCount(ctx context.Context, postID uint) error
	DecrementLikeCount(ctx context.Context, postID uint) error
}

type LikeRepository interface {
	Create(ctx context.Context, like *entities.Like) error
	Delete(ctx context.Context, userID, postID uint) error
	FindByUserAndPost(ctx context.Context, userID, postID uint) (*entities.Like, error)
	GetPostLikes(ctx context.Context, postID uint, limit, offset int) ([]*entities.Like, error)
}

type CommentRepository interface {
	Create(ctx context.Context, comment *entities.Comment) error
	GetByID(ctx context.Context, id uint) (*entities.Comment, error)
	GetByPostID(ctx context.Context, postID uint, limit, offset int) ([]*entities.Comment, error)
	Update(ctx context.Context, comment *entities.Comment) error
	Delete(ctx context.Context, id uint) error
}
