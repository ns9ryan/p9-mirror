import { createRouter, createWebHistory } from 'vue-router';
import NProgress from 'nprogress'; // progress bar
import 'nprogress/nprogress.css';

import { appRoutes } from './routes';
import {
  REDIRECT_MAIN,
  NOT_FOUND_ROUTE,
  NO_PERMISSION_ROUTE,
  SERVER_MENU_LAYOUT,
} from './routes/base';
import createRouteGuard from './guard';
import { AGENT_INIT_ROUTE_NAME } from './constants';

NProgress.configure({ showSpinner: false }); // NProgress Configuration

const router = createRouter({
  history: createWebHistory(),
  routes: [
    SERVER_MENU_LAYOUT,
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/login/index.vue'),
      meta: {
        requiresAuth: false,
      },
    },
    {
      path: '/agent-init',
      name: AGENT_INIT_ROUTE_NAME,
      component: () => import('@/views/agent-init/index.vue'),
      meta: {
        requiresAuth: true,
        hideInMenu: true,
      },
    },
    ...appRoutes,
    REDIRECT_MAIN,
    NO_PERMISSION_ROUTE,
    NOT_FOUND_ROUTE,
  ],
  scrollBehavior() {
    return { top: 0 };
  },
});

createRouteGuard(router);

export default router;
