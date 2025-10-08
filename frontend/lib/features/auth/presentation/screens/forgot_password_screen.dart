import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/auth_provider.dart';
import '../../../../shared/widgets/custom_text_field.dart';
import '../../../../shared/utils/validators.dart';
import '../../../../core/theme/app_theme.dart';

class ForgotPasswordScreen extends StatefulWidget {
  const ForgotPasswordScreen({super.key});

  @override
  State<ForgotPasswordScreen> createState() => _ForgotPasswordScreenState();
}

class _ForgotPasswordScreenState extends State<ForgotPasswordScreen> {
  final _formKey = GlobalKey<FormState>();
  final _emailController = TextEditingController();
  bool _emailSent = false;

  @override
  void dispose() {
    _emailController.dispose();
    super.dispose();
  }

  Future<void> _handleSendResetEmail() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    final authProvider = context.read<AuthProvider>();

    try {
      await authProvider.resetPassword(_emailController.text.trim());

      if (mounted) {
        setState(() {
          _emailSent = true;
        });
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
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Reset Password'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24.0),
            child: Consumer<AuthProvider>(
              builder: (context, authProvider, _) {
                if (_emailSent) {
                  return _buildSuccessView();
                }

                return _buildFormView(authProvider);
              },
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildFormView(AuthProvider authProvider) {
    return Form(
      key: _formKey,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          // Icon
          Icon(
            Icons.lock_reset,
            size: 80,
            color: AppTheme.primaryGreen,
          ),
          const SizedBox(height: 24),

          // Title
          Text(
            'Forgot Password?',
            style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                  color: AppTheme.primaryGreen,
                ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 16),

          // Instructions
          Text(
            'Enter your email address and we\'ll send you a link to reset your password.',
            style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                  color: Colors.grey[600],
                ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 32),

          // Email field
          EmailTextField(
            controller: _emailController,
            validator: Validators.validateEmail,
          ),
          const SizedBox(height: 24),

          // Send reset link button
          ElevatedButton(
            onPressed: authProvider.isLoading ? null : _handleSendResetEmail,
            child: authProvider.isLoading
                ? const SizedBox(
                    height: 20,
                    width: 20,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                    ),
                  )
                : const Text('Send Reset Link'),
          ),
          const SizedBox(height: 16),

          // Back to login
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Back to Sign In'),
          ),
        ],
      ),
    );
  }

  Widget _buildSuccessView() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        // Success icon
        Container(
          padding: const EdgeInsets.all(24),
          decoration: BoxDecoration(
            color: AppTheme.successGreen.withOpacity(0.1),
            shape: BoxShape.circle,
          ),
          child: Icon(
            Icons.check_circle_outline,
            size: 80,
            color: AppTheme.successGreen,
          ),
        ),
        const SizedBox(height: 24),

        // Success message
        Text(
          'Email Sent!',
          style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                fontWeight: FontWeight.bold,
                color: AppTheme.successGreen,
              ),
          textAlign: TextAlign.center,
        ),
        const SizedBox(height: 16),

        // Instructions
        Text(
          'We\'ve sent a password reset link to:',
          style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                color: Colors.grey[700],
              ),
          textAlign: TextAlign.center,
        ),
        const SizedBox(height: 8),

        // Email address
        Text(
          _emailController.text.trim(),
          style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                fontWeight: FontWeight.w600,
                color: AppTheme.primaryGreen,
              ),
          textAlign: TextAlign.center,
        ),
        const SizedBox(height: 24),

        // Additional instructions
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Colors.blue[50],
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: Colors.blue[200]!),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Icon(Icons.info_outline, color: Colors.blue[700]),
                  const SizedBox(width: 8),
                  Text(
                    'What\'s next?',
                    style: Theme.of(context).textTheme.titleMedium?.copyWith(
                          fontWeight: FontWeight.w600,
                          color: Colors.blue[700],
                        ),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              _buildStep('Check your email inbox'),
              _buildStep('Click the reset link in the email'),
              _buildStep('Enter your new password'),
              _buildStep('Sign in with your new password'),
            ],
          ),
        ),
        const SizedBox(height: 32),

        // Back to login button
        ElevatedButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('Back to Sign In'),
        ),
        const SizedBox(height: 16),

        // Didn't receive email
        Text(
          'Didn\'t receive the email? Check your spam folder.',
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: Colors.grey[600],
              ),
          textAlign: TextAlign.center,
        ),
      ],
    );
  }

  Widget _buildStep(String text) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        children: [
          Container(
            width: 6,
            height: 6,
            decoration: BoxDecoration(
              color: Colors.blue[700],
              shape: BoxShape.circle,
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              text,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Colors.grey[800],
                  ),
            ),
          ),
        ],
      ),
    );
  }
}
