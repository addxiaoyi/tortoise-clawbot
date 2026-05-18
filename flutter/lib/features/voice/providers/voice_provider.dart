import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Voice Wake 状态
class VoiceWakeState {
  final bool isInitialized;
  final bool isListening;
  final String? currentWakeWord;
  final List<WakeWord> wakeWords;
  final String? error;
  final double sensitivity;
  final VoiceWakeStatus status;

  VoiceWakeState({
    this.isInitialized = false,
    this.isListening = false,
    this.currentWakeWord,
    this.wakeWords = const [],
    this.error,
    this.sensitivity = 0.5,
    this.status = VoiceWakeStatus.idle,
  });

  VoiceWakeState copyWith({
    bool? isInitialized,
    bool? isListening,
    String? currentWakeWord,
    List<WakeWord>? wakeWords,
    String? error,
    double? sensitivity,
    VoiceWakeStatus? status,
  }) {
    return VoiceWakeState(
      isInitialized: isInitialized ?? this.isInitialized,
      isListening: isListening ?? this.isListening,
      currentWakeWord: currentWakeWord ?? this.currentWakeWord,
      wakeWords: wakeWords ?? this.wakeWords,
      error: error,
      sensitivity: sensitivity ?? this.sensitivity,
      status: status ?? this.status,
    );
  }
}

/// 唤醒词
class WakeWord {
  final String id;
  final String name;
  final double sensitivity;
  final bool isActive;
  final DateTime createdAt;

  WakeWord({
    required this.id,
    required this.name,
    this.sensitivity = 0.5,
    this.isActive = true,
    DateTime? createdAt,
  }) : createdAt = createdAt ?? DateTime.now();
}

/// 语音唤醒状态
enum VoiceWakeStatus {
  idle,
  initializing,
  ready,
  listening,
  error,
}

/// Voice Wake Notifier
class VoiceWakeNotifier extends StateNotifier<VoiceWakeState> {
  VoiceWakeNotifier() : super(VoiceWakeState());

  /// 初始化
  Future<void> initialize() async {
    state = state.copyWith(status: VoiceWakeStatus.initializing);

    try {
      // 模拟初始化
      await Future.delayed(const Duration(milliseconds: 500));

      // 添加默认唤醒词
      final defaultWakeWords = [
        WakeWord(id: '1', name: 'Hey Tortoise', sensitivity: 0.6),
        WakeWord(id: '2', name: 'Hey Assistant', sensitivity: 0.5),
      ];

      state = state.copyWith(
        isInitialized: true,
        status: VoiceWakeStatus.ready,
        wakeWords: defaultWakeWords,
        currentWakeWord: 'Hey Tortoise',
      );
    } catch (e) {
      state = state.copyWith(
        status: VoiceWakeStatus.error,
        error: e.toString(),
      );
    }
  }

  /// 开始监听
  Future<void> startListening({String? wakeWord}) async {
    if (!state.isInitialized) {
      await initialize();
    }

    state = state.copyWith(
      isListening: true,
      currentWakeWord: wakeWord ?? state.currentWakeWord,
      status: VoiceWakeStatus.listening,
    );
  }

  /// 停止监听
  void stopListening() {
    state = state.copyWith(
      isListening: false,
      status: VoiceWakeStatus.ready,
    );
  }

  /// 添加唤醒词
  void addWakeWord(String name, {double sensitivity = 0.5}) {
    final wakeWord = WakeWord(
      id: DateTime.now().millisecondsSinceEpoch.toString(),
      name: name,
      sensitivity: sensitivity,
    );

    state = state.copyWith(
      wakeWords: [...state.wakeWords, wakeWord],
    );
  }

  /// 删除唤醒词
  void removeWakeWord(String id) {
    state = state.copyWith(
      wakeWords: state.wakeWords.where((w) => w.id != id).toList(),
    );
  }

  /// 更新唤醒词
  void updateWakeWord(String id, {String? name, double? sensitivity, bool? isActive}) {
    final wakeWords = state.wakeWords.map((w) {
      if (w.id == id) {
        return WakeWord(
          id: w.id,
          name: name ?? w.name,
          sensitivity: sensitivity ?? w.sensitivity,
          isActive: isActive ?? w.isActive,
          createdAt: w.createdAt,
        );
      }
      return w;
    }).toList();

    state = state.copyWith(wakeWords: wakeWords);
  }

  /// 设置灵敏度
  void setSensitivity(double sensitivity) {
    state = state.copyWith(sensitivity: sensitivity);
  }

  /// 测试唤醒词
  Future<void> testWakeWord(String name) async {
    state = state.copyWith(status: VoiceWakeStatus.listening);

    await Future.delayed(const Duration(seconds: 2));

    state = state.copyWith(status: VoiceWakeStatus.ready);
  }

  /// 重置
  void reset() {
    state = VoiceWakeState();
  }
}

/// Provider
final voiceWakeProvider = StateNotifierProvider<VoiceWakeNotifier, VoiceWakeState>((ref) {
  return VoiceWakeNotifier();
});
