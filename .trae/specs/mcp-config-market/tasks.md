# Tasks

## Phase 1: MCP 配置页面基础

- [ ] Task 1: 创建 MCPStore 状态管理
  - 创建 useMCPStore.ts
  - 实现 MCPServer CRUD 操作
  - 持久化到 localStorage

- [ ] Task 2: 创建 MCPServer 类型定义
  - 定义 MCPServer 接口
  - 定义 MCPMarketItem 接口
  - 定义服务器状态类型

## Phase 2: MCP 市场组件

- [ ] Task 3: 创建 MCP Marketplace 组件
  - 接入 mcpmark.com API
  - 展示热门 MCP 服务器列表
  - 实现搜索过滤功能

- [ ] Task 4: 实现市场安装功能
  - 解析安装命令
  - 自动填充配置
  - 一键安装流程

## Phase 3: MCP 配置组件

- [ ] Task 5: 创建 MCP 配置表单组件
  - 服务器名称/命令/参数输入
  - 环境变量配置
  - 表单验证

- [ ] Task 6: 创建服务器状态卡片
  - 实时状态展示
  - 连接/断开按钮
  - 日志查看入口

## Phase 4: 集成与页面

- [ ] Task 7: 创建 MCPConfigPage 页面
  - 整合市场和配置组件
  - 添加引导文案
  - 响应式布局

- [ ] Task 8: 集成到 SkillsPage
  - 新增 MCP 标签页
  - Tab 导航实现
  - 状态保持

## Phase 5: 验证与优化

- [ ] Task 9: 功能验证
  - 构建测试
  - 浏览器预览验证
  - 手动测试安装流程

## Task Dependencies
- Task 3 依赖 Task 1, 2
- Task 4 依赖 Task 3
- Task 5 依赖 Task 1
- Task 6 依赖 Task 1
- Task 7 依赖 Task 4, 5, 6
- Task 8 依赖 Task 7
- Task 9 依赖 Task 8