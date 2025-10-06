FROM golang:1.25 AS builder
# Устанавливаем gcc для CGO (Ubuntu уже имеет его)
RUN apt-get update && apt-get install -y gcc
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Сборка из каталога с main (cmd/bot)
# CGO_ENABLED=1 нужен для PostgreSQL драйвера
RUN CGO_ENABLED=1 GOOS=linux go build -o bot ./cmd/bot

FROM ubuntu:22.04
# Устанавливаем ca-certificates для HTTPS запросов
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /root/
COPY --from=builder /app/bot .
COPY --from=builder /app/locales ./locales

CMD ["./bot"]
