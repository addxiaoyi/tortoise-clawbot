import 'package:flutter/material.dart';

/// 深色模式检测服务
class DarkModeService {
  static DarkModeService? _instance;
  static DarkModeService get instance => _instance ??= DarkModeService._();
  DarkModeService._();

  /// 检查系统是否使用深色模式
  bool isSystemDarkMode(BuildContext context) {
    final brightness = MediaQuery.platformBrightnessOf(context);
    return brightness == Brightness.dark;
  }

  /// 检查是否跟随系统
  bool shouldFollowSystem(ThemeMode themeMode) {
    return themeMode == ThemeMode.system;
  }

  /// 获取当前应该使用的主题模式
  ThemeMode getEffectiveThemeMode(ThemeMode themeMode, BuildContext context) {
    if (themeMode == ThemeMode.system) {
      return isSystemDarkMode(context) ? ThemeMode.dark : ThemeMode.light;
    }
    return themeMode;
  }

  /// 获取主题模式名称
  String getThemeModeName(ThemeMode mode) {
    switch (mode) {
      case ThemeMode.system:
        return '跟随系统';
      case ThemeMode.light:
        return '浅色';
      case ThemeMode.dark:
        return '深色';
    }
  }

  /// 获取主题模式图标
  IconData getThemeModeIcon(ThemeMode mode) {
    switch (mode) {
      case ThemeMode.system:
        return Icons.brightness_auto;
      case ThemeMode.light:
        return Icons.light_mode;
      case ThemeMode.dark:
        return Icons.dark_mode;
    }
  }
}
