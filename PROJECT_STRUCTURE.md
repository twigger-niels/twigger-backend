# Project Structure

This document explains how the Twigger backend repository is organized.

## Philosophy

As a single developer working on a multi-service system, the structure balances:
- **Clarity**: Easy to find what you need
- **Scalability**: Simple to add new services
- **Context switching**: Service-specific docs live with their code
- **System overview**: Root docs provide the big picture

## Directory Layout

```
twigger-backend/
├── docs/                           # System-wide documentation
│   ├── README.md                   # System architecture overview
│   ├── API_GATEWAY.md             # API gateway design
│   ├── DEPLOYMENT.md              # Deployment guide
│   └── swagger/                   # API documentation
│
├── backend/
│   ├── plant-service/
│   │   ├── docs/                  # Plant service specific docs
│   │   │   ├── architecture.md    # Service architecture
│   │   │   ├── prd.md            # Product requirements
│   │   │   ├── tasks.md          # Task tracking
│   │   │   └── CLAUDE.md         # Service-specific AI instructions
│   │   ├── domain/
│   │   ├── infrastructure/
│   │   └── ...
│   │
│   ├── garden-service/
│   │   ├── docs/
│   │   ├── domain/
│   │   └── ...
│   │
│   └── shared/
│       ├── mocks/
│       ├── validators/
│       └── docs/
│           └── COMMON_PATTERNS.md  # Shared utilities documentation
│
├── cmd/                           # Executable commands
│   ├── api-gateway/
│   ├── migrate/
│   └── ...
│
├── internal/                      # Shared internal packages
│   ├── api-gateway/
│   ├── db/
│   └── ...
│
├── migrations/                    # Database migrations
│
├── tests/                        # Integration tests
│
├── CLAUDE.md                     # Root AI instructions (general patterns)
├── PROJECT_STRUCTURE.md          # This file
├── go.mod
└── go.sum
```

## Documentation Organization

### Root `/docs/` - System-Wide Documentation

**Purpose**: Cross-cutting concerns that affect multiple services

**Contains**:
- System architecture diagrams
- Service communication patterns
- Shared infrastructure (database, authentication, deployment)
- API gateway documentation
- Swagger/OpenAPI specifications
- Deployment guides

**When to add here**: Documentation that affects or describes interactions between multiple services

### Root `/CLAUDE.md` - General Development Patterns

**Purpose**: Development guidelines applicable across all services

**Contains**:
- Code style guidelines (Go, SQL, testing)
- Common gotchas (all 45+ critical patterns)
- Testing patterns (unit, integration, performance)
- Git workflow and commit conventions
- Localization patterns
- PostGIS/spatial patterns

**When to update**: When you discover a pattern or gotcha that applies to multiple services

### Service-Specific `/backend/{service}/docs/`

**Purpose**: Service boundary documentation

**Contains**:
- Service architecture
- Business requirements (PRD)
- Task tracking and implementation progress
- Service-specific CLAUDE.md (only if patterns differ significantly from root)

**When to add here**: Documentation specific to one service's domain logic

### `/backend/shared/docs/` - Shared Utilities

**Purpose**: Documentation for reusable packages

**Contains**:
- Common patterns documentation
- Shared validator usage
- Mock usage guidelines
- Shared middleware documentation

## File Naming Conventions

- **UPPERCASE.md**: Important reference documents (README, CLAUDE, ARCHITECTURE)
- **lowercase.md**: Working documents (tasks, notes, drafts)
- **PascalCase.md**: Specific topic guides (API_GATEWAY, DEPLOYMENT)

## Working with This Structure

### Starting Work on a Service

1. Open the service directory: `cd backend/plant-service`
2. Review service docs: `docs/tasks.md`, `docs/architecture.md`
3. Check root CLAUDE.md for development patterns
4. Run service-specific tests

### Adding a New Service

1. Create directory: `backend/{new-service}/`
2. Add docs folder: `backend/{new-service}/docs/`
3. Create service docs: `architecture.md`, `prd.md`, `tasks.md`
4. Update `docs/README.md` with service description

### Updating System Documentation

1. **Service-specific change**: Update `backend/{service}/docs/`
2. **Cross-service pattern**: Update root `CLAUDE.md`
3. **System architecture change**: Update `docs/README.md` or `docs/ARCHITECTURE.md`
4. **API change**: Update `docs/swagger/` or `docs/API_GATEWAY.md`

## Migration from Old Structure

The `/epics/` folder has been deprecated. Content has been moved to:
- Service-specific docs → `backend/{service}/docs/`
- System-wide docs → `docs/`

## Benefits of This Organization

### For Solo Development
- **Clear context**: Know exactly where to look for information
- **Fast navigation**: Service docs next to code
- **Easy cleanup**: Delete service folder removes all related docs

### For Future Growth
- **Team onboarding**: Clear documentation hierarchy
- **Service independence**: Each service fully documented
- **Microservices ready**: Services can be extracted to separate repos

### For AI Assistance
- **Scoped instructions**: Root CLAUDE.md for general patterns, service CLAUDE.md for specifics
- **Clear boundaries**: AI can focus on service-specific context
- **Discoverable patterns**: All gotchas in one place (root CLAUDE.md)

## Questions?

If the right location for documentation is unclear:
1. Does it affect multiple services? → `docs/`
2. Is it a development pattern? → `CLAUDE.md`
3. Is it service-specific? → `backend/{service}/docs/`
4. Is it about shared utilities? → `backend/shared/docs/`
