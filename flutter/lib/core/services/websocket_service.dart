import 'dart:async';
import 'dart:convert';
import 'package:web_socket_channel/web_socket_channel.dart';

/// WebSocket 服务
class WebSocketService {
  static WebSocketService? _instance;
  
  WebSocketChannel? _channel;
  final _messageController = StreamController<Map<String, dynamic>>.broadcast();
  bool _isConnected = false;
  String? _baseUrl;
  String? _apiKey;
  
  WebSocketService._();
  
  static WebSocketService get instance {
    _instance ??= WebSocketService._();
    return _instance!;
  }
  
  /// 初始化
  void init({required String baseUrl, required String apiKey}) {
    _baseUrl = baseUrl.replaceAll('http', 'ws');
    _apiKey = apiKey;
  }
  
  /// 连接状态
  bool get isConnected => _isConnected;
  
  /// 消息流
  Stream<Map<String, dynamic>> get messages => _messageController.stream;
  
  /// 连接
  Future<void> connect() async {
    if (_channel != null) return;
    
    final wsUrl = '$_baseUrl/ws';
    _channel = WebSocketChannel.connect(
      Uri.parse(wsUrl),
      protocols: ['Bearer $_apiKey'],
    );
    
    _channel!.stream.listen(
      (data) => _handleMessage(data),
      onError: (error) => _handleError(error),
      onDone: () => _handleDisconnect(),
    );
    
    _isConnected = true;
  }
  
  /// 断开连接
  Future<void> disconnect() async {
    await _channel?.sink.close();
    _channel = null;
    _isConnected = false;
  }
  
  /// 发送消息
  void send(Map<String, dynamic> message) {
    if (_channel == null) return;
    _channel!.sink.add(jsonEncode(message));
  }
  
  /// 发送聊天消息
  void sendChat(String sessionId, String content) {
    send({
      'type': 'chat',
      'payload': {
        'session_id': sessionId,
        'content': content,
      },
    });
  }
  
  /// 发送认证
  void authenticate(String userId) {
    send({
      'type': 'auth',
      'payload': {'user_id': userId},
    });
  }
  
  /// 发送心跳
  void ping() {
    send({'type': 'ping'});
  }
  
  void _handleMessage(dynamic data) {
    try {
      final message = jsonDecode(data as String) as Map<String, dynamic>;
      _messageController.add(message);
    } catch (e) {
      // 忽略解析错误
    }
  }
  
  void _handleError(dynamic error) {
    _isConnected = false;
  }
  
  void _handleDisconnect() {
    _isConnected = false;
    _channel = null;
  }
  
  void dispose() {
    _messageController.close();
  }
}
