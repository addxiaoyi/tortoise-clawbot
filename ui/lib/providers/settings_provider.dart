import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

class SettingsProvider extends ChangeNotifier {
  ThemeMode _themeMode = ThemeMode.dark;
  String _language = 'en';
  bool _p2pEnabled = true;
  bool _encryptionEnabled = true;

  ThemeMode get themeMode => _themeMode;
  String get language => _language;
  bool get p2pEnabled => _p2pEnabled;
  bool get encryptionEnabled => _encryptionEnabled;

  SettingsProvider() {
    _loadSettings();
  }

  Future<void> _loadSettings() async {
    final prefs = await SharedPreferences.getInstance();
    _themeMode = ThemeMode.values[prefs.getInt('themeMode') ?? 2];
    _language = prefs.getString('language') ?? 'en';
    _p2pEnabled = prefs.getBool('p2pEnabled') ?? true;
    _encryptionEnabled = prefs.getBool('encryptionEnabled') ?? true;
    notifyListeners();
  }

  Future<void> setThemeMode(ThemeMode mode) async {
    _themeMode = mode;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setInt('themeMode', mode.index);
    notifyListeners();
  }

  Future<void> setLanguage(String lang) async {
    _language = lang;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('language', lang);
    notifyListeners();
  }

  Future<void> setP2PEnabled(bool enabled) async {
    _p2pEnabled = enabled;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool('p2pEnabled', enabled);
    notifyListeners();
  }

  Future<void> setEncryptionEnabled(bool enabled) async {
    _encryptionEnabled = enabled;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool('encryptionEnabled', enabled);
    notifyListeners();
  }
}
