package database

import (
	"context"
	"database/sql"
)

// CharacteristicTranslator provides translation for enum characteristic values
type CharacteristicTranslator struct {
	db *sql.DB
}

// NewCharacteristicTranslator creates a new translator
func NewCharacteristicTranslator(db *sql.DB) *CharacteristicTranslator {
	return &CharacteristicTranslator{db: db}
}

// Translate translates a characteristic value to the specified language
// Uses the translate_characteristic database function
func (ct *CharacteristicTranslator) Translate(ctx context.Context, characteristicType, value, languageID string) (string, error) {
	query := `SELECT translate_characteristic($1, $2, $3)`

	var translated string
	err := ct.db.QueryRowContext(ctx, query, characteristicType, value, languageID).Scan(&translated)
	if err != nil {
		// If translation fails, return original value
		return value, nil
	}

	return translated, nil
}

// TranslateArray translates an array of characteristic values
func (ct *CharacteristicTranslator) TranslateArray(ctx context.Context, characteristicType string, values []string, languageID string) ([]string, error) {
	if len(values) == 0 {
		return values, nil
	}

	translated := make([]string, len(values))
	for i, value := range values {
		t, err := ct.Translate(ctx, characteristicType, value, languageID)
		if err != nil {
			// If translation fails for one value, use original
			translated[i] = value
		} else {
			translated[i] = t
		}
	}

	return translated, nil
}
