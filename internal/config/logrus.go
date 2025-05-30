package config

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type LogrusConfig struct {
	Level        string `json:"level"`
	Format       string `json:"format"`
	Output       string `json:"output"`
	ReportCaller bool   `json:"report_caller"`
}

func NewLogrusLogger(cfg *Config) *logrus.Logger {
	logger := logrus.New()

	level := strings.ToLower(getEnv("LOG_LEVEL", "info"))
	switch level {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn", "warning":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logger.SetLevel(logrus.FatalLevel)
	case "panic":
		logger.SetLevel(logrus.PanicLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	logFormat := getEnv("LOG_FORMAT", "")
	if cfg.Server.Environment == "production" || logFormat == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "caller",
			},
		})
	} else {

		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:          true,
			TimestampFormat:        "2006-01-02 15:04:05",
			ForceColors:            true,
			DisableLevelTruncation: true,
			PadLevelText:           true,
		})
	}

	logOutput := getEnv("LOG_OUTPUT", "stdout")
	switch logOutput {
	case "stderr":
		logger.SetOutput(os.Stderr)
	case "stdout", "":
		logger.SetOutput(os.Stdout)
	default:

		if file, err := os.OpenFile(logOutput, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
			logger.SetOutput(file)
		} else {
			logger.SetOutput(os.Stdout)
			logger.Warnf("Failed to open log file %s, falling back to stdout: %v", logOutput, err)
		}
	}

	reportCaller := getEnv("LOG_REPORT_CALLER", "false")
	if reportCaller == "true" || cfg.Server.Environment == "development" {
		logger.SetReportCaller(true)
	}

	if cfg.Server.Environment == "production" {
		logger.AddHook(NewCustomHook(logrus.AllLevels))
	} else {
		logger.AddHook(NewCustomHook([]logrus.Level{logrus.DebugLevel, logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel}))
	}

	return logger
}

type CustomHook struct {
	levels []logrus.Level
}

func NewCustomHook(levels []logrus.Level) *CustomHook {
	return &CustomHook{
		levels: levels,
	}
}

func (hook *CustomHook) Levels() []logrus.Level {
	return hook.levels
}

func (hook *CustomHook) Fire(entry *logrus.Entry) error {

	entry.Data["service"] = "linkedin-clone"
	entry.Data["version"] = "1.0.0"

	if _, exists := entry.Data["password"]; exists {
		entry.Data["password"] = "[REDACTED]"
	}
	if _, exists := entry.Data["token"]; exists {
		entry.Data["token"] = "[REDACTED]"
	}
	if _, exists := entry.Data["secret"]; exists {
		entry.Data["secret"] = "[REDACTED]"
	}

	return nil
}

type ContextualLogger struct {
	*logrus.Logger
	context logrus.Fields
}

func NewContextualLogger(cfg *Config) *ContextualLogger {
	logger := NewLogrusLogger(cfg)

	hook := NewCustomHook(logrus.AllLevels)
	logger.AddHook(hook)

	return &ContextualLogger{
		Logger: logger,
		context: logrus.Fields{
			"service":     "linkedin-clone",
			"version":     "1.0.0",
			"environment": cfg.Server.Environment,
		},
	}
}

func (cl *ContextualLogger) WithField(key string, value interface{}) *logrus.Entry {
	return cl.Logger.WithFields(cl.context).WithField(key, value)
}

func (cl *ContextualLogger) WithFields(fields logrus.Fields) *logrus.Entry {

	mergedFields := make(logrus.Fields)
	for k, v := range cl.context {
		mergedFields[k] = v
	}
	for k, v := range fields {
		mergedFields[k] = v
	}
	return cl.Logger.WithFields(mergedFields)
}

func (cl *ContextualLogger) WithError(err error) *logrus.Entry {
	return cl.Logger.WithFields(cl.context).WithError(err)
}

type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
	PanicLevel LogLevel = "panic"
)

func SetLogLevel(logger *logrus.Logger, level LogLevel) {
	switch level {
	case DebugLevel:
		logger.SetLevel(logrus.DebugLevel)
	case InfoLevel:
		logger.SetLevel(logrus.InfoLevel)
	case WarnLevel:
		logger.SetLevel(logrus.WarnLevel)
	case ErrorLevel:
		logger.SetLevel(logrus.ErrorLevel)
	case FatalLevel:
		logger.SetLevel(logrus.FatalLevel)
	case PanicLevel:
		logger.SetLevel(logrus.PanicLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}
}

type StructuredLogger struct {
	*ContextualLogger
}

func NewStructuredLogger(cfg *Config) *StructuredLogger {
	return &StructuredLogger{
		ContextualLogger: NewContextualLogger(cfg),
	}
}

func (sl *StructuredLogger) LogHTTPRequest(method, path, userAgent, ip string, statusCode int, duration string) {
	sl.WithFields(logrus.Fields{
		"http_method":   method,
		"http_path":     path,
		"http_status":   statusCode,
		"user_agent":    userAgent,
		"client_ip":     ip,
		"response_time": duration,
		"event_type":    "http_request",
	}).Info("HTTP request processed")
}

func (sl *StructuredLogger) LogDatabaseQuery(query string, duration string, rowsAffected int64) {
	sl.WithFields(logrus.Fields{
		"db_query":         query,
		"db_duration":      duration,
		"db_rows_affected": rowsAffected,
		"event_type":       "database_query",
	}).Debug("Database query executed")
}

func (sl *StructuredLogger) LogUserAction(userID uint, action, resource string, metadata map[string]interface{}) {
	fields := logrus.Fields{
		"user_id":    userID,
		"action":     action,
		"resource":   resource,
		"event_type": "user_action",
	}

	for k, v := range metadata {
		fields[k] = v
	}

	sl.WithFields(fields).Info("User action logged")
}

func (sl *StructuredLogger) LogSecurityEvent(eventType, description string, severity string, metadata map[string]interface{}) {
	fields := logrus.Fields{
		"security_event_type": eventType,
		"description":         description,
		"severity":            severity,
		"event_type":          "security",
	}

	for k, v := range metadata {
		fields[k] = v
	}

	entry := sl.WithFields(fields)

	switch severity {
	case "critical", "high":
		entry.Error("Security event detected")
	case "medium":
		entry.Warn("Security event detected")
	default:
		entry.Info("Security event detected")
	}
}
