package repository

import (
	"context"
	"twigger-backend/backend/auth-service/domain/entity"

	"github.com/google/uuid"
)

// SessionRepository defines the interface for session data access
type SessionRepository interface {
	// Core CRUD operations
	Create(ctx context.Context, session *entity.Session) error
	GetByID(ctx context.Context, sessionID uuid.UUID) (*entity.Session, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Session, error)
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Session, error)

	// Session management
	Revoke(ctx context.Context, sessionID uuid.UUID) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
	RevokeByDeviceID(ctx context.Context, userID uuid.UUID, deviceID string) error

	// Cleanup
	DeleteExpired(ctx context.Context) (int64, error)
}
