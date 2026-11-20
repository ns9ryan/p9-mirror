const TOKEN_KEY = 'token';
const REFRESH_TOKEN_KEY = 'refresh-token';
const USER_ROLE_KEY = 'user-role';
const USER_NAME_KEY = 'user-name';
const PREVIEW_SESSION_KEY = 'preview-session';

/** 总网预览分站：登录态只放在当前标签页，避免多个预览窗口互相覆盖 */
const isPreviewSession = () => {
  return sessionStorage.getItem(PREVIEW_SESSION_KEY) === '1';
};

const markPreviewSession = () => {
  sessionStorage.setItem(PREVIEW_SESSION_KEY, '1');
};

const authStorage = (): Storage => {
  return isPreviewSession() ? sessionStorage : localStorage;
};

const clearAuthKeys = (storage: Storage) => {
  storage.removeItem(TOKEN_KEY);
  storage.removeItem(REFRESH_TOKEN_KEY);
  storage.removeItem(USER_ROLE_KEY);
  storage.removeItem(USER_NAME_KEY);
};

/** 本标签改为正式登录前调用：丢掉预览态，不碰其他标签页的 localStorage */
const endPreviewSession = () => {
  clearAuthKeys(sessionStorage);
  sessionStorage.removeItem(PREVIEW_SESSION_KEY);
};

const isLogin = () => {
  return !!authStorage().getItem(TOKEN_KEY);
};

const getToken = () => {
  return authStorage().getItem(TOKEN_KEY);
};

const setToken = (token: string) => {
  authStorage().setItem(TOKEN_KEY, token);
};

const getRefreshToken = () => {
  return authStorage().getItem(REFRESH_TOKEN_KEY);
};

const setRefreshToken = (token: string) => {
  authStorage().setItem(REFRESH_TOKEN_KEY, token);
};

const getUserRole = () => {
  return authStorage().getItem(USER_ROLE_KEY) || '';
};

const setUserRole = (role: string) => {
  authStorage().setItem(USER_ROLE_KEY, role);
};

const getUserName = () => {
  return authStorage().getItem(USER_NAME_KEY) || '';
};

const setUserName = (name: string) => {
  authStorage().setItem(USER_NAME_KEY, name);
};

const clearToken = () => {
  clearAuthKeys(authStorage());
};

export {
  isPreviewSession,
  markPreviewSession,
  endPreviewSession,
  isLogin,
  getToken,
  setToken,
  getRefreshToken,
  setRefreshToken,
  getUserRole,
  setUserRole,
  getUserName,
  setUserName,
  clearToken,
};
