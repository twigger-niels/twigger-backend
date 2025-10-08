# Pre-Deployment Checklist

This document contains all tests and external system configurations that must be completed before deploying to production.

## Table of Contents

1. [Critical Security Fixes Testing](#critical-security-fixes-testing)
2. [External System Configuration](#external-system-configuration)
3. [Authentication Flow Testing](#authentication-flow-testing)
4. [Performance Testing](#performance-testing)
5. [Database Verification](#database-verification)
6. [Environment Variables](#environment-variables)

---

## Critical Security Fixes Testing

### 1. Account Takeover Prevention (CRITICAL-3)

**What Changed**: Provider linking now requires users to sign in with their original provider.

**Test Case 1: Prevent Cross-Provider Account Takeover**
```
Steps:
1. Register user with Google: alice@example.com
2. Attempt to sign in with Facebook using same email: alice@example.com
3. Verify behavior

Expected Result:
- Backend returns error: "This email is already registered with google.com. Please sign in using google.com"
- Account is NOT linked automatically
- User remains logged out

Audit Log Check:
- Event type: "account_linking_blocked" should appear in auth_audit_log
- Check metadata includes: existing_provider, attempted_provider, reason
```

**Test Case 2: Same Provider Re-authentication**
```
Steps:
1. Register user with Google: bob@example.com
2. Sign out
3. Sign in again with Google: bob@example.com
4. Verify behavior

Expected Result:
- User successfully authenticated
- No new user created (existing account used)
- Last login timestamp updated
```

**Files Changed**:
- `backend/auth-service/domain/service/auth_service.go:192-205`

---

### 2. Race Condition in Username Generation (CRITICAL-1)

**What Changed**: Username generation now uses database constraints with retry logic (10 attempts) instead of pre-check.

**Test Case 1: Concurrent Registration with Same Email Prefix**
```
Setup:
- Use load testing tool (e.g., Apache Bench, k6, or Artillery)
- Simulate 50 concurrent registrations with emails: test01@example.com through test50@example.com

Steps:
1. Send 50 concurrent POST requests to /api/v1/auth/register
2. All should complete successfully within 10 seconds
3. Check database for username conflicts

Expected Result:
- All 50 users created successfully
- Usernames: test01 (or test01_a3f9c2b1 if conflict), test02, test03, etc.
- No database constraint violations
- No failed registrations

Database Query to Verify:
```sql
SELECT username, COUNT(*)
FROM users
WHERE email LIKE 'test%@example.com'
GROUP BY username
HAVING COUNT(*) > 1;
-- Should return 0 rows (no duplicates)
```
```

**Test Case 2: Username Generation Exhaustion**
```
This is a stress test to verify the system doesn't crash when retries are exhausted.

Steps:
1. Manually create users with usernames: john_doe, john_doe_aaaaaaaa through john_doe_zzzzzzzz
2. Attempt to register with email: john.doe@example.com
3. Verify behavior

Expected Result:
- After 10 retry attempts, returns error: "failed to create user after 10 attempts: username generation exhausted"
- No partial user creation (transaction rolled back)
- Audit log shows registration failure
```

**Test Case 3: Custom Username Conflict**
```
Steps:
1. Register user with custom username: "plantlover"
2. Attempt to register another user with same custom username: "plantlover"
3. Verify behavior

Expected Result:
- Second registration fails with: "Username already taken. Please choose another."
- First user remains in database
- No retry attempts (custom usernames don't retry)
```

**Files Changed**:
- `backend/auth-service/domain/service/auth_service.go:593-664`

---

### 3. Information Disclosure Prevention (CRITICAL-2)

**What Changed**: Firebase UIDs removed from error logs, replaced with opaque error IDs.

**Test Case 1: Failed Email Verification Log Check**
```
Steps:
1. Create Firebase account but don't verify email
2. Attempt to call POST /api/v1/auth/register
3. Check backend logs

Expected Result:
- User sees: "Please verify your email address before completing registration..."
- Backend log shows: "email verification required [error_id: a3f9c2b1-...]"
- Log does NOT contain Firebase UID
- Log does NOT contain email_verified: false details

Log Pattern to Verify:
✅ GOOD: "email verification required [error_id: uuid]"
❌ BAD: "email not verified | Firebase UID: abc123"
```

**Test Case 2: Failed Authentication Log Check**
```
Steps:
1. Send malformed Firebase token to POST /api/v1/auth/verify
2. Check backend logs

Expected Result:
- User sees: "authentication failed"
- Backend log shows: "authentication failed [error_id: uuid]"
- Log does NOT contain Firebase UID or sensitive details

Security Check:
- Attacker cannot enumerate valid Firebase UIDs from logs
- Attacker cannot determine if email exists in system
```

**Files Changed**:
- `internal/api-gateway/handlers/auth_handler.go:110, 143, 287, 327`

---

## External System Configuration

### Firebase Console

**Email/Password Provider**
```
Location: Firebase Console > Authentication > Sign-in method

Settings to Verify:
✅ Email/Password: ENABLED
✅ Email link (passwordless sign-in): DISABLED (unless needed)
✅ Email enumeration protection: ENABLED (recommended)

Email Verification Settings:
✅ Require email verification: YES (enforced in backend)
✅ Template customization: Configure branded email templates
```

**Social Providers**
```
Location: Firebase Console > Authentication > Sign-in method

Google Sign-In:
✅ Status: ENABLED
✅ Web SDK configuration: Client ID configured
✅ Authorized domains: Include production domain (e.g., app.twigger.com)

Facebook Login:
✅ Status: ENABLED
✅ App ID: 763251526584065 (update for production)
✅ App Secret: Configured
✅ OAuth redirect URI: Whitelisted in Facebook Developer Console

Apple Sign-In (if applicable):
□ Status: TBD for Phase 4
```

**Authorized Domains**
```
Location: Firebase Console > Authentication > Settings > Authorized domains

Required Entries:
✅ localhost (for development)
✅ Production domain (e.g., app.twigger.com)
✅ Staging domain (e.g., staging.twigger.com)

Remove before production:
❌ Any test domains
❌ Developer personal domains
```

**Security Rules**
```
Location: Firebase Console > Firestore/Storage > Rules

Verify:
✅ No test mode rules in production
✅ All reads/writes require authentication
✅ User can only access their own data
```

### Facebook Developer Console

**OAuth Settings**
```
Location: Facebook App > Settings > Basic

Required Configuration:
✅ App ID: Matches Firebase configuration
✅ App Secret: Stored in Firebase
✅ Privacy Policy URL: Must be publicly accessible
✅ Terms of Service URL: Must be publicly accessible
✅ App Domains: Include production domain

Valid OAuth Redirect URIs:
✅ https://twigger-prod.firebaseapp.com/__/auth/handler
✅ https://app.twigger.com (if custom domain)
```

**App Review (Required for Production)**
```
Facebook requires app review before public access.

Required Permissions:
✅ public_profile (default)
✅ email (default)

Testing:
- Add test users in Facebook Developer Console > Roles > Test Users
- Verify login flow works for test users
- Submit for review before production launch
```

### Google Cloud Platform

**OAuth 2.0 Credentials**
```
Location: Google Cloud Console > APIs & Services > Credentials

Web Client ID:
✅ Authorized JavaScript origins: Include production domain
✅ Authorized redirect URIs: Include Firebase auth handler

Example:
- Authorized origins: https://app.twigger.com
- Redirect URIs: https://twigger-prod.firebaseapp.com/__/auth/handler
```

**API Restrictions**
```
Location: Google Cloud Console > APIs & Services > Credentials

For Production API Key:
✅ Application restrictions: HTTP referrers (web sites)
✅ Allowed referrers: Only production domains
✅ API restrictions: Restrict to Identity Toolkit API
```

---

## Authentication Flow Testing

### Email/Password Registration

**Complete Flow Test**
```
Steps:
1. Navigate to registration screen
2. Enter email: newuser@example.com
3. Enter password: Test123!@#
4. Confirm password: Test123!@#
5. Check terms of service
6. Click "Create Account"

Expected Behavior:
✅ Email verification screen appears immediately (no back button needed)
✅ Verification email received within 1 minute
✅ Email contains verification link
✅ Click link → redirects to Firebase verification success page
✅ Return to app → click "I've verified my email"
✅ Redirected to home page
✅ User profile shows auto-generated username (e.g., "newuser")

Database Verification:
```sql
SELECT user_id, email, username, email_verified, provider
FROM users
WHERE email = 'newuser@example.com';

Expected:
- email_verified: true
- provider: password
- username: newuser (or newuser_a3f9c2b1 if conflict)
```
```

**Username Auto-Generation Verification**
```
Test Cases:
| Email                    | Expected Username    | Notes                          |
|--------------------------|----------------------|--------------------------------|
| john.doe@example.com     | john_doe             | Period replaced with underscore|
| alice+test@example.com   | alice_test           | Plus replaced with underscore  |
| bob-smith@example.com    | bob_smith            | Hyphen replaced with underscore|
| admin@company.co.uk      | admin                | Domain ignored                 |
| test@example.com (dup)   | test_a3f9c2b1        | UUID suffix on conflict        |
```

### Social Login Testing

**Google Sign-In Flow**
```
Steps:
1. Click "Continue with Google"
2. Select Google account
3. Grant permissions (if first time)

Expected Behavior:
✅ Redirects to home page immediately (no email verification)
✅ User profile shows Google photo
✅ Username auto-generated from email
✅ Provider set to "google.com"
✅ Workspace created automatically

Database Verification:
- email_verified: true (Google provides verified emails)
- firebase_uid: Should start with Google UID pattern
- photo_url: Google profile picture URL
```

**Facebook Sign-In Flow**
```
Steps:
1. Click "Continue with Facebook"
2. Authorize app (if first time)

Expected Behavior:
✅ Same as Google flow
✅ Provider set to "facebook.com"
✅ Photo from Facebook profile (if available)

HTTPS Requirement:
⚠️ Facebook login only works on HTTPS in production
⚠️ Test on staging environment with valid SSL certificate
```

**Account Linking Prevention**
```
Steps:
1. Register with Google: test@example.com
2. Sign out
3. Attempt to sign in with Facebook using same email: test@example.com

Expected Behavior:
✅ Error message: "This email is already registered with google.com. Please sign in using google.com"
✅ User remains logged out
✅ No account created or modified
✅ Audit log shows "account_linking_blocked" event
```

---

## Performance Testing

### Concurrent User Registration

**Load Test Configuration**
```javascript
// Example using k6 (https://k6.io)
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 20 },  // Ramp up to 20 users
    { duration: '1m', target: 50 },   // Stay at 50 users
    { duration: '30s', target: 0 },   // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'], // 95% of requests under 2s
    http_req_failed: ['rate<0.01'],    // Less than 1% failure rate
  },
};

export default function () {
  const url = 'https://api.twigger.com/api/v1/auth/register';
  const payload = JSON.stringify({
    device_id: `device_${__VU}_${__ITER}`,
  });
  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${__ENV.FIREBASE_TOKEN}`,
    },
  };

  let response = http.post(url, payload, params);
  check(response, {
    'status is 200 or 201': (r) => r.status === 200 || r.status === 201,
    'has user_id': (r) => JSON.parse(r.body).user.user_id !== undefined,
  });

  sleep(1);
}
```

**Success Criteria**
- ✅ 95% of requests complete in under 2 seconds
- ✅ Less than 1% failure rate
- ✅ No database deadlocks or constraint violations
- ✅ All usernames unique (no duplicates)

### Database Query Performance

**Check Index Usage**
```sql
-- Verify indexes exist and are being used
EXPLAIN ANALYZE
SELECT * FROM users
WHERE email = 'test@example.com'
AND deleted_at IS NULL;

-- Should use index on email column
-- Execution time should be < 5ms

EXPLAIN ANALYZE
SELECT * FROM users
WHERE firebase_uid = 'abc123'
AND deleted_at IS NULL;

-- Should use index on firebase_uid column
-- Execution time should be < 5ms

EXPLAIN ANALYZE
SELECT COUNT(*) FROM users
WHERE username = 'john_doe'
AND deleted_at IS NULL;

-- Should use unique index on username column
-- Execution time should be < 1ms
```

---

## Database Verification

### Migration Status

**Check All Migrations Applied**
```bash
# Run migration status check
DATABASE_URL="postgres://user:pass@host:5432/twigger" \
  go run cmd/migrate/main.go version

Expected Output:
✅ All migrations up to date
✅ No pending migrations
✅ Current version: [latest migration number]
```

**Required Tables**
```sql
-- Verify all required tables exist
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
AND table_type = 'BASE TABLE'
ORDER BY table_name;

Required Tables:
✅ users
✅ workspaces
✅ workspace_members
✅ sessions
✅ auth_audit_log
✅ user_providers (if using multi-provider)
```

### Data Integrity Checks

**Username Uniqueness**
```sql
-- Check for duplicate usernames (should return 0 rows)
SELECT username, COUNT(*) as count
FROM users
WHERE deleted_at IS NULL
GROUP BY username
HAVING COUNT(*) > 1;

-- Expected: 0 rows
```

**Firebase UID Uniqueness**
```sql
-- Check for duplicate Firebase UIDs (should return 0 rows)
SELECT firebase_uid, COUNT(*) as count
FROM users
WHERE firebase_uid IS NOT NULL
AND deleted_at IS NULL
GROUP BY firebase_uid
HAVING COUNT(*) > 1;

-- Expected: 0 rows
```

**Orphaned Workspaces**
```sql
-- Check for workspaces without owners (should return 0 rows)
SELECT w.workspace_id, w.owner_id
FROM workspaces w
LEFT JOIN users u ON w.owner_id = u.user_id
WHERE u.user_id IS NULL;

-- Expected: 0 rows
```

**Audit Log Integrity**
```sql
-- Verify audit logs are being created
SELECT
  event_type,
  COUNT(*) as count,
  MAX(created_at) as last_event
FROM auth_audit_log
GROUP BY event_type
ORDER BY event_type;

Expected Event Types:
✅ user_registered
✅ user_login
✅ user_logout
✅ account_linking_blocked (if tested)
```

---

## Environment Variables

### Required Environment Variables (Production)

**Backend API Gateway**
```bash
# Database
DATABASE_URL="postgres://user:pass@prod-db-host:5432/twigger?sslmode=require"

# Firebase
FIREBASE_PROJECT_ID="twigger-prod"
FIREBASE_CREDENTIALS_PATH="./secrets/firebase-admin-key.json"

# Authentication
AUTH_ENABLED="true"

# Environment
ENVIRONMENT="production"
LOG_LEVEL="info"  # Not "debug" in production

# Server
PORT="8080"
```

**Secrets Management**
```bash
# NEVER commit these to version control:
❌ firebase-admin-key.json
❌ .env files with real credentials
❌ Database passwords

# Use secret management service:
✅ Google Secret Manager
✅ AWS Secrets Manager
✅ Azure Key Vault
✅ HashiCorp Vault
```

### Frontend Environment Variables

**Flutter Web Configuration**
```dart
// lib/core/config/app_config.dart

Production Values:
✅ apiBaseUrl: 'https://api.twigger.com'
✅ enableLogging: false (disable debug logs)
✅ environment: 'production'
✅ apiTimeout: Duration(seconds: 30)

// Firebase configuration in firebase_options.dart
// Generated via: flutterfire configure --project=twigger-prod
```

---

## Final Pre-Deployment Checklist

### Security Review

- [ ] All 3 critical security fixes tested and verified
- [ ] Firebase UIDs not exposed in logs
- [ ] Account linking prevention working
- [ ] Username race conditions resolved
- [ ] Error messages don't leak sensitive information
- [ ] HTTPS enforced on all production endpoints
- [ ] CORS configured for production domain only
- [ ] Database credentials rotated from development values

### Configuration Review

- [ ] Firebase authorized domains updated for production
- [ ] Facebook OAuth redirect URIs whitelisted
- [ ] Google Cloud API restrictions configured
- [ ] Environment variables set correctly (no dev values)
- [ ] Secret files (.env, service accounts) NOT committed to git
- [ ] Database migrations all applied
- [ ] Database indexes verified

### Testing Review

- [ ] Email verification flow tested end-to-end
- [ ] Social login tested (Google, Facebook)
- [ ] Account linking prevention verified
- [ ] Concurrent registration load tested (50+ users)
- [ ] Username generation tested with conflicts
- [ ] Error logging verified (no sensitive data)
- [ ] Audit logs populated correctly

### Performance Review

- [ ] API response times under 2 seconds (p95)
- [ ] Database queries using indexes (checked with EXPLAIN)
- [ ] No N+1 query issues
- [ ] Session management tested under load
- [ ] Connection pooling configured (max_connections)

### Monitoring Setup

- [ ] Application logs forwarded to monitoring service
- [ ] Error alerting configured (e.g., Sentry, Rollbar)
- [ ] Database performance monitoring enabled
- [ ] Audit log alerting for suspicious activity
- [ ] Uptime monitoring configured (e.g., UptimeRobot)

### Documentation Review

- [ ] API documentation updated
- [ ] Architecture diagrams current
- [ ] Runbooks created for common issues
- [ ] Rollback plan documented
- [ ] On-call rotation defined

---

## Rollback Plan

### If Critical Issues Found in Production

**Backend Rollback**
```bash
# Revert to previous version
kubectl rollout undo deployment/api-gateway  # If using Kubernetes

# Or redeploy previous version
docker pull twigger-backend:previous-tag
docker-compose up -d
```

**Database Rollback**
```bash
# Only if new migration causes issues
DATABASE_URL="production-db-url" \
  go run cmd/migrate/main.go down 1

# Verify rollback
DATABASE_URL="production-db-url" \
  go run cmd/migrate/main.go version
```

**Quick Fixes Without Full Rollback**
```bash
# Hotfix for critical security issue:
1. Create hotfix branch from production
2. Apply minimal fix
3. Test in staging
4. Deploy directly to production
5. Merge back to main
```

---

## Post-Deployment Verification

### Smoke Tests (First 15 Minutes)

- [ ] Health check endpoint responds: `GET /health`
- [ ] New user can register via email/password
- [ ] Email verification email received and works
- [ ] Existing user can log in
- [ ] Google sign-in works
- [ ] Facebook sign-in works
- [ ] No error spikes in monitoring
- [ ] Database connections stable
- [ ] Response times normal

### Extended Monitoring (First 24 Hours)

- [ ] User registration rate normal
- [ ] No unusual audit log patterns
- [ ] No repeated authentication failures
- [ ] Database query performance stable
- [ ] No memory leaks detected
- [ ] Session management working correctly

---

## Contact Information

**On-Call Engineer**: [TBD]
**Database Admin**: [TBD]
**Firebase Admin**: [TBD]

**Escalation**:
1. Check monitoring dashboards
2. Review error logs
3. Check audit logs for security issues
4. Contact on-call engineer
5. Initiate rollback if necessary

---

## Revision History

| Date       | Version | Changes                                    | Author |
|------------|---------|-------------------------------------------|--------|
| 2025-10-08 | 1.0     | Initial version with 3 critical fixes     | Claude |

---

**Next Review Date**: [Set quarterly review date]
