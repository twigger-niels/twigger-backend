package cache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// Cache TTL constants
const (
	PlantTTL           = 1 * time.Hour      // Individual plants cached for 1 hour
	SearchTTL          = 15 * time.Minute   // Search results cached for 15 minutes
	CompanionTTL       = 1 * time.Hour      // Companion relationships cached for 1 hour
	GrowingConditionsTTL = 2 * time.Hour    // Growing conditions cached for 2 hours
	PhysicalCharsTTL   = 2 * time.Hour      // Physical characteristics cached for 2 hours
)

// Key prefixes
const (
	PlantPrefix             = "plant:"
	SearchPrefix            = "search:"
	CompanionPrefix         = "companion:"
	GrowingConditionsPrefix = "gc:"
	PhysicalCharsPrefix     = "pc:"
	CountPrefix             = "count:"
)

// PlantKey generates a cache key for a plant by ID (deprecated - use PlantKeyWithLanguage)
func PlantKey(plantID string) string {
	return fmt.Sprintf("%s%s", PlantPrefix, plantID)
}

// PlantKeyWithLanguage generates a language-aware cache key for a plant
func PlantKeyWithLanguage(plantID, languageID string, countryID *string) string {
	if countryID != nil {
		return fmt.Sprintf("%s%s:%s:%s", PlantPrefix, plantID, languageID, *countryID)
	}
	return fmt.Sprintf("%s%s:%s", PlantPrefix, plantID, languageID)
}

// SearchKey generates a cache key for search results (deprecated - use SearchKeyWithLanguage)
// Uses MD5 hash of query and filter parameters for consistent keys
func SearchKey(query string, filter interface{}) string {
	// Serialize filter to JSON for hashing
	filterJSON, err := json.Marshal(filter)
	if err != nil {
		// Fallback to just query if filter serialization fails
		return fmt.Sprintf("%s%s", SearchPrefix, hashString(query))
	}

	// Combine query and filter
	combined := fmt.Sprintf("%s:%s", query, string(filterJSON))
	hash := hashString(combined)

	return fmt.Sprintf("%s%s", SearchPrefix, hash)
}

// SearchKeyWithLanguage generates a language-aware cache key for search results
func SearchKeyWithLanguage(query string, filter interface{}, languageID string, countryID *string) string {
	// Serialize filter to JSON for hashing
	filterJSON, err := json.Marshal(filter)
	if err != nil {
		filterJSON = []byte("{}")
	}

	// Include language and country in the cache key
	countryPart := "global"
	if countryID != nil {
		countryPart = *countryID
	}

	combined := fmt.Sprintf("%s:%s:%s:%s", query, string(filterJSON), languageID, countryPart)
	hash := hashString(combined)

	return fmt.Sprintf("%s%s", SearchPrefix, hash)
}

// CompanionKey generates a cache key for companion relationships
func CompanionKey(plantID string) string {
	return fmt.Sprintf("%s%s", CompanionPrefix, plantID)
}

// CompanionFilterKey generates a cache key with filter applied
func CompanionFilterKey(plantID string, filter interface{}) string {
	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return CompanionKey(plantID)
	}
	hash := hashString(string(filterJSON))
	return fmt.Sprintf("%s%s:%s", CompanionPrefix, plantID, hash)
}

// GrowingConditionsKey generates a cache key for growing conditions
func GrowingConditionsKey(plantID, countryID string) string {
	return fmt.Sprintf("%s%s:%s", GrowingConditionsPrefix, plantID, countryID)
}

// PhysicalCharacteristicsKey generates a cache key for physical characteristics
func PhysicalCharacteristicsKey(plantID string) string {
	return fmt.Sprintf("%s%s", PhysicalCharsPrefix, plantID)
}

// CountKey generates a cache key for count queries
func CountKey(filter interface{}) string {
	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return fmt.Sprintf("%sall", CountPrefix)
	}
	hash := hashString(string(filterJSON))
	return fmt.Sprintf("%s%s", CountPrefix, hash)
}

// InvalidatePlantPattern returns a pattern to invalidate all plant-related caches
func InvalidatePlantPattern(plantID string) string {
	return fmt.Sprintf("*%s*", plantID)
}

// InvalidateSearchPattern returns a pattern to invalidate all search caches
func InvalidateSearchPattern() string {
	return fmt.Sprintf("%s*", SearchPrefix)
}

// hashString creates an MD5 hash of a string for use in cache keys
func hashString(s string) string {
	hasher := md5.New()
	hasher.Write([]byte(s))
	return hex.EncodeToString(hasher.Sum(nil))
}
