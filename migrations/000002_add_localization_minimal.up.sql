-- Minimal localization tables needed for API Gateway Part 5
-- This is a subset of 005_add_localization.sql with only tables that exist

-- Plant common names with country+language context
CREATE TABLE IF NOT EXISTS plant_common_names (
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
CREATE INDEX IF NOT EXISTS idx_plant_common_names_plant ON plant_common_names(plant_id);
CREATE INDEX IF NOT EXISTS idx_plant_common_names_language ON plant_common_names(language_id);
CREATE INDEX IF NOT EXISTS idx_plant_common_names_country ON plant_common_names(country_id);
CREATE INDEX IF NOT EXISTS idx_plant_common_names_lookup ON plant_common_names(plant_id, language_id, country_id);
CREATE INDEX IF NOT EXISTS idx_plant_common_names_name_lang ON plant_common_names(language_id, common_name);

-- Plant descriptions with country+language context
CREATE TABLE IF NOT EXISTS plant_descriptions (
    description_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    plant_id UUID NOT NULL REFERENCES plants(plant_id) ON DELETE CASCADE,
    language_id UUID NOT NULL REFERENCES languages(language_id),
    country_id UUID REFERENCES countries(country_id), -- NULL for universal

    -- Different types of descriptions
    short_description TEXT, -- 1-2 sentences
    full_description TEXT, -- Comprehensive description
    care_instructions TEXT, -- How to care for this plant

    -- Context
    is_primary BOOLEAN DEFAULT FALSE, -- Primary description for this language/country
    source_id UUID REFERENCES data_sources(source_id),

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_plant_desc_per_context
        UNIQUE(plant_id, language_id, country_id)
);

CREATE INDEX IF NOT EXISTS idx_plant_descriptions_plant ON plant_descriptions(plant_id);
CREATE INDEX IF NOT EXISTS idx_plant_descriptions_lookup ON plant_descriptions(plant_id, language_id, country_id);

-- Characteristic translations (for enums like sun_requirement, soil_type, etc.)
CREATE TABLE IF NOT EXISTS characteristic_translations (
    translation_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    language_id UUID NOT NULL REFERENCES languages(language_id),

    -- The characteristic being translated
    characteristic_category VARCHAR(50) NOT NULL, -- e.g., 'sun_requirement', 'soil_type'
    characteristic_value VARCHAR(100) NOT NULL, -- e.g., 'full_sun', 'loamy'

    -- Translation
    translated_label VARCHAR(200) NOT NULL,
    translated_description TEXT,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_characteristic_translation
        UNIQUE(language_id, characteristic_category, characteristic_value)
);

CREATE INDEX IF NOT EXISTS idx_characteristic_translations_lookup
    ON characteristic_translations(language_id, characteristic_category, characteristic_value);

-- Companion benefits localization
CREATE TABLE IF NOT EXISTS companion_benefits_i18n (
    benefit_i18n_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    relationship_id UUID NOT NULL REFERENCES companion_relationships(relationship_id) ON DELETE CASCADE,
    language_id UUID NOT NULL REFERENCES languages(language_id),

    -- Localized benefit description
    benefit_description TEXT NOT NULL,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_companion_benefit_lang
        UNIQUE(relationship_id, language_id)
);

CREATE INDEX IF NOT EXISTS idx_companion_benefits_i18n_relationship
    ON companion_benefits_i18n(relationship_id);
CREATE INDEX IF NOT EXISTS idx_companion_benefits_i18n_language
    ON companion_benefits_i18n(language_id);

-- Country names localization
CREATE TABLE IF NOT EXISTS country_names_i18n (
    country_name_i18n_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    country_id UUID NOT NULL REFERENCES countries(country_id) ON DELETE CASCADE,
    language_id UUID NOT NULL REFERENCES languages(language_id),

    -- Localized country name
    country_name VARCHAR(100) NOT NULL,
    country_name_formal VARCHAR(200), -- Formal/official name

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_country_name_lang
        UNIQUE(country_id, language_id)
);

CREATE INDEX IF NOT EXISTS idx_country_names_i18n_country
    ON country_names_i18n(country_id);
CREATE INDEX IF NOT EXISTS idx_country_names_i18n_language
    ON country_names_i18n(language_id);
