import '../../../core/constants/app_constants.dart';

/// Workspace model representing a user's workspace/garden
class WorkspaceModel {
  final String workspaceId;
  final String name;
  final WorkspaceRole role;
  final DateTime? createdAt;
  final DateTime? joinedAt;

  WorkspaceModel({
    required this.workspaceId,
    required this.name,
    required this.role,
    this.createdAt,
    this.joinedAt,
  });

  /// Factory constructor to create WorkspaceModel from JSON
  factory WorkspaceModel.fromJson(Map<String, dynamic> json) {
    return WorkspaceModel(
      workspaceId: json['workspace_id'] as String,
      name: json['name'] as String,
      role: WorkspaceRoleExtension.fromString(json['role'] as String? ?? 'viewer'),
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : null,
      joinedAt: json['joined_at'] != null
          ? DateTime.parse(json['joined_at'] as String)
          : null,
    );
  }

  /// Convert WorkspaceModel to JSON
  Map<String, dynamic> toJson() {
    return {
      'workspace_id': workspaceId,
      'name': name,
      'role': role.value,
      'created_at': createdAt?.toIso8601String(),
      'joined_at': joinedAt?.toIso8601String(),
    };
  }

  /// Create a copy of WorkspaceModel with updated fields
  WorkspaceModel copyWith({
    String? workspaceId,
    String? name,
    WorkspaceRole? role,
    DateTime? createdAt,
    DateTime? joinedAt,
  }) {
    return WorkspaceModel(
      workspaceId: workspaceId ?? this.workspaceId,
      name: name ?? this.name,
      role: role ?? this.role,
      createdAt: createdAt ?? this.createdAt,
      joinedAt: joinedAt ?? this.joinedAt,
    );
  }

  /// Check if user is owner
  bool get isOwner => role == WorkspaceRole.owner;

  /// Check if user is member
  bool get isMember => role == WorkspaceRole.member;

  /// Check if user is viewer
  bool get isViewer => role == WorkspaceRole.viewer;

  /// Check if user has write permissions
  bool get canWrite => role == WorkspaceRole.owner || role == WorkspaceRole.member;

  /// Check if user has admin permissions
  bool get canAdmin => role == WorkspaceRole.owner;

  /// Get role display name
  String get roleDisplayName {
    switch (role) {
      case WorkspaceRole.owner:
        return 'Owner';
      case WorkspaceRole.member:
        return 'Member';
      case WorkspaceRole.viewer:
        return 'Viewer';
    }
  }

  @override
  String toString() {
    return 'WorkspaceModel(workspaceId: $workspaceId, name: $name, role: ${role.value})';
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;

    return other is WorkspaceModel &&
        other.workspaceId == workspaceId &&
        other.name == name &&
        other.role == role;
  }

  @override
  int get hashCode => workspaceId.hashCode ^ name.hashCode ^ role.hashCode;
}
