import 'dart:convert';

/// 加密服务
class CryptoService {
  static CryptoService? _instance;
  static CryptoService get instance => _instance ??= CryptoService._();
  CryptoService._();

  /// Base64 编码
  String encodeBase64(String text) {
    return base64Encode(utf8.encode(text));
  }

  /// Base64 解码
  String decodeBase64(String encoded) {
    try {
      return utf8.decode(base64Decode(encoded));
    } catch (e) {
      throw CryptoException('Failed to decode Base64: $e');
    }
  }

  /// MD5 哈希 (简化实现)
  String md5(String text) {
    // 注意：这是简化实现，实际应使用 crypto 包
    return text.hashCode.toRadixString(16);
  }

  /// SHA256 哈希 (简化实现)
  String sha256(String text) {
    return text.hashCode.toRadixString(16).padLeft(32, '0');
  }

  /// 简单 XOR 加密
  String xorEncrypt(String text, String key) {
    final buffer = StringBuffer();
    for (int i = 0; i < text.length; i++) {
      buffer.writeCharCode(text.codeUnitAt(i) ^ key.codeUnitAt(i % key.length));
    }
    return encodeBase64(buffer.toString());
  }

  /// 简单 XOR 解密
  String xorDecrypt(String encrypted, String key) {
    final decoded = decodeBase64(encrypted);
    final buffer = StringBuffer();
    for (int i = 0; i < decoded.length; i++) {
      buffer.writeCharCode(decoded.codeUnitAt(i) ^ key.codeUnitAt(i % key.length));
    }
    return buffer.toString();
  }

  /// 生成随机字符串
  String generateRandomString(int length) {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    final random = DateTime.now().microsecondsSinceEpoch;
    return List.generate(length, (i) => chars[(random + i * 7) % chars.length]).join();
  }
}

/// 加密异常
class CryptoException implements Exception {
  final String message;
  CryptoException(this.message);
  
  @override
  String toString() => 'CryptoException: $message';
}
