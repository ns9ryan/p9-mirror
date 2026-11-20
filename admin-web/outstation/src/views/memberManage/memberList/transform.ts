import {
  MEMBER_GENDER,
  MEMBER_STATUS,
  type MemberBaseItem,
  type MemberDetail,
} from '@/api/member';
import { formatSiteTime } from '@/utils/site-time';
import type {
  AccountStatus,
  Gender,
  MemberDetailRecord,
  MemberRecord,
} from './types';

const STATUS_BY_CODE: Record<number, AccountStatus> = {
  [MEMBER_STATUS.NORMAL]: 'normal',
  [MEMBER_STATUS.FROZEN]: 'frozen',
  [MEMBER_STATUS.DISABLED]: 'disabled',
};

const GENDER_BY_CODE: Record<number, Gender> = {
  [MEMBER_GENDER.UNSET]: 'unset',
  [MEMBER_GENDER.MALE]: 'male',
  [MEMBER_GENDER.FEMALE]: 'female',
  [MEMBER_GENDER.SECRET]: 'secret',
};

function formatTime(ms: number | null | undefined) {
  return formatSiteTime(ms, undefined, '');
}

export default function toMemberRecord(
  item: MemberBaseItem & { phone?: string | null; email?: string | null }
): MemberRecord {
  const { levels, stats } = item;
  const fixedLevels = levels?.fixed_levels || [];
  return {
    id: item.id,
    code: item.code || '',
    account: item.account || '',
    nickname: item.nickname || '',
    remark: item.remark || '',
    phone: item.phone || '',
    email: item.email || '',
    vipLevel: levels?.vip_level ?? 0,
    registerType: item.register_method || '',
    autoLevelId: levels?.auto_level_id || 0,
    autoLevelName: levels?.auto_level_name || '',
    fixedLevelIds: fixedLevels.map((level) => level.id),
    fixedLevelNames: fixedLevels.map((level) => level.name),
    registerSource: item.register_channel || '',
    channelSource: '',
    parentAgent: item.agent_account || '',
    parentAgentId: item.agent_id ? String(item.agent_id) : '',
    balance: null,
    unsettled: null,
    totalDepositAmount: stats?.total_deposit ?? 0,
    totalDepositCount: stats?.total_deposit_count ?? 0,
    totalWithdrawAmount: stats?.total_withdrawal ?? 0,
    totalWithdrawCount: stats?.total_withdrawal_count ?? 0,
    depositWithdrawDiff: stats?.total_deposit_withdrawal_diff ?? 0,
    firstDepositAmount: null,
    periodDeposit: stats?.current_period_deposit ?? 0,
    periodWithdraw: stats?.current_period_withdrawal ?? 0,
    periodBet: stats?.current_period_bet ?? 0,
    periodValidBet: stats?.current_period_valid_bet ?? 0,
    periodWinLoss: stats?.current_period_win_loss ?? 0,
    totalWinLoss: stats?.total_win_loss ?? 0,
    periodBonus: null,
    referralStatus: item.inviter_member_id ? 'normal' : 'none',
    registeredAt: formatTime(item.created_at),
    lastLoginAt: formatTime(item.last_login_at),
    lastLoginIp: item.last_login_ip || '',
    status: STATUS_BY_CODE[item.status] || 'normal',
    betBanned: !!item.betting_disabled,
  };
}

export function toMemberDetailRecord(item: MemberDetail): MemberDetailRecord {
  return {
    ...toMemberRecord(item),
    avatarId: item.avatar_id || 0,
    avatarFrameId: item.avatar_frame_id || 0,
    birthday: item.birthday || '',
    gender: GENDER_BY_CODE[item.gender] || 'unset',
    regionCode: item.region_code || '',
    agentLineCode: item.agent_line_code || '',
    chatBanned: !!item.chat_disabled,
    withdrawBanned: !!item.withdrawal_disabled,
    registerIp: item.register_ip || '',
    registerDeviceCode: item.register_device_code || '',
    registerDomain: item.register_domain || '',
    inviterMemberId: item.inviter_member_id
      ? String(item.inviter_member_id)
      : '',
  };
}
