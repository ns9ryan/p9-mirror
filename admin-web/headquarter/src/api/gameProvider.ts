import axios from 'axios';

/** 1 启用 / 2 停用 */
export const GAME_PROVIDER_STATUS = {
  ENABLED: 1,
  DISABLED: 2,
} as const;

export interface GameProviderRecord {
  id: number;
  provider_code: string;
  /** 厂商名称（多语言 JSON，取 default 展示） */
  name: string;
  logo_url: string;
  sort_no: number;
  /** 1 启用 / 2 停用 */
  status: number;
  source_id: number;
  source_provider_code: string;
  source_name_i18n: string;
  source_logo_url: string;
  source_sort_no: number;
  source_status: number;
  /** 0 未删除 / 1 已删除 */
  is_deleted: number;
  created_at: number;
  updated_at: number;
}

export interface GameProviderListParams {
  /** 页码，默认 1 */
  page?: number;
  /** 每页条数，默认 10，最大 100 */
  page_size?: number;
  /** 厂商编码 */
  provider_code?: string;
  /** 厂商名称（模糊搜索 source_name_i18n.default） */
  name?: string;
  /** 1 启用 / 2 停用 */
  status?: number;
  /** 0 未删除 / 1 已删除，默认 0 */
  is_deleted?: number;
  /** 排序字段：id / sort_no / created_at */
  sort_by?: string;
  /** 排序方向：asc / desc */
  sort_order?: string;
}

export interface GameProviderListRes {
  list: GameProviderRecord[];
  total: number;
}

/** 游戏厂商列表 GET /admin/game/game-provider/list（走 18003） */
export function queryGameProviderList(params: GameProviderListParams = {}) {
  return axios.get<GameProviderListRes>('/game/game-provider/list', { params });
}

export interface UpdateGameProviderParams {
  id: number;
  name: string;
  logo_url: string;
  sort_no: number;
  /** 1 启用 / 2 停用 */
  status: number;
  /** 接口文档示例中出现的扩展字段，默认不传 */
  force_logout?: boolean;
}

/** 更新游戏厂商 POST /admin/game/game-provider/update（走 18003） */
export function updateGameProvider(data: UpdateGameProviderParams) {
  return axios.post<null>('/game/game-provider/update', data);
}

export interface GameProviderSyncApplyResult {
  created: number;
  deleted: number;
  failed: number;
  skipped: number;
  updated: number;
}

export interface GameProviderSyncDiff {
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

export interface GameProviderSyncStats {
  conflict_total: number;
  create_total: number;
  delete_total: number;
  error_total: number;
  local_total: number;
  noop_total: number;
  remote_total: number;
  update_total: number;
}

export interface GameProviderSyncResult {
  apply: GameProviderSyncApplyResult;
  preview: {
    diffs: GameProviderSyncDiff[];
    stats: GameProviderSyncStats;
  };
}

export interface GameProviderPreviewParams {
  /** 页码 */
  page?: number;
  /** 每页条数 */
  page_size?: number;
  /** 是否跳过 */
  is_skip?: boolean;
}

export interface GameProviderPreviewResult {
  diffs: GameProviderSyncDiff[];
  page: number;
  page_size: number;
  total_count: number;
  stats: GameProviderSyncStats;
}

/** 厂商同步预览 POST /admin/game/sync/vendors/preview（走 18003） */
export function previewGameProviderSync(data: GameProviderPreviewParams = {}) {
  return axios.post<GameProviderPreviewResult>('/game/sync/vendors/preview', data);
}

/** 厂商同步可同步字段 */
export type GameProviderSyncCol =
  | 'name'
  | 'sort_no'
  | 'source_provider_code'
  | 'source_id'
  | 'source_name_i18n'
  | 'source_status'
  | 'status';

export interface GameProviderRunParams {
  /** 可选：指定同步字段，可多选 */
  sync_cols?: GameProviderSyncCol[];
}

/** 厂商同步 POST /admin/game/sync/vendors/run（走 18003） */
export function runGameProviderSync(data: GameProviderRunParams = {}) {
  return axios.post<GameProviderSyncResult>('/game/sync/vendors/run', data);
}

