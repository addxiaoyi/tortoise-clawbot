# Checklist - 完善多核心 Agent 支持与新手老手切换

## 阶段 1: 功能验证

### Task 1.1: 入场动画播放
- [x] IntroAnimation 组件正常渲染 ✅
- [x] Canvas 动画正常显示 ✅
- [x] 跳过按钮功能正常 ✅
- [x] 动画结束后触发 onComplete ✅

### Task 1.2: 核心选择流程
- [x] phase === "core-agent" 正确显示 ✅
- [x] 标题"选择你的核心操控对象"显示 ✅
- [x] 副标题"选择一个最适合你的 AI Agent 系统"显示 ✅
- [x] OpenClaw 卡片显示（🐉 图标、名称、描述、优点、缺点）✅
- [x] myclaw 卡片显示（🐢 图标、名称、描述、优点、缺点）✅
- [x] nanobot 卡片显示（🤖 图标、名称、描述、优点、缺点）✅
- [x] 点击选择后高亮正确 ✅
- [x] 选择后"继续"按钮可用 ✅

### Task 1.3: 模式选择流程
- [x] phase === "mode" 正确显示 ✅
- [x] 标题"选择你的模式"显示 ✅
- [x] "老手"选项显示（图标 + 描述）✅
- [x] "新手"选项显示（图标 + 描述）✅
- [x] 返回按钮功能正常 ✅
- [x] 选择模式后调用 onComplete ✅

### Task 1.4: 状态持久化
- [x] useCoreAgentStore 持久化到 localStorage ✅
- [x] useOnboardingStore 持久化到 localStorage ✅
- [x] 重启后核心选择保持 ✅
- [x] 重启后模式选择保持 ✅

## 阶段 2: 错误处理验证

### Task 2.1: App.tsx Tauri API
- [x] getCurrent() 调用添加 try-catch ✅
- [x] onOpenUrl() 调用添加 try-catch ✅
- [x] handleDragStart() 添加 null 检查 ✅
- [x] 浏览器环境不崩溃 ✅

### Task 2.2: AboutPage.tsx openUrl
- [x] openUrl(REPO_URL) 添加 try-catch ✅
- [x] openUrl(RELEASES_URL) 添加 try-catch ✅
- [x] 浏览器环境回退到 window.open ✅

### Task 2.3: SkillsPage.tsx openUrl
- [x] openUrl 调用已有 .catch() 处理 ✅

## 阶段 3: UI/UX 验证

### Task 3.1: 中文字符串完整
- [x] 核心选择标题：选择你的核心操控对象 ✅
- [x] 核心选择副标题：选择一个最适合你的 AI Agent 系统 ✅
- [x] 优点标签：优点 ✅
- [x] 缺点标签：缺点 ✅
- [x] 继续按钮：继续 ✅
- [x] 选择提示：请选择一个核心开始 ✅
- [x] 模式选择标题：选择你的模式 ✅
- [x] 老手选项：老手 ✅
- [x] 新手选项：新手 ✅
- [x] 返回按钮：← 返回选择核心对象 ✅
- [x] 关于我们当前核心：当前核心 ✅

### Task 3.2: 插件兼容性标签
- [x] gateway 显示 🔵 仅 OpenClaw ✅
- [x] feishu-doc 显示 🟢 仅 myclaw ✅
- [x] multi-agent-orchestration 显示 🟣 仅 nanobot ✅
- [x] tavily-search 显示 ✅ 全平台 ✅
- [x] coding-agent 显示 ✅ 全平台 ✅

## 阶段 4: 构建验证

### Task 4.1: TypeScript 类型检查
- [x] npx tsc --noEmit 无错误 ✅

### Task 4.2: Vite 构建验证
- [x] npm run build 成功 ✅
- [x] dist/ 目录生成正确 ✅

## 验证场景

### 场景 1: 首次启动
1. 应用启动 → 显示入场动画 ✅
2. 动画播放 → 跳过或等待结束 ✅
3. 显示核心选择 → 选择一个核心 ✅
4. 点击继续 → 显示模式选择 ✅
5. 选择模式 → 进入主界面 ✅

### 场景 2: 重启应用
1. 应用启动 → 显示入场动画 ✅
2. 动画结束 → 检测到已选择核心和模式 ✅
3. 跳过核心和模式选择 → 直接进入主界面 ✅

### 场景 3: 在关于页面重新选择
1. 在主界面 → 进入关于页面 ✅
2. 点击重新选择核心 → 跳转到核心选择 ✅
3. 重新选择核心 → 进入主界面 ✅

## 核心数据结构验证

### OpenClaw 数据
- [x] id: 'openclaw' ✅
- [x] name: 'OpenClaw' ✅
- [x] icon: '🐉' ✅
- [x] color: '#007AFF' ✅
- [x] description: '成熟的开源 AI Agent 系统' ✅
- [x] advantages.length === 4 ✅
- [x] disadvantages.length === 2 ✅
- [x] exclusiveSkillCount === 19 ✅

### myclaw 数据
- [x] id: 'myclaw' ✅
- [x] name: 'myclaw' ✅
- [x] icon: '🐢' ✅
- [x] color: '#34C759' ✅
- [x] description: '多渠道集成的 AI 助手' ✅
- [x] advantages.length === 4 ✅
- [x] disadvantages.length === 2 ✅
- [x] exclusiveSkillCount === 16 ✅

### nanobot 数据
- [x] id: 'nanobot' ✅
- [x] name: 'nanobot' ✅
- [x] icon: '🤖' ✅
- [x] color: '#AF52DE' ✅
- [x] description: 'AI同事工作流系统' ✅
- [x] advantages.length === 4 ✅
- [x] disadvantages.length === 2 ✅
- [x] exclusiveSkillCount === 14 ✅
