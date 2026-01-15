import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'app.dart';
import 'core/theme/theme.dart';
import 'core/di/providers.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  
  // Initialize Hive
  // await Hive.initFlutter();
  
  runApp(
    ProviderScope(
      overrides: [],
      child: const TortoiseApp(),
    ),
  );
}
