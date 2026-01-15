# 乌龟助手 (Tortoise) - Supabase & 桌面端「全量集成」保姆级指南 (v2.0)

本指南配合 Supabase Dashboard 的 **Integrations** 页面使用（路径形如 `https://supabase.com/dashboard/project/<YOUR_PROJECT_REF>/integrations`）。通过这些集成，Tortoise 可具备：**账户同步、实时通道、自动化任务分发**等能力。

---

## 1. 桌面开发环境配置 (Desktop Setup)

在开始集成前，请确保您的本地电脑已安装以下核心组件：

1. **Node.js (v18.0+)**: 用于运行 Vite 前端和桌面逻辑。
2. **Rust & Cargo**: Tauri 2.0 的核心渲染与后端逻辑依赖。
3. **Supabase CLI**: 用于本地开发及 Edge Functions 部署。
   * 安装命令：`npm install supabase --save-dev` 或 `scoop install supabase` (Windows)。
   * 登录授权：`npx supabase login`。

---

## 2. 核心云端扩展集成 (Integrations & Extensions)

请在 Dashboard → Integrations 中按需配置以下拓展：

### A. 身份验证与通信 (Resend + GitHub)

* **Resend (强烈推荐)**: 替代默认的邮件发送器。
  * **作用**: 解决 Magic Link 登录延迟，确保账号 0 秒登录。
  * **配置**: 在 Dashboard -> Integrations 搜索 Resend 并填入 API Key。
* **GitHub OAuth**: 在 Authentication -> Providers 中开启。
  * **作用**: 允许开发者一键同步 GitHub 代码库、Issue 状态及个人指纹。

### B. 性能监控与自动化 (Sentry + Vercel)

* **Sentry (必选)**: 集成在桌面运行时的错误追踪。
  * **作用**: 当用户的本地代理崩溃时，自动将日志通过 Supabase 实时通道回传到 Sentry。
* **Vercel Integration**:
  * **作用**: 托管 Tortoise 社区镜像和在线控制台，实现 Web & Desktop 的无缝切换。

### C. 数据流转与 Webhooks (QQ/WeChat 联动)

* **Webhooks (Database Webhooks)**:
  * **作用**: 当数据库中的 `tasks` 或 `community_posts` 表有更新时，自动触发外部接口。
  * **示例**: 自动将社区中排名前三的 Prompt 推送到关联的 **微信/QQ 群** 中。

---

## 3. 必须开启的数据库服务 (SQL Editor)

请在 SQL Editor 中运行以下「性能增强」脚本，激活高效的数据同步逻辑：

```sql
-- 1. 开启极致实时复制 (高性能分发)
begin;
  drop publication if exists supabase_realtime;
  create publication supabase_realtime;
commit;
alter publication supabase_realtime add table profiles, announcements, tasks, community_posts;

-- 2. 部署 Edge Functions 触发器 (自动分析拖拽文件)
-- 当 profiles 里的 remote_command 被写入时，自动触发清理或重启
create or replace function handle_remote_action()
returns trigger as $$
begin
  -- 此处可集成外部 Webhook 调用或 Sentry 日志记录
  return new;
end;
$$ language plpgsql;
```

---

## 4. 推荐安装的高级桌面插件 (Extensions for App)

为了实现「推特般」的社交功能和「CLI-Anything」的自动化，建议在应用中加载以下扩展：

1. **SQLite Extension**: 用于本地缓存大型向量数据（如社区缓存），减少频繁请求 Supabase。
2. **Tauri-plugin-shell**: 必须权限，用于 CLI-Anything 执行外部系统命令。
3. **Tauri-plugin-upload**: 用于实现拖拽图片时的分片极速上传到 Supabase Storage。

---

## 5. 项目 `.env` 全量模板 (Full config)

请确保您的 `.env` 文件包含以下字段，以支持所有功能的完美运行：

```env
# Supabase 基础 (必备) — URL 形如 https://<project-ref>.supabase.co
VITE_SUPABASE_URL=https://YOUR_PROJECT_REF.supabase.co
VITE_SUPABASE_ANON_KEY=你的匿名Key

# 社交与控制 (建议)
VITE_GITHUB_CLIENT_ID=用于GitHub登录
VITE_RESEND_API_KEY=用于极速邮件

# 代理潜能 (高级)
VITE_SENTRY_DSN=用于错误追溯
VITE_WECHAT_WEBHOOK=用于微信通知
VITE_QQ_BOT_API=用于QQ任务分发
```

---

## 6. 进阶：如何将代理潜能发挥到极致？

1. **开启 Edge Functions**: 在本地运行 `npx supabase functions deploy research-agent`。
2. **集成 Search API**: 在 Integrations 中接入 **Tavily** 或 **SerpApi**，使您的研究模式具备真实的实时联网搜索能力。
3. **多代理协同 (Hive Mode)**: 在数据库中开启 `profiles` 表的 `settings` 漫游，让分布式运行的代理通过同一个 Supabase 项目同步状态。

---

**⚠️ 注意：** 所有的配置更改后，请在桌面应用中点击侧边栏顶部的「同步」图标，或在集成中心点击「刷新环境检测」以确保配置已生效。
