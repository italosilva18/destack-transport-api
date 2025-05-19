.PHONY: build run test clean docker docker-compose

# Variáveis
APP_NAME=destack-api
BUILD_DIR=./tmp

build:
	@echo "Building $(APP_NAME)..."
	@go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

run: build
	@echo "Running $(APP_NAME)..."
	@$(BUILD_DIR)/$(APP_NAME)

test:
	@echo "Running tests..."
	@go test -v ./...

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)

docker:
	@echo "Building Docker image..."
	@docker build -t $(APP_NAME) .

docker-compose:
	@echo "Starting services with Docker Compose..."
	@docker-compose up --build