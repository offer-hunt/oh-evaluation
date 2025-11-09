# Этап 1: Сборка
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем статичный бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o evaluation-service ./cmd/evaluation-service

# Этап 2: Запуск
FROM alpine:latest

# TLS корни, чтобы ходить по HTTPS (например, к JWKS)
RUN apk add --no-cache ca-certificates

WORKDIR /root/

# Копируем бинарник из этапа сборки
COPY --from=builder /app/evaluation-service .

# Открываем порт
EXPOSE 8080

# Запускаем приложение
ENTRYPOINT ["./evaluation-service"]
