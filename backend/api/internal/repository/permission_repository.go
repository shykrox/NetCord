package repository

import (
	"context"
	"errors"
	"time"

	"netcord/backend/api/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PermissionRepository interface {
	UserPermissions(ctx context.Context, serverID, userID uuid.UUID) (int64, error)
	CreateRole(ctx context.Context, role models.Role) (models.Role, error)
	ListRoles(ctx context.Context, serverID uuid.UUID) ([]models.Role, error)
	GetRole(ctx context.Context, roleID uuid.UUID) (models.Role, error)
	UpdateRole(ctx context.Context, role models.Role) (models.Role, error)
	DeleteRole(ctx context.Context, roleID uuid.UUID) error
	AssignRole(ctx context.Context, serverID, userID, roleID uuid.UUID) error
	RemoveRole(ctx context.Context, serverID, userID, roleID uuid.UUID) error
	CreateInvite(ctx context.Context, invite models.ServerInvite) (models.ServerInvite, error)
	GetInvite(ctx context.Context, code string) (models.ServerInvite, error)
	JoinInvite(ctx context.Context, code string, userID uuid.UUID) (models.ServerInvite, error)
	CreateAuditLog(ctx context.Context, serverID, actorID uuid.UUID, action, targetType string, targetID uuid.UUID, metadata string) error
}

type PostgresPermissionRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresPermissionRepository(pool *pgxpool.Pool) *PostgresPermissionRepository {
	return &PostgresPermissionRepository{pool: pool}
}

func (r *PostgresPermissionRepository) UserPermissions(ctx context.Context, serverID, userID uuid.UUID) (int64, error) {
	var ownerID uuid.UUID
	var rolePermissions int64
	err := r.pool.QueryRow(ctx, `
		SELECT s.owner_id, COALESCE(bit_or(r.permissions), 0)::bigint
		FROM servers s
		JOIN server_members sm ON sm.server_id = s.id AND sm.user_id = $2
		LEFT JOIN member_roles mr ON mr.server_id = s.id AND mr.user_id = sm.user_id
		LEFT JOIN roles r ON r.id = mr.role_id
		WHERE s.id = $1
		GROUP BY s.owner_id
	`, serverID, userID).Scan(&ownerID, &rolePermissions)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrServerNotFound
		}
		return 0, err
	}
	if ownerID == userID {
		return models.AllPermissions, nil
	}
	permissions := models.DefaultMemberPermissions | rolePermissions
	if permissions&models.PermissionAdministrator != 0 {
		return models.AllPermissions, nil
	}
	return permissions, nil
}

func (r *PostgresPermissionRepository) CreateRole(ctx context.Context, role models.Role) (models.Role, error) {
	created, err := scanRole(r.pool.QueryRow(ctx, `
		INSERT INTO roles (id, server_id, name, color, position, permissions)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, server_id, name, color, position, permissions, created_at, updated_at
	`, role.ID, role.ServerID, role.Name, role.Color, role.Position, role.Permissions))
	if err != nil {
		if isUniqueViolation(err) {
			return models.Role{}, ErrRoleConflict
		}
		return models.Role{}, err
	}
	return created, nil
}

func (r *PostgresPermissionRepository) ListRoles(ctx context.Context, serverID uuid.UUID) ([]models.Role, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, server_id, name, color, position, permissions, created_at, updated_at
		FROM roles
		WHERE server_id = $1
		ORDER BY position ASC, created_at ASC
	`, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRoles(rows)
}

func (r *PostgresPermissionRepository) GetRole(ctx context.Context, roleID uuid.UUID) (models.Role, error) {
	role, err := scanRole(r.pool.QueryRow(ctx, `
		SELECT id, server_id, name, color, position, permissions, created_at, updated_at
		FROM roles
		WHERE id = $1
	`, roleID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Role{}, ErrRoleNotFound
		}
		return models.Role{}, err
	}
	return role, nil
}

func (r *PostgresPermissionRepository) UpdateRole(ctx context.Context, role models.Role) (models.Role, error) {
	updated, err := scanRole(r.pool.QueryRow(ctx, `
		UPDATE roles
		SET name = $2, color = $3, position = $4, permissions = $5, updated_at = now()
		WHERE id = $1
		RETURNING id, server_id, name, color, position, permissions, created_at, updated_at
	`, role.ID, role.Name, role.Color, role.Position, role.Permissions))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Role{}, ErrRoleNotFound
		}
		if isUniqueViolation(err) {
			return models.Role{}, ErrRoleConflict
		}
		return models.Role{}, err
	}
	return updated, nil
}

func (r *PostgresPermissionRepository) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM roles WHERE id = $1`, roleID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrRoleNotFound
	}
	return nil
}

func (r *PostgresPermissionRepository) AssignRole(ctx context.Context, serverID, userID, roleID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		INSERT INTO member_roles (server_id, user_id, role_id)
		SELECT $1, $2, r.id
		FROM roles r
		JOIN server_members sm ON sm.server_id = r.server_id AND sm.user_id = $2
		WHERE r.id = $3 AND r.server_id = $1
		ON CONFLICT DO NOTHING
	`, serverID, userID, roleID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrRoleNotFound
	}
	return nil
}

func (r *PostgresPermissionRepository) RemoveRole(ctx context.Context, serverID, userID, roleID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM member_roles
		WHERE server_id = $1 AND user_id = $2 AND role_id = $3
	`, serverID, userID, roleID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrRoleNotFound
	}
	return nil
}

func (r *PostgresPermissionRepository) CreateInvite(ctx context.Context, invite models.ServerInvite) (models.ServerInvite, error) {
	created, err := scanInvite(r.pool.QueryRow(ctx, `
		INSERT INTO server_invites (id, server_id, created_by, code, max_uses, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, server_id, created_by, code, max_uses, uses, expires_at, created_at
	`, invite.ID, invite.ServerID, invite.CreatedBy, invite.Code, invite.MaxUses, invite.ExpiresAt))
	if err != nil {
		if isUniqueViolation(err) {
			return models.ServerInvite{}, ErrInviteUnavailable
		}
		return models.ServerInvite{}, err
	}
	return created, nil
}

func (r *PostgresPermissionRepository) GetInvite(ctx context.Context, code string) (models.ServerInvite, error) {
	invite, err := scanInviteWithServer(r.pool.QueryRow(ctx, `
		SELECT i.id, i.server_id, i.created_by, i.code, i.max_uses, i.uses, i.expires_at, i.created_at,
			s.id, s.owner_id, s.name, s.description, s.icon_url, s.created_at, s.updated_at
		FROM server_invites i
		JOIN servers s ON s.id = i.server_id
		WHERE i.code = $1
	`, code))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ServerInvite{}, ErrInviteNotFound
		}
		return models.ServerInvite{}, err
	}
	if inviteUnavailable(invite) {
		return models.ServerInvite{}, ErrInviteUnavailable
	}
	return invite, nil
}

func (r *PostgresPermissionRepository) JoinInvite(ctx context.Context, code string, userID uuid.UUID) (models.ServerInvite, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.ServerInvite{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	invite, err := scanInviteWithServer(tx.QueryRow(ctx, `
		SELECT i.id, i.server_id, i.created_by, i.code, i.max_uses, i.uses, i.expires_at, i.created_at,
			s.id, s.owner_id, s.name, s.description, s.icon_url, s.created_at, s.updated_at
		FROM server_invites i
		JOIN servers s ON s.id = i.server_id
		WHERE i.code = $1
		FOR UPDATE OF i
	`, code))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ServerInvite{}, ErrInviteNotFound
		}
		return models.ServerInvite{}, err
	}
	if inviteUnavailable(invite) {
		return models.ServerInvite{}, ErrInviteUnavailable
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO server_members (server_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`, invite.ServerID, userID, models.ServerRoleMember); err != nil {
		return models.ServerInvite{}, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE server_invites
		SET uses = uses + 1
		WHERE id = $1
	`, invite.ID); err != nil {
		return models.ServerInvite{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.ServerInvite{}, err
	}
	invite.Uses++
	return invite, nil
}

func (r *PostgresPermissionRepository) CreateAuditLog(ctx context.Context, serverID, actorID uuid.UUID, action, targetType string, targetID uuid.UUID, metadata string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO audit_logs (id, server_id, actor_id, action, target_type, target_id, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)
	`, uuid.New(), serverID, actorID, action, targetType, targetID, metadata)
	return err
}

func scanRole(row pgx.Row) (models.Role, error) {
	var role models.Role
	err := row.Scan(&role.ID, &role.ServerID, &role.Name, &role.Color, &role.Position, &role.Permissions, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return models.Role{}, err
	}
	return role, nil
}

func scanRoles(rows pgx.Rows) ([]models.Role, error) {
	roles := make([]models.Role, 0)
	for rows.Next() {
		role, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return roles, nil
}

func scanInvite(row pgx.Row) (models.ServerInvite, error) {
	var invite models.ServerInvite
	var maxUses pgtype.Int4
	var expiresAt pgtype.Timestamptz
	err := row.Scan(&invite.ID, &invite.ServerID, &invite.CreatedBy, &invite.Code, &maxUses, &invite.Uses, &expiresAt, &invite.CreatedAt)
	if err != nil {
		return models.ServerInvite{}, err
	}
	invite.MaxUses = nullableInt(maxUses)
	invite.ExpiresAt = nullableTime(expiresAt)
	return invite, nil
}

func scanInviteWithServer(row pgx.Row) (models.ServerInvite, error) {
	invite, err := scanInviteWithServerColumns(row)
	if err != nil {
		return models.ServerInvite{}, err
	}
	return invite, nil
}

func scanInviteWithServerColumns(row pgx.Row) (models.ServerInvite, error) {
	var invite models.ServerInvite
	var server models.Server
	var maxUses pgtype.Int4
	var expiresAt pgtype.Timestamptz
	err := row.Scan(
		&invite.ID, &invite.ServerID, &invite.CreatedBy, &invite.Code, &maxUses, &invite.Uses, &expiresAt, &invite.CreatedAt,
		&server.ID, &server.OwnerID, &server.Name, &server.Description, &server.IconURL, &server.CreatedAt, &server.UpdatedAt,
	)
	if err != nil {
		return models.ServerInvite{}, err
	}
	invite.MaxUses = nullableInt(maxUses)
	invite.ExpiresAt = nullableTime(expiresAt)
	publicServer := server.Public()
	invite.Server = &publicServer
	return invite, nil
}

func nullableInt(value pgtype.Int4) *int {
	if !value.Valid {
		return nil
	}
	v := int(value.Int32)
	return &v
}

func inviteUnavailable(invite models.ServerInvite) bool {
	if invite.ExpiresAt != nil && time.Now().UTC().After(*invite.ExpiresAt) {
		return true
	}
	return invite.MaxUses != nil && invite.Uses >= *invite.MaxUses
}
