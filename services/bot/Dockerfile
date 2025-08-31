FROM golang:1.21-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Сборка из каталога с main (cmd/bot)
RUN CGO_ENABLED=0 GOOS=linux go build -o bot ./cmd/bot

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bot .
COPY --from=builder /app/locales ./locales

CMD ["./bot"]
