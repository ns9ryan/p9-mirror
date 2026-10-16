import type { RouteRecordRaw } from 'vue-router';
import type { ServerMenuItem } from '@/api/user';

// group nodes (has children, no own page) must not re-wrap DEFAULT_LAYOUT,
// otherwise the navbar/sider/padding renders twice nested
const BLANK_LAYOUT = () => import('@/layout/blank-layout.vue');

const viewModules = import.meta.glob('../views/**/*.vue');
const componentAliases: Record<string, string> = {
  'dashboard/index': 'dashboard/workplace/index',
  'system/user/index': 'account-management/user-manage/index',
  'system/user': 'account-management/user-manage/index',
  'system/userGroup/index': 'account-management/userGroup-manage/index',
  'system/userGroup': 'account-management/userGroup-manage/index',
  'system/webmaster/index': 'account-management/webmaster-manage/index',
  'system/webmaster': 'account-management/webmaster-manage/index',
  'system/menu/index': 'system/menu/index',
  'system/menu': 'system/menu/index',
  'system/role/index': 'system/role/index',
  'system/role': 'system/role/index',
  'system/apiManage/index': 'system/apiManage/index',
  'system/apiManage': 'system/apiManage/index',
  // 后端菜单 path 为 /system/api，component 常为空时会回退用 path 解析
  'system/api/index': 'system/apiManage/index',
  'system/api': 'system/apiManage/index',
  // 注意：不要把父级 system/log 映射到具体页面。
  // 父菜单有 children 时应走 BLANK_LAYOUT；若映射到 action 页面，
  // 子路由（login/error）URL 会变，但页面仍渲染操作日志。
  'system/log/action/index': 'system/log/action/index',
  'system/log/action': 'system/log/action/index',
  'system/log/login/index': 'system/log/login/index',
  'system/log/login': 'system/log/login/index',
  'system/log/error/index': 'system/log/error/index',
  'system/log/error': 'system/log/error/index',
};

function normalizeKey(path: string) {
  return path
    .replace(/^\/+/, '')
    .replace(/^_+/, '')
    .replace(/\.vue$/, '');
}

function resolveComponent(component?: string, hasChildren = false) {
  if (!component) return hasChildren ? BLANK_LAYOUT : undefined;

  const raw = normalizeKey(component);
  const mapped =
    componentAliases[raw] ||
    componentAliases[component] ||
    componentAliases[`${raw}/index`] ||
    raw;
  const normalizedPath = normalizeKey(mapped);

  const view =
    viewModules[`../views/${normalizedPath}.vue`] ||
    viewModules[`../views/${normalizedPath}/index.vue`];

  return view || (hasChildren ? BLANK_LAYOUT : undefined);
}

function getRoutePath(menu: ServerMenuItem, parentPath?: string) {
  if (menu.path.startsWith('/') || !parentPath) return menu.path;
  return menu.path;
}

// 接口未下发 icon 时按 path/name 关键字兜底，后端补齐后会被 menu.icon 覆盖
// 注意：更具体的关键字要写在前面（find 取第一条命中）
const FALLBACK_ICON_RULES: Array<{ keywords: string[]; icon: string }> = [
  { keywords: ['dashboard'], icon: 'icon-dashboard' },
  { keywords: ['logquery', 'log'], icon: 'icon-file' },
  // 基础数据
  { keywords: ['timezone'], icon: 'icon-clock-circle' },
  { keywords: ['currency'], icon: 'icon-stamp' },
  { keywords: ['region', 'country'], icon: 'icon-location' },
  { keywords: ['basedata', 'base-data', 'base_data'], icon: 'icon-storage' },
  // IP 地理
  { keywords: ['strategy', 'policy'], icon: 'icon-branch' },
  { keywords: ['localip', 'local-ip', 'local_ip', 'ipgeo/local'], icon: 'icon-computer' },
  { keywords: ['ipgeo', 'ipmanage', 'integration'], icon: 'icon-public' },
  // 游戏管理（子项先于父级）
  { keywords: ['category'], icon: 'icon-tags' },
  { keywords: ['channel'], icon: 'icon-share-alt' },
  { keywords: ['manufacturer', 'provider'], icon: 'icon-tool' },
  { keywords: ['indiegame'], icon: 'icon-gift' },
  { keywords: ['synclist'], icon: 'icon-sync' },
  // 分站管理（子项先于父级；含中文 title 兜底）
  { keywords: ['分站列表', 'stationmanage', 'stationlist', 'station/list'], icon: 'icon-list' },
  { keywords: ['域名', 'domain'], icon: 'icon-cloud' },
  { keywords: ['管理员账号', 'webmaster', 'adminaccount', 'admin-account'], icon: 'icon-user' },
  { keywords: ['语言分配', 'langalloc', 'languagealloc', 'language-assign'], icon: 'icon-language' },
  { keywords: ['经营地', 'businessloc', 'business-loc', 'operate-region'], icon: 'icon-location' },
  { keywords: ['代理子线路', '代理线路', 'proxysub', 'proxy-line', 'agentline', 'agent-line'], icon: 'icon-share-alt' },
  { keywords: ['游戏资源', 'gameresource', 'game-resource', 'game_resource'], icon: 'icon-gift' },
  { keywords: ['分站管理', 'outstation'], icon: 'icon-home' },
  { keywords: ['gamemanage', 'game'], icon: 'icon-trophy' },
  { keywords: ['system/user', 'accountmanagement', 'usermanage'], icon: 'icon-user' },
  { keywords: ['system/role', 'role'], icon: 'icon-user-group' },
  { keywords: ['system/menu', 'menu'], icon: 'icon-menu' },
  { keywords: ['system/dept', 'dept'], icon: 'icon-apps' },
  { keywords: ['system/api', 'apimanage'], icon: 'icon-link' },
  { keywords: ['system/log', 'operatelog'], icon: 'icon-file' },
  { keywords: ['system/dict', 'dict'], icon: 'icon-storage' },
  { keywords: ['system/i18n', 'i18n'], icon: 'icon-language' },
  { keywords: ['thirdlogin'], icon: 'icon-lock' },
  { keywords: ['token'], icon: 'icon-safe' },
  { keywords: ['position'], icon: 'icon-idcard' },
  { keywords: ['param'], icon: 'icon-settings' },
  { keywords: ['system'], icon: 'icon-settings' },
];
const DEFAULT_FALLBACK_ICON = 'icon-apps';

function resolveFallbackIcon(menu: ServerMenuItem) {
  // title 一并参与匹配，便于后端 path/name 不规范时用中文菜单名兜底
  const key = `${menu.path || ''}/${menu.name || ''}/${menu.title || ''}`.toLowerCase();
  const rule = FALLBACK_ICON_RULES.find(({ keywords }) =>
    keywords.some((keyword) => key.includes(keyword.toLowerCase()))
  );
  return rule?.icon || DEFAULT_FALLBACK_ICON;
}

function formatMenu(
  menu: ServerMenuItem,
  parentPath?: string
): RouteRecordRaw | null {
  if (menu.menu_type === 2 || !menu.path || !menu.name) return null;

  const children = (menu.children || [])
    .map((item) => formatMenu(item, menu.path || parentPath))
    .filter((item): item is RouteRecordRaw => Boolean(item));

  const hasChildren = children.length > 0;
  const component = resolveComponent(menu.component || menu.path, hasChildren);

  const route: any = {
    path: getRoutePath(menu, parentPath),
    name: menu.name,
    component,
    meta: {
      title: menu.title,
      locale: menu.title,
      icon: menu.icon || resolveFallbackIcon(menu),
      order: menu.sort || 0,
      requiresAuth: true,
      hideInMenu: Boolean(menu.hide_menu),
      permission: menu.permission,
    },
  };

  if (menu.redirect) {
    route.redirect = menu.redirect;
  }

  if (hasChildren) {
    route.children = children;
  }

  return route as RouteRecordRaw;
}

export default function formatServerMenus(
  menus: ServerMenuItem[]
): RouteRecordRaw[] {
  return menus
    .map((menu) => formatMenu(menu))
    .filter((menu): menu is RouteRecordRaw => Boolean(menu));
}
