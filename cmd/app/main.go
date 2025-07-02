package main

import (
	"context"
	authRepo "linked-clone/internal/api/auth/repository"
	"linked-clone/internal/background"
	"linked-clone/internal/config"
	"linked-clone/internal/config/server"
	"linked-clone/internal/infrastructure/database"
	"linked-clone/pkg/auth"
	"linked-clone/pkg/logger"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	loggerService := logger.NewStructuredLogger()

	cfg, err := config.Load()
	if err != nil {
		loggerService.Fatal("Failed to load configuration", "error", err)
	}

	loggerService.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
		Event:   "application_startup",
		Entity:  "application",
		Success: true,
		Details: map[string]interface{}{
			"environment": cfg.Server.Environment,
			"port":        cfg.Server.Port,
			"version":     "1.0.0",
		},
	})

	db, err := database.NewPostgreSQLConnection(cfg.Database)
	if err != nil {
		loggerService.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
			Event:   "database_connection_failed",
			Entity:  "database",
			Success: false,
			Error:   err.Error(),
		})
		loggerService.Fatal("Failed to connect to database", "error", err)
	}

	loggerService.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
		Event:   "database_connected",
		Entity:  "database",
		Success: true,
		Details: map[string]interface{}{
			"host": cfg.Database.Host,
			"port": cfg.Database.Port,
			"name": cfg.Database.DBName,
		},
	})

	migrationsDir := filepath.Join("migrations")
	if err := database.RunMigrations(cfg.Database, migrationsDir); err != nil {
		loggerService.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
			Event:   "database_migration_failed",
			Entity:  "database",
			Success: false,
			Error:   err.Error(),
		})
		loggerService.Fatal("Failed to run database migrations", "error", err)
	}

	loggerService.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
		Event:   "database_migrated",
		Entity:  "database",
		Success: true,
	})

	srv, err := server.NewServer(cfg, db, loggerService)
	if err != nil {
		loggerService.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
			Event:   "server_initialization_failed",
			Entity:  "server",
			Success: false,
			Error:   err.Error(),
		})
		loggerService.Fatal("Failed to create server", "error", err)
	}

	sessionRepo := authRepo.NewSessionRepository(db)
	jwtService := auth.NewJWTService(cfg.JWT.SecretKey, cfg.JWT.ExpiryHours, sessionRepo)

	backgroundWorker := background.NewBackgroundWorker(background.WorkerConfig{
		CleanupInterval: 5 * time.Minute,
		JWTService:      jwtService,
		Logger:          loggerService,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	backgroundWorker.Start(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		loggerService.Info("Starting server",
			"port", cfg.Server.Port,
			"environment", cfg.Server.Environment)

		if err := srv.Start(); err != nil {
			loggerService.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
				Event:   "server_start_failed",
				Entity:  "server",
				Success: false,
				Error:   err.Error(),
			})
			loggerService.Fatal("Failed to start server", "error", err)
		}
	}()

	loggerService.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
		Event:   "server_started",
		Entity:  "server",
		Success: true,
		Details: map[string]interface{}{
			"port":              cfg.Server.Port,
			"environment":       cfg.Server.Environment,
			"pid":               os.Getpid(),
			"background_worker": "enabled",
		},
	})

	loggerService.Info("Server started successfully",
		"port", cfg.Server.Port,
		"environment", cfg.Server.Environment,
		"pid", os.Getpid(),
		"background_worker", "running")
	loggerService.Info("Press Ctrl+C to shutdown...")

	<-quit

	loggerService.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
		Event:   "shutdown_initiated",
		Entity:  "server",
		Success: true,
	})

	loggerService.Info("Shutting down server...")

	cancel()

	backgroundWorker.Stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	done := make(chan error, 1)

	go func() {
		done <- srv.Shutdown()
	}()

	select {
	case err := <-done:
		if err != nil {
			loggerService.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
				Event:   "shutdown_failed",
				Entity:  "server",
				Success: false,
				Error:   err.Error(),
			})
			loggerService.Error("Failed to shutdown server gracefully", "error", err)
		} else {
			loggerService.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
				Event:   "shutdown_completed",
				Entity:  "server",
				Success: true,
			})
			loggerService.Info("Server shutdown completed successfully")
		}
	case <-shutdownCtx.Done():
		loggerService.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
			Event:   "shutdown_timeout",
			Entity:  "server",
			Success: false,
			Error:   "shutdown timeout exceeded",
		})
		loggerService.Error("Server shutdown timeout exceeded")
	}

	if sqlDB, err := db.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			loggerService.Error("Failed to close database connection", "error", err)
		} else {
			loggerService.Info("Database connection closed")
		}
	}

	loggerService.Info("Application terminated")
}
