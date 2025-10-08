import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:firebase_core/firebase_core.dart';
import 'package:flutter_facebook_auth/flutter_facebook_auth.dart';
import 'package:provider/provider.dart';
import 'firebase_options.dart';
import 'core/theme/app_theme.dart';
import 'features/auth/presentation/providers/auth_provider.dart';
import 'features/auth/services/auth_service.dart';
import 'features/auth/data/auth_api_client.dart';
import 'shared/widgets/auth_wrapper.dart';

/// Main entry point for the Twigger app
///
/// Initializes Firebase and sets up Provider for state management
void main() async {
  // Ensure Flutter is initialized before Firebase
  WidgetsFlutterBinding.ensureInitialized();

  // Initialize Firebase with platform-specific options
  try {
    await Firebase.initializeApp(
      options: DefaultFirebaseOptions.currentPlatform,
    );
  } catch (e) {
    // Firebase initialization failed
    debugPrint('Firebase initialization failed: $e');
  }

  // Initialize Facebook Auth for web
  if (kIsWeb) {
    await FacebookAuth.i.webAndDesktopInitialize(
      appId: '763251526584065',
      cookie: true,
      xfbml: true,
      version: 'v18.0',
    );
  }

  runApp(const TwiggerApp());
}

class TwiggerApp extends StatelessWidget {
  const TwiggerApp({super.key});

  @override
  Widget build(BuildContext context) {
    // Setup dependency injection and state management
    return MultiProvider(
      providers: [
        // Auth API Client (singleton)
        Provider<AuthApiClient>(
          create: (_) => AuthApiClient(),
        ),

        // Auth Service (singleton)
        Provider<AuthService>(
          create: (context) => AuthService(
            apiClient: context.read<AuthApiClient>(),
          ),
        ),

        // Auth Provider (state management)
        ChangeNotifierProvider<AuthProvider>(
          create: (context) => AuthProvider(
            authService: context.read<AuthService>(),
          ),
        ),
      ],
      child: MaterialApp(
        title: 'Twigger',
        debugShowCheckedModeBanner: false,

        // Theme configuration
        theme: AppTheme.lightTheme,
        darkTheme: AppTheme.darkTheme,
        themeMode: ThemeMode.system,

        // Home screen with authentication routing
        home: const AuthWrapper(),
      ),
    );
  }
}
