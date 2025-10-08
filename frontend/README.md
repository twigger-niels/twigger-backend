# Twigger Frontend

Flutter mobile app for Twigger - Your AI-Powered Garden Planner.

## 🎯 What's Built

This is a complete authentication system implementation including:

### Authentication Features
- **Email/Password Authentication**
  - Registration with username, email, and password
  - Password strength indicator
  - Email verification flow with auto-checking
  - Password reset via email
  - Form validation

- **Social Authentication**
  - Google Sign-In
  - Facebook Login
  - Account linking handled by backend

- **User Profile**
  - User information display
  - Workspace management
  - Provider badges (Email/Google/Facebook)
  - Sign out with confirmation

### App Structure
- **Bottom Navigation** with 5 tabs (Home, Calendar, Camera, Add, Profile)
- **Placeholder Screens** for non-auth features
- **Material Design 3** with custom garden/nature theme
- **Clean Architecture** with feature-based folder structure
- **Provider State Management** for authentication flows

## 📁 Project Structure

```
lib/
├── core/
│   ├── config/
│   │   └── app_config.dart           # API URLs, timeouts, validation rules
│   ├── constants/
│   │   └── app_constants.dart        # Enums, constants
│   └── theme/
│       └── app_theme.dart            # Material 3 theme with green colors
│
├── features/
│   ├── auth/
│   │   ├── data/
│   │   │   └── auth_api_client.dart  # REST API client for backend
│   │   ├── domain/
│   │   │   ├── user_model.dart       # User data model
│   │   │   └── workspace_model.dart  # Workspace data model
│   │   ├── services/
│   │   │   └── auth_service.dart     # Firebase auth + backend integration
│   │   └── presentation/
│   │       ├── providers/
│   │       │   └── auth_provider.dart        # State management
│   │       └── screens/
│   │           ├── login_screen.dart         # Login with 3 methods
│   │           ├── register_screen.dart      # Registration with validation
│   │           ├── email_verification_screen.dart  # Email verification flow
│   │           ├── forgot_password_screen.dart     # Password reset
│   │           └── splash_screen.dart        # Initial loading screen
│   │
│   ├── profile/
│   │   └── presentation/screens/
│   │       └── profile_screen.dart   # User profile and logout
│   │
│   └── [home, calendar, camera, add]/
│       └── presentation/screens/
│           └── *_screen.dart         # Placeholder screens
│
├── shared/
│   ├── utils/
│   │   └── validators.dart           # Form validators, password strength
│   └── widgets/
│       ├── auth_wrapper.dart         # Routes based on auth state
│       ├── custom_text_field.dart    # Reusable input components
│       ├── loading_indicator.dart    # Loading UI components
│       └── main_navigation_shell.dart # Bottom navigation shell
│
└── main.dart                          # App entry point with Firebase init
```

## 🚀 Getting Started

### Prerequisites

- Flutter SDK 3.9+ ([install guide](https://flutter.dev/docs/get-started/install))
- Dart 3.0+
- Firebase project ([console](https://console.firebase.google.com/))
- Node.js (for Firebase CLI)

### 1. Install Dependencies

```bash
cd frontend
flutter pub get
```

### 2. Configure Firebase

⚠️ **REQUIRED**: Firebase must be configured for the app to run.

```bash
# Install Firebase CLI (if not already installed)
npm install -g firebase-tools

# Login to Firebase
firebase login

# Configure Firebase for all platforms (iOS, Android, Web)
flutterfire configure
```

Follow the prompts to:
1. Select your Firebase project (`twigger-prod`)
2. Choose platforms (iOS, Android, Web)
3. Let it create `firebase_options.dart` automatically

### 3. Enable Authentication Methods in Firebase Console

Go to [Firebase Console](https://console.firebase.google.com/) → Authentication → Sign-in method:

1. **Email/Password**: Enable
2. **Google**: Enable and configure OAuth consent screen
3. **Facebook**: Enable and add Facebook App ID and secret

### 4. Configure Backend API URL

Update `lib/core/config/app_config.dart` with your backend URL:

```dart
static const String apiBaseUrl = String.fromEnvironment(
  'API_BASE_URL',
  defaultValue: 'http://localhost:8080/api/v1', // Change for production
);
```

Or set environment variable when running:

```bash
flutter run --dart-define=API_BASE_URL=https://your-backend-url.com/api/v1
```

### 5. Run the App

```bash
# Run on connected device/emulator
flutter run

# Run with specific device
flutter devices  # List available devices
flutter run -d <device-id>

# Run with backend URL
flutter run --dart-define=API_BASE_URL=http://10.0.2.2:8080/api/v1  # Android emulator
flutter run --dart-define=API_BASE_URL=http://localhost:8080/api/v1  # iOS simulator
```

## 🔧 Development

### Hot Reload

After making changes, save the file to hot reload:
- VS Code: `Ctrl+S` or `Cmd+S`
- Android Studio: `Ctrl+S` or `Cmd+S`
- Terminal: Press `r`

### Debug Mode

```bash
flutter run --debug
```

### Release Build

```bash
# Android APK
flutter build apk --release

# iOS (requires Mac + Xcode)
flutter build ios --release
```

## 🧪 Testing

### Run Tests

```bash
# All tests
flutter test

# Specific file
flutter test test/auth/auth_service_test.dart

# With coverage
flutter test --coverage
```

### Manual Testing Checklist

- [ ] Register with email/password
- [ ] Verify email via link
- [ ] Login with email/password
- [ ] Login with Google
- [ ] Login with Facebook
- [ ] Forgot password flow
- [ ] Profile screen displays user info
- [ ] Sign out
- [ ] App state persists across restarts

## 📱 Platform-Specific Setup

### Android

1. **SHA-1 Certificate** (for Google Sign-In):
```bash
cd android
./gradlew signingReport
```
Add SHA-1 to Firebase Console → Project Settings → Android app

2. **Minimum SDK**: API 21 (Android 5.0) - already configured in `android/app/build.gradle`

### iOS

1. **Xcode**: Install from App Store
2. **CocoaPods**:
```bash
cd ios
pod install
```
3. **Bundle ID**: Update in Xcode to match Firebase project

### Web

1. **Enable Web in Firebase**: Add web app in Firebase Console
2. **CORS**: Configure backend to allow web origin
3. **API URL**: Use full URL (not localhost) for production

## 🔐 Security Notes

### Backend Integration

The app expects these REST endpoints:

- `POST /api/v1/auth/verify` - Verify Firebase token, get/create user
- `POST /api/v1/auth/register` - Register new user with username (optional fallback)
- `GET /api/v1/auth/me` - Get current user profile
- `POST /api/v1/auth/logout` - Logout user

### Token Management

- Firebase ID tokens are automatically refreshed by Firebase SDK
- Tokens are sent in `Authorization: Bearer <token>` header
- Secure storage via `flutter_secure_storage` for sensitive data

### Environment Variables

Never commit:
- `google-services.json` (Android)
- `GoogleService-Info.plist` (iOS)
- `firebase_options.dart` (generated)
- API keys or secrets

## 🎨 Customization

### Theme Colors

Edit `lib/core/theme/app_theme.dart`:

```dart
static const Color primaryGreen = Color(0xFF2E7D32);
static const Color secondaryGreen = Color(0xFF66BB6A);
```

### App Name

Update:
- `pubspec.yaml`: `name: twigger`
- `android/app/src/main/AndroidManifest.xml`: `android:label`
- `ios/Runner/Info.plist`: `CFBundleName`

### App Icon

Use `flutter_launcher_icons` package:

```bash
flutter pub run flutter_launcher_icons
```

## 🐛 Troubleshooting

### Firebase Initialization Failed

**Error**: `Firebase initialization failed: No Firebase App '[DEFAULT]' has been created`

**Solution**: Run `flutterfire configure` to generate `firebase_options.dart`

### Google Sign-In Not Working (Android)

**Error**: `PlatformException(sign_in_failed)`

**Solution**:
1. Add SHA-1 certificate to Firebase Console
2. Re-download `google-services.json`
3. Rebuild app

### Backend Connection Failed

**Error**: `SocketException: Failed host lookup`

**Solution**:
- Android Emulator: Use `http://10.0.2.2:8080` (not `localhost`)
- iOS Simulator: Use `http://localhost:8080`
- Real Device: Use computer's IP address

### Email Verification Not Working

**Issue**: User clicks link but app doesn't detect verification

**Solution**:
1. Check Firebase Console → Authentication → Templates → Email verification
2. Ensure email was sent successfully
3. Try manual "I've Verified" button after clicking link
4. Check Firebase SDK is refreshing token (`reloadUserInfo()`)

## 📚 Next Steps

### Immediate
1. Configure Firebase for all platforms
2. Test complete auth flow on iOS/Android
3. Ensure backend `/auth/verify` endpoint is working

### Short Term
- Add error logging (Sentry, Firebase Crashlytics)
- Add analytics (Firebase Analytics)
- Implement remember me / biometric auth
- Add deep linking for email verification
- Add loading skeletons
- Add animations/transitions

### Medium Term
- Implement GraphQL client for plant data
- Build home screen with garden list
- Build calendar screen
- Implement camera feature
- Add plant search and filtering

## 📖 Resources

- [Flutter Documentation](https://flutter.dev/docs)
- [Firebase Flutter Setup](https://firebase.google.com/docs/flutter/setup)
- [Provider State Management](https://pub.dev/packages/provider)
- [Material Design 3](https://m3.material.io/)
- [Twigger Backend API](../docs/api/)

## 🤝 Contributing

This is a solo project. For questions or issues, refer to the main project documentation.

---

**Built with Flutter 3.9+ • Firebase • Material Design 3**
