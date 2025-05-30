package routes

import (
	"github.com/gin-gonic/gin"
	"linked-clone/internal/middleware"
	"time"
)

func JobRoutes(rg *gin.RouterGroup, deps *Dependencies) {
	authMiddleware := middleware.AuthMiddleware(deps.JWTService, deps.Logger)

	jobs := rg.Group("/jobs")
	{

		jobs.GET("",
			middleware.RateLimitMiddleware(time.Minute, 100, deps.Logger),
			deps.JobHandler.GetJobs)

		jobs.GET("/search",
			middleware.RateLimitMiddleware(time.Minute, 50, deps.Logger),
			deps.JobHandler.SearchJobs)

		jobs.GET("/:id",
			middleware.RateLimitMiddleware(time.Minute, 200, deps.Logger),
			deps.JobHandler.GetJob)

		jobs.POST("",
			authMiddleware,
			middleware.RateLimitMiddleware(time.Minute, 10, deps.Logger),
			deps.JobHandler.CreateJob)

		jobs.PUT("/:id",
			authMiddleware,
			middleware.RateLimitMiddleware(time.Minute, 20, deps.Logger),
			deps.JobHandler.UpdateJob)

		jobs.DELETE("/:id",
			authMiddleware,
			middleware.RateLimitMiddleware(time.Minute, 10, deps.Logger),
			deps.JobHandler.DeleteJob)

		jobs.GET("/:id/applications",
			authMiddleware,
			middleware.RateLimitMiddleware(time.Minute, 50, deps.Logger),
			deps.JobHandler.GetJobApplications)

		jobs.GET("/my/jobs",
			authMiddleware,
			middleware.RateLimitMiddleware(time.Minute, 50, deps.Logger),
			deps.JobHandler.GetMyJobs)

		jobs.GET("/my/applications",
			authMiddleware,
			middleware.RateLimitMiddleware(time.Minute, 50, deps.Logger),
			deps.JobHandler.GetMyApplications)

		jobs.POST("/:id/apply",
			authMiddleware,
			middleware.RateLimitMiddleware(time.Minute, 5, deps.Logger),
			middleware.FileUploadMiddleware(5<<20, []string{".pdf", ".doc", ".docx"}),
			middleware.SecurityMonitoring(deps.Logger),
			deps.JobHandler.ApplyJob,
		)
	}
}
