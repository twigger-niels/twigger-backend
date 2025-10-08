import 'package:flutter/foundation.dart';
import 'package:firebase_auth/firebase_auth.dart';
import '../../services/auth_service.dart';
import '../../domain/user_model.dart';
import '../../../../core/constants/app_constants.dart';
import '../../../../core/config/app_config.dart';

/// Authentication provider managing auth state
class AuthProvider extends ChangeNotifier {
  final AuthService _authService;

  // State
  AuthState _state = AuthState.initial;
  UserModel? _currentUser;
  String? _errorMessage;
  bool _isLoading = false;

  // Getters
  AuthState get state => _state;
  UserModel? get currentUser => _currentUser;
  String? get errorMessage => _errorMessage;
  bool get isLoading => _isLoading;
  bool get isAuthenticated => _state == AuthState.authenticated;

  AuthProvider({AuthService? authService})
      : _authService = authService ?? AuthService() {
    // Initialize - check auth state on startup
    _initialize();
  }

  /// Initialize and listen to Firebase auth state changes
  void _initialize() {
    _authService.authStateChanges.listen((User? user) {
      if (user == null) {
        // User signed out
        _state = AuthState.unauthenticated;
        _currentUser = null;
        notifyListeners();
      } else {
        // User signed in - check if email verification needed
        if (_authService.needsEmailVerification()) {
          _state = AuthState.emailVerificationPending;
          notifyListeners();
        }
      }
    });
  }

  // ===== EMAIL/PASSWORD AUTHENTICATION =====

  /// Register with email and password (username auto-generated from email)
  Future<void> registerWithEmail({
    required String email,
    required String password,
  }) async {
    try {
      _setLoading(true);
      _setError(null);

      // Register user (creates Firebase account + sends verification email)
      final user = await _authService.registerUser(
        email: email,
        password: password,
      );

      _currentUser = user;
      _state = AuthState.emailVerificationPending;

      notifyListeners();
    } catch (e) {
      _setError(e.toString());
      _state = AuthState.error;
      notifyListeners();
      rethrow;
    } finally {
      _setLoading(false);
    }
  }

  /// Sign in with email and password
  Future<void> signInWithEmail({
    required String email,
    required String password,
  }) async {
    try {
      _setLoading(true);
      _setError(null);
      _state = AuthState.authenticating;
      notifyListeners();

      // Sign in with Firebase
      final userCredential = await _authService.signInWithEmail(
        email: email,
        password: password,
      );

      // Check if email verification is needed
      if (_authService.needsEmailVerification()) {
        _state = AuthState.emailVerificationPending;
        notifyListeners();
        return;
      }

      // Complete authentication with backend
      final user = await _authService.completeAuthentication(
        userCredential: userCredential,
      );

      _currentUser = user;
      _state = AuthState.authenticated;
      notifyListeners();
    } catch (e) {
      _setError(e.toString());
      _state = AuthState.error;
      notifyListeners();
      rethrow;
    } finally {
      _setLoading(false);
    }
  }

  /// Send email verification
  Future<void> sendEmailVerification() async {
    try {
      _setLoading(true);
      _setError(null);

      await _authService.sendEmailVerification();

      notifyListeners();
    } catch (e) {
      _setError(e.toString());
      rethrow;
    } finally {
      _setLoading(false);
    }
  }

  /// Check if email is verified and complete registration/authentication
  Future<bool> checkEmailVerified() async {
    try {
      _setLoading(true);
      _setError(null);

      final isVerified = await _authService.isEmailVerified();

      if (AppConfig.enableLogging) {
        print('checkEmailVerified: isVerified = $isVerified');
      }

      if (isVerified) {
        // Email is verified - complete registration with backend
        final user = _authService.currentUser;
        if (user != null) {
          if (AppConfig.enableLogging) {
            print('Email verified! Completing registration...');
          }
          // For email/password users, complete registration (calls /auth/register)
          // This will auto-generate username and create user in backend
          final userModel = await _authService.completeRegistration();

          if (AppConfig.enableLogging) {
            print('Registration completed! User: ${userModel.username}');
          }

          _currentUser = userModel;
          _state = AuthState.authenticated;
          notifyListeners();
          return true;
        }
      }

      return false;
    } catch (e) {
      if (AppConfig.enableLogging) {
        print('checkEmailVerified ERROR: $e');
      }
      _setError(e.toString());
      return false;
    } finally {
      _setLoading(false);
    }
  }

  /// Send password reset email
  Future<void> resetPassword(String email) async {
    try {
      _setLoading(true);
      _setError(null);

      await _authService.sendPasswordResetEmail(email);

      notifyListeners();
    } catch (e) {
      _setError(e.toString());
      rethrow;
    } finally {
      _setLoading(false);
    }
  }

  // ===== GOOGLE SIGN-IN =====

  /// Sign in with Google
  Future<void> signInWithGoogle() async {
    try {
      _setLoading(true);
      _setError(null);
      _state = AuthState.authenticating;
      notifyListeners();

      // Sign in with Google via Firebase
      final userCredential = await _authService.signInWithGoogle();

      if (userCredential == null) {
        // User cancelled
        _state = AuthState.unauthenticated;
        notifyListeners();
        return;
      }

      // Complete authentication with backend
      final user = await _authService.completeAuthentication(
        userCredential: userCredential,
      );

      _currentUser = user;
      _state = AuthState.authenticated;
      notifyListeners();
    } catch (e) {
      _setError(e.toString());
      _state = AuthState.error;
      notifyListeners();
      rethrow;
    } finally {
      _setLoading(false);
    }
  }

  // ===== FACEBOOK LOGIN =====

  /// Sign in with Facebook
  Future<void> signInWithFacebook() async {
    try {
      _setLoading(true);
      _setError(null);
      _state = AuthState.authenticating;
      notifyListeners();

      // Sign in with Facebook via Firebase
      final userCredential = await _authService.signInWithFacebook();

      if (userCredential == null) {
        // User cancelled
        _state = AuthState.unauthenticated;
        notifyListeners();
        return;
      }

      // Complete authentication with backend
      final user = await _authService.completeAuthentication(
        userCredential: userCredential,
      );

      _currentUser = user;
      _state = AuthState.authenticated;
      notifyListeners();
    } catch (e) {
      _setError(e.toString());
      _state = AuthState.error;
      notifyListeners();
      rethrow;
    } finally {
      _setLoading(false);
    }
  }

  // ===== SIGN OUT =====

  /// Sign out
  Future<void> signOut() async {
    try {
      _setLoading(true);
      _setError(null);

      await _authService.signOut();

      _currentUser = null;
      _state = AuthState.unauthenticated;
      notifyListeners();
    } catch (e) {
      _setError(e.toString());
      rethrow;
    } finally {
      _setLoading(false);
    }
  }

  // ===== USER PROFILE =====

  /// Refresh user profile from backend
  Future<void> refreshUserProfile() async {
    try {
      _setLoading(true);
      _setError(null);

      final user = await _authService.getUserProfile();

      if (user != null) {
        _currentUser = user;
        notifyListeners();
      }
    } catch (e) {
      _setError(e.toString());
    } finally {
      _setLoading(false);
    }
  }

  // ===== HELPER METHODS =====

  void _setLoading(bool loading) {
    _isLoading = loading;
  }

  void _setError(String? error) {
    _errorMessage = error;
  }

  /// Clear error message
  void clearError() {
    _errorMessage = null;
    notifyListeners();
  }

  /// Check authentication state on app start
  Future<void> checkAuthState() async {
    final user = _authService.currentUser;

    if (user == null) {
      _state = AuthState.unauthenticated;
    } else if (_authService.needsEmailVerification()) {
      _state = AuthState.emailVerificationPending;
    } else {
      // User is signed in - get profile from backend
      try {
        final userModel = await _authService.getUserProfile();
        if (userModel != null) {
          _currentUser = userModel;
          _state = AuthState.authenticated;
        } else {
          _state = AuthState.unauthenticated;
        }
      } catch (e) {
        _state = AuthState.unauthenticated;
      }
    }

    notifyListeners();
  }
}
