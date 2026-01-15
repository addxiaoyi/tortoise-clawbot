/// 聊天消息模型
class ChatMessage {
  final String id;
  final String sessionId;
  final String role;
  final String content;
  final DateTime createdAt;
  final String? model;
  
  ChatMessage({
    required this.id,
    required this.sessionId,
    required this.role,
    required this.content,
    required this.createdAt,
    this.model,
  });
  
  factory ChatMessage.fromJson(Map<String, dynamic> json) {
    return ChatMessage(
      id: json['id']?.toString() ?? '',
      sessionId: json['sessionId']?.toString() ?? '',
      role: json['role']?.toString() ?? 'user',
      content: json['content']?.toString() ?? '',
      createdAt: json['createdAt'] != null 
        ? DateTime.parse(json['createdAt'].toString()) 
        : DateTime.now(),
      model: json['model']?.toString(),
    );
  }
  
  Map<String, dynamic> toJson() => {
    'id': id,
    'sessionId': sessionId,
    'role': role,
    'content': content,
    'createdAt': createdAt.toIso8601String(),
    if (model != null) 'model': model,
  };
}

/// 聊天会话模型
class ChatSession {
  final String id;
  final String title;
  final List<ChatMessage> messages;
  final DateTime createdAt;
  final DateTime updatedAt;
  final String? model;
  final String? channel;
  
  ChatSession({
    required this.id,
    required this.title,
    required this.messages,
    required this.createdAt,
    required this.updatedAt,
    this.model,
    this.channel,
  });
  
  ChatSession copyWith({
    String? id,
    String? title,
    List<ChatMessage>? messages,
    DateTime? createdAt,
    DateTime? updatedAt,
    String? model,
    String? channel,
  }) {
    return ChatSession(
      id: id ?? this.id,
      title: title ?? this.title,
      messages: messages ?? this.messages,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
      model: model ?? this.model,
      channel: channel ?? this.channel,
    );
  }
  
  factory ChatSession.fromJson(Map<String, dynamic> json) {
    return ChatSession(
      id: json['id']?.toString() ?? '',
      title: json['title']?.toString() ?? '新对话',
      messages: (json['messages'] as List?)
          ?.map((m) => ChatMessage.fromJson(m as Map<String, dynamic>))
          .toList() ?? [],
      createdAt: json['createdAt'] != null 
        ? DateTime.parse(json['createdAt'].toString()) 
        : DateTime.now(),
      updatedAt: json['updatedAt'] != null 
        ? DateTime.parse(json['updatedAt'].toString()) 
        : DateTime.now(),
      model: json['model']?.toString(),
      channel: json['channel']?.toString(),
    );
  }
  
  Map<String, dynamic> toJson() => {
    'id': id,
    'title': title,
    'messages': messages.map((m) => m.toJson()).toList(),
    'createdAt': createdAt.toIso8601String(),
    'updatedAt': updatedAt.toIso8601String(),
    if (model != null) 'model': model,
    if (channel != null) 'channel': channel,
  };
}

/// AI提供商模型
class AIProvider {
  final String id;
  final String name;
  final String apiKey;
  final String baseUrl;
  final bool isActive;
  
  AIProvider({
    required this.id,
    required this.name,
    required this.apiKey,
    required this.baseUrl,
    this.isActive = true,
  });
  
  factory AIProvider.fromJson(Map<String, dynamic> json) {
    return AIProvider(
      id: json['id']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      apiKey: json['apiKey']?.toString() ?? '',
      baseUrl: json['baseUrl']?.toString() ?? '',
      isActive: json['isActive'] as bool? ?? true,
    );
  }
  
  Map<String, dynamic> toJson() => {
    'id': id,
    'name': name,
    'apiKey': apiKey,
    'baseUrl': baseUrl,
    'isActive': isActive,
  };
}
