package service

import (
	"context"
	"fmt"
	"strings"

	"twigger-backend/backend/plant-service/domain/entity"
	"twigger-backend/backend/plant-service/domain/repository"
	"twigger-backend/backend/plant-service/pkg/types"
)

// PlantService provides business logic for plant operations
type PlantService struct {
	repo repository.PlantRepository
}

// NewPlantService creates a new plant service
func NewPlantService(repo repository.PlantRepository) *PlantService {
	return &PlantService{
		repo: repo,
	}
}

// GetPlant retrieves a plant by ID with optional growing conditions and characteristics
func (s *PlantService) GetPlant(ctx context.Context, plantID string, includeDetails bool) (*entity.Plant, error) {
	if plantID == "" {
		return nil, entity.ErrInvalidPlantID
	}

	plant, err := s.repo.FindByID(ctx, plantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plant: %w", err)
	}

	if includeDetails {
		// Load physical characteristics
		if pc, err := s.repo.GetPhysicalCharacteristics(ctx, plantID); err == nil && pc != nil {
			plant.PhysicalCharacteristics = pc
		}
		// Note: Growing conditions require country_id, so we don't load them here
	}

	return plant, nil
}

// GetPlantWithConditions retrieves a plant with growing conditions for a specific country
func (s *PlantService) GetPlantWithConditions(ctx context.Context, plantID, countryID string) (*entity.Plant, error) {
	plant, err := s.GetPlant(ctx, plantID, true)
	if err != nil {
		return nil, err
	}

	if countryID != "" {
		gc, err := s.repo.GetGrowingConditions(ctx, plantID, countryID)
		if err == nil && gc != nil {
			plant.GrowingConditions = gc
		}
	}

	return plant, nil
}

// SearchPlants performs a search with filtering and ranking
func (s *PlantService) SearchPlants(ctx context.Context, query string, filter *repository.SearchFilter) (*repository.SearchResult, error) {
	// Validate and sanitize query
	query = strings.TrimSpace(query)
	if len(query) > 200 {
		return nil, entity.ErrInvalidSearchQuery
	}

	// Use default filter if not provided
	if filter == nil {
		filter = repository.DefaultSearchFilter()
	}

	// Validate filter
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	// Perform search
	result, err := s.repo.Search(ctx, query, filter)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Apply additional ranking if query is provided
	if query != "" {
		s.rankSearchResults(result.Plants, query)
	}

	return result, nil
}

// FindByBotanicalName finds a plant by exact botanical name
func (s *PlantService) FindByBotanicalName(ctx context.Context, botanicalName string) (*entity.Plant, error) {
	botanicalName = strings.TrimSpace(botanicalName)
	if botanicalName == "" {
		return nil, fmt.Errorf("botanical name is required")
	}

	return s.repo.FindByBotanicalName(ctx, botanicalName)
}

// FindPlantsByFamily retrieves all plants in a family
func (s *PlantService) FindPlantsByFamily(ctx context.Context, familyName string, limit, offset int) ([]*entity.Plant, error) {
	familyName = strings.TrimSpace(familyName)
	if familyName == "" {
		return nil, fmt.Errorf("family name is required")
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	return s.repo.FindByFamily(ctx, familyName, limit, offset)
}

// FindPlantsByGenus retrieves all plants in a genus
func (s *PlantService) FindPlantsByGenus(ctx context.Context, genusName string, limit, offset int) ([]*entity.Plant, error) {
	genusName = strings.TrimSpace(genusName)
	if genusName == "" {
		return nil, fmt.Errorf("genus name is required")
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	return s.repo.FindByGenus(ctx, genusName, limit, offset)
}

// GetCompanionPlants retrieves companion relationships for a plant
func (s *PlantService) GetCompanionPlants(ctx context.Context, plantID string, beneficialOnly bool) ([]*entity.Companion, error) {
	if plantID == "" {
		return nil, entity.ErrInvalidPlantID
	}

	filter := &entity.CompanionFilter{
		PlantID:        &plantID,
		BeneficialOnly: beneficialOnly,
	}

	return s.repo.GetCompanions(ctx, plantID, filter)
}

// GetBeneficialCompanions retrieves only beneficial companion plants
func (s *PlantService) GetBeneficialCompanions(ctx context.Context, plantID string) ([]*entity.Companion, error) {
	return s.GetCompanionPlants(ctx, plantID, true)
}

// GetAntagonisticPlants retrieves plants that should not be planted together
func (s *PlantService) GetAntagonisticPlants(ctx context.Context, plantID string) ([]*entity.Companion, error) {
	if plantID == "" {
		return nil, entity.ErrInvalidPlantID
	}

	relType := types.RelationshipAntagonistic
	filter := &entity.CompanionFilter{
		PlantID:          &plantID,
		RelationshipType: &relType,
	}

	return s.repo.GetCompanions(ctx, plantID, filter)
}

// RecommendPlants recommends plants based on growing conditions
func (s *PlantService) RecommendPlants(ctx context.Context, hardinessZone string, sunReq types.SunRequirement, limit int) ([]*entity.Plant, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	filter := &repository.GrowingConditionsFilter{
		HardinessZone:   &hardinessZone,
		SunRequirements: []types.SunRequirement{sunReq},
		Limit:           limit,
		Offset:          0,
	}

	return s.repo.FindByGrowingConditions(ctx, filter)
}

// ValidatePlantCompatibility checks if two plants are compatible for companion planting
func (s *PlantService) ValidatePlantCompatibility(ctx context.Context, plantAID, plantBID string) (*CompatibilityResult, error) {
	if plantAID == "" || plantBID == "" {
		return nil, fmt.Errorf("both plant IDs are required")
	}

	if plantAID == plantBID {
		return nil, fmt.Errorf("cannot check compatibility with self")
	}

	// Get companion relationships for plant A
	companions, err := s.repo.GetCompanions(ctx, plantAID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get companions: %w", err)
	}

	// Check if plant B is in the companions
	for _, c := range companions {
		otherID, err := c.GetOtherPlantID(plantAID)
		if err != nil {
			continue
		}

		if otherID == plantBID {
			return &CompatibilityResult{
				Compatible:       c.IsBeneficial() || c.IsNeutral(),
				RelationshipType: c.RelationshipType,
				Benefits:         c.Benefits,
				RecommendedDistanceM: c.GetRecommendedDistanceM(),
			}, nil
		}
	}

	// No relationship found - assume neutral
	return &CompatibilityResult{
		Compatible:       true,
		RelationshipType: types.RelationshipNeutral,
		Benefits:         []string{},
	}, nil
}

// CreatePlant creates a new plant with validation
func (s *PlantService) CreatePlant(ctx context.Context, plant *entity.Plant) error {
	// Validate plant
	if err := plant.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if plant with same botanical name already exists
	existing, err := s.repo.FindByBotanicalName(ctx, plant.FullBotanicalName)
	if err == nil && existing != nil {
		return entity.ErrPlantAlreadyExists
	}

	// Create plant
	return s.repo.Create(ctx, plant)
}

// rankSearchResults applies additional ranking logic to search results
func (s *PlantService) rankSearchResults(plants []*entity.Plant, query string) {
	// Calculate search scores for each plant
	scores := make(map[string]int)
	for _, plant := range plants {
		scores[plant.PlantID] = plant.SearchScore(query)
	}

	// Sort plants by score (bubble sort for simplicity - good enough for small result sets)
	for i := 0; i < len(plants); i++ {
		for j := i + 1; j < len(plants); j++ {
			if scores[plants[j].PlantID] > scores[plants[i].PlantID] {
				plants[i], plants[j] = plants[j], plants[i]
			}
		}
	}
}

// CompatibilityResult represents the compatibility between two plants
type CompatibilityResult struct {
	Compatible           bool                   `json:"compatible"`
	RelationshipType     types.RelationshipType `json:"relationship_type"`
	Benefits             []string               `json:"benefits,omitempty"`
	RecommendedDistanceM *float64               `json:"recommended_distance_m,omitempty"`
}

// GetPlantStatistics returns statistics about plants in the database
func (s *PlantService) GetPlantStatistics(ctx context.Context) (*PlantStatistics, error) {
	total, err := s.repo.Count(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to count plants: %w", err)
	}

	return &PlantStatistics{
		TotalPlants: total,
	}, nil
}

// PlantStatistics contains database statistics
type PlantStatistics struct {
	TotalPlants     int64 `json:"total_plants"`
	TotalFamilies   int64 `json:"total_families,omitempty"`
	TotalGenera     int64 `json:"total_genera,omitempty"`
	TotalSpecies    int64 `json:"total_species,omitempty"`
	TotalCultivars  int64 `json:"total_cultivars,omitempty"`
}
