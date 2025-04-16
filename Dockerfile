# Multi-stage build for TermPOS
# Stage 1: Build environment
FROM golang:1.19-alpine AS builder

# Install build dependencies
RUN apk add --no-cache make gcc libc-dev git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 go build -o termpos ./cmd/pos

# Stage 2: Runtime environment
FROM debian:bullseye-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    sqlite3 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/termpos /app/termpos

# Create necessary directories
RUN mkdir -p /app/data /app/config /app/backups

# Set environment variables
ENV POS_DB_PATH=/app/data/pos.db
ENV POS_CONFIG_PATH=/app/config/config.json

# Expose port for agent mode
EXPOSE 8000

# Set volume mounts for persistent data
VOLUME ["/app/data", "/app/config", "/app/backups"]

# Set entrypoint
ENTRYPOINT ["/app/termpos"]

# Default command
CMD ["--mode", "classic"]