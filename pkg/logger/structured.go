package logger

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	EventTypeHTTPRequest   = "http_request"
	EventTypeDBQuery       = "database_query"
	EventTypeUserAction    = "user_action"
	EventTypeSecurity      = "security_event"
	EventTypeError         = "error"
	EventTypeAuth          = "auth_event"
	EventTypeFileUpload    = "file_upload"
	EventTypeEmailSent     = "email_sent"
	EventTypeRateLimit     = "rate_limit"
	EventTypeValidation    = "validation_error"
	EventTypeBusinessLogic = "business_logic"
	EventTypeIntegration   = "external_integration"
)

type StructuredLogger interface {
	Logger

	LogHTTPRequest(ctx context.Context, req HTTPRequestLog)
	LogDatabaseQuery(ctx context.Context, query DBQueryLog)
	LogUserAction(ctx context.Context, action UserActionLog)
	LogSecurityEvent(ctx context.Context, event SecurityEventLog)
	LogAuthEvent(ctx context.Context, event AuthEventLog)
	LogFileUpload(ctx context.Context, upload FileUploadLog)
	LogEmailEvent(ctx context.Context, email EmailEventLog)
	LogValidationError(ctx context.Context, validation ValidationErrorLog)
	LogBusinessEvent(ctx context.Context, event BusinessEventLog)
	LogIntegrationEvent(ctx context.Context, event IntegrationEventLog)

	WithContext(ctx context.Context) StructuredLogger
	WithTraceID(traceID string) StructuredLogger
	WithUserID(userID uint) StructuredLogger
	WithRequestID(requestID string) StructuredLogger
}

type HTTPRequestLog struct {
	Method       string        `json:"method"`
	Path         string        `json:"path"`
	Query        string        `json:"query,omitempty"`
	StatusCode   int           `json:"status_code"`
	Latency      time.Duration `json:"latency"`
	UserAgent    string        `json:"user_agent"`
	IP           string        `json:"ip"`
	UserID       uint          `json:"user_id,omitempty"`
	RequestSize  int64         `json:"request_size,omitempty"`
	ResponseSize int64         `json:"response_size,omitempty"`
	Referer      string        `json:"referer,omitempty"`
}

type DBQueryLog struct {
	Query        string        `json:"query"`
	Args         []interface{} `json:"args,omitempty"`
	Duration     time.Duration `json:"duration"`
	RowsAffected int64         `json:"rows_affected"`
	Error        string        `json:"error,omitempty"`
	Table        string        `json:"table,omitempty"`
	Operation    string        `json:"operation"`
}

type UserActionLog struct {
	UserID      uint                   `json:"user_id"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resource_id,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	IP          string                 `json:"ip,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Success     bool                   `json:"success"`
	ErrorReason string                 `json:"error_reason,omitempty"`
}

type SecurityEventLog struct {
	EventType   string                 `json:"event_type"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"`
	IP          string                 `json:"ip"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	UserID      uint                   `json:"user_id,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Blocked     bool                   `json:"blocked"`
}

type AuthEventLog struct {
	UserID     uint   `json:"user_id,omitempty"`
	Email      string `json:"email,omitempty"`
	Action     string `json:"action"`
	Success    bool   `json:"success"`
	IP         string `json:"ip"`
	UserAgent  string `json:"user_agent,omitempty"`
	FailReason string `json:"fail_reason,omitempty"`
	TokenType  string `json:"token_type,omitempty"`
}

type FileUploadLog struct {
	UserID      uint          `json:"user_id"`
	FileName    string        `json:"file_name"`
	FileSize    int64         `json:"file_size"`
	ContentType string        `json:"content_type"`
	Destination string        `json:"destination"`
	Success     bool          `json:"success"`
	Error       string        `json:"error,omitempty"`
	Duration    time.Duration `json:"duration"`
}

type EmailEventLog struct {
	To       string        `json:"to"`
	Subject  string        `json:"subject"`
	Type     string        `json:"type"`
	Success  bool          `json:"success"`
	Error    string        `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
	Provider string        `json:"provider,omitempty"`
}

type ValidationErrorLog struct {
	Field    string `json:"field"`
	Value    string `json:"value,omitempty"`
	Rule     string `json:"rule"`
	Message  string `json:"message"`
	Resource string `json:"resource"`
	UserID   uint   `json:"user_id,omitempty"`
}

type BusinessEventLog struct {
	Event    string                 `json:"event"`
	Entity   string                 `json:"entity"`
	EntityID string                 `json:"entity_id,omitempty"`
	UserID   uint                   `json:"user_id,omitempty"`
	Success  bool                   `json:"success"`
	Details  map[string]interface{} `json:"details,omitempty"`
	Duration time.Duration          `json:"duration,omitempty"`
	Error    string                 `json:"error,omitempty"`
}

type IntegrationEventLog struct {
	Service    string                 `json:"service"`
	Operation  string                 `json:"operation"`
	Success    bool                   `json:"success"`
	Duration   time.Duration          `json:"duration"`
	StatusCode int                    `json:"status_code,omitempty"`
	Error      string                 `json:"error,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

type structuredLogger struct {
	logger    *logrus.Logger
	entry     *logrus.Entry
	context   map[string]interface{}
	traceID   string
	requestID string
	userID    uint
}

var _ StructuredLogger = (*structuredLogger)(nil)

func NewStructuredLogger() StructuredLogger {
	logger := logrus.New()

	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
			logrus.FieldKeyFunc:  "caller",
		},
	})

	return &structuredLogger{
		logger:  logger,
		entry:   logrus.NewEntry(logger),
		context: make(map[string]interface{}),
	}
}

func (s *structuredLogger) WithContext(ctx context.Context) StructuredLogger {
	newLogger := *s

	if traceID := ctx.Value("trace_id"); traceID != nil {
		newLogger.traceID = traceID.(string)
	}
	if requestID := ctx.Value("request_id"); requestID != nil {
		newLogger.requestID = requestID.(string)
	}
	if userID := ctx.Value("user_id"); userID != nil {
		newLogger.userID = userID.(uint)
	}

	return &newLogger
}

func (s *structuredLogger) WithTraceID(traceID string) StructuredLogger {
	newLogger := *s
	newLogger.traceID = traceID
	return &newLogger
}

func (s *structuredLogger) WithUserID(userID uint) StructuredLogger {
	newLogger := *s
	newLogger.userID = userID
	return &newLogger
}

func (s *structuredLogger) WithRequestID(requestID string) StructuredLogger {
	newLogger := *s
	newLogger.requestID = requestID
	return &newLogger
}

func (s *structuredLogger) Debug(msg string, fields ...interface{}) {
	s.log(logrus.DebugLevel, msg, fields...)
}

func (s *structuredLogger) Info(msg string, fields ...interface{}) {
	s.log(logrus.InfoLevel, msg, fields...)
}

func (s *structuredLogger) Warn(msg string, fields ...interface{}) {
	s.log(logrus.WarnLevel, msg, fields...)
}

func (s *structuredLogger) Error(msg string, fields ...interface{}) {
	s.log(logrus.ErrorLevel, msg, fields...)
}

func (s *structuredLogger) Fatal(msg string, fields ...interface{}) {
	s.log(logrus.FatalLevel, msg, fields...)
}

func (s *structuredLogger) WithField(key string, value interface{}) Logger {
	return &structuredLogger{
		logger: s.logger,
		entry:  s.entry.WithField(key, value),
	}
}

func (s *structuredLogger) WithFields(fields map[string]interface{}) Logger {
	return &structuredLogger{
		logger: s.logger,
		entry:  s.entry.WithFields(logrus.Fields(fields)),
	}
}

func (s *structuredLogger) WithError(err error) Logger {
	return &structuredLogger{
		logger: s.logger,
		entry:  s.entry.WithError(err),
	}
}

func (s *structuredLogger) LogHTTPRequest(ctx context.Context, req HTTPRequestLog) {
	fields := s.getBaseFields(EventTypeHTTPRequest)
	fields["method"] = req.Method
	fields["path"] = req.Path
	fields["query"] = req.Query
	fields["status_code"] = req.StatusCode
	fields["latency"] = req.Latency.String()
	fields["latency_ms"] = req.Latency.Milliseconds()
	fields["user_agent"] = req.UserAgent
	fields["ip"] = req.IP
	fields["request_size"] = req.RequestSize
	fields["response_size"] = req.ResponseSize
	fields["referer"] = req.Referer

	if req.UserID > 0 {
		fields["user_id"] = req.UserID
	}

	level := logrus.InfoLevel
	if req.StatusCode >= 500 {
		level = logrus.ErrorLevel
	} else if req.StatusCode >= 400 {
		level = logrus.WarnLevel
	}

	s.logger.WithFields(fields).Log(level, "HTTP request processed")
}

func (s *structuredLogger) LogDatabaseQuery(ctx context.Context, query DBQueryLog) {
	fields := s.getBaseFields(EventTypeDBQuery)
	fields["query"] = query.Query
	fields["duration"] = query.Duration.String()
	fields["duration_ms"] = query.Duration.Milliseconds()
	fields["rows_affected"] = query.RowsAffected
	fields["table"] = query.Table
	fields["operation"] = query.Operation

	if query.Error != "" {
		fields["error"] = query.Error
		s.logger.WithFields(fields).Error("Database query failed")
	} else {

		if query.Duration > time.Second {
			s.logger.WithFields(fields).Warn("Slow database query")
		} else {
			s.logger.WithFields(fields).Debug("Database query executed")
		}
	}
}

func (s *structuredLogger) LogUserAction(ctx context.Context, action UserActionLog) {
	fields := s.getBaseFields(EventTypeUserAction)
	fields["user_id"] = action.UserID
	fields["action"] = action.Action
	fields["resource"] = action.Resource
	fields["resource_id"] = action.ResourceID
	fields["success"] = action.Success
	fields["ip"] = action.IP
	fields["user_agent"] = action.UserAgent

	for k, v := range action.Details {
		fields[k] = v
	}

	if action.Success {
		s.logger.WithFields(fields).Info("User action performed")
	} else {
		fields["error_reason"] = action.ErrorReason
		s.logger.WithFields(fields).Warn("User action failed")
	}
}

func (s *structuredLogger) LogSecurityEvent(ctx context.Context, event SecurityEventLog) {
	fields := s.getBaseFields(EventTypeSecurity)
	fields["security_event_type"] = event.EventType
	fields["description"] = event.Description
	fields["severity"] = event.Severity
	fields["ip"] = event.IP
	fields["user_agent"] = event.UserAgent
	fields["blocked"] = event.Blocked

	if event.UserID > 0 {
		fields["user_id"] = event.UserID
	}

	for k, v := range event.Details {
		fields[k] = v
	}

	level := logrus.InfoLevel
	switch event.Severity {
	case "critical", "high":
		level = logrus.ErrorLevel
	case "medium":
		level = logrus.WarnLevel
	}

	s.logger.WithFields(fields).Log(level, "Security event detected")
}

func (s *structuredLogger) LogAuthEvent(ctx context.Context, event AuthEventLog) {
	fields := s.getBaseFields(EventTypeAuth)
	fields["action"] = event.Action
	fields["success"] = event.Success
	fields["ip"] = event.IP
	fields["user_agent"] = event.UserAgent
	fields["token_type"] = event.TokenType

	if event.UserID > 0 {
		fields["user_id"] = event.UserID
	}
	if event.Email != "" {

		fields["email"] = maskEmail(event.Email)
	}

	if event.Success {
		s.logger.WithFields(fields).Info("Authentication event")
	} else {
		fields["fail_reason"] = event.FailReason
		s.logger.WithFields(fields).Warn("Authentication failed")
	}
}

func (s *structuredLogger) LogFileUpload(ctx context.Context, upload FileUploadLog) {
	fields := s.getBaseFields(EventTypeFileUpload)
	fields["user_id"] = upload.UserID
	fields["file_name"] = upload.FileName
	fields["file_size"] = upload.FileSize
	fields["content_type"] = upload.ContentType
	fields["destination"] = upload.Destination
	fields["success"] = upload.Success
	fields["duration"] = upload.Duration.String()

	if upload.Success {
		s.logger.WithFields(fields).Info("File uploaded successfully")
	} else {
		fields["error"] = upload.Error
		s.logger.WithFields(fields).Error("File upload failed")
	}
}

func (s *structuredLogger) LogEmailEvent(ctx context.Context, email EmailEventLog) {
	fields := s.getBaseFields(EventTypeEmailSent)
	fields["to"] = maskEmail(email.To)
	fields["subject"] = email.Subject
	fields["type"] = email.Type
	fields["success"] = email.Success
	fields["duration"] = email.Duration.String()
	fields["provider"] = email.Provider

	if email.Success {
		s.logger.WithFields(fields).Info("Email sent successfully")
	} else {
		fields["error"] = email.Error
		s.logger.WithFields(fields).Error("Email sending failed")
	}
}

func (s *structuredLogger) LogValidationError(ctx context.Context, validation ValidationErrorLog) {
	fields := s.getBaseFields(EventTypeValidation)
	fields["field"] = validation.Field
	fields["value"] = validation.Value
	fields["rule"] = validation.Rule
	fields["message"] = validation.Message
	fields["resource"] = validation.Resource

	if validation.UserID > 0 {
		fields["user_id"] = validation.UserID
	}

	s.logger.WithFields(fields).Warn("Validation error")
}

func (s *structuredLogger) LogBusinessEvent(ctx context.Context, event BusinessEventLog) {
	fields := s.getBaseFields(EventTypeBusinessLogic)
	fields["event"] = event.Event
	fields["entity"] = event.Entity
	fields["entity_id"] = event.EntityID
	fields["success"] = event.Success

	if event.UserID > 0 {
		fields["user_id"] = event.UserID
	}
	if event.Duration > 0 {
		fields["duration"] = event.Duration.String()
	}

	for k, v := range event.Details {
		fields[k] = v
	}

	if event.Success {
		s.logger.WithFields(fields).Info("Business event processed")
	} else {
		fields["error"] = event.Error
		s.logger.WithFields(fields).Error("Business event failed")
	}
}

func (s *structuredLogger) LogIntegrationEvent(ctx context.Context, event IntegrationEventLog) {
	fields := s.getBaseFields(EventTypeIntegration)
	fields["service"] = event.Service
	fields["operation"] = event.Operation
	fields["success"] = event.Success
	fields["duration"] = event.Duration.String()
	fields["status_code"] = event.StatusCode
	fields["request_id"] = event.RequestID

	for k, v := range event.Details {
		fields[k] = v
	}

	if event.Success {
		s.logger.WithFields(fields).Info("External integration successful")
	} else {
		fields["error"] = event.Error
		s.logger.WithFields(fields).Error("External integration failed")
	}
}

func (s *structuredLogger) getBaseFields(eventType string) logrus.Fields {
	fields := logrus.Fields{
		"event_type": eventType,
		"service":    "linkedin-clone",
		"version":    "1.0.0",
	}

	if s.traceID != "" {
		fields["trace_id"] = s.traceID
	}
	if s.requestID != "" {
		fields["request_id"] = s.requestID
	}
	if s.userID > 0 {
		fields["user_id"] = s.userID
	}

	if pc, file, line, ok := runtime.Caller(3); ok {
		funcName := runtime.FuncForPC(pc).Name()
		if idx := strings.LastIndex(funcName, "/"); idx >= 0 {
			funcName = funcName[idx+1:]
		}
		fields["caller"] = fmt.Sprintf("%s:%d:%s", file, line, funcName)
	}

	return fields
}

func (s *structuredLogger) log(level logrus.Level, msg string, fields ...interface{}) {
	logFields := s.getBaseFields("")

	if len(fields)%2 == 0 {
		for i := 0; i < len(fields); i += 2 {
			if key, ok := fields[i].(string); ok {
				logFields[key] = fields[i+1]
			}
		}
	}

	s.logger.WithFields(logFields).Log(level, msg)
}

func maskEmail(email string) string {
	if email == "" {
		return ""
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***"
	}

	username := parts[0]
	domain := parts[1]

	if len(username) <= 2 {
		return "***@" + domain
	}

	masked := string(username[0]) + strings.Repeat("*", len(username)-2) + string(username[len(username)-1])
	return masked + "@" + domain
}
