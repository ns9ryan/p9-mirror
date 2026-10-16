import axios from 'axios';

/** 同步状态：0 待同步 / 1 同步成功 / 2 同步失败（后端未提供明确枚举，按常见语义处理） */
export const GAME_SYNC_STATUS = {
  PENDING: 0,
  SUCCESS: 1,
  FAILED: 2,
} as const;

/** 各业务页查询检查点时使用的 sync_scope */
export const GAME_SYNC_SCOPE = {
  CATEGORY: 'CATEGORY',
  CHANNEL: 'CHANNEL',
  CURRENCY: 'CURRENCY',
  PROVIDER: 'PROVIDER',
  GAME: 'GAME',
} as const;

export type GameSyncScope =
  (typeof GAME_SYNC_SCOPE)[keyof typeof GAME_SYNC_SCOPE];

export interface GameSyncCheckpointRecord {
  id: number;
  /** 同步范围 */
  sync_scope: string;
  /** 检查点值 */
  checkpoint_value: string;
  /** 同步进度百分数（0-100，文档可能未列出） */
  progress: number;
  /** 本地总数 */
  local_total: number;
  /** 远程总数 */
  remote_total: number;
  /** 新增数量 */
  created_count: number;
  /** 更新数量 */
  updated_count: number;
  /** 删除数量 */
  deleted_count: number;
  /** 失败数量 */
  failed_count: number;
  /** 同步状态 */
  sync_status: number;
  /** 最后一次错误信息 */
  last_error_message: string;
  /** 最后一次同步时间（秒级时间戳） */
  last_sync_at: number;
  /** 最后一次同步成功时间（秒级时间戳） */
  last_success_at: number;
  created_at: number;
  updated_at: number;
}

export interface GameSyncCheckpointListParams {
  /** 页码，默认 1 */
  page?: number;
  /** 每页条数，默认 15，最大 100 */
  page_size?: number;
  /** 同步范围（可选） */
  sync_scope?: string;
  /** last_success_at 范围过滤 - 开始时间戳（秒，可选） */
  start_time?: number;
  /** last_success_at 范围过滤 - 结束时间戳（秒，可选） */
  end_time?: number;
}

export interface GameSyncCheckpointListRes {
  list: GameSyncCheckpointRecord[];
  total: number;
}

/** 游戏同步检查点列表 GET /admin/game/game-sync-checkpoint/list（走 18003） */
export function queryGameSyncCheckpointList(
  params: GameSyncCheckpointListParams = {}
) {
  return axios.get<GameSyncCheckpointListRes>('/game/game-sync-checkpoint/list', {
    params,
  });
}

/** 游戏同步检查点详情 GET /admin/game/game-sync-checkpoint/get（走 18003） */
export function getGameSyncCheckpointDetail(syncScope: string) {
  return axios.get<GameSyncCheckpointRecord>('/game/game-sync-checkpoint/get', {
    params: { sync_scope: syncScope },
  });
}
