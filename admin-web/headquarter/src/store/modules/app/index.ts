import { defineStore } from 'pinia';
import { Notification } from '@arco-design/web-vue';
import type { NotificationReturn } from '@arco-design/web-vue/es/notification/interface';
import type { RouteRecordRaw, RouteRecordNormalized } from 'vue-router';
import defaultSettings from '@/config/settings.json';
import { queryMenuList, type AdminMenuRecord } from '@/api/menu';
import { getMenuList, getUserPerm, type ServerMenuItem } from '@/api/user';
import formatServerMenus from '@/router/server-menu';
import { AppState } from './types';

function isMenuDisabled(disabled?: number | boolean | string) {
  return disabled === true || Number(disabled) === 1;
}

function asMenuNodes(data: unknown): Array<AdminMenuRecord | ServerMenuItem> {
  if (Array.isArray(data)) return data;
  if (
    data &&
    typeof data === 'object' &&
    Array.isArray((data as { list?: unknown }).list)
  ) {
    return (data as { list: AdminMenuRecord[] }).list;
  }
  return [];
}

/** 只收集按钮节点上的权限码。停用按钮不会进入 enabled。 */
function collectButtonPermissions(
  items: Array<AdminMenuRecord | ServerMenuItem>,
  mode: 'enabled' | 'disabled'
) {
  const codes = new Set<string>();
  const walk = (list: Array<AdminMenuRecord | ServerMenuItem>) => {
    list.forEach((item) => {
      if (Number(item.menu_type) === 2) {
        const code = item.permission?.trim();
        const disabled = isMenuDisabled(item.disabled);
        if (code && (mode === 'disabled' ? disabled : !disabled)) {
          codes.add(code);
        }
      }
      if (item.children?.length) walk(item.children);
    });
  };
  walk(items);
  return codes;
}

const useAppStore = defineStore('app', {
  state: (): AppState => ({ ...defaultSettings }),

  getters: {
    appCurrentSetting(state: AppState): AppState {
      return { ...state };
    },
    appDevice(state: AppState) {
      return state.device;
    },
    appAsyncMenus(state: AppState): RouteRecordNormalized[] {
      return state.serverMenu as unknown as RouteRecordNormalized[];
    },
  },

  actions: {
    // Update app settings
    updateSettings(partial: Partial<AppState>) {
      // @ts-ignore-next-line
      this.$patch(partial);
    },

    // Change theme color
    toggleTheme(dark: boolean) {
      if (dark) {
        this.theme = 'dark';
        document.body.setAttribute('arco-theme', 'dark');
      } else {
        this.theme = 'light';
        document.body.removeAttribute('arco-theme');
      }
    },
    toggleDevice(device: string) {
      this.device = device;
    },
    toggleMenu(value: boolean) {
      this.hideMenu = value;
    },
    async fetchServerMenuConfig(): Promise<RouteRecordRaw[]> {
      let notifyInstance: NotificationReturn | null = null;
      try {
        notifyInstance = Notification.info({
          id: 'menuNotice', // Keep the instance id the same
          content: 'loading',
          closable: true,
        });
        const [menuRes, permRes] = await Promise.all([
          getMenuList(),
          getUserPerm(),
        ]);
        const { data } = menuRes;
        const { data: permData } = permRes;
        let list: ServerMenuItem[] = [];
        if (Array.isArray(data)) {
          list = data;
        } else if (Array.isArray((data as { list?: unknown })?.list)) {
          list = (data as { list: ServerMenuItem[] }).list;
        }
        const menus = formatServerMenus(list);
        this.serverMenu = menus as unknown as RouteRecordNormalized[];
        const fromPerm = Array.isArray(permData?.permissions)
          ? permData.permissions.filter((code) => !!code?.trim())
          : [];
        const enabledButtons = collectButtonPermissions(list, 'enabled');
        const disabledButtons = collectButtonPermissions(list, 'disabled');
        try {
          const { data: allMenus } = await queryMenuList({ silentError: true });
          collectButtonPermissions(asMenuNodes(allMenus), 'disabled').forEach(
            (code) => {
              disabledButtons.add(code);
              enabledButtons.delete(code);
            }
          );
        } catch {
          // 无菜单管理权限时忽略，仍使用 /menu/role 里的按钮状态
        }
        const allowed = new Set(enabledButtons);
        disabledButtons.forEach((code) => allowed.delete(code));
        let permissions = fromPerm;
        if (enabledButtons.size) {
          permissions = fromPerm.length
            ? fromPerm.filter((code) => allowed.has(code))
            : [...allowed];
        } else {
          permissions = fromPerm.filter((code) => !disabledButtons.has(code));
        }
        this.serverPermissions = permissions;
        this.disabledPermissions = [...disabledButtons];
        this.serverMenuLoaded = true;
        notifyInstance = Notification.success({
          id: 'menuNotice',
          content: 'success',
          closable: true,
        });
        return menus;
      } catch (error) {
        // 失败不标记 loaded，允许下次导航重试，避免永久 404
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        notifyInstance = Notification.error({
          id: 'menuNotice',
          content: 'error',
          closable: true,
        });
        throw error;
      }
    },
    setMenuPermissionDisabled(code: string, disabled: boolean) {
      const current = new Set(this.disabledPermissions || []);
      const permissions = new Set(this.serverPermissions || []);
      if (disabled) {
        current.add(code);
        permissions.delete(code);
      } else {
        current.delete(code);
        permissions.add(code);
      }
      this.disabledPermissions = [...current];
      this.serverPermissions = [...permissions];
    },
    clearServerMenu() {
      this.serverMenu = [];
      this.serverPermissions = [];
      this.disabledPermissions = [];
      this.serverMenuLoaded = false;
    },
  },
});

export default useAppStore;
