import 'dart:async';
import 'package:dio/dio.dart';

/// AI Provider 接口
abstract class AIProvider {
  Future<ChatCompletionResponse> chat({
    required String model,
    required List<ChatMessage> messages,
    double? temperature,
    int? maxTokens,
  });
}

/// Chat 消息
class ChatMessage {
  final String role;
  final String content;
  ChatMessage({required this.role, required this.content});
  Map<String, dynamic> toJson() => {'role': role, 'content': content};
}

/// Chat 完成响应
class ChatCompletionResponse {
  final String id;
  final String model;
  final List<Choice> choices;
  ChatCompletionResponse({required this.id, required this.model, required this.choices});
}

/// Choice
class Choice {
  final ChatMessage message;
  final String? finishReason;
  Choice({required this.message, this.finishReason});
}

/// AI 响应
class AIResponse {
  final String content;
  final String model;
  final String provider;
  AIResponse({required this.content, required this.model, required this.provider});
}

/// AI 事件
class AIEvent {
  final AIEventType type;
  final String? provider;
  final String? model;
  AIEvent({required this.type, this.provider, this.model});
}

enum AIEventType { initialized, request, response, error }

/// AI 引擎异常
class AIException implements Exception {
  final String message;
  AIException(this.message);
  @override
  String toString() => message;
}

/// AI 引擎核心
class AIEngine {
  final Map<String, AIProvider> _providers = {};
  String _defaultProvider = 'openai';
  String _defaultModel = 'gpt-4';
  final _eventController = StreamController<AIEvent>.broadcast();

  Stream<AIEvent> get events => _eventController.stream;

  Future<void> initialize({required Map<String, String> apiKeys, String? defaultProvider, String? defaultModel}) async {
    if (apiKeys.containsKey('openai')) {
      _providers['openai'] = OpenAIProvider(apiKey: apiKeys['openai']!);
    }
    if (apiKeys.containsKey('anthropic')) {
      _providers['anthropic'] = AnthropicProvider(apiKey: apiKeys['anthropic']!);
    }
    if (_providers.isEmpty) {
      throw AIException('至少需要一个 AI API Key');
    }
    _defaultProvider = defaultProvider ?? _providers.keys.first;
    _defaultModel = defaultModel ?? 'gpt-4';
    _eventController.add(AIEvent(type: AIEventType.initialized, provider: _defaultProvider, model: _defaultModel));
  }

  Future<AIResponse> chat({required String prompt, List<ChatMessage>? history, String? provider, String? model}) async {
    final effectiveProvider = provider ?? _defaultProvider;
    final effectiveModel = model ?? _defaultModel;
    final aiProvider = _providers[effectiveProvider];
    if (aiProvider == null) {
      throw AIException('未找到 AI Provider: $effectiveProvider');
    }
    final messages = <ChatMessage>[];
    if (history != null) messages.addAll(history);
    messages.add(ChatMessage(role: 'user', content: prompt));
    try {
      final response = await aiProvider.chat(model: effectiveModel, messages: messages);
      return AIResponse(content: response.choices.first.message.content, model: effectiveModel, provider: effectiveProvider);
    } catch (e) {
      throw AIException(e.toString());
    }
  }

  void dispose() {
    _eventController.close();
  }
}

/// OpenAI Provider
class OpenAIProvider implements AIProvider {
  final Dio _dio;
  final String apiKey;

  OpenAIProvider({required this.apiKey}) : _dio = Dio(BaseOptions(
    baseUrl: 'https://api.openai.com/v1',
    headers: {'Authorization': 'Bearer $apiKey'},
  ));

  @override
  Future<ChatCompletionResponse> chat({required String model, required List<ChatMessage> messages, double? temperature, int? maxTokens}) async {
    final response = await _dio.post('/chat/completions', data: {
      'model': model,
      'messages': messages.map((m) => m.toJson()).toList(),
      if (temperature != null) 'temperature': temperature,
      if (maxTokens != null) 'max_tokens': maxTokens,
    });
    final data = response.data as Map<String, dynamic>;
    final choices = (data['choices'] as List<dynamic>).map((c) {
      final choiceData = c as Map<String, dynamic>;
      final messageData = choiceData['message'] as Map<String, dynamic>;
      return Choice(
        message: ChatMessage(
          role: messageData['role'] as String,
          content: messageData['content'] as String,
        ),
        finishReason: choiceData['finish_reason'] as String?,
      );
    }).toList();
    return ChatCompletionResponse(
      id: data['id'] as String,
      model: model,
      choices: choices,
    );
  }
}

/// Anthropic Provider
class AnthropicProvider implements AIProvider {
  final Dio _dio;
  AnthropicProvider({required String apiKey}) : _dio = Dio(BaseOptions(
    baseUrl: 'https://api.anthropic.com/v1',
    headers: {'x-api-key': apiKey, 'anthropic-version': '2023-06-01', 'content-type': 'application/json'},
  ));

  @override
  Future<ChatCompletionResponse> chat({required String model, required List<ChatMessage> messages, double? temperature, int? maxTokens}) async {
    final systemMsg = messages.where((m) => m.role == 'system').toList();
    final filteredMessages = messages.where((m) => m.role != 'system').toList();
    final requestData = <String, dynamic>{
      'model': model,
      'messages': filteredMessages.map((m) => {'role': m.role, 'content': m.content}).toList(),
      'max_tokens': maxTokens ?? 4096,
    };
    if (systemMsg.isNotEmpty) requestData['system'] = systemMsg.first.content;
    if (temperature != null) requestData['temperature'] = temperature;
    final response = await _dio.post('/messages', data: requestData);
    final data = response.data as Map<String, dynamic>;
    String text = '';
    final content = data['content'] as List<dynamic>;
    for (final block in content) {
      final blockData = block as Map<String, dynamic>;
      if (blockData['type'] == 'text') {
        text = (blockData['text'] as String?) ?? '';
        break;
      }
    }
    final id = (data['id'] as String?) ?? DateTime.now().millisecondsSinceEpoch.toString();
    return ChatCompletionResponse(
      id: id,
      model: model,
      choices: [Choice(message: ChatMessage(role: 'assistant', content: text))],
    );
  }
}
