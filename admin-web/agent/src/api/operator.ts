import axios from 'axios';

/** 1 草稿 / 2 已完成 */
export const OPERATOR_CREATION_STATUS = {
  DRAFT: 1,
  COMPLETED: 2,
} as const;

/** 1 未发布 / 2 发布中 / 3 已发布 / 4 发布失败 */
export const OPERATOR_PUBLISH_STATUS = {
  UNPUBLISHED: 1,
  PUBLISHING: 2,
  PUBLISHED: 3,
  FAILED: 4,
} as const;

/** 1 正常 / 2 暂停 / 3 关闭 */
export const OPERATOR_STATUS = {
  NORMAL: 1,
  PAUSED: 2,
  CLOSED: 3,
} as const;

export interface OperatorRecord {
  /** 分站 ID */
  id: number;
  /** 分站业务编码，由后端自动生成 */
  code: string;
  /** 分站名称 */
  name: string;
  /** IANA 时区编码 */
  timezone_code: string;
  /** 结算币种编码 */
  settlement_currency_code: string;
  /** 1 草稿 / 2 已完成 */
  creation_status: number;
  /** 1 未发布 / 2 发布中 / 3 已发布 / 4 发布失败 */
  publish_status: number;
  /** 1 正常 / 2 暂停 / 3 关闭 */
  status: number;
  /** 总网内部备注 */
  remark: string;
  /** 最近一次成功发布时间，未发布过为 null */
  published_at: number | null;
  created_at: number;
  updated_at: number;
}

/** 发布中、或已有过成功发布时间的分站，不允许再修改时区与结算币种 */
export function isOperatorBaseCurrencyLocked(record: OperatorRecord) {
  return (
    record.publish_status === OPERATOR_PUBLISH_STATUS.PUBLISHING ||
    Boolean(record.published_at)
  );
}

/** 仅未发布、发布失败的分站允许删除 */
export function isOperatorDeletable(record: OperatorRecord) {
  return (
    record.publish_status === OPERATOR_PUBLISH_STATUS.UNPUBLISHED ||
    record.publish_status === OPERATOR_PUBLISH_STATUS.FAILED
  );
}

export interface OperatorListParams {
  /** 页码，从 1 开始 */
  page: number;
  /** 每页数量，范围 1-100 */
  page_size: number;
  /** 搜索关键字，匹配分站编码或名称，最大 100 个字符 */
  keyword?: string;
  /** 1 草稿 / 2 已完成 */
  creation_status?: number;
  /** 1 未发布 / 2 发布中 / 3 已发布 / 4 发布失败 */
  publish_status?: number;
  /** 1 正常 / 2 暂停 / 3 关闭 */
  status?: number;
}

export interface OperatorListRes {
  total: number;
  list: OperatorRecord[];
}

/** 分站管理列表 GET /admin/operator/list，按创建时间倒序 */
export function queryOperatorList(params: OperatorListParams) {
  return axios.get<OperatorListRes>('/operator/list', { params });
}

export interface OperatorIdParams {
  /** 分站 ID，大于 0 */
  id: number;
}

/** 按分站 ID 获取分站详情 GET /admin/operator/get，data 即为 OperatorRecord */
export function getOperator(params: OperatorIdParams) {
  return axios.get<OperatorRecord>('/operator/get', { params });
}

export interface CreateOperatorParams {
  /** 分站名称，非空，最大 100 个字符 */
  name: string;
  /** 时区编码，需在基础数据中存在且已启用，最大 64 个字符 */
  timezone_code: string;
  /** 结算币种编码，需在基础数据中存在且已启用，最大 16 个字符 */
  settlement_currency_code: string;
  /** 1 正常 / 2 暂停 / 3 关闭，不传默认 1 */
  status?: number;
  /** 总网内部备注，最大 1000 个字符 */
  remark?: string;
}

export interface CreateOperatorRes {
  id: number;
  code: string;
}

/**
 * 创建分站基础信息 POST /admin/operator/create
 * 分站编码由后端生成；新建后 creation_status=1、publish_status=1
 */
export function createOperator(data: CreateOperatorParams) {
  return axios.post<CreateOperatorRes>('/operator/create', data);
}

export interface UpdateOperatorParams {
  /** 分站 ID，大于 0 */
  id: number;
  /** 分站名称，非空，最大 100 个字符 */
  name?: string;
  /** 时区编码，需在基础数据中存在且已启用 */
  timezone_code?: string;
  /** 结算币种编码，需在基础数据中存在且已启用 */
  settlement_currency_code?: string;
  /** 1 正常 / 2 暂停 / 3 关闭 */
  status?: number;
  /** 总网内部备注，最大 1000 个字符 */
  remark?: string;
}

/** 修改分站基础信息 POST /admin/operator/update，除 id 外至少传一个字段 */
export function updateOperator(data: UpdateOperatorParams) {
  return axios.post<Record<string, never>>('/operator/update', data);
}

/**
 * 完成分站创建 POST /admin/operator/complete，将 creation_status 置为 2
 * 后端不会在此校验档案、域名与资源是否完整，需前端在创建流程结束后调用
 */
export function completeOperator(data: OperatorIdParams) {
  return axios.post<Record<string, never>>('/operator/complete', data);
}

/**
 * 发布分站 POST /admin/operator/publish
 * 后端发布逻辑仍是 TODO，调用会返回成功结构但不会真正完成发布，
 * 因此不能据此判断发布是否成功
 */
export function publishOperator(data: OperatorIdParams) {
  return axios.post<Record<string, never>>('/operator/publish', data);
}

/** 删除分站 POST /admin/operator/delete，仅未发布 / 发布失败可删除 */
export function deleteOperator(data: OperatorIdParams) {
  return axios.post<Record<string, never>>('/operator/delete', data);
}

export interface CreateOperatorProfileParams {
  /** 分站 ID，大于 0 */
  operator_id: number;
  /** 公司名称，最大 200 个字符 */
  company_name?: string;
  /** 主要联系人名称，最大 100 个字符 */
  contact_name?: string;
  /** 主要联系人邮箱，邮箱格式，最大 255 个字符 */
  contact_email?: string;
  /** 总网内部档案备注，最大 1000 个字符 */
  remark?: string;
}

export interface CreateOperatorProfileRes {
  id: number;
}

/** 创建分站档案 POST /admin/operator/profile/create */
export function createOperatorProfile(data: CreateOperatorProfileParams) {
  return axios.post<CreateOperatorProfileRes>('/operator/profile/create', data);
}

export interface UpdateOperatorProfileParams {
  /** 分站 ID，大于 0；除 operator_id 外至少传一个字段 */
  operator_id: number;
  company_name?: string;
  contact_name?: string;
  contact_email?: string;
  remark?: string;
}

/** 修改分站档案 POST /admin/operator/profile/update */
export function updateOperatorProfile(data: UpdateOperatorProfileParams) {
  return axios.post<Record<string, never>>('/operator/profile/update', data);
}

export interface OperatorProfileRecord {
  id: number;
  operator_id: number;
  company_name: string;
  contact_name: string;
  contact_email: string;
  remark: string;
  created_at: number;
  updated_at: number;
}

export interface OperatorProfileGetRes {
  /** 分站存在但尚未创建档案时为 null */
  profile: OperatorProfileRecord | null;
}

/** 获取分站档案 GET /admin/operator/profile/get?operator_id= */
export function getOperatorProfile(params: { operator_id: number }) {
  return axios.get<OperatorProfileGetRes>('/operator/profile/get', { params });
}

/** 1 分站后台 / 2 代理后台 / 3 会员 H5 */
export const OPERATOR_DOMAIN_TYPE = {
  ADMIN: 1,
  AGENT: 2,
  H5: 3,
} as const;

/** 1 启用 / 2 停用 */
export const OPERATOR_DOMAIN_STATUS = {
  ENABLED: 1,
  DISABLED: 2,
} as const;

export interface CreateOperatorDomainParams {
  /** 分站 ID，大于 0 */
  operator_id: number;
  /** FQDN，不含协议和端口，最大 253 个字符 */
  domain_name: string;
  /** 1 分站后台 / 2 代理后台 / 3 会员 H5 */
  domain_type: number;
  /** 1 启用 / 2 停用，不传默认由后端处理 */
  status?: number;
  /** 总网内部备注，最大 1000 个字符 */
  remark?: string;
}

export interface CreateOperatorDomainRes {
  id: number;
}

/** 创建分站域名 POST /admin/operator/domain/create */
export function createOperatorDomain(data: CreateOperatorDomainParams) {
  return axios.post<CreateOperatorDomainRes>('/operator/domain/create', data);
}

export interface OperatorDomainRecord {
  id: number;
  operator_id: number;
  domain_name: string;
  domain_type: number;
  status: number;
  remark: string;
  created_at: number;
  updated_at: number;
}

export interface OperatorDomainListParams {
  page: number;
  page_size: number;
  operator_id?: number;
  keyword?: string;
  domain_type?: number;
  status?: number;
}

export interface OperatorDomainListRes {
  total: number;
  list: OperatorDomainRecord[];
}

/** 分站域名分页列表 GET /admin/operator/domain/list */
export function queryOperatorDomainList(params: OperatorDomainListParams) {
  return axios.get<OperatorDomainListRes>('/operator/domain/list', { params });
}

export interface UpdateOperatorDomainParams {
  /** 域名 ID，大于 0；除 id 外至少传一个字段 */
  id: number;
  domain_name?: string;
  domain_type?: number;
  status?: number;
  remark?: string;
}

/** 修改分站域名 POST /admin/operator/domain/update */
export function updateOperatorDomain(data: UpdateOperatorDomainParams) {
  return axios.post<Record<string, never>>('/operator/domain/update', data);
}

/** 按域名 ID 获取详情 GET /admin/operator/domain/get，data 即为 OperatorDomainRecord */
export function getOperatorDomain(params: { id: number }) {
  return axios.get<OperatorDomainRecord>('/operator/domain/get', { params });
}

/** 删除分站域名 POST /admin/operator/domain/delete */
export function deleteOperatorDomain(data: { id: number }) {
  return axios.post<Record<string, never>>('/operator/domain/delete', data);
}

export interface OperatorLanguageAllocationItem {
  language_code: string;
  allocated_at: number;
}

export interface BasicResourceAllocationRecord {
  /** 分站 ID */
  operator_id: number;
  /** 分站业务编码 */
  operator_code: string;
  /** 分站名称 */
  operator_name: string;
  /** 已分配语言数量 */
  language_count: number;
  /** 已分配经营地区数量 */
  region_count: number;
  /** 已分配代理子线路数量 */
  agent_line_count: number;
}

export interface BasicResourceAllocationListParams {
  /** 页码，从 1 开始 */
  page: number;
  /** 每页数量，范围 1-100 */
  page_size: number;
  /** 搜索关键字，匹配分站编码或名称，最大 100 个字符 */
  keyword?: string;
}

export interface BasicResourceAllocationListRes {
  total: number;
  list: BasicResourceAllocationRecord[];
}

/** 获取各分站基础资源分配汇总列表 GET /admin/operator/basic-resource-allocation/list */
export function queryBasicResourceAllocationList(
  params: BasicResourceAllocationListParams
) {
  return axios.get<BasicResourceAllocationListRes>(
    '/operator/basic-resource-allocation/list',
    { params }
  );
}

export interface GameResourceAllocationRecord {
  /** 记录 ID */
  id: number;
  /** 分站业务编码 */
  op_code: string;
  /** 分站名称 */
  op_name: string;
  /** 已分配游戏数量 */
  game_count: number;
  /** 已分配游戏分类数量 */
  game_category_count: number;
  /** 已分配游戏厂商数量 */
  game_provider_count: number;
  /** 已分配游戏渠道数量 */
  game_channel_count: number;
  /** 最近更新时间 */
  updated_at: number;
}

export interface GameResourceAllocationListParams {
  /** 页码，从 1 开始 */
  page: number;
  /** 每页数量，范围 1-100 */
  page_size: number;
  /** 分站业务编码，最大 64 个字符 */
  op_code?: string;
  /** 分站状态筛选：1 正常 / 2 暂停 / 3 关闭 */
  status?: number;
}

export interface GameResourceAllocationListRes {
  total: number;
  page: number;
  page_size: number;
  items: GameResourceAllocationRecord[];
}

/** 获取各分站游戏资源分配汇总列表 GET /admin/operator/game-allocation/list */
export function queryGameResourceAllocationList(
  params: GameResourceAllocationListParams
) {
  return axios.get<GameResourceAllocationListRes>(
    '/operator/game-allocation/list',
    { params }
  );
}

/** 1 启用 / 2 停用（分站游戏资源） */
export const OPERATOR_GAME_ITEM_STATUS = {
  ENABLED: 1,
  DISABLED: 2,
} as const;

/** 1 勾选 / 2 不勾选（分站游戏资源分配） */
export const OPERATOR_GAME_CHECK_STATUS = {
  CHECKED: 1,
  UNCHECKED: 2,
} as const;

interface OperatorGameResourceListRes<T> {
  total: number;
  page: number;
  page_size: number;
  items: T[];
}

interface OperatorGameResourceBaseRecord {
  id: number;
  op_code: string;
  name: string;
  /** 1 启用 / 2 停用 */
  status: number;
  /** 1 勾选 / 2 不勾选 */
  check_status: number;
  remark: string;
  deleted_at: number;
  created_at: number;
  updated_at: number;
}

export interface OperatorGameCategoryRecord
  extends OperatorGameResourceBaseRecord {
  category_code: string;
}

export interface OperatorGameProviderRecord
  extends OperatorGameResourceBaseRecord {
  provider_code: string;
}

export interface OperatorGameChannelRecord
  extends OperatorGameResourceBaseRecord {
  channel_code: string;
}

export interface OperatorGameItemRecord extends OperatorGameResourceBaseRecord {
  game_code: string;
}

interface OperatorGameResourceListParams {
  page: number;
  page_size: number;
  op_code?: string;
  status?: number;
}

export interface OperatorGameCategoryListParams
  extends OperatorGameResourceListParams {
  category_code?: string;
}

export interface OperatorGameProviderListParams
  extends OperatorGameResourceListParams {
  provider_code?: string;
}

export interface OperatorGameChannelListParams
  extends OperatorGameResourceListParams {
  channel_code?: string;
}

export interface OperatorGameItemListParams
  extends OperatorGameResourceListParams {
  game_code?: string;
}

/** 获取分站游戏分类列表 GET /admin/operator/game-category/list */
export function queryOperatorGameCategoryList(
  params: OperatorGameCategoryListParams
) {
  return axios.get<OperatorGameResourceListRes<OperatorGameCategoryRecord>>(
    '/operator/game-category/list',
    { params }
  );
}

/** 获取分站游戏厂商列表 GET /admin/operator/game-provider/list */
export function queryOperatorGameProviderList(
  params: OperatorGameProviderListParams
) {
  return axios.get<OperatorGameResourceListRes<OperatorGameProviderRecord>>(
    '/operator/game-provider/list',
    { params }
  );
}

/** 获取分站游戏渠道列表 GET /admin/operator/game-channel/list */
export function queryOperatorGameChannelList(
  params: OperatorGameChannelListParams
) {
  return axios.get<OperatorGameResourceListRes<OperatorGameChannelRecord>>(
    '/operator/game-channel/list',
    { params }
  );
}

/** 获取分站游戏列表 GET /admin/operator/game/list */
export function queryOperatorGameItemList(params: OperatorGameItemListParams) {
  return axios.get<OperatorGameResourceListRes<OperatorGameItemRecord>>(
    '/operator/game/list',
    { params }
  );
}

export interface OperatorGameResourceBatchResult {
  total: number;
  success: number;
  failed: number;
}

export interface SaveOperatorGameCategoryAllocationItem {
  /** 分类业务编码 */
  code: string;
  /** 1 勾选 / 2 不勾选 */
  check_status: number;
}

export interface SaveOperatorGameCategoryAllocationResult {
  total: number;
  created: number;
  deleted: number;
  exist: number;
  failed: number;
}

export interface SaveOperatorGameProviderAllocationItem {
  /** 提供商业务编码 */
  code: string;
  /** 1 勾选 / 2 不勾选 */
  check_status: number;
}

export interface SaveOperatorGameProviderAllocationResult {
  total: number;
  created: number;
  deleted: number;
  exist: number;
  failed: number;
}

export interface SaveOperatorGameChannelAllocationItem {
  /** 渠道业务编码 */
  code: string;
  /** 1 勾选 / 2 不勾选 */
  check_status: number;
}

export interface SaveOperatorGameChannelAllocationResult {
  total: number;
  created: number;
  deleted: number;
  exist: number;
  failed: number;
}

export interface SaveOperatorGameAllocationItem {
  /** 游戏业务编码 */
  code: string;
  /** 1 勾选 / 2 不勾选 */
  check_status: number;
}

export interface SaveOperatorGameAllocationResult {
  total: number;
  created: number;
  deleted: number;
  exist: number;
  failed: number;
}

/** 保存分站游戏分类分配 POST /admin/operator/game-category/save-allocation */
export function batchCreateOperatorGameCategory(data: {
  /** 分站业务编码 */
  op_code: string;
  items: SaveOperatorGameCategoryAllocationItem[];
}) {
  return axios.post<SaveOperatorGameCategoryAllocationResult>(
    '/operator/game-category/save-allocation',
    data
  );
}

/** 保存分站游戏厂商分配 POST /admin/operator/game-provider/save-allocation */
export function batchCreateOperatorGameProvider(data: {
  /** 分站业务编码 */
  op_code: string;
  items: SaveOperatorGameProviderAllocationItem[];
}) {
  return axios.post<SaveOperatorGameProviderAllocationResult>(
    '/operator/game-provider/save-allocation',
    data
  );
}

/** 保存分站游戏渠道分配 POST /admin/operator/game-channel/save-allocation */
export function batchCreateOperatorGameChannel(data: {
  /** 分站业务编码 */
  op_code: string;
  items: SaveOperatorGameChannelAllocationItem[];
}) {
  return axios.post<SaveOperatorGameChannelAllocationResult>(
    '/operator/game-channel/save-allocation',
    data
  );
}

/** 保存分站游戏分配 POST /admin/operator/game/save-allocation */
export function batchCreateOperatorGameItem(data: {
  /** 分站业务编码 */
  op_code: string;
  items: SaveOperatorGameAllocationItem[];
}) {
  return axios.post<SaveOperatorGameAllocationResult>(
    '/operator/game/save-allocation',
    data
  );
}

/** 批量修改分站游戏分类状态 POST /admin/operator/game-category/batch-update-status */
export function batchUpdateOperatorGameCategoryStatus(data: {
  ids: number[];
  status: number;
}) {
  return axios.post<OperatorGameResourceBatchResult>(
    '/operator/game-category/batch-update-status',
    data
  );
}

/** 批量修改分站游戏厂商状态 POST /admin/operator/game-provider/batch-update-status */
export function batchUpdateOperatorGameProviderStatus(data: {
  ids: number[];
  status: number;
}) {
  return axios.post<OperatorGameResourceBatchResult>(
    '/operator/game-provider/batch-update-status',
    data
  );
}

/** 批量修改分站游戏渠道状态 POST /admin/operator/game-channel/batch-update-status */
export function batchUpdateOperatorGameChannelStatus(data: {
  ids: number[];
  status: number;
}) {
  return axios.post<OperatorGameResourceBatchResult>(
    '/operator/game-channel/batch-update-status',
    data
  );
}

/** 批量修改分站游戏状态 POST /admin/operator/game/batch-update-status */
export function batchUpdateOperatorGameItemStatus(data: {
  ids: number[];
  status: number;
}) {
  return axios.post<OperatorGameResourceBatchResult>(
    '/operator/game/batch-update-status',
    data
  );
}

/** 获取分站已分配语言 GET /admin/operator/language-allocation/list */
export function queryOperatorLanguageAllocation(params: {
  operator_id: number;
}) {
  return axios.get<{ list: OperatorLanguageAllocationItem[] }>(
    '/operator/language-allocation/list',
    { params }
  );
}

/**
 * 保存分站语言分配 POST /admin/operator/language-allocation/save
 * 全量覆盖；传空数组表示清空
 */
export function saveOperatorLanguageAllocation(data: {
  operator_id: number;
  language_codes?: string[];
}) {
  return axios.post<Record<string, never>>(
    '/operator/language-allocation/save',
    data
  );
}

export interface OperatorRegionAllocationItem {
  region_code: string;
  allocated_at: number;
}

/** 获取分站已分配经营地区 GET /admin/operator/region-allocation/list */
export function queryOperatorRegionAllocation(params: { operator_id: number }) {
  return axios.get<{ list: OperatorRegionAllocationItem[] }>(
    '/operator/region-allocation/list',
    { params }
  );
}

/**
 * 保存分站经营地区分配 POST /admin/operator/region-allocation/save
 * 全量覆盖；传空数组表示清空
 */
export function saveOperatorRegionAllocation(data: {
  operator_id: number;
  region_codes?: string[];
}) {
  return axios.post<Record<string, never>>(
    '/operator/region-allocation/save',
    data
  );
}

export interface OperatorAgentLineAllocationItem {
  agent_line_code: string;
  allocated: boolean;
  allocated_at: number | null;
}

/**
 * 获取分站代理子线路分配情况 GET /admin/operator/agent-line-allocation/list
 * 与语言/地区不同：返回全部可选线路，用 allocated 标记是否已分配
 */
export function queryOperatorAgentLineAllocation(params: {
  operator_id: number;
}) {
  return axios.get<{ list: OperatorAgentLineAllocationItem[] }>(
    '/operator/agent-line-allocation/list',
    { params }
  );
}

/**
 * 保存分站代理子线路分配 POST /admin/operator/agent-line-allocation/save
 * 全量覆盖；传空数组表示清空；后端会自动去重
 */
export function saveOperatorAgentLineAllocation(data: {
  operator_id: number;
  /** 最终选中的代理子线路编码，单项最大 32 个字符 */
  agent_line_codes?: string[];
}) {
  return axios.post<Record<string, never>>(
    '/operator/agent-line-allocation/save',
    data
  );
}

/** 1 启用 / 2 停用 */
export const OPERATOR_ADMIN_STATUS = {
  ENABLED: 1,
  DISABLED: 2,
} as const;

export interface OperatorAdminRecord {
  id: number;
  operator_id: number;
  operator_name: string;
  username: string;
  display_name: string;
  status: number;
  created_at: number;
  updated_at: number;
}

export interface OperatorAdminListParams {
  /** 页码，从 1 开始 */
  page: number;
  /** 每页数量，范围 1-100 */
  page_size: number;
  /** 分站 ID，大于 0 */
  operator_id?: number;
  /** 匹配账号或显示名称，最大 100 个字符 */
  keyword?: string;
  /** 1 启用 / 2 停用 */
  status?: number;
}

export interface OperatorAdminListRes {
  total: number;
  list: OperatorAdminRecord[];
}

/** 获取分站管理员分页列表 GET /admin/operator/admin/list */
export function queryOperatorAdminList(params: OperatorAdminListParams) {
  return axios.get<OperatorAdminListRes>('/operator/admin/list', { params });
}

export interface CreateOperatorAdminParams {
  /** 分站 ID，大于 0 */
  operator_id: number;
  /** 账号，最大 64 个字符 */
  username: string;
  /** 明文密码，长度 6-32 */
  password: string;
  /** 显示名称，最大 100 个字符 */
  display_name: string;
  /** 1 启用 / 2 停用 */
  status?: number;
}

export interface CreateOperatorAdminRes {
  id: number;
}

/** 创建分站管理员 POST /admin/operator/admin/create */
export function createOperatorAdmin(data: CreateOperatorAdminParams) {
  return axios.post<CreateOperatorAdminRes>('/operator/admin/create', data);
}

export interface UpdateOperatorAdminParams {
  /** 管理员 ID，大于 0 */
  id: number;
  /** 账号，最大 64 个字符 */
  username?: string;
  /** 明文密码，长度 6-32；不传表示不修改 */
  password?: string;
  /** 显示名称，最大 100 个字符 */
  display_name?: string;
  /** 1 启用 / 2 停用 */
  status?: number;
}

/** 修改分站管理员 POST /admin/operator/admin/update，除 id 外至少传一个字段 */
export function updateOperatorAdmin(data: UpdateOperatorAdminParams) {
  return axios.post<Record<string, never>>('/operator/admin/update', data);
}

export interface UpdateOperatorAdminStatusParams {
  /** 管理员 ID，大于 0 */
  id: number;
  /** 1 启用 / 2 停用 */
  status: number;
}

/** 更新分站管理员启停状态 POST /admin/operator/admin/updateStatus */
export function updateOperatorAdminStatus(
  data: UpdateOperatorAdminStatusParams
) {
  return axios.post<Record<string, never>>(
    '/operator/admin/updateStatus',
    data
  );
}

export interface ResetOperatorAdminPasswordParams {
  /** 管理员 ID，大于 0 */
  id: number;
  /** 新明文密码，长度 6-32 */
  password: string;
}

/** 重置分站管理员密码 POST /admin/operator/admin/resetPassword */
export function resetOperatorAdminPassword(
  data: ResetOperatorAdminPasswordParams
) {
  return axios.post<Record<string, never>>(
    '/operator/admin/resetPassword',
    data
  );
}
