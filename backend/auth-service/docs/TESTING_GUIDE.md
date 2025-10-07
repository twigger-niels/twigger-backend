# Authentication Testing Guide

## Quick Start

### Run in Development Mode (No Firebase)
```bash
# Terminal 1: Start services
docker-compose up -d postgres redis

# Terminal 2: Start API Gateway (auth bypass)
DATABASE_URL="postgres://postgres:postgres@localhost:5432/twigger?sslmode=disable" \
AUTH_ENABLED=false \
ENVIRONMENT=development \
go run cmd/api-gateway/main.go
```

### Run with Firebase (Production Mode)
```bash
# Terminal 1: Start services
docker-compose up -d postgres redis

# Terminal 2: Start API Gateway with Firebase
DATABASE_URL="postgres://postgres:postgres@localhost:5432/twigger?sslmode=disable" \
FIREBASE_PROJECT_ID=twigger-prod \
FIREBASE_CREDENTIALS_PATH=./secrets/firebase-admin-key.json \
AUTH_ENABLED=true \
ENVIRONMENT=development \
LOG_LEVEL=debug \
go run cmd/api-gateway/main.go
```

---

## Test Endpoints (Development Mode)

When `AUTH_ENABLED=false`, any token works:

### 1. Verify Authentication
```bash
curl -X POST http://localhost:8080/api/v1/auth/verify \
  -H "Authorization: Bearer any-token" \
  -H "Content-Type: application/json" \
  -d '{"device_id": "test-device-123"}'
```

**Expected Response** (first time):
```json
{
  "user": {
    "user_id": "...",
    "firebase_uid": "dev-firebase-user",
    "email": "dev@example.com",
    "username": "dev_...",
    "email_verified": true,
    "provider": "email",
    "created_at": "2025-01-27T..."
  },
  "workspaces": [
    {
      "workspace_id": "...",
      "owner_id": "...",
      "name": "dev_...'s Garden",
      "created_at": "2025-01-27T..."
    }
  ],
  "session_id": "...",
  "is_new_user": true
}
```

### 2. Get Current User
```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer any-token"
```

### 3. Logout
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer any-token" \
  -H "Content-Type: application/json" \
  -d '{"device_id": "test-device-123", "revoke_all": false}'
```

---

## Test with Real Firebase Tokens

### Get a Firebase Token (JavaScript/Browser)

Create `test-firebase.html`:

```html
<!DOCTYPE html>
<html>
<head>
    <title>Firebase Auth Test</title>
    <script src="https://www.gstatic.com/firebasejs/10.7.1/firebase-app-compat.js"></script>
    <script src="https://www.gstatic.com/firebasejs/10.7.1/firebase-auth-compat.js"></script>
    <style>
        body { font-family: Arial; padding: 20px; }
        button { padding: 10px 20px; margin: 10px; font-size: 16px; }
        pre { background: #f4f4f4; padding: 15px; border-radius: 5px; }
        .token { word-break: break-all; }
    </style>
</head>
<body>
    <h1>Firebase Authentication Test</h1>

    <button id="google-signin">🔐 Sign in with Google</button>
    <button id="signout">Sign Out</button>
    <button id="refresh-token">Refresh Token</button>

    <h3>User Info:</h3>
    <pre id="user-info">Not signed in</pre>

    <h3>Firebase Token:</h3>
    <pre id="token" class="token">No token yet</pre>

    <h3>Test API:</h3>
    <button id="test-verify">Test /auth/verify</button>
    <button id="test-me">Test /auth/me</button>
    <button id="test-logout">Test /auth/logout</button>

    <h3>API Response:</h3>
    <pre id="api-response"></pre>

    <script>
        const firebaseConfig = {
            apiKey: "YOUR-WEB-API-KEY",
            authDomain: "twigger-prod.firebaseapp.com",
            projectId: "twigger-prod"
        };

        firebase.initializeApp(firebaseConfig);
        const auth = firebase.auth();

        let currentToken = null;

        auth.onAuthStateChanged(async (user) => {
            if (user) {
                document.getElementById('user-info').textContent = JSON.stringify({
                    uid: user.uid,
                    email: user.email,
                    displayName: user.displayName,
                    photoURL: user.photoURL,
                    emailVerified: user.emailVerified
                }, null, 2);

                currentToken = await user.getIdToken();
                document.getElementById('token').textContent = currentToken;
            } else {
                document.getElementById('user-info').textContent = 'Not signed in';
                document.getElementById('token').textContent = 'No token yet';
                currentToken = null;
            }
        });

        document.getElementById('google-signin').addEventListener('click', async () => {
            const provider = new firebase.auth.GoogleAuthProvider();
            try {
                await auth.signInWithPopup(provider);
            } catch (error) {
                alert('Error: ' + error.message);
            }
        });

        document.getElementById('signout').addEventListener('click', () => {
            auth.signOut();
        });

        document.getElementById('refresh-token').addEventListener('click', async () => {
            const user = auth.currentUser;
            if (user) {
                currentToken = await user.getIdToken(true);
                document.getElementById('token').textContent = currentToken;
                alert('Token refreshed!');
            } else {
                alert('Not signed in');
            }
        });

        async function callAPI(url, method = 'GET', body = null) {
            if (!currentToken) {
                alert('Please sign in first');
                return;
            }

            try {
                const options = {
                    method,
                    headers: {
                        'Authorization': 'Bearer ' + currentToken,
                        'Content-Type': 'application/json'
                    }
                };

                if (body) {
                    options.body = JSON.stringify(body);
                }

                const response = await fetch(url, options);
                const data = await response.json();

                document.getElementById('api-response').textContent = JSON.stringify(data, null, 2);
            } catch (error) {
                document.getElementById('api-response').textContent = 'Error: ' + error.message;
            }
        }

        document.getElementById('test-verify').addEventListener('click', () => {
            callAPI('http://localhost:8080/api/v1/auth/verify', 'POST', {
                device_id: 'browser-test-' + Date.now()
            });
        });

        document.getElementById('test-me').addEventListener('click', () => {
            callAPI('http://localhost:8080/api/v1/auth/me');
        });

        document.getElementById('test-logout').addEventListener('click', () => {
            callAPI('http://localhost:8080/api/v1/auth/logout', 'POST', {
                revoke_all: false
            });
        });
    </script>
</body>
</html>
```

**Setup**:
1. Get your Web API Key from Firebase Console → Project settings → General
2. Replace `YOUR-WEB-API-KEY` in the HTML
3. Open `test-firebase.html` in browser
4. Click "Sign in with Google"
5. Test the API buttons

---

## Test with cURL (Production Tokens)

### 1. Get Token from Browser Test
Copy the token from the HTML test above, then:

```bash
TOKEN="paste-your-firebase-token-here"

# Test verify endpoint
curl -X POST http://localhost:8080/api/v1/auth/verify \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"device_id": "curl-test"}'

# Test me endpoint
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

---

## Database Inspection

### Check Created Users
```bash
docker-compose exec postgres psql -U postgres -d twigger -c "
  SELECT user_id, email, username, firebase_uid, provider, email_verified, created_at
  FROM users
  ORDER BY created_at DESC
  LIMIT 5;
"
```

### Check Workspaces
```bash
docker-compose exec postgres psql -U postgres -d twigger -c "
  SELECT w.workspace_id, w.name, w.owner_id, wm.role
  FROM workspaces w
  JOIN workspace_members wm ON w.workspace_id = wm.workspace_id
  ORDER BY w.created_at DESC
  LIMIT 5;
"
```

### Check Sessions
```bash
docker-compose exec postgres psql -U postgres -d twigger -c "
  SELECT session_id, user_id, device_id, ip_address, created_at, expires_at, revoked_at
  FROM auth_sessions
  ORDER BY created_at DESC
  LIMIT 5;
"
```

### Check Audit Log
```bash
docker-compose exec postgres psql -U postgres -d twigger -c "
  SELECT id, user_id, event_type, success, ip_address, created_at
  FROM auth_audit_log
  ORDER BY created_at DESC
  LIMIT 10;
"
```

---

## Common Issues

### Issue: "Firebase not initialized"
```bash
# Check environment variables
echo $FIREBASE_PROJECT_ID
echo $FIREBASE_CREDENTIALS_PATH

# Verify file exists
ls -l $FIREBASE_CREDENTIALS_PATH
```

### Issue: "Invalid token"
- Token expired (15-minute lifetime)
- Wrong project ID
- Token from different environment

**Solution**: Refresh token or get new one

### Issue: "Connection refused"
```bash
# Check if services are running
docker-compose ps

# Check if API Gateway is running
curl http://localhost:8080/health
```

### Issue: Database connection error
```bash
# Check PostgreSQL is running
docker-compose exec postgres pg_isready -U postgres

# Test connection
psql postgres://postgres:postgres@localhost:5432/twigger -c "SELECT 1"
```

---

## Expected API Response Codes

| Endpoint | Method | Success | Errors |
|----------|--------|---------|--------|
| `/health` | GET | 200 | 500 |
| `/api/v1/auth/verify` | POST | 200 | 400, 401, 500 |
| `/api/v1/auth/me` | GET | 200 | 401, 404, 500 |
| `/api/v1/auth/logout` | POST | 200 | 401, 500 |

---

## Performance Benchmarks

Target performance (p95):
- Token verification: < 50ms
- User lookup: < 20ms
- Complete auth (existing): < 100ms
- Complete auth (new user): < 300ms

Test with:
```bash
# Install Apache Bench
# apt-get install apache2-utils  # Ubuntu
# brew install apache2  # Mac

# Benchmark /health endpoint
ab -n 1000 -c 10 http://localhost:8080/health

# Benchmark with auth (requires valid token)
ab -n 100 -c 5 -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/auth/me
```

---

## Next Steps

1. ✅ Test development mode (no Firebase)
2. ⏳ Configure Firebase providers
3. ⏳ Test with real Firebase tokens
4. ⏳ Run integration tests
5. ⏳ Performance testing

---

**Last Updated**: 2025-01-27
**For Production Setup**: See [FIREBASE_SETUP.md](./FIREBASE_SETUP.md)
