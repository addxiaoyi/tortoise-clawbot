-- 供 Go BFF /api/me 读取的 profiles 扩展列（与 Tortoise Supabase 表对齐）。
-- 在已存在 public.profiles 的库中执行一次即可；列已存在时跳过。

ALTER TABLE public.profiles ADD COLUMN IF NOT EXISTS roles text[] DEFAULT ARRAY[]::text[];
ALTER TABLE public.profiles ADD COLUMN IF NOT EXISTS permissions text[] DEFAULT ARRAY[]::text[];
ALTER TABLE public.profiles ADD COLUMN IF NOT EXISTS is_banned boolean DEFAULT false;
ALTER TABLE public.profiles ADD COLUMN IF NOT EXISTS rewards bigint DEFAULT 0;

-- Tortoise 侧栏等使用单列 role（user/admin/moderator）；若尚无该列可打开下一行。
ALTER TABLE public.profiles ADD COLUMN IF NOT EXISTS role text DEFAULT 'user';
