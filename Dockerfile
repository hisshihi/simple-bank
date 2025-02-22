# Build stage
# Используем официальный образ Go
FROM golang:1.23.6-alpine3.21 AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем текущий каталог в образ
COPY . .

# Собираем бинарный файл
RUN go build -o simple-bank-go main.go

# Run stage
FROM alpine:3.21

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем бинарный файл из builder stage
COPY --from=builder /app/simple-bank-go .
COPY app.env .

# Открываем порт 8080
EXPOSE 8080

# Запускаем бинарный файл
CMD ["./simple-bank-go"]