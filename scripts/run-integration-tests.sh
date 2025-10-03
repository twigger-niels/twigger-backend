#!/bin/bash

# Plant Service Integration Test Runner
# Starts test database, runs tests, and cleans up

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="$ROOT_DIR/docker-compose.test.yml"
TEST_DB_HOST="${TEST_DB_HOST:-localhost}"
TEST_DB_PORT="${TEST_DB_PORT:-5433}"
TEST_DB_NAME="${TEST_DB_NAME:-plantdb_test}"
TEST_DB_USER="${TEST_DB_USER:-plant_api_test}"
TEST_DB_PASSWORD="${TEST_DB_PASSWORD:-test_password_123}"

# Parse command line arguments
RUN_BENCHMARKS=false
VERBOSE=false
SPECIFIC_TEST=""
PARALLEL=1

while [[ $# -gt 0 ]]; do
    case $1 in
        -b|--bench)
            RUN_BENCHMARKS=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -t|--test)
            SPECIFIC_TEST="$2"
            shift 2
            ;;
        -p|--parallel)
            PARALLEL="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -b, --bench          Run benchmarks instead of tests"
            echo "  -v, --verbose        Enable verbose output"
            echo "  -t, --test NAME      Run specific test by name"
            echo "  -p, --parallel N     Run tests in parallel (default: 1)"
            echo "  -h, --help           Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                                    # Run all integration tests"
            echo "  $0 -v                                 # Run with verbose output"
            echo "  $0 -t FindByID                        # Run specific test"
            echo "  $0 -b                                 # Run benchmarks"
            echo "  $0 -p 4                               # Run tests in parallel"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

echo -e "${GREEN}=== Plant Service Integration Tests ===${NC}"
echo ""

# Function to check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        echo -e "${RED}Error: Docker is not running${NC}"
        echo "Please start Docker and try again"
        exit 1
    fi
}

# Function to start test database
start_test_db() {
    echo -e "${YELLOW}Starting test database...${NC}"

    cd "$ROOT_DIR"
    docker-compose -f "$COMPOSE_FILE" up -d

    echo -e "${YELLOW}Waiting for database to be ready...${NC}"

    # Wait for database to be healthy
    max_attempts=30
    attempt=0

    while [ $attempt -lt $max_attempts ]; do
        if docker-compose -f "$COMPOSE_FILE" exec -T postgres-test pg_isready -U "$TEST_DB_USER" -d "$TEST_DB_NAME" > /dev/null 2>&1; then
            echo -e "${GREEN}✓ Database is ready${NC}"
            break
        fi

        attempt=$((attempt + 1))
        echo -n "."
        sleep 1
    done

    if [ $attempt -eq $max_attempts ]; then
        echo -e "${RED}Error: Database failed to start${NC}"
        docker-compose -f "$COMPOSE_FILE" logs postgres-test
        exit 1
    fi

    # Wait a bit more for PostGIS extension
    sleep 2

    # Verify PostGIS
    echo -e "${YELLOW}Verifying PostGIS extension...${NC}"
    if docker-compose -f "$COMPOSE_FILE" exec -T postgres-test psql -U "$TEST_DB_USER" -d "$TEST_DB_NAME" -c "SELECT PostGIS_version();" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PostGIS is available${NC}"
    else
        echo -e "${RED}Error: PostGIS extension not available${NC}"
        exit 1
    fi

    echo ""
}

# Function to stop test database
stop_test_db() {
    echo ""
    echo -e "${YELLOW}Stopping test database...${NC}"
    cd "$ROOT_DIR"
    docker-compose -f "$COMPOSE_FILE" down
    echo -e "${GREEN}✓ Database stopped${NC}"
}

# Function to run tests
run_tests() {
    echo -e "${GREEN}Running integration tests...${NC}"
    echo ""

    cd "$ROOT_DIR"

    # Build test command
    TEST_CMD="go test -tags=integration"

    if [ "$VERBOSE" = true ]; then
        TEST_CMD="$TEST_CMD -v"
    fi

    if [ -n "$SPECIFIC_TEST" ]; then
        TEST_CMD="$TEST_CMD -run $SPECIFIC_TEST"
    fi

    if [ "$PARALLEL" -gt 1 ]; then
        TEST_CMD="$TEST_CMD -parallel $PARALLEL"
    fi

    TEST_CMD="$TEST_CMD ./backend/plant-service/infrastructure/database/..."

    echo -e "${YELLOW}Command: $TEST_CMD${NC}"
    echo ""

    # Export test environment variables
    export TEST_DB_HOST
    export TEST_DB_PORT
    export TEST_DB_NAME
    export TEST_DB_USER
    export TEST_DB_PASSWORD
    export TEST_DB_SSLMODE="disable"

    # Run tests
    if eval "$TEST_CMD"; then
        echo ""
        echo -e "${GREEN}✓ All tests passed${NC}"
        return 0
    else
        echo ""
        echo -e "${RED}✗ Some tests failed${NC}"
        return 1
    fi
}

# Function to run benchmarks
run_benchmarks() {
    echo -e "${GREEN}Running benchmarks...${NC}"
    echo ""

    cd "$ROOT_DIR"

    # Build benchmark command
    BENCH_CMD="go test -tags=integration -bench=. -benchmem"

    if [ "$VERBOSE" = true ]; then
        BENCH_CMD="$BENCH_CMD -v"
    fi

    if [ -n "$SPECIFIC_TEST" ]; then
        BENCH_CMD="$BENCH_CMD -bench=$SPECIFIC_TEST"
    fi

    BENCH_CMD="$BENCH_CMD ./backend/plant-service/infrastructure/database/..."

    echo -e "${YELLOW}Command: $BENCH_CMD${NC}"
    echo ""

    # Export test environment variables
    export TEST_DB_HOST
    export TEST_DB_PORT
    export TEST_DB_NAME
    export TEST_DB_USER
    export TEST_DB_PASSWORD
    export TEST_DB_SSLMODE="disable"

    # Run benchmarks
    if eval "$BENCH_CMD"; then
        echo ""
        echo -e "${GREEN}✓ Benchmarks completed${NC}"
        return 0
    else
        echo ""
        echo -e "${RED}✗ Benchmarks failed${NC}"
        return 1
    fi
}

# Main execution
main() {
    # Trap to ensure cleanup on exit
    trap stop_test_db EXIT

    check_docker
    start_test_db

    if [ "$RUN_BENCHMARKS" = true ]; then
        run_benchmarks
    else
        run_tests
    fi
}

# Run main function
main
