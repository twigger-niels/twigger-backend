package cache

import (
	"context"
	"fmt"

	"twigger-backend/backend/plant-service/domain/entity"
	"twigger-backend/backend/plant-service/domain/repository"
	"twigger-backend/backend/plant-service/pkg/types"
)

// CachedPlantRepository wraps a PlantRepository with caching
type CachedPlantRepository struct {
	repo  repository.PlantRepository
	cache *RedisCache
}

// NewCachedPlantRepository creates a new cached repository
func NewCachedPlantRepository(repo repository.PlantRepository, cache *RedisCache) *CachedPlantRepository {
	return &CachedPlantRepository{
		repo:  repo,
		cache: cache,
	}
}

// FindByID retrieves a plant by ID with caching
func (r *CachedPlantRepository) FindByID(ctx context.Context, plantID string) (*entity.Plant, error) {
	key := PlantKey(plantID)

	// Try cache first
	var plant entity.Plant
	err := r.cache.GetJSON(ctx, key, &plant)
	if err == nil {
		return &plant, nil
	}

	// Cache miss - fetch from database
	result, err := r.repo.FindByID(ctx, plantID)
	if err != nil {
		return nil, err
	}

	// Store in cache (ignore errors to not block the response)
	_ = r.cache.SetJSON(ctx, key, result, PlantTTL)

	return result, nil
}

// FindByIDs retrieves multiple plants (partial caching implementation)
func (r *CachedPlantRepository) FindByIDs(ctx context.Context, plantIDs []string) ([]*entity.Plant, error) {
	// For simplicity, delegate to repo
	// A full implementation would check cache for each ID
	return r.repo.FindByIDs(ctx, plantIDs)
}

// Search performs search with caching
func (r *CachedPlantRepository) Search(ctx context.Context, query string, filter *repository.SearchFilter) (*repository.SearchResult, error) {
	key := SearchKey(query, filter)

	// Try cache first
	var result repository.SearchResult
	err := r.cache.GetJSON(ctx, key, &result)
	if err == nil {
		return &result, nil
	}

	// Cache miss - perform search
	searchResult, err := r.repo.Search(ctx, query, filter)
	if err != nil {
		return nil, err
	}

	// Store in cache
	_ = r.cache.SetJSON(ctx, key, searchResult, SearchTTL)

	return searchResult, nil
}

// GetGrowingConditions retrieves growing conditions with caching
func (r *CachedPlantRepository) GetGrowingConditions(ctx context.Context, plantID, countryID string) (*types.GrowingConditions, error) {
	key := GrowingConditionsKey(plantID, countryID)

	// Try cache first
	var gc types.GrowingConditions
	err := r.cache.GetJSON(ctx, key, &gc)
	if err == nil {
		return &gc, nil
	}

	// Cache miss - fetch from database
	result, err := r.repo.GetGrowingConditions(ctx, plantID, countryID)
	if err != nil {
		return nil, err
	}

	if result != nil {
		_ = r.cache.SetJSON(ctx, key, result, GrowingConditionsTTL)
	}

	return result, nil
}

// GetPhysicalCharacteristics retrieves physical characteristics with caching
func (r *CachedPlantRepository) GetPhysicalCharacteristics(ctx context.Context, plantID string) (*types.PhysicalCharacteristics, error) {
	key := PhysicalCharacteristicsKey(plantID)

	// Try cache first
	var pc types.PhysicalCharacteristics
	err := r.cache.GetJSON(ctx, key, &pc)
	if err == nil {
		return &pc, nil
	}

	// Cache miss - fetch from database
	result, err := r.repo.GetPhysicalCharacteristics(ctx, plantID)
	if err != nil {
		return nil, err
	}

	if result != nil {
		_ = r.cache.SetJSON(ctx, key, result, PhysicalCharsTTL)
	}

	return result, nil
}

// GetCompanions retrieves companions with caching
func (r *CachedPlantRepository) GetCompanions(ctx context.Context, plantID string, filter *entity.CompanionFilter) ([]*entity.Companion, error) {
	key := CompanionFilterKey(plantID, filter)

	// Try cache first
	var companions []*entity.Companion
	err := r.cache.GetJSON(ctx, key, &companions)
	if err == nil {
		return companions, nil
	}

	// Cache miss - fetch from database
	result, err := r.repo.GetCompanions(ctx, plantID, filter)
	if err != nil {
		return nil, err
	}

	_ = r.cache.SetJSON(ctx, key, result, CompanionTTL)

	return result, nil
}

// Create creates a plant and invalidates relevant caches
func (r *CachedPlantRepository) Create(ctx context.Context, plant *entity.Plant) error {
	err := r.repo.Create(ctx, plant)
	if err != nil {
		return err
	}

	// Invalidate search caches
	_ = r.cache.DeletePattern(ctx, InvalidateSearchPattern())

	return nil
}

// Update updates a plant and invalidates caches
func (r *CachedPlantRepository) Update(ctx context.Context, plant *entity.Plant) error {
	err := r.repo.Update(ctx, plant)
	if err != nil {
		return err
	}

	// Invalidate plant cache
	_ = r.cache.Delete(ctx, PlantKey(plant.PlantID))

	// Invalidate search caches
	_ = r.cache.DeletePattern(ctx, InvalidateSearchPattern())

	return nil
}

// Delete removes a plant and invalidates caches
func (r *CachedPlantRepository) Delete(ctx context.Context, plantID string) error {
	err := r.repo.Delete(ctx, plantID)
	if err != nil {
		return err
	}

	// Invalidate all caches related to this plant
	_ = r.cache.DeletePattern(ctx, InvalidatePlantPattern(plantID))

	return nil
}

// Delegate methods that don't benefit much from caching

func (r *CachedPlantRepository) FindByBotanicalName(ctx context.Context, botanicalName string) (*entity.Plant, error) {
	return r.repo.FindByBotanicalName(ctx, botanicalName)
}

func (r *CachedPlantRepository) FindByCommonName(ctx context.Context, commonName string) ([]*entity.Plant, error) {
	return r.repo.FindByCommonName(ctx, commonName)
}

func (r *CachedPlantRepository) FindByFamily(ctx context.Context, familyName string, limit, offset int) ([]*entity.Plant, error) {
	return r.repo.FindByFamily(ctx, familyName, limit, offset)
}

func (r *CachedPlantRepository) FindByGenus(ctx context.Context, genusName string, limit, offset int) ([]*entity.Plant, error) {
	return r.repo.FindByGenus(ctx, genusName, limit, offset)
}

func (r *CachedPlantRepository) FindBySpecies(ctx context.Context, genusName, speciesName string) ([]*entity.Plant, error) {
	return r.repo.FindBySpecies(ctx, genusName, speciesName)
}

func (r *CachedPlantRepository) FindByGrowingConditions(ctx context.Context, filter *repository.GrowingConditionsFilter) ([]*entity.Plant, error) {
	return r.repo.FindByGrowingConditions(ctx, filter)
}

func (r *CachedPlantRepository) GetCompanionsByType(ctx context.Context, plantID string, relType types.RelationshipType) ([]*entity.Companion, error) {
	return r.repo.GetCompanionsByType(ctx, plantID, relType)
}

func (r *CachedPlantRepository) CreateCompanionRelationship(ctx context.Context, companion *entity.Companion) error {
	err := r.repo.CreateCompanionRelationship(ctx, companion)
	if err != nil {
		return err
	}

	// Invalidate companion caches for both plants
	_ = r.cache.DeletePattern(ctx, fmt.Sprintf("%s%s*", CompanionPrefix, companion.PlantAID))
	_ = r.cache.DeletePattern(ctx, fmt.Sprintf("%s%s*", CompanionPrefix, companion.PlantBID))

	return nil
}

func (r *CachedPlantRepository) DeleteCompanionRelationship(ctx context.Context, relationshipID string) error {
	// Note: We don't know which plants are affected without querying first
	// For simplicity, just invalidate all companion caches
	err := r.repo.DeleteCompanionRelationship(ctx, relationshipID)
	if err != nil {
		return err
	}

	_ = r.cache.DeletePattern(ctx, fmt.Sprintf("%s*", CompanionPrefix))

	return nil
}

func (r *CachedPlantRepository) BulkCreate(ctx context.Context, plants []*entity.Plant) error {
	err := r.repo.BulkCreate(ctx, plants)
	if err != nil {
		return err
	}

	// Invalidate search caches
	_ = r.cache.DeletePattern(ctx, InvalidateSearchPattern())

	return nil
}

func (r *CachedPlantRepository) Count(ctx context.Context, filter *repository.SearchFilter) (int64, error) {
	key := CountKey(filter)

	// Try cache first
	var count int64
	err := r.cache.GetJSON(ctx, key, &count)
	if err == nil {
		return count, nil
	}

	// Cache miss
	result, err := r.repo.Count(ctx, filter)
	if err != nil {
		return 0, err
	}

	_ = r.cache.SetJSON(ctx, key, result, SearchTTL)

	return result, nil
}
