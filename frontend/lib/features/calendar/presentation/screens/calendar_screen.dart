import 'package:flutter/material.dart';
import '../../../../core/theme/app_theme.dart';

class CalendarScreen extends StatelessWidget {
  const CalendarScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Calendar'),
      ),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(24.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.calendar_today,
                size: 100,
                color: AppTheme.primaryGreen.withOpacity(0.5),
              ),
              const SizedBox(height: 24),
              Text(
                'Garden Calendar',
                style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                      color: AppTheme.primaryGreen,
                    ),
              ),
              const SizedBox(height: 16),
              Text(
                'Track your garden tasks and events here',
                style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                      color: Colors.grey[600],
                    ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 32),
              Text(
                'Coming Soon!',
                style: Theme.of(context).textTheme.titleLarge?.copyWith(
                      color: AppTheme.secondaryGreen,
                      fontWeight: FontWeight.w600,
                    ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
