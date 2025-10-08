import 'package:flutter/material.dart';
import '../../../../core/theme/app_theme.dart';

class AddScreen extends StatelessWidget {
  const AddScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Add'),
      ),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(24.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.add_circle,
                size: 100,
                color: AppTheme.primaryGreen.withOpacity(0.5),
              ),
              const SizedBox(height: 24),
              Text(
                'Add Plants',
                style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                      color: AppTheme.primaryGreen,
                    ),
              ),
              const SizedBox(height: 16),
              Text(
                'Add new plants to your garden',
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
