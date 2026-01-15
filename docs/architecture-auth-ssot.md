# 身份与角色：单一事实来源（SSOT）

更广义的全栈拓扑（各组件职责、环境变量索引、CI、本地联调顺序、多仓拆分映射）见 **`docs/full-stack-architecture.md`**。

本仓库同时存在 **Supabase（Tortoise 桌面 / Web 前端）** 与 **SuperTokens + Go BFF（`cloud/cmd/server`）** 两条能力线。为避免「同一用户两套身份、角色不一致」，约定如下。

## 组件职责

| 组件 | 职责 |
|------|------|
| **Supabase Auth** | Tortoise 主登录、JWT、`auth.users` 与 `public.profiles` 行数据（含 `role`、`is_banned`）。 |
| **SuperTokens** | 可选：独立 Web / BFF 会话、Email-Password、与 Go BFF 的 `/auth` 路由。 |
| **Postgres `profiles`** | 业务档案与展示用角色；**Go BFF** 在配置 `DATABASE_URL` 时读取同一表，与 SuperTokens `userroles` 按 `CLOUD_PROFILE_ROLES_MODE` 合并。 |

## 推荐部署模式

### 模式 A：Tortoise 为主（默认产品形态）

- 用户身份以 **Supabase** 为准。
- `cloud` BFF **可选**：若仅桌面端使用 Tortoise，可不部署 BFF。
- 若仍部署 BFF（例如未来统一 Web 控制台），请将 **同一 Postgres**（含 `profiles`）接到 `DATABASE_URL`，并采用 `CLOUD_PROFILE_ROLES_MODE=db`，使库内 `role` 优先于 SuperTokens，避免双源冲突。

### 模式 B：BFF 为主（仅 SuperTokens 会话）

- 用户只走 SuperTokens，**不在** Tortoise 内使用 Supabase 登录。
- `profiles` 仍可由 BFF 读写；Tortoise 若连接 Supabase，应改为只读或关闭登录相关页。

## 管理员与敏感操作

- **Supabase**：通过迁移 `admin_profile_set` RPC + RLS 限制 `role` / `is_banned` 的随意修改；详见 `supabase/README.md`。
- **SuperTokens**：管理员权限仍可通过 Dashboard 配置 `userroles`；与库角色合并策略由 `CLOUD_PROFILE_*_MODE` 控制。

## 环境变量索引

| 位置 | 用途 |
|------|------|
| 仓库根 `.env.example` | OpenClaw / 网关 / 模型 Key |
| `tortoise/.env.example` | `VITE_SUPABASE_*`、可选 `VITE_ST_*` |
| `cloud/env.example` | `SUPERTOKENS_*`、`DATABASE_URL`、`CLOUD_*` |

变更身份方案时，请同步更新本文档与 `SUPABASE_INTEGRATION_GUIDE.md` 中的集成步骤。
