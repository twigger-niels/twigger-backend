import 'dart:async';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/auth_provider.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../../core/config/app_config.dart';

class EmailVerificationScreen extends StatefulWidget {
  const EmailVerificationScreen({super.key});

  @override
  State<EmailVerificationScreen> createState() => _EmailVerificationScreenState();
}

class _EmailVerificationScreenState extends State<EmailVerificationScreen> {
  Timer? _verificationCheckTimer;
  Timer? _resendCooldownTimer;
  int _resendCooldownSeconds = 0;
  bool _isResending = false;

  @override
  void initState() {
    super.initState();
    _startVerificationCheck();
  }

  @override
  void dispose() {
    _verificationCheckTimer?.cancel();
    _resendCooldownTimer?.cancel();
    super.dispose();
  }

  void _startVerificationCheck() {
    // Check email verification status every 5 seconds
    _verificationCheckTimer = Timer.periodic(
      AppConfig.emailVerificationCheckInterval,
      (timer) async {
        final authProvider = context.read<AuthProvider>();
        final isVerified = await authProvider.checkEmailVerified();

        if (isVerified && mounted) {
          timer.cancel();
          // Navigation handled by AuthWrapper
        }
      },
    );
  }

  Future<void> _resendVerificationEmail() async {
    if (_resendCooldownSeconds > 0) return;

    setState(() {
      _isResending = true;
    });

    final authProvider = context.read<AuthProvider>();

    try {
      await authProvider.sendEmailVerification();

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Verification email sent! Please check your inbox.'),
            backgroundColor: AppTheme.successGreen,
          ),
        );

        // Start cooldown timer
        setState(() {
          _resendCooldownSeconds = AppConfig.resendEmailCooldown.inSeconds;
        });

        _resendCooldownTimer = Timer.periodic(
          const Duration(seconds: 1),
          (timer) {
            setState(() {
              _resendCooldownSeconds--;
            });

            if (_resendCooldownSeconds <= 0) {
              timer.cancel();
            }
          },
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(e.toString().replaceAll('Exception: ', '')),
            backgroundColor: AppTheme.errorRed,
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() {
          _isResending = false;
        });
      }
    }
  }

  Future<void> _checkManually() async {
    final authProvider = context.read<AuthProvider>();
    final isVerified = await authProvider.checkEmailVerified();

    if (mounted && !isVerified) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Email not yet verified. Please check your inbox.'),
          backgroundColor: AppTheme.warningOrange,
        ),
      );
    }
  }

  Future<void> _signOut() async {
    final authProvider = context.read<AuthProvider>();
    await authProvider.signOut();
  }

  @override
  Widget build(BuildContext context) {
    final authProvider = context.watch<AuthProvider>();
    final email = authProvider.currentUser?.email ?? '';

    return Scaffold(
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24.0),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // Email icon
                Icon(
                  Icons.mark_email_unread_outlined,
                  size: 100,
                  color: AppTheme.primaryGreen,
                ),
                const SizedBox(height: 32),

                // Title
                Text(
                  'Verify Your Email',
                  style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                        color: AppTheme.primaryGreen,
                      ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 16),

                // Instructions
                Text(
                  'We sent a verification email to:',
                  style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                        color: Colors.grey[700],
                      ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 8),

                // Email address
                Text(
                  email,
                  style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                        fontWeight: FontWeight.w600,
                        color: AppTheme.primaryGreen,
                      ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 24),

                // Instructions
                Text(
                  'Please check your email and click the verification link to continue.',
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: Colors.grey[600],
                      ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 8),

                // Auto-check notice
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: AppTheme.primaryGreen.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Row(
                    children: [
                      Icon(
                        Icons.info_outline,
                        color: AppTheme.primaryGreen,
                        size: 20,
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          'We\'re checking automatically. You\'ll be redirected once verified.',
                          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                color: AppTheme.primaryGreen,
                              ),
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 32),

                // Check manually button
                OutlinedButton.icon(
                  onPressed: _checkManually,
                  icon: const Icon(Icons.refresh),
                  label: const Text('I\'ve Verified My Email'),
                  style: OutlinedButton.styleFrom(
                    foregroundColor: AppTheme.primaryGreen,
                    side: const BorderSide(color: AppTheme.primaryGreen),
                    minimumSize: const Size(double.infinity, 50),
                  ),
                ),
                const SizedBox(height: 16),

                // Resend email button
                ElevatedButton.icon(
                  onPressed: _resendCooldownSeconds > 0 || _isResending
                      ? null
                      : _resendVerificationEmail,
                  icon: _isResending
                      ? const SizedBox(
                          height: 16,
                          width: 16,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                          ),
                        )
                      : const Icon(Icons.send),
                  label: Text(
                    _resendCooldownSeconds > 0
                        ? 'Resend Email ($_resendCooldownSeconds s)'
                        : 'Resend Verification Email',
                  ),
                ),
                const SizedBox(height: 32),

                // Divider
                Divider(color: Colors.grey[400]),
                const SizedBox(height: 16),

                // Back to login
                TextButton(
                  onPressed: _signOut,
                  child: const Text('Sign Out'),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
