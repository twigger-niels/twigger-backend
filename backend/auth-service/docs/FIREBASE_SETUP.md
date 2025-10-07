# Firebase Setup Guide for Twigger Authentication

## Overview
This guide walks you through configuring Firebase Authentication for the Twigger backend, including social login providers (Google, Apple, Facebook).

## Prerequisites
- Firebase project created: `twigger-prod`
- Google Cloud Platform access (for service account)
- Apple Developer account (for Apple Sign-In, $99/year)
- Facebook Developer account (for Facebook Login, free)

---

## Step 1: Download Service Account Key

### 1.1 Navigate to Firebase Console
1. Go to [Firebase Console](https://console.firebase.google.com)
2. Select your project: `twigger-prod`

### 1.2 Generate Service Account Key
1. Click the gear icon (⚙️) → **Project settings**
2. Go to **Service accounts** tab
3. Click **Generate new private key**
4. Click **Generate key** to download `twigger-prod-firebase-adminsdk-xxxxx.json`

### 1.3 Store Securely
**⚠️ NEVER commit this file to git!**

```bash
# Local development
mkdir -p secrets
mv ~/Downloads/twigger-prod-firebase-adminsdk-*.json secrets/firebase-admin-key.json

# Add to .gitignore (should already be there)
echo "secrets/" >> .gitignore
```

---

## Step 2: Configure Google Sign-In

### 2.1 Enable Google Provider
1. In Firebase Console → **Authentication** → **Sign-in method**
2. Click **Google** → **Enable**
3. Add support email: your-email@example.com
4. Click **Save**

### 2.2 Configure OAuth Consent Screen (GCP Console)
1. Go to [GCP Console](https://console.cloud.google.com)
2. Select project: `twigger-prod`
3. Go to **APIs & Services** → **OAuth consent screen**
4. Choose **External** → **Create**
5. Fill in:
   - App name: `Twigger`
   - User support email: your-email@example.com
   - Developer contact: your-email@example.com
6. Click **Save and Continue**
7. Skip **Scopes** (default is fine)
8. Add test users if needed
9. Click **Save and Continue**

### 2.3 Get OAuth Client IDs
For Firebase to work with Google Sign-In on different platforms:

#### Web Client ID
1. In Firebase Console → **Authentication** → **Sign-in method** → **Google**
2. Expand **Web SDK configuration**
3. Copy **Web client ID** (format: `xxxxx.apps.googleusercontent.com`)

#### iOS Client ID
1. In Firebase Console → **Project settings** → **General**
2. Scroll to **Your apps** → Click iOS app (or add one)
3. Copy **iOS client ID**

#### Android Client ID
1. In Firebase Console → **Project settings** → **General**
2. Scroll to **Your apps** → Click Android app (or add one)
3. Add SHA-1 fingerprint (for production)
4. Copy **Android client ID**

---

## Step 3: Configure Apple Sign-In

### 3.1 Apple Developer Account Setup
**Requirement**: Apple Developer Program membership ($99/year)

1. Go to [Apple Developer](https://developer.apple.com/account)
2. Go to **Certificates, Identifiers & Profiles**

### 3.2 Create App ID
1. Click **Identifiers** → **+** (Add)
2. Select **App IDs** → **Continue**
3. Fill in:
   - Description: `Twigger iOS`
   - Bundle ID: `com.twigger.app` (or your bundle ID)
4. Check **Sign in with Apple** capability
5. Click **Continue** → **Register**

### 3.3 Create Service ID
1. Click **Identifiers** → **+** (Add)
2. Select **Services IDs** → **Continue**
3. Fill in:
   - Description: `Twigger Web`
   - Identifier: `com.twigger.web`
4. Check **Sign in with Apple**
5. Click **Configure** next to Sign in with Apple:
   - Primary App ID: Select your App ID from above
   - Domains: `twigger.com` (or your domain)
   - Return URLs: `https://twigger-prod.firebaseapp.com/__/auth/handler`
6. Click **Continue** → **Register**

### 3.4 Create Key
1. Click **Keys** → **+** (Add)
2. Fill in:
   - Key Name: `Twigger Apple Sign-In Key`
3. Check **Sign in with Apple**
4. Click **Configure** → Select your Primary App ID
5. Click **Continue** → **Register**
6. **Download** the key file (`.p8`) - **YOU CAN ONLY DOWNLOAD ONCE!**
7. Note the **Key ID** (e.g., `ABC123DEFG`)

### 3.5 Enable in Firebase
1. In Firebase Console → **Authentication** → **Sign-in method**
2. Click **Apple** → **Enable**
3. Fill in:
   - **Services ID**: `com.twigger.web` (from Step 3.3)
   - **Apple Team ID**: Found in Apple Developer → Membership
   - **Key ID**: From Step 3.4
   - **Private Key**: Open the `.p8` file and paste contents
4. Click **Save**

---

## Step 4: Configure Facebook Login

### 4.1 Create Facebook App
1. Go to [Facebook Developers](https://developers.facebook.com/)
2. Click **My Apps** → **Create App**
3. Select **Consumer** → **Next**
4. Fill in:
   - App name: `Twigger`
   - App contact email: your-email@example.com
5. Click **Create App**

### 4.2 Add Facebook Login Product
1. In app dashboard → Click **Add Product**
2. Find **Facebook Login** → Click **Set Up**
3. Select **Web** (or iOS/Android)
4. Skip quick start

### 4.3 Configure OAuth Redirect URIs
1. In left sidebar → **Facebook Login** → **Settings**
2. Add to **Valid OAuth Redirect URIs**:
   ```
   https://twigger-prod.firebaseapp.com/__/auth/handler
   ```
3. Click **Save Changes**

### 4.4 Get App Credentials
1. In left sidebar → **Settings** → **Basic**
2. Copy **App ID** (e.g., `1234567890`)
3. Click **Show** next to **App Secret** → Copy

### 4.5 Enable in Firebase
1. In Firebase Console → **Authentication** → **Sign-in method**
2. Click **Facebook** → **Enable**
3. Fill in:
   - **App ID**: From Step 4.4
   - **App secret**: From Step 4.4
4. Copy the **OAuth redirect URI** shown
5. Go back to Facebook app settings and add this URI if different
6. Click **Save** in Firebase

### 4.6 Make App Live
1. In Facebook app dashboard → **Settings** → **Basic**
2. Toggle **App Mode** to **Live** (top right)

---

## Step 5: Environment Configuration

### 5.1 Local Development (.env)
Create `.env` file in project root:

```bash
# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/twigger?sslmode=disable

# Firebase
FIREBASE_PROJECT_ID=twigger-prod
FIREBASE_CREDENTIALS_PATH=./secrets/firebase-admin-key.json
AUTH_ENABLED=true

# Server
ENVIRONMENT=development
LOG_LEVEL=debug
PORT=8080

# Redis (optional for caching)
REDIS_HOST=localhost
REDIS_PORT=6379
```

### 5.2 Production (Cloud Run)
Set environment variables in Cloud Run:

```bash
# Deploy with environment variables
gcloud run deploy api-gateway \
  --set-env-vars="FIREBASE_PROJECT_ID=twigger-prod" \
  --set-env-vars="AUTH_ENABLED=true" \
  --set-env-vars="ENVIRONMENT=production" \
  --set-env-vars="LOG_LEVEL=info"
```

**For credentials**, use Secret Manager:

```bash
# Upload service account key to Secret Manager
gcloud secrets create firebase-admin-key \
  --data-file=./secrets/firebase-admin-key.json

# Grant Cloud Run service account access
gcloud secrets add-iam-policy-binding firebase-admin-key \
  --member="serviceAccount:YOUR-SERVICE-ACCOUNT@PROJECT.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"

# Mount secret in Cloud Run
gcloud run deploy api-gateway \
  --update-secrets="/secrets/firebase-admin-key.json=firebase-admin-key:latest"

# Set environment variable
gcloud run services update api-gateway \
  --set-env-vars="FIREBASE_CREDENTIALS_PATH=/secrets/firebase-admin-key.json"
```

---

## Step 6: Test Firebase Integration

### 6.1 Start the API Gateway
```bash
# Load environment variables
export $(cat .env | xargs)

# Start server
go run cmd/api-gateway/main.go
```

Expected output:
```
Successfully connected to database
Successfully initialized Firebase
Starting API Gateway on port 8080
```

### 6.2 Test Health Check
```bash
curl http://localhost:8080/health
```

### 6.3 Get a Firebase Token (Web)
Create a simple HTML file to test:

```html
<!DOCTYPE html>
<html>
<head>
    <title>Firebase Test</title>
    <script src="https://www.gstatic.com/firebasejs/10.7.1/firebase-app-compat.js"></script>
    <script src="https://www.gstatic.com/firebasejs/10.7.1/firebase-auth-compat.js"></script>
</head>
<body>
    <button id="google-signin">Sign in with Google</button>
    <pre id="token"></pre>

    <script>
        const firebaseConfig = {
            apiKey: "YOUR-API-KEY",
            authDomain: "twigger-prod.firebaseapp.com",
            projectId: "twigger-prod"
        };

        firebase.initializeApp(firebaseConfig);

        const provider = new firebase.auth.GoogleAuthProvider();

        document.getElementById('google-signin').addEventListener('click', async () => {
            try {
                const result = await firebase.auth().signInWithPopup(provider);
                const token = await result.user.getIdToken();
                document.getElementById('token').textContent = token;
                console.log('Token:', token);
            } catch (error) {
                console.error('Error:', error);
            }
        });
    </script>
</body>
</html>
```

Replace `YOUR-API-KEY` with your Firebase Web API Key from:
**Firebase Console** → **Project settings** → **General** → **Web API Key**

### 6.4 Test Auth Endpoints
```bash
# Get the token from the HTML test above, then:

# Verify token and complete authentication
curl -X POST http://localhost:8080/api/v1/auth/verify \
  -H "Authorization: Bearer YOUR-FIREBASE-TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"device_id": "test-device-123"}'

# Get current user
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer YOUR-FIREBASE-TOKEN"

# Logout
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer YOUR-FIREBASE-TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"revoke_all": false}'
```

---

## Troubleshooting

### Error: "Firebase not initialized"
**Solution**:
- Check `FIREBASE_PROJECT_ID` is set
- Check `FIREBASE_CREDENTIALS_PATH` points to valid file
- Verify service account key is not corrupted

### Error: "Invalid token"
**Possible causes**:
1. Token expired (15-minute lifetime)
2. Wrong Firebase project
3. Token from different environment (dev vs prod)

**Solution**:
```bash
# Verify project ID matches
echo $FIREBASE_PROJECT_ID

# Check token claims (decode JWT at jwt.io)
# Ensure "aud" and "iss" match your project
```

### Error: "failed to verify token: ID token has been revoked"
**Solution**: User signed out or account disabled. Get a fresh token.

### Google Sign-In doesn't work on mobile
**Solution**:
- Add SHA-1 fingerprints to Firebase project (Android)
- Configure iOS client ID in GoogleService-Info.plist (iOS)

### Apple Sign-In returns "invalid_client"
**Solution**:
- Verify Services ID matches in Firebase
- Check return URL is exactly: `https://twigger-prod.firebaseapp.com/__/auth/handler`
- Ensure App ID has Sign in with Apple capability enabled

---

## Security Checklist

- [ ] Service account key stored in Secret Manager (production)
- [ ] Service account key NOT committed to git
- [ ] OAuth redirect URIs only include production domains
- [ ] Firebase security rules configured
- [ ] Rate limiting enabled (100 req/min)
- [ ] Audit logging enabled
- [ ] HTTPS enforced (production)
- [ ] CORS configured for allowed origins only

---

## Next Steps

1. ✅ Firebase project created
2. ✅ Service account key generated
3. ⏳ Configure social providers (Google, Apple, Facebook)
4. ⏳ Test authentication flow
5. ⏳ Deploy to Cloud Run

See [SETUP.md](./SETUP.md) for general authentication setup and [README.md](../README.md) for architecture overview.

---

**Last Updated**: 2025-01-27
**Project**: twigger-prod
**Status**: Ready for provider configuration
