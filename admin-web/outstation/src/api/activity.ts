import axios from 'axios';
import type { AxiosRequestConfig } from 'axios';

type SilentConfig = AxiosRequestConfig & { silentError?: boolean };

/** 活动类型 */
export const ACTIVITY_TYPE = {
  DEPOSIT: 1,
  /** 洗码 */
  BET: 2,
  REFER: 3,
  LUCK_WHEEL: 4,
  ASSIST: 5,
  BLIND_BOX: 6,
  NEWCOMER: 7,
  /** 全民代理 */
  AGENT: 8,
  /** 救援金 */
  RESCUE: 9,
  LUCKY_BET: 10,
  VIP: 11,
  RED_PACKET: 12,
  /** 图文(广告) */
  IMAGE_AD: 13,
  SIGN_IN: 14,
} as const;

/** 活动状态：1 草稿，2 进行中，3 已结束；仅草稿可发布 */
export const ACTIVITY_STATUS = {
  DRAFT: 1,
  RUNNING: 2,
  ENDED: 3,
} as const;

/** 参与模式：1 自动，2 手动 */
export const ACTIVITY_APPLY_MODE = {
  AUTO: 1,
  MANUAL: 2,
} as const;

/** 活动所属：1 大厅，2 电子，3 体育 */
export const ACTIVITY_SECTION = {
  LOBBY: 1,
  ELECTRONIC: 2,
  SPORTS: 3,
} as const;

/** 循环模式 */
export const ACTIVITY_CYCLE_MODE = {
  NONE: 1,
  DAILY: 2,
  WEEKLY: 3,
  MONTHLY: 4,
  INFINITE: 5,
} as const;

/** 派发模式 */
export const ACTIVITY_PAYOUT_MODE = {
  /** 手动领取，过期作废 */
  MANUAL_EXPIRE_VOID: 1,
  /** 手动领取，过期自动领取 */
  MANUAL_EXPIRE_AUTO: 2,
  AUTO: 3,
} as const;

/** 派发时机：1 实时，2 次日 */
export const ACTIVITY_PAYOUT_TIMING = {
  REALTIME: 1,
  NEXT_DAY: 2,
} as const;

/** 奖金模式 */
export const ACTIVITY_BONUS_MODE = {
  FIXED_MULTI: 1,
  RATE_MULTI: 2,
  FIXED_LADDER_MAX: 3,
  RATE_LADDER_MAX: 4,
  SINGLE: 5,
} as const;

/** 投注奖励领取模式 */
export const ACTIVITY_BET_CLAIM_MODE = {
  NONE: 0,
  EACH: 1,
  HIGHEST: 2,
  CUMULATIVE: 3,
} as const;

/** 充值奖励类型，单选 */
export const ACTIVITY_DEPOSIT_REWARD = {
  FIRST: 1,
  SECOND: 2,
  CUMULATIVE: 3,
  SINGLE: 4,
} as const;

/** 幸运转盘触发方式 */
export const ACTIVITY_WHEEL_TRIGGER = {
  BET: 1,
  DEPOSIT: 2,
  BET_OR_DEPOSIT: 3,
} as const;

/** 洗码基数：1 本金+赠送，2 仅赠送 */
export const ACTIVITY_WAGERING_BASE = {
  PRINCIPAL_AND_BONUS: 1,
  BONUS_ONLY: 2,
} as const;

/** allowed_clients 元素 */
export const ACTIVITY_CLIENT = {
  PC_WEB: 1,
  MOBILE_WEB: 2,
  ANDROID: 3,
  IOS: 4,
} as const;

/** apply_limit 元素 */
export const ACTIVITY_APPLY_LIMIT = {
  DEVICE: 1,
  IP: 2,
} as const;

/** claim_limit 元素：完成对应操作后才可领取 */
export const ACTIVITY_CLAIM_LIMIT = {
  DEPOSIT: 1,
  BANK_CARD: 2,
  SMS: 3,
  EMAIL: 4,
  REAL_NAME: 5,
} as const;

/** 金额字段均为结算币种最小单位 */
export interface ActivityRewardSchema {
  id?: number;
  /** 达标阈值，按活动类型解释为金额或次数 */
  threshold?: number;
  amount?: number;
  rate?: number;
  max_bonus?: number;
  min_bonus?: number;
  count?: number;
}

export interface ActivityReferConfig {
  min_deposit_amount?: number;
  min_bet_amount?: number;
  ip_limit?: number;
  device_limit?: number;
}

export interface ActivityLuckWheelConfig {
  wheel_trigger_mode?: number;
  bet_schemas?: ActivityRewardSchema[];
  deposit_schemas?: ActivityRewardSchema[];
}

export interface ActivitySettings {
  deposit_reward_type?: number;
  bet_claim_mode?: number;
  bonus_mode?: number;
  reward_schemas?: ActivityRewardSchema[];
  refer_config?: ActivityReferConfig;
  /** type 为幸运转盘时必填 */
  luck_wheel?: ActivityLuckWheelConfig;
}

export interface ActivityI18nItem {
  /** 如 zh-CN、zh-HK、en-US */
  lang: string;
  name?: string;
  rule?: string;
  /** 活动说明/副标题，最长 1024 个字符 */
  description?: string;
  /** 该语言的活动图片地址 */
  image?: string;
}

/** 时间字段为 Unix 秒，不是毫秒 */
export interface CreateActivityParams {
  type: number;
  /** 不传默认草稿 */
  status?: number;
  apply_mode: number;
  /** 活动所属：1 大厅，2 电子，3 体育 */
  section: number;
  cycle_mode?: number;
  allowed_fixed_level?: number[];
  allowed_auto_level?: number[];
  allowed_clients?: number[];
  allowed_game_ids?: number[];
  allowed_agent_ids?: number[];
  apply_limit?: number[];
  claim_limit?: number[];
  is_permanent?: boolean;
  open_at_sec?: number;
  close_at_sec?: number;
  release_at_sec?: number;
  auto_audit?: boolean;
  auto_audit_max_amt?: number;
  payout_mode?: number;
  payout_timing?: number;
  claim_expire_sec?: number;
  settings?: ActivitySettings;
  wagering_multiplier?: number;
  wagering_base?: number;
  show_in_list?: boolean;
  show_in_sidebar?: boolean;
  /** 数值越大越靠前 */
  sort_weight?: number;
  i18n?: ActivityI18nItem[];
}

/** 状态不可改，编辑后仍为草稿 */
export interface UpdateActivityParams
  extends Omit<CreateActivityParams, 'status'> {
  id: number;
}

export interface ActivityDetail
  extends Required<
    Omit<
      CreateActivityParams,
      'open_at_sec' | 'close_at_sec' | 'release_at_sec' | 'settings'
    >
  > {
  id: number;
  open_at_sec: number | null;
  close_at_sec: number | null;
  release_at_sec: number | null;
  settings: ActivitySettings | null;
  create_uid: number;
  update_uid: number;
  version: number;
  created_at_sec: number;
  updated_at_sec: number;
}

export interface ActivityListParams {
  /** 从 1 开始 */
  page: number;
  /** 1~100 */
  page_size: number;
  /** 按 lang 对应的 i18n.name 模糊匹配 */
  name?: string;
  type?: number;
  client?: number;
  status?: number;
  /** 同时决定名称检索和列表 name 的语言，默认 zh-CN */
  lang?: string;
}

/** 数组字段无值时接口实际返回 null，不是文档写的 [] */
export interface ActivityListItem {
  id: number;
  type: number;
  status: number;
  /** 按 lang 解析出的名称 */
  name: string;
  /** 按 lang 解析出的图片地址，未上传时为空 */
  image?: string | null;
  allowed_clients: number[] | null;
  open_at_sec: number | null;
  close_at_sec: number | null;
  release_at_sec: number | null;
  sort_weight: number;
  show_in_list: boolean;
  updated_at_sec: number;
  auto_audit: boolean;
  auto_audit_max_amt: number;
  allowed_fixed_level: number[] | null;
  allowed_auto_level: number[] | null;
  wagering_multiplier: number;
  allowed_game_ids: number[] | null;
  is_permanent: boolean;
  created_at_sec: number;
  create_uid: number;
  update_uid: number;
  /** 活动所属：1 大厅，2 电子，3 体育 */
  section: number;
}

export interface ActivityListRes {
  total: number;
  list: ActivityListItem[];
}

/** 创建活动 POST /admin/operator/activity/create；幸运转盘必须带 settings.luck_wheel */
export function createActivity(data: CreateActivityParams) {
  return axios.post<{ id: number }>('/operator/activity/create', data);
}

/** 编辑活动 POST /admin/operator/activity/update；仅草稿可编辑，i18n 按语言整批替换 */
export function updateActivity(data: UpdateActivityParams) {
  return axios.post<{ id: number }>('/operator/activity/update', data);
}

/** 活动详情 GET /admin/operator/activity/detail?id= */
export function getActivityDetail(id: number, silentError = false) {
  return axios.get<ActivityDetail>('/operator/activity/detail', {
    params: { id },
    silentError,
  } as SilentConfig);
}

/** 活动列表 GET /admin/operator/activity/list */
export function queryActivityList(params: ActivityListParams) {
  return axios.get<ActivityListRes>('/operator/activity/list', { params });
}

/** 关闭活动 POST /admin/operator/activity/close，状态置为已结束 */
export function closeActivity(id: number) {
  return axios.post<{ id: number }>('/operator/activity/close', { id });
}

/** 发布活动 POST /admin/operator/activity/publish；仅草稿可发布，成功后为进行中 */
export function publishActivity(id: number) {
  return axios.post<{ id: number }>('/operator/activity/publish', { id });
}

export interface ActivityEnumOption {
  /** 文案的翻译 key，如 activity.type.1，由分组名和 value 拼成 */
  key: string;
  value: number;
  /** 按请求头 X-Lang 翻译 */
  label: string;
}

/** 各组枚举选项 */
export interface ActivityEnumGroups {
  type: ActivityEnumOption[];
  status: ActivityEnumOption[];
  apply_mode: ActivityEnumOption[];
  section: ActivityEnumOption[];
  cycle_mode: ActivityEnumOption[];
  payout_mode: ActivityEnumOption[];
  payout_timing: ActivityEnumOption[];
  bonus_mode: ActivityEnumOption[];
  bet_claim_mode: ActivityEnumOption[];
  deposit_reward_type: ActivityEnumOption[];
  wheel_trigger_mode: ActivityEnumOption[];
  wagering_base: ActivityEnumOption[];
  client_type: ActivityEnumOption[];
  apply_limit_type: ActivityEnumOption[];
  claim_limit_type: ActivityEnumOption[];
}

export interface ActivityEnumParams extends ActivityEnumGroups {
  /** 本站结算币种代码，如 JPY */
  currency: string;
  /** 本站时区，IANA 名称，如 Asia/Tokyo */
  timezone: string;
}

/** 活动枚举参数 GET /admin/operator/activity/enum-params，供表单下拉选项使用 */
export function getActivityEnumParams(silentError = false) {
  return axios.get<ActivityEnumParams>('/operator/activity/enum-params', {
    silentError,
  } as SilentConfig);
}

/** 订单状态：1 待审核，2 审核通过，3 审核拒绝，4 待领取，5 已领取，6 已过期，7 待派发，8 已派发 */
export const ACTIVITY_ORDER_STATUS = {
  PENDING: 1,
  APPROVED: 2,
  REJECTED: 3,
  TO_CLAIM: 4,
  CLAIMED: 5,
  EXPIRED: 6,
  TO_PAYOUT: 7,
  PAID: 8,
} as const;

/** 审核列表的审核类型：1 人工审核，2 风控拦截 */
export const ACTIVITY_AUDIT_STATUS = {
  MANUAL: 1,
  RISK: 2,
} as const;

/** 订单时间筛选字段：1 领取时间，2 申请时间 */
export const ACTIVITY_TIME_FIELD = {
  PAYOUT: 1,
  APPLIED: 2,
} as const;

/** 订单审核动作：1 通过，2 拒绝 */
export const ACTIVITY_AUDIT_ACTION = {
  APPROVE: 1,
  REJECT: 2,
} as const;

/** 订单列表和审核列表的公共筛选；传了起止时间时 time_field 必须是 1 或 2 */
export interface ActivityOrderFilter {
  /** 从 1 开始 */
  page: number;
  /** 1~100 */
  page_size: number;
  /** 精确匹配 */
  order_no?: string;
  activity_id?: number;
  member_id?: number;
  type?: number;
  time_field?: number;
  /** Unix 秒，含本秒 */
  start_sec?: number;
  end_sec?: number;
}

export interface ActivityOrderListParams extends ActivityOrderFilter {
  status?: number;
}

export interface ActivityAuditListParams extends ActivityOrderFilter {
  /** 不传不限 */
  audit_status?: number;
}

/** 活动名称随 X-Lang，没有对应语言时为空；列表不下发订单自增 ID */
export interface ActivityOrderItem {
  order_no: string;
  activity_id: number;
  activity_name: string;
  type: number;
  member_id: number;
  /** 赠送金额，最小单位 */
  rewards_total: number;
  /** 渠道来源，同客户端类型 */
  channel_source: number;
  payout_mode: number;
  status: number;
  /** 以下时间为 Unix 秒，没有时不返回 */
  claim_expire_at_sec?: number;
  payout_at_sec?: number;
  applied_at_sec?: number;
  approved_at_sec?: number;
  /** 审批人后台用户 ID，没有为 0 */
  approver_uid?: number;
}

export interface ActivityOrderListRes {
  total: number;
  list: ActivityOrderItem[];
}

export interface AuditActivityOrderParams {
  order_no: string;
  action: number;
  /** 为空时后端按动作使用默认备注 */
  remark?: string;
}

/** 活动订单列表 GET /admin/operator/activity/order/list */
export function queryActivityOrderList(params: ActivityOrderListParams) {
  return axios.get<ActivityOrderListRes>('/operator/activity/order/list', {
    params,
  });
}

/** 活动订单审核列表 GET /admin/operator/activity/order/audit/list，只返回待审核订单 */
export function queryActivityAuditList(
  params: ActivityAuditListParams,
  silentError = false
) {
  return axios.get<ActivityOrderListRes>(
    '/operator/activity/order/audit/list',
    { params, silentError } as SilentConfig
  );
}

/** 活动订单审核 POST /admin/operator/activity/order/audit；仅待审核订单可审，引擎未启动时不能通过 */
export function auditActivityOrder(data: AuditActivityOrderParams) {
  return axios.post<{ order_no: string }>(
    '/operator/activity/order/audit',
    data
  );
}
