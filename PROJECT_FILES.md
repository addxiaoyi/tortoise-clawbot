# 🐢 Tortoise 项目文件清单

## 前端 (ui/web)

### 配置文件
- [x] `package.json` - NPM 依赖配置
- [x] `vite.config.ts` - Vite 构建配置
- [x] `tsconfig.json` - TypeScript 配置
- [x] `tsconfig.node.json` - Node 类型配置
- [x] `tailwind.config.js` - Tailwind CSS 配置
- [x] `postcss.config.js` - PostCSS 配置
- [x] `index.html` - HTML 入口

### 源代码
- [x] `src/main.tsx` - React 入口
- [x] `src/App.tsx` - 根组件 + 路由
- [x] `src/index.css` - 全局样式

### 组件 (src/components)
- [x] `Layout.tsx` - 页面布局
- [x] `Sidebar.tsx` - 侧边栏导航
- [x] `Header.tsx` - 顶部状态栏

### 页面 (src/pages)
- [x] `Dashboard.tsx` - 仪表盘页面
- [x] `Chat.tsx` - 对话页面
- [x] `Memory.tsx` - 记忆管理页面
- [x] `Plugins.tsx` - 插件管理页面
- [x] `Settings.tsx` - 设置页面

### 服务 (src/services)
- [x] `api.ts` - API 客户端

### 状态管理 (src/store)
- [x] `appStore.ts` - Zustand 状态存储

### 脚本
- [x] `run-web.bat` - Windows 启动脚本
- [x] `README.md` - 前端文档

---

## 后端 (server)

### 入口
- [x] `go.mod` - Go 模块配置
- [x] `cmd/api/main.go` - API 服务器入口

### API 层 (internal/api)
- [x] `server.go` - 服务器主文件 + 路由
- [x] `handlers.go` - API 处理器

### 存储层 (internal/store)
- [x] `session_store.go` - 会话存储
- [x] `message_store.go` - 消息存储
- [x] `memory_store.go` - 记忆存储
- [x] `plugin_store.go` - 插件存储
- [x] `utils.go` - 工具函数

### 脚本
- [x] `build-api.bat` - 构建脚本
- [x] `run-api.bat` - 启动脚本
- [x] `README.md` - 后端文档

---

## 根目录

### 脚本
- [x] `start-all.bat` - 一键启动前后端
- [x] `stop-all.bat` - 停止所有服务

### 文档
- [x] `README.md` - 项目主文档
- [x] `PROJECT_FILES.md` - 本文件

---

## 功能实现状态

### 前端功能 ✅
| 模块 | 状态 | 说明 |
|------|------|------|
| 仪表盘 | ✅ 完成 | 系统状态、统计数据、快速操作 |
| 对话页面 | ✅ 完成 | 会话列表、消息发送、AI回复模拟 |
| 记忆管理 | ✅ 完成 | 三层记忆、搜索、添加/删除 |
| 插件管理 | ✅ 完成 | 插件列表、启用/禁用、安装 |
| 设置页面 | ✅ 完成 | 通用、外观、连接配置 |
| 状态管理 | ✅ 完成 | Zustand 统一状态管理 |
| API 服务 | ✅ 完成 | Axios API 客户端 |
| 路由 | ✅ 完成 | React Router 页面路由 |
| 样式 | ✅ 完成 | Tailwind CSS 深色主题 |

### 后端功能 ✅
| 模块 | 状态 | 说明 |
|------|------|------|
| REST API | ✅ 完成 | Gin 框架实现 |
| 会话管理 | ✅ 完成 | CRUD + 内存存储 |
| 消息处理 | ✅ 完成 | 消息存储 + AI响应模拟 |
| 记忆系统 | ✅ 完成 | 三层记忆存储 + 搜索 |
| 插件管理 | ✅ 完成 | 插件列表 + 启用控制 |
| CORS | ✅ 完成 | 支持跨域请求 |
| 示例数据 | ✅ 完成 | 预置演示数据 |

---

## 启动方式

### 方式一：一键启动
```batch
双击 start-all.bat
```

### 方式二：分别启动

**终端 1 - 后端 API：**
```bash
cd server
go mod tidy
go run ./cmd/api
```

**终端 2 - 前端 Web：**
```bash
cd ui/web
npm install
npm run dev
```

---

## 访问地址

| 服务 | 地址 |
|------|------|
| Web UI | http://localhost:3000 |
| API Server | http://localhost:18792 |
| API 文档 | http://localhost:18792/api/v1 |

---

## 技术栈

| 层级 | 技术 |
|------|------|
| 前端框架 | React 18 + TypeScript |
| 构建工具 | Vite |
| 路由 | React Router v6 |
| 状态管理 | Zustand |
| UI 样式 | Tailwind CSS |
| 图标 | Lucide React |
| HTTP 客户端 | Axios |
| 后端框架 | Go + Gin |
| 数据存储 | 内存存储 |

---

## 项目特色

1. **现代深色主题** - 专业级深色 UI 设计
2. **响应式布局** - 适配各种屏幕尺寸
3. **流畅动画** - 微交互和过渡效果
4. **完整功能** - 仪表盘、对话、记忆、插件、设置
5. **RESTful API** - 标准化的后端接口
6. **示例数据** - 开箱即用的演示内容
7. **一键启动** - Windows 批处理脚本

---

最后更新: 2026-02-15
