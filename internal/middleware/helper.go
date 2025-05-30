package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

func GetTraceID(c *gin.Context) string {
	if traceID, exists := c.Get(TraceIDKey); exists {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

func GetSpanID(c *gin.Context) string {
	if spanID, exists := c.Get(SpanIDKey); exists {
		if id, ok := spanID.(string); ok {
			return id
		}
	}
	return ""
}

func GetCorrelationID(c *gin.Context) string {
	if correlationID, exists := c.Get("correlation_id"); exists {
		if id, ok := correlationID.(string); ok {
			return id
		}
	}
	return ""
}

func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

func SetTraceContext(c *gin.Context, traceID, spanID string) {
	c.Set(TraceIDKey, traceID)
	c.Set(SpanIDKey, spanID)

	ctx := context.WithValue(c.Request.Context(), TraceIDKey, traceID)
	ctx = context.WithValue(ctx, SpanIDKey, spanID)
	c.Request = c.Request.WithContext(ctx)
}

func GetUserContext(c *gin.Context) (uint, string, string) {
	userID := GetUserID(c)
	email := GetUserEmail(c)
	username := GetUsername(c)
	return userID, email, username
}

func GetTraceIDFromContext(ctx context.Context) string {
	if traceID := ctx.Value(TraceIDKey); traceID != nil {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

func GetSpanIDFromContext(ctx context.Context) string {
	if spanID := ctx.Value(SpanIDKey); spanID != nil {
		if id, ok := spanID.(string); ok {
			return id
		}
	}
	return ""
}

func GetUserIDFromContext(ctx context.Context) uint {
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(uint); ok {
			return id
		}
	}
	return 0
}
