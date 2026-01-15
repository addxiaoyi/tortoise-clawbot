# 多仓库技能系统智能集成规范

## Why

tortoise-clawbot 项目需要从 12 个相关开源仓库中汲取最佳特性，实现功能最大化集成。当前规范已完成设计但代码未实现，且源码在 GitHub 推送时被清理，需重建。

## What Changes

### 1. 仓库价值评估（按 Stars 排序）

| 仓库 | Stars | 核心价值 | 集成优先级 |
|------|-------|---------|-----------|
| CherryHQ/cherry-studio | 42k | 300+ AI 助手、助理市场 | P0 |
| VoltAgent/awesome-openclaw-skills | 40k | 5400+ 技能精选列表 | P0 |
| volcengine/OpenViking | 17.3k | 上下文数据库、记忆系统 | P0 |
| NVIDIA/NemoClaw | 14.6k | 安全托管推理 | P1 |
| kepano/obsidian-skills | 15k | Obsidian 技能集 | P1 |
| HKUDS/ClawWork | 7.5k | AI 同事工作流 | P1 |
| getclawe/clawe | 684 | 多智能体协调系统 | P1 |
| AlexAnys/opencrew | 338 | Slack 团队协同 | P2 |
| daggerhashimoto/openclaw-nerve | 207 | Web 控制台、语音、看板 | P2 |
| jacksonjp0311-gif/Clawbot-skills | 11 | 智能体群、稳定性内核 | P2 |
| roman-rysenadvanced/clawbot_glm | 4 | GLM 模型集成 | P3 |
| supreeth-ravi/mongoclaw | 7 | MongoDB 集合操作 | P3 |

### 2. 功能去重矩阵

| 功能类别 | 合并方案 | 保留实现 |
|---------|---------|---------|
| 多智能体协调 | clawe + opencrew → 统一 | ClaweMCP 协调器 |
| 上下文管理 | OpenViking 为主 | L0/L1/L2 分层 |
| 技能市场 | cherry + awesome 整合 | AssistantMarketplace |
| Web 控制台 | openclaw-nerve | 实时看板 |
| 记忆系统 | OpenViking | 自进化记忆 |

### 3. 新增功能模块

#### P0 - 核心模块
1. **助理市场 (AssistantMarketplace)**
   - 300+ 预设助手
   - 分类浏览与搜索
   - 一键安装

2. **上下文引擎 (ContextEngine)**
   - L0/L1/L2 分层上下文
   - 文件系统范式
   - 自进化记忆

3. **技能精选库 (SkillsLibrary)**
   - awesome-openclaw-skills 精选
   - 分类索引
   - 技能详情

#### P1 - 重要模块
4. **多智能体协调 (MultiAgentDashboard)**
   - clawe 风格看板
   - A2A 协议
   - 任务分配

5. **安全沙箱 (SandboxSecurity)**
   - NemoClaw 概念
   - 隔离执行
   - NVIDIA 安全推理

6. **专业技能包 (SpecializedSkills)**
   - obsidian-skills (Markdown/Canvas)
   - ClawWork (AI 同事)

#### P2 - 增强模块
7. **增强控制台 (EnhancedConsole)**
   - 实时看板
   - 语音对话

8. **Slack 集成 (SlackIntegration)**
   - opencrew Slack

### 4. 数据结构设计

```typescript
// 助理格式
interface Assistant {
  id: string;
  name: string;
  description: string;
  avatar?: string;
  capabilities: Capability[];
  prompts: string[];
  category: string;
  source: 'builtin' | 'community';
}

// 技能格式
interface Skill {
  id: string;
  name: string;
  description: string;
  tags: string[];
  capabilities: Capability[];
  installCommand?: string;
  sourceRepo: string;
}

// 上下文条目
interface ContextEntry {
  level: 'L0' | 'L1' | 'L2';
  content: string;
  timestamp: number;
  evolution?: string;
}
```

## Impact

### 影响的规范
- SkillsPage - 三标签页架构
- ConfigPage - 助手配置
- AIChatPage - 上下文注入

### 影响的代码
- src/components/AssistantMarketplace.tsx
- src/components/ContextEngine.tsx
- src/components/MultiAgentDashboard.tsx
- src/components/SkillsLibrary.tsx
- src/pages/SkillsPage.tsx

## ADDED Requirements

### Requirement: Assistant Marketplace
系统 SHALL 提供助理市场，用户可浏览安装 300+ 精选助手

#### Scenario: Browse and install
- **WHEN** 用户访问助理市场
- **THEN** 展示分类助手，支持搜索和一键安装

### Requirement: Context Engine
系统 SHALL 提供分层上下文管理，支持 L0/L1/L2 三级

#### Scenario: Context persistence
- **WHEN** AI 对话时
- **THEN** 自动管理上下文，自动注入相关记忆

### Requirement: Multi-Agent Coordination
系统 SHALL 支持多智能体任务协调

#### Scenario: Agent collaboration
- **WHEN** 用户创建多智能体任务
- **THEN** 任务分配执行，状态实时同步

## MODIFIED Requirements

### Requirement: SkillsPage Extension
SkillsPage 扩展为支持助理、技能双轨制市场

## REMOVED Requirements

### Requirement: Redundant Features
合并重复的多智能体实现和多个市场入口
