import { inject, provide, ref } from 'vue';
import type { InjectionKey, Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  closeActivity,
  getActivityDetail,
  getActivityEnumParams,
  publishActivity,
  queryActivityList,
  type ActivityDetail,
  type ActivityEnumParams,
  type ActivityListParams,
} from '@/api/activity';
import {
  queryAutoLevelList,
  queryFixedLevelList,
  type AutoLevelRecord,
  type FixedLevelRecord,
} from '@/api/member';
import { queryEnabledLangs } from '@/api/i18n';
import { GAME_STATUS, queryGameList, type GameRecord } from '@/api/game';
import { getCachedOperatorCode } from '@/utils/operator-site';
import { setSiteTimezone } from './site-time';
import type { ActivityQuery, ActivityRecord } from './types';

const LEVEL_PAGE_SIZE = 100;

export interface ActivityPage {
  current: number;
  pageSize: number;
}

export interface ActivityStore {
  /** 当前页活动 */
  activities: Ref<ActivityRecord[]>;
  /** 符合条件的活动总数 */
  total: Ref<number>;
  loading: Ref<boolean>;
  fetchActivities: (query: ActivityQuery, page: ActivityPage) => Promise<void>;
  /** 按上一次的条件重新拉取当前页 */
  reload: () => Promise<void>;
  /** 当前查看的活动详情，未加载或加载失败时为 null */
  detail: Ref<ActivityDetail | null>;
  detailLoading: Ref<boolean>;
  /** 传 null 时清空详情 */
  fetchDetail: (id: number | null) => Promise<void>;
  findActivity: (id: number | null) => ActivityRecord | null;
  /** 活动变更后刷新列表，该活动的详情已打开时一并刷新 */
  refresh: (id: number) => Promise<void>;
  /** 发布、关闭成功后会调用 refresh */
  publish: (id: number) => Promise<void>;
  close: (id: number) => Promise<void>;
  autoLevels: Ref<AutoLevelRecord[]>;
  fixedLevels: Ref<FixedLevelRecord[]>;
  /** 层级只加载一次（各取第一页，上限 100 个） */
  loadLevels: () => Promise<void>;
  /** 未限制层级时返回空数组 */
  levelNames: (record: ActivityRecord) => string[];
  levelGroups: (record: ActivityRecord) => { auto: string[]; fixed: string[] };
  /** 已启用的游戏，供新增弹窗选择和详情里把游戏 ID 翻译成名称；只加载一次 */
  games: Ref<GameRecord[]>;
  loadGames: () => Promise<void>;
  /** 取不到时显示为 ID */
  gameName: (id: number) => string;
  /** 语言码 → 显示名，只加载一次；取不到时原样显示语言码 */
  langLabel: (lang: string) => string;
  loadLangs: () => Promise<void>;
  /** 表单枚举选项，文案随 X-Lang；未加载或加载失败时为 null */
  enumParams: Ref<ActivityEnumParams | null>;
  /** 成功后只加载一次，失败的下次调用重试 */
  loadEnumParams: () => Promise<void>;
}

/** 用 Symbol.for 保证热更新重新执行本模块后 key 不变，否则已缓存的页面注入不到 */
const ACTIVITY_STORE_KEY: InjectionKey<ActivityStore> = Symbol.for(
  'activityCenter.store'
);

/** 详情没有 name 和 image，按当前语言从 i18n 中取，取不到时用第一条 */
function toRecord(detail: ActivityDetail, lang: string): ActivityRecord {
  const i18n = detail.i18n || [];
  const item = i18n.find((entry) => entry.lang === lang) || i18n[0];
  return {
    id: detail.id,
    type: detail.type,
    status: detail.status,
    name: item?.name || '',
    image: item?.image || '',
    allowed_clients: detail.allowed_clients || [],
    open_at_sec: detail.open_at_sec,
    close_at_sec: detail.close_at_sec,
    release_at_sec: detail.release_at_sec,
    sort_weight: detail.sort_weight,
    show_in_list: detail.show_in_list,
    section: detail.section,
    updated_at_sec: detail.updated_at_sec,
    auto_audit: detail.auto_audit,
    auto_audit_max_amt: detail.auto_audit_max_amt,
    allowed_fixed_level: detail.allowed_fixed_level || [],
    allowed_auto_level: detail.allowed_auto_level || [],
    wagering_multiplier: detail.wagering_multiplier,
    allowed_game_ids: detail.allowed_game_ids || [],
    is_permanent: detail.is_permanent,
    created_at_sec: detail.created_at_sec,
    create_uid: detail.create_uid,
    update_uid: detail.update_uid,
  };
}

function createActivityStore(): ActivityStore {
  const { locale } = useI18n();

  const activities = ref<ActivityRecord[]>([]);
  const total = ref(0);
  const loading = ref(false);
  let ticket = 0;
  let lastRequest: { query: ActivityQuery; page: ActivityPage } | null = null;

  const detail = ref<ActivityDetail | null>(null);
  const detailLoading = ref(false);
  let detailTicket = 0;

  async function searchById(id: number) {
    if (id <= 0) return [];
    try {
      const { data } = await getActivityDetail(id, true);
      return data ? [toRecord(data, locale.value)] : [];
    } catch {
      // 不存在的 ID 按无结果处理
      return [];
    }
  }

  async function fetchActivities(query: ActivityQuery, page: ActivityPage) {
    lastRequest = { query, page };
    ticket += 1;
    const current = ticket;
    loading.value = true;
    try {
      if (query.id != null) {
        const list = await searchById(query.id);
        if (current !== ticket) return;
        activities.value = list;
        total.value = list.length;
        return;
      }
      const params: ActivityListParams = {
        page: page.current,
        page_size: page.pageSize,
        name: query.name || undefined,
        type: query.type,
        client: query.client,
        status: query.status,
        lang: locale.value,
      };
      const { data } = await queryActivityList(params);
      if (current !== ticket) return;
      activities.value = data?.list || [];
      total.value = data?.total || 0;
    } catch {
      if (current !== ticket) return;
      activities.value = [];
      total.value = 0;
    } finally {
      if (current === ticket) loading.value = false;
    }
  }

  function reload() {
    if (!lastRequest) return Promise.resolve();
    return fetchActivities(lastRequest.query, lastRequest.page);
  }

  async function fetchDetail(id: number | null) {
    detailTicket += 1;
    const current = detailTicket;
    if (detail.value?.id !== id) detail.value = null;
    if (id == null) {
      detailLoading.value = false;
      return;
    }
    detailLoading.value = true;
    try {
      const { data } = await getActivityDetail(id);
      if (current !== detailTicket) return;
      detail.value = data || null;
    } catch {
      if (current !== detailTicket) return;
      detail.value = null;
    } finally {
      if (current === detailTicket) detailLoading.value = false;
    }
  }

  function findActivity(id: number | null) {
    if (id == null) return null;
    const record = activities.value.find((item) => item.id === id);
    if (record) return record;
    return detail.value?.id === id
      ? toRecord(detail.value, locale.value)
      : null;
  }

  async function refresh(id: number) {
    await Promise.all([
      reload(),
      detail.value?.id === id ? fetchDetail(id) : Promise.resolve(),
    ]);
  }

  async function publish(id: number) {
    await publishActivity(id);
    await refresh(id);
  }

  async function close(id: number) {
    await closeActivity(id);
    await refresh(id);
  }

  const autoLevels = ref<AutoLevelRecord[]>([]);
  const fixedLevels = ref<FixedLevelRecord[]>([]);
  let levelsRequest: Promise<void> | null = null;

  function loadLevels() {
    if (!levelsRequest) {
      levelsRequest = (async () => {
        const params = { page: 1, page_size: LEVEL_PAGE_SIZE };
        const [auto, fixed] = await Promise.allSettled([
          queryAutoLevelList(params),
          queryFixedLevelList(params),
        ]);
        if (auto.status === 'fulfilled') {
          autoLevels.value = auto.value.data?.list || [];
        }
        if (fixed.status === 'fulfilled') {
          fixedLevels.value = fixed.value.data?.list || [];
        }
        // 任一失败时允许下次重试
        if (auto.status === 'rejected' || fixed.status === 'rejected') {
          levelsRequest = null;
        }
      })();
    }
    return levelsRequest;
  }

  /** 已删除或不在第一页的层级取不到名称，显示为 ID */
  function levelGroups(record: ActivityRecord) {
    const lookup = (
      ids: number[] | null | undefined,
      levels: { id: number; name: string }[]
    ) =>
      (ids || []).map(
        (id) => levels.find((level) => level.id === id)?.name || `ID ${id}`
      );
    return {
      auto: lookup(record.allowed_auto_level, autoLevels.value),
      fixed: lookup(record.allowed_fixed_level, fixedLevels.value),
    };
  }

  function levelNames(record: ActivityRecord) {
    const { auto, fixed } = levelGroups(record);
    return [...auto, ...fixed];
  }

  const games = ref<GameRecord[]>([]);
  let gamesRequest: Promise<void> | null = null;
  const GAME_PAGE_SIZE = 100;
  /** 防止 total 异常时无限翻页 */
  const GAME_MAX_PAGES = 50;

  function loadGames() {
    if (!gamesRequest) {
      gamesRequest = (async () => {
        const list: GameRecord[] = [];
        try {
          for (let page = 1; page <= GAME_MAX_PAGES; page += 1) {
            // 逐页请求，后一页依赖前一页的数量判断是否结束
            // eslint-disable-next-line no-await-in-loop
            const { data } = await queryGameList({
              page,
              page_size: GAME_PAGE_SIZE,
              status: GAME_STATUS.ENABLED,
              sort_by: 'sort_no',
              sort_order: 'asc',
            });
            const rows = data?.list || [];
            list.push(...rows);
            if (
              rows.length < GAME_PAGE_SIZE ||
              list.length >= (data?.total ?? 0)
            ) {
              break;
            }
          }
          games.value = list;
        } catch {
          gamesRequest = null;
        }
      })();
    }
    return gamesRequest;
  }

  function gameName(id: number) {
    return games.value.find((game) => game.id === id)?.name || `ID ${id}`;
  }

  const langLabels = ref<Record<string, string>>({});
  let langsRequest: Promise<void> | null = null;

  function loadLangs() {
    if (!langsRequest) {
      langsRequest = (async () => {
        const operatorCode = getCachedOperatorCode();
        try {
          const { data } = await queryEnabledLangs({
            params: operatorCode ? { operator_code: operatorCode } : undefined,
            silentError: true,
          });
          langLabels.value = Object.fromEntries(
            (data?.list || []).map((item) => [
              item.lang,
              item.i18n_name || item.name,
            ])
          );
        } catch {
          langsRequest = null;
        }
      })();
    }
    return langsRequest;
  }

  function langLabel(lang: string) {
    return langLabels.value[lang] || lang;
  }

  const enumParams = ref<ActivityEnumParams | null>(null);
  let enumParamsRequest: Promise<void> | null = null;

  function loadEnumParams() {
    if (!enumParamsRequest) {
      enumParamsRequest = (async () => {
        try {
          const { data } = await getActivityEnumParams(true);
          enumParams.value = data || null;
          setSiteTimezone(data?.timezone);
        } catch {
          enumParamsRequest = null;
        }
      })();
    }
    return enumParamsRequest;
  }

  return {
    activities,
    total,
    loading,
    fetchActivities,
    reload,
    detail,
    detailLoading,
    fetchDetail,
    findActivity,
    refresh,
    publish,
    close,
    autoLevels,
    fixedLevels,
    loadLevels,
    levelNames,
    levelGroups,
    langLabel,
    loadLangs,
    enumParams,
    loadEnumParams,
    games,
    loadGames,
    gameName,
  };
}

export function provideActivityStore() {
  const store = createActivityStore();
  provide(ACTIVITY_STORE_KEY, store);
  return store;
}

export function useActivityStore() {
  const store = inject(ACTIVITY_STORE_KEY);
  if (!store) throw new Error('useActivityStore 需要在活动中心页面内使用');
  return store;
}
