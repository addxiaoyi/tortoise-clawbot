import 'dart:async';

/// Voice Tools - 语音工具
/// 语音交互能力
class VoiceTools {
  /// 语音识别
  SpeechRecognition? _speechRecognition;

  /// 语音合成
  SpeechSynthesis? _speechSynthesis;

  /// 当前状态
  VoiceStatus status = VoiceStatus.idle;

  /// 回调
  Function(String text)? onResult;
  Function(String error)? onError;

  /// 初始化
  Future<void> initialize() async {
    _speechRecognition = SpeechRecognition();
    _speechSynthesis = SpeechSynthesis();

    _speechRecognition!.onResult = (text) {
      onResult?.call(text);
    };

    _speechRecognition!.onError = (error) {
      onError?.call(error);
    };
  }

  /// 开始监听
  Future<void> startListening({
    String language = 'zh-CN',
    bool continuous = false,
  }) async {
    if (status == VoiceStatus.listening) return;

    status = VoiceStatus.listening;
    await _speechRecognition?.start(
      language: language,
      continuous: continuous,
    );
  }

  /// 停止监听
  Future<void> stopListening() async {
    if (status != VoiceStatus.listening) return;

    await _speechRecognition?.stop();
    status = VoiceStatus.idle;
  }

  /// 说话
  Future<void> speak(String text, {
    String voice = 'default',
    double rate = 1.0,
    double pitch = 1.0,
  }) async {
    status = VoiceStatus.speaking;
    await _speechSynthesis?.speak(text, voice: voice, rate: rate, pitch: pitch);
    status = VoiceStatus.idle;
  }

  /// 停止说话
  Future<void> stopSpeaking() async {
    await _speechSynthesis?.stop();
    status = VoiceStatus.idle;
  }

  /// 获取可用语音
  List<String> getAvailableVoices() {
    return _speechSynthesis?.getVoices() ?? [];
  }

  /// 销毁
  void dispose() {
    _speechRecognition?.dispose();
    _speechSynthesis?.dispose();
  }
}

/// 语音状态
enum VoiceStatus {
  idle,
  listening,
  speaking,
}

/// 语音识别 (简化接口)
class SpeechRecognition {
  Function(String text)? onResult;
  Function(String error)? onError;

  Future<void> start({
    String language = 'zh-CN',
    bool continuous = false,
  }) async {
    // 实际实现使用 speech_to_text 插件
  }

  Future<void> stop() async {}

  void dispose() {}
}

/// 语音合成 (简化接口)
class SpeechSynthesis {
  Future<void> speak(String text, {
    String voice = 'default',
    double rate = 1.0,
    double pitch = 1.0,
  }) async {
    // 实际实现使用 flutter_tts 插件
  }

  Future<void> stop() async {}

  List<String> getVoices() => ['default', 'female', 'male'];

  void dispose() {}
}

/// Voice Activity Detection - 语音活动检测
class VoiceActivityDetection {
  /// 是否启用
  bool enabled = true;

  /// 能量阈值
  double energyThreshold = 0.5;

  /// 最后活动时间
  DateTime? lastActivity;

  /// 检测到语音活动
  Function()? onVoiceActivity;

  /// 语音结束
  Function(Duration duration)? onVoiceEnd;

  /// 计时器
  Timer? _silenceTimer;

  /// 静音超时时间 (秒)
  static const int silenceTimeout = 3;

  /// 检测能量
  void detectEnergy(double energy) {
    if (!enabled) return;

    if (energy > energyThreshold) {
      // 语音活动
      lastActivity = DateTime.now();
      _silenceTimer?.cancel();
      onVoiceActivity?.call();
    } else {
      // 静音
      if (lastActivity != null && _silenceTimer == null) {
        // 开始静音计时
        _silenceTimer = Timer(
          const Duration(seconds: silenceTimeout),
          () {
            final duration = DateTime.now().difference(lastActivity!);
            onVoiceEnd?.call(duration);
            lastActivity = null;
            _silenceTimer = null;
          },
        );
      }
    }
  }
}

/// Wake Word Detection - 唤醒词检测
class WakeWordDetector {
  /// 唤醒词列表
  final List<String> wakeWords;

  /// 检测回调
  Function(String word)? onWakeWordDetected;

  /// 当前状态
  bool isListening = false;

  WakeWordDetector({
    this.wakeWords = const ['小 Tortoise', 'Hey Tortoise'],
  });

  /// 开始监听
  void startListening() {
    isListening = true;
  }

  /// 停止监听
  void stopListening() {
    isListening = false;
  }

  /// 检测音频
  void detect(List<double> audioData) {
    if (!isListening) return;

    // 简化实现：实际会使用专门的唤醒词模型
    // 这里只是演示接口
  }

  /// 添加唤醒词
  void addWakeWord(String word) {
    if (!wakeWords.contains(word)) {
      wakeWords.add(word);
    }
  }

  /// 移除唤醒词
  void removeWakeWord(String word) {
    wakeWords.remove(word);
  }
}

/// Voice Commands - 语音命令
class VoiceCommands {
  /// 命令映射
  final Map<String, VoiceCommand> _commands = {};

  /// 注册命令
  void register(VoiceCommand command) {
    _commands[command.pattern] = command;
  }

  /// 执行命令
  Future<bool> execute(String input) async {
    for (final entry in _commands.entries) {
      if (_matches(entry.key, input)) {
        await entry.value.handler();
        return true;
      }
    }
    return false;
  }

  /// 匹配命令
  bool _matches(String pattern, String input) {
    // 简单匹配：包含关系
    // 实际实现会使用更复杂的 NLP
    return input.toLowerCase().contains(pattern.toLowerCase());
  }

  /// 获取所有命令
  List<VoiceCommand> getAll() => _commands.values.toList();
}

/// 语音命令
class VoiceCommand {
  final String pattern;
  final String description;
  final String category;
  final Future<void> Function() handler;

  VoiceCommand({
    required this.pattern,
    required this.description,
    required this.category,
    required this.handler,
  });
}
