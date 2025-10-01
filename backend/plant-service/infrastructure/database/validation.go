package database

import (
	"fmt"
	"regexp"

	"github.com/google/uuid"
)

// Language ID validation patterns
var (
	// ISO 639-1 language codes (2 characters) or ISO 639-3 (3 characters)
	languageCodeRegex = regexp.MustCompile(`^[a-z]{2,3}$`)

	// UUID pattern for language_id
	uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

// ValidateLanguageID validates that a language ID is either a valid UUID or ISO language code
func ValidateLanguageID(languageID string) error {
	if languageID == "" {
		return fmt.Errorf("language_id is required")
	}

	// Check if it's a valid UUID
	if _, err := uuid.Parse(languageID); err == nil {
		return nil
	}

	// Check if it's a valid ISO language code (2-3 lowercase letters)
	if languageCodeRegex.MatchString(languageID) {
		return nil
	}

	return fmt.Errorf("invalid language_id format: %s (must be UUID or ISO 639 code)", languageID)
}

// ValidateCountryID validates that a country ID is either a valid UUID or ISO country code
func ValidateCountryID(countryID *string) error {
	if countryID == nil {
		return nil // Country is optional
	}

	if *countryID == "" {
		return fmt.Errorf("country_id cannot be empty string (use nil for no country)")
	}

	// Check if it's a valid UUID
	if _, err := uuid.Parse(*countryID); err == nil {
		return nil
	}

	// Check if it's a valid ISO 3166-1 alpha-2 country code (2 uppercase letters)
	countryCodeRegex := regexp.MustCompile(`^[A-Z]{2}$`)
	if countryCodeRegex.MatchString(*countryID) {
		return nil
	}

	return fmt.Errorf("invalid country_id format: %s (must be UUID or ISO 3166-1 alpha-2 code)", *countryID)
}

// ValidatePlantID validates that a plant ID is a valid UUID
func ValidatePlantID(plantID string) error {
	if plantID == "" {
		return fmt.Errorf("plant_id is required")
	}

	if _, err := uuid.Parse(plantID); err != nil {
		return fmt.Errorf("invalid plant_id format: %s (must be UUID)", plantID)
	}

	return nil
}
