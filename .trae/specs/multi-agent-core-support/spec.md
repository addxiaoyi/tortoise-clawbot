# 多 Agent 核心操控对象支持规范

## Why
当前 tortoise 应用仅支持 OpenClaw 作为核心操控对象。根据用户需求，需要扩展支持 `myclaw`（stellarlinkco/myclaw）和 `nanobot`（HKUDS/nanobot）两个新兴的轻量级 AI 助手项目。

同时，为了实现智能化管理与维护，需要实现：
- **自动更新机制**：低侵入性、高稳定性、用户透明
- **实时推送功能**：及时性、准确性、安全性
- **自适应生态环境**：根据核心对象和外部环境变化智能调整

## What Changes

### 1. 新增多 Agent 核心对象支持
- OpenClaw（原有）：成熟稳定的 AI 助手框架，生态完善
- myclaw（新增）：基于 agentsdk-go 的个人 AI 助手，支持 CLI Agent 和 Gateway 模式，多渠道集成
- nanobot（新增）：超轻量级 OpenClaw 实现，代码量减少 99%，支持 MCP、多种渠道和 provider

### 2. 核心对象选择流程重构
- 在入场动画完成后、模式选择（老手/新手）之前，增加核心对象选择页面
- 展示三个核心对象的名称、图标、优缺点描述
- 用户选择后，后续的模式选择和引导流程保持不变

### 3. 插件生态差异化展示
- 根据不同核心对象展示不同的可用插件
- 仅 OpenClaw 可用的插件显示"仅 OpenClaw"标签
- 突出各核心对象的独特优势

### 4. 自动更新机制
- **版本检测**：启动时自动检测最新版本，运行时定期检查
- **差量更新**：支持增量更新，减少下载量
- **热更新**：更新过程用户无感知，后台静默下载
- **回滚机制**：更新失败自动回滚到上一稳定版本
- **更新日志**：展示版本差异和更新内容

### 5. 实时推送系统
- **WebSocket 长连接**：维护持久连接，实时接收服务端推送
- **消息队列**：本地消息队列保障离线消息不丢失
- **多渠道推送**：支持系统通知、App 内通知、邮件等多渠道
- **推送过滤**：用户可自定义推送类型和频率
- **安全性**：推送消息加密传输，身份验证

### 6. 自适应生态环境
- **环境感知**：自动检测可用 Provider、网络状态、系统资源
- **智能切换**：在 Provider 或核心对象不可用时自动切换到备选方案
- **性能优化**：根据系统资源自动调整运行参数
- **健康检查**：定期检查各组件健康状态，自动修复或提醒

### 7. 入场动画行为调整
- 入场动画默认每次启动都播放
- 确保每次用户进入应用都能看到一致的入场体验

### 8. 关于我们页面增强
- 新增"切换核心对象"功能，允许用户重新选择
- 保留"重置启动动画"功能
- 新增"检查更新"和"更新历史"功能

## Impact
- Affected specs: onboarding flow, about page, plugins/skills display, auto-update, push system
- Affected code:
  - `src/store/useOnboardingStore.ts` - 新增 coreAgent 状态
  - `src/components/SplashScreen.tsx` - 重构入场动画和选择流程
  - `src/pages/AboutPage.tsx` - 新增核心对象切换、更新检查功能
  - `src/pages/SkillsPage.tsx` - 插件生态差异化展示
  - `src/lib/auto-updater.ts` - 新增自动更新模块
  - `src/lib/push-service.ts` - 新增实时推送服务
  - `src/lib/environment-adapter.ts` - 新增环境适配器

## ADDED Requirements

### Requirement: 核心对象选择
系统 SHALL 在入场动画之后、模式选择之前展示核心对象选择界面。

#### Scenario: 完整启动流程
- **WHEN** 用户启动应用
- **THEN** 依次展示：入场动画(3200ms) → 核心对象选择 → 老手/新手模式选择 → 主页

#### Scenario: 核心对象展示
- **WHEN** 用户进入核心对象选择页
- **THEN** 展示三个选项，每个选项包含：
  - 名称与图标
  - 核心优势（2-3 个 bullet points）
  - 潜在不足（1-2 个 bullet points）

### Requirement: 核心对象插件生态差异化
系统 SHALL 根据用户选择的 coreAgent 展示不同的插件生态。

#### Scenario: OpenClaw 插件生态
- **WHEN** 用户选择 OpenClaw 作为核心对象
- **THEN** 展示完整插件生态：
  - Gateway 插件、GitHub 插件、Slack 插件、Notion 插件、Memory 插件等全部可用
  - 无限制标签

#### Scenario: myclaw 插件生态
- **WHEN** 用户选择 myclaw 作为核心对象
- **THEN** 展示 myclaw 特有插件：
  - Gateway 编排、Telegram/Feishu/WeCom/WhatsApp 多渠道集成
  - MCP 支持（实验中）
  - **标注"仅 OpenClaw 可用"**：部分高级插件如 Notion 集成、复杂 Memory 调度等

#### Scenario: nanobot 插件生态
- **WHEN** 用户选择 nanobot 作为核心对象
- **THEN** 展示 nanobot 特有插件：
  - 超轻量设计、实时行数统计
  - MCP 支持、DingTalk 插件
  - 多 Provider 支持（OpenAI、Anthropic、Azure、DeepSeek、Moonshot、VolcEngine 等）
  - **标注"仅 OpenClaw 可用"**：部分重型插件如沙盒执行、高级 Notion 集成等

#### Scenario: 插件可用性标签
- **WHEN** 插件仅支持特定核心对象时
- **THEN** 显示"仅 OpenClaw"或"仅 myclaw"等标签
- **THEN** 不可用的插件显示为灰色或禁用状态

### Requirement: 核心对象持久化
系统 SHALL 将用户选择的核心对象持久化存储。

#### Scenario: 选择确认
- **WHEN** 用户选择核心对象并进入主页
- **THEN** 将选择保存到本地存储，下次启动直接进入主页（动画仍播放）

### Requirement: 关于页切换核心对象
系统 SHALL 在关于我们页面提供核心对象重新选择入口。

#### Scenario: 用户在关于页点击切换
- **WHEN** 用户点击"切换核心对象"按钮
- **THEN** 返回核心对象选择页面，选择后可重新开始引导

### Requirement: 入场动画全局播放
系统 SHALL 在每次启动时都播放入场动画。

#### Scenario: 动画播放策略
- **WHEN** 用户启动应用
- **THEN** 无论之前是否跳过，每次都播放完整入场动画

### Requirement: 自动更新机制
系统 SHALL 实现低侵入性、高稳定性的自动更新机制。

#### Scenario: 启动时版本检测
- **WHEN** 用户启动应用
- **THEN** 后台静默检测最新版本，不阻塞主流程
- **THEN** 如有新版本，后台下载差量更新包

#### Scenario: 运行时定期检查
- **WHEN** 应用运行中
- **THEN** 每 30 分钟自动检查一次更新（如用户在线）
- **THEN** 空闲时预加载下一版本，减少用户等待

#### Scenario: 更新下载
- **WHEN** 检测到新版本
- **THEN** 在空闲网络带宽时后台下载
- **THEN** 下载进度不显示，除非用户主动点击"检查更新"

#### Scenario: 更新安装
- **WHEN** 下载完成且用户下次启动时
- **THEN** 提示用户"发现新版本，是否更新？"
- **THEN** 用户可选择"立即更新"或"稍后更新"
- **THEN** 更新过程显示进度条，但可后台进行

#### Scenario: 更新失败回滚
- **WHEN** 更新过程中应用崩溃或中断
- **THEN** 自动回滚到上一稳定版本
- **THEN** 记录错误日志供开发者分析

#### Scenario: 更新历史
- **WHEN** 用户访问关于页
- **THEN** 可查看历史更新记录和当前版本

### Requirement: 实时推送系统
系统 SHALL 实现及时、准确、安全的实时推送功能。

#### Scenario: 长连接维护
- **WHEN** 用户已登录且应用在前台
- **THEN** 维护 WebSocket 长连接接收实时推送
- **THEN** 连接断开时自动重连，使用指数退避策略

#### Scenario: 消息接收
- **WHEN** 收到推送消息
- **THEN** 根据消息类型展示系统通知或应用内通知
- **THEN** 消息加密存储到本地消息队列

#### Scenario: 离线消息处理
- **WHEN** 应用离线时收到推送
- **THEN** 消息暂存服务端，用户上线后同步
- **THEN** 展示消息到达时间戳

#### Scenario: 推送偏好设置
- **WHEN** 用户进入设置页
- **THEN** 可配置接收哪些类型的推送（版本更新、插件推荐、系统公告等）
- **THEN** 可设置免打扰时段

#### Scenario: 推送安全性
- **WHEN** 接收推送消息
- **THEN** 验证消息签名确保来源可靠
- **THEN** 敏感操作需二次确认

### Requirement: 自适应生态环境
系统 SHALL 实现智能化管理与维护，自适应生态环境变化。

#### Scenario: 环境感知
- **WHEN** 应用启动或网络状态变化
- **THEN** 自动检测可用 Provider、网络延迟、系统资源
- **THEN** 评估各核心对象的可用性和性能

#### Scenario: 智能切换
- **WHEN** 当前 Provider 或核心对象不可用
- **THEN** 自动切换到最优备选方案
- **THEN** 切换过程对用户透明，无需手动干预

#### Scenario: 性能自适应
- **WHEN** 系统资源紧张（低内存、低 CPU）
- **THEN** 自动降低动画复杂度、延迟非紧急任务
- **THEN** 保障核心功能流畅运行

#### Scenario: 健康检查
- **WHEN** 应用运行中
- **THEN** 定期检查 Gateway、Provider、插件等组件状态
- **THEN** 发现问题时自动修复或提醒用户

## MODIFIED Requirements

### Requirement: 入场动画跳过机制调整
原"下次启动跳过动画"选项移除，不再影响启动行为。

### Requirement: 启动模式选择页面
模式选择（老手/新手）仍然保留，但位置调整到核心对象选择之后。

## REMOVED Requirements
无

## 核心对象详细信息

### OpenClaw
- **官网**: https://github.com/stellarlinkco/openclaw
- **描述**: 成熟的 AI 助手框架，生态完善，功能丰富
- **核心优势**:
  - 完整插件生态：Gateway、GitHub、Slack、Notion、Memory、Testing 等
  - 成熟稳定：经过大量用户验证
  - 社区活跃：文档完善、插件丰富
- **潜在不足**:
  - 学习曲线较陡
  - 资源占用相对较高

### myclaw
- **官网**: https://github.com/stellarlinkco/myclaw
- **描述**: 基于 agentsdk-go 的个人 AI 助手，支持 CLI Agent 和 Gateway 模式
- **核心优势**:
  - 多渠道集成：Telegram、Feishu、WeCom、WhatsApp
  - Gateway 模式：完整的渠道编排 + cron + heartbeat
  - 多 Provider 支持：Anthropic 和 OpenAI
  - 轻量高效：Go 语言实现，性能优异
- **潜在不足**:
  - 生态较新、文档待完善
  - 部分 OpenClaw 插件暂未支持
- **不支持的插件**: Notion 深度集成、高级 Memory 调度、沙盒执行等

### nanobot
- **官网**: https://github.com/HKUDS/nanobot
- **描述**: 超轻量级 OpenClaw 实现，代码量减少 99%
- **核心优势**:
  - 极简轻量：99% 更少代码，实时行数统计
  - MCP 支持：Model Context Protocol 实验性支持
  - 超多 Provider：OpenAI、Anthropic、Azure、DeepSeek、Moonshot、Kimi、VolcEngine、Ollama、vLLM 等
  - 多渠道支持：Slack、Discord、Telegram、Feishu、DingTalk、QQ、WhatsApp、Matrix 等
  - 创意功能：视频剪辑、Code Review、Planning、Documentation 等技能
- **潜在不足**:
  - 相对年轻、功能在持续迭代
  - 部分重型插件暂未支持
- **不支持的插件**: 沙盒执行、高级 Notion 集成、复杂 Memory 深度调度等

## 技术架构

### 自动更新模块 (auto-updater.ts)
```
- VersionChecker: 版本检测
- UpdateDownloader: 差量下载
- UpdateInstaller: 更新安装
- RollbackManager: 回滚管理
- UpdateLogger: 更新日志
```

### 实时推送模块 (push-service.ts)
```
- WebSocketClient: 长连接管理
- MessageQueue: 本地消息队列
- NotificationManager: 通知管理
- PushPreferences: 推送偏好
- SecurityValidator: 安全验证
```

### 环境适配器 (environment-adapter.ts)
```
- EnvironmentDetector: 环境检测
- ProviderSelector: Provider 选择
- HealthMonitor: 健康监控
- PerformanceOptimizer: 性能优化
- AutoRecovery: 自动恢复
```