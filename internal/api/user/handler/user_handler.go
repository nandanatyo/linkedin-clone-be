package handler

import (
	"linked-clone/internal/api/user/dto"
	"linked-clone/internal/api/user/service"
	"linked-clone/internal/middleware"
	"linked-clone/pkg/logger"
	"linked-clone/pkg/response"
	validation "linked-clone/pkg/validator"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService service.UserService
	validator   validation.Validator
	logger      logger.Logger
}

func NewUserHandler(userService service.UserService, validator validation.Validator, logger logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   validator,
		logger:      logger,
	}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	profile, err := h.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get profile", "error", err)
		response.Error(c, http.StatusNotFound, "Profile not found", err.Error())
		return
	}

	response.Success(c, profile)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationErrors(c, err)
		return
	}

	profile, err := h.userService.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Failed to update profile", "error", err)
		response.Error(c, http.StatusBadRequest, "Update failed", err.Error())
		return
	}

	response.Success(c, profile)
}

func (h *UserHandler) UploadProfilePicture(c *gin.Context) {
	userID := middleware.GetUserID(c)

	file, err := c.FormFile("image")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "No image file provided", err.Error())
		return
	}

	result, err := h.userService.UploadProfilePicture(c.Request.Context(), userID, file)
	if err != nil {
		h.logger.Error("Failed to upload profile picture", "error", err)
		response.Error(c, http.StatusInternalServerError, "Upload failed", err.Error())
		return
	}

	response.Success(c, result)
}

func (h *UserHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.Error(c, http.StatusBadRequest, "Search query is required", "")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	users, err := h.userService.SearchUsers(c.Request.Context(), query, limit, offset)
	if err != nil {
		h.logger.Error("Failed to search users", "error", err)
		response.Error(c, http.StatusInternalServerError, "Search failed", err.Error())
		return
	}

	response.Success(c, gin.H{
		"users":  users,
		"query":  query,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get user", "error", err)
		response.Error(c, http.StatusNotFound, "User not found", err.Error())
		return
	}

	response.Success(c, user)
}
