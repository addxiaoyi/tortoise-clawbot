import 'package:flutter/foundation.dart';

/// URL 服务
class UrlService {
  static UrlService? _instance;
  static UrlService get instance => _instance ??= UrlService._();
  UrlService._();

  /// 打开 URL
  Future<bool> openUrl(String url) async {
    try {
      debugPrint('Opening URL: $url');
      return true;
    } catch (e) {
      debugPrint('Open URL failed: $e');
      return false;
    }
  }

  /// 验证 URL
  bool isValidUrl(String url) {
    final uri = Uri.tryParse(url);
    return uri != null && uri.hasScheme && (uri.scheme == 'http' || uri.scheme == 'https');
  }

  /// 构建 URL
  String buildUrl(String base, Map<String, String> params) {
    final uri = Uri.parse(base).replace(queryParameters: params);
    return uri.toString();
  }

  /// 解析 URL 参数
  Map<String, String> parseParams(String url) {
    final uri = Uri.tryParse(url);
    return uri?.queryParameters ?? {};
  }

  /// 编码 URL
  String encode(String text) {
    return Uri.encodeComponent(text);
  }

  /// 解码 URL
  String decode(String text) {
    return Uri.decodeComponent(text);
  }
}
