# Checklist - 快捷键系统验证

## 阶段 1: 项目基础结构

- [ ] `src/App.tsx` 文件存在且正确导出 App 组件
- [ ] `src/components/Sidebar.tsx` 文件存在且导出 Sidebar 组件
- [ ] `src/components/CommandPalette.tsx` 文件存在且可正常打开/关闭
- [ ] `src/styles/globals.css` 包含必要的 CSS 变量和样式

## 阶段 2: 快捷键 Hook

- [ ] `src/hooks/useKeyboardShortcut.ts` 正确实现快捷键注册
- [ ] `useKeyboardShortcut` 支持 key、ctrl、shift、alt、meta 参数
- [ ] `useKeyboardShortcut` 支持 enabled 参数控制是否启用
- [ ] 组件卸载时正确取消注册快捷键

## 阶段 3: 全局快捷键

- [ ] `Ctrl+K` / `Cmd+K` 打开命令面板
- [ ] `Escape` 关闭命令面板
- [ ] `Ctrl+,` / `Cmd+,` 打开设置
- [ ] `Ctrl+1` / `Cmd+1` 切换到状态页
- [ ] `Ctrl+2` / `Cmd+2` 切换到配置页
- [ ] `Ctrl+3` / `Cmd+3` 切换到飞书页
- [ ] `Ctrl+4` / `Cmd+4` 切换到技能页

## 阶段 4: 编辑操作快捷键

- [ ] `Ctrl+C` / `Cmd+C` 复制功能正常
- [ ] `Ctrl+V` / `Cmd+V` 粘贴功能正常
- [ ] `Ctrl+Z` / `Cmd+Z` 撤销功能正常
- [ ] `Ctrl+Shift+Z` / `Cmd+Shift+Z` 重做功能正常
- [ ] `Ctrl+A` / `Cmd+A` 全选功能正常
- [ ] `Ctrl+X` / `Cmd+X` 剪切功能正常

## 阶段 5: 应用状态处理

- [ ] 输入框聚焦时编辑快捷键正常响应
- [ ] 模态框打开时快捷键优先传递给模态框
- [ ] 模态框关闭后快捷键恢复正常响应

## 阶段 6: 快捷键提示 UI

- [ ] 命令面板中显示各命令的快捷键
- [ ] 快捷键提示随主题变化
- [ ] 快捷键提示在不同页面正确显示

## 阶段 7: 跨平台验证

- [ ] Windows 系统快捷键正常
- [ ] macOS 系统快捷键正常
- [ ] Linux 系统快捷键正常

## 阶段 8: TypeScript 类型检查

- [ ] `npx tsc --noEmit` 通过无错误
- [ ] 所有组件正确导入和导出
- [ ] Hook 返回值类型正确
