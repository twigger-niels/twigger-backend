-- ============================================================================
-- COMPREHENSIVE PLANT DATABASE SCHEMA - PRODUCTION READY WITH POSTGIS
-- PostgreSQL 17+ with PostGIS
-- ============================================================================

-- Drop existing simple tables (they will be replaced)
DROP TABLE IF EXISTS garden_analysis;
DROP TABLE IF EXISTS plant_placements;
DROP TABLE IF EXISTS garden_zones;
DROP TABLE IF EXISTS gardens;
DROP TABLE IF EXISTS plants;
DROP TABLE IF EXISTS workspace_members;
DROP TABLE IF EXISTS workspaces;
DROP TABLE IF EXISTS users;

-- Drop existing enum types
DROP TYPE IF EXISTS soil_type;
DROP TYPE IF EXISTS water_requirement;
DROP TYPE IF EXISTS sun_requirement;
DROP TYPE IF EXISTS growing_season;
DROP TYPE IF EXISTS plant_category;
DROP TYPE IF EXISTS subscription_tier;
DROP TYPE IF EXISTS user_role;

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS postgis_topology;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS btree_gist;
CREATE EXTENSION IF NOT EXISTS ltree;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- SECTION 1: MEASUREMENT STANDARDIZATION DOMAINS
-- ============================================================================

-- Temperature measurements (always in Celsius)
CREATE DOMAIN temperature_c AS NUMERIC(4,1)
    CHECK (VALUE >= -273.1 AND VALUE <= 100.0);

-- Length measurements
CREATE DOMAIN length_mm AS NUMERIC(7,1) CHECK (VALUE >= 0);
CREATE DOMAIN length_cm AS NUMERIC(6,1) CHECK (VALUE >= 0);
CREATE DOMAIN length_m AS NUMERIC(5,2) CHECK (VALUE >= 0);

-- Weight measurements
CREATE DOMAIN weight_g AS NUMERIC(8,1) CHECK (VALUE >= 0);
CREATE DOMAIN weight_kg AS NUMERIC(8,3) CHECK (VALUE >= 0);

-- Area measurements
CREATE DOMAIN area_m2 AS NUMERIC(10,2) CHECK (VALUE >= 0);

-- Time measurements
CREATE DOMAIN days AS INTEGER CHECK (VALUE >= 0);
CREATE DOMAIN hours AS INTEGER CHECK (VALUE >= 0 AND VALUE <= 24);
CREATE DOMAIN years AS INTEGER CHECK (VALUE >= 0);

-- Percentage (0-100)
CREATE DOMAIN percentage AS NUMERIC(5,2)
    CHECK (VALUE >= 0 AND VALUE <= 100);

-- pH measurement (0-14)
CREATE DOMAIN ph_value AS NUMERIC(3,1)
    CHECK (VALUE >= 0 AND VALUE <= 14);

-- Rating (1-5 stars)
CREATE DOMAIN rating AS INTEGER
    CHECK (VALUE >= 1 AND VALUE <= 5);

-- ============================================================================
-- SECTION 2: CUSTOM TYPES
-- ============================================================================

-- Plant classification
CREATE TYPE plant_type AS ENUM (
    'tree', 'shrub', 'perennial', 'annual', 'biennial',
    'bulb', 'grass', 'fern', 'climber', 'aquatic',
    'succulent', 'palm', 'bamboo', 'orchid', 'vine'
);

-- Confidence scoring
CREATE TYPE confidence_level AS ENUM (
    'very_low',    -- 0-20%
    'low',         -- 20-40%
    'moderate',    -- 40-60%
    'probable',    -- 60-80%
    'very_high',   -- 80-95%
    'confirmed'    -- 95-100%
);

-- Seasons
CREATE TYPE season AS ENUM (
    'spring', 'summer', 'autumn', 'winter', 'all_year'
);

-- Environmental conditions
CREATE TYPE sun_requirement AS ENUM (
    'full_sun', 'partial_sun', 'partial_shade', 'full_shade',
    'morning_sun', 'afternoon_shade', 'dappled_shade'
);

CREATE TYPE water_needs AS ENUM (
    'very_dry', 'dry', 'moderate', 'moist',
    'wet', 'aquatic', 'bog'
);

CREATE TYPE soil_drainage AS ENUM (
    'very_well_drained', 'well_drained', 'moderate_drainage',
    'poorly_drained', 'waterlogged'
);

-- Growth characteristics
CREATE TYPE growth_rate AS ENUM (
    'very_slow', 'slow', 'moderate', 'fast', 'very_fast'
);

-- Composite types
CREATE TYPE temp_range AS (
    min_c temperature_c,
    max_c temperature_c,
    optimal_c temperature_c
);

CREATE TYPE ph_range AS (
    min_ph ph_value,
    max_ph ph_value,
    optimal_ph ph_value
);

CREATE TYPE size_range AS (
    min_m length_m,
    typical_m length_m,
    max_m length_m
);

-- ============================================================================
-- SECTION 3: CORE INFRASTRUCTURE TABLES
-- ============================================================================

-- Countries with spatial boundaries
CREATE TABLE countries (
    country_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    country_code CHAR(2) UNIQUE NOT NULL,
    country_name VARCHAR(100) NOT NULL,

    -- Climate systems
    climate_systems TEXT[] NOT NULL,
    default_climate_system VARCHAR(20),

    -- Spatial data
    country_boundary GEOMETRY(MultiPolygon, 4326),

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Create spatial index for country boundaries
CREATE INDEX idx_countries_boundary ON countries USING GIST(country_boundary);

-- Climate zones with spatial data
CREATE TABLE climate_zones (
    zone_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    country_id UUID NOT NULL REFERENCES countries(country_id),
    zone_system VARCHAR(20) NOT NULL, -- 'USDA', 'RHS', etc.
    zone_code VARCHAR(10) NOT NULL,

    -- Spatial polygon for the zone
    zone_geometry GEOMETRY(MultiPolygon, 4326) NOT NULL,

    -- Temperature ranges
    min_temp_c temperature_c,
    max_temp_c temperature_c,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_zone UNIQUE (country_id, zone_system, zone_code)
);

CREATE INDEX idx_climate_zones_geometry ON climate_zones USING GIST(zone_geometry);

-- Languages
CREATE TABLE languages (
    language_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    language_code VARCHAR(5) UNIQUE NOT NULL,
    language_name VARCHAR(50) NOT NULL,
    native_name VARCHAR(50),
    is_active BOOLEAN DEFAULT TRUE,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Data sources with simplified tracking
CREATE TABLE data_sources (
    source_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    source_name VARCHAR(200) NOT NULL,
    source_type TEXT CHECK (source_type IN (
        'botanical_garden', 'university', 'government_db',
        'commercial_nursery', 'book', 'website', 'expert', 'observation'
    )),

    -- Simple reliability
    reliability_score rating DEFAULT 3,

    -- Basic metadata
    website_url TEXT,
    last_verified DATE,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- SECTION 4: PLANT TAXONOMY (SIMPLIFIED)
-- ============================================================================

-- Plant families
CREATE TABLE plant_families (
    family_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    family_name VARCHAR(100) UNIQUE NOT NULL,
    common_name VARCHAR(100),

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Plant genera
CREATE TABLE plant_genera (
    genus_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    family_id UUID NOT NULL REFERENCES plant_families(family_id),
    genus_name VARCHAR(100) NOT NULL,

    UNIQUE(family_id, genus_name),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Plant species
CREATE TABLE plant_species (
    species_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    genus_id UUID NOT NULL REFERENCES plant_genera(genus_id),
    species_name VARCHAR(100) NOT NULL,

    -- Basic characteristics
    plant_type plant_type NOT NULL,

    UNIQUE(genus_id, species_name),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Cultivars (simplified)
CREATE TABLE cultivars (
    cultivar_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    species_id UUID NOT NULL REFERENCES plant_species(species_id),

    -- Essential naming
    cultivar_name VARCHAR(100) NOT NULL,
    trade_name VARCHAR(100),

    -- Key legal info only
    patent_number VARCHAR(50),
    patent_expiry DATE,
    propagation_restricted BOOLEAN DEFAULT FALSE,

    UNIQUE(species_id, cultivar_name),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Combined plants table
CREATE TABLE plants (
    plant_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    species_id UUID NOT NULL REFERENCES plant_species(species_id),
    cultivar_id UUID REFERENCES cultivars(cultivar_id),

    -- Search optimization - simplified for now (will be updated with proper generated column)
    full_botanical_name TEXT,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_plant_cultivar UNIQUE(species_id, cultivar_id)
);

-- Simple synonym tracking
CREATE TABLE plant_synonyms (
    synonym_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    current_plant_id UUID NOT NULL REFERENCES plants(plant_id),
    old_name TEXT NOT NULL,
    date_deprecated DATE,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- SECTION 5: COUNTRY-SPECIFIC PLANT INFORMATION
-- ============================================================================

-- Country-specific plant information
CREATE TABLE country_plants (
    country_plant_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    country_id UUID NOT NULL REFERENCES countries(country_id),
    plant_id UUID NOT NULL REFERENCES plants(plant_id),

    -- Regional status
    native_status TEXT CHECK (native_status IN (
        'native', 'endemic', 'naturalized', 'introduced',
        'invasive', 'cultivated_only'
    )),

    -- Legal status
    legal_status TEXT CHECK (legal_status IN (
        'prohibited', 'restricted', 'unrestricted', 'protected'
    )),

    -- Native range as spatial data
    native_range GEOMETRY(MultiPolygon, 4326),

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_country_plant UNIQUE (country_id, plant_id)
);

CREATE INDEX idx_country_plants_native_range ON country_plants USING GIST(native_range);

-- ============================================================================
-- SECTION 6: GROWING CONDITIONS WITH SPATIAL CONTEXT
-- ============================================================================

-- Growing conditions assertions
CREATE TABLE growing_conditions_assertions (
    assertion_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    country_plant_id UUID NOT NULL REFERENCES country_plants(country_plant_id),
    source_id UUID NOT NULL REFERENCES data_sources(source_id),

    -- Confidence
    confidence confidence_level DEFAULT 'probable',

    -- Climate zones (using int4range for PostgreSQL compatibility)
    hardiness_zones TEXT, -- Will store as text for now
    heat_zones TEXT,      -- Will store as text for now

    -- Sun and shade
    sun_requirements sun_requirement[],
    shade_tolerance BOOLEAN DEFAULT FALSE,

    -- Water and humidity
    water_needs water_needs,
    humidity_preference percentage,

    -- Soil requirements
    soil_types TEXT[],
    soil_drainage soil_drainage,
    ph_preference ph_range,

    -- Tolerances
    salt_tolerant BOOLEAN DEFAULT FALSE,
    wind_tolerant BOOLEAN DEFAULT FALSE,
    drought_tolerant BOOLEAN DEFAULT FALSE,

    -- Temporal aspects
    flowering_months INTEGER[],
    fruiting_months INTEGER[],

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_growing_assertion UNIQUE (country_plant_id, source_id)
);

-- Create indexes for foreign keys
CREATE INDEX idx_growing_conditions_country_plant ON growing_conditions_assertions(country_plant_id);
CREATE INDEX idx_growing_conditions_source ON growing_conditions_assertions(source_id);

-- ============================================================================
-- SECTION 7: PHYSICAL CHARACTERISTICS
-- ============================================================================

-- Physical characteristics using JSONB for flexibility
CREATE TABLE physical_characteristics (
    characteristic_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    plant_id UUID NOT NULL REFERENCES plants(plant_id),
    source_id UUID REFERENCES data_sources(source_id),

    -- Core measurements (structured)
    mature_height size_range,
    mature_spread size_range,
    growth_rate growth_rate,

    -- Flexible traits in JSONB
    traits JSONB DEFAULT '{}',

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_characteristics_source UNIQUE (plant_id, source_id)
);

-- Create GIN index on JSONB for performance
CREATE INDEX idx_physical_traits ON physical_characteristics USING GIN(traits);

-- ============================================================================
-- SECTION 8: USER GARDENS WITH SPATIAL FEATURES
-- ============================================================================

-- Users table
CREATE TABLE users (
    user_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,

    -- User's primary location (for climate detection)
    location GEOGRAPHY(Point, 4326),
    detected_hardiness_zone VARCHAR(10),

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_location ON users USING GIST(location);

-- User gardens with spatial boundaries
CREATE TABLE gardens (
    garden_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(user_id),
    garden_name VARCHAR(200) NOT NULL,

    -- Garden boundary and location
    boundary GEOMETRY(Polygon, 4326),
    location GEOGRAPHY(Point, 4326), -- Center point for distance calculations

    -- Environmental data
    elevation_m length_m,
    slope_degrees NUMERIC(3,1) CHECK (slope_degrees >= 0 AND slope_degrees <= 90),
    aspect TEXT CHECK (aspect IN ('N','NE','E','SE','S','SW','W','NW','flat')),

    -- Detected zones (from spatial queries)
    hardiness_zone VARCHAR(10),

    -- Garden metadata
    garden_type TEXT CHECK (garden_type IN (
        'ornamental', 'vegetable', 'mixed', 'orchard', 'greenhouse'
    )),

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_gardens_boundary ON gardens USING GIST(boundary);
CREATE INDEX idx_gardens_location ON gardens USING GIST(location);
CREATE INDEX idx_gardens_user ON gardens(user_id);

-- Garden zones/beds within a garden
CREATE TABLE garden_zones (
    zone_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    garden_id UUID NOT NULL REFERENCES gardens(garden_id),
    zone_name VARCHAR(100),
    zone_type TEXT CHECK (zone_type IN (
        'bed', 'border', 'lawn', 'path', 'water', 'structure', 'compost'
    )),

    -- Spatial representation
    geometry GEOMETRY(Polygon, 4326) NOT NULL,
    area_m2 area_m2 GENERATED ALWAYS AS (ST_Area(geometry::geography)) STORED,

    -- Zone characteristics
    soil_type TEXT,
    soil_amended BOOLEAN DEFAULT FALSE,
    irrigation_type TEXT CHECK (irrigation_type IN (
        'none', 'drip', 'sprinkler', 'soaker', 'manual'
    )),

    -- Sun exposure
    sun_hours_summer hours,
    sun_hours_winter hours,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_garden_zones_geometry ON garden_zones USING GIST(geometry);
CREATE INDEX idx_garden_zones_garden ON garden_zones(garden_id);

-- Garden features (trees, buildings, structures)
CREATE TABLE garden_features (
    feature_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    garden_id UUID NOT NULL REFERENCES gardens(garden_id),
    feature_type TEXT CHECK (feature_type IN (
        'tree', 'shrub', 'building', 'fence', 'wall',
        'greenhouse', 'shed', 'pond', 'path'
    )),
    feature_name VARCHAR(200),

    -- Spatial representation (point or polygon)
    geometry GEOMETRY(Geometry, 4326) NOT NULL,

    -- Height for shade calculations
    height_m length_m,

    -- For trees specifically
    canopy_diameter_m length_m,
    deciduous BOOLEAN,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_garden_features_geometry ON garden_features USING GIST(geometry);
CREATE INDEX idx_garden_features_garden ON garden_features(garden_id);

-- Planted specimens in the garden
CREATE TABLE garden_plants (
    garden_plant_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    garden_id UUID NOT NULL REFERENCES gardens(garden_id),
    zone_id UUID REFERENCES garden_zones(zone_id),
    plant_id UUID NOT NULL REFERENCES plants(plant_id),

    -- Location and timing
    location GEOMETRY(Point, 4326) NOT NULL,
    planted_date DATE DEFAULT CURRENT_DATE,
    removed_date DATE,

    -- Plant specifics
    quantity INTEGER DEFAULT 1,
    plant_source TEXT, -- 'seed', 'cutting', 'nursery', etc.

    -- Health tracking
    health_status TEXT CHECK (health_status IN (
        'thriving', 'healthy', 'struggling', 'diseased', 'dead'
    )),

    -- Notes and observations
    notes TEXT,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_garden_plants_location ON garden_plants USING GIST(location);
CREATE INDEX idx_garden_plants_garden ON garden_plants(garden_id);
CREATE INDEX idx_garden_plants_zone ON garden_plants(zone_id);
CREATE INDEX idx_garden_plants_plant ON garden_plants(plant_id);

-- ============================================================================
-- SECTION 9: COMPANION PLANTING
-- ============================================================================

CREATE TABLE companion_relationships (
    relationship_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    plant_a_id UUID NOT NULL REFERENCES plants(plant_id),
    plant_b_id UUID NOT NULL REFERENCES plants(plant_id),

    relationship_type TEXT CHECK (relationship_type IN (
        'beneficial', 'antagonistic', 'neutral'
    )),

    -- Benefits
    benefits TEXT[],

    -- Optimal spacing for companions
    optimal_distance_m length_m,
    max_distance_m length_m,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT no_self_companion CHECK (plant_a_id != plant_b_id),
    CONSTRAINT unique_companion UNIQUE (plant_a_id, plant_b_id)
);

CREATE INDEX idx_companions_plant_a ON companion_relationships(plant_a_id);
CREATE INDEX idx_companions_plant_b ON companion_relationships(plant_b_id);

-- ============================================================================
-- SECTION 10: INDEXES FOR PERFORMANCE
-- ============================================================================

-- Text search indexes
CREATE INDEX idx_plants_botanical_name ON plants USING GIN(to_tsvector('english', COALESCE(full_botanical_name, '')));

-- Foreign key indexes (critical for performance)
CREATE INDEX idx_cultivars_species ON cultivars(species_id);
CREATE INDEX idx_plant_species_genus ON plant_species(genus_id);
CREATE INDEX idx_plant_genera_family ON plant_genera(family_id);
CREATE INDEX idx_country_plants_country ON country_plants(country_id);
CREATE INDEX idx_country_plants_plant ON country_plants(plant_id);