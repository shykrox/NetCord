package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	PermissionViewServer int64 = 1 << iota
	PermissionManageServer
	PermissionManageRoles
	PermissionManageChannels
	PermissionViewChannel
	PermissionSendMessages
	PermissionManageMessages
	PermissionUploadFiles
	PermissionAddReactions
	PermissionConnectVoice
	PermissionSpeak
	PermissionStream
	PermissionMuteMembers
	PermissionDeafenMembers
	PermissionMoveMembers
	PermissionKickMembers
	PermissionBanMembers
	PermissionCreateInvite
	PermissionViewAuditLog
	PermissionAdministrator
)

const (
	AllPermissions           int64 = (1 << 20) - 1
	DefaultMemberPermissions int64 = PermissionViewServer |
		PermissionViewChannel |
		PermissionSendMessages |
		PermissionUploadFiles |
		PermissionConnectVoice |
		PermissionSpeak |
		PermissionCreateInvite
)

type Role struct {
	ID          uuid.UUID
	ServerID    uuid.UUID
	Name        string
	Color       *string
	Position    int
	Permissions int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PublicRole struct {
	ID          uuid.UUID `json:"id"`
	ServerID    uuid.UUID `json:"server_id"`
	Name        string    `json:"name"`
	Color       *string   `json:"color"`
	Position    int       `json:"position"`
	Permissions int64     `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (r Role) Public() PublicRole {
	return PublicRole{
		ID:          r.ID,
		ServerID:    r.ServerID,
		Name:        r.Name,
		Color:       r.Color,
		Position:    r.Position,
		Permissions: r.Permissions,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

type ServerInvite struct {
	ID        uuid.UUID
	ServerID  uuid.UUID
	CreatedBy uuid.UUID
	Code      string
	MaxUses   *int
	Uses      int
	ExpiresAt *time.Time
	CreatedAt time.Time
	Server    *PublicServer
}

type PublicServerInvite struct {
	ID        uuid.UUID     `json:"id"`
	ServerID  uuid.UUID     `json:"server_id"`
	CreatedBy uuid.UUID     `json:"created_by"`
	Code      string        `json:"code"`
	MaxUses   *int          `json:"max_uses"`
	Uses      int           `json:"uses"`
	ExpiresAt *time.Time    `json:"expires_at"`
	CreatedAt time.Time     `json:"created_at"`
	Server    *PublicServer `json:"server,omitempty"`
}

func (i ServerInvite) Public() PublicServerInvite {
	return PublicServerInvite{
		ID:        i.ID,
		ServerID:  i.ServerID,
		CreatedBy: i.CreatedBy,
		Code:      i.Code,
		MaxUses:   i.MaxUses,
		Uses:      i.Uses,
		ExpiresAt: i.ExpiresAt,
		CreatedAt: i.CreatedAt,
		Server:    i.Server,
	}
}
