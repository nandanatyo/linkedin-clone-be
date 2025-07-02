package main

import (
	"flag"
	"fmt"
	"linked-clone/internal/config"
	"linked-clone/internal/infrastructure/database"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	migrationsDir := filepath.Join("internal", "infrastructure", "database", "migrations")

	var command string
	var name string

	flag.StringVar(&command, "command", "", "Migration command: up, down, status, create, reset")
	flag.StringVar(&name, "name", "", "Migration name (for create command)")
	flag.Parse()

	if command == "" {
		fmt.Println("Usage:")
		fmt.Println("  go run cmd/migrate/main.go -command=up              # Run all pending migrations")
		fmt.Println("  go run cmd/migrate/main.go -command=down            # Rollback last migration")
		fmt.Println("  go run cmd/migrate/main.go -command=status          # Show migration status")
		fmt.Println("  go run cmd/migrate/main.go -command=create -name=migration_name  # Create new migration")
		fmt.Println("  go run cmd/migrate/main.go -command=reset           # Reset all migrations")
		os.Exit(1)
	}

	switch command {
	case "up":
		if err := database.RunMigrations(cfg.Database, migrationsDir); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		fmt.Println("Migrations completed successfully")

	case "status":
		if err := database.MigrationStatus(cfg.Database, migrationsDir); err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}

	case "create":
		if name == "" {
			log.Fatal("Migration name is required for create command")
		}
		if err := database.CreateMigration(migrationsDir, name); err != nil {
			log.Fatalf("Failed to create migration: %v", err)
		}
		fmt.Printf("Migration created successfully: %s\n", name)

	default:
		log.Fatalf("Unknown command: %s", command)
	}
}
