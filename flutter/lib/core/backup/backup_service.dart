import 'dart:convert';
import '../storage/storage_service.dart';
import '../session/session_manager.dart';
import '../config/config_manager.dart';

/// 备份数据
class BackupData {
  final DateTime createdAt;
  final String version;
  final Map<String, dynamic> sessions;
  final Map<String, dynamic> config;
  final Map<String, dynamic> settings;

  BackupData({
    required this.createdAt,
    required this.version,
    required this.sessions,
    required this.config,
    required this.settings,
  });

  Map<String, dynamic> toJson() => {
    'createdAt': createdAt.toIso8601String(),
    'version': version,
    'sessions': sessions,
    'config': config,
    'settings': settings,
  };

  factory BackupData.fromJson(Map<String, dynamic> json) {
    return BackupData(
      createdAt: DateTime.parse(json['createdAt'] as String),
      version: json['version'] as String,
      sessions: _castMap(json['sessions']),
      config: _castMap(json['config']),
      settings: _castMap(json['settings']),
    );
  }
  
  static Map<String, dynamic> _castMap(dynamic value) {
    if (value == null) return {};
    if (value is Map) return Map<String, dynamic>.from(value);
    return {};
  }
}

/// 备份服务
class BackupService {
  static BackupService? _instance;
  static BackupService get instance => _instance ??= BackupService._();
  BackupService._();

  final StorageService _storage = StorageService.instance;
  static const String _currentVersion = '1.0.0';

  /// 创建备份
  Future<BackupData> createBackup() async {
    final sessions = <String, dynamic>{};
    final sessionManager = SessionManager.instance;
    
    for (final session in sessionManager.sessions) {
      sessions[session.id] = session.toJson();
    }

    final configManager = ConfigManager.instance;

    return BackupData(
      createdAt: DateTime.now(),
      version: _currentVersion,
      sessions: sessions,
      config: <String, dynamic>{
        'themeMode': configManager.themeMode.name,
        'locale': configManager.locale.languageCode,
        'model': _storage.getString('ai_model') ?? 'gpt-4',
      },
      settings: <String, dynamic>{},
    );
  }

  /// 导出为 JSON 字符串
  Future<String> exportToJson() async {
    final backup = await createBackup();
    return const JsonEncoder.withIndent('  ').convert(backup.toJson());
  }

  /// 从 JSON 字符串导入
  Future<void> importFromJson(String jsonString) async {
    try {
      final json = jsonDecode(jsonString) as Map<String, dynamic>;
      final backup = BackupData.fromJson(json);

      if (!_isVersionCompatible(backup.version)) {
        throw BackupException('版本不兼容: ${backup.version}');
      }

      await _restoreSessions(backup.sessions);
      await _restoreConfig(backup.config);

    } catch (e) {
      throw BackupException('导入失败: $e');
    }
  }

  bool _isVersionCompatible(String version) {
    return version.startsWith('1.');
  }

  Future<void> _restoreSessions(Map<String, dynamic> sessions) async {
    for (final entry in sessions.entries) {
      final value = entry.value;
      if (value is Map) {
        await _storage.saveSession(entry.key, Map<String, dynamic>.from(value));
      }
    }
  }

  Future<void> _restoreConfig(Map<String, dynamic> config) async {
    if (config['themeMode'] != null) {
      await _storage.setString('theme_mode', config['themeMode'].toString());
    }
    if (config['locale'] != null) {
      await _storage.setString('locale', config['locale'].toString());
    }
    if (config['model'] != null) {
      await _storage.setString('ai_model', config['model'].toString());
    }
  }

  List<BackupInfo> listBackups() {
    return [];
  }
}

class BackupInfo {
  final String id;
  final DateTime createdAt;
  final String version;
  final int size;

  BackupInfo({
    required this.id,
    required this.createdAt,
    required this.version,
    required this.size,
  });
}

class BackupException implements Exception {
  final String message;
  BackupException(this.message);

  @override
  String toString() => 'BackupException: $message';
}
