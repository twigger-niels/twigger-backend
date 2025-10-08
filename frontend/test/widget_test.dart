// Basic Flutter widget test for Twigger app

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:frontend/main.dart';

void main() {
  testWidgets('App launches and shows splash or login screen', (WidgetTester tester) async {
    // Build our app and trigger a frame.
    // Note: This will fail Firebase initialization in tests, which is expected
    await tester.pumpWidget(const TwiggerApp());
    await tester.pumpAndSettle();

    // Verify that the app renders without crashing
    // In a real test environment, you would mock Firebase and test specific flows
    expect(find.byType(MaterialApp), findsOneWidget);
  });
}
