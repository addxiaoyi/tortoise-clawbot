import 'package:hive_flutter/hive_flutter.dart';

class StorageService {
  static const String settingsBox = 'settings';
  static const String sessionsBox = 'sessions';
  static const String memoriesBox = 'memories';
  static const String cacheBox = 'cache';
  
  static late Box _settingsBox;
  static late Box _sessionsBox;
  static late Box _memoriesBox;
  static late Box _cacheBox;
  
  static Future<void> init() async {
    _settingsBox = await Hive.openBox(settingsBox);
    _sessionsBox = await Hive.openBox(sessionsBox);
    _memoriesBox = await Hive.openBox(memoriesBox);
    _cacheBox = await Hive.openBox(cacheBox);
  }
  
  // 设置操作
  static Box get settings => _settingsBox;
  static Box get sessions => _sessionsBox;
  static Box get memories => _memoriesBox;
  static Box get cache => _cacheBox;
  
  // 通用方法
  static Future<void> put(String box, String key, dynamic value) async {
    final b = _getBox(box);
    await b.put(key, value);
  }
  
  static dynamic get(String box, String key, {dynamic defaultValue}) {
    final b = _getBox(box);
    return b.get(key, defaultValue: defaultValue);
  }
  
  static Future<void> delete(String box, String key) async {
    final b = _getBox(box);
    await b.delete(key);
  }
  
  static Future<void> clear(String box) async {
    final b = _getBox(box);
    await b.clear();
  }
  
  static Box _getBox(String name) {
    switch (name) {
      case settingsBox:
        return _settingsBox;
      case sessionsBox:
        return _sessionsBox;
      case memoriesBox:
        return _memoriesBox;
      case cacheBox:
        return _cacheBox;
      default:
        return _settingsBox;
    }
  }
}
