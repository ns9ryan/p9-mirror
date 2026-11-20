export const WHITE_LIST = [
  { name: 'notFound', children: [] },
  { name: 'login', children: [] },
  { name: 'noPermission', children: [] },
  /** 本地 OSS 对接 demo，无需服务端菜单即可访问 /demo/oss */
  { name: 'OssDemo', children: [] },
];

export const NOT_FOUND = {
  name: 'notFound',
};

/** 角色没有分配任何菜单权限时跳转的提示页 */
export const NO_PERMISSION = {
  name: 'noPermission',
};

export const REDIRECT_ROUTE_NAME = 'Redirect';

export const DEFAULT_ROUTE_NAME = 'Workplace';

export const DEFAULT_ROUTE = {
  title: 'menu.dashboard.workplace',
  name: DEFAULT_ROUTE_NAME,
  fullPath: '/dashboard/workplace',
};
