package logger

import (
	"time"

	"github.com/gin-gonic/gin"
)

func GinMiddleware(logger Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx := WithLogger(c.Request.Context(), logger)
		c.Request = c.Request.WithContext(ctx)

		if requestID := c.GetString("request_id"); requestID != "" {
			logger = logger.WithField(FieldRequestID, requestID)
		}

		c.Set("logger", logger)

		c.Next()
	}
}

func GetGinLogger(c *gin.Context) Logger {
	if logger, exists := c.Get("logger"); exists {
		if l, ok := logger.(Logger); ok {
			return l
		}
	}
	return NewLogger()
}

func LogHTTPRequest(c *gin.Context, logger Logger, startTime time.Time) {
	latency := time.Since(startTime)

	fields := Fields{
		FieldMethod:     c.Request.Method,
		FieldPath:       c.Request.URL.Path,
		FieldStatusCode: c.Writer.Status(),
		FieldLatency:    latency.String(),
		FieldIP:         c.ClientIP(),
		FieldUserAgent:  c.Request.UserAgent(),
	}

	if userID, exists := c.Get("user_id"); exists {
		fields[FieldUserID] = userID
	}

	if c.Request.URL.RawQuery != "" {
		fields["query"] = c.Request.URL.RawQuery
	}

	logger.WithFields(fields).Info("HTTP request processed")
}
