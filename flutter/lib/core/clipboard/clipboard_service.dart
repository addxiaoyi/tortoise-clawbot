import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';

/// 剪贴板服务
class ClipboardService {
  static ClipboardService? _instance;
  static ClipboardService get instance => _instance ??= ClipboardService._();
  ClipboardService._();

  /// 复制到剪贴板
  Future<bool> copy(String text) async {
    try {
      await Clipboard.setData(ClipboardData(text: text));
      debugPrint('Copied to clipboard: ${text.length} chars');
      return true;
    } catch (e) {
      debugPrint('Failed to copy to clipboard: $e');
      return false;
    }
  }

  /// 从剪贴板获取
  Future<String?> paste() async {
    try {
      final data = await Clipboard.getData(Clipboard.kTextPlain);
      return data?.text;
    } catch (e) {
      debugPrint('Failed to paste from clipboard: $e');
      return null;
    }
  }

  /// 检查剪贴板是否有内容
  Future<bool> hasContent() async {
    try {
      final data = await Clipboard.getData(Clipboard.kTextPlain);
      return data?.text?.isNotEmpty ?? false;
    } catch (e) {
      return false;
    }
  }

  /// 清空剪贴板
  Future<void> clear() async {
    try {
      await Clipboard.setData(const ClipboardData(text: ''));
    } catch (e) {
      debugPrint('Failed to clear clipboard: $e');
    }
  }
}
