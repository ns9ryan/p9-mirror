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

const rewritePrefix =
  (from: string, to: string) => (requestPath: string) =>
    requestPath.replace(new RegExp(`^${from}`), to);

export default mergeConfig(
  {
    mode: 'development',
    server: {
      host: true,
      open: true,
      // 与总网 headquarter(默认 5173) 错开，便于同时本地联调
      port: 5174,
      proxy: {
        // OSS 本地联调（docs: Base URL http://127.0.0.1:8020，写在通用 /p9 之前）
        '/p9/oss': proxyTo(18006),

        // 三方集成（IP 地理等）→ 192.168.0.10:18005
        '/p9': proxyTo(18005),

        // 总网跳分站换票 POST /admin/previewToken → 38000 的 /core/previewToken
        '/admin/previewToken': {
          ...proxyTo(38000),
          rewrite: rewritePrefix('/admin', '/core'),
        },

        '/admin/core': {
          ...proxyTo(38000),
          rewrite: rewritePrefix('/admin/core', '/core'),
        },
        '/admin/base': {
          ...proxyTo(18001),
          rewrite: rewritePrefix('/admin/base', '/admin'),
        },
        // 浏览器请求 /admin/game/...，转发到 38003 时改为 /operator-game/...
        '/admin/game': {
          ...proxyTo(38003),
          rewrite: rewritePrefix('/admin/game', '/operator-game'),
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
