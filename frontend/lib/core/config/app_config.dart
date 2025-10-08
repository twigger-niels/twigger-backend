/// App configuration
/// Environment-specific settings and constants
class AppConfig {
  // API Configuration
  static const String apiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://localhost:8080/api/v1',
  );

  static const String environment = String.fromEnvironment(
    'ENVIRONMENT',
    defaultValue: 'development',
  );

  // Feature Flags
  static const bool enableLogging = true; // Debug only

  // Timeouts
  static const Duration apiTimeout = Duration(seconds: 30);
  static const Duration emailVerificationCheckInterval = Duration(seconds: 5);
  static const Duration resendEmailCooldown = Duration(seconds: 60);

  // Validation Rules
  static const int minPasswordLength = 8;
  static const int minUsernameLength = 3;
  static const int maxUsernameLength = 20;

  // App Info
  static const String appName = 'Twigger';
  static const String appVersion = '1.0.0';

  // Firebase
  static const String firebaseProjectId = 'twigger-prod';

  // Helper methods
  static bool get isDevelopment => environment == 'development';
  static bool get isProduction => environment == 'production';
  static bool get isStaging => environment == 'staging';

  // Device identification (placeholder for real device ID)
  static String get deviceId => 'flutter-${DateTime.now().millisecondsSinceEpoch}';
}
