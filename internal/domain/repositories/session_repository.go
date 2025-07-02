package repositories

import (
	"context"
	"linked-clone/internal/domain/entities"
	"time"
)

type SessionRepository interface {
	Create(ctx context.Context, session *entities.Session) error
	GetByRefreshToken(ctx context.Context, refreshToken string) (*entities.Session, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*entities.Session, error)
	GetUserActiveSessions(ctx context.Context, userID uint, limit, offset int) ([]*entities.Session, error)
	Update(ctx context.Context, session *entities.Session) error
	UpdateLastUsedAt(ctx context.Context, sessionID uint, lastUsedAt time.Time) error
	RevokeSession(ctx context.Context, sessionID uint) error
	RevokeUserSessions(ctx context.Context, userID uint) error
	RevokeSessionByToken(ctx context.Context, refreshToken string) error
	DeleteExpiredSessions(ctx context.Context) error
	Delete(ctx context.Context, id uint) error
}
