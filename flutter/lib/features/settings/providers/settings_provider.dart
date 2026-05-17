import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

// Theme mode provider
final themeModeProvider = StateProvider<ThemeMode>((ref) => ThemeMode.system);

// AI provider selection
final selectedAiProviderProvider = StateProvider<String>((ref) => 'openai');

// Settings state
class SettingsState {
  final ThemeMode themeMode;
  final String aiProvider;
  final String aiModel;
  final String apiKey;
  final String apiEndpoint;
  final bool voiceEnabled;
  final String wakeWord;
  final double wakeSensitivity;
  final bool notificationsEnabled;
  final String language;

  const SettingsState({
    this.themeMode = ThemeMode.system,
    this.aiProvider = 'openai',
    this.aiModel = 'gpt-4o',
    this.apiKey = '',
    this.apiEndpoint = 'https://api.openai.com/v1',
    this.voiceEnabled = false,
    this.wakeWord = 'Hey Tortoise',
    this.wakeSensitivity = 0.7,
    this.notificationsEnabled = true,
    this.language = 'zh-CN',
  });

  SettingsState copyWith({
    ThemeMode? themeMode,
    String? aiProvider,
    String? aiModel,
    String? apiKey,
    String? apiEndpoint,
    bool? voiceEnabled,
    String? wakeWord,
    double? wakeSensitivity,
    bool? notificationsEnabled,
    String? language,
  }) {
    return SettingsState(
      themeMode: themeMode ?? this.themeMode,
      aiProvider: aiProvider ?? this.aiProvider,
      aiModel: aiModel ?? this.aiModel,
      apiKey: apiKey ?? this.apiKey,
      apiEndpoint: apiEndpoint ?? this.apiEndpoint,
      voiceEnabled: voiceEnabled ?? this.voiceEnabled,
      wakeWord: wakeWord ?? this.wakeWord,
      wakeSensitivity: wakeSensitivity ?? this.wakeSensitivity,
      notificationsEnabled: notificationsEnabled ?? this.notificationsEnabled,
      language: language ?? this.language,
    );
  }
}

class SettingsNotifier extends StateNotifier<SettingsState> {
  SettingsNotifier() : super(const SettingsState());

  void updateThemeMode(ThemeMode mode) {
    state = state.copyWith(themeMode: mode);
  }

  void updateAiProvider(String provider) {
    state = state.copyWith(aiProvider: provider);
    // Set default model for provider
    switch (provider) {
      case 'openai':
        state = state.copyWith(aiModel: 'gpt-4o');
        state = state.copyWith(apiEndpoint: 'https://api.openai.com/v1');
        break;
      case 'anthropic':
        state = state.copyWith(aiModel: 'claude-3-5-sonnet');
        state = state.copyWith(apiEndpoint: 'https://api.anthropic.com/v1');
        break;
      case 'google':
        state = state.copyWith(aiModel: 'gemini-2.0-flash');
        state = state.copyWith(apiEndpoint: 'https://generativelanguage.googleapis.com/v1');
        break;
    }
  }

  void updateAiModel(String model) {
    state = state.copyWith(aiModel: model);
  }

  void updateApiKey(String key) {
    state = state.copyWith(apiKey: key);
  }

  void updateApiEndpoint(String endpoint) {
    state = state.copyWith(apiEndpoint: endpoint);
  }

  void updateVoiceEnabled(bool enabled) {
    state = state.copyWith(voiceEnabled: enabled);
  }

  void updateWakeWord(String word) {
    state = state.copyWith(wakeWord: word);
  }

  void updateWakeSensitivity(double sensitivity) {
    state = state.copyWith(wakeSensitivity: sensitivity);
  }

  void updateNotificationsEnabled(bool enabled) {
    state = state.copyWith(notificationsEnabled: enabled);
  }

  void updateLanguage(String language) {
    state = state.copyWith(language: language);
  }
}

final settingsProvider = StateNotifierProvider<SettingsNotifier, SettingsState>((ref) {
  return SettingsNotifier();
});
