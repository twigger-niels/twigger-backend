# Firebase Setup Required

## Current Status
✅ Flutter app running on Chrome
✅ Backend API running with CORS fixed
✅ Firebase SDK initialized

❌ Firebase Authentication not configured
❌ Google Sign-In not configured for web

---

## Errors You're Seeing

### 1. **identitytoolkit.googleapis.com 400 Error**
```
Failed to load resource: the server responded with a status of 400
```

**Cause**: Email/Password authentication is not enabled in Firebase Console.

**Fix**: Go to [Firebase Console](https://console.firebase.google.com/project/twigger-prod/authentication/providers)
1. Click **Authentication** → **Sign-in method**
2. Click **Email/Password**
3. Enable both toggles:
   - ✅ Email/Password
   - ✅ Email link (passwordless sign-in) - Optional
4. Click **Save**

### 2. **Google Sign-In Error**
```
Access to fetch at 'http://localhost:8080/api/v1/auth/register' blocked by CORS
```

**Cause 1 (CORS)**: ✅ **FIXED** - Backend now allows all localhost origins

**Cause 2 (Google)**: Google Sign-In needs web client ID configuration

**Fix for Google Sign-In**:

#### Step 1: Enable Google Sign-In in Firebase
1. Go to [Firebase Console](https://console.firebase.google.com/project/twigger-prod/authentication/providers)
2. Click **Google** under Sign-in providers
3. Enable the toggle
4. Copy the **Web SDK configuration** → **Web client ID**
5. Click **Save**

#### Step 2: Add Client ID to Flutter Web
Edit `frontend/web/index.html` and add this line inside the `<head>` section (before `</head>`):

```html
<meta name="google-signin-client_id" content="YOUR_WEB_CLIENT_ID_HERE">
```

Replace `YOUR_WEB_CLIENT_ID_HERE` with the Web client ID from Step 1.

Example:
```html
<meta name="google-signin-client_id" content="53871324357-ld0us2jumg5hp2gkq09kao2luil0h0ke.apps.googleusercontent.com">
```

**Note**: I can see this client ID in `firebase_options.dart` (line 72), but you should verify it's the web client ID in Firebase Console.

### 3. **Facebook Login** (Not configured yet)
To enable Facebook login:
1. Create a Facebook App at https://developers.facebook.com/
2. Get App ID and App Secret
3. Add to Firebase Console → Authentication → Facebook
4. Configure OAuth redirect URI

---

## Quick Start Guide

### Minimum Setup (Email/Password Only)

**5 minutes to get auth working:**

1. **Enable Email/Password in Firebase Console**
   - Go to: https://console.firebase.google.com/project/twigger-prod/authentication/providers
   - Click "Email/Password" → Enable → Save

2. **Test the app**
   - Refresh the Flutter app in Chrome
   - Try registering a new user
   - Should work! ✅

### Full Setup (All Auth Methods)

**20 minutes for complete auth:**

1. ✅ Email/Password (see above)
2. Configure Google Sign-In (see Step 2 above)
3. Configure Facebook Login (optional, more complex)

---

## Testing Authentication

### Test Email/Password Registration

1. Open Flutter app in Chrome: http://localhost:54548 (or whatever port)
2. Click **"Don't have an account? Register"**
3. Fill in:
   - Username: `testuser`
   - Email: `test@example.com`
   - Password: `Test1234!`
   - Confirm password: `Test1234!`
4. Check "I agree to Terms & Conditions"
5. Click **Register**

**Expected Flow**:
- Firebase creates user account
- Email verification screen appears
- Check email for verification link
- Click link
- App detects verification and logs you in
- Backend creates user record via `/api/v1/auth/verify`
- You see the home screen!

### Test Email/Password Login

1. Click **Login** (if you registered)
2. Enter email and password
3. Click **Sign In**

**Expected**:
- Logged in immediately
- Home screen with bottom navigation

### Test Google Sign-In

1. Click **"Continue with Google"** button
2. Select Google account
3. Authorize app

**Expected**:
- Backend creates/links account
- Logged in
- Home screen appears

---

## Backend Integration

The Flutter app calls these endpoints:

1. **POST /api/v1/auth/verify** - Called after Firebase auth
   - Sends: Firebase ID token
   - Returns: User data with workspaces
   - Backend creates user if first time

2. **GET /api/v1/auth/me** - Get current user
   - Requires: Authorization header with Firebase token
   - Returns: User profile

3. **POST /api/v1/auth/logout** - Logout user
   - Requires: Authorization header
   - Invalidates session

All working! Backend is listening on `localhost:8080` with CORS enabled for all localhost origins.

---

## Troubleshooting

### "CORS policy" error persists
- ✅ Should be fixed now
- Backend allows all `http://localhost:*` origins
- If still seeing error, restart Flutter app with hot restart (R)

### "No Firebase App '[DEFAULT]'" error
- ✅ Should be fixed
- `main.dart` now imports `firebase_options.dart`
- If still seeing, clear cache: `flutter clean && flutter pub get`

### "ClientID not set" for Google
- This is expected if you haven't added the `<meta>` tag yet
- Google Sign-In won't work until you add it
- Email/Password will work fine without it

### Email verification not detected
- Click "I've Verified My Email" button manually
- Or wait up to 5 seconds for auto-detection
- Check that you clicked the link in your email

### Backend connection refused
- Ensure backend is running: `netstat -ano | findstr :8080`
- Should show: `LISTENING 54956` (or similar PID)
- If not running, start it (see main README)

---

## Next Steps

1. **Enable Email/Password in Firebase Console** (5 min)
2. **Test registration flow** (2 min)
3. **Optional**: Configure Google Sign-In (10 min)
4. **Optional**: Configure Facebook Login (20 min)

Once email/password is enabled, your auth system is fully functional! 🎉
