export type QuickRange = 'today' | 'd7' | 'd30';

export type SortDirection = 'ascend' | 'descend' | '';

export type WinLossSortField =
  | 'betCount'
  | 'bettorCount'
  | 'validBet'
  | 'memberWinLoss'
  | 'costFee';

export type IdType = 'memberId' | 'agentId';

export interface WinLossRecord {
  date: string;
  currency: string;
  betCount: number;
  bettorCount: number;
  validBet: number;
  memberWinLoss: number;
  costFee: number;
  memberIds: string[];
  agentIds: string[];
}
