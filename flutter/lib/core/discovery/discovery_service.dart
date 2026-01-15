import 'dart:async';
import 'package:dio/dio.dart';

/// 设备发现服务 (Web 兼容版)
class DiscoveryService {
  static DiscoveryService? _instance;
  static DiscoveryService get instance => _instance ??= DiscoveryService._();
  DiscoveryService._();

  final _devicesController = StreamController<List<DiscoveredDevice>>.broadcast();
  final List<DiscoveredDevice> _devices = [];
  Timer? _scanTimer;
  bool _isRunning = false;
  int _port = 18792;

  Stream<List<DiscoveredDevice>> get devicesStream => _devicesController.stream;
  List<DiscoveredDevice> get devices => List.unmodifiable(_devices);
  bool get isRunning => _isRunning;

  Future<void> initialize({int? port}) async {
    _port = port ?? _port;
  }

  Future<void> startDiscovery() async {
    if (_isRunning) return;
    _isRunning = true;
    _startScanning();
    // 定期扫描
    _scanTimer?.cancel();
    _scanTimer = Timer.periodic(const Duration(seconds: 30), (_) => _scan());
  }

  Future<void> stopDiscovery() async {
    _isRunning = false;
    _scanTimer?.cancel();
  }

  Future<void> refresh() async {
    await _scan();
  }

  Future<void> _startScanning() async {
    await _scan();
  }

  Future<void> _scan() async {
    // 扫描本地和常见地址
    final hosts = ['localhost', '127.0.0.1'];
    
    for (final host in hosts) {
      try {
        final dio = Dio();
        final response = await dio.get(
          'http://$host:$_port/health',
          options: Options(
            connectTimeout: const Duration(seconds: 2),
            receiveTimeout: const Duration(seconds: 2),
          ),
        );
        if (response.statusCode == 200) {
          _addOrUpdateDevice(DiscoveredDevice(
            id: '$host:$_port',
            name: 'Tortoise Gateway',
            address: host,
            port: _port,
            type: DeviceType.gateway,
            discoveredAt: DateTime.now(),
          ));
        }
      } catch (_) {
        // 忽略连接失败的 host
      }
    }
    
    // 清理过期设备 (5分钟前的)
    _devices.removeWhere(
      (d) => DateTime.now().difference(d.discoveredAt).inMinutes > 5
    );
    _devicesController.add(List.from(_devices));
  }

  void _addOrUpdateDevice(DiscoveredDevice device) {
    final index = _devices.indexWhere((d) => d.id == device.id);
    if (index >= 0) {
      _devices[index] = device;
    } else {
      _devices.add(device);
    }
  }

  void dispose() {
    stopDiscovery();
    _devicesController.close();
  }
}

/// 发现的设备
class DiscoveredDevice {
  final String id;
  final String name;
  final String address;
  final int port;
  final DeviceType type;
  final DateTime discoveredAt;

  DiscoveredDevice({
    required this.id,
    required this.name,
    required this.address,
    required this.port,
    required this.type,
    required this.discoveredAt,
  });

  String get url => 'http://$address:$port';
}

/// 设备类型
enum DeviceType {
  gateway,
  mobile,
  desktop,
  web,
}
