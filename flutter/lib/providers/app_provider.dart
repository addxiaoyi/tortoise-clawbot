import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../core/config/config_manager.dart';
import '../core/storage/storage_service.dart';

/// 应用状态
class AppState {
  final ThemeMode themeMode;
  final Locale locale;
  final bool isInitialized;
  final String? apiUrl;
  final String? currentModel;

  AppState({
    this.themeMode = ThemeMode.system,
    this.locale = const Locale('zh', 'CN'),
    this.isInitialized = false,
    this.apiUrl,
    this.currentModel = 'gpt-4',
  });

  AppState copyWith({
    ThemeMode? themeMode,
    Locale? locale,
    bool? isInitialized,
    String? apiUrl,
    String? currentModel,
  }) {
    return AppState(
      themeMode: themeMode ?? this.themeMode,
      locale: locale ?? this.locale,
      isInitialized: isInitialized ?? this.isInitialized,
      apiUrl: apiUrl ?? this.apiUrl,
      currentModel: currentModel ?? this.currentModel,
    );
  }
}

/// 应用状态管理器
class AppNotifier extends StateNotifier<AppState> {
  AppNotifier() : super(AppState()) {
    _initialize();
  }

  final _config = ConfigManager.instance;
  final _storage = StorageService.instance;

  void _initialize() {
    state = AppState(
      themeMode: _config.themeMode,
      locale: _config.locale,
      isInitialized: true,
      apiUrl: _config.apiUrl,
      currentModel: _config.defaultModel,
    );
  }

  Future<void> setThemeMode(ThemeMode mode) async {
    await _config.setThemeMode(mode);
    state = state.copyWith(themeMode: mode);
  }

  Future<void> setLocale(Locale locale) async {
    await _config.setLocale(locale);
    state = state.copyWith(locale: locale);
  }

  Future<void> setApiUrl(String url) async {
    await _config.setApiUrl(url);
    state = state.copyWith(apiUrl: url);
  }

  Future<void> setModel(String model) async {
    await _config.setDefaultModel(model);
    state = state.copyWith(currentModel: model);
  }
}

/// Provider
final appProvider = StateNotifierProvider<AppNotifier, AppState>((ref) {
  return AppNotifier();
});
