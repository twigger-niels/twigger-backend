package repository

import (
	"context"
	"twigger-backend/backend/auth-service/domain/entity"

	"github.com/google/uuid"
)

// WorkspaceRepository defines the interface for workspace data access
type WorkspaceRepository interface {
	// Core CRUD operations
	Create(ctx context.Context, workspace *entity.Workspace) error
	GetByID(ctx context.Context, workspaceID uuid.UUID) (*entity.Workspace, error)
	GetByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*entity.Workspace, error)
	Update(ctx context.Context, workspace *entity.Workspace) error
	Delete(ctx context.Context, workspaceID uuid.UUID) error

	// Member management
	AddMember(ctx context.Context, member *entity.WorkspaceMember) error
	RemoveMember(ctx context.Context, workspaceID, userID uuid.UUID) error
	GetMembers(ctx context.Context, workspaceID uuid.UUID) ([]*entity.WorkspaceMember, error)
	GetMemberRole(ctx context.Context, workspaceID, userID uuid.UUID) (string, error)
	UpdateMemberRole(ctx context.Context, workspaceID, userID uuid.UUID, role string) error

	// Workspace queries
	GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]*entity.Workspace, error)
	IsMember(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error)
}
