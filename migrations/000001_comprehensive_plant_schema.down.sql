-- Drop all tables in reverse order (respecting foreign key dependencies)
DROP TABLE IF EXISTS companion_relationships;
DROP TABLE IF EXISTS garden_plants;
DROP TABLE IF EXISTS garden_features;
DROP TABLE IF EXISTS garden_zones;
DROP TABLE IF EXISTS gardens;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS physical_characteristics;
DROP TABLE IF EXISTS growing_conditions_assertions;
DROP TABLE IF EXISTS country_plants;
DROP TABLE IF EXISTS plant_synonyms;
DROP TABLE IF EXISTS plants;
DROP TABLE IF EXISTS cultivars;
DROP TABLE IF EXISTS plant_species;
DROP TABLE IF EXISTS plant_genera;
DROP TABLE IF EXISTS plant_families;
DROP TABLE IF EXISTS data_sources;
DROP TABLE IF EXISTS languages;
DROP TABLE IF EXISTS climate_zones;
DROP TABLE IF EXISTS countries;

-- Drop composite types
DROP TYPE IF EXISTS size_range;
DROP TYPE IF EXISTS ph_range;
DROP TYPE IF EXISTS temp_range;

-- Drop enum types
DROP TYPE IF EXISTS growth_rate;
DROP TYPE IF EXISTS soil_drainage;
DROP TYPE IF EXISTS water_needs;
DROP TYPE IF EXISTS sun_requirement;
DROP TYPE IF EXISTS season;
DROP TYPE IF EXISTS confidence_level;
DROP TYPE IF EXISTS plant_type;

-- Drop domains
DROP DOMAIN IF EXISTS rating;
DROP DOMAIN IF EXISTS ph_value;
DROP DOMAIN IF EXISTS percentage;
DROP DOMAIN IF EXISTS years;
DROP DOMAIN IF EXISTS hours;
DROP DOMAIN IF EXISTS days;
DROP DOMAIN IF EXISTS area_m2;
DROP DOMAIN IF EXISTS weight_kg;
DROP DOMAIN IF EXISTS weight_g;
DROP DOMAIN IF EXISTS length_m;
DROP DOMAIN IF EXISTS length_cm;
DROP DOMAIN IF EXISTS length_mm;
DROP DOMAIN IF EXISTS temperature_c;