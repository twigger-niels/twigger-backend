#!/bin/bash
# Run garden service integration tests
# Requires PostgreSQL with PostGIS running

set -e

echo "================================================"
echo "Garden Service Integration Tests"
echo "================================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if PostgreSQL is running
echo "[1/4] Checking PostgreSQL connection..."
if ! psql -U postgres -c "SELECT version();" > /dev/null 2>&1; then
    echo -e "${RED}ERROR: PostgreSQL is not running or not accessible${NC}"
    echo "Please start PostgreSQL and try again"
    exit 1
fi
echo -e "${GREEN}✓ PostgreSQL is running${NC}"
echo ""

# Check if test database exists, create if not
echo "[2/4] Setting up test database..."
if ! psql -U postgres -lqt | cut -d \| -f 1 | grep -qw plantdb_test; then
    echo "Creating test database plantdb_test..."
    psql -U postgres -c "CREATE DATABASE plantdb_test;" > /dev/null 2>&1 || true
fi

# Enable PostGIS
psql -U postgres -d plantdb_test -c "CREATE EXTENSION IF NOT EXISTS postgis;" > /dev/null 2>&1 || true
psql -U postgres -d plantdb_test -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";" > /dev/null 2>&1 || true
echo -e "${GREEN}✓ Test database ready${NC}"
echo ""

# Set test database URL
export TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/plantdb_test?sslmode=disable"

# Run integration tests
echo "[3/4] Running garden repository tests..."
echo ""
cd "$(dirname "$0")/../backend/garden-service"
go test -v -tags=integration ./infrastructure/persistence/...

TEST_EXIT_CODE=$?

echo ""
echo "[4/4] Test Summary"
echo "================================================"
if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
else
    echo -e "${RED}✗ Some tests failed. See output above.${NC}"
fi
echo "================================================"

exit $TEST_EXIT_CODE
