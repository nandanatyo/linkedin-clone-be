package middleware

import (
	"linked-clone/internal/domain/repositories"
	"linked-clone/pkg/logger"
	"linked-clone/pkg/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func PremiumMiddleware(userRepo repositories.UserRepository, logger logger.Logger) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			response.Error(c, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
			c.Abort()
			return
		}

		user, err := userRepo.GetByID(c.Request.Context(), userID)
		if err != nil {
			logger.Error("Failed to get user for premium check", "error", err, "user_id", userID)
			response.Error(c, http.StatusInternalServerError, "Failed to verify premium status", "")
			c.Abort()
			return
		}

		if !user.IsPremium {
			response.Error(c, http.StatusForbidden, "Premium subscription required", "This feature requires a premium subscription")
			c.Abort()
			return
		}

		if user.PremiumUntil != nil && user.PremiumUntil.Before(time.Now()) {
			response.Error(c, http.StatusForbidden, "Premium subscription expired", "Your premium subscription has expired")
			c.Abort()
			return
		}

		c.Next()
	})
}
