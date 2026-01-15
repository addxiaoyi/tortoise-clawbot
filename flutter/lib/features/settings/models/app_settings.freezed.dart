// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'app_settings.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

T _$identity<T>(T value) => value;

final _privateConstructorUsedError = UnsupportedError(
    'It seems like you constructed your class using `MyClass._()`. This constructor is only meant to be used by freezed and you are not supposed to need it nor use it.\nPlease check the documentation here for more information: https://github.com/rrousselGit/freezed#adding-getters-and-methods-to-our-models');

/// @nodoc
mixin _$AppSettings {
  @ThemeModeConverter()
  ThemeMode get themeMode => throw _privateConstructorUsedError;
  String get apiBaseUrl => throw _privateConstructorUsedError;
  String get wsUrl => throw _privateConstructorUsedError;
  String get aiProvider => throw _privateConstructorUsedError;
  String get aiModel => throw _privateConstructorUsedError;
  double get temperature => throw _privateConstructorUsedError;
  int get maxTokens => throw _privateConstructorUsedError;
  Map<String, String> get apiKeys => throw _privateConstructorUsedError;
  Map<String, ChannelConfig> get channels => throw _privateConstructorUsedError;
  Map<String, bool> get plugins => throw _privateConstructorUsedError;
  bool get autoConnect => throw _privateConstructorUsedError;
  bool get showTimestamps => throw _privateConstructorUsedError;
  bool get enableNotifications => throw _privateConstructorUsedError;
  bool get enableSounds => throw _privateConstructorUsedError;
  String get memoryStorage =>
      throw _privateConstructorUsedError; // local, redis, supabase
  String get activeSessionId => throw _privateConstructorUsedError;

  @JsonKey(ignore: true)
  $AppSettingsCopyWith<AppSettings> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $AppSettingsCopyWith<$Res> {
  factory $AppSettingsCopyWith(
          AppSettings value, $Res Function(AppSettings) then) =
      _$AppSettingsCopyWithImpl<$Res, AppSettings>;
  @useResult
  $Res call(
      {@ThemeModeConverter() ThemeMode themeMode,
      String apiBaseUrl,
      String wsUrl,
      String aiProvider,
      String aiModel,
      double temperature,
      int maxTokens,
      Map<String, String> apiKeys,
      Map<String, ChannelConfig> channels,
      Map<String, bool> plugins,
      bool autoConnect,
      bool showTimestamps,
      bool enableNotifications,
      bool enableSounds,
      String memoryStorage,
      String activeSessionId});
}

/// @nodoc
class _$AppSettingsCopyWithImpl<$Res, $Val extends AppSettings>
    implements $AppSettingsCopyWith<$Res> {
  _$AppSettingsCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? themeMode = null,
    Object? apiBaseUrl = null,
    Object? wsUrl = null,
    Object? aiProvider = null,
    Object? aiModel = null,
    Object? temperature = null,
    Object? maxTokens = null,
    Object? apiKeys = null,
    Object? channels = null,
    Object? plugins = null,
    Object? autoConnect = null,
    Object? showTimestamps = null,
    Object? enableNotifications = null,
    Object? enableSounds = null,
    Object? memoryStorage = null,
    Object? activeSessionId = null,
  }) {
    return _then(_value.copyWith(
      themeMode: null == themeMode
          ? _value.themeMode
          : themeMode // ignore: cast_nullable_to_non_nullable
              as ThemeMode,
      apiBaseUrl: null == apiBaseUrl
          ? _value.apiBaseUrl
          : apiBaseUrl // ignore: cast_nullable_to_non_nullable
              as String,
      wsUrl: null == wsUrl
          ? _value.wsUrl
          : wsUrl // ignore: cast_nullable_to_non_nullable
              as String,
      aiProvider: null == aiProvider
          ? _value.aiProvider
          : aiProvider // ignore: cast_nullable_to_non_nullable
              as String,
      aiModel: null == aiModel
          ? _value.aiModel
          : aiModel // ignore: cast_nullable_to_non_nullable
              as String,
      temperature: null == temperature
          ? _value.temperature
          : temperature // ignore: cast_nullable_to_non_nullable
              as double,
      maxTokens: null == maxTokens
          ? _value.maxTokens
          : maxTokens // ignore: cast_nullable_to_non_nullable
              as int,
      apiKeys: null == apiKeys
          ? _value.apiKeys
          : apiKeys // ignore: cast_nullable_to_non_nullable
              as Map<String, String>,
      channels: null == channels
          ? _value.channels
          : channels // ignore: cast_nullable_to_non_nullable
              as Map<String, ChannelConfig>,
      plugins: null == plugins
          ? _value.plugins
          : plugins // ignore: cast_nullable_to_non_nullable
              as Map<String, bool>,
      autoConnect: null == autoConnect
          ? _value.autoConnect
          : autoConnect // ignore: cast_nullable_to_non_nullable
              as bool,
      showTimestamps: null == showTimestamps
          ? _value.showTimestamps
          : showTimestamps // ignore: cast_nullable_to_non_nullable
              as bool,
      enableNotifications: null == enableNotifications
          ? _value.enableNotifications
          : enableNotifications // ignore: cast_nullable_to_non_nullable
              as bool,
      enableSounds: null == enableSounds
          ? _value.enableSounds
          : enableSounds // ignore: cast_nullable_to_non_nullable
              as bool,
      memoryStorage: null == memoryStorage
          ? _value.memoryStorage
          : memoryStorage // ignore: cast_nullable_to_non_nullable
              as String,
      activeSessionId: null == activeSessionId
          ? _value.activeSessionId
          : activeSessionId // ignore: cast_nullable_to_non_nullable
              as String,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$AppSettingsImplCopyWith<$Res>
    implements $AppSettingsCopyWith<$Res> {
  factory _$$AppSettingsImplCopyWith(
          _$AppSettingsImpl value, $Res Function(_$AppSettingsImpl) then) =
      __$$AppSettingsImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {@ThemeModeConverter() ThemeMode themeMode,
      String apiBaseUrl,
      String wsUrl,
      String aiProvider,
      String aiModel,
      double temperature,
      int maxTokens,
      Map<String, String> apiKeys,
      Map<String, ChannelConfig> channels,
      Map<String, bool> plugins,
      bool autoConnect,
      bool showTimestamps,
      bool enableNotifications,
      bool enableSounds,
      String memoryStorage,
      String activeSessionId});
}

/// @nodoc
class __$$AppSettingsImplCopyWithImpl<$Res>
    extends _$AppSettingsCopyWithImpl<$Res, _$AppSettingsImpl>
    implements _$$AppSettingsImplCopyWith<$Res> {
  __$$AppSettingsImplCopyWithImpl(
      _$AppSettingsImpl _value, $Res Function(_$AppSettingsImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? themeMode = null,
    Object? apiBaseUrl = null,
    Object? wsUrl = null,
    Object? aiProvider = null,
    Object? aiModel = null,
    Object? temperature = null,
    Object? maxTokens = null,
    Object? apiKeys = null,
    Object? channels = null,
    Object? plugins = null,
    Object? autoConnect = null,
    Object? showTimestamps = null,
    Object? enableNotifications = null,
    Object? enableSounds = null,
    Object? memoryStorage = null,
    Object? activeSessionId = null,
  }) {
    return _then(_$AppSettingsImpl(
      themeMode: null == themeMode
          ? _value.themeMode
          : themeMode // ignore: cast_nullable_to_non_nullable
              as ThemeMode,
      apiBaseUrl: null == apiBaseUrl
          ? _value.apiBaseUrl
          : apiBaseUrl // ignore: cast_nullable_to_non_nullable
              as String,
      wsUrl: null == wsUrl
          ? _value.wsUrl
          : wsUrl // ignore: cast_nullable_to_non_nullable
              as String,
      aiProvider: null == aiProvider
          ? _value.aiProvider
          : aiProvider // ignore: cast_nullable_to_non_nullable
              as String,
      aiModel: null == aiModel
          ? _value.aiModel
          : aiModel // ignore: cast_nullable_to_non_nullable
              as String,
      temperature: null == temperature
          ? _value.temperature
          : temperature // ignore: cast_nullable_to_non_nullable
              as double,
      maxTokens: null == maxTokens
          ? _value.maxTokens
          : maxTokens // ignore: cast_nullable_to_non_nullable
              as int,
      apiKeys: null == apiKeys
          ? _value._apiKeys
          : apiKeys // ignore: cast_nullable_to_non_nullable
              as Map<String, String>,
      channels: null == channels
          ? _value._channels
          : channels // ignore: cast_nullable_to_non_nullable
              as Map<String, ChannelConfig>,
      plugins: null == plugins
          ? _value._plugins
          : plugins // ignore: cast_nullable_to_non_nullable
              as Map<String, bool>,
      autoConnect: null == autoConnect
          ? _value.autoConnect
          : autoConnect // ignore: cast_nullable_to_non_nullable
              as bool,
      showTimestamps: null == showTimestamps
          ? _value.showTimestamps
          : showTimestamps // ignore: cast_nullable_to_non_nullable
              as bool,
      enableNotifications: null == enableNotifications
          ? _value.enableNotifications
          : enableNotifications // ignore: cast_nullable_to_non_nullable
              as bool,
      enableSounds: null == enableSounds
          ? _value.enableSounds
          : enableSounds // ignore: cast_nullable_to_non_nullable
              as bool,
      memoryStorage: null == memoryStorage
          ? _value.memoryStorage
          : memoryStorage // ignore: cast_nullable_to_non_nullable
              as String,
      activeSessionId: null == activeSessionId
          ? _value.activeSessionId
          : activeSessionId // ignore: cast_nullable_to_non_nullable
              as String,
    ));
  }
}

/// @nodoc

class _$AppSettingsImpl implements _AppSettings {
  const _$AppSettingsImpl(
      {@ThemeModeConverter() this.themeMode = ThemeMode.system,
      this.apiBaseUrl = 'http://localhost:8080',
      this.wsUrl = 'ws://localhost:8080/ws',
      this.aiProvider = 'openai',
      this.aiModel = 'gpt-4-turbo-preview',
      this.temperature = 0.7,
      this.maxTokens = 4000,
      final Map<String, String> apiKeys = const {},
      final Map<String, ChannelConfig> channels = const {},
      final Map<String, bool> plugins = const {},
      this.autoConnect = false,
      this.showTimestamps = true,
      this.enableNotifications = true,
      this.enableSounds = true,
      this.memoryStorage = 'local',
      this.activeSessionId = 'main'})
      : _apiKeys = apiKeys,
        _channels = channels,
        _plugins = plugins;

  @override
  @JsonKey()
  @ThemeModeConverter()
  final ThemeMode themeMode;
  @override
  @JsonKey()
  final String apiBaseUrl;
  @override
  @JsonKey()
  final String wsUrl;
  @override
  @JsonKey()
  final String aiProvider;
  @override
  @JsonKey()
  final String aiModel;
  @override
  @JsonKey()
  final double temperature;
  @override
  @JsonKey()
  final int maxTokens;
  final Map<String, String> _apiKeys;
  @override
  @JsonKey()
  Map<String, String> get apiKeys {
    if (_apiKeys is EqualUnmodifiableMapView) return _apiKeys;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableMapView(_apiKeys);
  }

  final Map<String, ChannelConfig> _channels;
  @override
  @JsonKey()
  Map<String, ChannelConfig> get channels {
    if (_channels is EqualUnmodifiableMapView) return _channels;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableMapView(_channels);
  }

  final Map<String, bool> _plugins;
  @override
  @JsonKey()
  Map<String, bool> get plugins {
    if (_plugins is EqualUnmodifiableMapView) return _plugins;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableMapView(_plugins);
  }

  @override
  @JsonKey()
  final bool autoConnect;
  @override
  @JsonKey()
  final bool showTimestamps;
  @override
  @JsonKey()
  final bool enableNotifications;
  @override
  @JsonKey()
  final bool enableSounds;
  @override
  @JsonKey()
  final String memoryStorage;
// local, redis, supabase
  @override
  @JsonKey()
  final String activeSessionId;

  @override
  String toString() {
    return 'AppSettings(themeMode: $themeMode, apiBaseUrl: $apiBaseUrl, wsUrl: $wsUrl, aiProvider: $aiProvider, aiModel: $aiModel, temperature: $temperature, maxTokens: $maxTokens, apiKeys: $apiKeys, channels: $channels, plugins: $plugins, autoConnect: $autoConnect, showTimestamps: $showTimestamps, enableNotifications: $enableNotifications, enableSounds: $enableSounds, memoryStorage: $memoryStorage, activeSessionId: $activeSessionId)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$AppSettingsImpl &&
            (identical(other.themeMode, themeMode) ||
                other.themeMode == themeMode) &&
            (identical(other.apiBaseUrl, apiBaseUrl) ||
                other.apiBaseUrl == apiBaseUrl) &&
            (identical(other.wsUrl, wsUrl) || other.wsUrl == wsUrl) &&
            (identical(other.aiProvider, aiProvider) ||
                other.aiProvider == aiProvider) &&
            (identical(other.aiModel, aiModel) || other.aiModel == aiModel) &&
            (identical(other.temperature, temperature) ||
                other.temperature == temperature) &&
            (identical(other.maxTokens, maxTokens) ||
                other.maxTokens == maxTokens) &&
            const DeepCollectionEquality().equals(other._apiKeys, _apiKeys) &&
            const DeepCollectionEquality().equals(other._channels, _channels) &&
            const DeepCollectionEquality().equals(other._plugins, _plugins) &&
            (identical(other.autoConnect, autoConnect) ||
                other.autoConnect == autoConnect) &&
            (identical(other.showTimestamps, showTimestamps) ||
                other.showTimestamps == showTimestamps) &&
            (identical(other.enableNotifications, enableNotifications) ||
                other.enableNotifications == enableNotifications) &&
            (identical(other.enableSounds, enableSounds) ||
                other.enableSounds == enableSounds) &&
            (identical(other.memoryStorage, memoryStorage) ||
                other.memoryStorage == memoryStorage) &&
            (identical(other.activeSessionId, activeSessionId) ||
                other.activeSessionId == activeSessionId));
  }

  @override
  int get hashCode => Object.hash(
      runtimeType,
      themeMode,
      apiBaseUrl,
      wsUrl,
      aiProvider,
      aiModel,
      temperature,
      maxTokens,
      const DeepCollectionEquality().hash(_apiKeys),
      const DeepCollectionEquality().hash(_channels),
      const DeepCollectionEquality().hash(_plugins),
      autoConnect,
      showTimestamps,
      enableNotifications,
      enableSounds,
      memoryStorage,
      activeSessionId);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$AppSettingsImplCopyWith<_$AppSettingsImpl> get copyWith =>
      __$$AppSettingsImplCopyWithImpl<_$AppSettingsImpl>(this, _$identity);
}

abstract class _AppSettings implements AppSettings {
  const factory _AppSettings(
      {@ThemeModeConverter() final ThemeMode themeMode,
      final String apiBaseUrl,
      final String wsUrl,
      final String aiProvider,
      final String aiModel,
      final double temperature,
      final int maxTokens,
      final Map<String, String> apiKeys,
      final Map<String, ChannelConfig> channels,
      final Map<String, bool> plugins,
      final bool autoConnect,
      final bool showTimestamps,
      final bool enableNotifications,
      final bool enableSounds,
      final String memoryStorage,
      final String activeSessionId}) = _$AppSettingsImpl;

  @override
  @ThemeModeConverter()
  ThemeMode get themeMode;
  @override
  String get apiBaseUrl;
  @override
  String get wsUrl;
  @override
  String get aiProvider;
  @override
  String get aiModel;
  @override
  double get temperature;
  @override
  int get maxTokens;
  @override
  Map<String, String> get apiKeys;
  @override
  Map<String, ChannelConfig> get channels;
  @override
  Map<String, bool> get plugins;
  @override
  bool get autoConnect;
  @override
  bool get showTimestamps;
  @override
  bool get enableNotifications;
  @override
  bool get enableSounds;
  @override
  String get memoryStorage;
  @override // local, redis, supabase
  String get activeSessionId;
  @override
  @JsonKey(ignore: true)
  _$$AppSettingsImplCopyWith<_$AppSettingsImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
mixin _$ChannelConfig {
  bool get enabled => throw _privateConstructorUsedError;
  String get botToken => throw _privateConstructorUsedError;
  String get apiKey => throw _privateConstructorUsedError;
  String get webhookUrl => throw _privateConstructorUsedError;
  Map<String, dynamic> get settings => throw _privateConstructorUsedError;

  @JsonKey(ignore: true)
  $ChannelConfigCopyWith<ChannelConfig> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $ChannelConfigCopyWith<$Res> {
  factory $ChannelConfigCopyWith(
          ChannelConfig value, $Res Function(ChannelConfig) then) =
      _$ChannelConfigCopyWithImpl<$Res, ChannelConfig>;
  @useResult
  $Res call(
      {bool enabled,
      String botToken,
      String apiKey,
      String webhookUrl,
      Map<String, dynamic> settings});
}

/// @nodoc
class _$ChannelConfigCopyWithImpl<$Res, $Val extends ChannelConfig>
    implements $ChannelConfigCopyWith<$Res> {
  _$ChannelConfigCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? enabled = null,
    Object? botToken = null,
    Object? apiKey = null,
    Object? webhookUrl = null,
    Object? settings = null,
  }) {
    return _then(_value.copyWith(
      enabled: null == enabled
          ? _value.enabled
          : enabled // ignore: cast_nullable_to_non_nullable
              as bool,
      botToken: null == botToken
          ? _value.botToken
          : botToken // ignore: cast_nullable_to_non_nullable
              as String,
      apiKey: null == apiKey
          ? _value.apiKey
          : apiKey // ignore: cast_nullable_to_non_nullable
              as String,
      webhookUrl: null == webhookUrl
          ? _value.webhookUrl
          : webhookUrl // ignore: cast_nullable_to_non_nullable
              as String,
      settings: null == settings
          ? _value.settings
          : settings // ignore: cast_nullable_to_non_nullable
              as Map<String, dynamic>,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$ChannelConfigImplCopyWith<$Res>
    implements $ChannelConfigCopyWith<$Res> {
  factory _$$ChannelConfigImplCopyWith(
          _$ChannelConfigImpl value, $Res Function(_$ChannelConfigImpl) then) =
      __$$ChannelConfigImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {bool enabled,
      String botToken,
      String apiKey,
      String webhookUrl,
      Map<String, dynamic> settings});
}

/// @nodoc
class __$$ChannelConfigImplCopyWithImpl<$Res>
    extends _$ChannelConfigCopyWithImpl<$Res, _$ChannelConfigImpl>
    implements _$$ChannelConfigImplCopyWith<$Res> {
  __$$ChannelConfigImplCopyWithImpl(
      _$ChannelConfigImpl _value, $Res Function(_$ChannelConfigImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? enabled = null,
    Object? botToken = null,
    Object? apiKey = null,
    Object? webhookUrl = null,
    Object? settings = null,
  }) {
    return _then(_$ChannelConfigImpl(
      enabled: null == enabled
          ? _value.enabled
          : enabled // ignore: cast_nullable_to_non_nullable
              as bool,
      botToken: null == botToken
          ? _value.botToken
          : botToken // ignore: cast_nullable_to_non_nullable
              as String,
      apiKey: null == apiKey
          ? _value.apiKey
          : apiKey // ignore: cast_nullable_to_non_nullable
              as String,
      webhookUrl: null == webhookUrl
          ? _value.webhookUrl
          : webhookUrl // ignore: cast_nullable_to_non_nullable
              as String,
      settings: null == settings
          ? _value._settings
          : settings // ignore: cast_nullable_to_non_nullable
              as Map<String, dynamic>,
    ));
  }
}

/// @nodoc

class _$ChannelConfigImpl implements _ChannelConfig {
  const _$ChannelConfigImpl(
      {this.enabled = false,
      this.botToken = '',
      this.apiKey = '',
      this.webhookUrl = '',
      final Map<String, dynamic> settings = const {}})
      : _settings = settings;

  @override
  @JsonKey()
  final bool enabled;
  @override
  @JsonKey()
  final String botToken;
  @override
  @JsonKey()
  final String apiKey;
  @override
  @JsonKey()
  final String webhookUrl;
  final Map<String, dynamic> _settings;
  @override
  @JsonKey()
  Map<String, dynamic> get settings {
    if (_settings is EqualUnmodifiableMapView) return _settings;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableMapView(_settings);
  }

  @override
  String toString() {
    return 'ChannelConfig(enabled: $enabled, botToken: $botToken, apiKey: $apiKey, webhookUrl: $webhookUrl, settings: $settings)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$ChannelConfigImpl &&
            (identical(other.enabled, enabled) || other.enabled == enabled) &&
            (identical(other.botToken, botToken) ||
                other.botToken == botToken) &&
            (identical(other.apiKey, apiKey) || other.apiKey == apiKey) &&
            (identical(other.webhookUrl, webhookUrl) ||
                other.webhookUrl == webhookUrl) &&
            const DeepCollectionEquality().equals(other._settings, _settings));
  }

  @override
  int get hashCode => Object.hash(runtimeType, enabled, botToken, apiKey,
      webhookUrl, const DeepCollectionEquality().hash(_settings));

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$ChannelConfigImplCopyWith<_$ChannelConfigImpl> get copyWith =>
      __$$ChannelConfigImplCopyWithImpl<_$ChannelConfigImpl>(this, _$identity);
}

abstract class _ChannelConfig implements ChannelConfig {
  const factory _ChannelConfig(
      {final bool enabled,
      final String botToken,
      final String apiKey,
      final String webhookUrl,
      final Map<String, dynamic> settings}) = _$ChannelConfigImpl;

  @override
  bool get enabled;
  @override
  String get botToken;
  @override
  String get apiKey;
  @override
  String get webhookUrl;
  @override
  Map<String, dynamic> get settings;
  @override
  @JsonKey(ignore: true)
  _$$ChannelConfigImplCopyWith<_$ChannelConfigImpl> get copyWith =>
      throw _privateConstructorUsedError;
}
