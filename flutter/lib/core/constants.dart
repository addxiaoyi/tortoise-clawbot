// Tortoise App - 核心常量
library;

class AppConstants {
  AppConstants._();

  // 版本信息
  static const String appName = 'Tortoise';
  static const String appVersion = '0.1.0';
  static const int appBuildNumber = 1;

  // API 配置
  static const String defaultApiUrl = 'http://localhost:18792';
  static const Duration apiTimeout = Duration(seconds: 30);
  static const Duration streamTimeout = Duration(minutes: 5);

  // 存储键名
  static const String keyApiUrl = 'api_url';
  static const String keyApiKey = 'api_key';
  static const String keyOpenAiKey = 'openai_api_key';
  static const String keyAnthropicKey = 'anthropic_api_key';
  static const String keyTelegramToken = 'telegram_bot_token';
  static const String keyDiscordToken = 'discord_bot_token';
  static const String keyDefaultModel = 'default_model';
  static const String keyThemeMode = 'theme_mode';
  static const String keyLanguage = 'language';
  static const String keyAutoConnect = 'auto_connect';
  static const String keyEnableDiscovery = 'enable_discovery';

  // 默认值
  static const String defaultModel = 'gpt-4';
  static const List<String> supportedModels = [
    'gpt-4-turbo-preview',
    'gpt-4',
    'gpt-3.5-turbo',
    'claude-3-opus-20240229',
    'claude-3-sonnet-20240229',
    'claude-3-haiku-20240307',
  ];

  // 渠道类型
  static const String channelTelegram = 'telegram';
  static const String channelDiscord = 'discord';
  static const String channelSlack = 'slack';
  static const String channelWebSocket = 'websocket';

  // 主题
  static const String themeLight = 'light';
  static const String themeDark = 'dark';
  static const String themeSystem = 'system';
}
