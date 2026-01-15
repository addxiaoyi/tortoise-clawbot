// lib/core/theme/colors.dart

import 'package:flutter/material.dart';

class AppColors {
  AppColors._();

  // Primary colors
  static const Color primary = Color(0xFF6366F1);
  static const Color secondary = Color(0xFF8B5CF6);
  static const Color tertiary = Color(0xFF14B8A6);

  // Semantic colors
  static const Color success = Color(0xFF22C55E);
  static const Color warning = Color(0xFFF59E0B);
  static const Color error = Color(0xFFEF4444);
  static const Color info = Color(0xFF3B82F6);

  // Light theme
  static const Color background = Color(0xFFF8FAFC);
  static const Color surface = Color(0xFFFFFFFF);
  static const Color surfaceVariant = Color(0xFFF1F5F9);
  static const Color onSurface = Color(0xFF1E293B);
  static const Color onSurfaceVariant = Color(0xFF64748B);

  // Dark theme
  static const Color darkBackground = Color(0xFF0F172A);
  static const Color darkSurface = Color(0xFF1E293B);
  static const Color darkSurfaceVariant = Color(0xFF334155);
  static const Color darkOnSurface = Color(0xFFF8FAFC);
  static const Color darkOnSurfaceVariant = Color(0xFF94A3B8);

  // Message bubbles
  static const Color userBubble = Color(0xFF6366F1);
  static const Color userBubbleDark = Color(0xFF4F46E5);
  static const Color assistantBubble = Color(0xFFE2E8F0);
  static const Color assistantBubbleDark = Color(0xFF334155);

  // Agent role colors
  static const Color orchestratorColor = Color(0xFF8B5CF6);
  static const Color specialistColor = Color(0xFF14B8A6);
  static const Color researcherColor = Color(0xFFF59E0B);
  static const Color coderColor = Color(0xFF22C55E);
  static const Color criticColor = Color(0xFFEF4444);
}
