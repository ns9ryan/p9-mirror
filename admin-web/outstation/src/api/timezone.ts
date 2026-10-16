import axios from 'axios';

/** 1 启用 / 2 停用 */
export const TIMEZONE_STATUS = {
  ENABLED: 1,
  DISABLED: 2,
} as const;

/** 时区基础数据，无创建接口 */
export interface TimezoneRecord {
  /** 时区 ID */
  id: number;
  /** IANA 时区编码 */
  code: string;
  /** 名称翻译 Key */
  name_key: string;
  /** 当前语言名称（开发阶段调试用） */
  name: string;
  /** 1 启用 / 2 停用 */
  status: number;
  /** 排序值，数值越小越靠前 */
  sort_no: number;
}

export interface TimezoneListParams {
  /** 页码，从 1 开始，默认 1 */
  page?: number;
  /** 每页数量，范围 1-100，默认 20 */
  page_size?: number;
  /** 1 启用 / 2 停用 */
  status?: number;
}

export interface TimezoneListRes {
  total: number;
  list: TimezoneRecord[];
}

/** 分页时区管理列表 GET /admin/timezone/list（走 18001） */
export function queryTimezoneList(params: TimezoneListParams = {}) {
  return axios.get<TimezoneListRes>('/base/timezone/list', { params });
}

export interface TimezoneListAllParams {
  /** 1 启用 / 2 停用；不传则不筛选 */
  status?: number;
}

export interface TimezoneListAllRes {
  list: TimezoneRecord[];
}

/** 不分页获取全部时区 GET /admin/timezone/list-all（走 18001） */
export function queryTimezoneListAll(params: TimezoneListAllParams = {}) {
  return axios.get<TimezoneListAllRes>('/base/timezone/list-all', { params });
}

export interface TimezoneGetParams {
  /** 时区 ID，大于 0 */
  id: number;
}

/** 根据 ID 获取时区 GET /admin/timezone/get（走 18001），data 即为 TimezoneInfo */
export function getTimezone(params: TimezoneGetParams) {
  return axios.get<TimezoneRecord>('/base/timezone/get', { params });
}

export interface UpdateTimezoneParams {
  /** 时区 ID，大于 0 */
  id: number;
  /** 1 启用 / 2 停用；当前仅允许修改状态，不可改 code / name_key */
  status: number;
}

/** 修改时区（仅允许改 status） POST /admin/timezone/update（走 18001） */
export function updateTimezone(data: UpdateTimezoneParams) {
  return axios.post<Record<string, never>>('/base/timezone/update', data);
}

export interface ReorderTimezoneParams {
  /** 需要移动的时区 ID，大于 0 */
  id: number;
  /** 目标时区 ID，大于 0，且不可与 id 相同 */
  target_id: number;
}

/** 调整时区排序 POST /admin/timezone/reorder（走 18001） */
export function reorderTimezone(data: ReorderTimezoneParams) {
  return axios.post<Record<string, never>>('/base/timezone/reorder', data);
}
