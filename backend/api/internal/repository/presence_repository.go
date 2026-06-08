package repository

import (
	"context"
	"errors"

	"netcord/backend/api/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PresenceRepository interface {
	UpsertPresence(ctx context.Context, presence models.UserPresence) (models.UserPresence, error)
	GetPresence(ctx context.Context, userID uuid.UUID) (models.UserPresence, error)
}

func (r *PostgresUserRepository) UpsertPresence(ctx context.Context, presence models.UserPresence) (models.UserPresence, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO user_presence (user_id, status, custom_status, last_seen_at, updated_at)
		VALUES ($1, $2, $3, now(), now())
		ON CONFLICT (user_id) DO UPDATE
		SET status = EXCLUDED.status,
			custom_status = EXCLUDED.custom_status,
			last_seen_at = CASE
				WHEN EXCLUDED.status = 'offline' THEN now()
				ELSE user_presence.last_seen_at
			END,
			updated_at = now()
		RETURNING user_id, status, custom_status, last_seen_at, updated_at
	`, presence.UserID, presence.Status, presence.CustomStatus)

	return scanPresence(row)
}

func (r *PostgresUserRepository) GetPresence(ctx context.Context, userID uuid.UUID) (models.UserPresence, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT user_id, status, custom_status, last_seen_at, updated_at
		FROM user_presence
		WHERE user_id = $1
	`, userID)

	presence, err := scanPresence(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UserPresence{}, ErrUserNotFound
		}
		return models.UserPresence{}, err
	}
	return presence, nil
}

func scanPresence(row pgx.Row) (models.UserPresence, error) {
	var presence models.UserPresence
	err := row.Scan(
		&presence.UserID,
		&presence.Status,
		&presence.CustomStatus,
		&presence.LastSeenAt,
		&presence.UpdatedAt,
	)
	if err != nil {
		return models.UserPresence{}, err
	}
	return presence, nil
}
