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
	ID          uuid.UUID
	ServerID    uuid.UUID
	ChannelID   uuid.UUID
	AuthorID    uuid.UUID
	Content     string
	Attachments []MessageAttachment
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PublicMessage struct {
	ID          uuid.UUID          `json:"id"`
	ServerID    uuid.UUID          `json:"server_id"`
	ChannelID   uuid.UUID          `json:"channel_id"`
	AuthorID    uuid.UUID          `json:"author_id"`
	Content     string             `json:"content"`
	Attachments []PublicAttachment `json:"attachments"`
	CreatedAt   time.Time          `json:"created_at"`
}

func (m Message) Public() PublicMessage {
	return PublicMessage{
		ID:          m.ID,
		ServerID:    m.ServerID,
		ChannelID:   m.ChannelID,
		AuthorID:    m.AuthorID,
		Content:     m.Content,
		Attachments: publicAttachments(m.Attachments),
		CreatedAt:   m.CreatedAt,
	}
}

type MessageAttachment struct {
	ID               uuid.UUID
	UploaderID       uuid.UUID
	MessageID        *uuid.UUID
	ServerID         *uuid.UUID
	ChannelID        *uuid.UUID
	Bucket           string
	ObjectKey        string
	OriginalFilename string
	ContentType      string
	SizeBytes        int64
	CreatedAt        time.Time
}

type PublicAttachment struct {
	ID               uuid.UUID `json:"id"`
	OriginalFilename string    `json:"original_filename"`
	ContentType      string    `json:"content_type"`
	SizeBytes        int64     `json:"size_bytes"`
	DownloadURL      string    `json:"download_url"`
	CreatedAt        time.Time `json:"created_at"`
}

func (a MessageAttachment) Public() PublicAttachment {
	return PublicAttachment{
		ID:               a.ID,
		OriginalFilename: a.OriginalFilename,
		ContentType:      a.ContentType,
		SizeBytes:        a.SizeBytes,
		DownloadURL:      "/files/" + a.ID.String(),
		CreatedAt:        a.CreatedAt,
	}
}

func publicAttachments(attachments []MessageAttachment) []PublicAttachment {
	public := make([]PublicAttachment, 0, len(attachments))
	for _, attachment := range attachments {
		public = append(public, attachment.Public())
	}
	return public
}
