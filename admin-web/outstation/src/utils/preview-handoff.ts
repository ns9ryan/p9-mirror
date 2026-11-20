import type { LoginRes } from '@/api/user';

export const PREVIEW_READY_TYPE = 'p9-preview-ready';
export const PREVIEW_LOGIN_TYPE = 'p9-preview-login';

let handoffAttempted = false;

/** 向打开本页的总网要预览登录态。同一轮页面加载只等一次 */
export function waitForPreviewHandoff(
  timeout = 5000
): Promise<LoginRes | null> {
  const opener = window.opener as Window | null;
  if (handoffAttempted || !opener) return Promise.resolve(null);
  handoffAttempted = true;

  return new Promise((resolve) => {
    let timer = 0;

    const onMessage = (event: MessageEvent) => {
      if (event.source !== opener) return;
      const { data } = event;
      if (!data || data.type !== PREVIEW_LOGIN_TYPE) return;
      const accessToken = data.token?.access_token;
      if (typeof accessToken !== 'string' || !accessToken) return;
      window.clearTimeout(timer);
      window.removeEventListener('message', onMessage);
      const user = data.user || {};
      resolve({
        token: {
          access_token: accessToken,
          refresh_token:
            typeof data.token.refresh_token === 'string'
              ? data.token.refresh_token
              : '',
          expire: Number(data.token.expire ?? 0),
          refresh_expire: Number(data.token.refresh_expire ?? 0),
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
      });
    };

    timer = window.setTimeout(() => {
      window.removeEventListener('message', onMessage);
      resolve(null);
    }, timeout);

    window.addEventListener('message', onMessage);
    opener.postMessage({ type: PREVIEW_READY_TYPE }, '*');
  });
}
