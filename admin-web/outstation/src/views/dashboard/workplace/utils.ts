import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import timezonePlugin from 'dayjs/plugin/timezone';
import type { MemberTrendPoint, TrendGranularity } from '@/api/report';

dayjs.extend(utc);
dayjs.extend(timezonePlugin);

export const DATE_FMT = 'YYYY-MM-DD';

/** 指定时区的当前时间；未指定时区时退回本机时区 */
export function nowIn(tz?: string) {
  return tz ? dayjs().tz(tz) : dayjs();
}

export type QuickRange = 'today' | 'd7' | 'd30';

export type DateRange = [string, string];

export function buildQuickRange(key: QuickRange, now = dayjs()): DateRange {
  const end = now.format(DATE_FMT);
  const offset = { today: 0, d7: 6, d30: 29 }[key];
  return [now.subtract(offset, 'day').format(DATE_FMT), end];
}

export function detectQuickRange(
  range: string[],
  keys: QuickRange[] = ['today', 'd7', 'd30'],
  now = dayjs()
): QuickRange | '' {
  if (!range[0] || !range[1]) return '';
  return (
    keys.find((key) => {
      const [start, end] = buildQuickRange(key, now);
      return range[0] === start && range[1] === end;
    }) || ''
  );
}

/** 与当前区间等长、紧挨在前面的区间 */
export function previousRange([start, end]: DateRange): DateRange {
  const days = dayjs(end).diff(dayjs(start), 'day') + 1;
  return [
    dayjs(start).subtract(days, 'day').format(DATE_FMT),
    dayjs(start).subtract(1, 'day').format(DATE_FMT),
  ];
}

export function granularityOf([start, end]: DateRange): TrendGranularity {
  return start === end ? 'hour' : 'day';
}

/** 桶 key：按天为 YYYY-MM-DD，按小时为 YYYY-MM-DD HH */
export function bucketKey(date: string, granularity: TrendGranularity) {
  const text = String(date || '').replace('T', ' ');
  return granularity === 'hour' ? text.slice(0, 13) : text.slice(0, 10);
}

export function buildBuckets(
  [start, end]: DateRange,
  granularity: TrendGranularity
) {
  const buckets: string[] = [];
  if (granularity === 'hour') {
    for (let h = 0; h < 24; h += 1) {
      buckets.push(`${start} ${String(h).padStart(2, '0')}`);
    }
    return buckets;
  }
  let cursor = dayjs(start);
  const last = dayjs(end);
  while (!cursor.isAfter(last, 'day')) {
    buckets.push(cursor.format(DATE_FMT));
    cursor = cursor.add(1, 'day');
  }
  return buckets;
}

export function toCountMap(
  points: MemberTrendPoint[],
  granularity: TrendGranularity
) {
  const map = new Map<string, number>();
  points.forEach((item) => {
    const key = bucketKey(item.date, granularity);
    map.set(key, (map.get(key) || 0) + (Number(item.count) || 0));
  });
  return map;
}

export function bucketLabel(bucket: string, granularity: TrendGranularity) {
  return granularity === 'hour' ? `${bucket.slice(11, 13)}:00` : bucket;
}

/** 桶是否还没到（今天之后的日期 / 当前小时之后） */
export function isFutureBucket(
  bucket: string,
  granularity: TrendGranularity,
  now = dayjs()
) {
  if (granularity === 'hour') {
    return bucket > now.format('YYYY-MM-DD HH');
  }
  return bucket > now.format(DATE_FMT);
}

export function formatCount(value?: number | null) {
  if (value == null || Number.isNaN(value)) return '-';
  return Number(value).toLocaleString();
}

export function formatRate(value?: number | null) {
  if (value == null || Number.isNaN(value)) return '-';
  return `${Number.isInteger(value) ? value : value.toFixed(2)}%`;
}

/** 环比百分比；上期为 0 时无法计算返回 null */
export function changeRate(current: number, previous: number) {
  if (!previous) return current ? null : 0;
  return ((current - previous) / previous) * 100;
}

export function formatChange(value: number | null) {
  if (value == null) return '-';
  const fixed = Math.abs(value) >= 100 ? value.toFixed(0) : value.toFixed(2);
  return `${value > 0 ? '+' : ''}${Number(fixed)}%`;
}

export function sum(list: (number | null)[]) {
  return list.reduce<number>((acc, item) => acc + (item || 0), 0);
}
