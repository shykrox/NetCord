package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"netcord/backend/api/internal/models"
	"netcord/backend/api/internal/repository"

	"github.com/google/uuid"
)

type CreateRoleInput struct {
	Name        string  `json:"name"`
	Color       *string `json:"color"`
	Position    int     `json:"position"`
	Permissions int64   `json:"permissions"`
}

type UpdateRoleInput struct {
	Name        *string `json:"name"`
	Color       *string `json:"color"`
	Position    *int    `json:"position"`
	Permissions *int64  `json:"permissions"`
}

type CreateInviteInput struct {
	MaxUses   *int       `json:"max_uses"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type RolesResponse struct {
	Roles []models.PublicRole `json:"roles"`
}

type PermissionService struct {
	permissions repository.PermissionRepository
}

func NewPermissionService(permissions repository.PermissionRepository) *PermissionService {
	return &PermissionService{permissions: permissions}
}

func (s *PermissionService) RequirePermission(ctx context.Context, serverID, userID uuid.UUID, permission int64) error {
	permissions, err := s.permissions.UserPermissions(ctx, serverID, userID)
	if err != nil {
		return mapPermissionRepositoryError(err)
	}
	if permissions&models.PermissionAdministrator != 0 || permissions&permission != 0 {
		return nil
	}
	return ErrForbidden
}

func (s *PermissionService) CreateRole(ctx context.Context, userID, serverID uuid.UUID, input CreateRoleInput) (models.PublicRole, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Color != nil {
		color := strings.TrimSpace(*input.Color)
		input.Color = &color
	}
	if fields := validateCreateRoleInput(input); len(fields) > 0 {
		return models.PublicRole{}, &ValidationError{Fields: fields}
	}
	if err := s.RequirePermission(ctx, serverID, userID, models.PermissionManageRoles); err != nil {
		return models.PublicRole{}, err
	}

	role := models.Role{
		ID:          uuid.New(),
		ServerID:    serverID,
		Name:        input.Name,
		Color:       input.Color,
		Position:    input.Position,
		Permissions: input.Permissions,
	}
	created, err := s.permissions.CreateRole(ctx, role)
	if err != nil {
		return models.PublicRole{}, mapPermissionRepositoryError(err)
	}
	_ = s.permissions.CreateAuditLog(ctx, serverID, userID, "role.create", "role", created.ID, "{}")
	return created.Public(), nil
}

func (s *PermissionService) ListRoles(ctx context.Context, userID, serverID uuid.UUID) (RolesResponse, error) {
	if err := s.RequirePermission(ctx, serverID, userID, models.PermissionViewServer); err != nil {
		return RolesResponse{}, err
	}
	roles, err := s.permissions.ListRoles(ctx, serverID)
	if err != nil {
		return RolesResponse{}, mapPermissionRepositoryError(err)
	}
	return RolesResponse{Roles: publicRoles(roles)}, nil
}

func (s *PermissionService) UpdateRole(ctx context.Context, userID, roleID uuid.UUID, input UpdateRoleInput) (models.PublicRole, error) {
	role, err := s.permissions.GetRole(ctx, roleID)
	if err != nil {
		return models.PublicRole{}, mapPermissionRepositoryError(err)
	}
	if err := s.RequirePermission(ctx, role.ServerID, userID, models.PermissionManageRoles); err != nil {
		return models.PublicRole{}, err
	}

	if input.Name != nil {
		role.Name = strings.TrimSpace(*input.Name)
	}
	if input.Color != nil {
		color := strings.TrimSpace(*input.Color)
		role.Color = &color
	}
	if input.Position != nil {
		role.Position = *input.Position
	}
	if input.Permissions != nil {
		role.Permissions = *input.Permissions
	}
	if fields := validateRole(role); len(fields) > 0 {
		return models.PublicRole{}, &ValidationError{Fields: fields}
	}

	updated, err := s.permissions.UpdateRole(ctx, role)
	if err != nil {
		return models.PublicRole{}, mapPermissionRepositoryError(err)
	}
	_ = s.permissions.CreateAuditLog(ctx, role.ServerID, userID, "role.update", "role", role.ID, "{}")
	return updated.Public(), nil
}

func (s *PermissionService) DeleteRole(ctx context.Context, userID, roleID uuid.UUID) error {
	role, err := s.permissions.GetRole(ctx, roleID)
	if err != nil {
		return mapPermissionRepositoryError(err)
	}
	if err := s.RequirePermission(ctx, role.ServerID, userID, models.PermissionManageRoles); err != nil {
		return err
	}
	if err := s.permissions.DeleteRole(ctx, roleID); err != nil {
		return mapPermissionRepositoryError(err)
	}
	_ = s.permissions.CreateAuditLog(ctx, role.ServerID, userID, "role.delete", "role", role.ID, "{}")
	return nil
}

func (s *PermissionService) AssignRole(ctx context.Context, actorID, serverID, targetUserID, roleID uuid.UUID) error {
	if err := s.RequirePermission(ctx, serverID, actorID, models.PermissionManageRoles); err != nil {
		return err
	}
	if err := s.permissions.AssignRole(ctx, serverID, targetUserID, roleID); err != nil {
		return mapPermissionRepositoryError(err)
	}
	_ = s.permissions.CreateAuditLog(ctx, serverID, actorID, "member_role.assign", "role", roleID, "{}")
	return nil
}

func (s *PermissionService) RemoveRole(ctx context.Context, actorID, serverID, targetUserID, roleID uuid.UUID) error {
	if err := s.RequirePermission(ctx, serverID, actorID, models.PermissionManageRoles); err != nil {
		return err
	}
	if err := s.permissions.RemoveRole(ctx, serverID, targetUserID, roleID); err != nil {
		return mapPermissionRepositoryError(err)
	}
	_ = s.permissions.CreateAuditLog(ctx, serverID, actorID, "member_role.remove", "role", roleID, "{}")
	return nil
}

func (s *PermissionService) CreateInvite(ctx context.Context, userID, serverID uuid.UUID, input CreateInviteInput) (models.PublicServerInvite, error) {
	if fields := validateCreateInviteInput(input); len(fields) > 0 {
		return models.PublicServerInvite{}, &ValidationError{Fields: fields}
	}
	if err := s.RequirePermission(ctx, serverID, userID, models.PermissionCreateInvite); err != nil {
		return models.PublicServerInvite{}, err
	}

	invite := models.ServerInvite{
		ID:        uuid.New(),
		ServerID:  serverID,
		CreatedBy: userID,
		Code:      randomInviteCode(),
		MaxUses:   input.MaxUses,
		ExpiresAt: input.ExpiresAt,
	}
	created, err := s.permissions.CreateInvite(ctx, invite)
	if err != nil {
		return models.PublicServerInvite{}, mapPermissionRepositoryError(err)
	}
	_ = s.permissions.CreateAuditLog(ctx, serverID, userID, "invite.create", "invite", created.ID, "{}")
	return created.Public(), nil
}

func (s *PermissionService) GetInvite(ctx context.Context, code string) (models.PublicServerInvite, error) {
	invite, err := s.permissions.GetInvite(ctx, strings.TrimSpace(code))
	if err != nil {
		return models.PublicServerInvite{}, mapPermissionRepositoryError(err)
	}
	return invite.Public(), nil
}

func (s *PermissionService) JoinInvite(ctx context.Context, userID uuid.UUID, code string) (models.PublicServerInvite, error) {
	invite, err := s.permissions.JoinInvite(ctx, strings.TrimSpace(code), userID)
	if err != nil {
		return models.PublicServerInvite{}, mapPermissionRepositoryError(err)
	}
	_ = s.permissions.CreateAuditLog(ctx, invite.ServerID, userID, "invite.join", "invite", invite.ID, "{}")
	return invite.Public(), nil
}

func mapPermissionRepositoryError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, repository.ErrServerNotFound),
		errors.Is(err, repository.ErrRoleNotFound),
		errors.Is(err, repository.ErrInviteNotFound):
		return ErrNotFound
	case errors.Is(err, repository.ErrRoleConflict):
		return &ValidationError{Fields: map[string]string{"name": "role name already exists"}}
	case errors.Is(err, repository.ErrInviteUnavailable):
		return ErrForbidden
	default:
		return err
	}
}

func publicRoles(roles []models.Role) []models.PublicRole {
	public := make([]models.PublicRole, 0, len(roles))
	for _, role := range roles {
		public = append(public, role.Public())
	}
	return public
}

func randomInviteCode() string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return uuid.NewString()
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}
