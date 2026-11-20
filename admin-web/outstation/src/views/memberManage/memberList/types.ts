export type AccountStatus = 'normal' | 'frozen' | 'disabled';
export type RegisterType = 'email' | 'username' | 'phone';
export type ReferralStatus = 'normal' | 'none';
export type TimeType = 'register' | 'login';
export type QuickRange = 'today' | 'd7' | 'd30';

/**
 * 列表行。金额为结算币种最小单位；接口尚未提供的字段为 null，展示为 --。
 * period* 为当前结算周期统计，不是自然日
 */
export interface MemberRecord {
  id: number;
  /** 12 位会员对外编号 */
  code: string;
  account: string;
  nickname: string;
  remark: string;
  /** 列表接口暂不填充，始终为空 */
  phone: string;
  /** 列表接口暂不填充，始终为空 */
  email: string;
  vipLevel: number;
  /** 接口 register_method 原值，如 account */
  registerType: string;
  /** 每个会员都有一个自动层级，同时可关联零到多个固定层级 */
  autoLevelId: number;
  autoLevelName: string;
  /** 与 fixedLevelNames 一一对应，按 ID 升序 */
  fixedLevelIds: number[];
  fixedLevelNames: string[];
  /** 接口 register_channel 原值，如 admin */
  registerSource: string;
  channelSource: string;
  parentAgent: string;
  parentAgentId: string;
  balance: number | null;
  unsettled: number | null;
  totalDepositAmount: number;
  totalDepositCount: number;
  totalWithdrawAmount: number;
  totalWithdrawCount: number;
  depositWithdrawDiff: number;
  firstDepositAmount: number | null;
  periodDeposit: number;
  periodWithdraw: number;
  periodBet: number;
  periodValidBet: number;
  periodWinLoss: number;
  totalWinLoss: number;
  periodBonus: number | null;
  referralStatus: ReferralStatus;
  registeredAt: string;
  lastLoginAt: string;
  lastLoginIp: string;
  status: AccountStatus;
  betBanned: boolean;
}

export type Gender = 'unset' | 'male' | 'female' | 'secret';

/** 详情接口补充的字段；手机号、邮箱、真实姓名、身份ID、钱包余额详情接口不返回 */
export interface MemberDetailRecord extends MemberRecord {
  /** 0 表示未设置，图片地址从头像资源列表解析 */
  avatarId: number;
  avatarFrameId: number;
  birthday: string;
  gender: Gender;
  regionCode: string;
  agentLineCode: string;
  chatBanned: boolean;
  withdrawBanned: boolean;
  registerIp: string;
  registerDeviceCode: string;
  registerDomain: string;
  inviterMemberId: string;
}

export interface MemberQuery {
  timeType: TimeType;
  dateRange: string[];
  keyword: string;
  autoLevelId: number | '';
  fixedLevelId: number | '';
  registerChannel: string;
  registerMethod: string;
  status: '' | AccountStatus;
  betBanned: boolean;
  noPhone: boolean;
  noEmail: boolean;
  deposited: boolean;
  notDeposited: boolean;
}

export type MemberLayout = 'classic' | 'v2';
export type MemberSortKey =
  | 'registeredAt'
  | 'totalDeposit'
  | 'depositCount'
  | 'totalWithdraw'
  | 'withdrawCount';
export type SortOrder = 'asc' | 'desc';

/** 新版列表的单行筛选条件 */
export interface MemberBriefQuery {
  keyword: string;
  status: '' | AccountStatus;
  agentId: number | '';
  registerChannel: string;
  registerMethod: string;
}
