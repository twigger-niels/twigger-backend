# Plant Service Integration Tests

This directory contains integration test helpers and utilities for testing the Plant Service against a real PostgreSQL database with PostGIS.

## Prerequisites

- Docker and Docker Compose installed
- Go 1.21+ installed
- PostgreSQL client tools (optional, for manual verification)

## Quick Start

### 1. Start Test Database

```bash
# From repository root
docker-compose -f docker-compose.test.yml up -d

# Wait for database to be ready
docker-compose -f docker-compose.test.yml ps

# Verify PostGIS is installed
docker-compose -f docker-compose.test.yml exec postgres-test psql -U plant_api_test -d plantdb_test -c "SELECT PostGIS_version();"
```

### 2. Run Integration Tests

```bash
# Run all integration tests
go test -tags=integration ./backend/plant-service/infrastructure/database/...

# Run specific test
go test -tags=integration -run TestPostgresPlantRepository_FindByID_Integration ./backend/plant-service/infrastructure/database/

# Run with verbose output
go test -tags=integration -v ./backend/plant-service/infrastructure/database/...

# Run tests in parallel
go test -tags=integration -parallel 4 ./backend/plant-service/infrastructure/database/...
```

### 3. Run Benchmarks

```bash
# Run all benchmarks
go test -tags=integration -bench=. ./backend/plant-service/infrastructure/database/

# Run specific benchmark
go test -tags=integration -bench=BenchmarkPlantRepository_FindByIDs_BatchLoading ./backend/plant-service/infrastructure/database/

# Run benchmarks with memory profiling
go test -tags=integration -bench=. -benchmem ./backend/plant-service/infrastructure/database/

# Run benchmarks for longer duration (more accurate)
go test -tags=integration -bench=. -benchtime=10s ./backend/plant-service/infrastructure/database/
```

### 4. Stop Test Database

```bash
# Stop and remove containers
docker-compose -f docker-compose.test.yml down

# Stop and remove containers + volumes (clean slate)
docker-compose -f docker-compose.test.yml down -v
```

## Test Database Configuration

The test database uses the following configuration:

- **Host:** localhost
- **Port:** 5433 (different from default 5432 to avoid conflicts)
- **Database:** plantdb_test
- **User:** plant_api_test
- **Password:** test_password_123

You can override these values using environment variables:

```bash
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5433
export TEST_DB_USER=plant_api_test
export TEST_DB_PASSWORD=test_password_123
export TEST_DB_NAME=plantdb_test
export TEST_DB_SSLMODE=disable

go test -tags=integration ./backend/plant-service/infrastructure/database/...
```

## Test Structure

### Test Helpers (`test_helpers.go`)

Provides utilities for:
- **SetupTestDB()**: Creates database connection and runs migrations
- **TeardownTestDB()**: Cleans up test data and closes connection
- **SeedTestLanguages()**: Inserts test language data (en, es, fr, de)
- **SeedTestCountries()**: Inserts test country data (US, MX, GB, DE)
- **SeedTestPlantHierarchy()**: Creates test family/genus/species hierarchy

### Integration Tests

- **postgres_plant_repository_integration_test.go**: CRUD operations, localization, search
- **postgres_plant_repository_bench_test.go**: Performance benchmarks

### Build Tags

All integration tests use the `// +build integration` build tag to separate them from unit tests.

Run unit tests only (default):
```bash
go test ./backend/plant-service/domain/service/...
```

Run integration tests only:
```bash
go test -tags=integration ./backend/plant-service/infrastructure/database/...
```

## Test Coverage

### Integration Test Coverage

- ✅ **FindByID**: Single plant retrieval with localization
- ✅ **FindByIDs**: Batch loading with N+1 prevention
- ✅ **Create**: Plant creation with validation
- ✅ **Update**: Plant updates with timestamp tracking
- ✅ **Delete**: Plant deletion with cascade
- ✅ **FindByBotanicalName**: Case-insensitive botanical name search
- ✅ **Search**: Full-text search by botanical and common names
- ✅ **Localization**: 4-tier fallback (country+lang → lang → en → empty)
- ✅ **Batch Loading**: Performance comparison N+1 vs batch

### Benchmark Coverage

- **Batch Size Scaling**: Tests 10, 20, 50, 100 plants per batch
- **N+1 Comparison**: Measures N+1 problem vs batch loading
- **Search Performance**: Botanical and common name searches
- **Localization Performance**: Language-specific lookups

## Expected Benchmark Results

Based on ADR-008, batch loading should show:

- **96% query reduction**: 51 queries → 2 queries for 50 plants
- **~200ms savings**: For typical 20-result search pages
- **90% database load reduction**: Linear scaling vs O(n)

Example output:
```
BenchmarkPlantRepository_FindByIDs_BatchLoading/BatchSize10-8    500   3.2 ms/op
BenchmarkPlantRepository_FindByIDs_BatchLoading/BatchSize20-8    300   4.5 ms/op
BenchmarkPlantRepository_FindByIDs_BatchLoading/BatchSize50-8    200   8.1 ms/op
BenchmarkPlantRepository_FindByIDs_BatchLoading/BatchSize100-8   100  14.2 ms/op

BenchmarkPlantRepository_FindByID_SingleVsBatch/N+1_Problem-8     10  250.0 ms/op
BenchmarkPlantRepository_FindByID_SingleVsBatch/BatchLoading-8   200    8.5 ms/op
```

## Troubleshooting

### Database Connection Issues

```bash
# Check if database is running
docker-compose -f docker-compose.test.yml ps

# View database logs
docker-compose -f docker-compose.test.yml logs postgres-test

# Restart database
docker-compose -f docker-compose.test.yml restart postgres-test
```

### Migration Issues

```bash
# Manually run migrations
docker-compose -f docker-compose.test.yml exec postgres-test psql -U plant_api_test -d plantdb_test -f /docker-entrypoint-initdb.d/migrations/000001_comprehensive_plant_schema.up.sql

# Check if tables exist
docker-compose -f docker-compose.test.yml exec postgres-test psql -U plant_api_test -d plantdb_test -c "\dt"

# Check if PostGIS functions exist
docker-compose -f docker-compose.test.yml exec postgres-test psql -U plant_api_test -d plantdb_test -c "SELECT proname FROM pg_proc WHERE proname LIKE '%st_%';"
```

### Test Failures

```bash
# Run with verbose output
go test -tags=integration -v ./backend/plant-service/infrastructure/database/...

# Run specific failing test
go test -tags=integration -run TestPostgresPlantRepository_Localization_Integration ./backend/plant-service/infrastructure/database/ -v

# Clean database and retry
docker-compose -f docker-compose.test.yml down -v
docker-compose -f docker-compose.test.yml up -d
# Wait 10 seconds for initialization
go test -tags=integration ./backend/plant-service/infrastructure/database/...
```

## CI/CD Integration

Example GitHub Actions workflow:

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  integration-test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:17
        env:
          POSTGRES_DB: plantdb_test
          POSTGRES_USER: plant_api_test
          POSTGRES_PASSWORD: test_password_123
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5433:5432

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install PostGIS
        run: |
          docker exec ${{ job.services.postgres.id }} apt-get update
          docker exec ${{ job.services.postgres.id }} apt-get install -y postgis

      - name: Run Integration Tests
        run: go test -tags=integration ./backend/plant-service/infrastructure/database/...
        env:
          TEST_DB_HOST: localhost
          TEST_DB_PORT: 5433
```

## Performance Testing

### Load Testing

```bash
# Run benchmarks with different load patterns
go test -tags=integration -bench=BenchmarkPlantRepository_FindByIDs_BatchLoading -benchtime=100x ./backend/plant-service/infrastructure/database/

# CPU profiling
go test -tags=integration -bench=. -cpuprofile=cpu.prof ./backend/plant-service/infrastructure/database/
go tool pprof cpu.prof

# Memory profiling
go test -tags=integration -bench=. -memprofile=mem.prof ./backend/plant-service/infrastructure/database/
go tool pprof mem.prof
```

### Query Analysis

```bash
# Enable query logging in PostgreSQL
docker-compose -f docker-compose.test.yml exec postgres-test psql -U plant_api_test -d plantdb_test -c "ALTER SYSTEM SET log_statement = 'all';"
docker-compose -f docker-compose.test.yml restart postgres-test

# Run test and view queries
go test -tags=integration -run TestPostgresPlantRepository_FindByIDs_Integration ./backend/plant-service/infrastructure/database/ -v

# View logs
docker-compose -f docker-compose.test.yml logs postgres-test | grep "SELECT\|INSERT\|UPDATE"
```

## Contributing

When adding new repository methods:

1. Add integration test in `postgres_*_integration_test.go`
2. Add benchmark if performance-critical in `postgres_*_bench_test.go`
3. Update this README with new test coverage
4. Ensure tests clean up after themselves (use `TeardownTestDB`)
5. Use build tag `// +build integration`

## References

- [Testing in Go](https://golang.org/pkg/testing/)
- [PostGIS Documentation](https://postgis.net/documentation/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [ADR-008: Batch Loading](../../../../epics/plant-database/architecture.md#adr-008-batch-loading-for-n1-query-prevention)
