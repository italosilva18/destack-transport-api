.PHONY: build run test clean docker docker-compose deps tidy fmt lint

# Variáveis
APP_NAME=destack-api
BUILD_DIR=./tmp

# Comandos principais
build:
	@echo "Building $(APP_NAME)..."
	@go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

run: build
	@echo "Running $(APP_NAME)..."
	@$(BUILD_DIR)/$(APP_NAME)

dev:
	@echo "Running in development mode..."
	@go run ./cmd/server

test:
	@echo "Running tests..."
	@go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# Docker
docker:
	@echo "Building Docker image..."
	@docker build -t $(APP_NAME) .

docker-up:
	@echo "Starting all services with Docker..."
	@docker-compose up -d

docker-down:
	@echo "Stopping all services..."
	@docker-compose down

docker-restart:
	@echo "Restarting all services..."
	@docker-compose restart

docker-logs:
	@echo "Showing logs..."
	@docker-compose logs -f

docker-build:
	@echo "Building services..."
	@docker-compose build --no-cache

docker-dev:
	@echo "Starting development environment..."
	@docker-compose -f docker-compose.dev.yml up -d

docker-dev-down:
	@echo "Stopping development environment..."
	@docker-compose -f docker-compose.dev.yml down

docker-clean:
	@echo "Cleaning Docker resources..."
	@docker-compose down -v
	@docker system prune -f

docker-shell:
	@echo "Opening shell in API container..."
	@docker-compose exec api sh

docker-db-shell:
	@echo "Opening PostgreSQL shell..."
	@docker-compose exec postgres psql -U postgres -d destack_transport

# Dependências
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod verify

tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

# Ferramentas de desenvolvimento
fmt:
	@echo "Formatting code..."
	@go fmt ./...

lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install it with:"; \
		echo "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Banco de dados
db-init:
	@echo "Initializing database..."
	@psql -U postgres -f scripts/init-db.sql

# Geração de código
gen-swagger:
	@echo "Generating Swagger documentation..."
	@if command -v swag > /dev/null; then \
		swag init -g cmd/server/main.go; \
	else \
		echo "swag not installed. Install it with:"; \
		echo "go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

# Ajuda
help:
	@echo "Available commands:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Build and run the application"
	@echo "  make dev            - Run in development mode"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make docker         - Build Docker image"
	@echo "  make docker-compose - Start with Docker Compose"
	@echo "  make deps           - Install dependencies"
	@echo "  make tidy           - Tidy dependencies"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo "  make db-init        - Initialize database"
	@echo "  make help           - Show this help message"