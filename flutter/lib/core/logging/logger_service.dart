import 'package:flutter/foundation.dart';

/// 日志级别
enum LogLevel {
  debug,
  info,
  warning,
  error,
}

/// 日志条目
class LogEntry {
  final LogLevel level;
  final String message;
  final String? tag;
  final DateTime timestamp;
  final Object? error;
  final StackTrace? stackTrace;

  LogEntry({
    required this.level,
    required this.message,
    this.tag,
    DateTime? timestamp,
    this.error,
    this.stackTrace,
  }) : timestamp = timestamp ?? DateTime.now();

  @override
  String toString() {
    final buffer = StringBuffer();
    buffer.write('[$timestamp] ');
    buffer.write('[${level.name.toUpperCase()}]');
    if (tag != null) buffer.write(' [$tag]');
    buffer.write(' $message');
    if (error != null) buffer.write(' - $error');
    return buffer.toString();
  }
}

/// 日志服务
class LoggerService {
  static LoggerService? _instance;
  static LoggerService get instance => _instance ??= LoggerService._();
  LoggerService._();

  final List<LogEntry> _logs = [];
  final List<void Function(LogEntry)> _listeners = [];
  LogLevel _minLevel = LogLevel.debug;
  bool _enableConsole = true;
  int _maxLogs = 1000;

  LogLevel get minLevel => _minLevel;
  List<LogEntry> get logs => List.unmodifiable(_logs);

  /// 初始化
  void initialize({LogLevel minLevel = LogLevel.debug, bool enableConsole = true, int maxLogs = 1000}) {
    _minLevel = minLevel;
    _enableConsole = enableConsole;
    _maxLogs = maxLogs;
    info('LoggerService initialized', tag: 'Logger');
  }

  /// 设置最小日志级别
  void setMinLevel(LogLevel level) {
    _minLevel = level;
  }

  /// 添加监听器
  void addListener(void Function(LogEntry) listener) {
    _listeners.add(listener);
  }

  /// 移除监听器
  void removeListener(void Function(LogEntry) listener) {
    _listeners.remove(listener);
  }

  /// 记录调试日志
  void debug(String message, {String? tag, Object? error, StackTrace? stackTrace}) {
    _log(LogLevel.debug, message, tag: tag, error: error, stackTrace: stackTrace);
  }

  /// 记录信息日志
  void info(String message, {String? tag, Object? error, StackTrace? stackTrace}) {
    _log(LogLevel.info, message, tag: tag, error: error, stackTrace: stackTrace);
  }

  /// 记录警告日志
  void warning(String message, {String? tag, Object? error, StackTrace? stackTrace}) {
    _log(LogLevel.warning, message, tag: tag, error: error, stackTrace: stackTrace);
  }

  /// 记录错误日志
  void error(String message, {String? tag, Object? error, StackTrace? stackTrace}) {
    _log(LogLevel.error, message, tag: tag, error: error, stackTrace: stackTrace);
  }

  void _log(LogLevel level, String message, {String? tag, Object? error, StackTrace? stackTrace}) {
    if (level.index < _minLevel.index) return;

    final entry = LogEntry(
      level: level,
      message: message,
      tag: tag,
      error: error,
      stackTrace: stackTrace,
    );

    _logs.add(entry);
    if (_logs.length > _maxLogs) {
      _logs.removeAt(0);
    }

    if (_enableConsole) {
      _printToConsole(entry);
    }

    for (final listener in _listeners) {
      listener(entry);
    }
  }

  void _printToConsole(LogEntry entry) {
    switch (entry.level) {
      case LogLevel.debug:
        debugPrint(entry.toString());
        break;
      case LogLevel.info:
        debugPrint(entry.toString());
        break;
      case LogLevel.warning:
        debugPrint('[WARNING] ${entry.toString()}');
        break;
      case LogLevel.error:
        debugPrint('[ERROR] ${entry.toString()}');
        if (entry.stackTrace != null) {
          debugPrintStack(stackTrace: entry.stackTrace);
        }
        break;
    }
  }

  /// 清空日志
  void clear() {
    _logs.clear();
  }

  /// 导出日志
  String exportLogs() {
    return _logs.map((e) => e.toString()).join('\n');
  }
}
