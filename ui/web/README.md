# Tortoise Web UI

Tortoise AI Agent 框架的 Web 管理界面

## 功能特性

- 🎨 **现代深色主题** - 精心设计的深色 UI
- 💬 **对话管理** - 创建、切换、删除会话
- 🧠 **记忆系统** - 三层记忆管理（工作/语义/情景）
- 🧩 **插件管理** - 安装、启用、禁用插件
- ⚙️ **配置设置** - 丰富的配置选项

## 快速开始

### 安装依赖

```bash
cd ui/web
npm install
```

### 启动开发服务器

```bash
npm run dev
```

访问 http://localhost:3000

### 构建生产版本

```bash
npm run build
```

## 技术栈

- **框架**: React 18 + TypeScript
- **构建工具**: Vite
- **路由**: React Router v6
- **状态管理**: Zustand
- **样式**: Tailwind CSS
- **图标**: Lucide React
- **HTTP 客户端**: Axios
- **日期处理**: date-fns

## 项目结构

```
src/
├── components/     # 可复用组件
│   ├── Layout.tsx     # 页面布局
│   ├── Sidebar.tsx   # 侧边栏
│   └── Header.tsx    # 顶部栏
├── pages/          # 页面组件
│   ├── Dashboard.tsx # 仪表盘
│   ├── Chat.tsx      # 对话页面
│   ├── Memory.tsx    # 记忆管理
│   ├── Plugins.tsx   # 插件管理
│   └── Settings.tsx  # 设置页面
├── services/       # API 服务
│   └── api.ts       # API 客户端
├── store/          # 状态管理
│   └── appStore.ts  # Zustand Store
└── App.tsx         # 根组件
```

## 页面预览

### 仪表盘
- 系统运行状态
- 统计数据卡片
- 最近会话列表
- 快速操作入口

### 对话页面
- 会话列表侧边栏
- 消息显示区域
- 输入框和发送按钮
- 新建会话弹窗

### 记忆管理
- 三层记忆类型统计
- 记忆搜索和筛选
- 添加/删除记忆
- 重要性评分

### 插件管理
- 插件列表和状态
- 启用/禁用切换
- 插件安装功能
- 工具列表展示

### 设置页面
- 通用设置
- 外观主题
- 连接配置
- 高级选项

## API 对接

前端默认连接 `http://localhost:18792`，可通过修改 `src/services/api.ts` 中的 `baseURL` 来更改。

### API 端点

| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/v1/health` | GET | 健康检查 |
| `/api/v1/sessions` | GET/POST | 会话列表/创建 |
| `/api/v1/sessions/:id` | GET/DELETE | 会话详情/删除 |
| `/api/v1/sessions/:id/messages` | GET/POST | 消息列表/发送 |
| `/api/v1/memories` | GET/POST | 记忆列表/添加 |
| `/api/v1/plugins` | GET | 插件列表 |
| `/api/v1/plugins/:id` | PATCH/DELETE | 插件更新/删除 |

## 开发指南

### 添加新页面

1. 在 `src/pages/` 创建页面组件
2. 在 `src/App.tsx` 添加路由
3. 在 `src/store/appStore.ts` 添加相关状态

### 添加新 API

在 `src/services/api.ts` 中添加方法：

```typescript
async getNewEndpoint() {
  const response = await this.client.get('/api/v1/new')
  return response.data
}
```

## 许可证

Apache 2.0
