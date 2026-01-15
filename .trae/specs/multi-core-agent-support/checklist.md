# Checklist - 多核心 Agent 支持

## 阶段 1: 核心定义数据 (全中文)

- [ ] Task 1.1: src/lib/coreAgents.ts 创建完成
  - CoreAgentDef 接口定义
  - OpenClaw 数据完整
  - myclaw 数据完整
  - nanobot 数据完整
  - advantages, disadvantages, capabilities, defaultSkills, modelProviders
  - **所有文本必须为中文**

- [ ] Task 1.2: src/lib/pluginCompatibility.ts 创建完成
  - PLUGIN_COMPATIBILITY 常量定义
  - 每个技能对应支持的核心列表
  - getCompatibilityLabel(skillId, selectedCore) 函数
  - 返回正确的标签格式

- [ ] Task 1.3: src/store/useCoreAgentStore.ts 创建完成
  - selected, lastSelected, hasCompletedSelection 状态
  - localStorage 持久化
  - setSelectedCore 方法
  - resetSelection 方法

## 阶段 2: 核心选择组件 (全中文界面)

- [ ] Task 2.1: src/components/CoreAgentSelector.tsx 创建完成
  - 标题文本：选择你的核心操控对象
  - 副标题文本：选择一个最适合你的 AI Agent 系统
  - 三个核心卡片并排渲染
  - 每个卡片显示：图标、名称、颜色、描述
  - 优点列表渲染
  - 缺点列表渲染
  - 选中状态高亮（border + 背景色）
  - 继续按钮文字：继续
  - 选择提示文字：请选择一个核心开始

- [ ] Task 2.2: src/components/SplashScreen.tsx 更新完成
  - 入场动画后显示 CoreAgentSelector
  - 选择完成后进入主界面
  - useCoreAgentStore 集成

## 阶段 3: 关于我们界面 (全中文)

- [ ] Task 3.1: src/pages/AboutPage.tsx 更新完成
  - 显示当前已选核心（图标 + 名称 + 描述）
  - 当前核心标签文字：当前核心
  - 重新选择按钮文字：重新选择核心
  - 按钮点击跳转核心选择

## 阶段 4: 核心特定技能 (全中文显示)

- [ ] Task 4.1: SkillsPage CURATED_SECTIONS 更新完成
  - OpenClaw 专属技能分类（19+）
  - myclaw 专属技能分类（16+）
  - nanobot 专属技能分类（14+）
  - 全平台技能分类

- [ ] Task 4.2: SkillsPage CORE_AGENT_CAPABILITIES 更新完成
  - OpenClaw 能力映射完整
  - myclaw 能力映射完整
  - nanobot 能力映射完整

- [ ] Task 4.3: 技能卡片渲染更新完成
  - getCompatibilityLabel(skillId, selectedCore) 集成
  - 插件兼容性标签显示
  - 标签颜色区分核心（🔵 🟢 🟣 ✅）

## 阶段 5: 构建验证

- [ ] Task 5.1: TypeScript 检查通过
  - npx tsc --noEmit 无错误

- [ ] Task 5.2: Vite 构建通过
  - npm run build 成功

- [ ] Task 5.3: Tauri 构建通过
  - npm run tauri build 成功

## 功能验证点

### 核心选择流程
- [ ] 应用启动显示入场动画
- [ ] 动画后显示核心选择界面
- [ ] 核心选择界面标题显示：选择你的核心操控对象
- [ ] 核心选择界面副标题显示：选择一个最适合你的 AI Agent 系统
- [ ] 三个核心卡片正确展示
- [ ] 每个卡片显示优缺点
- [ ] 优点标签显示：优点
- [ ] 缺点标签显示：缺点
- [ ] 点击选择后高亮
- [ ] 继续按钮显示：继续
- [ ] 未选择时提示：请选择一个核心开始

### 关于我们界面
- [ ] 显示当前核心信息
- [ ] 当前核心标签显示：当前核心
- [ ] 重新选择按钮显示：重新选择核心
- [ ] 点击跳转核心选择

### 插件兼容性显示
- [ ] OpenClaw 专属技能显示 🔵 仅 OpenClaw
- [ ] myclaw 专属技能显示 🟢 仅 myclaw
- [ ] nanobot 专属技能显示 🟣 仅 nanobot
- [ ] 全平台技能显示 ✅ 全平台

## OpenClaw 专属技能 (19+)

gateway, github, slack, discord, notion-page, notion-database, notion-wiki, memory-setup, agent-memory, self-improvement, confluence, confluence-cloud, sharepoint, langchain, llamaindex, autogen, crewai, dspy, traceloop

## myclaw 专属技能 (16+)

telegram, feishu-doc, feishu-drive, feishu-perm, feishu-wiki, feishu-meeting, feishu-calendar, feishu-approval, wecom-doc, wecom-meeting, dingtalk-doc, dingtalk-drive, dingtalk-meeting, lark-doc, lark-sheet, lark-base

## nanobot 专属技能 (14+)

multi-agent-orchestration, a2a-protocol, heartbeat-schedule, context-tiered, sandbox-security, mongodb-change-stream, knowledge-沉淀, agent-team-orchestration, agent-spawner, agent-swarm, agent-orchestrator, agent-registry, agent-dispatch, context-engine

## 全平台技能

tavily-search, duckduckgo-search, google-search, screenshot, ocr-text, video-subtitle, videocut, lmstudio-local, ollama-local, deepseek-api, qwen-api, moonshot-api, volcengine-api, glm-api, langfuse, langsmith, coding-agent

## 中文界面文本清单

### 核心选择界面
- [ ] 标题：选择你的核心操控对象
- [ ] 副标题：选择一个最适合你的 AI Agent 系统
- [ ] 继续按钮：继续
- [ ] 选择提示：请选择一个核心开始
- [ ] 优点标签：优点
- [ ] 缺点标签：缺点

### 关于我们界面
- [ ] 应用名称：tortoise
- [ ] 当前核心标签：当前核心
- [ ] 重新选择按钮：重新选择核心

### 插件兼容性标签
- [ ] 🔵 仅 OpenClaw
- [ ] 🟢 仅 myclaw
- [ ] 🟣 仅 nanobot
- [ ] ✅ 全平台
