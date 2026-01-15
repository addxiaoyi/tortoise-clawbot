# Supabase 迁移

本目录存放 **Postgres / Supabase** 侧 SQL 迁移，与 Tortoise 桌面端使用的 `public.profiles` 等表对齐。

## 应用方式

1. 安装 [Supabase CLI](https://supabase.com/docs/guides/cli)，在项目根或 `supabase/` 下登录：`npx supabase login`。
2. 关联远程项目：`npx supabase link --project-ref <你的-project-ref>`。
3. 推送迁移：`npx supabase db push`（或在 Dashboard → SQL Editor 中**按文件名顺序**手动执行 `migrations/*.sql`）。

## 迁移内容摘要

| 文件 | 说明 |
|------|------|
| `20250405120000_profiles_rls_audit_rpc.sql` | `profiles` 行级安全、审计表 `audit_profile_actions`、触发器与 `admin_profile_set` RPC |
| `20250406120000_profiles_permissions.sql` | `profiles.permissions`（text[]）、非管理员不可自改、`admin_profile_set` 支持 `p_permissions` |

**首次管理员**：迁移不会自动创建管理员。在 Dashboard 的 SQL Editor 中执行（将 `YOUR_USER_UUID` 换成真实用户 id）：

```sql
update public.profiles set role = 'admin' where id = 'YOUR_USER_UUID';
```

## 本地开发

若使用 `supabase start` 本地栈，将 `migrations/` 同步到本地实例后，Tortoise 的 `VITE_SUPABASE_*` 指向本地 API URL 即可联调。
