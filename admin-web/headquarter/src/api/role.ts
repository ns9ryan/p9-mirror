import axios from 'axios';
import type { AxiosRequestConfig } from 'axios';

export interface RoleListParams {
  page?: number;
  page_size?: number;
  /** 角色名称模糊搜索（若后端支持） */
  role_name?: string;
  /** 角色编码模糊搜索（若后端支持） */
  role_code?: string;
  /** 1 启用 / 2 停用 */
  status?: number;
}

export interface RoleRecord {
  id: number;
  name?: string;
  role_name?: string;
  role_code?: string;
  code?: string;
  display_name?: string;
  description?: string;
  /** 1 启用 / 2 停用 */
  status?: number;
  sort_no?: number;
  /** 系统角色不可删除；内置角色 role_code / is_system 不可改 */
  is_system?: boolean | number;
  updated_by?: string;
  updated_at?: number;
  created_at?: number;
  /** 已绑定菜单 ID（详情接口可能返回） */
  menu_ids?: number[];
}

export interface RoleListRes {
  list: RoleRecord[];
  total: number;
}

/** 角色列表 POST /admin/core/role/list */
export function queryRoleList(
  data: RoleListParams = {},
  config: AxiosRequestConfig & { silentError?: boolean } = {}
) {
  return axios.post<RoleListRes>('/core/role/list', data, config);
}

/** 角色详情 GET /admin/core/role/detail?id= */
export function queryRoleDetail(id: number) {
  return axios.get<RoleRecord>('/core/role/detail', { params: { id } });
}

export interface CreateRoleParams {
  role_name: string;
  role_code: string;
  description?: string;
  /** 1 启用 / 2 停用 */
  status?: number;
  sort_no?: number;
}

/** 新增角色 POST /admin/core/role/create */
export function createRole(data: CreateRoleParams) {
  return axios.post<RoleRecord>('/core/role/create', data);
}

export interface UpdateRoleParams {
  id: number;
  role_name?: string;
  description?: string;
  /** 1 启用 / 2 停用；停用会清理 Casbin */
  status?: number;
  sort_no?: number;
}

/**
 * 更新角色 POST /admin/core/role/update
 * 内置角色的 role_code / is_system 不可改
 */
export function updateRole(data: UpdateRoleParams) {
  return axios.post<null>('/core/role/update', data);
}

export interface DeleteRoleParams {
  /** 单个角色 ID；若同时传 ids 则忽略本字段 */
  id?: number;
  /** 批量角色 ID，优先于 id */
  ids?: number[];
}

/**
 * 删除角色 POST /admin/core/role/delete
 * - is_system 角色不可删
 * - 已绑定用户的角色不可删
 * - 删除时会按 (role_code, v1) 清理 Casbin 权限
 */
export function deleteRole(data: DeleteRoleParams) {
  return axios.post<null>('/core/role/delete', data);
}

export interface SetRoleMenusParams {
  /** 角色主键，必须本租户 */
  role_id: number;
  /** 全量替换该角色菜单；超管角色禁止减菜单 */
  menu_ids: number[];
}

/** 更新角色菜单权限（全量替换） POST /admin/core/authority/menu/update */
export function setRoleMenus(data: SetRoleMenusParams) {
  return axios.post<null>('/core/authority/menu/update', data);
}

export interface RoleMenuIdsParams {
  /** 角色主键，与 id 二选一，优先本字段 */
  role_id?: number;
  /** 角色主键别名，role_id 未传时使用 */
  id?: number;
}

export interface RoleMenuIdsRes {
  /** 该角色已分配（应勾选）的菜单 ID 列表 */
  menu_ids: number[];
}

/** 获取角色已勾选的菜单权限 POST /admin/core/authority/menu/role */
export function queryRoleMenuIds(data: RoleMenuIdsParams) {
  return axios.post<RoleMenuIdsRes>('/core/authority/menu/role', data);
}

export interface RoleApiItem {
  path: string;
  method: string;
}

export interface SetRoleApisParams {
  /** 角色主键，必须本租户 */
  role_id: number;
  /** 全量替换该角色 API；超管角色禁止减 API；后端会合并 is_required */
  data: RoleApiItem[];
}

/** 更新角色接口权限（全量替换） POST /admin/core/authority/api/update */
export function setRoleApis(data: SetRoleApisParams) {
  return axios.post<null>('/core/authority/api/update', data);
}

export interface RoleApiParams {
  /** 角色主键，与 id 二选一，优先本字段 */
  role_id?: number;
  /** 角色主键别名，role_id 未传时使用 */
  id?: number;
}

/** 角色已分配 API 列表（拦截器解包后可能是数组，或带 data/list/apis 的对象） */
export type RoleApiRes =
  | RoleApiItem[]
  | {
      data?: RoleApiItem[];
      list?: RoleApiItem[];
      apis?: RoleApiItem[];
    };

/** 获取角色已勾选的接口权限 POST /admin/core/authority/api/role */
export function queryRoleApis(data: RoleApiParams) {
  return axios.post<RoleApiRes>('/core/authority/api/role', data);
}
