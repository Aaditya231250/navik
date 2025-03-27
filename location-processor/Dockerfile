# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Install build dependencies for librdkafka
RUN apk add --no-cache gcc musl-dev pkgconf

# Copy dependencies first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled and musl tags
RUN CGO_ENABLED=1 GOOS=linux go build -tags musl -o /app/main

# Runtime stage
FROM alpine:3.19
WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add librdkafka-dev

# Copy built binary
COPY --from=builder /app/main /app/main

# Create log directory and set permissions
RUN mkdir -p /data/logs && chown -R 1000:1000 /data/logs && chmod -R 755 /data

# Set up entrypoint
ENTRYPOINT ["/app/main"]
