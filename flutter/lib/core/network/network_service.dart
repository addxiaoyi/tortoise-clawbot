import 'dart:async';
import 'package:dio/dio.dart';
import '../logging/logger_service.dart';

/// 网络服务 - 统一的 HTTP 客户端
class NetworkService {
  static NetworkService? _instance;
  static NetworkService get instance => _instance ??= NetworkService._();
  NetworkService._();

  late final Dio _dio;
  final LoggerService _logger = LoggerService.instance;

  /// 默认超时时间
  static const Duration defaultConnectTimeout = Duration(seconds: 30);
  static const Duration defaultReceiveTimeout = Duration(seconds: 30);
  static const Duration defaultSendTimeout = Duration(seconds: 30);

  /// 初始化
  void initialize({
    String? baseUrl,
    Map<String, String>? headers,
    Duration? connectTimeout,
    Duration? receiveTimeout,
  }) {
    _dio = Dio(BaseOptions(
      baseUrl: baseUrl ?? '',
      connectTimeout: connectTimeout ?? defaultConnectTimeout,
      receiveTimeout: receiveTimeout ?? defaultReceiveTimeout,
      headers: {
        'Content-Type': 'application/json',
        ...?headers,
      },
    ));

    _dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) {
        _logger.debug('${options.method} ${options.uri}', tag: 'Network');
        handler.next(options);
      },
      onResponse: (response, handler) {
        _logger.debug('${response.statusCode} ${response.requestOptions.uri}', tag: 'Network');
        handler.next(response);
      },
      onError: (error, handler) {
        _logger.error(
          '${error.response?.statusCode ?? 'N/A'} ${error.requestOptions.uri}',
          tag: 'Network',
          error: error.message,
        );
        handler.next(error);
      },
    ));

    _logger.info('NetworkService initialized', tag: 'Network');
  }

  /// GET 请求
  Future<Response<T>> get<T>(
    String path, {
    Map<String, dynamic>? queryParameters,
    Options? options,
    CancelToken? cancelToken,
  }) {
    return _dio.get<T>(
      path,
      queryParameters: queryParameters,
      options: options,
      cancelToken: cancelToken,
    );
  }

  /// POST 请求
  Future<Response<T>> post<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    Options? options,
    CancelToken? cancelToken,
  }) {
    return _dio.post<T>(
      path,
      data: data,
      queryParameters: queryParameters,
      options: options,
      cancelToken: cancelToken,
    );
  }

  /// PUT 请求
  Future<Response<T>> put<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    Options? options,
    CancelToken? cancelToken,
  }) {
    return _dio.put<T>(
      path,
      data: data,
      queryParameters: queryParameters,
      options: options,
      cancelToken: cancelToken,
    );
  }

  /// DELETE 请求
  Future<Response<T>> delete<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    Options? options,
    CancelToken? cancelToken,
  }) {
    return _dio.delete<T>(
      path,
      data: data,
      queryParameters: queryParameters,
      options: options,
      cancelToken: cancelToken,
    );
  }

  /// PATCH 请求
  Future<Response<T>> patch<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    Options? options,
    CancelToken? cancelToken,
  }) {
    return _dio.patch<T>(
      path,
      data: data,
      queryParameters: queryParameters,
      options: options,
      cancelToken: cancelToken,
    );
  }

  /// 健康检查
  Future<bool> healthCheck(String url) async {
    try {
      final response = await _dio.get(url);
      return response.statusCode == 200;
    } catch (_) {
      return false;
    }
  }

  /// 取消所有请求
  void cancelAll(CancelToken cancelToken) {
    cancelToken.cancel('Cancelled by user');
  }
}
