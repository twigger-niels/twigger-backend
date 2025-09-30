# Entity-Relationship Diagram

## Comprehensive Plant Database Schema

### Core Infrastructure Layer

```mermaid
erDiagram
    COUNTRIES {
        uuid country_id PK
        char(2) country_code UK
        varchar(100) country_name
        text[] climate_systems
        varchar(20) default_climate_system
        geometry country_boundary
        timestamptz created_at
        timestamptz updated_at
    }

    CLIMATE_ZONES {
        uuid zone_id PK
        uuid country_id FK
        varchar(20) zone_system
        varchar(10) zone_code
        geometry zone_geometry
        temperature_c min_temp_c
        temperature_c max_temp_c
        timestamptz created_at
    }

    LANGUAGES {
        uuid language_id PK
        varchar(5) language_code UK
        varchar(50) language_name
        varchar(50) native_name
        boolean is_active
        timestamptz created_at
    }

    DATA_SOURCES {
        uuid source_id PK
        varchar(200) source_name
        text source_type
        rating reliability_score
        text website_url
        date last_verified
        timestamptz created_at
    }

    COUNTRIES ||--o{ CLIMATE_ZONES : contains
```

### Plant Taxonomy Layer

```mermaid
erDiagram
    PLANT_FAMILIES {
        uuid family_id PK
        varchar(100) family_name UK
        varchar(100) common_name
        timestamptz created_at
    }

    PLANT_GENERA {
        uuid genus_id PK
        uuid family_id FK
        varchar(100) genus_name
        timestamptz created_at
    }

    PLANT_SPECIES {
        uuid species_id PK
        uuid genus_id FK
        varchar(100) species_name
        plant_type plant_type
        timestamptz created_at
    }

    CULTIVARS {
        uuid cultivar_id PK
        uuid species_id FK
        varchar(100) cultivar_name
        varchar(100) trade_name
        varchar(50) patent_number
        date patent_expiry
        boolean propagation_restricted
        timestamptz created_at
    }

    PLANTS {
        uuid plant_id PK
        uuid species_id FK
        uuid cultivar_id FK
        text full_botanical_name
        timestamptz created_at
    }

    PLANT_SYNONYMS {
        uuid synonym_id PK
        uuid current_plant_id FK
        text old_name
        date date_deprecated
        timestamptz created_at
    }

    PLANT_FAMILIES ||--o{ PLANT_GENERA : contains
    PLANT_GENERA ||--o{ PLANT_SPECIES : contains
    PLANT_SPECIES ||--o{ CULTIVARS : has
    PLANT_SPECIES ||--o{ PLANTS : "species of"
    CULTIVARS ||--o{ PLANTS : "cultivar of"
    PLANTS ||--o{ PLANT_SYNONYMS : "has synonyms"
```

### Plant Information Layer

```mermaid
erDiagram
    PLANTS {
        uuid plant_id PK
        uuid species_id FK
        uuid cultivar_id FK
        text full_botanical_name
        timestamptz created_at
    }

    COUNTRIES {
        uuid country_id PK
        char(2) country_code UK
        varchar(100) country_name
    }

    DATA_SOURCES {
        uuid source_id PK
        varchar(200) source_name
        rating reliability_score
    }

    COUNTRY_PLANTS {
        uuid country_plant_id PK
        uuid country_id FK
        uuid plant_id FK
        text native_status
        text legal_status
        geometry native_range
        timestamptz created_at
        timestamptz updated_at
    }

    GROWING_CONDITIONS_ASSERTIONS {
        uuid assertion_id PK
        uuid country_plant_id FK
        uuid source_id FK
        confidence_level confidence
        text hardiness_zones
        text heat_zones
        sun_requirement[] sun_requirements
        boolean shade_tolerance
        water_needs water_needs
        percentage humidity_preference
        text[] soil_types
        soil_drainage soil_drainage
        ph_range ph_preference
        boolean salt_tolerant
        boolean wind_tolerant
        boolean drought_tolerant
        integer[] flowering_months
        integer[] fruiting_months
        timestamptz created_at
    }

    PHYSICAL_CHARACTERISTICS {
        uuid characteristic_id PK
        uuid plant_id FK
        uuid source_id FK
        size_range mature_height
        size_range mature_spread
        growth_rate growth_rate
        jsonb traits
        timestamptz created_at
    }

    COMPANION_RELATIONSHIPS {
        uuid relationship_id PK
        uuid plant_a_id FK
        uuid plant_b_id FK
        text relationship_type
        text[] benefits
        length_m optimal_distance_m
        length_m max_distance_m
        timestamptz created_at
    }

    COUNTRIES ||--o{ COUNTRY_PLANTS : "plants in"
    PLANTS ||--o{ COUNTRY_PLANTS : "found in"
    COUNTRY_PLANTS ||--o{ GROWING_CONDITIONS_ASSERTIONS : "conditions for"
    DATA_SOURCES ||--o{ GROWING_CONDITIONS_ASSERTIONS : sources
    PLANTS ||--o{ PHYSICAL_CHARACTERISTICS : "characteristics of"
    DATA_SOURCES ||--o{ PHYSICAL_CHARACTERISTICS : sources
    PLANTS ||--o{ COMPANION_RELATIONSHIPS : "plant_a"
    PLANTS ||--o{ COMPANION_RELATIONSHIPS : "plant_b"
```

### User Gardens Layer

```mermaid
erDiagram
    USERS {
        uuid user_id PK
        varchar(255) email UK
        varchar(100) username UK
        geography location
        varchar(10) detected_hardiness_zone
        timestamptz created_at
    }

    GARDENS {
        uuid garden_id PK
        uuid user_id FK
        varchar(200) garden_name
        geometry boundary
        geography location
        length_m elevation_m
        numeric slope_degrees
        text aspect
        varchar(10) hardiness_zone
        text garden_type
        timestamptz created_at
        timestamptz updated_at
    }

    GARDEN_ZONES {
        uuid zone_id PK
        uuid garden_id FK
        varchar(100) zone_name
        text zone_type
        geometry geometry
        area_m2 area_m2
        text soil_type
        boolean soil_amended
        text irrigation_type
        hours sun_hours_summer
        hours sun_hours_winter
        timestamptz created_at
    }

    GARDEN_FEATURES {
        uuid feature_id PK
        uuid garden_id FK
        text feature_type
        varchar(200) feature_name
        geometry geometry
        length_m height_m
        length_m canopy_diameter_m
        boolean deciduous
        timestamptz created_at
    }

    GARDEN_PLANTS {
        uuid garden_plant_id PK
        uuid garden_id FK
        uuid zone_id FK
        uuid plant_id FK
        geometry location
        date planted_date
        date removed_date
        integer quantity
        text plant_source
        text health_status
        text notes
        timestamptz created_at
        timestamptz updated_at
    }

    USERS ||--o{ GARDENS : owns
    GARDENS ||--o{ GARDEN_ZONES : "divided into"
    GARDENS ||--o{ GARDEN_FEATURES : contains
    GARDENS ||--o{ GARDEN_PLANTS : "plants in"
    GARDEN_ZONES ||--o{ GARDEN_PLANTS : "planted in"
    PLANTS ||--o{ GARDEN_PLANTS : "planted as"
```

## Key Relationships

### Primary Relationships
1. **Taxonomic Hierarchy**: Families → Genera → Species → Cultivars → Plants
2. **Geographic Context**: Countries → Climate Zones → Country Plants
3. **User Context**: Users → Gardens → Zones → Plant Placements
4. **Data Quality**: Data Sources → Growing Conditions & Physical Characteristics

### Spatial Relationships
- Countries have spatial boundaries (MultiPolygon)
- Climate zones have spatial geometry (MultiPolygon)
- Gardens have boundaries and locations (Polygon + Point)
- Garden zones have geometry (Polygon)
- Garden features have geometry (Point/Polygon)
- Plant placements have locations (Point)

### Constraints and Rules
1. **Unique Constraints**:
   - One cultivar per species with same name
   - One plant per species-cultivar combination
   - One country-plant combination per country

2. **Spatial Constraints**:
   - Garden zones must be within garden boundaries
   - Plant placements must be within zone boundaries

3. **Data Quality**:
   - Source reliability scoring (1-5 stars)
   - Confidence levels for growing conditions
   - Multiple sources can provide data for same plant

## Domain Types

### Measurement Domains
- `temperature_c`: -273.1 to 100.0°C
- `length_m`: Non-negative lengths in meters
- `area_m2`: Non-negative areas in square meters
- `percentage`: 0-100%
- `ph_value`: 0-14 pH scale
- `rating`: 1-5 star rating

### Composite Types
- `temp_range`: min, max, optimal temperatures
- `ph_range`: min, max, optimal pH values
- `size_range`: min, typical, max sizes

### Enum Types
- `plant_type`: tree, shrub, perennial, annual, etc.
- `confidence_level`: very_low to confirmed
- `sun_requirement`: full_sun to full_shade
- `water_needs`: very_dry to aquatic
- `soil_drainage`: very_well_drained to waterlogged
- `growth_rate`: very_slow to very_fast

## Indexes and Performance

### Spatial Indexes (GIST)
- Country boundaries
- Climate zone geometry
- Garden boundaries and locations
- Garden zone geometry
- Garden feature geometry
- Plant placement locations

### Text Search Indexes (GIN)
- Plant botanical names (full-text search)
- Physical characteristics traits (JSONB)

### Foreign Key Indexes
- All foreign key relationships for optimal join performance
- Critical for maintaining referential integrity at scale