package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Meta      *MetaInfo   `json:"meta,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type ErrorInfo struct {
	Code    string                  `json:"code,omitempty"`
	Message string                  `json:"message"`
	Details interface{}             `json:"details,omitempty"`
	Fields  []ValidationErrorDetail `json:"fields,omitempty"`
}

type MetaInfo struct {
	Page       int `json:"page,omitempty"`
	Limit      int `json:"limit,omitempty"`
	Offset     int `json:"offset,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

const (
	ErrCodeValidation   = "VALIDATION_ERROR"
	ErrCodeUnauthorized = "UNAUTHORIZED"
	ErrCodeForbidden    = "FORBIDDEN"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeConflict     = "CONFLICT"
	ErrCodeInternal     = "INTERNAL_ERROR"
	ErrCodeRateLimit    = "RATE_LIMIT_EXCEEDED"
	ErrCodeBadRequest   = "BAD_REQUEST"
	ErrCodeTimeout      = "TIMEOUT"
	ErrCodeServiceError = "SERVICE_ERROR"
)

func Success(c *gin.Context, data interface{}) {
	respond(c, http.StatusOK, true, "", data, nil, nil)
}

func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	respond(c, http.StatusOK, true, message, data, nil, nil)
}

func SuccessWithMeta(c *gin.Context, data interface{}, meta *MetaInfo) {
	respond(c, http.StatusOK, true, "", data, nil, meta)
}

func Created(c *gin.Context, data interface{}) {
	respond(c, http.StatusCreated, true, "Resource created successfully", data, nil, nil)
}

func CreatedWithMessage(c *gin.Context, message string, data interface{}) {
	respond(c, http.StatusCreated, true, message, data, nil, nil)
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Error(c *gin.Context, statusCode int, message, details string) {
	errorInfo := &ErrorInfo{
		Code:    getErrorCodeFromStatus(statusCode),
		Message: message,
		Details: details,
	}
	respond(c, statusCode, false, "", nil, errorInfo, nil)
}

func ErrorWithCode(c *gin.Context, statusCode int, code, message, details string) {
	errorInfo := &ErrorInfo{
		Code:    code,
		Message: message,
		Details: details,
	}
	respond(c, statusCode, false, "", nil, errorInfo, nil)
}

func ValidationErrors(c *gin.Context, err error) {
	var validationErrors []ValidationErrorDetail

	if validatorErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validatorErrors {
			validationErrors = append(validationErrors, ValidationErrorDetail{
				Field:   fieldError.Field(),
				Tag:     fieldError.Tag(),
				Message: getValidationMessage(fieldError),
				Value:   fieldError.Value().(string),
			})
		}
	}

	errorInfo := &ErrorInfo{
		Code:    ErrCodeValidation,
		Message: "Validation failed",
		Fields:  validationErrors,
	}
	respond(c, http.StatusBadRequest, false, "", nil, errorInfo, nil)
}

func BadRequest(c *gin.Context, message, details string) {
	ErrorWithCode(c, http.StatusBadRequest, ErrCodeBadRequest, message, details)
}

func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized access"
	}
	ErrorWithCode(c, http.StatusUnauthorized, ErrCodeUnauthorized, message, "")
}

func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "Access forbidden"
	}
	ErrorWithCode(c, http.StatusForbidden, ErrCodeForbidden, message, "")
}

func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "Resource not found"
	}
	ErrorWithCode(c, http.StatusNotFound, ErrCodeNotFound, message, "")
}

func Conflict(c *gin.Context, message, details string) {
	if message == "" {
		message = "Resource conflict"
	}
	ErrorWithCode(c, http.StatusConflict, ErrCodeConflict, message, details)
}

func InternalServerError(c *gin.Context, message, details string) {
	if message == "" {
		message = "Internal server error"
	}
	ErrorWithCode(c, http.StatusInternalServerError, ErrCodeInternal, message, details)
}

func TooManyRequests(c *gin.Context, message string) {
	if message == "" {
		message = "Too many requests"
	}
	ErrorWithCode(c, http.StatusTooManyRequests, ErrCodeRateLimit, message, "")
}

func RequestTimeout(c *gin.Context, message string) {
	if message == "" {
		message = "Request timeout"
	}
	ErrorWithCode(c, http.StatusRequestTimeout, ErrCodeTimeout, message, "")
}

func UnprocessableEntity(c *gin.Context, message, details string) {
	if message == "" {
		message = "Unprocessable entity"
	}
	ErrorWithCode(c, http.StatusUnprocessableEntity, ErrCodeValidation, message, details)
}

func ServiceError(c *gin.Context, service, message string) {
	ErrorWithCode(c, http.StatusServiceUnavailable, ErrCodeServiceError,
		"Service temporarily unavailable", "External service error: "+service+" - "+message)
}

func respond(c *gin.Context, statusCode int, success bool, message string, data interface{}, errorInfo *ErrorInfo, meta *MetaInfo) {
	response := APIResponse{
		Success:   success,
		Message:   message,
		Data:      data,
		Error:     errorInfo,
		Meta:      meta,
		RequestID: getRequestID(c),
		Timestamp: time.Now().UTC(),
	}

	c.JSON(statusCode, response)
}

func getRequestID(c *gin.Context) string {
	if requestID := c.GetString("request_id"); requestID != "" {
		return requestID
	}
	return c.GetHeader("X-Request-ID")
}

func getErrorCodeFromStatus(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return ErrCodeBadRequest
	case http.StatusUnauthorized:
		return ErrCodeUnauthorized
	case http.StatusForbidden:
		return ErrCodeForbidden
	case http.StatusNotFound:
		return ErrCodeNotFound
	case http.StatusConflict:
		return ErrCodeConflict
	case http.StatusTooManyRequests:
		return ErrCodeRateLimit
	case http.StatusRequestTimeout:
		return ErrCodeTimeout
	case http.StatusInternalServerError:
		return ErrCodeInternal
	default:
		return "UNKNOWN_ERROR"
	}
}

func getValidationMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return fieldError.Field() + " is required"
	case "email":
		return fieldError.Field() + " must be a valid email address"
	case "min":
		return fieldError.Field() + " must be at least " + fieldError.Param() + " characters long"
	case "max":
		return fieldError.Field() + " must be at most " + fieldError.Param() + " characters long"
	case "len":
		return fieldError.Field() + " must be exactly " + fieldError.Param() + " characters long"
	case "alphanum":
		return fieldError.Field() + " must contain only alphanumeric characters"
	case "url":
		return fieldError.Field() + " must be a valid URL"
	case "oneof":
		return fieldError.Field() + " must be one of: " + fieldError.Param()
	case "gte":
		return fieldError.Field() + " must be greater than or equal to " + fieldError.Param()
	case "lte":
		return fieldError.Field() + " must be less than or equal to " + fieldError.Param()
	case "gt":
		return fieldError.Field() + " must be greater than " + fieldError.Param()
	case "lt":
		return fieldError.Field() + " must be less than " + fieldError.Param()
	case "eqfield":
		return fieldError.Field() + " must equal " + fieldError.Param()
	case "nefield":
		return fieldError.Field() + " must not equal " + fieldError.Param()
	case "unique":
		return fieldError.Field() + " must be unique"
	case "numeric":
		return fieldError.Field() + " must be a valid number"
	case "alpha":
		return fieldError.Field() + " must contain only alphabetic characters"
	default:
		return fieldError.Field() + " is invalid"
	}
}

func CreateMeta(page, limit, total int) *MetaInfo {
	totalPages := (total + limit - 1) / limit
	offset := (page - 1) * limit

	return &MetaInfo{
		Page:       page,
		Limit:      limit,
		Offset:     offset,
		Total:      total,
		TotalPages: totalPages,
	}
}
