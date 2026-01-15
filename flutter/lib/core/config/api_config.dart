/// API配置常量
class ApiConfig {
  // API地址配置
  static const String defaultApiUrl = 'http://localhost:8080';
  static const String defaultWsUrl = 'ws://localhost:8080/ws';
  
  // API版本
  static const String apiVersion = 'v1';
  
  // 超时配置
  static const Duration connectTimeout = Duration(seconds: 30);
  static const Duration receiveTimeout = Duration(seconds: 60);
  
  // 端点
  static const String healthEndpoint = '/health';
  static const String sessionsEndpoint = '/api/v1/sessions';
  static const String messagesEndpoint = '/api/v1/messages';
  static const String configEndpoint = '/api/v1/config';
  static const String channelsEndpoint = '/api/v1/channels';
  static const String pluginsEndpoint = '/api/v1/plugins';
  static const String memoryEndpoint = '/api/v1/memory';
}

/// AI提供商配置
class AIProviders {
  static const String openai = 'openai';
  static const String anthropic = 'anthropic';
  static const String google = 'google';
  static const String ollama = 'ollama';
  static const String groq = 'groq';
  static const String openrouter = 'openrouter';
  static const String deepseek = 'deepseek';
  static const String mistral = 'mistral';
  static const String togetherai = 'togetherai';
  static const String perplexity = 'perplexity';
  
  // OpenAI 模型列表
  static const List<String> openaiModels = [
    'gpt-4-turbo-preview',
    'gpt-4',
    'gpt-4-32k',
    'gpt-3.5-turbo',
    'gpt-3.5-turbo-16k',
  ];
  
  // Anthropic 模型列表
  static const List<String> anthropicModels = [
    'claude-3-opus-20240229',
    'claude-3-sonnet-20240229',
    'claude-3-haiku-20240307',
    'claude-2.1',
    'claude-2.0',
    'claude-instant-1.2',
  ];
  
  // 提供商显示名称
  static const Map<String, String> names = {
    openai: 'OpenAI',
    anthropic: 'Anthropic Claude',
    google: 'Google Gemini',
    ollama: 'Ollama (本地)',
    groq: 'Groq',
    openrouter: 'OpenRouter',
    deepseek: 'DeepSeek',
    mistral: 'Mistral AI',
    togetherai: 'Together AI',
    perplexity: 'Perplexity',
  };
  
  // API基础URL
  static const Map<String, String> baseUrls = {
    openai: 'https://api.openai.com/v1',
    anthropic: 'https://api.anthropic.com/v1',
    google: 'https://generativelanguage.googleapis.com/v1beta',
    deepseek: 'https://attachment.cdn/img/3fdd23a6aee0.png',
  };
  
  // 默认模型
  static const Map<String, String> defaultModels = {
    openai: 'gpt-4-turbo-preview',
    anthropic: 'claude-3-opus-20240229',
    google: 'gemini-pro',
    ollama: 'llama3',
    groq: 'mixtral-8x7b-32768',
    openrouter: 'openai/gpt-4-turbo',
    deepseek: 'deepseek-chat',
    mistral: 'mistral-medium',
    togetherai: 'meta-llama/Llama-3-70b-chat-hf',
    perplexity: 'pplx-70b-online',
  };
}

/// 消息渠道配置
class Channels {
  static const String telegram = 'telegram';
  static const String discord = 'discord';
  static const String slack = 'slack';
  static const String whatsapp = 'whatsapp';
  static const String matrix = 'matrix';
  static const String signal = 'signal';
  static const String email = 'email';
  static const String sms = 'sms';
  static const String web = 'web';
  
  static const List<String> all = [
    telegram,
    discord,
    slack,
    whatsapp,
    matrix,
    signal,
    email,
    sms,
    web,
  ];
  
  static const Map<String, String> names = {
    telegram: 'Telegram',
    discord: 'Discord',
    slack: 'Slack',
    whatsapp: 'WhatsApp',
    matrix: 'Matrix',
    signal: 'Signal',
    email: 'Email',
    sms: 'SMS',
    web: 'Web',
  };
}
