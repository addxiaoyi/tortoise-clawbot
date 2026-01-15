import 'dart:async';
import 'dart:convert';
import 'package:dio/dio.dart';

/// OpenAI 真实 API 服务
class OpenAIService {
  final String apiKey;
  final String baseUrl;
  late final Dio _dio;

  OpenAIService({
    required this.apiKey,
    this.baseUrl = 'https://api.openai.com/v1',
  }) {
    _dio = Dio(BaseOptions(
      baseUrl: baseUrl,
      headers: {
        'Authorization': 'Bearer $apiKey',
        'Content-Type': 'application/json',
      },
    ));
  }

  /// 聊天补全
  Future<ChatResponse> chat(ChatRequest request) async {
    try {
      final response = await _dio.post<Map<String, dynamic>>(
        '/chat/completions',
        data: request.toJson(),
      );
      return ChatResponse.fromJson(response.data ?? <String, dynamic>{});
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }

  /// 流式聊天
  Stream<String> streamChat(ChatRequest request) async* {
    try {
      final response = await _dio.post<ResponseBody>(
        '/chat/completions',
        data: request.toJson(),
        options: Options(responseType: ResponseType.stream),
      );

      final stream = response.data?.stream as Stream<List<int>>?;
      if (stream == null) return;

      await for (final chunk in stream) {
        final text = utf8.decode(chunk);
        for (final line in text.split('\n')) {
          if (line.startsWith('data: ')) {
            final data = line.substring(6);
            if (data == '[DONE]') return;
            try {
              final json = jsonDecode(data) as Map<String, dynamic>;
              final delta = json['choices']?[0]?['delta']?['content'];
              if (delta != null) {
                yield delta.toString();
              }
            } catch (_) {}
          }
        }
      }
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }

  /// 获取可用模型列表
  Future<List<Model>> getModels() async {
    try {
      final response = await _dio.get<Map<String, dynamic>>('/models');
      final data = response.data?['data'] as List?;
      if (data == null) return [];
      return data.map((m) => Model.fromJson(m as Map<String, dynamic>)).toList();
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }

  Exception _handleError(DioException e) {
    switch (e.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
        return Exception('连接超时');
      case DioExceptionType.badResponse:
        final statusCode = e.response?.statusCode;
        final message = e.response?.data?['error']?['message'] ?? '未知错误';
        return Exception('API错误 ($statusCode): $message');
      default:
        return Exception('网络错误: ${e.message}');
    }
  }
}

/// 聊天请求
class ChatRequest {
  final String model;
  final List<ChatMessage> messages;
  final double? temperature;
  final int? maxTokens;
  final double? topP;
  final int? n;
  final bool? stream;
  final List<String>? stop;

  ChatRequest({
    required this.model,
    required this.messages,
    this.temperature,
    this.maxTokens,
    this.topP,
    this.n,
    this.stream,
    this.stop,
  });

  Map<String, dynamic> toJson() {
    final json = <String, dynamic>{
      'model': model,
      'messages': messages.map((m) => m.toJson()).toList(),
    };
    if (temperature != null) json['temperature'] = temperature;
    if (maxTokens != null) json['max_tokens'] = maxTokens;
    if (topP != null) json['top_p'] = topP;
    if (n != null) json['n'] = n;
    if (stream != null) json['stream'] = stream;
    if (stop != null) json['stop'] = stop;
    return json;
  }
}

/// 聊天消息
class ChatMessage {
  final String role;
  final String content;

  ChatMessage({
    required this.role,
    required this.content,
  });

  Map<String, dynamic> toJson() => {'role': role, 'content': content};
}

/// 聊天响应
class ChatResponse {
  final String id;
  final String model;
  final List<ChatChoice> choices;
  final ChatUsage usage;

  ChatResponse({
    required this.id,
    required this.model,
    required this.choices,
    required this.usage,
  });

  factory ChatResponse.fromJson(Map<String, dynamic> json) {
    final choicesData = json['choices'] as List? ?? [];
    final usageData = json['usage'] as Map<String, dynamic>? ?? {};
    
    return ChatResponse(
      id: json['id']?.toString() ?? '',
      model: json['model']?.toString() ?? '',
      choices: choicesData.map((c) => ChatChoice.fromJson(c as Map<String, dynamic>)).toList(),
      usage: ChatUsage.fromJson(usageData),
    );
  }
}

/// 聊天选择
class ChatChoice {
  final int index;
  final ChatMessage message;
  final String finishReason;

  ChatChoice({
    required this.index,
    required this.message,
    required this.finishReason,
  });

  factory ChatChoice.fromJson(Map<String, dynamic> json) {
    return ChatChoice(
      index: json['index'] as int? ?? 0,
      message: ChatMessage(
        role: json['message']?['role']?.toString() ?? '',
        content: json['message']?['content']?.toString() ?? '',
      ),
      finishReason: json['finish_reason']?.toString() ?? '',
    );
  }
}

/// 使用量
class ChatUsage {
  final int promptTokens;
  final int completionTokens;
  final int totalTokens;

  ChatUsage({
    required this.promptTokens,
    required this.completionTokens,
    required this.totalTokens,
  });

  factory ChatUsage.fromJson(Map<String, dynamic> json) {
    return ChatUsage(
      promptTokens: json['prompt_tokens'] as int? ?? 0,
      completionTokens: json['completion_tokens'] as int? ?? 0,
      totalTokens: json['total_tokens'] as int? ?? 0,
    );
  }
}

/// 模型
class Model {
  final String id;
  final String object;
  final String ownedBy;
  final bool ready;

  Model({
    required this.id,
    required this.object,
    required this.ownedBy,
    required this.ready,
  });

  factory Model.fromJson(Map<String, dynamic> json) {
    return Model(
      id: json['id']?.toString() ?? '',
      object: json['object']?.toString() ?? '',
      ownedBy: json['owned_by']?.toString() ?? '',
      ready: json['ready'] as bool? ?? true,
    );
  }
}
