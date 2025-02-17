.PHONY: build run test lint clean

# Build variables
BINARY_NAME=chat-server
GO_FILES=$(shell find . -name '*.go')

# Go commands
GO=go
GOTEST=$(GO) test
GOBUILD=$(GO) build

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/server

# Run the application
run:
	$(GO) run ./cmd/server

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out

# Run linter
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out

# Install dependencies
deps:
	$(GO) mod download

# Update dependencies
deps-update:
	$(GO) get -u ./...
	$(GO) mod tidy

# Generate mocks (requires mockgen)
mocks:
	mockgen -source=internal/repository/interfaces.go -destination=internal/repository/mocks/repository_mocks.go
	mockgen -source=internal/service/user_service.go -destination=internal/service/mocks/user_service_mocks.go
	mockgen -source=internal/service/group_service.go -destination=internal/service/mocks/group_service_mocks.go
	mockgen -source=internal/service/message_service.go -destination=internal/service/mocks/message_service_mocks.go

# Create initial database
init-db:
	psql -U postgres -c "CREATE DATABASE chat;"
	cqlsh -e "CREATE KEYSPACE IF NOT EXISTS chat WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};"

# Run all pre-commit checks
check: lint test

# Help
help:
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  lint           - Run linter"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Install dependencies"
	@echo "  deps-update    - Update dependencies"
	@echo "  mocks          - Generate mocks"
	@echo "  init-db        - Create initial database"
	@echo "  check          - Run all pre-commit checks"

# Default target
default: build 