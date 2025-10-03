package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"twigger-backend/backend/plant-service/domain/entity"
)

// LookupTableCache provides in-memory caching for lookup tables (languages, families, genera)
// These tables change infrequently and are accessed frequently, making them ideal for caching
type LookupTableCache struct {
	languages    map[string]*entity.Language      // key: language_id
	languageCode map[string]*entity.Language      // key: language_code
	families     map[string]*entity.PlantFamily   // key: family_id
	familyName   map[string]*entity.PlantFamily   // key: family_name
	genera       map[string]*entity.PlantGenus    // key: genus_id
	genusName    map[string]*entity.PlantGenus    // key: genus_name
	mu           sync.RWMutex
	ttl          time.Duration
	lastRefresh  time.Time
}

// NewLookupTableCache creates a new lookup table cache with specified TTL
func NewLookupTableCache(ttl time.Duration) *LookupTableCache {
	if ttl == 0 {
		ttl = 24 * time.Hour // Default: 24 hours
	}

	return &LookupTableCache{
		languages:    make(map[string]*entity.Language),
		languageCode: make(map[string]*entity.Language),
		families:     make(map[string]*entity.PlantFamily),
		familyName:   make(map[string]*entity.PlantFamily),
		genera:       make(map[string]*entity.PlantGenus),
		genusName:    make(map[string]*entity.PlantGenus),
		ttl:          ttl,
	}
}

// --- Language Cache ---

// SetLanguages populates the language cache
func (c *LookupTableCache) SetLanguages(languages []*entity.Language) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.languages = make(map[string]*entity.Language, len(languages))
	c.languageCode = make(map[string]*entity.Language, len(languages))

	for _, lang := range languages {
		c.languages[lang.LanguageID] = lang
		c.languageCode[lang.LanguageCode] = lang
	}

	c.lastRefresh = time.Now()
}

// GetLanguageByID retrieves a language from cache by ID
func (c *LookupTableCache) GetLanguageByID(languageID string) (*entity.Language, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isExpired() {
		return nil, false
	}

	lang, ok := c.languages[languageID]
	return lang, ok
}

// GetLanguageByCode retrieves a language from cache by code
func (c *LookupTableCache) GetLanguageByCode(code string) (*entity.Language, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isExpired() {
		return nil, false
	}

	lang, ok := c.languageCode[code]
	return lang, ok
}

// GetAllLanguages retrieves all cached languages
func (c *LookupTableCache) GetAllLanguages() ([]*entity.Language, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isExpired() {
		return nil, false
	}

	languages := make([]*entity.Language, 0, len(c.languages))
	for _, lang := range c.languages {
		languages = append(languages, lang)
	}
	return languages, true
}

// --- Plant Family Cache ---

// SetFamilies populates the plant family cache
func (c *LookupTableCache) SetFamilies(families []*entity.PlantFamily) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.families = make(map[string]*entity.PlantFamily, len(families))
	c.familyName = make(map[string]*entity.PlantFamily, len(families))

	for _, family := range families {
		c.families[family.FamilyID] = family
		c.familyName[family.FamilyName] = family
	}

	c.lastRefresh = time.Now()
}

// GetFamilyByID retrieves a plant family from cache by ID
func (c *LookupTableCache) GetFamilyByID(familyID string) (*entity.PlantFamily, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isExpired() {
		return nil, false
	}

	family, ok := c.families[familyID]
	return family, ok
}

// GetFamilyByName retrieves a plant family from cache by name
func (c *LookupTableCache) GetFamilyByName(name string) (*entity.PlantFamily, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isExpired() {
		return nil, false
	}

	family, ok := c.familyName[name]
	return family, ok
}

// --- Plant Genus Cache ---

// SetGenera populates the plant genus cache
func (c *LookupTableCache) SetGenera(genera []*entity.PlantGenus) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.genera = make(map[string]*entity.PlantGenus, len(genera))
	c.genusName = make(map[string]*entity.PlantGenus, len(genera))

	for _, genus := range genera {
		c.genera[genus.GenusID] = genus
		c.genusName[genus.GenusName] = genus
	}

	c.lastRefresh = time.Now()
}

// GetGenusByID retrieves a plant genus from cache by ID
func (c *LookupTableCache) GetGenusByID(genusID string) (*entity.PlantGenus, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isExpired() {
		return nil, false
	}

	genus, ok := c.genera[genusID]
	return genus, ok
}

// GetGenusByName retrieves a plant genus from cache by name
func (c *LookupTableCache) GetGenusByName(name string) (*entity.PlantGenus, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isExpired() {
		return nil, false
	}

	genus, ok := c.genusName[name]
	return genus, ok
}

// --- Cache Management ---

// isExpired checks if the cache has expired (must be called with read lock held)
func (c *LookupTableCache) isExpired() bool {
	if c.lastRefresh.IsZero() {
		return true
	}
	return time.Since(c.lastRefresh) > c.ttl
}

// IsExpired checks if the cache has expired (thread-safe public method)
func (c *LookupTableCache) IsExpired() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isExpired()
}

// Clear clears all cached data
func (c *LookupTableCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.languages = make(map[string]*entity.Language)
	c.languageCode = make(map[string]*entity.Language)
	c.families = make(map[string]*entity.PlantFamily)
	c.familyName = make(map[string]*entity.PlantFamily)
	c.genera = make(map[string]*entity.PlantGenus)
	c.genusName = make(map[string]*entity.PlantGenus)
	c.lastRefresh = time.Time{}
}

// GetStats returns cache statistics
func (c *LookupTableCache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"languages_count": len(c.languages),
		"families_count":  len(c.families),
		"genera_count":    len(c.genera),
		"last_refresh":    c.lastRefresh,
		"ttl_seconds":     c.ttl.Seconds(),
		"is_expired":      c.isExpired(),
	}
}

// CachedLanguageRepository wraps a language repository with caching
type CachedLanguageRepository struct {
	repo  interface {
		FindByID(ctx context.Context, languageID string) (*entity.Language, error)
		FindByCode(ctx context.Context, languageCode string) (*entity.Language, error)
		FindAll(ctx context.Context) ([]*entity.Language, error)
	}
	cache *LookupTableCache
}

// NewCachedLanguageRepository creates a cached language repository
func NewCachedLanguageRepository(repo interface {
	FindByID(ctx context.Context, languageID string) (*entity.Language, error)
	FindByCode(ctx context.Context, languageCode string) (*entity.Language, error)
	FindAll(ctx context.Context) ([]*entity.Language, error)
}, cache *LookupTableCache) *CachedLanguageRepository {
	return &CachedLanguageRepository{
		repo:  repo,
		cache: cache,
	}
}

// FindByID finds language by ID with caching
func (r *CachedLanguageRepository) FindByID(ctx context.Context, languageID string) (*entity.Language, error) {
	// Try cache first
	if lang, ok := r.cache.GetLanguageByID(languageID); ok {
		return lang, nil
	}

	// Cache miss - fetch from database
	lang, err := r.repo.FindByID(ctx, languageID)
	if err != nil {
		return nil, err
	}

	// If cache is expired, refresh it
	if r.cache.IsExpired() {
		go r.refreshCache(context.Background())
	}

	return lang, nil
}

// FindByCode finds language by code with caching
func (r *CachedLanguageRepository) FindByCode(ctx context.Context, languageCode string) (*entity.Language, error) {
	// Try cache first
	if lang, ok := r.cache.GetLanguageByCode(languageCode); ok {
		return lang, nil
	}

	// Cache miss - fetch from database
	lang, err := r.repo.FindByCode(ctx, languageCode)
	if err != nil {
		return nil, err
	}

	// If cache is expired, refresh it
	if r.cache.IsExpired() {
		go r.refreshCache(context.Background())
	}

	return lang, nil
}

// refreshCache refreshes the entire language cache
func (r *CachedLanguageRepository) refreshCache(ctx context.Context) {
	languages, err := r.repo.FindAll(ctx)
	if err != nil {
		// Log error but don't fail - cache will retry on next access
		fmt.Printf("Failed to refresh language cache: %v\n", err)
		return
	}

	r.cache.SetLanguages(languages)
}
