package handler

import (
	"linked-clone/internal/api/post/dto"
	"linked-clone/internal/api/post/service"
	"linked-clone/internal/middleware"
	"linked-clone/pkg/logger"
	"linked-clone/pkg/response"
	validation "linked-clone/pkg/validator"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	postService service.PostService
	validator   validation.Validator
	logger      logger.Logger
}

func NewPostHandler(postService service.PostService, validator validation.Validator, logger logger.Logger) *PostHandler {
	return &PostHandler{
		postService: postService,
		validator:   validator,
		logger:      logger,
	}
}

func (h *PostHandler) CreatePost(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.CreatePostRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationErrors(c, err)
		return
	}

	file, _ := c.FormFile("image")

	post, err := h.postService.CreatePost(c.Request.Context(), userID, &req, file)
	if err != nil {
		h.logger.Error("Failed to create post", "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to create post", err.Error())
		return
	}

	response.Success(c, post)
}

func (h *PostHandler) GetPost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post ID", err.Error())
		return
	}

	post, err := h.postService.GetPost(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get post", "error", err)
		response.Error(c, http.StatusNotFound, "Post not found", err.Error())
		return
	}

	response.Success(c, post)
}

func (h *PostHandler) GetFeed(c *gin.Context) {
	userID := middleware.GetUserID(c)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	posts, err := h.postService.GetFeed(c.Request.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get feed", "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to get feed", err.Error())
		return
	}

	response.Success(c, gin.H{
		"posts":  posts,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *PostHandler) GetUserPosts(c *gin.Context) {
	idStr := c.Param("user_id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	posts, err := h.postService.GetUserPosts(c.Request.Context(), uint(userID), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get user posts", "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to get posts", err.Error())
		return
	}

	response.Success(c, gin.H{
		"posts":   posts,
		"user_id": userID,
		"limit":   limit,
		"offset":  offset,
	})
}

func (h *PostHandler) UpdatePost(c *gin.Context) {
	userID := middleware.GetUserID(c)

	idStr := c.Param("id")
	postID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post ID", err.Error())
		return
	}

	var req dto.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationErrors(c, err)
		return
	}

	post, err := h.postService.UpdatePost(c.Request.Context(), userID, uint(postID), &req)
	if err != nil {
		h.logger.Error("Failed to update post", "error", err)
		response.Error(c, http.StatusBadRequest, "Failed to update post", err.Error())
		return
	}

	response.Success(c, post)
}

func (h *PostHandler) DeletePost(c *gin.Context) {
	userID := middleware.GetUserID(c)

	idStr := c.Param("id")
	postID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post ID", err.Error())
		return
	}

	if err := h.postService.DeletePost(c.Request.Context(), userID, uint(postID)); err != nil {
		h.logger.Error("Failed to delete post", "error", err)
		response.Error(c, http.StatusBadRequest, "Failed to delete post", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Post deleted successfully"})
}

func (h *PostHandler) LikePost(c *gin.Context) {
	userID := middleware.GetUserID(c)

	idStr := c.Param("id")
	postID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post ID", err.Error())
		return
	}

	like, err := h.postService.LikePost(c.Request.Context(), userID, uint(postID))
	if err != nil {
		h.logger.Error("Failed to like post", "error", err)
		response.Error(c, http.StatusBadRequest, "Failed to like post", err.Error())
		return
	}

	response.Success(c, like)
}

func (h *PostHandler) UnlikePost(c *gin.Context) {
	userID := middleware.GetUserID(c)

	idStr := c.Param("id")
	postID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post ID", err.Error())
		return
	}

	if err := h.postService.UnlikePost(c.Request.Context(), userID, uint(postID)); err != nil {
		h.logger.Error("Failed to unlike post", "error", err)
		response.Error(c, http.StatusBadRequest, "Failed to unlike post", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Post unliked successfully"})
}

func (h *PostHandler) AddComment(c *gin.Context) {
	userID := middleware.GetUserID(c)

	idStr := c.Param("id")
	postID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post ID", err.Error())
		return
	}

	var req dto.AddCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationErrors(c, err)
		return
	}

	comment, err := h.postService.AddComment(c.Request.Context(), userID, uint(postID), &req)
	if err != nil {
		h.logger.Error("Failed to add comment", "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to add comment", err.Error())
		return
	}

	response.Success(c, comment)
}

func (h *PostHandler) GetComments(c *gin.Context) {
	idStr := c.Param("id")
	postID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post ID", err.Error())
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	comments, err := h.postService.GetComments(c.Request.Context(), uint(postID), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get comments", "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to get comments", err.Error())
		return
	}

	response.Success(c, gin.H{
		"comments": comments,
		"post_id":  postID,
		"limit":    limit,
		"offset":   offset,
	})
}
