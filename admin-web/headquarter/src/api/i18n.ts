import axios from 'axios';
import type { AxiosRequestConfig } from 'axios';

type RequestConfig = AxiosRequestConfig & { silentError?: boolean };

export interface I18nRecord {
  id: number;
  /** 站点：platform / operator / agent / user */
  i18n_code?: string;
  i18n_group: string;
  lang: string;
  trans_key: string;
  value: string;
  created_at: number;
  updated_at: number;
}

export interface I18nListParams {
  /** 站点：platform / operator / agent / user */
  i18n_code?: string;
  /** 分组 */
  i18n_group?: string;
  /** 语言 */
  lang?: string;
  /** 翻译键 */
  trans_key?: string;
  /** 页码 */
  page?: number;
  /** 每页条数 */
  page_size?: number;
}

export interface I18nListRes {
  list: I18nRecord[];
  total: number;
}

/** 多语言列表 POST /admin/core/i18n/list */
export function queryI18nList(data: I18nListParams = {}) {
  return axios.post<I18nListRes>('/core/i18n/list', data);
}

export interface CreateI18nParams {
  /** 站点：platform / operator / agent / user */
  i18n_code?: string;
  i18n_group: string;
  lang: string;
  trans_key: string;
  value: string;
}

/** 创建多语言 POST /admin/core/i18n/create */
export function createI18n(data: CreateI18nParams) {
  return axios.post<I18nRecord>('/core/i18n/create', data);
}

export interface UpdateI18nParams {
  id: number;
  i18n_group: string;
  lang: string;
  trans_key: string;
  value: string;
}

/** 更新多语言 POST /admin/core/i18n/update */
export function updateI18n(data: UpdateI18nParams) {
  return axios.post<{ result: string }>('/core/i18n/update', data);
}

export interface DeleteI18nParams {
  /** 单个 ID；若同时传 ids 则忽略本字段 */
  id?: number;
  /** 批量 ID，优先于 id */
  ids?: number[];
}

/** 删除多语言 POST /admin/core/i18n/delete */
export function deleteI18n(data: DeleteI18nParams) {
  return axios.post<{ result: string }>('/core/i18n/delete', data);
}

export interface DeleteI18nByKeyParams {
  /** 站点：platform / operator / agent / user */
  i18n_code: string;
  /** 分组 */
  i18n_group: string;
  /** 词条 key；可为短 key 或完整 key */
  trans_key: string;
}

/** 按 key 删除多语言（删除该 key 下全部语言版本）POST /admin/core/i18n/deleteByKey */
export function deleteI18nByKey(data: DeleteI18nByKeyParams) {
  return axios.post<{ result: string }>('/core/i18n/deleteByKey', data);
}

export interface UpdateI18nByKeyParams {
  /** 翻译键 */
  trans_key: string;
  /** 语言码 -> 译文，如 { zh: '登录', en: 'Login' } */
  data: Record<string, string>;
}

/** 按词条 key 更新多语言 POST /admin/core/i18n/updateByKey */
export function updateI18nByKey(data: UpdateI18nByKeyParams) {
  return axios.post<{ result: string }>('/core/i18n/updateByKey', data);
}

export interface I18nLangRecord {
  id: number;
  lang: string;
  name: string;
  /** 多语言显示名对应的词条 key */
  i18n_key?: string;
  /** 当前语言下的多语言显示名 */
  i18n_name?: string;
  /** 排序号，越小越靠前 */
  sort_no?: number;
  disabled: number;
  is_default: number;
  created_at: number;
  updated_at: number;
}

export interface I18nEnabledLangRes {
  list: I18nLangRecord[];
  total: number;
}

/**
 * 已启用语言列表 GET /admin/core/i18n/lang/enabled
 * 鉴权要求 query 带 operator_code（分站编码），否则返回 400
 */
export function queryEnabledLangs(config?: RequestConfig) {
  return axios.get<I18nEnabledLangRes>('/core/i18n/lang/enabled', config);
}

export interface I18nDictParams {
  /** 站点编码 */
  i18n_code: string;
  /** 分组，空则下发全部组 */
  i18n_group?: string;
  /** 语言，空则从服务端 ctx / X-Lang 取 */
  lang?: string;
}

/** 词条下发：data 为扁平 key -> 文案；兼容文档里的 { items } 包一层 */
export type I18nDictRes =
  | Record<string, string>
  | { items: Record<string, string> };

/** 词条下发 GET /admin/core/i18n/dict */
export function queryI18nDict(params: I18nDictParams, config?: RequestConfig) {
  return axios.get<I18nDictRes>('/core/i18n/dict', { ...config, params });
}

export interface I18nLangListParams {
  /** 页码，默认 1 */
  page?: number;
  /** 每页条数，默认 50 */
  page_size?: number;
  /** 语言码模糊筛选 */
  lang?: string;
  /** 0 启用 / 1 停用；不传不过滤 */
  disabled?: number;
}

export interface I18nLangListRes {
  list: I18nLangRecord[];
  total: number;
}

/** 语言列表 POST /admin/core/i18n/lang/list */
export function queryI18nLangList(data: I18nLangListParams = {}) {
  return axios.post<I18nLangListRes>('/core/i18n/lang/list', data);
}

export interface CreateI18nLangParams {
  lang: string;
  name: string;
  /** 多语言显示名对应的词条 key */
  i18n_key?: string;
  /** 0 启用 / 1 停用，默认 0 */
  disabled?: number;
  /** 0 / 1，设为 1 时会清空其它语言的默认标记 */
  is_default?: number;
}

/** 创建语言 POST /admin/core/i18n/lang/create */
export function createI18nLang(data: CreateI18nLangParams) {
  return axios.post<I18nLangRecord>('/core/i18n/lang/create', data);
}

export interface UpdateI18nLangParams {
  id: number;
  /** 有词条后不可改 */
  lang?: string;
  /** 传了则 trim 后不能为空 */
  name?: string;
  /** 多语言显示名对应的词条 key */
  i18n_key?: string;
  /** 0 启用 / 1 停用；当前默认语言不可停用 */
  disabled?: number;
  /** 0 / 1；不能把当前默认直接改为 0 */
  is_default?: number;
}

/** 更新语言 POST /admin/core/i18n/lang/update */
export function updateI18nLang(data: UpdateI18nLangParams) {
  return axios.post<{ result: string }>('/core/i18n/lang/update', data);
}

export interface DeleteI18nLangParams {
  /** 单个 ID；若同时传 ids 则忽略本字段 */
  id?: number;
  /** 批量 ID，优先于 id */
  ids?: number[];
}

/** 删除语言 POST /admin/core/i18n/lang/delete（不可删默认语言；有词条时也不可删） */
export function deleteI18nLang(data: DeleteI18nLangParams) {
  return axios.post<{ result: string }>('/core/i18n/lang/delete', data);
}

export interface ReorderI18nLangParams {
  /** 需要移动的语言 ID，大于 0 */
  id: number;
  /** 目标语言 ID，大于 0 */
  target_id: number;
}

/** 调整语言排序 POST /admin/core/i18n/lang/reorder */
export function reorderI18nLang(data: ReorderI18nLangParams) {
  return axios.post<{ result: string }>('/core/i18n/lang/reorder', data);
}

export interface ExportI18nParams {
  /** 语言码，必填 */
  lang: string;
  /** 站点精确筛选；空则导出该语言下全部站点 */
  i18n_code?: string;
  /** 分组精确筛选；空则导出该语言/站点下全部分组 */
  i18n_group?: string;
}

/** 导出多语言词条 POST /admin/core/i18n/export（返回原始 JSON 文件，非业务信封） */
export function exportI18n(data: ExportI18nParams) {
  return axios.post<Blob>('/core/i18n/export', data, {
    responseType: 'blob',
    silentError: true,
  } as RequestConfig);
}

export interface ImportI18nResult {
  created: number;
  updated: number;
  skipped: number;
}

/** 导入多语言词条 POST /admin/core/i18n/import（multipart: file, lang） */
export function importI18n(data: FormData) {
  return axios.post<ImportI18nResult>('/core/i18n/import', data);
}
