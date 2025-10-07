# Authentication Service Setup Guide

## Overview
This guide covers setting up the authentication service for local development and production deployment.

## Prerequisites
- Go 1.22+
- Docker & Docker Compose
- PostgreSQL 17 with PostGIS (via Docker)
- Firebase project (for production)

---

## Environment Variables

### Required Variables

```bash
# Database Configuration
DB_HOST=localhost                    # Database host (use /cloudsql/... for Cloud SQL)
DB_PORT=5432                        # Database port
DB_USER=postgres                    # Database user
DB_PASSWORD=postgres                # Database password
DB_NAME=twigger                     # Database name
DB_SSLMODE=disable                  # SSL mode (require for production)

# Firebase Configuration
FIREBASE_PROJECT_ID=twigger-dev     # Firebase project ID
FIREBASE_CREDENTIALS_PATH=/path/to/firebase-admin-key.json  # Service account key path
AUTH_ENABLED=true                   # Enable authentication (set false for dev bypass)

# Service Configuration
ENVIRONMENT=dev                     # Environment: dev, staging, prod
LOG_LEVEL=debug                     # Log level: debug, info, warn, error
PORT=8080                          # API Gateway port

# Redis Configuration (optional, for caching)
REDIS_HOST=localhost
REDIS_PORT=6379
```

### Optional Variables

```bash
# Auth-specific settings
SESSION_EXPIRY_HOURS=720           # Session expiry (default: 30 days)
MAX_FAILED_LOGINS=5                # Max failed login attempts before lockout
LOCKOUT_DURATION_MINUTES=15        # Account lockout duration

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_MINUTE=100 # Rate limit per IP
```

---

## Local Development Setup

### Step 1: Start Docker Services

```bash
# Start PostgreSQL, Firebase Emulator, and Redis
docker-compose up -d

# Verify services are running
docker-compose ps
```

Expected output:
```
NAME                          STATUS              PORTS
twigger-postgres             Up                  0.0.0.0:5432->5432/tcp
twigger-firebase-emulator    Up                  0.0.0.0:9099->9099/tcp, 0.0.0.0:4000->4000/tcp
twigger-redis                Up                  0.0.0.0:6379->6379/tcp
```

### Step 2: Run Database Migrations

```bash
# Run migrations
go run cmd/migrate/main.go up

# Verify migration 008 applied
psql -h localhost -U postgres -d twigger -c "\dt"
```

Expected tables:
- `users` (with auth fields: firebase_uid, provider, etc.)
- `workspaces`
- `workspace_members`
- `auth_sessions`
- `auth_audit_log` (partitioned)
- `linked_accounts`

### Step 3: Configure Environment

Create `.env` file in project root:

```bash
# Copy example env file
cp .env.example .env

# Edit with your settings
nano .env
```

For local development:
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=twigger
DB_SSLMODE=disable

# Use Firebase Emulator for local dev
FIREBASE_PROJECT_ID=twigger-dev
AUTH_ENABLED=false  # Bypass auth for initial development

ENVIRONMENT=dev
LOG_LEVEL=debug
PORT=8080

REDIS_HOST=localhost
REDIS_PORT=6379
```

### Step 4: Run the API Gateway

```bash
# Start the API Gateway
go run cmd/api-gateway/main.go
```

Expected output:
```
[INFO] Starting Twigger API Gateway
[INFO] Environment: dev
[INFO] Auth enabled: false (development mode)
[INFO] Server listening on :8080
```

### Step 5: Test Authentication Flow

With `AUTH_ENABLED=false`, test the endpoints:

```bash
# Health check
curl http://localhost:8080/health

# Test protected endpoint (should work in dev mode)
curl http://localhost:8080/api/v1/auth/me
```

---

## Firebase Setup (Production)

### Step 1: Create Firebase Project

1. Go to [Firebase Console](https://console.firebase.google.com)
2. Click "Add project"
3. Enter project name: `twigger-prod`
4. Disable Google Analytics (optional)
5. Click "Create project"

### Step 2: Enable Authentication Providers

1. In Firebase Console, go to **Authentication** → **Sign-in method**
2. Enable **Google**:
   - Add OAuth client IDs for iOS, Android, Web
3. Enable **Apple Sign-In**:
   - Requires Apple Developer account ($99/year)
   - Add Team ID, Key ID, and Private Key
4. Enable **Facebook Login** (optional):
   - Add Facebook App ID and App Secret

### Step 3: Generate Service Account Key

1. In Firebase Console, go to **Project Settings** → **Service accounts**
2. Click "Generate new private key"
3. Download `firebase-admin-key.json`
4. **IMPORTANT**: Never commit this file to git
5. Store in Google Secret Manager for production:

```bash
# Upload to Secret Manager
gcloud secrets create firebase-admin-key --data-file=firebase-admin-key.json

# Grant access to service account
gcloud secrets add-iam-policy-binding firebase-admin-key \
  --member="serviceAccount:YOUR-SERVICE-ACCOUNT@PROJECT.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

### Step 4: Configure Production Environment

```bash
# Production environment variables (Cloud Run)
FIREBASE_PROJECT_ID=twigger-prod
FIREBASE_CREDENTIALS_PATH=/secrets/firebase-admin-key.json
AUTH_ENABLED=true

DB_HOST=/cloudsql/PROJECT:REGION:INSTANCE
DB_USER=twigger-api
DB_PASSWORD=${SECRET_DB_PASSWORD}  # From Secret Manager
DB_NAME=twigger
DB_SSLMODE=require

ENVIRONMENT=prod
LOG_LEVEL=info
```

---

## Testing

### Unit Tests

```bash
# Run all unit tests
go test ./backend/auth-service/...

# Run with coverage
go test -cover ./backend/auth-service/...

# Run specific test
go test -run TestCompleteAuthentication_NewUser ./backend/auth-service/domain/service
```

### Integration Tests

```bash
# Start test database
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test -tags=integration ./backend/auth-service/infrastructure/persistence/...

# Cleanup
docker-compose -f docker-compose.test.yml down -v
```

---

## Database Maintenance

### Create Monthly Audit Partitions

```sql
-- Create partition for next month (run monthly)
CREATE TABLE auth_audit_log_2025_03 PARTITION OF auth_audit_log
    FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');
```

Automate with cron or Cloud Scheduler:

```bash
# crontab example (run on 1st of each month)
0 0 1 * * psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "CREATE TABLE IF NOT EXISTS auth_audit_log_$(date +%Y_%m) PARTITION OF auth_audit_log FOR VALUES FROM ('$(date +%Y-%m-01)') TO ('$(date -d '+1 month' +%Y-%m-01)');"
```

### Cleanup Expired Sessions

```bash
# Run daily to delete expired sessions
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "DELETE FROM auth_sessions WHERE expires_at < NOW();"
```

---

## Troubleshooting

### Firebase Token Verification Fails

**Problem**: `Invalid token: failed to verify signature`

**Solution**:
1. Check `FIREBASE_PROJECT_ID` matches your Firebase project
2. Verify service account key is correct
3. Ensure system time is synchronized (NTP)
4. Check Firebase Admin SDK logs

### Database Connection Fails

**Problem**: `pq: password authentication failed`

**Solution**:
1. Verify credentials in `.env`
2. Check PostgreSQL is running: `docker-compose ps`
3. Test connection: `psql -h localhost -U postgres -d twigger`
4. Check `pg_hba.conf` allows connections from your IP

### Migration 008 Fails

**Problem**: `ERROR: column "firebase_uid" already exists`

**Solution**:
```bash
# Rollback migration
go run cmd/migrate/main.go down

# Reapply
go run cmd/migrate/main.go up
```

### Auth Middleware Not Working

**Problem**: All requests return 401 Unauthorized

**Solution**:
1. Set `AUTH_ENABLED=false` for development
2. Check token format: `Authorization: Bearer <token>`
3. Verify Firebase project ID matches
4. Check middleware is registered in router

---

## Security Checklist

### Development
- [ ] `AUTH_ENABLED=false` for local development only
- [ ] Never commit `firebase-admin-key.json` to git
- [ ] Use `.env` file (not tracked in git)
- [ ] Use Firebase Emulator for local auth testing

### Production
- [ ] `AUTH_ENABLED=true`
- [ ] Store secrets in Google Secret Manager
- [ ] Use Cloud SQL proxy or private IP
- [ ] Enable SSL/TLS (`DB_SSLMODE=require`)
- [ ] Set up rate limiting
- [ ] Monitor audit logs for suspicious activity
- [ ] Rotate Firebase service account keys quarterly

---

## Next Steps

1. **Phase 2**: Implement API handlers (`POST /api/v1/auth/verify`, `POST /api/v1/auth/logout`, `GET /api/v1/auth/me`)
2. **Phase 3**: Configure social providers in Firebase Console
3. **Phase 4**: Add Redis caching for public keys and user lookups
4. **Phase 5**: Implement rate limiting and security monitoring

---

## Additional Resources

- [Firebase Admin SDK Go Documentation](https://firebase.google.com/docs/admin/setup#go)
- [Firebase Authentication REST API](https://firebase.google.com/docs/reference/rest/auth)
- [PostgreSQL Partitioning Guide](https://www.postgresql.org/docs/17/ddl-partitioning.html)
- [Twigger Architecture Documentation](./architecture.md)
- [Twigger PRD](./prd.md)

---

**Last Updated**: 2025-01-27
**Version**: 2.0 (Phase 1 Complete)
