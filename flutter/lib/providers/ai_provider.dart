import 'package:flutter/material.dart';
import '../core/ai/ai_engine.dart';
import '../core/storage/storage_service.dart';

/// AI Provider 状态管理
class AIState {
  final bool isInitialized;
  final bool isLoading;
  final String? error;
  final String currentModel;
  final String? lastResponse;

  AIState({
    this.isInitialized = false,
    this.isLoading = false,
    this.error,
    this.currentModel = 'gpt-4',
    this.lastResponse,
  });

  AIState copyWith({
    bool? isInitialized,
    bool? isLoading,
    String? error,
    String? currentModel,
    String? lastResponse,
  }) {
    return AIState(
      isInitialized: isInitialized ?? this.isInitialized,
      isLoading: isLoading ?? this.isLoading,
      error: error,
      currentModel: currentModel ?? this.currentModel,
      lastResponse: lastResponse ?? this.lastResponse,
    );
  }
}

/// AI Provider Notifier
class AINotifier extends ChangeNotifier {
  final AIEngine _engine = AIEngine();
  final StorageService _storage = StorageService.instance;
  AIState _state = AIState();

  AIState get state => _state;
  AIEngine get engine => _engine;

  AINotifier() {
    _loadConfig();
  }

  void _loadConfig() {
    final model = _storage.getString('ai_model') ?? 'gpt-4';
    _state = _state.copyWith(currentModel: model);
  }

  Future<void> initialize({required String provider, required String apiKey}) async {
    try {
      _state = _state.copyWith(isLoading: true, error: null);
      notifyListeners();

      await _engine.initialize(apiKeys: {provider: apiKey});

      _state = _state.copyWith(isInitialized: true, isLoading: false);
      notifyListeners();
    } catch (e) {
      _state = _state.copyWith(isLoading: false, error: e.toString());
      notifyListeners();
    }
  }

  Future<String?> sendMessage(String message) async {
    if (!_state.isInitialized) {
      return '请先配置 API Key';
    }

    try {
      _state = _state.copyWith(isLoading: true, error: null);
      notifyListeners();

      final response = await _engine.chat(prompt: message);

      _state = _state.copyWith(
        isLoading: false,
        lastResponse: response.content,
      );
      notifyListeners();

      return response.content;
    } catch (e) {
      _state = _state.copyWith(isLoading: false, error: e.toString());
      notifyListeners();
      return null;
    }
  }

  void setModel(String model) {
    _state = _state.copyWith(currentModel: model);
    _storage.setString('ai_model', model);
    notifyListeners();
  }

  void clearError() {
    _state = _state.copyWith(error: null);
    notifyListeners();
  }
}
