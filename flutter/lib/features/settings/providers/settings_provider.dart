import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/app_settings.dart';

/// 设置Provider
class SettingsNotifier extends StateNotifier<AppSettings> {
  SettingsNotifier() : super(AppSettings());

  void updateThemeMode(ThemeMode mode) {
    state = state.copyWith(themeMode: mode);
  }

  void updateLocale(String locale) {
    state = state.copyWith(locale: locale);
  }

  void updateDefaultModel(String model) {
    state = state.copyWith(defaultModel: model);
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

/// 设置Provider
final settingsProvider = StateNotifierProvider<SettingsNotifier, AppSettings>((ref) {
  return SettingsNotifier();
});

/// 主题模式Provider
final themeModeProvider = Provider<ThemeMode>((ref) {
  return ref.watch(settingsProvider).themeMode;
});

/// 语言Provider
final localeProvider = Provider<String>((ref) {
  return ref.watch(settingsProvider).locale;
});

/// 默认模型Provider
final defaultModelProvider = Provider<String>((ref) {
  return ref.watch(settingsProvider).defaultModel;
});
