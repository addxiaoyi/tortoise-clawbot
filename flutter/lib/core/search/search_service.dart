import '../session/session_manager.dart';

/// 搜索服务
class SearchService {
  static SearchService? _instance;
  static SearchService get instance => _instance ??= SearchService._();
  SearchService._();

  /// 搜索会话
  List<SearchResult> searchSessions(String query) {
    if (query.isEmpty) return [];
    
    final sessionManager = SessionManager.instance;
    final results = <SearchResult>[];
    
    for (final session in sessionManager.sessions) {
      // 搜索标题
      if (session.title.toLowerCase().contains(query.toLowerCase())) {
        results.add(SearchResult(
          type: SearchResultType.sessionTitle,
          title: session.title,
          subtitle: '会话标题',
          id: session.id,
        ));
      }
      
      // 搜索消息内容
      for (final message in session.messages) {
        if (message.content.toLowerCase().contains(query.toLowerCase())) {
          results.add(SearchResult(
            type: SearchResultType.messageContent,
            title: session.title,
            subtitle: message.content.length > 100 
                ? '${message.content.substring(0, 100)}...' 
                : message.content,
            id: session.id,
          ));
          break; // 每个会话只显示一次
        }
      }
    }
    
    return results;
  }

  /// 高亮匹配文本
  String highlightMatches(String text, String query) {
    if (query.isEmpty) return text;
    // 在实际应用中，这里会返回带有高亮标记的文本
    return text;
  }
}

/// 搜索结果
class SearchResult {
  final SearchResultType type;
  final String title;
  final String subtitle;
  final String id;

  SearchResult({
    required this.type,
    required this.title,
    required this.subtitle,
    required this.id,
  });
}

/// 搜索结果类型
enum SearchResultType {
  sessionTitle,
  messageContent,
}
