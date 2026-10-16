import { mergeConfig, type ProxyOptions } from 'vite';
import eslint from 'vite-plugin-eslint';
import baseConfig from './vite.config.base';

const DEV_HOST = 'http://192.168.0.15';
// const DEV_HOST = 'http://192.168.0.8';
/** 三方集成等服务（另一台机器）这个是本地的暂时不用了 */
// const P9_HOST = 'http://192.168.0.10';

const proxyTo = (port: number, host = DEV_HOST): ProxyOptions => ({
  target: `${host}:${port}`,
  changeOrigin: true,
  xfwd: true,
});

/**
 * 与 nginx location 一致，按前缀转发：
 * /admin/core/     → 18000（上游仍是 /core/...）
 * /admin/base/     → 18001（上游仍是 /admin/...）
 * /admin/game/     → 18003（上游仍是 /platform-game/...）
 * /admin/operator/ → 18002（上游路径不变）
 * /admin/dispatch/ → 28001（节点 /admin/node/...，调度任务 /admin/dispatch-task/...）
 */
const rewritePrefix =
  (from: string, to: string) => (requestPath: string) =>
    requestPath.replace(new RegExp(`^${from}`), to);

export default mergeConfig(
  {
    mode: 'development',
    server: {
      host: true,
      open: true,
      proxy: {
        // OSS 本地联调（docs: Base URL http://127.0.0.1:8020，写在通用 /p9 之前）
        '/p9/oss': proxyTo(18006),

        // 三方集成（IP 地理等）→ 192.168.0.10:18005
        '/p9': proxyTo(18005),

        // 总网预览分站换票 POST /admin/previewToken → 38000 的 /core/previewToken
        '/admin/previewToken': {
          ...proxyTo(38000),
          rewrite: rewritePrefix('/admin', '/core'),
        },

        // 节点与调度任务 → 28001
        // /admin/dispatch/node/... → /admin/node/...
        // /admin/dispatch/dispatch-task/... → /admin/dispatch-task/...
        '/admin/dispatch/': {
          ...proxyTo(28001),
          rewrite: rewritePrefix('/admin/dispatch', '/admin'),
        },

        '/admin/core': {
          ...proxyTo(18000),
          rewrite: rewritePrefix('/admin/core', '/core'),
        },
        '/admin/base': {
          ...proxyTo(18001),
          rewrite: rewritePrefix('/admin/base', '/admin'),
        },
        '/admin/game': {
          ...proxyTo(18003),
          rewrite: rewritePrefix('/admin/game', '/platform-game'),
        },
        '/admin/operator': proxyTo(18002),
      },
      fs: {
        strict: true,
      },
    },
    plugins: [
      eslint({
        cache: false,
        include: ['src/**/*.ts', 'src/**/*.tsx', 'src/**/*.vue'],
        exclude: ['node_modules'],
      }),
    ],
  },
  baseConfig
);
