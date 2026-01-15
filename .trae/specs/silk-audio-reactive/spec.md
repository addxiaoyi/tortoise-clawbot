# 丝线声音感知交互 Spec

## Why
丝线背景目前是静态的动画效果,缺乏与用户的互动感。需要增加麦克风/设备声音感知功能,让丝线能够根据声音强度做出反应,增强沉浸式体验。

## What Changes
- 增加麦克风音频输入监听
- 增加系统音量/音频输出感知(可选)
- 丝线根据声音强度做出反应(波动幅度、速度、粗细)

## Impact
- Affected specs: UI/UX 体验优化
- Affected code: `src/components/GelBackground.tsx`

## ADDED Requirements
### Requirement: 麦克风音频感知
系统应获取麦克风输入并转化为可视化效果:

#### Scenario: 麦克风权限允许
- **WHEN** 用户允许麦克风权限
- **THEN** 实时获取麦克风音量
- **THEN** 丝线根据音量强度做出反应

#### Scenario: 麦克风权限拒绝
- **WHEN** 用户拒绝麦克风权限
- **THEN** 优雅降级到普通动画模式
- **THEN** 不显示错误提示,保持静默

### Requirement: 丝线声音反应
丝线对声音做出以下反应:

#### Scenario: 音量增大
- **THEN** 丝线波动幅度增加(最大 2 倍)
- **THEN** 丝线流动速度加快
- **THEN** 丝线亮度/粗细略微增加

#### Scenario: 音量减小/静音
- **THEN** 丝线恢复平稳流动状态
- **THEN** 动画平滑过渡,不突兀

### Requirement: 性能优化
- 音频处理使用 Web Audio API
- 分析频率限制(60fps 同步)
- 麦克风监听可随时开启/关闭

## MODIFIED Requirements
### Requirement: 现有丝线效果
- 丝线效果保持不变
- 增加声音感知层

## REMOVED Requirements
- 无
