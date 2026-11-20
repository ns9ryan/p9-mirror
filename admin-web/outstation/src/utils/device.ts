const UDID_KEY = 'device-udid';

const generateUdid = () => {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) {
    return crypto.randomUUID();
  }
  // crypto.randomUUID 仅在安全上下文（https / localhost）可用，http 内网环境下回退
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
};

let cachedUdid = '';

/** 设备码：首次生成后持久化到 localStorage，退出登录不清除 */
const getUdid = () => {
  if (cachedUdid) return cachedUdid;
  let udid = '';
  try {
    udid = localStorage.getItem(UDID_KEY) || '';
    if (!udid) {
      udid = generateUdid();
      localStorage.setItem(UDID_KEY, udid);
    }
  } catch {
    udid = udid || generateUdid();
  }
  cachedUdid = udid;
  return udid;
};

export interface ParsedUserAgent {
  browser: string;
  os: string;
  mobile: boolean;
}

const BROWSER_RULES: [string, RegExp][] = [
  ['Edge', /Edg(?:e|A|iOS)?\/([\d.]+)/],
  ['Opera', /(?:OPR|Opera)\/([\d.]+)/],
  ['WeChat', /MicroMessenger\/([\d.]+)/],
  ['QQ', /QQBrowser\/([\d.]+)/],
  ['UC', /UCBrowser\/([\d.]+)/],
  ['Firefox', /(?:Firefox|FxiOS)\/([\d.]+)/],
  ['Chrome', /(?:Chrome|CriOS)\/([\d.]+)/],
  ['Safari', /Version\/([\d.]+).*Safari/],
];

const OS_RULES: [string, RegExp][] = [
  ['Windows', /Windows NT ([\d.]+)/],
  ['iOS', /(?:iPhone|iPad|iPod).*?OS ([\d_]+)/],
  ['Android', /Android ([\d.]+)/],
  ['macOS', /Mac OS X ([\d_.]+)/],
  ['Linux', /Linux/],
];

const WINDOWS_VERSIONS: Record<string, string> = {
  '10.0': '10/11',
  '6.3': '8.1',
  '6.2': '8',
  '6.1': '7',
};

const matchRule = (ua: string, rules: [string, RegExp][]) => {
  for (let i = 0; i < rules.length; i += 1) {
    const [name, reg] = rules[i];
    const match = ua.match(reg);
    if (match) return { name, version: (match[1] || '').replace(/_/g, '.') };
  }
  return null;
};

/** 将 User-Agent 解析为「浏览器 + 系统」，无法识别时返回空字段 */
const parseUserAgent = (ua?: string): ParsedUserAgent => {
  if (!ua) return { browser: '', os: '', mobile: false };
  const browser = matchRule(ua, BROWSER_RULES);
  const os = matchRule(ua, OS_RULES);
  let osText = os?.name || '';
  if (os?.name === 'Windows') {
    osText = `Windows ${WINDOWS_VERSIONS[os.version] || os.version}`;
  } else if (os?.version) {
    osText = `${os.name} ${os.version.split('.').slice(0, 2).join('.')}`;
  }
  return {
    browser: browser
      ? `${browser.name} ${browser.version.split('.')[0]}`.trim()
      : '',
    os: osText,
    mobile: /Mobile|Android|iPhone|iPad|iPod/i.test(ua),
  };
};

export { getUdid, parseUserAgent };
export default getUdid;
