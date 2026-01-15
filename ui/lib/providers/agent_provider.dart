import 'package:flutter/material.dart';

class AgentProvider extends ChangeNotifier {
  bool _isConnected = true;
  String _status = 'Active';

  bool get isConnected => _isConnected;
  String get status => _status;

  void setStatus(String status) {
    _status = status;
    notifyListeners();
  }

  void setConnected(bool connected) {
    _isConnected = connected;
    notifyListeners();
  }
}
