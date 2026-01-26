.PHONY: build run test clean install lint fmt help

# Build variables
BINARY_NAME := lazyoci
BUILD_DIR := bin
MAIN_PATH := ./cmd/lazyoci

# Go variables
GOFLAGS := -ldflags="-s -w"

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## build: Build the binary
build:
	@mkdir -p $(BUILD_DIR)
	go build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

## run: Run the application
run: build
	$(BUILD_DIR)/$(BINARY_NAME)

## test: Run tests
test:
	go test -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

## lint: Run linter
lint:
	golangci-lint run

## fmt: Format code
fmt:
	go fmt ./...
	goimports -w .

## clean: Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

## install: Install the binary
install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

## deps: Download dependencies
deps:
	go mod download
	go mod tidy

## dev: Run with hot reload (requires air)
dev:
	air -c .air.toml
