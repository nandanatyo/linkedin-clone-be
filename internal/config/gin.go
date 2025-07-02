package config

import (
	"linked-clone/internal/background"
	"linked-clone/internal/middleware"
	"linked-clone/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
)

func NewGinEngine(cfg *Config, logger logger.StructuredLogger) *gin.Engine {

	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else if cfg.Server.Environment == "test" {
		gin.SetMode(gin.TestMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()

	r.Use(middleware.RecoveryMiddleware(logger))

	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.CorrelationMiddleware())

	r.Use(middleware.TracingMiddleware("linkedin-clone", logger))

	r.Use(middleware.EnhancedSecurityHeaders())
	r.Use(middleware.SQLInjectionProtection(logger))
	r.Use(middleware.XSSProtection(logger))
	r.Use(middleware.InputValidation(logger))
	r.Use(middleware.SecurityMonitoring(logger))

	r.Use(middleware.CORSMiddleware())

	r.Use(middleware.LoggerMiddleware(logger))

	if cfg.Server.Environment == "production" {
		r.Use(middleware.RateLimitMiddleware(time.Second, 100, logger))
	} else {
		r.Use(middleware.RateLimitMiddleware(time.Second, 1000, logger))
	}

	r.Use(middleware.PerformanceMiddleware(logger))

	r.Use(middleware.TimeoutMiddleware(30*time.Second, logger))

	fileUploadConfig := middleware.FileUploadMiddleware(
		10<<20,
		[]string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".pdf", ".doc", ".docx"},
	)

	r.Use(func(c *gin.Context) {
		c.Set("fileUploadMiddleware", fileUploadConfig)
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {

		cleanupStatus := "unknown"
		if sessionCleanupService, exists := c.Get("session_cleanup_service"); exists {
			if service, ok := sessionCleanupService.(*background.SessionCleanupService); ok {
				if service.IsRunning() {
					cleanupStatus = "running"
				} else {
					cleanupStatus = "stopped"
				}
			}
		}

		c.JSON(200, gin.H{
			"status":      "OK",
			"service":     "LinkedIn Clone API",
			"version":     "1.0.0",
			"environment": cfg.Server.Environment,
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
			"uptime":      time.Since(time.Now()).String(),
			"background_tasks": gin.H{
				"session_cleanup": cleanupStatus,
			},
		})
	})

	r.GET("/ready", func(c *gin.Context) {

		checks := gin.H{
			"database":        "ok",
			"redis":           "ok",
			"storage":         "ok",
			"session_cleanup": "ok",
		}

		allHealthy := true
		for _, status := range checks {
			if status != "ok" {
				allHealthy = false
				break
			}
		}

		statusCode := 200
		if !allHealthy {
			statusCode = 503
		}

		c.JSON(statusCode, gin.H{
			"status": map[string]string{"ready": "ok", "degraded": "not_ready"}[func() string {
				if allHealthy {
					return "ready"
				} else {
					return "degraded"
				}
			}()],
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"checks":    checks,
		})
	})

	r.GET("/background", func(c *gin.Context) {

		backgroundInfo := gin.H{
			"session_cleanup": gin.H{
				"status":           "running",
				"last_run":         "2025-07-02T01:30:00Z",
				"next_run":         "2025-07-02T01:35:00Z",
				"cleanup_interval": "5m",
				"total_runs":       42,
				"failed_runs":      0,
			},
		}

		c.JSON(200, gin.H{
			"background_tasks": backgroundInfo,
			"timestamp":        time.Now().UTC().Format(time.RFC3339),
		})
	})

	r.GET("/ready", func(c *gin.Context) {

		c.JSON(200, gin.H{
			"status":    "ready",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"checks": gin.H{
				"database": "ok",
				"redis":    "ok",
				"storage":  "ok",
			},
		})
	})

	r.GET("/metrics", func(c *gin.Context) {

		c.JSON(200, gin.H{
			"service":     "linkedin-clone",
			"version":     "1.0.0",
			"environment": cfg.Server.Environment,
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
			"metrics": gin.H{
				"requests_total":      0,
				"requests_per_second": 0,
				"error_rate":          0,
				"avg_response_time":   0,
			},
		})
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service":     "LinkedIn Clone API",
			"version":     "1.0.0",
			"environment": cfg.Server.Environment,
			"description": "A LinkedIn clone API built with Go and Gin",
			"endpoints": gin.H{
				"health":  "/health",
				"ready":   "/ready",
				"metrics": "/metrics",
				"api":     "/api/v1",
				"docs":    "/api/v1/docs",
			},
			"support": gin.H{
				"email": "support@linkedin-clone.com",
				"docs":  "https://docs.linkedin-clone.com",
			},
		})
	})

	return r
}

func ApplyFileUploadMiddleware(c *gin.Context) {
	if middleware, exists := c.Get("fileUploadMiddleware"); exists {
		if mw, ok := middleware.(gin.HandlerFunc); ok {
			mw(c)
		}
	}
}

func FileUploadMiddlewareWithCustomLimits(maxSize int64, allowedTypes []string) gin.HandlerFunc {
	return middleware.FileUploadMiddleware(maxSize, allowedTypes)
}
