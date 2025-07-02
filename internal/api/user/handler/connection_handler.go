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

type ConnectionHandler struct {
	connectionService service.ConnectionService
	validator         validation.Validator
	logger            logger.Logger
}

func NewConnectionHandler(connectionService service.ConnectionService, validator validation.Validator, logger logger.Logger) *ConnectionHandler {
	return &ConnectionHandler{
		connectionService: connectionService,
		validator:         validator,
		logger:            logger,
	}
}

func (h *ConnectionHandler) SendConnectionRequest(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.ConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationErrors(c, err)
		return
	}

	connection, err := h.connectionService.SendConnectionRequest(c.Request.Context(), userID, req.UserID)
	if err != nil {
		h.logger.Error("Failed to send connection request", "error", err)
		response.Error(c, http.StatusBadRequest, "Failed to send connection request", err.Error())
		return
	}

	response.Success(c, connection)
}

func (h *ConnectionHandler) AcceptConnectionRequest(c *gin.Context) {
	userID := middleware.GetUserID(c)

	idStr := c.Param("id")
	connectionID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid connection ID", err.Error())
		return
	}

	connection, err := h.connectionService.AcceptConnectionRequest(c.Request.Context(), userID, uint(connectionID))
	if err != nil {
		h.logger.Error("Failed to accept connection request", "error", err)
		response.Error(c, http.StatusBadRequest, "Failed to accept connection request", err.Error())
		return
	}

	response.Success(c, connection)
}

func (h *ConnectionHandler) RejectConnectionRequest(c *gin.Context) {
	userID := middleware.GetUserID(c)

	idStr := c.Param("id")
	connectionID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid connection ID", err.Error())
		return
	}

	if err := h.connectionService.RejectConnectionRequest(c.Request.Context(), userID, uint(connectionID)); err != nil {
		h.logger.Error("Failed to reject connection request", "error", err)
		response.Error(c, http.StatusBadRequest, "Failed to reject connection request", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Connection request rejected successfully"})
}

func (h *ConnectionHandler) RemoveConnection(c *gin.Context) {
	userID := middleware.GetUserID(c)

	idStr := c.Param("id")
	connectionID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid connection ID", err.Error())
		return
	}

	if err := h.connectionService.RemoveConnection(c.Request.Context(), userID, uint(connectionID)); err != nil {
		h.logger.Error("Failed to remove connection", "error", err)
		response.Error(c, http.StatusBadRequest, "Failed to remove connection", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Connection removed successfully"})
}

func (h *ConnectionHandler) GetUserConnections(c *gin.Context) {
	userID := middleware.GetUserID(c)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	connections, err := h.connectionService.GetUserConnections(c.Request.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get user connections", "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to get connections", err.Error())
		return
	}

	response.Success(c, gin.H{
		"connections": connections,
		"limit":       limit,
		"offset":      offset,
	})
}

func (h *ConnectionHandler) GetConnectionRequests(c *gin.Context) {
	userID := middleware.GetUserID(c)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	connections, err := h.connectionService.GetConnectionRequests(c.Request.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get connection requests", "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to get connection requests", err.Error())
		return
	}

	response.Success(c, gin.H{
		"requests": connections,
		"limit":    limit,
		"offset":   offset,
	})
}

func (h *ConnectionHandler) GetSentRequests(c *gin.Context) {
	userID := middleware.GetUserID(c)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	connections, err := h.connectionService.GetSentRequests(c.Request.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get sent requests", "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to get sent requests", err.Error())
		return
	}

	response.Success(c, gin.H{
		"sent_requests": connections,
		"limit":         limit,
		"offset":        offset,
	})
}

func (h *ConnectionHandler) GetConnectionStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)

	userIDStr := c.Param("userId")
	targetUserID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	connection, err := h.connectionService.GetConnectionStatus(c.Request.Context(), userID, uint(targetUserID))
	if err != nil {
		h.logger.Error("Failed to get connection status", "error", err)
		response.Error(c, http.StatusNotFound, "No connection found", err.Error())
		return
	}

	response.Success(c, connection)
}

func (h *ConnectionHandler) GetMutualConnections(c *gin.Context) {
	userID := middleware.GetUserID(c)

	userIDStr := c.Param("userId")
	targetUserID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	connections, err := h.connectionService.GetMutualConnections(c.Request.Context(), userID, uint(targetUserID), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get mutual connections", "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to get mutual connections", err.Error())
		return
	}

	response.Success(c, gin.H{
		"mutual_connections": connections,
		"limit":              limit,
		"offset":             offset,
	})
}

func (h *ConnectionHandler) BlockUser(c *gin.Context) {
	userID := middleware.GetUserID(c)

	userIDStr := c.Param("userId")
	targetUserID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	if err := h.connectionService.BlockUser(c.Request.Context(), userID, uint(targetUserID)); err != nil {
		h.logger.Error("Failed to block user", "error", err)
		response.Error(c, http.StatusBadRequest, "Failed to block user", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "User blocked successfully"})
}

func (h *ConnectionHandler) UnblockUser(c *gin.Context) {
	userID := middleware.GetUserID(c)

	userIDStr := c.Param("userId")
	targetUserID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	if err := h.connectionService.UnblockUser(c.Request.Context(), userID, uint(targetUserID)); err != nil {
		h.logger.Error("Failed to unblock user", "error", err)
		response.Error(c, http.StatusBadRequest, "Failed to unblock user", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "User unblocked successfully"})
}
