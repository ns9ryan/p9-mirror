import { getOperatorCodeByDomain } from '@/api/user';

const STORAGE_KEY = 'outstation-operator-code';

let memoryCode = '';
let resolvePromise: Promise<string> | undefined;

function isIpOrLocalhost(host: string) {
  return (
    host === 'localhost' ||
    host.includes(':') ||
    /^\d{1,3}(\.\d{1,3}){3}$/.test(host)
  );
}

/** 读取已缓存的站点编码（内存优先，其次 sessionStorage） */
export function getCachedOperatorCode() {
  if (memoryCode) return memoryCode;
  try {
    return sessionStorage.getItem(STORAGE_KEY) || '';
  } catch {
    return '';
  }
}

/** 写入站点编码缓存（登录成功或域名解析成功时调用） */
export function setCachedOperatorCode(code: string) {
  memoryCode = code || '';
  try {
    if (code) sessionStorage.setItem(STORAGE_KEY, code);
    else sessionStorage.removeItem(STORAGE_KEY);
  } catch {
    // sessionStorage 不可用时仍保留内存值
  }
}

/**
 * 用地址栏域名换站点编码。
 * IP / localhost 直接返回空（后端不接受）；已缓存则复用；并发调用共用一个请求。
 */
export function resolveOperatorCodeFromDomain() {
  const cached = getCachedOperatorCode();
  if (cached) return Promise.resolve(cached);

  const { hostname } = window.location;
  if (isIpOrLocalhost(hostname)) return Promise.resolve('');

  if (!resolvePromise) {
    resolvePromise = getOperatorCodeByDomain({ domain_name: hostname })
      .then(({ data }) => {
        const code = data?.operator_code || '';
        if (code) setCachedOperatorCode(code);
        return code;
      })
      .catch(() => '')
      .finally(() => {
        resolvePromise = undefined;
      });
  }
  return resolvePromise;
}
