package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Username     string
	Email        string
	PasswordHash string
	DisplayName  *string
	AvatarURL    *string
	Status       string
	IsBot        bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PublicUser struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"`
	Status      string    `json:"status"`
	IsBot       bool      `json:"is_bot"`
	CreatedAt   time.Time `json:"created_at"`
}

const (
	PresenceOnline  = "online"
	PresenceIdle    = "idle"
	PresenceDND     = "dnd"
	PresenceOffline = "offline"
)

type UserPresence struct {
	UserID       uuid.UUID
	Status       string
	CustomStatus *string
	LastSeenAt   time.Time
	UpdatedAt    time.Time
}

type PublicPresence struct {
	UserID       uuid.UUID `json:"user_id"`
	Status       string    `json:"status"`
	CustomStatus *string   `json:"custom_status"`
	LastSeenAt   time.Time `json:"last_seen_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (p UserPresence) Public() PublicPresence {
	status := p.Status
	if status == "" {
		status = PresenceOffline
	}

	return PublicPresence{
		UserID:       p.UserID,
		Status:       status,
		CustomStatus: p.CustomStatus,
		LastSeenAt:   p.LastSeenAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

func (u User) Public() PublicUser {
	displayName := u.Username
	if u.DisplayName != nil && *u.DisplayName != "" {
		displayName = *u.DisplayName
	}

	status := u.Status
	if status == "" {
		status = "offline"
	}

	return PublicUser{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		DisplayName: displayName,
		AvatarURL:   u.AvatarURL,
		Status:      status,
		IsBot:       u.IsBot,
		CreatedAt:   u.CreatedAt,
	}
}
