# 🚀 Tortoise 项目启动指南

## 环境要求

- **Node.js** 18+ (用于前端)
- **Go** 1.21+ (用于后端)
- **Git** (可选)

---

## 快速启动

### 方式一：一键启动（推荐 Windows 用户）

```batch
双击运行: start-all.bat
```

### 方式二：分别启动

#### 步骤 1：启动后端 API 服务器

```bash
# 进入后端目录
cd server

# 下载依赖
go mod tidy

# 启动服务器
go run ./cmd/api
```

#### 步骤 2：启动前端 Web UI

```bash
# 新开一个终端，进入前端目录
cd ui/web

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

---

## 访问地址

启动成功后，打开浏览器访问：

| 服务 | 地址 | 说明 |
|------|------|------|
| **Web UI** | http://localhost:3000 | 前端界面 |
| **API Server** | http://localhost:18792 | 后端 API |
| **API 文档** | http://localhost:18792/api/v1/health | 健康检查 |

---

## 功能验证

### 1. 检查 API 是否正常运行

浏览器打开：
```
http://localhost:18792/api/v1/health
```

应该看到：
```json
{"status":"healthy","timestamp":1704067200}
```

### 2. 检查前端是否正常运行

浏览器打开：
```
http://localhost:3000
```

应该看到 Tortoise 管理界面。

---

## 常见问题

### Q1: 端口被占用

如果端口被占用，修改以下配置：

**后端端口** (server/cmd/api/main.go):
```go
addr := ":18792"  // 改为其他端口，如 :8080
```

**前端端口** (ui/web/vite.config.ts):
```ts
server: {
  port: 3000  // 改为其他端口，如 5173
}
```

### Q2: Go 依赖下载失败

```bash
# 设置代理
go env -w GOPROXY=https://goproxy.cn,direct

# 重新下载依赖
cd server
go mod tidy
```

### Q3: Node 依赖安装失败

```bash
# 清理缓存
npm cache clean --force

# 使用淘宝镜像
npm config set registry https://registry.npmmirror.com

# 重新安装
cd ui/web
rm -rf node_modules
npm install
```

### Q4: 前端无法连接后端

检查浏览器控制台是否有跨域错误。

确保后端已启动，并检查 API 地址配置：
- 文件：`ui/web/src/services/api.ts`
- 默认地址：`http://localhost:18792`

---

## 停止服务

### 方式一：使用停止脚本

```batch
双击运行: stop-all.bat
```

### 方式二：手动停止

在运行后端/前端的终端按 `Ctrl+C`

或使用任务管理器结束进程：
- `go.exe` (后端)
- `node.exe` (前端)

---

## 项目文件结构

```
tohelp/
├── ui/web/               # 前端 (React)
│   ├── src/
│   │   ├── components/   # 组件
│   │   ├── pages/       # 页面
│   │   ├── services/    # API
│   │   └── store/       # 状态
│   └── package.json
├── server/               # 后端 (Go)
│   ├── cmd/api/         # 入口
│   └── internal/
│       ├── api/         # API
│       └── store/       # 存储
├── start-all.bat        # 一键启动
└── stop-all.bat        # 停止服务
```

---

## 技术支持

如果遇到问题，请检查：

1. Node.js 版本：`node --version` (需要 18+)
2. Go 版本：`go version` (需要 1.21+)
3. 端口占用：`netstat -ano | findstr "3000 18792"`

---

**祝使用愉快！🐢**
