package main

// @title Auth Service API
// @version 1.0
// @description Сервис аутентификации и авторизации пользователей
// @host localhost:8085
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

import (
	"context"
	"database/sql"
	"github.com/medods/auth-service/internal/delivery/http"
	"github.com/medods/auth-service/internal/handler"
	"github.com/medods/auth-service/internal/repository/postgres"
	"github.com/medods/auth-service/internal/usecase"
	"github.com/medods/auth-service/pkg/jwt"
	"github.com/medods/auth-service/pkg/smtp"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	_ "github.com/medods/auth-service/docs"
	"github.com/medods/auth-service/internal/config"
	"github.com/medods/auth-service/internal/logger"
)

func main() {
	const op = "cmd.app.main"

	// Загрузка конфигурации
	cfg, err := config.InitConfig()
	if err != nil {
		slog.Error(op, "не удалось загрузить конфигурацию", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Настройка логирования
	if err = logger.SetupLogger(cfg); err != nil {
		slog.Error(op, "ошибка настройки логгера", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("Запуск приложения", "environment", cfg.Env)

	// Установка соединения с базой данных Psql
	db, err := sql.Open("postgres", cfg.Postgres.GetConnectionString())
	if err != nil {
		slog.Error(op, "Ошибка подключения к базе данных:", err)
		os.Exit(1)
	}
	defer db.Close()

	// Проверка соединения с базой данных
	if err = db.Ping(); err != nil {
		slog.Error(op, "Ошибка при проверке соединения с базой данных:", err)
		os.Exit(1)
	}

	// Репозиторий
	authRepo := postgres.NewRefreshTokenRepository(db)

	// PKG
	tokenManager := jwt.NewTokenManager(cfg.JWT.SecretKey, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	smtpManager := smtp.NewEmailSender(&cfg.SMTP)

	// UseCase
	authUseCase := usecase.NewAuthUseCase(tokenManager, authRepo, smtpManager)

	// Handler
	authHandler := handler.NewAuthHandler(authUseCase)

	r := gin.Default()

	http.SetupRoutes(r, authHandler)

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err = r.Run(":" + cfg.ServerConfig.Address); err != nil {
			slog.Error(op, "ошибка при старте сервера", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info(op, "остановка сервера...")
}
