import axios, { type AxiosRequestConfig } from 'axios';
import { getRefreshToken } from '@/utils/auth';
import { UserState } from '@/store/modules/user/types';

type AuthRequestConfig = AxiosRequestConfig & {
  skipAuthRefresh?: boolean;
  silentError?: boolean;
};

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

/** 分站用 URL 中的 operator_code 换票 POST /admin/previewToken */
export function previewToken(data: PreviewTokenData) {
  return axios.post<LoginRes>('/previewToken', data, {
    skipAuthRefresh: true,
  } as AuthRequestConfig);
}

/** 登出 POST /admin/logout：拉黑当前 access_token 与 refresh_token */
export function logout() {
  return axios.post<null>(
    '/core/logout',
    {
      refresh_token: getRefreshToken() || '',
    },
    { silentError: true } as AuthRequestConfig
  );
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

/** 当前角色按钮权限码 GET /admin/user/perm */
export function getUserPerm() {
  return axios.get<UserPermRes>('/core/user/perm');
}

export interface ChangeSelfPasswordParams {
  old_password: string;
  password: string;
}

/** 修改当前用户密码 POST /admin/user/password/self（成功后所有 token 失效） */
export function changeSelfPassword(data: ChangeSelfPasswordParams) {
  return axios.post<null>('/core/user/password/self', data);
}

export interface ChangeUserPasswordParams {
  user_id: number;
  password: string;
}

/** 修改他人密码 POST /admin/user/password（成功后轮换 salt） */
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

/** 用户管理列表 POST /admin/user/list */
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

/** 新增用户 POST /admin/user/create（禁止传 is_super_admin） */
export function createUser(data: CreateUserParams) {
  return axios.post<AdminUserRecord>('/core/user/create', data);
}

/** 用户详情 GET /admin/user/detail?id= */
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
 * 更新用户 POST /admin/user/update
 * operator_id / is_super_admin / salt 不可改
 */
export function updateUser(data: UpdateUserParams) {
  return axios.post<null>('/core/user/update', data);
}

export interface UpdateUserRolesParams {
  user_id: number;
  role_ids: number[];
}

/** 全量替换用户角色 POST /admin/user/roles（须同属分站） */
export function updateUserRoles(data: UpdateUserRolesParams) {
  return axios.post<null>('/core/user/roles', data);
}

export interface DeleteUserParams {
  /** 单个用户 ID；若同时传 ids 则忽略本字段 */
  id?: number;
  /** 批量用户 ID，优先于 id */
  ids?: number[];
}

/** 删除用户 POST /admin/user/delete */
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

/** 更新用户登录 IP 白名单 POST /admin/user/ipWhitelist */
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
