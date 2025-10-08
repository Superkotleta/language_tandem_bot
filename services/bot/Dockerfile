FROM golang:1.25-alpine AS builder
# Устанавливаем gcc и musl-dev для CGO в Alpine
RUN apk add --no-cache gcc musl-dev
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Сборка из каталога с main (cmd/bot)
# CGO_ENABLED=1 нужен для PostgreSQL драйвера
RUN CGO_ENABLED=1 GOOS=linux go build -o bot ./cmd/bot

FROM alpine:3.19
# Устанавливаем ca-certificates для HTTPS запросов
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /app/bot .
COPY --from=builder /app/locales ./locales

CMD ["./bot"]
