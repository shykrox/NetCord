package repository

import (
	"context"

	"netcord/backend/api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AIRepository interface {
	CreateJob(ctx context.Context, job models.AIJob) (models.AIJob, error)
}

type PostgresAIRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAIRepository(pool *pgxpool.Pool) *PostgresAIRepository {
	return &PostgresAIRepository{pool: pool}
}

func (r *PostgresAIRepository) CreateJob(ctx context.Context, job models.AIJob) (models.AIJob, error) {
	created, err := scanAIJob(r.pool.QueryRow(ctx, `
		INSERT INTO ai_jobs (id, server_id, channel_id, user_id, command, prompt, status, progress)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, server_id, channel_id, user_id, command, prompt, status, progress,
			result_message_id, error, created_at, updated_at, completed_at
	`, job.ID, job.ServerID, job.ChannelID, job.UserID, job.Command, job.Prompt, job.Status, job.Progress))
	if err != nil {
		return models.AIJob{}, err
	}
	return created, nil
}

func scanAIJob(row pgx.Row) (models.AIJob, error) {
	var job models.AIJob
	var resultMessageID pgtype.UUID
	var completedAt pgtype.Timestamptz
	err := row.Scan(
		&job.ID,
		&job.ServerID,
		&job.ChannelID,
		&job.UserID,
		&job.Command,
		&job.Prompt,
		&job.Status,
		&job.Progress,
		&resultMessageID,
		&job.Error,
		&job.CreatedAt,
		&job.UpdatedAt,
		&completedAt,
	)
	if err != nil {
		return models.AIJob{}, err
	}
	job.ResultMessageID = nullableUUID(resultMessageID)
	job.CompletedAt = nullableTime(completedAt)
	return job, nil
}
