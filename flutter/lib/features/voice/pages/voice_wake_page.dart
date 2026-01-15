import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:speech_to_text/speech_to_text.dart' as stt;
import 'package:flutter_tts/flutter_tts.dart';
import '../providers/voice_provider.dart';

// Voice wake word settings provider
final wakeWordProvider = StateProvider<String>((ref) => 'Hey Tortoise');
final wakeEnabledProvider = StateProvider<bool>((ref) => false);
final wakeSensitivityProvider = StateProvider<double>((ref) => 0.5);

class VoiceWakePage extends ConsumerStatefulWidget {
  const VoiceWakePage({super.key});

  @override
  ConsumerState<VoiceWakePage> createState() => _VoiceWakePageState();
}

class _VoiceWakePageState extends ConsumerState<VoiceWakePage>
    with SingleTickerProviderStateMixin {
  late AnimationController _pulseController;
  late Animation<double> _pulseAnimation;
  
  bool _isListening = false;
  bool _isProcessing = false;
  bool _speechEnabled = false;
  String _lastWords = '';
  double _soundLevel = 0.0;
  
  final stt.SpeechToText _speech = stt.SpeechToText();
  final FlutterTts _tts = FlutterTts();
  
  Timer? _silenceTimer;
  bool _wakeWordDetected = false;

  @override
  void initState() {
    super.initState();
    _initAnimations();
    _initSpeech();
    _initTts();
  }

  void _initAnimations() {
    _pulseController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1500),
    )..repeat(reverse: true);
    
    _pulseAnimation = Tween<double>(begin: 1.0, end: 1.3).animate(
      CurvedAnimation(parent: _pulseController, curve: Curves.easeInOut),
    );
  }

  Future<void> _initSpeech() async {
    _speechEnabled = await _speech.initialize(
      onStatus: _onSpeechStatus,
      onError: _onSpeechError,
    );
    
    if (_speechEnabled) {
      setState(() {});
    }
  }

  Future<void> _initTts() async {
    await _tts.setLanguage('zh-CN');
    await _tts.setSpeechRate(0.5);
    await _tts.setVolume(1.0);
    await _tts.setPitch(1.0);
  }

  @override
  void dispose() {
    _pulseController.dispose();
    _speech.stop();
    _tts.stop();
    _silenceTimer?.cancel();
    super.dispose();
  }

  void _onSpeechStatus(String status) {
    if (status == 'done' || status == 'notListening') {
      if (_isListening && !_wakeWordDetected) {
        _startListening();
      }
    }
  }

  void _onSpeechError(dynamic error) {
    debugPrint('Speech error: $error');
  }

  Future<void> _toggleListening() async {
    if (_isListening) {
      await _stopListening();
    } else {
      await _startListening();
    }
  }

  Future<void> _startListening() async {
    if (!_speechEnabled) {
      _showError('语音识别不可用');
      return;
    }

    setState(() {
      _isListening = true;
      _wakeWordDetected = false;
    });

    await _speech.listen(
      onResult: _onSpeechResult,
      listenFor: const Duration(seconds: 30),
      pauseFor: const Duration(seconds: 3),
      partialResults: true,
      onSoundLevelChange: (level) {
        setState(() {
          _soundLevel = (level + 2) / 2; // Normalize to 0-1
        });
      },
    );
  }

  Future<void> _stopListening() async {
    await _speech.stop();
    _silenceTimer?.cancel();
    
    setState(() {
      _isListening = false;
      _soundLevel = 0;
    });
  }

  void _onSpeechResult(result) {
    final words = result.recognizedWords;
    final confidence = result.confidence;
    
    setState(() {
      _lastWords = words;
    });

    if (words.isEmpty) return;

    // Check for wake word
    final wakeWord = ref.read(wakeWordProvider).toLowerCase();
    if (!_wakeWordDetected && words.toLowerCase().contains(wakeWord)) {
      _wakeWordDetected = true;
      _onWakeWordDetected();
    }

    // Reset silence timer on speech
    _silenceTimer?.cancel();
    _silenceTimer = Timer(const Duration(seconds: 3), () {
      if (_isListening && !_wakeWordDetected) {
        _startListening(); // Restart listening
      }
    });
  }

  void _onWakeWordDetected() async {
    await _speech.stop();
    setState(() {
      _isProcessing = true;
    });

    // Audio feedback
    await _tts.speak('我在');
    
    // Emit event for main app to handle
    if (mounted) {
      _showSuccess('唤醒成功！请说话...');
      
      // Continue listening for commands
      setState(() {
        _isProcessing = false;
        _wakeWordDetected = false;
      });
      
      await _startListening();
    }
  }

  Future<void> _testVoice() async {
    await _tts.speak('你好，我是 Tortoise');
  }

  void _showError(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message),
        backgroundColor: Colors.red,
      ),
    );
  }

  void _showSuccess(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message),
        backgroundColor: Colors.green,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final wakeWord = ref.watch(wakeWordProvider);
    final enabled = ref.watch(wakeEnabledProvider);
    final sensitivity = ref.watch(wakeSensitivityProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('语音唤醒'),
        actions: [
          IconButton(
            icon: const Icon(Icons.help_outline),
            onPressed: _showHelp,
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // Status Card
            _buildStatusCard(),
            const SizedBox(height: 24),
            
            // Wake Word Settings
            _buildWakeWordSettings(wakeWord),
            const SizedBox(height: 16),
            
            // Sensitivity Slider
            _buildSensitivitySlider(sensitivity),
            const SizedBox(height: 24),
            
            // Voice Activity Indicator
            _buildVoiceActivityIndicator(),
            const SizedBox(height: 24),
            
            // Control Buttons
            _buildControlButtons(),
            const SizedBox(height: 24),
            
            // Last Recognized Text
            if (_lastWords.isNotEmpty) _buildLastWordsCard(),
          ],
        ),
      ),
    );
  }

  Widget _buildStatusCard() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          children: [
            AnimatedBuilder(
              animation: _pulseAnimation,
              builder: (context, child) {
                return Transform.scale(
                  scale: _isListening ? _pulseAnimation.value : 1.0,
                  child: Container(
                    width: 120,
                    height: 120,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      color: _isListening
                          ? Colors.green.withOpacity(0.2)
                          : Colors.grey.withOpacity(0.2),
                      border: Border.all(
                        color: _isListening ? Colors.green : Colors.grey,
                        width: 3,
                      ),
                    ),
                    child: Icon(
                      _isListening ? Icons.mic : Icons.mic_off,
                      size: 60,
                      color: _isListening ? Colors.green : Colors.grey,
                    ),
                  ),
                );
              },
            ),
            const SizedBox(height: 16),
            Text(
              _getStatusText(),
              style: Theme.of(context).textTheme.titleLarge,
            ),
            if (_isProcessing)
              const Padding(
                padding: EdgeInsets.only(top: 8),
                child: CircularProgressIndicator(),
              ),
          ],
        ),
      ),
    );
  }

  String _getStatusText() {
    if (_isProcessing) return '处理中...';
    if (_isListening) return '正在监听...';
    return '点击开始监听';
  }

  Widget _buildWakeWordSettings(String wakeWord) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              '唤醒词设置',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 12),
            TextField(
              decoration: const InputDecoration(
                labelText: '唤醒词',
                hintText: '例如: Hey Tortoise',
                prefixIcon: Icon(Icons.record_voice_over),
              ),
              controller: TextEditingController(text: wakeWord),
              onChanged: (value) {
                ref.read(wakeWordProvider.notifier).state = value;
              },
            ),
            const SizedBox(height: 12),
            Wrap(
              spacing: 8,
              children: ['Hey Tortoise', '小乌龟', 'Tortoise', '你好'].map((word) {
                return ActionChip(
                  label: Text(word),
                  onPressed: () {
                    ref.read(wakeWordProvider.notifier).state = word;
                  },
                );
              }).toList(),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSensitivitySlider(double sensitivity) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  '灵敏度',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                Text(
                  '${(sensitivity * 100).round()}%',
                  style: Theme.of(context).textTheme.bodyLarge,
                ),
              ],
            ),
            Slider(
              value: sensitivity,
              min: 0.1,
              max: 1.0,
              divisions: 9,
              label: '${(sensitivity * 100).round()}%',
              onChanged: (value) {
                ref.read(wakeSensitivityProvider.notifier).state = value;
              },
            ),
            const Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('不敏感', style: TextStyle(color: Colors.grey)),
                Text('敏感', style: TextStyle(color: Colors.grey)),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildVoiceActivityIndicator() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              '语音活动',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 12),
            ClipRRect(
              borderRadius: BorderRadius.circular(4),
              child: LinearProgressIndicator(
                value: _soundLevel,
                minHeight: 8,
                backgroundColor: Colors.grey[300],
                valueColor: AlwaysStoppedAnimation<Color>(
                  _soundLevel > 0.5 ? Colors.green : Colors.blue,
                ),
              ),
            ),
            const SizedBox(height: 8),
            Text(
              _lastWords.isEmpty ? '说点什么...' : _lastWords,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                color: _lastWords.isEmpty ? Colors.grey : null,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildControlButtons() {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceEvenly,
      children: [
        ElevatedButton.icon(
          onPressed: _toggleListening,
          icon: Icon(_isListening ? Icons.stop : Icons.play_arrow),
          label: Text(_isListening ? '停止' : '开始'),
          style: ElevatedButton.styleFrom(
            backgroundColor: _isListening ? Colors.red : Colors.green,
            foregroundColor: Colors.white,
            padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
          ),
        ),
        OutlinedButton.icon(
          onPressed: _testVoice,
          icon: const Icon(Icons.volume_up),
          label: const Text('测试语音'),
        ),
      ],
    );
  }

  Widget _buildLastWordsCard() {
    return Card(
      color: Colors.blue[50],
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Icon(Icons.text_fields, size: 20),
                const SizedBox(width: 8),
                Text(
                  '识别内容',
                  style: Theme.of(context).textTheme.titleSmall,
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              _lastWords,
              style: Theme.of(context).textTheme.bodyLarge,
            ),
          ],
        ),
      ),
    );
  }

  void _showHelp() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('语音唤醒帮助'),
        content: const SingleChildScrollView(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                '如何使用',
                style: TextStyle(fontWeight: FontWeight.bold),
              ),
              SizedBox(height: 8),
              Text('1. 设置唤醒词（默认: Hey Tortoise）'),
              Text('2. 点击"开始"按钮'),
              Text('3. 说出唤醒词来激活助手'),
              Text('4. 唤醒后可继续下达语音命令'),
              SizedBox(height: 16),
              Text(
                '提示',
                style: TextStyle(fontWeight: FontWeight.bold),
              ),
              SizedBox(height: 8),
              Text('• 建议在安静环境使用'),
              Text('• 确保麦克风权限已开启'),
              Text('• 可调整灵敏度适应环境'),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('知道了'),
          ),
        ],
      ),
    );
  }
}
