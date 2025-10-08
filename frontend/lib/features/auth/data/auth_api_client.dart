import 'dart:convert';
import 'package:http/http.dart' as http;
import '../../../core/config/app_config.dart';

/// API client for authentication endpoints
class AuthApiClient {
  final String baseUrl;
  final http.Client httpClient;

  AuthApiClient({
    String? baseUrl,
    http.Client? httpClient,
  })  : baseUrl = baseUrl ?? AppConfig.apiBaseUrl,
        httpClient = httpClient ?? http.Client();

  /// Verify authentication with backend
  /// POST /api/v1/auth/verify
  Future<Map<String, dynamic>> verifyAuth({
    required String firebaseToken,
    String? deviceId,
  }) async {
    try {
      final response = await httpClient
          .post(
            Uri.parse('$baseUrl/auth/verify'),
            headers: {
              'Authorization': 'Bearer $firebaseToken',
              'Content-Type': 'application/json',
            },
            body: jsonEncode({
              if (deviceId != null) 'device_id': deviceId,
            }),
          )
          .timeout(AppConfig.apiTimeout);

      if (response.statusCode == 200) {
        return jsonDecode(response.body) as Map<String, dynamic>;
      } else if (response.statusCode == 429) {
        throw Exception('Too many requests. Please try again later.');
      } else if (response.statusCode == 401) {
        throw Exception('Unauthorized. Please sign in again.');
      } else if (response.statusCode == 403) {
        // Email not verified for password provider
        final errorData = jsonDecode(response.body);
        throw Exception(errorData['error'] ?? 'Forbidden');
      } else {
        throw Exception('Authentication failed: ${response.statusCode}');
      }
    } catch (e) {
      if (e.toString().contains('TimeoutException')) {
        throw Exception('Request timed out. Please check your connection.');
      }
      rethrow;
    }
  }

  /// Register new user (username auto-generated from email)
  /// POST /api/v1/auth/register
  Future<Map<String, dynamic>> registerUser({
    required String firebaseToken,
    String? deviceId,
  }) async {
    try {
      final response = await httpClient
          .post(
            Uri.parse('$baseUrl/auth/register'),
            headers: {
              'Authorization': 'Bearer $firebaseToken',
              'Content-Type': 'application/json',
            },
            body: jsonEncode({
              if (deviceId != null) 'device_id': deviceId,
            }),
          )
          .timeout(AppConfig.apiTimeout);

      if (response.statusCode == 200 || response.statusCode == 201) {
        return jsonDecode(response.body) as Map<String, dynamic>;
      } else if (response.statusCode == 429) {
        throw Exception('Too many requests. Please try again later.');
      } else {
        throw Exception('Registration failed: ${response.statusCode}');
      }
    } catch (e) {
      if (e.toString().contains('TimeoutException')) {
        throw Exception('Request timed out. Please check your connection.');
      }
      rethrow;
    }
  }

  /// Get user profile
  /// GET /api/v1/auth/me
  Future<Map<String, dynamic>?> getProfile({
    required String firebaseToken,
  }) async {
    try {
      final response = await httpClient
          .get(
            Uri.parse('$baseUrl/auth/me'),
            headers: {
              'Authorization': 'Bearer $firebaseToken',
            },
          )
          .timeout(AppConfig.apiTimeout);

      if (response.statusCode == 200) {
        return jsonDecode(response.body) as Map<String, dynamic>;
      } else if (response.statusCode == 401) {
        // Token expired or invalid
        return null;
      } else {
        throw Exception('Failed to get profile: ${response.statusCode}');
      }
    } catch (e) {
      if (e.toString().contains('TimeoutException')) {
        throw Exception('Request timed out. Please check your connection.');
      }
      if (AppConfig.enableLogging) {
        print('Error getting profile: $e');
      }
      return null;
    }
  }

  /// Logout
  /// POST /api/v1/auth/logout
  Future<void> logout({
    required String firebaseToken,
    bool revokeAll = false,
    String? deviceId,
  }) async {
    try {
      final response = await httpClient
          .post(
            Uri.parse('$baseUrl/auth/logout'),
            headers: {
              'Authorization': 'Bearer $firebaseToken',
              'Content-Type': 'application/json',
            },
            body: jsonEncode({
              'revoke_all': revokeAll,
              if (deviceId != null) 'device_id': deviceId,
            }),
          )
          .timeout(AppConfig.apiTimeout);

      if (response.statusCode != 200) {
        throw Exception('Logout failed: ${response.statusCode}');
      }
    } catch (e) {
      if (e.toString().contains('TimeoutException')) {
        throw Exception('Request timed out. Please check your connection.');
      }
      rethrow;
    }
  }

  /// Dispose HTTP client
  void dispose() {
    httpClient.close();
  }
}
