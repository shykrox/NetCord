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

type SocialRepository interface {
	CreateFriendRequest(ctx context.Context, request models.FriendRequest) (models.FriendRequest, error)
	ListFriendRequestsForUser(ctx context.Context, userID uuid.UUID) ([]models.FriendRequest, error)
	AcceptFriendRequest(ctx context.Context, requestID, recipientID uuid.UUID) (models.FriendRequest, error)
	DeclineFriendRequest(ctx context.Context, requestID, recipientID uuid.UUID) (models.FriendRequest, error)
	ListFriends(ctx context.Context, userID uuid.UUID) ([]models.Friend, error)
	RemoveFriend(ctx context.Context, userID, friendID uuid.UUID) error
	UsersBlocked(ctx context.Context, userA, userB uuid.UUID) (bool, error)
	CreateDMConversation(ctx context.Context, conversation models.DMConversation, memberIDs []uuid.UUID) (models.DMConversation, error)
	ListDMConversationsForUser(ctx context.Context, userID uuid.UUID) ([]models.DMConversation, error)
	GetDMConversationForUser(ctx context.Context, conversationID, userID uuid.UUID) (models.DMConversation, error)
	ListDMMemberIDs(ctx context.Context, conversationID uuid.UUID) ([]uuid.UUID, error)
	CreateDMMessage(ctx context.Context, message models.DMMessage) (models.DMMessage, error)
	ListDMMessagesForUser(ctx context.Context, conversationID, userID uuid.UUID, limit int) ([]models.DMMessage, error)
}

type PostgresSocialRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresSocialRepository(pool *pgxpool.Pool) *PostgresSocialRepository {
	return &PostgresSocialRepository{pool: pool}
}

func (r *PostgresSocialRepository) CreateFriendRequest(ctx context.Context, request models.FriendRequest) (models.FriendRequest, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO friend_requests (id, requester_id, recipient_id, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, requester_id, recipient_id, status, created_at, responded_at
	`, request.ID, request.RequesterID, request.RecipientID, request.Status)

	created, err := scanFriendRequest(row)
	if err != nil {
		if isUniqueViolation(err) {
			return models.FriendRequest{}, ErrFriendRequestConflict
		}
		return models.FriendRequest{}, err
	}
	return created, nil
}

func (r *PostgresSocialRepository) ListFriendRequestsForUser(ctx context.Context, userID uuid.UUID) ([]models.FriendRequest, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, requester_id, recipient_id, status, created_at, responded_at
		FROM friend_requests
		WHERE (requester_id = $1 OR recipient_id = $1)
			AND status = 'pending'
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFriendRequests(rows)
}

func (r *PostgresSocialRepository) AcceptFriendRequest(ctx context.Context, requestID, recipientID uuid.UUID) (models.FriendRequest, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.FriendRequest{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	request, err := scanFriendRequest(tx.QueryRow(ctx, `
		UPDATE friend_requests
		SET status = 'accepted', responded_at = now()
		WHERE id = $1
			AND recipient_id = $2
			AND status = 'pending'
		RETURNING id, requester_id, recipient_id, status, created_at, responded_at
	`, requestID, recipientID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.FriendRequest{}, ErrFriendRequestNotFound
		}
		return models.FriendRequest{}, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO friends (user_id, friend_id)
		VALUES ($1, $2), ($2, $1)
		ON CONFLICT DO NOTHING
	`, request.RequesterID, request.RecipientID)
	if err != nil {
		return models.FriendRequest{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.FriendRequest{}, err
	}
	return request, nil
}

func (r *PostgresSocialRepository) DeclineFriendRequest(ctx context.Context, requestID, recipientID uuid.UUID) (models.FriendRequest, error) {
	request, err := scanFriendRequest(r.pool.QueryRow(ctx, `
		UPDATE friend_requests
		SET status = 'declined', responded_at = now()
		WHERE id = $1
			AND recipient_id = $2
			AND status = 'pending'
		RETURNING id, requester_id, recipient_id, status, created_at, responded_at
	`, requestID, recipientID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.FriendRequest{}, ErrFriendRequestNotFound
		}
		return models.FriendRequest{}, err
	}
	return request, nil
}

func (r *PostgresSocialRepository) ListFriends(ctx context.Context, userID uuid.UUID) ([]models.Friend, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT u.id, u.username, f.created_at
		FROM friends f
		JOIN users u ON u.id = f.friend_id
		WHERE f.user_id = $1
		ORDER BY u.username ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	friends := make([]models.Friend, 0)
	for rows.Next() {
		var friend models.Friend
		if err := rows.Scan(&friend.UserID, &friend.Username, &friend.CreatedAt); err != nil {
			return nil, err
		}
		friends = append(friends, friend)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return friends, nil
}

func (r *PostgresSocialRepository) RemoveFriend(ctx context.Context, userID, friendID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM friends
		WHERE (user_id = $1 AND friend_id = $2)
			OR (user_id = $2 AND friend_id = $1)
	`, userID, friendID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrFriendNotFound
	}
	return nil
}

func (r *PostgresSocialRepository) UsersBlocked(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM blocked_users
			WHERE (blocker_id = $1 AND blocked_id = $2)
				OR (blocker_id = $2 AND blocked_id = $1)
		)
	`, userA, userB).Scan(&exists)
	return exists, err
}

func (r *PostgresSocialRepository) CreateDMConversation(ctx context.Context, conversation models.DMConversation, memberIDs []uuid.UUID) (models.DMConversation, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.DMConversation{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	created, err := scanDMConversation(tx.QueryRow(ctx, `
		INSERT INTO dm_conversations (id, type, name, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, type, name, created_by, created_at, updated_at
	`, conversation.ID, conversation.Type, conversation.Name, conversation.CreatedBy))
	if err != nil {
		return models.DMConversation{}, err
	}

	for _, memberID := range memberIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO dm_members (conversation_id, user_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, created.ID, memberID); err != nil {
			return models.DMConversation{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return models.DMConversation{}, err
	}

	return r.GetDMConversationForUser(ctx, created.ID, conversation.CreatedBy)
}

func (r *PostgresSocialRepository) ListDMConversationsForUser(ctx context.Context, userID uuid.UUID) ([]models.DMConversation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.type, c.name, c.created_by, c.created_at, c.updated_at
		FROM dm_conversations c
		JOIN dm_members dm ON dm.conversation_id = c.id
		WHERE dm.user_id = $1
		ORDER BY c.updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversations := make([]models.DMConversation, 0)
	for rows.Next() {
		conversation, err := scanDMConversation(rows)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, conversation)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := r.hydrateDMMembers(ctx, conversations); err != nil {
		return nil, err
	}
	return conversations, nil
}

func (r *PostgresSocialRepository) GetDMConversationForUser(ctx context.Context, conversationID, userID uuid.UUID) (models.DMConversation, error) {
	conversation, err := scanDMConversation(r.pool.QueryRow(ctx, `
		SELECT c.id, c.type, c.name, c.created_by, c.created_at, c.updated_at
		FROM dm_conversations c
		JOIN dm_members dm ON dm.conversation_id = c.id
		WHERE c.id = $1 AND dm.user_id = $2
	`, conversationID, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.DMConversation{}, ErrDMConversationNotFound
		}
		return models.DMConversation{}, err
	}

	conversations := []models.DMConversation{conversation}
	if err := r.hydrateDMMembers(ctx, conversations); err != nil {
		return models.DMConversation{}, err
	}
	return conversations[0], nil
}

func (r *PostgresSocialRepository) ListDMMemberIDs(ctx context.Context, conversationID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT user_id
		FROM dm_members
		WHERE conversation_id = $1
	`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *PostgresSocialRepository) CreateDMMessage(ctx context.Context, message models.DMMessage) (models.DMMessage, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.DMMessage{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var isMember bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM dm_members
			WHERE conversation_id = $1 AND user_id = $2
		)
	`, message.ConversationID, message.AuthorID).Scan(&isMember); err != nil {
		return models.DMMessage{}, err
	}
	if !isMember {
		return models.DMMessage{}, ErrDMConversationNotFound
	}

	created, err := scanDMMessage(tx.QueryRow(ctx, `
		INSERT INTO dm_messages (id, conversation_id, author_id, content)
		VALUES ($1, $2, $3, $4)
		RETURNING id, conversation_id, author_id, content, created_at, updated_at, edited_at
	`, message.ID, message.ConversationID, message.AuthorID, message.Content))
	if err != nil {
		return models.DMMessage{}, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE dm_conversations
		SET updated_at = now()
		WHERE id = $1
	`, message.ConversationID)
	if err != nil {
		return models.DMMessage{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.DMMessage{}, err
	}
	return created, nil
}

func (r *PostgresSocialRepository) ListDMMessagesForUser(ctx context.Context, conversationID, userID uuid.UUID, limit int) ([]models.DMMessage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT *
		FROM (
			SELECT m.id, m.conversation_id, m.author_id, m.content, m.created_at, m.updated_at, m.edited_at
			FROM dm_messages m
			JOIN dm_members dm ON dm.conversation_id = m.conversation_id
			WHERE m.conversation_id = $1 AND dm.user_id = $2
			ORDER BY m.created_at DESC, m.id DESC
			LIMIT $3
		) page
		ORDER BY created_at ASC, id ASC
	`, conversationID, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]models.DMMessage, 0)
	for rows.Next() {
		message, err := scanDMMessage(rows)
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

func scanFriendRequest(row pgx.Row) (models.FriendRequest, error) {
	var request models.FriendRequest
	var respondedAt pgtype.Timestamptz
	err := row.Scan(&request.ID, &request.RequesterID, &request.RecipientID, &request.Status, &request.CreatedAt, &respondedAt)
	if err != nil {
		return models.FriendRequest{}, err
	}
	request.RespondedAt = nullableTime(respondedAt)
	return request, nil
}

func scanFriendRequests(rows pgx.Rows) ([]models.FriendRequest, error) {
	requests := make([]models.FriendRequest, 0)
	for rows.Next() {
		request, err := scanFriendRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return requests, nil
}

func scanDMConversation(row pgx.Row) (models.DMConversation, error) {
	var conversation models.DMConversation
	err := row.Scan(&conversation.ID, &conversation.Type, &conversation.Name, &conversation.CreatedBy, &conversation.CreatedAt, &conversation.UpdatedAt)
	if err != nil {
		return models.DMConversation{}, err
	}
	return conversation, nil
}

func scanDMMessage(row pgx.Row) (models.DMMessage, error) {
	var message models.DMMessage
	var editedAt pgtype.Timestamptz
	err := row.Scan(&message.ID, &message.ConversationID, &message.AuthorID, &message.Content, &message.CreatedAt, &message.UpdatedAt, &editedAt)
	if err != nil {
		return models.DMMessage{}, err
	}
	message.EditedAt = nullableTime(editedAt)
	return message, nil
}

func (r *PostgresSocialRepository) hydrateDMMembers(ctx context.Context, conversations []models.DMConversation) error {
	if len(conversations) == 0 {
		return nil
	}

	conversationIDs := make([]uuid.UUID, 0, len(conversations))
	conversationIndex := make(map[uuid.UUID]int, len(conversations))
	for i, conversation := range conversations {
		conversationIDs = append(conversationIDs, conversation.ID)
		conversationIndex[conversation.ID] = i
	}

	rows, err := r.pool.Query(ctx, `
		SELECT dm.conversation_id, u.id, u.username, u.email, u.password_hash, u.display_name, u.avatar_url,
			COALESCE(up.status, 'offline') AS status, u.is_bot, u.created_at, u.updated_at
		FROM dm_members dm
		JOIN users u ON u.id = dm.user_id
		LEFT JOIN user_presence up ON up.user_id = u.id
		WHERE dm.conversation_id = ANY($1)
		ORDER BY u.username ASC
	`, conversationIDs)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var conversationID uuid.UUID
		var user models.User
		if err := rows.Scan(&conversationID, &user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.DisplayName, &user.AvatarURL, &user.Status, &user.IsBot, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return err
		}
		index, ok := conversationIndex[conversationID]
		if !ok {
			continue
		}
		conversations[index].Members = append(conversations[index].Members, user.Public())
	}
	return rows.Err()
}
