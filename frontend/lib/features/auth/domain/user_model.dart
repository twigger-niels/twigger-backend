import 'workspace_model.dart';

/// User model representing authenticated user data
class UserModel {
  final String userId;
  final String email;
  final String username;
  final String? photoUrl;
  final String provider;
  final bool emailVerified;
  final List<WorkspaceModel> workspaces;
  final DateTime? createdAt;
  final DateTime? lastLoginAt;

  UserModel({
    required this.userId,
    required this.email,
    required this.username,
    this.photoUrl,
    required this.provider,
    required this.emailVerified,
    this.workspaces = const [],
    this.createdAt,
    this.lastLoginAt,
  });

  /// Factory constructor to create UserModel from JSON
  factory UserModel.fromJson(Map<String, dynamic> json) {
    return UserModel(
      userId: json['user_id'] as String,
      email: json['email'] as String,
      username: json['username'] as String,
      photoUrl: json['photo_url'] as String?,
      provider: json['provider'] as String,
      emailVerified: json['email_verified'] as bool? ?? false,
      workspaces: (json['workspaces'] as List<dynamic>?)
              ?.map((w) => WorkspaceModel.fromJson(w as Map<String, dynamic>))
              .toList() ??
          [],
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : null,
      lastLoginAt: json['last_login_at'] != null
          ? DateTime.parse(json['last_login_at'] as String)
          : null,
    );
  }

  /// Convert UserModel to JSON
  Map<String, dynamic> toJson() {
    return {
      'user_id': userId,
      'email': email,
      'username': username,
      'photo_url': photoUrl,
      'provider': provider,
      'email_verified': emailVerified,
      'workspaces': workspaces.map((w) => w.toJson()).toList(),
      'created_at': createdAt?.toIso8601String(),
      'last_login_at': lastLoginAt?.toIso8601String(),
    };
  }

  /// Create a copy of UserModel with updated fields
  UserModel copyWith({
    String? userId,
    String? email,
    String? username,
    String? photoUrl,
    String? provider,
    bool? emailVerified,
    List<WorkspaceModel>? workspaces,
    DateTime? createdAt,
    DateTime? lastLoginAt,
  }) {
    return UserModel(
      userId: userId ?? this.userId,
      email: email ?? this.email,
      username: username ?? this.username,
      photoUrl: photoUrl ?? this.photoUrl,
      provider: provider ?? this.provider,
      emailVerified: emailVerified ?? this.emailVerified,
      workspaces: workspaces ?? this.workspaces,
      createdAt: createdAt ?? this.createdAt,
      lastLoginAt: lastLoginAt ?? this.lastLoginAt,
    );
  }

  /// Get display name (username or email)
  String get displayName => username.isNotEmpty ? username : email;

  /// Check if user has a profile photo
  bool get hasPhoto => photoUrl != null && photoUrl!.isNotEmpty;

  /// Get initials for avatar (first letter of username)
  String get initials {
    if (username.isNotEmpty) {
      return username[0].toUpperCase();
    } else if (email.isNotEmpty) {
      return email[0].toUpperCase();
    }
    return '?';
  }

  /// Check if email provider
  bool get isEmailProvider => provider == 'password';

  /// Check if social provider
  bool get isSocialProvider =>
      provider == 'google.com' ||
      provider == 'facebook.com' ||
      provider == 'apple.com';

  /// Get provider display name
  String get providerDisplayName {
    switch (provider) {
      case 'google.com':
        return 'Google';
      case 'facebook.com':
        return 'Facebook';
      case 'apple.com':
        return 'Apple';
      case 'password':
        return 'Email';
      default:
        return provider;
    }
  }

  @override
  String toString() {
    return 'UserModel(userId: $userId, email: $email, username: $username, provider: $provider)';
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;

    return other is UserModel &&
        other.userId == userId &&
        other.email == email &&
        other.username == username;
  }

  @override
  int get hashCode => userId.hashCode ^ email.hashCode ^ username.hashCode;
}
