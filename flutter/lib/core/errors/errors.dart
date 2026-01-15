// 应用异常类

/// 基础异常
class AppException implements Exception {
  final String code;
  final String message;
  final dynamic details;

  AppException({
    required this.code,
    required this.message,
    this.details,
  });

  @override
  String toString() => '[$code] $message';
}

/// 网络异常
class NetworkException extends AppException {
  NetworkException({
    required super.message,
    super.details,
  }) : super(code: 'NETWORK_ERROR');
}

/// API 异常
class ApiException extends AppException {
  final int? statusCode;

  ApiException({
    required super.message,
    this.statusCode,
    super.details,
  }) : super(code: 'API_ERROR');

  factory ApiException.fromStatusCode(int statusCode, [String? message]) {
    switch (statusCode) {
      case 400:
        return ApiException(message: message ?? '请求参数错误', statusCode: statusCode);
      case 401:
        return ApiException(message: message ?? '认证失败', statusCode: statusCode);
      case 403:
        return ApiException(message: message ?? '访问被拒绝', statusCode: statusCode);
      case 404:
        return ApiException(message: message ?? '资源不存在', statusCode: statusCode);
      case 429:
        return ApiException(message: message ?? '请求过于频繁', statusCode: statusCode);
      case 500:
        return ApiException(message: message ?? '服务器错误', statusCode: statusCode);
      default:
        return ApiException(message: message ?? '未知错误', statusCode: statusCode);
    }
  }
}

/// 存储异常
class StorageException extends AppException {
  StorageException({required super.message, super.details}) : super(code: 'STORAGE_ERROR');
}

/// 配置异常
class ConfigException extends AppException {
  ConfigException({required super.message, super.details}) : super(code: 'CONFIG_ERROR');
}

/// 渠道异常
class ChannelException extends AppException {
  final String? channelId;
  ChannelException({required super.message, this.channelId, super.details}) : super(code: 'CHANNEL_ERROR');
}

/// AI 引擎异常
class AIEngineException extends AppException {
  final String? provider;
  AIEngineException({required super.message, this.provider, super.details}) : super(code: 'AI_ENGINE_ERROR');
}

/// 会话异常
class SessionException extends AppException {
  SessionException({required super.message, super.details}) : super(code: 'SESSION_ERROR');
}

/// 插件异常
class PluginException extends AppException {
  final String? pluginId;
  PluginException({required super.message, this.pluginId, super.details}) : super(code: 'PLUGIN_ERROR');
}

/// 发现服务异常
class DiscoveryException extends AppException {
  DiscoveryException({required super.message, super.details}) : super(code: 'DISCOVERY_ERROR');
}
