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

const defaultChannelPosition = 0

type ServerRepository interface {
	CreateServer(ctx context.Context, server models.Server, ownerID uuid.UUID) (models.Server, error)
	ListServersForUser(ctx context.Context, userID uuid.UUID) ([]models.Server, error)
	GetServerForUser(ctx context.Context, serverID, userID uuid.UUID) (models.Server, error)
	CreateChannel(ctx context.Context, channel models.Channel) (models.Channel, error)
	ListChannelsForUser(ctx context.Context, serverID, userID uuid.UUID) ([]models.Channel, error)
	GetChannelForUser(ctx context.Context, channelID, userID uuid.UUID) (models.Channel, error)
	CreateMessage(ctx context.Context, message models.Message, attachmentIDs []uuid.UUID) (models.Message, error)
	ListMessagesForChannelUser(ctx context.Context, channelID, userID uuid.UUID, limit int) ([]models.Message, error)
}

type FileRepository interface {
	CreateAttachment(ctx context.Context, attachment models.MessageAttachment) (models.MessageAttachment, error)
	GetAttachmentForUser(ctx context.Context, attachmentID, userID uuid.UUID) (models.MessageAttachment, error)
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

func (r *PostgresServerRepository) CreateMessage(ctx context.Context, message models.Message, attachmentIDs []uuid.UUID) (models.Message, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.Message{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	row := tx.QueryRow(ctx, `
		INSERT INTO messages (id, server_id, channel_id, author_id, content)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, server_id, channel_id, author_id, content, created_at, updated_at
	`, message.ID, message.ServerID, message.ChannelID, message.AuthorID, message.Content)

	created, err := scanMessage(row)
	if err != nil {
		return models.Message{}, err
	}

	if len(attachmentIDs) > 0 {
		rows, err := tx.Query(ctx, `
			UPDATE message_attachments
			SET message_id = $1, server_id = $2, channel_id = $3
			WHERE id = ANY($4)
				AND uploader_id = $5
				AND message_id IS NULL
			RETURNING id, uploader_id, message_id, server_id, channel_id, bucket, object_key,
				original_filename, content_type, size_bytes, created_at
		`, created.ID, created.ServerID, created.ChannelID, attachmentIDs, created.AuthorID)
		if err != nil {
			return models.Message{}, err
		}

		attachments, err := scanAttachments(rows)
		if err != nil {
			return models.Message{}, err
		}
		if len(attachments) != len(attachmentIDs) {
			return models.Message{}, ErrAttachmentNotFound
		}
		created.Attachments = attachments
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Message{}, err
	}

	return created, nil
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

	messages, err := scanMessages(rows)
	if err != nil {
		return nil, err
	}

	if err := r.hydrateMessageAttachments(ctx, messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *PostgresServerRepository) CreateAttachment(ctx context.Context, attachment models.MessageAttachment) (models.MessageAttachment, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO message_attachments (
			id, uploader_id, bucket, object_key, original_filename, content_type, size_bytes
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, uploader_id, message_id, server_id, channel_id, bucket, object_key,
			original_filename, content_type, size_bytes, created_at
	`, attachment.ID, attachment.UploaderID, attachment.Bucket, attachment.ObjectKey,
		attachment.OriginalFilename, attachment.ContentType, attachment.SizeBytes)

	created, err := scanAttachment(row)
	if err != nil {
		return models.MessageAttachment{}, err
	}
	return created, nil
}

func (r *PostgresServerRepository) GetAttachmentForUser(ctx context.Context, attachmentID, userID uuid.UUID) (models.MessageAttachment, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT ma.id, ma.uploader_id, ma.message_id, ma.server_id, ma.channel_id, ma.bucket, ma.object_key,
			ma.original_filename, ma.content_type, ma.size_bytes, ma.created_at
		FROM message_attachments ma
		LEFT JOIN server_members sm ON sm.server_id = ma.server_id AND sm.user_id = $2
		WHERE ma.id = $1 AND (ma.uploader_id = $2 OR sm.user_id = $2)
	`, attachmentID, userID)

	attachment, err := scanAttachment(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.MessageAttachment{}, ErrAttachmentNotFound
		}
		return models.MessageAttachment{}, err
	}
	return attachment, nil
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

func (r *PostgresServerRepository) hydrateMessageAttachments(ctx context.Context, messages []models.Message) error {
	if len(messages) == 0 {
		return nil
	}

	messageIDs := make([]uuid.UUID, 0, len(messages))
	messageIndex := make(map[uuid.UUID]int, len(messages))
	for i, message := range messages {
		messageIDs = append(messageIDs, message.ID)
		messageIndex[message.ID] = i
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, uploader_id, message_id, server_id, channel_id, bucket, object_key,
			original_filename, content_type, size_bytes, created_at
		FROM message_attachments
		WHERE message_id = ANY($1)
		ORDER BY created_at ASC
	`, messageIDs)
	if err != nil {
		return err
	}

	attachments, err := scanAttachments(rows)
	if err != nil {
		return err
	}

	for _, attachment := range attachments {
		if attachment.MessageID == nil {
			continue
		}
		index, ok := messageIndex[*attachment.MessageID]
		if !ok {
			continue
		}
		messages[index].Attachments = append(messages[index].Attachments, attachment)
	}

	return nil
}

func scanAttachment(row pgx.Row) (models.MessageAttachment, error) {
	var attachment models.MessageAttachment
	var messageID pgtype.UUID
	var serverID pgtype.UUID
	var channelID pgtype.UUID

	err := row.Scan(
		&attachment.ID,
		&attachment.UploaderID,
		&messageID,
		&serverID,
		&channelID,
		&attachment.Bucket,
		&attachment.ObjectKey,
		&attachment.OriginalFilename,
		&attachment.ContentType,
		&attachment.SizeBytes,
		&attachment.CreatedAt,
	)
	if err != nil {
		return models.MessageAttachment{}, err
	}

	attachment.MessageID = nullableUUID(messageID)
	attachment.ServerID = nullableUUID(serverID)
	attachment.ChannelID = nullableUUID(channelID)

	return attachment, nil
}

func scanAttachments(rows pgx.Rows) ([]models.MessageAttachment, error) {
	defer rows.Close()

	attachments := make([]models.MessageAttachment, 0)
	for rows.Next() {
		attachment, err := scanAttachment(rows)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, attachment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return attachments, nil
}

func nullableUUID(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}

	id := uuid.UUID(value.Bytes)
	return &id
}
