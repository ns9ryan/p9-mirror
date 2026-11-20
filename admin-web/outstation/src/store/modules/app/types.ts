import type { RouteRecordNormalized } from 'vue-router';

export interface AppState {
  theme: string;
  colorWeak: boolean;
  navbar: boolean;
  menu: boolean;
  topMenu: boolean;
  hideMenu: boolean;
  menuCollapse: boolean;
  footer: boolean;
  themeColor: string;
  menuWidth: number;
  globalSettings: boolean;
  device: string;
  tabBar: boolean;
  menuFromServer: boolean;
  serverMenu: RouteRecordNormalized[];
  /** 是否已尝试拉取过服务端菜单（成功或失败都算，用于避免空菜单死循环） */
  serverMenuLoaded: boolean;
  /** 当前用户菜单树中的按钮权限码 */
  serverPermissions: string[];
  /** 菜单管理中已停用的按钮权限码，超级管理员同样不可用 */
  disabledPermissions: string[];
  [key: string]: unknown;
}
