export type AccountStatus = 'normal' | 'disabled';
export type RegisterType = 'email' | 'username' | 'phone';
export type ReferralStatus = 'normal' | 'none';
export type LevelKind = 'auto' | 'fixed';
export type KeywordType = 'id' | 'account' | 'phone' | 'email';
export type TimeType = 'register' | 'login';
export type LoginState = '' | 'today' | 'd7' | 'd30' | 'over30' | 'never';
export type QuickRange = 'today' | 'd7' | 'd30';

export interface MemberRecord {
  id: number;
  account: string;
  phone: string;
  email: string;
  vipLevel: number;
  registerType: RegisterType;
  level: string;
  levelKind: LevelKind;
  registerSource: string;
  channelSource: string;
  parentAgent: string;
  parentAgentId: string;
  balance: number;
  unsettled: number;
  totalDepositAmount: number;
  totalDepositCount: number;
  totalWithdrawAmount: number;
  totalWithdrawCount: number;
  depositWithdrawDiff: number;
  firstDepositAmount: number;
  todayDeposit: number;
  todayWithdraw: number;
  todayValidBet: number;
  todayWinLoss: number;
  todayBonus: number;
  referralStatus: ReferralStatus;
  registeredAt: string;
  lastLoginAt: string;
  lastLoginIp: string;
  status: AccountStatus;
  betBanned: boolean;
}

export interface MemberFormValue {
  account: string;
  phone: string;
  email: string;
  vipLevel: number;
  registerType: RegisterType;
  level: string;
  levelKind: LevelKind;
  registerSource: string;
  channelSource: string;
  parentAgent: string;
  parentAgentId: string;
  referralStatus: ReferralStatus;
  status: AccountStatus;
  betBanned: boolean;
}

export interface MemberQuery {
  timeType: TimeType;
  dateRange: string[];
  keywordType: KeywordType;
  keyword: string;
  autoLevel: string;
  fixedLevel: string;
  registerSource: string;
  status: '' | AccountStatus;
  channelSource: string;
  loginState: LoginState;
  betBanned: boolean;
  noPhone: boolean;
  noEmail: boolean;
  deposited: boolean;
  notDeposited: boolean;
}
