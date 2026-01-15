import 'dart:async';

/// 通知服务
class NotificationService {
  static NotificationService? _instance;
  final _notificationController = StreamController<Notification>.broadcast();
  final List<Notification> _notificationHistory = [];
  
  NotificationService._();
  
  static NotificationService get instance {
    _instance ??= NotificationService._();
    return _instance!;
  }
  
  /// 通知流
  Stream<Notification> get onNotification => _notificationController.stream;
  
  /// 历史记录
  List<Notification> get history => List.unmodifiable(_notificationHistory);
  
  /// 发送通知
  void notify({
    required String title,
    required String body,
    NotificationType type = NotificationType.info,
    String? actionUrl,
    Duration? duration,
  }) {
    final notification = Notification(
      id: DateTime.now().millisecondsSinceEpoch.toString(),
      title: title,
      body: body,
      type: type,
      actionUrl: actionUrl,
      timestamp: DateTime.now(),
    );
    
    _notificationHistory.insert(0, notification);
    _notificationController.add(notification);
    
    // 自动清除
    if (duration != null) {
      Future.delayed(duration, () => dismiss(notification.id));
    }
    
    // 限制历史记录数量
    if (_notificationHistory.length > 100) {
      _notificationHistory.removeLast();
    }
  }
  
  /// 成功通知
  void success(String title, String body) {
    notify(title: title, body: body, type: NotificationType.success);
  }
  
  /// 错误通知
  void error(String title, String body) {
    notify(title: title, body: body, type: NotificationType.error);
  }
  
  /// 警告通知
  void warning(String title, String body) {
    notify(title: title, body: body, type: NotificationType.warning);
  }
  
  /// 信息通知
  void info(String title, String body) {
    notify(title: title, body: body, type: NotificationType.info);
  }
  
  /// 清除通知
  void dismiss(String id) {
    _notificationHistory.removeWhere((n) => n.id == id);
  }
  
  /// 清除所有通知
  void dismissAll() {
    _notificationHistory.clear();
  }
  
  void dispose() {
    _notificationController.close();
  }
}

/// 通知模型
class Notification {
  final String id;
  final String title;
  final String body;
  final NotificationType type;
  final String? actionUrl;
  final DateTime timestamp;
  bool isRead;

  Notification({
    required this.id,
    required this.title,
    required this.body,
    required this.type,
    this.actionUrl,
    required this.timestamp,
    this.isRead = false,
  });
}

/// 通知类型
enum NotificationType {
  info,
  success,
  warning,
  error,
}
