package repository

import (
	"context"
	"errors"

	"netcord/backend/api/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultChannelPosition = 0

type ServerRepository interface {
	CreateServer(ctx context.Context, server models.Server, ownerID uuid.UUID) (models.Server, error)
	ListServersForUser(ctx context.Context, userID uuid.UUID) ([]models.Server, error)
	GetServerForUser(ctx context.Context, serverID, userID uuid.UUID) (models.Server, error)
	CreateChannel(ctx context.Context, channel models.Channel) (models.Channel, error)
	ListChannelsForUser(ctx context.Context, serverID, userID uuid.UUID) ([]models.Channel, error)
	GetChannelForUser(ctx context.Context, channelID, userID uuid.UUID) (models.Channel, error)
	CreateMessage(ctx context.Context, message models.Message) (models.Message, error)
	ListMessagesForChannelUser(ctx context.Context, channelID, userID uuid.UUID, limit int) ([]models.Message, error)
}

type PostgresServerRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresServerRepository(pool *pgxpool.Pool) *PostgresServerRepository {
	return &PostgresServerRepository{pool: pool}
}

func (r *PostgresServerRepository) CreateServer(ctx context.Context, server models.Server, ownerID uuid.UUID) (models.Server, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.Server{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	row := tx.QueryRow(ctx, `
		INSERT INTO servers (id, owner_id, name, description, icon_url)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, owner_id, name, description, icon_url, created_at, updated_at
	`, server.ID, ownerID, server.Name, server.Description, server.IconURL)

	created, err := scanServer(row)
	if err != nil {
		return models.Server{}, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO server_members (server_id, user_id, role)
		VALUES ($1, $2, $3)
	`, created.ID, ownerID, models.ServerRoleOwner)
	if err != nil {
		return models.Server{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Server{}, err
	}

	return created, nil
}

func (r *PostgresServerRepository) ListServersForUser(ctx context.Context, userID uuid.UUID) ([]models.Server, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT s.id, s.owner_id, s.name, s.description, s.icon_url, s.created_at, s.updated_at
		FROM servers s
		JOIN server_members sm ON sm.server_id = s.id
		WHERE sm.user_id = $1
		ORDER BY s.created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanServers(rows)
}

func (r *PostgresServerRepository) GetServerForUser(ctx context.Context, serverID, userID uuid.UUID) (models.Server, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT s.id, s.owner_id, s.name, s.description, s.icon_url, s.created_at, s.updated_at
		FROM servers s
		JOIN server_members sm ON sm.server_id = s.id
		WHERE s.id = $1 AND sm.user_id = $2
	`, serverID, userID)

	server, err := scanServer(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Server{}, ErrServerNotFound
		}
		return models.Server{}, err
	}
	return server, nil
}

func (r *PostgresServerRepository) CreateChannel(ctx context.Context, channel models.Channel) (models.Channel, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO channels (id, server_id, name, type, position)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, server_id, name, type, position, created_at, updated_at
	`, channel.ID, channel.ServerID, channel.Name, channel.Type, channel.Position)

	created, err := scanChannel(row)
	if err != nil {
		if isUniqueViolation(err) {
			return models.Channel{}, ErrChannelConflict
		}
		return models.Channel{}, err
	}
	return created, nil
}

func (r *PostgresServerRepository) ListChannelsForUser(ctx context.Context, serverID, userID uuid.UUID) ([]models.Channel, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.server_id, c.name, c.type, c.position, c.created_at, c.updated_at
		FROM channels c
		JOIN server_members sm ON sm.server_id = c.server_id
		WHERE c.server_id = $1 AND sm.user_id = $2
		ORDER BY c.position ASC, c.created_at ASC
	`, serverID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanChannels(rows)
}

func (r *PostgresServerRepository) GetChannelForUser(ctx context.Context, channelID, userID uuid.UUID) (models.Channel, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT c.id, c.server_id, c.name, c.type, c.position, c.created_at, c.updated_at
		FROM channels c
		JOIN server_members sm ON sm.server_id = c.server_id
		WHERE c.id = $1 AND sm.user_id = $2
	`, channelID, userID)

	channel, err := scanChannel(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Channel{}, ErrChannelNotFound
		}
		return models.Channel{}, err
	}
	return channel, nil
}

func (r *PostgresServerRepository) CreateMessage(ctx context.Context, message models.Message) (models.Message, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO messages (id, server_id, channel_id, author_id, content)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, server_id, channel_id, author_id, content, created_at, updated_at
	`, message.ID, message.ServerID, message.ChannelID, message.AuthorID, message.Content)

	return scanMessage(row)
}

func (r *PostgresServerRepository) ListMessagesForChannelUser(ctx context.Context, channelID, userID uuid.UUID, limit int) ([]models.Message, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT m.id, m.server_id, m.channel_id, m.author_id, m.content, m.created_at, m.updated_at
		FROM messages m
		JOIN channels c ON c.id = m.channel_id
		JOIN server_members sm ON sm.server_id = c.server_id
		WHERE m.channel_id = $1 AND sm.user_id = $2
		ORDER BY m.created_at ASC
		LIMIT $3
	`, channelID, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMessages(rows)
}

func scanServer(row pgx.Row) (models.Server, error) {
	var server models.Server
	err := row.Scan(
		&server.ID,
		&server.OwnerID,
		&server.Name,
		&server.Description,
		&server.IconURL,
		&server.CreatedAt,
		&server.UpdatedAt,
	)
	if err != nil {
		return models.Server{}, err
	}
	return server, nil
}

func scanServers(rows pgx.Rows) ([]models.Server, error) {
	servers := make([]models.Server, 0)
	for rows.Next() {
		server, err := scanServer(rows)
		if err != nil {
			return nil, err
		}
		servers = append(servers, server)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return servers, nil
}

func scanChannel(row pgx.Row) (models.Channel, error) {
	var channel models.Channel
	err := row.Scan(
		&channel.ID,
		&channel.ServerID,
		&channel.Name,
		&channel.Type,
		&channel.Position,
		&channel.CreatedAt,
		&channel.UpdatedAt,
	)
	if err != nil {
		return models.Channel{}, err
	}
	return channel, nil
}

func scanChannels(rows pgx.Rows) ([]models.Channel, error) {
	channels := make([]models.Channel, 0)
	for rows.Next() {
		channel, err := scanChannel(rows)
		if err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return channels, nil
}

func scanMessage(row pgx.Row) (models.Message, error) {
	var message models.Message
	err := row.Scan(
		&message.ID,
		&message.ServerID,
		&message.ChannelID,
		&message.AuthorID,
		&message.Content,
		&message.CreatedAt,
		&message.UpdatedAt,
	)
	if err != nil {
		return models.Message{}, err
	}
	return message, nil
}

func scanMessages(rows pgx.Rows) ([]models.Message, error) {
	messages := make([]models.Message, 0)
	for rows.Next() {
		message, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}
