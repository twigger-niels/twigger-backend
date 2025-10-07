package entity

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an authentication session
type Session struct {
	SessionID  uuid.UUID              `json:"session_id"`
	UserID     uuid.UUID              `json:"user_id"`
	DeviceID   *string                `json:"device_id,omitempty"`
	DeviceInfo map[string]interface{} `json:"device_info,omitempty"`
	IPAddress  *string                `json:"ip_address,omitempty"`
	UserAgent  *string                `json:"user_agent,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	ExpiresAt  time.Time              `json:"expires_at"`
	RevokedAt  *time.Time             `json:"revoked_at,omitempty"`
}

// IsActive returns true if the session is not revoked and not expired
func (s *Session) IsActive() bool {
	return s.RevokedAt == nil && time.Now().Before(s.ExpiresAt)
}

// IsRevoked returns true if the session has been revoked
func (s *Session) IsRevoked() bool {
	return s.RevokedAt != nil
}

// IsExpired returns true if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// Revoke marks the session as revoked
func (s *Session) Revoke() {
	now := time.Now()
	s.RevokedAt = &now
}
