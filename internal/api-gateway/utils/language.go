package utils

import (
	"context"
	"net/http"
	"strings"
)

// LanguageContext holds language and country information for localization
type LanguageContext struct {
	LanguageID string
	CountryID  *string
}

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	languageContextKey contextKey = "languageContext"
	userIDKey          contextKey = "userID"
)

// ExtractLanguageContext extracts language context from the request
// Priority: 1) User preferences (from auth context), 2) Accept-Language header, 3) Default to "en"
func ExtractLanguageContext(r *http.Request) *LanguageContext {
	ctx := r.Context()

	// Try to get from context (set by auth middleware from user preferences)
	if langCtx, ok := ctx.Value(languageContextKey).(*LanguageContext); ok {
		return langCtx
	}

	// Parse Accept-Language header
	acceptLang := r.Header.Get("Accept-Language")
	if acceptLang != "" {
		langID, countryID := parseAcceptLanguage(acceptLang)
		return &LanguageContext{
			LanguageID: langID,
			CountryID:  countryID,
		}
	}

	// Default to English
	return &LanguageContext{
		LanguageID: "en",
		CountryID:  nil,
	}
}

// parseAcceptLanguage parses the Accept-Language header
// Examples:
//   - "en-US" -> languageID: "en", countryID: "US"
//   - "es-MX" -> languageID: "es", countryID: "MX"
//   - "en" -> languageID: "en", countryID: nil
//   - "en-US, es-MX;q=0.9" -> languageID: "en", countryID: "US" (first match)
func parseAcceptLanguage(acceptLang string) (string, *string) {
	// Split by comma to get individual language tags
	tags := strings.Split(acceptLang, ",")
	if len(tags) == 0 {
		return "en", nil
	}

	// Take the first tag (highest priority)
	tag := strings.TrimSpace(tags[0])

	// Remove quality value if present (e.g., ";q=0.9")
	if idx := strings.Index(tag, ";"); idx != -1 {
		tag = tag[:idx]
	}

	// Parse language-country format (e.g., "en-US")
	parts := strings.Split(tag, "-")
	if len(parts) >= 2 {
		lang := strings.ToLower(parts[0])
		country := strings.ToUpper(parts[1])
		return lang, &country
	}

	// Only language code provided
	return strings.ToLower(parts[0]), nil
}

// SetLanguageContext stores language context in the request context
func SetLanguageContext(ctx context.Context, langCtx *LanguageContext) context.Context {
	return context.WithValue(ctx, languageContextKey, langCtx)
}

// GetLanguageContext retrieves language context from the request context
func GetLanguageContext(ctx context.Context) *LanguageContext {
	if langCtx, ok := ctx.Value(languageContextKey).(*LanguageContext); ok {
		return langCtx
	}
	// Default fallback
	return &LanguageContext{
		LanguageID: "en",
		CountryID:  nil,
	}
}

// SetUserID stores user ID in the request context
func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID retrieves user ID from the request context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(userIDKey).(string); ok {
		return userID
	}
	return ""
}

// GetUserIDOrEmpty safely retrieves user ID, returning empty string if not present
func GetUserIDOrEmpty(ctx context.Context) string {
	return GetUserID(ctx)
}
