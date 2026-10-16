export type AccountStatus = 'normal' | 'disabled';
export type WithdrawLimit = 'none' | 'forbidden' | 'unset';
export type KeywordType = 'id' | 'account';
export type TimeType = 'register' | 'login';
export type QuickRange = 'today' | 'd7' | 'd30';

export interface AgentRecord {
  id: number;
  account: string;
  level: string;
  parentAgent: string;
  topAgent: string;
  registerSource: string;
  directMembers: number;
  directAgents: number;
  allMembers: number;
  allAgents: number;
  shareRatio: number;
  balance: number;
  totalCommission: number;
  unpaidCommission: number;
  status: AccountStatus;
  withdrawLimit: WithdrawLimit;
  remark: string;
  promoCode: string;
  promoHost: string;
  visitUv: number;
  registeredAt: string;
  lastLoginAt: string;
}

export interface AgentTreeNode extends AgentRecord {
  depth: number;
  children?: AgentTreeNode[];
}

export interface AgentFormValue {
  account: string;
  level: string;
  parentAgent: string;
  topAgent: string;
  registerSource: string;
  shareRatio: number;
  withdrawLimit: WithdrawLimit;
  remark: string;
  status: AccountStatus;
  promoCode: string;
}

export interface AgentLevelSetting {
  layer: number;
  name: string;
}

export interface AgentSettings {
  levels: AgentLevelSetting[];
  maxShareRatio: number;
  applyShareMax: number;
  applyShareMin: number;
  defaultLevel: string;
  defaultShareRatio: number;
  defaultWithdrawLimit: WithdrawLimit;
}

export interface AgentQuery {
  timeType: TimeType;
  dateRange: string[];
  keywordType: KeywordType;
  keyword: string;
  registerSource: string;
  directMemberMin?: number;
  directMemberMax?: number;
  withdrawLimit: '' | WithdrawLimit;
  level: string;
  status: '' | AccountStatus;
  clientHost: string;
}
