export type DispatchMode = 'manual' | 'auto';
export type VipTab = 'level' | 'setting' | 'icon' | 'record' | 'change';
export type RewardStatus = 'claimed' | 'dispatched' | 'pending';

export interface VipLevelRecord {
  id: number;
  level: number;
  reward: number;
  promoteDeposit: number;
  promoteWash: number;
  keepDeposit: number;
  keepWash: number;
  memberCount: number;
  updatedAt: string;
}

export interface VipIconGroup {
  id: number;
  no: number;
  startLevel: number;
  endLevel: number;
  defaultIcon: string;
  /** 当前使用的图标 */
  icon: string;
  updatedAt: string;
}

/** 已上传、尚未保存的图标 */
export interface VipIconDraft {
  /** 预览用完整 URL */
  url: string;
  /** 保存时提交的对象 Key */
  key: string;
}

export type VipIconAction = 'upload' | 'save' | 'restore';

export interface VipIconBusy {
  no: number;
  action: VipIconAction;
}

export interface VipLevelFormValue {
  level: number;
  reward: number;
  promoteDeposit: number;
  promoteWash: number;
  keepDeposit: number;
  keepWash: number;
}

export interface VipSettings {
  levelEnabled: boolean;
  rewardEnabled: boolean;
  keepEnabled: boolean;
  dispatchMode: DispatchMode;
  washMultiple: number;
}

export interface VipRewardRecord {
  orderNo: string;
  account: string;
  nickname: string;
  level: number;
  reward: number;
  mode: DispatchMode;
  status: RewardStatus;
  promotedAt: string;
  /** 待领取时为空字符串 */
  claimedAt: string;
  washMultiple: number;
  washRequired: number;
  agentPath: string[];
  autoLayer: string;
  fixedLayers: string[];
  currentLevel: number;
}

export interface VipRewardQuery {
  account: string;
  nickname: string;
  level: number | undefined;
  status: RewardStatus | undefined;
  mode: DispatchMode | undefined;
  dateRange: string[];
}

export type LevelChangeType = 'promote' | 'demote' | 'manual' | 'recalc';

export interface VipLevelChangeRecord {
  id: string;
  account: string;
  memberId: number;
  nickname: string;
  fromLevel: number;
  toLevel: number;
  type: LevelChangeType;
  reason: string;
  /** 系统操作时为空字符串 */
  operator: string;
  changedAt: string;
  rewardOrderNo: string;
  remark: string;
}

export interface VipLevelChangeQuery {
  account: string;
  type: LevelChangeType | undefined;
  level: number | undefined;
  dateRange: string[];
}
