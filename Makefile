# Makefile for TermPOS

# Build variables
BINARY_NAME=termpos
VERSION=$(shell grep -o 'Version     = "[^"]*"' cmd/pos/version.go | cut -d'"' -f2)
BUILD_DATE=$(shell date +%Y-%m-%d)
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.GitCommit=${GIT_COMMIT}"

# Output directories
BIN_DIR=./bin

# Default target
.PHONY: all
all: clean build

# Build for the current platform
.PHONY: build
build:
	@echo "Building ${BINARY_NAME}..."
	@go build ${LDFLAGS} -o ${BINARY_NAME} ./cmd/pos

# Cross-platform builds
.PHONY: build-all
build-all: build-linux build-windows build-darwin

.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p ${BIN_DIR}/linux
	@GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BIN_DIR}/linux/${BINARY_NAME} ./cmd/pos

.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p ${BIN_DIR}/windows
	@GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BIN_DIR}/windows/${BINARY_NAME}.exe ./cmd/pos

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p ${BIN_DIR}/darwin
	@GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ${BIN_DIR}/darwin/${BINARY_NAME} ./cmd/pos

# Docker targets
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	@docker build -t ${BINARY_NAME}:${VERSION} .

.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	@docker run -it --rm -p 8000:8000 \
		-v termpos-data:/app/data \
		-v termpos-config:/app/config \
		-v termpos-backups:/app/backups \
		${BINARY_NAME}:${VERSION}

.PHONY: docker-compose-up
docker-compose-up:
	@echo "Starting with docker-compose..."
	@docker-compose up -d

.PHONY: docker-compose-down
docker-compose-down:
	@echo "Stopping docker-compose services..."
	@docker-compose down

# Clean up build artifacts
.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -f ${BINARY_NAME}
	@rm -f ${BINARY_NAME}.exe
	@rm -rf ${BIN_DIR}

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

# Initialize the project for first use
.PHONY: init
init:
	@echo "Initializing project..."
	@go run ./cmd/pos init

# Help information
.PHONY: help
help:
	@echo "TermPOS Make Targets:"
	@echo "  all           - Clean and build the application"
	@echo "  build         - Build for the current platform"
	@echo "  build-all     - Build for Linux, Windows, and macOS"
	@echo "  build-linux   - Build for Linux"
	@echo "  build-windows - Build for Windows"
	@echo "  build-darwin  - Build for macOS"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  docker-compose-up   - Start services with docker-compose"
	@echo "  docker-compose-down - Stop services with docker-compose"
	@echo "  clean         - Remove build artifacts"
	@echo "  test          - Run tests"
	@echo "  init          - Initialize the project for first use"
	@echo "  help          - Show this help message"