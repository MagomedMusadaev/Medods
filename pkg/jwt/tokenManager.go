package jwt

import (
	"errors"
	"fmt"
	"github.com/cespare/xxhash/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

type TokenManager struct {
	secretKey  string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type TokenClaims struct {
	UserID    uuid.UUID
	UserIP    string
	RefreshID string
	jwt.RegisteredClaims
}

func NewTokenManager(secretKey string, accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{
		secretKey:  secretKey,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (tm *TokenManager) GenerateTokenPair(userID uuid.UUID, userIP string) (TokenPair, string, error) {
	refreshID := uuid.NewString()

	accessToken, err := tm.generateAccessToken(userID, userIP, refreshID)
	if err != nil {
		return TokenPair{}, "", err
	}

	refreshToken, err := tm.generateRefreshToken(refreshID)
	if err != nil {
		return TokenPair{}, "", err
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, refreshID, nil
}

func (tm *TokenManager) generateAccessToken(userID uuid.UUID, userIP, refreshID string) (string, error) {
	claims := TokenClaims{
		UserID:    userID,
		UserIP:    userIP,
		RefreshID: refreshID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tm.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString([]byte(tm.secretKey))
}

func (tm *TokenManager) generateRefreshToken(refreshID string) (string, error) {
	claims := TokenClaims{
		RefreshID: refreshID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tm.refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString([]byte(tm.secretKey))
}

func (tm *TokenManager) ParseAccessToken(accessToken string) (*TokenClaims, error) {
	return tm.parseTokens(accessToken)
}

func (tm *TokenManager) ParseRefreshToken(refreshToken string) (*TokenClaims, error) {
	return tm.parseTokens(refreshToken)
}

func (tm *TokenManager) parseTokens(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS512 {
			return nil, ErrInvalidToken
		}
		return []byte(tm.secretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (tm *TokenManager) HashRefreshToken(refreshToken string) (string, error) {
	tokenBytes := []byte(refreshToken)

	hasher := xxhash.New()
	_, err := hasher.Write(tokenBytes)
	if err != nil {
		return "", err
	}

	hash := hasher.Sum64()

	return fmt.Sprintf("%x", hash), nil
}

func (tm *TokenManager) CompareRefreshToken(hash, refreshToken string) error {
	h := xxhash.New()
	h.Write([]byte(refreshToken))
	computedHash := h.Sum64()

	storedHashValue, err := parseStoredHash(hash)
	if err != nil {
		return fmt.Errorf("невалидный формат сохраненного хеша: %w", err)
	}

	if storedHashValue != computedHash {
		return fmt.Errorf("токен не соответствует сохраненному хешу")
	}

	return nil
}

func parseStoredHash(storedHash string) (uint64, error) {
	var hashValue uint64
	_, err := fmt.Sscanf(storedHash, "%x", &hashValue)
	if err != nil {
		return 0, err
	}
	return hashValue, nil
}

func (tm *TokenManager) GetAccessTTL() time.Duration {
	return tm.accessTTL
}

func (tm *TokenManager) GetRefreshTTL() time.Duration {
	return tm.refreshTTL
}
