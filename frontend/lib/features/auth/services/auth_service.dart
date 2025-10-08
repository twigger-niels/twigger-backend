import 'package:firebase_auth/firebase_auth.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:flutter_facebook_auth/flutter_facebook_auth.dart';
import '../data/auth_api_client.dart';
import '../domain/user_model.dart';
import '../../../core/config/app_config.dart';

/// Auth service handling Firebase authentication and backend integration
class AuthService {
  final FirebaseAuth _firebaseAuth;
  final GoogleSignIn _googleSignIn;
  final FacebookAuth _facebookAuth;
  final AuthApiClient _apiClient;

  AuthService({
    FirebaseAuth? firebaseAuth,
    GoogleSignIn? googleSignIn,
    FacebookAuth? facebookAuth,
    AuthApiClient? apiClient,
  })  : _firebaseAuth = firebaseAuth ?? FirebaseAuth.instance,
        _googleSignIn = googleSignIn ?? GoogleSignIn(scopes: ['email', 'profile']),
        _facebookAuth = facebookAuth ?? FacebookAuth.instance,
        _apiClient = apiClient ?? AuthApiClient();

  /// Get current Firebase user
  User? get currentUser => _firebaseAuth.currentUser;

  /// Auth state stream
  Stream<User?> get authStateChanges => _firebaseAuth.authStateChanges();

  // ===== EMAIL/PASSWORD AUTHENTICATION =====

  /// Sign up with email and password
  Future<UserCredential> signUpWithEmail({
    required String email,
    required String password,
  }) async {
    try {
      final userCredential = await _firebaseAuth.createUserWithEmailAndPassword(
        email: email,
        password: password,
      );

      // Send email verification
      await userCredential.user?.sendEmailVerification();

      return userCredential;
    } on FirebaseAuthException catch (e) {
      throw _handleFirebaseAuthException(e);
    }
  }

  /// Sign in with email and password
  Future<UserCredential> signInWithEmail({
    required String email,
    required String password,
  }) async {
    try {
      final userCredential = await _firebaseAuth.signInWithEmailAndPassword(
        email: email,
        password: password,
      );

      return userCredential;
    } on FirebaseAuthException catch (e) {
      throw _handleFirebaseAuthException(e);
    }
  }

  /// Send email verification
  Future<void> sendEmailVerification() async {
    try {
      await currentUser?.sendEmailVerification();
    } on FirebaseAuthException catch (e) {
      throw _handleFirebaseAuthException(e);
    }
  }

  /// Send password reset email
  Future<void> sendPasswordResetEmail(String email) async {
    try {
      await _firebaseAuth.sendPasswordResetEmail(email: email);
    } on FirebaseAuthException catch (e) {
      throw _handleFirebaseAuthException(e);
    }
  }

  /// Reload current user to get latest email verification status
  Future<void> reloadUser() async {
    await currentUser?.reload();
  }

  /// Check if email is verified
  Future<bool> isEmailVerified() async {
    if (AppConfig.enableLogging) {
      print('Checking email verification - before reload: ${currentUser?.emailVerified}');
    }

    await reloadUser();

    final verified = currentUser?.emailVerified ?? false;

    if (AppConfig.enableLogging) {
      print('After reload - emailVerified: $verified');
      print('Current user: ${currentUser?.email}');
      print('Current user UID: ${currentUser?.uid}');
    }

    return verified;
  }

  // ===== GOOGLE SIGN-IN =====

  /// Sign in with Google
  Future<UserCredential?> signInWithGoogle() async {
    try {
      // Trigger Google Sign-In flow
      final GoogleSignInAccount? googleUser = await _googleSignIn.signIn();

      if (googleUser == null) {
        // User cancelled the sign-in
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

      return userCredential;
    } on FirebaseAuthException catch (e) {
      throw _handleFirebaseAuthException(e);
    } catch (e) {
      throw Exception('Google sign-in failed: $e');
    }
  }

  // ===== FACEBOOK LOGIN =====

  /// Sign in with Facebook
  Future<UserCredential?> signInWithFacebook() async {
    try {
      if (AppConfig.enableLogging) {
        print('Starting Facebook login...');
      }

      // Trigger Facebook Login flow
      final LoginResult result = await _facebookAuth.login(
        permissions: ['email', 'public_profile'],
      );

      if (AppConfig.enableLogging) {
        print('Facebook login result status: ${result.status}');
      }

      if (result.status != LoginStatus.success) {
        // User cancelled or error occurred
        if (AppConfig.enableLogging) {
          print('Facebook login failed or cancelled: ${result.status}');
          if (result.message != null) {
            print('Facebook error message: ${result.message}');
          }
        }
        return null;
      }

      // Create Firebase credential
      final credential = FacebookAuthProvider.credential(
        result.accessToken!.token,
      );

      // Sign in to Firebase
      final userCredential = await _firebaseAuth.signInWithCredential(credential);

      if (AppConfig.enableLogging) {
        print('Facebook sign-in successful');
      }

      return userCredential;
    } on FirebaseAuthException catch (e) {
      throw _handleFirebaseAuthException(e);
    } catch (e) {
      if (AppConfig.enableLogging) {
        print('Facebook sign-in exception: $e');
      }
      throw Exception('Facebook sign-in failed: $e');
    }
  }

  // ===== BACKEND INTEGRATION =====

  /// Complete authentication with backend
  /// Called after successful Firebase authentication
  Future<UserModel> completeAuthentication({
    required UserCredential userCredential,
  }) async {
    try {
      // Get Firebase ID token
      final idToken = await userCredential.user?.getIdToken();

      if (idToken == null) {
        throw Exception('Failed to get Firebase ID token');
      }

      // Call backend /auth/verify endpoint
      final response = await _apiClient.verifyAuth(
        firebaseToken: idToken,
        deviceId: AppConfig.deviceId,
      );

      // Parse user data from backend response
      final userData = response['user'] as Map<String, dynamic>;
      final userModel = UserModel.fromJson(userData);

      return userModel;
    } catch (e) {
      throw Exception('Failed to complete authentication: $e');
    }
  }

  /// Get user profile from backend
  Future<UserModel?> getUserProfile() async {
    try {
      final idToken = await currentUser?.getIdToken();

      if (idToken == null) {
        return null;
      }

      final response = await _apiClient.getProfile(firebaseToken: idToken);

      if (response == null) {
        return null;
      }

      final userData = response['user'] as Map<String, dynamic>;
      return UserModel.fromJson(userData);
    } catch (e) {
      if (AppConfig.enableLogging) {
        print('Error getting user profile: $e');
      }
      return null;
    }
  }

  /// Register new user (creates Firebase account only)
  /// Backend registration with auto-generated username happens after email verification
  Future<UserModel> registerUser({
    required String email,
    required String password,
  }) async {
    try {
      // Create Firebase user and send verification email
      final userCredential = await signUpWithEmail(
        email: email,
        password: password,
      );

      // Return a temporary UserModel (backend registration happens after verification)
      return UserModel(
        userId: '',  // Will be set after backend registration
        email: email,
        username: '',  // Will be auto-generated by backend after verification
        provider: 'password',
        emailVerified: false,
      );
    } catch (e) {
      throw Exception('Registration failed: $e');
    }
  }

  /// Complete registration after email verification (calls /auth/register)
  Future<UserModel> completeRegistration() async {
    try {
      // Get Firebase ID token
      final idToken = await currentUser?.getIdToken();

      if (idToken == null) {
        throw Exception('Failed to get Firebase ID token');
      }

      // Call backend registration endpoint (username auto-generated)
      final response = await _apiClient.registerUser(
        firebaseToken: idToken,
        deviceId: AppConfig.deviceId,
      );

      // Parse user data
      final userData = response['user'] as Map<String, dynamic>;
      return UserModel.fromJson(userData);
    } catch (e) {
      throw Exception('Registration completion failed: $e');
    }
  }

  // ===== SIGN OUT =====

  /// Sign out from Firebase and backend
  Future<void> signOut() async {
    try {
      // Get token before signing out (for backend call)
      final idToken = await currentUser?.getIdToken();

      // Call backend logout
      if (idToken != null) {
        try {
          await _apiClient.logout(
            firebaseToken: idToken,
            revokeAll: true,
          );
        } catch (e) {
          // Non-blocking - continue with Firebase sign out even if backend fails
          if (AppConfig.enableLogging) {
            print('Backend logout failed: $e');
          }
        }
      }

      // Sign out from Firebase
      await _firebaseAuth.signOut();

      // Sign out from Google
      await _googleSignIn.signOut();

      // Sign out from Facebook
      await _facebookAuth.logOut();
    } catch (e) {
      throw Exception('Sign out failed: $e');
    }
  }

  // ===== HELPER METHODS =====

  /// Handle Firebase auth exceptions and convert to user-friendly messages
  Exception _handleFirebaseAuthException(FirebaseAuthException e) {
    String message;

    switch (e.code) {
      case 'user-not-found':
        message = 'No account found with this email.';
        break;
      case 'wrong-password':
        message = 'Incorrect password.';
        break;
      case 'email-already-in-use':
        message = 'An account already exists with this email.';
        break;
      case 'weak-password':
        message = 'Password is too weak. Use at least 8 characters.';
        break;
      case 'invalid-email':
        message = 'Invalid email address.';
        break;
      case 'user-disabled':
        message = 'This account has been disabled.';
        break;
      case 'too-many-requests':
        message = 'Too many attempts. Please try again later.';
        break;
      case 'operation-not-allowed':
        message = 'This sign-in method is not enabled.';
        break;
      case 'invalid-credential':
        message = 'Invalid credentials. Please try again.';
        break;
      case 'account-exists-with-different-credential':
        message = 'An account already exists with this email using a different sign-in method.';
        break;
      case 'requires-recent-login':
        message = 'This operation requires recent authentication. Please sign in again.';
        break;
      default:
        message = 'Authentication failed: ${e.message ?? 'Unknown error'}';
    }

    if (AppConfig.enableLogging) {
      print('FirebaseAuthException: ${e.code} - ${e.message}');
    }

    return Exception(message);
  }

  /// Get provider display name from Firebase provider ID
  String getProviderDisplayName(String providerId) {
    switch (providerId) {
      case 'google.com':
        return 'Google';
      case 'facebook.com':
        return 'Facebook';
      case 'apple.com':
        return 'Apple';
      case 'password':
        return 'Email';
      default:
        return providerId;
    }
  }

  /// Check if user needs email verification
  bool needsEmailVerification() {
    final user = currentUser;
    if (user == null) return false;

    // Only email/password users need email verification
    final isPasswordProvider = user.providerData
        .any((provider) => provider.providerId == 'password');

    return isPasswordProvider && !user.emailVerified;
  }
}
