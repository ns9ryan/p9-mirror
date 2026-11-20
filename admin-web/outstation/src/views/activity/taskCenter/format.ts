import type { TaskCategory, TaskPayoutMode } from '@/api/task';
import { TASK_REWARD_STATUS } from '@/api/task';
export { formatTime } from '@/utils/site-time';

export const CATEGORY_COLORS: Record<TaskCategory, string> = {
  ONCE: 'arcoblue',
  DAILY: 'green',
  WEEKLY: 'orangered',
  MONTHLY: 'purple',
};

/** 手动领取偏向需要用户操作，用橙色提示；自动派发用绿色 */
export const PAYOUT_MODE_COLORS: Record<TaskPayoutMode, string> = {
  MANUAL: 'orange',
  AUTO: 'green',
};

export const REWARD_STATUS_COLORS: Record<number, string> = {
  [TASK_REWARD_STATUS.PENDING_CLAIM]: 'arcoblue',
  [TASK_REWARD_STATUS.PENDING_PAYOUT]: 'orange',
  [TASK_REWARD_STATUS.TASK_REJECTED]: 'red',
  [TASK_REWARD_STATUS.PAYOUT_REJECTED]: 'red',
  [TASK_REWARD_STATUS.PAID]: 'green',
};

/** 金额为最小货币单位，币种精度未提供，按原值展示 */
export function formatAmount(value?: number | null) {
  if (value == null) return '--';
  return value.toLocaleString('en-US');
}

export function displayText(value?: string | null) {
  return value?.trim() || '--';
}
