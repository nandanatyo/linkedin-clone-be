package server

import (
	"context"
	"fmt"
	"linked-clone/internal/config"
	"linked-clone/internal/config/server/routes"
	"linked-clone/pkg/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Server struct {
	router     *gin.Engine
	httpServer *http.Server
	logger     logger.StructuredLogger
}

func NewServer(cfg *config.Config, db *gorm.DB, logger logger.StructuredLogger) (*Server, error) {

	router := config.NewGinEngine(cfg, logger)

	if err := routes.SetupRoutes(router, cfg, db, logger); err != nil {
		return nil, fmt.Errorf("failed to setup routes: %w", err)
	}

	httpServer := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	return &Server{
		router:     router,
		httpServer: httpServer,
		logger:     logger,
	}, nil
}

func (s *Server) Start() error {
	s.logger.Info("Starting server", "addr", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.logger.Info("Shutting down server...")
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) GetRouter() *gin.Engine {
	return s.router
}
