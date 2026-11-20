import type { AgentSettings } from './types';

export const DEFAULT_CLIENT_HOST = 'client-33.66isit.com';
export const PROMO_HOSTS = [DEFAULT_CLIENT_HOST, 'client-88.21isit.com'];
export const OPEN_AGENT_LEVELS = [
  { layer: 1, name: 'A1' },
  { layer: 2, name: 'A2' },
  { layer: 3, name: 'A3' },
  { layer: 4, name: 'A4' },
  { layer: 5, name: 'A5' },
  { layer: 6, name: 'A6' },
  { layer: 7, name: 'A7' },
];
export function createDefaultAgentSettings(): AgentSettings {
  return {
    levels: OPEN_AGENT_LEVELS.map((item) => ({ ...item })),
    maxShareRatio: 100,
    applyShareMax: 90,
    applyShareMin: 5,
    defaultLevel: 'A3',
    defaultShareRatio: 90,
    defaultWithdrawLimit: 'none',
  };
}
export const REGISTER_SOURCES = ['官网(PC注册)', '代理(代理添加)'];
export function levelLayer(level: string) {
  const layer = Number(level.replace(/\D/g, ''));
  return Number.isFinite(layer) ? layer : 0;
}

const LEVEL_COLOR: Record<string, string> = {
  A1: '#3b6cff',
  A2: '#3b82f6',
  A3: '#7c6bf2',
  A4: '#f59a23',
  A5: '#e85d8c',
  A6: '#14b8a6',
  A7: '#64748b',
};

const LAYER_COLOR_LIST = Object.values(LEVEL_COLOR);

export function levelColor(level: string, layer = -1) {
  if (LEVEL_COLOR[level]) return LEVEL_COLOR[level];
  if (layer >= 1) {
    return LAYER_COLOR_LIST[(layer - 1) % LAYER_COLOR_LIST.length];
  }
  if (layer === 0) return 'rgb(var(--primary-6))';
  return '#64748b';
}

export const FINANCE_ACCOUNT_TYPES = [
  { value: 'bank', category: 'bank' },
  { value: 'usdtTrc20', category: 'crypto' },
  { value: 'usdtErc20', category: 'crypto' },
  { value: 'ewallet', category: 'ewallet' },
] as const;

export const BANK_CODE_TYPES = ['SWIFT', 'IFSC', 'CLABE', 'IBAN'];

export const WALLET_TYPES = [
  { value: 'deposit', subs: ['online', 'manual'] },
  { value: 'withdraw', subs: ['apply', 'refund'] },
  { value: 'commission', subs: ['settle', 'receive'] },
  { value: 'transfer', subs: ['in', 'out'] },
  { value: 'adjust', subs: ['add', 'deduct'] },
] as const;

export function buildPromoLink(host: string, code: string) {
  const text = code.trim();
  if (!text) return '';
  if (/^https?:\/\//i.test(text)) return text;
  if (!host) return text;
  return `https://${host}?ref=${text}`;
}
