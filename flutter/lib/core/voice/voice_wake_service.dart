import 'dart:async';

/// 语音唤醒服务
class VoiceWakeService {
  static VoiceWakeService? _instance;
  static VoiceWakeService get instance => _instance ??= VoiceWakeService._();
  VoiceWakeService._();

  bool _isListening = false;
  Function(String wakeWord)? onWakeWordDetected;
  
  bool get isListening => _isListening;
  
  /// 开始监听
  Future<void> startListening() async {
    if (_isListening) return;
    _isListening = true;
  }
  
  /// 停止监听
  Future<void> stopListening() async {
    _isListening = false;
  }
  
  /// 模拟唤醒词检测
  void simulateWakeWord(String word) {
    onWakeWordDetected?.call(word);
  }
}

/// 唤醒词配置
class WakeWordConfig {
  final String name;
  final String? modelPath;
  final double sensitivity;
  final bool isActive;
  final DateTime createdAt;
  
  WakeWordConfig({
    required this.name,
    this.modelPath,
    this.sensitivity = 0.5,
    this.isActive = true,
    DateTime? createdAt,
  }) : createdAt = createdAt ?? DateTime.now();
  
  WakeWordConfig copyWith({
    String? name,
    String? modelPath,
    double? sensitivity,
    bool? isActive,
  }) {
    return WakeWordConfig(
      name: name ?? this.name,
      modelPath: modelPath ?? this.modelPath,
      sensitivity: sensitivity ?? this.sensitivity,
      isActive: isActive ?? this.isActive,
      createdAt: createdAt,
    );
  }
  
  Map<String, dynamic> toJson() => {
    'name': name,
    'modelPath': modelPath,
    'sensitivity': sensitivity,
    'isActive': isActive,
    'createdAt': createdAt.toIso8601String(),
  };
  
  factory WakeWordConfig.fromJson(Map<String, dynamic> json) {
    final sensitivityValue = json['sensitivity'];
    double sensitivity = 0.5;
    if (sensitivityValue is double) {
      sensitivity = sensitivityValue;
    } else if (sensitivityValue is int) {
      sensitivity = sensitivityValue.toDouble();
    } else if (sensitivityValue is String) {
      sensitivity = double.tryParse(sensitivityValue) ?? 0.5;
    }
    
    return WakeWordConfig(
      name: (json['name'] ?? '').toString(),
      modelPath: json['modelPath']?.toString(),
      sensitivity: sensitivity,
      isActive: json['isActive'] == true,
      createdAt: json['createdAt'] != null 
        ? DateTime.parse(json['createdAt'].toString()) 
        : DateTime.now(),
    );
  }
}
