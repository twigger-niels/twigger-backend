-- Migration 007 Rollback: Remove Spatial Indexes for Garden Service
-- This removes all GIST indexes created in 007_add_spatial_indexes.up.sql

-- Drop gardens table spatial indexes
DROP INDEX IF EXISTS idx_gardens_boundary;
DROP INDEX IF EXISTS idx_gardens_location;
DROP INDEX IF EXISTS idx_gardens_user_boundary;

-- Drop garden zones table spatial indexes
DROP INDEX IF EXISTS idx_garden_zones_geometry;
DROP INDEX IF EXISTS idx_garden_zones_garden_geometry;

-- Drop garden features table spatial indexes
DROP INDEX IF EXISTS idx_garden_features_geometry;
DROP INDEX IF EXISTS idx_garden_features_type_height;

-- Drop garden plants table spatial indexes
DROP INDEX IF EXISTS idx_garden_plants_location;
DROP INDEX IF EXISTS idx_garden_plants_garden_location;
DROP INDEX IF EXISTS idx_garden_plants_active;

-- Drop climate zones table spatial index (if exists)
DROP INDEX IF EXISTS idx_climate_zones_boundary;

-- Verify index removal
DO $$
DECLARE
    idx_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO idx_count
    FROM pg_indexes
    WHERE schemaname = 'public'
      AND indexname LIKE 'idx_garden%';

    RAISE NOTICE 'Remaining garden-related indexes: %', idx_count;
END $$;
