package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerConfig ServerConfig
	Postgres     Postgres
	Log          LogConfig
	JWT          JWT
	SMTP         SMTPConfig
	Env          string
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	UseTLS   bool
}

type ServerConfig struct {
	Address     string
	Timeout     time.Duration
	IdleTimeout time.Duration
}

type JWT struct {
	SecretKey     string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	SigningMethod string
}

type Postgres struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type LogConfig struct {
	FilePath    string
	Environment string
}

func (c *LogConfig) GetLogDir() string {
	return filepath.Dir(c.FilePath)
}

func InitConfig() (*Config, error) {
	// Поиск корня проекта
	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить рабочую директорию: %w", err)
	}

	rootDir := workDir
	for {
		if _, err = os.Stat(filepath.Join(rootDir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(rootDir)
		if parent == rootDir {
			return nil, errors.New("не удалось найти корень проекта (go.mod не найден)")
		}
		rootDir = parent
	}

	if err = godotenv.Load(filepath.Join(rootDir, ".env")); err != nil {
		return nil, fmt.Errorf("не удалось загрузить .env: %w", err)
	}

	parseDuration := func(envKey, defaultVal string) time.Duration {
		dur, err := time.ParseDuration(getEnv(envKey, defaultVal))
		if err != nil {
			panic(fmt.Sprintf("невалидная длительность %s: %v", envKey, err))
		}
		return dur
	}

	cfg := &Config{
		Env: getEnv("ENV", "development"),
		ServerConfig: ServerConfig{
			Address:     getEnv("HTTP_SERVER_ADDRESS", "8080"),
			Timeout:     parseDuration("HTTP_SERVER_TIMEOUT", "5s"),
			IdleTimeout: parseDuration("HTTP_SERVER_IDLE_TIMEOUT", "60s"),
		},
		JWT: JWT{
			SecretKey:     getEnv("JWT_SECRET_KEY", "my_secret_key"),
			AccessTTL:     parseDuration("JWT_ACCESS_TTL", "15m"),
			RefreshTTL:    parseDuration("JWT_REFRESH_TTL", "24h"),
			SigningMethod: getEnv("JWT_SIGNING_METHOD", "SHA512"),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:     getEnvAsInt("SMTP_PORT", 587),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", ""),
			UseTLS:   getEnv("SMTP_USE_TLS", "true") == "true",
		},
		Postgres: Postgres{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "postgres"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Log: LogConfig{
			FilePath:    filepath.Join(rootDir, getEnv("LOG_FILE", "logs/app.log")),
			Environment: getEnv("ENV", "development"),
		},
	}

	if err = cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Postgres.Host == "" {
		return fmt.Errorf("хост базы данных не может быть пустым")
	}
	if c.Postgres.Port <= 0 || c.Postgres.Port > 65535 {
		return fmt.Errorf("недопустимый порт базы данных")
	}
	if c.Postgres.User == "" {
		return fmt.Errorf("имя пользователя базы данных не может быть пустым")
	}
	if c.ServerConfig.Address == "" {
		return fmt.Errorf("адрес сервера не может быть пустым")
	}
	validEnvs := map[string]bool{"development": true, "production": true, "test": true}
	if !validEnvs[strings.ToLower(c.Env)] {
		return fmt.Errorf("недопустимое окружение: %s", c.Env)
	}
	return nil
}

func (c *Postgres) GetConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var i int
		if _, err := fmt.Sscanf(value, "%d", &i); err == nil {
			return i
		}
	}
	return defaultValue
}
