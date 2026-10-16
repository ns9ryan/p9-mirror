import { RouteLocationNormalized, RouteRecordRaw } from 'vue-router';
import { useAppStore, useUserStore } from '@/store';

export default function usePermission() {
  const userStore = useUserStore();
  const appStore = useAppStore();
  return {
    accessRouter(route: RouteLocationNormalized | RouteRecordRaw) {
      return (
        !route.meta?.requiresAuth ||
        !route.meta?.roles ||
        route.meta?.roles?.includes('*') ||
        route.meta?.roles?.includes(userStore.role)
      );
    },
    /**
     * 是否拥有按钮权限码。
     * 菜单里停用的按钮不会放行，超级管理员同样要落在当前仍启用的权限码里。
     */
    hasPermission(code: string) {
      if (!code) return false;
      if (appStore.disabledPermissions?.includes(code)) return false;
      const permissions = appStore.serverPermissions || [];
      if (userStore.role === '*' && !permissions.length) return true;
      return permissions.includes(code);
    },
    findFirstPermissionRoute(_routers: any, role = 'admin') {
      const cloneRouters = [..._routers];
      while (cloneRouters.length) {
        const firstElement = cloneRouters.shift();
        if (
          firstElement?.meta?.roles?.find((el: string[]) => {
            return el.includes('*') || el.includes(role);
          })
        )
          return { name: firstElement.name };
        if (firstElement?.children) {
          cloneRouters.push(...firstElement.children);
        }
      }
      return null;
    },
  };
}
