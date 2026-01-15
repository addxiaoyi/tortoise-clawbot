# Checklist - 多仓库技能系统智能集成

## 阶段 1: 环境恢复

- [ ] Task 1.1: 源码恢复完成
  - src/, src-tauri/, public/ 已恢复
  - 配置文件已恢复
  - 依赖已安装

- [ ] Task 1.2: 项目可构建
  - npm run tauri build 成功
  - Windows x64 exe 可生成

## 阶段 2: P0 核心组件

- [ ] Task 2.1: AssistantMarketplace 组件
  - cherry-studio 架构实现
  - 300+ 助手数据导入
  - 分类、搜索、安装功能

- [ ] Task 2.2: ContextEngine 组件
  - L0/L1/L2 分层实现
  - 文件系统范式
  - 自进化记忆

- [ ] Task 2.3: SkillsLibrary 组件
  - awesome 精选导入
  - 分类索引
  - 技能详情

## 阶段 3: P1 重要组件

- [ ] Task 2.4: MultiAgentDashboard 组件
  - 看板视图
  - A2A 协议
  - 任务分配

- [ ] Task 2.5: SpecializedSkills 组件
  - obsidian-skills
  - ClawWork 集成

## 阶段 4: P2 增强组件

- [ ] Task 2.6: EnhancedConsole 组件
  - 实时看板
  - 语音入口

- [ ] Task 2.7: SlackIntegration 组件
  - Slack 集成

## 阶段 5: SkillsPage 集成

- [ ] Task 3.1: 多标签页架构
  - 技能市场标签
  - 助理市场标签
  - 上下文引擎标签
  - 多智能体标签

- [ ] Task 3.2: 能力映射集成
  - CORE_AGENT_CAPABILITIES 扩展
  - 8+ 新能力映射

- [ ] Task 3.3: CURATED_SECTIONS 分类
  - 多智能体协作分类
  - 沙箱安全分类
  - 上下文管理分类

## 阶段 6: 数据填充

- [ ] Task 4.1: cherry-studio Top 12 助手
- [ ] Task 4.2: awesome 分类技能
- [ ] Task 4.3: OpenViking 上下文技能
- [ ] Task 4.4: obsidian 专业技能
- [ ] Task 4.5: ClawWork 工作流

## 阶段 7: 构建验证

- [ ] Task 5.1: TypeScript 检查通过
- [ ] Task 5.2: Vite 构建通过
- [ ] Task 5.3: Tauri 构建通过
- [ ] Task 5.4: exe 文件生成

## 阶段 8: GitHub 发布

- [ ] Task 6.1: 源码推送
- [ ] Task 6.2: Release 创建
- [ ] Task 6.3: README 更新

## 去重检查

- [ ] 多智能体: clawe + opencrew 合并
- [ ] 上下文: OpenViking 统一
- [ ] 市场: cherry + awesome 整合
- [ ] 无重复实现
