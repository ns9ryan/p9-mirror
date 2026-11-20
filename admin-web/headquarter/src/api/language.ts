import axios from 'axios';

/** 1 启用 / 2 停用 */
export const LANGUAGE_STATUS = {
  ENABLED: 1,
  DISABLED: 2,
} as const;

/** 多语言名称，key 为语言码（如 zh-CN / en-US），value 为该语言下的名称 */
export type LanguageI18nName = Record<string, string>;

export interface LanguageRecord {
  /** 语言 ID */
  id: number;
  /** 语言编码 */
  code: string;
  /** 多语言名称 */
  name_i18n: LanguageI18nName;
  /** 1 启用 / 2 停用 */
  status: number;
  /** 排序值，数值越小越靠前 */
  sort_no: number;
}

export interface LanguageListParams {
  /** 页码，从 1 开始，默认 1 */
  page?: number;
  /** 每页数量，范围 1-100，默认 20 */
  page_size?: number;
  /** 1 启用 / 2 停用 */
  status?: number;
}

export interface LanguageListRes {
  total: number;
  list: LanguageRecord[];
}

/** 分页语言管理列表 GET /admin/base/language/list（走 18001） */
export function queryLanguageList(params: LanguageListParams = {}) {
  return axios.get<LanguageListRes>('/base/language/list', { params });
}

export interface LanguageListAllParams {
  /** 1 启用 / 2 停用；不传则不筛选 */
  status?: number;
}

export interface LanguageListAllRes {
  list: LanguageRecord[];
}

/** 不分页获取全部语言 GET /admin/base/language/list-all（走 18001） */
export function queryLanguageListAll(params: LanguageListAllParams = {}) {
  return axios.get<LanguageListAllRes>('/base/language/list-all', { params });
}

export interface LanguageGetParams {
  /** 语言 ID，大于 0 */
  id: number;
}

export interface LanguageGetRes {
  language: LanguageRecord;
}

/** 根据 ID 获取语言 GET /admin/base/language/get（走 18001） */
export function getLanguage(params: LanguageGetParams) {
  return axios.get<LanguageGetRes>('/base/language/get', { params });
}

export interface CreateLanguageParams {
  /** 语言编码，最大 35 个字符 */
  code: string;
  /** 多语言名称 */
  name_i18n: LanguageI18nName;
  /** 1 启用 / 2 停用 */
  status?: number;
}

export interface CreateLanguageRes {
  id: number;
}

/** 创建语言 POST /admin/base/language/create（走 18001） */
export function createLanguage(data: CreateLanguageParams) {
  return axios.post<CreateLanguageRes>('/base/language/create', data);
}

export interface UpdateLanguageParams {
  /** 语言 ID，大于 0 */
  id: number;
  /** 多语言名称，不传时不修改 */
  name_i18n?: LanguageI18nName;
  /** 1 启用 / 2 停用 */
  status?: number;
}

/** 修改语言（不提供语言编码 code 的修改） POST /admin/base/language/update（走 18001） */
export function updateLanguage(data: UpdateLanguageParams) {
  return axios.post<Record<string, never>>('/base/language/update', data);
}

export interface ReorderLanguageParams {
  /** 需要移动的语言 ID，大于 0 */
  id: number;
  /** 目标语言 ID，大于 0 */
  target_id: number;
}

/** 调整语言排序 POST /admin/base/language/reorder（走 18001） */
export function reorderLanguage(data: ReorderLanguageParams) {
  return axios.post<Record<string, never>>('/base/language/reorder', data);
}
