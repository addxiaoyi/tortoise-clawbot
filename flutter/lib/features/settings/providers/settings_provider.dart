import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/app_settings.dart';

/// 设置状态
class SettingsState {
  final AppSettings settings;
  final bool isLoading;
  final String? error;

  const SettingsState({
    this.settings = const AppSettings(),
    this.isLoading = false,
    this.error,
  });

  SettingsState copyWith({
    AppSettings? settings,
    bool? isLoading,
    String? error,
  }) {
    return SettingsState(
      settings: settings ?? this.settings,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }
}

/// 设置管理
class SettingsNotifier extends StateNotifier<SettingsState> {
  SettingsNotifier() : super(const SettingsState());

  /// 更新 API 配置
  void updateApiConfig({
    String? apiKey,
    String? baseUrl,
    String? model,
  }) {
    final current = state.settings;
    state = state.copyWith(
      settings: current.copyWith(
        openaiKey: apiKey ?? current.openaiKey,
        apiBaseUrl: baseUrl ?? current.apiBaseUrl,
        defaultModel: model ?? current.defaultModel,
      ),
    );
  }

  /// 更新主题模式
  void updateThemeMode(String mode) {
    final current = state.settings;
    state = state.copyWith(
      settings: current.copyWith(themeMode: mode),
    );
  }

  /// 更新语言
  void updateLanguage(String language) {
    final current = state.settings;
    state = state.copyWith(
      settings: current.copyWith(language: language),
    );
  }

  /// 更新通知设置
  void updateNotifications({
    bool? pushEnabled,
    bool? soundEnabled,
    bool? emailEnabled,
  }) {
    final current = state.settings;
    state = state.copyWith(
      settings: current.copyWith(
        pushEnabled: pushEnabled ?? current.pushEnabled,
        soundEnabled: soundEnabled ?? current.soundEnabled,
        emailNotifications: emailEnabled ?? current.emailNotifications,
      ),
    );
  }

  /// 更新隐私设置
  void updatePrivacy({
    bool? analytics,
    bool? crashReporting,
  }) {
    final current = state.settings;
    state = state.copyWith(
      settings: current.copyWith(
        analyticsEnabled: analytics ?? current.analyticsEnabled,
        crashReporting: crashReporting ?? current.crashReporting,
      ),
    );
  }

  /// 重置设置
  void resetSettings() {
    state = const SettingsState();
  }

  /// 加载设置
  Future<void> loadSettings() async {
    state = state.copyWith(isLoading: true);
    // TODO: 从存储加载
    await Future.delayed(const Duration(milliseconds: 300));
    state = state.copyWith(isLoading: false);
  }

  /// 保存设置
  Future<void> saveSettings() async {
    state = state.copyWith(isLoading: true);
    // TODO: 保存到存储
    await Future.delayed(const Duration(milliseconds: 300));
    state = state.copyWith(isLoading: false);
  }
}

/// 设置 Provider
final settingsProvider = StateNotifierProvider<SettingsNotifier, SettingsState>((ref) {
  return SettingsNotifier();
});

/// 当前主题模式
final themeModeProvider = Provider<String>((ref) {
  return ref.watch(settingsProvider).settings.themeMode;
});

/// 当前语言
final languageProvider = Provider<String>((ref) {
  return ref.watch(settingsProvider).settings.language;
});

/// API 配置
final apiConfigProvider = Provider<Map<String, String>>((ref) {
  final settings = ref.watch(settingsProvider).settings;
  return {
    'apiKey': settings.openaiKey,
    'baseUrl': settings.apiBaseUrl,
    'model': settings.defaultModel,
  };
});
