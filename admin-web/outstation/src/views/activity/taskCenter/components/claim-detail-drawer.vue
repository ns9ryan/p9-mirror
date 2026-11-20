<template>
  <a-drawer
    :visible="record !== null"
    :width="480"
    :title="t('discountCenter.taskCenter.claims.detail.title')"
    unmount-on-close
    @cancel="close"
  >
    <template v-if="record">
      <div class="section-title">
        {{ t('discountCenter.taskCenter.claims.detail.base') }}
      </div>
      <dl class="detail-box">
        <div class="detail-row">
          <dt>{{ t('discountCenter.taskCenter.claims.column.orderNo') }}</dt>
          <dd class="mono">{{ record.order_no }}</dd>
        </div>
        <div class="detail-row">
          <dt>{{ t('discountCenter.taskCenter.claims.column.memberId') }}</dt>
          <dd>{{ record.member_id }}</dd>
        </div>
        <div class="detail-row">
          <dt>{{ t('discountCenter.taskCenter.column.name') }}</dt>
          <dd>{{ taskName }}</dd>
        </div>
        <div class="detail-row">
          <dt>{{ t('discountCenter.taskCenter.column.category') }}</dt>
          <dd>
            <a-tag :color="CATEGORY_COLORS[record.category as TaskCategory]">
              {{ t(`discountCenter.taskCenter.category.${record.category}`) }}
            </a-tag>
          </dd>
        </div>
        <div class="detail-row">
          <dt>{{ t('discountCenter.taskCenter.claims.column.periodKey') }}</dt>
          <dd>{{ displayText(record.period_key) }}</dd>
        </div>
        <div class="detail-row">
          <dt>{{ t('discountCenter.taskCenter.claims.column.amount') }}</dt>
          <dd class="amount">{{ formatAmount(record.amount) }}</dd>
        </div>
        <div class="detail-row">
          <dt>{{ t('discountCenter.taskCenter.column.payoutMode') }}</dt>
          <dd>
            {{ t(`discountCenter.taskCenter.payoutMode.${record.payout_mode}`) }}
          </dd>
        </div>
        <div class="detail-row">
          <dt>{{ t('discountCenter.taskCenter.claims.column.status') }}</dt>
          <dd>
            <a-tag :color="REWARD_STATUS_COLORS[record.status]">
              {{ t(`discountCenter.taskCenter.claims.status.${record.status}`) }}
            </a-tag>
          </dd>
        </div>
        <div class="detail-row">
          <dt>{{ t('discountCenter.taskCenter.claims.column.createdAt') }}</dt>
          <dd>{{ formatTime(record.created_at) }}</dd>
        </div>
        <div class="detail-row">
          <dt>{{ t('discountCenter.taskCenter.claims.column.payoutAt') }}</dt>
          <dd>{{ formatTime(record.payout_at) }}</dd>
        </div>
      </dl>

      <div class="section-title">
        {{ t('discountCenter.taskCenter.claims.detail.claim') }}
      </div>
      <dl class="detail-box">
        <div class="detail-row">
          <dt>{{ t('discountCenter.taskCenter.claims.column.claimIp') }}</dt>
          <dd>{{ displayText(record.claim_ip) }}</dd>
        </div>
        <div class="detail-row">
          <dt>{{ t('discountCenter.taskCenter.claims.column.deviceId') }}</dt>
          <dd class="mono">{{ displayText(record.device_id) }}</dd>
        </div>
        <div class="detail-row">
          <dt>{{ t('discountCenter.taskCenter.claims.column.walletId') }}</dt>
          <dd class="mono">{{ displayText(record.wallet_id) }}</dd>
        </div>
        <div class="detail-row">
          <dt>
            {{ t('discountCenter.taskCenter.claims.column.transactionId') }}
          </dt>
          <dd class="mono">
            {{ displayText(record.wallet_transaction_id) }}
          </dd>
        </div>
      </dl>
    </template>
    <template #footer>
      <a-button @click="close">
        {{ t('discountCenter.taskCenter.claims.detail.close') }}
      </a-button>
    </template>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { TaskCategory, TaskRewardItem } from '@/api/task';
  import {
    CATEGORY_COLORS,
    REWARD_STATUS_COLORS,
    displayText,
    formatAmount,
    formatTime,
  } from '../format';

  const props = defineProps<{
    record: TaskRewardItem | null;
    taskName?: string;
  }>();

  const emit = defineEmits<{
    (e: 'close'): void;
  }>();

  const { t } = useI18n();

  const taskName = computed(
    () =>
      props.taskName ||
      props.record?.task_code ||
      '--'
  );

  function close() {
    emit('close');
  }
</script>

<style scoped lang="less">
  .section-title {
    margin-bottom: 12px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 14px;
  }

  .detail-box {
    margin: 0 0 24px;
    padding: 4px 16px;
    background: var(--color-fill-1);
    border-radius: 6px;
  }

  .detail-row {
    display: flex;
    gap: 16px;
    align-items: flex-start;
    padding: 12px 0;
    border-bottom: 1px solid var(--color-border-2);

    &:last-child {
      border-bottom: none;
    }
  }

  dt {
    flex: 0 0 110px;
    color: var(--color-text-3);
    line-height: 22px;
  }

  dd {
    flex: 1;
    margin: 0;
    color: var(--color-text-1);
    line-height: 22px;
    word-break: break-all;
  }

  .mono {
    font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
    font-size: 13px;
  }

  .amount {
    color: rgb(var(--primary-6));
    font-variant-numeric: tabular-nums;
  }
</style>
