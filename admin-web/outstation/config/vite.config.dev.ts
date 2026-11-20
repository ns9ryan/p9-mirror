import { mergeConfig } from 'vite';
import eslint from 'vite-plugin-eslint';
import baseConfig from './vite.config.base';

const DEV_HOST = 'http://192.168.10.220';
// const DEV_HOST = 'http://192.168.0.8';
/** 三方集成等服务（另一台机器）这个是本地的暂时不用了 */
// const P9_HOST = 'http://192.168.0.10';

export default mergeConfig(
  {
    mode: 'development',
    server: {
      host: true,
      open: true,
      // 与总网 headquarter(默认 5173) 错开，便于同时本地联调
      port: 5174,
      // 前端按服务划分前缀 /admin/{服务名}，转发时换回后端原路径
      proxy: {
        // 核心：登录、换票、权限、菜单、i18n、日志
        '/admin/core': {
          target: `${DEV_HOST}:38000`,
          changeOrigin: true,
          xfwd: true,
          rewrite: (path: string) => path.replace('/admin/core', '/core'),
        },
        // 基础数据：地区、币种、时区、语言
        '/admin/base': {
          target: `${DEV_HOST}:18001`,
          changeOrigin: true,
          xfwd: true,
          rewrite: (path: string) => path.replace('/admin/base', '/admin'),
        },
        // 统一图片上传（分站 VIP 图标 / Logo），operator-admin :38010
        '/admin/resource': {
          target: `${DEV_HOST}:38010`,
          changeOrigin: true,
          xfwd: true,
          rewrite: (path: string) => path.replace('/admin/resource', '/resource'),
        },
        // 会员（层级等）。与运营商同前缀但不同服务，必须写在 /admin/operator 之前
        '/admin/operator/member': {
          target: `${DEV_HOST}:38010`,
          changeOrigin: true,
          xfwd: true,
          rewrite: (path: string) =>
            path.replace('/admin/operator/member', '/member'),
        },
        // 活动中心，同上需写在 /admin/operator 之前
        '/admin/operator/activity': {
          target: `${DEV_HOST}:38010`,
          changeOrigin: true,
          xfwd: true,
          rewrite: (path: string) =>
            path.replace('/admin/operator/activity', '/activity'),
        },
        // 任务中心，同上需写在 /admin/operator 之前
        '/admin/operator/task': {
          target: `${DEV_HOST}:38010`,
          changeOrigin: true,
          xfwd: true,
          rewrite: (path: string) =>
            path.replace('/admin/operator/task', '/task'),
        },
        // 本站配置（站点设置页）
        '/admin/site': {
          target: `${DEV_HOST}:38001`,
          // 后端按地址栏域名（Host）识别当前站点，需原样透传，不能被 changeOrigin 改成 IP
          changeOrigin: true,
          xfwd: true,
          rewrite: (path: string) => path.replace('/admin/site', '/operator'),
        },
        // 运营商
        '/admin/operator': {
          target: `${DEV_HOST}:18002`,
          changeOrigin: true,
          xfwd: true,
        },
        // 游戏
        '/admin/game': {
          target: `${DEV_HOST}:38003`,
          changeOrigin: true,
          xfwd: true,
          rewrite: (path: string) =>
            path.replace('/admin/game', '/operator-game'),
        },
        // 报表（端口待后端确认）
        '/admin/report': {
          target: `${DEV_HOST}:38009`,
          changeOrigin: true,
          xfwd: true,
          rewrite: (path: string) => path.replace('/admin/report', ''),
        },
        // 代理。手机区号、初始化状态接口后端不在 /agent 下，必须写在 /admin/agent 之前
        '/admin/agent/phone-calling-codes': {
          target: `${DEV_HOST}:38010`,
          changeOrigin: true,
          xfwd: true,
          rewrite: (path: string) => path.replace('/admin/agent', ''),
        },
        '/admin/agent/init/status': {
          target: `${DEV_HOST}:38010`,
          changeOrigin: true,
          xfwd: true,
          rewrite: (path: string) => path.replace('/admin/agent', ''),
        },
        '/admin/agent': {
          target: `${DEV_HOST}:38010`,
          changeOrigin: true,
          xfwd: true,
          rewrite: (path: string) => path.replace('/admin/agent', '/agent'),
        },
        // AI 聊天（mcp-agent）
        '/admin/chat': {
          target: `${DEV_HOST}:28003`,
          changeOrigin: true,
          xfwd: true,
          rewrite: (path: string) => path.replace('/admin/chat', '/chat'),
        },
        // OSS（docs: Base URL http://127.0.0.1:8020）
        '/p9/oss': {
          target: `${DEV_HOST}:18006`,
          changeOrigin: true,
          xfwd: true,
        },
        // 三方集成（IP 地理等）
        '/p9/integration': {
          target: `${DEV_HOST}:18005`,
          changeOrigin: true,
          xfwd: true,
        },
      },
      fs: {
        strict: true,
      },
    },
    // 每个文件编译前同步跑 ESLint 会拖慢首次加载，默认关闭，需要时用 pnpm dev:lint
    plugins: process.env.DEV_ESLINT
      ? [
          eslint({
            cache: true,
            include: ['src/**/*.ts', 'src/**/*.tsx', 'src/**/*.vue'],
            exclude: ['node_modules'],
          }),
        ]
      : [],
  },
  baseConfig
);
