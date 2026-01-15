/// 语言检测服务
class LanguageService {
  static LanguageService? _instance;
  static LanguageService get instance => _instance ??= LanguageService._();
  LanguageService._();

  /// 检测语言
  String detectLanguage(String text) {
    if (text.isEmpty) return 'unknown';

    // 简单的中英文检测
    final hasChinese = RegExp(r'[\u4e00-\u9fa5]').hasMatch(text);
    if (hasChinese) return 'zh';

    final hasJapanese = RegExp(r'[\u3040-\u309f\u30a0-\u30ff]').hasMatch(text);
    if (hasJapanese) return 'ja';

    final hasKorean = RegExp(r'[\uac00-\ud7af]').hasMatch(text);
    if (hasKorean) return 'ko';

    return 'en';
  }

  /// 获取语言名称
  String getLanguageName(String code) {
    switch (code) {
      case 'zh':
        return '中文';
      case 'en':
        return 'English';
      case 'ja':
        return '日本語';
      case 'ko':
        return '한국어';
      default:
        return code;
    }
  }

  /// 获取语言标志
  String getLanguageFlag(String code) {
    switch (code) {
      case 'zh':
        return '🇨🇳';
      case 'en':
        return '🇺🇸';
      case 'ja':
        return '🇯🇵';
      case 'ko':
        return '🇰🇷';
      default:
        return '🌐';
    }
  }

  /// 是否是中文
  bool isChinese(String text) => detectLanguage(text) == 'zh';

  /// 是否是英文
  bool isEnglish(String text) => detectLanguage(text) == 'en';
}
