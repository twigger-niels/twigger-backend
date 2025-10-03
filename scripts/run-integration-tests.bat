@echo off
REM Plant Service Integration Test Runner for Windows

setlocal enabledelayedexpansion

REM Get script directory
set SCRIPT_DIR=%~dp0
set ROOT_DIR=%SCRIPT_DIR%..

REM Configuration
set COMPOSE_FILE=%ROOT_DIR%\docker-compose.test.yml
set TEST_DB_HOST=localhost
set TEST_DB_PORT=5433
set TEST_DB_NAME=plantdb_test
set TEST_DB_USER=plant_api_test
set TEST_DB_PASSWORD=test_password_123
set TEST_DB_SSLMODE=disable

REM Parse command line arguments
set RUN_BENCHMARKS=false
set VERBOSE=false
set SPECIFIC_TEST=
set PARALLEL=1

:parse_args
if "%~1"=="" goto end_parse
if /i "%~1"=="-b" set RUN_BENCHMARKS=true& shift & goto parse_args
if /i "%~1"=="--bench" set RUN_BENCHMARKS=true& shift & goto parse_args
if /i "%~1"=="-v" set VERBOSE=true& shift & goto parse_args
if /i "%~1"=="--verbose" set VERBOSE=true& shift & goto parse_args
if /i "%~1"=="-t" set SPECIFIC_TEST=%~2& shift & shift & goto parse_args
if /i "%~1"=="--test" set SPECIFIC_TEST=%~2& shift & shift & goto parse_args
if /i "%~1"=="-p" set PARALLEL=%~2& shift & shift & goto parse_args
if /i "%~1"=="--parallel" set PARALLEL=%~2& shift & shift & goto parse_args
if /i "%~1"=="-h" goto show_help
if /i "%~1"=="--help" goto show_help
echo Unknown option: %~1
goto show_help

:end_parse

echo === Plant Service Integration Tests ===
echo.

REM Check if Docker is running
docker info >nul 2>&1
if errorlevel 1 (
    echo Error: Docker is not running
    echo Please start Docker Desktop and try again
    exit /b 1
)

REM Start test database
echo Starting test database...
cd /d "%ROOT_DIR%"
docker-compose -f "%COMPOSE_FILE%" up -d

echo Waiting for database to be ready...
set /a attempts=0
set /a max_attempts=30

:wait_loop
docker-compose -f "%COMPOSE_FILE%" exec -T postgres-test pg_isready -U %TEST_DB_USER% -d %TEST_DB_NAME% >nul 2>&1
if %errorlevel%==0 goto db_ready

set /a attempts+=1
if %attempts% geq %max_attempts% (
    echo Error: Database failed to start
    docker-compose -f "%COMPOSE_FILE%" logs postgres-test
    goto cleanup
)

echo|set /p=.
timeout /t 1 /nobreak >nul
goto wait_loop

:db_ready
echo.
echo Database is ready
timeout /t 2 /nobreak >nul

REM Verify PostGIS
echo Verifying PostGIS extension...
docker-compose -f "%COMPOSE_FILE%" exec -T postgres-test psql -U %TEST_DB_USER% -d %TEST_DB_NAME% -c "SELECT PostGIS_version();" >nul 2>&1
if errorlevel 1 (
    echo Error: PostGIS extension not available
    goto cleanup
)
echo PostGIS is available
echo.

REM Build test command
if "%RUN_BENCHMARKS%"=="true" (
    set TEST_CMD=go test -tags=integration -bench=. -benchmem
) else (
    set TEST_CMD=go test -tags=integration
)

if "%VERBOSE%"=="true" (
    set TEST_CMD=!TEST_CMD! -v
)

if not "%SPECIFIC_TEST%"=="" (
    if "%RUN_BENCHMARKS%"=="true" (
        set TEST_CMD=!TEST_CMD! -bench=%SPECIFIC_TEST%
    ) else (
        set TEST_CMD=!TEST_CMD! -run %SPECIFIC_TEST%
    )
)

if not "%PARALLEL%"=="1" (
    set TEST_CMD=!TEST_CMD! -parallel %PARALLEL%
)

set TEST_CMD=!TEST_CMD! .\backend\plant-service\infrastructure\database\...

REM Run tests or benchmarks
if "%RUN_BENCHMARKS%"=="true" (
    echo Running benchmarks...
) else (
    echo Running integration tests...
)
echo.
echo Command: !TEST_CMD!
echo.

REM Execute tests
call !TEST_CMD!
set TEST_RESULT=%errorlevel%

echo.
if %TEST_RESULT%==0 (
    if "%RUN_BENCHMARKS%"=="true" (
        echo Benchmarks completed successfully
    ) else (
        echo All tests passed
    )
) else (
    if "%RUN_BENCHMARKS%"=="true" (
        echo Benchmarks failed
    ) else (
        echo Some tests failed
    )
)

REM Cleanup
:cleanup
echo.
echo Stopping test database...
docker-compose -f "%COMPOSE_FILE%" down
echo Database stopped

exit /b %TEST_RESULT%

:show_help
echo Usage: %~nx0 [OPTIONS]
echo.
echo Options:
echo   -b, --bench          Run benchmarks instead of tests
echo   -v, --verbose        Enable verbose output
echo   -t, --test NAME      Run specific test by name
echo   -p, --parallel N     Run tests in parallel (default: 1)
echo   -h, --help           Show this help message
echo.
echo Examples:
echo   %~nx0                                    # Run all integration tests
echo   %~nx0 -v                                 # Run with verbose output
echo   %~nx0 -t FindByID                        # Run specific test
echo   %~nx0 -b                                 # Run benchmarks
echo   %~nx0 -p 4                               # Run tests in parallel
exit /b 0
