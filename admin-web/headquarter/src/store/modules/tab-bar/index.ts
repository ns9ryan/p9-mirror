import type { RouteLocationNormalized, RouteRecordRaw } from 'vue-router';
import { defineStore } from 'pinia';
import {
  DEFAULT_ROUTE,
  DEFAULT_ROUTE_NAME,
  REDIRECT_ROUTE_NAME,
} from '@/router/constants';
import { isString } from '@/utils/is';
import { TabBarState, TagProps } from './types';

const formatTag = (route: RouteLocationNormalized): TagProps => {
  const { name, meta, fullPath, query } = route;
  return {
    title: meta.locale || '',
    name: String(name),
    fullPath,
    query,
    ignoreCache: meta.ignoreCache,
  };
};

const BAN_LIST = [REDIRECT_ROUTE_NAME];

/** 服务端菜单模式下，首页标签取 dashboard 菜单，没有则取第一个可见叶子菜单 */
function findServerHomeTag(menus: RouteRecordRaw[]): TagProps | undefined {
  const leaves: TagProps[] = [];
  const walk = (list: RouteRecordRaw[], parentPath = '') => {
    list.forEach((item) => {
      if (item.meta?.hideInMenu) return;
      const fullPath = item.path.startsWith('/')
        ? item.path
        : `${parentPath}/${item.path}`.replace(/\/+/g, '/');
      if (item.children?.length) {
        walk(item.children, fullPath);
      } else if (typeof item.name === 'string') {
        leaves.push({
          title: String(item.meta?.locale || ''),
          name: item.name,
          fullPath,
          ignoreCache: Boolean(item.meta?.ignoreCache),
        });
      }
    });
  };
  walk(menus);
  return leaves.find((tag) => tag.fullPath.startsWith('/dashboard')) || leaves[0];
}

const useAppStore = defineStore('tabBar', {
  state: (): TabBarState => ({
    cacheTabList: new Set([DEFAULT_ROUTE_NAME]),
    tagList: [DEFAULT_ROUTE],
    homeTag: DEFAULT_ROUTE,
  }),

  getters: {
    getTabList(): TagProps[] {
      return this.tagList;
    },
    getCacheList(): string[] {
      return Array.from(this.cacheTabList);
    },
  },

  actions: {
    updateTabList(route: RouteLocationNormalized) {
      if (BAN_LIST.includes(route.name as string)) return;
      this.tagList.push(formatTag(route));
      if (!route.meta.ignoreCache) {
        this.cacheTabList.add(route.name as string);
      }
    },
    deleteTag(idx: number, tag: TagProps) {
      this.tagList.splice(idx, 1);
      this.cacheTabList.delete(tag.name);
    },
    addCache(name: string) {
      if (isString(name) && name !== '') this.cacheTabList.add(name);
    },
    deleteCache(tag: TagProps) {
      this.cacheTabList.delete(tag.name);
    },
    freshTabList(tags: TagProps[]) {
      this.tagList = tags;
      this.cacheTabList.clear();
      // 要先判断ignoreCache
      this.tagList
        .filter((el) => !el.ignoreCache)
        .map((el) => el.name)
        .forEach((x) => this.cacheTabList.add(x));
    },
    resetTabList() {
      this.tagList = [this.homeTag];
      this.cacheTabList.clear();
      if (!this.homeTag.ignoreCache) this.cacheTabList.add(this.homeTag.name);
    },
    syncHomeTag(menus: RouteRecordRaw[]) {
      const tag = findServerHomeTag(menus);
      if (!tag) return;
      const [prev] = this.tagList;
      this.homeTag = tag;
      if (prev?.name === tag.name && prev.fullPath === tag.fullPath) return;
      if (prev && prev.name !== tag.name) this.cacheTabList.delete(prev.name);
      this.tagList = [
        tag,
        ...this.tagList.slice(1).filter((el) => el.fullPath !== tag.fullPath),
      ];
      if (!tag.ignoreCache) this.cacheTabList.add(tag.name);
    },
  },
});

export default useAppStore;
