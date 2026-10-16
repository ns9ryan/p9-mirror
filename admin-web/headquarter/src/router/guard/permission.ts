import type {
  LocationQueryRaw,
  Router,
  RouteRecordNormalized,
  RouteRecordRaw,
} from 'vue-router';
import usePermission from '@/hooks/permission';
import { useUserStore, useAppStore } from '@/store';
import { clearToken, isLogin } from '@/utils/auth';
import {
  loadRemoteDict,
  syncLocaleWithEnabledLangs,
} from '@/utils/locale';
import { buildLoginRedirect } from '@/router/login-redirect';
import { appRoutes } from '../routes';
import { WHITE_LIST, NOT_FOUND, NO_PERMISSION } from '../constants';

/** 并发导航共用一次菜单请求，避免同时打出多次 /menu/role */
let serverMenuRequest: Promise<RouteRecordRaw[]> | null = null;

/**
 * 连续获取菜单失败的次数。用于防止 “notFound -> 重新拉菜单 -> 再次失败 -> notFound”
 * 这种死循环：一旦连续失败超过阈值，即使本地仍残留 token，也强制视为登录已失效，
 * 清空 token 并跳转登录页，避免页面卡死、接口被无限重复请求。
 */
let consecutiveMenuFailures = 0;
const MAX_CONSECUTIVE_MENU_FAILURES = 2;

function findFirstMenuRouteName(
  routes: RouteRecordNormalized[]
): string | undefined {
  const route = routes.find((item) => {
    if (!item.children?.length && typeof item.name === 'string') return true;
    const childRouteName = item.children?.length
      ? findFirstMenuRouteName(
          item.children as unknown as RouteRecordNormalized[]
        )
      : undefined;
    return Boolean(childRouteName);
  });
  if (!route) return undefined;

  return route.children?.length
    ? findFirstMenuRouteName(
        route.children as unknown as RouteRecordNormalized[]
      )
    : (route.name as string);
}

function findMenuNameByPath(
  routes: RouteRecordNormalized[],
  path: string,
  parentFullPath = ''
): string | undefined {
  const normalized = path.replace(/\/+$/, '') || '/';
  let matchedName: string | undefined;

  routes.some((route) => {
    const segment = route.path || '';
    const fullPath = segment.startsWith('/')
      ? segment.replace(/\/+$/, '') || '/'
      : `${parentFullPath}/${segment}`.replace(/\/+/g, '/').replace(/\/+$/, '') ||
        '/';

    if (route.children?.length) {
      const childName = findMenuNameByPath(
        route.children as unknown as RouteRecordNormalized[],
        path,
        fullPath === '/' ? '' : fullPath
      );
      if (childName) {
        matchedName = childName;
        return true;
      }
    }

    if (fullPath === normalized && typeof route.name === 'string') {
      matchedName = route.name;
      return true;
    }
    return false;
  });

  return matchedName;
}

function hasRegisteredServerMenus(
  router: Router,
  menus: RouteRecordNormalized[]
) {
  return menus.some(
    (menu) => typeof menu.name === 'string' && router.hasRoute(menu.name)
  );
}

function registerServerMenus(router: Router, menus: RouteRecordRaw[]) {
  menus.forEach((menu) => {
    if (typeof menu.name === 'string' && router.hasRoute(menu.name)) {
      router.removeRoute(menu.name);
    }
    router.addRoute('serverMenuRoot', menu);
  });
}

async function ensureServerMenus(
  router: Router,
  appStore: ReturnType<typeof useAppStore>
) {
  if (!appStore.serverMenuLoaded) {
    if (!serverMenuRequest) {
      // 菜单标题依赖 X-Lang；先校验本地语言是否仍启用，停用则回退默认语言。
      // 同步拉取远端词条（front.common.* 等），避免页内切语言刷新后只剩本地包、公共键显示成 key。
      serverMenuRequest = syncLocaleWithEnabledLangs()
        .then(async () => {
          await loadRemoteDict();
          return appStore.fetchServerMenuConfig();
        })
        .finally(() => {
          serverMenuRequest = null;
        });
    }
    const menus = await serverMenuRequest;
    registerServerMenus(router, menus);
    return menus;
  }

  const menus = appStore.appAsyncMenus as unknown as RouteRecordRaw[];
  // 刷新后 pinia 可能仍有菜单数据（HMR），但动态路由已丢失，需重新注册
  if (
    menus.length &&
    !hasRegisteredServerMenus(router, appStore.appAsyncMenus)
  ) {
    registerServerMenus(router, menus);
  }
  return menus;
}

export default function setupPermissionGuard(router: Router) {
  router.beforeEach(async (to, from, next) => {
    const appStore = useAppStore();
    const userStore = useUserStore();
    const Permission = usePermission();
    const permissionsAllow = Permission.accessRouter(to);
    if (appStore.menuFromServer) {
      // 针对来自服务端的菜单配置进行处理
      // Handle routing configuration from the server

      // 根据需要自行完善来源于服务端的菜单配置的permission逻辑
      // Refine the permission logic from the server's menu configuration as needed
      if (!appStore.serverMenuLoaded && to.name !== 'login') {
        try {
          const menus = await ensureServerMenus(router, appStore);
          consecutiveMenuFailures = 0;
          // 角色没有分配任何菜单权限：跳转到无权限提示页，而不是 404
          if (!menus.length && to.name !== 'noPermission') {
            next(NO_PERMISSION);
            return;
          }
          if (to.name === 'serverMenuRoot') {
            const targetRouteName = findFirstMenuRouteName(
              menus as unknown as RouteRecordNormalized[]
            );
            if (targetRouteName) {
              next({ name: targetRouteName, replace: true });
              return;
            }
          }
          // 动态路由刚注册时，to 可能仍是 notFound，按 path 重新匹配
          const resolved = router.resolve(to.fullPath);
          if (resolved.name && resolved.name !== 'notFound') {
            next({
              name: resolved.name,
              params: resolved.params,
              query: resolved.query,
              hash: resolved.hash,
              replace: true,
            });
            return;
          }
          const menuName = findMenuNameByPath(
            menus as unknown as RouteRecordNormalized[],
            to.path
          );
          if (menuName && router.hasRoute(menuName)) {
            next({ name: menuName, replace: true });
            return;
          }
          next({ path: to.fullPath, replace: true });
          return;
        } catch {
          consecutiveMenuFailures += 1;
          // 连续多次失败仍落在 notFound 上会不断重新拉菜单形成死循环，
          // 这种情况下即使本地 token 还在，也强制当作登录失效处理。
          const treatAsExpired =
            !isLogin() ||
            consecutiveMenuFailures >= MAX_CONSECUTIVE_MENU_FAILURES;
          // token 过期时保留原 path，避免先落到 notFound 再跳登录丢失 redirect
          if (treatAsExpired) {
            clearToken();
            consecutiveMenuFailures = 0;
            const redirect = buildLoginRedirect(to.fullPath, to.name);
            next({
              name: 'login',
              query: {
                ...to.query,
                ...(redirect ? { redirect } : {}),
              } as LocationQueryRaw,
            });
          } else {
            next(NOT_FOUND);
          }
          return;
        }
      }

      if (appStore.serverMenuLoaded && to.name !== 'login') {
        await ensureServerMenus(router, appStore);
        // 角色没有分配任何菜单权限：跳转到无权限提示页，而不是 404
        if (!appStore.appAsyncMenus.length && to.name !== 'noPermission') {
          next(NO_PERMISSION);
          return;
        }
        // 仍落到 notFound 时，尝试按菜单 path 找回路由名
        if (to.name === 'notFound') {
          const menuName = findMenuNameByPath(appStore.appAsyncMenus, to.path);
          if (menuName && router.hasRoute(menuName)) {
            next({ name: menuName, replace: true });
            return;
          }
        }
      }

      const serverMenuConfig = [...appStore.appAsyncMenus, ...WHITE_LIST];

      let exist = false;
      while (serverMenuConfig.length && !exist) {
        const element = serverMenuConfig.shift();
        if (element?.name === to.name) exist = true;

        if (element?.children) {
          serverMenuConfig.push(
            ...(element.children as unknown as RouteRecordNormalized[])
          );
        }
      }
      if (exist && permissionsAllow) {
        next();
      } else next(NOT_FOUND);
    } else {
      // eslint-disable-next-line no-lonely-if
      if (permissionsAllow) next();
      else {
        const destination =
          Permission.findFirstPermissionRoute(appRoutes, userStore.role) ||
          NOT_FOUND;
        next(destination);
      }
    }
  });
}
