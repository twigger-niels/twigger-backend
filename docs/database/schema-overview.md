# Database Schema Overview

## Comprehensive Plant Database Schema v5.0

This document provides a high-level overview of the comprehensive plant database schema designed for production use with multi-country, multilingual plant data and advanced spatial capabilities.

## Schema Architecture

### 1. **Core Infrastructure Layer**
Foundational tables for geographic and administrative data:

| Table | Purpose | Key Features |
|-------|---------|--------------|
| `countries` | Country definitions with spatial boundaries | MultiPolygon boundaries, climate systems |
| `climate_zones` | Climate zone mappings (USDA, RHS, etc.) | Spatial polygons, temperature ranges |
| `languages` | Multi-language support | ISO language codes, active status |
| `data_sources` | Source tracking and reliability | Reliability scoring, verification dates |

### 2. **Plant Taxonomy Layer**
Complete taxonomic hierarchy with scientific accuracy:

| Table | Purpose | Key Features |
|-------|---------|--------------|
| `plant_families` | Botanical families | Family names, common names |
| `plant_genera` | Genera within families | Genus classification |
| `plant_species` | Species definitions | Plant type, growth characteristics |
| `cultivars` | Cultivated varieties | Trade names, patent information |
| `plants` | Master plant catalog | Full botanical names, search optimization |
| `plant_synonyms` | Historical name tracking | Deprecated names, date tracking |

### 3. **Growing Conditions Layer**
Environmental requirements and adaptations:

| Table | Purpose | Key Features |
|-------|---------|--------------|
| `country_plants` | Country-specific plant data | Native status, legal restrictions, spatial ranges |
| `growing_conditions_assertions` | Environmental requirements | Sun, water, soil preferences with confidence levels |
| `physical_characteristics` | Plant physical traits | Size ranges, growth rates, flexible JSONB traits |
| `companion_relationships` | Plant compatibility | Beneficial/antagonistic relationships, spacing |

### 4. **User Gardens Layer**
Spatial garden management and plant tracking:

| Table | Purpose | Key Features |
|-------|---------|--------------|
| `users` | User accounts | Location detection, hardiness zones |
| `gardens` | User garden definitions | Spatial boundaries, environmental data |
| `garden_zones` | Garden subdivisions | Beds, borders, microclimates |
| `garden_features` | Physical features | Trees, buildings for shade analysis |
| `garden_plants` | Plant placements | Spatial locations, health tracking |

## Data Types and Standards

### Measurement Domains
Standardized measurement types with validation:

- **Temperature**: `temperature_c` (-273.1 to 100.0°C)
- **Length**: `length_m`, `length_cm`, `length_mm` (non-negative)
- **Area**: `area_m2` (non-negative square meters)
- **Weight**: `weight_g`, `weight_kg` (non-negative)
- **Time**: `days`, `hours`, `years` (non-negative integers)
- **Quality**: `percentage` (0-100%), `ph_value` (0-14), `rating` (1-5 stars)

### Enum Types
Controlled vocabularies for consistency:

- **`plant_type`**: tree, shrub, perennial, annual, biennial, bulb, grass, fern, climber, aquatic, succulent, palm, bamboo, orchid, vine
- **`confidence_level`**: very_low, low, moderate, probable, very_high, confirmed
- **`sun_requirement`**: full_sun, partial_sun, partial_shade, full_shade, morning_sun, afternoon_shade, dappled_shade
- **`water_needs`**: very_dry, dry, moderate, moist, wet, aquatic, bog
- **`soil_drainage`**: very_well_drained, well_drained, moderate_drainage, poorly_drained, waterlogged
- **`growth_rate`**: very_slow, slow, moderate, fast, very_fast

### Composite Types
Complex data structures for related measurements:

- **`temp_range`**: min, max, optimal temperatures
- **`ph_range`**: min, max, optimal pH values
- **`size_range`**: min, typical, max sizes

## Spatial Capabilities

### PostGIS Integration
Full spatial analysis support with:

- **GEOMETRY**: Planar coordinates for local calculations
- **GEOGRAPHY**: Spherical coordinates for global accuracy
- **SRID 4326**: WGS84 global coordinate system
- **Spatial Indexes**: GIST indexes for optimal performance

### Spatial Features
- Country and climate zone boundaries
- Garden mapping with zones and features
- Plant placement tracking
- Shade analysis from garden features
- Distance and area calculations
- Optimal planting location algorithms

## Performance Optimizations

### Indexing Strategy
Comprehensive indexing for optimal performance:

1. **Foreign Key Indexes**: All foreign key relationships
2. **Spatial Indexes**: GIST indexes on all geometry/geography columns
3. **Text Search Indexes**: GIN indexes for full-text search
4. **Composite Indexes**: Multi-column indexes for common queries

### Query Optimization
- Generated columns for calculated fields
- Materialized views for complex aggregations
- Spatial function optimization
- Connection pooling configuration

## Data Quality and Confidence

### Source Tracking
Every piece of growing condition and physical characteristic data is linked to:
- **Data Source**: Reliability-scored source
- **Confidence Level**: Six-level confidence scale
- **Verification Date**: Last verification timestamp

### Multi-Source Consensus
- Multiple sources can provide data for the same plant
- Confidence levels help prioritize conflicting information
- Source reliability scores weight the data quality

## Extensibility Features

### JSONB Flexibility
Physical characteristics include a flexible `traits` JSONB column for:
- Leaf shapes and colors
- Flower characteristics
- Bark textures
- Root depth information
- Custom properties

### Array Support
Multiple value support for:
- Common names (text arrays)
- Climate systems (text arrays)
- Sun requirements (enum arrays)
- Soil types (text arrays)
- Flowering/fruiting months (integer arrays)

## Integration Points

### External Systems
Schema designed for integration with:
- Weather APIs (for climate data)
- Plant databases (for data import)
- Mapping services (for spatial visualization)
- E-commerce platforms (for plant sales)

### API Support
Schema optimized for:
- REST API endpoints
- GraphQL queries
- Real-time garden monitoring
- Spatial analysis services

## Security and Compliance

### Data Protection
- No personal plant health data stored
- Location data anonymizable
- GDPR-compliant user data handling
- Audit trails for data modifications

### Access Control
- Row-level security capability
- Role-based access patterns
- Workspace isolation support
- Multi-tenant architecture ready

## Migration and Versioning

### Schema Evolution
- Forward-compatible design
- Migration scripts included
- Rollback capabilities
- Data preservation during updates

### Backward Compatibility
- Maintains compatibility with existing plant databases
- Standard botanical nomenclature
- Common garden management concepts
- Extensible without breaking changes

## Statistics and Scale

### Current Capacity
- **21 core tables** with full relationships
- **13 measurement domains** for data standardization
- **7 enum types** for controlled vocabularies
- **Full PostGIS spatial support** with analysis functions
- **Production-ready** with proper indexing and constraints

### Scalability Targets
- **10,000+ plant species** support
- **100,000+ gardens** capacity
- **1,000,000+ plant placements** tracking
- **Complex spatial queries** under 100ms
- **Multi-country deployment** ready

## Development Workflow

### Schema Management
1. **Development**: Local PostgreSQL with PostGIS
2. **Testing**: Automated schema validation
3. **Staging**: Cloud SQL replica testing
4. **Production**: Cloud SQL with backup/recovery

### Quality Assurance
- Comprehensive test suite
- Spatial function validation
- Performance benchmarking
- Data integrity checks

This schema represents a production-ready foundation for comprehensive plant database applications with advanced spatial capabilities and multi-source data management.