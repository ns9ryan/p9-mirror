import axios from 'axios';
import type { AxiosRequestConfig } from 'axios';
import {
  uploadResource,
  type ResourceUploadResult,
} from './resource';

type SilentConfig = AxiosRequestConfig & { silentError?: boolean };

/** 会员状态：1 正常，2 冻结，3 禁用 */
export const MEMBER_STATUS = {
  NORMAL: 1,
  FROZEN: 2,
  DISABLED: 3,
} as const;

/** 列表时间筛选：0 不筛选，1 注册时间，2 最后登录时间 */
export const MEMBER_TIME_TYPE = {
  NONE: 0,
  REGISTER: 1,
  LOGIN: 2,
} as const;

/** 列表排序字段；0 注册时间时服务端始终倒序，忽略 sort_order */
export const MEMBER_SORT_BY = {
  REGISTERED_AT: 0,
  TOTAL_DEPOSIT: 1,
  DEPOSIT_COUNT: 2,
  TOTAL_WITHDRAWAL: 3,
  WITHDRAWAL_COUNT: 4,
} as const;

/** 排序方向：0 默认降序，1 升序，2 降序 */
export const MEMBER_SORT_ORDER = {
  ASC: 1,
  DESC: 2,
} as const;

/**
 * 会员列表查询参数。可选项不筛选时不要传：
 * 服务端会拒绝不支持的枚举、无效 ID 和冲突条件，不会自动忽略
 */
export interface MemberListParams {
  /** 页码，从 1 开始 */
  page: number;
  /** 每页数量，1~100 */
  page_size: number;
  /** 模糊匹配会员账号、12 位会员编号或昵称 */
  keyword?: string;
  /** 为 0 时不可传时间边界；为 1/2 时至少传一个边界 */
  time_type?: number;
  /** Unix 毫秒，包含 */
  start_at?: number;
  /** Unix 毫秒，不包含；与 start_at 同传时须大于 start_at */
  end_at?: number;
  auto_level_id?: number;
  fixed_level_id?: number;
  /** 精确匹配，最长 32 字节 */
  register_method?: string;
  /** 精确匹配，最长 32 字节 */
  register_channel?: string;
  status?: number;
  agent_id?: number;
  betting_disabled?: boolean;
  no_phone?: boolean;
  no_email?: boolean;
  /** 与 not_deposited 互斥 */
  deposited?: boolean;
  not_deposited?: boolean;
  sort_by?: number;
  sort_order?: number;
}

export interface MemberLevels {
  auto_level_id: number;
  auto_level_name: string;
  /** 按 ID 升序，无记录为 [] */
  fixed_levels: { id: number; name: string }[];
  vip_level_id: number;
  vip_level: number;
}

/** 金额均为结算币种最小单位；无统计记录时为 0。输赢正数表示会员盈利 */
export interface MemberStats {
  total_deposit: number;
  total_deposit_count: number;
  total_withdrawal: number;
  total_withdrawal_count: number;
  total_deposit_withdrawal_diff: number;
  total_bet: number;
  total_valid_bet: number;
  total_water: number;
  total_win_loss: number;
  current_period_deposit: number;
  current_period_withdrawal: number;
  current_period_bet: number;
  current_period_valid_bet: number;
  current_period_water: number;
  current_period_win_loss: number;
  /** Unix 毫秒 */
  current_period_started_at: number | null;
}

/** 列表与详情共有的会员字段 */
export interface MemberBaseItem {
  id: number;
  /** 12 位会员对外编号 */
  code: string;
  account: string | null;
  nickname: string;
  agent_id: number;
  agent_account: string;
  levels: MemberLevels;
  status: number;
  chat_disabled: boolean;
  betting_disabled: boolean;
  withdrawal_disabled: boolean;
  register_method: string;
  register_channel: string;
  /** Unix 毫秒 */
  created_at: number;
  /** Unix 毫秒 */
  last_login_at: number | null;
  last_login_ip: string | null;
  inviter_member_id: number | null;
  remark: string | null;
  stats: MemberStats;
}

export interface MemberListItem extends MemberBaseItem {
  /** 列表暂不填充 */
  phone: string | null;
  /** 列表暂不填充 */
  email: string | null;
}

/** 会员性别：0 未设置，1 男，2 女，3 保密 */
export const MEMBER_GENDER = {
  UNSET: 0,
  MALE: 1,
  FEMALE: 2,
  SECRET: 3,
} as const;

/** 详情不返回手机号、邮箱、真实姓名、身份ID、钱包余额及钱包流水 */
export interface MemberDetail extends MemberBaseItem {
  /** 头像资源 ID，按 getMemberImages 解析图片；0 表示未设置 */
  avatar_id: number;
  /** 头像框资源 ID，按 getMemberImages 解析图片；0 表示未设置 */
  avatar_frame_id: number;
  /** YYYY-MM-DD */
  birthday: string | null;
  gender: number;
  /** 国家或地区编码，如 CN */
  region_code: string | null;
  agent_line_code: string;
  register_ip: string | null;
  register_device_code: string | null;
  register_domain: string | null;
}

export interface MemberListRes {
  total: number;
  list: MemberListItem[];
}

/** 会员列表 GET /admin/operator/member/list（后端 /member/list） */
export function queryMemberList(params: MemberListParams) {
  return axios.get<MemberListRes>('/operator/member/list', { params });
}

export interface MemberExportParams {
  /** 会员状态：1 正常，2 冻结，3 禁用 */
  status?: number;
  agent_type: number;
  include_demo: boolean;
  /** 导出列，按后端字段名传 */
  columns: string[];
}

export interface MemberExportRes {
  task_id: string;
  /** 预签名下载地址，有效期见 expires_at */
  url: string;
  key: string;
  rows: number;
  /** Unix 毫秒 */
  started_at: number;
  /** Unix 毫秒，下载链接过期时间 */
  expires_at: number;
  visibility: string;
}

/** 导出会员 POST /admin/operator/member/export/v2（后端 /member/export/v2），返回文件下载地址 */
export function exportMembers(data: MemberExportParams) {
  return axios.post<MemberExportRes>('/operator/member/export/v2', data);
}

/** 会员详情 GET /admin/operator/member/detail?member_id=（后端 /member/detail） */
export function getMemberDetail(memberId: number, silentError = false) {
  return axios.get<MemberDetail>('/operator/member/detail', {
    params: { member_id: memberId },
    silentError,
  } as SilentConfig);
}

export interface MemberImage {
  id: number;
  image_url: string;
}

/** 只含已配置图片的资源，无资源时为 [] */
export interface MemberImagesRes {
  avatars: MemberImage[];
  avatar_frames: MemberImage[];
}

/**
 * 头像与头像框资源 GET /admin/operator/member/images
 * 总站更新图片时 ID 不变，展示时始终按本接口返回的最新 URL
 */
export function getMemberImages() {
  return axios.get<MemberImagesRes>('/operator/member/images');
}

/**
 * 除 member_id 外均可选，至少传一个；省略或 null 保持原值。
 * real_name、identity_number 传空字符串清空；phone、email 不能清空，修改时必须携带对应验证信息
 */
export interface UpdateMemberContactParams {
  member_id: number;
  /** getMemberImages 返回的正整数 ID，不能传 URL */
  avatar_id?: number;
  avatar_frame_id?: number;
  /** 去首尾空白后非空，最多 100 个字符 */
  nickname?: string;
  /** 最多 100 个字符 */
  real_name?: string;
  /** 身份 ID / CPF，最多 100 个字符 */
  identity_number?: string;
  /** 去空格及前导 + 后为 5~20 位数字，含国家区号；同分站不可重复。修改时必须携带 phone_verification */
  phone?: string;
  /** 最多 254 个字符，服务端转小写；同分站不可重复。修改时必须携带 email_verification */
  email?: string;
  /** 与 phone 同传，来自 sendMemberContactBindingCode（渠道 1） */
  phone_verification?: ContactVerification;
  /** 与 email 同传，来自 sendMemberContactBindingCode（渠道 2） */
  email_verification?: ContactVerification;
}

/** 修改会员联系方式及头像 POST /admin/operator/member/contact/update */
export function updateMemberContact(data: UpdateMemberContactParams) {
  return axios.post<Record<string, never>>(
    '/operator/member/contact/update',
    data
  );
}

/** 发送验证码 HTTP 超时：文档要求 45 秒 */
const CONTACT_CODE_SEND_TIMEOUT = 45 * 1000;

/** 批量创建同步返回，服务端最长 10 分钟；客户端超时需略大于此值，并确认网关配置 */
const BATCH_CREATE_TIMEOUT = 11 * 60 * 1000;

/** 联系方式验证码渠道：1 手机短信，2 邮件 */
export const CONTACT_CHANNEL = {
  SMS: 1,
  EMAIL: 2,
} as const;

/** 验证码记录状态 */
export const CONTACT_BINDING_CODE_STATUS = {
  /** 待发送 */
  PENDING: 0,
  /** 已发送 */
  SENT: 1,
  /** 已使用 */
  USED: 2,
  /** 已作废 */
  VOIDED: 3,
  /** 发送失败 */
  SEND_FAILED: 4,
} as const;

/** 修改手机号或邮箱时必须携带的验证信息 */
export interface ContactVerification {
  record_id: number;
  /** 六位数字 */
  verification_code: string;
}

export interface SendContactBindingCodeParams {
  member_id: number;
  /** 1 手机短信，2 邮件 */
  channel: number;
  /** 手机号需含国际区号；邮箱需为合法邮箱 */
  target: string;
}

export interface SendContactBindingCodeRes {
  record_id: number;
  /** Unix 毫秒，验证码有效期 5 分钟 */
  expires_at: number;
}

/**
 * 发送联系方式验证码 POST /admin/operator/member/contact/binding-codes/send
 * 同一会员同一渠道 60 秒内只能申请一次；新请求会作废旧记录
 */
export function sendMemberContactBindingCode(
  data: SendContactBindingCodeParams
) {
  return axios.post<SendContactBindingCodeRes>(
    '/operator/member/contact/binding-codes/send',
    data,
    { timeout: CONTACT_CODE_SEND_TIMEOUT }
  );
}

export interface ContactBindingCodeListParams {
  page: number;
  page_size: number;
  id?: number;
  member_id?: number;
  channel?: number;
  status?: number;
  /** Unix 毫秒，包含 */
  created_from?: number;
  /** Unix 毫秒，不包含 */
  created_before?: number;
}

export interface ContactBindingCodeRecord {
  id: number;
  member_id: number;
  channel: number;
  status: number;
  failed_attempts: number;
  /** 已失效的验证码明文；仍可用于验证或只有摘要的记录为空字符串 */
  verification_code: string;
  /** Unix 毫秒 */
  created_at: number;
  /** Unix 毫秒，0 表示未发送 */
  sent_at: number;
  /** Unix 毫秒，过期与否由此判断，不改变 status */
  expires_at: number;
  /** Unix 毫秒，0 表示未核销 */
  used_at: number;
}

export interface ContactBindingCodeListRes {
  total: number;
  list: ContactBindingCodeRecord[];
}

/** 验证码记录列表 GET /admin/operator/member/contact/binding-codes/list，只查询不发送 */
export function queryContactBindingCodeList(
  params: ContactBindingCodeListParams
) {
  return axios.get<ContactBindingCodeListRes>(
    '/operator/member/contact/binding-codes/list',
    { params }
  );
}

/** 批量创建会员请求体；前缀与后缀至少填一项 */
export interface BatchCreateMembersParams {
  account_prefix?: string;
  account_suffix?: string;
  /** 1~2000 */
  count: number;
  /** 所选的正常分层代理 ID */
  agent_id: number;
}

export interface BatchCreatedMember {
  id: number;
  /** 12 位会员编号 */
  code: string;
  account: string;
}

export interface BatchCreateMembersRes {
  members: BatchCreatedMember[];
}

/**
 * 批量创建会员 POST /admin/operator/member/batch/create
 * 超时或断线后应先查询会员列表确认，不要直接重试，否则可能重复创建
 */
export function batchCreateMembers(data: BatchCreateMembersParams) {
  return axios.post<BatchCreateMembersRes>(
    '/operator/member/batch/create',
    data,
    { timeout: BATCH_CREATE_TIMEOUT }
  );
}

/**
 * 修改会员状态 POST /admin/operator/member/status/update
 * 只改 member.status，不影响禁言、投注、提款开关；silentError 供批量操作自行汇总失败
 */
export function updateMemberStatus(
  data: { member_id: number; status: number },
  silentError = false
) {
  return axios.post<Record<string, never>>(
    '/operator/member/status/update',
    data,
    { silentError } as SilentConfig
  );
}

/** 修改自动层级 POST /admin/operator/member/level/auto/update */
export function updateMemberAutoLevel(data: {
  member_id: number;
  auto_level_id: number;
}) {
  return axios.post<Record<string, never>>(
    '/operator/member/level/auto/update',
    data
  );
}

/**
 * 全量替换固定层级 POST /admin/operator/member/level/fixed/update
 * 空数组清空；最多 100 项，ID 须为正数
 */
export function updateMemberFixedLevels(data: {
  member_id: number;
  fixed_level_ids: number[];
}) {
  return axios.post<Record<string, never>>(
    '/operator/member/level/fixed/update',
    data
  );
}

/** 修改标签/备注 POST /admin/operator/member/remark/update；最多 200 个字符，空字符串清空 */
export function updateMemberRemark(data: {
  member_id: number;
  remark: string;
}) {
  return axios.post<Record<string, never>>(
    '/operator/member/remark/update',
    data
  );
}

/** chat 禁止发言，betting 禁止投注，withdrawal 禁止提款 */
export type MemberRestriction = 'chat' | 'betting' | 'withdrawal';

/** 禁止/解除禁止 POST /admin/operator/member/{chat|betting|withdrawal}/update，每个接口只改对应开关 */
export function updateMemberRestriction(
  kind: MemberRestriction,
  data: { member_id: number; disabled: boolean }
) {
  return axios.post<Record<string, never>>(
    `/operator/member/${kind}/update`,
    data
  );
}

export type MemberAuthErrorField =
  | 'sms_code_error_count'
  | 'email_code_error_count'
  | 'login_password_error_count'
  | 'withdrawal_password_error_count';

/**
 * 重置认证错误次数 POST /admin/operator/member/auth/errors/reset
 * 只清零指定计数，不改密码、不发验证码
 */
export function resetMemberAuthErrors(data: {
  member_id: number;
  field: MemberAuthErrorField;
}) {
  return axios.post<Record<string, never>>(
    '/operator/member/auth/errors/reset',
    data
  );
}

/** 后端严格校验请求字段，未支持的字段不要传，否则返回参数无效 */
export interface CreateMemberParams {
  /** 4~20 位英文字母、数字、- 或 _，厅内唯一 */
  account: string;
  /** 8~32 位 ASCII 可见字符，不含空格 */
  password: string;
  /** 最多 100 个字符；不传或空白时服务端使用账号 */
  nickname?: string;
  /** 所属分层代理 ID，须存在、状态正常且层级有效 */
  agent_id: number;
  /** 最多 200 个字符 */
  remark?: string;
}

/** 创建会员 POST /admin/operator/member/create；code 为 12 位会员编号字符串 */
export function createMember(data: CreateMemberParams) {
  return axios.post<{ id: number; code: string }>(
    '/operator/member/create',
    data
  );
}

export interface AutoLevelListParams {
  /** 页码，从 1 开始 */
  page: number;
  /** 每页数量，1~100 */
  page_size: number;
  /** 层级名称模糊查询，最多 100 个字符；不传或空串表示不按名称筛选 */
  name?: string;
}

export interface AutoLevelRecord {
  id: number;
  name: string;
  /** 晋升要求的最低累计存款金额 */
  min_deposit: number;
  /** 晋升要求的最低累计洗码量 */
  min_water: number;
  /** 当前归属该自动层级的会员数 */
  current_member_count: number;
  /** Unix 毫秒 */
  created_at: number;
  /** Unix 毫秒 */
  updated_at: number;
}

export interface AutoLevelListRes {
  total: number;
  list: AutoLevelRecord[];
}

/** 自动层级分页列表 GET /admin/operator/member/auto-level/list */
export function queryAutoLevelList(params: AutoLevelListParams) {
  return axios.get<AutoLevelListRes>('/operator/member/auto-level/list', {
    params,
  });
}

export interface FixedLevelRecord {
  id: number;
  name: string;
  /** 无说明时为空字符串 */
  description: string;
  /** 单笔提款金额上限，null 表示不额外限制 */
  withdrawal_limit: number | null;
  /** 当前关联到该固定层级的会员数 */
  current_member_count: number;
  /** Unix 毫秒 */
  created_at: number;
  /** Unix 毫秒 */
  updated_at: number;
}

export interface FixedLevelListRes {
  total: number;
  list: FixedLevelRecord[];
}

/** 固定层级分页列表 GET /admin/operator/member/fixed-level/list，参数同自动层级 */
export function queryFixedLevelList(params: AutoLevelListParams) {
  return axios.get<FixedLevelListRes>('/operator/member/fixed-level/list', {
    params,
  });
}

export interface CreateFixedLevelParams {
  /** 非空，最多 100 个字符 */
  name: string;
  /** 最多 255 个字符，不传按空字符串处理 */
  description?: string;
  /** 单笔提款金额上限，传入时必须 > 0；不传或 null 表示不额外限制 */
  withdrawal_limit?: number | null;
}

/** 创建固定层级 POST /admin/operator/member/fixed-level/create */
export function createFixedLevel(data: CreateFixedLevelParams) {
  return axios.post<{ id: number }>(
    '/operator/member/fixed-level/create',
    data
  );
}

export interface UpdateFixedLevelParams extends CreateFixedLevelParams {
  id: number;
}

/**
 * 修改固定层级 POST /admin/operator/member/fixed-level/update
 * 按整组字段覆盖：description 不传会清空，withdrawal_limit 不传或 null 会改为不限制
 */
export function updateFixedLevel(data: UpdateFixedLevelParams) {
  return axios.post<Record<string, never>>(
    '/operator/member/fixed-level/update',
    data
  );
}

/**
 * 删除固定层级 POST /admin/operator/member/fixed-level/delete
 * 服务端会先解除该层级与会员的关联再删除，不删除会员本身
 */
export function deleteFixedLevel(id: number) {
  return axios.post<Record<string, never>>(
    '/operator/member/fixed-level/delete',
    { id }
  );
}

export interface CreateAutoLevelParams {
  /** 非空，最多 100 个字符 */
  name: string;
  /** 最低累计存款金额，>= 0 */
  min_deposit: number;
  /** 最低累计洗码量，>= 0 */
  min_water: number;
}

/**
 * 创建自动层级 POST /admin/operator/member/auto-level/create
 * 两项门槛都必须严格大于当前分站所有已有自动层级的对应门槛
 */
export function createAutoLevel(data: CreateAutoLevelParams) {
  return axios.post<{ id: number }>('/operator/member/auto-level/create', data);
}

/** reward_water_multiplier 存储值放大倍数：显示倍率 = 接口值 / 100000 */
export const VIP_WATER_MULTIPLIER_SCALE = 100000;

/** 奖励派发模式：1 手动领取，2 自动派发 */
export type VipRewardGrantMode = 1 | 2;

export interface VipConfig {
  /** VIP 总开关 */
  vip_enabled: boolean;
  /** 晋级奖励开关 */
  upgrade_reward_enabled: boolean;
  /** 保级考核开关 */
  retention_check_enabled: boolean;
  reward_grant_mode: VipRewardGrantMode;
  /** 晋级奖励洗码倍率，放大 VIP_WATER_MULTIPLIER_SCALE 倍 */
  reward_water_multiplier: number;
  /** 最近一次从关闭切换为开启的时间，Unix 毫秒；尚无开启时间时为 null */
  last_enabled_at: number | null;
  /** Unix 毫秒 */
  updated_at: number;
}

/** VIP 基础配置 GET /admin/operator/member/vip/config（后端 /member/vip/config） */
export function getVipConfig() {
  return axios.get<VipConfig>('/operator/member/vip/config');
}

/** 修改 VIP 基础配置的请求体；字段均可选，只更新传入字段 */
export interface UpdateVipConfigParams {
  vip_enabled?: boolean;
  upgrade_reward_enabled?: boolean;
  retention_check_enabled?: boolean;
  /** 仅允许 1 或 2 */
  reward_grant_mode?: VipRewardGrantMode;
  /** >= 0，按 VIP_WATER_MULTIPLIER_SCALE 放大后提交 */
  reward_water_multiplier?: number;
}

/**
 * 修改 VIP 基础配置 POST /admin/operator/member/vip/config/update
 * 成功后不回写配置，需再调 getVipConfig 同步页面
 */
export function updateVipConfig(data: UpdateVipConfigParams) {
  return axios.post<Record<string, never>>(
    '/operator/member/vip/config/update',
    data
  );
}

export interface VipLevelListParams {
  /** 页码，从 1 开始 */
  page: number;
  /** 每页数量，1~100 */
  page_size: number;
}

export interface VipLevelApiRecord {
  id: number;
  /** VIP 等级编号，从 0 开始 */
  level: number;
  upgrade_deposit: number;
  upgrade_water: number;
  retention_deposit: number;
  retention_water: number;
  upgrade_reward: number;
  member_count: number;
  /** Unix 毫秒 */
  updated_at: number;
}

export interface VipLevelListRes {
  total: number;
  /** 按 level 升序 */
  list: VipLevelApiRecord[];
}

/** VIP 等级分页列表 GET /admin/operator/member/vip/level/list */
export function queryVipLevelList(params: VipLevelListParams) {
  return axios.get<VipLevelListRes>('/operator/member/vip/level/list', {
    params,
  });
}

/** 创建 VIP 等级的请求体；level 由服务端按当前最高等级 + 1 分配，不可提交 */
export interface CreateVipLevelParams {
  /** 以下字段均 >= 0 的整数 */
  upgrade_deposit: number;
  upgrade_water: number;
  retention_deposit: number;
  retention_water: number;
  upgrade_reward: number;
}

/**
 * 创建 VIP 等级 POST /admin/operator/member/vip/level/create
 * 分站未初始化 VIP 等级数据时会失败
 */
export function createVipLevel(data: CreateVipLevelParams) {
  return axios.post<{ id: number; level: number }>(
    '/operator/member/vip/level/create',
    data
  );
}

/** 修改 VIP 等级的请求体；除 id 外只更新传入字段，且至少传一个 */
export interface UpdateVipLevelParams extends Partial<CreateVipLevelParams> {
  id: number;
}

/**
 * 修改 VIP 等级 POST /admin/operator/member/vip/level/update
 * VIP0 不可修改；已有会员的等级不可修改晋级/保级门槛，仅可改 upgrade_reward
 */
export function updateVipLevel(data: UpdateVipLevelParams) {
  return axios.post<Record<string, never>>(
    '/operator/member/vip/level/update',
    data
  );
}

/**
 * 删除 VIP 等级 POST /admin/operator/member/vip/level/delete
 * 仅允许删除当前最高等级且会员数为 0；VIP0 禁止删除
 */
export function deleteVipLevel(id: number) {
  return axios.post<Record<string, never>>(
    '/operator/member/vip/level/delete',
    { id }
  );
}

export interface VipIconGroupRecord {
  /** 本站记录主键；本站尚无此组记录时为 0，不能用作行 key 或请求参数 */
  id: number;
  /** 图标组编号 0–9，上传 / 保存 / 恢复都按它定位 */
  group_no: number;
  /** group_no × 10 */
  start_level: number;
  /** start_level + 9 */
  end_level: number;
  /** 总站本组实时默认图完整 URL；总站未配置时为 "" */
  default_icon_url: string;
  /** 当前实际展示的完整 URL（本站自定义优先，否则总站默认）；都无图时为 "" */
  icon_url: string;
  /** 本站记录更新时间，Unix 毫秒；本站尚无记录时为 0 */
  updated_at: number;
}

/** VIP 图标上传结果（统一 /resource/upload 响应） */
export type VipIconUploadResult = ResourceUploadResult;

/** VIP 图标规格：PNG，宽 115 × 高 165，最大 200 KiB */
export const VIP_ICON_WIDTH = 115;
export const VIP_ICON_HEIGHT = 165;
export const VIP_ICON_MAX_SIZE = 204800;

/** VIP 图标组列表 GET /admin/operator/member/vip/icon/list，固定十组，按 group_no 升序 */
export function queryVipIconList() {
  return axios.get<{ list: VipIconGroupRecord[] }>(
    '/operator/member/vip/icon/list'
  );
}

/**
 * 上传本站 VIP 图标 POST /resource/upload
 * scene=vipicon，min_level=group_no×10；只上传不改配置，需再调用 updateVipIcon
 */
export function uploadVipIcon(groupNo: number, file: File) {
  return uploadResource({
    file,
    scene: 'vipicon',
    min_level: groupNo * 10,
  });
}

/**
 * 保存 VIP 图标 POST /admin/operator/member/vip/icon/update
 * icon_url 传同组上传返回的 key；不能传空串，清空必须走 resetVipIcon
 */
export function updateVipIcon(data: { group_no: number; icon_url: string }) {
  return axios.post<Record<string, never>>(
    '/operator/member/vip/icon/update',
    data
  );
}

/**
 * 恢复总站默认图 POST /admin/operator/member/vip/icon/reset
 * 清空本站该组自定义图，之后跟随总站默认图变化；立即生效，无需再保存
 */
export function resetVipIcon(groupNo: number) {
  return axios.post<Record<string, never>>(
    '/operator/member/vip/icon/reset',
    { group_no: groupNo }
  );
}

/** VIP 等级详情 GET /admin/operator/member/vip/level/detail?id=，data 同列表单项 */
export function getVipLevelDetail(id: number) {
  return axios.get<VipLevelApiRecord>('/operator/member/vip/level/detail', {
    params: { id },
  });
}
