// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'chat_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

String _$sessionsNotifierHash() => r'79c82b40895752c68173e86e839c0fc2c5ba415a';

/// 会话列表Provider
///
/// Copied from [SessionsNotifier].
@ProviderFor(SessionsNotifier)
final sessionsNotifierProvider =
    AutoDisposeNotifierProvider<SessionsNotifier, List<ChatSession>>.internal(
  SessionsNotifier.new,
  name: r'sessionsNotifierProvider',
  debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
      ? null
      : _$sessionsNotifierHash,
  dependencies: null,
  allTransitiveDependencies: null,
);

typedef _$SessionsNotifier = AutoDisposeNotifier<List<ChatSession>>;
String _$activeSessionHash() => r'a3ee526867c49edbaa48581508db9343a1f38293';

/// 当前活动会话ID
///
/// Copied from [ActiveSession].
@ProviderFor(ActiveSession)
final activeSessionProvider =
    AutoDisposeNotifierProvider<ActiveSession, String?>.internal(
  ActiveSession.new,
  name: r'activeSessionProvider',
  debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
      ? null
      : _$activeSessionHash,
  dependencies: null,
  allTransitiveDependencies: null,
);

typedef _$ActiveSession = AutoDisposeNotifier<String?>;
String _$messagesNotifierHash() => r'7a11c36ce141a0dbb424d5bc5d29431317e487f3';

/// 消息列表Provider
///
/// Copied from [MessagesNotifier].
@ProviderFor(MessagesNotifier)
final messagesNotifierProvider = AutoDisposeNotifierProvider<MessagesNotifier,
    Map<String, List<ChatMessage>>>.internal(
  MessagesNotifier.new,
  name: r'messagesNotifierProvider',
  debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
      ? null
      : _$messagesNotifierHash,
  dependencies: null,
  allTransitiveDependencies: null,
);

typedef _$MessagesNotifier
    = AutoDisposeNotifier<Map<String, List<ChatMessage>>>;
String _$chatStatusNotifierHash() =>
    r'e4fc2dd793f5de1a03e28531b30f9df5dbd13bce';

/// 聊天状态Provider
///
/// Copied from [ChatStatusNotifier].
@ProviderFor(ChatStatusNotifier)
final chatStatusNotifierProvider =
    AutoDisposeNotifierProvider<ChatStatusNotifier, ChatStatus>.internal(
  ChatStatusNotifier.new,
  name: r'chatStatusNotifierProvider',
  debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
      ? null
      : _$chatStatusNotifierHash,
  dependencies: null,
  allTransitiveDependencies: null,
);

typedef _$ChatStatusNotifier = AutoDisposeNotifier<ChatStatus>;
String _$aIServiceNotifierHash() => r'1f3d6ce7ca75199c010eeaf52c87b599a85453a1';

/// AI服务Provider
///
/// Copied from [AIServiceNotifier].
@ProviderFor(AIServiceNotifier)
final aIServiceNotifierProvider =
    AutoDisposeNotifierProvider<AIServiceNotifier, AIService?>.internal(
  AIServiceNotifier.new,
  name: r'aIServiceNotifierProvider',
  debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
      ? null
      : _$aIServiceNotifierHash,
  dependencies: null,
  allTransitiveDependencies: null,
);

typedef _$AIServiceNotifier = AutoDisposeNotifier<AIService?>;
String _$chatServiceNotifierHash() =>
    r'920d07b320fc77a4bce3300801a3299950cfc825';

/// 聊天服务Provider
///
/// Copied from [ChatServiceNotifier].
@ProviderFor(ChatServiceNotifier)
final chatServiceNotifierProvider =
    AutoDisposeNotifierProvider<ChatServiceNotifier, ChatService>.internal(
  ChatServiceNotifier.new,
  name: r'chatServiceNotifierProvider',
  debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
      ? null
      : _$chatServiceNotifierHash,
  dependencies: null,
  allTransitiveDependencies: null,
);

typedef _$ChatServiceNotifier = AutoDisposeNotifier<ChatService>;
// ignore_for_file: type=lint
// ignore_for_file: subtype_of_sealed_class, invalid_use_of_internal_member, invalid_use_of_visible_for_testing_member
