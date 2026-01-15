# Tasks - 完善多核心 Agent 支持与新手老手切换

## 阶段 1: 功能验证

- [x] Task 1.1: 验证入场动画播放
  - 检查 IntroAnimation 组件是否正常渲染 ✅
  - 检查 Canvas 动画是否正常 ✅
  - 检查跳过按钮功能 ✅

- [x] Task 1.2: 验证核心选择流程
  - 检查 phase === "core-agent" 时显示核心选择 ✅
  - 检查三个核心卡片是否正确渲染 ✅
  - 检查优点/缺点显示 ✅
  - 检查选中状态高亮 ✅

- [x] Task 1.3: 验证模式选择流程
  - 检查 phase === "mode" 时显示模式选择 ✅
  - 检查老手/新手选项 ✅
  - 检查返回按钮功能 ✅

- [x] Task 1.4: 验证状态持久化
  - 检查 useCoreAgentStore 持久化 ✅
  - 检查 useOnboardingStore 持久化 ✅
  - 验证重启后选择保持 ✅

## 阶段 2: 错误处理验证

- [x] Task 2.1: App.tsx Tauri API 错误处理
  - getCurrent() / onOpenUrl() try-catch 已添加 ✅
  - handleDragStart() try-catch 已添加 ✅

- [x] Task 2.2: AboutPage.tsx openUrl 错误处理
  - openUrl 调用添加 try-catch ✅
  - 浏览器环境回退到 window.open ✅

- [x] Task 2.3: SkillsPage.tsx openUrl 错误处理
  - openUrl 调用已有 .catch() 处理 ✅

## 阶段 3: UI/UX 验证

- [x] Task 3.1: 验证中文字符串完整
  - 核心选择界面文本完整 ✅
  - 模式选择界面文本完整 ✅
  - 关于我们页面文本完整 ✅

- [x] Task 3.2: 验证插件兼容性标签
  - getCompatibilityLabel 函数逻辑正确 ✅
  - OpenClaw 专属显示 🔵 仅 OpenClaw ✅
  - myclaw 专属显示 🟢 仅 myclaw ✅
  - nanobot 专属显示 🟣 仅 nanobot ✅
  - 全平台显示 ✅ 全平台 ✅

## 阶段 4: 构建验证

- [x] Task 4.1: TypeScript 类型检查
  - npx tsc --noEmit 无错误 ✅

- [x] Task 4.2: Vite 构建验证
  - npm run build 成功 ✅

## Task Dependencies

```
Task 1.1 → Task 1.2 → Task 1.3 → Task 1.4
Task 1.4 → Task 2.1 → Task 2.2 → Task 2.3
Task 2.3 → Task 3.1 → Task 3.2
Task 3.2 → Task 4.1 → Task 4.2
```

## 验证文件清单

### 核心文件
- src/lib/coreAgents.ts - 核心数据定义 ✅
- src/lib/pluginCompatibility.ts - 插件兼容性 ✅
- src/store/useCoreAgentStore.ts - 核心状态管理 ✅
- src/store/useOnboardingStore.ts -  onboarding 状态 ✅

### UI 组件
- src/components/SplashScreen.tsx - 入场动画 + 核心选择 + 模式选择 ✅
- src/components/CoreAgentSelector.tsx - 独立核心选择组件（备份）✅
- src/pages/AboutPage.tsx - 关于我们页面 ✅

### 错误处理
- src/App.tsx - getCurrent(), onOpenUrl(), handleDragStart() ✅
- src/pages/AboutPage.tsx - openUrl() ✅
