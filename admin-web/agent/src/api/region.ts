import axios from 'axios';

/** 1 启用 / 2 停用 */
export const REGION_STATUS = {
  ENABLED: 1,
  DISABLED: 2,
} as const;

/** 国家地区基础数据，无创建接口 */
export interface RegionRecord {
  /** 国家或地区 ID */
  id: number;
  /** 国家或地区编码 */
  code: string;
  /** 国际电话区号，不含 + */
  calling_code: string;
  /** 名称翻译 Key */
  name_key: string;
  /** 当前语言名称（开发阶段调试用） */
  name: string;
  /** 1 启用 / 2 停用 */
  status: number;
  /** 排序值，数值越小越靠前 */
  sort_no: number;
}

export interface RegionListParams {
  /** 页码，从 1 开始，默认 1 */
  page?: number;
  /** 每页数量，范围 1-100，默认 20 */
  page_size?: number;
  /** 1 启用 / 2 停用 */
  status?: number;
}

export interface RegionListRes {
  total: number;
  list: RegionRecord[];
}

/** 分页国家地区管理列表 GET /admin/region/list（走 18001） */
export function queryRegionList(params: RegionListParams = {}) {
  return axios.get<RegionListRes>('/base/region/list', { params });
}

export interface RegionListAllParams {
  /** 1 启用 / 2 停用；不传则不筛选 */
  status?: number;
}

export interface RegionListAllRes {
  list: RegionRecord[];
}

/** 不分页获取全部国家地区 GET /admin/region/list-all（走 18001） */
export function queryRegionListAll(params: RegionListAllParams = {}) {
  return axios.get<RegionListAllRes>('/base/region/list-all', { params });
}

export interface RegionGetParams {
  /** 国家或地区 ID，大于 0 */
  id: number;
}

/** 根据 ID 获取国家地区 GET /admin/region/get（走 18001），data 即为 RegionInfo */
export function getRegion(params: RegionGetParams) {
  return axios.get<RegionRecord>('/base/region/get', { params });
}

export interface UpdateRegionParams {
  /** 国家或地区 ID，大于 0 */
  id: number;
  /** 国际电话区号，不含 +，最大 3 个字符；传字符串如 "81" */
  calling_code?: string;
  /** 1 启用 / 2 停用 */
  status?: number;
}

/**
 * 修改国家地区 POST /admin/region/update（走 18001）
 * 当前仅允许改 calling_code / status，且至少传其中一个；
 * 不可改 code / name_key
 */
export function updateRegion(data: UpdateRegionParams) {
  return axios.post<Record<string, never>>('/base/region/update', data);
}

export interface ReorderRegionParams {
  /** 需要移动的国家或地区 ID，大于 0 */
  id: number;
  /** 目标国家或地区 ID，大于 0 */
  target_id: number;
}

/** 调整国家地区排序 POST /admin/region/reorder（走 18001） */
export function reorderRegion(data: ReorderRegionParams) {
  return axios.post<Record<string, never>>('/base/region/reorder', data);
}
