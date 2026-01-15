import 'package:flutter/material.dart';

/// 应用国际化
class AppLocalizations {
  final Locale locale;

  AppLocalizations(this.locale);

  static AppLocalizations? of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations);
  }

  static const LocalizationsDelegate<AppLocalizations> delegate = _AppLocalizationsDelegate();

  static const List<Locale> supportedLocales = [
    Locale('en', 'US'),
    Locale('zh', 'CN'),
  ];

  static final Map<String, Map<String, String>> _localizedValues = {
    'en_US': {
      // 首页
      'home': 'Home',
      'chat': 'Chat',
      'settings': 'Settings',
      'ai_status': 'AI Status',
      'ready': 'Ready',
      'not_configured': 'Not Configured',
      'quick_actions': 'Quick Actions',
      'new_chat': 'New Chat',
      'device_discovery': 'Device Discovery',
      'refresh': 'Refresh',
      'devices': 'Devices',
      'channels': 'Channels',
      'no_devices': 'No devices found',

      // 对话
      'start_chatting': 'Start chatting',
      'input_placeholder': 'Type a message...',
      'thinking': 'AI is thinking...',

      // 设置
      'appearance': 'Appearance',
      'theme_mode': 'Theme Mode',
      'follow_system': 'Follow System',
      'light': 'Light',
      'dark': 'Dark',
      'language': 'Language',
      'ai_settings': 'AI Settings',
      'openai_api': 'OpenAI API Key',
      'anthropic_api': 'Anthropic API Key',
      'model': 'Model',
      'message_channels': 'Message Channels',
      'telegram': 'Telegram',
      'discord': 'Discord',
      'not_connected': 'Not Connected',
      'connected': 'Connected',
      'about': 'About',
      'version': 'Version',

      // 操作
      'save': 'Save',
      'cancel': 'Cancel',
      'delete': 'Delete',
      'clear': 'Clear',
      'configure': 'Configure',
      'click_to_configure': 'Click to configure',
    },
    'zh_CN': {
      // 首页
      'home': '首页',
      'chat': '对话',
      'settings': '设置',
      'ai_status': 'AI 状态',
      'ready': '已就绪',
      'not_configured': '未配置',
      'quick_actions': '快速操作',
      'new_chat': '新对话',
      'device_discovery': '设备发现',
      'refresh': '刷新',
      'devices': '设备',
      'channels': '消息渠道',
      'no_devices': '未发现设备',

      // 对话
      'start_chatting': '开始对话吧',
      'input_placeholder': '输入消息...',
      'thinking': 'AI 正在思考...',

      // 设置
      'appearance': '外观',
      'theme_mode': '主题模式',
      'follow_system': '跟随系统',
      'light': '浅色',
      'dark': '深色',
      'language': '语言',
      'ai_settings': 'AI 设置',
      'openai_api': 'OpenAI API Key',
      'anthropic_api': 'Anthropic API Key',
      'model': '模型',
      'message_channels': '消息渠道',
      'telegram': 'Telegram',
      'discord': 'Discord',
      'not_connected': '未连接',
      'connected': '已连接',
      'about': '关于',
      'version': '版本',

      // 操作
      'save': '保存',
      'cancel': '取消',
      'delete': '删除',
      'clear': '清除',
      'configure': '配置',
      'click_to_configure': '点击配置',
    },
  };

  String _getKey(String key) {
    final localeKey = locale.languageCode == 'zh' ? 'zh_CN' : 'en_US';
    return _localizedValues[localeKey]?[key] ?? key;
  }

  String get home => _getKey('home');
  String get chat => _getKey('chat');
  String get settings => _getKey('settings');
  String get aiStatus => _getKey('ai_status');
  String get ready => _getKey('ready');
  String get notConfigured => _getKey('not_configured');
  String get quickActions => _getKey('quick_actions');
  String get newChat => _getKey('new_chat');
  String get deviceDiscovery => _getKey('device_discovery');
  String get refresh => _getKey('refresh');
  String get devices => _getKey('devices');
  String get channels => _getKey('channels');
  String get noDevices => _getKey('no_devices');
  String get startChatting => _getKey('start_chatting');
  String get inputPlaceholder => _getKey('input_placeholder');
  String get thinking => _getKey('thinking');
  String get appearance => _getKey('appearance');
  String get themeMode => _getKey('theme_mode');
  String get followSystem => _getKey('follow_system');
  String get light => _getKey('light');
  String get dark => _getKey('dark');
  String get language => _getKey('language');
  String get aiSettings => _getKey('ai_settings');
  String get openaiApi => _getKey('openai_api');
  String get anthropicApi => _getKey('anthropic_api');
  String get model => _getKey('model');
  String get messageChannels => _getKey('message_channels');
  String get telegram => _getKey('telegram');
  String get discord => _getKey('discord');
  String get notConnected => _getKey('not_connected');
  String get connected => _getKey('connected');
  String get about => _getKey('about');
  String get version => _getKey('version');
  String get save => _getKey('save');
  String get cancel => _getKey('cancel');
  String get delete => _getKey('delete');
  String get clear => _getKey('clear');
  String get configure => _getKey('configure');
  String get clickToConfigure => _getKey('click_to_configure');
}

class _AppLocalizationsDelegate extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  bool isSupported(Locale locale) {
    return ['en', 'zh'].contains(locale.languageCode);
  }

  @override
  Future<AppLocalizations> load(Locale locale) async {
    return AppLocalizations(locale);
  }

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}
