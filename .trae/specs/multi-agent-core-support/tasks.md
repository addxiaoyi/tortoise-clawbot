# Tasks

## Task 1: 扩展 useOnboardingStore 支持核心对象选择
- [ ] SubTask 1.1: 在 OnboardingState 中新增 `coreAgent: "openclaw" | "myclaw" | "nanobot" | null` 类型
- [ ] SubTask 1.2: 新增 `setCoreAgent` 方法
- [ ] SubTask 1.3: 更新 `completeOnboarding` 方法参数，接收 coreAgent 和 mode

## Task 2: 重构 SplashScreen 组件 - 核心对象选择页
- [ ] SubTask 2.1: 将现有的 ModeCard 模式选择逻辑拆分为三个阶段：入场动画 → 核心对象选择 → 模式选择
- [ ] SubTask 2.2: 创建 CoreAgentCard 组件，展示三个核心对象的名称、图标、优缺点
- [ ] SubTask 2.3: 实现核心对象选择的状态管理和回调

## Task 3: 重构 SplashScreen 组件 - 移除跳过动画选项
- [ ] SubTask 3.1: 移除"下次启动跳过动画"复选框
- [ ] SubTask 3.2: 确保每次启动都播放完整入场动画
- [ ] SubTask 3.3: 保留"重置启动动画"功能（用于关于页）

## Task 4: 更新 App.tsx 启动流程
- [ ] SubTask 4.1: 确保 App.tsx 正确传递 coreAgent 给 SplashScreen
- [ ] SubTask 4.2: 验证启动流程：入场动画 → 核心对象选择 → 模式选择 → 主页

## Task 5: 更新 SkillsPage 组件 - 插件生态差异化展示
- [ ] SubTask 5.1: 创建插件可用性配置，定义各核心对象的可用插件列表
- [ ] SubTask 5.2: 实现插件可用性判断逻辑（根据 coreAgent 判断）
- [ ] SubTask 5.3: 在插件卡片上显示"仅 OpenClaw"等标签
- [ ] SubTask 5.4: 不可用插件显示为灰色或禁用状态

## Task 6: 更新 AboutPage 组件 - 新增核心对象切换功能
- [ ] SubTask 6.1: 新增"切换核心对象"按钮
- [ ] SubTask 6.2: 实现点击逻辑：重置 onboardingComplete 并跳转回核心对象选择页
- [ ] SubTask 6.3: 展示当前已选择的核心对象名称

## Task 7: 实现自动更新模块 (auto-updater.ts)
- [ ] SubTask 7.1: 创建 VersionChecker - 版本检测（启动时 + 定期）
- [ ] SubTask 7.2: 创建 UpdateDownloader - 差量下载（后台静默）
- [ ] SubTask 7.3: 创建 UpdateInstaller - 更新安装（用户透明）
- [ ] SubTask 7.4: 创建 RollbackManager - 回滚管理（更新失败时）
- [ ] SubTask 7.5: 创建 UpdateLogger - 更新日志记录
- [ ] SubTask 7.6: 与 AboutPage 集成，展示更新状态和历史

## Task 8: 实现实时推送模块 (push-service.ts)
- [ ] SubTask 8.1: 创建 WebSocketClient - 长连接管理（自动重连）
- [ ] SubTask 8.2: 创建 MessageQueue - 本地消息队列（离线支持）
- [ ] SubTask 8.3: 创建 NotificationManager - 通知管理（系统+应用内）
- [ ] SubTask 8.4: 创建 PushPreferences - 推送偏好设置
- [ ] SubTask 8.5: 创建 SecurityValidator - 消息签名验证
- [ ] SubTask 8.6: 与 StatusPage/DiagnosisPage 集成

## Task 9: 实现环境适配器 (environment-adapter.ts)
- [ ] SubTask 9.1: 创建 EnvironmentDetector - 环境检测（Provider、网络、资源）
- [ ] SubTask 9.2: 创建 ProviderSelector - Provider 智能选择和切换
- [ ] SubTask 9.3: 创建 HealthMonitor - 健康检查（定期）
- [ ] SubTask 9.4: 创建 PerformanceOptimizer - 性能自适应（资源紧张时降级）
- [ ] SubTask 9.5: 创建 AutoRecovery - 自动恢复机制

## Task 10: 验证和测试
- [ ] SubTask 10.1: 验证启动流程完整性（动画 → 核心对象选择 → 模式选择 → 主页）
- [ ] SubTask 10.2: 验证核心对象选择功能
- [ ] SubTask 10.3: 验证 SkillsPage 插件差异化展示（不同核心对象显示不同插件和标签）
- [ ] SubTask 10.4: 验证关于页核心对象切换功能
- [ ] SubTask 10.5: 验证入场动画每次启动都播放
- [ ] SubTask 10.6: 验证自动更新机制（后台检测、差量下载、回滚）
- [ ] SubTask 10.7: 验证实时推送系统（长连接、离线消息、推送偏好）
- [ ] SubTask 10.8: 验证自适应环境功能（智能切换、性能降级）

## Task Dependencies
- Task 2 依赖 Task 1（需要先有 coreAgent 状态）
- Task 3 可独立进行
- Task 4 依赖 Task 1、2、3
- Task 5 可独立进行（在核心对象确定后）
- Task 6 可独立进行
- Task 7 可独立进行
- Task 8 可独立进行
- Task 9 可独立进行
- Task 10 依赖 Task 1-9