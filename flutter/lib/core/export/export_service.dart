import 'dart:convert';
import '../session/session_manager.dart';

/// 导出格式
enum ExportFormat {
  json,
  markdown,
  text,
  csv,
}

/// 导出服务
class ExportService {
  static ExportService? _instance;
  static ExportService get instance => _instance ??= ExportService._();
  ExportService._();

  /// 导出所有会话
  String export(ExportFormat format) {
    final sessionManager = SessionManager.instance;
    final sessions = sessionManager.sessions;
    
    switch (format) {
      case ExportFormat.json:
        return _exportJson(sessions.cast<Map<String, dynamic>>());
      case ExportFormat.markdown:
        return _exportMarkdown(sessions);
      case ExportFormat.text:
        return _exportText(sessions);
      case ExportFormat.csv:
        return _exportCsv(sessions);
    }
  }

  String _exportJson(List<Map<String, dynamic>> sessions) {
    return const JsonEncoder.withIndent('  ').convert(sessions);
  }

  String _exportMarkdown(List sessions) {
    final buffer = StringBuffer();
    buffer.writeln('# Tortoise 会话导出');
    buffer.writeln();
    buffer.writeln('导出时间: ${DateTime.now()}');
    buffer.writeln();
    
    for (final session in sessions) {
      buffer.writeln(_sessionToMarkdown(session));
      buffer.writeln();
    }
    
    return buffer.toString();
  }

  String _exportText(List sessions) {
    final buffer = StringBuffer();
    for (final session in sessions) {
      buffer.writeln('=== ${session.title} ===');
      buffer.writeln();
      for (final message in (session.messages as List)) {
        final role = message.role == 'user' ? '用户' : 'AI';
        buffer.writeln('[$role] ${message.content}');
        buffer.writeln();
      }
    }
    return buffer.toString();
  }

  String _exportCsv(List sessions) {
    final buffer = StringBuffer();
    buffer.writeln('Session,Role,Content,Timestamp');
    
    for (final session in sessions) {
      for (final message in (session.messages as List)) {
        buffer.writeln('"${session.title}","${message.role}","${message.content.replaceAll('"', '""')}","${session.updatedAt}"',
        );
      }
    }
    
    return buffer.toString();
  }

  String _sessionToMarkdown(dynamic session) {
    final buffer = StringBuffer();
    buffer.writeln('## ${session.title}');
    buffer.writeln();
    buffer.writeln('创建时间: ${session.createdAt}');
    buffer.writeln();
    
    for (final message in (session.messages as List)) {
      final role = message.role == 'user' ? '👤 用户' : '🤖 AI';
      buffer.writeln('### $role');
      buffer.writeln();
      buffer.writeln(message.content);
      buffer.writeln();
    }
    
    return buffer.toString();
  }
}
