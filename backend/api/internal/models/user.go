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
	CreatedAt   time.Time `json:"created_at"`
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
		CreatedAt:   u.CreatedAt,
	}
}
