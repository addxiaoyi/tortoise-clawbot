/// 验证服务
class ValidationService {
  static ValidationService? _instance;
  static ValidationService get instance => _instance ??= ValidationService._();
  ValidationService._();

  /// 验证邮箱
  bool isValidEmail(String email) {
    final regex = RegExp(r'^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$');
    return regex.hasMatch(email);
  }

  /// 验证 URL
  bool isValidUrl(String url) {
    final regex = RegExp(r'^https?:\/\/([\w-]+\.)+[\w-]+(\/[\w-./?%&=]*)?$');
    return regex.hasMatch(url);
  }

  /// 验证手机号
  bool isValidPhone(String phone) {
    final regex = RegExp(r'^1[3-9]\d{9}$');
    return regex.hasMatch(phone);
  }

  /// 验证密码强度
  PasswordStrength checkPasswordStrength(String password) {
    if (password.length < 6) return PasswordStrength.weak;
    
    int score = 0;
    if (password.length >= 8) score++;
    if (password.length >= 12) score++;
    if (RegExp(r'[A-Z]').hasMatch(password)) score++;
    if (RegExp(r'[a-z]').hasMatch(password)) score++;
    if (RegExp(r'[0-9]').hasMatch(password)) score++;
    if (RegExp(r'[!@#$%^&*(),.?":{}|<>]').hasMatch(password)) score++;
    
    if (score <= 2) return PasswordStrength.weak;
    if (score <= 4) return PasswordStrength.medium;
    return PasswordStrength.strong;
  }

  /// 验证 API Key 格式
  bool isValidApiKey(String key, ApiKeyType type) {
    switch (type) {
      case ApiKeyType.openai:
        return key.startsWith('sk-') && key.length > 40;
      case ApiKeyType.anthropic:
        return key.startsWith('sk-ant-') && key.length > 40;
      case ApiKeyType.telegram:
        return key.length > 30;
      case ApiKeyType.discord:
        return key.length > 50;
    }
  }

  /// 验证必填字段
  bool isNotEmpty(String? value) {
    return value != null && value.trim().isNotEmpty;
  }

  /// 验证长度范围
  bool isLengthInRange(String value, int min, int max) {
    return value.length >= min && value.length <= max;
  }

  /// 验证数字范围
  bool isNumberInRange(num value, num min, num max) {
    return value >= min && value <= max;
  }
}

/// 密码强度
enum PasswordStrength {
  weak,
  medium,
  strong,
}

/// API Key 类型
enum ApiKeyType {
  openai,
  anthropic,
  telegram,
  discord,
}

/// 验证结果
class ValidationResult {
  final bool isValid;
  final String? error;

  ValidationResult({required this.isValid, this.error});

  factory ValidationResult.valid() => ValidationResult(isValid: true);
  factory ValidationResult.invalid(String error) => ValidationResult(isValid: false, error: error);
}
