# Build
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o matcher ./cmd/matcher

# Runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates curl
WORKDIR /root/
COPY --from=builder /app/matcher /usr/local/bin/matcher
# copy migrations inside image
COPY migrations /migrations
ENV HTTP_PORT=8082
EXPOSE 8082
CMD ["/usr/local/bin/matcher"]
