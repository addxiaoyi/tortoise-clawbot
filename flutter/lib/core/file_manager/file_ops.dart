import 'dart:io';

/// File Operations - 文件操作
/// 本地文件管理和同步
class FileManager {
  /// 读取文件
  Future<String?> readFile(String path) async {
    try {
      final file = File(path);
      if (await file.exists()) {
        return await file.readAsString();
      }
    } catch (e) {
      // 错误处理
    }
    return null;
  }

  /// 写入文件
  Future<bool> writeFile(String path, String content) async {
    try {
      final file = File(path);
      await file.writeAsString(content);
      return true;
    } catch (e) {
      return false;
    }
  }

  /// 删除文件
  Future<bool> deleteFile(String path) async {
    try {
      final file = File(path);
      if (await file.exists()) {
        await file.delete();
        return true;
      }
    } catch (e) {
      // 错误处理
    }
    return false;
  }

  /// 列出目录
  Future<List<FileInfo>> listDirectory(String path, {bool recursive = false}) async {
    final dir = Directory(path);
    if (!await dir.exists()) return [];

    final files = <FileInfo>[];
    await for (final entity in dir.list(recursive: recursive)) {
      final stat = await entity.stat();
      files.add(FileInfo(
        name: entity.path.split(Platform.pathSeparator).last,
        path: entity.path,
        isDirectory: entity is Directory,
        size: stat.size,
        modified: stat.modified,
      ));
    }
    return files;
  }

  /// 复制文件
  Future<bool> copyFile(String source, String dest) async {
    try {
      final file = File(source);
      await file.copy(dest);
      return true;
    } catch (e) {
      return false;
    }
  }

  /// 移动文件
  Future<bool> moveFile(String source, String dest) async {
    try {
      final file = File(source);
      await file.rename(dest);
      return true;
    } catch (e) {
      return false;
    }
  }

  /// 创建目录
  Future<bool> createDirectory(String path) async {
    try {
      final dir = Directory(path);
      await dir.create(recursive: true);
      return true;
    } catch (e) {
      return false;
    }
  }

  /// 检查存在
  Future<bool> exists(String path) async {
    final file = File(path);
    return await file.exists();
  }

  /// 获取文件大小
  Future<int> getSize(String path) async {
    final file = File(path);
    if (await file.exists()) {
      return await file.length();
    }
    return 0;
  }
}

/// 文件信息
class FileInfo {
  final String name;
  final String path;
  final bool isDirectory;
  final int size;
  final DateTime modified;

  FileInfo({
    required this.name,
    required this.path,
    required this.isDirectory,
    required this.size,
    required this.modified,
  });
}
