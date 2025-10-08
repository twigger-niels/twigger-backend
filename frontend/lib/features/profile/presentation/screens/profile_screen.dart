import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../../auth/presentation/providers/auth_provider.dart';
import '../../../../core/theme/app_theme.dart';

class ProfileScreen extends StatelessWidget {
  const ProfileScreen({super.key});

  Future<void> _handleSignOut(BuildContext context) async {
    final shouldSignOut = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Sign Out'),
        content: const Text('Are you sure you want to sign out?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(context, true),
            style: ElevatedButton.styleFrom(
              backgroundColor: AppTheme.errorRed,
            ),
            child: const Text('Sign Out'),
          ),
        ],
      ),
    );

    if (shouldSignOut == true && context.mounted) {
      try {
        await context.read<AuthProvider>().signOut();
      } catch (e) {
        if (context.mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(e.toString().replaceAll('Exception: ', '')),
              backgroundColor: AppTheme.errorRed,
            ),
          );
        }
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Profile'),
      ),
      body: Consumer<AuthProvider>(
        builder: (context, authProvider, _) {
          final user = authProvider.currentUser;

          if (user == null) {
            return const Center(
              child: Text('No user data'),
            );
          }

          return SingleChildScrollView(
            child: Column(
              children: [
                const SizedBox(height: 32),

                // Profile photo
                CircleAvatar(
                  radius: 60,
                  backgroundColor: AppTheme.primaryGreen.withOpacity(0.2),
                  backgroundImage: user.hasPhoto
                      ? NetworkImage(user.photoUrl!)
                      : null,
                  child: !user.hasPhoto
                      ? Text(
                          user.initials,
                          style: const TextStyle(
                            fontSize: 40,
                            fontWeight: FontWeight.bold,
                            color: AppTheme.primaryGreen,
                          ),
                        )
                      : null,
                ),
                const SizedBox(height: 16),

                // Username
                Text(
                  user.username,
                  style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                        fontWeight: FontWeight.bold,
                        color: AppTheme.primaryGreen,
                      ),
                ),
                const SizedBox(height: 8),

                // Email
                Text(
                  user.email,
                  style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                        color: Colors.grey[600],
                      ),
                ),
                const SizedBox(height: 24),

                // Provider badge
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 16,
                    vertical: 8,
                  ),
                  decoration: BoxDecoration(
                    color: AppTheme.primaryGreen.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(20),
                    border: Border.all(color: AppTheme.primaryGreen),
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(
                        _getProviderIcon(user.provider),
                        size: 20,
                        color: AppTheme.primaryGreen,
                      ),
                      const SizedBox(width: 8),
                      Text(
                        'Signed in with ${user.providerDisplayName}',
                        style: TextStyle(
                          color: AppTheme.primaryGreen,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 32),

                // User info card
                Card(
                  margin: const EdgeInsets.symmetric(horizontal: 16),
                  child: Column(
                    children: [
                      _buildInfoTile(
                        context,
                        icon: Icons.email_outlined,
                        title: 'Email',
                        subtitle: user.email,
                      ),
                      const Divider(height: 1),
                      _buildInfoTile(
                        context,
                        icon: Icons.verified_user_outlined,
                        title: 'Email Verified',
                        subtitle: user.emailVerified ? 'Yes' : 'No',
                        trailing: Icon(
                          user.emailVerified ? Icons.check_circle : Icons.cancel,
                          color: user.emailVerified
                              ? AppTheme.successGreen
                              : AppTheme.warningOrange,
                        ),
                      ),
                      const Divider(height: 1),
                      _buildInfoTile(
                        context,
                        icon: Icons.calendar_today_outlined,
                        title: 'Member Since',
                        subtitle: user.createdAt != null
                            ? _formatDate(user.createdAt!)
                            : 'N/A',
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 24),

                // Workspaces section
                if (user.workspaces.isNotEmpty) ...[
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 16),
                    child: Align(
                      alignment: Alignment.centerLeft,
                      child: Text(
                        'Workspaces',
                        style: Theme.of(context).textTheme.titleLarge?.copyWith(
                              fontWeight: FontWeight.bold,
                              color: AppTheme.primaryGreen,
                            ),
                      ),
                    ),
                  ),
                  const SizedBox(height: 12),
                  ...user.workspaces.map(
                    (workspace) => Card(
                      margin: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 4,
                      ),
                      child: ListTile(
                        leading: CircleAvatar(
                          backgroundColor: AppTheme.secondaryGreen,
                          child: Icon(
                            workspace.isOwner
                                ? Icons.star
                                : Icons.people_outline,
                            color: Colors.white,
                          ),
                        ),
                        title: Text(workspace.name),
                        subtitle: Text(workspace.roleDisplayName),
                        trailing: workspace.isOwner
                            ? const Icon(Icons.arrow_forward_ios, size: 16)
                            : null,
                      ),
                    ),
                  ),
                  const SizedBox(height: 24),
                ],

                // Settings section (placeholder)
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  child: Column(
                    children: [
                      ListTile(
                        leading: const Icon(Icons.settings_outlined),
                        title: const Text('Settings'),
                        trailing: const Icon(Icons.arrow_forward_ios, size: 16),
                        onTap: () {
                          ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(
                              content: Text('Settings coming soon!'),
                            ),
                          );
                        },
                      ),
                      ListTile(
                        leading: const Icon(Icons.help_outline),
                        title: const Text('Help & Support'),
                        trailing: const Icon(Icons.arrow_forward_ios, size: 16),
                        onTap: () {
                          ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(
                              content: Text('Help & Support coming soon!'),
                            ),
                          );
                        },
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 24),

                // Sign out button
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  child: ElevatedButton.icon(
                    onPressed: authProvider.isLoading
                        ? null
                        : () => _handleSignOut(context),
                    icon: const Icon(Icons.logout),
                    label: const Text('Sign Out'),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: AppTheme.errorRed,
                      minimumSize: const Size(double.infinity, 50),
                    ),
                  ),
                ),
                const SizedBox(height: 32),

                // App version
                Text(
                  'Version 1.0.0',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: Colors.grey[500],
                      ),
                ),
                const SizedBox(height: 32),
              ],
            ),
          );
        },
      ),
    );
  }

  Widget _buildInfoTile(
    BuildContext context, {
    required IconData icon,
    required String title,
    required String subtitle,
    Widget? trailing,
  }) {
    return ListTile(
      leading: Icon(icon, color: AppTheme.primaryGreen),
      title: Text(title),
      subtitle: Text(subtitle),
      trailing: trailing,
    );
  }

  IconData _getProviderIcon(String provider) {
    switch (provider) {
      case 'google.com':
        return Icons.g_mobiledata;
      case 'facebook.com':
        return Icons.facebook;
      case 'apple.com':
        return Icons.apple;
      case 'password':
        return Icons.email;
      default:
        return Icons.person;
    }
  }

  String _formatDate(DateTime date) {
    final months = [
      'Jan',
      'Feb',
      'Mar',
      'Apr',
      'May',
      'Jun',
      'Jul',
      'Aug',
      'Sep',
      'Oct',
      'Nov',
      'Dec'
    ];
    return '${months[date.month - 1]} ${date.day}, ${date.year}';
  }
}
