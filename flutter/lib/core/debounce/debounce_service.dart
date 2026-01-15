import 'dart:async';

/// 防抖服务
class Debouncer {
  final Duration delay;
  Timer? _timer;

  Debouncer({this.delay = const Duration(milliseconds: 300)});

  void run(void Function() action) {
    _timer?.cancel();
    _timer = Timer(delay, action);
  }

  void cancel() {
    _timer?.cancel();
  }

  void dispose() {
    _timer?.cancel();
    _timer = null;
  }
}

/// 节流服务
class Throttler {
  final Duration interval;
  Timer? _timer;
  bool _isThrottled = false;

  Throttler({this.interval = const Duration(milliseconds: 300)});

  void run(void Function() action) {
    if (_isThrottled) return;
    action();
    _isThrottled = true;
    _timer?.cancel();
    _timer = Timer(interval, () {
      _isThrottled = false;
    });
  }

  void cancel() {
    _timer?.cancel();
    _isThrottled = false;
  }

  void dispose() {
    _timer?.cancel();
    _timer = null;
    _isThrottled = false;
  }
}

/// 延迟执行
Future<void> delay(Duration duration) async {
  await Future.delayed(duration);
}

/// 重试服务
class RetryService {
  static Future<T> retry<T>({
    required Future<T> Function() action,
    int maxAttempts = 3,
    Duration delayBetweenAttempts = const Duration(seconds: 1),
    bool Function(Exception)? shouldRetry,
  }) async {
    Exception? lastException;

    for (int i = 0; i < maxAttempts; i++) {
      try {
        return await action();
      } catch (e) {
        lastException = e as Exception;
        if (shouldRetry != null && !shouldRetry(lastException)) {
          rethrow;
        }
        if (i < maxAttempts - 1) {
          await delay(delayBetweenAttempts);
        }
      }
    }

    throw lastException!;
  }
}
