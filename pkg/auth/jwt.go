package auth

import (
	"errors"
	"github.com/google/uuid"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

type TokenResponse struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	ExpiresAt        time.Time `json:"expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type JWTService interface {
	GenerateTokens(userID uint, email, username string) (*TokenResponse, error)
	ValidateToken(tokenString string) (*JWTClaims, error)
	ValidateRefreshToken(tokenString string) (*JWTClaims, error)
	RefreshAccessToken(refreshToken string) (*TokenResponse, error)
}

type jwtService struct {
	secretKey          string
	accessTokenExpiry  int
	refreshTokenExpiry int
}

func NewJWTService(secretKey string, accessTokenExpiryHours int) JWTService {
	refreshTokenExpiryDays := 30
	if accessTokenExpiryHours > 24 {
		refreshTokenExpiryDays = accessTokenExpiryHours / 24 * 2
	}

	return &jwtService{
		secretKey:          secretKey,
		accessTokenExpiry:  accessTokenExpiryHours,
		refreshTokenExpiry: refreshTokenExpiryDays,
	}
}

func (s *jwtService) GenerateTokens(userID uint, email, username string) (*TokenResponse, error) {

	accessExpiresAt := time.Now().Add(time.Duration(s.accessTokenExpiry) * time.Hour)
	accessClaims := &JWTClaims{
		UserID:    userID,
		Email:     email,
		Username:  username,
		TokenType: "access",
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

	refreshExpiresAt := time.Now().Add(time.Duration(s.refreshTokenExpiry) * 24 * time.Hour)
	refreshClaims := &JWTClaims{
		UserID:    userID,
		Email:     email,
		Username:  username,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "linkedin-clone",
			Subject:   "refresh_token",
			ID:        uuid.NewString(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.secretKey))
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:      accessTokenString,
		RefreshToken:     refreshTokenString,
		ExpiresAt:        accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
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

func (s *jwtService) ValidateRefreshToken(tokenString string) (*JWTClaims, error) {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, errors.New("invalid token type: expected refresh token")
	}

	return claims, nil
}

func (s *jwtService) RefreshAccessToken(refreshToken string) (*TokenResponse, error) {

	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	return s.GenerateTokens(claims.UserID, claims.Email, claims.Username)
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

func (s *jwtService) GenerateToken(userID uint, email, username string) (*TokenResponse, error) {
	tokens, err := s.GenerateTokens(userID, email, username)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken: tokens.AccessToken,
		ExpiresAt:   tokens.ExpiresAt,
	}, nil
}
