# Используем официальный образ Go
FROM golang:1.22.2

# Создаем рабочую директорию
WORKDIR /app

# Копируем файлы модулей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN go build -o main .

# Открываем порт
EXPOSE 5002

# Запускаем приложение
CMD ["./main"]

