package handler

import (
	"context"
	"github.com/medods/auth-service/pkg/jwt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthTokenUseCase interface {
	GenerateTokens(ctx context.Context, userID uuid.UUID, userIP string) (*jwt.TokenPair, error)
	RefreshTokens(ctx context.Context, refreshTokenBase64 string, userIP string) (*jwt.TokenPair, error)
}

type AuthHandler struct {
	tokenUseCase AuthTokenUseCase
}

func NewAuthHandler(tokenUseCase AuthTokenUseCase) *AuthHandler {
	return &AuthHandler{
		tokenUseCase: tokenUseCase,
	}
}

// refreshRequest представляет запрос на обновление токенов
type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// @Summary Генерация токенов
// @Description Генерирует пару access и refresh токенов для пользователя
// @Tags auth
// @Produce json
// @Param user_id query string true "ID пользователя в формате UUID"
// @Success 200 {object} jwt.TokenPair "Успешная генерация токенов"
// @Failure 400 {object} map[string]string "Ошибка валидации (неправильный формат user_id или отсутствует параметр)"
// @Failure 409 {object} map[string]string "Сессия уже существует (токены уже были сгенерированы для данного пользователя)"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /auth/tokens [post]
func (h *AuthHandler) GenerateTokens(c *gin.Context) {
	const op = "handler.auth.GenerateTokens"

	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id обязателен"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		slog.Error(op, "невалидный формат user_id", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный формат user_id"})
		return
	}

	userIP := c.ClientIP()

	// Генерируем токены
	tokens, err := h.tokenUseCase.GenerateTokens(c.Request.Context(), userID, userIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}
	if tokens == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "сессия уже существует"})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// @Summary Обновление токенов
// @Description Обновляет пару токенов используя refresh токен
// @Tags auth
// @Accept json
// @Produce json
// @Param request body refreshRequest true "Refresh токен"
// @Success 200 {object} jwt.TokenPair "Успешное обновление токенов"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 401 {object} map[string]string "Неверный или истекший токен"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshTokens(c *gin.Context) {
	const op = "handler.auth.RefreshTokens"

	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(op, "невалидное тело запроса", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидное тело запроса"})
		return
	}

	userIP := c.ClientIP()

	// Обновляем токены
	tokens, err := h.tokenUseCase.RefreshTokens(c.Request.Context(), req.RefreshToken, userIP)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "неверный или истекший refresh токен"})
		return
	}

	c.JSON(http.StatusOK, tokens)
}
