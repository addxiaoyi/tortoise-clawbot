# 医生诊断页面完善任务列表

## Task 1: 创建健康度评分组件
创建 `HealthScoreCard` 组件，展示系统健康度评分和趋势
- [x] 创建 `tortoise/src/components/HealthScoreCard.tsx`
- [x] 实现圆形进度条展示评分(0-100)
- [x] 添加评分等级颜色映射(优秀/良好/一般/差)
- [x] 实现趋势箭头(上升/下降/持平)
- [x] 添加与上次诊断的对比显示
- [x] 支持点击展开详情

## Task 2: 创建检查模块卡片组件
创建 `CheckModuleCard` 组件，展示四大检查模块的详细信息
- [x] 创建 `tortoise/src/components/CheckModuleCard.tsx`
- [x] 实现展开/收起功能
- [x] 添加模块图标和状态指示
- [x] 实现子检查项列表展示
- [x] 添加每个子项的状态图标(通过/警告/错误)
- [x] 支持点击子项查看详情

## Task 3: 创建诊断历史记录组件
创建 `DiagnosisHistory` 组件，管理诊断历史记录
- [x] 创建 `tortoise/src/components/DiagnosisHistory.tsx`
- [x] 实现历史记录列表展示
- [x] 添加时间格式化显示
- [x] 实现历史记录对比功能
- [x] 添加删除单条历史记录功能
- [x] 实现历史记录持久化存储

## Task 4: 创建修复进度面板组件
创建 `FixProgressPanel` 组件，优化修复流程展示
- [x] 创建 `tortoise/src/components/FixProgressPanel.tsx`
- [x] 实现步骤进度条
- [x] 添加当前步骤高亮
- [x] 实现实时日志输出区域
- [x] 添加预计剩余时间显示
- [x] 实现修复结果统计展示

## Task 5: 扩展AppStore添加诊断历史状态
在 `useAppStore` 中添加诊断历史相关的状态和方法
- [x] 添加 `diagnosisHistory` 状态
- [x] 添加 `addDiagnosisRecord` 方法
- [x] 添加 `clearDiagnosisHistory` 方法
- [x] 添加 `getDiagnosisHistory` 方法
- [x] 实现历史记录本地存储持久化

## Task 6: 重构DiagnosisPage主页面
整合所有新组件，完善主页面功能
- [x] 在页面顶部添加 `HealthScoreCard`
- [x] 替换原有检查项列表为 `CheckModuleCard`
- [x] 集成 `DiagnosisHistory` 侧边栏
- [x] 使用 `FixProgressPanel` 替换原有修复进度展示
- [x] 移除页面底部的占位提示文本
- [x] 优化整体布局和样式

## Task 7: 添加诊断报告导出功能
实现诊断报告导出
- [x] 实现报告数据组装逻辑
- [x] 添加导出按钮和文件保存对话框
- [x] 支持导出为JSON格式

## Task 8: 完善后端接口调用
在 `tauri.ts` 中添加必要的后端接口
- [x] `runFullDiagnosis` 接口已存在
- [x] `runFullFix` 接口已存在
- [x] 各种修复操作接口已存在

# Task Dependencies
- Task 6 依赖于 Task 1, 2, 3, 4, 5
- Task 7 依赖于 Task 6
- Task 8 可以与 Task 1-5 并行进行
