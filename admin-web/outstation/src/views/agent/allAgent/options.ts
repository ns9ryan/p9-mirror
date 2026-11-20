export const PROMO_HOST = 'client-33.66isit.com';
export const REGISTER_SOURCES = ['用户注册', '全民代理活动'];

export function buildPromoLink(host: string, code: string) {
  return `https://${host}?ref=${code}`;
}
