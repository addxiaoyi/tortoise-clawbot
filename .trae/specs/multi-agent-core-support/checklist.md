# Checklist

## useOnboardingStore 核心对象状态
- [ ] `coreAgent` 类型定义正确（"openclaw" | "myclaw" | "nanobot" | null）
- [ ] `setCoreAgent` 方法已添加
- [ ] `completeOnboarding` 方法已更新支持 coreAgent 参数

## SplashScreen 核心对象选择页
- [ ] 核心对象选择页正确展示三个选项
- [ ] OpenClaw 选项正确显示名称、图标、核心优势、潜在不足
- [ ] myclaw 选项正确显示名称、图标、核心优势（多渠道、Gateway、轻量高效）、潜在不足
- [ ] nanobot 选项正确显示名称、图标、核心优势（超轻量、MCP、多Provider）、潜在不足
- [ ] 点击选项后正确调用 onSelectCoreAgent 回调
- [ ] 选择后正确过渡到模式选择页

## SplashScreen 入场动画行为
- [ ] 每次启动都播放入场动画（无跳过选项）
- [ ] 3200ms 动画时长正确
- [ ] 动画各阶段（序章、环现、标升、文浮、余韵）正常
- [ ] 跳过按钮功能正常

## App.tsx 启动流程
- [ ] SplashScreen 正确传递 coreAgent 和 mode 参数
- [ ] 启动流程：入场动画 → 核心对象选择 → 模式选择 → 主页

## SkillsPage 插件生态差异化
- [ ] 各核心对象的可用插件列表已定义
- [ ] OpenClaw 展示全部插件，无限制标签
- [ ] myclaw 展示可用插件，标注不支持的插件为"仅 OpenClaw 可用"
- [ ] nanobot 展示可用插件，标注不支持的插件为"仅 OpenClaw 可用"
- [ ] 不可用插件显示为灰色或禁用状态
- [ ] 插件卡片上正确显示"仅 OpenClaw"等标签

## AboutPage 核心对象切换
- [ ] 显示当前已选择的核心对象名称
- [ ] "切换核心对象"按钮存在且可点击
- [ ] 点击后正确重置 onboardingComplete 并跳转
- [ ] "重置启动动画"按钮保留且功能正常
- [ ] "检查更新"按钮功能正常
- [ ] "更新历史"可查看历史更新记录

## 自动更新模块 (auto-updater.ts)
- [ ] VersionChecker 实现正确（启动检测 + 定期检测）
- [ ] UpdateDownloader 实现差量下载，后台静默执行
- [ ] UpdateInstaller 用户透明安装
- [ ] RollbackManager 更新失败自动回滚
- [ ] UpdateLogger 正确记录更新日志
- [ ] 与 AboutPage 正确集成

## 实时推送模块 (push-service.ts)
- [ ] WebSocketClient 长连接维护正常
- [ ] 自动重连使用指数退避策略
- [ ] MessageQueue 离线消息不丢失
- [ ] NotificationManager 系统+应用内通知正常
- [ ] PushPreferences 用户可配置推送类型和免打扰时段
- [ ] SecurityValidator 消息签名验证正确

## 环境适配器 (environment-adapter.ts)
- [ ] EnvironmentDetector 环境检测正常（Provider、网络、资源）
- [ ] ProviderSelector 智能选择和切换正常
- [ ] HealthMonitor 健康检查定期执行
- [ ] PerformanceOptimizer 资源紧张时正确降级
- [ ] AutoRecovery 自动恢复机制正常