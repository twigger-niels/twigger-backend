package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"twigger-backend/backend/auth-service/domain/entity"
	"twigger-backend/backend/auth-service/domain/repository"

	"github.com/google/uuid"
)

// PostgresWorkspaceRepository implements WorkspaceRepository using PostgreSQL
type PostgresWorkspaceRepository struct {
	db *sql.DB
}

// NewPostgresWorkspaceRepository creates a new PostgresWorkspaceRepository
func NewPostgresWorkspaceRepository(db *sql.DB) repository.WorkspaceRepository {
	return &PostgresWorkspaceRepository{db: db}
}

// Create inserts a new workspace
func (r *PostgresWorkspaceRepository) Create(ctx context.Context, workspace *entity.Workspace) error {
	query := `
		INSERT INTO workspaces (workspace_id, owner_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	// Generate UUID if not provided
	if workspace.WorkspaceID == uuid.Nil {
		workspace.WorkspaceID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	if workspace.CreatedAt.IsZero() {
		workspace.CreatedAt = now
	}
	if workspace.UpdatedAt.IsZero() {
		workspace.UpdatedAt = now
	}

	_, err := r.db.ExecContext(ctx, query,
		workspace.WorkspaceID,
		workspace.OwnerID,
		workspace.Name,
		workspace.CreatedAt,
		workspace.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	return nil
}

// GetByID retrieves a workspace by workspace_id
func (r *PostgresWorkspaceRepository) GetByID(ctx context.Context, workspaceID uuid.UUID) (*entity.Workspace, error) {
	query := `
		SELECT workspace_id, owner_id, name, created_at, updated_at
		FROM workspaces
		WHERE workspace_id = $1
	`

	workspace := &entity.Workspace{}
	err := r.db.QueryRowContext(ctx, query, workspaceID).Scan(
		&workspace.WorkspaceID,
		&workspace.OwnerID,
		&workspace.Name,
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("workspace not found: %s", workspaceID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	return workspace, nil
}

// GetByOwnerID retrieves all workspaces owned by a user
func (r *PostgresWorkspaceRepository) GetByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*entity.Workspace, error) {
	query := `
		SELECT workspace_id, owner_id, name, created_at, updated_at
		FROM workspaces
		WHERE owner_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspaces by owner: %w", err)
	}
	defer rows.Close()

	var workspaces []*entity.Workspace
	for rows.Next() {
		workspace := &entity.Workspace{}
		err := rows.Scan(
			&workspace.WorkspaceID,
			&workspace.OwnerID,
			&workspace.Name,
			&workspace.CreatedAt,
			&workspace.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		workspaces = append(workspaces, workspace)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workspaces: %w", err)
	}

	return workspaces, nil
}

// Update updates an existing workspace
func (r *PostgresWorkspaceRepository) Update(ctx context.Context, workspace *entity.Workspace) error {
	query := `
		UPDATE workspaces
		SET name = $2, updated_at = $3
		WHERE workspace_id = $1
	`

	workspace.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		workspace.WorkspaceID,
		workspace.Name,
		workspace.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("workspace not found: %s", workspace.WorkspaceID)
	}

	return nil
}

// Delete deletes a workspace
func (r *PostgresWorkspaceRepository) Delete(ctx context.Context, workspaceID uuid.UUID) error {
	query := `DELETE FROM workspaces WHERE workspace_id = $1`

	result, err := r.db.ExecContext(ctx, query, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("workspace not found: %s", workspaceID)
	}

	return nil
}

// AddMember adds a member to a workspace
func (r *PostgresWorkspaceRepository) AddMember(ctx context.Context, member *entity.WorkspaceMember) error {
	query := `
		INSERT INTO workspace_members (workspace_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (workspace_id, user_id) DO UPDATE
		SET role = EXCLUDED.role
	`

	if member.JoinedAt.IsZero() {
		member.JoinedAt = time.Now()
	}

	_, err := r.db.ExecContext(ctx, query,
		member.WorkspaceID,
		member.UserID,
		member.Role,
		member.JoinedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to add workspace member: %w", err)
	}

	return nil
}

// RemoveMember removes a member from a workspace
func (r *PostgresWorkspaceRepository) RemoveMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
	query := `
		DELETE FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, workspaceID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove workspace member: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("member not found in workspace")
	}

	return nil
}

// GetMembers retrieves all members of a workspace
func (r *PostgresWorkspaceRepository) GetMembers(ctx context.Context, workspaceID uuid.UUID) ([]*entity.WorkspaceMember, error) {
	query := `
		SELECT workspace_id, user_id, role, joined_at
		FROM workspace_members
		WHERE workspace_id = $1
		ORDER BY joined_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace members: %w", err)
	}
	defer rows.Close()

	var members []*entity.WorkspaceMember
	for rows.Next() {
		member := &entity.WorkspaceMember{}
		err := rows.Scan(
			&member.WorkspaceID,
			&member.UserID,
			&member.Role,
			&member.JoinedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace member: %w", err)
		}
		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workspace members: %w", err)
	}

	return members, nil
}

// GetMemberRole retrieves the role of a member in a workspace
func (r *PostgresWorkspaceRepository) GetMemberRole(ctx context.Context, workspaceID, userID uuid.UUID) (string, error) {
	query := `
		SELECT role
		FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2
	`

	var role string
	err := r.db.QueryRowContext(ctx, query, workspaceID, userID).Scan(&role)

	if err == sql.ErrNoRows {
		return "", fmt.Errorf("member not found in workspace")
	}
	if err != nil {
		return "", fmt.Errorf("failed to get member role: %w", err)
	}

	return role, nil
}

// UpdateMemberRole updates the role of a member in a workspace
func (r *PostgresWorkspaceRepository) UpdateMemberRole(ctx context.Context, workspaceID, userID uuid.UUID, role string) error {
	query := `
		UPDATE workspace_members
		SET role = $3
		WHERE workspace_id = $1 AND user_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, workspaceID, userID, role)
	if err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("member not found in workspace")
	}

	return nil
}

// GetUserWorkspaces retrieves all workspaces for a user
func (r *PostgresWorkspaceRepository) GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]*entity.Workspace, error) {
	query := `
		SELECT w.workspace_id, w.owner_id, w.name, w.created_at, w.updated_at
		FROM workspaces w
		INNER JOIN workspace_members wm ON w.workspace_id = wm.workspace_id
		WHERE wm.user_id = $1
		ORDER BY w.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []*entity.Workspace
	for rows.Next() {
		workspace := &entity.Workspace{}
		err := rows.Scan(
			&workspace.WorkspaceID,
			&workspace.OwnerID,
			&workspace.Name,
			&workspace.CreatedAt,
			&workspace.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		workspaces = append(workspaces, workspace)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workspaces: %w", err)
	}

	return workspaces, nil
}

// IsMember checks if a user is a member of a workspace
func (r *PostgresWorkspaceRepository) IsMember(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM workspace_members
			WHERE workspace_id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, workspaceID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check workspace membership: %w", err)
	}

	return exists, nil
}
