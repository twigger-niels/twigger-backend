-- ============================================================================
-- MIGRATION: Add Comprehensive Localization Support
-- Version: 5.1
-- Description: Adds country+language specific localization for all reference data
-- ============================================================================

-- ============================================================================
-- SECTION 1: PLANT COMMON NAMES (Country + Language Specific)
-- ============================================================================

-- Plant common names with country+language context
CREATE TABLE plant_common_names (
    common_name_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    plant_id UUID NOT NULL REFERENCES plants(plant_id) ON DELETE CASCADE,
    language_id UUID NOT NULL REFERENCES languages(language_id),
    country_id UUID REFERENCES countries(country_id), -- NULL means global
    
    -- The localized name
    common_name VARCHAR(200) NOT NULL,
    
    -- Context and usage
    is_primary BOOLEAN DEFAULT FALSE, -- Primary name for this language/country
    is_colloquial BOOLEAN DEFAULT FALSE, -- Informal/slang name
    region VARCHAR(100), -- Specific region within country
    
    -- Metadata
    source_id UUID REFERENCES data_sources(source_id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure uniqueness at the right level
    CONSTRAINT unique_plant_name_per_context 
        UNIQUE(plant_id, language_id, country_id, common_name)
);

-- Indexes for performance
CREATE INDEX idx_plant_common_names_plant ON plant_common_names(plant_id);
CREATE INDEX idx_plant_common_names_language ON plant_common_names(language_id);
CREATE INDEX idx_plant_common_names_country ON plant_common_names(country_id);

-- CRITICAL: Composite index for the most common query pattern (FindByID with language fallback)
CREATE INDEX idx_plant_common_names_lookup ON plant_common_names(plant_id, language_id, country_id);

-- CRITICAL: For FindByCommonName searches across languages
CREATE INDEX idx_plant_common_names_name_lang ON plant_common_names(language_id, common_name);

-- Full-text search index
CREATE INDEX idx_plant_common_names_search ON plant_common_names
    USING GIN(to_tsvector('simple', common_name));

-- ============================================================================
-- SECTION 2: PLANT DESCRIPTIONS (Localized)
-- ============================================================================

CREATE TABLE plant_descriptions (
    description_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    plant_id UUID NOT NULL REFERENCES plants(plant_id) ON DELETE CASCADE,
    language_id UUID NOT NULL REFERENCES languages(language_id),
    country_id UUID REFERENCES countries(country_id), -- NULL for universal
    
    -- Different types of descriptions
    description_type VARCHAR(50) NOT NULL CHECK (description_type IN (
        'general', 'appearance', 'habitat', 'cultivation', 
        'uses', 'history', 'ecology', 'identification'
    )),
    
    -- The localized content
    title VARCHAR(200),
    content TEXT NOT NULL,
    
    -- Metadata
    source_id UUID REFERENCES data_sources(source_id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_description_per_type 
        UNIQUE(plant_id, language_id, country_id, description_type)
);

CREATE INDEX idx_plant_descriptions_plant ON plant_descriptions(plant_id);
CREATE INDEX idx_plant_descriptions_language ON plant_descriptions(language_id);
CREATE INDEX idx_plant_descriptions_type ON plant_descriptions(description_type);

-- ============================================================================
-- SECTION 3: LOCALIZED PLANT PROBLEMS
-- ============================================================================

-- Replace the text arrays in plant_problems with localized versions
CREATE TABLE plant_problems_i18n (
    problem_i18n_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    problem_id UUID NOT NULL REFERENCES plant_problems(problem_id) ON DELETE CASCADE,
    language_id UUID NOT NULL REFERENCES languages(language_id),
    country_id UUID REFERENCES countries(country_id),
    
    -- Localized content
    problem_name VARCHAR(200) NOT NULL,
    symptoms TEXT[] NOT NULL,
    treatments TEXT[] NOT NULL,
    prevention TEXT[],
    
    -- Regional variations
    regional_notes TEXT,
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_problem_translation 
        UNIQUE(problem_id, language_id, country_id)
);

CREATE INDEX idx_plant_problems_i18n_problem ON plant_problems_i18n(problem_id);
CREATE INDEX idx_plant_problems_i18n_language ON plant_problems_i18n(language_id);

-- ============================================================================
-- SECTION 4: CHARACTERISTIC VALUE TRANSLATIONS
-- ============================================================================

-- For translating enum values and controlled vocabularies
CREATE TABLE characteristic_translations (
    translation_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    language_id UUID NOT NULL REFERENCES languages(language_id),
    
    -- What we're translating
    characteristic_type VARCHAR(100) NOT NULL, -- 'sun_requirement', 'soil_type', etc.
    characteristic_value VARCHAR(100) NOT NULL, -- 'full_sun', 'clay', etc.
    
    -- The translation
    translated_value VARCHAR(200) NOT NULL,
    translated_description TEXT, -- Optional longer explanation
    
    -- Context
    is_technical BOOLEAN DEFAULT FALSE, -- Technical vs common term
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_characteristic_translation 
        UNIQUE(language_id, characteristic_type, characteristic_value)
);

CREATE INDEX idx_characteristic_trans_language ON characteristic_translations(language_id);
CREATE INDEX idx_characteristic_trans_type ON characteristic_translations(characteristic_type);

-- ============================================================================
-- SECTION 5: COMPANION PLANT BENEFITS (Localized)
-- ============================================================================

CREATE TABLE companion_benefits_i18n (
    benefit_i18n_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    relationship_id UUID NOT NULL REFERENCES companion_relationships(relationship_id) ON DELETE CASCADE,
    language_id UUID NOT NULL REFERENCES languages(language_id),
    
    -- Localized benefits description
    benefit_description TEXT NOT NULL,
    scientific_explanation TEXT,
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_companion_benefit_translation 
        UNIQUE(relationship_id, language_id)
);

CREATE INDEX idx_companion_benefits_i18n_rel ON companion_benefits_i18n(relationship_id);
CREATE INDEX idx_companion_benefits_i18n_lang ON companion_benefits_i18n(language_id);

-- ============================================================================
-- SECTION 6: COUNTRY NAMES (Localized)
-- ============================================================================

CREATE TABLE country_names_i18n (
    country_name_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    country_id UUID NOT NULL REFERENCES countries(country_id) ON DELETE CASCADE,
    language_id UUID NOT NULL REFERENCES languages(language_id),
    
    -- Localized country name
    country_name VARCHAR(100) NOT NULL,
    official_name VARCHAR(200), -- "Federal Republic of Germany"
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_country_translation 
        UNIQUE(country_id, language_id)
);

CREATE INDEX idx_country_names_i18n_country ON country_names_i18n(country_id);
CREATE INDEX idx_country_names_i18n_language ON country_names_i18n(language_id);

-- ============================================================================
-- SECTION 7: USER PREFERENCES ENHANCEMENT
-- ============================================================================

-- Add language and measurement preferences to users
ALTER TABLE users 
ADD COLUMN preferred_language_id UUID REFERENCES languages(language_id),
ADD COLUMN measurement_system VARCHAR(20) DEFAULT 'metric' 
    CHECK (measurement_system IN ('metric', 'imperial'));

-- Create index on language preference for filtering
CREATE INDEX idx_users_language ON users(preferred_language_id);

-- ============================================================================
-- SECTION 8: PHYSICAL TRAITS LOCALIZATION
-- ============================================================================

-- For translating the JSONB traits in physical_characteristics
CREATE TABLE physical_traits_i18n (
    trait_i18n_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    characteristic_id UUID NOT NULL REFERENCES physical_characteristics(characteristic_id) ON DELETE CASCADE,
    language_id UUID NOT NULL REFERENCES languages(language_id),
    
    -- Localized traits as JSONB
    -- Keys match the original traits JSONB, values are translated
    traits_localized JSONB NOT NULL,
    /* Example:
    {
        "leaf_shape": "palmada",
        "leaf_color": "verde",
        "flower_color": ["rojo", "rosa"],
        "bark_texture": "lisa"
    }
    */
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_traits_translation 
        UNIQUE(characteristic_id, language_id)
);

CREATE INDEX idx_physical_traits_i18n_char ON physical_traits_i18n(characteristic_id);
CREATE INDEX idx_physical_traits_i18n_lang ON physical_traits_i18n(language_id);
CREATE INDEX idx_physical_traits_i18n_traits ON physical_traits_i18n USING GIN(traits_localized);

-- ============================================================================
-- SECTION 9: GROWING CONDITIONS LOCALIZATION
-- ============================================================================

-- For soil types and other text arrays in growing_conditions_assertions
CREATE TABLE growing_conditions_i18n (
    growing_i18n_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    assertion_id UUID NOT NULL REFERENCES growing_conditions_assertions(assertion_id) ON DELETE CASCADE,
    language_id UUID NOT NULL REFERENCES languages(language_id),
    
    -- Localized soil types
    soil_types_localized TEXT[],
    
    -- Additional localized notes
    special_requirements TEXT,
    regional_tips TEXT,
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_growing_translation 
        UNIQUE(assertion_id, language_id)
);

CREATE INDEX idx_growing_conditions_i18n_assertion ON growing_conditions_i18n(assertion_id);
CREATE INDEX idx_growing_conditions_i18n_language ON growing_conditions_i18n(language_id);

-- ============================================================================
-- SECTION 10: HELPER FUNCTIONS FOR LOCALIZATION
-- ============================================================================

-- Function to get plant common names for a specific language and country
CREATE OR REPLACE FUNCTION get_plant_names(
    p_plant_id UUID,
    p_language_id UUID,
    p_country_id UUID DEFAULT NULL
) RETURNS TABLE (
    common_name VARCHAR(200),
    is_primary BOOLEAN,
    is_country_specific BOOLEAN
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        pcn.common_name,
        pcn.is_primary,
        pcn.country_id IS NOT NULL AS is_country_specific
    FROM plant_common_names pcn
    WHERE pcn.plant_id = p_plant_id
      AND pcn.language_id = p_language_id
      AND (pcn.country_id = p_country_id OR 
           (p_country_id IS NOT NULL AND pcn.country_id IS NULL) OR
           p_country_id IS NULL)
    ORDER BY 
        pcn.country_id IS NOT NULL DESC, -- Country-specific first
        pcn.is_primary DESC,
        pcn.common_name;
END;
$$ LANGUAGE plpgsql;

-- Function to get translated characteristic value
CREATE OR REPLACE FUNCTION translate_characteristic(
    p_characteristic_type VARCHAR(100),
    p_characteristic_value VARCHAR(100),
    p_language_id UUID
) RETURNS VARCHAR(200) AS $$
DECLARE
    v_translated VARCHAR(200);
BEGIN
    SELECT translated_value INTO v_translated
    FROM characteristic_translations
    WHERE characteristic_type = p_characteristic_type
      AND characteristic_value = p_characteristic_value
      AND language_id = p_language_id;
    
    -- Return original if no translation found
    RETURN COALESCE(v_translated, p_characteristic_value);
END;
$$ LANGUAGE plpgsql;

-- Function to get localized plant description
CREATE OR REPLACE FUNCTION get_plant_description(
    p_plant_id UUID,
    p_description_type VARCHAR(50),
    p_language_id UUID,
    p_country_id UUID DEFAULT NULL
) RETURNS TEXT AS $$
DECLARE
    v_description TEXT;
BEGIN
    -- Try country-specific first
    IF p_country_id IS NOT NULL THEN
        SELECT content INTO v_description
        FROM plant_descriptions
        WHERE plant_id = p_plant_id
          AND description_type = p_description_type
          AND language_id = p_language_id
          AND country_id = p_country_id;
    END IF;
    
    -- Fall back to global for this language
    IF v_description IS NULL THEN
        SELECT content INTO v_description
        FROM plant_descriptions
        WHERE plant_id = p_plant_id
          AND description_type = p_description_type
          AND language_id = p_language_id
          AND country_id IS NULL;
    END IF;
    
    RETURN v_description;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- SECTION 11: VIEWS FOR COMMON LOCALIZATION QUERIES
-- ============================================================================

-- View for plant information in user's language
CREATE OR REPLACE VIEW v_plants_localized AS
SELECT 
    p.plant_id,
    p.full_botanical_name,
    u.user_id,
    u.preferred_language_id,
    (
        SELECT array_agg(pcn.common_name ORDER BY pcn.is_primary DESC)
        FROM plant_common_names pcn
        WHERE pcn.plant_id = p.plant_id
          AND pcn.language_id = u.preferred_language_id
    ) AS common_names,
    (
        SELECT pd.content
        FROM plant_descriptions pd
        WHERE pd.plant_id = p.plant_id
          AND pd.language_id = u.preferred_language_id
          AND pd.description_type = 'general'
        LIMIT 1
    ) AS description
FROM plants p
CROSS JOIN users u;

-- ============================================================================
-- SECTION 12: MIGRATION OF EXISTING DATA
-- ============================================================================

-- Migrate existing problem names to English translations
INSERT INTO plant_problems_i18n (problem_id, language_id, problem_name, symptoms, treatments, prevention)
SELECT 
    pp.problem_id,
    (SELECT language_id FROM languages WHERE language_code = 'en'),
    pp.problem_name,
    pp.symptoms,
    pp.treatments,
    pp.prevention
FROM plant_problems pp;

-- Migrate country names to English
INSERT INTO country_names_i18n (country_id, language_id, country_name)
SELECT 
    c.country_id,
    (SELECT language_id FROM languages WHERE language_code = 'en'),
    c.country_name
FROM countries c;

-- ============================================================================
-- SECTION 13: COMMENTS FOR DOCUMENTATION
-- ============================================================================

COMMENT ON TABLE plant_common_names IS 'Localized common names for plants with country and language context';
COMMENT ON TABLE plant_descriptions IS 'Localized descriptions for plants by type and language';
COMMENT ON TABLE characteristic_translations IS 'Translations for enum values and controlled vocabularies';
COMMENT ON TABLE plant_problems_i18n IS 'Localized problem descriptions, symptoms, and treatments';
COMMENT ON FUNCTION get_plant_names IS 'Returns localized plant names with country-specific precedence';
COMMENT ON FUNCTION translate_characteristic IS 'Translates characteristic values to user language';

-- ============================================================================
-- END OF MIGRATION
-- ============================================================================