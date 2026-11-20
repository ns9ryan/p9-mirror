export type QuickRange = 'today' | 'd7' | 'd30';

export type SortDirection = 'ascend' | 'descend' | '';

export type GameSortField =
  | 'betCount'
  | 'bettorCount'
  | 'validBet'
  | 'memberWinLoss';

export interface GameReportRecord {
  date: string;
  currency: string;
  betCount: number;
  bettorCount: number;
  validBet: number;
  memberWinLoss: number;
  testBetCount: number;
  testBettorCount: number;
  testValidBet: number;
  testMemberWinLoss: number;
}

export interface GameReportRow {
  date: string;
  currency: string;
  betCount: number;
  bettorCount: number;
  validBet: number;
  memberWinLoss: number;
}
