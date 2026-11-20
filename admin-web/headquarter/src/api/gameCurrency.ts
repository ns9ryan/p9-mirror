import axios from 'axios';

/** 1 启用 / 2 停用 */
export const GAME_CURRENCY_STATUS = {
  ENABLED: 1,
  DISABLED: 2,
} as const;

export interface GameCurrencyRecord {
  id: number;
  game_id: number;
  currency_id: number;
  game_code: string;
  /** 游戏名称（多语言 JSON，取 default 展示） */
  game_name: string;
  currency_code: string;
  /** 币种名称（多语言 JSON，取 default 展示） */
  currency_name: string;
  /** 1 启用 / 2 停用 */
  status: number;
  source_status: number;
  /** 0 未删除 / 1 已删除 */
  is_deleted: number;
  created_at: number;
  updated_at: number;
}

export interface GameCurrencyListParams {
  /** 页码，默认 1 */
  page?: number;
  /** 每页条数，默认 10，最大 100 */
  page_size?: number;
  /** 游戏 ID 筛选 */
  game_id?: number;
  /** 1 启用 / 2 停用 */
  status?: number;
  /** 0 未删除 / 1 已删除，默认 0 */
  is_deleted?: number;
  /** 排序字段：id / created_at */
  sort_by?: string;
  /** 排序方向：asc / desc */
  sort_order?: string;
}

export interface GameCurrencyListRes {
  list: GameCurrencyRecord[];
  total: number;
}

/** 游戏币种列表 GET /admin/game/game-currency/list（走 18003） */
export function queryGameCurrencyList(params: GameCurrencyListParams = {}) {
  return axios.get<GameCurrencyListRes>('/game/game-currency/list', { params });
}

export interface UpdateGameCurrencyParams {
  id: number;
  /** 1 启用 / 2 停用 */
  status: number;
  /** 接口文档示例中出现的扩展字段，默认不传 */
  force_logout?: boolean;
}

/** 更新游戏币种 POST /admin/game/game-currency/update（走 18003） */
export function updateGameCurrency(data: UpdateGameCurrencyParams) {
  return axios.post<null>('/game/game-currency/update', data);
}

export interface GameCurrencySyncApplyResult {
  created: number;
  deleted: number;
  failed: number;
  skipped: number;
  updated: number;
}

export interface GameCurrencySyncDiff {
  action: string;
  conflict_type: string;
  details: string;
  object_code: string;
  object_id: number;
  object_type: string;
  reason: string;
  remote_code: string;
  remote_id: number;
}

export interface GameCurrencySyncStats {
  conflict_total: number;
  create_total: number;
  delete_total: number;
  error_total: number;
  local_total: number;
  noop_total: number;
  remote_total: number;
  update_total: number;
}

export interface GameCurrencySyncResult {
  apply: GameCurrencySyncApplyResult;
  preview: {
    diffs: GameCurrencySyncDiff[];
    stats: GameCurrencySyncStats;
  };
}

export interface GameCurrencyPreviewParams {
  /** 页码 */
  page?: number;
  /** 每页条数 */
  page_size?: number;
  /** 是否跳过 */
  is_skip?: boolean;
}

export interface GameCurrencyPreviewResult {
  diffs: GameCurrencySyncDiff[];
  page: number;
  page_size: number;
  total_count: number;
  stats: GameCurrencySyncStats;
}

/** 币种同步预览 POST /admin/game/sync/currencies/preview（走 18003） */
export function previewGameCurrencySync(data: GameCurrencyPreviewParams = {}) {
  return axios.post<GameCurrencyPreviewResult>('/game/sync/currencies/preview', data);
}

/** 币种同步可同步字段 */
export type GameCurrencySyncCol =
  | 'source_currency_code'
  | 'source_id'
  | 'source_status'
  | 'status';

export interface GameCurrencyRunParams {
  /** 可选：指定同步字段，可多选 */
  sync_cols?: GameCurrencySyncCol[];
}

/** 币种同步 POST /admin/game/sync/currencies/run（走 18003） */
export function runGameCurrencySync(data: GameCurrencyRunParams = {}) {
  return axios.post<GameCurrencySyncResult>('/game/sync/currencies/run', data);
}
