import 'package:hive_flutter/hive_flutter.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// 存储服务 - 统一管理所有持久化存储
class StorageService {
  static StorageService? _instance;
  static StorageService get instance => _instance ??= StorageService._();

  StorageService._();

  // 普通存储
  SharedPreferences? _prefs;
  
  // 安全存储 (敏感数据)
  final _secureStorage = const FlutterSecureStorage();
  
  // Hive Boxes
  final Map<String, Box> _hiveBoxes = {};

  /// 初始化
  Future<void> init() async {
    _prefs = await SharedPreferences.getInstance();
    
    // 打开 Hive Boxes
    _hiveBoxes['sessions'] = await Hive.openBox('sessions');
    _hiveBoxes['messages'] = await Hive.openBox('messages');
    _hiveBoxes['memories'] = await Hive.openBox('memories');
    _hiveBoxes['config'] = await Hive.openBox('config');
    _hiveBoxes['cache'] = await Hive.openBox('cache');
  }

  // ========== SharedPreferences ==========

  String? getString(String key) => _prefs?.getString(key);
  
  Future<void> setString(String key, String value) async {
    await _prefs?.setString(key, value);
  }

  bool? getBool(String key) => _prefs?.getBool(key);
  
  Future<void> setBool(String key, bool value) async {
    await _prefs?.setBool(key, value);
  }

  int? getInt(String key) => _prefs?.getInt(key);
  
  Future<void> setInt(String key, int value) async {
    await _prefs?.setInt(key, value);
  }

  double? getDouble(String key) => _prefs?.getDouble(key);
  
  Future<void> setDouble(String key, double value) async {
    await _prefs?.setDouble(key, value);
  }

  Future<void> remove(String key) async {
    await _prefs?.remove(key);
  }

  Future<void> clear() async {
    await _prefs?.clear();
  }

  // ========== Secure Storage ==========

  Future<void> setSecure(String key, String value) async {
    await _secureStorage.write(key: key, value: value);
  }

  Future<String?> getSecure(String key) async {
    return await _secureStorage.read(key: key);
  }

  Future<void> deleteSecure(String key) async {
    await _secureStorage.delete(key: key);
  }

  // ========== API Keys ==========

  Future<void> saveApiKeys({String? openai, String? anthropic, String? telegram, String? discord, String? slack}) async {
    if (openai != null) await setSecure('openai_api_key', openai);
    if (anthropic != null) await setSecure('anthropic_api_key', anthropic);
    if (telegram != null) await setSecure('telegram_bot_token', telegram);
    if (discord != null) await setSecure('discord_bot_token', discord);
    if (slack != null) await setSecure('slack_bot_token', slack);
  }

  Future<Map<String, String>> getApiKeys() async {
    final tokens = <String, String>{};
    final openai = await getSecure('openai_api_key');
    final anthropic = await getSecure('anthropic_api_key');
    final telegram = await getSecure('telegram_bot_token');
    final discord = await getSecure('discord_bot_token');
    final slack = await getSecure('slack_bot_token');
    if (openai != null) tokens['openai'] = openai;
    if (anthropic != null) tokens['anthropic'] = anthropic;
    if (telegram != null) tokens['telegram'] = telegram;
    if (discord != null) tokens['discord'] = discord;
    if (slack != null) tokens['slack'] = slack;
    return tokens;
  }

  // ========== Sessions ==========

  Future<void> saveSession(String id, Map<String, dynamic> data) async {
    final box = _hiveBoxes['sessions'];
    await box?.put(id, data);
  }

  Map<String, dynamic>? getSession(String id) {
    final box = _hiveBoxes['sessions'];
    final data = box?.get(id);
    if (data == null) return null;
    return Map<String, dynamic>.from(data is Map ? data : {});
  }

  List<Map<String, dynamic>> getAllSessions() {
    final box = _hiveBoxes['sessions'];
    if (box == null) return [];
    return box.values
        .whereType<Map>()
        .map((v) => Map<String, dynamic>.from(v as Map))
        .toList();
  }

  Future<void> deleteSession(String id) async {
    final box = _hiveBoxes['sessions'];
    await box?.delete(id);
  }

  // ========== Messages ==========

  Future<void> saveMessage(String sessionId, Map<String, dynamic> message) async {
    final box = _hiveBoxes['messages'];
    final messages = _getMessageList(box, sessionId);
    messages.add(message);
    await box?.put(sessionId, messages);
  }

  List<Map<String, dynamic>> getMessages(String sessionId) {
    final box = _hiveBoxes['messages'];
    return _getMessageList(box, sessionId);
  }

  List<Map<String, dynamic>> _getMessageList(Box? box, String sessionId) {
    final data = box?.get(sessionId);
    if (data == null) return [];
    if (data is List) {
      return data.whereType<Map>().map((v) => Map<String, dynamic>.from(v as Map)).toList();
    }
    return [];
  }

  Future<void> clearMessages(String sessionId) async {
    final box = _hiveBoxes['messages'];
    await box?.delete(sessionId);
  }

  // ========== Config ==========

  Future<void> saveConfig(String key, dynamic value) async {
    final box = _hiveBoxes['config'];
    await box?.put(key, value);
  }

  T? getConfig<T>(String key) {
    final box = _hiveBoxes['config'];
    return box?.get(key) as T?;
  }

  // ========== Cache ==========

  Future<void> saveCache(String key, dynamic value) async {
    final box = _hiveBoxes['cache'];
    await box?.put(key, value);
  }

  T? getCache<T>(String key) {
    final box = _hiveBoxes['cache'];
    return box?.get(key) as T?;
  }

  Future<void> clearCache() async {
    await _hiveBoxes['cache']?.clear();
  }

  // ========== Cleanup ==========

  Future<void> dispose() async {
    await _prefs?.clear();
    await _secureStorage.deleteAll();
    for (final box in _hiveBoxes.values) {
      await box.close();
    }
    _hiveBoxes.clear();
  }
}
