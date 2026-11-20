import dayjs from 'dayjs';
import {
  siteDayStartMs,
  siteNextDayStartMs,
  siteNow,
} from '@/utils/site-time';
import type { AllAgentQuery, AllAgentRecord, QuickRange } from './types';

export function createAllAgentQuery(): AllAgentQuery {
  return {
    timeType: 'register',
    dateRange: [],
    keyword: '',
    registerSource: '',
    subCountMin: undefined,
    subCountMax: undefined,
    withdrawLimit: '',
    status: '',
  };
}

export function buildQuickRange(
  key: QuickRange,
  now = siteNow()
): [string, string] {
  const end = now.format('YYYY-MM-DD');
  const offsetByKey: Record<QuickRange, number> = {
    today: 0,
    d7: 6,
    d30: 29,
  };
  return [now.subtract(offsetByKey[key], 'day').format('YYYY-MM-DD'), end];
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

function inDateRange(value: string, range: string[]) {
  if (!range[0] || !range[1]) return true;
  if (!value) return false;
  // 按分站时区的整天区间，左闭右开
  const time = dayjs(value).valueOf();
  return (
    time >= siteDayStartMs(range[0]) && time < siteNextDayStartMs(range[1])
  );
}

export function matchAllAgent(record: AllAgentRecord, query: AllAgentQuery) {
  const time =
    query.timeType === 'login' ? record.lastLoginAt : record.registeredAt;
  if (!inDateRange(time, query.dateRange)) return false;

  const keyword = query.keyword.trim();
  if (keyword && !String(record.id).includes(keyword)) return false;

  if (query.registerSource && record.registerSource !== query.registerSource) {
    return false;
  }
  if (query.subCountMin != null && record.subCount < query.subCountMin) {
    return false;
  }
  if (query.subCountMax != null && record.subCount > query.subCountMax) {
    return false;
  }
  if (query.withdrawLimit && record.withdrawLimit !== query.withdrawLimit) {
    return false;
  }
  if (query.status && record.status !== query.status) return false;
  return true;
}
