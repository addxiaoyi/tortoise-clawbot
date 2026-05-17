/// Tortoise API 配置
class ApiConfig {
  // Gateway 地址
  static const String defaultGateway = 'http://localhost:8080';
  
  // WebSocket 地址
  static const String defaultWebSocket = 'ws://localhost:8080/ws';
  
  // API 版本
  static const String apiVersion = 'v1';
  
  // 超时设置
  static const Duration connectTimeout = Duration(seconds: 30);
  static const Duration receiveTimeout = Duration(seconds: 60);
  
  // 重试配置
  static const int maxRetries = 3;
  static const Duration retryDelay = Duration(seconds: 1);
  
  // 端点
  static const Map<String, String> endpoints = {
    'health': '/health',
    'status': '/api/v1/status',
    'chat': '/api/v1/chat',
    'sessions': '/api/v1/sessions',
    'messages': '/api/v1/messages',
    'agents': '/api/v1/agents',
    'plugins': '/api/v1/plugins',
    'channels': '/api/v1/channels',
    'memory': '/api/v1/memory',
    'skills': '/api/v1/skills',
    'marketplace': '/api/v1/marketplace',
    'config': '/api/v1/config',
  };
  
  // 获取完整 URL
  static String getUrl(String endpoint) {
    return '$defaultGateway${endpoints[endpoint] ?? endpoint}';
  }
}

/// 消息渠道常量
class Channels {
  static const String telegram = 'telegram';
  static const String discord = 'discord';
  static const String slack = 'slack';
  static const String whatsapp = 'whatsapp';
  static const String signal = 'signal';
  static const String matrix = 'matrix';
  static const String email = 'email';
  static const String sms = 'sms';
  static const String web = 'web';
  
  static const List<String> all = [
    telegram,
    discord,
    slack,
    whatsapp,
    signal,
    matrix,
    email,
    sms,
    web,
  ];
  
  static const Map<String, String> names = {
    telegram: 'Telegram',
    discord: 'Discord',
    slack: 'Slack',
    whatsapp: 'WhatsApp',
    signal: 'Signal',
    matrix: 'Matrix',
    email: 'Email',
    sms: 'SMS',
    web: 'Web',
  };
  
  static const Map<String, String> descriptions = {
    telegram: 'Telegram 是一款跨平台的即时通讯软件',
    discord: 'Discord 是专为社区设计的聊天软件',
    slack: 'Slack 是企业级团队协作平台',
    whatsapp: 'WhatsApp 是全球流行的即时通讯应用',
    signal: 'Signal 是注重隐私的加密通讯应用',
    matrix: 'Matrix 是一个去中心化通讯协议',
    email: 'Email 邮件是最传统的电子通讯方式',
    sms: 'SMS 短信是基础的移动通讯',
    web: 'Web 渠道提供网页端访问',
  };
}
