-- Migration 007: Add Spatial Indexes for Garden Service
-- Purpose: Create GIST indexes for all spatial columns to optimize PostGIS queries
-- Performance: Spatial queries without indexes can be 100-1000x slower
-- Reference: backend/garden-service/SPATIAL_QUERIES.md (Required Indexes section)

-- Gardens table spatial indexes
CREATE INDEX IF NOT EXISTS idx_gardens_boundary
ON gardens USING GIST(boundary);

CREATE INDEX IF NOT EXISTS idx_gardens_location
ON gardens USING GIST(location);

COMMENT ON INDEX idx_gardens_boundary IS 'GIST index for garden boundary polygons - supports ST_Contains, ST_Overlaps, ST_Area queries';
COMMENT ON INDEX idx_gardens_location IS 'GIST index for garden center point - supports ST_DWithin radius searches';

-- Garden zones table spatial indexes
CREATE INDEX IF NOT EXISTS idx_garden_zones_geometry
ON garden_zones USING GIST(geometry);

COMMENT ON INDEX idx_garden_zones_geometry IS 'GIST index for zone polygons - supports overlap detection and containment checks';

-- Garden features table spatial indexes
CREATE INDEX IF NOT EXISTS idx_garden_features_geometry
ON garden_features USING GIST(geometry);

COMMENT ON INDEX idx_garden_features_geometry IS 'GIST index for feature geometries (Point/Polygon mix) - supports spatial relationships';

-- Garden plants table spatial indexes
CREATE INDEX IF NOT EXISTS idx_garden_plants_location
ON garden_plants USING GIST(location);

COMMENT ON INDEX idx_garden_plants_location IS 'GIST index for plant locations - supports spacing checks with ST_DWithin and zone containment';

-- Climate zones table spatial indexes (for hardiness zone detection)
-- Note: climate_zones table may not exist yet (depends on Part 4 implementation)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'climate_zones') THEN
        CREATE INDEX IF NOT EXISTS idx_climate_zones_boundary
        ON climate_zones USING GIST(boundary);

        COMMENT ON INDEX idx_climate_zones_boundary IS 'GIST index for climate zone boundaries - supports ST_Contains for hardiness zone detection';
    END IF;
END $$;

-- Composite indexes for common query patterns
-- These combine spatial and non-spatial columns for specific use cases

-- Gardens by user with spatial filtering
CREATE INDEX IF NOT EXISTS idx_gardens_user_boundary
ON gardens USING GIST(user_id, boundary);

COMMENT ON INDEX idx_gardens_user_boundary IS 'Composite index for user-specific spatial queries';

-- Zones by garden with spatial filtering
CREATE INDEX IF NOT EXISTS idx_garden_zones_garden_geometry
ON garden_zones(garden_id, geometry);

COMMENT ON INDEX idx_garden_zones_garden_geometry IS 'Composite index for garden-specific zone queries';

-- Plants by garden with spatial filtering
CREATE INDEX IF NOT EXISTS idx_garden_plants_garden_location
ON garden_plants(garden_id, location);

COMMENT ON INDEX idx_garden_plants_garden_location IS 'Composite index for garden-specific plant location queries';

-- Active plants only (for FindActivePlants query optimization)
CREATE INDEX IF NOT EXISTS idx_garden_plants_active
ON garden_plants(garden_id, removed_date)
WHERE removed_date IS NULL;

COMMENT ON INDEX idx_garden_plants_active IS 'Partial index for active plants (not removed) - speeds up FindActivePlants queries';

-- Features by type for shade calculation (Part 4 dependency)
CREATE INDEX IF NOT EXISTS idx_garden_features_type_height
ON garden_features(garden_id, feature_type, height_m)
WHERE height_m IS NOT NULL;

COMMENT ON INDEX idx_garden_features_type_height IS 'Partial index for features with height - supports shade analysis queries';

-- Verify index creation and display statistics
DO $$
DECLARE
    idx_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO idx_count
    FROM pg_indexes
    WHERE schemaname = 'public'
      AND indexname LIKE 'idx_garden%';

    RAISE NOTICE 'Successfully created % spatial indexes for Garden Service', idx_count;
END $$;
