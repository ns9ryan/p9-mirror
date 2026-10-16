import axios from 'axios';
import { getRefreshToken } from '@/utils/auth';
import { UserState } from '@/store/modules/user/types';

export interface LoginData {
  username: string;
  password: string;
  operator_code?: string;
}

export interface LoginUser {
  id: number;
  user_code: string;
  username: string;
  display_name: string;
  operator_id: number;
  is_super_admin: boolean;
  status: number;
  role_codes: string[];
  role_names?: string[];
  home_path: string;
  mobile?: string;
  email?: string;
  created_at?: number;
  last_login_at?: number;
  /** 登录 IP 白名单 */
  ip_whitelist?: string[];
  /** 0 关闭 / 1 开启 */
  ip_whitelist_enabled?: number;
}

export interface LoginRes {
  token: {
    access_token: string;
    refresh_token: string;
    expire: number;
    refresh_expire: number;
  };
  user: LoginUser;
}
export function login(data: LoginData) {
  return axios.post<LoginRes>('/core/login', data);
}

export interface PreviewTokenData {
  operator_code: string;
}

/** 总网预览分站换票 POST /admin/previewToken，body 只传 operator_code */
export function previewToken(data: PreviewTokenData) {
  return axios.post<LoginRes>('/previewToken', data);
}

/** 换票响应可能是登录结构，也可能只在 data 里带 access_token */
export function readPreviewLogin(payload: unknown): LoginRes {
  const root =
    payload && typeof payload === 'object'
      ? (payload as Record<string, any>)
      : {};
  const nested =
    root.data && typeof root.data === 'object'
      ? (root.data as Record<string, any>)
      : null;
  const body =
    root.token || root.access_token || root.user ? root : nested || root;
  const tokenObj =
    body.token && typeof body.token === 'object'
      ? (body.token as Record<string, any>)
      : null;
  const accessToken =
    tokenObj?.access_token ||
    body.access_token ||
    (typeof body.token === 'string' ? body.token : '');
  const refreshToken = tokenObj?.refresh_token || body.refresh_token || '';
  const user = body.user || {};
  return {
    token: {
      access_token: accessToken,
      refresh_token: refreshToken,
      expire: Number(tokenObj?.expire ?? body.expire ?? 0),
      refresh_expire: Number(
        tokenObj?.refresh_expire ?? body.refresh_expire ?? 0
      ),
    },
    user: {
      id: Number(user.id ?? 0),
      user_code: String(user.user_code ?? ''),
      username: String(user.username ?? ''),
      display_name: String(user.display_name || user.username || ''),
      operator_id: Number(user.operator_id ?? 0),
      is_super_admin: Boolean(user.is_super_admin),
      status: Number(user.status ?? 1),
      role_codes: Array.isArray(user.role_codes) ? user.role_codes : [],
      home_path: String(user.home_path || '/dashboard/workplace'),
    },
  };
}

/** 登出 POST /admin/core/logout：拉黑当前 access_token 与 refresh_token */
export function logout() {
  return axios.post<null>('/core/logout', {
    refresh_token: getRefreshToken() || '',
  });
}

export function getUserInfo() {
  return axios.post<UserState>('/api/user/info');
}

export interface ServerMenuItem {
  id: number;
  parent_id: number;
  menu_type: number;
  path: string;
  name: string;
  component: string;
  redirect?: string;
  title: string;
  icon?: string;
  permission?: string;
  hide_menu?: number | boolean;
  /** 1 停用 */
  disabled?: number | boolean;
  sort?: number;
  children?: ServerMenuItem[];
}

export function getMenuList() {
  return axios.get<ServerMenuItem[]>('/core/menu/role');
}

export interface UserPermRes {
  permissions: string[];
}

/** 当前角色按钮权限码 GET /admin/core/user/perm */
export function getUserPerm() {
  return axios.get<UserPermRes>('/core/user/perm');
}

export interface ChangeSelfPasswordParams {
  old_password: string;
  password: string;
}

/** 修改当前用户密码 POST /admin/core/user/password/self（成功后所有 token 失效） */
export function changeSelfPassword(data: ChangeSelfPasswordParams) {
  return axios.post<null>('/core/user/password/self', data);
}

export interface ChangeUserPasswordParams {
  user_id: number;
  password: string;
}

/** 修改他人密码 POST /admin/core/user/password（成功后轮换 salt） */
export function changeUserPassword(data: ChangeUserPasswordParams) {
  return axios.post<null>('/core/user/password', data);
}

export interface UserListParams {
  page?: number;
  page_size?: number;
  /** 后台账号模糊搜索 */
  username?: string;
  /** 后台昵称模糊搜索 */
  display_name?: string;
  /** 绑定的角色 ID */
  role_ids?: number[];
  /** 1 启用 / 2 停用 */
  status?: number;
}

export type AdminUserRecord = LoginUser;

export interface UserListRes {
  list: AdminUserRecord[];
  total: number;
}

/** 用户管理列表 POST /admin/core/user/list */
export function queryUserList(data: UserListParams) {
  return axios.post<UserListRes>('/core/user/list', data);
}

export interface CreateUserParams {
  username: string;
  password: string;
  display_name: string;
  mobile?: string;
  email?: string;
  status?: number;
  role_ids?: number[];
}

/** 新增用户 POST /admin/core/user/create（禁止传 is_super_admin） */
export function createUser(data: CreateUserParams) {
  return axios.post<AdminUserRecord>('/core/user/create', data);
}

/** 用户详情 GET /admin/core/user/detail?id= */
export function queryUserDetail(id: number) {
  return axios.get<AdminUserRecord>('/core/user/detail', { params: { id } });
}

export interface UpdateUserParams {
  id: number;
  display_name?: string;
  mobile?: string;
  email?: string;
  status?: number;
}

/**
 * 更新用户 POST /admin/core/user/update
 * operator_id / is_super_admin / salt 不可改
 */
export function updateUser(data: UpdateUserParams) {
  return axios.post<null>('/core/user/update', data);
}

export interface UpdateUserRolesParams {
  user_id: number;
  role_ids: number[];
}

/** 全量替换用户角色 POST /admin/core/user/roles（须同属分站） */
export function updateUserRoles(data: UpdateUserRolesParams) {
  return axios.post<null>('/core/user/roles', data);
}

export interface DeleteUserParams {
  /** 单个用户 ID；若同时传 ids 则忽略本字段 */
  id?: number;
  /** 批量用户 ID，优先于 id */
  ids?: number[];
}

/** 删除用户 POST /admin/core/user/delete */
export function deleteUser(data: DeleteUserParams) {
  return axios.post<null>('/core/user/delete', data);
}

export interface UpdateUserIpWhitelistParams {
  id: number;
  ip_whitelist: string[];
  /** 0 关闭 / 1 开启 */
  ip_whitelist_enabled: number;
}

export interface UpdateUserIpWhitelistRes {
  result: string;
}

/** 更新用户登录 IP 白名单 POST /admin/core/user/ipWhitelist */
export function updateUserIpWhitelist(data: UpdateUserIpWhitelistParams) {
  return axios.post<UpdateUserIpWhitelistRes>('/core/user/ipWhitelist', data);
}

export type {
  RoleListParams,
  RoleRecord,
  RoleListRes,
  DeleteRoleParams,
} from './role';
export { queryRoleList, deleteRole } from './role';
