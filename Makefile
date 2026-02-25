.PHONY: all build proto docker-up docker-down test clean

# Variables
PROJECT_NAME=cloud_storage
DOCKER_COMPOSE=docker compose

all: proto build

# Generate protobuf files
proto:
	@echo "Generating protobuf files..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		internal/api/auth.proto
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		internal/api/metadata.proto
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		internal/api/file.proto
	@echo "Protobuf files generated."

# Build project
build: proto
	@echo "Building services..."
	@go build -o bin/auth-service ./cmd/auth
	@go build -o bin/metadata-service ./cmd/metadata
	@go build -o bin/file-service ./cmd/file
	@go build -o bin/gateway ./cmd/gateway
	@echo "Build complete."

# Build Docker images
docker-build: build
	@echo "Building Docker images..."
	@docker build -f deployments/docker/Dockerfile.auth -t cloud-storage/auth-service:latest .
	@docker build -f deployments/docker/Dockerfile.metadata -t cloud-storage/metadata-service:latest .
	@docker build -f deployments/docker/Dockerfile.file -t cloud-storage/file-service:latest .
	@docker build -f deployments/docker/Dockerfile.gateway -t cloud-storage/gateway:latest .
	@echo "Docker images built."

# Start services with Docker Compose
docker-up:
	@echo "Starting services with Docker Compose..."
	@$(DOCKER_COMPOSE) up -d

# Stop Docker Compose
docker-down:
	@echo "Stopping Docker Compose services..."
	@$(DOCKER_COMPOSE) down

# View logs
logs:
	@$(DOCKER_COMPOSE) logs -f

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@go clean
	@echo "Clean complete."

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies downloaded."

# Help
help:
	@echo "Available targets:"
	@echo "  make proto        - Generate protobuf files"
	@echo "  make build        - Build all services"
	@echo "  make docker-build - Build Docker images"
	@echo "  make docker-up    - Start services with Docker Compose"
	@echo "  make docker-down  - Stop services"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make deps         - Download dependencies"
	@echo "  make help         - Show help"