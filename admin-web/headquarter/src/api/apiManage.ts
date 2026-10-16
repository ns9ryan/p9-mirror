import axios from 'axios';

export interface ApiRecord {
  id: number;
  description: string;
  /** 列表展示用（已按当前语言翻译） */
  trans_description?: string;
  api_group: string;
  /** 列表展示用（已按当前语言翻译的分组名） */
  trans_api_group?: string;
  method: string;
  path: string;
  /** 0 否 / 1 是 */
  is_required: number;
  service_name: string;
  created_at: number;
  updated_at: number;
}

export interface ApiListParams {
  page?: number;
  page_size?: number;
  /** 接口说明模糊搜索 */
  description?: string;
  /** 请求路径模糊搜索 */
  path?: string;
  /** 接口分组 */
  api_group?: string;
  /** GET / POST / PUT / PATCH / DELETE */
  method?: string;
  /** 所属服务 */
  service_name?: string;
}

export interface ApiListRes {
  list: ApiRecord[];
  total: number;
}

/** 接口列表 POST /admin/core/api/list */
export function queryApiList(data: ApiListParams = {}) {
  return axios.post<ApiListRes | ApiRecord[]>('/core/api/list', data);
}

export interface CreateApiParams {
  /** GET / POST / PUT / PATCH / DELETE */
  method: string;
  /** 须以 / 开头；与 method 组合唯一 */
  path: string;
  description?: string;
  api_group?: string;
  /** 1 分配权限时必选；默认 0 */
  is_required?: number;
  service_name?: string;
}

/** 新增接口 POST /admin/core/api/create */
export function createApi(data: CreateApiParams) {
  return axios.post<ApiRecord>('/core/api/create', data);
}

export interface UpdateApiParams {
  id: number;
  description?: string;
  api_group?: string;
  /** 变更 method/path 会同步 Casbin */
  method?: string;
  path?: string;
  is_required?: number;
  service_name?: string;
}

/** 更新接口 POST /admin/core/api/update */
export function updateApi(data: UpdateApiParams) {
  return axios.post<null>('/core/api/update', data);
}

export interface DeleteApiParams {
  /** 单个接口 ID；若同时传 ids 则忽略本字段 */
  id?: number;
  /** 批量接口 ID，优先于 id；会同步清理 Casbin */
  ids?: number[];
}

/** 删除接口 POST /admin/core/api/delete */
export function deleteApi(data: DeleteApiParams) {
  return axios.post<null>('/core/api/delete', data);
}
