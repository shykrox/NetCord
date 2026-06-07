package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	ServerRoleOwner  = "owner"
	ServerRoleMember = "member"
	ChannelTypeText  = "text"
)

type Server struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	Name        string
	Description *string
	IconURL     *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PublicServer struct {
	ID          uuid.UUID `json:"id"`
	OwnerID     uuid.UUID `json:"owner_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	IconURL     *string   `json:"icon_url"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s Server) Public() PublicServer {
	return PublicServer{
		ID:          s.ID,
		OwnerID:     s.OwnerID,
		Name:        s.Name,
		Description: s.Description,
		IconURL:     s.IconURL,
		CreatedAt:   s.CreatedAt,
	}
}

type Channel struct {
	ID        uuid.UUID
	ServerID  uuid.UUID
	Name      string
	Type      string
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PublicChannel struct {
	ID        uuid.UUID `json:"id"`
	ServerID  uuid.UUID `json:"server_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
}

func (c Channel) Public() PublicChannel {
	channelType := c.Type
	if channelType == "" {
		channelType = ChannelTypeText
	}

	return PublicChannel{
		ID:        c.ID,
		ServerID:  c.ServerID,
		Name:      c.Name,
		Type:      channelType,
		Position:  c.Position,
		CreatedAt: c.CreatedAt,
	}
}

type Message struct {
	ID        uuid.UUID
	ServerID  uuid.UUID
	ChannelID uuid.UUID
	AuthorID  uuid.UUID
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PublicMessage struct {
	ID        uuid.UUID `json:"id"`
	ServerID  uuid.UUID `json:"server_id"`
	ChannelID uuid.UUID `json:"channel_id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func (m Message) Public() PublicMessage {
	return PublicMessage{
		ID:        m.ID,
		ServerID:  m.ServerID,
		ChannelID: m.ChannelID,
		AuthorID:  m.AuthorID,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
}
