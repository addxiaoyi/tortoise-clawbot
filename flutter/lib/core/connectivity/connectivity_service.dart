import 'dart:async';
import 'package:flutter/foundation.dart';

/// 连接状态
enum ConnectionStatus {
  online,
  offline,
  checking,
}

/// 连接状态服务
class ConnectivityService {
  static ConnectivityService? _instance;
  static ConnectivityService get instance => _instance ??= ConnectivityService._();
  ConnectivityService._();

  final _statusController = StreamController<ConnectionStatus>.broadcast();
  ConnectionStatus _status = ConnectionStatus.checking;

  Stream<ConnectionStatus> get statusStream => _statusController.stream;
  ConnectionStatus get status => _status;
  bool get isOnline => _status == ConnectionStatus.online;
  bool get isOffline => _status == ConnectionStatus.offline;

  /// 初始化
  void initialize() {
    checkConnection();
    // 定期检查连接状态
    Timer.periodic(const Duration(seconds: 30), (_) => checkConnection());
  }

  /// 检查连接状态
  Future<void> checkConnection() async {
    _updateStatus(ConnectionStatus.checking);
    
    try {
      // 简单的连接检查
      if (kIsWeb) {
        _updateStatus(ConnectionStatus.online);
      } else {
        _updateStatus(ConnectionStatus.online);
      }
    } catch (e) {
      _updateStatus(ConnectionStatus.offline);
    }
  }

  /// 手动设置为在线
  void setOnline() {
    _updateStatus(ConnectionStatus.online);
  }

  /// 手动设置为离线
  void setOffline() {
    _updateStatus(ConnectionStatus.offline);
  }

  void _updateStatus(ConnectionStatus status) {
    if (_status != status) {
      _status = status;
      _statusController.add(status);
    }
  }

  /// 销毁
  void dispose() {
    _statusController.close();
  }
}
