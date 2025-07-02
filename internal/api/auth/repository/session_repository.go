package repository

import (
	"context"
	"linked-clone/internal/domain/entities"
	"linked-clone/internal/domain/repositories"
	"time"

	"gorm.io/gorm"
)

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) repositories.SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *entities.Session) error {

	if session.UserAgent != nil && *session.UserAgent == "" {
		session.UserAgent = nil
	}
	if session.IPAddress != nil && *session.IPAddress == "" {
		session.IPAddress = nil
	}

	return r.db.WithContext(ctx).Create(session).Error
}

func (r *sessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*entities.Session, error) {
	var session entities.Session
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("refresh_token = ? AND status = ? AND expires_at > ?",
			refreshToken, entities.SessionActive, time.Now()).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*entities.Session, error) {
	var session entities.Session
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("token_hash = ? AND status = ? AND expires_at > ?",
			tokenHash, entities.SessionActive, time.Now()).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) GetUserActiveSessions(ctx context.Context, userID uint, limit, offset int) ([]*entities.Session, error) {
	var sessions []*entities.Session
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ? AND expires_at > ?",
			userID, entities.SessionActive, time.Now()).
		Order("last_used_at DESC NULLS LAST, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&sessions).Error
	return sessions, err
}

func (r *sessionRepository) Update(ctx context.Context, session *entities.Session) error {

	if session.UserAgent != nil && *session.UserAgent == "" {
		session.UserAgent = nil
	}
	if session.IPAddress != nil && *session.IPAddress == "" {
		session.IPAddress = nil
	}

	return r.db.WithContext(ctx).Save(session).Error
}

func (r *sessionRepository) UpdateLastUsedAt(ctx context.Context, sessionID uint, lastUsedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&entities.Session{}).
		Where("id = ?", sessionID).
		Update("last_used_at", lastUsedAt).Error
}

func (r *sessionRepository) RevokeSession(ctx context.Context, sessionID uint) error {
	return r.db.WithContext(ctx).Model(&entities.Session{}).
		Where("id = ?", sessionID).
		Update("status", entities.SessionRevoked).Error
}

func (r *sessionRepository) RevokeUserSessions(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&entities.Session{}).
		Where("user_id = ? AND status = ?", userID, entities.SessionActive).
		Update("status", entities.SessionRevoked).Error
}

func (r *sessionRepository) RevokeSessionByToken(ctx context.Context, refreshToken string) error {
	return r.db.WithContext(ctx).Model(&entities.Session{}).
		Where("refresh_token = ?", refreshToken).
		Update("status", entities.SessionRevoked).Error
}

func (r *sessionRepository) DeleteExpiredSessions(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ? OR status = ?", time.Now(), entities.SessionExpired).
		Delete(&entities.Session{}).Error
}

func (r *sessionRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entities.Session{}, id).Error
}
