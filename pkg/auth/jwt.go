package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"linked-clone/internal/domain/entities"
	"linked-clone/internal/domain/repositories"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTClaims struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	TokenType string `json:"token_type"`
	SessionID uint   `json:"session_id,omitempty"`
	jwt.RegisteredClaims
}

type TokenResponse struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	ExpiresAt        time.Time `json:"expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
	SessionID        uint      `json:"session_id,omitempty"`
}

type JWTService interface {
	GenerateTokens(ctx context.Context, userID uint, email, username, userAgent, ipAddress string) (*TokenResponse, error)
	ValidateToken(tokenString string) (*JWTClaims, error)
	ValidateRefreshToken(ctx context.Context, refreshToken string) (*JWTClaims, error)
	RefreshAccessToken(ctx context.Context, refreshToken, userAgent, ipAddress string) (*TokenResponse, error)
	RevokeSession(ctx context.Context, sessionID uint) error
	RevokeUserSessions(ctx context.Context, userID uint) error
	RevokeRefreshToken(ctx context.Context, refreshToken string) error
	GetUserActiveSessions(ctx context.Context, userID uint, limit, offset int) ([]*entities.Session, error)
	CleanupExpiredSessions(ctx context.Context) error
}

type jwtService struct {
	secretKey          string
	accessTokenExpiry  int
	refreshTokenExpiry int
	sessionRepo        repositories.SessionRepository
}

func NewJWTService(secretKey string, accessTokenExpiryHours int, sessionRepo repositories.SessionRepository) JWTService {
	refreshTokenExpiryDays := 30
	if accessTokenExpiryHours > 24 {
		refreshTokenExpiryDays = accessTokenExpiryHours / 24 * 2
	}

	return &jwtService{
		secretKey:          secretKey,
		accessTokenExpiry:  accessTokenExpiryHours,
		refreshTokenExpiry: refreshTokenExpiryDays,
		sessionRepo:        sessionRepo,
	}
}

// pkg/auth/jwt.go - UPDATED GENERATE TOKENS METHOD

func (s *jwtService) GenerateTokens(ctx context.Context, userID uint, email, username, userAgent, ipAddress string) (*TokenResponse, error) {
	// Generate refresh token
	refreshToken, err := s.generateSecureToken()
	if err != nil {
		return nil, err
	}

	// Create token hash for storage
	tokenHash := s.hashToken(refreshToken)

	// Set expiration times
	accessExpiresAt := time.Now().Add(time.Duration(s.accessTokenExpiry) * time.Hour)
	refreshExpiresAt := time.Now().Add(time.Duration(s.refreshTokenExpiry) * 24 * time.Hour)

	// Create session record
	session := &entities.Session{
		UserID:       userID,
		RefreshToken: refreshToken,
		TokenHash:    tokenHash,
		Status:       entities.SessionActive,
		ExpiresAt:    refreshExpiresAt,
		LastUsedAt:   nil, // Will be set on first use
	}

	// Set user agent and IP address safely using helper methods
	session.SetUserAgent(userAgent)
	session.SetIPAddress(ipAddress)

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	// Generate access token with session ID
	accessClaims := &JWTClaims{
		UserID:    userID,
		Email:     email,
		Username:  username,
		TokenType: "access",
		SessionID: session.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "linkedin-clone",
			Subject:   "access_token",
			ID:        uuid.NewString(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.secretKey))
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:      accessTokenString,
		RefreshToken:     refreshToken,
		ExpiresAt:        accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
		SessionID:        session.ID,
	}, nil
}

func (s *jwtService) RefreshAccessToken(ctx context.Context, refreshToken, userAgent, ipAddress string) (*TokenResponse, error) {
	// Validate refresh token and get claims
	claims, err := s.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	// Get session to update user agent and IP if changed
	session, err := s.sessionRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, errors.New("session not found")
	}

	// Update session info if changed
	updated := false

	// Check user agent
	currentUA := ""
	if session.UserAgent != nil {
		currentUA = *session.UserAgent
	}
	if currentUA != userAgent {
		session.SetUserAgent(userAgent)
		updated = true
	}

	// Check IP address
	currentIP := ""
	if session.IPAddress != nil {
		currentIP = *session.IPAddress
	}
	// Normalize IP for comparison
	normalizedIP := ipAddress
	if ipAddress == "::1" {
		normalizedIP = "127.0.0.1"
	}
	if currentIP != normalizedIP && normalizedIP != "" {
		session.SetIPAddress(ipAddress)
		updated = true
	}

	if updated {
		s.sessionRepo.Update(ctx, session)
	}

	// Generate new access token
	accessExpiresAt := time.Now().Add(time.Duration(s.accessTokenExpiry) * time.Hour)

	accessClaims := &JWTClaims{
		UserID:    claims.UserID,
		Email:     claims.Email,
		Username:  claims.Username,
		TokenType: "access",
		SessionID: claims.SessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "linkedin-clone",
			Subject:   "access_token",
			ID:        uuid.NewString(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.secretKey))
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:      accessTokenString,
		RefreshToken:     refreshToken, // Keep the same refresh token
		ExpiresAt:        accessExpiresAt,
		RefreshExpiresAt: session.ExpiresAt,
		SessionID:        session.ID,
	}, nil
}

func (s *jwtService) ValidateToken(tokenString string) (*JWTClaims, error) {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "access" {
		return nil, errors.New("invalid token type: expected access token")
	}

	return claims, nil
}

func (s *jwtService) ValidateRefreshToken(ctx context.Context, refreshToken string) (*JWTClaims, error) {

	session, err := s.sessionRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if session.Status != entities.SessionActive {
		return nil, errors.New("session is not active")
	}

	if session.ExpiresAt.Before(time.Now()) {

		session.Status = entities.SessionExpired
		s.sessionRepo.Update(ctx, session)
		return nil, errors.New("refresh token expired")
	}

	now := time.Now()
	s.sessionRepo.UpdateLastUsedAt(ctx, session.ID, now)

	return &JWTClaims{
		UserID:    session.UserID,
		Email:     session.User.Email,
		Username:  session.User.Username,
		TokenType: "refresh",
		SessionID: session.ID,
	}, nil
}

func (s *jwtService) RevokeSession(ctx context.Context, sessionID uint) error {
	return s.sessionRepo.RevokeSession(ctx, sessionID)
}

func (s *jwtService) RevokeUserSessions(ctx context.Context, userID uint) error {
	return s.sessionRepo.RevokeUserSessions(ctx, userID)
}

func (s *jwtService) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	return s.sessionRepo.RevokeSessionByToken(ctx, refreshToken)
}

func (s *jwtService) GetUserActiveSessions(ctx context.Context, userID uint, limit, offset int) ([]*entities.Session, error) {
	return s.sessionRepo.GetUserActiveSessions(ctx, userID, limit, offset)
}

func (s *jwtService) CleanupExpiredSessions(ctx context.Context) error {
	return s.sessionRepo.DeleteExpiredSessions(ctx)
}

func (s *jwtService) parseToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *jwtService) generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *jwtService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
