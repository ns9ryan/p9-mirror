import dayjs from 'dayjs';
import { parseSiteTime, siteNow } from '@/utils/site-time';
import {
  MEMBER_SORT_BY,
  MEMBER_SORT_ORDER,
  MEMBER_STATUS,
  MEMBER_TIME_TYPE,
  type MemberListParams,
} from '@/api/member';
import type {
  AccountStatus,
  MemberBriefQuery,
  MemberQuery,
  MemberSortKey,
  QuickRange,
  SortOrder,
} from './types';

type FilterParams = Omit<MemberListParams, 'page' | 'page_size'>;

const STATUS_CODE: Record<AccountStatus, number> = {
  normal: MEMBER_STATUS.NORMAL,
  frozen: MEMBER_STATUS.FROZEN,
  disabled: MEMBER_STATUS.DISABLED,
};

const SORT_BY: Record<MemberSortKey, number> = {
  registeredAt: MEMBER_SORT_BY.REGISTERED_AT,
  totalDeposit: MEMBER_SORT_BY.TOTAL_DEPOSIT,
  depositCount: MEMBER_SORT_BY.DEPOSIT_COUNT,
  totalWithdraw: MEMBER_SORT_BY.TOTAL_WITHDRAWAL,
  withdrawCount: MEMBER_SORT_BY.WITHDRAWAL_COUNT,
};

export function createMemberQuery(): MemberQuery {
  return {
    timeType: 'register',
    dateRange: [],
    keyword: '',
    autoLevelId: '',
    fixedLevelId: '',
    registerChannel: '',
    registerMethod: '',
    status: '',
    betBanned: false,
    noPhone: false,
    noEmail: false,
    deposited: false,
    notDeposited: false,
  };
}

/** 快捷日期按分站时区的"今天"计算 */
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

function optionalText(value: string) {
  const text = value.trim();
  return text || undefined;
}

/** 厅时区某一天 00:00 的 Unix 毫秒（parseSiteTime 返回秒，需换算） */
function siteDayStartMs(date: string) {
  return parseSiteTime(`${date} 00:00:00`) * 1000;
}

/** 日期按分站时区整天计算；接口区间左闭右开，结束日取次日 0 点 */
function timeParams(query: MemberQuery): FilterParams {
  const [start, end] = query.dateRange || [];
  if (!start && !end) return {};
  return {
    time_type:
      query.timeType === 'login'
        ? MEMBER_TIME_TYPE.LOGIN
        : MEMBER_TIME_TYPE.REGISTER,
    start_at: start ? siteDayStartMs(start) : undefined,
    end_at: end
      ? siteDayStartMs(dayjs(end).add(1, 'day').format('YYYY-MM-DD'))
      : undefined,
  };
}

export function buildMemberParams(query: MemberQuery): FilterParams {
  return {
    ...timeParams(query),
    keyword: optionalText(query.keyword),
    auto_level_id: query.autoLevelId || undefined,
    fixed_level_id: query.fixedLevelId || undefined,
    register_channel: optionalText(query.registerChannel),
    register_method: optionalText(query.registerMethod),
    status: query.status ? STATUS_CODE[query.status] : undefined,
    betting_disabled: query.betBanned || undefined,
    no_phone: query.noPhone || undefined,
    no_email: query.noEmail || undefined,
    deposited: query.deposited || undefined,
    not_deposited: query.notDeposited || undefined,
  };
}

export function createMemberBriefQuery(): MemberBriefQuery {
  return {
    keyword: '',
    status: '',
    agentId: '',
    registerChannel: '',
    registerMethod: '',
  };
}

export function buildMemberBriefParams(query: MemberBriefQuery): FilterParams {
  return {
    keyword: optionalText(query.keyword),
    status: query.status ? STATUS_CODE[query.status] : undefined,
    agent_id: query.agentId || undefined,
    register_channel: optionalText(query.registerChannel),
    register_method: optionalText(query.registerMethod),
  };
}

export function buildSortParams(
  key: MemberSortKey,
  order: SortOrder
): FilterParams {
  if (key === 'registeredAt') return {};
  return {
    sort_by: SORT_BY[key],
    sort_order:
      order === 'asc' ? MEMBER_SORT_ORDER.ASC : MEMBER_SORT_ORDER.DESC,
  };
}
