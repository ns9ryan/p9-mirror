import {
  ACTIVITY_STATUS,
  ACTIVITY_TYPE,
  type ActivityEnumGroups,
  type ActivityEnumOption,
  type ActivityEnumParams,
  type ActivityI18nItem,
  type ActivityRewardSchema,
  type ActivitySettings,
} from '@/api/activity';
import { SUPPORTED_TYPE_CODES } from './create-payload';
import { DATE_TIME, formatSiteTime } from './site-time';
import type { ActivityRecord, ActivityStatus } from './types';

type Translate = (key: string, named?: Record<string, unknown>) => string;

const PREFIX = 'discountCenter.activityCenter';

const STATUS_CODES: Record<ActivityStatus, number> = {
  draft: ACTIVITY_STATUS.DRAFT,
  running: ACTIVITY_STATUS.RUNNING,
  ended: ACTIVITY_STATUS.ENDED,
};

export const STATUS_COLORS: Record<ActivityStatus, string> = {
  draft: 'gray',
  running: 'green',
  ended: 'red',
};

function keyOf<K extends string>(codes: Record<K, number>, code: number) {
  return (Object.keys(codes) as K[]).find((key) => codes[key] === code);
}

export function statusKey(code: number) {
  return keyOf(STATUS_CODES, code);
}

export function statusColor(code: number) {
  const key = statusKey(code);
  return key ? STATUS_COLORS[key] : 'gray';
}

/** 状态、类型、客户端的文案来自枚举接口，取不到时显示编号 */
export function statusText(code: number, params: ActivityEnumParams | null) {
  return enumLabel(params, 'status', code);
}

export function typeText(code: number, params: ActivityEnumParams | null) {
  return enumLabel(params, 'type', code);
}

export function clientNames(
  codes: number[],
  params: ActivityEnumParams | null
) {
  return codes.map((code) => enumLabel(params, 'client_type', code));
}

/** 未限制客户端时返回「全部」 */
export function clientsText(
  codes: number[] | null,
  t: Translate,
  params: ActivityEnumParams | null
) {
  if (!codes?.length) return t(`${PREFIX}.all`);
  return clientNames(codes, params).join('、');
}

export function displayText(value?: string | null) {
  return value || '---';
}

/** 接口时间均为 Unix 秒，按本站时区显示 */
export function formatTime(sec?: number | null, format = DATE_TIME) {
  if (!sec) return '';
  return formatSiteTime(sec, format);
}

export const TYPE_COLORS: Record<number, string> = {
  [ACTIVITY_TYPE.DEPOSIT]: 'arcoblue',
  [ACTIVITY_TYPE.BET]: 'orange',
  [ACTIVITY_TYPE.REFER]: 'green',
  [ACTIVITY_TYPE.LUCK_WHEEL]: 'purple',
  [ACTIVITY_TYPE.ASSIST]: 'cyan',
  [ACTIVITY_TYPE.BLIND_BOX]: 'magenta',
  [ACTIVITY_TYPE.NEWCOMER]: 'gold',
  [ACTIVITY_TYPE.AGENT]: 'lime',
  [ACTIVITY_TYPE.RESCUE]: 'red',
  [ACTIVITY_TYPE.LUCKY_BET]: 'orangered',
  [ACTIVITY_TYPE.VIP]: 'pinkpurple',
  [ACTIVITY_TYPE.RED_PACKET]: 'red',
  [ACTIVITY_TYPE.IMAGE_AD]: 'blue',
  [ACTIVITY_TYPE.SIGN_IN]: 'green',
};

export function activityTimeText(record: ActivityRecord, t: Translate) {
  if (record.is_permanent) return t(`${PREFIX}.permanent`);
  const open = formatTime(record.open_at_sec);
  const close = formatTime(record.close_at_sec);
  if (!open && !close) return '';
  return `${open || '---'} ~ ${close || '---'}`;
}

/** 金额为结算币种最小单位，币种精度未提供，按原值展示 */
export function formatAmount(value?: number | null) {
  if (value == null) return '';
  return value.toLocaleString('en-US');
}

export function autoAuditText(record: ActivityRecord, t: Translate) {
  if (!record.auto_audit) return t(`${PREFIX}.autoReview.off`);
  return t(`${PREFIX}.autoReview.limit`, {
    amount: formatAmount(record.auto_audit_max_amt),
  });
}

export function wageringText(record: ActivityRecord, t: Translate) {
  if (!record.wagering_multiplier) return '';
  return t(`${PREFIX}.wageringTimes`, { count: record.wagering_multiplier });
}

export type EnumParamGroup = keyof ActivityEnumGroups;

const PARAM_GROUPS = {
  applyMode: 'apply_mode',
  section: 'section',
  cycleMode: 'cycle_mode',
  payoutMode: 'payout_mode',
  payoutTiming: 'payout_timing',
  bonusMode: 'bonus_mode',
  betClaimMode: 'bet_claim_mode',
  wageringBase: 'wagering_base',
  wheelTrigger: 'wheel_trigger_mode',
  applyLimit: 'apply_limit_type',
  claimLimit: 'claim_limit_type',
  depositReward: 'deposit_reward_type',
} as const;

export type EnumGroup = keyof typeof PARAM_GROUPS;

/** 枚举参数接口返回的选项，接口未返回时为空 */
export function enumOptions(
  params: ActivityEnumParams | null,
  group: EnumParamGroup
): ActivityEnumOption[] {
  return params?.[group] ?? [];
}

/** 接口取不到对应文案时显示编号 */
export function enumLabel(
  params: ActivityEnumParams | null,
  group: EnumParamGroup,
  value: number
) {
  const option = params?.[group]?.find((item) => item.value === value);
  return option ? option.label : `#${value}`;
}

/** 无编码时显示为空 */
export function enumText(
  group: EnumGroup,
  code: number | null | undefined,
  params: ActivityEnumParams | null
) {
  if (!code) return '';
  return enumLabel(params, PARAM_GROUPS[group], code);
}

/** 1 首充、2 次存、3 累计、4 单笔；0 视为未设置 */
export function depositRewardText(
  code: number | null | undefined,
  params: ActivityEnumParams | null
) {
  return enumText('depositReward', code, params);
}

export function claimExpireText(sec: number | null | undefined, t: Translate) {
  if (!sec) return '';
  return sec % 86400 === 0
    ? t(`${PREFIX}.expireDays`, { count: sec / 86400 })
    : t(`${PREFIX}.expireSeconds`, { count: sec });
}

/** 除奖金模式外各类型特有的规则字段：投注的领取模式、转盘的触发方式 */
export function extraRuleItems(
  type: number,
  settings: ActivitySettings | null | undefined,
  t: Translate,
  params: ActivityEnumParams | null
) {
  if (type === ACTIVITY_TYPE.BET) {
    return [
      {
        label: t(`${PREFIX}.detail.betClaimMode`),
        value: enumText('betClaimMode', settings?.bet_claim_mode, params),
      },
    ];
  }
  if (type === ACTIVITY_TYPE.LUCK_WHEEL) {
    return [
      {
        label: t(`${PREFIX}.detail.wheelTrigger`),
        value: enumText(
          'wheelTrigger',
          settings?.luck_wheel?.wheel_trigger_mode,
          params
        ),
      },
    ];
  }
  return [];
}

/**
 * tier：达标阈值 → 奖励；refer：count 为邀请人数，threshold 不用；
 * prize：转盘奖品，rate 为开奖权重
 */
export type SchemaKind = 'tier' | 'refer' | 'prize';

export interface SchemaGroup {
  /** 只有一组时为空 */
  label: string;
  kind: SchemaKind;
  items: ActivityRewardSchema[];
}

export function rewardSchemaGroups(
  type: number,
  settings: ActivitySettings | null | undefined,
  t: Translate
): SchemaGroup[] {
  if (!settings) return [];
  let groups: SchemaGroup[];
  if (type === ACTIVITY_TYPE.LUCK_WHEEL) {
    groups = [
      {
        label: t(`${PREFIX}.schema.betTrigger`),
        kind: 'tier',
        items: settings.luck_wheel?.bet_schemas || [],
      },
      {
        label: t(`${PREFIX}.schema.depositTrigger`),
        kind: 'tier',
        items: settings.luck_wheel?.deposit_schemas || [],
      },
      {
        label: t(`${PREFIX}.schema.prizes`),
        kind: 'prize',
        items: settings.reward_schemas || [],
      },
    ];
  } else {
    groups = [
      {
        label: '',
        kind: type === ACTIVITY_TYPE.REFER ? 'refer' : 'tier',
        items: settings.reward_schemas || [],
      },
    ];
  }
  return groups.filter((group) => group.items.length);
}

export function schemaText(
  schema: ActivityRewardSchema,
  kind: SchemaKind,
  t: Translate
) {
  const parts: string[] = [];
  if (kind === 'refer') {
    parts.push(t(`${PREFIX}.detail.inviteCount`, { value: schema.count ?? 0 }));
  } else if (kind === 'tier' && schema.threshold) {
    parts.push(
      t(`${PREFIX}.detail.threshold`, {
        value: formatAmount(schema.threshold),
      })
    );
  }
  if (schema.amount) {
    parts.push(
      t(`${PREFIX}.detail.amount`, { value: formatAmount(schema.amount) })
    );
  }
  if (schema.rate) {
    parts.push(
      kind === 'prize'
        ? t(`${PREFIX}.detail.weight`, { value: schema.rate })
        : t(`${PREFIX}.detail.rate`, { value: schema.rate })
    );
  }
  if (schema.min_bonus) {
    parts.push(
      t(`${PREFIX}.detail.minBonus`, { value: formatAmount(schema.min_bonus) })
    );
  }
  if (schema.max_bonus) {
    parts.push(
      t(`${PREFIX}.detail.maxBonus`, { value: formatAmount(schema.max_bonus) })
    );
  }
  if (kind !== 'refer' && schema.count) {
    parts.push(t(`${PREFIX}.detail.count`, { value: schema.count }));
  }
  return parts.join('，');
}

export interface ActivityImage {
  url: string;
  lang: string;
}

/** 每种语言一张图，未上传图片的语言不显示 */
export function activityImages(
  i18n: ActivityI18nItem[] | null | undefined
): ActivityImage[] {
  return (i18n || [])
    .filter((item) => item.image)
    .map((item) => ({ url: item.image as string, lang: item.lang }));
}

/** 接口只返回操作人 UID，0 表示无 */
export function uidText(uid?: number | null) {
  return uid ? `UID ${uid}` : '';
}

/** 仅草稿可编辑，且弹窗需支持该活动类型 */
export function canEdit(record: ActivityRecord) {
  return (
    record.status === ACTIVITY_STATUS.DRAFT &&
    SUPPORTED_TYPE_CODES.includes(record.type)
  );
}

export function canPublish(record: ActivityRecord) {
  return record.status === ACTIVITY_STATUS.DRAFT;
}

export function canClose(record: ActivityRecord) {
  return record.status !== ACTIVITY_STATUS.ENDED;
}
