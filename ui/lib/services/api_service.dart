import 'dart:convert';
import 'package:http/http.dart' as http;

class ApiService {
  static const String baseUrl = 'http://localhost:8080/api';
  
  final http.Client _client;

  ApiService({http.Client? client}) : _client = client ?? http.Client();

  // Agent APIs
  Future<Map<String, dynamic>> getAgentStatus() async {
    final response = await _client.get(Uri.parse('$baseUrl/agent/status'));
    return json.decode(response.body);
  }

  Future<Map<String, dynamic>> sendMessage(String message) async {
    final response = await _client.post(
      Uri.parse('$baseUrl/agent/chat'),
      headers: {'Content-Type': 'application/json'},
      body: json.encode({'message': message}),
    );
    return json.decode(response.body);
  }

  Future<List<Map<String, dynamic>>> getConversations() async {
    final response = await _client.get(Uri.parse('$baseUrl/conversations'));
    return List<Map<String, dynamic>>.from(json.decode(response.body));
  }

  Future<Map<String, dynamic>> getConversation(String id) async {
    final response = await _client.get(Uri.parse('$baseUrl/conversations/$id'));
    return json.decode(response.body);
  }

  Future<Map<String, dynamic>> deleteConversation(String id) async {
    final response = await _client.delete(Uri.parse('$baseUrl/conversations/$id'));
    return json.decode(response.body);
  }

  // Channel APIs
  Future<List<Map<String, dynamic>>> getChannels() async {
    final response = await _client.get(Uri.parse('$baseUrl/channels'));
    return List<Map<String, dynamic>>.from(json.decode(response.body));
  }

  Future<Map<String, dynamic>> connectChannel(String id) async {
    final response = await _client.post(Uri.parse('$baseUrl/channels/$id/connect'));
    return json.decode(response.body);
  }

  Future<Map<String, dynamic>> disconnectChannel(String id) async {
    final response = await _client.post(Uri.parse('$baseUrl/channels/$id/disconnect'));
    return json.decode(response.body);
  }

  // Plugin APIs
  Future<List<Map<String, dynamic>>> getPlugins() async {
    final response = await _client.get(Uri.parse('$baseUrl/plugins'));
    return List<Map<String, dynamic>>.from(json.decode(response.body));
  }

  Future<Map<String, dynamic>> installPlugin(String id) async {
    final response = await _client.post(Uri.parse('$baseUrl/plugins/$id/install'));
    return json.decode(response.body);
  }

  Future<Map<String, dynamic>> uninstallPlugin(String id) async {
    final response = await _client.delete(Uri.parse('$baseUrl/plugins/$id'));
    return json.decode(response.body);
  }

  Future<Map<String, dynamic>> enablePlugin(String id) async {
    final response = await _client.post(Uri.parse('$baseUrl/plugins/$id/enable'));
    return json.decode(response.body);
  }

  Future<Map<String, dynamic>> disablePlugin(String id) async {
    final response = await _client.post(Uri.parse('$baseUrl/plugins/$id/disable'));
    return json.decode(response.body);
  }

  // Memory APIs
  Future<List<Map<String, dynamic>>> getMemories() async {
    final response = await _client.get(Uri.parse('$baseUrl/memory'));
    return List<Map<String, dynamic>>.from(json.decode(response.body));
  }

  Future<Map<String, dynamic>> addMemory(String content, double importance) async {
    final response = await _client.post(
      Uri.parse('$baseUrl/memory'),
      headers: {'Content-Type': 'application/json'},
      body: json.encode({'content': content, 'importance': importance}),
    );
    return json.decode(response.body);
  }

  Future<Map<String, dynamic>> deleteMemory(String id) async {
    final response = await _client.delete(Uri.parse('$baseUrl/memory/$id'));
    return json.decode(response.body);
  }

  Future<List<Map<String, dynamic>>> searchMemories(String query) async {
    final response = await _client.get(
      Uri.parse('$baseUrl/memory/search?q=${Uri.encodeComponent(query)}'),
    );
    return List<Map<String, dynamic>>.from(json.decode(response.body));
  }

  // Settings APIs
  Future<Map<String, dynamic>> getSettings() async {
    final response = await _client.get(Uri.parse('$baseUrl/settings'));
    return json.decode(response.body);
  }

  Future<Map<String, dynamic>> updateSettings(Map<String, dynamic> settings) async {
    final response = await _client.put(
      Uri.parse('$baseUrl/settings'),
      headers: {'Content-Type': 'application/json'},
      body: json.encode(settings),
    );
    return json.decode(response.body);
  }

  // Model APIs
  Future<List<Map<String, dynamic>>> getAvailableModels() async {
    final response = await _client.get(Uri.parse('$baseUrl/models'));
    return List<Map<String, dynamic>>.from(json.decode(response.body));
  }

  Future<Map<String, dynamic>> switchModel(String modelId) async {
    final response = await _client.post(
      Uri.parse('$baseUrl/models/switch'),
      headers: {'Content-Type': 'application/json'},
      body: json.encode({'model': modelId}),
    );
    return json.decode(response.body);
  }

  // Stats APIs
  Future<Map<String, dynamic>> getStats() async {
    final response = await _client.get(Uri.parse('$baseUrl/stats'));
    return json.decode(response.body);
  }

  Future<Map<String, dynamic>> getMemoryStats() async {
    final response = await _client.get(Uri.parse('$baseUrl/stats/memory'));
    return json.decode(response.body);
  }

  Future<Map<String, dynamic>> getUsageStats() async {
    final response = await _client.get(Uri.parse('$baseUrl/stats/usage'));
    return json.decode(response.body);
  }
}
