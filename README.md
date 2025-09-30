# Twigger Backend

Plant database backend system with PostgreSQL + PostGIS and Go.

## Prerequisites

- Go 1.21+
- Google Cloud SDK
- Cloud SQL Proxy
- Access to the Cloud SQL instance: `twigger:us-central1:dev-twigger-db1`

## Quick Start

1. **Set up authentication:**
   ```bash
   gcloud auth login
   gcloud config set project YOUR_PROJECT_ID
   ```

2. **Configure Cloud SQL authorized networks:**
   - Go to [Cloud SQL Console](https://console.cloud.google.com/sql)
   - Select your instance `dev-twigger-db1`
   - Go to **Connections** → **Authorized networks**
   - Add your IP: `82.217.141.244/32` (your current IP)

2. **Install Cloud SQL Proxy:**
   ```bash
   make install-proxy
   ```

3. **Start Cloud SQL Proxy** (in a separate terminal):
   ```bash
   make proxy
   ```

4. **Run migrations:**
   ```bash
   make migrate-up
   ```

5. **Start the development server:**
   ```bash
   make dev
   ```

6. **Test the health endpoint:**
   ```bash
   curl http://localhost:8080/health
   ```

## Available Commands

- `make help` - Show available commands
- `make proxy` - Start Cloud SQL Proxy
- `make dev` - Run development server
- `make build` - Build the application
- `make test` - Run tests
- `make migrate-up` - Run database migrations
- `make migrate-down` - Rollback one migration
- `make migrate-reset` - Reset database (all migrations)
- `make clean` - Clean build artifacts

## Database Schema

The comprehensive plant database schema includes:

**Core Infrastructure:**
- Countries and climate zones with spatial boundaries
- Multi-language support
- Data source tracking and reliability scoring

**Plant Taxonomy:**
- Complete taxonomic hierarchy (families → genera → species → cultivars)
- Scientific naming with synonyms
- Plant type classification

**Growing Conditions:**
- Country-specific plant information
- Environmental requirements (sun, water, soil, pH)
- Climate zone mappings (USDA, RHS)
- Spatial native range data

**User Gardens:**
- Spatial garden boundaries with PostGIS
- Garden zones/beds with characteristics
- Garden features (trees, buildings) for shade analysis
- Plant placement tracking with health status

**Advanced Features:**
- Companion planting relationships
- Physical characteristics with JSONB flexibility
- Measurement standardization domains
- Spatial analysis functions

## Environment Variables

Copy `.env.example` to `.env` and configure:

- `CLOUD_SQL_INSTANCE` - Cloud SQL instance connection name
- `CLOUD_SQL_PROXY` - Set to `true` for local development
- `DB_USER` - Database username
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name
- `PORT` - Server port (default: 8080)
- `GOOGLE_APPLICATION_CREDENTIALS` - Path to service account key

## Architecture

```
cmd/
├── main.go           # Main application server
└── migrate/          # Migration CLI tool

internal/
└── db/               # Database utilities

migrations/           # SQL migration files

.env.example         # Environment variables template
Makefile            # Development commands
```

## Backup and Recovery

**Automated Backups:** ✅ Configured and verified
- **Schedule:** Daily at 02:00 UTC
- **Retention:** 14 daily backups
- **Point-in-time recovery:** 7 days
- **Status:** All backups successful

**Commands:**
- `./scripts/verify-backups.sh` - Verify backup health
- See `docs/deployment/backup-recovery.md` for full procedures

## Part 1 Complete! ✅

All setup tasks completed:
- ✅ Cloud SQL PostgreSQL 17 instance with PostGIS
- ✅ Authorized networks configured (IP: 82.217.141.244/32)
- ✅ Cloud SQL Proxy setup for local development
- ✅ Database schema with migrations
- ✅ Connection pooling (pgxpool)
- ✅ Health check endpoint (`/health`)
- ✅ Automated backups and point-in-time recovery

**Database ready for development!**

### 📊 Schema Statistics:
- **21 tables** with comprehensive plant data structure
- **13 measurement domains** for data standardization
- **7 enum types** for controlled vocabularies
- **Full PostGIS spatial support** with analysis functions
- **Production-ready** with proper indexing and constraints

## 📚 Documentation

Comprehensive documentation is available in the `/docs` folder:

- **[📚 Documentation Index](./docs/README.md)** - Complete documentation overview
- **[🏗️ System Architecture](./docs/architecture/system-overview.md)** - System design and components
- **[🗄️ Database Schema](./docs/database/schema-overview.md)** - Complete database documentation
- **[🗺️ Spatial Queries](./docs/database/spatial-queries.md)** - PostGIS operations and examples
- **[🚀 Cloud SQL Setup](./docs/deployment/cloud-sql-setup.md)** - Infrastructure setup guide

Ready to proceed with Part 2: Plant Domain Service implementation.