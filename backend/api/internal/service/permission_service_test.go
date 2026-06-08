package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"netcord/backend/api/internal/models"
	"netcord/backend/api/internal/repository"

	"github.com/google/uuid"
)

func TestCreateRoleRequiresManageRoles(t *testing.T) {
	repo := newFakePermissionRepository()
	service := NewPermissionService(repo)
	serverID := uuid.New()
	userID := uuid.New()
	repo.permissions[[2]uuid.UUID{serverID, userID}] = models.DefaultMemberPermissions

	_, err := service.CreateRole(context.Background(), userID, serverID, CreateRoleInput{
		Name:        "Mods",
		Permissions: models.PermissionManageMessages,
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestCreateRoleAllowsOwnerPermissions(t *testing.T) {
	repo := newFakePermissionRepository()
	service := NewPermissionService(repo)
	serverID := uuid.New()
	userID := uuid.New()
	repo.permissions[[2]uuid.UUID{serverID, userID}] = models.AllPermissions

	role, err := service.CreateRole(context.Background(), userID, serverID, CreateRoleInput{
		Name:        "Mods",
		Permissions: models.PermissionManageMessages,
	})
	if err != nil {
		t.Fatalf("create role: %v", err)
	}
	if role.Name != "Mods" {
		t.Fatalf("expected role name Mods, got %q", role.Name)
	}
}

type fakePermissionRepository struct {
	permissions map[[2]uuid.UUID]int64
	roles       map[uuid.UUID]models.Role
}

func newFakePermissionRepository() *fakePermissionRepository {
	return &fakePermissionRepository{
		permissions: make(map[[2]uuid.UUID]int64),
		roles:       make(map[uuid.UUID]models.Role),
	}
}

func (r *fakePermissionRepository) UserPermissions(ctx context.Context, serverID, userID uuid.UUID) (int64, error) {
	permissions, ok := r.permissions[[2]uuid.UUID{serverID, userID}]
	if !ok {
		return 0, repository.ErrServerNotFound
	}
	return permissions, nil
}

func (r *fakePermissionRepository) CreateRole(ctx context.Context, role models.Role) (models.Role, error) {
	now := time.Now().UTC()
	role.CreatedAt = now
	role.UpdatedAt = now
	r.roles[role.ID] = role
	return role, nil
}

func (r *fakePermissionRepository) ListRoles(ctx context.Context, serverID uuid.UUID) ([]models.Role, error) {
	roles := make([]models.Role, 0)
	for _, role := range r.roles {
		if role.ServerID == serverID {
			roles = append(roles, role)
		}
	}
	return roles, nil
}

func (r *fakePermissionRepository) GetRole(ctx context.Context, roleID uuid.UUID) (models.Role, error) {
	role, ok := r.roles[roleID]
	if !ok {
		return models.Role{}, repository.ErrRoleNotFound
	}
	return role, nil
}

func (r *fakePermissionRepository) UpdateRole(ctx context.Context, role models.Role) (models.Role, error) {
	r.roles[role.ID] = role
	return role, nil
}

func (r *fakePermissionRepository) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	delete(r.roles, roleID)
	return nil
}

func (r *fakePermissionRepository) AssignRole(ctx context.Context, serverID, userID, roleID uuid.UUID) error {
	return nil
}

func (r *fakePermissionRepository) RemoveRole(ctx context.Context, serverID, userID, roleID uuid.UUID) error {
	return nil
}

func (r *fakePermissionRepository) CreateInvite(ctx context.Context, invite models.ServerInvite) (models.ServerInvite, error) {
	return invite, nil
}

func (r *fakePermissionRepository) GetInvite(ctx context.Context, code string) (models.ServerInvite, error) {
	return models.ServerInvite{}, repository.ErrInviteNotFound
}

func (r *fakePermissionRepository) JoinInvite(ctx context.Context, code string, userID uuid.UUID) (models.ServerInvite, error) {
	return models.ServerInvite{}, repository.ErrInviteNotFound
}

func (r *fakePermissionRepository) CreateAuditLog(ctx context.Context, serverID, actorID uuid.UUID, action, targetType string, targetID uuid.UUID, metadata string) error {
	return nil
}
