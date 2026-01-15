import 'package:flutter/foundation.dart';

/// 平台服务 - 提供平台特定功能
class PlatformService {
  static PlatformService? _instance;
  static PlatformService get instance => _instance ??= PlatformService._();
  PlatformService._();

  /// 当前平台
  AppPlatform get platform {
    if (kIsWeb) return AppPlatform.web;
    // 注意: 在 web 环境中无法使用 dart:io
    // 这里简化处理，实际应根据平台选择器返回
    return AppPlatform.web;
  }

  /// 是否是 Web 平台
  bool get isWeb => platform == AppPlatform.web;
  
  /// 是否是移动端
  bool get isMobile => platform == AppPlatform.ios || platform == AppPlatform.android;
  
  /// 是否是桌面端
  bool get isDesktop => platform == AppPlatform.macos || 
                        platform == AppPlatform.windows || 
                        platform == AppPlatform.linux;

  /// 获取平台信息
  Map<String, dynamic> getPlatformInfo() {
    return {
      'platform': platform.name,
      'isWeb': isWeb,
      'isMobile': isMobile,
      'isDesktop': isDesktop,
    };
  }

  /// 检查权限
  Future<bool> checkPermission(String permission) async {
    // 根据平台检查权限
    return true;
  }

  /// 请求权限
  Future<bool> requestPermission(String permission) async {
    // 请求权限
    return true;
  }

  /// 获取设备信息
  Map<String, String> getDeviceInfo() {
    return {
      'platform': platform.name,
      'timestamp': DateTime.now().toIso8601String(),
    };
  }
}

/// 应用平台
enum AppPlatform {
  ios,
  android,
  web,
  windows,
  macos,
  linux,
  unknown,
}
