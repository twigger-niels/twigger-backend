# Product Requirements Document (PRD)
## Plant Database & Garden Management System

### Executive Summary
A comprehensive plant database system with spatial garden mapping capabilities, designed to help gardeners plan, manage, and optimize their gardens using scientific plant data and spatial analysis.

### Problem Statement
Gardeners currently lack a unified system that combines:
- Accurate, multi-source plant information
- Spatial garden planning with real-world constraints
- Scientific analysis of garden conditions (shade, drainage, frost)
- Companion planting and spacing recommendations
- Multi-country support with localized climate data

### Target Users
1. **Primary:** Home gardeners wanting to optimize their garden layout
2. **Secondary:** Professional landscapers planning client gardens
3. **Tertiary:** Community gardens managing shared spaces

### Core Features

#### P0 - Must Have (MVP)
- [ ] Plant database with 10,000+ species
- [x] **Multi-language support (country + language specific)** ✅ COMPLETED
- [x] **Localized plant names and descriptions** ✅ COMPLETED
  - 8 localization tables implemented (plant_common_names, plant_descriptions, etc.)
  - 4-tier fallback chain (Country+Language → Language → English → Raw)
  - Language-aware caching and batch loading
  - Composite indexes for performance
- [ ] Garden boundary drawing on maps
- [x] Basic plant search and filtering ✅ COMPLETED (Part 2 - 85% complete)
- [ ] User authentication (Firebase)
- [ ] Add plants to garden locations
- [ ] View garden layout

#### P1 - Should Have
- [ ] Garden zones (beds, paths, etc.)
- [ ] Shade calculation based on features
- [x] Companion planting suggestions ✅ COMPLETED (Part 2 - with localization)
- [ ] Multi-workspace support
- [ ] Plant spacing validation
- [x] Growing conditions matching ✅ PARTIALLY COMPLETE (basic implementation done, filtering TODO)

#### P2 - Nice to Have
- [ ] Frost pocket detection
- [ ] Drainage/slope analysis
- [ ] Seasonal planning
- [ ] Garden sharing/collaboration
- [ ] Photo observations
- [ ] Planting reminders

### Technical Requirements

#### Performance
- Plant search: <100ms response time
- Garden rendering: <200ms for 1000 plants
- API requests: <200ms p95 latency
- Support 10,000 concurrent users

#### Scalability
- Database: Support 1M+ plants
- Storage: 10TB for images
- Gardens: 100K+ gardens
- API: 1000 requests/second

#### Reliability
- 99.9% uptime (8.76 hours downtime/year)
- Automated backups every 24 hours
- Point-in-time recovery
- Multi-region failover (future)

#### Security
- Firebase Authentication
- Row-level security in database
- Workspace isolation
- API rate limiting
- HTTPS only

### User Stories

#### As a Gardener
- I want to search for plants suitable for my climate zone
- I want to draw my garden boundaries on a map
- I want to see where shadows fall in my garden
- I want to know which plants grow well together
- I want to track what I planted where and when

#### As a Professional Landscaper
- I want to manage multiple client gardens
- I want to generate planting plans
- I want to analyze drainage patterns
- I want to share garden plans with clients

#### As a Garden Administrator
- I want to invite collaborators to my garden
- I want to control who can edit vs view
- I want to track changes over time

### Success Metrics

#### User Engagement
- Daily Active Users (DAU): 10,000
- Monthly Active Users (MAU): 50,000
- Average session duration: 15 minutes
- Gardens created per month: 5,000

#### Technical Metrics
- API latency p50: <100ms
- API latency p99: <500ms
- Error rate: <0.1%
- Database query time: <50ms

#### Business Metrics
- User retention (30 day): 40%
- User retention (90 day): 25%
- Free to paid conversion: 5%
- Customer satisfaction (NPS): >50

### Constraints

#### Technical
- Must use PostgreSQL with PostGIS
- Must support offline read mode
- Must work on mobile devices
- Must handle conflicting plant data

#### Business
- Initial budget: Development only
- Timeline: 3-month MVP
- Team size: 3-5 developers
- No marketing budget initially

#### Legal/Compliance
- GDPR compliance for EU users
- Data residency requirements
- Plant patent restrictions
- Image copyright compliance

### Out of Scope (v1)
- Native mobile apps (Flutter web only)
- Social features (forums, chat)
- E-commerce (plant/seed sales)
- IoT sensor integration
- AI plant identification from photos
- Pest/disease diagnosis
- Weather integration (beyond basic zones)

### Risks & Mitigation

| Risk                      | Impact | Probability | Mitigation                           |
| ------------------------- | ------ | ----------- | ------------------------------------ |
| Plant data accuracy       | High   | Medium      | Multiple sources, confidence scoring |
| PostGIS complexity        | High   | Low         | Extensive testing, simple fallbacks  |
| Firebase costs            | Medium | Medium      | Rate limiting, caching               |
| User adoption             | High   | Medium      | Focus on core gardener needs         |
| Spatial calculations slow | Medium | Low         | Pre-calculate, cache results         |

### Dependencies
- Firebase Auth service
- Google Maps API
- Cloud SQL (PostgreSQL 17)
- PostGIS extensions
- Multiple plant data sources

### Timeline

#### Phase 1: Foundation (Weeks 1-4)
- Database setup with PostGIS
- Basic plant data import
- Authentication system
- Core API development

#### Phase 2: Garden Features (Weeks 5-8)
- Garden boundary drawing
- Zone management
- Plant placement
- Basic spatial queries

#### Phase 3: Analysis (Weeks 9-12)
- Shade calculation
- Companion planting
- Spacing validation
- Recommendations engine

#### Phase 4: Polish (Weeks 13-16)
- Performance optimization
- UI refinement
- Testing & bug fixes
- Documentation

### Open Questions
1. Should we support imperial and metric units from day 1?
2. How do we handle plant name conflicts between sources?
3. What's the minimum viable climate zone support?
4. Should garden history be preserved (temporal data)?
5. How detailed should shade calculations be?

### Appendix

#### Competitor Analysis
- **Gardenize:** Good plant diary, lacks spatial planning
- **GrowVeg:** Good planning, limited plant database
- **PlantNet:** Great plant ID, no garden management
- **iScape:** Professional focused, expensive

#### Technical Decisions Made
- PostgreSQL 17 over MySQL (PostGIS support)
- GraphQL over REST-only (complex queries)
- Firebase over Auth0 (better integration)
- Go over Node.js (performance)
- Flutter over React Native (web-first)

#### Recent Architectural Decisions (Part 2 Implementation)
See `architecture.md` for detailed ADRs (ADR-008 through ADR-012):
- **Batch Loading**: N+1 query prevention via batch methods (96% query reduction)
- **Composite Indexes**: Multi-column indexes for localization queries
- **Language-Aware Caching**: Separate cache entries per language/country with pattern invalidation
- **Repository Validation**: Input validation at repository boundary (UUIDs + ISO codes)
- **Service Layer Defaults**: Temporary English hardcoding with TODO markers for API integration