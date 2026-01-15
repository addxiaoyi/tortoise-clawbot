# Tasks - 快捷键系统实现

## 阶段 1: 项目基础结构恢复

- [ ] Task 1.1: 创建 `src/App.tsx` - 应用主入口，包含全局布局和路由
  - 创建 Sidebar 侧边栏布局
  - 创建主内容区域
  - 添加命令面板组件

- [ ] Task 1.2: 创建 `src/components/Sidebar.tsx` - 侧边栏导航组件
  - 实现页面导航功能
  - 实现收起/展开状态
  - 实现主题切换按钮

- [ ] Task 1.3: 创建 `src/components/CommandPalette.tsx` - 命令面板组件
  - 实现搜索功能
  - 实现命令列表
  - 实现快捷键打开/关闭

- [ ] Task 1.4: 创建 `src/styles/globals.css` - 全局样式
  - CSS 变量定义
  - 全局样式重置
  - 滚动条样式

## 阶段 2: 核心页面恢复

- [ ] Task 2.1: 创建 `src/pages/StatusPage.tsx` - 状态页面
  - 显示运行状态
  - 显示版本信息
  - 快捷键提示显示

- [ ] Task 2.2: 创建 `src/pages/ConfigPage.tsx` - 配置页面
  - API 配置表单
  - 配置保存功能

- [ ] Task 2.3: 创建其他基础页面组件
  - FeishuPage
  - SkillsPage
  - DiagnosisPage
  - AboutPage

## 阶段 3: 快捷键系统实现

- [ ] Task 3.1: 创建 `src/hooks/useKeyboardShortcut.ts` - 快捷键 Hook
  - 实现快捷键注册/注销
  - 实现快捷键冲突检测
  - 实现作用域支持

- [ ] Task 3.2: 创建 `src/components/KeyboardShortcutHints.tsx` - 快捷键提示组件
  - 显示快捷键标签
  - 主题适配

- [ ] Task 3.3: 在 App.tsx 中集成全局快捷键
  - 注册全局快捷键
  - 实现快捷键处理逻辑

- [ ] Task 3.4: 实现编辑操作快捷键
  - Ctrl+C/V/Z/A/X 等
  - 输入框内编辑操作优先

- [ ] Task 3.5: 实现导航快捷键
  - Ctrl+1/2/3/4 页面切换
  - Ctrl+, 设置页面

## 阶段 4: 快捷键状态管理

- [ ] Task 4.1: 实现应用状态检测
  - 检测输入框聚焦状态
  - 检测模态框打开状态

- [ ] Task 4.2: 实现快捷键作用域切换
  - 根据应用状态切换快捷键响应
  - 优先级处理

## 阶段 5: 测试与验证

- [ ] Task 5.1: 验证快捷键在不同操作系统下正常工作
- [ ] Task 5.2: 验证快捷键在输入框聚焦时正确处理
- [ ] Task 5.3: 验证快捷键在模态框打开时正确处理
- [ ] Task 5.4: 验证快捷键提示 UI 显示正确

## Task Dependencies

- Task 1.1 依赖: 无
- Task 1.2 依赖: Task 1.1 (App.tsx 需要 Sidebar)
- Task 1.3 依赖: Task 1.1
- Task 1.4 依赖: Task 1.1
- Task 2.1 依赖: Task 1.1
- Task 3.1 依赖: Task 1.1
- Task 3.2 依赖: Task 1.1
- Task 3.3 依赖: Task 1.1, Task 3.1
- Task 4.1 依赖: Task 3.1
- Task 5.1 依赖: Task 3.3, Task 4.1
