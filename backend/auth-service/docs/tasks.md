# Tasks Tracking - Authentication Service
**Last Updated:** 2025-01-27
**Version:** 2.0 (Aligned with Go-based Twigger System Architecture)

---

## Overview
This document tracks all development tasks for the Authentication Service. The auth service provides Firebase-based identity management integrated with the existing workspace-based multi-tenant architecture.

## Progress Summary
| Phase | Status | Completion | Priority | Blockers |
|-------|--------|-----------|----------|----------|
| Phase 1: Foundation (Week 1-2) | 📋 Not Started | 0% | P0 | None |
| Phase 2: Complete Auth Flow (Week 3-4) | 📋 Not Started | 0% | P0 | Phase 1 |
| Phase 3: Social Providers (Week 5-6) | 📋 Not Started | 0% | P1 | Phase 2 |
| Phase 4: Production Ready (Week 7-8) | 📋 Not Started | 0% | P1 | Phase 3 |

## Recent Major Achievements
- ✅ **Architecture & PRD Documentation Complete (2025-01-27)**: Aligned auth-service docs with Go-based system architecture, workspace integration, and existing localization infrastructure

## Task Status Legend
- 📋 **TODO**: Not started
- 🚧 **IN PROGRESS**: Currently being worked on
- ✅ **DONE**: Completed and tested
- 🔍 **IN REVIEW**: Code complete, awaiting review
- ❌ **BLOCKED**: Cannot proceed due to dependency

---

## Phase 1: Foundation (Week 1-2)
**Owner**: Unassigned | **Status**: 📋 Not Started | **Priority**: P0 (Must do first)
**Goal**: Token verification and user lookup working

### Firebase Integration Tasks
- [ ] 📋 Add Firebase Admin SDK to go.mod
  - **Package**: `firebase.google.com/go/v4`
  - **Files**: `go.mod`, `go.sum`
- [ ] 📋 Implement Firebase token verification in API Gateway middleware
  - **File**: `internal/api-gateway/middleware/auth.go` (line 129-150)
  - **Replace**: Mock implementation with real Firebase SDK
  - **Function**: `verifyFirebaseToken(ctx context.Context, token string) (*auth.Token, error)`
- [ ] 📋 Configure Firebase service account credentials
  - **File**: `configs/firebase-admin-key.json`
  - **Env**: `FIREBASE_CREDENTIALS_PATH` environment variable
  - **Secret Manager**: Store in Google Secret Manager for production
- [ ] 📋 Add Firebase emulator configuration for local development
  - **File**: `docker-compose.yml` (add firebase-emulator service)
  - **Port**: 9099 for Auth emulator
  - **Env**: `FIREAUTH_EMULATOR_HOST=localhost:9099`

### Database Migration Tasks
- [ ] 📋 Create migration 008: Add auth fields to users table
  - **File**: `migrations/008_add_auth_fields.up.sql`
  - **Fields**: `firebase_uid`, `email_verified`, `phone_number`, `photo_url`, `provider`, `last_login_at`, `deleted_at`
  - **Indexes**: `idx_users_firebase_uid`, `idx_users_deleted_at`, `idx_users_last_login`
  - **See**: Appendix C in PRD for full SQL
- [ ] 📋 Create auth_sessions table
  - **File**: `migrations/008_add_auth_fields.up.sql`
  - **Purpose**: Track sessions for audit and revocation
  - **Indexes**: `idx_auth_sessions_user_id`, `idx_auth_sessions_expires_at`
- [ ] 📋 Create auth_audit_log table (partitioned by month)
  - **File**: `migrations/008_add_auth_fields.up.sql`
  - **Partitioning**: Monthly partitions for performance
  - **Indexes**: `idx_auth_audit_user_id`, `idx_auth_audit_created_at`, `idx_auth_audit_event_type`
- [ ] 📋 Create linked_accounts table
  - **File**: `migrations/008_add_auth_fields.up.sql`
  - **Purpose**: Track social provider connections
  - **Constraint**: `UNIQUE(provider, provider_user_id)`
- [ ] 📋 Run migration 008 on development database
  - **Command**: `go run cmd/migrate/main.go up`
  - **Verify**: Check tables exist with `\dt auth_*` in psql

### Domain Model Tasks
- [ ] 📋 Create User entity
  - **File**: `backend/auth-service/domain/entity/user.go`
  - **Fields**: Match extended `users` table schema
  - **Localization**: Include `PreferredLanguageID`, `CountryID` (reuses existing system)
  - **See**: Architecture section 5.1 for struct definition
- [ ] 📋 Create Session entity
  - **File**: `backend/auth-service/domain/entity/session.go`
  - **Fields**: `SessionID`, `UserID`, `DeviceID`, `IPAddress`, `CreatedAt`, `ExpiresAt`, `RevokedAt`
- [ ] 📋 Create AuditEvent entity
  - **File**: `backend/auth-service/domain/entity/audit_event.go`
  - **Fields**: `EventType`, `Success`, `Metadata`, `IPAddress`, `UserAgent`

### Repository Layer Tasks
- [ ] 📋 Create UserRepository interface
  - **File**: `backend/auth-service/domain/repository/user_repository.go`
  - **Methods**: `Create`, `GetByID`, `GetByFirebaseUID`, `GetByEmail`, `Update`, `SoftDelete`, `UpdateLastLogin`
  - **See**: Architecture section 5.2 for full interface
- [ ] 📋 Implement PostgresUserRepository
  - **File**: `backend/auth-service/infrastructure/persistence/postgres_user_repository.go`
  - **Pattern**: Follow existing plant/garden service repository patterns
  - **Gotchas**: All fields in SELECT/INSERT/UPDATE, timestamp handling, nullable fields use pointers
- [ ] 📋 Create SessionRepository interface
  - **File**: `backend/auth-service/domain/repository/session_repository.go`
  - **Methods**: `Create`, `GetByID`, `GetByUserID`, `Revoke`, `RevokeAllForUser`, `DeleteExpired`
- [ ] 📋 Implement PostgresSessionRepository
  - **File**: `backend/auth-service/infrastructure/persistence/postgres_session_repository.go`
  - **Partitioning**: Handle partitioned auth_audit_log table

### Service Layer Tasks
- [ ] 📋 Create AuthService
  - **File**: `backend/auth-service/domain/service/auth_service.go`
  - **Dependencies**: `UserRepository`, `SessionRepository`, `WorkspaceRepository` (from existing system)
  - **Methods**: `CompleteAuthentication`, `GetUser`, `DeleteUser`, `ExportUserData`
  - **See**: Architecture section 5.3 for implementation example
- [ ] 📋 Implement CompleteAuthentication method
  - **Logic**: Check if user exists by firebase_uid
  - **Existing User**: Update `last_login_at`, insert audit log
  - **New User**: Create user + default workspace + workspace_member (transaction)
  - **Return**: User entity with workspace list
- [ ] 📋 Create SessionService
  - **File**: `backend/auth-service/domain/service/session_service.go`
  - **Methods**: `CreateSession`, `RevokeSession`, `GetActiveSessions`
  - **Audit**: Log all session events

### Unit Testing Tasks
- [ ] 📋 Write UserRepository tests
  - **File**: `backend/auth-service/infrastructure/persistence/postgres_user_repository_test.go`
  - **Tests**: All CRUD operations, firebase_uid lookup, soft delete
  - **Pattern**: Use mocks, no database required
- [ ] 📋 Write AuthService tests
  - **File**: `backend/auth-service/domain/service/auth_service_test.go`
  - **Tests**: CompleteAuthentication (new user, existing user), error handling
  - **Mocks**: Mock repositories
  - **Target**: 80%+ code coverage

### Documentation Tasks
- [ ] 📋 Update API Gateway README with auth flow
  - **File**: `cmd/api-gateway/README.md`
  - **Add**: Firebase token verification steps, user context propagation
- [ ] 📋 Document environment variables
  - **File**: `backend/auth-service/docs/environment.md`
  - **Variables**: `FIREBASE_PROJECT_ID`, `FIREBASE_CREDENTIALS_PATH`, `AUTH_ENABLED`

---

## Phase 2: Complete Authentication Flow (Week 3-4)
**Owner**: Unassigned | **Status**: 📋 Not Started | **Priority**: P0
**Goal**: End-to-end auth flow from client to database

### API Handler Tasks
- [ ] 📋 Create AuthHandler
  - **File**: `backend/auth-service/handlers/auth_handler.go`
  - **Endpoints**: `POST /api/v1/auth/verify`, `POST /api/v1/auth/logout`, `GET /api/v1/auth/me`
  - **Pattern**: Follow existing plant/garden handler patterns
- [ ] 📋 Implement POST /api/v1/auth/verify
  - **Request**: JWT in Authorization header, optional device_id in body
  - **Response**: User profile + workspace list + session_id
  - **Logic**: Call AuthService.CompleteAuthentication
  - **See**: PRD section 14 for API spec
- [ ] 📋 Implement POST /api/v1/auth/logout
  - **Request**: JWT in header, optional device_id and revoke_all_sessions in body
  - **Response**: Success message + sessions_revoked count
  - **Logic**: Revoke session(s), log audit event
- [ ] 📋 Implement GET /api/v1/auth/me
  - **Request**: JWT in header
  - **Response**: Full user profile + workspace memberships
  - **Logic**: Get user by firebase_uid, load workspaces

### Workspace Integration Tasks
- [ ] 📋 Verify workspaces table exists in schema
  - **File**: Check `migrations/000001_comprehensive_plant_schema.up.sql`
  - **Fields**: `workspace_id`, `owner_id`, `name`, `created_at`
  - **Risk**: High priority - schema must exist before implementation
- [ ] 📋 Verify workspace_members table exists
  - **File**: Check `migrations/000001_comprehensive_plant_schema.up.sql`
  - **Fields**: `workspace_id`, `user_id`, `role`, `joined_at`
  - **Constraint**: PRIMARY KEY (workspace_id, user_id)
- [ ] 📋 Create or extend WorkspaceRepository interface
  - **File**: Check if exists, else create `backend/shared/domain/repository/workspace_repository.go`
  - **Methods**: `Create`, `GetByID`, `AddMember`, `GetUserWorkspaces`
- [ ] 📋 Update AuthService to create workspace on registration
  - **Logic**: In CompleteAuthentication for new users
  - **Workspace Name**: `"{username}'s Garden"`
  - **Transaction**: User + Workspace + WorkspaceMember in single transaction

### Audit Logging Tasks
- [ ] 📋 Create AuditService
  - **File**: `backend/auth-service/domain/service/audit_service.go`
  - **Methods**: `LogEvent`, `GetUserAuditLog`
  - **Events**: `user_registered`, `user_login`, `user_logout`, `session_revoked`
- [ ] 📋 Implement audit logging in AuthService
  - **Location**: All auth operations (CompleteAuthentication, logout, etc.)
  - **Fields**: `user_id`, `event_type`, `success`, `ip_address`, `user_agent`, `metadata`
- [ ] 📋 Add audit log partition management
  - **Task**: Monthly partition creation
  - **File**: Create script `scripts/create-audit-partitions.sh`
  - **Automation**: Cron job or Cloud Scheduler

### Integration Testing Tasks
- [ ] 📋 Create integration test infrastructure
  - **File**: `backend/auth-service/tests/integration/setup_test.go`
  - **Docker**: PostgreSQL + PostGIS + Firebase emulator
  - **Pattern**: Use existing `docker-compose.test.yml` pattern
- [ ] 📋 Write integration tests for auth flow
  - **File**: `backend/auth-service/tests/integration/auth_flow_test.go`
  - **Tests**: New user registration, existing user login, workspace creation
  - **Database**: Real PostgreSQL with schema
- [ ] 📋 Write integration tests for audit logging
  - **File**: `backend/auth-service/tests/integration/audit_test.go`
  - **Tests**: Events logged correctly, partitions working

### Router Integration Tasks
- [ ] 📋 Register auth endpoints in API Gateway router
  - **File**: `internal/api-gateway/router/router.go`
  - **Routes**: `/api/v1/auth/*` routes
  - **Middleware**: RequireAuth for /me, /logout; no auth for /verify (verify is the auth)
- [ ] 📋 Update API Gateway main.go to initialize auth service
  - **File**: `cmd/api-gateway/main.go`
  - **Add**: AuthService, UserRepository, SessionRepository initialization
  - **Inject**: Into Handlers struct

---

## Phase 3: Social Providers (Week 5-6)
**Owner**: Unassigned | **Status**: 📋 Not Started | **Priority**: P1
**Goal**: Google, Apple, Facebook login functional in production

### Firebase Console Configuration
- [ ] 📋 Create Firebase project: `twigger-prod`
  - **Console**: https://console.firebase.google.com
  - **Enable**: Authentication service
  - **Generate**: Service account key for backend
- [ ] 📋 Enable Google Sign-In provider
  - **OAuth**: Configure client IDs for iOS, Android, Web
  - **Scopes**: email, profile
  - **Test**: With real Google account
- [ ] 📋 Enable Apple Sign-In provider
  - **Requirement**: Apple Developer account ($99/year)
  - **Config**: Team ID, Key ID, Private Key
  - **Services ID**: Register bundle IDs
  - **Test**: Requires real iOS device or Mac Simulator
- [ ] 📋 Enable Facebook Login provider
  - **Requirement**: Facebook App ID
  - **Config**: App Secret, OAuth redirect URIs
  - **Test**: With real Facebook account

### Provider-Specific Implementation
- [ ] 📋 Handle Google-specific user attributes
  - **Fields**: `photo_url` from Google profile
  - **Validation**: Email verified by default from Google
- [ ] 📋 Handle Apple-specific privacy features
  - **Hide Email**: Handle private relay emails
  - **First Login**: Name only provided on first auth
  - **Store**: Capture name before Apple hides it
- [ ] 📋 Handle Facebook-specific data
  - **Permissions**: Request email, public_profile
  - **Graph API**: Fetch profile picture URL
  - **Verification**: Email verified status from Facebook

### Account Linking Tasks
- [ ] 📋 Implement account linking logic
  - **Scenario**: User signs in with Google, later tries Facebook with same email
  - **Logic**: Link accounts instead of creating duplicate
  - **Table**: `linked_accounts` table
  - **UI**: Prompt user to confirm account link
- [ ] 📋 Handle provider conflicts
  - **Scenario**: Email registered with password, tries Google with same email
  - **Options**: Merge accounts OR require password verification first
  - **Security**: Prevent account takeover via email alone

### Testing Tasks
- [ ] 📋 Test Google Sign-In on all platforms
  - **iOS**: Native SDK flow
  - **Android**: Native SDK flow
  - **Web**: OAuth redirect flow
  - **Verify**: User profile synced correctly
- [ ] 📋 Test Apple Sign-In (iOS/Mac only)
  - **Device**: Real device or Xcode simulator
  - **Private Email**: Test hide email feature
  - **Verify**: Backend handles private relay
- [ ] 📋 Test Facebook Login on all platforms
  - **iOS**: Native SDK
  - **Android**: Native SDK
  - **Web**: OAuth flow
  - **Limited Login**: Test iOS 14.5+ privacy features

---

## Phase 4: Production Readiness (Week 7-8)
**Owner**: Unassigned | **Status**: 📋 Not Started | **Priority**: P1
**Goal**: Secure, monitored, production-ready system

### Caching Layer Tasks
- [ ] 📋 Add Redis for Firebase public key caching
  - **Purpose**: Cache public keys for 4 hours (per Firebase docs)
  - **Implementation**: In-memory sync.Map OR Redis
  - **Invalidation**: Refresh every 4 hours
- [ ] 📋 Implement user lookup caching
  - **Key**: `user:firebase:{uid}` → User JSON
  - **TTL**: 5 minutes
  - **Invalidation**: On user update (pattern: `user:firebase:*`)
- [ ] 📋 Cache session validation
  - **Key**: `session:{session_id}` → Valid/Revoked boolean
  - **TTL**: Match session expiration
  - **Invalidation**: On session revoke

### Rate Limiting Tasks
- [ ] 📋 Implement rate limiting per endpoint
  - **Limits**:
    - `/auth/verify`: 100 req/min per IP
    - `/auth/logout`: 20 req/min per IP
    - `/auth/me`: 200 req/min per user
  - **Algorithm**: Token bucket with Redis
  - **Response**: 429 Too Many Requests with Retry-After header

### Security Audit Tasks
- [ ] 📋 OWASP Top 10 compliance check
  - **A01**: Workspace isolation enforced
  - **A02**: TLS 1.3 everywhere
  - **A03**: Parameterized queries only
  - **A07**: Firebase handles auth
  - **Checklist**: See Architecture section 13.1
- [ ] 📋 Dependency vulnerability scan
  - **Tool**: `go list -m all | nancy` or Snyk
  - **Frequency**: Weekly in CI/CD
  - **Action**: Update dependencies with vulnerabilities
- [ ] 📋 Secret management review
  - **Check**: No secrets in code or git history
  - **Tool**: Google Secret Manager
  - **Rotation**: Document rotation procedures

### Load Testing Tasks
- [ ] 📋 Simulate 1,000 concurrent authentications
  - **Tool**: K6 or Vegeta
  - **Scenario**: Mix of new user registrations + existing user logins
  - **Metrics**: p95 latency, error rate, throughput
  - **Target**: < 100ms p95 for existing user, < 300ms for new user
- [ ] 📋 Database connection pool tuning
  - **Current**: 25 max, 5 idle, 5min lifetime
  - **Monitor**: Pool utilization, wait times
  - **Adjust**: Based on load test results
- [ ] 📋 API Gateway scaling test
  - **Cloud Run**: Auto-scaling 2-100 instances
  - **Test**: Verify horizontal scaling works
  - **Target**: Handle 1,000 concurrent users

### Monitoring & Alerting Tasks
- [ ] 📋 Set up Cloud Monitoring dashboard
  - **Metrics**: Authentication success rate, token verification latency, error rates
  - **Charts**: Time series, distribution, heatmaps
  - **Access**: Share with team
- [ ] 📋 Configure critical alerts
  - **Alert 1**: Token verification failures > 5% → Page on-call
  - **Alert 2**: Database connection pool exhausted → Page on-call
  - **Alert 3**: Auth latency > 500ms p95 → Slack alert
  - **Channels**: PagerDuty for critical, Slack for warnings
- [ ] 📋 Set up distributed tracing
  - **Tool**: Cloud Trace or Jaeger
  - **Traces**: Full request flow (Gateway → AuthService → Database)
  - **Context**: Propagate trace IDs across services

### Documentation Tasks
- [ ] 📋 Generate Swagger/OpenAPI documentation
  - **Tool**: swag for Go annotations
  - **File**: `docs/swagger/auth-api.yaml`
  - **Endpoints**: All 3 auth endpoints documented
- [ ] 📋 Write deployment runbook
  - **File**: `backend/auth-service/docs/deployment-runbook.md`
  - **Content**: Pre-deployment checklist, rollback procedures, common issues
- [ ] 📋 Create incident response guide
  - **File**: `backend/auth-service/docs/incident-response.md`
  - **Scenarios**: Token verification failures, database down, Firebase outage
  - **Actions**: Step-by-step resolution procedures
- [ ] 📋 Document API for client developers
  - **File**: `backend/auth-service/docs/client-integration-guide.md`
  - **Examples**: Request/response samples, error handling, token refresh

---

## Bug Fixes & Issues
*Track bugs discovered during development*

### Critical Issues
- [ ] 📋 Verify workspace schema exists in migrations
  - **Risk**: HIGH - Implementation cannot proceed without workspace tables
  - **Action**: Immediate schema verification required
  - **Status**: Needs investigation

### High Priority Issues
- [ ] 📋 Firebase Admin SDK requires service account credentials
  - **Issue**: Cannot deploy without credential file
  - **Solution**: Store in Google Secret Manager
  - **Status**: Pending credential generation

### Medium Priority Issues
- [ ] 📋 Apple Sign-In requires paid developer account
  - **Impact**: Cannot test without $99/year account
  - **Workaround**: Test on real device with developer account
  - **Status**: Blocked on budget approval

### Low Priority Issues
- [ ] 📋 Firebase emulator doesn't support all OAuth providers
  - **Impact**: Must test social login with real Firebase project
  - **Workaround**: Use integration tests with Firebase project
  - **Status**: Acceptable limitation

---

## Technical Debt
*Track technical debt to be addressed post-MVP*

- [ ] 📋 Add MFA support (TOTP)
- [ ] 📋 Implement email verification flows
- [ ] 📋 Add device management UI
- [ ] 📋 Implement SAML SSO (enterprise feature)
- [ ] 📋 Add SCIM user provisioning
- [ ] 📋 Optimize audit log queries (currently partitioned monthly, consider yearly)
- [ ] 📋 Add Redis Sentinel for cache high availability
- [ ] 📋 Implement passwordless authentication (magic links)

---

## Metrics & Success Criteria

### Performance Metrics
- [ ] 📋 Token verification < 50ms (p95) ← **Target**
- [ ] 📋 User lookup < 20ms (p95) ← **Target**
- [ ] 📋 Complete auth (existing user) < 100ms (p95) ← **Target**
- [ ] 📋 Complete auth (new user) < 300ms (p95) ← **Target**
- [ ] 📋 Support 1,000 concurrent users ← **Target**

### Quality Metrics
- [ ] 📋 >80% unit test coverage ← **Target**
- [ ] 📋 <0.1% error rate ← **Target**
- [ ] 📋 Zero critical security issues ← **Target**
- [ ] 📋 All integration tests passing ← **Target**
- [ ] 📋 Documentation complete ← **Target**

### Business Metrics
- [ ] 📋 80% signup completion rate ← **Target**
- [ ] 📋 95% successful login rate ← **Target**
- [ ] 📋 < 2% auth-related support tickets ← **Target**
- [ ] 📋 99.9% uptime (Firebase SLA) ← **Target**

---

## Notes & Decisions

### Key Decisions Made
- **2025-01-27**: Using Firebase Authentication for provider management (no custom OAuth)
- **2025-01-27**: Extending existing `users` table instead of creating separate auth schema
- **2025-01-27**: Reusing existing localization system (no auth-specific translations)
- **2025-01-27**: Workspace-based multi-tenancy from day one (no single-user mode)
- **2025-01-27**: Go-native implementation (no Node.js/TypeScript)
- **2025-01-27**: Single Cloud SQL instance (no multi-region database for MVP)

### Open Questions
- [ ] Should we implement passwordless auth (magic links) in Phase 2 or defer to Phase 5?
- [ ] What's the workspace invitation flow? (Email? Link?)
- [ ] Should we support multiple workspaces per user from day one?
- [ ] Do we need organization-level accounts (above workspace)?
- [ ] What's the account deletion retention period? (Currently 30 days soft delete)

### Risks & Mitigations
- **Risk**: Firebase service outage
  - **Mitigation**: Cache public keys (4-hour TTL), graceful degradation for existing sessions
- **Risk**: JWT verification performance bottleneck
  - **Mitigation**: Cache public keys in-memory, pre-warm cache on startup
- **Risk**: Workspace table doesn't exist in schema
  - **Mitigation**: **IMMEDIATE ACTION REQUIRED** - Verify schema before Phase 2
- **Risk**: Database connection pool exhaustion
  - **Mitigation**: Monitor pool metrics, implement request queuing, add read replicas

---

## Sprint Planning

### Sprint 1 (Week 1-2): Foundation ← **Current Sprint**
- Complete Phase 1: Foundation
- Deliverable: Token verification + user lookup working
- Key Tasks: AUTH-SDK, Migration 008, User entity, UserRepository, AuthService

### Sprint 2 (Week 3-4): Complete Auth Flow
- Complete Phase 2: Complete Authentication Flow
- Deliverable: End-to-end auth from client to database
- Key Tasks: API handlers, workspace integration, audit logging

### Sprint 3 (Week 5-6): Social Providers
- Complete Phase 3: Social Providers
- Deliverable: Google, Apple, Facebook login working
- Key Tasks: Firebase console config, provider-specific handling, account linking

### Sprint 4 (Week 7-8): Production Ready
- Complete Phase 4: Production Readiness
- Deliverable: Secure, monitored, production deployment
- Key Tasks: Caching, rate limiting, security audit, load testing, monitoring

---

## Quick Reference

### Current State
- **Phase**: Phase 1 (Foundation)
- **Active Branch**: `feature/auth-phase-1-foundation`
- **Database Status**: Migration 008 pending creation
- **Firebase Status**: Project pending creation
- **Test Coverage**: 0% (target: 80%)
- **Next Major Milestone**: Token verification working (Week 2)

### Environment URLs
- **Local Backend**: http://localhost:8080
- **Firebase Console**: https://console.firebase.google.com
- **GCP Console**: https://console.cloud.google.com
- **Database**: Cloud SQL PostgreSQL 17

### Key Files Reference
- **Architecture**: `backend/auth-service/docs/architecture.md`
- **PRD**: `backend/auth-service/docs/prd.md`
- **Existing Middleware**: `internal/api-gateway/middleware/auth.go` (to be enhanced)
- **Migration Example**: See PRD Appendix C for SQL
- **User Table**: `migrations/000001_comprehensive_plant_schema.up.sql` (line 400)

---

*Last Updated: 2025-01-27*
*Next Review: Upon completion of Phase 1*
*Version: 2.0 (Go-aligned)*
