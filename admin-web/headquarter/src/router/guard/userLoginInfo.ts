import type { Router, LocationQueryRaw } from 'vue-router';
import NProgress from 'nprogress'; // progress bar

import { useUserStore } from '@/store';
import { isLogin } from '@/utils/auth';
import { buildLoginRedirect } from '@/router/login-redirect';

export default function setupUserLoginInfoGuard(router: Router) {
  router.beforeEach(async (to, from, next) => {
    NProgress.start();
    const userStore = useUserStore();
    if (isLogin()) {
      if (userStore.name) {
        next();
      } else {
        try {
          await userStore.info();
          next();
        } catch (error) {
          await userStore.logout();
          const redirect = buildLoginRedirect(to.fullPath, to.name);
          next({
            name: 'login',
            query: {
              ...to.query,
              ...(redirect ? { redirect } : {}),
            } as LocationQueryRaw,
          });
        }
      }
    } else {
      if (to.name === 'login') {
        next();
        return;
      }
      const redirect = buildLoginRedirect(to.fullPath, to.name);
      next({
        name: 'login',
        query: {
          ...to.query,
          ...(redirect ? { redirect } : {}),
        } as LocationQueryRaw,
      });
    }
  });
}
