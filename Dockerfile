# Multi-stage Dockerfile for Jta

# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/jta \
    ./cmd/jta

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -S jta && adduser -S jta -G jta

# Set working directory
WORKDIR /home/jta

# Copy binary from builder
COPY --from=builder /app/jta /usr/local/bin/jta

# Copy examples (optional)
COPY examples /home/jta/examples

# Change ownership
RUN chown -R jta:jta /home/jta

# Switch to non-root user
USER jta

# Set entrypoint
ENTRYPOINT ["jta"]

# Default command (show help)
CMD ["--help"]
