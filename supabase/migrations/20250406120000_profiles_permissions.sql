-- profiles.permissions：功能门禁（与 Tortoise tortoise.page.* / tortoise.bundle.* 对齐）
-- 扩展 admin_profile_set；非管理员不可自改 permissions（与 role / is_banned 一致）

begin;

alter table public.profiles
  add column if not exists permissions text[] not null default '{}';

-- 非管理员更新自己行时锁定 role / is_banned / permissions
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
  new.permissions := coalesce(old.permissions, '{}');
  return new;
end;
$$;

-- 审计：敏感字段含 permissions
create or replace function public.profiles_audit_sensitive_changes()
returns trigger
language plpgsql
security definer
set search_path = public
as $$
begin
  if new.role is distinct from old.role
     or new.is_banned is distinct from old.is_banned
     or coalesce(new.permissions, '{}') is distinct from coalesce(old.permissions, '{}') then
    insert into public.audit_profile_actions (actor_id, target_id, action, old_value, new_value)
    values (
      auth.uid(),
      new.id,
      'profiles_sensitive_update',
      jsonb_build_object(
        'role', old.role,
        'is_banned', old.is_banned,
        'permissions', to_jsonb(coalesce(old.permissions, '{}'))
      ),
      jsonb_build_object(
        'role', new.role,
        'is_banned', new.is_banned,
        'permissions', to_jsonb(coalesce(new.permissions, '{}'))
      )
    );
  end if;
  return null;
end;
$$;

drop function if exists public.admin_profile_set(uuid, text, boolean);

create or replace function public.admin_profile_set(
  p_target_id uuid,
  p_role text default null,
  p_is_banned boolean default null,
  p_permissions text[] default null
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
    permissions = case when p_permissions is null then p.permissions else p_permissions end,
    updated_at = now()
  where p.id = p_target_id;

  return json_build_object('ok', true);
end;
$$;

grant execute on function public.admin_profile_set(uuid, text, boolean, text[]) to authenticated;

commit;
