import axios, { type AxiosRequestConfig } from 'axios';

/** 0 目录 / 1 菜单 / 2 按钮 */
export const MENU_TYPE = {
  DIR: 0,
  MENU: 1,
  BUTTON: 2,
} as const;

export interface AdminMenuRecord {
  id: number;
  parent_id: number;
  menu_type: number;
  path?: string;
  name?: string;
  component?: string;
  redirect?: string;
  /** 多语言 key，如 menu.route.dashboard */
  title: string;
  /** 列表/树展示用（已按当前语言翻译） */
  trans_title?: string;
  icon?: string;
  permission?: string;
  hide_menu?: number;
  sort?: number;
  disabled?: number;
  created_at?: number;
  updated_at?: number;
  children?: AdminMenuRecord[];
}

/** 菜单管理列表 POST /admin/menu/list，无请求体 */
export function queryMenuList(
  config: AxiosRequestConfig & { silentError?: boolean } = {}
) {
  return axios.post<AdminMenuRecord[]>('/core/menu/list', undefined, config);
}

export interface SaveMenuParams {
  id?: number;
  parent_id: number;
  menu_type: number;
  path?: string;
  name?: string;
  component?: string;
  redirect?: string;
  title: string;
  icon?: string;
  permission?: string;
  hide_menu?: number;
  sort?: number;
  disabled?: number;
}

/** 新增菜单 POST /admin/menu/create */
export function createMenu(data: SaveMenuParams) {
  return axios.post<AdminMenuRecord>('/core/menu/create', data);
}

/** 更新菜单参数：按 id 部分更新，其余字段均可省略 */
export interface UpdateMenuParams {
  id: number;
  parent_id?: number;
  menu_type?: number;
  path?: string;
  name?: string;
  component?: string;
  redirect?: string;
  title?: string;
  icon?: string;
  permission?: string;
  hide_menu?: number;
  sort?: number;
  disabled?: number;
}

/** 更新菜单 POST /admin/menu/update：按 id 部分更新，禁止挂到自己的子孙 */
export function updateMenu(data: UpdateMenuParams) {
  return axios.post<null>('/core/menu/update', data);
}

export interface DeleteMenuParams {
  id?: number;
  ids?: number[];
}

/** 删除菜单 POST /admin/menu/delete */
export function deleteMenu(data: DeleteMenuParams) {
  return axios.post<null>('/core/menu/delete', data);
}
