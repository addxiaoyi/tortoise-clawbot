import 'dart:convert';
import 'package:dio/dio.dart';

/// Anthropic Claude API 服务
class AnthropicService {
  final String apiKey;
  final String baseUrl;
  late final Dio _dio;

  AnthropicService({
    required this.apiKey,
    this.baseUrl = 'https://api.anthropic.com',
  }) {
    _dio = Dio(BaseOptions(
      baseUrl: baseUrl,
      headers: {
        'x-api-key': apiKey,
        'anthropic-version': '2023-06-01',
        'Content-Type': 'application/json',
      },
    ));
  }

  /// 发送消息
  Future<ClaudeResponse> sendMessage(ClaudeRequest request) async {
    try {
      final response = await _dio.post(
        '/v1/messages',
        data: request.toJson(),
      );
      return ClaudeResponse.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }

  /// 流式发送消息
  Stream<String> streamMessage(ClaudeRequest request) async* {
    try {
      final response = await _dio.post(
        '/v1/messages',
        data: request.toJson(),
        options: Options(responseType: ResponseType.stream),
      );

      final stream = response.data.stream as Stream<List<int>>;
      await for (final chunk in stream) {
        final text = utf8.decode(chunk);
        // 解析 SSE 格式
        for (final line in text.split('\n')) {
          if (line.startsWith('data: ')) {
            final data = line.substring(6);
            if (data == '[DONE]') return;
            try {
              final json = _parseSSE(data);
              if (json['type'] == 'content_block_delta' &&
                  json['delta']?['type'] == 'text_delta' &&
                  json['delta']?['text'] != null) {
                yield json['delta']['text'] as String;
              }
            } catch (_) {}
          }
        }
      }
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }

  Map<String, dynamic> _parseSSE(String data) {
    try {
      return {'type': 'content_block_delta', 'delta': {'text': data}};
    } catch (_) {
      return <String, dynamic>{};
    }
  }

  Exception _handleError(DioException e) {
    switch (e.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
        return Exception('连接超时，请检查网络');
      case DioExceptionType.badResponse:
        final statusCode = e.response?.statusCode;
        final error = e.response?.data?['error'] as Map<String, dynamic>?;
        final message = error?['message'] ?? error?['type'] ?? '未知错误';
        return Exception('Claude API 错误 ($statusCode): $message');
      default:
        return Exception('网络错误: ${e.message}');
    }
  }
}

/// Claude 请求
class ClaudeRequest {
  final String model;
  final List<ClaudeMessage> messages;
  final int? maxTokens;
  final double? temperature;
  final String? systemPrompt;
  final List<String>? stopSequences;
  final bool stream;

  ClaudeRequest({
    required this.model,
    required this.messages,
    this.maxTokens,
    this.temperature,
    this.systemPrompt,
    this.stopSequences,
    this.stream = false,
  });

  Map<String, dynamic> toJson() {
    final json = <String, dynamic>{
      'model': model,
      'messages': messages.map((m) => m.toJson()).toList(),
      'max_tokens': maxTokens ?? 1024,
      'stream': stream,
    };
    if (temperature != null) json['temperature'] = temperature;
    if (systemPrompt != null) json['system'] = systemPrompt;
    if (stopSequences != null) json['stop_sequences'] = stopSequences;
    return json;
  }
}

/// Claude 消息
class ClaudeMessage {
  final String role;
  final String content;

  ClaudeMessage({
    required this.role,
    required this.content,
  });

  Map<String, dynamic> toJson() {
    return {'role': role, 'content': content};
  }
}

/// Claude 响应
class ClaudeResponse {
  final String id;
  final String type;
  final String role;
  final List<ContentBlock> content;
  final String? stopReason;
  final String? stopSequence;
  final int usageInputTokens;
  final int usageOutputTokens;

  ClaudeResponse.fromJson(Map<String, dynamic> json)
      : id = json['id'] as String? ?? '',
        type = json['type'] as String? ?? '',
        role = json['role'] as String? ?? '',
        content = (json['content'] as List?)
                ?.map((c) => ContentBlock.fromJson(c as Map<String, dynamic>))
                .toList() ??
            [],
        stopReason = json['stop_reason'] as String?,
        stopSequence = json['stop_sequence'] as String?,
        usageInputTokens = json['usage']?['input_tokens'] as int? ?? 0,
        usageOutputTokens = json['usage']?['output_tokens'] as int? ?? 0;

  String get text {
    return content
        .where((c) => c.type == 'text')
        .map((c) => c.text ?? '')
        .join('\n');
  }
}

/// 内容块
class ContentBlock {
  final String type;
  final String? text;

  ContentBlock.fromJson(Map<String, dynamic> json)
      : type = json['type'] as String? ?? '',
        text = json['text'] as String?;
}
