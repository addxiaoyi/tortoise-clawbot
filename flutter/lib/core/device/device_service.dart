import 'package:flutter/foundation.dart';

/// 设备信息服务
class DeviceService {
  static DeviceService? _instance;
  static DeviceService get instance => _instance ??= DeviceService._();
  DeviceService._();

  /// 获取设备信息
  Map<String, dynamic> getDeviceInfo() {
    return {
      'platform': platform,
      'operatingSystem': operatingSystem,
      'isWeb': isWeb,
      'isAndroid': isAndroid,
      'isIOS': isIOS,
      'isMacOS': isMacOS,
      'isWindows': isWindows,
      'isLinux': isLinux,
      'timestamp': DateTime.now().toIso8601String(),
    };
  }

  /// 获取平台名称
  String get platform {
    if (kIsWeb) return 'web';
    return 'unknown';
  }

  /// 获取操作系统
  String get operatingSystem {
    if (kIsWeb) return 'Web Browser';
    return 'Unknown';
  }

  /// 是否是 Web
  bool get isWeb => kIsWeb;

  /// 是否是 Android
  bool get isAndroid => !kIsWeb && platform == 'android';

  /// 是否是 iOS
  bool get isIOS => !kIsWeb && platform == 'ios';

  /// 是否是 macOS
  bool get isMacOS => !kIsWeb && platform == 'macos';

  /// 是否是 Windows
  bool get isWindows => !kIsWeb && platform == 'windows';

  /// 是否是 Linux
  bool get isLinux => !kIsWeb && platform == 'linux';

  /// 是否是移动端
  bool get isMobile => isAndroid || isIOS;

  /// 是否是桌面端
  bool get isDesktop => isMacOS || isWindows || isLinux;

  /// 获取应用版本
  String get appVersion => '1.0.0';

  /// 获取构建号
  String get buildNumber => '1';
}
