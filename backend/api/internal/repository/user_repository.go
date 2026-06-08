package repository

import (
	"context"
	"errors"

	"netcord/backend/api/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(ctx context.Context, user models.User) (models.User, error)
	GetByEmail(ctx context.Context, email string) (models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (models.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, displayName *string, status string) (models.User, error)
	SetAvatarURL(ctx context.Context, userID uuid.UUID, avatarURL string) (models.User, error)
	SetBannerURL(ctx context.Context, userID uuid.UUID, bannerURL string) (models.User, error)
}

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user models.User) (models.User, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO users (id, username, email, password_hash, display_name, avatar_url, banner_url, status, is_bot)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, username, email, password_hash, display_name, avatar_url, banner_url, status, is_bot, created_at, updated_at
	`, user.ID, user.Username, user.Email, user.PasswordHash, user.DisplayName, user.AvatarURL, user.BannerURL, user.Status, user.IsBot)

	created, err := scanUser(row)
	if err != nil {
		if isUniqueViolation(err) {
			return models.User{}, ErrUserConflict
		}
		return models.User{}, err
	}

	return created, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (models.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, username, email, password_hash, display_name, avatar_url, banner_url, status, is_bot, created_at, updated_at
		FROM users
		WHERE email = $1
	`, email)

	return scanUserOrNotFound(row)
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (models.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, username, email, password_hash, display_name, avatar_url, banner_url, status, is_bot, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id)

	return scanUserOrNotFound(row)
}

func (r *PostgresUserRepository) UpdateProfile(ctx context.Context, userID uuid.UUID, displayName *string, status string) (models.User, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE users
		SET display_name = $2, status = $3, updated_at = now()
		WHERE id = $1
		RETURNING id, username, email, password_hash, display_name, avatar_url, banner_url, status, is_bot, created_at, updated_at
	`, userID, displayName, status)
	return scanUserOrNotFound(row)
}

func (r *PostgresUserRepository) SetAvatarURL(ctx context.Context, userID uuid.UUID, avatarURL string) (models.User, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE users
		SET avatar_url = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, username, email, password_hash, display_name, avatar_url, banner_url, status, is_bot, created_at, updated_at
	`, userID, avatarURL)
	return scanUserOrNotFound(row)
}

func (r *PostgresUserRepository) SetBannerURL(ctx context.Context, userID uuid.UUID, bannerURL string) (models.User, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE users
		SET banner_url = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, username, email, password_hash, display_name, avatar_url, banner_url, status, is_bot, created_at, updated_at
	`, userID, bannerURL)
	return scanUserOrNotFound(row)
}

func scanUserOrNotFound(row pgx.Row) (models.User, error) {
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}
	return user, nil
}

func scanUser(row pgx.Row) (models.User, error) {
	var user models.User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&user.BannerURL,
		&user.Status,
		&user.IsBot,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
