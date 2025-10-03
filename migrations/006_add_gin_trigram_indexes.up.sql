-- Migration 006: Add GIN trigram indexes for ILIKE searches
-- This migration adds pg_trgm extension and creates GIN indexes for text search performance

-- Enable pg_trgm extension for trigram matching (required for GIN text search)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Plant families: family_name and common_name searches
CREATE INDEX IF NOT EXISTS idx_plant_families_family_name_trgm
ON plant_families USING GIN (family_name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_plant_families_common_name_trgm
ON plant_families USING GIN (common_name gin_trgm_ops);

-- Plant genera: genus_name searches
CREATE INDEX IF NOT EXISTS idx_plant_genera_genus_name_trgm
ON plant_genera USING GIN (genus_name gin_trgm_ops);

-- Plant species: species_name and full_botanical_name searches
CREATE INDEX IF NOT EXISTS idx_plant_species_species_name_trgm
ON plant_species USING GIN (species_name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_plant_species_botanical_name_trgm
ON plant_species USING GIN (full_botanical_name gin_trgm_ops);

-- Cultivars: cultivar_name and trade_name searches
CREATE INDEX IF NOT EXISTS idx_cultivars_cultivar_name_trgm
ON cultivars USING GIN (cultivar_name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_cultivars_trade_name_trgm
ON cultivars USING GIN (trade_name gin_trgm_ops);

-- Plant synonyms: old_name searches
CREATE INDEX IF NOT EXISTS idx_plant_synonyms_old_name_trgm
ON plant_synonyms USING GIN (old_name gin_trgm_ops);

-- Plant common names: common_name searches (for localized searches)
CREATE INDEX IF NOT EXISTS idx_plant_common_names_name_trgm
ON plant_common_names USING GIN (common_name gin_trgm_ops);

-- Performance notes:
-- GIN indexes with pg_trgm provide fast ILIKE/LIKE queries with wildcards
-- These indexes support pattern matching queries like: WHERE name ILIKE '%search%'
-- Index size: approximately 2-3x the column data size
-- Build time: may take several minutes on large datasets (>100K rows)
