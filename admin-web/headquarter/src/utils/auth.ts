const TOKEN_KEY = 'token';
const REFRESH_TOKEN_KEY = 'refresh-token';
const USER_ROLE_KEY = 'user-role';
const USER_NAME_KEY = 'user-name';

const isLogin = () => {
  return !!localStorage.getItem(TOKEN_KEY);
};

const getToken = () => {
  return localStorage.getItem(TOKEN_KEY);
};

const setToken = (token: string) => {
  localStorage.setItem(TOKEN_KEY, token);
};

const getRefreshToken = () => {
  return localStorage.getItem(REFRESH_TOKEN_KEY);
};

const setRefreshToken = (token: string) => {
  localStorage.setItem(REFRESH_TOKEN_KEY, token);
};

const getUserRole = () => {
  return localStorage.getItem(USER_ROLE_KEY) || '';
};

const setUserRole = (role: string) => {
  localStorage.setItem(USER_ROLE_KEY, role);
};

const getUserName = () => {
  return localStorage.getItem(USER_NAME_KEY) || '';
};

const setUserName = (name: string) => {
  localStorage.setItem(USER_NAME_KEY, name);
};

const clearToken = () => {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
  localStorage.removeItem(USER_ROLE_KEY);
  localStorage.removeItem(USER_NAME_KEY);
};

export {
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
