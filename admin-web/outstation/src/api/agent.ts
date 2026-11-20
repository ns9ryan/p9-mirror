import axios from 'axios';
import type { AxiosRequestConfig } from 'axios';

/**
 * 状态。文档只说明筛选传 0 表示不限，示例值为 1。
 * 按本项目惯例：1 正常，2 停用。
 */
export const AGENT_STATUS = {
  NORMAL: 1,
  DISABLED: 2,
} as const;

/**
 * 出款限制。文档未给枚举，示例值为 1，筛选传 0 表示不限。
 * 先按 1 不限制、2 限制出款。
 */
export const AGENT_PAYOUT = {
  NONE: 1,
  FORBIDDEN: 2,
} as const;

/** 1 本人，2 上级 */
export const AGENT_FILTER = {
  SELF: 1,
  PARENT: 2,
} as const;

/** 文档示例为 1，未列出其他取值 */
export const AGENT_TYPE = {
  DEFAULT: 1,
} as const;

export interface AgentListParams {
  /** 页码，从 1 */
  page?: number;
  /** 每页条数，默认 20，上限 100 */
  page_size?: number;
  /** 注册来源，0 不限 */
  register_source?: number;
  /** 层级，0 不限 */
  level?: number;
  /** 状态，0 不限 */
  status?: number;
  /** 出款限制，0 不限 */
  payout_restricted?: number;
  /** 1 注册时间，2 最后登录，0 不筛 */
  time_filter_type?: number;
  /** Unix 秒，需与 time_end 成对 */
  time_start?: number;
  time_end?: number;
  /**
   * 1 直属会员，2 直属代理，3 全部下级代理，4 全部会员，
   * 5 占成，6 累计佣金，7 已领取佣金。0 不筛
   */
  number_filter_type?: number;
  number_min?: number;
  number_max?: number;
  /** 1 本人，2 上级，0 不筛 */
  agent_filter_type?: number;
  /** 纯数字按 ID，否则账号模糊 */
  agent_keyword?: string;
}

export interface AgentListItem {
  id: number;
  account: string;
  level: number;
  level_alias: string;
  parent_id: number;
  parent_account: string;
  top_id: number;
  top_account: string;
  register_source: number;
  direct_member_count: number;
  direct_agent_count: number;
  total_member_count: number;
  total_agent_count: number;
  /** 百分比 0~100 */
  share_rate: number;
  wallet_balance: number;
  total_commission: number;
  unpaid_commission: number;
  status: number;
  payout_restricted: number;
  remark: string;
  promo_link: string;
  visit_uv: number;
  /** Unix 秒 */
  register_at: number;
  /** Unix 秒，0 表示没有 */
  last_login_at: number;
}

export interface AgentListRes {
  list: AgentListItem[];
  total: number;
}

/** 分页代理列表 POST /agent/list */
export function listAgents(data: AgentListParams, silentError = false) {
  return axios.post<AgentListRes>('/agent/list', data, {
    silentError,
  } as AxiosRequestConfig);
}

/**
 * 代理申请审核状态，筛选传 0 表示不限。
 * 文档只给出 1 待审核、通过后为 2，拒绝的取值按审批 action 顺延推断。
 */
export const AGENT_APPLY_AUDIT = {
  PENDING: 1,
  APPROVED: 2,
  REJECTED: 3,
  RISK_REJECTED: 4,
} as const;

/** 审批动作：1 通过，2 人工拒绝，3 风控拒绝 */
export const AGENT_APPLY_ACTION = {
  APPROVE: 1,
  REJECT: 2,
  RISK_REJECT: 3,
} as const;

/** 代理申请的账号类型：1 代理账号，2 会员账号 */
export const AGENT_APPLY_ACCOUNT_TYPE = {
  AGENT: 1,
  MEMBER: 2,
} as const;

export interface AgentApplyListParams {
  /** 页码，从 1 */
  page?: number;
  /** 每页条数，默认 20，上限 100 */
  page_size?: number;
  /** 0 不限 */
  audit_status?: number;
  /** 1 代理账号，2 会员账号，0 不限类型 */
  account_type?: number;
  /** 账号模糊 */
  account?: string;
  /** 申请时间 Unix 秒，0 不限 */
  time_start?: number;
  time_end?: number;
}

export interface AgentApplyItem {
  id: number;
  account: string;
  realname: string;
  mobile: string;
  register_source: number;
  /** 百分比 0~100 */
  apply_share_rate: number;
  member_account: string;
  member_code: string;
  audit_status: number;
  reject_reason: string;
  apply_remark: string;
  /** Unix 秒 */
  register_at: number;
  /** Unix 秒，0 表示未审核 */
  audited_at: number;
  auditor: string;
  agent_id: number;
}

export interface AgentApplyListRes {
  list: AgentApplyItem[];
  total: number;
}

/** 分页查询代理申请 POST /agent/apply/list */
export function listAgentApplies(
  data: AgentApplyListParams,
  silentError = false
) {
  return axios.post<AgentApplyListRes>('/agent/apply/list', data, {
    silentError,
  } as AxiosRequestConfig);
}

export interface ReviewAgentApplyParams {
  /** 申请单 ID */
  id: number;
  action: number;
  /** 拒绝时必填 */
  reject_reason?: string;
}

export interface ReviewAgentApplyRes {
  apply: Pick<
    AgentApplyItem,
    'id' | 'account' | 'audit_status' | 'agent_id' | 'auditor' | 'audited_at'
  >;
  /** 通过时返回的明文登录密码，仅返回这一次 */
  password?: string;
}

/**
 * 审批申请 POST /agent/apply/review，仅待审核可审。
 * 通过时按现金正式线指定代理开户。
 */
export function reviewAgentApply(data: ReviewAgentApplyParams) {
  return axios.post<ReviewAgentApplyRes>('/agent/apply/review', data);
}

/** 资料接口的账号状态：1 正常，2 冻结，3 停用 */
export const AGENT_PROFILE_STATUS = {
  NORMAL: 1,
  FROZEN: 2,
  DISABLED: 3,
} as const;

/** 资料接口的出款限制：1 不限制，2 限制本人，3 限制全线 */
export const AGENT_PROFILE_PAYOUT = {
  NONE: 1,
  SELF: 2,
  LINE: 3,
} as const;

export interface AgentProfile {
  agent_id: number;
  /** 主表备注 */
  remark: string;
  status: number;
  /** 1 PC，2 app，3 管理员，4 代理 */
  register_source: number;
  promo_code: string;
  parent_account: string;
  /** Unix 秒 */
  register_at: number;
  /** Unix 秒，0 表示没有 */
  last_login_at: number;
  last_login_ip: string;
  last_login_udid: string;
  /** 百分比 0~100 */
  share_rate: number;
  login_count: number;
  /** 直属会员数 */
  invite_member_count: number;
  payout_restricted: number;
  /** 1 现金，2 信用 */
  agent_type: number;
  country: string;
  currency: string;
  avatar_url: string;
  realname: string;
  telegram: string;
  facebook: string;
  whatsapp: string;
  mobile: string;
  nickname: string;
  id_document_no: string;
  zalo: string;
  twitter: string;
  line: string;
  email: string;
}

/** 按代理 ID 查询资料 GET /agent/profile */
export function getAgentProfile(agentId: number) {
  return axios.get<AgentProfile>('/agent/profile', {
    params: { agent_id: agentId },
  });
}

/** POST /agent/profile/update 的 field_type，0 表示同时更新下列全部字段 */
export const AGENT_PROFILE_FIELD = {
  ALL: 0,
  AVATAR_URL: 1,
  REALNAME: 2,
  TELEGRAM: 3,
  FACEBOOK: 4,
  WHATSAPP: 5,
  MOBILE: 6,
  NICKNAME: 7,
  ID_DOCUMENT_NO: 8,
  ZALO: 9,
  TWITTER: 10,
  LINE: 11,
  EMAIL: 12,
} as const;

export type AgentProfileFields = Pick<
  AgentProfile,
  | 'avatar_url'
  | 'realname'
  | 'telegram'
  | 'facebook'
  | 'whatsapp'
  | 'mobile'
  | 'nickname'
  | 'id_document_no'
  | 'zalo'
  | 'twitter'
  | 'line'
  | 'email'
>;

export type UpdateAgentProfileParams = {
  agent_id: number;
  field_type: number;
} & Partial<AgentProfileFields>;

/**
 * 按 field_type 更新资料（不含备注）POST /agent/profile/update。
 * 被选中的字段省略或传空字符串表示清空，data 与 GET /agent/profile 相同。
 */
export function updateAgentProfile(data: UpdateAgentProfileParams) {
  return axios.post<AgentProfile>('/agent/profile/update', data);
}

export interface ResetAgentPasswordRes {
  /** 服务端生成的明文密码，仅返回这一次 */
  password: string;
}

/** 重置登录密码（服务端随机生成）POST /agent/password/reset */
export function resetAgentPassword(agentId: number) {
  return axios.post<ResetAgentPasswordRes>('/agent/password/reset', {
    agent_id: agentId,
  });
}

/**
 * 修改占成 POST /agent/share-rate/update，share_rate 为百分比 0~100。
 * 不能超过上级占成和配置上限，也不能低于直属下级里最高的占成，相等允许。
 */
export function updateAgentShareRate(agentId: number, shareRate: number) {
  return axios.post<Record<string, never>>('/agent/share-rate/update', {
    agent_id: agentId,
    share_rate: shareRate,
  });
}

/** 修改出款限制 POST /agent/payout-restricted/update，取值见 AGENT_PROFILE_PAYOUT */
export function updateAgentPayoutRestricted(
  agentId: number,
  payoutRestricted: number
) {
  return axios.post<Record<string, never>>('/agent/payout-restricted/update', {
    agent_id: agentId,
    payout_restricted: payoutRestricted,
  });
}

/** 修改主表备注 POST /agent/remark/update，空字符串表示清空 */
export function updateAgentRemark(agentId: number, remark: string) {
  return axios.post<Record<string, never>>('/agent/remark/update', {
    agent_id: agentId,
    remark,
  });
}

/** 重置提现密码 POST /agent/withdraw-password/reset，传明文，服务端只存哈希 */
export function resetAgentWithdrawPassword(agentId: number, password: string) {
  return axios.post<Record<string, never>>('/agent/withdraw-password/reset', {
    agent_id: agentId,
    password,
  });
}

/** 登录类型：0 不限，1 登录，2 注册 */
export const AGENT_LOGIN_TYPE = {
  ALL: 0,
  LOGIN: 1,
  REGISTER: 2,
} as const;

/** 设备类型：0 不限，1 PC网页，2 H5，3 Android，4 iOS，5 PC客户端 */
export const AGENT_LOGIN_DEVICE = {
  ALL: 0,
  PC: 1,
  H5: 2,
  ANDROID: 3,
  IOS: 4,
  PC_CLIENT: 5,
} as const;

/** 关键字类型：0 不筛，1 代理ID，2 登录IP，3 设备码，4 登录地点 */
export const AGENT_LOGIN_KEYWORD = {
  NONE: 0,
  AGENT_ID: 1,
  IP: 2,
  UDID: 3,
  LOCATION: 4,
} as const;

/** 登录结果：1 成功，其余为失败（文案见 result_label） */
export const AGENT_LOGIN_RESULT_SUCCESS = 1;

export interface AgentLoginLogParams {
  /** 页码，默认 1 */
  page?: number;
  /** 每页条数，默认 20，上限 100 */
  page_size?: number;
  /** 登录时间起止 Unix 秒，0 不限 */
  time_start?: number;
  time_end?: number;
  login_type?: number;
  device_type?: number;
  keyword_type?: number;
  keyword?: string;
}

export interface AgentLoginLogItem {
  id: number;
  agent_id: number;
  account: string;
  login_type: number;
  login_type_label: string;
  login_ip: string;
  login_location: string;
  udid: string;
  login_device: string;
  device_type: number;
  device_type_label: string;
  /** Unix 秒 */
  login_at: number;
  result: number;
  result_label: string;
}

export interface AgentLoginLogRes {
  list: AgentLoginLogItem[];
  total: number;
}

/** 分页查询代理登录日志 POST /agent/login-log/list */
export function listAgentLoginLogs(data: AgentLoginLogParams) {
  return axios.post<AgentLoginLogRes>('/agent/login-log/list', data);
}

/** 操作结果：0 不限，1 成功，2 失败 */
export const AGENT_OP_RESULT = {
  ALL: 0,
  SUCCESS: 1,
  FAIL: 2,
} as const;

/** 操作目标类型：0 不限，1 代理，2 会员 */
export const AGENT_OP_TARGET_TYPE = {
  ALL: 0,
  AGENT: 1,
  MEMBER: 2,
} as const;

/** 操作人类型：1 分站管理员，2 分站代理 */
export const AGENT_OP_ADMIN_TYPE = {
  ADMIN: 1,
  AGENT: 2,
} as const;

export interface AgentOpLogParams {
  /** 操作目标 ID，0 或不传表示不限 */
  target_id?: number;
  /** 取值见 AGENT_OP_TARGET_TYPE */
  target_type?: number;
  /** 页码，默认 1 */
  page?: number;
  /** 每页条数，默认 20，上限 100 */
  page_size?: number;
  /** 操作时间起止 Unix 秒，0 不限 */
  time_start?: number;
  time_end?: number;
  /** 取值见 AGENT_OP_RESULT */
  result?: number;
}

export interface AgentOpLogItem {
  id: number;
  op_target_id: number;
  /** 取值见 AGENT_OP_TARGET_TYPE */
  op_target_type: number;
  op_target_type_label: string;
  op_target_account: string;
  admin_id: number;
  /** 取值见 AGENT_OP_ADMIN_TYPE */
  admin_type: number;
  admin_type_label: string;
  admin_account: string;
  /** 完整 i18n key，如 agent.opAction.share_rate */
  action: string;
  action_label: string;
  /** 密码类为 ***；整份资料更新时为 JSON */
  before_value: string;
  after_value: string;
  client_ip: string;
  client_ip_location: string;
  login_device: string;
  udid: string;
  /** Unix 秒 */
  operated_at: number;
  result: number;
  result_label: string;
  /** 已按当前语言翻译 */
  fail_reason: string;
}

export interface AgentOpLogRes {
  list: AgentOpLogItem[];
  total: number;
}

/** 分页查询资料 / 账号变更记录 GET /agent/op-logs */
export function listAgentOpLogs(params: AgentOpLogParams) {
  return axios.get<AgentOpLogRes>('/agent/op-logs', { params });
}

/** 关联搜索的关键字类型：0 不筛，1 登录IP，2 设备码 */
export const AGENT_ASSOC_KEYWORD = {
  NONE: 0,
  IP: 1,
  UDID: 2,
} as const;

export interface AgentLoginAssocParams {
  /** 页码，默认 1 */
  page?: number;
  /** 每页条数，默认 20，上限 100 */
  page_size?: number;
  /** 登录时间起止 Unix 秒，0 不限 */
  time_start?: number;
  time_end?: number;
  /** 0 不限 */
  agent_id?: number;
  /** 取值见 AGENT_LOGIN_DEVICE */
  device_type?: number;
  /** 取值见 AGENT_ASSOC_KEYWORD */
  keyword_type?: number;
  keyword?: string;
}

export interface AgentLoginAssocItem {
  id: number;
  agent_id: number;
  account: string;
  login_ip: string;
  login_location: string;
  udid: string;
  login_device: string;
  device_type: number;
  device_type_label: string;
  /** Unix 秒 */
  login_at: number;
  /** 时间范围与设备类型范围内的同 IP 代理数，不受 agent_id / keyword 影响 */
  same_ip_agent_count: number;
  same_device_agent_count: number;
}

export interface AgentLoginAssocRes {
  list: AgentLoginAssocItem[];
  total: number;
}

/** 关联检索登录日志 POST /agent/login-log/associations */
export function listAgentLoginAssociations(data: AgentLoginAssocParams) {
  return axios.post<AgentLoginAssocRes>('/agent/login-log/associations', data);
}

export interface AgentTreeNodeItem {
  id: number;
  account: string;
  line_code: string;
  level: number;
  level_alias: string;
  direct_agent_count: number;
}

/**
 * 某一层的直属下级 GET /agent/tree/node。
 * 不传 parent_id（或传 0）返回顶层。
 */
export function listAgentTreeNodes(parentId?: number) {
  return axios.get<AgentTreeNodeItem[]>('/agent/tree/node', {
    params: parentId ? { parent_id: parentId } : {},
  });
}

export interface AgentTreeAncestorItem {
  id: number;
  account: string;
  level: number;
  level_alias: string;
  line_code: string;
  /** 百分比 0~100 */
  share_rate: number;
}

/**
 * 按代理 ID / 账号精确查找当前代理及其全部上级 GET /agent/tree/ancestors。
 * 当前代理在数组最前，其后依次为直接上级直到顶层。
 * 纯数字先按 ID 匹配，没有再按账号匹配；其他只按账号匹配。
 */
export function listAgentTreeAncestors(keyword: string) {
  return axios.get<AgentTreeAncestorItem[]>('/agent/tree/ancestors', {
    params: { keyword },
  });
}

export interface AgentCreateFields {
  account: string;
  code?: string;
  password: string;
  /** 6 位纯数字 */
  withdraw_password?: string;
  agent_type: number;
  agent_line_code: string;
  /** 万分比，10000 = 100% */
  share_rate: number;
  region_code?: string;
  currency_code?: string;
  promo_code?: string;
  remark?: string;
}

export interface CreateAgentParams {
  agent: AgentCreateFields;
  /** true 时可走指定线 / root 逻辑 */
  is_init?: boolean;
  /** 上级 ID。0 且非初始化时会失败 */
  parent_id?: number;
}

export interface CreateAgentRes {
  id: number;
  account: string;
  code: string;
  parent_id: number;
  top_id: number;
  level: number;
  share_rate: number;
}

/** 分站添加代理 POST /agent/create */
export function createAgent(data: CreateAgentParams) {
  return axios.post<CreateAgentRes>('/agent/create', data);
}

/** 营运商代理（按线路分组） */
export interface OperatorAgentItem {
  id: number;
  account: string;
  level: number;
  level_alias: string;
  line_code: string;
  /** 百分比 0~100；部分接口可能不下发 */
  share_rate?: number;
}

/** 代理线路编码（与 operator-base 已开通线路一致） */
export const AGENT_LINE = {
  CASH_PRODUCTION: 'CASH_PRODUCTION',
  CASH_DEMO: 'CASH_DEMO',
  CASH_TEST: 'CASH_TEST',
  CREDIT_DEMO: 'CREDIT_DEMO',
  CREDIT_TEST: 'CREDIT_TEST',
} as const;

export interface OperatorAgentGroup {
  line_code: string;
  agents: OperatorAgentItem[];
}

export interface OperatorAgentListRes {
  groups: OperatorAgentGroup[];
}

/**
 * 按层级查询代理，并按 agent_line_code 分组。
 * GET /agent/operator/list，业务上常用 level=2。
 */
export function listOperatorAgents(level: number) {
  return axios.get<OperatorAgentListRes>('/agent/operator/list', {
    params: { level },
  });
}

export interface AgentLineDefault {
  line_code: string;
  /** 0 表示未配置 */
  agent_id: number;
  /** 百分比 0~100，未配置为 0 */
  share_rate: number;
}

export interface AgentLineDefaultsRes {
  list: AgentLineDefault[];
}

/** 当前分站已开通代理线的默认代理及占成 GET /agent/line/defaults */
export function getAgentLineDefaults() {
  return axios.get<AgentLineDefaultsRes>('/agent/line/defaults');
}

export interface AgentInitLine {
  line_code: string;
  /** 该线路至少有一名 1 级代理 */
  initialized: boolean;
}

export interface AgentInitStatus {
  /** 等级已就绪且每条已开通线路都已初始化；没有开通线路时只看等级 */
  ready: boolean;
  /** agent_level 至少 3 层 */
  levels_ready: boolean;
  level_count: number;
  lines: AgentInitLine[];
}

/** 已开通代理线是否完成初始化、代理等级是否就绪 GET /init/status */
export function getAgentInitStatus() {
  return axios.get<AgentInitStatus>('/agent/init/status');
}

export interface AgentSearchParams {
  /** 纯数字按 ID/账号模糊，否则按账号模糊 */
  keyword: string;
  /** 代理线编码 */
  line_code?: string;
  /** <=0 走 CRM 规则（通常 >=2） */
  level?: number;
  /** 默认 20，上限 50 */
  limit?: number;
}

export interface AgentSearchItem {
  id: number;
  account: string;
  level: number;
  level_alias: string;
  line_code: string;
  /** 百分比 0~100 */
  share_rate: number;
}

export interface AgentSearchRes {
  list: AgentSearchItem[];
}

export interface PhoneCallingCode {
  /** ISO 3166-1 alpha-2 */
  region: string;
  /** 国际区号，不含 + */
  code: string;
}

/** 支持的国际手机区号 GET /agent/phone-calling-codes */
export function getPhoneCallingCodes() {
  return axios.get<{ list: PhoneCallingCode[] }>('/agent/phone-calling-codes');
}

/** 关键字搜索代理（具体选人）POST /agent/search */
export function searchAgents(data: AgentSearchParams, silentError = false) {
  return axios.post<AgentSearchRes>('/agent/search', data, {
    silentError,
  } as AxiosRequestConfig);
}

export interface AgentLevelInput {
  /** 1~7，同一请求里不可重复 */
  level: number;
  name: string;
  alias: string;
}

export interface AgentLevelRecord extends AgentLevelInput {
  id: number;
}

/** 当前租户代理等级列表 GET /agent/level/list */
export function listAgentLevels() {
  return axios.get<{ list: AgentLevelRecord[] }>('/agent/level/list');
}

/** 批量保存代理等级 POST /agent/level/update */
export function updateAgentLevels(list: AgentLevelInput[]) {
  return axios.post<{ list: AgentLevelRecord[] }>('/agent/level/update', {
    list,
  });
}

/** 与 GET /agent/config 的 data 相同。指定代理 ID 为 0 表示清空 */
export interface AgentConfig {
  /** 配置主键；无配置时为 0。更新请求不传 */
  id?: number;
  /** 最高占成，百分比 0~100 */
  max_share_rate: number;
  /** 申请最低占成 */
  apply_min_share_rate: number;
  /** 申请最高占成；须 ≥ min 且 ≤ max */
  apply_max_share_rate: number;
  /** 现金代理最大会员数 */
  cash_max_member_count: number;
  /** 现金 IP 最大可注册代理数 */
  cash_max_agent_per_ip: number;
  /** 现金正式线指定代理 */
  cash_production_agent_id: number;
  /** 现金试玩线指定代理 */
  cash_demo_agent_id: number;
  /** 现金测试线指定代理 */
  cash_test_agent_id: number;
  /** 信誉试玩线指定代理 */
  credit_demo_agent_id: number;
  /** 信誉测试线指定代理 */
  credit_test_agent_id: number;
}

/** 当前代理配置 GET /agent/config */
export function getAgentConfig() {
  return axios.get<AgentConfig>('/agent/config');
}

/** 保存代理配置 POST /agent/config/update */
export function updateAgentConfig(data: AgentConfig) {
  return axios.post<AgentConfig>('/agent/config/update', data);
}
