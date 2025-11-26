# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install CA certificates for HTTPS requests
RUN apk add --no-cache ca-certificates

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with optimizations for smaller binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o opgl-gateway main.go

# Production stage
FROM alpine:3.19

WORKDIR /app

# Install CA certificates and curl for health checks
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copy binary from builder
COPY --from=builder /app/opgl-gateway .

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check using POST method
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f -X POST http://localhost:8080/health || exit 1

# Run the application
CMD ["./opgl-gateway"]
