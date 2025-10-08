# Troubleshooting Guide - Firebase Authentication

**Version:** 1.0
**Date:** 2025-10-08
**Platforms:** iOS, Android, Web (Flutter)

---

## Overview

This guide provides solutions to common issues encountered when integrating Firebase Authentication with the Twigger backend. Issues are categorized by platform and component for easy navigation.

---

## Table of Contents

1. [Firebase Initialization Issues](#firebase-initialization-issues)
2. [iOS-Specific Issues](#ios-specific-issues)
3. [Android-Specific Issues](#android-specific-issues)
4. [Web-Specific Issues](#web-specific-issues)
5. [Backend Integration Issues](#backend-integration-issues)
6. [Authentication Flow Issues](#authentication-flow-issues)
7. [Network and API Issues](#network-and-api-issues)
8. [Database Issues](#database-issues)
9. [Performance Issues](#performance-issues)
10. [Security and Permissions Issues](#security-and-permissions-issues)

---

## Firebase Initialization Issues

### Error: "Firebase not initialized"

**Symptoms:**
```
[ERROR] Firebase has not been correctly initialized
```

**Causes:**
1. `Firebase.initializeApp()` not called
2. Called after widget tree builds
3. Configuration file missing

**Solutions:**

**Solution 1:** Ensure initialization in `main()`
```dart
void main() async {
  WidgetsFlutterBinding.ensureInitialized();  // REQUIRED

  await Firebase.initializeApp(
    options: DefaultFirebaseOptions.currentPlatform,
  );

  runApp(const MyApp());
}
```

**Solution 2:** Check configuration file exists
```bash
# iOS
ls ios/Runner/GoogleService-Info.plist

# Android
ls android/app/google-services.json

# Flutter generated
ls lib/firebase_options.dart
```

**Solution 3:** Regenerate configuration
```bash
flutterfire configure --project=twigger-prod
```

---

### Error: "No Firebase App '[DEFAULT]' has been created"

**Symptoms:**
```dart
FirebaseException: No Firebase App '[DEFAULT]' has been created
```

**Cause:** Attempting to use Firebase before initialization completes

**Solution:**
```dart
// WRONG
void main() {
  Firebase.initializeApp();  // Not awaited
  runApp(const MyApp());
}

// CORRECT
void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await Firebase.initializeApp(  // Await initialization
    options: DefaultFirebaseOptions.currentPlatform,
  );
  runApp(const MyApp());
}
```

---

### Error: "Duplicate app registration"

**Symptoms:**
```
Firebase app already initialized
```

**Cause:** `Firebase.initializeApp()` called multiple times

**Solution:**
```dart
// Initialize only once in main()
void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Check if already initialized (for hot reload)
  if (Firebase.apps.isEmpty) {
    await Firebase.initializeApp(
      options: DefaultFirebaseOptions.currentPlatform,
    );
  }

  runApp(const MyApp());
}
```

---

## iOS-Specific Issues

### Google Sign-In doesn't work on iOS

**Symptoms:**
- Button taps do nothing
- No Google login screen appears
- Error: "Missing URL scheme"

**Diagnosis:**
```bash
# Check if URL scheme configured
grep -A 10 "CFBundleURLTypes" ios/Runner/Info.plist
```

**Solutions:**

**Solution 1:** Add URL scheme to `Info.plist`

1. Open `ios/Runner/GoogleService-Info.plist`
2. Find `REVERSED_CLIENT_ID` (e.g., `com.googleusercontent.apps.123456`)
3. Add to `ios/Runner/Info.plist`:

```xml
<key>CFBundleURLTypes</key>
<array>
    <dict>
        <key>CFBundleTypeRole</key>
        <string>Editor</string>
        <key>CFBundleURLSchemes</key>
        <array>
            <string>com.googleusercontent.apps.YOUR-CLIENT-ID</string>
        </array>
    </dict>
</array>
```

**Solution 2:** Verify Bundle ID matches Firebase Console

```bash
# Check Bundle ID in Xcode
open ios/Runner.xcworkspace

# Compare with Firebase Console → Project Settings → iOS app
```

Must be identical (e.g., `com.twigger.app`)

**Solution 3:** Rebuild after configuration changes
```bash
cd ios
pod deintegrate
pod install
cd ..
flutter clean
flutter run
```

---

### Error: "The operation couldn't be completed" (iOS)

**Symptoms:**
- Google Sign-In starts, then immediately fails
- Console shows: "The operation couldn't be completed"

**Cause:** Missing Google Sign-In framework or incorrect configuration

**Solution:**
```bash
# Clean and reinstall pods
cd ios
rm -rf Pods Podfile.lock
pod install --repo-update
cd ..

# Rebuild
flutter clean
flutter build ios
```

---

### Facebook Login shows blank screen (iOS)

**Symptoms:**
- Tap "Continue with Facebook" → Blank screen
- No Facebook login UI

**Causes:**
1. Facebook SDK not configured
2. Missing Facebook app ID in `Info.plist`

**Solution:** Add to `ios/Runner/Info.plist`:

```xml
<key>CFBundleURLTypes</key>
<array>
    <dict>
        <key>CFBundleURLSchemes</key>
        <array>
            <string>fbYOUR-FACEBOOK-APP-ID</string>
        </array>
    </dict>
</array>

<key>FacebookAppID</key>
<string>YOUR-FACEBOOK-APP-ID</string>

<key>FacebookDisplayName</key>
<string>Twigger</string>

<key>LSApplicationQueriesSchemes</key>
<array>
    <string>fbapi</string>
    <string>fb-messenger-share-api</string>
</array>
```

---

## Android-Specific Issues

### Error: "Status{statusCode=DEVELOPER_ERROR}"

**Symptoms:**
```
Sign in failed: PlatformException(sign_in_failed, Status{statusCode=DEVELOPER_ERROR}, null, null)
```

**Causes:**
1. Missing SHA-1 fingerprint in Firebase Console
2. Wrong package name
3. Google Services plugin not applied

**Solutions:**

**Solution 1:** Add SHA-1 fingerprint

```bash
# Get debug SHA-1
cd android
./gradlew signingReport

# Copy SHA-1 (under "Variant: debug")
# Example: 1A:2B:3C:4D:5E:6F:...

# Add to Firebase Console:
# Project Settings → Your apps → Android → Add fingerprint
```

**Solution 2:** Verify package name matches

```gradle
// android/app/build.gradle
android {
    defaultConfig {
        applicationId "com.twigger.app"  // MUST match Firebase Console
    }
}
```

**Solution 3:** Ensure Google Services plugin applied

```gradle
// android/app/build.gradle (at the BOTTOM of file)
apply plugin: 'com.google.gms.google-services'
```

**Solution 4:** Clean and rebuild
```bash
cd android
./gradlew clean
cd ..
flutter clean
flutter run
```

---

### Error: "google-services.json is missing"

**Symptoms:**
```
Execution failed for task ':app:processDebugGoogleServices'.
> File google-services.json is missing.
```

**Diagnosis:**
```bash
ls android/app/google-services.json
```

**Solution:**
1. Download from Firebase Console → Project Settings → Android app
2. Place in `android/app/` directory
3. Verify file structure:

```json
{
  "project_info": {
    "project_id": "twigger-prod"
  },
  "client": [
    {
      "client_info": {
        "mobilesdk_app_id": "1:...",
        "android_client_info": {
          "package_name": "com.twigger.app"
        }
      }
    }
  ]
}
```

---

### Facebook Login fails on Android

**Symptoms:**
- "App Not Setup" error
- Facebook login starts then fails

**Causes:**
1. Facebook App ID not in `strings.xml`
2. Package name not registered in Facebook Developer Console

**Solution 1:** Add to `android/app/src/main/res/values/strings.xml`:

```xml
<resources>
    <string name="app_name">Twigger</string>
    <string name="facebook_app_id">YOUR-FACEBOOK-APP-ID</string>
    <string name="fb_login_protocol_scheme">fbYOUR-FACEBOOK-APP-ID</string>
</resources>
```

**Solution 2:** Add to `android/app/src/main/AndroidManifest.xml`:

```xml
<application>
    <!-- Add inside <application> tag -->
    <meta-data
        android:name="com.facebook.sdk.ApplicationId"
        android:value="@string/facebook_app_id"/>

    <activity
        android:name="com.facebook.FacebookActivity"
        android:configChanges="keyboard|keyboardHidden|screenLayout|screenSize|orientation"
        android:label="@string/app_name" />

    <activity
        android:name="com.facebook.CustomTabActivity"
        android:exported="true">
        <intent-filter>
            <action android:name="android.intent.action.VIEW" />
            <category android:name="android.intent.category.DEFAULT" />
            <category android:name="android.intent.category.BROWSABLE" />
            <data android:scheme="@string/fb_login_protocol_scheme" />
        </intent-filter>
    </activity>
</application>
```

**Solution 3:** Add package name and key hash to Facebook Developer Console

```bash
# Get key hash for Facebook
keytool -exportcert -alias androiddebugkey -keystore ~/.android/debug.keystore | openssl sha1 -binary | openssl base64

# Add to Facebook Developers → Your App → Settings → Android
```

---

## Web-Specific Issues

### Error: "Popup blocked by browser"

**Symptoms:**
- Google/Facebook login popup doesn't appear
- Console error: "Popup blocked"

**Solution 1:** Use redirect flow instead of popup

```dart
// Instead of signInWithPopup
await FirebaseAuth.instance.signInWithRedirect(GoogleAuthProvider());

// Handle redirect result
final result = await FirebaseAuth.instance.getRedirectResult();
```

**Solution 2:** Inform user to allow popups

```dart
try {
  await authService.signInWithGoogle();
} catch (e) {
  if (e.toString().contains('popup')) {
    showDialog(
      context: context,
      builder: (_) => AlertDialog(
        title: Text('Allow Popups'),
        content: Text('Please enable popups for this site to sign in.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text('OK'),
          ),
        ],
      ),
    );
  }
}
```

---

### CORS error on API calls (Web)

**Symptoms:**
```
Access to fetch at 'http://localhost:8080/api/v1/auth/verify' from origin 'http://localhost:5000'
has been blocked by CORS policy
```

**Cause:** Backend doesn't allow requests from web app origin

**Solution:** Add CORS middleware to backend

```go
// internal/api-gateway/middleware/cors.go
func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}

// Use in router
router.Use(middleware.CORS())
```

**Production:** Restrict origins to your domain
```go
allowedOrigins := []string{
    "https://app.twigger.com",
    "https://staging.twigger.com",
}
```

---

### Safari: "Cross-site tracking prevention"

**Symptoms:**
- Google Sign-In works in Chrome but not Safari
- Error about third-party cookies

**Solution 1:** Inform users to adjust Safari settings

Safari → Preferences → Privacy → Uncheck "Prevent cross-site tracking"

**Solution 2:** Use SameSite=None for cookies (backend)

```go
// Set cookie with SameSite=None; Secure
http.SetCookie(w, &http.Cookie{
    Name:     "session",
    Value:    sessionID,
    SameSite: http.SameSiteNoneMode,
    Secure:   true,  // HTTPS required
})
```

---

## Backend Integration Issues

### Error: "Authentication failed: 401"

**Symptoms:**
```
POST /api/v1/auth/verify → 401 Unauthorized
```

**Causes:**
1. Invalid Firebase token
2. Token expired
3. Backend can't verify token (Firebase credentials issue)

**Diagnosis:**
```bash
# Get token from Flutter app
print('Token: $token');

# Decode token at https://jwt.io
# Check:
# - "iss" claim = "https://securetoken.google.com/twigger-prod"
# - "aud" claim = "twigger-prod"
# - "exp" not expired
```

**Solution 1:** Check backend logs

```bash
# Look for Firebase initialization errors
grep -i "firebase" api-gateway.log

# Expected:
# "Successfully initialized Firebase"
```

**Solution 2:** Verify backend environment variables

```bash
echo $FIREBASE_PROJECT_ID  # Should be: twigger-prod
echo $FIREBASE_CREDENTIALS_PATH  # Should point to valid JSON file
cat $FIREBASE_CREDENTIALS_PATH  # Should be valid JSON
```

**Solution 3:** Force token refresh (client)

```dart
// Get fresh token
final token = await FirebaseAuth.instance.currentUser?.getIdToken(true);  // true = force refresh
```

---

### Error: "Authentication failed: 429 Too Many Requests"

**Symptoms:**
```
POST /api/v1/auth/verify → 429 Too Many Requests
Retry-After: 60
```

**Cause:** Rate limiting enforced (5 requests/minute per IP)

**Solution 1:** Implement exponential backoff (client)

```dart
Future<void> retryAuthentication() async {
  int attempts = 0;
  const maxAttempts = 3;

  while (attempts < maxAttempts) {
    try {
      await authService.completeAuthentication();
      return;
    } on HttpException catch (e) {
      if (e.message.contains('429')) {
        attempts++;
        final delaySeconds = math.pow(2, attempts).toInt();
        await Future.delayed(Duration(seconds: delaySeconds));
      } else {
        rethrow;
      }
    }
  }

  throw Exception('Max retry attempts reached');
}
```

**Solution 2:** Cache authentication response (don't call repeatedly)

```dart
// Cache user profile
SharedPreferences prefs = await SharedPreferences.getInstance();
prefs.setString('user_profile', jsonEncode(profile));

// Reuse cached data if fresh
final cached = prefs.getString('user_profile');
if (cached != null) {
  final cacheTime = DateTime.parse(prefs.getString('profile_cache_time')!);
  if (DateTime.now().difference(cacheTime).inMinutes < 5) {
    return jsonDecode(cached);  // Use cached profile
  }
}
```

---

### Error: "User profile not syncing"

**Symptoms:**
- User signs in successfully
- Database has user record
- BUT: Photo URL or email not showing in app

**Diagnosis:**
```sql
-- Check user in database
SELECT user_id, email, username, photo_url, provider
FROM users
WHERE email = 'test@example.com';
```

**Cause:** Client not calling `/api/v1/auth/verify` after Firebase sign-in

**Solution:** Ensure `completeAuthentication()` called

```dart
Future<UserCredential?> signInWithGoogle() async {
  // Firebase sign-in
  final userCredential = await _firebaseAuth.signInWithCredential(credential);

  // CRITICAL: Call backend to sync profile
  await _completeAuthentication(userCredential);  // <--- Must call this

  return userCredential;
}
```

---

## Authentication Flow Issues

### User created multiple times (duplicate accounts)

**Symptoms:**
```sql
SELECT COUNT(*) FROM users WHERE email = 'test@example.com';
-- Returns: 2 or more
```

**Cause:** Account linking not working - creating new user instead of linking

**Diagnosis:**
```sql
-- Check if multiple users with same email
SELECT user_id, email, provider, firebase_uid, created_at
FROM users
WHERE email = 'test@example.com'
ORDER BY created_at;
```

**Expected:** Only 1 user, multiple providers in `linked_accounts`

**Solution:** Verify backend account linking logic

```go
// Backend should check email first
existingUser, err := s.userRepo.GetByEmail(ctx, email)
if err == nil && existingUser != nil {
    // Link new provider to existing account
    return s.handleAccountLinking(ctx, existingUser, newFirebaseUID, provider, ...)
}
```

**Database Fix (if duplicates exist):**
```sql
-- Find duplicates
SELECT email, COUNT(*) FROM users GROUP BY email HAVING COUNT(*) > 1;

-- Manually merge (keep oldest user, delete newer)
-- This is complex - contact backend team
```

---

### Photo URL not updating

**Symptoms:**
- User signs in with Google (has photo)
- Later signs in with Facebook (different photo)
- Photo URL remains Google's photo (expected behavior)

**Expected Behavior:** Photo URL only updates if user has NO photo

**Verification:**
```sql
SELECT email, photo_url FROM users WHERE email = 'test@example.com';
-- Should show first provider's photo URL
```

**If you WANT to force update photo:**

```sql
-- Manually update photo URL (not recommended)
UPDATE users
SET photo_url = 'https://graph.facebook.com/...'
WHERE email = 'test@example.com';
```

---

### Account linking not tracking providers

**Symptoms:**
- User signs in with Google, then Facebook
- Only 1 entry in `linked_accounts` table

**Diagnosis:**
```sql
SELECT la.provider, la.linked_at
FROM linked_accounts la
JOIN users u ON la.user_id = u.user_id
WHERE u.email = 'test@example.com'
ORDER BY la.linked_at;
```

**Expected:** 2 rows (Google and Facebook)

**Cause:** Backend `LinkProvider()` not called or failing silently

**Solution:** Check backend logs for errors

```bash
grep -i "linkprovider" api-gateway.log
grep -i "linked_accounts" api-gateway.log
```

**Backend fix:** Ensure `LinkProvider()` called in both flows

```go
// After creating new user
s.userRepo.LinkProvider(ctx, user.UserID, provider, firebaseUID)

// After linking existing account
s.userRepo.LinkProvider(ctx, existingUser.UserID, newProvider, newFirebaseUID)
```

---

## Network and API Issues

### Timeout errors

**Symptoms:**
```
SocketException: Connection timed out
```

**Solutions:**

**Solution 1:** Increase timeout duration

```dart
final response = await http.post(
  Uri.parse('$apiBaseUrl/auth/verify'),
  headers: {...},
  body: {...},
).timeout(
  const Duration(seconds: 15),  // Increase from default 10s
  onTimeout: () => throw TimeoutException('Request timed out'),
);
```

**Solution 2:** Check backend is running

```bash
curl http://localhost:8080/health
```

**Solution 3:** Check network connectivity

```dart
import 'package:connectivity_plus/connectivity_plus.dart';

final connectivity = await Connectivity().checkConnectivity();
if (connectivity == ConnectivityResult.none) {
  throw Exception('No internet connection');
}
```

---

### SSL/TLS errors on Android

**Symptoms:**
```
HandshakeException: Handshake error in client
```

**Cause:** Self-signed certificates or Android network security config

**Solution (development only):**

Create `android/app/src/main/res/xml/network_security_config.xml`:

```xml
<?xml version="1.0" encoding="utf-8"?>
<network-security-config>
    <base-config cleartextTrafficPermitted="true">
        <trust-anchors>
            <certificates src="system" />
        </trust-anchors>
    </base-config>
    <domain-config cleartextTrafficPermitted="true">
        <domain includeSubdomains="true">localhost</domain>
        <domain includeSubdomains="true">10.0.2.2</domain>
    </domain-config>
</network-security-config>
```

Add to `AndroidManifest.xml`:
```xml
<application
    android:networkSecurityConfig="@xml/network_security_config">
```

**PRODUCTION:** Use only HTTPS, no cleartext traffic!

---

## Database Issues

### Error: "relation 'users' does not exist"

**Symptoms:**
Backend logs show:
```
ERROR: relation "users" does not exist
```

**Cause:** Migrations not run

**Solution:**
```bash
# Run migrations
DATABASE_URL="postgres://postgres:postgres@localhost:5432/twigger?sslmode=disable" \
  go run cmd/migrate/main.go up

# Verify tables exist
psql postgres://postgres:postgres@localhost:5432/twigger -c "\dt"
```

Expected tables: `users`, `workspaces`, `workspace_members`, `auth_sessions`, `linked_accounts`, `auth_audit_log`

---

### Error: "duplicate key value violates unique constraint"

**Symptoms:**
```
ERROR: duplicate key value violates unique constraint "users_firebase_uid_key"
```

**Cause:** Attempting to create user with existing `firebase_uid`

**Diagnosis:**
```sql
SELECT * FROM users WHERE firebase_uid = '{firebase-uid}';
```

**Solution:** Backend should check for existing user first

```go
// Backend logic
existingUser, err := s.userRepo.GetByFirebaseUID(ctx, firebaseUID)
if err == nil && existingUser != nil {
    // User exists - update instead of create
    return s.handleExistingUserLogin(ctx, existingUser, ...)
}
```

---

## Performance Issues

### Slow authentication (> 5 seconds)

**Diagnosis:**

**Step 1:** Time each component

```dart
// Client-side
final sw1 = Stopwatch()..start();
final credential = await _firebaseAuth.signInWithCredential(googleCred);
print('Firebase sign-in: ${sw1.elapsedMilliseconds}ms');

final sw2 = Stopwatch()..start();
await _completeAuthentication(credential);
print('Backend call: ${sw2.elapsedMilliseconds}ms');
```

**Step 2:** Check backend API latency

```bash
time curl -X POST http://localhost:8080/api/v1/auth/verify \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{"device_id": "test"}'
```

**Target:** < 300ms

**Solutions:**

**If Firebase is slow:**
- Check internet connection
- Firebase might be experiencing issues (check status page)

**If backend is slow:**
```sql
-- Check database query performance
EXPLAIN ANALYZE
SELECT * FROM users WHERE firebase_uid = '{uid}';

-- Should use index, execution time < 20ms
```

**Add database indexes if missing:**
```sql
CREATE INDEX IF NOT EXISTS idx_users_firebase_uid ON users(firebase_uid);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
```

---

## Security and Permissions Issues

### Error: "Permission denied" (iOS Camera/Photos)

**Symptoms:**
- User tries to update profile photo
- App crashes or shows "Permission denied"

**Solution:** Add to `ios/Runner/Info.plist`:

```xml
<key>NSCameraUsageDescription</key>
<string>We need access to your camera to update your profile photo</string>
<key>NSPhotoLibraryUsageDescription</key>
<string>We need access to your photo library to update your profile photo</string>
```

**Request permission before accessing:**
```dart
import 'package:permission_handler/permission_handler.dart';

Future<void> pickImage() async {
  final status = await Permission.photos.request();

  if (status.isGranted) {
    // Proceed with image picker
  } else {
    // Show error message
  }
}
```

---

### Firebase token exposed in logs

**Symptoms:**
Console shows:
```
Token: eyJhbGciOiJSUzI1NiIsImtpZCI6...
```

**Risk:** Token can be stolen from logs

**Solution:** Remove debug logging in production

```dart
void logDebug(String message) {
  if (kDebugMode) {  // Only log in debug mode
    print('[DEBUG] $message');
  }
}

// NEVER log tokens
// print('Token: $token');  // ❌ DON'T DO THIS
```

**ProGuard (Android):** Remove logs in release builds

```proguard
-assumenosideeffects class android.util.Log {
    public static *** d(...);
    public static *** v(...);
}
```

---

## Diagnostic Commands Reference

### Check Firebase Auth Status (Client)

```dart
final user = FirebaseAuth.instance.currentUser;
if (user != null) {
  print('Signed in: ${user.email}');
  print('Provider: ${user.providerData.first.providerId}');
  print('UID: ${user.uid}');
  print('Email verified: ${user.emailVerified}');

  // Get token and check expiration
  final token = await user.getIdToken();
  print('Token: ${token?.substring(0, 20)}...');
} else {
  print('Not signed in');
}
```

### Check Backend Health

```bash
# Health check
curl http://localhost:8080/health

# Verify Firebase initialized
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer {valid-token}"
# Should return 200 or 401, not 500
```

### Check Database State

```sql
-- Count users
SELECT COUNT(*) FROM users;

-- Recent users
SELECT email, provider, created_at FROM users ORDER BY created_at DESC LIMIT 10;

-- Check account linking
SELECT u.email, la.provider, la.linked_at
FROM users u
JOIN linked_accounts la ON u.user_id = la.user_id
ORDER BY u.created_at DESC, la.linked_at;

-- Check audit logs
SELECT user_id, event_type, success, created_at
FROM auth_audit_log
ORDER BY created_at DESC
LIMIT 20;
```

---

## Getting Help

### Before Asking for Help

1. **Check logs:**
   - Flutter: `flutter logs`
   - iOS: Xcode Console
   - Android: `adb logcat`
   - Backend: `tail -f api-gateway.log`

2. **Try clean rebuild:**
```bash
flutter clean
cd ios && pod deintegrate && pod install && cd ..
cd android && ./gradlew clean && cd ..
flutter run
```

3. **Search this troubleshooting guide**

### Information to Provide

When reporting an issue, include:

- **Platform:** iOS/Android/Web
- **Flutter version:** `flutter --version`
- **Error message:** Full stack trace
- **Steps to reproduce:** Detailed steps
- **Backend logs:** Relevant error lines
- **Database state:** Result of diagnostic queries
- **Firebase Console:** Any errors in Authentication tab

---

## Common Questions

**Q: Why does Google Sign-In work on Android but not iOS?**

**A:** Different platform configurations required. Check:
- iOS: URL scheme in `Info.plist`
- Android: SHA-1 fingerprint in Firebase Console

---

**Q: Can I test without a real Google/Facebook account?**

**A:** Yes, create test accounts in Firebase Console:
- Firebase Console → Authentication → Users → Add user
- Email: `test@example.com`, password: `testpass123`
- Note: These are only for email/password, not social sign-in

---

**Q: How do I reset everything and start fresh?**

**A:**

```bash
# Flutter
flutter clean
rm -rf ios/Pods ios/Podfile.lock
rm -rf android/.gradle android/build

# Database
psql postgres://postgres:postgres@localhost:5432/twigger
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
CREATE EXTENSION IF NOT EXISTS postgis;
\q

# Run migrations
go run cmd/migrate/main.go up

# Rebuild app
flutter pub get
cd ios && pod install && cd ..
flutter run
```

---

**Last Updated:** 2025-10-08
**Related Docs:** [client-integration-guide.md](./client-integration-guide.md), [frontend-testing-guide.md](./frontend-testing-guide.md)
