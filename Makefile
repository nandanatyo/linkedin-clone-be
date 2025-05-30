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