package middleware

import (
	"bytes"
	"io"
	"linked-clone/pkg/logger"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func LoggerMiddleware(logger logger.Logger) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()

		var requestBody []byte
		if c.Request.Body != nil && !containsSensitiveData(c.Request.URL.Path) {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		w := &responseWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w

		c.Next()

		latency := time.Since(start)
		requestID := c.GetString("request_id")

		fields := map[string]interface{}{
			"request_id":    requestID,
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"query":         c.Request.URL.RawQuery,
			"status":        c.Writer.Status(),
			"latency":       latency.String(),
			"latency_ms":    latency.Milliseconds(),
			"user_agent":    c.Request.UserAgent(),
			"ip":            c.ClientIP(),
			"response_size": w.body.Len(),
		}

		if userID := GetUserID(c); userID != 0 {
			fields["user_id"] = userID
		}

		if len(requestBody) > 0 && len(requestBody) < 1000 {
			fields["request_body"] = string(requestBody)
		}

		if c.Writer.Status() >= 400 && w.body.Len() < 1000 {
			fields["response_body"] = w.body.String()
		}

		if c.Writer.Status() >= 500 {
			logger.Error("HTTP Request - Server Error", fields)
		} else if c.Writer.Status() >= 400 {
			logger.Warn("HTTP Request - Client Error", fields)
		} else {
			logger.Info("HTTP Request", fields)
		}
	})
}

func containsSensitiveData(path string) bool {
	sensitivePaths := []string{
		"/auth/login",
		"/auth/register",
		"/auth/reset-password",
		"/auth/forgot-password",
		"/auth/verify-email",
	}

	for _, sensitivePath := range sensitivePaths {
		if strings.Contains(strings.ToLower(path), sensitivePath) {
			return true
		}
	}
	return false
}
