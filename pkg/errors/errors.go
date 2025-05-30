package errors

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

const (
	ErrCodeValidation     = "VALIDATION_ERROR"
	ErrCodeAuthentication = "AUTHENTICATION_ERROR"
	ErrCodeAuthorization  = "AUTHORIZATION_ERROR"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeConflict       = "CONFLICT"
	ErrCodeRateLimit      = "RATE_LIMIT_EXCEEDED"
	ErrCodeInternal       = "INTERNAL_ERROR"
	ErrCodeExternal       = "EXTERNAL_SERVICE_ERROR"
	ErrCodeDatabase       = "DATABASE_ERROR"
	ErrCodeNetwork        = "NETWORK_ERROR"
	ErrCodeTimeout        = "TIMEOUT_ERROR"
	ErrCodeFileUpload     = "FILE_UPLOAD_ERROR"
	ErrCodeEmailService   = "EMAIL_SERVICE_ERROR"
	ErrCodeCacheService   = "CACHE_SERVICE_ERROR"
)

const (
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

type AppError struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Details     string                 `json:"details,omitempty"`
	Cause       error                  `json:"-"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Severity    string                 `json:"severity"`
	Retryable   bool                   `json:"retryable"`
	UserMessage string                 `json:"user_message,omitempty"`
	HTTPStatus  int                    `json:"http_status"`
	Component   string                 `json:"component,omitempty"`
	Operation   string                 `json:"operation,omitempty"`
}

func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s - %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func (e *AppError) Is(target error) bool {
	if targetErr, ok := target.(*AppError); ok {
		return e.Code == targetErr.Code
	}
	return false
}

func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

func (e *AppError) WithComponent(component string) *AppError {
	e.Component = component
	return e
}

func (e *AppError) WithOperation(operation string) *AppError {
	e.Operation = operation
	return e
}

func New(code, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Timestamp:  time.Now().UTC(),
		StackTrace: getStackTrace(),
		Severity:   SeverityMedium,
		HTTPStatus: getHTTPStatusFromCode(code),
	}
}

func Newf(code, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:       code,
		Message:    fmt.Sprintf(format, args...),
		Timestamp:  time.Now().UTC(),
		StackTrace: getStackTrace(),
		Severity:   SeverityMedium,
		HTTPStatus: getHTTPStatusFromCode(code),
	}
}

func Wrap(err error, code, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Cause:      err,
		Details:    err.Error(),
		Timestamp:  time.Now().UTC(),
		StackTrace: getStackTrace(),
		Severity:   SeverityMedium,
		HTTPStatus: getHTTPStatusFromCode(code),
	}
}

func Wrapf(err error, code, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:       code,
		Message:    fmt.Sprintf(format, args...),
		Cause:      err,
		Details:    err.Error(),
		Timestamp:  time.Now().UTC(),
		StackTrace: getStackTrace(),
		Severity:   SeverityMedium,
		HTTPStatus: getHTTPStatusFromCode(code),
	}
}

func ValidationError(message string) *AppError {
	return &AppError{
		Code:        ErrCodeValidation,
		Message:     message,
		Timestamp:   time.Now().UTC(),
		Severity:    SeverityLow,
		HTTPStatus:  400,
		UserMessage: "Please check your input and try again.",
		Retryable:   false,
	}
}

func AuthenticationError(message string) *AppError {
	return &AppError{
		Code:        ErrCodeAuthentication,
		Message:     message,
		Timestamp:   time.Now().UTC(),
		Severity:    SeverityMedium,
		HTTPStatus:  401,
		UserMessage: "Authentication required. Please log in.",
		Retryable:   false,
	}
}

func AuthorizationError(message string) *AppError {
	return &AppError{
		Code:        ErrCodeAuthorization,
		Message:     message,
		Timestamp:   time.Now().UTC(),
		Severity:    SeverityMedium,
		HTTPStatus:  403,
		UserMessage: "You don't have permission to perform this action.",
		Retryable:   false,
	}
}

func NotFoundError(resource string) *AppError {
	return &AppError{
		Code:        ErrCodeNotFound,
		Message:     fmt.Sprintf("%s not found", resource),
		Timestamp:   time.Now().UTC(),
		Severity:    SeverityLow,
		HTTPStatus:  404,
		UserMessage: "The requested resource was not found.",
		Retryable:   false,
	}
}

func ConflictError(message string) *AppError {
	return &AppError{
		Code:        ErrCodeConflict,
		Message:     message,
		Timestamp:   time.Now().UTC(),
		Severity:    SeverityMedium,
		HTTPStatus:  409,
		UserMessage: "The operation conflicts with existing data.",
		Retryable:   false,
	}
}

func RateLimitError(message string) *AppError {
	return &AppError{
		Code:        ErrCodeRateLimit,
		Message:     message,
		Timestamp:   time.Now().UTC(),
		Severity:    SeverityMedium,
		HTTPStatus:  429,
		UserMessage: "Too many requests. Please try again later.",
		Retryable:   true,
	}
}

func InternalError(message string) *AppError {
	return &AppError{
		Code:        ErrCodeInternal,
		Message:     message,
		Timestamp:   time.Now().UTC(),
		Severity:    SeverityHigh,
		HTTPStatus:  500,
		UserMessage: "An internal error occurred. Please try again later.",
		Retryable:   true,
	}
}

func DatabaseError(err error, operation string) *AppError {
	return &AppError{
		Code:        ErrCodeDatabase,
		Message:     fmt.Sprintf("Database error during %s", operation),
		Cause:       err,
		Details:     err.Error(),
		Timestamp:   time.Now().UTC(),
		Severity:    SeverityHigh,
		HTTPStatus:  500,
		UserMessage: "A database error occurred. Please try again later.",
		Retryable:   true,
		Operation:   operation,
	}
}

func ExternalServiceError(service string, err error) *AppError {
	return &AppError{
		Code:        ErrCodeExternal,
		Message:     fmt.Sprintf("External service error: %s", service),
		Cause:       err,
		Details:     err.Error(),
		Timestamp:   time.Now().UTC(),
		Severity:    SeverityHigh,
		HTTPStatus:  502,
		UserMessage: "An external service is currently unavailable. Please try again later.",
		Retryable:   true,
		Component:   service,
	}
}

func TimeoutError(operation string) *AppError {
	return &AppError{
		Code:        ErrCodeTimeout,
		Message:     fmt.Sprintf("Operation timeout: %s", operation),
		Timestamp:   time.Now().UTC(),
		Severity:    SeverityMedium,
		HTTPStatus:  408,
		UserMessage: "The operation took too long. Please try again.",
		Retryable:   true,
		Operation:   operation,
	}
}

func FileUploadError(message string) *AppError {
	return &AppError{
		Code:        ErrCodeFileUpload,
		Message:     message,
		Timestamp:   time.Now().UTC(),
		Severity:    SeverityMedium,
		HTTPStatus:  400,
		UserMessage: "File upload failed. Please check the file and try again.",
		Retryable:   true,
	}
}

func getStackTrace() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])

	var trace []string
	frames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := frames.Next()
		if strings.Contains(frame.File, "runtime/") {
			if !more {
				break
			}
			continue
		}

		trace = append(trace, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))

		if !more {
			break
		}
	}

	return strings.Join(trace, "\n")
}

func getHTTPStatusFromCode(code string) int {
	switch code {
	case ErrCodeValidation:
		return 400
	case ErrCodeAuthentication:
		return 401
	case ErrCodeAuthorization:
		return 403
	case ErrCodeNotFound:
		return 404
	case ErrCodeConflict:
		return 409
	case ErrCodeRateLimit:
		return 429
	case ErrCodeTimeout:
		return 408
	case ErrCodeInternal, ErrCodeDatabase:
		return 500
	case ErrCodeExternal:
		return 502
	default:
		return 500
	}
}

func IsRetryable(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Retryable
	}
	return false
}

func GetSeverity(err error) string {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Severity
	}
	return SeverityMedium
}

func GetHTTPStatus(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.HTTPStatus
	}
	return 500
}

func GetUserMessage(err error) string {
	if appErr, ok := err.(*AppError); ok && appErr.UserMessage != "" {
		return appErr.UserMessage
	}
	return "An error occurred. Please try again later."
}
