/// App-wide constants and enums

class AppConstants {
  // Routes
  static const String splashRoute = '/';
  static const String loginRoute = '/login';
  static const String registerRoute = '/register';
  static const String forgotPasswordRoute = '/forgot-password';
  static const String emailVerificationRoute = '/email-verification';
  static const String homeRoute = '/home';

  // Storage Keys
  static const String authTokenKey = 'auth_token';
  static const String userDataKey = 'user_data';
  static const String rememberMeKey = 'remember_me';

  // Error Messages
  static const String networkErrorMessage =
      'Network error. Please check your connection.';
  static const String genericErrorMessage =
      'Something went wrong. Please try again.';
  static const String authErrorMessage = 'Authentication failed. Please try again.';

  // Success Messages
  static const String loginSuccessMessage = 'Logged in successfully!';
  static const String registerSuccessMessage = 'Account created successfully!';
  static const String logoutSuccessMessage = 'Logged out successfully!';
  static const String passwordResetEmailSent =
      'Password reset email sent. Please check your inbox.';
  static const String emailVerificationSent =
      'Verification email sent. Please check your inbox.';

  // Regex Patterns
  static final RegExp emailRegex = RegExp(
    r'^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$',
  );
  static final RegExp usernameRegex = RegExp(
    r'^[a-zA-Z0-9_]+$',
  );

  // Validation
  static const int minPasswordLength = 8;
  static const int minUsernameLength = 3;
  static const int maxUsernameLength = 20;
}

/// Auth state enum
enum AuthState {
  initial,
  authenticating,
  authenticated,
  unauthenticated,
  emailVerificationPending,
  error,
}

/// Provider types
enum AuthProvider {
  email,
  google,
  facebook,
  apple,
}

extension AuthProviderExtension on AuthProvider {
  String get value {
    switch (this) {
      case AuthProvider.email:
        return 'password';
      case AuthProvider.google:
        return 'google.com';
      case AuthProvider.facebook:
        return 'facebook.com';
      case AuthProvider.apple:
        return 'apple.com';
    }
  }
}

/// Workspace roles
enum WorkspaceRole {
  owner,
  member,
  viewer,
}

extension WorkspaceRoleExtension on WorkspaceRole {
  String get value {
    switch (this) {
      case WorkspaceRole.owner:
        return 'owner';
      case WorkspaceRole.member:
        return 'member';
      case WorkspaceRole.viewer:
        return 'viewer';
    }
  }

  static WorkspaceRole fromString(String role) {
    switch (role.toLowerCase()) {
      case 'owner':
        return WorkspaceRole.owner;
      case 'member':
        return WorkspaceRole.member;
      case 'viewer':
        return WorkspaceRole.viewer;
      default:
        return WorkspaceRole.viewer;
    }
  }
}
