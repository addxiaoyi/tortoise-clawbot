# 项目完整性检查规范

## Why

需要全面检查 tortoise 项目的完整性，确保所有功能模块正常可用，包括核心选择、页面完整性、错误处理和构建状态。

## What Changes

### 1. 项目结构检查
- 核心文件存在性验证
- 组件完整性检查
- 页面文件检查

### 2. 功能完整性检查
- 入场动画 + 核心选择 + 模式选择
- 关于我们页面
- 技能页面
- 错误处理

### 3. 构建状态检查
- TypeScript 类型检查
- Vite 构建验证

## Impact

### 受影响的代码
- src/components/SplashScreen.tsx
- src/components/CoreAgentSelector.tsx
- src/pages/AboutPage.tsx
- src/pages/SkillsPage.tsx
- src/lib/coreAgents.ts
- src/lib/pluginCompatibility.ts
- src/store/useCoreAgentStore.ts
- src/store/useOnboardingStore.ts
- src/App.tsx

## 检查清单

### 项目结构
- [ ] src/lib/coreAgents.ts 存在且完整
- [ ] src/lib/pluginCompatibility.ts 存在且完整
- [ ] src/store/useCoreAgentStore.ts 存在且完整
- [ ] src/store/useOnboardingStore.ts 存在且完整
- [ ] src/components/SplashScreen.tsx 存在且完整
- [ ] src/components/CoreAgentSelector.tsx 存在
- [ ] src/pages/AboutPage.tsx 存在且完整

### 功能完整性
- [ ] 入场动画播放
- [ ] 核心选择（OpenClaw/myclaw/nanobot）
- [ ] 模式选择（老手/新手）
- [ ] 关于我们页面显示当前核心
- [ ] 重新选择核心功能

### 错误处理
- [ ] App.tsx Tauri API 错误处理
- [ ] AboutPage.tsx openUrl 错误处理

### 构建状态
- [ ] TypeScript 类型检查通过
- [ ] Vite 构建成功
