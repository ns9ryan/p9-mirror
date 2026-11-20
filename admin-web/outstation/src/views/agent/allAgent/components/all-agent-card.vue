<template>
  <div class="agent-item" :class="{ 'is-selected': selected }">
    <div class="agent-row" @click="toggleExpanded">
      <span class="check-cell" @click.stop>
        <a-checkbox :model-value="selected" @change="onCheck" />
      </span>
      <button
        type="button"
        class="fold"
        :aria-expanded="expanded"
        :aria-label="
          expanded ? t('allAgent.card.collapse') : t('allAgent.card.expand')
        "
        @click.stop="toggleExpanded"
      >
        <AgentFoldIcon :open="expanded" />
      </button>
      <span class="agent-cell">
        <strong>{{ record.id }}</strong>
        <small>
          {{ t('allAgent.column.parentId') }}：{{
            displayText(record.parentId)
          }}
        </small>
      </span>
      <span :class="['num', { 'count-link': record.subCount > 0 }]">
        {{ record.subCount }}
      </span>
      <span class="num">{{ record.subTotal }}</span>
      <span class="num is-balance">{{ formatAmount(record.commission) }}</span>
      <span class="num">{{ formatAmount(record.unclaimedCommission) }}</span>
      <span :class="['withdraw', record.withdrawLimit]">
        {{ t(`allAgent.withdraw.${record.withdrawLimit}`) }}
      </span>
      <span :class="['state', record.status]">
        {{ t(`allAgent.status.${record.status}`) }}
      </span>
      <span class="num">{{ displayText(record.registeredAt) }}</span>
      <span class="ops" @click.stop>
        <a-dropdown trigger="click" @select="onBet">
          <a-button type="text" size="small">
            {{ t('allAgent.card.betLimit') }}
            <icon-down />
          </a-button>
          <template #content>
            <a-doption
              v-for="action in BET_ACTIONS"
              :key="action"
              :value="action"
            >
              {{ t(`allAgent.action.${action}`) }}
            </a-doption>
          </template>
        </a-dropdown>
        <a-button type="text" size="small" @click="emit('remark')">
          {{ t('allAgent.action.remark') }}
        </a-button>
      </span>
    </div>

    <div class="agent-panel" :class="{ 'is-open': expanded }">
      <div class="agent-panel__inner">
        <div class="detail-grid">
          <section>
            <h4><i class="dot is-green"></i>{{ t('allAgent.card.team') }}</h4>
            <p>
              <span>{{ t('allAgent.column.memberId') }}</span>
              <strong>{{ record.id }}</strong>
            </p>
            <p>
              <span>{{ t('allAgent.column.parentId') }}</span>
              {{ displayText(record.parentId) }}
            </p>
            <p>
              <span>{{ t('allAgent.column.subCount') }}</span>
              <em :class="{ 'count-link': record.subCount > 0 }">
                {{ record.subCount }}
              </em>
            </p>
            <p>
              <span>{{ t('allAgent.column.subTotal') }}</span>
              {{ record.subTotal }}
            </p>
            <p>
              <span>{{ t('allAgent.search.registerSource') }}</span>
              {{ displayText(record.registerSource) }}
            </p>
          </section>

          <section>
            <h4><i class="dot is-orange"></i>{{ t('allAgent.card.bet') }}</h4>
            <p>
              <span>{{ t('allAgent.column.betLevel1') }}</span>
              <em class="num">{{ formatAmount(record.betLevel1) }}</em>
            </p>
            <p>
              <span>{{ t('allAgent.column.betLevel2') }}</span>
              <em class="num">{{ formatAmount(record.betLevel2) }}</em>
            </p>
            <p>
              <span>{{ t('allAgent.column.betLevel3') }}</span>
              <em class="num">{{ formatAmount(record.betLevel3) }}</em>
            </p>
          </section>

          <section>
            <h4>
              <i class="dot is-blue"></i>{{ t('allAgent.card.commission') }}
            </h4>
            <div class="fund-list">
              <p v-for="item in fundItems" :key="item.label">
                <span>{{ item.label }}</span>
                <em class="num" :class="`is-${item.tone}`">{{ item.value }}</em>
              </p>
            </div>
          </section>

          <section>
            <h4><i class="dot is-lime"></i>{{ t('allAgent.card.promo') }}</h4>
            <p class="link-line" @click.stop>
              <span>{{ t('allAgent.column.link') }}</span>
              <template v-if="record.promoCode">
                <a
                  class="promo-link"
                  :href="promoLink"
                  :title="promoLink"
                  target="_blank"
                >
                  {{ promoLink }}
                </a>
                <a-button type="text" size="mini" @click="copyLink">
                  {{ t('allAgent.copy') }}
                </a-button>
              </template>
              <template v-else>{{ t('allAgent.empty') }}</template>
            </p>
            <p>
              <span>{{ t('allAgent.column.uv') }}</span>
              {{ record.visitUv }}
            </p>
            <p>
              <span>{{ t('allAgent.column.lastLogin') }}</span>
              {{ displayText(record.lastLoginAt) }}
            </p>
            <p class="ellipsis-line" :title="record.remark">
              <span>{{ t('allAgent.column.remark') }}</span>
              <b>{{ displayText(record.remark) }}</b>
            </p>
          </section>

          <section>
            <h4>
              <i class="dot is-purple"></i>{{ t('allAgent.column.status') }}
            </h4>
            <p>
              <span>{{ t('allAgent.column.status') }}</span>
              <em :class="['state', record.status]">
                {{ t(`allAgent.status.${record.status}`) }}
              </em>
            </p>
            <p>
              <span>{{ t('allAgent.column.withdrawLimit') }}</span>
              <em :class="['withdraw', record.withdrawLimit]">
                {{ t(`allAgent.withdraw.${record.withdrawLimit}`) }}
              </em>
            </p>
            <p class="tag-line">
              <span>{{ t('allAgent.card.betLimit') }}</span>
              <a-space wrap :size="4">
                <a-tag v-if="record.selfBetBanned" size="small" color="red">
                  {{ t('allAgent.card.selfBanned') }}
                </a-tag>
                <a-tag v-if="record.lineBetBanned" size="small" color="red">
                  {{ t('allAgent.card.lineBanned') }}
                </a-tag>
                <a-tag
                  v-if="!record.selfBetBanned && !record.lineBetBanned"
                  size="small"
                  color="green"
                >
                  {{ t('allAgent.card.betNormal') }}
                </a-tag>
              </a-space>
            </p>
            <p>
              <span>{{ t('allAgent.column.registeredAt') }}</span>
              {{ displayText(record.registeredAt) }}
            </p>
          </section>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import AgentFoldIcon from '@/views/agent/list/components/agent-fold-icon.vue';
  import { displayText, formatAmount } from '../format';
  import { buildPromoLink } from '../options';
  import type { AllAgentRecord, BetAction } from '../types';

  const BET_ACTIONS: BetAction[] = [
    'limitSelf',
    'limitLine',
    'liftSelf',
    'liftLine',
  ];

  const props = defineProps<{
    record: AllAgentRecord;
    selected: boolean;
    expanded: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'update:selected', value: boolean): void;
    (e: 'update:expanded', value: boolean): void;
    (e: 'bet', action: BetAction): void;
    (e: 'remark'): void;
  }>();

  const { t } = useI18n();

  type FundTone = 'balance' | 'up' | 'down' | 'zero';

  function positiveTone(value: number): FundTone {
    return value > 0 ? 'up' : 'zero';
  }

  const fundItems = computed(() => {
    const { record } = props;
    return [
      {
        label: t('allAgent.column.commission'),
        value: formatAmount(record.commission),
        tone: 'balance' as FundTone,
      },
      {
        label: t('allAgent.column.claimed'),
        value: formatAmount(record.claimedCommission),
        tone: positiveTone(record.claimedCommission),
      },
      {
        label: t('allAgent.column.pending'),
        value: formatAmount(record.pendingCommission),
        tone: positiveTone(record.pendingCommission),
      },
      {
        label: t('allAgent.column.unclaimed'),
        value: formatAmount(record.unclaimedCommission),
        tone: positiveTone(record.unclaimedCommission),
      },
      {
        label: t('allAgent.column.frozen'),
        value: formatAmount(record.frozenCommission),
        tone: (record.frozenCommission > 0 ? 'down' : 'zero') as FundTone,
      },
      {
        label: t('allAgent.column.cancelled'),
        value: formatAmount(record.cancelledCommission),
        tone: (record.cancelledCommission > 0 ? 'down' : 'zero') as FundTone,
      },
    ];
  });

  const promoLink = computed(() =>
    buildPromoLink(props.record.promoHost, props.record.promoCode)
  );

  function onCheck(value: boolean | (string | number | boolean)[]) {
    emit('update:selected', value === true);
  }

  function toggleExpanded() {
    emit('update:expanded', !props.expanded);
  }

  function onBet(value: string | number | Record<string, unknown> | undefined) {
    const action = BET_ACTIONS.find((item) => item === value);
    if (action) emit('bet', action);
  }

  async function copyLink() {
    try {
      await navigator.clipboard.writeText(promoLink.value);
      Message.success(t('allAgent.message.copied'));
    } catch {
      Message.warning(t('allAgent.message.copyFail'));
    }
  }
</script>

<style scoped lang="less">
  .agent-item.is-selected {
    background: rgb(var(--primary-1));
  }

  .agent-row {
    display: grid;
    grid-template-columns: var(--agent-cols);
    gap: 8px;
    align-items: center;
    min-height: 64px;
    padding: 0 16px;
    color: var(--color-text-1);
    font-size: 13px;
    cursor: pointer;
  }

  .agent-row > span {
    min-width: 0;
  }

  .check-cell {
    display: inline-flex;
    align-items: center;
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

  .agent-cell strong,
  .agent-cell small {
    display: block;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .agent-cell strong {
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 14px;
  }

  .agent-cell small {
    margin-top: 2px;
    color: var(--color-text-3);
    font-size: 12px;
  }

  .num {
    font-variant-numeric: tabular-nums;
  }

  .count-link {
    color: rgb(var(--primary-6));
    font-weight: 600;
  }

  .state,
  .withdraw {
    font-style: normal;
  }

  .state.normal {
    color: rgb(var(--green-6));
  }

  .state.unset {
    color: var(--color-text-3);
  }

  .withdraw.limited {
    color: rgb(var(--red-6));
  }

  .ops {
    display: flex;
    gap: 4px;
  }

  .ops :deep(.arco-btn) {
    padding: 0 6px;
  }

  .agent-panel {
    display: grid;
    grid-template-rows: 0fr;
    transition: grid-template-rows 0.2s ease;
  }

  .agent-panel.is-open {
    grid-template-rows: 1fr;
  }

  .agent-panel__inner {
    min-height: 0;
    overflow: hidden;
  }

  .detail-grid {
    display: grid;
    grid-template-columns:
      minmax(0, 1fr) minmax(0, 0.9fr) minmax(0, 1.6fr) minmax(0, 1.3fr)
      minmax(0, 1fr);
    margin: 0 16px 16px;
    padding: 4px 0;
    background: var(--color-fill-2);
    border-radius: 8px;
  }

  section {
    min-width: 0;
    padding: 14px 18px;
  }

  section + section {
    border-left: 1px solid var(--color-border-2);
  }

  h4 {
    display: flex;
    gap: 8px;
    align-items: center;
    margin: 0 0 10px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 14px;
  }

  p {
    display: flex;
    gap: 8px;
    align-items: baseline;
    margin: 0 0 8px;
    color: var(--color-text-1);
    font-size: 13px;
    line-height: 1.6;
  }

  p > span {
    flex: none;
    color: var(--color-text-3);
  }

  p strong {
    font-weight: 600;
  }

  p em {
    font-style: normal;
  }

  .ellipsis-line b {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    font-weight: normal;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .tag-line {
    flex-wrap: wrap;
    align-items: center;
  }

  .link-line {
    align-items: center;
  }

  .link-line :deep(.arco-btn) {
    flex: none;
    padding: 0 4px;
  }

  .promo-link {
    min-width: 0;
    overflow: hidden;
    color: rgb(var(--primary-6));
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .fund-list {
    display: grid;
    grid-template-columns: 1fr 1fr;
    column-gap: 16px;
  }

  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }

  .is-green {
    background: #14c9c9;
  }

  .is-orange {
    background: #f59a23;
  }

  .is-blue {
    background: #3b6cff;
  }

  .is-lime {
    background: #00b42a;
  }

  .is-purple {
    background: #7c6bf2;
  }

  .is-balance {
    color: rgb(var(--orange-6));
    font-weight: 600;
  }

  .is-up {
    color: rgb(var(--green-6));
    font-weight: 600;
  }

  .is-down {
    color: rgb(var(--red-6));
    font-weight: 600;
  }

  @media (max-width: 1400px) {
    .detail-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    section + section {
      border-left: none;
    }
  }
</style>
