import axios from 'axios';
import type { AxiosRequestConfig, AxiosResponse } from 'axios';
import { Message, Modal } from '@arco-design/web-vue';
import { useUserStore } from '@/store';
import {
  clearToken,
  getRefreshToken,
  getToken,
  setRefreshToken,
  setToken,
} from '@/utils/auth';

export interface HttpResponse<T = unknown> {
  status?: number;
  msg: string;
  code: number;
  data?: T;
}

interface TokenPayload {
  access_token: string;
  refresh_token: string;
}

interface AuthRequestConfig extends AxiosRequestConfig {
  skipAuthRefresh?: boolean;
  hasRetried?: boolean;
  /** 该请求失败时不弹出全局错误提示（由调用方自行处理，例如可选的辅助数据加载） */
  silentError?: boolean;
}

let refreshTokenPromise: Promise<string> | undefined;

if (import.meta.env.VITE_API_BASE_URL) {
  axios.defaults.baseURL = import.meta.env.VITE_API_BASE_URL;
}

const refreshAccessToken = async () => {
  const refreshToken = getRefreshToken();
  if (!refreshToken) {
    throw new Error('登录已过期，请重新登录');
  }
  if (!refreshTokenPromise) {
    refreshTokenPromise = axios
      .post<{ token: TokenPayload }>(
        '/core/refresh',
        { refresh_token: refreshToken },
        { skipAuthRefresh: true } as AuthRequestConfig
      )
      .then((res) => {
        const { token } = res.data;
        setToken(token.access_token);
        setRefreshToken(token.refresh_token);
        return token.access_token;
      })
      .finally(() => {
        refreshTokenPromise = undefined;
      });
  }
  return refreshTokenPromise;
};

/**
 * 统一处理 token 失效：无论是业务 code=498 还是真实 HTTP 状态码 498，
 * 都尝试刷新 access_token 后重发原请求；刷新失败则清空本地登录态。
 * 返回 null 表示当前请求不满足重试条件（已重试过 / 已跳过鉴权刷新）。
 */
function retryWithRefreshedToken(config: AxiosRequestConfig | undefined) {
  const requestConfig = config as AuthRequestConfig | undefined;
  if (
    !requestConfig ||
    requestConfig.skipAuthRefresh ||
    requestConfig.hasRetried
  ) {
    return null;
  }
  requestConfig.hasRetried = true;
  return refreshAccessToken()
    .then((accessToken) => {
      requestConfig.headers = requestConfig.headers || {};
      requestConfig.headers.Authorization = `Bearer ${accessToken}`;
      return axios(requestConfig);
    })
    .catch((error) => {
      clearToken();
      return Promise.reject(error);
    });
}

axios.interceptors.request.use(
  (config: AxiosRequestConfig) => {
    // let each request carry token
    // this example using the JWT token
    // Authorization is a custom headers key
    // please modify it according to the actual situation
    const requestConfig = config as AuthRequestConfig;
    const token = requestConfig.skipAuthRefresh ? null : getToken();
    if (!config.headers) {
      config.headers = {};
    }
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    // 后端根据 X-Lang 翻译内置菜单标题/角色名/接口描述及响应 msg，缺省 zh-CN
    config.headers['X-Lang'] = localStorage.getItem('arco-locale') || 'zh-CN';
    return config;
  },
  (error) => {
    // do something
    return Promise.reject(error);
  }
);
// add response interceptors
axios.interceptors.response.use(
  (response: AxiosResponse<HttpResponse>) => {
    // 文件下载等 blob / arraybuffer 响应不是业务信封：
    // 仍按项目惯例解包成 { code, msg, data }，并把 headers 带上供解析文件名
    if (
      response.config.responseType === 'blob' ||
      response.config.responseType === 'arraybuffer' ||
      response.data instanceof Blob
    ) {
      return {
        code: 0,
        msg: 'ok',
        data: response.data,
        headers: response.headers,
      };
    }
    const res = response.data;
    // Existing APIs use 20000; the admin APIs return 0 for success.
    if (![0, 20000].includes(res.code)) {
      if (res.code === 498) {
        const retry = retryWithRefreshedToken(response.config);
        if (retry) return retry as unknown as AxiosResponse<HttpResponse>;
        clearToken();
      }
      const silentError = (response.config as AuthRequestConfig)?.silentError;
      if (!silentError) {
        Message.error({
          content: res.msg || 'Error',
          duration: 5 * 1000,
        });
      }
      // 50008: Illegal token; 50012: Other clients logged in; 50014: Token expired;
      if (
        [50008, 50012, 50014].includes(res.code) &&
        response.config.url !== '/api/user/info'
      ) {
        Modal.error({
          title: 'Confirm logout',
          content:
            'You have been logged out, you can cancel to stay on this page, or log in again',
          okText: 'Re-Login',
          async onOk() {
            const userStore = useUserStore();

            await userStore.logout();
            window.location.reload();
          },
        });
      }
      return Promise.reject(new Error(res.msg || 'Error'));
    }
    return res;
  },
  (error) => {
    // 部分后端直接用真实 HTTP 状态码 498 表示 token 失效/未授权
    // （而不是 HTTP 200 + 响应体 code=498），这里同样走刷新重试逻辑，
    // 否则 token 永远不会被清空，会导致路由守卫反复重试陷入死循环。
    if (error?.response?.status === 498) {
      const retry = retryWithRefreshedToken(error.config);
      if (retry) return retry;
      clearToken();
    }
    const silentError = (error?.config as AuthRequestConfig)?.silentError;
    if (!silentError) {
      Message.error({
        content: error?.response?.data?.msg || error.msg || 'Request Error',
        duration: 5 * 1000,
      });
    }
    return Promise.reject(error);
  }
);
