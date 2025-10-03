@echo off
REM Run garden service integration tests
REM Requires PostgreSQL with PostGIS running

echo ================================================
echo Garden Service Integration Tests
echo ================================================
echo.

REM Check if PostgreSQL is running
echo [1/4] Checking PostgreSQL connection...
psql -U postgres -c "SELECT version();" > nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: PostgreSQL is not running or not accessible
    echo Please start PostgreSQL and try again
    exit /b 1
)
echo ✓ PostgreSQL is running
echo.

REM Check if test database exists, create if not
echo [2/4] Setting up test database...
psql -U postgres -lqt | findstr plantdb_test > nul
if %errorlevel% neq 0 (
    echo Creating test database plantdb_test...
    psql -U postgres -c "CREATE DATABASE plantdb_test;" > nul 2>&1
)

REM Enable PostGIS
psql -U postgres -d plantdb_test -c "CREATE EXTENSION IF NOT EXISTS postgis;" > nul 2>&1
psql -U postgres -d plantdb_test -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";" > nul 2>&1
echo ✓ Test database ready
echo.

REM Set test database URL
set TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5432/plantdb_test?sslmode=disable

REM Run integration tests
echo [3/4] Running garden repository tests...
echo.
cd /d "%~dp0\..\backend\garden-service"
go test -v -tags=integration ./infrastructure/persistence/...

set TEST_EXIT_CODE=%errorlevel%

echo.
echo [4/4] Test Summary
echo ================================================
if %TEST_EXIT_CODE% equ 0 (
    echo ✓ All tests passed!
) else (
    echo ✗ Some tests failed. See output above.
)
echo ================================================

exit /b %TEST_EXIT_CODE%
