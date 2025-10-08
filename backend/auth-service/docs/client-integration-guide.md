# Client Integration Guide - Firebase Authentication

**Version:** 1.0
**Date:** 2025-10-08
**Target Platforms:** iOS, Android, Web (Flutter)

---

## Overview

This guide provides step-by-step instructions for integrating Firebase Authentication into Twigger's Flutter-based clients (iOS, Android, Web). After completing this guide, your client app will:

- Support Google Sign-In and Facebook Login
- Automatically create user accounts and workspaces
- Handle account linking across multiple providers
- Sync user profiles with the backend
- Manage session tokens and automatic refresh

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Flutter Project Setup](#flutter-project-setup)
3. [iOS Configuration](#ios-configuration)
4. [Android Configuration](#android-configuration)
5. [Web Configuration](#web-configuration)
6. [Authentication Implementation](#authentication-implementation)
7. [Backend Integration](#backend-integration)
8. [Error Handling](#error-handling)
9. [Testing](#testing)
10. [Best Practices](#best-practices)

---

## Prerequisites

### Required Tools
- Flutter SDK 3.16+ (`flutter --version`)
- Xcode 15+ (for iOS development)
- Android Studio (for Android development)
- Firebase CLI (`npm install -g firebase-tools`)
- Git

### Firebase Project Access
- Firebase project: `twigger-prod`
- Access to Firebase Console: https://console.firebase.google.com
- Permissions to download configuration files

### Backend Information
- **API Gateway URL** (local): `http://localhost:8080`
- **API Gateway URL** (production): `https://api.twigger.com` (or your domain)
- **Auth Endpoints**:
  - `POST /api/v1/auth/verify` - Complete authentication
  - `GET /api/v1/auth/me` - Get user profile
  - `POST /api/v1/auth/logout` - Logout

---

## Flutter Project Setup

### 1. Add Dependencies

Add these packages to your `pubspec.yaml`:

```yaml
dependencies:
  flutter:
    sdk: flutter

  # Firebase Core (required)
  firebase_core: ^2.24.2
  firebase_auth: ^4.16.0

  # Social Sign-In Providers
  google_sign_in: ^6.2.1
  flutter_facebook_auth: ^6.0.4

  # HTTP Client
  http: ^1.2.0

  # State Management (choose one)
  provider: ^6.1.1  # or riverpod, bloc, etc.

  # Secure Storage
  flutter_secure_storage: ^9.0.0

dev_dependencies:
  flutter_test:
    sdk: flutter
```

### 2. Install Packages

```bash
flutter pub get
```

### 3. Initialize FlutterFire CLI

```bash
# Install FlutterFire CLI
dart pub global activate flutterfire_cli

# Configure Firebase for all platforms
flutterfire configure --project=twigger-prod
```

This will:
- Generate `lib/firebase_options.dart`
- Download platform-specific configuration files
- Configure iOS and Android apps in Firebase Console

---

## iOS Configuration

### 1. Download Configuration File

1. Go to [Firebase Console](https://console.firebase.google.com) → `twigger-prod`
2. Click iOS app (or add one if not exists)
3. Download `GoogleService-Info.plist`
4. Open `ios/Runner.xcworkspace` in Xcode
5. Drag `GoogleService-Info.plist` into the `Runner` folder in Xcode
6. Ensure "Copy items if needed" is checked

### 2. Configure Bundle Identifier

In Xcode:
1. Select `Runner` target
2. General tab → Identity
3. Set Bundle Identifier: `com.twigger.app` (or your bundle ID)
4. Ensure it matches Firebase Console

### 3. Add URL Schemes (for Google Sign-In)

1. In Xcode, open `ios/Runner/Info.plist`
2. Add URL scheme from `GoogleService-Info.plist`:

```xml
<key>CFBundleURLTypes</key>
<array>
    <dict>
        <key>CFBundleTypeRole</key>
        <string>Editor</string>
        <key>CFBundleURLSchemes</key>
        <array>
            <!-- Get this from GoogleService-Info.plist → REVERSED_CLIENT_ID -->
            <string>com.googleusercontent.apps.YOUR-CLIENT-ID</string>
        </array>
    </dict>
</array>
```

### 4. Update iOS Deployment Target

Ensure `ios/Podfile` has minimum iOS 12:

```ruby
platform :ios, '12.0'
```

### 5. Install CocoaPods

```bash
cd ios
pod install
cd ..
```

### 6. Configure Permissions (Info.plist)

Add camera/photo library access for profile photos:

```xml
<key>NSCameraUsageDescription</key>
<string>We need access to your camera to update your profile photo</string>
<key>NSPhotoLibraryUsageDescription</key>
<string>We need access to your photo library to update your profile photo</string>
```

---

## Android Configuration

### 1. Download Configuration File

1. Go to Firebase Console → `twigger-prod` → Android app
2. Download `google-services.json`
3. Place in `android/app/` directory

### 2. Configure Package Name

In `android/app/build.gradle`:

```gradle
android {
    defaultConfig {
        applicationId "com.twigger.app"  // Must match Firebase Console
        minSdkVersion 21  // Minimum for Firebase Auth
        targetSdkVersion 34
    }
}
```

### 3. Add Google Services Plugin

In `android/build.gradle` (project-level):

```gradle
buildscript {
    dependencies {
        classpath 'com.google.gms:google-services:4.4.0'
    }
}
```

In `android/app/build.gradle` (app-level), add at the **bottom**:

```gradle
apply plugin: 'com.google.gms.google-services'
```

### 4. Add SHA-1 Fingerprint to Firebase

**Debug Mode** (for local testing):

```bash
# Get debug SHA-1
cd android
./gradlew signingReport

# Copy SHA-1 from output (under "debug" variant)
```

**Release Mode** (for production):

```bash
# Generate keystore if you don't have one
keytool -genkey -v -keystore ~/twigger-release-key.jks -keyalg RSA -keysize 2048 -validity 10000 -alias twigger

# Get release SHA-1
keytool -list -v -keystore ~/twigger-release-key.jks -alias twigger
```

Add both SHA-1 fingerprints to Firebase Console:
1. Firebase Console → Project Settings → Your apps → Android app
2. Click "Add fingerprint"
3. Paste SHA-1 and save

### 5. Configure ProGuard (for release builds)

In `android/app/proguard-rules.pro`:

```proguard
-keep class com.google.firebase.** { *; }
-keep class com.google.android.gms.** { *; }
```

---

## Web Configuration

### 1. Initialize Firebase in Web

Edit `web/index.html`, add before `</body>`:

```html
<script src="https://www.gstatic.com/firebasejs/10.7.1/firebase-app-compat.js"></script>
<script src="https://www.gstatic.com/firebasejs/10.7.1/firebase-auth-compat.js"></script>
```

### 2. Configure Firebase

The `flutterfire configure` command already generated `lib/firebase_options.dart`. No additional web-specific config needed.

### 3. Update Web Index

Ensure `web/index.html` has proper meta tags:

```html
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
```

---

## Authentication Implementation

### 1. Initialize Firebase

In `lib/main.dart`:

```dart
import 'package:firebase_core/firebase_core.dart';
import 'firebase_options.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  await Firebase.initializeApp(
    options: DefaultFirebaseOptions.currentPlatform,
  );

  runApp(const MyApp());
}
```

### 2. Create Auth Service

Create `lib/services/auth_service.dart`:

```dart
import 'package:firebase_auth/firebase_auth.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:flutter_facebook_auth/flutter_facebook_auth.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';

class AuthService {
  final FirebaseAuth _firebaseAuth = FirebaseAuth.instance;
  final GoogleSignIn _googleSignIn = GoogleSignIn(
    scopes: ['email', 'profile'],
  );

  // Backend API base URL
  static const String _apiBaseUrl = 'http://localhost:8080/api/v1';

  // Get current user
  User? get currentUser => _firebaseAuth.currentUser;

  // Auth state stream
  Stream<User?> get authStateChanges => _firebaseAuth.authStateChanges();

  /// Sign in with Google
  Future<UserCredential?> signInWithGoogle() async {
    try {
      // Trigger Google Sign-In flow
      final GoogleSignInAccount? googleUser = await _googleSignIn.signIn();

      if (googleUser == null) {
        // User canceled the sign-in
        return null;
      }

      // Obtain auth details
      final GoogleSignInAuthentication googleAuth = await googleUser.authentication;

      // Create Firebase credential
      final credential = GoogleAuthProvider.credential(
        accessToken: googleAuth.accessToken,
        idToken: googleAuth.idToken,
      );

      // Sign in to Firebase
      final userCredential = await _firebaseAuth.signInWithCredential(credential);

      // Complete authentication with backend
      await _completeAuthentication(userCredential);

      return userCredential;
    } catch (e) {
      print('Error signing in with Google: $e');
      rethrow;
    }
  }

  /// Sign in with Facebook
  Future<UserCredential?> signInWithFacebook() async {
    try {
      // Trigger Facebook Login flow
      final LoginResult result = await FacebookAuth.instance.login(
        permissions: ['email', 'public_profile'],
      );

      if (result.status != LoginStatus.success) {
        // User canceled or error occurred
        return null;
      }

      // Create Firebase credential
      final credential = FacebookAuthProvider.credential(
        result.accessToken!.token,
      );

      // Sign in to Firebase
      final userCredential = await _firebaseAuth.signInWithCredential(credential);

      // Complete authentication with backend
      await _completeAuthentication(userCredential);

      return userCredential;
    } catch (e) {
      print('Error signing in with Facebook: $e');
      rethrow;
    }
  }

  /// Complete authentication with backend
  Future<Map<String, dynamic>> _completeAuthentication(
    UserCredential userCredential,
  ) async {
    try {
      // Get Firebase ID token
      final idToken = await userCredential.user?.getIdToken();

      if (idToken == null) {
        throw Exception('Failed to get ID token');
      }

      // Call backend /auth/verify endpoint
      final response = await http.post(
        Uri.parse('$_apiBaseUrl/auth/verify'),
        headers: {
          'Authorization': 'Bearer $idToken',
          'Content-Type': 'application/json',
        },
        body: jsonEncode({
          'device_id': await _getDeviceId(),
        }),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        print('Authentication completed: ${data['user']['email']}');
        return data;
      } else if (response.statusCode == 429) {
        throw Exception('Too many requests. Please try again later.');
      } else {
        throw Exception('Authentication failed: ${response.statusCode}');
      }
    } catch (e) {
      print('Error completing authentication: $e');
      rethrow;
    }
  }

  /// Get user profile from backend
  Future<Map<String, dynamic>?> getUserProfile() async {
    try {
      final idToken = await currentUser?.getIdToken();

      if (idToken == null) {
        return null;
      }

      final response = await http.get(
        Uri.parse('$_apiBaseUrl/auth/me'),
        headers: {
          'Authorization': 'Bearer $idToken',
        },
      );

      if (response.statusCode == 200) {
        return jsonDecode(response.body);
      } else {
        return null;
      }
    } catch (e) {
      print('Error getting user profile: $e');
      return null;
    }
  }

  /// Sign out
  Future<void> signOut() async {
    try {
      // Get token before signing out (for backend call)
      final idToken = await currentUser?.getIdToken();

      // Call backend logout
      if (idToken != null) {
        await http.post(
          Uri.parse('$_apiBaseUrl/auth/logout'),
          headers: {
            'Authorization': 'Bearer $idToken',
            'Content-Type': 'application/json',
          },
          body: jsonEncode({
            'revoke_all': false,
          }),
        );
      }

      // Sign out from Firebase
      await _firebaseAuth.signOut();
      await _googleSignIn.signOut();
      await FacebookAuth.instance.logOut();
    } catch (e) {
      print('Error signing out: $e');
      rethrow;
    }
  }

  /// Get device ID (implement based on your needs)
  Future<String> _getDeviceId() async {
    // Use flutter_secure_storage or device_info_plus to get a persistent device ID
    // For now, return a placeholder
    return 'flutter-client-${DateTime.now().millisecondsSinceEpoch}';
  }
}
```

### 3. Create Login Screen

Create `lib/screens/login_screen.dart`:

```dart
import 'package:flutter/material.dart';
import '../services/auth_service.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({Key? key}) : super(key: key);

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final AuthService _authService = AuthService();
  bool _isLoading = false;

  Future<void> _signInWithGoogle() async {
    setState(() => _isLoading = true);

    try {
      final userCredential = await _authService.signInWithGoogle();

      if (userCredential != null) {
        // Navigate to home screen
        if (mounted) {
          Navigator.of(context).pushReplacementNamed('/home');
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Sign in failed: $e')),
        );
      }
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  Future<void> _signInWithFacebook() async {
    setState(() => _isLoading = true);

    try {
      final userCredential = await _authService.signInWithFacebook();

      if (userCredential != null) {
        // Navigate to home screen
        if (mounted) {
          Navigator.of(context).pushReplacementNamed('/home');
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Sign in failed: $e')),
        );
      }
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: Padding(
            padding: const EdgeInsets.all(24.0),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                // App Logo
                Icon(
                  Icons.local_florist,
                  size: 100,
                  color: Colors.green,
                ),
                const SizedBox(height: 24),

                // Title
                const Text(
                  'Welcome to Twigger',
                  style: TextStyle(
                    fontSize: 28,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const SizedBox(height: 8),

                // Subtitle
                const Text(
                  'Plan your perfect garden',
                  style: TextStyle(
                    fontSize: 16,
                    color: Colors.grey,
                  ),
                ),
                const SizedBox(height: 48),

                // Google Sign-In Button
                ElevatedButton.icon(
                  onPressed: _isLoading ? null : _signInWithGoogle,
                  icon: Image.asset(
                    'assets/google_logo.png',  // Add Google logo asset
                    height: 24,
                  ),
                  label: const Text('Continue with Google'),
                  style: ElevatedButton.styleFrom(
                    minimumSize: const Size(double.infinity, 50),
                    backgroundColor: Colors.white,
                    foregroundColor: Colors.black87,
                  ),
                ),
                const SizedBox(height: 16),

                // Facebook Sign-In Button
                ElevatedButton.icon(
                  onPressed: _isLoading ? null : _signInWithFacebook,
                  icon: const Icon(Icons.facebook),
                  label: const Text('Continue with Facebook'),
                  style: ElevatedButton.styleFrom(
                    minimumSize: const Size(double.infinity, 50),
                    backgroundColor: const Color(0xFF1877F2),
                    foregroundColor: Colors.white,
                  ),
                ),

                // Loading Indicator
                if (_isLoading)
                  const Padding(
                    padding: EdgeInsets.only(top: 24),
                    child: CircularProgressIndicator(),
                  ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
```

### 4. Add Auth State Listener

In `lib/main.dart`:

```dart
import 'package:flutter/material.dart';
import 'package:firebase_core/firebase_core.dart';
import 'firebase_options.dart';
import 'services/auth_service.dart';
import 'screens/login_screen.dart';
import 'screens/home_screen.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await Firebase.initializeApp(
    options: DefaultFirebaseOptions.currentPlatform,
  );
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Twigger',
      theme: ThemeData(
        primarySwatch: Colors.green,
        useMaterial3: true,
      ),
      home: const AuthWrapper(),
      routes: {
        '/login': (context) => const LoginScreen(),
        '/home': (context) => const HomeScreen(),
      },
    );
  }
}

class AuthWrapper extends StatelessWidget {
  const AuthWrapper({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    final authService = AuthService();

    return StreamBuilder(
      stream: authService.authStateChanges,
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.waiting) {
          return const Scaffold(
            body: Center(child: CircularProgressIndicator()),
          );
        }

        if (snapshot.hasData) {
          return const HomeScreen();
        }

        return const LoginScreen();
      },
    );
  }
}
```

---

## Backend Integration

### Authentication Flow

```
1. User taps "Sign in with Google" → Flutter
2. Google Sign-In SDK flow → Firebase
3. Firebase returns JWT token → Flutter
4. Flutter calls POST /api/v1/auth/verify with JWT → Backend
5. Backend verifies token, creates/updates user → Database
6. Backend returns user profile + workspace → Flutter
7. Flutter stores user data and navigates to home screen
```

### API Endpoints

#### 1. Complete Authentication

**Endpoint:** `POST /api/v1/auth/verify`

**Headers:**
```
Authorization: Bearer {firebase-jwt-token}
Content-Type: application/json
```

**Request Body:**
```json
{
  "device_id": "flutter-ios-12345"
}
```

**Response (200 OK):**
```json
{
  "user": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "username": "user123",
    "photo_url": "https://lh3.googleusercontent.com/...",
    "provider": "google.com",
    "email_verified": true,
    "created_at": "2025-10-08T10:00:00Z"
  },
  "workspaces": [
    {
      "workspace_id": "660e8400-e29b-41d4-a716-446655440001",
      "name": "user123's Garden",
      "role": "owner",
      "created_at": "2025-10-08T10:00:00Z"
    }
  ],
  "session_id": "770e8400-e29b-41d4-a716-446655440002",
  "is_new_user": true
}
```

**Response (429 Too Many Requests):**
```json
{
  "error": "Too many requests. Please try again later.",
  "code": 429
}
```

#### 2. Get User Profile

**Endpoint:** `GET /api/v1/auth/me`

**Headers:**
```
Authorization: Bearer {firebase-jwt-token}
```

**Response (200 OK):**
```json
{
  "user": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "username": "user123",
    "photo_url": "https://lh3.googleusercontent.com/...",
    "provider": "google.com",
    "email_verified": true,
    "last_login_at": "2025-10-08T11:30:00Z",
    "created_at": "2025-10-08T10:00:00Z"
  },
  "workspaces": [...]
}
```

#### 3. Logout

**Endpoint:** `POST /api/v1/auth/logout`

**Headers:**
```
Authorization: Bearer {firebase-jwt-token}
Content-Type: application/json
```

**Request Body:**
```json
{
  "device_id": "flutter-ios-12345",
  "revoke_all": false
}
```

**Response (200 OK):**
```json
{
  "message": "Logged out successfully",
  "sessions_revoked": 1
}
```

### Token Refresh

Firebase SDK automatically refreshes tokens. No client-side action needed.

To manually refresh:

```dart
final token = await FirebaseAuth.instance.currentUser?.getIdToken(true);
```

### HTTP Interceptor (Automatic Token Refresh)

Create `lib/services/http_client.dart`:

```dart
import 'package:http/http.dart' as http;
import 'package:firebase_auth/firebase_auth.dart';
import 'dart:convert';

class AuthenticatedHttpClient {
  static const String baseUrl = 'http://localhost:8080/api/v1';

  Future<http.Response> get(String path) async {
    final token = await _getToken();
    return http.get(
      Uri.parse('$baseUrl$path'),
      headers: {'Authorization': 'Bearer $token'},
    );
  }

  Future<http.Response> post(String path, Map<String, dynamic> body) async {
    final token = await _getToken();
    return http.post(
      Uri.parse('$baseUrl$path'),
      headers: {
        'Authorization': 'Bearer $token',
        'Content-Type': 'application/json',
      },
      body: jsonEncode(body),
    );
  }

  Future<String> _getToken() async {
    final user = FirebaseAuth.instance.currentUser;
    if (user == null) {
      throw Exception('User not authenticated');
    }

    // Force refresh if token is about to expire
    final token = await user.getIdToken(true);
    return token ?? '';
  }
}
```

---

## Error Handling

### Common Errors

#### 1. Firebase Errors

```dart
try {
  await authService.signInWithGoogle();
} on FirebaseAuthException catch (e) {
  switch (e.code) {
    case 'user-disabled':
      // User account has been disabled
      showError('Your account has been disabled');
      break;
    case 'user-not-found':
      // User doesn't exist
      showError('Account not found');
      break;
    case 'network-request-failed':
      // Network error
      showError('Network error. Please check your connection');
      break;
    default:
      showError('Authentication failed: ${e.message}');
  }
}
```

#### 2. Backend Errors

```dart
final response = await http.post(...);

switch (response.statusCode) {
  case 200:
    // Success
    return jsonDecode(response.body);
  case 401:
    // Unauthorized - token invalid
    await authService.signOut();
    showError('Session expired. Please sign in again');
    break;
  case 429:
    // Rate limited
    showError('Too many requests. Please wait a moment');
    break;
  case 500:
    // Server error
    showError('Server error. Please try again later');
    break;
  default:
    showError('Unexpected error occurred');
}
```

### Retry Logic

```dart
Future<T> retryOperation<T>(
  Future<T> Function() operation, {
  int maxAttempts = 3,
}) async {
  int attempts = 0;

  while (attempts < maxAttempts) {
    try {
      return await operation();
    } catch (e) {
      attempts++;
      if (attempts >= maxAttempts) rethrow;

      // Exponential backoff
      await Future.delayed(Duration(seconds: attempts * 2));
    }
  }

  throw Exception('Max retry attempts reached');
}

// Usage
final profile = await retryOperation(() => authService.getUserProfile());
```

---

## Testing

### Unit Tests

Create `test/auth_service_test.dart`:

```dart
import 'package:flutter_test/flutter_test.dart';
import 'package:mockito/mockito.dart';
import 'package:firebase_auth/firebase_auth.dart';

void main() {
  group('AuthService', () {
    test('signInWithGoogle returns UserCredential on success', () async {
      // Add test implementation
    });

    test('signOut clears all sessions', () async {
      // Add test implementation
    });
  });
}
```

### Integration Tests

Create `integration_test/auth_flow_test.dart`:

```dart
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:your_app/main.dart' as app;

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('complete sign-in flow', (WidgetTester tester) async {
    app.main();
    await tester.pumpAndSettle();

    // Find and tap Google Sign-In button
    final googleButton = find.text('Continue with Google');
    expect(googleButton, findsOneWidget);
    await tester.tap(googleButton);
    await tester.pumpAndSettle();

    // Verify navigation to home screen
    // Add assertions
  });
}
```

### Manual Testing Checklist

See [frontend-testing-guide.md](./frontend-testing-guide.md) for comprehensive manual testing procedures.

---

## Best Practices

### 1. Secure Token Storage

**DON'T** store Firebase tokens in SharedPreferences (insecure).

**DO** let Firebase SDK handle token storage automatically:

```dart
// Firebase SDK handles this - no manual storage needed
final user = FirebaseAuth.instance.currentUser;
final token = await user?.getIdToken();
```

### 2. Handle Token Expiration

```dart
// Tokens expire after 1 hour - SDK auto-refreshes
// Force refresh if needed:
final token = await user?.getIdToken(true);  // true = force refresh
```

### 3. Network Error Handling

```dart
Future<T?> safeApiCall<T>(Future<T> Function() call) async {
  try {
    return await call().timeout(
      const Duration(seconds: 10),
      onTimeout: () => throw TimeoutException('Request timed out'),
    );
  } on SocketException {
    showError('No internet connection');
    return null;
  } on TimeoutException {
    showError('Request timed out');
    return null;
  } catch (e) {
    showError('Unexpected error: $e');
    return null;
  }
}
```

### 4. Loading States

```dart
class _MyWidgetState extends State<MyWidget> {
  bool _isLoading = false;

  Future<void> _handleAction() async {
    setState(() => _isLoading = true);

    try {
      // Your async operation
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }
}
```

### 5. Environment Configuration

Use different API URLs per environment:

```dart
class Config {
  static const String environment = String.fromEnvironment(
    'ENVIRONMENT',
    defaultValue: 'development',
  );

  static String get apiBaseUrl {
    switch (environment) {
      case 'production':
        return 'https://api.twigger.com/api/v1';
      case 'staging':
        return 'https://staging-api.twigger.com/api/v1';
      default:
        return 'http://localhost:8080/api/v1';
    }
  }
}

// Run with: flutter run --dart-define=ENVIRONMENT=production
```

### 6. Logging (Production-Safe)

```dart
import 'package:flutter/foundation.dart';

void logDebug(String message) {
  if (kDebugMode) {
    print('[DEBUG] $message');
  }
}

void logError(String message, [Object? error, StackTrace? stackTrace]) {
  if (kDebugMode) {
    print('[ERROR] $message');
    if (error != null) print('Error: $error');
    if (stackTrace != null) print('Stack trace: $stackTrace');
  }

  // In production, send to error tracking service
  // Sentry.captureException(error, stackTrace: stackTrace);
}
```

---

## Next Steps

1. ✅ Complete Firebase setup
2. ✅ Integrate authentication in Flutter app
3. ⏳ Test on all platforms (iOS, Android, Web)
4. ⏳ Add error tracking (Sentry, Firebase Crashlytics)
5. ⏳ Implement profile editing
6. ⏳ Add workspace switching

---

## Troubleshooting

### Firebase not initialized
- Check `Firebase.initializeApp()` called in `main()`
- Verify `GoogleService-Info.plist` (iOS) or `google-services.json` (Android) exists

### Google Sign-In doesn't work
- iOS: Check URL scheme in `Info.plist`
- Android: Add SHA-1 fingerprint to Firebase Console
- Web: Check OAuth redirect URI configuration

### Backend returns 401 Unauthorized
- Token might be expired - force refresh
- Check `FIREBASE_PROJECT_ID` matches in backend
- Verify service account key is correct

### Rate limiting (429 error)
- Backend limits: 5 req/min for `/auth/verify`
- Implement exponential backoff
- Cache user profile locally

For more troubleshooting, see [troubleshooting.md](./troubleshooting.md).

---

**Last Updated:** 2025-10-08
**Maintainer:** Twigger Backend Team
**Related Docs:** [FIREBASE_SETUP.md](./FIREBASE_SETUP.md), [architecture.md](./architecture.md)
