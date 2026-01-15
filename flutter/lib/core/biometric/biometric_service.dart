import 'package:flutter/foundation.dart';

/// 生物识别类型
enum BiometricType {
  fingerprint,
  face,
  iris,
  none,
}

/// 生物识别服务
class BiometricService {
  static BiometricService? _instance;
  static BiometricService get instance => _instance ??= BiometricService._();
  BiometricService._();

  bool _isAvailable = false;

  /// 检查是否可用
  Future<bool> isAvailable() async {
    // Web 平台不支持生物识别
    debugPrint('Biometric not available on web');
    return false;
  }

  /// 获取生物识别类型
  Future<BiometricType> getBiometricType() async {
    return BiometricType.none;
  }

  /// 认证
  Future<bool> authenticate({String reason = '请进行身份验证'}) async {
    debugPrint('Biometric authentication: $reason');
    return false;
  }

  /// 是否已启用
  bool get isEnabled => _isAvailable;
}
