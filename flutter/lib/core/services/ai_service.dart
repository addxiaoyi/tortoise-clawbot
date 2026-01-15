import 'package:dio/dio.dart';
import '../config/api_config.dart';

/// AI服务类 - 支持多种AI提供商
class AIService {
  final String provider;
  final String apiKey;
  final String baseUrl;
  late final Dio _dio;
  
  AIService({
    required this.provider,
    required this.apiKey,
    String? baseUrl,
  }) : baseUrl = baseUrl ?? AIProviders.baseUrls[provider] ?? ApiConfig.defaultApiUrl {
    _dio = Dio(
      BaseOptions(
        baseUrl: this.baseUrl,
        connectTimeout: ApiConfig.connectTimeout,
        receiveTimeout: ApiConfig.receiveTimeout,
        headers: _getHeaders(),
      ),
    );
  }
  
  Map<String, String> _getHeaders() {
    switch (provider) {
      case AIProviders.openai:
      case AIProviders.openrouter:
      case AIProviders.groq:
      case AIProviders.deepseek:
      case AIProviders.togetherai:
      case AIProviders.perplexity:
      case AIProviders.mistral:
        return {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer $apiKey',
        };
      case AIProviders.anthropic:
        return {
          'Content-Type': 'application/json',
          'x-api-key': apiKey,
          'anthropic-version': '2023-06-01',
        };
      case AIProviders.google:
        return {
          'Content-Type': 'application/json',
        };
      case AIProviders.ollama:
        return {
          'Content-Type': 'application/json',
        };
      default:
        return {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer $apiKey',
        };
    }
  }
  
  Future<AIResponse> chat({
    required String model,
    required List<AIMessage> messages,
    int? maxTokens,
    double? temperature,
  }) async {
    try {
      final endpoint = _getEndpoint(model);
      final data = _buildRequestBody(model: model, messages: messages, maxTokens: maxTokens, temperature: temperature);
      final response = await _dio.post<Map<String, dynamic>>(endpoint, data: data);
      return _parseResponse(response.data ?? <String, dynamic>{});
    } on DioException catch (e) {
      throw AIException(message: e.message ?? '请求失败', code: e.response?.statusCode?.toString());
    }
  }
  
  String _getEndpoint(String model) {
    switch (provider) {
      case AIProviders.openai:
      case AIProviders.openrouter:
      case AIProviders.groq:
      case AIProviders.deepseek:
      case AIProviders.togetherai:
      case AIProviders.perplexity:
      case AIProviders.mistral:
        return '/chat/completions';
      case AIProviders.anthropic:
        return '/v1/messages';
      case AIProviders.google:
        return '/v1beta/models/$model:generateContent';
      case AIProviders.ollama:
        return '/api/chat';
      default:
        return '/chat/completions';
    }
  }
  
  Map<String, dynamic> _buildRequestBody({
    required String model,
    required List<AIMessage> messages,
    int? maxTokens,
    double? temperature,
  }) {
    switch (provider) {
      case AIProviders.anthropic:
        return {
          'model': model,
          'messages': messages.map((m) => m.toJson()).toList(),
          'max_tokens': maxTokens ?? 4096,
        };
      case AIProviders.google:
        return {
          'contents': messages.map((m) => {'role': m.role == 'user' ? 'user' : 'model', 'parts': [{'text': m.content}]}).toList(),
        };
      case AIProviders.ollama:
        return {'model': model, 'messages': messages.map((m) => m.toJson()).toList(), 'stream': false};
      default:
        return {
          'model': model,
          'messages': messages.map((m) => m.toJson()).toList(),
          if (maxTokens != null) 'max_tokens': maxTokens,
          if (temperature != null) 'temperature': temperature,
        };
    }
  }
  
  AIResponse _parseResponse(Map<String, dynamic> data) {
    String content = '';
    String model = '';
    AIUsage usage = AIUsage(promptTokens: 0, completionTokens: 0, totalTokens: 0);
    
    switch (provider) {
      case AIProviders.anthropic:
        if (data['content'] is List && (data['content'] as List).isNotEmpty) {
          content = (data['content'] as List)[0]['text']?.toString() ?? '';
        }
        model = data['model']?.toString() ?? '';
        final usageData = data['usage'] as Map<String, dynamic>? ?? {};
        usage = AIUsage(
          promptTokens: usageData['input_tokens'] as int? ?? 0,
          completionTokens: usageData['output_tokens'] as int? ?? 0,
          totalTokens: (usageData['input_tokens'] as int? ?? 0) + (usageData['output_tokens'] as int? ?? 0),
        );
        break;
      case AIProviders.google:
        final candidates = data['candidates'] as List? ?? [];
        if (candidates.isNotEmpty) {
          content = candidates[0]['content']?['parts']?[0]?['text']?.toString() ?? '';
        }
        model = data['modelVersion']?.toString() ?? '';
        break;
      default:
        final choices = data['choices'] as List? ?? [];
        if (choices.isNotEmpty) {
          content = choices[0]['message']?['content']?.toString() ?? '';
        }
        model = data['model']?.toString() ?? '';
        final usageMap = data['usage'] as Map<String, dynamic>? ?? {};
        usage = AIUsage(
          promptTokens: usageMap['prompt_tokens'] as int? ?? 0,
          completionTokens: usageMap['completion_tokens'] as int? ?? 0,
          totalTokens: usageMap['total_tokens'] as int? ?? 0,
        );
    }
    
    return AIResponse(content: content, model: model, usage: usage);
  }
  
  Future<List<String>> getModels() async {
    switch (provider) {
      case AIProviders.openai:
        return AIProviders.openaiModels;
      case AIProviders.anthropic:
        return AIProviders.anthropicModels;
      case AIProviders.ollama:
        final response = await _dio.get<Map<String, dynamic>>('/api/tags');
        final models = response.data?['models'] as List?;
        if (models == null) return [];
        return models.map((m) => m['name']?.toString() ?? '').toList();
      default:
        return [];
    }
  }
}

class AIMessage {
  final String role;
  final String content;
  AIMessage({required this.role, required this.content});
  Map<String, dynamic> toJson() => {'role': role, 'content': content};
}

class AIResponse {
  final String content;
  final String model;
  final AIUsage usage;
  AIResponse({required this.content, required this.model, required this.usage});
}

class AIUsage {
  final int promptTokens;
  final int completionTokens;
  final int totalTokens;
  AIUsage({required this.promptTokens, required this.completionTokens, required this.totalTokens});
}

class AIException implements Exception {
  final String message;
  final String? code;
  AIException({required this.message, this.code});
  @override
  String toString() => 'AIException: $message (code: $code)';
}
