package repository

import (
	"context"

	"twigger-backend/backend/plant-service/domain/entity"
	"twigger-backend/backend/plant-service/pkg/types"
)

// PlantRepository defines the interface for plant data access
type PlantRepository interface {
	// Basic CRUD operations (language context required for common names)
	FindByID(ctx context.Context, plantID, languageID string, countryID *string) (*entity.Plant, error)
	FindByIDs(ctx context.Context, plantIDs []string, languageID string, countryID *string) ([]*entity.Plant, error)
	Create(ctx context.Context, plant *entity.Plant) error
	Update(ctx context.Context, plant *entity.Plant) error
	Delete(ctx context.Context, plantID string) error

	// Search operations (with localization)
	Search(ctx context.Context, query string, filter *SearchFilter, languageID string, countryID *string) (*SearchResult, error)
	FindByBotanicalName(ctx context.Context, botanicalName, languageID string, countryID *string) (*entity.Plant, error)
	FindByCommonName(ctx context.Context, commonName, languageID string, countryID *string) ([]*entity.Plant, error)

	// Taxonomy queries (with localization)
	FindByFamily(ctx context.Context, familyName string, languageID string, countryID *string, limit, offset int) ([]*entity.Plant, error)
	FindByGenus(ctx context.Context, genusName string, languageID string, countryID *string, limit, offset int) ([]*entity.Plant, error)
	FindBySpecies(ctx context.Context, genusName, speciesName, languageID string, countryID *string) ([]*entity.Plant, error)

	// Growing conditions queries
	GetGrowingConditions(ctx context.Context, plantID, countryID, languageID string) (*types.GrowingConditions, error)
	FindByGrowingConditions(ctx context.Context, filter *GrowingConditionsFilter) ([]*entity.Plant, error)

	// Physical characteristics queries
	GetPhysicalCharacteristics(ctx context.Context, plantID, languageID string) (*types.PhysicalCharacteristics, error)

	// Companion plant queries (with localization support)
	GetCompanions(ctx context.Context, plantID, languageID string, countryID *string, filter *entity.CompanionFilter) ([]*entity.Companion, error)
	GetCompanionsByType(ctx context.Context, plantID, languageID string, countryID *string, relType types.RelationshipType) ([]*entity.Companion, error)
	CreateCompanionRelationship(ctx context.Context, companion *entity.Companion) error
	DeleteCompanionRelationship(ctx context.Context, relationshipID string) error

	// Bulk operations
	BulkCreate(ctx context.Context, plants []*entity.Plant) error

	// Statistics
	Count(ctx context.Context, filter *SearchFilter) (int64, error)
}

// SearchFilter represents search and filter criteria
type SearchFilter struct {
	// Text search
	Query string

	// Taxonomy filters
	FamilyName  *string
	GenusName   *string
	SpeciesName *string
	PlantType   *types.PlantType

	// Growing condition filters
	HardinessZone  *string
	SunRequirement *types.SunRequirement
	WaterNeeds     *types.WaterNeeds

	// Physical characteristic filters
	MinHeight *float64 // meters
	MaxHeight *float64 // meters
	GrowthRate *types.GrowthRate

	// Trait filters
	Evergreen *bool
	Deciduous *bool
	Toxic     *bool

	// Pagination
	Limit  int
	Cursor *string // ID of last item from previous page (for cursor-based pagination)

	// Sorting
	SortBy    SortField
	SortOrder SortOrder
}

// SortField represents fields that can be used for sorting
type SortField string

const (
	SortByBotanicalName SortField = "botanical_name"
	SortByCommonName    SortField = "common_name"
	SortByFamilyName    SortField = "family_name"
	SortByGenusName     SortField = "genus_name"
	SortByRelevance     SortField = "relevance" // For search results
	SortByCreatedAt     SortField = "created_at"
)

// SortOrder represents sort direction
type SortOrder string

const (
	SortAsc  SortOrder = "ASC"
	SortDesc SortOrder = "DESC"
)

// SearchResult represents search results with metadata
type SearchResult struct {
	Plants     []*entity.Plant `json:"plants"`
	Total      int64           `json:"total"`
	Limit      int             `json:"limit"`
	NextCursor *string         `json:"next_cursor,omitempty"` // Cursor for next page
	HasMore    bool            `json:"has_more"`
	Query      string          `json:"query,omitempty"`
}

// GrowingConditionsFilter represents filter criteria for growing conditions
type GrowingConditionsFilter struct {
	// Climate zone matching
	HardinessZone *string
	HeatZone      *string

	// Environmental requirements
	SunRequirements []types.SunRequirement
	WaterNeeds      *types.WaterNeeds
	SoilDrainage    *types.SoilDrainage

	// Tolerances
	DroughtTolerant *bool
	SaltTolerant    *bool
	WindTolerant    *bool

	// pH range
	MinPH *float64
	MaxPH *float64

	// Temporal
	FloweringMonth *int // 1-12
	FruitingMonth  *int // 1-12

	// Minimum confidence level
	MinConfidence *types.ConfidenceLevel

	// Pagination
	Limit  int
	Cursor *string
}

// DefaultSearchFilter returns a SearchFilter with default values
func DefaultSearchFilter() *SearchFilter {
	return &SearchFilter{
		Limit:     20,
		Cursor:    nil,
		SortBy:    SortByRelevance,
		SortOrder: SortDesc,
	}
}

// DefaultGrowingConditionsFilter returns a GrowingConditionsFilter with default values
func DefaultGrowingConditionsFilter() *GrowingConditionsFilter {
	minConfidence := types.ConfidenceProbable
	return &GrowingConditionsFilter{
		MinConfidence: &minConfidence,
		Limit:         20,
		Cursor:        nil,
	}
}
