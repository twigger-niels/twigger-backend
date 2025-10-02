package entity

import (
	"fmt"
	"time"
)

// Language represents a language for localization
type Language struct {
	LanguageID   string    `json:"language_id"`
	LanguageCode string    `json:"language_code"` // ISO 639-1 (e.g., "en", "es", "zh")
	LanguageName string    `json:"language_name"`
	NativeName   *string   `json:"native_name,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// Validate validates the language entity
func (l *Language) Validate() error {
	if l.LanguageCode == "" {
		return fmt.Errorf("language_code is required")
	}

	if len(l.LanguageCode) > 5 {
		return fmt.Errorf("language_code must be 5 characters or less (ISO 639-1 or 639-3)")
	}

	if l.LanguageName == "" {
		return fmt.Errorf("language_name is required")
	}

	return nil
}
