import 'dart:async';
import 'package:dio/dio.dart';

/// API 服务 - 与后端通信
class ApiService {
  static ApiService? _instance;
  late final Dio _dio;
  String? _baseUrl;
  String? _apiKey;

  ApiService._();

  static ApiService get instance {
    _instance ??= ApiService._();
    return _instance!;
  }

  /// 初始化
  void init({required String baseUrl, required String apiKey}) {
    _baseUrl = baseUrl;
    _apiKey = apiKey;

    _dio = Dio(BaseOptions(
      baseUrl: baseUrl,
      connectTimeout: const Duration(seconds: 30),
      receiveTimeout: const Duration(seconds: 30),
      headers: {
        'Authorization': 'Bearer $apiKey',
        'Content-Type': 'application/json',
      },
    ));

    // 添加拦截器
    _dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) {
        options.headers['Authorization'] = 'Bearer $apiKey';
        return handler.next(options);
      },
      onError: (error, handler) {
        _handleError(error);
        return handler.next(error);
      },
    ));
  }

  /// 是否已初始化
  bool get isInitialized => _baseUrl != null && _apiKey != null;

  // ========== 会话 API ==========

  /// 列出会话
  Future<List<Map<String, dynamic>>> listSessions() async {
    final response = await _dio.get<Map<String, dynamic>>('/api/v1/sessions');
    final data = response.data?['sessions'];
    if (data == null) return [];
    return (data as List).cast<Map<String, dynamic>>();
  }

  /// 创建会话
  Future<Map<String, dynamic>> createSession({
    required String title,
    String? aiProvider,
    String? model,
  }) async {
    final response = await _dio.post<Map<String, dynamic>>('/api/v1/sessions', data: {
      'title': title,
      if (aiProvider != null) 'ai_provider': aiProvider,
      if (model != null) 'model': model,
    });
    return response.data ?? <String, dynamic>{};
  }

  /// 获取会话
  Future<Map<String, dynamic>> getSession(String id) async {
    final response = await _dio.get<Map<String, dynamic>>('/api/v1/sessions/$id');
    return response.data ?? <String, dynamic>{};
  }

  /// 删除会话
  Future<void> deleteSession(String id) async {
    await _dio.delete('/api/v1/sessions/$id');
  }

  /// 获取会话消息
  Future<List<Map<String, dynamic>>> getSessionMessages(String sessionId) async {
    final response = await _dio.get<Map<String, dynamic>>('/api/v1/sessions/$sessionId/messages');
    final data = response.data?['messages'];
    if (data == null) return [];
    return (data as List).cast<Map<String, dynamic>>();
  }

  // ========== AI 聊天 API ==========

  /// 发送消息
  Future<Map<String, dynamic>> sendMessage({
    required String model,
    required List<Map<String, String>> messages,
    double? temperature,
    int? maxTokens,
    bool stream = false,
  }) async {
    final response = await _dio.post<Map<String, dynamic>>('/api/v1/chat/completions', data: {
      'model': model,
      'messages': messages,
      if (temperature != null) 'temperature': temperature,
      if (maxTokens != null) 'max_tokens': maxTokens,
      'stream': stream,
    });
    return response.data ?? <String, dynamic>{};
  }

  /// 流式聊天
  Stream<String> streamChat({
    required String model,
    required List<Map<String, String>> messages,
  }) async* {
    final response = await _dio.post<ResponseBody>(
      '/api/v1/chat/completions/stream',
      data: {
        'model': model,
        'messages': messages,
        'stream': true,
      },
      options: Options(responseType: ResponseType.stream),
    );

    final stream = response.data?.stream as Stream<List<int>>?;
    if (stream == null) return;

    await for (final chunk in stream) {
      final text = String.fromCharCodes(chunk);
      for (final line in text.split('\n')) {
        if (line.startsWith('data: ')) {
          yield line.substring(6);
        }
      }
    }
  }

  // ========== 渠道 API ==========

  /// 列出渠道
  Future<List<Map<String, dynamic>>> listChannels() async {
    final response = await _dio.get<Map<String, dynamic>>('/api/v1/channels');
    final data = response.data?['channels'];
    if (data == null) return [];
    return (data as List).cast<Map<String, dynamic>>();
  }

  /// 连接渠道
  Future<void> connectChannel(String id) async {
    await _dio.post('/api/v1/channels/$id/connect');
  }

  /// 断开渠道
  Future<void> disconnectChannel(String id) async {
    await _dio.post('/api/v1/channels/$id/disconnect');
  }

  // ========== 记忆 API ==========

  /// 列出记忆
  Future<List<Map<String, dynamic>>> listMemories() async {
    final response = await _dio.get<Map<String, dynamic>>('/api/v1/memory');
    final data = response.data?['memories'];
    if (data == null) return [];
    return (data as List).cast<Map<String, dynamic>>();
  }

  /// 创建记忆
  Future<Map<String, dynamic>> createMemory({
    required String title,
    required String content,
    String? type,
    List<String>? tags,
  }) async {
    final response = await _dio.post<Map<String, dynamic>>('/api/v1/memory', data: {
      'title': title,
      'content': content,
      if (type != null) 'type': type,
      if (tags != null) 'tags': tags,
    });
    return response.data ?? <String, dynamic>{};
  }

  // ========== 插件 API ==========

  /// 列出插件
  Future<List<Map<String, dynamic>>> listPlugins() async {
    final response = await _dio.get<Map<String, dynamic>>('/api/v1/plugins');
    final data = response.data?['plugins'];
    if (data == null) return [];
    return (data as List).cast<Map<String, dynamic>>();
  }

  /// 安装插件
  Future<void> installPlugin(String id) async {
    await _dio.post('/api/v1/plugins/install', data: {'id': id});
  }

  /// 启用插件
  Future<void> enablePlugin(String id) async {
    await _dio.post('/api/v1/plugins/$id/enable');
  }

  /// 禁用插件
  Future<void> disablePlugin(String id) async {
    await _dio.post('/api/v1/plugins/$id/disable');
  }

  // ========== 配置 API ==========

  /// 获取配置
  Future<Map<String, dynamic>> getConfig() async {
    final response = await _dio.get<Map<String, dynamic>>('/api/v1/config');
    return response.data ?? <String, dynamic>{};
  }

  /// 更新配置
  Future<void> updateConfig(Map<String, dynamic> config) async {
    await _dio.put('/api/v1/config', data: config);
  }

  // ========== 健康检查 ==========

  /// 健康检查
  Future<bool> healthCheck() async {
    try {
      final response = await _dio.get<Map<String, dynamic>>('/health');
      return response.data?['status'] == 'ok';
    } catch (_) {
      return false;
    }
  }

  // ========== 错误处理 ==========

  Exception _handleError(DioException error) {
    switch (error.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
        return Exception('连接超时，请检查网络');
      case DioExceptionType.badResponse:
        final statusCode = error.response?.statusCode;
        final message = error.response?.data?['error'] ?? '未知错误';
        return Exception('服务器错误 ($statusCode): $message');
      case DioExceptionType.cancel:
        return Exception('请求已取消');
      default:
        return Exception('网络错误: ${error.message}');
    }
  }
}
