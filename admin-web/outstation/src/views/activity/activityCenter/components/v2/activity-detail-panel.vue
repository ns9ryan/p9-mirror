<template>
  <aside class="detail-panel">
    <div class="panel-head">
      <div class="panel-title">
        <span>{{ t('discountCenter.activityCenter.v2.panel.title') }}</span>
        <a-button
          type="text"
          size="mini"
          :aria-label="t('discountCenter.activityCenter.v2.panel.close')"
          @click="emit('close')"
        >
          <template #icon><icon-close /></template>
        </a-button>
      </div>

      <div class="profile">
        <a-image
          class="cover"
          :src="record.image || imageSrc"
          :width="52"
          :height="52"
          :preview="false"
          fit="cover"
        />
        <div class="profile-main">
          <div class="profile-name">
            <strong class="ellipsis" :title="record.name">
              {{ displayText(record.name) }}
            </strong>
            <a-tag size="small" :color="statusColor(record.status)">
              {{ statusText(record.status, enumParams) }}
            </a-tag>
          </div>
          <div class="profile-sub ellipsis">
            {{ typeText(record.type, enumParams) }}
          </div>
          <div class="profile-sub num">
            {{ t('discountCenter.activityCenter.columns.activityId') }}:
            {{ record.id }}
          </div>
        </div>
      </div>

      <div
        v-if="canEdit(record) || canPublish(record) || canClose(record)"
        class="profile-actions"
      >
        <a-button
          v-if="canEdit(record)"
          size="small"
          @click="emit('action', 'edit')"
        >
          {{ t('discountCenter.activityCenter.card.edit') }}
        </a-button>
        <a-button
          v-if="canPublish(record)"
          type="primary"
          size="small"
          @click="emit('action', 'publish')"
        >
          {{ t('discountCenter.activityCenter.card.publish') }}
        </a-button>
        <a-button
          v-if="canClose(record)"
          size="small"
          status="danger"
          @click="emit('action', 'close')"
        >
          {{ t('discountCenter.activityCenter.card.close') }}
        </a-button>
      </div>

      <a-tabs
        class="panel-tabs"
        size="small"
        :active-key="activeSection"
        @change="scrollToSection"
      >
        <a-tab-pane
          v-for="item in SECTIONS"
          :key="item"
          :title="t(`discountCenter.activityCenter.v2.section.${item}`)"
        />
      </a-tabs>
    </div>

    <a-spin :loading="loading" class="panel-spin">
      <div ref="bodyRef" class="panel-body" @scroll="onBodyScroll">
        <section data-section="setting" class="panel-card">
          <h3>{{ t('discountCenter.activityCenter.v2.section.setting') }}</h3>
          <dl class="info-list">
            <template v-for="item in settingItems" :key="item.label">
              <dt>{{ item.label }}</dt>
              <dd :title="item.value">{{ item.value }}</dd>
            </template>
          </dl>
        </section>

        <section data-section="bonus" class="panel-card">
          <h3>{{ t('discountCenter.activityCenter.v2.section.bonus') }}</h3>
          <div class="stat-grid">
            <div v-for="item in bonusItems" :key="item.label" class="stat-item">
              <span class="stat-icon" :class="`is-${item.tone}`">
                <component :is="item.icon" />
              </span>
              <div class="stat-main">
                <span class="stat-label">{{ item.label }}</span>
                <strong
                  class="num ellipsis"
                  :class="{ 'is-wash': item.highlight }"
                  :title="item.value"
                >
                  {{ item.value }}
                </strong>
              </div>
            </div>
          </div>
          <dl class="info-list payout-list">
            <template v-for="item in payoutItems" :key="item.label">
              <dt>{{ item.label }}</dt>
              <dd :title="item.value">{{ item.value }}</dd>
            </template>
          </dl>
          <div
            v-for="group in schemaGroups"
            :key="group.label || group.kind"
            class="tag-group schema-group"
          >
            <span class="tag-label">
              {{
                group.label ||
                t('discountCenter.activityCenter.detail.rewardSchemas')
              }}
            </span>
            <ol class="schema-list">
              <li v-for="(item, index) in group.items" :key="index">
                {{ schemaText(item, group.kind, t) }}
              </li>
            </ol>
          </div>
        </section>

        <section data-section="participation" class="panel-card">
          <h3>{{
            t('discountCenter.activityCenter.v2.section.participation')
          }}</h3>
          <div v-for="group in tagGroups" :key="group.label" class="tag-group">
            <span class="tag-label">{{ group.label }}</span>
            <a-space wrap :size="6">
              <a-tag
                v-for="item in group.items"
                :key="item"
                size="small"
                :color="group.color"
              >
                {{ item }}
              </a-tag>
            </a-space>
          </div>
        </section>

        <section data-section="content" class="panel-card">
          <h3>{{ t('discountCenter.activityCenter.v2.section.content') }}</h3>
          <ActivityLocales v-if="info?.i18n?.length" :items="info.i18n" />
          <span v-else class="tag-label">---</span>
          <div v-if="images.length" class="tag-group image-group">
            <span class="tag-label">
              {{ t('discountCenter.activityCenter.expand.images') }}
            </span>
            <a-image-preview-group infinite>
              <div class="image-list">
                <figure
                  v-for="image in images"
                  :key="image.lang"
                  class="image-figure"
                >
                  <a-image
                    class="content-image"
                    :src="image.url"
                    :width="160"
                    :height="60"
                    fit="cover"
                  />
                  <figcaption>{{ langLabel(image.lang) }}</figcaption>
                </figure>
              </div>
            </a-image-preview-group>
          </div>
        </section>

        <section data-section="record" class="panel-card">
          <h3>{{ t('discountCenter.activityCenter.v2.section.record') }}</h3>
          <dl class="info-list">
            <template v-for="item in recordItems" :key="item.label">
              <dt>{{ item.label }}</dt>
              <dd :title="item.value">{{ item.value }}</dd>
            </template>
          </dl>
        </section>
      </div>
    </a-spin>
  </aside>
</template>

<script lang="ts" setup>
  import { computed, nextTick, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { ACTIVITY_TYPE, type ActivityDetail } from '@/api/activity';
  import {
    activityImages,
    activityTimeText,
    autoAuditText,
    canClose,
    canEdit,
    canPublish,
    claimExpireText,
    clientNames,
    depositRewardText,
    displayText,
    enumText,
    extraRuleItems,
    formatAmount,
    formatTime,
    rewardSchemaGroups,
    schemaText,
    statusColor,
    statusText,
    typeText,
    uidText,
    wageringText,
    type EnumGroup,
  } from '../../options';
  import { useActivityStore } from '../../store';
  import type { ActivityAction, ActivityRecord } from '../../types';
  import ActivityLocales from '../activity-locales.vue';

  const props = defineProps<{
    record: ActivityRecord;
    /** 详情接口返回的完整配置，加载中或失败时为 null */
    detail: ActivityDetail | null;
    loading?: boolean;
    /** 活动没有图片时的占位图 */
    imageSrc: string;
  }>();

  const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'action', action: ActivityAction): void;
  }>();

  const { t } = useI18n();
  const {
    levelNames,
    langLabel,
    loadLangs,
    gameName,
    loadGames,
    enumParams,
    loadEnumParams,
  } = useActivityStore();

  loadLangs();
  loadEnumParams();

  const enumValue = (group: EnumGroup, code: number | null | undefined) =>
    enumText(group, code, enumParams.value);

  // 游戏较多，只在当前活动限定了游戏时才加载
  watch(
    () => props.record.allowed_game_ids?.length,
    (count) => {
      if (count) loadGames();
    },
    { immediate: true }
  );

  const PREFIX = 'discountCenter.activityCenter';

  type SectionKey =
    | 'setting'
    | 'bonus'
    | 'participation'
    | 'content'
    | 'record';

  const SECTIONS: SectionKey[] = [
    'setting',
    'bonus',
    'participation',
    'content',
    'record',
  ];

  const bodyRef = ref<HTMLElement>();
  const activeSection = ref<SectionKey>('setting');

  /** 详情与当前记录不是同一条时视为未加载 */
  const info = computed(() =>
    props.detail?.id === props.record.id ? props.detail : null
  );

  const yesNo = (value?: boolean) =>
    value == null ? '' : t(`${PREFIX}.${value ? 'yes' : 'no'}`);

  const withDisplay = (items: { label: string; value: string }[]) =>
    items.map((item) => ({ ...item, value: displayText(item.value) }));

  const settingItems = computed(() => {
    const { record } = props;
    return withDisplay([
      {
        label: t(`${PREFIX}.columns.activityName`),
        value: record.name,
      },
      {
        label: t(`${PREFIX}.card.activityType`),
        value: typeText(record.type, enumParams.value),
      },
      {
        label: t(`${PREFIX}.detail.applyMode`),
        value: enumValue('applyMode', info.value?.apply_mode),
      },
      {
        label: t(`${PREFIX}.detail.section`),
        value: enumValue('section', info.value?.section ?? record.section),
      },
      {
        label: t(`${PREFIX}.detail.cycleMode`),
        value: enumValue('cycleMode', info.value?.cycle_mode),
      },
      {
        label: t(`${PREFIX}.columns.activityTime`),
        value: activityTimeText(record, t),
      },
      {
        label: t(`${PREFIX}.card.releaseAt`),
        value: formatTime(record.release_at_sec),
      },
      {
        label: t(`${PREFIX}.columns.listed`),
        value: yesNo(record.show_in_list),
      },
      {
        label: t(`${PREFIX}.detail.showInSidebar`),
        value: yesNo(info.value?.show_in_sidebar),
      },
      {
        label: t(`${PREFIX}.columns.sort`),
        value: String(record.sort_weight),
      },
    ]);
  });

  const bonusItems = computed(() => {
    const { record } = props;
    return [
      {
        label: t(`${PREFIX}.columns.washMultiplier`),
        value: displayText(wageringText(record, t)),
        icon: 'icon-sync',
        tone: 'red',
        highlight: !!record.wagering_multiplier,
      },
      {
        label: t(`${PREFIX}.detail.wageringBase`),
        value: displayText(
          enumValue('wageringBase', info.value?.wagering_base)
        ),
        icon: 'icon-layers',
        tone: 'green',
      },
      {
        label: t(`${PREFIX}.card.bonusMode`),
        value: displayText(
          enumValue('bonusMode', info.value?.settings?.bonus_mode)
        ),
        icon: 'icon-gift',
        tone: 'orange',
      },
      {
        label: t(`${PREFIX}.columns.autoReviewConfig`),
        value: autoAuditText(record, t),
        icon: 'icon-check-circle',
        tone: 'blue',
      },
    ];
  });

  const payoutItems = computed(() => {
    const settings = info.value?.settings;
    const items = [
      {
        label: t(`${PREFIX}.detail.payoutMode`),
        value: enumValue('payoutMode', info.value?.payout_mode),
      },
      {
        label: t(`${PREFIX}.detail.payoutTiming`),
        value: enumValue('payoutTiming', info.value?.payout_timing),
      },
      {
        label: t(`${PREFIX}.detail.claimExpire`),
        value: claimExpireText(info.value?.claim_expire_sec, t),
      },
      ...extraRuleItems(props.record.type, settings, t, enumParams.value),
    ];
    if (props.record.type === ACTIVITY_TYPE.DEPOSIT) {
      items.push({
        label: t(`${PREFIX}.detail.depositRewardType`),
        value: depositRewardText(
          settings?.deposit_reward_type,
          enumParams.value
        ),
      });
    }
    const refer = settings?.refer_config;
    if (props.record.type === ACTIVITY_TYPE.REFER && refer) {
      items.push(
        {
          label: t(`${PREFIX}.detail.referMinDeposit`),
          value: formatAmount(refer.min_deposit_amount),
        },
        {
          label: t(`${PREFIX}.detail.referMinBet`),
          value: formatAmount(refer.min_bet_amount),
        },
        {
          label: t(`${PREFIX}.detail.referIpLimit`),
          value: refer.ip_limit ? String(refer.ip_limit) : '',
        },
        {
          label: t(`${PREFIX}.detail.referDeviceLimit`),
          value: refer.device_limit ? String(refer.device_limit) : '',
        }
      );
    }
    return withDisplay(items);
  });

  const schemaGroups = computed(() =>
    rewardSchemaGroups(props.record.type, info.value?.settings, t)
  );

  const images = computed(() => activityImages(info.value?.i18n));

  const allText = computed(() => t(`${PREFIX}.all`));

  const tagGroups = computed(() => {
    const { record } = props;
    const orAll = (items: string[]) => (items.length ? items : [allText.value]);
    const groups = [
      {
        label: t(`${PREFIX}.card.client`),
        items: orAll(clientNames(record.allowed_clients || [], enumParams.value)),
        color: undefined as string | undefined,
      },
      {
        label: t(`${PREFIX}.card.participantLevel`),
        items: orAll(levelNames(record)),
        color: 'green',
      },
      {
        label: t(`${PREFIX}.card.games`),
        items: orAll((record.allowed_game_ids || []).map(gameName)),
        color: 'arcoblue',
      },
    ];
    const agents = (info.value?.allowed_agent_ids || []).map(
      (id) => `ID ${id}`
    );
    if (agents.length) {
      groups.push({
        label: t(`${PREFIX}.detail.agents`),
        items: agents,
        color: 'cyan',
      });
    }
    const applyLimit = (info.value?.apply_limit || []).map((code) =>
      enumValue('applyLimit', code)
    );
    const claimLimit = (info.value?.claim_limit || []).map((code) =>
      enumValue('claimLimit', code)
    );
    if (applyLimit.length) {
      groups.push({
        label: t(`${PREFIX}.detail.applyLimit`),
        items: applyLimit,
        color: 'orange',
      });
    }
    if (claimLimit.length) {
      groups.push({
        label: t(`${PREFIX}.detail.claimLimit`),
        items: claimLimit,
        color: 'orange',
      });
    }
    return groups;
  });

  const recordItems = computed(() => {
    const { record } = props;
    return withDisplay([
      {
        label: t(`${PREFIX}.card.createdBy`),
        value: uidText(record.create_uid),
      },
      {
        label: t(`${PREFIX}.card.createdAt`),
        value: formatTime(record.created_at_sec),
      },
      {
        label: t(`${PREFIX}.card.updatedBy`),
        value: uidText(record.update_uid),
      },
      {
        label: t(`${PREFIX}.card.updatedAt`),
        value: formatTime(record.updated_at_sec),
      },
      {
        label: t(`${PREFIX}.detail.version`),
        value: info.value ? String(info.value.version) : '',
      },
    ]);
  });

  function sectionEl(key: SectionKey) {
    return bodyRef.value?.querySelector<HTMLElement>(`[data-section="${key}"]`);
  }

  function scrollToSection(key: string | number) {
    const el = sectionEl(key as SectionKey);
    if (!bodyRef.value || !el) return;
    activeSection.value = key as SectionKey;
    bodyRef.value.scrollTo({ top: el.offsetTop - 12, behavior: 'smooth' });
  }

  function onBodyScroll() {
    const body = bodyRef.value;
    if (!body) return;
    const atBottom =
      body.scrollTop + body.clientHeight >= body.scrollHeight - 4;
    if (atBottom) {
      activeSection.value = SECTIONS[SECTIONS.length - 1];
      return;
    }
    const current = SECTIONS.filter((key) => {
      const el = sectionEl(key);
      return el && el.offsetTop - 24 <= body.scrollTop;
    }).pop();
    activeSection.value = current || 'setting';
  }

  watch(
    () => props.record.id,
    async () => {
      activeSection.value = 'setting';
      await nextTick();
      bodyRef.value?.scrollTo({ top: 0 });
    }
  );
</script>

<style scoped lang="less">
  .detail-panel {
    display: flex;
    flex-direction: column;
    min-height: 0;
    overflow: hidden;
    background: var(--color-bg-2);
    border-radius: 4px;
  }

  .panel-head {
    flex: none;
    padding: 16px 20px 0;
    border-bottom: 1px solid var(--color-neutral-3);
  }

  .panel-title {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 16px;
  }

  .profile {
    display: flex;
    gap: 14px;
    align-items: center;
  }

  .profile :deep(.cover) {
    flex: none;
    overflow: hidden;
    border-radius: 8px;

    .arco-image-img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
  }

  .profile-main {
    flex: 1;
    min-width: 0;
  }

  .profile-name {
    display: flex;
    gap: 8px;
    align-items: center;
    min-width: 0;

    strong {
      min-width: 0;
      color: var(--color-text-1);
      font-size: 17px;
    }
  }

  .profile-sub {
    margin-top: 2px;
    color: var(--color-text-3);
    font-size: 12px;
  }

  .profile-actions {
    display: flex;
    gap: 8px;
    margin-top: 14px;
  }

  .panel-tabs {
    margin-top: 8px;

    :deep(.arco-tabs-content) {
      display: none;
    }

    :deep(.arco-tabs-nav::before) {
      display: none;
    }
  }

  .panel-spin {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
  }

  .panel-body {
    position: relative;
    flex: 1;
    min-height: 0;
    padding: 12px;
    overflow-y: auto;
    background: var(--color-fill-1);
  }

  .panel-card {
    padding: 16px;
    background: var(--color-bg-2);
    border-radius: 6px;

    & + & {
      margin-top: 12px;
    }

    h3 {
      margin: 0 0 12px;
      color: var(--color-text-1);
      font-weight: 600;
      font-size: 14px;
    }
  }

  .info-list {
    display: grid;
    grid-template-columns: 96px minmax(0, 1fr);
    row-gap: 10px;
    column-gap: 12px;
    align-items: center;
    margin: 0;
    font-size: 13px;

    dt {
      color: var(--color-text-3);
    }

    dd {
      margin: 0;
      overflow: hidden;
      color: var(--color-text-1);
      white-space: nowrap;
      text-overflow: ellipsis;
    }
  }

  .payout-list {
    margin-top: 14px;
  }

  .stat-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
  }

  .stat-item {
    display: flex;
    gap: 10px;
    min-width: 0;
    padding: 12px;
    background: var(--color-fill-1);
    border-radius: 6px;
  }

  .stat-icon {
    display: inline-flex;
    flex: none;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    font-size: 16px;
    border-radius: 8px;

    &.is-blue {
      color: rgb(var(--arcoblue-6));
      background: rgb(var(--arcoblue-1));
    }

    &.is-green {
      color: rgb(var(--green-6));
      background: rgb(var(--green-1));
    }

    &.is-red {
      color: rgb(var(--red-6));
      background: rgb(var(--red-1));
    }

    &.is-orange {
      color: rgb(var(--orange-6));
      background: rgb(var(--orange-1));
    }
  }

  .stat-main {
    display: flex;
    flex-direction: column;
    min-width: 0;
    line-height: 1.5;

    strong {
      color: var(--color-text-1);
      font-size: 16px;
    }
  }

  .stat-label {
    color: var(--color-text-2);
    font-size: 12px;
  }

  .tag-group {
    display: flex;
    flex-direction: column;
    gap: 8px;

    & + & {
      margin-top: 14px;
    }
  }

  .schema-group,
  .image-group {
    margin-top: 14px;
  }

  .image-list {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }

  .image-figure {
    margin: 0;

    figcaption {
      margin-top: 4px;
      color: var(--color-text-3);
      font-size: 12px;
      text-align: center;
    }
  }

  :deep(.content-image) {
    overflow: hidden;
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
  }

  .schema-list {
    margin: 0;
    padding-left: 20px;
    color: var(--color-text-1);
    font-size: 13px;
    line-height: 1.8;
  }

  .tag-label {
    color: var(--color-text-3);
    font-size: 13px;
  }

  .num {
    font-variant-numeric: tabular-nums;
  }

  .ellipsis {
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .stat-main strong.is-wash {
    color: rgb(var(--red-6));
  }
</style>
