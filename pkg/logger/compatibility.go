package logger

import "context"

type LoggerAdapter struct {
	basicLogger Logger
}

func NewLoggerAdapter(logger Logger) StructuredLogger {
	if structuredLogger, ok := logger.(StructuredLogger); ok {
		return structuredLogger
	}
	return &LoggerAdapter{basicLogger: logger}
}

func (l *LoggerAdapter) Debug(msg string, fields ...interface{}) {
	l.basicLogger.Debug(msg, fields...)
}

func (l *LoggerAdapter) Info(msg string, fields ...interface{}) {
	l.basicLogger.Info(msg, fields...)
}

func (l *LoggerAdapter) Warn(msg string, fields ...interface{}) {
	l.basicLogger.Warn(msg, fields...)
}

func (l *LoggerAdapter) Error(msg string, fields ...interface{}) {
	l.basicLogger.Error(msg, fields...)
}

func (l *LoggerAdapter) Fatal(msg string, fields ...interface{}) {
	l.basicLogger.Fatal(msg, fields...)
}

func (l *LoggerAdapter) WithField(key string, value interface{}) Logger {
	return l.basicLogger.WithField(key, value)
}

func (l *LoggerAdapter) WithFields(fields map[string]interface{}) Logger {
	return l.basicLogger.WithFields(fields)
}

func (l *LoggerAdapter) WithError(err error) Logger {
	return l.basicLogger.WithError(err)
}

func (l *LoggerAdapter) LogHTTPRequest(ctx context.Context, req HTTPRequestLog) {
	l.basicLogger.Info("HTTP Request",
		"method", req.Method,
		"path", req.Path,
		"status", req.StatusCode,
		"latency", req.Latency.String(),
		"ip", req.IP,
		"user_agent", req.UserAgent)
}

func (l *LoggerAdapter) LogDatabaseQuery(ctx context.Context, query DBQueryLog) {
	if query.Error != "" {
		l.basicLogger.Error("Database Query Failed",
			"query", query.Query,
			"duration", query.Duration.String(),
			"error", query.Error)
	} else {
		l.basicLogger.Debug("Database Query",
			"operation", query.Operation,
			"table", query.Table,
			"duration", query.Duration.String(),
			"rows_affected", query.RowsAffected)
	}
}

func (l *LoggerAdapter) LogUserAction(ctx context.Context, action UserActionLog) {
	if action.Success {
		l.basicLogger.Info("User Action",
			"user_id", action.UserID,
			"action", action.Action,
			"resource", action.Resource,
			"ip", action.IP)
	} else {
		l.basicLogger.Warn("User Action Failed",
			"user_id", action.UserID,
			"action", action.Action,
			"resource", action.Resource,
			"error", action.ErrorReason,
			"ip", action.IP)
	}
}

func (l *LoggerAdapter) LogSecurityEvent(ctx context.Context, event SecurityEventLog) {
	l.basicLogger.Error("Security Event",
		"event_type", event.EventType,
		"description", event.Description,
		"severity", event.Severity,
		"ip", event.IP,
		"blocked", event.Blocked)
}

func (l *LoggerAdapter) LogAuthEvent(ctx context.Context, event AuthEventLog) {
	if event.Success {
		l.basicLogger.Info("Auth Event",
			"action", event.Action,
			"user_id", event.UserID,
			"ip", event.IP)
	} else {
		l.basicLogger.Warn("Auth Event Failed",
			"action", event.Action,
			"fail_reason", event.FailReason,
			"ip", event.IP)
	}
}

func (l *LoggerAdapter) LogFileUpload(ctx context.Context, upload FileUploadLog) {
	if upload.Success {
		l.basicLogger.Info("File Upload",
			"user_id", upload.UserID,
			"file_name", upload.FileName,
			"file_size", upload.FileSize,
			"duration", upload.Duration.String())
	} else {
		l.basicLogger.Error("File Upload Failed",
			"user_id", upload.UserID,
			"file_name", upload.FileName,
			"error", upload.Error)
	}
}

func (l *LoggerAdapter) LogEmailEvent(ctx context.Context, email EmailEventLog) {
	if email.Success {
		l.basicLogger.Info("Email Sent",
			"to", email.To,
			"type", email.Type,
			"duration", email.Duration.String())
	} else {
		l.basicLogger.Error("Email Failed",
			"to", email.To,
			"type", email.Type,
			"error", email.Error)
	}
}

func (l *LoggerAdapter) LogValidationError(ctx context.Context, validation ValidationErrorLog) {
	l.basicLogger.Warn("Validation Error",
		"field", validation.Field,
		"rule", validation.Rule,
		"message", validation.Message,
		"resource", validation.Resource)
}

func (l *LoggerAdapter) LogBusinessEvent(ctx context.Context, event BusinessEventLog) {
	if event.Success {
		l.basicLogger.Info("Business Event",
			"event", event.Event,
			"entity", event.Entity,
			"entity_id", event.EntityID,
			"user_id", event.UserID)
	} else {
		l.basicLogger.Error("Business Event Failed",
			"event", event.Event,
			"entity", event.Entity,
			"error", event.Error)
	}
}

func (l *LoggerAdapter) LogIntegrationEvent(ctx context.Context, event IntegrationEventLog) {
	if event.Success {
		l.basicLogger.Info("Integration Event",
			"service", event.Service,
			"operation", event.Operation,
			"duration", event.Duration.String())
	} else {
		l.basicLogger.Error("Integration Event Failed",
			"service", event.Service,
			"operation", event.Operation,
			"error", event.Error)
	}
}

func (l *LoggerAdapter) WithContext(ctx context.Context) StructuredLogger {
	return l
}

func (l *LoggerAdapter) WithTraceID(traceID string) StructuredLogger {
	newLogger := l.basicLogger.WithField("trace_id", traceID)
	return &LoggerAdapter{basicLogger: newLogger}
}

func (l *LoggerAdapter) WithUserID(userID uint) StructuredLogger {
	newLogger := l.basicLogger.WithField("user_id", userID)
	return &LoggerAdapter{basicLogger: newLogger}
}

func (l *LoggerAdapter) WithRequestID(requestID string) StructuredLogger {
	newLogger := l.basicLogger.WithField("request_id", requestID)
	return &LoggerAdapter{basicLogger: newLogger}
}
