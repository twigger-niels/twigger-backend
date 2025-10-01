-- ============================================================================
-- MIGRATION ADDENDUM: Performance Indexes for Localized Queries
-- Related to: 005_add_localization.sql
-- Purpose: Add critical indexes discovered during code review
-- ============================================================================

-- ============================================================================
-- INDEXES FOR TAXONOMY FILTERING (Used in Search and FindByFamily/Genus)
-- ============================================================================

-- For filtering by plant_type in search queries
CREATE INDEX IF NOT EXISTS idx_plant_species_type ON plant_species(plant_type);

-- For filtering by family_name in FindByFamily queries
CREATE INDEX IF NOT EXISTS idx_plant_families_name ON plant_families(family_name);

-- For filtering by genus_name in FindByGenus queries
CREATE INDEX IF NOT EXISTS idx_plant_genera_name ON plant_genera(genus_name);

-- For filtering by species_name in FindBySpecies queries
CREATE INDEX IF NOT EXISTS idx_plant_species_name ON plant_species(species_name);

-- ============================================================================
-- COMPOSITE INDEXES FOR COMMON QUERY PATTERNS
-- ============================================================================

-- For queries filtering by family and then sorting/limiting
CREATE INDEX IF NOT EXISTS idx_plant_genera_family_name ON plant_genera(family_id, genus_name);

-- For queries filtering by genus and then sorting/limiting
CREATE INDEX IF NOT EXISTS idx_plant_species_genus_name ON plant_species(genus_id, species_name);

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON INDEX idx_plant_common_names_lookup IS 'Critical for loadCommonNames batch queries - covers (plant_id, language_id, country_id)';
COMMENT ON INDEX idx_plant_common_names_name_lang IS 'Critical for FindByCommonName multi-language search';
COMMENT ON INDEX idx_plant_species_type IS 'Used by Search filter on plant_type';
COMMENT ON INDEX idx_plant_families_name IS 'Used by FindByFamily queries';
COMMENT ON INDEX idx_plant_genera_name IS 'Used by FindByGenus queries';
