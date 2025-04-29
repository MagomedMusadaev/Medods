# Medods - Сервис аутентификации

## Описание проекта
Medods - это сервис аутентификации с поддержкой JWT токенов и механизмом обновления токенов. 
Сервис обеспечивает безопасную аутентификацию пользователей с использованием пары Access/Refresh токенов.

## Архитектура

Проект построен с использованием Clean Architecture, что обеспечивает высокую гибкость и масштабируемость.

- **Domain Layer** - содержит бизнес-модели
- **Use Case Layer** - реализует бизнес-логику
- **Repository Layer** - отвечает за работу с данными
- **Delivery & Handler Layers** - обрабатывают HTTP запросы

Основные компоненты:
- **TokenManager** - управление JWT токенами
- **AuthUseCase** - бизнес-логика аутентификации
- **AuthHandler** - HTTP обработчики
- **SMTPManager** - отправка уведомлений

## Технологии

- Go 1.21+
- JWT (SHA512)
- PostgreSQL 14+
- Gin Framework
- Slog (для логирования)
- Docker & Docker Compose
- Swagger (API документация)

## Структура проекта

```
├── cmd/           # Точка входа приложения
│   └── app/      # Основное приложение
├── internal/      # Внутренняя логика
│   ├── config/    # Конфигурация
│   ├── domain/    # Бизнес-модели
│   ├── handler/   # HTTP обработчики
│   ├── repository/# Работа с БД
│   └── usecase/   # Бизнес-логика
├── migrations/    # SQL миграции
├── pkg/           # Общие пакеты
│   ├── jwt/      # Работа с JWT
│   └── smtp/     # Email уведомления
└── docs/         # Swagger документация
```

## Требования

- Go 1.21 или выше
- Docker и Docker Compose
- PostgreSQL 14 или выше (если запуск без Docker)

## Установка и запуск

### Через Docker

1. Клонируйте репозиторий
2. Создайте `.env` файл на основе `.env.example` (он описан ниже)
3. Запустите через Docker Compose:
   ```bash
   docker-compose up -d
   ```

### Локальный запуск

1. Клонируйте репозиторий
2. Создайте `.env` файл
3. Создайте базу данных PostgreSQL
4. Установите зависимости:
   ```bash
   go mod download
   ```
5. Примените миграции:
   ```bash
   goose postgres "host=localhost port=5432 user=postgres password=postgres dbname=auth_db sslmode=disable" up
   ```
6. Запустите сервер:
   ```bash
   go run cmd/app/main.go
   ```

## Конфигурация

Настройка через `.env` файл:

```env
# Окружение
ENV=development

# HTTP-сервер
HTTP_SERVER_ADDRESS=8085
HTTP_SERVER_TIMEOUT=5s
HTTP_SERVER_IDLE_TIMEOUT=60s

# JWT
JWT_SECRET_KEY=your-secret-key-here
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=24h
JWT_SIGNING_METHOD=SHA512

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_NAME=auth_db
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSLMODE=disable

# SMTP (если не настроено, уведомления будут в консоли)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=
SMTP_PASSWORD=

# Логи
LOG_FILE=logs/app.log

```

## API Endpoints

### Генерация токенов

```http
POST /auth/tokens?user_id=<uuid>
Content-Type: application/json

Response 200:
{
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}

Response 400:
{
    "error": "невалидный формат user_id"
}

Response 409:
{
    "error": "сессия уже существует"
}
```

### Обновление токенов

```http
POST /auth/refresh
Content-Type: application/json

Request:
{
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}

Response 200:
{
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}

Response 401:
{
    "error": "неверный или истекший refresh токен"
}
```

## Механизм работы токенов

В системе реализована связь между Access и Refresh токенами через уникальный RefreshID, который хранится в базе данных:

1. Access токен:
   - Генерируется как JWT с использованием алгоритма SHA512
   - Содержит информацию о пользователе и RefreshID
   - Не хранится в базе данных
   - Используется для авторизации запросов

2. Refresh токен:
   - Хранится в базе данных в виде bcrypt хеша
   - Связан с конкретным RefreshID
   - Используется для получения новой пары токенов

## Безопасность

- Access токен (JWT) не хранится в базе данных
- Refresh токен хранится в виде bcrypt хеша
- Проверка IP адреса при обновлении токенов
- Защита от повторного использования Refresh токенов
- Отправка уведомлений при изменении IP адреса (через SMTP или в консоль)
- Настраиваемое время жизни токенов
- Безопасное хранение конфигурации через переменные окружения

## Статус реализации

Все требования технического задания успешно реализованы:
- ✅ Аутентификация с использованием JWT
- ✅ Механизм обновления токенов
- ✅ Защита от повторного использования токенов
- ✅ Уведомления при смене IP адреса
- ✅ Безопасное хранение данных

## Email уведомления

В проекте реализована полноценная отправка email уведомлений через SMTP. Для работы необходимо указать корректные SMTP параметры в .env файле. Если параметры не указаны, уведомления будут выводиться в консоль для удобства разработки и тестирования.

## Мониторинг и логирование

- Структурированное логирование через slog
- Логи сохраняются в `logs/app.log`
- Уведомления о подозрительной активности через email
- Детальное логирование ошибок с контекстом