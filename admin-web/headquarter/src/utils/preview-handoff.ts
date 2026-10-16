import type { LoginRes } from '@/api/user';

export const PREVIEW_READY_TYPE = 'p9-preview-ready';
export const PREVIEW_LOGIN_TYPE = 'p9-preview-login';

/** 打开分站页面，等分站就绪后把换票结果交给它，不把 operator_code 放进地址栏 */
export function openOutstationWithSession(session: LoginRes, rawUrl: string) {
  const url = new URL(rawUrl);
  url.searchParams.delete('operator_code');
  const child = window.open(url.toString(), '_blank');
  if (!child) return false;

  const { origin } = url;
  const onMessage = (event: MessageEvent) => {
    if (event.origin !== origin || event.source !== child) return;
    if (event.data?.type !== PREVIEW_READY_TYPE) return;
    child.postMessage(
      {
        type: PREVIEW_LOGIN_TYPE,
        token: session.token,
        user: session.user,
      },
      origin
    );
  };
  window.addEventListener('message', onMessage);
  window.setTimeout(() => {
    window.removeEventListener('message', onMessage);
  }, 10000);
  return true;
}
