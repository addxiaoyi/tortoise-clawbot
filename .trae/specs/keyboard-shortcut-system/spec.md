# 快捷键系统规范

## Why
应用程序需要一套完整、可用的快捷键系统，使用户能够高效地通过键盘操作完成常用任务，提高用户体验和操作效率。同时需要确保快捷键在不同操作系统和不同应用状态下都能正常工作。

## What Changes

### 核心变更
- 创建全局快捷键管理系统 (`useKeyboardShortcut` hook)
- 实现命令面板快捷键 (Command Palette)
- 实现编辑操作快捷键 (复制、粘贴、撤销、重做等)
- 实现导航快捷键 (页面切换、焦点移动)
- 实现特定功能快捷键 (设置、帮助、主题切换)
- 添加快捷键提示 UI 组件
- 确保快捷键在不同应用状态下正确响应

### 项目恢复
由于项目文件被删除，需要重新创建以下核心文件：
- `App.tsx` - 应用主入口
- `Sidebar.tsx` - 侧边栏导航
- `StatusPage.tsx` - 状态页面
- `ConfigPage.tsx` - 配置页面
- 其他必要页面和组件

## Impact
- 受影响的功能：全局快捷键、命令面板、编辑操作、导航系统
- 受影响的代码：主要前端文件

## ADDED Requirements

### Requirement: 全局快捷键管理系统
系统 SHALL 提供统一的快捷键注册和管理机制

#### Scenario: 快捷键注册
- **WHEN** 组件挂载时注册快捷键
- **THEN** 快捷键应在指定条件下触发对应操作
- **AND** 组件卸载时应自动取消注册

#### Scenario: 快捷键冲突检测
- **WHEN** 注册重复快捷键时
- **THEN** 应给出警告或覆盖已有快捷键

### Requirement: 命令面板快捷键
系统 SHALL 提供命令面板的快捷键触发机制

#### Scenario: 打开命令面板
- **WHEN** 用户按下 `Ctrl+K` (Windows/Linux) 或 `Cmd+K` (macOS)
- **THEN** 命令面板应打开

#### Scenario: 关闭命令面板
- **WHEN** 命令面板打开时用户按下 `Escape`
- **THEN** 命令面板应关闭

### Requirement: 编辑操作快捷键
系统 SHALL 支持常用编辑操作的快捷键

| 操作 | Windows/Linux | macOS |
|------|--------------|-------|
| 复制 | Ctrl+C | Cmd+C |
| 粘贴 | Ctrl+V | Cmd+V |
| 撤销 | Ctrl+Z | Cmd+Z |
| 重做 | Ctrl+Shift+Z | Cmd+Shift+Z |
| 全选 | Ctrl+A | Cmd+A |
| 剪切 | Ctrl+X | Cmd+X |

### Requirement: 导航快捷键
系统 SHALL 支持页面导航的快捷键

| 操作 | Windows/Linux | macOS |
|------|--------------|-------|
| 状态页 | Ctrl+1 | Cmd+1 |
| 配置页 | Ctrl+2 | Cmd+2 |
| 飞书页 | Ctrl+3 | Cmd+3 |
| 技能页 | Ctrl+4 | Cmd+4 |

### Requirement: 快捷键提示显示
系统 SHALL 在 UI 中显示可用的快捷键提示

#### Scenario: 显示快捷键提示
- **WHEN** 界面上有可用的快捷键操作
- **THEN** 应在相应位置显示快捷键提示
- **AND** 提示应随主题变化

### Requirement: 应用状态下的快捷键响应
系统 SHALL 根据应用状态正确响应快捷键

#### Scenario: 输入框聚焦时
- **WHEN** 用户在输入框中聚焦时按下快捷键
- **THEN** 应优先执行输入框的编辑操作
- **AND** 全局快捷键应在输入框失焦后恢复

#### Scenario: 模态框打开时
- **WHEN** 模态框打开时按下快捷键
- **THEN** 快捷键应优先传递给模态框处理
- **AND** 模态框关闭后恢复正常快捷键响应

## MODIFIED Requirements

### Requirement: 已有功能
无 - 本次为全新实现

## REMOVED Requirements

### Requirement: 旧功能
无

## 技术实现细节

### 快捷键优先级
1. 输入框内编辑操作 (最高)
2. 模态框/对话框操作
3. 全局快捷键 (最低)

### 操作系统检测
- 使用 `navigator.platform` 检测操作系统
- Windows: `Win32` / `Windows`
- macOS: `MacIntel` / `MacPPC`
- Linux: `Linux x86_64` / `Linux i686`

### 快捷键 Hook 设计
```typescript
interface ShortcutConfig {
  key: string;
  ctrl?: boolean;
  shift?: boolean;
  alt?: boolean;
  meta?: boolean;
  handler: () => void;
  enabled?: boolean;
  scope?: 'global' | 'input' | 'modal';
}
```
