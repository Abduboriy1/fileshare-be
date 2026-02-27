# ============================================================
# Fileshare Backend - Makefile
# ============================================================
# Targets:
#   build   - Compile the binary
#   run     - Run the server
#   dev     - Run with air (hot reload)
#   migrate - Run auto-migration (starts the app)
#   test    - Run tests
#   lint    - Run golangci-lint
#   clean   - Remove build artifacts
# ============================================================

BINARY_NAME=fileshare-be
BUILD_DIR=bin

.PHONY: build run dev migrate test lint clean

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/main.go

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

dev:
	air

migrate:
	go run ./cmd/main.go

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)
	go clean
