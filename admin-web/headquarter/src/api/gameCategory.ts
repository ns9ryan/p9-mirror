import axios from 'axios';

/** 1 启用 / 2 停用 */
export const GAME_CATEGORY_STATUS = {
  ENABLED: 1,
  DISABLED: 2,
} as const;

export interface GameCategoryRecord {
  id: number;
  category_code: string;
  /** 分类名称（多语言 JSON，取 default 展示） */
  name: string;
  sort_no: number;
  /** 1 启用 / 2 停用 */
  status: number;
  source_id: number;
  source_category_code: string;
  source_name_i18n: string;
  source_sort_no: number;
  source_status: number;
  /** 0 未删除 / 1 已删除 */
  is_deleted: number;
  created_at: number;
  updated_at: number;
}

export interface GameCategoryListParams {
  /** 页码，默认 1 */
  page?: number;
  /** 每页条数，默认 10，最大 100 */
  page_size?: number;
  /** 分类编码 */
  category_code?: string;
  /** 分类名称（模糊搜索 source_name_i18n.default） */
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

export interface GameCategoryListRes {
  list: GameCategoryRecord[];
  total: number;
}

/** 游戏分类列表 GET /admin/game/game-category/list（走 18003） */
export function queryGameCategoryList(params: GameCategoryListParams = {}) {
  return axios.get<GameCategoryListRes>('/game/game-category/list', { params });
}

export interface UpdateGameCategoryParams {
  id: number;
  name: string;
  sort_no: number;
  /** 1 启用 / 2 停用 */
  status: number;
  /** 接口文档示例中出现的扩展字段，默认不传 */
  force_logout?: boolean;
}

/** 更新游戏分类 POST /admin/game/game-category/update（走 18003） */
export function updateGameCategory(data: UpdateGameCategoryParams) {
  return axios.post<null>('/game/game-category/update', data);
}

export interface GameCategorySyncApplyResult {
  created: number;
  deleted: number;
  failed: number;
  skipped: number;
  updated: number;
}

export interface GameCategorySyncDiff {
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

export interface GameCategorySyncStats {
  conflict_total: number;
  create_total: number;
  delete_total: number;
  error_total: number;
  local_total: number;
  noop_total: number;
  remote_total: number;
  update_total: number;
}

export interface GameCategorySyncResult {
  apply: GameCategorySyncApplyResult;
  preview: {
    diffs: GameCategorySyncDiff[];
    stats: GameCategorySyncStats;
  };
}

export interface GameCategoryPreviewParams {
  /** 页码 */
  page?: number;
  /** 每页条数 */
  page_size?: number;
  /** 是否跳过 */
  is_skip?: boolean;
}

export interface GameCategoryPreviewResult {
  diffs: GameCategorySyncDiff[];
  page: number;
  page_size: number;
  total_count: number;
  total_page: 0
  stats: GameCategorySyncStats;
}

/** 分类同步预览 POST /admin/game/sync/categories/preview（走 18003） */
export function previewGameCategorySync(data: GameCategoryPreviewParams = {}) {
  return axios.post<GameCategoryPreviewResult>('/game/sync/categories/preview', data);
}

/** 分类同步可同步字段 */
export type GameCategorySyncCol =
  | 'name'
  | 'sort_no'
  | 'source_category_code'
  | 'source_id'
  | 'source_name_i18n'
  | 'source_status'
  | 'status';

export interface GameCategoryRunParams {
  /** 可选：指定同步字段，可多选 */
  sync_cols?: GameCategorySyncCol[];
}

/** 分类同步 POST /admin/game/sync/categories/run（走 18003） */
export function runGameCategorySync(data: GameCategoryRunParams = {}) {
  return axios.post<GameCategorySyncResult>('/game/sync/categories/run', data);
}
