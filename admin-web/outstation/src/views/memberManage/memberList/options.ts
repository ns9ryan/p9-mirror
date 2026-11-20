export const AUTO_LEVELS = [
  '默认层级',
  '十元层级',
  '百元层级',
  '千元层级',
  '万元层级',
  '十万层级',
  '百万层级',
  '千万层级',
  '三千万层级',
];
export const FIXED_LEVELS = ['大三层级', '黑名单', '测试层级'];
export const REGISTER_SOURCES = ['官网(PC注册)', '官网(H5注册)', '代理推广'];
export const CHANNEL_PRESETS = ['text1'];

export function levelsOf(kind: 'auto' | 'fixed') {
  return kind === 'auto' ? AUTO_LEVELS : FIXED_LEVELS;
}
