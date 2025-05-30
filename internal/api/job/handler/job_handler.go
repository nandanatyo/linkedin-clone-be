package handler

import (
	"linked-clone/internal/api/job/dto"
	"linked-clone/internal/api/job/service"
	"linked-clone/internal/middleware"
	"linked-clone/pkg/logger"
	"linked-clone/pkg/response"
	validation "linked-clone/pkg/validator"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type JobHandler struct {
	jobService service.JobService
	validator  validation.Validator
	logger     logger.Logger
}

func NewJobHandler(jobService service.JobService, validator validation.Validator, logger logger.Logger) *JobHandler {
	return &JobHandler{
		jobService: jobService,
		validator:  validator,
		logger:     logger,
	}
}

func (h *JobHandler) CreateJob(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationErrors(c, err)
		return
	}

	job, err := h.jobService.CreateJob(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Failed to create job", "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to create job", err.Error())
		return
	}

	response.Success(c, job)
}

func (h *JobHandler) GetJob(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid job ID", err.Error())
		return
	}

	job, err := h.jobService.GetJob(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get job", "error", err)
		response.Error(c, http.StatusNotFound, "Job not found", err.Error())
		return
	}

	response.Success(c, job)
}

func (h *JobHandler) GetJobs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filters := make(map[string]interface{})
	if jobType := c.Query("job_type"); jobType != "" {
		filters["job_type"] = jobType
	}
	if expLevel := c.Query("experience_level"); expLevel != "" {
		filters["experience_level"] = expLevel
	}
	if location := c.Query("location"); location != "" {
		filters["location"] = location
	}

	jobs, err := h.jobService.GetAllJobs(c.Request.Context(), filters, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get jobs", "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to get jobs", err.Error())
		return
	}

	response.Success(c, gin.H{
		"jobs":    jobs,
		"limit":   limit,
		"offset":  offset,
		"filters": filters,
	})
}

func (h *JobHandler) SearchJobs(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.Error(c, http.StatusBadRequest, "Search query is required", "")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filters := make(map[string]interface{})
	if jobType := c.Query("job_type"); jobType != "" {
		filters["job_type"] = jobType
	}
	if expLevel := c.Query("experience_level"); expLevel != "" {
		filters["experience_level"] = expLevel
	}
	if location := c.Query("location"); location != "" {
		filters["location"] = location
	}

	jobs, err := h.jobService.SearchJobs(c.Request.Context(), query, filters, limit, offset)
	if err != nil {
		h.logger.Error("Failed to search jobs", "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to search jobs", err.Error())
		return
	}

	response.Success(c, gin.H{
		"jobs":    jobs,
		"query":   query,
		"limit":   limit,
		"offset":  offset,
		"filters": filters,
	})
}

func (h *JobHandler) UpdateJob(c *gin.Context) {
	userID := middleware.GetUserID(c)

	idStr := c.Param("id")
	jobID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid job ID", err.Error())
		return
	}

	var req dto.UpdateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationErrors(c, err)
		return
	}

	job, err := h.jobService.UpdateJob(c.Request.Context(), userID, uint(jobID), &req)
	if err != nil {
		h.logger.Error("Failed to update job", "error", err)
		response.Error(c, http.StatusBadRequest, "Failed to update job", err.Error())
		return
	}

	response.Success(c, job)
}

func (h *JobHandler) DeleteJob(c *gin.Context) {
	userID := middleware.GetUserID(c)

	idStr := c.Param("id")
	jobID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid job ID", err.Error())
		return
	}

	if err := h.jobService.DeleteJob(c.Request.Context(), userID, uint(jobID)); err != nil {
		h.logger.Error("Failed to delete job", "error", err)
		response.Error(c, http.StatusBadRequest, "Failed to delete job", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Job deleted successfully"})
}

func (h *JobHandler) ApplyJob(c *gin.Context) {
	userID := middleware.GetUserID(c)

	idStr := c.Param("id")
	jobID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid job ID", err.Error())
		return
	}

	var req dto.ApplyJobRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationErrors(c, err)
		return
	}

	resume, _ := c.FormFile("resume")

	application, err := h.jobService.ApplyJob(c.Request.Context(), userID, uint(jobID), &req, resume)
	if err != nil {
		h.logger.Error("Failed to apply for job", "error", err)
		response.Error(c, http.StatusBadRequest, "Failed to apply for job", err.Error())
		return
	}

	response.Success(c, application)
}

func (h *JobHandler) GetMyApplications(c *gin.Context) {
	userID := middleware.GetUserID(c)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	applications, err := h.jobService.GetUserApplications(c.Request.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get applications", "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to get applications", err.Error())
		return
	}

	response.Success(c, gin.H{
		"applications": applications,
		"limit":        limit,
		"offset":       offset,
	})
}

func (h *JobHandler) GetJobApplications(c *gin.Context) {
	userID := middleware.GetUserID(c)

	idStr := c.Param("id")
	jobID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid job ID", err.Error())
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	applications, err := h.jobService.GetJobApplications(c.Request.Context(), userID, uint(jobID), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get job applications", "error", err)
		response.Error(c, http.StatusBadRequest, "Failed to get applications", err.Error())
		return
	}

	response.Success(c, gin.H{
		"applications": applications,
		"job_id":       jobID,
		"limit":        limit,
		"offset":       offset,
	})
}

func (h *JobHandler) GetMyJobs(c *gin.Context) {
	userID := middleware.GetUserID(c)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	jobs, err := h.jobService.GetUserJobs(c.Request.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get user jobs", "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to get jobs", err.Error())
		return
	}

	response.Success(c, gin.H{
		"jobs":   jobs,
		"limit":  limit,
		"offset": offset,
	})
}
