import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

// Theme providers
final themeModeProvider = StateProvider<ThemeMode>((ref) => ThemeMode.dark);

// Connection state providers
final connectionStateProvider = StateProvider<ConnectionState>((ref) => ConnectionState.disconnected);

enum ConnectionState {
  connected,
  connecting,
  disconnected,
  error,
}

// Settings providers
final settingsProvider = StateNotifierProvider<SettingsNotifier, SettingsState>((ref) {
  return SettingsNotifier();
});

class SettingsState {
  final String apiEndpoint;
  final String apiKey;
  final String defaultModel;
  final double temperature;
  final int maxTokens;
  
  const SettingsState({
    this.apiEndpoint = 'http://localhost:18792',
    this.apiKey = '',
    this.defaultModel = 'gpt-4o',
    this.temperature = 0.7,
    this.maxTokens = 4096,
  });
  
  SettingsState copyWith({
    String? apiEndpoint,
    String? apiKey,
    String? defaultModel,
    double? temperature,
    int? maxTokens,
  }) {
    return SettingsState(
      apiEndpoint: apiEndpoint ?? this.apiEndpoint,
      apiKey: apiKey ?? this.apiKey,
      defaultModel: defaultModel ?? this.defaultModel,
      temperature: temperature ?? this.temperature,
      maxTokens: maxTokens ?? this.maxTokens,
    );
  }
}

class SettingsNotifier extends StateNotifier<SettingsState> {
  SettingsNotifier() : super(const SettingsState());
  
  void updateEndpoint(String endpoint) {
    state = state.copyWith(apiEndpoint: endpoint);
  }
  
  void updateApiKey(String apiKey) {
    state = state.copyWith(apiKey: apiKey);
  }
  
  void updateModel(String model) {
    state = state.copyWith(defaultModel: model);
  }
  
  void updateTemperature(double temp) {
    state = state.copyWith(temperature: temp);
  }
  
  void updateMaxTokens(int tokens) {
    state = state.copyWith(maxTokens: tokens);
  }
}

// Sessions provider
final sessionsProvider = StateNotifierProvider<SessionsNotifier, List<Session>>((ref) {
  return SessionsNotifier();
});

class Session {
  final String id;
  final String name;
  final DateTime createdAt;
  final int messageCount;
  final bool isActive;
  
  const Session({
    required this.id,
    required this.name,
    required this.createdAt,
    this.messageCount = 0,
    this.isActive = true,
  });
}

class SessionsNotifier extends StateNotifier<List<Session>> {
  SessionsNotifier() : super([]);
  
  void addSession(Session session) {
    state = [...state, session];
  }
  
  void removeSession(String id) {
    state = state.where((s) => s.id != id).toList();
  }
  
  void updateSession(Session session) {
    state = state.map((s) => s.id == session.id ? session : s).toList();
  }
}

// Active session provider
final activeSessionProvider = StateProvider<String?>((ref) => null);
