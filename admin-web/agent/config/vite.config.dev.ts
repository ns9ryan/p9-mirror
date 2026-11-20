import { mergeConfig, type ProxyOptions } from 'vite';
import eslint from 'vite-plugin-eslint';
import baseConfig from './vite.config.base';

/** 代理管理平台后端 */
const DEV_TARGET = 'http://192.168.0.18:38010';

const proxyToBackend = (): ProxyOptions => ({
  target: DEV_TARGET,
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
      // 与总网 headquarter(5173)、分站 outstation(5174) 错开
      port: 5175,
      proxy: {
        '/p9/oss': proxyToBackend(),
        '/p9': proxyToBackend(),

        '/admin/previewToken': {
          ...proxyToBackend(),
          rewrite: rewritePrefix('/admin', ''),
        },
        // 浏览器 /admin/core/i18n/dict → 38010 的 /i18n/dict
        '/admin/core': {
          ...proxyToBackend(),
          rewrite: rewritePrefix('/admin/core', ''),
        },
        '/admin/base': {
          ...proxyToBackend(),
          rewrite: rewritePrefix('/admin/base', '/admin'),
        },
        '/admin/game': {
          ...proxyToBackend(),
          rewrite: rewritePrefix('/admin/game', '/operator-game'),
        },
        '/admin/operator': proxyToBackend(),
        '/admin': proxyToBackend(),
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
