import axios from 'axios';

export const TASK_CATEGORIES = ['ONCE', 'DAILY', 'WEEKLY', 'MONTHLY'] as const;
export type TaskCategory = (typeof TASK_CATEGORIES)[number];

/** 任务状态：1 启用，2 禁用 */
export const TASK_STATUS = {
  ENABLED: 1,
  DISABLED: 2,
} as const;

export const TASK_PAYOUT_MODES = ['MANUAL', 'AUTO'] as const;
export type TaskPayoutMode = (typeof TASK_PAYOUT_MODES)[number];

/** 奖励订单状态 */
export const TASK_REWARD_STATUS = {
  /** 待领取 */
  PENDING_CLAIM: 1,
  /** 待派发 */
  PENDING_PAYOUT: 2,
  /** 任务已拒绝 */
  TASK_REJECTED: 11,
  /** 派发被拒绝 */
  PAYOUT_REJECTED: 12,
  /** 已派发 */
  PAID: 99,
} as const;

/** 金额为最小货币单位，时间为 Unix 秒 */
export interface TaskConfigItem {
  id: number;
  operator_code: string;
  task_code: string;
  category: TaskCategory;
  name: string;
  /** 国际化 JSON 字符串 */
  name_i18n: string;
  description: string;
  description_i18n: string;
  /** IANA 时区，如 Asia/Shanghai */
  timezone: string;
  status: number;
  /** 完成条件 USD 金额 */
  completion_amount: number;
  reward_amount: number;
  payout_mode: TaskPayoutMode;
  /** 升序 */
  sort: number;
  created_at: number;
  updated_at: number;
}

export interface TaskConfigListParams {
  /** 从 1 开始 */
  page: number;
  /** 最大 100 */
  page_size: number;
  category?: TaskCategory;
  status?: number;
}

export interface TaskConfigListRes {
  total: number;
  page: number;
  page_size: number;
  list: TaskConfigItem[];
}

/** 只允许修改以下字段，未传的字段保持不变 */
export interface UpdateTaskConfigParams {
  id: number;
  timezone?: string;
  status?: number;
  completion_amount?: number;
  reward_amount?: number;
  payout_mode?: TaskPayoutMode;
  sort?: number;
}

export interface TaskRewardItem {
  id: number;
  operator_code: string;
  member_id: number;
  task_code: string;
  kafka_id: string;
  category: TaskCategory;
  /** ALL / 日期 / 周 / 月 */
  period_key: string;
  payout_mode: TaskPayoutMode;
  amount: number;
  status: number;
  order_no: string;
  wallet_transaction_id: string;
  device_id: string;
  wallet_id: string;
  claim_ip: string;
  /** 未派发时为 0 */
  payout_at: number;
  created_at: number;
  updated_at: number;
}

export interface TaskRewardListParams {
  page: number;
  page_size: number;
  member_id?: number;
  task_code?: string;
  category?: TaskCategory;
  status?: number;
  order_no?: string;
  created_at_start?: number;
  created_at_end?: number;
}

export interface TaskRewardListRes {
  total: number;
  page: number;
  page_size: number;
  list: TaskRewardItem[];
}

/** 任务配置列表 GET /admin/operator/task/config/list，按 sort 升序 */
export function queryTaskConfigList(params: TaskConfigListParams) {
  return axios.get<TaskConfigListRes>('/operator/task/config/list', {
    params,
  });
}

/** 更新任务配置 POST /admin/operator/task/config/update */
export function updateTaskConfig(data: UpdateTaskConfigParams) {
  return axios.post<{ task_config: TaskConfigItem }>(
    '/operator/task/config/update',
    data
  );
}

/** 会员任务奖励订单 GET /admin/operator/task/reward/list */
export function queryTaskRewardList(params: TaskRewardListParams) {
  return axios.get<TaskRewardListRes>('/operator/task/reward/list', {
    params,
  });
}

/** 只允许将订单状态从 1（待领取）改为 2（待派发） */
export interface UpdateTaskRewardStatusParams {
  operator_code: string;
  id: number;
  status: typeof TASK_REWARD_STATUS.PENDING_PAYOUT;
}

/** 更新会员任务奖励状态 POST /admin/operator/task/reward/update-status */
export function updateTaskRewardStatus(data: UpdateTaskRewardStatusParams) {
  return axios.post<{ member_task_reward: TaskRewardItem }>(
    '/operator/task/reward/update-status',
    data
  );
}
