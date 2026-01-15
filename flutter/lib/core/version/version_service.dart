import 'package:flutter/foundation.dart';

/// 版本信息服务
class VersionService {
  static VersionService? _instance;
  static VersionService get instance => _instance ??= VersionService._();
  VersionService._();

  /// 当前版本
  String get currentVersion => '1.0.0';

  /// 构建号
  String get buildNumber => '1';

  /// 版本名称 (如 1.0.0)
  String get versionName => currentVersion;

  /// 比较版本
  int compareVersions(String v1, String v2) {
    final parts1 = v1.split('.').map(int.parse).toList();
    final parts2 = v2.split('.').map(int.parse).toList();

    for (int i = 0; i < 3; i++) {
      final p1 = i < parts1.length ? parts1[i] : 0;
      final p2 = i < parts2.length ? parts2[i] : 0;
      if (p1 != p2) return p1 - p2;
    }
    return 0;
  }

  /// 检查是否是最新版本
  bool isLatestVersion(String latestVersion) {
    return compareVersions(currentVersion, latestVersion) >= 0;
  }

  /// 检查是否有更新
  bool hasUpdate(String latestVersion) {
    return compareVersions(currentVersion, latestVersion) < 0;
  }

  /// 获取版本信息
  Map<String, String> getVersionInfo() {
    return {
      'version': currentVersion,
      'build': buildNumber,
      'timestamp': DateTime.now().toIso8601String(),
    };
  }

  /// 打印版本信息
  void printVersionInfo() {
    debugPrint('App Version: $versionName ($buildNumber)');
  }
}
