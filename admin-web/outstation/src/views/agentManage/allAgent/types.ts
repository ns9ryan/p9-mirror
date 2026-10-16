export type AccountStatus = 'normal' | 'unset';
export type WithdrawLimit = 'none' | 'limited';
export type QuickRange = 'today' | 'd7' | 'd30';

export interface AllAgentRecord {
  id: number;
  parentId: string;
  subCount: number;
  subTotal: number;
  betLevel1: number;
  betLevel2: number;
  betLevel3: number;
  claimedCommission: number;
  pendingCommission: number;
  commission: number;
  frozenCommission: number;
  cancelledCommission: number;
  unclaimedCommission: number;
  withdrawLimit: WithdrawLimit;
  remark: string;
  registerSource: string;
  promoHost: string;
  promoCode: string;
  visitUv: number;
  registeredAt: string;
  lastLoginAt: string;
  status: AccountStatus;
  selfBetBanned: boolean;
  lineBetBanned: boolean;
}

export interface AllAgentQuery {
  dateRange: string[];
  keyword: string;
  registerSource: string;
  subCountMin?: number;
  subCountMax?: number;
  withdrawLimit: '' | WithdrawLimit;
  status: '' | AccountStatus;
}
