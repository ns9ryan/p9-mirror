import type { MemberRecord } from './types';

export type MemberTagKey =
  | 'noPhone'
  | 'noEmail'
  | 'betBanned'
  | 'deposited'
  | 'notDeposited';

export function formatAmount(value: number) {
  return value.toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

export function formatAmountCount(amount: number, count: number) {
  return `${formatAmount(amount)}(${count})`;
}

export function displayText(value: string) {
  const text = value.trim();
  return text || '--';
}

export function memberTagKeys(record: MemberRecord): MemberTagKey[] {
  const tags: MemberTagKey[] = [];
  if (!record.phone) tags.push('noPhone');
  if (!record.email) tags.push('noEmail');
  if (record.betBanned) tags.push('betBanned');
  if (record.totalDepositCount > 0) tags.push('deposited');
  else tags.push('notDeposited');
  return tags;
}
