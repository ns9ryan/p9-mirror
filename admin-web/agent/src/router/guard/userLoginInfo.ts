import type { RouteRecordRaw, Router, LocationQueryRaw } from 'vue-router';
import NProgress from 'nprogress'; // progress bar
import { Message } from '@arco-design/web-vue';

import { useAppStore, useUserStore } from '@/store';
import { isLogin } from '@/utils/auth';
import { syncLocaleWithEnabledLangs } from '@/utils/locale';
import { waitForPreviewHandoff } from '@/utils/preview-handoff';
import {
  buildLoginRedirect,
  resolvePostLoginTarget,
} from '@/router/login-redirect';

export default function setupUserLoginInfoGuard(router: Router) {
  router.beforeEach(async (to, from, next) => {
    NProgress.start();
    const userStore = useUserStore();
    const handoff = await waitForPreviewHandoff();
    if (handoff) {
      try {
        userStore.applyPreviewHandoff(handoff);
        await syncLocaleWithEnabledLangs();
        const appStore = useAppStore();
        let menus: RouteRecordRaw[] = [];
        if (appStore.menuFromServer) {
          menus = await appStore.fetchServerMenuConfig();
          menus.forEach((menu) => {
            if (typeof menu.name === 'string' && router.hasRoute(menu.name)) {
              router.removeRoute(menu.name);
            }
            router.addRoute('serverMenuRoot', menu);
          });
        }
        next(resolvePostLoginTarget(router, undefined, menus));
      } catch (error) {
        Message.error((error as Error).message || '预览登录失败');
        next({ name: 'login', replace: true });
      }
      return;
    }
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
