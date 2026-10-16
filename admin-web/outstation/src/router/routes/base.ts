import type { RouteRecordRaw } from 'vue-router';
import { REDIRECT_ROUTE_NAME } from '@/router/constants';

export const DEFAULT_LAYOUT = () => import('@/layout/default-layout.vue');

export const SERVER_MENU_LAYOUT: RouteRecordRaw = {
  path: '/',
  name: 'serverMenuRoot',
  component: DEFAULT_LAYOUT,
  children: [],
  meta: {
    requiresAuth: true,
    hideInMenu: true,
  },
};

export const REDIRECT_MAIN: RouteRecordRaw = {
  path: '/redirect',
  name: 'redirectWrapper',
  component: DEFAULT_LAYOUT,
  meta: {
    requiresAuth: true,
    hideInMenu: true,
  },
  children: [
    {
      path: '/redirect/:path',
      name: REDIRECT_ROUTE_NAME,
      component: () => import('@/views/redirect/index.vue'),
      meta: {
        requiresAuth: true,
        hideInMenu: true,
      },
    },
  ],
};

export const NOT_FOUND_ROUTE: RouteRecordRaw = {
  path: '/:pathMatch(.*)*',
  name: 'notFound',
  component: () => import('@/views/not-found/index.vue'),
};

/** 角色没有分配任何菜单权限时的提示页，独立于业务 Layout 之外 */
export const NO_PERMISSION_ROUTE: RouteRecordRaw = {
  path: '/no-permission',
  name: 'noPermission',
  component: () => import('@/views/no-permission/index.vue'),
  meta: {
    requiresAuth: false,
    hideInMenu: true,
  },
};
