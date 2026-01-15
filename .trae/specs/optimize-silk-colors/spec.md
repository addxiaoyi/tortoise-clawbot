# 丝线背景配色与疏密优化 Spec

## Why
当前丝线背景配色与部分主题不协调,疏密不够自然均衡。需要检查每个主题的配色方案,确保对比度鲜明、视觉效果专业舒适,同时优化丝线的疏密变化。

## What Changes
- 优化各主题的丝线配色方案
- 确保配色与主题背景形成鲜明对比
- 优化丝线疏密分布,使其更自然均衡
- 调整丝线速度和波动幅度

## Impact
- Affected specs: UI/UX 体验优化
- Affected code: `src/components/GelBackground.tsx`

## ADDED Requirements
### Requirement: 主题配色优化
各主题丝线配色应符合以下标准:

#### Scenario: 浅色主题
- **THEN** 丝线颜色应比背景稍深,形成柔和对比
- **THEN** 避免使用过亮颜色,保持专业低调

#### Scenario: 深色主题
- **THEN** 丝线颜色应比背景更亮,确保可见性
- **THEN** 避免过于刺眼的荧光色

### Requirement: 丝线疏密优化
- **THEN** 丝线分布应疏密有致,避免过于密集或稀疏
- **THEN** 波动幅度应有变化,部分丝线平缓、部分波动明显
- **THEN** 丝线速度应有差异化,增加自然感

### Requirement: 视觉舒适度
- **THEN** 丝线整体透明度应适中,不抢眼
- **THEN** 阴影效果柔和,不影响阅读

## MODIFIED Requirements
### Requirement: 现有丝线效果
- 优化配色方案
- 优化疏密分布

## REMOVED Requirements
- 无
