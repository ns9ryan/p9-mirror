import type { RewardStatus, VipLevelChangeRecord } from './types';

export function levelColor(level: number) {
  if (level <= 1) return 'gray';
  if (level <= 3) return 'arcoblue';
  if (level <= 5) return 'purple';
  if (level <= 7) return 'orangered';
  return 'red';
}

export const statusColor: Record<RewardStatus, string> = {
  claimed: 'green',
  dispatched: 'green',
  pending: 'orange',
};

export function changeColor(record: VipLevelChangeRecord) {
  return record.toLevel < record.fromLevel ? 'red' : 'green';
}
