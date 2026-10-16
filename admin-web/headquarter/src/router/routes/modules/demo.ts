import { DEFAULT_LAYOUT } from '../base';
import { AppRouteRecordRaw } from '../types';

const DEMO: AppRouteRecordRaw = {
  path: '/demo',
  name: 'Demo',
  component: DEFAULT_LAYOUT,
  meta: {
    locale: 'menu.demo',
    icon: 'icon-experiment',
    requiresAuth: true,
    order: 99,
  },
  children: [
    {
      path: 'oss',
      name: 'OssDemo',
      component: () => import('@/views/demo/oss/index.vue'),
      meta: {
        locale: 'menu.demo.oss',
        requiresAuth: true,
        roles: ['*'],
      },
    },
  ],
};

export default DEMO;
