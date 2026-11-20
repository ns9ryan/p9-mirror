import type { ActivityListItem } from '@/api/activity';

export type ActivityLayout = 'classic' | 'v2';

export type ActivityStatus = 'draft' | 'running' | 'ended';

/** 接口没有删除；编辑和发布仅限草稿 */
export type ActivityAction = 'detail' | 'edit' | 'publish' | 'close';

export type ActivityRecord = ActivityListItem;

export interface ActivityQuery {
  /** 活动名称，按当前语言模糊匹配 */
  name?: string;
  /** 传了 ID 时只按详情接口精确查找，忽略其它条件 */
  id?: number;
  type?: number;
  client?: number;
  status?: number;
}
