import axios from 'axios';

/** 1 启用 / 2 停用 */
export const GAME_STATUS = {
  ENABLED: 1,
  DISABLED: 2,
} as const;

export interface GameCurrencyInfo {
  currency_id: number;
  /** 币种名称 */
  currency_name: string;
}

export interface GameRecord {
  id: number;
  game_code: string;
  /** 游戏名称 */
  name: string;
  image_url: string;
  category_id: number;
  /** 分类编码，列表接口返回 */
  category_code?: string;
  /** 分类名称，列表若直接返回则优先展示 */
  category_name?: string;
  category_name_i18n?: string;
  provider_id: number;
  /** 厂商编码，列表接口返回 */
  provider_code?: string;
  provider_key: string;
  /** 厂商名称，列表若直接返回则优先展示 */
  provider_name?: string;
  provider_name_i18n?: string;
  channel_id: number;
  /** 渠道编码，列表接口返回 */
  channel_code?: string;
  /** 渠道名称，列表若直接返回则优先展示 */
  channel_name?: string;
  channel_name_i18n?: string;
  game_currency_info: GameCurrencyInfo[];
  sort_no: number;
  /** 是否支持内嵌 */
  supports_embed: boolean;
  /** 是否支持跳转 */
  supports_redirect: boolean;
  /** 1 启用 / 2 停用 */
  status: number;
  source_id: number;
  source_game_code: string;
  source_provider_key: string;
  source_name_i18n: string;
  source_image_url: string;
  source_sort_no: number;
  source_status: number;
  /** 0 未删除 / 1 已删除 */
  is_deleted: number;
  created_at: number;
  updated_at: number;
}

export interface GameListParams {
  /** 页码，默认 1 */
  page?: number;
  /** 每页条数，默认 15，最大 100 */
  page_size?: number;
  /** 游戏编码 */
  game_code?: string;
  /** 游戏名称（模糊搜索） */
  name?: string;
  /** 分类编码 */
  category_code?: string;
  /** 厂商编码 */
  provider_code?: string;
  /** 渠道编码 */
  channel_code?: string;
  /** 1 启用 / 2 停用 */
  status?: number;
  /** 0 未删除 / 1 已删除，默认 0 */
  is_deleted?: number;
  /** 排序字段：id / sort_no / created_at */
  sort_by?: string;
  /** 排序方向：asc / desc */
  sort_order?: string;
}

export interface GameListRes {
  list: GameRecord[];
  total: number;
}

/** 独立游戏列表 GET /admin/game/list（走 18003） */
export function queryGameList(params: GameListParams = {}) {
  return axios.get<GameListRes>('/game/game/list', { params });
}

/** 独立游戏详情 GET /admin/game/get（走 18003） */
export function getGame(id: number) {
  return axios.get<GameRecord>('/game/game/get', { params: { id } });
}

export interface UpdateGameParams {
  id: number;
  name: string;
  image_url: string;
  sort_no: number;
  provider_key: string;
  supports_embed: boolean;
  supports_redirect: boolean;
  /** 1 启用 / 2 停用 */
  status: number;
  /** 接口文档示例中出现的扩展字段，默认不传 */
  force_logout?: boolean;
}

/** 更新独立游戏 POST /admin/game/update（走 18003） */
export function updateGame(data: UpdateGameParams) {
  return axios.post<GameRecord>('/game/game/update', data);
}

export interface GameSyncApplyResult {
  created: number;
  deleted: number;
  failed: number;
  skipped: number;
  updated: number;
}

export interface GameSyncDiff {
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

export interface GameSyncStats {
  conflict_total: number;
  create_total: number;
  delete_total: number;
  error_total: number;
  local_total: number;
  noop_total: number;
  remote_total: number;
  update_total: number;
}

export interface GameSyncResult {
  apply: GameSyncApplyResult;
  preview: {
    diffs: GameSyncDiff[];
    stats: GameSyncStats;
  };
}

export interface GamePreviewParams {
  /** 页码 */
  page?: number;
  /** 每页条数 */
  page_size?: number;
  /** 是否跳过 */
  is_skip?: boolean;
}

export interface GamePreviewResult {
  diffs: GameSyncDiff[];
  page: number;
  page_size: number;
  total_count: number;
  stats: GameSyncStats;
}

/** 游戏同步预览 POST /admin/sync/games/preview（走 18003） */
export function previewGameSync(data: GamePreviewParams = {}) {
  return axios.post<GamePreviewResult>('/game/sync/games/preview', data);
}

/** 游戏同步可同步字段 */
export type GameSyncCol =
  | 'name'
  | 'image_url'
  | 'sort_no'
  | 'provider_key'
  | 'source_game_code'
  | 'source_id'
  | 'source_provider_key'
  | 'source_name_i18n'
  | 'source_image_url'
  | 'source_sort_no'
  | 'source_status'
  | 'status';

export interface GameRunParams {
  /** 可选：指定同步字段，可多选 */
  sync_cols?: GameSyncCol[];
}

/** 游戏同步 POST /admin/sync/games/run（走 18003） */
export function runGameSync(data: GameRunParams = {}) {
  return axios.post<GameSyncResult>('/game/sync/games/run', data);
}
