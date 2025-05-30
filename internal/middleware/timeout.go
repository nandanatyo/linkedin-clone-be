package middleware

import (
	"context"
	"linked-clone/pkg/logger"
	"linked-clone/pkg/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func TimeoutMiddleware(timeout time.Duration, logger logger.Logger) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		finished := make(chan struct{})

		go func() {
			defer close(finished)
			c.Next()
		}()

		select {
		case <-finished:

			return
		case <-ctx.Done():

			logger.Warn("Request timeout", map[string]interface{}{
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
				"timeout":    timeout.String(),
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
				"request_id": c.GetString("request_id"),
			})

			response.Error(c, http.StatusRequestTimeout, "Request timeout", "Request took too long to process")
			c.Abort()
			return
		}
	})
}
