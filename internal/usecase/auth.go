package usecase

import (
	"context"
	"fmt"
	"github.com/medods/auth-service/internal/domain"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/medods/auth-service/pkg/jwt"
)

type AuthTokenRepo interface {
	SaveRefreshSession(ctx context.Context, session *domain.RefreshSession) error
	GetRefreshSession(ctx context.Context, refreshID string) (*domain.RefreshSession, error)
	DeleteRefreshSession(ctx context.Context, refreshToken string) error
	FindSessionByUserID(ctx context.Context, userID string) (bool, error)
}

type TokenManager interface {
	GenerateTokenPair(userID uuid.UUID, userIP string) (jwt.TokenPair, string, error)
	HashRefreshToken(refreshToken string) (string, error)
	CompareRefreshToken(hash, refreshToken string) error

	ParseAccessToken(accessToken string) (*jwt.TokenClaims, error)
	ParseRefreshToken(refreshToken string) (*jwt.TokenClaims, error)

	GetAccessTTL() time.Duration
	GetRefreshTTL() time.Duration
}

type SMTPManager interface {
	SendAlert(subject, body string) error
}

type AuthUseCase struct {
	tokenManager    TokenManager
	tokenRepository AuthTokenRepo
	emailSender     SMTPManager
}

func NewAuthUseCase(tokenManager *jwt.TokenManager, tokenRepo AuthTokenRepo, emailSender SMTPManager) *AuthUseCase {
	return &AuthUseCase{
		tokenManager:    tokenManager,
		tokenRepository: tokenRepo,
		emailSender:     emailSender,
	}
}

func (uc *AuthUseCase) GenerateTokens(ctx context.Context, userID uuid.UUID, userIP string) (*jwt.TokenPair, error) {
	const op = "usecase.auth.GenerateTokens"

	sessionExists, err := uc.tokenRepository.FindSessionByUserID(ctx, userID.String())
	if err != nil {
		return nil, fmt.Errorf("внутренняя ошибка при проверке сессии пользователя")
	}

	if sessionExists {
		slog.Warn(op,
			"сессия для пользователя уже существует",
			slog.String("userID", userID.String()),
		)
		return nil, nil
	}

	tokenPair, refreshID, err := uc.tokenManager.GenerateTokenPair(userID, userIP)
	if err != nil {
		slog.Error(op,
			"ошибка генерации токенов",
			slog.String("error", err.Error()),
			slog.Any("userID", userID),
		)
		return nil, err
	}

	ttlRefresh := uc.tokenManager.GetRefreshTTL()

	refreshHash, err := uc.tokenManager.HashRefreshToken(tokenPair.RefreshToken)
	if err != nil {
		slog.Error(op,
			"ошибка при хэшировании refresh токена",
			slog.String("error", err.Error()),
			slog.Any("userID", userID),
		)
		return nil, err
	}

	session := &domain.RefreshSession{
		ID:        refreshID,                  // Связь сессии через UUID
		UserID:    userID,                     // ID пользователя (UUID)
		TokenHash: refreshHash,                // Захешированный refresh токен
		UserIP:    userIP,                     // IP пользователя
		ExpiresAt: time.Now().Add(ttlRefresh), // Срок действия
		CreatedAt: time.Now(),                 // Время создания
	}

	if err = uc.tokenRepository.SaveRefreshSession(ctx, session); err != nil {
		return nil, err
	}

	return &tokenPair, nil
}

func (uc *AuthUseCase) RefreshTokens(ctx context.Context, refreshToken string, userIP string) (*jwt.TokenPair, error) {
	const op = "usecase.auth.RefreshTokens"

	refreshClaim, err := uc.tokenManager.ParseRefreshToken(refreshToken)
	if err != nil {
		slog.Error(op,
			"ошибка при парсинге токена",
			slog.String("error", err.Error()),
			slog.Any("refreshToken", refreshToken),
		)
		return nil, err
	}

	session, err := uc.tokenRepository.GetRefreshSession(ctx, refreshClaim.RefreshID)
	if err != nil {
		slog.Error(op,
			"ошибка при получении сессии",
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("внутренняя ошибка при обработке refresh токена")
	}

	if session == nil {
		slog.Warn(op, "сессия refresh токена не найдена")
		return nil, fmt.Errorf("refresh токен не найден")
	}

	if err = uc.tokenManager.CompareRefreshToken(session.TokenHash, refreshToken); err != nil {
		slog.Warn(op, "несоответствие refresh токена", slog.String("refresh_token", refreshToken))
		return nil, fmt.Errorf("неверный refresh токен")
	}

	if time.Now().After(session.ExpiresAt) {
		_ = uc.tokenRepository.DeleteRefreshSession(ctx, refreshToken)

		slog.Warn(op, "refresh токен истёк", slog.String("refresh_token", refreshToken))
		return nil, fmt.Errorf("refresh токен истёк")
	}

	if session.UserIP != userIP {
		slog.Warn(op,
			"несоответствие IP адреса",
			slog.String("stored_ip", session.UserIP),
			slog.String("current_ip", userIP),
		)

		err = uc.emailSender.SendAlert("Предупреждение: Несоответствие IP", fmt.Sprintf("IP адрес для токена %s изменился. Старый: %s, Новый: %s", refreshToken, session.UserIP, userIP))
		if err != nil {
			slog.Error(op, "не удалось отправить уведомление", slog.String("error", err.Error()))
		}
	}

	if err = uc.tokenRepository.DeleteRefreshSession(ctx, refreshClaim.RefreshID); err != nil {
		return nil, fmt.Errorf("внутренняя ошибка при удалении старой сессии")
	}

	return uc.GenerateTokens(ctx, session.UserID, userIP)
}
