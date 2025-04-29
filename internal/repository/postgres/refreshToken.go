package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/medods/auth-service/internal/domain"
	"log/slog"
)

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) SaveRefreshSession(ctx context.Context, session *domain.RefreshSession) error {
	const op = "repository.postgres.SaveRefreshSession"

	query := `
		INSERT INTO refresh_sessions (id, user_id, token_hash, user_ip, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.TokenHash,
		session.UserIP,
		session.CreatedAt,
		session.ExpiresAt,
	)

	if err != nil {
		slog.Error(op,
			"ошибка при сохранении сессии",
			slog.String("user_id", session.UserID.String()),
			slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *RefreshTokenRepository) GetRefreshSession(ctx context.Context, refreshID string) (*domain.RefreshSession, error) {
	const op = "repository.postgres.GetRefreshSession"

	var session domain.RefreshSession
	query := `
		SELECT id, user_id, token_hash, user_ip, created_at, expires_at
		FROM refresh_sessions
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, refreshID).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.UserIP,
		&session.CreatedAt,
		&session.ExpiresAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *RefreshTokenRepository) DeleteRefreshSession(ctx context.Context, refreshID string) error {
	const op = "repository.postgres.DeleteRefreshSession"

	query := `DELETE FROM refresh_sessions WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, refreshID)

	if err != nil {
		slog.Error(op,
			"ошибка при сохранении сессии",
			slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *RefreshTokenRepository) FindSessionByUserID(ctx context.Context, userID string) (bool, error) {
	const op = "repository.postgres.FindSessionByUserID"

	var count int

	query := `SELECT COUNT(*) FROM refresh_sessions WHERE user_id = $1`

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		slog.Error(op,
			"ошибка при получени данных с базы",
			slog.String("user_id", userID),
			slog.String("error", err.Error()))
		return false, err
	}

	return count > 0, nil
}
