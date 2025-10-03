-- Migration 006 Rollback: Remove GIN trigram indexes

-- Drop plant common names indexes
DROP INDEX IF EXISTS idx_plant_common_names_name_trgm;

-- Drop plant synonym indexes
DROP INDEX IF EXISTS idx_plant_synonyms_old_name_trgm;

-- Drop cultivar indexes
DROP INDEX IF EXISTS idx_cultivars_trade_name_trgm;
DROP INDEX IF EXISTS idx_cultivars_cultivar_name_trgm;

-- Drop plant species indexes
DROP INDEX IF EXISTS idx_plant_species_botanical_name_trgm;
DROP INDEX IF EXISTS idx_plant_species_species_name_trgm;

-- Drop plant genera indexes
DROP INDEX IF EXISTS idx_plant_genera_genus_name_trgm;

-- Drop plant families indexes
DROP INDEX IF EXISTS idx_plant_families_common_name_trgm;
DROP INDEX IF EXISTS idx_plant_families_family_name_trgm;

-- Note: We do NOT drop pg_trgm extension as other applications may depend on it
-- To manually remove: DROP EXTENSION IF EXISTS pg_trgm CASCADE;
