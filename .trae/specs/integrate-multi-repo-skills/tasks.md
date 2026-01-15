# Tasks - 多仓库技能系统智能集成

## 阶段 1: 环境恢复 (Environment Recovery)

- [ ] Task 1.1: 从备份恢复项目源码到 tortoise 目录
  - 恢复 src/, src-tauri/, public/
  - 恢复配置文件 package.json, tsconfig.json 等
  - 恢复 node_modules 依赖

- [ ] Task 1.2: 验证项目可构建
  - 运行 npm run tauri build
  - 确保 Windows x64 exe 可正常生成

## 阶段 2: 核心组件实现 (Core Components)

### P0 组件

- [ ] Task 2.1: 创建 AssistantMarketplace 组件
  - 基于 cherry-studio 架构
  - 300+ 预设助手数据
  - 分类、搜索、安装功能

- [ ] Task 2.2: 创建 ContextEngine 组件
  - 基于 OpenViking L0/L1/L2 分层
  - 文件系统范式上下文
  - 自进化记忆系统

- [ ] Task 2.3: 创建 SkillsLibrary 组件
  - 精选 awesome-openclaw-skills
  - 分类索引
  - 技能详情展示

### P1 组件

- [ ] Task 2.4: 创建 MultiAgentDashboard 组件
  - clawe 风格看板
  - A2A 协议支持
  - 任务分配系统

- [ ] Task 2.5: 创建 SpecializedSkills 组件
  - obsidian-skills 导入
  - ClawWork 工作流集成

### P2 组件

- [ ] Task 2.6: 创建 EnhancedConsole 组件
  - 实时看板 (来自 nerve)
  - 语音对话入口

- [ ] Task 2.7: 创建 SlackIntegration 组件
  - opencrew Slack 集成

## 阶段 3: SkillsPage 集成 (SkillsPage Integration)

- [ ] Task 3.1: 重构 SkillsPage 为多标签页架构
  - 标签 1: 技能市场 (原功能)
  - 标签 2: 助理市场 (AssistantMarketplace)
  - 标签 3: 上下文引擎 (ContextEngine)
  - 标签 4: 多智能体 (MultiAgentDashboard)

- [ ] Task 3.2: 集成 CORE_AGENT_CAPABILITIES 能力映射
  - 新增 8+ 能力映射
  - gateway, github, slack, discord 等

- [ ] Task 3.3: 集成 CURATED_SECTIONS 分类
  - 多智能体协作
  - 沙箱安全
  - 上下文管理

## 阶段 4: 技能数据填充 (Skill Data Population)

- [ ] Task 4.1: 导入 cherry-studio 精选助手 (Top 12)
- [ ] Task 4.2: 导入 awesome-openclaw-skills 分类技能
- [ ] Task 4.3: 导入 OpenViking 上下文技能
- [ ] Task 4.4: 导入 obsidian-skills 专业技能
- [ ] Task 4.5: 导入 ClawWork 工作流技能

## 阶段 5: 构建验证 (Build Verification)

- [ ] Task 5.1: TypeScript 类型检查 (npx tsc --noEmit)
- [ ] Task 5.2: Vite 构建验证 (npm run build)
- [ ] Task 5.3: Tauri 构建验证 (npm run tauri build)
- [ ] Task 5.4: 生成 Windows x64 exe

## 阶段 6: GitHub 发布 (GitHub Release)

- [ ] Task 6.1: 推送源码到 GitHub
- [ ] Task 6.2: 创建 Release 上传构建文件
- [ ] Task 6.3: 更新 README

## Task Dependencies

```
Task 1.1 → Task 1.2
Task 1.2 → Task 2.1, Task 2.2, Task 2.3, Task 2.4, Task 2.5, Task 2.6, Task 2.7
Task 2.1, Task 2.2, Task 2.3, Task 2.4 → Task 3.1
Task 3.1 → Task 4.1, Task 4.2, Task 4.3, Task 4.4, Task 4.5
Task 4.* → Task 5.1
Task 5.1 → Task 5.2 → Task 5.3 → Task 5.4
Task 5.4 → Task 6.1 → Task 6.2 → Task 6.3
```

## 集成来源映射

| 组件 | 来源仓库 | Stars |
|------|---------|-------|
| AssistantMarketplace | CherryHQ/cherry-studio | 42k |
| ContextEngine | volcengine/OpenViking | 17.3k |
| SkillsLibrary | VoltAgent/awesome-openclaw-skills | 40k |
| MultiAgentDashboard | getclawe/clawe + AlexAnys/opencrew | 684+338 |
| SpecializedSkills | kepano/obsidian-skills + HKUDS/ClawWork | 15k+7.5k |
| EnhancedConsole | daggerhashimoto/openclaw-nerve | 207 |
| SlackIntegration | AlexAnys/opencrew | 338 |
