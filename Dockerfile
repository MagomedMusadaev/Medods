FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Копируем .env файл в контейнер
COPY .env .env

RUN go build -o main ./cmd/app

EXPOSE 8085

CMD ["sh", "-c", "echo 'Starting app' && ./main"]