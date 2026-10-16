import type { Router } from 'vue-router';
import NProgress from 'nprogress'; // progress bar
import { setRouteEmitter } from '@/utils/route-listener';
import setupUserLoginInfoGuard from './userLoginInfo';
import setupPermissionGuard from './permission';

function setupPageGuard(router: Router) {
  router.beforeEach(async (to) => {
    // emit route change
    setRouteEmitter(to);
  });
}

/** 等懒加载组件解析、导航确认后再关掉进度条 */
function setupProgressGuard(router: Router) {
  router.afterEach(() => {
    NProgress.done();
  });
  router.onError(() => {
    NProgress.done();
  });
}

export default function createRouteGuard(router: Router) {
  setupPageGuard(router);
  setupUserLoginInfoGuard(router);
  setupPermissionGuard(router);
  setupProgressGuard(router);
}
