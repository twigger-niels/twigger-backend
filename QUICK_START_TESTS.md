# Quick Start - Integration Tests

## Current Situation

The integration tests are ready to run, but we need a PostgreSQL database with PostGIS. You have two options:

## Option 1: Use Docker Desktop (Recommended)

### Step 1: Start Docker Desktop
1. Open Docker Desktop application on Windows
2. Wait for it to fully start (whale icon in system tray should be stable)

### Step 2: Run Tests
```bash
# Using the test runner script (does everything automatically)
scripts\run-integration-tests.bat -v

# OR manually
docker-compose -f docker-compose.test.yml up -d
timeout /t 15 /nobreak
go test -tags=integration -v ./backend/plant-service/infrastructure/database/...
```

### Expected Output
```
=== RUN   TestPostgresPlantRepository_FindByID_Integration
=== RUN   TestPostgresPlantRepository_FindByID_Integration/successful_retrieval_with_English_common_names
=== RUN   TestPostgresPlantRepository_FindByID_Integration/plant_not_found
=== RUN   TestPostgresPlantRepository_FindByID_Integration/invalid_plant_ID
--- PASS: TestPostgresPlantRepository_FindByID_Integration (0.15s)
    --- PASS: TestPostgresPlantRepository_FindByID_Integration/successful_retrieval_with_English_common_names (0.05s)
    --- PASS: TestPostgresPlantRepository_FindByID_Integration/plant_not_found (0.02s)
    --- PASS: TestPostgresPlantRepository_FindByID_Integration/invalid_plant_ID (0.01s)
...
PASS
ok  	twigger-backend/backend/plant-service/infrastructure/database	2.345s
```

## Option 2: Use Existing PostgreSQL Installation

If you have PostgreSQL installed locally:

### Step 1: Install PostGIS Extension
```sql
-- Connect to your database
psql -U postgres

-- Install PostGIS
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

### Step 2: Configure Test Environment
```bash
# Set environment variables
set TEST_DB_HOST=localhost
set TEST_DB_PORT=5432
set TEST_DB_USER=postgres
set TEST_DB_PASSWORD=your_password
set TEST_DB_NAME=plantdb_test
set TEST_DB_SSLMODE=disable

# Run tests
go test -tags=integration -v ./backend/plant-service/infrastructure/database/...
```

## Option 3: Quick Cloud Database Test (If Accessible)

If your cloud database at 162.222.181.26 is accessible:

```bash
set USE_CLOUD_DB=true
go test -tags=integration -run TestPostgresPlantRepository_FindByID_Integration -v ./backend/plant-service/infrastructure/database/
```

**Note**: Currently getting connection timeout - the database may be:
- Behind a firewall
- Requires VPN connection
- IP whitelist needed
- Not currently running

## Recommended Next Steps

**For immediate testing:**

1. **Start Docker Desktop** (easiest option)
   - Open Docker Desktop
   - Wait for startup (~30 seconds)
   - Run: `scripts\run-integration-tests.bat -v`

2. **Review test results**
   - All 11 test suites should pass
   - Expected runtime: ~3-5 seconds
   - Database will auto-cleanup

3. **Run benchmarks**
   ```bash
   scripts\run-integration-tests.bat -b
   ```
   - Measures N+1 query performance
   - Expected: 96% query reduction
   - Shows batch loading improvements

## What Happens Next

Once tests pass:
- ✅ Integration tests validated
- ✅ Localization confirmed working
- ✅ Batch loading performance verified
- ✅ Part 2 testing complete (95% → 100%)

Then we can move to:
- Data import scripts
- API endpoint documentation
- Service-level unit tests
- Mark Part 2 as complete! 🎉

## Troubleshooting

### Docker Desktop Won't Start
- Restart your computer
- Reinstall Docker Desktop
- Check Windows Hyper-V is enabled

### Port 5433 Already in Use
```bash
# Find what's using the port
netstat -ano | findstr :5433

# Stop the conflicting service or change port in docker-compose.test.yml
```

### Tests Fail After Database Starts
```bash
# Clean database and retry
docker-compose -f docker-compose.test.yml down -v
docker-compose -f docker-compose.test.yml up -d
timeout /t 15 /nobreak
go test -tags=integration -v ./backend/plant-service/infrastructure/database/...
```

## Current Status

✅ Integration test infrastructure complete
✅ 11 comprehensive test suites implemented
✅ Performance benchmarks ready
✅ Documentation complete

⏸️ **Waiting for**: Docker Desktop to start OR PostgreSQL connection

🎯 **Next action**: Start Docker Desktop and run `scripts\run-integration-tests.bat -v`
