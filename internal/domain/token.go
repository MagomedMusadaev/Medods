package domain

import (
	"github.com/google/uuid"
	"time"
)

type RefreshSession struct {
	ID        string // refresh_id
	UserID    uuid.UUID
	TokenHash string
	UserIP    string
	CreatedAt time.Time
	ExpiresAt time.Time
}
