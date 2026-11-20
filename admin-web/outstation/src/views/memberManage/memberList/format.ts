import type { AccountStatus, MemberDetailRecord, MemberRecord } from './types';

type Translate = (key: string) => string;

export const STATUS_COLORS: Record<AccountStatus, string> = {
  normal: 'green',
  frozen: 'orange',
  disabled: 'red',
};

const VIP_COLORS = ['gray', 'arcoblue', 'blue', 'purple', 'orange', 'red'];

export function vipColor(level: number) {
  return VIP_COLORS[Math.min(Math.max(level, 0), VIP_COLORS.length - 1)];
}

/** 列表只返回手机号 / 邮箱为空，不能据此判断“无手机 / 无邮箱”，故不作标记 */
export type MemberTagKey = 'betBanned' | 'deposited' | 'notDeposited';

/** null 表示接口尚未提供该字段 */
export function formatAmount(value: number | null) {
  if (value == null) return '--';
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

/** 文档只确认了 account / admin，phone、email 按字面对应；未登记的取值原样展示 */
const REGISTER_METHOD_KEYS: Record<string, string> = {
  account: 'memberManage.registerType.username',
  phone: 'memberManage.registerType.phone',
  email: 'memberManage.registerType.email',
};

const REGISTER_CHANNEL_KEYS: Record<string, string> = {
  admin: 'memberManage.registerChannel.admin',
};

export const REGISTER_METHODS = Object.keys(REGISTER_METHOD_KEYS);
export const REGISTER_CHANNELS = Object.keys(REGISTER_CHANNEL_KEYS);

export function registerMethodText(value: string, t: Translate) {
  const key = REGISTER_METHOD_KEYS[value];
  return key ? t(key) : displayText(value);
}

export function registerChannelText(value: string, t: Translate) {
  const key = REGISTER_CHANNEL_KEYS[value];
  return key ? t(key) : displayText(value);
}

export function restrictionText(record: MemberDetailRecord, t: Translate) {
  const keys = [
    record.chatBanned && 'memberManage.restriction.chat',
    record.betBanned && 'memberManage.restriction.bet',
    record.withdrawBanned && 'memberManage.restriction.withdraw',
  ].filter((key): key is string => !!key);
  if (!keys.length) return t('memberManage.restriction.none');
  return keys.map((key) => t(key)).join('、');
}

/** 未关联固定层级时返回空字符串，配合 displayText 显示 -- */
export function fixedLevelsText(record: MemberRecord) {
  return record.fixedLevelNames.join('、');
}

export function memberTagKeys(record: MemberRecord): MemberTagKey[] {
  const tags: MemberTagKey[] = [];
  if (record.betBanned) tags.push('betBanned');
  if (record.totalDepositCount > 0) tags.push('deposited');
  else tags.push('notDeposited');
  return tags;
}
