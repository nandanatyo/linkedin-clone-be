package handler

import (
	"context"
	"linked-clone/internal/api/auth/dto"
	"linked-clone/internal/api/auth/service"
	"linked-clone/internal/middleware"
	"linked-clone/pkg/errors"
	"linked-clone/pkg/logger"
	"linked-clone/pkg/response"
	validation "linked-clone/pkg/validator"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
	validator   validation.Validator
	logger      logger.StructuredLogger
}

func NewAuthHandler(authService service.AuthService, validator validation.Validator, logger logger.StructuredLogger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator,
		logger:      logger,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	ctx := c.Request.Context()
	traceID := middleware.GetTraceID(c)

	h.logger.WithTraceID(traceID).LogUserAction(ctx, logger.UserActionLog{
		Action:    "register_attempt",
		Resource:  "user",
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Success:   false,
	})

	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := errors.ValidationError("Invalid request body").
			WithContext("raw_error", err.Error()).
			WithComponent("auth_handler").
			WithOperation("register")

		h.logger.WithTraceID(traceID).LogValidationError(ctx, logger.ValidationErrorLog{
			Field:    "request_body",
			Rule:     "json_binding",
			Message:  err.Error(),
			Resource: "register",
		})

		response.BadRequest(c, appErr.Message, appErr.Details)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		h.logger.WithTraceID(traceID).LogValidationError(ctx, logger.ValidationErrorLog{
			Field:    "validation",
			Rule:     "struct_validation",
			Message:  err.Error(),
			Resource: "register",
		})

		response.ValidationErrors(c, err)
		return
	}

	h.logger.WithTraceID(traceID).Info("Processing registration request",
		"email", req.Email,
		"username", req.Username,
		"full_name", req.FullName)

	// Add gin context to the context for extractRequestInfo method
	ctxWithGin := context.WithValue(ctx, "gin_context", c)

	result, err := h.authService.Register(ctxWithGin, &req)
	if err != nil {
		switch {
		case err.Error() == "email already registered":
			appErr := errors.ConflictError("Email already registered").
				WithContext("email", req.Email).
				WithComponent("auth_service").
				WithOperation("register")

			h.logger.WithTraceID(traceID).LogUserAction(ctx, logger.UserActionLog{
				Action:      "register_failed",
				Resource:    "user",
				IP:          c.ClientIP(),
				UserAgent:   c.Request.UserAgent(),
				Success:     false,
				ErrorReason: "email_already_exists",
				Details: map[string]interface{}{
					"email": req.Email,
				},
			})

			response.Conflict(c, appErr.Message, appErr.Details)
			return

		case err.Error() == "username already taken":
			appErr := errors.ConflictError("Username already taken").
				WithContext("username", req.Username).
				WithComponent("auth_service").
				WithOperation("register")

			h.logger.WithTraceID(traceID).LogUserAction(ctx, logger.UserActionLog{
				Action:      "register_failed",
				Resource:    "user",
				IP:          c.ClientIP(),
				UserAgent:   c.Request.UserAgent(),
				Success:     false,
				ErrorReason: "username_already_exists",
				Details: map[string]interface{}{
					"username": req.Username,
				},
			})

			response.Conflict(c, appErr.Message, appErr.Details)
			return

		default:
			appErr := errors.InternalError("Registration failed").
				WithContext("original_error", err.Error()).
				WithComponent("auth_service").
				WithOperation("register")

			h.logger.WithTraceID(traceID).Error("Registration failed",
				"error", err.Error(),
				"email", req.Email,
				"username", req.Username)

			response.InternalServerError(c, appErr.Message, appErr.UserMessage)
			return
		}
	}

	h.logger.WithTraceID(traceID).LogUserAction(ctx, logger.UserActionLog{
		UserID:    result.User.ID,
		Action:    "register_success",
		Resource:  "user",
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Success:   true,
		Details: map[string]interface{}{
			"user_id":  result.User.ID,
			"email":    result.User.Email,
			"username": result.User.Username,
		},
	})

	response.Created(c, result)
}

func (h *AuthHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()
	traceID := middleware.GetTraceID(c)

	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := errors.ValidationError("Invalid request body").
			WithContext("raw_error", err.Error()).
			WithComponent("auth_handler").
			WithOperation("login")

		response.BadRequest(c, appErr.Message, appErr.Details)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationErrors(c, err)
		return
	}

	h.logger.WithTraceID(traceID).LogAuthEvent(ctx, logger.AuthEventLog{
		Email:     req.Email,
		Action:    "login_attempt",
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Success:   false,
	})

	// Add gin context to the context for extractRequestInfo method
	ctxWithGin := context.WithValue(ctx, "gin_context", c)

	result, err := h.authService.Login(ctxWithGin, &req)
	if err != nil {
		h.logger.WithTraceID(traceID).LogAuthEvent(ctx, logger.AuthEventLog{
			Email:      req.Email,
			Action:     "login_failed",
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			Success:    false,
			FailReason: err.Error(),
		})

		h.logger.WithTraceID(traceID).LogSecurityEvent(ctx, logger.SecurityEventLog{
			EventType:   "failed_login",
			Description: "Failed login attempt",
			Severity:    "medium",
			IP:          c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
			Details: map[string]interface{}{
				"email": req.Email,
				"error": err.Error(),
			},
			Blocked: false,
		})

		switch {
		case err.Error() == "invalid email or password":
			appErr := errors.AuthenticationError("Invalid credentials").
				WithComponent("auth_service").
				WithOperation("login")
			response.Unauthorized(c, appErr.Message)
			return
		default:
			appErr := errors.InternalError("Login failed").
				WithContext("original_error", err.Error()).
				WithComponent("auth_service").
				WithOperation("login")
			response.InternalServerError(c, appErr.Message, appErr.UserMessage)
			return
		}
	}

	h.logger.WithTraceID(traceID).LogAuthEvent(ctx, logger.AuthEventLog{
		UserID:    result.User.ID,
		Email:     result.User.Email,
		Action:    "login_success",
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Success:   true,
		TokenType: "access_token",
	})

	response.Success(c, result)
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	ctx := c.Request.Context()
	traceID := middleware.GetTraceID(c)
	userID := middleware.GetUserID(c)

	var req dto.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := errors.ValidationError("Invalid request body").
			WithContext("raw_error", err.Error()).
			WithComponent("auth_handler").
			WithOperation("verify_email")
		response.BadRequest(c, appErr.Message, appErr.Details)
		return
	}

	req.UserID = userID

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationErrors(c, err)
		return
	}

	h.logger.WithTraceID(traceID).LogUserAction(ctx, logger.UserActionLog{
		UserID:    userID,
		Action:    "email_verification_attempt",
		Resource:  "user",
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Success:   false,
	})

	if err := h.authService.VerifyEmail(ctx, &req); err != nil {
		h.logger.WithTraceID(traceID).LogUserAction(ctx, logger.UserActionLog{
			UserID:      userID,
			Action:      "email_verification_failed",
			Resource:    "user",
			IP:          c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
			Success:     false,
			ErrorReason: err.Error(),
		})

		switch {
		case err.Error() == "verification code expired or invalid":
			appErr := errors.ValidationError("Verification code expired or invalid").
				WithComponent("auth_service").
				WithOperation("verify_email")
			response.BadRequest(c, appErr.Message, appErr.Details)
			return
		case err.Error() == "invalid verification code":
			appErr := errors.ValidationError("Invalid verification code").
				WithComponent("auth_service").
				WithOperation("verify_email")
			response.BadRequest(c, appErr.Message, appErr.Details)
			return
		default:
			appErr := errors.InternalError("Email verification failed").
				WithContext("original_error", err.Error()).
				WithComponent("auth_service").
				WithOperation("verify_email")
			response.InternalServerError(c, appErr.Message, appErr.UserMessage)
			return
		}
	}

	h.logger.WithTraceID(traceID).LogUserAction(ctx, logger.UserActionLog{
		UserID:    userID,
		Action:    "email_verification_success",
		Resource:  "user",
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Success:   true,
	})

	response.Success(c, gin.H{"message": "Email verified successfully"})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	ctx := c.Request.Context()
	traceID := middleware.GetTraceID(c)

	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := errors.ValidationError("Invalid request body").
			WithContext("raw_error", err.Error()).
			WithComponent("auth_handler").
			WithOperation("forgot_password")
		response.BadRequest(c, appErr.Message, appErr.Details)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationErrors(c, err)
		return
	}

	h.logger.WithTraceID(traceID).LogAuthEvent(ctx, logger.AuthEventLog{
		Email:     req.Email,
		Action:    "password_reset_request",
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Success:   true,
	})

	if err := h.authService.ForgotPassword(ctx, &req); err != nil {
		h.logger.WithTraceID(traceID).Error("Password reset failed",
			"error", err.Error(),
			"email", req.Email)
	}

	response.Success(c, gin.H{"message": "Reset code sent to your email"})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	ctx := c.Request.Context()
	traceID := middleware.GetTraceID(c)

	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := errors.ValidationError("Invalid request body").
			WithContext("raw_error", err.Error()).
			WithComponent("auth_handler").
			WithOperation("reset_password")
		response.BadRequest(c, appErr.Message, appErr.Details)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationErrors(c, err)
		return
	}

	h.logger.WithTraceID(traceID).LogAuthEvent(ctx, logger.AuthEventLog{
		Email:     req.Email,
		Action:    "password_reset_attempt",
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Success:   false,
	})

	if err := h.authService.ResetPassword(ctx, &req); err != nil {
		h.logger.WithTraceID(traceID).LogAuthEvent(ctx, logger.AuthEventLog{
			Email:      req.Email,
			Action:     "password_reset_failed",
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			Success:    false,
			FailReason: err.Error(),
		})

		switch {
		case err.Error() == "reset code expired or invalid":
			appErr := errors.ValidationError("Reset code expired or invalid").
				WithComponent("auth_service").
				WithOperation("reset_password")
			response.BadRequest(c, appErr.Message, appErr.Details)
			return
		case err.Error() == "invalid reset code":
			appErr := errors.ValidationError("Invalid reset code").
				WithComponent("auth_service").
				WithOperation("reset_password")
			response.BadRequest(c, appErr.Message, appErr.Details)
			return
		default:
			appErr := errors.InternalError("Password reset failed").
				WithContext("original_error", err.Error()).
				WithComponent("auth_service").
				WithOperation("reset_password")
			response.InternalServerError(c, appErr.Message, appErr.UserMessage)
			return
		}
	}

	h.logger.WithTraceID(traceID).LogAuthEvent(ctx, logger.AuthEventLog{
		Email:     req.Email,
		Action:    "password_reset_success",
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Success:   true,
	})

	response.Success(c, gin.H{"message": "Password reset successfully"})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	ctx := c.Request.Context()
	traceID := middleware.GetTraceID(c)

	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := errors.ValidationError("Invalid request body").
			WithContext("raw_error", err.Error()).
			WithComponent("auth_handler").
			WithOperation("refresh_token")

		h.logger.WithTraceID(traceID).LogValidationError(ctx, logger.ValidationErrorLog{
			Field:    "request_body",
			Rule:     "json_binding",
			Message:  err.Error(),
			Resource: "refresh_token",
		})

		response.BadRequest(c, appErr.Message, appErr.Details)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		h.logger.WithTraceID(traceID).LogValidationError(ctx, logger.ValidationErrorLog{
			Field:    "validation",
			Rule:     "struct_validation",
			Message:  err.Error(),
			Resource: "refresh_token",
		})

		response.ValidationErrors(c, err)
		return
	}

	h.logger.WithTraceID(traceID).LogAuthEvent(ctx, logger.AuthEventLog{
		Action:    "token_refresh_attempt",
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Success:   false,
		TokenType: "refresh_token",
	})

	ctxWithGin := context.WithValue(ctx, "gin_context", c)

	result, err := h.authService.RefreshToken(ctxWithGin, &req)
	if err != nil {
		h.logger.WithTraceID(traceID).LogAuthEvent(ctx, logger.AuthEventLog{
			Action:     "token_refresh_failed",
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			Success:    false,
			FailReason: err.Error(),
			TokenType:  "refresh_token",
		})

		h.logger.WithTraceID(traceID).LogSecurityEvent(ctx, logger.SecurityEventLog{
			EventType:   "invalid_refresh_token",
			Description: "Failed refresh token attempt",
			Severity:    "medium",
			IP:          c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
			Details: map[string]interface{}{
				"error": err.Error(),
			},
			Blocked: false,
		})

		switch {
		case err.Error() == "invalid refresh token":
			appErr := errors.AuthenticationError("Invalid refresh token").
				WithComponent("auth_service").
				WithOperation("refresh_token")
			response.Unauthorized(c, appErr.Message)
			return
		case err.Error() == "user not found":
			appErr := errors.NotFoundError("User").
				WithComponent("auth_service").
				WithOperation("refresh_token")
			response.NotFound(c, appErr.Message)
			return
		default:
			appErr := errors.InternalError("Token refresh failed").
				WithContext("original_error", err.Error()).
				WithComponent("auth_service").
				WithOperation("refresh_token")
			response.InternalServerError(c, appErr.Message, appErr.UserMessage)
			return
		}
	}

	h.logger.WithTraceID(traceID).LogAuthEvent(ctx, logger.AuthEventLog{
		UserID:    result.User.ID,
		Email:     result.User.Email,
		Action:    "token_refresh_success",
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Success:   true,
		TokenType: "access_token",
	})

	h.logger.WithTraceID(traceID).LogBusinessEvent(ctx, logger.BusinessEventLog{
		Event:    "token_refreshed",
		Entity:   "auth_token",
		EntityID: string(rune(result.User.ID)),
		UserID:   result.User.ID,
		Success:  true,
		Details: map[string]interface{}{
			"user_id":  result.User.ID,
			"email":    result.User.Email,
			"username": result.User.Username,
		},
	})

	response.Success(c, result)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	ctx := c.Request.Context()
	traceID := middleware.GetTraceID(c)
	userID := middleware.GetUserID(c)

	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationErrors(c, err)
		return
	}

	h.logger.WithTraceID(traceID).LogAuthEvent(ctx, logger.AuthEventLog{
		UserID:    userID,
		Action:    "logout_attempt",
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Success:   false,
	})

	if err := h.authService.Logout(ctx, req.RefreshToken); err != nil {
		h.logger.WithTraceID(traceID).LogAuthEvent(ctx, logger.AuthEventLog{
			UserID:     userID,
			Action:     "logout_failed",
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			Success:    false,
			FailReason: err.Error(),
		})

		response.InternalServerError(c, "Logout failed", err.Error())
		return
	}

	h.logger.WithTraceID(traceID).LogAuthEvent(ctx, logger.AuthEventLog{
		UserID:    userID,
		Action:    "logout_success",
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Success:   true,
	})

	response.Success(c, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) GetActiveSessions(c *gin.Context) {
	ctx := c.Request.Context()
	userID := middleware.GetUserID(c)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	sessions, err := h.authService.GetUserActiveSessions(ctx, userID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get active sessions", "error", err)
		response.InternalServerError(c, "Failed to get sessions", err.Error())
		return
	}

	response.Success(c, gin.H{
		"sessions": sessions,
		"limit":    limit,
		"offset":   offset,
	})
}

func (h *AuthHandler) RevokeSession(c *gin.Context) {
	ctx := c.Request.Context()
	userID := middleware.GetUserID(c)

	sessionIDStr := c.Param("sessionId")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid session ID", err.Error())
		return
	}

	if err := h.authService.RevokeSession(ctx, userID, uint(sessionID)); err != nil {
		h.logger.Error("Failed to revoke session", "error", err)
		response.InternalServerError(c, "Failed to revoke session", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Session revoked successfully"})
}

func (h *AuthHandler) RevokeAllSessions(c *gin.Context) {
	ctx := c.Request.Context()
	userID := middleware.GetUserID(c)

	if err := h.authService.RevokeAllUserSessions(ctx, userID); err != nil {
		h.logger.Error("Failed to revoke all sessions", "error", err)
		response.InternalServerError(c, "Failed to revoke sessions", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "All sessions revoked successfully"})
}
