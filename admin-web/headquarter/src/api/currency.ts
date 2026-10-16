import axios from 'axios';

/** 1 启用 / 2 停用 */
export const CURRENCY_STATUS = {
  ENABLED: 1,
  DISABLED: 2,
} as const;

/** 1 法定货币 / 2 虚拟货币 */
export const CURRENCY_TYPE = {
  FIAT: 1,
  VIRTUAL: 2,
} as const;

/** 货币基础数据，无创建接口 */
export interface CurrencyRecord {
  /** 货币 ID */
  id: number;
  /** 货币编码 */
  code: string;
  /** 名称翻译 Key */
  name_key: string;
  /** 当前语言名称（开发阶段调试用） */
  name: string;
  /** 1 法定货币 / 2 虚拟货币 */
  currency_type: number;
  /** 货币符号 */
  symbol: string;
  /** 金额换算倍率，如 USD=100、VND=1 */
  amount_factor: number;
  /** 1 启用 / 2 停用 */
  status: number;
  /** 排序值，数值越小越靠前 */
  sort_no: number;
}

export interface CurrencyListParams {
  /** 页码，从 1 开始，默认 1 */
  page?: number;
  /** 每页数量，范围 1-100，默认 20 */
  page_size?: number;
  /** 1 启用 / 2 停用 */
  status?: number;
}

export interface CurrencyListRes {
  total: number;
  list: CurrencyRecord[];
}

/** 分页货币管理列表 GET /admin/base/currency/list（走 18001） */
export function queryCurrencyList(params: CurrencyListParams = {}) {
  return axios.get<CurrencyListRes>('/base/currency/list', { params });
}

export interface CurrencyListAllParams {
  /** 1 启用 / 2 停用；不传则不筛选 */
  status?: number;
}

export interface CurrencyListAllRes {
  list: CurrencyRecord[];
}

/** 不分页获取全部货币 GET /admin/base/currency/list-all（走 18001） */
export function queryCurrencyListAll(params: CurrencyListAllParams = {}) {
  return axios.get<CurrencyListAllRes>('/base/currency/list-all', { params });
}

export interface CurrencyGetParams {
  /** 货币 ID，大于 0 */
  id: number;
}

/** 根据 ID 获取货币 GET /admin/base/currency/get（走 18001），data 即为 CurrencyInfo */
export function getCurrency(params: CurrencyGetParams) {
  return axios.get<CurrencyRecord>('/base/currency/get', { params });
}

export interface UpdateCurrencyParams {
  /** 货币 ID，大于 0 */
  id: number;
  /** 货币符号，最大 16 个字符；传了则不可为空 */
  symbol?: string;
  /** 1 启用 / 2 停用 */
  status?: number;
}

/**
 * 修改货币 POST /admin/base/currency/update（走 18001）
 * 当前仅允许改 symbol / status，且至少传其中一个；
 * 不可改 code / name_key / currency_type / amount_factor
 */
export function updateCurrency(data: UpdateCurrencyParams) {
  return axios.post<Record<string, never>>('/base/currency/update', data);
}

export interface ReorderCurrencyParams {
  /** 需要移动的货币 ID，大于 0 */
  id: number;
  /** 目标货币 ID，大于 0，且不可与 id 相同 */
  target_id: number;
}

/** 调整货币排序 POST /admin/base/currency/reorder（走 18001） */
export function reorderCurrency(data: ReorderCurrencyParams) {
  return axios.post<Record<string, never>>('/base/currency/reorder', data);
}
