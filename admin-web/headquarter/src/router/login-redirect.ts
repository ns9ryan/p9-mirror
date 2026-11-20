import type { LocationQuery, RouteLocationRaw, Router } from 'vue-router';
import type { RouteRecordRaw } from 'vue-router';

/** 不可作为登录后跳转目标的路由名 */
export const INVALID_LOGIN_REDIRECT_NAMES = new Set([
  'notFound',
  'login',
  'serverMenuRoot',
  'redirectWrapper',
  'Redirect',
]);

export function findFirstMenuRouteName(
  menus: RouteRecordRaw[]
): string | undefined {
  const menu = menus.find((item) => {
    if (item.children?.length) {
      return Boolean(findFirstMenuRouteName(item.children));
    }
    return typeof item.name === 'string';
  });
  if (!menu) return undefined;
  return menu.children?.length
    ? findFirstMenuRouteName(menu.children)
    : (menu.name as string);
}

/** 从 query 取出单个 redirect 字符串 */
export function getRedirectQuery(
  query: LocationQuery
): string | undefined {
  const raw = query.redirect;
  if (Array.isArray(raw)) {
    return typeof raw[0] === 'string' ? raw[0] : undefined;
  }
  return typeof raw === 'string' ? raw : undefined;
}

/** 是否可作为登录回跳地址（path 或合法路由名） */
export function isUsableLoginRedirect(
  redirect: string | undefined,
  router: Router
): boolean {
  if (!redirect) return false;
  if (INVALID_LOGIN_REDIRECT_NAMES.has(redirect)) return false;
  if (redirect.startsWith('/')) {
    const pathOnly = redirect.split('?')[0]?.split('#')[0] || '/';
    if (pathOnly === '/' || pathOnly === '/login') return false;
    return true;
  }
  return (
    router.hasRoute(redirect) && !INVALID_LOGIN_REDIRECT_NAMES.has(redirect)
  );
}

/**
 * 登录成功后解析跳转目标。
 * redirect 优先为 fullPath（如 /system/role），兼容历史的路由 name。
 */
export function resolvePostLoginTarget(
  router: Router,
  redirect: string | undefined,
  menus: RouteRecordRaw[],
  fallbackName = 'Workplace'
): RouteLocationRaw {
  const firstMenu = findFirstMenuRouteName(menus);

  if (isUsableLoginRedirect(redirect, router) && redirect) {
    if (redirect.startsWith('/')) {
      const resolved = router.resolve(redirect);
      const name = resolved.name ? String(resolved.name) : '';
      if (name && !INVALID_LOGIN_REDIRECT_NAMES.has(name)) {
        return redirect;
      }
      // 动态路由刚注册时 resolve 可能仍是 notFound，直接按 path 推一次即可
      const pathOnly = redirect.split('?')[0] || '';
      if (pathOnly && pathOnly !== '/') {
        return redirect;
      }
    } else if (router.hasRoute(redirect)) {
      return { name: redirect };
    }
  }

  if (firstMenu && router.hasRoute(firstMenu)) {
    return { name: firstMenu };
  }
  if (router.hasRoute(fallbackName)) {
    return { name: fallbackName };
  }
  return { path: '/' };
}

/** 跳登录页时写入的 redirect：保留原 path，丢掉无效目标 */
export function buildLoginRedirect(fullPath: string, name?: string | symbol | null) {
  if (name && INVALID_LOGIN_REDIRECT_NAMES.has(String(name))) {
    // name 无效时仍可能带着真实 path（catch-all 命中前的业务地址）
    const pathOnly = fullPath.split('?')[0]?.split('#')[0] || '';
    if (pathOnly && pathOnly !== '/' && pathOnly !== '/login') {
      return fullPath;
    }
    return undefined;
  }
  if (typeof name === 'string' && name === 'login') {
    return undefined;
  }
  const pathOnly = fullPath.split('?')[0]?.split('#')[0] || '';
  if (!pathOnly || pathOnly === '/login') {
    return undefined;
  }
  return fullPath;
}
