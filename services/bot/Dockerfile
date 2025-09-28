FROM golang:1.22-alpine AS builder
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bot-api ./cmd/bot-api

# Final stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

# Copy the binary
COPY --from=builder /app/bot-api /usr/local/bin/bot-api

# Copy locales
COPY --from=builder /app/locales ./locales

# Set environment variables
ENV HTTP_PORT=8082

# Expose port
EXPOSE 8082

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8082/healthz || exit 1

# Run the application
CMD ["bot-api"]
