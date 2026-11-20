export type AccountStatus = 'normal' | 'disabled';
export type WithdrawLimit = 'none' | 'forbidden' | 'unset';
export type KeywordType = 'id' | 'account';
export type TimeType = 'register' | 'login';
export type QuickRange = 'today' | 'd7' | 'd30';

export interface AgentRecord {
  id: number;
  account: string;
  /** 层级名称，对应接口 level_alias */
  level: string;
  /** 层级数字，对应接口 level；没有时为 0 */
  layer: number;
  /** 上级 ID，顶层为 0 */
  parentId: number;
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
  /** 代理线路，树接口 line_code */
  lineCode?: string;
  visitUv: number;
  registeredAt: string;
  lastLoginAt: string;
  /** 列表接口已补全。树节点只有摘要时为 false */
  detailReady?: boolean;
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

export interface InfoItem {
  key: string;
  label: string;
  value?: string;
  /** 右侧操作按钮文案，不传则不显示 */
  action?: string;
  /** 独占一整行 */
  full?: boolean;
}

export type FieldEditorKind =
  | 'text'
  | 'textarea'
  | 'share'
  | 'phone'
  | 'withdrawPassword';

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
