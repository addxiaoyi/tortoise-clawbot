import 'package:http/http.dart' as http;
import 'dart:convert';

// API 基础地址
const String DEFAULT_API_BASE = 'http://localhost:18792';

// API 响应类型
class ApiResponse<T> {
  final bool success;
  final T? data;
  final String? error;

  ApiResponse({required this.success, this.data, this.error});
}

// AI 提供商
class AIProvider {
  final String id;
  final String name;
  final bool enabled;
  final String apiKey;
  final String model;
  final String baseUrl;

  AIProvider({
    required this.id,
    required this.name,
    this.enabled = false,
    this.apiKey = '',
    required this.model,
    required this.baseUrl,
  });

  factory AIProvider.fromJson(Map<String, dynamic> json) {
    return AIProvider(
      id: json['id'] ?? '',
      name: json['name'] ?? '',
      enabled: json['enabled'] ?? false,
      apiKey: json['api_key'] ?? '',
      model: json['model'] ?? '',
      baseUrl: json['base_url'] ?? '',
    );
  }

  Map<String, dynamic> toJson() => {
    'id': id,
    'name': name,
    'enabled': enabled,
    'api_key': apiKey,
    'model': model,
    'base_url': baseUrl,
  };
}

// 渠道状态
class ChannelStatus {
  final String type;
  final String name;
  final bool connected;
  final int messageCount;
  final DateTime? lastMessage;

  ChannelStatus({
    required this.type,
    required this.name,
    required this.connected,
    this.messageCount = 0,
    this.lastMessage,
  });

  factory ChannelStatus.fromJson(Map<String, dynamic> json) {
    return ChannelStatus(
      type: json['type'] ?? '',
      name: json['name'] ?? '',
      connected: json['connected'] ?? false,
      messageCount: json['message_count'] ?? 0,
      lastMessage: json['last_message'] != null
          ? DateTime.tryParse(json['last_message'])
          : null,
    );
  }
}

// 统计
class Stats {
  final int requestsTotal;
  final int requestsSuccess;
  final int requestsFailed;
  final double avgLatencyMs;
  final int tokensUsed;
  final double costUsd;

  Stats({
    required this.requestsTotal,
    required this.requestsSuccess,
    required this.requestsFailed,
    required this.avgLatencyMs,
    required this.tokensUsed,
    required this.costUsd,
  });

  factory Stats.fromJson(Map<String, dynamic> json) {
    return Stats(
      requestsTotal: json['requests_total'] ?? 0,
      requestsSuccess: json['requests_success'] ?? 0,
      requestsFailed: json['requests_failed'] ?? 0,
      avgLatencyMs: (json['avg_latency_ms'] ?? 0).toDouble(),
      tokensUsed: json['tokens_used'] ?? 0,
      costUsd: (json['cost_usd'] ?? 0).toDouble(),
    );
  }
}

// API 服务类
class ApiService {
  final String baseUrl;
  final String? apiKey;
  final http.Client _client;

  ApiService({
    String? baseUrl,
    this.apiKey,
    http.Client? client,
  })  : baseUrl = baseUrl ?? DEFAULT_API_BASE,
        _client = client ?? http.Client();

  Map<String, String> get _headers => {
        'Content-Type': 'application/json',
        if (apiKey != null && apiKey!.isNotEmpty) 'Authorization': 'Bearer $apiKey',
      };

  // 健康检查
  Future<ApiResponse<Map<String, dynamic>>> health() async {
    try {
      final response = await _client.get(
        Uri.parse('$baseUrl/api/v1/health'),
        headers: _headers,
      );
      if (response.statusCode == 200) {
        return ApiResponse(
          success: true,
          data: jsonDecode(response.body),
        );
      }
      return ApiResponse(success: false, error: 'Health check failed');
    } catch (e) {
      return ApiResponse(success: false, error: e.toString());
    }
  }

  // 获取配置
  Future<ApiResponse<Map<String, dynamic>>> getConfig() async {
    try {
      final response = await _client.get(
        Uri.parse('$baseUrl/api/v1/config'),
        headers: _headers,
      );
      if (response.statusCode == 200) {
        return ApiResponse(
          success: true,
          data: jsonDecode(response.body),
        );
      }
      return ApiResponse(success: false, error: 'Failed to get config');
    } catch (e) {
      return ApiResponse(success: false, error: e.toString());
    }
  }

  // 更新配置
  Future<ApiResponse<Map<String, dynamic>>> updateConfig(Map<String, dynamic> updates) async {
    try {
      final response = await _client.patch(
        Uri.parse('$baseUrl/api/v1/config'),
        headers: _headers,
        body: jsonEncode(updates),
      );
      if (response.statusCode == 200) {
        return ApiResponse(
          success: true,
          data: jsonDecode(response.body),
        );
      }
      return ApiResponse(success: false, error: 'Failed to update config');
    } catch (e) {
      return ApiResponse(success: false, error: e.toString());
    }
  }

  // AI 聊天
  Future<ApiResponse<Map<String, dynamic>>> chat({
    required String model,
    required List<Map<String, String>> messages,
    double? temperature,
    int? maxTokens,
  }) async {
    try {
      final response = await _client.post(
        Uri.parse('$baseUrl/api/v1/ai/chat'),
        headers: _headers,
        body: jsonEncode({
          'model': model,
          'messages': messages,
          if (temperature != null) 'temperature': temperature,
          if (maxTokens != null) 'max_tokens': maxTokens,
        }),
      );
      if (response.statusCode == 200) {
        return ApiResponse(
          success: true,
          data: jsonDecode(response.body),
        );
      }
      return ApiResponse(success: false, error: 'Chat failed');
    } catch (e) {
      return ApiResponse(success: false, error: e.toString());
    }
  }

  // 获取 AI 提供商
  Future<ApiResponse<List<AIProvider>>> getAIProviders() async {
    try {
      final response = await _client.get(
        Uri.parse('$baseUrl/api/v1/ai/providers'),
        headers: _headers,
      );
      if (response.statusCode == 200) {
        final List<dynamic> data = jsonDecode(response.body);
        return ApiResponse(
          success: true,
          data: data.map((p) => AIProvider.fromJson(p)).toList(),
        );
      }
      return ApiResponse(success: false, error: 'Failed to get providers');
    } catch (e) {
      return ApiResponse(success: false, error: e.toString());
    }
  }

  // 获取渠道列表
  Future<ApiResponse<List<ChannelStatus>>> getChannels() async {
    try {
      final response = await _client.get(
        Uri.parse('$baseUrl/api/v1/channels'),
        headers: _headers,
      );
      if (response.statusCode == 200) {
        final List<dynamic> data = jsonDecode(response.body);
        return ApiResponse(
          success: true,
          data: data.map((c) => ChannelStatus.fromJson(c)).toList(),
        );
      }
      return ApiResponse(success: false, error: 'Failed to get channels');
    } catch (e) {
      return ApiResponse(success: false, error: e.toString());
    }
  }

  // 连接渠道
  Future<ApiResponse<void>> connectChannel(String type) async {
    try {
      final response = await _client.post(
        Uri.parse('$baseUrl/api/v1/channels/$type/connect'),
        headers: _headers,
      );
      if (response.statusCode == 200) {
        return ApiResponse(success: true);
      }
      return ApiResponse(success: false, error: 'Failed to connect channel');
    } catch (e) {
      return ApiResponse(success: false, error: e.toString());
    }
  }

  // 断开渠道
  Future<ApiResponse<void>> disconnectChannel(String type) async {
    try {
      final response = await _client.post(
        Uri.parse('$baseUrl/api/v1/channels/$type/disconnect'),
        headers: _headers,
      );
      if (response.statusCode == 200) {
        return ApiResponse(success: true);
      }
      return ApiResponse(success: false, error: 'Failed to disconnect channel');
    } catch (e) {
      return ApiResponse(success: false, error: e.toString());
    }
  }

  // 获取统计
  Future<ApiResponse<Stats>> getStats() async {
    try {
      final response = await _client.get(
        Uri.parse('$baseUrl/api/v1/stats'),
        headers: _headers,
      );
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return ApiResponse(
          success: true,
          data: Stats.fromJson(data['ai'] ?? {}),
        );
      }
      return ApiResponse(success: false, error: 'Failed to get stats');
    } catch (e) {
      return ApiResponse(success: false, error: e.toString());
    }
  }

  // 创建会话
  Future<ApiResponse<String>> createSession({String? userId, String? model}) async {
    try {
      final response = await _client.post(
        Uri.parse('$baseUrl/api/v1/sessions'),
        headers: _headers,
        body: jsonEncode({
          if (userId != null) 'user_id': userId,
          if (model != null) 'model': model,
        }),
      );
      if (response.statusCode == 201) {
        final data = jsonDecode(response.body);
        return ApiResponse(success: true, data: data['id']);
      }
      return ApiResponse(success: false, error: 'Failed to create session');
    } catch (e) {
      return ApiResponse(success: false, error: e.toString());
    }
  }

  // 获取会话列表
  Future<ApiResponse<List<Map<String, dynamic>>>> getSessions() async {
    try {
      final response = await _client.get(
        Uri.parse('$baseUrl/api/v1/sessions'),
        headers: _headers,
      );
      if (response.statusCode == 200) {
        final List<dynamic> data = jsonDecode(response.body);
        return ApiResponse(
          success: true,
          data: data.cast<Map<String, dynamic>>(),
        );
      }
      return ApiResponse(success: false, error: 'Failed to get sessions');
    } catch (e) {
      return ApiResponse(success: false, error: e.toString());
    }
  }

  // 发送消息
  Future<ApiResponse<Map<String, dynamic>>> sendMessage(String sessionId, String content) async {
    try {
      final response = await _client.post(
        Uri.parse('$baseUrl/api/v1/sessions/$sessionId/messages'),
        headers: _headers,
        body: jsonEncode({'content': content}),
      );
      if (response.statusCode == 200) {
        return ApiResponse(
          success: true,
          data: jsonDecode(response.body),
        );
      }
      return ApiResponse(success: false, error: 'Failed to send message');
    } catch (e) {
      return ApiResponse(success: false, error: e.toString());
    }
  }

  void dispose() {
    _client.close();
  }
}

// 导出单例
final apiService = ApiService();
