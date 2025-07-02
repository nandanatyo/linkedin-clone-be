package main

import (
	"context"
	"linked-clone/internal/config"
	"linked-clone/internal/config/server"
	"linked-clone/internal/infrastructure/database"
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

	structuredLogger := logger.NewStructuredLogger()

	cfg, err := config.Load()
	if err != nil {
		structuredLogger.Fatal("Failed to load configuration", "error", err)
	}

	structuredLogger.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
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
		structuredLogger.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
			Event:   "database_connection_failed",
			Entity:  "database",
			Success: false,
			Error:   err.Error(),
		})
		structuredLogger.Fatal("Failed to connect to database", "error", err)
	}

	structuredLogger.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
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
		structuredLogger.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
			Event:   "database_migration_failed",
			Entity:  "database",
			Success: false,
			Error:   err.Error(),
		})
		structuredLogger.Fatal("Failed to run database migrations", "error", err)
	}

	structuredLogger.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
		Event:   "database_migrated",
		Entity:  "database",
		Success: true,
	})

	srv, err := server.NewServer(cfg, db, structuredLogger)
	if err != nil {
		structuredLogger.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
			Event:   "server_initialization_failed",
			Entity:  "server",
			Success: false,
			Error:   err.Error(),
		})
		structuredLogger.Fatal("Failed to create server", "error", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		structuredLogger.Info("Starting server",
			"port", cfg.Server.Port,
			"environment", cfg.Server.Environment)

		if err := srv.Start(); err != nil {
			structuredLogger.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
				Event:   "server_start_failed",
				Entity:  "server",
				Success: false,
				Error:   err.Error(),
			})
			structuredLogger.Fatal("Failed to start server", "error", err)
		}
	}()

	structuredLogger.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
		Event:   "server_started",
		Entity:  "server",
		Success: true,
		Details: map[string]interface{}{
			"port":        cfg.Server.Port,
			"environment": cfg.Server.Environment,
			"pid":         os.Getpid(),
		},
	})

	structuredLogger.Info("Server started successfully",
		"port", cfg.Server.Port,
		"environment", cfg.Server.Environment,
		"pid", os.Getpid())
	structuredLogger.Info("Press Ctrl+C to shutdown...")

	<-quit

	structuredLogger.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
		Event:   "shutdown_initiated",
		Entity:  "server",
		Success: true,
	})

	structuredLogger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- srv.Shutdown()
	}()

	select {
	case err := <-done:
		if err != nil {
			structuredLogger.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
				Event:   "shutdown_failed",
				Entity:  "server",
				Success: false,
				Error:   err.Error(),
			})
			structuredLogger.Error("Failed to shutdown server gracefully", "error", err)
		} else {
			structuredLogger.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
				Event:   "shutdown_completed",
				Entity:  "server",
				Success: true,
			})
			structuredLogger.Info("Server shutdown completed successfully")
		}
	case <-shutdownCtx.Done():
		structuredLogger.LogBusinessEvent(context.Background(), logger.BusinessEventLog{
			Event:   "shutdown_timeout",
			Entity:  "server",
			Success: false,
			Error:   "shutdown timeout exceeded",
		})
		structuredLogger.Error("Server shutdown timeout exceeded")
	}

	if sqlDB, err := db.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			structuredLogger.Error("Failed to close database connection", "error", err)
		} else {
			structuredLogger.Info("Database connection closed")
		}
	}

	structuredLogger.Info("Application terminated")
}
