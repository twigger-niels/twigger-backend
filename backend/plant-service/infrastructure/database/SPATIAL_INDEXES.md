# Spatial Indexes Documentation

## Overview
This document describes the required GIST (Generalized Search Tree) indexes for spatial operations in the plant database. These indexes are critical for performance of PostGIS spatial queries.

## Required GIST Indexes

### 1. Countries Table
**Index Name**: `idx_countries_boundary`
**Column**: `country_boundary` (geometry)
**Type**: GIST

```sql
CREATE INDEX idx_countries_boundary
ON countries USING GIST (country_boundary);
```

**Used By**:
- `FindByPoint(latitude, longitude)` - Find country containing a point
- Spatial joins with climate zones

**Performance Impact**:
- Without index: Sequential scan O(n) - ~500ms for 200 countries
- With index: Index scan O(log n) - ~5ms

---

### 2. Climate Zones Table
**Index Name**: `idx_climate_zones_geometry`
**Column**: `zone_geometry` (geometry)
**Type**: GIST

```sql
CREATE INDEX idx_climate_zones_geometry
ON climate_zones USING GIST (zone_geometry);
```

**Used By**:
- `FindByPoint(latitude, longitude, zoneSystem)` - Find zone containing a point
- `ST_Contains()` queries for zone membership
- Spatial joins with gardens

**Performance Impact**:
- Without index: Sequential scan O(n) - ~2000ms for 50,000 zones
- With index: Index scan O(log n) - ~10ms

---

### 3. Country Plants Table (Optional)
**Index Name**: `idx_country_plants_native_range`
**Column**: `native_range_geojson` (geometry)
**Type**: GIST

```sql
CREATE INDEX idx_country_plants_native_range
ON country_plants USING GIST (native_range_geojson);
```

**Used By**:
- Finding plants native to specific geographic regions
- Spatial analysis of plant distributions

**Performance Impact**:
- Useful for large datasets (>10,000 country-plant relationships)
- Optional for MVP

---

## Index Maintenance

### Checking Index Usage
```sql
-- Check if indexes exist
SELECT schemaname, tablename, indexname, indexdef
FROM pg_indexes
WHERE indexname LIKE '%boundary%' OR indexname LIKE '%geometry%';

-- Check index size
SELECT
    schemaname,
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) AS index_size
FROM pg_stat_user_indexes
WHERE indexrelname LIKE 'idx_%boundary%' OR indexrelname LIKE 'idx_%geometry%';

-- Check if indexes are being used
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan AS times_used,
    idx_tup_read AS tuples_read,
    idx_tup_fetch AS tuples_fetched
FROM pg_stat_user_indexes
WHERE indexrelname LIKE 'idx_%boundary%' OR indexrelname LIKE 'idx_%geometry%';
```

### Rebuilding Indexes
If indexes become bloated or corrupted:

```sql
-- Rebuild specific index
REINDEX INDEX idx_countries_boundary;
REINDEX INDEX idx_climate_zones_geometry;

-- Rebuild all indexes on a table
REINDEX TABLE countries;
REINDEX TABLE climate_zones;
```

### Monitoring Performance
Use `EXPLAIN ANALYZE` to verify index usage:

```sql
-- Should show "Index Scan using idx_countries_boundary"
EXPLAIN ANALYZE
SELECT * FROM countries
WHERE ST_Contains(country_boundary, ST_SetSRID(ST_MakePoint(-122.4194, 37.7749), 4326));

-- Should show "Index Scan using idx_climate_zones_geometry"
EXPLAIN ANALYZE
SELECT * FROM climate_zones
WHERE ST_Contains(zone_geometry, ST_SetSRID(ST_MakePoint(-122.4194, 37.7749), 4326))
  AND zone_system = 'USDA';
```

---

## Performance Characteristics

### Index Build Time
| Table | Rows | Index Build Time | Index Size |
|-------|------|------------------|------------|
| countries | 200 | ~50ms | ~2MB |
| climate_zones | 50,000 | ~5s | ~100MB |
| country_plants | 100,000 | ~10s | ~200MB |

### Query Performance Comparison
| Query Type | Without Index | With Index | Improvement |
|------------|---------------|------------|-------------|
| Point-in-polygon (countries) | 500ms | 5ms | 100x |
| Point-in-polygon (zones) | 2000ms | 10ms | 200x |
| Spatial join | 30s | 150ms | 200x |

---

## Troubleshooting

### Issue: Queries not using index
**Symptoms**: `EXPLAIN ANALYZE` shows "Seq Scan" instead of "Index Scan"

**Solutions**:
1. Verify index exists: `\d countries` in psql
2. Update statistics: `ANALYZE countries;`
3. Check query uses correct SRID: `ST_SetSRID(ST_MakePoint(lng, lat), 4326)`
4. Ensure geometry column is not NULL in query

### Issue: Index build fails
**Symptoms**: `ERROR: could not create unique index`

**Solutions**:
1. Check for NULL geometries: `SELECT COUNT(*) FROM countries WHERE country_boundary IS NULL;`
2. Validate geometries: `SELECT ST_IsValid(country_boundary) FROM countries;`
3. Fix invalid geometries: `UPDATE countries SET country_boundary = ST_MakeValid(country_boundary);`

---

## References
- PostGIS GIST Indexes: https://postgis.net/docs/using_postgis_dbmanagement.html#gist_indexes
- PostgreSQL Index Types: https://www.postgresql.org/docs/current/indexes-types.html
- ST_Contains Performance: https://postgis.net/docs/ST_Contains.html
