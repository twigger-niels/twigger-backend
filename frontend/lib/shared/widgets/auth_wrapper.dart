import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../core/constants/app_constants.dart';
import '../../features/auth/presentation/providers/auth_provider.dart' as auth;
import '../../features/auth/presentation/screens/login_screen.dart';
import '../../features/auth/presentation/screens/email_verification_screen.dart';
import '../../features/auth/presentation/screens/splash_screen.dart';
import 'main_navigation_shell.dart';

/// AuthWrapper determines which screen to show based on authentication state
///
/// Routes:
/// - initial/authenticating -> SplashScreen
/// - authenticated -> MainNavigationShell (home with bottom nav)
/// - unauthenticated/error -> LoginScreen
/// - emailVerificationPending -> EmailVerificationScreen
class AuthWrapper extends StatefulWidget {
  const AuthWrapper({super.key});

  @override
  State<AuthWrapper> createState() => _AuthWrapperState();
}

class _AuthWrapperState extends State<AuthWrapper> {
  @override
  void initState() {
    super.initState();
    // Check auth state when app starts
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<auth.AuthProvider>().checkAuthState();
    });
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<auth.AuthProvider>(
      builder: (context, authProvider, child) {
        // Show splash screen during initial loading or authentication
        if (authProvider.state == AuthState.initial ||
            authProvider.state == AuthState.authenticating) {
          return const SplashScreen();
        }

        // User is authenticated and email is verified (or doesn't need verification)
        if (authProvider.state == AuthState.authenticated) {
          return const MainNavigationShell();
        }

        // User signed up but hasn't verified email yet
        if (authProvider.state == AuthState.emailVerificationPending) {
          return const EmailVerificationScreen();
        }

        // User is not authenticated or there was an error
        // Default to login screen
        return const LoginScreen();
      },
    );
  }
}
