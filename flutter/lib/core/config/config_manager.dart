import 'package:flutter/material.dart';
import '../storage/storage_service.dart';

/// 配置管理器 - 统一管理应用配置
class ConfigManager {
  static ConfigManager? _instance;
  static ConfigManager get instance => _instance ??= ConfigManager._();

  ConfigManager._();

  final StorageService _storage = StorageService.instance;

  // ============== AI 配置 ==============

  /// 获取 API URL
  String get apiUrl => 
    _storage.getString('api_url') ?? 'http://localhost:18792';

  /// 设置 API URL
  Future<void> setApiUrl(String url) => 
    _storage.setString('api_url', url);

  /// 获取默认模型
  String get defaultModel => 
    _storage.getString('default_model') ?? 'gpt-4';

  /// 设置默认模型
  Future<void> setDefaultModel(String model) => 
    _storage.setString('default_model', model);

  /// 获取温度参数
  double get temperature => 
    _storage.getDouble('temperature') ?? 0.7;

  /// 设置温度参数
  Future<void> setTemperature(double value) => 
    _storage.setDouble('temperature', value);

  /// 获取最大 Token
  int get maxTokens => 
    _storage.getInt('max_tokens') ?? 2048;

  /// 设置最大 Token
  Future<void> setMaxTokens(int value) => 
    _storage.setInt('max_tokens', value);

  // ============== 主题配置 ==============

  /// 获取主题模式
  ThemeMode get themeMode {
    final value = _storage.getString('theme_mode') ?? 'system';
    switch (value) {
      case 'light':
        return ThemeMode.light;
      case 'dark':
        return ThemeMode.dark;
      default:
        return ThemeMode.system;
    }
  }

  /// 设置主题模式
  Future<void> setThemeMode(ThemeMode mode) {
    String value;
    switch (mode) {
      case ThemeMode.light:
        value = 'light';
        break;
      case ThemeMode.dark:
        value = 'dark';
        break;
      case ThemeMode.system:
        value = 'system';
        break;
    }
    return _storage.setString('theme_mode', value);
  }

  // ============== 语言配置 ==============

  /// 获取语言
  Locale get locale {
    final code = _storage.getString('language') ?? 'zh';
    if (code.contains('_')) {
      final parts = code.split('_');
      return Locale(parts[0], parts[1]);
    }
    return Locale(code);
  }

  /// 设置语言
  Future<void> setLocale(Locale locale) {
    return _storage.setString('language', locale.languageCode);
  }

  // ============== 发现服务配置 ==============

  /// 获取发现服务是否启用
  bool get isDiscoveryEnabled => _storage.getBool('discovery_enabled') ?? false;

  /// 设置发现服务
  Future<void> setDiscoveryEnabled(bool enabled) =>
    _storage.setBool('discovery_enabled', enabled);

  /// 获取发现端口
  int get discoveryPort =>
    _storage.getInt('discovery_port') ?? 18792;

  /// 设置发现端口
  Future<void> setDiscoveryPort(int port) =>
    _storage.setInt('discovery_port', port);

  // ============== 连接配置 ==============

  /// 获取自动连接状态
  bool get autoConnect =>
    _storage.getBool('auto_connect') ?? true;

  /// 设置自动连接状态
  Future<void> setAutoConnect(bool enabled) =>
    _storage.setBool('auto_connect', enabled);

  // ============== 全部配置 ==============

  /// 导出配置
  Map<String, dynamic> exportConfig() {
    return {
      'api_url': apiUrl,
      'default_model': defaultModel,
      'temperature': temperature,
      'max_tokens': maxTokens,
      'theme_mode': themeMode.name,
      'language': locale.toString(),
      'auto_connect': autoConnect,
      'discovery_enabled': isDiscoveryEnabled,
      'discovery_port': discoveryPort,
    };
  }

  /// 导入配置
  Future<void> importConfig(Map<String, dynamic> config) async {
    if (config['api_url'] != null) await setApiUrl(config['api_url'] as String);
    if (config['default_model'] != null) await setDefaultModel(config['default_model'] as String);
    if (config['temperature'] != null) await setTemperature((config['temperature'] as num).toDouble());
    if (config['max_tokens'] != null) await setMaxTokens((config['max_tokens'] as num).toInt());
    if (config['theme_mode'] != null) {
      await setThemeMode(ThemeMode.values.firstWhere(
        (e) => e.name == config['theme_mode'],
        orElse: () => ThemeMode.system,
      ));
    }
    if (config['auto_connect'] != null) await setAutoConnect(config['auto_connect'] as bool);
    if (config['discovery_enabled'] != null) await setDiscoveryEnabled(config['discovery_enabled'] as bool);
    if (config['discovery_port'] != null) await setDiscoveryPort((config['discovery_port'] as num).toInt());
  }

  /// 重置为默认配置
  Future<void> resetToDefaults() async {
    await _storage.clear();
  }
}
