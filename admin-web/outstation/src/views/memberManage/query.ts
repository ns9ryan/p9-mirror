import dayjs from 'dayjs';
import type { MemberQuery, MemberRecord, QuickRange } from './types';

export function createMemberQuery(): MemberQuery {
  return {
    timeType: 'register',
    dateRange: [],
    keywordType: 'id',
    keyword: '',
    autoLevel: '',
    fixedLevel: '',
    registerSource: '',
    status: '',
    channelSource: '',
    loginState: '',
    betBanned: false,
    noPhone: false,
    noEmail: false,
    deposited: false,
    notDeposited: false,
  };
}

export function buildQuickRange(
  key: QuickRange,
  now = dayjs()
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
  now = dayjs()
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
  const time = dayjs(value);
  const start = dayjs(range[0]).startOf('day');
  const end = dayjs(range[1]).endOf('day');
  return !time.isBefore(start) && !time.isAfter(end);
}

function matchLoginState(
  record: MemberRecord,
  query: MemberQuery,
  now = dayjs()
) {
  const { loginState } = query;
  if (!loginState) return true;
  if (loginState === 'never') return !record.lastLoginAt;
  if (!record.lastLoginAt) return false;
  const login = dayjs(record.lastLoginAt);
  if (loginState === 'today') return login.isSame(now, 'day');
  if (loginState === 'd7') {
    return !login.isBefore(now.startOf('day').subtract(6, 'day'));
  }
  if (loginState === 'd30') {
    return !login.isBefore(now.startOf('day').subtract(29, 'day'));
  }
  return login.isBefore(now.startOf('day').subtract(30, 'day'));
}

export function matchMember(
  record: MemberRecord,
  query: MemberQuery,
  now = dayjs()
) {
  const timeValue =
    query.timeType === 'register' ? record.registeredAt : record.lastLoginAt;
  if (!inDateRange(timeValue, query.dateRange)) return false;

  const keyword = query.keyword.trim().toLowerCase();
  if (keyword) {
    const source = {
      id: String(record.id),
      account: record.account,
      phone: record.phone,
      email: record.email,
    }[query.keywordType];
    if (!source.toLowerCase().includes(keyword)) return false;
  }

  if (query.autoLevel) {
    if (record.levelKind !== 'auto' || record.level !== query.autoLevel) {
      return false;
    }
  }
  if (query.fixedLevel) {
    if (record.levelKind !== 'fixed' || record.level !== query.fixedLevel) {
      return false;
    }
  }
  if (query.registerSource && record.registerSource !== query.registerSource) {
    return false;
  }
  if (query.status && record.status !== query.status) return false;
  if (query.channelSource && record.channelSource !== query.channelSource) {
    return false;
  }
  if (!matchLoginState(record, query, now)) return false;
  if (query.betBanned && !record.betBanned) return false;
  if (query.noPhone && record.phone) return false;
  if (query.noEmail && record.email) return false;
  if (query.deposited && record.totalDepositCount <= 0) return false;
  if (query.notDeposited && record.totalDepositCount > 0) return false;
  return true;
}
