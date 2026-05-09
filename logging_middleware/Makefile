.PHONY: all build run test lint proto clean

APP_NAME=logging_middleware
BIN_DIR=build
API_BIN=$(BIN_DIR)/api
WORKER_BIN=$(BIN_DIR)/worker

all: lint test build

build:
	@echo "Building $(APP_NAME)..."
	go build -o $(API_BIN) ./cmd/api
	go build -o $(WORKER_BIN) ./cmd/worker

run-api:
	@echo "Running API..."
	go run ./cmd/api

run-worker:
	@echo "Running Worker..."
	go run ./cmd/worker

test:
	@echo "Running tests..."
	go test -v -race ./...

lint:
	@echo "Running linter..."
	golangci-lint run

proto:
	@echo "Generating Protobuf files..."
	buf generate

clean:
	@echo "Cleaning..."
	rm -rf $(BIN_DIR)
