// 数据模型

class AgentModel {
  final String id;
  final String name;
  final String model;
  final bool isActive;
  final DateTime? lastActive;

  AgentModel({
    required this.id,
    required this.name,
    required this.model,
    this.isActive = false,
    this.lastActive,
  });

  factory AgentModel.fromJson(Map<String, dynamic> json) {
    return AgentModel(
      id: json['id'] ?? '',
      name: json['name'] ?? '',
      model: json['model'] ?? '',
      isActive: json['is_active'] ?? false,
      lastActive: json['last_active'] != null 
          ? DateTime.parse(json['last_active']) 
          : null,
    );
  }

  Map<String, dynamic> toJson() => {
    'id': id,
    'name': name,
    'model': model,
    'is_active': isActive,
    'last_active': lastActive?.toIso8601String(),
  };
}

enum ChannelType {
  discord,
  telegram,
  whatsapp,
  slack,
  email,
  matrix,
  signal,
}

class ChannelModel {
  final String id;
  final String name;
  final ChannelType type;
  final bool isConnected;
  final DateTime? lastMessage;

  ChannelModel({
    required this.id,
    required this.name,
    required this.type,
    this.isConnected = false,
    this.lastMessage,
  });

  factory ChannelModel.fromJson(Map<String, dynamic> json) {
    return ChannelModel(
      id: json['id'] ?? '',
      name: json['name'] ?? '',
      type: ChannelType.values.firstWhere(
        (e) => e.name == json['type'],
        orElse: () => ChannelType.discord,
      ),
      isConnected: json['is_connected'] ?? false,
      lastMessage: json['last_message'] != null
          ? DateTime.parse(json['last_message'])
          : null,
    );
  }
}

enum PluginType {
  channel,
  skill,
  tool,
  integration,
}

class PluginModel {
  final String id;
  final String name;
  final String description;
  final String version;
  final bool isEnabled;
  final PluginType type;

  PluginModel({
    required this.id,
    required this.name,
    required this.description,
    required this.version,
    this.isEnabled = false,
    required this.type,
  });

  factory PluginModel.fromJson(Map<String, dynamic> json) {
    return PluginModel(
      id: json['id'] ?? '',
      name: json['name'] ?? '',
      description: json['description'] ?? '',
      version: json['version'] ?? '1.0.0',
      isEnabled: json['is_enabled'] ?? false,
      type: PluginType.values.firstWhere(
        (e) => e.name == json['type'],
        orElse: () => PluginType.tool,
      ),
    );
  }
}

enum MessageStatus {
  sending,
  sent,
  delivered,
  read,
  error,
}

class MessageModel {
  final String id;
  final String content;
  final bool isUser;
  final DateTime timestamp;
  final MessageStatus status;
  final List<String>? attachments;

  MessageModel({
    required this.id,
    required this.content,
    required this.isUser,
    required this.timestamp,
    this.status = MessageStatus.sent,
    this.attachments,
  });

  factory MessageModel.fromJson(Map<String, dynamic> json) {
    return MessageModel(
      id: json['id'] ?? '',
      content: json['content'] ?? '',
      isUser: json['is_user'] ?? false,
      timestamp: DateTime.parse(json['timestamp'] ?? DateTime.now().toIso8601String()),
      status: MessageStatus.values.firstWhere(
        (e) => e.name == json['status'],
        orElse: () => MessageStatus.sent,
      ),
      attachments: json['attachments'] != null
          ? List<String>.from(json['attachments'])
          : null,
    );
  }
}

enum MemoryType {
  shortTerm,
  mediumTerm,
  longTerm,
}

class MemoryModel {
  final String id;
  final String content;
  final MemoryType type;
  final double importance;
  final DateTime createdAt;
  final int accessCount;

  MemoryModel({
    required this.id,
    required this.content,
    required this.type,
    required this.importance,
    required this.createdAt,
    this.accessCount = 0,
  });

  factory MemoryModel.fromJson(Map<String, dynamic> json) {
    return MemoryModel(
      id: json['id'] ?? '',
      content: json['content'] ?? '',
      type: MemoryType.values.firstWhere(
        (e) => e.name == json['type'],
        orElse: () => MemoryType.shortTerm,
      ),
      importance: (json['importance'] ?? 0.5).toDouble(),
      createdAt: DateTime.parse(json['created_at'] ?? DateTime.now().toIso8601String()),
      accessCount: json['access_count'] ?? 0,
    );
  }
}

class SettingsModel {
  final String modelProvider;
  final String model;
  final String thinkingMode;
  final bool darkMode;
  final String language;
  final bool p2pEnabled;
  final bool encryptionEnabled;

  SettingsModel({
    this.modelProvider = 'openai',
    this.model = 'gpt-4',
    this.thinkingMode = 'balanced',
    this.darkMode = true,
    this.language = 'en',
    this.p2pEnabled = true,
    this.encryptionEnabled = true,
  });

  factory SettingsModel.fromJson(Map<String, dynamic> json) {
    return SettingsModel(
      modelProvider: json['model_provider'] ?? 'openai',
      model: json['model'] ?? 'gpt-4',
      thinkingMode: json['thinking_mode'] ?? 'balanced',
      darkMode: json['dark_mode'] ?? true,
      language: json['language'] ?? 'en',
      p2pEnabled: json['p2p_enabled'] ?? true,
      encryptionEnabled: json['encryption_enabled'] ?? true,
    );
  }

  Map<String, dynamic> toJson() => {
    'model_provider': modelProvider,
    'model': model,
    'thinking_mode': thinkingMode,
    'dark_mode': darkMode,
    'language': language,
    'p2p_enabled': p2pEnabled,
    'encryption_enabled': encryptionEnabled,
  };
}
