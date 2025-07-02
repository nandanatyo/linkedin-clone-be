package middleware

import (
	"context"
	"linked-clone/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	TraceIDKey   = "trace_id"
	SpanIDKey    = "span_id"
	ParentIDKey  = "parent_id"
	ServiceKey   = "service"
	OperationKey = "operation"
)

type TraceSpan struct {
	TraceID   string                 `json:"trace_id"`
	SpanID    string                 `json:"span_id"`
	ParentID  string                 `json:"parent_id,omitempty"`
	Service   string                 `json:"service"`
	Operation string                 `json:"operation"`
	StartTime time.Time              `json:"start_time"`
	EndTime   time.Time              `json:"end_time"`
	Duration  time.Duration          `json:"duration"`
	Tags      map[string]interface{} `json:"tags,omitempty"`
	Logs      []TraceLog             `json:"logs,omitempty"`
	Error     bool                   `json:"error"`
	ErrorMsg  string                 `json:"error_msg,omitempty"`
}

type TraceLog struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

func TracingMiddleware(serviceName string, structuredLogger logger.StructuredLogger) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {

		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = generateTraceID()
		}

		spanID := generateSpanID()

		parentID := c.GetHeader("X-Parent-Span-ID")

		span := &TraceSpan{
			TraceID:   traceID,
			SpanID:    spanID,
			ParentID:  parentID,
			Service:   serviceName,
			Operation: c.Request.Method + " " + c.FullPath(),
			StartTime: time.Now(),
			Tags: map[string]interface{}{
				"http.method":      c.Request.Method,
				"http.url":         c.Request.URL.String(),
				"http.path":        c.Request.URL.Path,
				"http.user_agent":  c.Request.UserAgent(),
				"http.remote_addr": c.ClientIP(),
			},
		}

		c.Header("X-Trace-ID", traceID)
		c.Header("X-Span-ID", spanID)

		c.Set(TraceIDKey, traceID)
		c.Set(SpanIDKey, spanID)
		c.Set(ParentIDKey, parentID)
		c.Set(ServiceKey, serviceName)
		c.Set(OperationKey, span.Operation)

		ctx := context.WithValue(c.Request.Context(), TraceIDKey, traceID)
		ctx = context.WithValue(ctx, SpanIDKey, spanID)
		ctx = context.WithValue(ctx, "trace_span", span)
		c.Request = c.Request.WithContext(ctx)

		if userID := GetUserID(c); userID != 0 {
			span.Tags["user.id"] = userID
			ctx = context.WithValue(ctx, "user_id", userID)
			c.Request = c.Request.WithContext(ctx)
		}

		c.Next()

		span.EndTime = time.Now()
		span.Duration = span.EndTime.Sub(span.StartTime)
		span.Tags["http.status_code"] = c.Writer.Status()
		span.Tags["http.response_size"] = c.Writer.Size()

		if c.Writer.Status() >= 400 {
			span.Error = true
			if len(c.Errors) > 0 {
				span.ErrorMsg = c.Errors.String()
			}
		}

		structuredLogger.WithTraceID(traceID).LogBusinessEvent(ctx, logger.BusinessEventLog{
			Event:    "request_trace",
			Entity:   "http_request",
			EntityID: spanID,
			Success:  !span.Error,
			Duration: span.Duration,
			Details: map[string]interface{}{
				"span":      span,
				"trace_id":  traceID,
				"span_id":   spanID,
				"parent_id": parentID,
			},
		})
	})
}

func TraceDBOperation(ctx context.Context, operation, table, query string, structuredLogger logger.StructuredLogger) func(error, time.Duration, int64) {
	startTime := time.Now()
	traceID := getTraceIDFromContext(ctx)
	spanID := generateSpanID()
	parentSpanID := getSpanIDFromContext(ctx)

	span := &TraceSpan{
		TraceID:   traceID,
		SpanID:    spanID,
		ParentID:  parentSpanID,
		Service:   "database",
		Operation: operation,
		StartTime: startTime,
		Tags: map[string]interface{}{
			"db.type":      "postgresql",
			"db.operation": operation,
			"db.table":     table,
			"db.query":     query,
		},
	}

	return func(err error, duration time.Duration, rowsAffected int64) {
		span.EndTime = time.Now()
		span.Duration = duration
		span.Tags["db.rows_affected"] = rowsAffected

		if err != nil {
			span.Error = true
			span.ErrorMsg = err.Error()
		}

		structuredLogger.WithTraceID(traceID).LogDatabaseQuery(ctx, logger.DBQueryLog{
			Query:        query,
			Duration:     duration,
			RowsAffected: rowsAffected,
			Error:        span.ErrorMsg,
			Table:        table,
			Operation:    operation,
		})
	}
}

func TraceExternalCall(ctx context.Context, service, operation string, structuredLogger logger.StructuredLogger) func(int, error, time.Duration) {
	startTime := time.Now()
	traceID := getTraceIDFromContext(ctx)
	spanID := generateSpanID()
	parentSpanID := getSpanIDFromContext(ctx)

	span := &TraceSpan{
		TraceID:   traceID,
		SpanID:    spanID,
		ParentID:  parentSpanID,
		Service:   service,
		Operation: operation,
		StartTime: startTime,
		Tags: map[string]interface{}{
			"external.service":   service,
			"external.operation": operation,
		},
	}

	return func(statusCode int, err error, duration time.Duration) {
		span.EndTime = time.Now()
		span.Duration = duration
		span.Tags["external.status_code"] = statusCode

		if err != nil || statusCode >= 400 {
			span.Error = true
			if err != nil {
				span.ErrorMsg = err.Error()
			}
		}

		structuredLogger.WithTraceID(traceID).LogIntegrationEvent(ctx, logger.IntegrationEventLog{
			Service:    service,
			Operation:  operation,
			Success:    !span.Error,
			Duration:   duration,
			StatusCode: statusCode,
			Error:      span.ErrorMsg,
			RequestID:  traceID,
		})
	}
}

func PerformanceMiddleware(structuredLogger logger.StructuredLogger) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		fields := map[string]interface{}{
			"method":        method,
			"path":          path,
			"status":        status,
			"latency":       latency.String(),
			"latency_ms":    latency.Milliseconds(),
			"request_size":  c.Request.ContentLength,
			"response_size": c.Writer.Size(),
		}

		if traceID := getTraceIDFromContext(c.Request.Context()); traceID != "" {
			fields["trace_id"] = traceID
		}

		if latency > 5*time.Second {
			structuredLogger.WithFields(fields).Warn("Slow request detected")
		} else if latency > 1*time.Second {
			structuredLogger.WithFields(fields).Info("Request completed")
		} else {
			structuredLogger.WithFields(fields).Debug("Request completed")
		}

		if status >= 500 {
			structuredLogger.WithFields(fields).Error("Server error response")
		} else if status >= 400 {
			structuredLogger.WithFields(fields).Warn("Client error response")
		}
	})
}

func CorrelationMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {

		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = generateCorrelationID()
		}

		c.Header("X-Correlation-ID", correlationID)

		c.Set("correlation_id", correlationID)
		ctx := context.WithValue(c.Request.Context(), "correlation_id", correlationID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	})
}

func generateTraceID() string {
	return uuid.New().String()
}

func generateSpanID() string {
	return uuid.New().String()[:16]
}

func generateCorrelationID() string {
	return uuid.New().String()
}

func getTraceIDFromContext(ctx context.Context) string {
	if traceID := ctx.Value(TraceIDKey); traceID != nil {
		return traceID.(string)
	}
	return ""
}

func getSpanIDFromContext(ctx context.Context) string {
	if spanID := ctx.Value(SpanIDKey); spanID != nil {
		return spanID.(string)
	}
	return ""
}
