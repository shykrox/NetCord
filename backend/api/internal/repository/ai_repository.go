package repository

import (
	"context"
	"errors"

	"netcord/backend/api/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AIRepository interface {
	CreateJob(ctx context.Context, job models.AIJob) (models.AIJob, error)
	ListJobsForUser(ctx context.Context, userID uuid.UUID, limit int) ([]models.AIJob, error)
	GetJobForUser(ctx context.Context, jobID, userID uuid.UUID) (models.AIJob, error)
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

func (r *PostgresAIRepository) ListJobsForUser(ctx context.Context, userID uuid.UUID, limit int) ([]models.AIJob, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT j.id, j.server_id, j.channel_id, j.user_id, j.command, j.prompt, j.status, j.progress,
			j.result_message_id, j.error, j.created_at, j.updated_at, j.completed_at
		FROM ai_jobs j
		JOIN server_members sm ON sm.server_id = j.server_id AND sm.user_id = $1
		ORDER BY j.created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]models.AIJob, 0)
	for rows.Next() {
		job, err := scanAIJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *PostgresAIRepository) GetJobForUser(ctx context.Context, jobID, userID uuid.UUID) (models.AIJob, error) {
	job, err := scanAIJob(r.pool.QueryRow(ctx, `
		SELECT j.id, j.server_id, j.channel_id, j.user_id, j.command, j.prompt, j.status, j.progress,
			j.result_message_id, j.error, j.created_at, j.updated_at, j.completed_at
		FROM ai_jobs j
		JOIN server_members sm ON sm.server_id = j.server_id AND sm.user_id = $2
		WHERE j.id = $1
	`, jobID, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.AIJob{}, ErrAIJobNotFound
		}
		return models.AIJob{}, err
	}
	return job, nil
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
