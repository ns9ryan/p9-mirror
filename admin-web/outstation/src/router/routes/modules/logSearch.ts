import { DEFAULT_LAYOUT } from '../base';
import { AppRouteRecordRaw } from '../types';

const LOG_SEARCH: AppRouteRecordRaw = {
  path: '/log-search',
  name: 'logSearch',
  component: DEFAULT_LAYOUT,
  meta: {
    locale: 'menu.logSearch',
    requiresAuth: true,
    icon: 'icon-settings',
    order: 3,
  },
  children: [
    {
      path: 'countryCurrencyLanguage',
      name: 'CountryCurrencyLanguage',
      component: () => import('@/views/sys-config/country-currency-language/index.vue'),
      meta: {
        locale: 'menu.sysConfig.countryCurrencyLanguage',
        requiresAuth: true,
        roles: ['*'],
      },
    },
    // {
    //   path: 'userGroup-manage',
    //   name: 'UserManageGroup',
    //   component: () => import('@/views/account-management/userGroup-manage/index.vue'),
    //   meta: {
    //     locale: 'menu.accountManagement.userGroupManage',
    //     requiresAuth: true,
    //     roles: ['*'],
    //   },
    // },
    // {
    //   path: 'webmaster-manage',
    //   name: 'WebMasterManage',
    //   component: () => import('@/views/account-management/webmaster-manage/index.vue'),
    //   meta: {
    //     locale: 'menu.accountManagement.webmasterManage',
    //     requiresAuth: true,
    //     roles: ['*'],
    //   },
    // },
  ],
};
export default LOG_SEARCH;
