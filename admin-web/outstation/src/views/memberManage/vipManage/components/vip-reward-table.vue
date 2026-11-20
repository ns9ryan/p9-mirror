<template>
  <a-table
    row-key="orderNo"
    :loading="loading"
    :columns="columns"
    :data="pageRows"
    :pagination="false"
    :bordered="false"
    :scroll="{ x: 1180 }"
  >
    <template #level="{ record }">
      <a-tag size="small" :color="levelColor(record.level)">
        VIP{{ record.level }}
      </a-tag>
    </template>
    <template #reward="{ record }">{{ formatAmount(record.reward) }}</template>
    <template #mode="{ record }">
      {{ t(`vipManage.record.modeText.${record.mode}`) }}
    </template>
    <template #status="{ record }">
      <a-tag size="small" bordered :color="statusColor[record.status]">
        {{ t(`vipManage.record.statusText.${record.status}`) }}
      </a-tag>
    </template>
    <template #claimedAt="{ record }">{{ record.claimedAt || '-' }}</template>
    <template #washRequired="{ record }">
      {{ formatAmount(record.washRequired) }}
    </template>
    <template #operations="{ record }">
      <a-space :size="16" class="operation-links">
        <a-link @click="detail = record">
          {{ t('vipManage.record.detail') }}
        </a-link>
        <a-link
          v-if="record.status === 'pending'"
          @click="confirmDispatch(record)"
        >
          {{ t('vipManage.record.dispatchNow') }}
        </a-link>
      </a-space>
    </template>
    <template #empty>
      <a-empty :description="t('vipManage.record.empty')" />
    </template>
  </a-table>
  <div class="table-footer">
    <span class="total-text">
      {{ t('vipManage.record.total', { total: records.length, size }) }}
    </span>
    <a-pagination
      v-model:current="current"
      v-model:page-size="size"
      :total="records.length"
      :page-size-options="[10, 20, 50, 100]"
      :page-size-props="{ popupContainer: 'body' }"
      show-page-size
    />
  </div>

  <VipRewardDetailDrawer
    :record="detail"
    @close="detail = null"
    @view-changes="onViewChanges"
  />
</template>

<script lang="ts" setup>
  import { computed, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Modal } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import { formatAmount } from '../../memberList/format';
  import { levelColor, statusColor } from '../format';
  import type { VipRewardRecord } from '../types';
  import VipRewardDetailDrawer from './vip-reward-detail-drawer.vue';

  const props = defineProps<{
    records: VipRewardRecord[];
    loading: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'dispatch', record: VipRewardRecord): void;
    (e: 'viewChanges', account: string): void;
  }>();

  function onViewChanges(account: string) {
    detail.value = null;
    emit('viewChanges', account);
  }

  const { t } = useI18n();
  const current = ref(1);
  const size = ref(10);
  const detail = ref<VipRewardRecord | null>(null);

  const columns = computed<TableColumnData[]>(() => [
    {
      title: t('vipManage.record.account'),
      dataIndex: 'account',
      width: 120,
    },
    {
      title: t('vipManage.record.nickname'),
      dataIndex: 'nickname',
      width: 120,
    },
    {
      title: t('vipManage.record.level'),
      slotName: 'level',
      width: 100,
    },
    {
      title: t('vipManage.record.reward'),
      slotName: 'reward',
      width: 110,
      align: 'right',
    },
    {
      title: t('vipManage.record.mode'),
      slotName: 'mode',
      width: 110,
    },
    {
      title: t('vipManage.record.status'),
      slotName: 'status',
      width: 100,
    },
    {
      title: t('vipManage.record.time'),
      slotName: 'claimedAt',
      width: 180,
    },
    {
      title: t('vipManage.record.washRequired'),
      slotName: 'washRequired',
      width: 130,
      align: 'right',
    },
    {
      title: t('vipManage.column.operations'),
      slotName: 'operations',
      width: 170,
    },
  ]);

  const pageRows = computed(() =>
    props.records.slice(
      (current.value - 1) * size.value,
      current.value * size.value
    )
  );

  watch(
    () => props.records.length,
    (total) => {
      const maxPage = Math.max(1, Math.ceil(total / size.value));
      if (current.value > maxPage) current.value = maxPage;
    }
  );

  defineExpose({
    resetPage() {
      current.value = 1;
    },
  });

  function confirmDispatch(record: VipRewardRecord) {
    Modal.confirm({
      title: t('vipManage.record.dispatchTitle'),
      content: t('vipManage.record.dispatchContent', {
        account: record.account,
        amount: formatAmount(record.reward),
      }),
      okText: t('vipManage.record.dispatchNow'),
      cancelText: t('memberManage.confirm.cancel'),
      onOk() {
        emit('dispatch', record);
      },
    });
  }
</script>

<style scoped lang="less">
  :deep(.arco-table-td .arco-table-cell) {
    font-variant-numeric: tabular-nums;
  }

  .operation-links :deep(.arco-link) {
    padding: 0;
  }

  .table-footer {
    display: flex;
    gap: 12px;
    align-items: center;
    justify-content: space-between;
    margin-top: 16px;
  }

  .total-text {
    color: var(--color-text-2);
  }

  :deep(.arco-pagination-options .arco-select) {
    min-width: 100px;
  }
</style>
