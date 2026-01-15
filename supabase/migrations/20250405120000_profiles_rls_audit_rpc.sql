-- Profiles RLS、审计表、管理员 RPC（Tortoise / Supabase）
-- 在已有 public.profiles（id uuid 主键，通常对齐 auth.users）上执行。
-- 首次将某用户设为管理员请在 SQL Editor 中手动 UPDATE，或通过 Dashboard。

begin;

-- Tortoise 列表依赖 updated_at；若尚无则补充
alter table public.profiles add column if not exists updated_at timestamptz default now();

-- ---------------------------------------------------------------------------
-- 审计表：记录 role / is_banned 变更（含 Dashboard 直接改表）
-- ---------------------------------------------------------------------------
create table if not exists public.audit_profile_actions (
  id uuid primary key default gen_random_uuid(),
  actor_id uuid references auth.users (id) on delete set null,
  target_id uuid not null,
  action text not null,
  old_value jsonb,
  new_value jsonb,
  created_at timestamptz not null default now()
);

create index if not exists audit_profile_actions_target_idx
  on public.audit_profile_actions (target_id, created_at desc);
create index if not exists audit_profile_actions_actor_time_idx
  on public.audit_profile_actions (actor_id, created_at desc);

alter table public.audit_profile_actions enable row level security;

-- 仅管理员可读审计（应用侧也可用 service role 绕过）
create or replace function public.is_profile_admin(check_id uuid)
returns boolean
language sql
stable
security definer
set search_path = public
as $$
  select exists (
    select 1 from public.profiles p
    where p.id = check_id and p.role = 'admin'
  );
$$;

grant execute on function public.is_profile_admin(uuid) to authenticated;

-- 非管理员不得改他人行的敏感列：在 BEFORE UPDATE 中锁定 role / is_banned
create or replace function public.profiles_lock_sensitive_for_non_admin()
returns trigger
language plpgsql
security definer
set search_path = public
as $$
declare
  adm boolean;
begin
  select public.is_profile_admin(auth.uid()) into adm;
  if coalesce(adm, false) then
    return new;
  end if;
  if new.id is distinct from auth.uid() then
    raise exception 'forbidden' using errcode = '42501';
  end if;
  new.role := old.role;
  new.is_banned := old.is_banned;
  return new;
end;
$$;

drop trigger if exists profiles_lock_sensitive_for_non_admin on public.profiles;
create trigger profiles_lock_sensitive_for_non_admin
  before update on public.profiles
  for each row
  execute function public.profiles_lock_sensitive_for_non_admin();

-- 变更敏感字段后写审计
create or replace function public.profiles_audit_sensitive_changes()
returns trigger
language plpgsql
security definer
set search_path = public
as $$
begin
  if new.role is distinct from old.role or new.is_banned is distinct from old.is_banned then
    insert into public.audit_profile_actions (actor_id, target_id, action, old_value, new_value)
    values (
      auth.uid(),
      new.id,
      'profiles_sensitive_update',
      jsonb_build_object('role', old.role, 'is_banned', old.is_banned),
      jsonb_build_object('role', new.role, 'is_banned', new.is_banned)
    );
  end if;
  return null;
end;
$$;

drop trigger if exists profiles_audit_sensitive_changes on public.profiles;
create trigger profiles_audit_sensitive_changes
  after update on public.profiles
  for each row
  execute function public.profiles_audit_sensitive_changes();

-- ---------------------------------------------------------------------------
-- RLS：profiles
-- ---------------------------------------------------------------------------
alter table public.profiles enable row level security;

drop policy if exists profiles_select_own_or_admin on public.profiles;
create policy profiles_select_own_or_admin
  on public.profiles
  for select
  to authenticated
  using (
    auth.uid() = id
    or public.is_profile_admin(auth.uid())
  );

drop policy if exists profiles_insert_own on public.profiles;
create policy profiles_insert_own
  on public.profiles
  for insert
  to authenticated
  with check (auth.uid() = id);

drop policy if exists profiles_update_own_or_admin on public.profiles;
create policy profiles_update_own_or_admin
  on public.profiles
  for update
  to authenticated
  using (
    auth.uid() = id
    or public.is_profile_admin(auth.uid())
  )
  with check (
    auth.uid() = id
    or public.is_profile_admin(auth.uid())
  );

-- 审计表：仅管理员可读
drop policy if exists audit_select_admin on public.audit_profile_actions;
create policy audit_select_admin
  on public.audit_profile_actions
  for select
  to authenticated
  using (public.is_profile_admin(auth.uid()));

-- ---------------------------------------------------------------------------
-- 管理员 RPC：速率限制 + 统一更新入口（前端应优先调用此函数）
-- ---------------------------------------------------------------------------
create or replace function public.admin_profile_set(
  p_target_id uuid,
  p_role text default null,
  p_is_banned boolean default null
)
returns json
language plpgsql
security definer
set search_path = public
as $$
declare
  v_actor uuid := auth.uid();
  v_cnt int;
begin
  if v_actor is null then
    raise exception 'unauthorized' using errcode = 'P0001';
  end if;
  if not public.is_profile_admin(v_actor) then
    raise exception 'forbidden' using errcode = 'P0001';
  end if;

  select count(*)::int into v_cnt
  from public.audit_profile_actions
  where actor_id = v_actor
    and created_at > now() - interval '1 minute';

  if v_cnt >= 60 then
    raise exception 'rate_limit' using errcode = 'P0001';
  end if;

  if not exists (select 1 from public.profiles where id = p_target_id) then
    raise exception 'not_found' using errcode = 'P0001';
  end if;

  update public.profiles p
  set
    role = coalesce(p_role, p.role),
    is_banned = coalesce(p_is_banned, p.is_banned),
    updated_at = now()
  where p.id = p_target_id;

  return json_build_object('ok', true);
end;
$$;

grant execute on function public.admin_profile_set(uuid, text, boolean) to authenticated;

commit;
