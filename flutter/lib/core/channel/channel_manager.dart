import 'package:flutter/foundation.dart';
import 'dart:async';

/// 渠道管理器 - 统一管理所有消息渠道
class ChannelManager {
  final Map<String, Channel> _channels = {};
  final _eventController = StreamController<ChannelEvent>.broadcast();
  
  Stream<ChannelEvent> get events => _eventController.stream;
  
  void register(Channel channel) {
    _channels[channel.id] = channel;
    channel.events.listen((event) {
      _eventController.add(event.copyWith(channelId: channel.id));
    });
  }

  void registerTelegram({required String token, Set<int>? allowedChats}) {
    register(TelegramChannel(botToken: token, allowedChats: allowedChats));
  }

  void registerDiscord({required String botToken, String? guildId}) {
    register(DiscordChannel(botToken: botToken, guildId: guildId));
  }

  Channel? getChannel(String id) => _channels[id];

  Future<void> connectAll() async {
    for (final channel in _channels.values) {
      if (channel is ConnectableChannel) {
        try {
          await channel.connect();
        } catch (e) {
          debugPrint('Channel ${channel.id} 连接失败: $e');
        }
      }
    }
  }

  Future<void> disconnectAll() async {
    for (final channel in _channels.values) {
      if (channel is ConnectableChannel) {
        await channel.disconnect();
      }
    }
  }

  Map<String, ChannelStatus> getStatus() {
    return _channels.map((id, channel) => MapEntry(id, channel.status));
  }

  void dispose() {
    disconnectAll();
    _eventController.close();
    for (final channel in _channels.values) {
      channel.dispose();
    }
    _channels.clear();
  }
}

/// 渠道基类
abstract class Channel {
  String get id;
  String get type;
  String get name;
  ChannelStatus get status;
  Stream<ChannelEvent> get events;
  void dispose();
}

/// 可连接的渠道
abstract class ConnectableChannel implements Channel {
  Future<void> connect();
  Future<void> disconnect();
}

/// 渠道状态
enum ChannelStatus {
  disconnected,
  connecting,
  connected,
  error,
  reconnecting,
}

/// 渠道事件
class ChannelEvent {
  final String? channelId;
  final ChannelEventType type;
  final dynamic data;
  final String? error;
  final DateTime timestamp;

  ChannelEvent({
    this.channelId,
    required this.type,
    this.data,
    this.error,
    DateTime? timestamp,
  }) : timestamp = timestamp ?? DateTime.now();

  ChannelEvent copyWith({
    String? channelId,
    ChannelEventType? type,
    dynamic data,
    String? error,
  }) {
    return ChannelEvent(
      channelId: channelId ?? this.channelId,
      type: type ?? this.type,
      data: data ?? this.data,
      error: error ?? this.error,
      timestamp: timestamp,
    );
  }
}

enum ChannelEventType {
  connected,
  disconnected,
  message,
  error,
}

/// Telegram 渠道
class TelegramChannel implements ConnectableChannel {
  final String botToken;
  final Set<int>? allowedChats;
  
  @override
  final String id;
  @override
  String get type => 'telegram';
  @override
  String get name => 'Telegram';
  
  ChannelStatus _status = ChannelStatus.disconnected;
  final _controller = StreamController<ChannelEvent>.broadcast();

  TelegramChannel({required this.botToken, this.allowedChats}) 
      : id = 'telegram_$botToken';

  @override
  ChannelStatus get status => _status;
  
  @override
  Stream<ChannelEvent> get events => _controller.stream;

  void _updateStatus(ChannelStatus status) {
    _status = status;
  }

  @override
  Future<void> connect() async {
    _updateStatus(ChannelStatus.connecting);
    _updateStatus(ChannelStatus.connected);
    _controller.add(ChannelEvent(type: ChannelEventType.connected));
  }

  @override
  Future<void> disconnect() async {
    _updateStatus(ChannelStatus.disconnected);
    _controller.add(ChannelEvent(type: ChannelEventType.disconnected));
  }

  @override
  void dispose() {
    _controller.close();
  }
}

/// Discord 渠道
class DiscordChannel implements ConnectableChannel {
  final String botToken;
  final String? guildId;
  
  @override
  final String id;
  @override
  String get type => 'discord';
  @override
  String get name => 'Discord';
  
  ChannelStatus _status = ChannelStatus.disconnected;
  final _controller = StreamController<ChannelEvent>.broadcast();

  DiscordChannel({required this.botToken, this.guildId}) 
      : id = 'discord_$botToken';

  @override
  ChannelStatus get status => _status;
  
  @override
  Stream<ChannelEvent> get events => _controller.stream;

  void _updateStatus(ChannelStatus status) {
    _status = status;
  }

  @override
  Future<void> connect() async {
    _updateStatus(ChannelStatus.connecting);
    _updateStatus(ChannelStatus.connected);
    _controller.add(ChannelEvent(type: ChannelEventType.connected));
  }

  @override
  Future<void> disconnect() async {
    _updateStatus(ChannelStatus.disconnected);
    _controller.add(ChannelEvent(type: ChannelEventType.disconnected));
  }

  @override
  void dispose() {
    _controller.close();
  }
}
