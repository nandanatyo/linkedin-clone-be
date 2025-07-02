package server

import (
	"context"
	"fmt"
	"linked-clone/internal/background"
	"linked-clone/internal/config"
	"linked-clone/internal/config/server/routes"
	"linked-clone/pkg/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Server struct {
	router                *gin.Engine
	httpServer            *http.Server
	logger                logger.StructuredLogger
	sessionCleanupService *background.SessionCleanupService
}

func NewServer(cfg *config.Config, db *gorm.DB, logger logger.StructuredLogger) (*Server, error) {

	router := config.NewGinEngine(cfg, logger)

	deps, err := routes.InitializeDependencies(cfg, db, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize dependencies: %w", err)
	}

	if err := routes.SetupRoutes(router, cfg, db, logger); err != nil {
		return nil, fmt.Errorf("failed to setup routes: %w", err)
	}

	sessionCleanupService := background.NewSessionCleanupService(deps.JWTService, logger)

	httpServer := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	return &Server{
		router:                router,
		httpServer:            httpServer,
		logger:                logger,
		sessionCleanupService: sessionCleanupService,
	}, nil
}

func (s *Server) Start() error {

	ctx := context.Background()
	s.sessionCleanupService.Start(ctx)

	s.logger.Info("Starting HTTP server", "addr", s.httpServer.Addr)
	s.logger.Info("Session cleanup service started")

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown() error {

	s.sessionCleanupService.Stop()
	s.logger.Info("Session cleanup service stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.logger.Info("Shutting down HTTP server...")
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

func (s *Server) IsHealthy() map[string]interface{} {
	return map[string]interface{}{
		"server_status":           "running",
		"cleanup_service_running": s.sessionCleanupService.IsRunning(),
		"timestamp":               time.Now().UTC().Format(time.RFC3339),
	}
}
