# cloud/web

放在 `cloud/` 目录下的最小网站骨架（Vite + 原生 JS）。

## 启动

```bash
cd cloud/web
npm install
npm run dev
```

默认地址：`http://127.0.0.1:5173`

## 环境变量

复制 `.env.example` 为 `.env`：

- `VITE_CLOUD_API_BASE`：Go BFF 地址（示例 `http://127.0.0.1:8080`）
- `VITE_CLOUD_API_BASE_PATH`：预留给后续登录路由拼接（示例 `/auth`）

## 当前功能

- 显示当前 API Base 与 Auth Base Path
- 提供邮箱+密码登录/注册表单，调用 `${VITE_CLOUD_API_BASE}${VITE_CLOUD_API_BASE_PATH}/signin` 与 `/signup`
- 提供退出登录按钮，调用 `${VITE_CLOUD_API_BASE}${VITE_CLOUD_API_BASE_PATH}/signout`
- 请求头包含 `rid: emailpassword`，请求体为 SuperTokens `formFields` 结构
- 登录后可点击按钮调用 `${VITE_CLOUD_API_BASE}/api/me` 检查会话态
- 页面加载时会自动请求 `/api/me` 并显示当前会话状态（已登录/未登录）
- 常见认证错误会转换为更易读的中文提示（如密码错误、邮箱已存在）
- 请求进行中会统一禁用操作按钮，避免重复点击产生并发请求
- 输出区为带时间戳的操作日志，并提供“一键清空日志”按钮
- 支持“复制日志到剪贴板”和“导出日志为 txt 文件”
- 支持日志过滤（全部 / 仅错误）与“仅导出错误日志”
- 保留 `${VITE_CLOUD_API_BASE}/health` 健康检查
