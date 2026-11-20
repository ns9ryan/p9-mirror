import {
  ACTIVITY_APPLY_MODE,
  ACTIVITY_BONUS_MODE,
  ACTIVITY_SECTION,
  ACTIVITY_STATUS,
  ACTIVITY_TYPE,
  ACTIVITY_WHEEL_TRIGGER,
  type ActivityDetail,
  type ActivitySettings,
  type CreateActivityParams,
  type UpdateActivityParams,
} from '@/api/activity';
import { formatSiteTime, parseSiteTime } from './site-time';

/** 多选里代表「全部」的选项 */
export const ALL_OPTION = '全部';

/**
 * 新增弹窗里会提交给接口的字段。
 * 枚举字段直接存接口编码（见 @/api/activity 的 ACTIVITY_* 常量）；
 * 层级、奖金模式等仍为弹窗自身的标识
 */
export interface CreateFormInput {
  /** 活动类型编码（ACTIVITY_TYPE） */
  type: number;
  /** image 选填，未上传时为空字符串 */
  locales: {
    lang: string;
    name: string;
    /** 副标题，对应 i18n.description */
    description: string;
    content: string;
    image: string;
  }[];
  showInList: boolean;
  /** section：1 大厅、2 电子、3 体育 */
  section: number;
  timeType: string;
  timeRange: string[];
  /** deposit_reward_type */
  rechargeType?: number;
  /** cycle_mode */
  loopMode?: number;
  /** apply_mode */
  participationMode: number;
  /** payout_mode */
  distributionMode: number;
  /** payout_timing */
  distributionTime: number;
  claimExpireDays?: number;
  /** bet_claim_mode */
  claimMode?: number;
  memberLevels: string[];
  bonusMode: string;
  bonusTiers: { threshold?: number; reward?: number }[];
  /** 推荐会员：单独人数奖励 / 累计人数奖励 */
  referralRewardMode?: string;
  /** 单独人数奖励：每个新增会员赠送金额 */
  referralSingleAmount?: number;
  withdrawMultiple: string;
  /** wagering_base */
  giftMode: number;
  /** allowed_clients，选「全部」时带 ALL_OPTION */
  clientList: (number | typeof ALL_OPTION)[];
  /** apply_limit */
  activityRestrictions: number[];
  /** claim_limit */
  claimRestrictions: number[];
  inviteeRules: Record<
    'recharge' | 'turnover' | 'sameIp' | 'sameDevice',
    { checked: boolean; value?: number }
  >;
  drawConditions: string[];
  /** 幸运轮盘：有效投注达标阶梯 */
  wheelBetTiers: { threshold?: number }[];
  /** 幸运轮盘：充值达标阶梯 */
  wheelDepositTiers: { threshold?: number }[];
  /** 指定游戏开关 */
  specifiedGame: boolean;
  /** 开关打开时勾选的游戏 ID */
  gameIds: number[];
}

/** 当前活动类型下哪些表单项可见，隐藏的项不校验也不提交 */
export interface CreateFormVisibility {
  activityTime: boolean;
  specifiedGame: boolean;
  rechargeType: boolean;
  loopMode: boolean;
  participation: boolean;
  distribution: boolean;
  claimExpireDays: boolean;
  claimMode: boolean;
  bonusMode: boolean;
  memberLevels: boolean;
  withdraw: boolean;
  clientList: boolean;
  restrictions: boolean;
}

/** 弹窗的折叠区块，校验不通过时展开问题所在的区块 */
export type FormSection = 'basic' | 'settings' | 'rewards' | 'rules';

export interface FormError {
  section: FormSection;
  message: string;
}

const formError = (section: FormSection, message: string): FormError => ({
  section,
  message,
});

/** 弹窗的阶梯奖励是「达标后取最高一档」 */
const BONUS_MODES: Record<string, number> = {
  fixed: ACTIVITY_BONUS_MODE.FIXED_LADDER_MAX,
  ratio: ACTIVITY_BONUS_MODE.RATE_LADDER_MAX,
};

const WHEEL_BY_BET = '达标有效投注赠送1次机会';
const WHEEL_BY_DEPOSIT = '达标充值金额赠送1次机会';

/** 金额字段为 int64，币种精度未定，按输入值取整提交 */
const toInt = (value?: number) =>
  value == null ? undefined : Math.round(value);

/** 0 表示未设置，回填时保留弹窗默认值 */
const codeOrDefault = (code?: number | null) => code || undefined;

const findLabel = (table: Record<string, number>, code?: number | null) =>
  code == null
    ? undefined
    : Object.keys(table).find((label) => table[label] === code);

function levelIds(levels: string[]) {
  const pick = (prefix: string) =>
    levels
      .filter((value) => value.startsWith(prefix))
      .map((value) => Number(value.slice(prefix.length)));
  return { auto: pick('auto:'), fixed: pick('fixed:') };
}

/** 一个都没勾选时为 undefined */
function wheelTrigger(conditions: string[]) {
  const byBet = conditions.includes(WHEEL_BY_BET);
  const byDeposit = conditions.includes(WHEEL_BY_DEPOSIT);
  if (byBet && byDeposit) return ACTIVITY_WHEEL_TRIGGER.BET_OR_DEPOSIT;
  if (byBet) return ACTIVITY_WHEEL_TRIGGER.BET;
  if (byDeposit) return ACTIVITY_WHEEL_TRIGGER.DEPOSIT;
  return undefined;
}

/** 各活动类型专属表单项的处理，对应接口的 settings */
interface TypeHandler {
  validate: (
    form: CreateFormInput,
    show: CreateFormVisibility
  ) => FormError | null;
  toSettings: (
    form: CreateFormInput,
    show: CreateFormVisibility
  ) => ActivitySettings;
  /** 编辑回填 */
  fromSettings: (settings: ActivitySettings) => Partial<CreateFormInput>;
}

const depositHandler: TypeHandler = {
  validate: (form, show) =>
    show.bonusMode && !form.bonusMode
      ? formError('rewards', '请选择奖金模式')
      : null,

  toSettings: (form, show) => {
    const settings: ActivitySettings = {};
    if (show.rechargeType && form.rechargeType != null) {
      settings.deposit_reward_type = form.rechargeType;
    }
    if (show.bonusMode) {
      const ratio = form.bonusMode === 'ratio';
      settings.bonus_mode = BONUS_MODES[form.bonusMode];
      settings.reward_schemas = form.bonusTiers.map((tier, index) => ({
        id: index + 1,
        threshold: toInt(tier.threshold),
        ...(ratio
          ? { rate: toInt(tier.reward) }
          : { amount: toInt(tier.reward) }),
      }));
    }
    return settings;
  },

  fromSettings: (settings) => {
    const values: Partial<CreateFormInput> = {
      rechargeType: settings.deposit_reward_type,
    };
    const bonusMode = findLabel(BONUS_MODES, settings.bonus_mode);
    if (bonusMode) {
      values.bonusMode = bonusMode;
      values.bonusTiers = (settings.reward_schemas || []).map((schema) => ({
        threshold: schema.threshold,
        reward: bonusMode === 'ratio' ? schema.rate : schema.amount,
      }));
    }
    return values;
  },
};

const betHandler: TypeHandler = {
  validate: (form, show) => {
    if (show.claimMode && form.claimMode == null) {
      return formError('rewards', '请选择领取模式');
    }
    if (show.bonusMode && !form.bonusMode) {
      return formError('rewards', '请选择奖金模式');
    }
    return null;
  },

  toSettings: (form, show) => {
    const settings: ActivitySettings = {};
    if (show.claimMode && form.claimMode != null) {
      settings.bet_claim_mode = form.claimMode;
    }
    if (show.bonusMode) {
      const ratio = form.bonusMode === 'ratio';
      settings.bonus_mode = BONUS_MODES[form.bonusMode];
      settings.reward_schemas = form.bonusTiers.map((tier, index) => ({
        id: index + 1,
        threshold: toInt(tier.threshold),
        ...(ratio
          ? { rate: toInt(tier.reward) }
          : { amount: toInt(tier.reward) }),
      }));
    }
    return settings;
  },

  fromSettings: (settings) => {
    const values: Partial<CreateFormInput> = {
      claimMode: codeOrDefault(settings.bet_claim_mode),
    };
    const bonusMode = findLabel(BONUS_MODES, settings.bonus_mode);
    if (bonusMode) {
      values.bonusMode = bonusMode;
      values.bonusTiers = (settings.reward_schemas || []).map((schema) => ({
        threshold: schema.threshold,
        reward: bonusMode === 'ratio' ? schema.rate : schema.amount,
      }));
    }
    return values;
  },
};

const referHandler: TypeHandler = {
  validate: (form) => {
    if (!form.referralRewardMode) {
      return formError('settings', '请选择奖励模式');
    }
    if (
      form.referralRewardMode === '单独人数奖励' &&
      form.referralSingleAmount == null
    ) {
      return formError('settings', '请填写奖励模式具体金额');
    }
    if (form.referralRewardMode === '累计人数奖励') {
      if (
        !form.bonusTiers.length ||
        form.bonusTiers.some(
          (tier) => tier.threshold == null || tier.reward == null
        )
      ) {
        return formError('settings', '请填写奖励模式具体金额');
      }
    }
    if (
      Object.values(form.inviteeRules).some(
        (rule) => rule.checked && rule.value == null
      )
    ) {
      return formError('settings', '请填写已勾选的被邀请会员有效判定数值');
    }
    return null;
  },

  toSettings: (form) => {
    const rules = form.inviteeRules;
    const value = (key: keyof CreateFormInput['inviteeRules']) =>
      rules[key].checked ? toInt(rules[key].value) : undefined;
    const settings: ActivitySettings = {
      refer_config: {
        min_deposit_amount: value('recharge'),
        min_bet_amount: value('turnover'),
        ip_limit: value('sameIp'),
        device_limit: value('sameDevice'),
      },
    };
    if (form.referralRewardMode === '单独人数奖励') {
      settings.bonus_mode = ACTIVITY_BONUS_MODE.SINGLE;
      settings.reward_schemas = [
        { id: 1, amount: toInt(form.referralSingleAmount) },
      ];
    } else if (form.referralRewardMode === '累计人数奖励') {
      settings.bonus_mode = ACTIVITY_BONUS_MODE.FIXED_LADDER_MAX;
      settings.reward_schemas = form.bonusTiers.map((tier, index) => ({
        id: index + 1,
        threshold: toInt(tier.threshold),
        amount: toInt(tier.reward),
      }));
    }
    return settings;
  },

  fromSettings: (settings) => {
    const values: Partial<CreateFormInput> = {};
    const refer = settings.refer_config;
    if (refer) {
      const rule = (value?: number) =>
        value ? { checked: true, value } : { checked: false };
      values.inviteeRules = {
        recharge: rule(refer.min_deposit_amount),
        turnover: rule(refer.min_bet_amount),
        sameIp: rule(refer.ip_limit),
        sameDevice: rule(refer.device_limit),
      };
    }
    if (settings.bonus_mode === ACTIVITY_BONUS_MODE.SINGLE) {
      values.referralRewardMode = '单独人数奖励';
      values.referralSingleAmount = settings.reward_schemas?.[0]?.amount;
    } else if (settings.bonus_mode === ACTIVITY_BONUS_MODE.FIXED_LADDER_MAX) {
      values.referralRewardMode = '累计人数奖励';
      values.bonusTiers = (settings.reward_schemas || []).map((schema) => ({
        threshold: schema.threshold,
        reward: schema.amount,
      }));
    }
    return values;
  },
};

const toWheelSchemas = (tiers: { threshold?: number }[]) =>
  tiers.map((tier, index) => ({
    id: index + 1,
    threshold: toInt(tier.threshold),
    count: 1,
  }));

const incompleteWheelTiers = (tiers: { threshold?: number }[]) =>
  !tiers.length || tiers.some((tier) => tier.threshold == null);

const luckWheelHandler: TypeHandler = {
  validate: (form) => {
    if (!wheelTrigger(form.drawConditions)) {
      return formError('settings', '请选择抽奖需要条件');
    }
    if (
      form.drawConditions.includes(WHEEL_BY_BET) &&
      incompleteWheelTiers(form.wheelBetTiers)
    ) {
      return formError('settings', '请填写有效投注达标金额');
    }
    if (
      form.drawConditions.includes(WHEEL_BY_DEPOSIT) &&
      incompleteWheelTiers(form.wheelDepositTiers)
    ) {
      return formError('settings', '请填写充值达标金额');
    }
    return null;
  },

  toSettings: (form) => ({
    luck_wheel: {
      wheel_trigger_mode: wheelTrigger(form.drawConditions),
      bet_schemas: form.drawConditions.includes(WHEEL_BY_BET)
        ? toWheelSchemas(form.wheelBetTiers)
        : [],
      deposit_schemas: form.drawConditions.includes(WHEEL_BY_DEPOSIT)
        ? toWheelSchemas(form.wheelDepositTiers)
        : [],
    },
  }),

  fromSettings: (settings) => {
    const luck = settings.luck_wheel;
    const trigger = luck?.wheel_trigger_mode;
    if (!trigger) return {};
    const byBet = trigger !== ACTIVITY_WHEEL_TRIGGER.DEPOSIT;
    const byDeposit = trigger !== ACTIVITY_WHEEL_TRIGGER.BET;
    return {
      drawConditions: [
        ...(byBet ? [WHEEL_BY_BET] : []),
        ...(byDeposit ? [WHEEL_BY_DEPOSIT] : []),
      ],
      wheelBetTiers: (luck?.bet_schemas || []).map((schema) => ({
        threshold: schema.threshold,
      })),
      wheelDepositTiers: (luck?.deposit_schemas || []).map((schema) => ({
        threshold: schema.threshold,
      })),
    };
  },
};

/** 弹窗能新建、编辑的类型及其处理；其余类型的专属配置接口暂未提供字段 */
const TYPE_HANDLERS: Record<number, TypeHandler> = {
  [ACTIVITY_TYPE.DEPOSIT]: depositHandler,
  [ACTIVITY_TYPE.BET]: betHandler,
  [ACTIVITY_TYPE.REFER]: referHandler,
  [ACTIVITY_TYPE.LUCK_WHEEL]: luckWheelHandler,
};

export const SUPPORTED_TYPE_CODES = Object.keys(TYPE_HANDLERS).map(Number);

/** 提交前校验，返回第一个问题；全部通过时返回 null */
export function validateCreateForm(
  form: CreateFormInput,
  show: CreateFormVisibility
): FormError | null {
  const handler = TYPE_HANDLERS[form.type];
  if (!handler) {
    return formError(
      'basic',
      '该活动类型暂未开放，目前仅支持充值、洗码、推荐会员、幸运轮盘'
    );
  }

  if (
    form.section !== ACTIVITY_SECTION.LOBBY &&
    form.section !== ACTIVITY_SECTION.ELECTRONIC &&
    form.section !== ACTIVITY_SECTION.SPORTS
  ) {
    return formError('settings', '请选择活动所属');
  }

  if (show.activityTime && form.timeType !== 'permanent') {
    const [start, end] = form.timeRange || [];
    if (!start || !end) return formError('settings', '请选择活动时间');
    if (parseSiteTime(end) <= parseSiteTime(start)) {
      return formError('settings', '活动结束时间需晚于开始时间');
    }
  }

  if (show.memberLevels && !form.memberLevels.length) {
    return formError('settings', '请选择会员层级');
  }

  if (show.specifiedGame && form.specifiedGame && !form.gameIds.length) {
    return formError(
      'rewards',
      '请选择指定洗码游戏，或关闭「按平台定向选择」'
    );
  }

  if (
    show.clientList &&
    !form.clientList.some((client) => client !== ALL_OPTION)
  ) {
    return formError('rewards', '请选择参与客户端');
  }

  if (show.withdraw && !/^\d*$/.test(form.withdrawMultiple.trim())) {
    return formError('rewards', '洗码倍数需为非负整数');
  }

  return handler.validate(form, show);
}

/**
 * 按弹窗表单生成创建活动的请求体，一律保存为草稿。
 * 只做格式转换，调用前需先通过 validateCreateForm
 */
export default function buildCreatePayload(
  form: CreateFormInput,
  show: CreateFormVisibility
): CreateActivityParams {
  const data: CreateActivityParams = {
    type: form.type,
    status: ACTIVITY_STATUS.DRAFT,
    apply_mode: show.participation
      ? form.participationMode
      : ACTIVITY_APPLY_MODE.AUTO,
    section: form.section,
    show_in_list: form.showInList,
    i18n: form.locales.map((item) => ({
      lang: item.lang,
      name: item.name.trim(),
      description: item.description.trim() || undefined,
      rule: item.content,
      image: item.image || undefined,
    })),
  };

  if (show.activityTime) {
    data.is_permanent = form.timeType === 'permanent';
    if (!data.is_permanent) {
      // 时间选择器里填的是本站时区的时间
      data.open_at_sec = parseSiteTime(form.timeRange[0]);
      data.close_at_sec = parseSiteTime(form.timeRange[1]);
    }
  }

  // 未打开指定游戏时不传，表示全部游戏
  if (show.specifiedGame && form.specifiedGame) {
    data.allowed_game_ids = form.gameIds;
  }

  if (show.loopMode && form.loopMode != null) {
    data.cycle_mode = form.loopMode;
  }

  if (show.distribution) {
    data.payout_mode = form.distributionMode;
    data.payout_timing = form.distributionTime;
  }

  if (show.claimExpireDays && form.claimExpireDays) {
    data.claim_expire_sec = form.claimExpireDays * 86400;
  }

  // 选「全部」时不传层级，表示不限制
  if (show.memberLevels && !form.memberLevels.includes(ALL_OPTION)) {
    const ids = levelIds(form.memberLevels);
    data.allowed_auto_level = ids.auto;
    data.allowed_fixed_level = ids.fixed;
  }

  if (show.clientList) {
    data.allowed_clients = form.clientList.filter(
      (client): client is number => client !== ALL_OPTION
    );
  }

  if (show.restrictions) {
    data.apply_limit = form.activityRestrictions;
    data.claim_limit = form.claimRestrictions;
  }

  if (show.withdraw) {
    const multiple = form.withdrawMultiple.trim();
    // 编辑时清空输入需要提交 0，否则会保留原倍数
    data.wagering_multiplier = multiple ? Number(multiple) : 0;
    data.wagering_base = form.giftMode;
  }

  data.settings = TYPE_HANDLERS[form.type]?.toSettings(form, show) ?? {};

  return data;
}

/**
 * 编辑时以详情为底，弹窗不管理的字段（代理、排序、自动审核等）原样带回，避免被清空；
 * 层级和游戏在弹窗里选「全部」时不会出现在请求体里，需显式传空数组才能取消限制
 */
export function buildUpdatePayload(
  detail: ActivityDetail,
  data: CreateActivityParams
): UpdateActivityParams {
  const defined = Object.fromEntries(
    Object.entries(data).filter(
      ([key, value]) => key !== 'status' && value !== undefined
    )
  );
  return {
    type: detail.type,
    apply_mode: detail.apply_mode,
    section: detail.section,
    cycle_mode: detail.cycle_mode,
    allowed_clients: detail.allowed_clients || [],
    allowed_agent_ids: detail.allowed_agent_ids || [],
    apply_limit: detail.apply_limit || [],
    claim_limit: detail.claim_limit || [],
    is_permanent: detail.is_permanent,
    open_at_sec: detail.open_at_sec ?? undefined,
    close_at_sec: detail.close_at_sec ?? undefined,
    release_at_sec: detail.release_at_sec ?? undefined,
    auto_audit: detail.auto_audit,
    auto_audit_max_amt: detail.auto_audit_max_amt,
    payout_mode: detail.payout_mode,
    payout_timing: detail.payout_timing,
    claim_expire_sec: detail.claim_expire_sec,
    wagering_multiplier: detail.wagering_multiplier,
    wagering_base: detail.wagering_base,
    show_in_sidebar: detail.show_in_sidebar,
    sort_weight: detail.sort_weight,
    allowed_auto_level: [],
    allowed_fixed_level: [],
    allowed_game_ids: [],
    ...defined,
    id: detail.id,
  };
}

export type DetailFormResult =
  | { ok: true; type: number; values: Partial<CreateFormInput> }
  | { ok: false; message: string };

/**
 * buildCreatePayload 的逆过程，供编辑时回填弹窗；取不到对应选项的字段不返回，保留弹窗默认值。
 * getLoopOptions 返回该类型可选的循环模式，不在其中的循环模式不回填
 */
export function detailToForm(
  detail: ActivityDetail,
  getLoopOptions: (type: number) => number[]
): DetailFormResult {
  const { type } = detail;
  const handler = TYPE_HANDLERS[type];
  if (!handler) {
    return {
      ok: false,
      message: '该活动类型暂不支持编辑，目前仅支持充值、洗码、推荐会员、幸运轮盘',
    };
  }
  const gameIds = detail.allowed_game_ids || [];
  const autoLevels = detail.allowed_auto_level || [];
  const fixedLevels = detail.allowed_fixed_level || [];
  const clients = detail.allowed_clients || [];
  const hasRange =
    !detail.is_permanent && !!detail.open_at_sec && !!detail.close_at_sec;

  const values: Partial<CreateFormInput> = {
    locales: (detail.i18n || []).map((item) => ({
      lang: item.lang,
      name: item.name || '',
      description: item.description || '',
      content: item.rule || '',
      image: item.image || '',
    })),
    showInList: detail.show_in_list,
    section: codeOrDefault(detail.section) ?? ACTIVITY_SECTION.LOBBY,
    timeType: hasRange ? 'range' : 'permanent',
    timeRange: hasRange
      ? [
          formatSiteTime(detail.open_at_sec as number),
          formatSiteTime(detail.close_at_sec as number),
        ]
      : [],
    specifiedGame: gameIds.length > 0,
    gameIds,
    loopMode:
      detail.cycle_mode != null &&
      getLoopOptions(type).includes(detail.cycle_mode)
        ? detail.cycle_mode
        : undefined,
    participationMode: codeOrDefault(detail.apply_mode),
    distributionMode: codeOrDefault(detail.payout_mode),
    distributionTime: codeOrDefault(detail.payout_timing),
    claimExpireDays: detail.claim_expire_sec
      ? Math.max(1, Math.round(detail.claim_expire_sec / 86400))
      : undefined,
    memberLevels:
      autoLevels.length || fixedLevels.length
        ? [
            ...autoLevels.map((id) => `auto:${id}`),
            ...fixedLevels.map((id) => `fixed:${id}`),
          ]
        : [ALL_OPTION],
    withdrawMultiple: detail.wagering_multiplier
      ? String(detail.wagering_multiplier)
      : '',
    giftMode: codeOrDefault(detail.wagering_base),
    clientList: clients.length ? clients : [ALL_OPTION],
    activityRestrictions: detail.apply_limit || [],
    claimRestrictions: detail.claim_limit || [],
    ...handler.fromSettings(detail.settings || {}),
  };

  return {
    ok: true,
    type,
    values: Object.fromEntries(
      Object.entries(values).filter(([, value]) => value !== undefined)
    ),
  };
}
