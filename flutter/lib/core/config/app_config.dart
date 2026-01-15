import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../features/settings/models/app_settings.dart';

/// App 配置 Provider
class AppConfigNotifier extends StateNotifier<AppSettings> {
  AppConfigNotifier() : super(AppSettings());

  void updateThemeMode(ThemeMode mode) {
    state = state.copyWith(themeMode: mode);
  }

  void updateApiBaseUrl(String url) {
    state = state.copyWith(apiBaseUrl: url);
  }

  void updateAiProvider(String provider) {
    state = state.copyWith(aiProvider: provider);
  }

  void updateApiKeys(Map<String, String> keys) {
    state = state.copyWith(apiKeys: keys);
  }
}

/// App 配置 Provider
final appConfigProvider = StateNotifierProvider<AppConfigNotifier, AppSettings>((ref) {
  return AppConfigNotifier();
});
