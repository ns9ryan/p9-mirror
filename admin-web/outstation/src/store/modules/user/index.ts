import { defineStore } from 'pinia';
import { Message } from '@arco-design/web-vue';
import {
  login as userLogin,
  logout as userLogout,
  previewToken,
  getUserInfo,
  LoginData,
  LoginRes,
} from '@/api/user';
import {
  setToken,
  setRefreshToken,
  clearToken,
  getUserRole,
  setUserRole,
  getUserName,
  setUserName,
  isPreviewSession,
  markPreviewSession,
  endPreviewSession,
} from '@/utils/auth';
import { removeRouteListener } from '@/utils/route-listener';
import { UserState, RoleType } from './types';
import useAppStore from '../app';

/** 换票响应可能是登录结构，也可能只在 data 里带 access_token */
function readPreviewSession(payload: unknown): LoginRes {
  const root =
    payload && typeof payload === 'object'
      ? (payload as Record<string, any>)
      : {};
  const nested =
    root.data && typeof root.data === 'object'
      ? (root.data as Record<string, any>)
      : null;
  const body =
    root.token || root.access_token || root.user ? root : nested || root;
  const tokenObj =
    body.token && typeof body.token === 'object'
      ? (body.token as Record<string, any>)
      : null;
  const accessToken =
    tokenObj?.access_token ||
    body.access_token ||
    (typeof body.token === 'string' ? body.token : '');
  const refreshToken = tokenObj?.refresh_token || body.refresh_token || '';
  const user = body.user || {};
  return {
    token: {
      access_token: accessToken,
      refresh_token: refreshToken,
      expire: Number(tokenObj?.expire ?? body.expire ?? 0),
      refresh_expire: Number(
        tokenObj?.refresh_expire ?? body.refresh_expire ?? 0
      ),
    },
    user: {
      id: Number(user.id ?? 0),
      user_code: String(user.user_code ?? ''),
      username: String(user.username ?? ''),
      display_name: String(user.display_name || user.username || ''),
      operator_id: Number(user.operator_id ?? 0),
      is_super_admin: Boolean(user.is_super_admin),
      status: Number(user.status ?? 1),
      role_codes: Array.isArray(user.role_codes) ? user.role_codes : [],
      home_path: String(user.home_path || '/dashboard/workplace'),
    },
  };
}

const useUserStore = defineStore('user', {
  state: (): UserState => ({
    name: getUserName() || undefined,
    avatar: undefined,
    job: undefined,
    organization: undefined,
    location: undefined,
    email: undefined,
    introduction: undefined,
    personalWebsite: undefined,
    jobName: undefined,
    organizationName: undefined,
    locationName: undefined,
    phone: undefined,
    registrationDate: undefined,
    accountId: undefined,
    certification: undefined,
    role: (getUserRole() as RoleType) || '',
  }),

  getters: {
    userInfo(state: UserState): UserState {
      return { ...state };
    },
  },

  actions: {
    switchRoles() {
      return new Promise((resolve) => {
        this.role = this.role === 'user' ? 'admin' : 'user';
        resolve(this.role);
      });
    },
    // Set user's information
    setInfo(partial: Partial<UserState>) {
      this.$patch(partial);
    },

    // Reset user's information
    resetInfo() {
      this.$reset();
    },

    // Get user's information
    async info() {
      const res = await getUserInfo();

      this.setInfo(res.data);
      if (res.data.name) {
        setUserName(res.data.name);
      }
    },

    applyLoginSession(data: LoginRes) {
      const { token, user } = data;
      if (!token?.access_token) {
        throw new Error('换票未返回登录 token');
      }
      setToken(token.access_token);
      if (token.refresh_token) setRefreshToken(token.refresh_token);
      const roleCodes = user?.role_codes || [];
      const role = (
        user?.is_super_admin ? '*' : roleCodes[0] || ''
      ) as RoleType;
      const name = user?.display_name || user?.username || 'preview';
      this.setInfo({
        accountId: user?.user_code,
        name,
        role,
      });
      setUserRole(role);
      setUserName(name);
    },

    // Login
    async login(loginForm: LoginData) {
      try {
        const res = await userLogin(loginForm);
        endPreviewSession();
        this.applyLoginSession(res.data);
      } catch (err) {
        clearToken();
        throw err;
      }
    },

    /** 总网已换好的预览登录态：先标记，再写入当前标签页的 sessionStorage */
    applyPreviewHandoff(data: LoginRes) {
      markPreviewSession();
      this.applyLoginSession(data);
    },

    /** 用分站代码换登录态 */
    async loginByPreviewToken(operatorCode: string) {
      try {
        const res = await previewToken({ operator_code: operatorCode });
        markPreviewSession();
        this.applyLoginSession(readPreviewSession(res));
      } catch (err) {
        if (isPreviewSession()) clearToken();
        throw err;
      }
    },
    logoutCallBack() {
      const appStore = useAppStore();
      clearToken();
      // 预览标记留在本标签页，退出后不再读到其他标签页的 localStorage 登录态
      this.resetInfo();
      removeRouteListener();
      appStore.clearServerMenu();
    },
    // Logout
    async logout() {
      const previewSession = isPreviewSession();
      try {
        // 预览 token 只能查看，登出接口会被拒绝
        if (!previewSession) await userLogout();
      } catch (err) {
        const error = err as {
          message?: string;
          response?: { data?: { msg?: string } };
        };
        const message = error.response?.data?.msg || error.message || '';
        if (!message.includes('预览')) {
          Message.error(message || '登出失败');
        }
      } finally {
        this.logoutCallBack();
      }
    },
  },
});

export default useUserStore;
