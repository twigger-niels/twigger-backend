package entity

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	// Identity
	UserID      uuid.UUID  `json:"user_id"`
	FirebaseUID *string    `json:"firebase_uid,omitempty"`
	Email       string     `json:"email"`
	Username    string     `json:"username"`

	// Auth Provider
	EmailVerified bool    `json:"email_verified"`
	PhoneNumber   *string `json:"phone_number,omitempty"`
	PhotoURL      *string `json:"photo_url,omitempty"`
	Provider      string  `json:"provider"` // 'email', 'google.com', 'facebook.com', 'apple.com'

	// Localization (reuses existing system)
	PreferredLanguageID *uuid.UUID `json:"preferred_language_id,omitempty"`
	CountryID           *uuid.UUID `json:"country_id,omitempty"`

	// Spatial (existing field from user table)
	Location              *string `json:"location,omitempty"` // WKT format from geography(Point, 4326)
	DetectedHardinessZone *string `json:"detected_hardiness_zone,omitempty"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// IsDeleted returns true if the user has been soft deleted
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

// IsActive returns true if the user is not deleted
func (u *User) IsActive() bool {
	return !u.IsDeleted()
}

// HasFirebaseUID returns true if the user has a Firebase UID
func (u *User) HasFirebaseUID() bool {
	return u.FirebaseUID != nil && *u.FirebaseUID != ""
}

// UpdateLastLogin updates the last login timestamp
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
}

// SoftDelete marks the user as deleted
func (u *User) SoftDelete() {
	now := time.Now()
	u.DeletedAt = &now
}
