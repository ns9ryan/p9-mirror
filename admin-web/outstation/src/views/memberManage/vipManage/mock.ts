import dayjs from 'dayjs';
import type {
  DispatchMode,
  LevelChangeType,
  RewardStatus,
  VipLevelChangeQuery,
  VipLevelChangeRecord,
  VipLevelRecord,
  VipRewardQuery,
  VipRewardRecord,
} from './types';

export function createVipLevelSeed(): VipLevelRecord[] {
  const row = (
    level: number,
    reward: number,
    promote: [number, number],
    keep: [number, number],
    memberCount: number,
    updatedAt: string
  ): VipLevelRecord => ({
    id: level + 100,
    level,
    reward,
    promoteDeposit: promote[0],
    promoteWash: promote[1],
    keepDeposit: keep[0],
    keepWash: keep[1],
    memberCount,
    updatedAt,
  });

  return [
    row(0, 0, [0, 0], [0, 0], 12456, '2026-10-10 11:20:05'),
    row(1, 50, [500, 1000], [300, 600], 8320, '2026-10-14 09:21:18'),
    row(2, 100, [2000, 5000], [1000, 3000], 5678, '2026-10-12 16:45:37'),
    row(3, 300, [10000, 30000], [5000, 15000], 2348, '2026-10-10 11:20:05'),
    row(4, 800, [50000, 150000], [20000, 60000], 986, '2026-10-09 10:15:08'),
    row(
      5,
      2000,
      [200000, 500000],
      [100000, 300000],
      320,
      '2026-10-08 18:26:41'
    ),
    row(
      6,
      5000,
      [500000, 1500000],
      [250000, 800000],
      128,
      '2026-10-07 14:02:33'
    ),
    row(
      7,
      10000,
      [1000000, 3000000],
      [500000, 1500000],
      46,
      '2026-10-06 20:11:09'
    ),
    row(
      8,
      28000,
      [3000000, 8000000],
      [1500000, 4000000],
      12,
      '2026-10-05 09:30:00'
    ),
    row(
      9,
      66000,
      [8000000, 20000000],
      [4000000, 10000000],
      3,
      '2026-10-03 16:48:52'
    ),
  ];
}

const NICKNAMES = [
  '小明',
  '阿华',
  '测试用户',
  '开心果',
  '大赢家',
  '玩家001',
  '幸运星',
  '风轻云淡',
  '阿杰',
  '晨曦',
  '追风少年',
  '小确幸',
  '星辰',
  '一路向北',
  '甜甜圈',
  '老王',
];

const LEVEL_POOL = [1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 5, 5, 6, 7, 8, 9];
const AGENT_PATHS = [
  ['顶级代理', '代理A', '代理B'],
  ['顶级代理', '代理A'],
  ['顶级代理', '代理C'],
  ['顶级代理'],
];
const AUTO_LAYERS = ['A1', 'A2', 'B1', 'B2'];
const FIXED_LAYERS = [['白金会员', '活动特权'], ['白金会员'], ['活动特权'], []];
const TIME_FORMAT = 'YYYY-MM-DD HH:mm:ss';

export function createVipRewardSeed(
  levels: VipLevelRecord[],
  washMultiple: number,
  total = 128
): VipRewardRecord[] {
  let seed = 20261019;
  const random = () => {
    seed = (seed * 16807) % 2147483647;
    return seed / 2147483647;
  };
  const pick = <T>(list: T[]) => list[Math.floor(random() * list.length)];
  const rewardByLevel = new Map(
    levels.map((item) => [item.level, item.reward])
  );
  let cursor = dayjs('2026-10-19 10:35:20');

  return Array.from({ length: total }, (_, index) => {
    const level = LEVEL_POOL[Math.floor(random() * LEVEL_POOL.length)];
    const reward = rewardByLevel.get(level) ?? 0;
    const mode: DispatchMode = random() < 0.45 ? 'auto' : 'manual';
    let status: RewardStatus = 'dispatched';
    if (mode === 'manual') status = random() < 0.2 ? 'pending' : 'claimed';
    const promotedAt = cursor;
    cursor = cursor.subtract(2 + Math.floor(random() * 10), 'hour');
    let claimedAt = '';
    if (status === 'dispatched') claimedAt = promotedAt.format(TIME_FORMAT);
    if (status === 'claimed') {
      claimedAt = promotedAt
        .add(5 + Math.floor(random() * 600), 'minute')
        .format(TIME_FORMAT);
    }
    const no = String(index + 1).padStart(3, '0');
    return {
      orderNo: `VR${promotedAt.format('YYYYMMDDHHmmss')}${no}`,
      account: `member${no}`,
      nickname: NICKNAMES[index % NICKNAMES.length],
      level,
      reward,
      mode,
      status,
      promotedAt: promotedAt.format(TIME_FORMAT),
      claimedAt,
      washMultiple,
      washRequired: reward * washMultiple,
      agentPath: pick(AGENT_PATHS),
      autoLayer: pick(AUTO_LAYERS),
      fixedLayers: pick(FIXED_LAYERS),
      currentLevel: Math.min(9, level + (random() < 0.8 ? 0 : 1)),
    };
  });
}

export function createRewardQuery(): VipRewardQuery {
  return {
    account: '',
    nickname: '',
    level: undefined,
    status: undefined,
    mode: undefined,
    dateRange: [],
  };
}

export function filterVipRewards(
  list: VipRewardRecord[],
  query: VipRewardQuery
) {
  const account = query.account.trim();
  const nickname = query.nickname.trim();
  const [start, end] = query.dateRange;
  return list.filter((item) => {
    if (account && !item.account.includes(account)) return false;
    if (nickname && !item.nickname.includes(nickname)) return false;
    if (query.level != null && item.level !== query.level) return false;
    if (query.status && item.status !== query.status) return false;
    if (query.mode && item.mode !== query.mode) return false;
    if (start && end) {
      if (!item.claimedAt) return false;
      const time = dayjs(item.claimedAt);
      if (time.isBefore(dayjs(start)) || time.isAfter(dayjs(end))) return false;
    }
    return true;
  });
}

const CHANGE_REASONS: Record<Exclude<LevelChangeType, 'promote'>, string[]> = {
  demote: [
    '保级周期内未达到保级条件，自动降级',
    '近30天有效洗码量未达标，自动降级',
  ],
  manual: [
    '客服申请调整等级，给予新用户体验',
    '违反平台规则，手动降级',
    '用户等级补发，活动奖励',
    '运营活动奖励，维持当前等级',
  ],
  recalc: ['根据最新等级规则重新计算', '晋级条件调整后重新计算'],
};
const OPERATORS = ['Admin', 'operator001'];
const CHANGE_REMARKS: Record<LevelChangeType, string> = {
  promote: '晋级奖励已按规则生成',
  demote: '降级后已领取的晋级奖励不会收回',
  manual: '本次调整仅影响当前VIP等级，不补发晋级奖励',
  recalc: '规则调整后批量重算，不补发晋级奖励',
};

function memberIdOf(account: string) {
  return 100000 + Number(account.replace(/\D/g, ''));
}

export function createLevelChangeSeed(
  rewards: VipRewardRecord[],
  extra = 40
): VipLevelChangeRecord[] {
  let seed = 20261022;
  const random = () => {
    seed = (seed * 16807) % 2147483647;
    return seed / 2147483647;
  };
  const pick = <T>(list: T[]) => list[Math.floor(random() * list.length)];

  const promotions = rewards.map(
    (item): VipLevelChangeRecord => ({
      id: `LC${item.orderNo.slice(2)}`,
      account: item.account,
      memberId: memberIdOf(item.account),
      nickname: item.nickname,
      fromLevel: item.level - 1,
      toLevel: item.level,
      type: 'promote',
      reason: '累计有效存款和有效洗码量达到晋级条件',
      operator: '',
      changedAt: item.promotedAt,
      rewardOrderNo: item.orderNo,
      remark: CHANGE_REMARKS.promote,
    })
  );

  const base = dayjs('2026-10-19 12:00:00');
  const others = Array.from({ length: extra }, (_, index) => {
    const member = pick(rewards);
    const type = pick<Exclude<LevelChangeType, 'promote'>>([
      'demote',
      'manual',
      'recalc',
    ]);
    const fromLevel = 1 + Math.floor(random() * 8);
    let toLevel = fromLevel - 1;
    if (type === 'manual') toLevel = fromLevel + pick([-1, 0, 1]);
    if (type === 'recalc') toLevel = fromLevel + pick([-1, 1]);
    const changedAt = base
      .subtract(Math.floor(random() * 30 * 24 * 60), 'minute')
      .format(TIME_FORMAT);
    return {
      id: `LC${dayjs(changedAt).format('YYYYMMDDHHmmss')}9${String(
        index
      ).padStart(2, '0')}`,
      account: member.account,
      memberId: memberIdOf(member.account),
      nickname: member.nickname,
      fromLevel,
      toLevel,
      type,
      reason: pick(CHANGE_REASONS[type]),
      operator: type === 'manual' ? pick(OPERATORS) : '',
      changedAt,
      rewardOrderNo: '',
      remark: CHANGE_REMARKS[type],
    };
  });

  return [...promotions, ...others].sort((a, b) =>
    b.changedAt.localeCompare(a.changedAt)
  );
}

export function createChangeQuery(): VipLevelChangeQuery {
  return {
    account: '',
    type: undefined,
    level: undefined,
    dateRange: [],
  };
}

export function filterLevelChanges(
  list: VipLevelChangeRecord[],
  query: VipLevelChangeQuery
) {
  const account = query.account.trim();
  const [start, end] = query.dateRange;
  return list.filter((item) => {
    if (account && !item.account.includes(account)) return false;
    if (query.type && item.type !== query.type) return false;
    if (
      query.level != null &&
      item.fromLevel !== query.level &&
      item.toLevel !== query.level
    ) {
      return false;
    }
    if (start && end) {
      const time = dayjs(item.changedAt);
      if (time.isBefore(dayjs(start).startOf('day'))) return false;
      if (time.isAfter(dayjs(end).endOf('day'))) return false;
    }
    return true;
  });
}
