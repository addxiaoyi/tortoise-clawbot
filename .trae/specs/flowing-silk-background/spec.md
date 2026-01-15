# 专业流动丝线背景设计 Spec

## Why
当前波纹动画配色诡异,与主题不协调。需要替换为专业优雅的流动丝线效果,全屏渲染,鼠标跟随交互,配色与主题系统协调统一。

## What Changes
- 移除现有的波纹凝胶效果
- 实现全局全屏流动丝线背景
- 丝线颜色随主题自适应(浅色/深色/各主题)
- 鼠标跟随交互 - 丝线自然向鼠标位置流动
- 专业优雅的视觉风格

## Impact
- Affected specs: UI/UX 体验优化
- Affected code: `src/components/GelBackground.tsx`

## ADDED Requirements
### Requirement: 流动丝线背景
系统应提供专业优雅的流动丝线背景:

#### Scenario: 页面加载
- **WHEN** 用户打开应用
- **THEN** 多条半透明丝线从左向右平滑流动
- **THEN** 丝线呈渐变透明,不影响内容阅读

#### Scenario: 鼠标移动
- **WHEN** 用户在窗口内移动鼠标
- **THEN** 丝线自然向鼠标位置偏移,产生引力效果
- **THEN** 偏移量平滑过渡,无突兀感

#### Scenario: 主题切换
- **WHEN** 用户切换主题(浅色/深色/VS Code/GitHub/Nord/Solarized)
- **THEN** 丝线颜色自动适配新主题
- **THEN** 配色保持专业低调,与主题协调

#### Scenario: 触摸设备
- **WHEN** 用户在触摸屏上滑动
- **THEN** 产生类似的丝线流动效果

### Requirement: 性能优化
- 丝线数量适中(6-10条)
- 60fps 流畅渲染
- 不影响页面交互响应

## MODIFIED Requirements
- 移除现有的波纹凝胶效果组件

## REMOVED Requirements
- 移除 GelBackground 中的波纹效果实现
