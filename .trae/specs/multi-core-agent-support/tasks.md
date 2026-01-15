# Tasks - 多核心 Agent 支持

## 阶段 1: 核心定义数据 (全中文)

- [ ] Task 1.1: 创建 src/lib/coreAgents.ts
  - 定义 CoreAgentDef 接口 (中文注释)
  - OpenClaw 数据完整（id, name, icon, color, description, advantages, disadvantages, capabilities, defaultSkills, modelProviders）
  - myclaw 数据完整
  - nanobot 数据完整
  - **所有文本必须为中文**

- [ ] Task 1.2: 创建 src/lib/pluginCompatibility.ts
  - 定义 PLUGIN_COMPATIBILITY 常量
  - 每个技能对应支持的核心列表
  - getCompatibilityLabel(skillId, selectedCore) 函数
  - 返回 "🔵 仅 OpenClaw" / "🟢 仅 myclaw" / "🟣 仅 nanobot" / "✅ 全平台"

- [ ] Task 1.3: 创建 src/store/useCoreAgentStore.ts
  - 核心选择状态管理
  - selected, lastSelected, hasCompletedSelection
  - localStorage 持久化
  - setSelectedCore, resetSelection 方法

## 阶段 2: 核心选择组件 (全中文界面)

- [ ] Task 2.1: 创建 src/components/CoreAgentSelector.tsx
  - 标题：选择你的核心操控对象
  - 副标题：选择一个最适合你的 AI Agent 系统
  - 三个核心卡片并排展示
  - 每个卡片显示：图标、名称、颜色、描述、优点列表、缺点列表
  - 优点标题：优点
  - 缺点标题：缺点
  - 选中状态高亮（border + 背景色）
  - 继续按钮文字：继续
  - 选择提示文字：请选择一个核心开始

- [ ] Task 2.2: 更新 src/components/SplashScreen.tsx
  - 入场动画后显示 CoreAgentSelector
  - 选择完成后进入主界面
  - 使用 useCoreAgentStore 判断是否显示选择

## 阶段 3: 关于我们界面修改 (全中文)

- [ ] Task 3.1: 更新 src/pages/AboutPage.tsx
  - 显示当前已选核心（图标 + 名称 + 描述）
  - 重新选择按钮文字：重新选择核心
  - 点击跳转核心选择界面

## 阶段 4: 核心特定技能 (全中文显示)

- [ ] Task 4.1: 更新 SkillsPage 的 CURATED_SECTIONS
  - OpenClaw 专属技能：gateway, github, slack, notion-page, memory-setup 等
  - myclaw 专属技能：telegram, feishu-doc, feishu-drive 等
  - nanobot 专属技能：multi-agent-orchestration, a2a-protocol 等
  - 全平台技能：tavily-search, screenshot 等

- [ ] Task 4.2: 更新 SkillsPage 的 CORE_AGENT_CAPABILITIES
  - OpenClaw 能力映射：gateway, github, slack, notion, memory
  - myclaw 能力映射：telegram, feishu, wecom, dingtalk, lark
  - nanobot 能力映射：multi-agent, context-tiered, sandbox

- [ ] Task 4.3: 更新技能卡片渲染
  - 显示插件兼容性标签
  - getCompatibilityLabel(skillId, selectedCore)
  - 标签颜色区分核心

## 阶段 5: 构建验证 (构建检查)

- [ ] Task 5.1: TypeScript 类型检查 (npx tsc --noEmit)
- [ ] Task 5.2: Vite 构建验证 (npm run build)
- [ ] Task 5.3: Tauri 构建验证 (npm run tauri build)

## Task Dependencies

```
Task 1.1 → Task 1.2 → Task 1.3
Task 1.3 → Task 2.1
Task 2.1 → Task 2.2
Task 2.2 → Task 3.1
Task 3.1 → Task 4.1
Task 4.1 → Task 4.2 → Task 4.3
Task 4.3 → Task 5.1
Task 5.1 → Task 5.2 → Task 5.3
```

## 核心信息详细 (全中文)

### OpenClaw
- **ID**: openclaw
- **图标**: 🐉
- **颜色**: #007AFF
- **名称**: OpenClaw
- **描述**: 成熟的开源 AI Agent 系统
- **优点**: 完整插件生态、Gateway、GitHub、Slack、Notion、Memory
- **缺点**: 学习曲线较陡、资源占用相对较高
- **专属技能数**: 19+
- **支持渠道**: Slack, Discord, Notion, GitHub, Memory, Gateway

### myclaw
- **ID**: myclaw
- **图标**: 🐢
- **颜色**: #34C759
- **名称**: myclaw
- **描述**: 多渠道集成的 AI 助手
- **优点**: 多渠道集成(Telegram/飞书/企业微信/WhatsApp)、Gateway模式、部署简单
- **缺点**: 插件生态相对较小，社区活跃度较低
- **专属技能数**: 16+
- **支持渠道**: Telegram, 飞书, 企业微信, WhatsApp, 钉钉

### nanobot
- **ID**: nanobot
- **图标**: 🤖
- **颜色**: #AF52DE
- **名称**: nanobot
- **描述**: AI同事工作流系统
- **优点**: 多Agent协作、经济基准测试、多模型支持
- **缺点**: 项目较新、文档完善度待提升
- **专属技能数**: 14+
- **支持渠道**: 多Agent内部协作

## 插件兼容性示例 (中文标签)

| 技能ID | OpenClaw | myclaw | nanobot | 显示标签 |
|--------|----------|--------|---------|---------|
| gateway | ✅ | ❌ | ❌ | 🔵 仅 OpenClaw |
| feishu-doc | ❌ | ✅ | ❌ | 🟢 仅 myclaw |
| multi-agent-orchestration | ❌ | ❌ | ✅ | 🟣 仅 nanobot |
| tavily-search | ✅ | ✅ | ✅ | ✅ 全平台 |
| coding-agent | ✅ | ✅ | ✅ | ✅ 全平台 |

## 中文界面文本汇总

### 核心选择界面
- 标题：选择你的核心操控对象
- 副标题：选择一个最适合你的 AI Agent 系统
- 继续按钮：继续
- 选择提示：请选择一个核心开始
- 优点标签：优点
- 缺点标签：缺点

### 关于我们界面
- 应用名称：tortoise
- 当前核心标签：当前核心
- 重新选择按钮：重新选择核心

### 插件兼容性标签
- 🔵 仅 OpenClaw
- 🟢 仅 myclaw
- 🟣 仅 nanobot
- ✅ 全平台
