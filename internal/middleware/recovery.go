package middleware

import (
	"linked-clone/pkg/logger"
	"linked-clone/pkg/response"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func RecoveryMiddleware(logger logger.Logger) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered", map[string]interface{}{
					"error":      err,
					"stack":      string(debug.Stack()),
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
					"ip":         c.ClientIP(),
					"user_agent": c.Request.UserAgent(),
					"request_id": c.GetString("request_id"),
				})

				response.Error(c, http.StatusInternalServerError, "Internal server error", "An unexpected error occurred")
				c.Abort()
			}
		}()

		c.Next()
	})
}
