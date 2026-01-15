import 'package:flutter/material.dart';

/// 应用设置模型
class AppSettings {
  final ThemeMode themeMode;
  final String locale;
  final String defaultModel;
  final String? apiBaseUrl;
  final String? aiProvider;
  final Map<String, String>? apiKeys;
  final bool autoSpeechToText;
  final bool autoWakeWord;

  AppSettings({
    this.themeMode = ThemeMode.system,
    this.locale = 'zh-CN',
    this.defaultModel = 'gpt-4',
    this.apiBaseUrl,
    this.aiProvider,
    this.apiKeys,
    this.autoSpeechToText = false,
    this.autoWakeWord = false,
  });

  AppSettings copyWith({
    ThemeMode? themeMode,
    String? locale,
    String? defaultModel,
    String? apiBaseUrl,
    String? aiProvider,
    Map<String, String>? apiKeys,
    bool? autoSpeechToText,
    bool? autoWakeWord,
  }) {
    return AppSettings(
      themeMode: themeMode ?? this.themeMode,
      locale: locale ?? this.locale,
      defaultModel: defaultModel ?? this.defaultModel,
      apiBaseUrl: apiBaseUrl ?? this.apiBaseUrl,
      aiProvider: aiProvider ?? this.aiProvider,
      apiKeys: apiKeys ?? this.apiKeys,
      autoSpeechToText: autoSpeechToText ?? this.autoSpeechToText,
      autoWakeWord: autoWakeWord ?? this.autoWakeWord,
    );
  }
}
