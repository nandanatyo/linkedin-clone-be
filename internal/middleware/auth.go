package middleware

import (
	"linked-clone/pkg/auth"
	"linked-clone/pkg/logger"
	"linked-clone/pkg/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserIDKey           = "user_id"
	UserEmailKey        = "user_email"
	UsernameKey         = "username"
)

func AuthMiddleware(jwtService auth.JWTService, logger logger.Logger) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "Authorization header required", "")
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			response.Error(c, http.StatusUnauthorized, "Invalid authorization header format", "")
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, BearerPrefix)
		if token == "" {
			response.Error(c, http.StatusUnauthorized, "Token required", "")
			c.Abort()
			return
		}

		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			logger.Error("Token validation failed", "error", err)
			response.Error(c, http.StatusUnauthorized, "Invalid or expired token", err.Error())
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)
		c.Set(UsernameKey, claims.Username)

		c.Next()
	})
}

func RefreshTokenMiddleware(jwtService auth.JWTService, logger logger.Logger) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "Authorization header required", "")
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			response.Error(c, http.StatusUnauthorized, "Invalid authorization header format", "")
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, BearerPrefix)
		if token == "" {
			response.Error(c, http.StatusUnauthorized, "Token required", "")
			c.Abort()
			return
		}

		claims, err := jwtService.ValidateRefreshToken(c, token)
		if err != nil {
			logger.Error("Refresh token validation failed", "error", err)
			response.Error(c, http.StatusUnauthorized, "Invalid or expired refresh token", err.Error())
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)
		c.Set(UsernameKey, claims.Username)

		c.Next()
	})
}

func GetUserID(c *gin.Context) uint {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return 0
	}
	return userID.(uint)
}

func GetUserEmail(c *gin.Context) string {
	email, exists := c.Get(UserEmailKey)
	if !exists {
		return ""
	}
	return email.(string)
}

func GetUsername(c *gin.Context) string {
	username, exists := c.Get(UsernameKey)
	if !exists {
		return ""
	}
	return username.(string)
}
