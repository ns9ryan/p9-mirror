<template>
  <div class="activity-item" :class="{ 'is-expanded': expanded }">
    <div class="activity-row" @click="toggleExpanded">
      <button
        type="button"
        class="fold"
        :aria-expanded="expanded"
        :aria-label="
          expanded
            ? t('discountCenter.activityCenter.card.collapse')
            : t('discountCenter.activityCenter.card.expand')
        "
        @click.stop="toggleExpanded"
      >
        <icon-right class="fold-caret" :class="{ 'is-open': expanded }" />
      </button>

      <div class="cell name-cell">
        <a-image
          class="thumb"
          :src="record.image || imageSrc"
          :width="40"
          :height="40"
          :preview="false"
          fit="cover"
        />
        <div class="stack">
          <strong class="ellipsis" :title="record.name">
            {{ displayText(record.name) }}
          </strong>
          <div class="meta">
            <a-tag
              size="small"
              class="type-tag"
              :color="TYPE_COLORS[record.type] || 'gray'"
            >
              {{ typeText(record.type, enumParams) }}
            </a-tag>
            <span class="num muted">ID {{ record.id }}</span>
          </div>
        </div>
      </div>

      <div class="cell">
        <span class="status" :class="`is-${statusKey(record.status)}`">
          <i class="status-dot"></i>
          {{ statusText(record.status, enumParams) }}
        </span>
      </div>

      <div class="cell stack num">
        <template v-if="record.is_permanent">
          <span>{{ t('discountCenter.activityCenter.permanent') }}</span>
        </template>
        <template v-else>
          <span class="time-line">
            <small>{{ t('discountCenter.activityCenter.row.start') }}</small>
            {{ displayText(formatTime(record.open_at_sec, MINUTE)) }}
          </span>
          <span class="time-line">
            <small>{{ t('discountCenter.activityCenter.row.end') }}</small>
            {{ displayText(formatTime(record.close_at_sec, MINUTE)) }}
          </span>
        </template>
      </div>

      <div class="cell stack">
        <a-tooltip :content="levelTip" :disabled="!levelTip">
          <span class="ellipsis">{{ levelSummary }}</span>
        </a-tooltip>
        <small class="ellipsis" :title="clientText">{{ clientText }}</small>
      </div>

      <div class="cell stack">
        <span class="kv">
          <small>{{ t('discountCenter.activityCenter.row.wagering') }}</small>
          <em :class="{ 'is-wash': !!record.wagering_multiplier }">
            {{ displayText(wageringText(record, t)) }}
          </em>
        </span>
        <span class="kv">
          <small>{{ t('discountCenter.activityCenter.row.autoAudit') }}</small>
          <em :class="{ muted: !record.auto_audit }">{{ autoAudit }}</em>
        </span>
      </div>

      <div class="cell stack">
        <span :class="record.show_in_list ? 'is-listed' : 'muted'">
          {{
            t(
              `discountCenter.activityCenter.row.${
                record.show_in_list ? 'listed' : 'hidden'
              }`
            )
          }}
        </span>
        <small class="num">
          {{
            t('discountCenter.activityCenter.row.sort', {
              value: record.sort_weight,
            })
          }}
        </small>
      </div>

      <div class="cell stack num">
        <span>{{
          displayText(formatTime(record.created_at_sec, MINUTE))
        }}</span>
        <small class="ellipsis">
          {{ displayText(uidText(record.create_uid)) }}
        </small>
      </div>

      <div class="cell ops" @click.stop>
        <template v-if="actions.length">
          <a-button
            v-for="action in actions"
            :key="action"
            type="text"
            size="small"
            :status="action === 'close' ? 'danger' : undefined"
            @click="emit('action', action)"
          >
            {{ t(`discountCenter.activityCenter.card.${action}`) }}
          </a-button>
        </template>
        <span v-else class="muted">—</span>
      </div>
    </div>

    <div class="activity-panel" :class="{ 'is-open': expanded }">
      <div class="activity-panel__inner">
        <a-spin :loading="detailLoading" class="detail">
          <div v-if="detailFailed" class="detail-failed">
            {{ t('discountCenter.activityCenter.expand.loadFailed') }}
            <a-link @click.stop="loadDetail">
              {{ t('discountCenter.activityCenter.expand.retry') }}
            </a-link>
          </div>

          <template v-else>
            <section class="detail-section">
              <h5 class="section-title">
                {{ t('discountCenter.activityCenter.expand.config') }}
              </h5>
              <dl class="kv-grid">
                <div
                  v-for="item in configItems"
                  :key="item.label"
                  class="kv-item"
                >
                  <dt>{{ item.label }}</dt>
                  <dd :title="item.value">{{ item.value }}</dd>
                </div>
              </dl>
            </section>

            <!-- 三栏并排共用宽度，避免每组独占一行时右侧大片留白 -->
            <div class="detail-columns">
              <section class="detail-section">
                <h5 class="section-title">
                  {{ t('discountCenter.activityCenter.expand.scope') }}
                </h5>
                <div class="tag-rows">
                  <div v-for="row in tagRows" :key="row.label" class="tag-row">
                    <span class="tag-row-label">{{ row.label }}</span>
                    <div class="tag-list">
                      <a-tag
                        v-for="(item, index) in row.items"
                        :key="index"
                        size="small"
                        :color="row.color"
                      >
                        {{ item }}
                      </a-tag>
                    </div>
                  </div>
                </div>
              </section>

              <section class="detail-section">
                <h5 class="section-title">
                  {{ t('discountCenter.activityCenter.expand.rewards') }}
                </h5>
                <div v-if="schemaGroups.length" class="schema-groups">
                  <div
                    v-for="group in schemaGroups"
                    :key="group.label || group.kind"
                  >
                    <span v-if="group.label" class="schema-group-label">
                      {{ group.label }}
                    </span>
                    <ol class="schema-list">
                      <li v-for="(item, index) in group.items" :key="index">
                        <b>{{ index + 1 }}</b>
                        {{ schemaText(item, group.kind, t) }}
                      </li>
                    </ol>
                  </div>
                </div>
                <span v-else class="muted">
                  {{ t('discountCenter.activityCenter.expand.noRewards') }}
                </span>
              </section>

              <section class="detail-section">
                <h5 class="section-title">
                  {{ t('discountCenter.activityCenter.expand.content') }}
                </h5>
                <ActivityLocales
                  v-if="detail?.i18n?.length"
                  :items="detail.i18n"
                />
                <span v-else class="muted">---</span>
              </section>
            </div>

            <div v-if="images.length" class="detail-group">
              <span class="group-title">
                {{ t('discountCenter.activityCenter.expand.images') }}
              </span>
              <a-image-preview-group infinite>
                <div class="image-list" @click.stop>
                  <figure
                    v-for="image in images"
                    :key="image.lang"
                    class="image-figure"
                  >
                    <a-image
                      class="detail-image"
                      :src="image.url"
                      :width="200"
                      :height="75"
                      fit="cover"
                    />
                    <figcaption>{{ langLabel(image.lang) }}</figcaption>
                  </figure>
                </div>
              </a-image-preview-group>
            </div>
          </template>

          <div class="detail-foot num">
            <span v-if="detail">
              {{
                t('discountCenter.activityCenter.expand.version', {
                  value: detail.version,
                })
              }}
              ·
            </span>
            {{
              t('discountCenter.activityCenter.expand.updated', {
                time: displayText(formatTime(record.updated_at_sec, MINUTE)),
                user: displayText(uidText(record.update_uid)),
              })
            }}
          </div>
        </a-spin>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { computed, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import {
    ACTIVITY_TYPE,
    getActivityDetail,
    type ActivityDetail,
  } from '@/api/activity';
  import ActivityLocales from './activity-locales.vue';
  import {
    activityImages,
    autoAuditText,
    canClose,
    canEdit,
    canPublish,
    claimExpireText,
    clientNames,
    clientsText,
    depositRewardText,
    displayText,
    enumText,
    extraRuleItems,
    formatAmount,
    formatTime,
    rewardSchemaGroups,
    schemaText,
    statusKey,
    statusText,
    TYPE_COLORS,
    typeText,
    uidText,
    wageringText,
    type EnumGroup,
  } from '../options';
  import { useActivityStore } from '../store';
  import type { ActivityAction, ActivityRecord } from '../types';

  const props = defineProps<{
    record: ActivityRecord;
    expanded: boolean;
    /** 活动没有图片时的占位图 */
    imageSrc: string;
  }>();

  const emit = defineEmits<{
    (e: 'update:expanded', value: boolean): void;
    (e: 'action', value: ActivityAction): void;
  }>();

  const MINUTE = 'YYYY-MM-DD HH:mm';
  const PREFIX = 'discountCenter.activityCenter';

  const { t } = useI18n();
  const {
    levelNames,
    levelGroups,
    langLabel,
    loadLangs,
    gameName,
    loadGames,
    enumParams,
    loadEnumParams,
  } = useActivityStore();

  const enumValue = (group: EnumGroup, code: number | null | undefined) =>
    enumText(group, code, enumParams.value);

  // 语言名、游戏名和枚举文案只在展开的详情里显示，游戏较多，未限定游戏时不加载
  watch(
    () => props.expanded,
    (expanded) => {
      if (!expanded) return;
      loadLangs();
      loadEnumParams();
      if (props.record.allowed_game_ids?.length) loadGames();
    },
    { immediate: true }
  );

  const autoAudit = computed(() => autoAuditText(props.record, t));

  /** 行内只显示层级数量，名称放在悬浮提示里，避免长列表把单元格撑乱 */
  const levelSummary = computed(() => {
    const count = levelNames(props.record).length;
    return count
      ? t(`${PREFIX}.row.levelCount`, { count })
      : t(`${PREFIX}.row.allLevels`);
  });

  const levelTip = computed(() => {
    const { auto, fixed } = levelGroups(props.record);
    return [
      auto.length && `${t(`${PREFIX}.row.autoLevels`)}：${auto.join('、')}`,
      fixed.length && `${t(`${PREFIX}.row.fixedLevels`)}：${fixed.join('、')}`,
    ]
      .filter(Boolean)
      .join('；');
  });

  const clientText = computed(() =>
    props.record.allowed_clients?.length
      ? clientsText(props.record.allowed_clients, t, enumParams.value)
      : t(`${PREFIX}.row.allClients`)
  );

  const actions = computed(() => {
    const list: ActivityAction[] = [];
    if (canEdit(props.record)) list.push('edit');
    if (canPublish(props.record)) list.push('publish');
    if (canClose(props.record)) list.push('close');
    return list;
  });

  function toggleExpanded() {
    emit('update:expanded', !props.expanded);
  }

  /** 列表接口只有概要，展开时按需拉取详情；各行独立加载，互不影响 */
  const detail = ref<ActivityDetail | null>(null);
  const detailLoading = ref(false);
  const detailFailed = ref(false);
  let ticket = 0;

  async function loadDetail() {
    ticket += 1;
    const current = ticket;
    detailLoading.value = true;
    detailFailed.value = false;
    try {
      const { data } = await getActivityDetail(props.record.id, true);
      if (current !== ticket) return;
      detail.value = data || null;
    } catch {
      if (current !== ticket) return;
      detailFailed.value = true;
    } finally {
      if (current === ticket) detailLoading.value = false;
    }
  }

  // 发布、关闭后列表会刷新，更新时间或状态变化时重新拉取已展开的详情
  watch(
    () => [props.expanded, props.record.status, props.record.updated_at_sec],
    () => {
      if (props.expanded) loadDetail();
    },
    { immediate: true }
  );

  const withDisplay = (items: { label: string; value: string }[]) =>
    items.map((item) => ({ ...item, value: displayText(item.value) }));

  const yesNo = (value?: boolean) =>
    value == null ? '' : t(`${PREFIX}.${value ? 'yes' : 'no'}`);

  const configItems = computed(() => {
    const info = detail.value;
    const { record } = props;
    const settings = info?.settings;
    const payout = [
      {
        label: t(`${PREFIX}.detail.payoutMode`),
        value: enumValue('payoutMode', info?.payout_mode),
      },
      {
        label: t(`${PREFIX}.detail.payoutTiming`),
        value: enumValue('payoutTiming', info?.payout_timing),
      },
      {
        label: t(`${PREFIX}.detail.claimExpire`),
        value: claimExpireText(info?.claim_expire_sec, t),
      },
      {
        label: t(`${PREFIX}.detail.wageringBase`),
        value: enumValue('wageringBase', info?.wagering_base),
      },
      {
        label: t(`${PREFIX}.card.bonusMode`),
        value: enumValue('bonusMode', settings?.bonus_mode),
      },
      ...extraRuleItems(record.type, settings, t, enumParams.value),
    ];
    if (record.type === ACTIVITY_TYPE.DEPOSIT) {
      payout.push({
        label: t(`${PREFIX}.detail.depositRewardType`),
        value: depositRewardText(
          settings?.deposit_reward_type,
          enumParams.value
        ),
      });
    }
    const refer = settings?.refer_config;
    if (record.type === ACTIVITY_TYPE.REFER && refer) {
      payout.push(
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
    return withDisplay([
      {
        label: t(`${PREFIX}.detail.applyMode`),
        value: enumValue('applyMode', info?.apply_mode),
      },
      {
        label: t(`${PREFIX}.detail.section`),
        value: enumValue('section', info?.section ?? record.section),
      },
      {
        label: t(`${PREFIX}.detail.cycleMode`),
        value: enumValue('cycleMode', info?.cycle_mode),
      },
      {
        label: t(`${PREFIX}.card.releaseAt`),
        value: formatTime(record.release_at_sec, MINUTE),
      },
      {
        label: t(`${PREFIX}.detail.showInSidebar`),
        value: yesNo(info?.show_in_sidebar),
      },
      ...payout,
    ]);
  });

  const tagRows = computed(() => {
    const { record } = props;
    const all = t(`${PREFIX}.all`);
    const { auto, fixed } = levelGroups(record);
    const clients = clientNames(record.allowed_clients || [], enumParams.value);
    const rows: { label: string; items: string[]; color?: string }[] = [
      {
        label: t(`${PREFIX}.card.client`),
        items: clients.length ? clients : [all],
      },
    ];
    if (auto.length || fixed.length) {
      if (auto.length) {
        rows.push({
          label: t(`${PREFIX}.row.autoLevels`),
          items: auto,
          color: 'green',
        });
      }
      if (fixed.length) {
        rows.push({
          label: t(`${PREFIX}.row.fixedLevels`),
          items: fixed,
          color: 'arcoblue',
        });
      }
    } else {
      rows.push({
        label: t(`${PREFIX}.card.participantLevel`),
        items: [all],
      });
    }
    const games = (record.allowed_game_ids || []).map(gameName);
    rows.push({
      label: t(`${PREFIX}.card.games`),
      items: games.length ? games : [t(`${PREFIX}.allGames`)],
      color: games.length ? 'purple' : undefined,
    });
    const agents = (detail.value?.allowed_agent_ids || []).map(
      (id) => `ID ${id}`
    );
    if (agents.length) {
      rows.push({
        label: t(`${PREFIX}.detail.agents`),
        items: agents,
        color: 'cyan',
      });
    }
    const applyLimit = (detail.value?.apply_limit || []).map((code) =>
      enumValue('applyLimit', code)
    );
    const claimLimit = (detail.value?.claim_limit || []).map((code) =>
      enumValue('claimLimit', code)
    );
    if (applyLimit.length) {
      rows.push({
        label: t(`${PREFIX}.detail.applyLimit`),
        items: applyLimit,
        color: 'orange',
      });
    }
    if (claimLimit.length) {
      rows.push({
        label: t(`${PREFIX}.detail.claimLimit`),
        items: claimLimit,
        color: 'orange',
      });
    }
    return rows;
  });

  const schemaGroups = computed(() =>
    rewardSchemaGroups(props.record.type, detail.value?.settings, t)
  );

  const images = computed(() => activityImages(detail.value?.i18n));
</script>

<style scoped lang="less">
  .activity-row {
    display: grid;
    grid-template-columns: var(--activity-cols);
    gap: 16px;
    align-items: center;
    min-height: 72px;
    padding: 12px 20px 12px 12px;
    color: var(--color-text-1);
    font-size: 13px;
    cursor: pointer;
    transition: background-color 0.15s ease;
  }

  .activity-row:hover,
  .is-expanded {
    background: var(--color-fill-1);
  }

  .cell {
    min-width: 0;
  }

  .fold {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    padding: 0;
    color: var(--color-text-3);
    background: transparent;
    border: none;
    cursor: pointer;
  }

  .fold-caret {
    transition: transform 0.2s ease;
  }

  .fold-caret.is-open {
    color: rgb(var(--primary-6));
    transform: rotate(90deg);
  }

  .name-cell {
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .name-cell .stack {
    flex: 1;
  }

  :deep(.thumb) {
    flex: none;
    overflow: hidden;
    border-radius: 6px;

    .arco-image-img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
  }

  .stack {
    display: flex;
    flex-direction: column;
    min-width: 0;
    line-height: 1.5;
  }

  .stack > * {
    min-width: 0;
  }

  .stack strong {
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 14px;
  }

  .stack small {
    margin-top: 2px;
    color: var(--color-text-3);
    font-size: 12px;
  }

  .meta {
    display: flex;
    gap: 8px;
    align-items: center;
    margin-top: 4px;
    font-size: 12px;
  }

  .type-tag {
    flex: none;
  }

  .muted {
    color: var(--color-text-3);
  }

  .status {
    display: inline-flex;
    gap: 6px;
    align-items: center;
    white-space: nowrap;
  }

  .status-dot {
    width: 6px;
    height: 6px;
    background: rgb(var(--gray-5));
    border-radius: 50%;
  }

  .status.is-running .status-dot {
    background: rgb(var(--green-6));
    box-shadow: 0 0 0 3px rgb(var(--green-1));
  }

  .status.is-ended {
    color: var(--color-text-3);
  }

  .status.is-ended .status-dot {
    background: rgb(var(--red-5));
  }

  .time-line,
  .kv {
    display: flex;
    gap: 6px;
    align-items: baseline;
    white-space: nowrap;
  }

  .time-line small,
  .kv small {
    flex: none;
    margin-top: 0;
  }

  .kv em {
    overflow: hidden;
    font-style: normal;
    text-overflow: ellipsis;
  }

  .is-listed {
    color: rgb(var(--green-6));
  }

  .is-wash {
    color: rgb(var(--red-6));
    font-weight: 600;
  }

  .ellipsis {
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .num {
    font-variant-numeric: tabular-nums;
  }

  .ops {
    display: flex;
    gap: 4px;
    justify-content: flex-end;
  }

  .ops :deep(.arco-btn) {
    padding: 0 6px;
  }

  .activity-panel {
    display: grid;
    grid-template-rows: 0fr;
    transition: grid-template-rows 0.2s ease;
  }

  .activity-panel.is-open {
    grid-template-rows: 1fr;
  }

  .activity-panel__inner {
    min-height: 0;
    overflow: hidden;
  }

  // 与名称列左对齐（折叠按钮 28px + 行左内边距 12px + 列间距 16px）
  .detail {
    display: block;
    margin: 0 20px 16px 56px;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 8px;
  }

  .detail-section {
    min-width: 0;
    padding: 16px 20px;
  }

  .section-title {
    display: flex;
    gap: 6px;
    align-items: center;
    margin: 0 0 12px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 13px;
    line-height: 20px;

    &::before {
      width: 3px;
      height: 12px;
      background: rgb(var(--primary-6));
      border-radius: 2px;
      content: '';
    }
  }

  .detail-columns {
    display: grid;
    grid-template-columns: minmax(0, 1.2fr) minmax(0, 1fr) minmax(0, 1.3fr);
    border-top: 1px solid var(--color-border-1);

    > .detail-section + .detail-section {
      border-left: 1px dashed var(--color-border-2);
    }
  }

  // 图片等低频内容仍按整行排列
  .detail-group {
    display: grid;
    grid-template-columns: 72px minmax(0, 1fr);
    gap: 16px;
    padding: 14px 20px;
    border-top: 1px solid var(--color-border-1);
  }

  .group-title {
    color: var(--color-text-2);
    font-weight: 500;
    font-size: 13px;
    line-height: 20px;
  }

  .kv-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    gap: 12px 24px;
    margin: 0;
  }

  .kv-item {
    min-width: 0;

    dt {
      margin-bottom: 2px;
      color: var(--color-text-3);
      font-size: 12px;
      line-height: 18px;
    }

    dd {
      margin: 0;
      overflow: hidden;
      color: var(--color-text-1);
      font-size: 13px;
      line-height: 20px;
      white-space: nowrap;
      text-overflow: ellipsis;
    }
  }

  .tag-rows {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .tag-row {
    display: flex;
    gap: 12px;
    align-items: flex-start;
  }

  .tag-row-label {
    flex: none;
    width: 56px;
    color: var(--color-text-3);
    font-size: 12px;
    line-height: 24px;
  }

  .tag-list {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    min-width: 0;
  }

  .schema-groups {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .schema-group-label {
    display: block;
    margin-bottom: 6px;
    color: var(--color-text-3);
    font-size: 12px;
  }

  .schema-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin: 0;
    padding: 0;
    list-style: none;

    li {
      display: flex;
      gap: 8px;
      align-items: center;
      padding: 5px 10px 5px 6px;
      color: var(--color-text-1);
      font-size: 13px;
      background: var(--color-fill-1);
      border-radius: 6px;
    }

    b {
      display: inline-flex;
      flex: none;
      align-items: center;
      justify-content: center;
      width: 20px;
      height: 20px;
      color: rgb(var(--primary-6));
      font-weight: 600;
      font-size: 12px;
      background: rgb(var(--primary-1));
      border-radius: 50%;
    }
  }

  // 窄屏三栏放不下时改为上下堆叠
  @media (max-width: 1440px) {
    .detail-columns {
      grid-template-columns: minmax(0, 1fr);

      > .detail-section + .detail-section {
        border-top: 1px dashed var(--color-border-2);
        border-left: none;
      }
    }
  }

  .image-list {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
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

  :deep(.detail-image) {
    overflow: hidden;
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
  }

  .detail-failed {
    display: flex;
    gap: 8px;
    align-items: center;
    padding: 16px 0;
    color: var(--color-text-3);
  }

  .detail-foot {
    padding: 10px 20px;
    color: var(--color-text-3);
    font-size: 12px;
    text-align: right;
    background: var(--color-fill-1);
    border-top: 1px solid var(--color-border-1);
    border-radius: 0 0 8px 8px;
  }

  .detail-failed {
    padding: 16px 20px;
  }
</style>
