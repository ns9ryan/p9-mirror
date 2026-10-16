<template>
  <div class="agent-list-row">
    <div class="agent-list-header" @click="toggleExpanded">
      <button
        type="button"
        class="expand-btn"
        :aria-expanded="expanded"
        :aria-label="
          expanded ? t('allAgent.card.collapse') : t('allAgent.card.expand')
        "
        @click.stop="toggleExpanded"
      >
        <icon-down class="expand-icon" :class="{ 'is-open': expanded }" />
      </button>
      <span>{{ t('allAgent.column.memberId') }}：{{ record.id }}</span>
      <span>
        {{ t('allAgent.column.parentId') }}：{{ displayText(record.parentId) }}
      </span>
      <span>
        {{ t('allAgent.column.subCount') }}：
        <em :class="{ 'count-link': record.subCount > 0 }">
          {{ record.subCount }}
        </em>
      </span>
      <span>
        {{ t('allAgent.column.status') }}：
        <a-tag size="small" :color="statusColor">
          {{ t(`allAgent.status.${record.status}`) }}
        </a-tag>
      </span>
      <span>
        {{ t('allAgent.column.registeredAt') }}：{{
          displayText(record.registeredAt)
        }}
      </span>
      <a-dropdown v-if="!expanded" trigger="click" @select="onOperate">
        <a-button size="small" class="header-actions" @click.stop>
          {{ t('allAgent.column.operate') }}
        </a-button>
        <template #content>
          <a-doption value="limitSelf">
            {{ t('allAgent.action.limitSelf') }}
          </a-doption>
          <a-doption value="limitLine">
            {{ t('allAgent.action.limitLine') }}
          </a-doption>
          <a-doption value="liftSelf">
            {{ t('allAgent.action.liftSelf') }}
          </a-doption>
          <a-doption value="liftLine">
            {{ t('allAgent.action.liftLine') }}
          </a-doption>
          <a-doption value="remark">
            {{ t('allAgent.action.remark') }}
          </a-doption>
        </template>
      </a-dropdown>
    </div>

    <div class="agent-list-panel" :class="{ 'is-open': expanded }">
      <div class="agent-list-panel__inner">
        <div class="agent-list-body">
          <div class="info-block">
            <div class="block-title">
              <span class="block-mark"></span>{{ t('allAgent.card.team') }}
            </div>
            <div class="block-content compact-list">
              <div class="primary-name">{{ record.id }}</div>
              <div>
                <strong>{{ t('allAgent.column.parentId') }}：</strong>
                {{ displayText(record.parentId) }}
              </div>
              <div>
                <strong>{{ t('allAgent.column.subCount') }}：</strong>
                <span :class="{ 'count-link': record.subCount > 0 }">
                  {{ record.subCount }}
                </span>
              </div>
              <div>
                <strong>{{ t('allAgent.column.subTotal') }}：</strong>
                {{ record.subTotal }}
              </div>
              <div>
                <strong>{{ t('allAgent.search.registerSource') }}：</strong>
                {{ displayText(record.registerSource) }}
              </div>
            </div>
          </div>

          <div class="info-block">
            <div class="block-title">
              <span class="block-mark orange"></span>
              {{ t('allAgent.card.bet') }}
            </div>
            <div class="block-content compact-list">
              <div>
                <strong>{{ t('allAgent.column.betLevel1') }}：</strong>
                <span class="num">{{ formatAmount(record.betLevel1) }}</span>
              </div>
              <div>
                <strong>{{ t('allAgent.column.betLevel2') }}：</strong>
                <span class="num">{{ formatAmount(record.betLevel2) }}</span>
              </div>
              <div>
                <strong>{{ t('allAgent.column.betLevel3') }}：</strong>
                <span class="num">{{ formatAmount(record.betLevel3) }}</span>
              </div>
            </div>
          </div>

          <div class="info-block">
            <div class="block-title">
              <span class="block-mark blue"></span>
              {{ t('allAgent.card.commission') }}
            </div>
            <div class="block-content compact-list">
              <div>
                <strong>{{ t('allAgent.column.claimed') }}：</strong>
                <span class="num">
                  {{ formatAmount(record.claimedCommission) }}
                </span>
              </div>
              <div>
                <strong>{{ t('allAgent.column.pending') }}：</strong>
                <span class="num">
                  {{ formatAmount(record.pendingCommission) }}
                </span>
              </div>
              <div>
                <strong>{{ t('allAgent.column.commission') }}：</strong>
                <span class="num">{{ formatAmount(record.commission) }}</span>
              </div>
              <div>
                <strong>{{ t('allAgent.column.frozen') }}：</strong>
                <span class="num">
                  {{ formatAmount(record.frozenCommission) }}
                </span>
              </div>
              <div>
                <strong>{{ t('allAgent.column.cancelled') }}：</strong>
                <span class="num">
                  {{ formatAmount(record.cancelledCommission) }}
                </span>
              </div>
              <div>
                <strong>{{ t('allAgent.column.unclaimed') }}：</strong>
                <span class="num">
                  {{ formatAmount(record.unclaimedCommission) }}
                </span>
              </div>
            </div>
          </div>

          <div class="info-block">
            <div class="block-title">
              <span class="block-mark green"></span>
              {{ t('allAgent.card.promo') }}
            </div>
            <div class="block-content compact-list">
              <div>
                <strong>{{ t('allAgent.column.withdrawLimit') }}：</strong>
                {{ t(`allAgent.withdraw.${record.withdrawLimit}`) }}
              </div>
              <div class="remark-line" :title="record.remark">
                <strong>{{ t('allAgent.column.remark') }}：</strong>
                {{ displayText(record.remark) }}
              </div>
              <div class="link-line" @click.stop>
                <strong>{{ t('allAgent.column.link') }}：</strong>
                <template v-if="record.promoCode">
                  <a class="promo-link" :href="promoLink" target="_blank">
                    {{ promoLink }}
                  </a>
                  <a-button type="text" size="mini" @click="copyLink">
                    {{ t('allAgent.copy') }}
                  </a-button>
                </template>
                <span v-else>{{ t('allAgent.empty') }}</span>
              </div>
              <div>
                <strong>{{ t('allAgent.column.uv') }}：</strong>
                {{ record.visitUv }}
              </div>
              <div>
                <strong>{{ t('allAgent.column.lastLogin') }}：</strong>
                {{ displayText(record.lastLoginAt) }}
              </div>
            </div>
          </div>

          <div class="info-block status-block">
            <div class="block-title">
              <span class="block-mark purple"></span>
              {{ t('allAgent.column.status') }}
            </div>
            <div class="block-content status-wrap">
              <a-tag size="medium" :color="statusColor">
                {{ t(`allAgent.status.${record.status}`) }}
              </a-tag>
            </div>
          </div>
        </div>

        <div class="agent-list-footer">
          <span>
            {{ t('allAgent.column.registeredAt') }}：{{
              displayText(record.registeredAt)
            }}
          </span>
          <span>
            {{ t('allAgent.column.lastLogin') }}：{{
              displayText(record.lastLoginAt)
            }}
          </span>
          <span>{{ t('allAgent.column.uv') }}：{{ record.visitUv }}</span>
          <div class="row-actions" @click.stop>
            <a-link @click="emit('bet', 'limitSelf')">
              {{ t('allAgent.action.limitSelf') }}
            </a-link>
            <a-link @click="emit('bet', 'limitLine')">
              {{ t('allAgent.action.limitLine') }}
            </a-link>
            <a-link @click="emit('bet', 'liftSelf')">
              {{ t('allAgent.action.liftSelf') }}
            </a-link>
            <a-link @click="emit('bet', 'liftLine')">
              {{ t('allAgent.action.liftLine') }}
            </a-link>
            <a-link @click="emit('remark')">
              {{ t('allAgent.action.remark') }}
            </a-link>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { displayText, formatAmount } from '../format';
  import { buildPromoLink } from '../options';
  import type { AllAgentRecord } from '../types';

  type BetAction = 'limitSelf' | 'limitLine' | 'liftSelf' | 'liftLine';

  const props = defineProps<{
    record: AllAgentRecord;
    expanded: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'update:expanded', value: boolean): void;
    (e: 'bet', action: BetAction): void;
    (e: 'remark'): void;
  }>();

  const { t } = useI18n();

  const statusColor = computed(() =>
    props.record.status === 'normal' ? 'green' : 'gray'
  );
  const promoLink = computed(() =>
    buildPromoLink(props.record.promoHost, props.record.promoCode)
  );

  function toggleExpanded() {
    emit('update:expanded', !props.expanded);
  }

  function onOperate(
    value: string | number | Record<string, unknown> | undefined
  ) {
    if (value === 'remark') {
      emit('remark');
      return;
    }
    if (
      value === 'limitSelf' ||
      value === 'limitLine' ||
      value === 'liftSelf' ||
      value === 'liftLine'
    ) {
      emit('bet', value);
    }
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
  .agent-list-row {
    overflow: hidden;
    background-color: var(--color-bg-2);
  }

  .agent-list-header {
    display: flex;
    flex-wrap: wrap;
    gap: 16px;
    align-items: center;
    min-height: 38px;
    padding: 8px 16px;
    color: var(--color-text-2);
    font-size: 13px;
    line-height: 1.6;
    background-color: var(--color-fill-1);
    border-bottom: 1px solid var(--color-border-2);
    cursor: pointer;
  }

  .agent-list-header span {
    display: inline-flex;
    gap: 6px;
    align-items: center;
    white-space: nowrap;
  }

  .expand-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    padding: 0;
    color: var(--color-text-2);
    background: transparent;
    border: none;
    border-radius: 4px;
    cursor: pointer;
  }

  .expand-btn:hover {
    background: var(--color-fill-3);
  }

  .expand-icon {
    transform: rotate(-90deg);
    transition: transform 0.2s ease;
  }

  .expand-icon.is-open {
    transform: rotate(0deg);
  }

  .header-actions {
    margin-left: auto;
  }

  .count-link {
    color: rgb(var(--primary-6));
    font-style: normal;
  }

  .agent-list-panel {
    display: grid;
    grid-template-rows: 0fr;
    transition: grid-template-rows 0.2s ease;
  }

  .agent-list-panel.is-open {
    grid-template-rows: 1fr;
  }

  .agent-list-panel__inner {
    min-height: 0;
    overflow: hidden;
  }

  .agent-list-body {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    background-color: var(--color-bg-2);
  }

  .info-block {
    min-height: 168px;
    padding: 14px 16px;
    background-color: var(--color-bg-2);
    border-right: 1px solid var(--color-border-2);
    border-bottom: 1px solid var(--color-border-2);
  }

  .block-title {
    display: flex;
    gap: 8px;
    align-items: center;
    margin-bottom: 10px;
    color: var(--color-text-2);
    font-weight: 600;
    font-size: 14px;
  }

  .block-mark {
    display: inline-block;
    width: 10px;
    height: 10px;
    background: linear-gradient(135deg, #34d399, #10b981);
    border-radius: 3px;
  }

  .block-mark.orange {
    background: linear-gradient(135deg, #fbbf24, #f59e0b);
  }

  .block-mark.blue {
    background: linear-gradient(135deg, #60a5fa, #2563eb);
  }

  .block-mark.green {
    background: linear-gradient(135deg, #4ade80, #16a34a);
  }

  .block-mark.purple {
    background: linear-gradient(135deg, #c084fc, #8b5cf6);
  }

  .block-content {
    display: flex;
    flex-direction: column;
    gap: 4px;
    color: var(--color-text-2);
    line-height: 1.6;
  }

  .primary-name {
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 16px;
  }

  .compact-list {
    color: var(--color-text-2);
    font-size: 14px;
  }

  .compact-list strong {
    color: var(--color-text-1);
    font-weight: 600;
  }

  .remark-line {
    word-break: break-all;
  }

  .link-line {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    align-items: center;
  }

  .promo-link {
    max-width: 220px;
    overflow: hidden;
    color: rgb(var(--primary-6));
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .num {
    font-variant-numeric: tabular-nums;
  }

  .status-block {
    display: flex;
    flex-direction: column;
  }

  .status-wrap {
    align-items: flex-start;
  }

  .agent-list-footer {
    display: flex;
    flex-wrap: wrap;
    gap: 12px 20px;
    align-items: center;
    min-height: 46px;
    padding: 8px 16px;
    color: var(--color-text-2);
    font-size: 13px;
    border-top: 1px solid var(--color-border-2);
  }

  .row-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 4px 12px;
    margin-left: auto;
  }
</style>
