import { ref } from 'vue';
import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import timezone from 'dayjs/plugin/timezone';
import { getSiteInfo } from '@/api/siteConfig';

dayjs.extend(utc);
dayjs.extend(timezone);

export const DATE_TIME = 'YYYY-MM-DD HH:mm:ss';

/** 接口时区（IANA）；未加载前按 UTC，不用本机时区 */
const siteTimezone = ref('');

let loadPromise: Promise<void> | null = null;

function isValidTimezone(tz: string) {
  try {
    Intl.DateTimeFormat(undefined, { timeZone: tz });
    return true;
  } catch {
    return false;
  }
}

function activeTimezone() {
  return siteTimezone.value || 'UTC';
}

/** 秒 / 毫秒都兼容；小于 1e12 视为秒 */
function toMs(value: number) {
  return value < 1e12 ? value * 1000 : value;
}

export function getSiteTimezone() {
  return siteTimezone.value;
}

/** 写入本站时区；无效值清空后回退 UTC */
export function setSiteTimezone(tz?: string | null) {
  siteTimezone.value = tz && isValidTimezone(tz) ? tz : '';
}

export function clearSiteTimezone() {
  siteTimezone.value = '';
  loadPromise = null;
}

/**
 * 从 /site/info 拉取本站时区，登录后调用一次即可。
 * 并发调用共用一个请求；已有有效时区则跳过。
 */
export function ensureSiteTimezone() {
  if (siteTimezone.value) return Promise.resolve();
  if (!loadPromise) {
    loadPromise = getSiteInfo()
      .then(({ data }) => {
        const code =
          data?.timezone?.code || data?.site_config?.timezone_code || '';
        setSiteTimezone(code);
      })
      .catch(() => {
        // 拉不到时保持 UTC，避免落到本机时区
      })
      .finally(() => {
        loadPromise = null;
      });
  }
  return loadPromise;
}

/** 当前本站时间（未配置时区时为 UTC） */
export function siteNow() {
  return dayjs().tz(activeTimezone());
}

/**
 * Unix 秒/毫秒 → 本站时区文本。
 * @param empty 空值占位，默认 '--'
 */
export function formatSiteTime(
  value?: number | null,
  format = DATE_TIME,
  empty = '--'
) {
  if (value == null || value === 0) return empty;
  return dayjs(toMs(value)).tz(activeTimezone()).format(format);
}

/** 表格时间展示别名，空值返回 '--' */
export function formatTime(value?: number | null, format = DATE_TIME) {
  return formatSiteTime(value, format);
}

/** 本站时区的时间文本 → Unix 秒 */
export function parseSiteTime(text: string) {
  return dayjs.tz(text, activeTimezone()).unix();
}

/** 把任意时间（字符串、毫秒、Date）转成本站时区的 dayjs 对象 */
export function siteDayjs(value: string | number | Date) {
  return dayjs(value).tz(activeTimezone());
}

/** 本站时区某天 00:00:00 → Unix 秒；date 为 YYYY-MM-DD */
export function siteDayStartSec(date: string) {
  return parseSiteTime(`${date} 00:00:00`);
}

/** 本站时区某天 23:59:59 → Unix 秒（等同于 endOf('day').unix()） */
export function siteDayEndSec(date: string) {
  return parseSiteTime(`${date} 23:59:59`);
}

/** 本站时区某天 00:00:00 → Unix 毫秒 */
export function siteDayStartMs(date: string) {
  return siteDayStartSec(date) * 1000;
}

/** 本站时区某天的次日 00:00:00 → Unix 毫秒，用于左闭右开区间的结束边界 */
export function siteNextDayStartMs(date: string) {
  return siteDayStartMs(dayjs(date).add(1, 'day').format('YYYY-MM-DD'));
}
