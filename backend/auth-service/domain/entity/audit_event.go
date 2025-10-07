package entity

import (
	"time"

	"github.com/google/uuid"
)

// AuditEvent represents an authentication audit log entry
type AuditEvent struct {
	ID         int64                  `json:"id"`
	UserID     *uuid.UUID             `json:"user_id,omitempty"`
	EventType  AuditEventType         `json:"event_type"`
	Success    bool                   `json:"success"`
	IPAddress  *string                `json:"ip_address,omitempty"`
	UserAgent  *string                `json:"user_agent,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

// AuditEventType represents the type of audit event
type AuditEventType string

const (
	EventUserRegistered AuditEventType = "user_registered"
	EventUserLogin      AuditEventType = "user_login"
	EventUserLogout     AuditEventType = "user_logout"
	EventTokenRefresh   AuditEventType = "token_refresh"
	EventSessionRevoked AuditEventType = "session_revoked"
	EventAccountDeleted AuditEventType = "account_deleted"
	EventPasswordReset  AuditEventType = "password_reset"
	EventAccountLinked  AuditEventType = "account_linked"
)

// IsValid checks if the event type is valid
func (e AuditEventType) IsValid() bool {
	switch e {
	case EventUserRegistered, EventUserLogin, EventUserLogout,
		EventTokenRefresh, EventSessionRevoked, EventAccountDeleted,
		EventPasswordReset, EventAccountLinked:
		return true
	default:
		return false
	}
}

// LinkedAccount represents a social provider linked to a user account
type LinkedAccount struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	Provider       string    `json:"provider"` // 'google.com', 'facebook.com', 'apple.com'
	ProviderUserID string    `json:"provider_user_id"`
	LinkedAt       time.Time `json:"linked_at"`
}
