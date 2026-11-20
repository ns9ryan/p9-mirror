import type { TaskPayoutMode } from '@/api/task';

export type TaskClient = 'web' | 'mobile' | 'android' | 'ios';

export type DispatchMode = 'manual' | 'auto';

export interface TaskEditValue {
  reward_amount: number;
  completion_amount: number;
  payout_mode: TaskPayoutMode;
  timezone: string;
  sort: number;
}

export interface TaskConfig {
  enabled: boolean;
  dispatchMode: DispatchMode;
  autoLevelIds: number[];
  fixedLevelIds: number[];
  washMultiplier: number | undefined;
  clients: TaskClient[];
  oncePerDevice: boolean;
  effect: 'immediate';
}
