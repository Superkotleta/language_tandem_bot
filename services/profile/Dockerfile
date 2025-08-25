# Build
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o profile ./cmd/profile

# Runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates curl
WORKDIR /root/
COPY --from=builder /app/profile /usr/local/bin/profile
# copy migrations inside image
COPY migrations /migrations
ENV HTTP_PORT=8081
EXPOSE 8081
CMD ["profile"]
