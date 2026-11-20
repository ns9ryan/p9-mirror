import { DEFAULT_LAYOUT } from '../base';
import { AppRouteRecordRaw } from '../types';

const ACCOUNT_MANAGEMENT: AppRouteRecordRaw = {
  path: '/account-management',
  name: 'accountManagement',
  component: DEFAULT_LAYOUT,
  meta: {
    locale: 'menu.accountManagement',
    requiresAuth: true,
    icon: 'icon-user',
    order: 3,
  },
  children: [
    {
      path: 'user-manage',
      name: 'UserManage',
      component: () => import('@/views/account-management/user-manage/index.vue'),
      meta: {
        locale: 'menu.accountManagement.userManage',
        requiresAuth: true,
        roles: ['*'],
      },
    },
    {
      path: 'userGroup-manage',
      name: 'UserManageGroup',
      component: () => import('@/views/account-management/userGroup-manage/index.vue'),
      meta: {
        locale: 'menu.accountManagement.userGroupManage',
        requiresAuth: true,
        roles: ['*'],
      },
    },
    {
      path: 'webmaster-manage',
      name: 'WebMasterManage',
      component: () => import('@/views/account-management/webmaster-manage/index.vue'),
      meta: {
        locale: 'menu.accountManagement.webmasterManage',
        requiresAuth: true,
        roles: ['*'],
      },
    },
  ],
};
export default ACCOUNT_MANAGEMENT;