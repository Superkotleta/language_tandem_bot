# Builder stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o bot cmd/bot/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies (postgresql-client for migrations if needed)
RUN apk add --no-cache ca-certificates postgresql-client

# Copy binary from builder
COPY --from=builder /build/bot /app/bot

# Copy locales and migrations
COPY locales/ /app/locales/
COPY migrations/ /app/migrations/

# Set environment variables
ENV LOCALES_PATH=/app/locales

# Run the bot
CMD ["/app/bot"]

