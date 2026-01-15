# 多核心 Agent 支持规范

## Why

当前 tortoise 项目仅支持 OpenClaw 作为核心操控对象，但用户需要同时支持 OpenClaw、myclaw (stellarlinkco)、nanobot (HKUDS) 三个核心系统。需要在入场动画后让用户选择核心，并支持在关于我们界面重新选择。每个核心有不同的插件生态和功能，需要明确展示。**所有界面必须完整支持中文。**

## What Changes

### 1. 多核心架构

| 核心 | 来源仓库 | Stars | 特点 | 专属插件 |
|------|---------|-------|------|---------|
| **OpenClaw** | openclaw-skills | 40k+ | 完整插件生态、Gateway、GitHub、Slack、Notion、Memory | gateway, github, slack, discord, notion-page, notion-database, memory-setup, agent-memory, self-improvement, confluence, sharepoint, langchain, llamaindex, autogen, crewai, dspy |
| **myclaw** | stellarlinkco/myclaw | - | 多渠道集成、Telegram、飞书、企业微信、WhatsApp、Gateway模式 | telegram, feishu-doc, feishu-drive, feishu-perm, feishu-wiki, feishu-meeting, feishu-calendar, wecom-doc, wecom-meeting, dingtalk-doc, dingtalk-drive, lark-doc, lark-sheet, lark-base |
| **nanobot** | HKUDS/nanobot | 7.5k | AI同事工作流、经济基准测试、多模型支持 | multi-agent-orchestration, a2a-protocol, heartbeat-schedule, context-tiered, sandbox-security, mongodb-change-stream, knowledge-沉淀, agent-team-orchestration, agent-spawner, agent-swarm |

### 2. 入场动画流程修改

```
应用启动
    ↓
显示入场动画 (默认每次播放)
    ↓
显示核心选择界面 (新)
    ↓
用户选择核心 + 显示优缺点
    ↓
进入主界面
```

### 3. 核心选择界面设计

**界面布局:**
- 三个核心并排展示卡片
- 每个卡片显示：图标、名称、描述、优点列表、缺点列表
- 用户点击选择后高亮确认
- "继续" 按钮进入主界面

**所有文本必须为中文:**
- 标题：选择你的核心操控对象
- 副标题：选择一个最适合你的 AI Agent 系统
- 继续按钮：继续
- 选择提示：请选择一个核心开始

### 4. 插件兼容性显示

**技能/插件状态 (中文):**
- `✅ 全平台` - 所有核心都支持
- `🔵 仅 OpenClaw` - 仅 OpenClaw 支持
- `🟢 仅 myclaw` - 仅 myclaw 支持
- `🟣 仅 nanobot` - 仅 nanobot 支持

**示例:**
- `gateway` → 🔵 仅 OpenClaw
- `feishu-doc` → 🟢 仅 myclaw
- `multi-agent-orchestration` → 🟣 仅 nanobot
- `tavily-search` → ✅ 全平台

## Impact

### 受影响的规范
- SplashScreen - 入场动画流程
- AboutPage/关于我们 - 新增核心管理
- SkillsPage - 动态加载核心相关技能，插件兼容性标签

### 受影响的代码
- src/components/SplashScreen.tsx - 核心选择流程
- src/pages/AboutPage.tsx - 关于我们界面
- src/store/useOnboardingStore.ts - 核心选择状态
- src/lib/coreAgents.ts - 核心定义
- src/lib/pluginCompatibility.ts - 插件兼容性判断

## ADDED Requirements

### Requirement: Core Agent Selection
系统 SHALL 在入场动画后显示核心选择界面，用户可选择 OpenClaw、myclaw 或 nanobot

#### Scenario: First launch
- **WHEN** 用户首次启动应用
- **THEN** 显示入场动画，然后显示核心选择界面

#### Scenario: Core selection
- **WHEN** 用户在核心选择界面
- **THEN** 展示三个核心的优缺点，用户点击选择后高亮

### Requirement: Core Re-selection
系统 SHALL 在关于我们界面提供重新选择核心功能

#### Scenario: Re-select core
- **WHEN** 用户点击"重新选择核心"
- **THEN** 跳转到核心选择界面

### Requirement: Core-specific Skills Display
系统 SHALL 根据用户选择的核心里动态加载相关技能，并显示插件兼容性标签

#### Scenario: OpenClaw skills
- **WHEN** 用户选择 OpenClaw
- **THEN** 加载 OpenClaw 专属技能集，插件显示"🔵 仅 OpenClaw"标签

#### Scenario: myclaw skills
- **WHEN** 用户选择 myclaw
- **THEN** 加载 myclaw 专属技能集，插件显示"🟢 仅 myclaw"标签

#### Scenario: nanobot skills
- **WHEN** 用户选择 nanobot
- **THEN** 加载 nanobot 专属技能集，插件显示"🟣 仅 nanobot"标签

#### Scenario: Cross-platform skills
- **WHEN** 技能被多个核心支持
- **THEN** 显示"✅ 全平台"标签

## MODIFIED Requirements

### Requirement: SplashScreen Animation
入场动画默认每次进入应用程序都播放

#### Scenario: Animation playback
- **WHEN** 应用启动
- **THEN** 播放入场动画（默认每次都播放）

## 数据结构设计

```typescript
// 核心 Agent 定义
interface CoreAgentDef {
  id: 'openclaw' | 'myclaw' | 'nanobot';
  name: string;
  icon: string;
  color: string;
  description: string;  // 中文描述
  advantages: string[];      // 中文优点列表
  disadvantages: string[];     // 中文缺点列表
  capabilities: string[];     // 支持的能力ID列表
  defaultSkills: string[];    // 默认技能ID列表
  modelProviders: string[];    // 支持的模型提供商
}

// 插件兼容性
interface PluginCompatibility {
  skillId: string;
  supported: ('openclaw' | 'myclaw' | 'nanobot')[];
  displayLabel: string;  // "🔵 仅 OpenClaw" / "🟢 仅 myclaw" / "🟣 仅 nanobot" / "✅ 全平台"
}

// 用户选择状态
interface CoreSelectionState {
  selected: 'openclaw' | 'myclaw' | 'nanobot' | null;
  lastSelected: 'openclaw' | 'myclaw' | 'nanobot' | null;
  hasCompletedSelection: boolean;
}
```

## 核心优缺点对比 (全中文)

### OpenClaw
**优点:**
- 完整插件生态：Gateway、GitHub、Slack、Notion、Memory 等
- 成熟稳定：经过大量用户验证
- 社区活跃：文档完善、插件丰富
- 企业级功能：Confluence、SharePoint 集成

**缺点:**
- 学习曲线较陡
- 资源占用相对较高

**专属技能:**
gateway, github, slack, discord, notion-page, notion-database, notion-wiki, memory-setup, agent-memory, self-improvement, confluence, confluence-cloud, sharepoint, langchain, llamaindex, autogen, crewai, dspy, traceloop

### myclaw
**优点:**
- 多渠道集成：Telegram、飞书、企业微信、WhatsApp
- Gateway 模式：完整渠道编排 + cron + heartbeat
- 部署简单
- 国内生态完善：飞书、钉钉、企业微信

**缺点:**
- 插件生态相对较小
- 社区活跃度较低

**专属技能:**
telegram, feishu-doc, feishu-drive, feishu-perm, feishu-wiki, feishu-meeting, feishu-calendar, feishu-approval, wecom-doc, wecom-meeting, dingtalk-doc, dingtalk-drive, dingtalk-meeting, lark-doc, lark-sheet, lark-base

### nanobot
**优点:**
- AI同事工作流：多Agent协作
- 经济基准测试：真实专业任务评估
- 多模型支持：GPT-4o、Claude、Gemini、Qwen、GLM-4
- 沙箱安全：隔离执行环境

**缺点:**
- 项目较新
- 文档完善度待提升

**专属技能:**
multi-agent-orchestration, a2a-protocol, heartbeat-schedule, context-tiered, sandbox-security, mongodb-change-stream, knowledge-沉淀, agent-team-orchestration, agent-spawner, agent-swarm, agent-orchestrator, agent-registry, agent-dispatch, context-engine

## 全平台技能 (跨核心)

以下技能所有核心都支持：
- tavily-search, duckduckgo-search, google-search
- screenshot, ocr-text, video-subtitle
- videocut
- lmstudio-local, ollama-local (本地模型)
- deepseek-api, qwen-api, moonshot-api, volcengine-api, glm-api (国内API)
- langfuse, langsmith (观测)

## 关于我们界面 (全中文)

**显示内容:**
- 应用名称： tortoise
- 当前核心：[核心图标] [核心名称]
- 核心描述：[核心描述]
- 重新选择按钮：重新选择核心

## 核心选择界面文本 (全中文)

**标题:** 选择你的核心操控对象
**副标题:** 选择一个最适合你的 AI Agent 系统

**OpenClaw 卡片:**
- 名称: OpenClaw
- 描述: 成熟的开源 AI Agent 系统
- 优点标题: 优点
- 缺点标题: 缺点

**myclaw 卡片:**
- 名称: myclaw
- 描述: 多渠道集成的 AI 助手
- 优点标题: 优点
- 缺点标题: 缺点

**nanobot 卡片:**
- 名称: nanobot
- 描述: AI同事工作流系统
- 优点标题: 优点
- 缺点标题: 缺点

**按钮:**
- 继续按钮: 继续
- 选择提示: 请选择一个核心开始
