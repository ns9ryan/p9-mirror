import axios from 'axios';

/** 1 启用 / 2 禁用 */
export const GAME_PROVIDER_STATUS = {
  ENABLED: 1,
  DISABLED: 2,
} as const;

export interface GameProviderRecord {
  id: number;
  provider_code: string;
  channel_code: string;
  /** 供应商名称 */
  name: string;
  channel_name: string;
  logo_url: string;
  sort_no: number;
  /** 1 启用 / 2 禁用 */
  status: number;
  created_at: number;
  updated_at: number;
}

export interface GameProviderListParams {
  /** 页码，默认 1 */
  page?: number;
  /** 每页大小，默认 15 */
  page_size?: number;
  /** 供应商编码 */
  provider_code?: string;
  /** 供应商名称 */
  name?: string;
  /** 1 启用 / 2 禁用 */
  status?: number;
}

export interface GameProviderListRes {
  list: GameProviderRecord[];
  total: number;
}

/** 游戏供应商列表 GET /admin/game-provider/list（走 38003） */
export function queryGameProviderList(params: GameProviderListParams = {}) {
  return axios.get<GameProviderListRes>('/game/game-provider/list', { params });
}

/** 游戏供应商详情 GET /admin/game-provider/get（走 38003） */
export function getGameProvider(id: number) {
  return axios.get<GameProviderRecord>('/game/game-provider/get', {
    params: { id },
  });
}

export interface UpdateGameProviderParams {
  id: number;
  /** 排序号 */
  sort_no?: number;
  /** 1 启用 / 2 禁用 */
  status?: number;
}

/** 更新游戏供应商 POST /admin/game-provider/update（走 38003），响应同供应商详情 */
export function updateGameProvider(data: UpdateGameProviderParams) {
  return axios.post<GameProviderRecord>('/game/game-provider/update', data);
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

/** 厂商同步预览 POST /admin/sync/vendors/preview（走 18003） */
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

/** 厂商同步 POST /admin/sync/vendors/run（走 18003） */
export function runGameProviderSync(data: GameProviderRunParams = {}) {
  return axios.post<GameProviderSyncResult>('/game/sync/vendors/run', data);
}

