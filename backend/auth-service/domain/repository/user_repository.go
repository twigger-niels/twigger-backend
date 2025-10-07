package repository

import (
	"context"
	"twigger-backend/backend/auth-service/domain/entity"

	"github.com/google/uuid"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Core CRUD operations
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, userID uuid.UUID) (*entity.User, error)
	GetByFirebaseUID(ctx context.Context, firebaseUID string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	SoftDelete(ctx context.Context, userID uuid.UUID) error

	// Auth-specific operations
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	LinkProvider(ctx context.Context, userID uuid.UUID, provider, providerUserID string) error
	GetLinkedAccounts(ctx context.Context, userID uuid.UUID) ([]*entity.LinkedAccount, error)

	// Workspace-related operations
	GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]*entity.Workspace, error)
}
