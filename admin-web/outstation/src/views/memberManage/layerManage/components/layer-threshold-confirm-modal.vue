<template>
  <a-modal
    :visible="visible"
    :width="560"
    :title="t('layerManage.thresholdConfirm.title')"
    :mask-closable="false"
    unmount-on-close
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    :ok-text="t('layerManage.thresholdConfirm.ok')"
    :cancel-text="t('memberManage.confirm.cancel')"
    @ok="emit('confirm')"
    @cancel="close"
  >
    <div class="confirm-desc">
      {{ t('layerManage.thresholdConfirm.desc') }}
    </div>
    <div class="confirm-name">
      <span class="name-label">{{ t('layerManage.column.name') }}：</span>
      <span class="name-value">{{ name }}</span>
    </div>

    <div class="compare">
      <div class="compare-card">
        <div class="card-head">
          {{ t('layerManage.thresholdConfirm.before') }}
        </div>
        <div v-for="row in rows" :key="row.label" class="card-row">
          <span class="row-label">{{ row.label }}</span>
          <span class="row-value">{{ formatAmount(row.before) }}</span>
        </div>
      </div>
      <icon-arrow-right class="compare-arrow" />
      <div class="compare-card is-after">
        <div class="card-head">
          {{ t('layerManage.thresholdConfirm.after') }}
        </div>
        <div v-for="row in rows" :key="row.label" class="card-row">
          <span class="row-label">{{ row.label }}</span>
          <span
            class="row-value"
            :class="{ 'is-changed': row.before !== row.after }"
          >
            {{ formatAmount(row.after) }}
          </span>
        </div>
      </div>
    </div>

    <div v-if="impact.orderChanged" class="order-change">
      <div class="order-title">
        <icon-swap class="order-icon" />
        {{ t('layerManage.thresholdConfirm.orderTitle') }}
      </div>
      <div class="order-desc">
        {{
          t('layerManage.thresholdConfirm.orderDesc', {
            from: impact.oldIndex + 1,
            to: impact.newIndex + 1,
          })
        }}
      </div>
      <div class="order-chain">
        <template v-if="impact.prev">
          <span class="chain-item">{{ impact.prev.name }}</span>
          <icon-arrow-right class="chain-arrow" />
        </template>
        <span class="chain-item is-current">{{ name }}</span>
        <template v-if="impact.next">
          <icon-arrow-right class="chain-arrow" />
          <span class="chain-item">{{ impact.next.name }}</span>
        </template>
      </div>
    </div>

    <div class="confirm-warning">
      <icon-exclamation-circle-fill class="warning-icon" />
      <div>
        <div class="warning-main">
          {{ t('layerManage.thresholdConfirm.warning') }}
        </div>
        <div class="warning-count">
          {{
            impact.affected > 0
              ? t('layerManage.thresholdConfirm.affected', {
                  count: impact.affected.toLocaleString('en-US'),
                })
              : t('layerManage.thresholdConfirm.affectedNone')
          }}
        </div>
        <div class="warning-sub">
          {{ t('layerManage.thresholdConfirm.async') }}
        </div>
      </div>
    </div>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { formatAmount } from '../../memberList/format';
  import sortAutoLayers from '../sort';
  import type { LayerRecord } from '../types';

  const props = defineProps<{
    visible: boolean;
    record: LayerRecord;
    name: string;
    after: { deposit: number; wash: number };
    layers: LayerRecord[];
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'confirm'): void;
  }>();

  const { t } = useI18n();

  const before = computed(() => ({
    deposit: props.record.deposit ?? 0,
    wash: props.record.wash ?? 0,
  }));

  const rows = computed(() => [
    {
      label: t('layerManage.form.deposit'),
      before: before.value.deposit,
      after: props.after.deposit,
    },
    {
      label: t('layerManage.form.wash'),
      before: before.value.wash,
      after: props.after.wash,
    },
  ]);

  /**
   * 前端只有各层级的会员数，没有会员的累计存款和洗码量，
   * 这里给出可能变动人数的上限，实际人数以后台重算结果为准。
   */
  const impact = computed(() => {
    const { record, after, layers } = props;
    const oldOrder = sortAutoLayers(layers);
    const newOrder = sortAutoLayers(
      layers.map((item) =>
        item.id === record.id
          ? { ...item, deposit: after.deposit, wash: after.wash }
          : item
      )
    );
    const oldIndex = oldOrder.findIndex((item) => item.id === record.id);
    const newIndex = newOrder.findIndex((item) => item.id === record.id);
    const orderChanged = oldIndex !== newIndex;
    const belowAfter = new Set(
      newOrder.slice(0, newIndex).map((item) => item.id)
    );
    const raised =
      after.deposit > before.value.deposit || after.wash > before.value.wash;
    const lowered =
      after.deposit < before.value.deposit || after.wash < before.value.wash;

    let affected = raised || orderChanged ? record.memberCount : 0;
    if (lowered) {
      oldOrder.slice(0, oldIndex).forEach((item, index) => {
        if (!belowAfter.has(item.id)) return;
        const upper = oldOrder[index + 1];
        if (
          after.deposit < (upper.deposit ?? 0) ||
          after.wash < (upper.wash ?? 0)
        ) {
          affected += item.memberCount;
        }
      });
    }
    oldOrder.slice(oldIndex + 1).forEach((item) => {
      if (belowAfter.has(item.id)) affected += item.memberCount;
    });

    return {
      affected,
      orderChanged,
      oldIndex,
      newIndex,
      prev: newOrder[newIndex - 1] ?? null,
      next: newOrder[newIndex + 1] ?? null,
    };
  });

  function close() {
    emit('update:visible', false);
  }
</script>

<style scoped lang="less">
  .confirm-desc {
    margin-bottom: 16px;
    color: var(--color-text-2);
    font-size: 13px;
  }

  .confirm-name {
    padding: 12px 16px;
    background: var(--color-fill-2);
    border-radius: 6px;
  }

  .name-label {
    color: var(--color-text-3);
    font-size: 13px;
  }

  .name-value {
    margin-left: 8px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 15px;
  }

  .compare {
    display: flex;
    gap: 16px;
    align-items: center;
    margin-top: 16px;
  }

  .compare-card {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    border: 1px solid var(--color-border-2);
    border-radius: 6px;

    &.is-after {
      background: rgb(var(--primary-1));
      border-color: rgb(var(--primary-3));

      .card-head {
        color: rgb(var(--primary-6));
        background: rgb(var(--primary-2));
      }
    }
  }

  .card-head {
    padding: 8px 16px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 13px;
    background: var(--color-fill-2);
  }

  .card-row {
    display: flex;
    gap: 12px;
    align-items: center;
    justify-content: space-between;
    padding: 6px 16px;
    font-size: 13px;

    &:last-child {
      padding-bottom: 12px;
    }

    &:nth-child(2) {
      padding-top: 12px;
    }
  }

  .row-label {
    color: var(--color-text-2);
  }

  .row-value {
    color: var(--color-text-1);
    font-weight: 600;

    &.is-changed {
      color: rgb(var(--primary-6));
    }
  }

  .compare-arrow {
    flex-shrink: 0;
    color: rgb(var(--primary-6));
    font-size: 18px;
  }

  .order-change {
    margin-top: 16px;
    padding: 12px 16px;
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
  }

  .order-title {
    display: flex;
    gap: 8px;
    align-items: center;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 13px;
  }

  .order-icon {
    color: rgb(var(--primary-6));
    font-size: 16px;
  }

  .order-desc {
    margin-top: 6px;
    color: var(--color-text-3);
    font-size: 12px;
  }

  .order-chain {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
    margin-top: 10px;
  }

  .chain-item {
    padding: 2px 10px;
    color: var(--color-text-2);
    font-size: 12px;
    line-height: 20px;
    background: var(--color-fill-2);
    border-radius: 4px;

    &.is-current {
      color: rgb(var(--primary-6));
      font-weight: 600;
      background: rgb(var(--primary-1));
      border: 1px solid rgb(var(--primary-3));
    }
  }

  .chain-arrow {
    color: var(--color-text-3);
    font-size: 12px;
  }

  .confirm-warning {
    display: flex;
    gap: 10px;
    align-items: flex-start;
    margin-top: 16px;
    padding: 14px 16px;
    background: rgb(var(--warning-1));
    border: 1px solid rgb(var(--warning-3));
    border-radius: 6px;
  }

  .warning-icon {
    flex-shrink: 0;
    margin-top: 1px;
    color: rgb(var(--warning-6));
    font-size: 18px;
  }

  .warning-main {
    color: rgb(var(--warning-6));
    font-weight: 600;
    font-size: 13px;
    line-height: 22px;
  }

  .warning-count {
    margin-top: 6px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 13px;
  }

  .warning-sub {
    margin-top: 4px;
    color: var(--color-text-3);
    font-size: 12px;
  }
</style>
