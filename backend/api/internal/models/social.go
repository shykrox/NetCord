package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	FriendRequestPending  = "pending"
	FriendRequestAccepted = "accepted"
	FriendRequestDeclined = "declined"
	DMTypeDirect          = "dm"
	DMTypeGroup           = "group"
)

type FriendRequest struct {
	ID          uuid.UUID
	RequesterID uuid.UUID
	RecipientID uuid.UUID
	Status      string
	CreatedAt   time.Time
	RespondedAt *time.Time
}

type PublicFriendRequest struct {
	ID          uuid.UUID  `json:"id"`
	RequesterID uuid.UUID  `json:"requester_id"`
	RecipientID uuid.UUID  `json:"recipient_id"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	RespondedAt *time.Time `json:"responded_at"`
}

func (r FriendRequest) Public() PublicFriendRequest {
	return PublicFriendRequest{
		ID:          r.ID,
		RequesterID: r.RequesterID,
		RecipientID: r.RecipientID,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt,
		RespondedAt: r.RespondedAt,
	}
}

type Friend struct {
	UserID    uuid.UUID
	Username  string
	CreatedAt time.Time
}

type PublicFriend struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

func (f Friend) Public() PublicFriend {
	return PublicFriend{
		UserID:    f.UserID,
		Username:  f.Username,
		CreatedAt: f.CreatedAt,
	}
}

type DMConversation struct {
	ID        uuid.UUID
	Type      string
	Name      *string
	CreatedBy uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Members   []PublicUser
}

type PublicDMConversation struct {
	ID        uuid.UUID    `json:"id"`
	Type      string       `json:"type"`
	Name      *string      `json:"name"`
	CreatedBy uuid.UUID    `json:"created_by"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	Members   []PublicUser `json:"members"`
}

func (c DMConversation) Public() PublicDMConversation {
	return PublicDMConversation{
		ID:        c.ID,
		Type:      c.Type,
		Name:      c.Name,
		CreatedBy: c.CreatedBy,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		Members:   c.Members,
	}
}

type DMMessage struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	AuthorID       uuid.UUID
	Content        string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	EditedAt       *time.Time
}

type PublicDMMessage struct {
	ID             uuid.UUID  `json:"id"`
	ConversationID uuid.UUID  `json:"conversation_id"`
	AuthorID       uuid.UUID  `json:"author_id"`
	Content        string     `json:"content"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	EditedAt       *time.Time `json:"edited_at"`
}

func (m DMMessage) Public() PublicDMMessage {
	return PublicDMMessage{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		AuthorID:       m.AuthorID,
		Content:        m.Content,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		EditedAt:       m.EditedAt,
	}
}
