package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"linked-clone/internal/config"
	"linked-clone/pkg/logger"
)

func SetupRoutes(router *gin.Engine, cfg *config.Config, db *gorm.DB, logger logger.StructuredLogger) error {
	deps, err := InitializeDependencies(cfg, db, logger)
	if err != nil {
		return err
	}

	v1 := router.Group("/api/v1")
	{

		AuthRoutes(v1, deps)

		UserRoutes(v1, deps)

		PostRoutes(v1, deps)

		JobRoutes(v1, deps)

	}

	return nil
}
