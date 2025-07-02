package routes

import (
	"github.com/gin-gonic/gin"
	"linked-clone/internal/middleware"
	"os"
	"time"
)

func AuthRoutes(rg *gin.RouterGroup, deps *Dependencies) {
	authMiddleware := middleware.AuthMiddleware(deps.JWTService, deps.Logger)

	_ = middleware.CSRFProtection(os.Getenv("CSRF_SECRET"), deps.Logger)

	auth := rg.Group("/auth")
	{

		auth.POST("/register",
			middleware.RateLimitMiddleware(time.Minute, 5, deps.Logger),
			deps.AuthHandler.Register)

		auth.POST("/login",
			middleware.RateLimitMiddleware(time.Minute, 10, deps.Logger),
			deps.AuthHandler.Login)

		auth.POST("/forgot-password",
			middleware.RateLimitMiddleware(time.Minute, 3, deps.Logger),
			deps.AuthHandler.ForgotPassword)

		auth.POST("/reset-password",
			middleware.RateLimitMiddleware(time.Minute, 5, deps.Logger),
			deps.AuthHandler.ResetPassword)

		auth.POST("/refresh",
			middleware.RateLimitMiddleware(time.Minute, 20, deps.Logger),
			deps.AuthHandler.RefreshToken)

		auth.POST("/verify-email",
			authMiddleware,
			middleware.RateLimitMiddleware(time.Minute, 10, deps.Logger),
			deps.AuthHandler.VerifyEmail)

		auth.POST("/logout",
			authMiddleware,
			middleware.RateLimitMiddleware(time.Minute, 10, deps.Logger),
			deps.AuthHandler.Logout)

		auth.GET("/sessions",
			authMiddleware,
			middleware.RateLimitMiddleware(time.Minute, 30, deps.Logger),
			deps.AuthHandler.GetActiveSessions)

		auth.DELETE("/sessions/:sessionId",
			authMiddleware,
			middleware.RateLimitMiddleware(time.Minute, 20, deps.Logger),
			deps.AuthHandler.RevokeSession)

		auth.DELETE("/sessions",
			authMiddleware,
			middleware.RateLimitMiddleware(time.Minute, 5, deps.Logger),
			deps.AuthHandler.RevokeAllSessions)
	}
}
