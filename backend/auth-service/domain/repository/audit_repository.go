package repository

import (
	"context"
	"time"
	"twigger-backend/backend/auth-service/domain/entity"

	"github.com/google/uuid"
)

// AuditRepository defines the interface for audit log data access
type AuditRepository interface {
	// Log operations
	LogEvent(ctx context.Context, event *entity.AuditEvent) error

	// Query operations
	GetUserEvents(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*entity.AuditEvent, error)
	GetUserEventsByType(ctx context.Context, userID uuid.UUID, eventType entity.AuditEventType, limit int) ([]*entity.AuditEvent, error)
	GetEventsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.AuditEvent, error)

	// Analytics
	CountEventsByType(ctx context.Context, eventType entity.AuditEventType, startDate, endDate time.Time) (int64, error)
	GetFailedLoginAttempts(ctx context.Context, userID uuid.UUID, since time.Time) (int, error)
}
