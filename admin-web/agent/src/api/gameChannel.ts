import axios from 'axios';

/** 1 启用 / 2 禁用 */
export const GAME_CHANNEL_STATUS = {
  ENABLED: 1,
  DISABLED: 2,
} as const;

export interface GameChannelRecord {
  id: number;
  channel_code: string;
  provider_count: number;
  game_count: number;
  /** 渠道名称 */
  name: string;
  sort_no: number;
  load_type: number;
  /** 1 启用 / 2 禁用 */
  status: number;
  created_at: number;
  updated_at: number;
}

export interface GameChannelListParams {
  /** 页码，默认 1 */
  page?: number;
  /** 每页大小，默认 15 */
  page_size?: number;
  /** 渠道编码 */
  channel_code?: string;
  /** 渠道名称 */
  name?: string;
  /** 1 启用 / 2 禁用 */
  status?: number;
}

export interface GameChannelListRes {
  list: GameChannelRecord[];
  total: number;
}

/** 游戏渠道列表 GET /admin/game-channel/list（走 38003） */
export function queryGameChannelList(params: GameChannelListParams = {}) {
  return axios.get<GameChannelListRes>('/game/game-channel/list', { params });
}

/** 游戏渠道详情 GET /admin/game-channel/get（走 38003） */
export function getGameChannel(id: number) {
  return axios.get<GameChannelRecord>('/game/game-channel/get', {
    params: { id },
  });
}

export interface UpdateGameChannelParams {
  id: number;
  /** 排序号 */
  sort_no?: number;
  /** 1 启用 / 2 禁用 */
  status?: number;
}

/** 更新游戏渠道 POST /admin/game-channel/update（走 38003），响应同渠道详情 */
export function updateGameChannel(data: UpdateGameChannelParams) {
  return axios.post<GameChannelRecord>('/game/game-channel/update', data);
}

export interface GameChannelSyncApplyResult {
  created: number;
  deleted: number;
  failed: number;
  skipped: number;
  updated: number;
}

export interface GameChannelSyncDiff {
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

export interface GameChannelSyncStats {
  conflict_total: number;
  create_total: number;
  delete_total: number;
  error_total: number;
  local_total: number;
  noop_total: number;
  remote_total: number;
  update_total: number;
}

export interface GameChannelSyncResult {
  apply: GameChannelSyncApplyResult;
  preview: {
    diffs: GameChannelSyncDiff[];
    stats: GameChannelSyncStats;
  };
}

export interface GameChannelPreviewParams {
  /** 页码 */
  page?: number;
  /** 每页条数 */
  page_size?: number;
  /** 是否跳过 */
  is_skip?: boolean;
}

export interface GameChannelPreviewResult {
  diffs: GameChannelSyncDiff[];
  page: number;
  page_size: number;
  total_count: number;
  stats: GameChannelSyncStats;
}

/** 渠道同步预览 POST /admin/sync/channel/preview（走 18003） */
export function previewGameChannelSync(data: GameChannelPreviewParams = {}) {
  return axios.post<GameChannelPreviewResult>('/game/sync/channel/preview', data);
}

/** 渠道同步可同步字段 */
export type GameChannelSyncCol =
  | 'name'
  | 'sort_no'
  | 'source_channel_code'
  | 'source_id'
  | 'source_name_i18n'
  | 'source_status'
  | 'status';

export interface GameChannelRunParams {
  /** 可选：指定同步字段，可多选 */
  sync_cols?: GameChannelSyncCol[];
}

/** 渠道同步 POST /admin/sync/channel/run（走 18003） */
export function runGameChannelSync(data: GameChannelRunParams = {}) {
  return axios.post<GameChannelSyncResult>('/game/sync/channel/run', data);
}
