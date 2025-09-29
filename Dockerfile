# Multi-stage build for Go application
FROM golang:1.25.1-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o miners-scheduler .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

ENV TZ=Europe/Riga

# Create non-root user
RUN addgroup -g 1001 miners && \
    adduser -D -u 1001 -G miners miners

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/miners-scheduler .

# Copy configuration example
COPY config.json /app/config.json

# Create directories for logs and data
RUN mkdir -p /app/logs /app/data && \
    chown -R miners:miners /app

# Switch to non-root user
USER miners

# Expose health check port (if configured)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Default command
ENTRYPOINT ["./miners-scheduler"]
