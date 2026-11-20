import { siteNow } from '@/utils/site-time';
import type { IdType, QuickRange, WinLossRecord } from './types';

export function buildQuickRange(
  key: QuickRange,
  now = siteNow()
): [string, string] {
  const end = now.format('YYYY-MM-DD');
  if (key === 'd30') {
    return [now.subtract(1, 'month').format('YYYY-MM-DD'), end];
  }
  const offset = key === 'd7' ? 6 : 0;
  return [now.subtract(offset, 'day').format('YYYY-MM-DD'), end];
}

export function detectQuickRange(
  range: string[],
  now = siteNow()
): QuickRange | '' {
  if (!range[0] || !range[1]) return '';
  const keys: QuickRange[] = ['today', 'd7', 'd30'];
  return (
    keys.find((key) => {
      const [start, end] = buildQuickRange(key, now);
      return range[0] === start && range[1] === end;
    }) || ''
  );
}

export function matchWinLoss(
  record: WinLossRecord,
  range: string[],
  idType: IdType,
  keyword: string
) {
  if (range[0] && range[1]) {
    if (record.date < range[0] || record.date > range[1]) return false;
  }
  const text = keyword.trim();
  if (!text) return true;
  const ids = idType === 'agentId' ? record.agentIds : record.memberIds;
  return ids.some((id) => id.includes(text));
}
