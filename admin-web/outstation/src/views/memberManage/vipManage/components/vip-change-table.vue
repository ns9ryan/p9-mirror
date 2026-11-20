<template>
  <a-table
    row-key="id"
    :columns="columns"
    :data="pageRows"
    :pagination="false"
    :bordered="false"
    :scroll="{ x: 1180 }"
  >
    <template #from="{ record }">
      <VipLevelBadge :level="record.fromLevel" />
    </template>
    <template #to="{ record }">
      <VipLevelBadge :level="record.toLevel" />
    </template>
    <template #type="{ record }">
      <a-tag size="small" bordered :color="changeColor(record)">
        {{ t(`vipManage.change.typeText.${record.type}`) }}
      </a-tag>
    </template>
    <template #operator="{ record }">
      {{ record.operator || t('vipManage.change.system') }}
    </template>
    <template #operations="{ record }">
      <a-link class="detail-link" @click="detail = record">
        {{ t('vipManage.record.detail') }}
      </a-link>
    </template>
    <template #empty>
      <a-empty :description="t('vipManage.change.empty')" />
    </template>
  </a-table>
  <div class="table-pagination">
    <a-pagination
      v-model:current="current"
      v-model:page-size="size"
      :total="records.length"
      :page-size-options="[15, 30, 50, 100]"
      :page-size-props="{ popupContainer: 'body' }"
      show-total
      show-page-size
    />
  </div>

  <VipChangeDetailDrawer :record="detail" @close="detail = null" />
</template>

<script lang="ts" setup>
  import { computed, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import { changeColor } from '../format';
  import type { VipLevelChangeRecord } from '../types';
  import VipChangeDetailDrawer from './vip-change-detail-drawer.vue';
  import VipLevelBadge from './vip-level-badge.vue';

  const props = defineProps<{
    records: VipLevelChangeRecord[];
  }>();

  const { t } = useI18n();
  const current = ref(1);
  const size = ref(15);
  const detail = ref<VipLevelChangeRecord | null>(null);

  const columns = computed<TableColumnData[]>(() => [
    {
      title: t('vipManage.record.account'),
      dataIndex: 'account',
      width: 120,
    },
    {
      title: t('vipManage.change.from'),
      slotName: 'from',
      width: 120,
    },
    {
      title: t('vipManage.change.to'),
      slotName: 'to',
      width: 120,
    },
    {
      title: t('vipManage.change.type'),
      slotName: 'type',
      width: 120,
    },
    {
      title: t('vipManage.change.reason'),
      dataIndex: 'reason',
      width: 260,
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('vipManage.change.operator'),
      slotName: 'operator',
      width: 110,
    },
    {
      title: t('vipManage.change.time'),
      dataIndex: 'changedAt',
      width: 180,
    },
    {
      title: t('vipManage.column.operations'),
      slotName: 'operations',
      width: 100,
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
</script>

<style scoped lang="less">
  :deep(.arco-table-td .arco-table-cell) {
    font-variant-numeric: tabular-nums;
  }

  .detail-link {
    padding: 0;
  }

  .table-pagination {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }

  :deep(.arco-pagination-options .arco-select) {
    min-width: 100px;
  }
</style>
