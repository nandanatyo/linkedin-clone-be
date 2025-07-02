# Makefile
.PHONY: test test-auth test-clean test-setup test-db

# Setup test database
test-db:
	@echo "Setting up test database..."
	@dropdb linkedin_clone_test 2>/dev/null || true
	@createdb linkedin_clone_test
	@echo "Test database ready!"

# Clean test artifacts
test-clean:
	@echo "Cleaning test artifacts..."
	@go clean -testcache
	@rm -f *.prof

# Run all tests
test: test-clean
	@echo "Running all tests..."
	@go test ./test/... -v -count=1

# Run auth tests only
test-auth: test-clean
	@echo "Running auth tests..."
	@go test ./test/auth_test.go ./test/base_test.go -v

# Run tests with coverage
test-coverage: test-clean
	@echo "Running tests with coverage..."
	@go test ./test/... -v -cover -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html

# Run tests with race detection
test-race: test-clean
	@echo "Running tests with race detection..."
	@go test ./test/... -v -race

# Setup and run tests
test-setup: test-db test

.PHONY: migrate-up migrate-down migrate-status migrate-create migrate-reset

# Run all pending migrations
migrate-up:
	go run cmd/migrate/main.go -command=up

# Show migration status
migrate-status:
	go run cmd/migrate/main.go -command=status

# Create a new migration
migrate-create:
	@read -p "Enter migration name: " name; \
	go run cmd/migrate/main.go -command=create -name=$$name

# Run migrations for development
dev-migrate:
	go run cmd/migrate/main.go -command=up

# Database setup for development
dev-setup: dev-migrate
	@echo "Database setup completed"

.DEFAULT_GOAL := help
help:
	@echo "Available commands:"
	@echo "  migrate-up      - Run all pending migrations"
	@echo "  migrate-status  - Show migration status"
	@echo "  migrate-create  - Create a new migration (interactive)"
	@echo "  dev-migrate     - Run migrations for development"
	@echo "  dev-setup       - Complete database setup for development"
