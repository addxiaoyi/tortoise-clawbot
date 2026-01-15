# MCP 配置与市场集成规范

## Why

用户需要通过可视化界面自助配置 MCP (Model Context Protocol) 服务器，对接热门 MCP 市场实现一键安装，提升工具扩展能力。

## What Changes

### 1. 新增功能模块

#### MCP 配置页面
- 可视化 MCP 服务器管理
- 手动添加/编辑/删除 MCP 配置
- 连接状态实时监测

#### MCP 市场对接
- 接入 mcpmark.com 热门市场
- 展示热门 MCP 服务器列表
- 一键安装/卸载市场应用

### 2. 数据结构

```typescript
interface MCPServer {
  id: string;
  name: string;
  command: string;
  args: string[];
  env: Record<string, string>;
  enabled: boolean;
  status: 'connected' | 'disconnected' | 'error';
}

interface MCPMarketItem {
  id: string;
  name: string;
  description: string;
  author: string;
  stars: number;
  installCommand: string;
  categories: string[];
  verified: boolean;
}
```

## Impact

### 影响的规范
- SkillsPage - 新增第四标签页 MCP
- ConfigPage - MCP 配置入口

### 影响的代码
- src/pages/MCPConfigPage.tsx (新增)
- src/components/MCPMarketplace.tsx (新增)
- src/store/useMCPStore.ts (新增)

## ADDED Requirements

### Requirement: MCP Server Management
系统 SHALL 提供可视化 MCP 服务器管理界面

#### Scenario: Add MCP server
- **WHEN** 用户点击添加 MCP 服务器
- **THEN** 展示配置表单，保存后自动连接

### Requirement: MCP Market
系统 SHALL 对接 MCP 市场，展示热门服务器

#### Scenario: Install from market
- **WHEN** 用户浏览 MCP 市场并点击安装
- **THEN** 自动配置并启动 MCP 服务器

## MODIFIED Requirements

### Requirement: SkillsPage Extension
SkillsPage 新增 MCP 标签页

## REMOVED Requirements

无

## 技术实现

### 市场 API
- mcpmark.com 提供公开 API
- 备用: mcp.so, smithery.ai

### 存储
- localStorage 存储 MCP 配置
- Tauri fs API 读写配置文件