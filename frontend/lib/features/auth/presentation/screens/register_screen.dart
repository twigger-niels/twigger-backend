import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/auth_provider.dart' as auth;
import '../../../../shared/widgets/custom_text_field.dart';
import '../../../../shared/utils/validators.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../../core/constants/app_constants.dart';

class RegisterScreen extends StatefulWidget {
  const RegisterScreen({super.key});

  @override
  State<RegisterScreen> createState() => _RegisterScreenState();
}

class _RegisterScreenState extends State<RegisterScreen> {
  final _formKey = GlobalKey<FormState>();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  final _confirmPasswordController = TextEditingController();
  bool _agreedToTerms = false;
  int _passwordStrength = 0;

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    _confirmPasswordController.dispose();
    super.dispose();
  }

  void _updatePasswordStrength(String password) {
    setState(() {
      _passwordStrength = Validators.getPasswordStrength(password);
    });
  }

  Future<void> _handleRegister() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    if (!_agreedToTerms) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Please agree to the terms and conditions'),
          backgroundColor: AppTheme.errorRed,
        ),
      );
      return;
    }

    final authProvider = context.read<auth.AuthProvider>();

    try {
      await authProvider.registerWithEmail(
        email: _emailController.text.trim(),
        password: _passwordController.text,
      );

      // Registration successful - pop back to let AuthWrapper show EmailVerificationScreen
      if (mounted) {
        // Use popUntil to ensure we go back to the root (AuthWrapper)
        Navigator.of(context).popUntil((route) => route.isFirst);
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
        title: const Text('Create Account'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24.0),
          child: Consumer<auth.AuthProvider>(
            builder: (context, authProvider, _) {
              return Form(
                key: _formKey,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    // Welcome message
                    Text(
                      'Join Twigger',
                      style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                            fontWeight: FontWeight.bold,
                            color: AppTheme.primaryGreen,
                          ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      'Create your account to start planning your garden',
                      style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                            color: Colors.grey[600],
                          ),
                    ),
                    const SizedBox(height: 32),

                    // Email field
                    EmailTextField(
                      controller: _emailController,
                      validator: Validators.validateEmail,
                    ),
                    const SizedBox(height: 16),

                    // Password field
                    PasswordTextField(
                      controller: _passwordController,
                      validator: Validators.validatePassword,
                    ),
                    const SizedBox(height: 8),

                    // Password strength indicator
                    if (_passwordController.text.isNotEmpty) ...[
                      Row(
                        children: [
                          Expanded(
                            child: LinearProgressIndicator(
                              value: _passwordStrength / 4,
                              backgroundColor: Colors.grey[300],
                              valueColor: AlwaysStoppedAnimation<Color>(
                                _passwordStrength <= 1
                                    ? AppTheme.errorRed
                                    : _passwordStrength == 2
                                        ? AppTheme.warningOrange
                                        : _passwordStrength == 3
                                            ? Colors.yellow[700]!
                                            : AppTheme.successGreen,
                              ),
                            ),
                          ),
                          const SizedBox(width: 12),
                          Text(
                            Validators.getPasswordStrengthLabel(_passwordStrength),
                            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                  color: _passwordStrength <= 1
                                      ? AppTheme.errorRed
                                      : _passwordStrength == 2
                                          ? AppTheme.warningOrange
                                          : _passwordStrength == 3
                                              ? Colors.yellow[700]!
                                              : AppTheme.successGreen,
                                  fontWeight: FontWeight.w600,
                                ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 16),
                    ],

                    // Confirm password field
                    CustomTextField(
                      controller: _confirmPasswordController,
                      label: 'Confirm Password',
                      hintText: 'Re-enter your password',
                      obscureText: true,
                      prefixIcon: const Icon(Icons.lock_outline),
                      validator: (value) => Validators.validatePasswordConfirmation(
                        value,
                        _passwordController.text,
                      ),
                    ),
                    const SizedBox(height: 16),

                    // Terms and conditions checkbox
                    Row(
                      children: [
                        Checkbox(
                          value: _agreedToTerms,
                          onChanged: (value) {
                            setState(() {
                              _agreedToTerms = value ?? false;
                            });
                          },
                          activeColor: AppTheme.primaryGreen,
                        ),
                        Expanded(
                          child: GestureDetector(
                            onTap: () {
                              setState(() {
                                _agreedToTerms = !_agreedToTerms;
                              });
                            },
                            child: Text(
                              'I agree to the Terms and Conditions',
                              style: Theme.of(context).textTheme.bodyMedium,
                            ),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 24),

                    // Register button
                    ElevatedButton(
                      onPressed: authProvider.isLoading ? null : _handleRegister,
                      child: authProvider.isLoading
                          ? const SizedBox(
                              height: 20,
                              width: 20,
                              child: CircularProgressIndicator(
                                strokeWidth: 2,
                                valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                              ),
                            )
                          : const Text('Create Account'),
                    ),
                    const SizedBox(height: 24),

                    // Sign in link
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Text(
                          'Already have an account? ',
                          style: TextStyle(color: Colors.grey[600]),
                        ),
                        TextButton(
                          onPressed: () => Navigator.pop(context),
                          child: const Text('Sign In'),
                        ),
                      ],
                    ),
                  ],
                ),
              );
            },
          ),
        ),
      ),
    );
  }
}
