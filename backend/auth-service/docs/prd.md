# Product Requirements Document: Authentication System
**Version:** 2.0 (Aligned with Twigger System Architecture)
**Date:** 2025-01-27
**Status:** Draft

---

## 1. Executive Summary

The authentication system provides secure, Firebase-based identity management for the Twigger gardening platform. This lightweight Go-based service integrates with the existing workspace-based multi-tenant architecture, enabling users to create accounts, sign in, and manage their garden planning profiles across iOS, Android, and web platforms. The system leverages Firebase Authentication for provider management and token lifecycle, while the backend handles workspace membership, user profiles, and security audit logging.

**Key Differentiators:**
- Firebase-managed authentication (no custom password handling)
- Workspace-based multi-tenancy from day one
- Reuses existing localization infrastructure (90+ languages)
- Minimal database footprint (extends existing schema)
- Go-native implementation matching plant/garden services

---

## 2. Business Objectives

### Primary Goals
- Enable secure user registration and authentication across all platforms
- Support multiple authentication methods (Google, Apple, Facebook, Email)
- Seamlessly integrate with workspace-based garden management
- Leverage existing localization system (no auth-specific translations needed)
- Build trust through enterprise-grade security practices (Firebase + Cloud SQL)
- Ensure compliance with data protection laws (GDPR, CCPA)

### Success Metrics (Aligned with Current Stage: Part 4)
- **User Acquisition:** 80% signup completion rate (simpler flow = higher conversion)
- **Authentication Success:** 95% successful login rate
- **Security:** Zero critical security incidents
- **Performance:**
  - < 100ms for existing user authentication (p95)
  - < 300ms for new user registration with workspace creation (p95)
  - < 50ms for token verification (p95)
- **User Satisfaction:** < 2% authentication-related support tickets
- **System Reliability:** 99.9% uptime (leveraging Firebase SLA)

---

## 3. User Personas

### Primary Persona: Home Gardener
- **Age:** 25-65
- **Tech Savviness:** Moderate
- **Needs:** Quick, simple signup process; immediate access to garden planning
- **Pain Points:** Remembering passwords, complex registration, slow onboarding
- **Preferred Auth:** Google Sign-In (one-tap, no password)
- **Expected Flow:** Login → Automatic workspace created → Start planning garden

### Secondary Persona: Professional Landscaper
- **Age:** 30-50
- **Tech Savviness:** High
- **Needs:** Secure account management, multi-device access, workspace sharing
- **Pain Points:** Security concerns, client data privacy, team collaboration
- **Preferred Auth:** Google/Apple Sign-In with potential for 2FA (Phase 2)
- **Expected Flow:** Login → Create/join multiple workspaces → Manage client gardens

---

## 4. Functional Requirements

### 4.1 User Registration (MVP - Phase 1)

**Social Registration (Primary Flow):**
- **Providers Supported:**
  - Google Sign-In (one-tap, auto-populate email/name/photo)
  - Apple Sign-In (required for iOS, privacy-focused)
  - Facebook Login (optional, for social reach)

- **Backend Flow:**
  1. Client receives Firebase JWT after successful provider auth
  2. Client sends JWT to `POST /api/v1/auth/verify`
  3. Backend verifies JWT signature and extracts `firebase_uid`, `email`, `provider`
  4. If new user:
     - Create user record in `users` table with `firebase_uid`
     - Generate unique username from email
     - Create default workspace: `{username}'s Garden`
     - Add user as workspace admin
     - Insert audit log event: `user_registered`
  5. If existing user:
     - Update `last_login_at` timestamp
     - Insert audit log event: `user_login`
  6. Return user profile + workspace list

- **Automatic Data Population:**
  - Email (from Firebase token)
  - Display name (from provider profile)
  - Photo URL (from provider profile)
  - Email verification status (from Firebase)

**Email/Password Registration (Phase 2, Optional):**
- Firebase handles password requirements (min 8 chars, complexity rules)
- Email verification link sent by Firebase
- Backend only stores `firebase_uid` after verification

### 4.2 Authentication

**Sign In Methods (Firebase-Managed):**
- Google OAuth 2.0
- Apple Sign-In
- Facebook Login
- Email/password (Phase 2)
- Biometric authentication (client-side, Firebase token refresh)

**Session Management:**
- **Access Token:** Firebase JWT (15-minute expiration, signed by Firebase)
- **Refresh Token:** Firebase refresh token (30-day lifecycle, managed by client SDK)
- **Backend Session Tracking:**
  - Audit only (not required for authentication)
  - Track device_id, IP, user agent for security monitoring
  - Enable remote session revocation via `auth_sessions.revoked_at`

**Token Flow:**
1. Client authenticates with Firebase → receives JWT + refresh token
2. Client includes JWT in `Authorization: Bearer {jwt}` header
3. API Gateway verifies JWT signature using Firebase public keys
4. API Gateway extracts `firebase_uid` and loads user from database
5. API Gateway sets user context (user_id, workspace_id, language) in request
6. When access token expires, client refreshes via Firebase SDK (no backend call)

### 4.3 Account Management

**Profile Management:**
- Edit username, display name, photo URL
- Update preferred language (uses existing `languages` table)
- Set location (geography point for climate detection)
- View/manage linked social accounts

**Workspace Management (Core Feature):**
- Create new workspaces (e.g., "Home Garden", "Client: Smith")
- Invite members to workspaces (email invitation)
- Manage workspace roles (admin, member, viewer)
- Switch active workspace (via header or query param)
- Leave/delete workspace

**Account Deletion (GDPR Compliant):**
- Soft delete: Set `users.deleted_at` timestamp
- Revoke all active sessions in `auth_sessions`
- Anonymize audit logs (keep event counts, remove PII)
- Transfer workspace ownership or delete if sole owner
- Hard delete after 30-day retention period

**Data Portability:**
- Export user profile, gardens, plants, zones, analysis results
- Return JSON archive via `GET /api/v1/auth/export`

### 4.4 Security Features

**Phase 1 (MVP):**
- Firebase JWT validation with public key caching
- Rate limiting: 100 requests/minute per IP (at API Gateway)
- Audit logging: All auth events (login, logout, registration)
- Suspicious activity detection: Login from new country → email notification
- Account lockout: After 5 failed token verifications

**Phase 2 (Post-MVP):**
- Multi-factor authentication (Firebase supports TOTP via Admin SDK)
- SMS OTP for high-value actions (via Firebase Phone Auth)
- Device management UI (view/revoke sessions)
- Advanced anomaly detection (geographic, time-based)

### 4.5 Localization & Internationalization

**System Integration (Reuses Existing Infrastructure):**
- **Language Support:** Leverage existing `languages` table (90+ languages with 4-tier fallback)
- **Country Detection:** Use existing `countries` table for regional settings
- **Localization Patterns:** Follow existing service patterns:
  - User's `preferred_language_id` stored in `users` table
  - Client includes `Accept-Language` header
  - API Gateway detects locale and sets context
  - Services use existing localization service for translations

**Auth-Specific Localized Content:**
- Welcome messages: `localized_content` table with key `auth.welcome`
- Email templates: Go templates with localized strings (via existing system)
- Error messages: Translated at client or via localization service

**Name Handling:**
- Support Unicode characters (full UTF-8 in PostgreSQL)
- Single `username` field (flexible for all cultures)
- Display name optional (from social providers)

---

## 5. Non-Functional Requirements

### 5.1 Performance

| Operation | Target (p95) | Rationale |
|-----------|--------------|-----------|
| Token verification | < 50ms | Lightweight JWT signature check |
| User lookup (firebase_uid) | < 20ms | Indexed query, single row |
| Complete auth (existing user) | < 100ms | Lookup + update last_login |
| Complete auth (new user) | < 300ms | Insert user + workspace + member |
| Workspace list retrieval | < 50ms | Indexed query on workspace_members |

**Scalability Target:**
- 1,000 concurrent users (current stage)
- Path to 10,000+ concurrent users (Cloud Run auto-scaling)

### 5.2 Security

**Alignment with Existing Architecture:**
- **Authentication:** Firebase Authentication (OWASP-compliant, SOC 2 certified)
- **Database:** Cloud SQL PostgreSQL with automatic encryption at rest (AES-256)
- **Transport:** TLS 1.3 for all connections (enforced at Load Balancer)
- **Secrets:** Google Secret Manager for Firebase credentials
- **Audit:** Complete audit trail in `auth_audit_log` (partitioned by month)

**Security Standards:**
- OWASP Top 10 compliance
- GDPR compliance (right to erasure, data portability)
- CCPA compliance (data export, do-not-sell)

### 5.3 Scalability (Current Stage: Part 4)

**Immediate Needs:**
- Support 100 concurrent users during development/beta
- Horizontal scaling with Cloud Run (2-10 instances)
- Single Cloud SQL instance (shared with plant/garden services)

**Future Growth Path:**
- Cloud Run auto-scaling (0-100 instances)
- Read replicas for user lookups (when needed)
- Redis caching for Firebase public keys and user profiles
- Database connection pooling (already implemented: 25 max, 5 idle)

### 5.4 Usability

**Onboarding Simplicity:**
- **Target:** 1-click social login → garden planning in < 10 seconds
- **Flow:**
  1. Click "Sign in with Google" (1 click)
  2. Google consent screen (1 click to approve)
  3. Auto-redirected to app with workspace ready
  4. Start adding plants immediately

**Multi-Platform Consistency:**
- Identical authentication flow on iOS, Android, Web
- Firebase SDKs handle platform-specific UI (Google One Tap, Apple Sign-In sheet)

### 5.5 Data Privacy

**GDPR Compliance:**
- **Lawful Basis:** Consent (accepted during registration)
- **Data Minimization:** Only store firebase_uid, email, username, location (optional)
- **Right to Access:** `GET /api/v1/auth/me` returns all user data
- **Right to Erasure:** `DELETE /api/v1/auth/me` soft deletes user
- **Right to Portability:** `GET /api/v1/auth/export` returns JSON archive
- **Data Breach Notification:** Audit logs enable rapid incident response

**CCPA Compliance:**
- Do-not-sell flag (not applicable - no data selling)
- California-specific privacy policy (via `regional_settings` if needed)

---

## 6. User Stories

### Epic: User Registration

```gherkin
Feature: Social Login Registration

Scenario: New user registers with Google
  Given I am on the login screen
  When I tap "Sign in with Google"
  And I approve Google consent
  Then I see a loading indicator for < 2 seconds
  And I am redirected to the garden planning screen
  And I have a default workspace named "{MyName}'s Garden"
  And I receive a welcome email in my preferred language

Acceptance Criteria:
- Complete flow in < 10 seconds
- Auto-populate email, name, photo from Google profile
- Email verification inherited from Google (email_verified = true)
- Workspace created with user as admin
- Audit log records registration event
```

### Epic: Secure Authentication

```gherkin
Feature: Returning User Login

Scenario: Existing user logs in with Google
  Given I previously registered with Google
  When I tap "Sign in with Google"
  And I approve Google consent
  Then I am logged in within 5 seconds
  And I see my list of workspaces
  And I see my last active workspace selected
  And my last_login_at timestamp is updated

Acceptance Criteria:
- No additional registration steps
- Workspace list loads with gardens visible
- Security audit log records login event with IP/device
```

### Epic: Multi-Workspace Management

```gherkin
Feature: Create and Switch Workspaces

Scenario: User creates a workspace for a client
  Given I am logged in as a landscaper
  When I tap "Create Workspace"
  And I enter name "Client: Smith Residence"
  Then I see the new workspace in my list
  And I am set as the workspace admin
  And I can switch between "My Personal Garden" and "Client: Smith"

Acceptance Criteria:
- Workspace creation < 200ms
- Workspace switcher in app header
- All gardens scoped to active workspace
- Invite members button visible (Phase 2)
```

### Epic: Account Security

```gherkin
Feature: Suspicious Login Detection

Scenario: User logs in from a new country
  Given I previously logged in from the United States
  When I log in from France
  Then I receive an email notification "New login from France"
  And the email includes device info and login time
  And the email has a "Wasn't you? Secure your account" link

Acceptance Criteria:
- Email sent within 1 minute of login
- Email in user's preferred language
- Link to revoke session (Phase 2) or change password
```

---

## 7. Technical Constraints

### System Architecture Alignment
- **Language:** Go 1.22+ (no Node.js/TypeScript)
- **Database:** Existing Cloud SQL PostgreSQL instance (no separate auth database)
- **Localization:** Reuse existing system (no auth-specific translation tables)
- **Domain Design:** Follow `entity.go`, `repository.go`, `service.go` patterns
- **API Gateway:** Integrate with existing `internal/api-gateway/middleware/auth.go`

### Firebase Limitations
- Firebase free tier: 10K verifications/day (sufficient for MVP)
- JWT expiration: 15 minutes (hardcoded by Firebase)
- Refresh token: 30-day lifecycle (hardcoded by Firebase)
- Social providers: Must be configured in Firebase console

### Database Constraints
- Existing `users` table: Extend with `firebase_uid`, `provider`, `last_login_at`
- Connection pool: 25 max connections (shared with plant/garden services)
- Query performance: Must maintain < 50ms p95 for all auth queries

---

## 8. Dependencies

### External Services
- **Firebase Authentication** (Primary Dependency)
  - Google Sign-In OAuth
  - Apple Sign-In OAuth
  - Facebook Login OAuth
  - JWT public key endpoint
  - Client SDKs (Flutter, Web)

### Internal Dependencies
- **Existing Database Schema:** `users`, `workspaces`, `workspace_members` tables
- **Existing Localization Service:** Language detection and translation
- **API Gateway Middleware:** `internal/api-gateway/middleware/auth.go`
- **Audit Logging:** Infrastructure for `auth_audit_log` table

### Development Dependencies
- Go Firebase Admin SDK: `firebase.google.com/go/v4`
- PostgreSQL driver: `github.com/lib/pq`
- Testing: Firebase emulator for local development

---

## 9. Risks and Mitigations

| Risk | Impact | Probability | Mitigation |
|------|---------|------------|------------|
| Firebase service outage | High | Low | Cache public keys (4-hour TTL), graceful degradation for existing sessions |
| JWT verification performance bottleneck | Medium | Medium | Cache public keys in-memory, pre-warm cache on startup |
| Workspace table doesn't exist in schema | High | Medium | **Immediate action:** Verify schema, create migration if needed |
| Localization system not compatible | Medium | Low | Review existing patterns, adapt if necessary |
| Database connection pool exhaustion | High | Low | Monitor pool metrics, implement request queuing, add read replicas |

---

## 10. Success Criteria

### Launch Criteria (MVP - Phase 1)
- ✅ Google Sign-In functional on all platforms
- ✅ Apple Sign-In functional on iOS
- ✅ User → Workspace mapping functional
- ✅ Token verification < 50ms (p95)
- ✅ New user registration < 300ms (p95)
- ✅ Audit logging operational
- ✅ Integration tests passing (Firebase emulator)
- ✅ Security review completed

### Post-Launch Metrics (30 Days)
- **Adoption:** 80% of users choose Google Sign-In (vs email/password)
- **Performance:** 95% of authentications complete in < 500ms
- **Reliability:** 99.9% uptime (measured via health checks)
- **Security:** Zero authentication-related security incidents
- **User Satisfaction:** < 2% support tickets related to login issues

---

## 11. Future Enhancements

### Phase 2 (Post-MVP, Q2 2025)
- Email/password authentication (Firebase built-in)
- Multi-factor authentication (TOTP via Firebase)
- Workspace invitations and member management
- Device management UI (view/revoke sessions)
- Password reset flows (Firebase-managed)

### Phase 3 (Enterprise Features, Q3 2025)
- SAML SSO integration (for enterprise customers)
- SCIM user provisioning (automated user management)
- Advanced audit logs with compliance exports
- Organization-level authentication policies
- API key authentication (for programmatic access)

---

## 12. Out of Scope (Explicitly)

**Not Building:**
- ❌ Custom password hashing (Firebase handles this)
- ❌ Custom email verification (Firebase handles this)
- ❌ Custom OAuth implementation (Firebase handles this)
- ❌ Auth-specific localization tables (reuse existing system)
- ❌ Multi-region database deployment (single Cloud SQL for MVP)
- ❌ Custom refresh token logic (Firebase SDK handles this)
- ❌ User provisioning API (manual registration only in Phase 1)

**Why:**
- Firebase provides these features securely out-of-the-box
- Reduces development time by 4-6 weeks
- Eliminates security vulnerabilities from custom implementations
- Aligns with "build less, leverage more" principle

---

## 13. Implementation Phases

### Phase 1: Foundation (Week 1-2)
**Goal:** Token verification and user lookup working

- [ ] Implement Firebase SDK in API Gateway middleware
- [ ] Create migration 008: Add `firebase_uid`, `provider`, `last_login_at` to `users`
- [ ] Create auth service structure: `backend/auth-service/domain/`
- [ ] Implement `GetByFirebaseUID` in user repository
- [ ] Add Firebase emulator to `docker-compose.yml`
- [ ] Write unit tests for token verification

**Deliverable:** Can verify Firebase JWT and retrieve user profile

### Phase 2: Complete Authentication Flow (Week 3-4)
**Goal:** New user registration with workspace creation

- [ ] Implement `CompleteAuthentication` service method
- [ ] Add workspace creation logic (verify `workspaces` table exists)
- [ ] Create `auth_sessions` and `auth_audit_log` tables
- [ ] Add audit logging for all auth events
- [ ] Implement `POST /api/v1/auth/verify` endpoint
- [ ] Write integration tests with test database

**Deliverable:** End-to-end auth flow from client to database

### Phase 3: Social Providers (Week 5-6)
**Goal:** Google, Apple, Facebook login functional

- [ ] Configure Firebase console (Google OAuth, Apple Sign-In, Facebook)
- [ ] Test provider-specific flows on iOS/Android/Web
- [ ] Create `linked_accounts` table for provider tracking
- [ ] Implement account linking logic (if user changes provider)
- [ ] Add provider-specific user attributes (photo_url, email_verified)

**Deliverable:** All three social providers working in production

### Phase 4: Production Readiness (Week 7-8)
**Goal:** Secure, monitored, production-ready

- [ ] Add Redis caching for Firebase public keys
- [ ] Implement rate limiting (100 req/min per IP)
- [ ] Security audit (OWASP Top 10 checklist)
- [ ] Load testing (1,000 concurrent authentications)
- [ ] Set up Cloud Monitoring alerts
- [ ] Documentation: API specs, runbooks, architecture diagrams

**Deliverable:** Production deployment with monitoring

---

## 14. API Specifications

### POST /api/v1/auth/verify
**Description:** Verify Firebase JWT and complete backend authentication

**Request:**
```http
POST /api/v1/auth/verify
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "device_id": "device-uuid-123" // Optional, for session tracking
}
```

**Response (200 OK - Existing User):**
```json
{
  "user": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "firebase_uid": "firebase-uid-123",
    "email": "user@example.com",
    "username": "user_example",
    "email_verified": true,
    "photo_url": "https://lh3.googleusercontent.com/...",
    "provider": "google.com",
    "preferred_language_id": "550e8400-e29b-41d4-a716-446655440001",
    "last_login_at": "2025-01-27T10:30:00Z",
    "created_at": "2025-01-15T08:00:00Z"
  },
  "workspaces": [
    {
      "workspace_id": "650e8400-e29b-41d4-a716-446655440000",
      "name": "My Garden",
      "role": "admin",
      "created_at": "2025-01-15T08:00:05Z"
    }
  ],
  "session_id": "750e8400-e29b-41d4-a716-446655440000"
}
```

**Response (200 OK - New User):**
```json
{
  "user": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "firebase_uid": "firebase-uid-456",
    "email": "newuser@example.com",
    "username": "newuser_example",
    "email_verified": true,
    "provider": "apple.com",
    "created_at": "2025-01-27T10:35:00Z"
  },
  "workspaces": [
    {
      "workspace_id": "650e8400-e29b-41d4-a716-446655440001",
      "name": "newuser_example's Garden",
      "role": "admin",
      "created_at": "2025-01-27T10:35:01Z"
    }
  ],
  "session_id": "750e8400-e29b-41d4-a716-446655440001",
  "is_new_user": true
}
```

**Response (401 Unauthorized):**
```json
{
  "error": "Invalid or expired token",
  "code": "AUTH_INVALID_TOKEN"
}
```

### GET /api/v1/auth/me
**Description:** Get current user profile and workspaces

**Request:**
```http
GET /api/v1/auth/me
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response (200 OK):**
```json
{
  "user": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "username": "user_example",
    "email_verified": true,
    "photo_url": "https://lh3.googleusercontent.com/...",
    "provider": "google.com",
    "preferred_language_id": "550e8400-e29b-41d4-a716-446655440001",
    "location": {"lat": 37.7749, "lng": -122.4194}
  },
  "workspaces": [
    {
      "workspace_id": "650e8400-e29b-41d4-a716-446655440000",
      "name": "My Garden",
      "role": "admin",
      "member_count": 1
    },
    {
      "workspace_id": "650e8400-e29b-41d4-a716-446655440002",
      "name": "Client: Smith",
      "role": "member",
      "member_count": 3
    }
  ]
}
```

### POST /api/v1/auth/logout
**Description:** Logout and revoke session

**Request:**
```http
POST /api/v1/auth/logout
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "device_id": "device-uuid-123", // Optional
  "revoke_all_sessions": false    // Default: false
}
```

**Response (200 OK):**
```json
{
  "message": "Logged out successfully",
  "sessions_revoked": 1
}
```

---

## 15. Compliance and Documentation

### Internal Documentation
- ✅ Architecture: `backend/auth-service/docs/architecture.md` (Go-aligned version)
- ✅ PRD: `backend/auth-service/docs/prd.md` (this document)
- [ ] API Documentation: Auto-generated Swagger/OpenAPI
- [ ] Runbook: Incident response procedures
- [ ] Privacy Policy: GDPR/CCPA compliance documentation

### External Standards
- [Firebase Auth Best Practices](https://firebase.google.com/docs/auth/admin/manage-users)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [GDPR Article 17 - Right to Erasure](https://gdpr-info.eu/art-17-gdpr/)
- [PostgreSQL Row-Level Security](https://www.postgresql.org/docs/17/ddl-rowsecurity.html)

---

## 16. Appendices

### A. Comparison: Original PRD vs Aligned PRD

| Aspect | Original PRD | Aligned PRD (v2.0) |
|--------|--------------|-------------------|
| Backend Language | Node.js/TypeScript | Go 1.22+ |
| Firebase Integration | Admin SDK + full sync | Token verification only |
| Database Strategy | Auth-specific tables | Extend existing `users` table |
| Localization | Auth-specific tables (90 languages) | Reuse existing system |
| Deployment | Multi-region (US, EU, APAC) | Single region (shared Cloud SQL) |
| Scale Target | 100K DAU globally | 1K concurrent (current stage) |
| Workspace Integration | Not mentioned | Core feature from day one |

### B. Firebase Configuration Checklist

**Firebase Console Setup:**
- [ ] Create Firebase project: `twigger-prod`
- [ ] Enable Google Sign-In provider
- [ ] Enable Apple Sign-In provider (requires Apple Developer account)
- [ ] Enable Facebook Login provider (requires Facebook App ID)
- [ ] Configure OAuth redirect URIs (iOS bundle ID, Android package name, web domain)
- [ ] Download `google-services.json` (Android) and `GoogleService-Info.plist` (iOS)
- [ ] Generate service account key for backend (Go Firebase Admin SDK)

**Environment Variables:**
```bash
FIREBASE_PROJECT_ID=twigger-prod
FIREBASE_CREDENTIALS_PATH=/secrets/firebase-admin-key.json
AUTH_ENABLED=true  # Set to false for local development bypass
```

### C. Migration Script (008_add_auth_fields.up.sql)

```sql
-- Migration: 008_add_auth_fields.up.sql
-- Add authentication fields to existing users table

ALTER TABLE users
ADD COLUMN firebase_uid VARCHAR(128) UNIQUE,
ADD COLUMN email_verified BOOLEAN DEFAULT FALSE,
ADD COLUMN phone_number VARCHAR(20),
ADD COLUMN photo_url TEXT,
ADD COLUMN provider VARCHAR(50),
ADD COLUMN last_login_at TIMESTAMPTZ,
ADD COLUMN deleted_at TIMESTAMPTZ;

-- Add indexes
CREATE INDEX idx_users_firebase_uid ON users(firebase_uid) WHERE firebase_uid IS NOT NULL;
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_users_last_login ON users(last_login_at DESC);

-- Create auth_sessions table
CREATE TABLE auth_sessions (
    session_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    device_id VARCHAR(255),
    device_info JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_auth_sessions_user_id ON auth_sessions(user_id);
CREATE INDEX idx_auth_sessions_expires_at ON auth_sessions(expires_at);

-- Create auth_audit_log table (partitioned)
CREATE TABLE auth_audit_log (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(user_id),
    event_type VARCHAR(50) NOT NULL,
    success BOOLEAN NOT NULL,
    ip_address INET,
    user_agent TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
) PARTITION BY RANGE (created_at);

-- Create first partition
CREATE TABLE auth_audit_log_2025_01 PARTITION OF auth_audit_log
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE INDEX idx_auth_audit_user_id ON auth_audit_log(user_id);
CREATE INDEX idx_auth_audit_created_at ON auth_audit_log(created_at);
CREATE INDEX idx_auth_audit_event_type ON auth_audit_log(event_type);

-- Create linked_accounts table
CREATE TABLE linked_accounts (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    linked_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider, provider_user_id)
);

CREATE INDEX idx_linked_accounts_user_id ON linked_accounts(user_id);
```

---

**PRD Version**: 2.0 (Aligned with Go-based Twigger System Architecture)
**Last Updated**: 2025-01-27
**Next Review**: Upon completion of Phase 1 implementation
**Status**: Ready for implementation
