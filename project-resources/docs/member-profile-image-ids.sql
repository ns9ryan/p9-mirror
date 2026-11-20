BEGIN;

ALTER TABLE public.member_profile
  ADD COLUMN IF NOT EXISTS avatar_id bigint NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS avatar_frame_id bigint NOT NULL DEFAULT 0;

ALTER TABLE public.member_profile
  DROP COLUMN IF EXISTS avatar_url,
  DROP COLUMN IF EXISTS avatar_frame_url;

COMMENT ON COLUMN public.member_profile.avatar_id IS '总站头像ID，0表示未设置';
COMMENT ON COLUMN public.member_profile.avatar_frame_id IS '总站头像框ID，0表示未设置';

COMMIT;