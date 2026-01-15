# Nanobot (Tortoise) Perfection & Configuration Plan

Based on the `SUPABASE_INTEGRATION_GUIDE.md` and project requirements, I am finalizing the configuration for full functionality.

## ✅ Completed Tasks
- **Environment Configuration**: Created `.env` and root `.env` with Supabase, GitHub, Resend, and Sentry placeholders.
- **Tauri Plugin Expansion (Native)**: Added `tauri-plugin-sql` and `tauri-plugin-upload` to `Cargo.toml`.
- **Rust Integration**: Initialized new plugins in `src-tauri/src/lib.rs`.
- **Capability Management**: Added permissions for `sql`, `upload`, and `dialog` in `capabilities/default.json`.
- **Error Tracking**: Integrated `@sentry/react` in `src/main.tsx` for production error tracking.

## 🛠️ In-Progress Tasks
- **Dependency Syncing**: `npm install` is running to install Sentry and Tauri plugins.
- **Plugin Configuration**: Finalizing `tauri.conf.json` for all plugins.

## 📋 Outstanding Requirements (User Actions Needed)
The following steps require your direct interaction with the Supabase/Third-party dashboards:

### 1. Database Setup (Supabase SQL Editor)
Execute the following SQL in your Supabase dashboard to enable Realtime and Profile features:
```sql
-- 启用 Realtime
alter publication supabase_realtime add table profiles;

-- 极速邮件通知触发器
create extension if not exists http with schema extensions;
create or replace function public.notify_on_new_login()
returns trigger as $$
begin
  perform
    net.http_post(
      url:='https://api.resend.com/emails',
      headers:='{"Authorization": "Bearer YOUR_RESEND_KEY"}'::jsonb,
      body:=json_build_object('from', 'tortoise@yourdomain.com', 'to', new.email, 'subject', 'New Login Detection')::jsonb
    );
  return new;
end;
$$ language plpgsql security definer;
```

### 2. Environment Variables
Populate the following in `.env`:
- `VITE_SUPABASE_ANON_KEY`: From Supabase Project Settings -> API.
- `VITE_GITHUB_CLIENT_ID`: Create an OAuth App in GitHub Developer Settings.
- `VITE_SENTRY_DSN`: Create a project in Sentry (React).

### 3. Deep Link Configuration
Ensure the `Redirect URL` in Supabase Auth Settings is set to `tortoise://`.

---
## 🚀 Next Steps for Me
1. Verify `npm install` completion.
2. Implement additional UI diagnostic tools in `DiagnosisPage` if requested.
3. Optimize the `RemoteCommandManager` for better reliability.
