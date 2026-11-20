-- CRM 历史 VIP 图标数据兼容迁移。
-- 在 192.168.10.220:15441 / p9_operator 执行，先停止 CRM RPC。
-- 使用与 init/member_vip.go 相同的默认图标，保留运营自定义 icon_url。
BEGIN;
DO $$
BEGIN
  IF current_database() <> 'p9_operator' THEN
    RAISE EXCEPTION 'Wrong database: %, expected p9_operator', current_database();
  END IF;
END $$;

-- 兼容自动迁移已回滚、新字段尚未创建的情况。
ALTER TABLE public.member_vip_icon_group
  ADD COLUMN IF NOT EXISTS default_icon_url text;

WITH defaults(group_no, icon_url) AS (
  VALUES
    (0, 'https://file.98ent.com/avatar/5afb77abe42d79378bb6be95ffec1a3d.png'),
    (1, 'https://file.98ent.com/avatar/c8bf063626f6d8e45751ce866590f97f.png'),
    (2, 'https://file.98ent.com/avatar/bb81b3fddfa4d802b9d4e675bade638e.png'),
    (3, 'https://file.98ent.com/avatar/5c6e7ff8673b6b5d30e67615c90b7527.png'),
    (4, 'https://file.98ent.com/avatar/f530bae5a30150d3f0dbeb82b4aec9fe.png'),
    (5, 'https://file.98ent.com/avatar/22eeec8f2579b1f59cfee516f3233914.png'),
    (6, 'https://file.98ent.com/avatar/e5e51f03020faa2f0ec273d2b49a90b1.png'),
    (7, 'https://file.98ent.com/avatar/c7a81a2ea4303262f88fcd833d45f9dc.png'),
    (8, 'https://file.98ent.com/avatar/dc20dd718a7516d3c8b54a6cf3f4f913.png'),
    (9, 'https://file.98ent.com/avatar/53fbe7099751f1f3bbce70689ce619d3.png')
)
UPDATE public.member_vip_icon_group AS g
SET default_icon_url = d.icon_url
FROM defaults AS d
WHERE g.group_no = d.group_no
  AND (g.default_icon_url IS NULL OR btrim(g.default_icon_url) = '');

-- 未识别的历史分组需要人工核对，不用空字符串掩盖问题。
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM public.member_vip_icon_group
    WHERE default_icon_url IS NULL OR btrim(default_icon_url) = ''
  ) THEN
    RAISE EXCEPTION 'Unmapped VIP group exists; inspect group_no before retrying';
  END IF;
END $$;

ALTER TABLE public.member_vip_icon_group
  ALTER COLUMN default_icon_url SET NOT NULL;
COMMIT;

-- 应返回 0。
SELECT count(*) AS missing_default_icons
FROM public.member_vip_icon_group
WHERE default_icon_url IS NULL OR btrim(default_icon_url) = '';
