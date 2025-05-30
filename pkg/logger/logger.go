package logger

import (
	"github.com/sirupsen/logrus"
	"os"
)

type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})

	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
}

type logrusLogger struct {
	logger *logrus.Logger
	entry  *logrus.Entry
}

func NewLogger() Logger {
	logger := logrus.New()

	env := os.Getenv("ENVIRONMENT")
	logLevel := os.Getenv("LOG_LEVEL")

	if logLevel != "" {
		if level, err := logrus.ParseLevel(logLevel); err == nil {
			logger.SetLevel(level)
		} else {
			logger.SetLevel(logrus.InfoLevel)
		}
	} else if env == "production" {
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetLevel(logrus.DebugLevel)
	}

	if env == "production" || os.Getenv("LOG_FORMAT") == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
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

	logger.SetOutput(os.Stdout)

	return &logrusLogger{
		logger: logger,
		entry:  logrus.NewEntry(logger),
	}
}

func NewLoggerWithConfig(logger *logrus.Logger) Logger {
	return &logrusLogger{
		logger: logger,
		entry:  logrus.NewEntry(logger),
	}
}

func (l *logrusLogger) Debug(msg string, fields ...interface{}) {
	l.entry.WithFields(convertFields(fields...)).Debug(msg)
}

func (l *logrusLogger) Info(msg string, fields ...interface{}) {
	l.entry.WithFields(convertFields(fields...)).Info(msg)
}

func (l *logrusLogger) Warn(msg string, fields ...interface{}) {
	l.entry.WithFields(convertFields(fields...)).Warn(msg)
}

func (l *logrusLogger) Error(msg string, fields ...interface{}) {
	l.entry.WithFields(convertFields(fields...)).Error(msg)
}

func (l *logrusLogger) Fatal(msg string, fields ...interface{}) {
	l.entry.WithFields(convertFields(fields...)).Fatal(msg)
}

func (l *logrusLogger) WithField(key string, value interface{}) Logger {
	return &logrusLogger{
		logger: l.logger,
		entry:  l.entry.WithField(key, value),
	}
}

func (l *logrusLogger) WithFields(fields map[string]interface{}) Logger {
	logrusFields := make(logrus.Fields)
	for k, v := range fields {
		logrusFields[k] = v
	}

	return &logrusLogger{
		logger: l.logger,
		entry:  l.entry.WithFields(logrusFields),
	}
}

func (l *logrusLogger) WithError(err error) Logger {
	return &logrusLogger{
		logger: l.logger,
		entry:  l.entry.WithError(err),
	}
}

func convertFields(fields ...interface{}) logrus.Fields {
	logrusFields := logrus.Fields{}

	if len(fields)%2 == 0 {
		for i := 0; i < len(fields); i += 2 {
			if key, ok := fields[i].(string); ok {
				logrusFields[key] = fields[i+1]
			}
		}
	}

	if len(fields) == 1 {
		if fieldMap, ok := fields[0].(map[string]interface{}); ok {
			for k, v := range fieldMap {
				logrusFields[k] = v
			}
		}
	}

	return logrusFields
}
