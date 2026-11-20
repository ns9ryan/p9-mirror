<template>
  <div class="container list-page">
    <a-card class="general-card search-card" :bordered="false">
      <a-form :model="form" layout="inline" class="search-form">
        <div class="search-fields">
          <a-form-item>
            <a-select v-model="form.dateField" class="date-type">
              <a-option value="date">
                {{ t('gameReport.search.date') }}
              </a-option>
            </a-select>
          </a-form-item>
          <a-form-item>
            <a-space :size="12" align="center">
              <a-radio-group
                v-model="quickRange"
                class="quick-range"
                type="button"
                @change="onQuickRange"
                @click="onQuickRangeClick"
              >
                <a-radio value="today">
                  {{ t('gameReport.search.today') }}
                </a-radio>
                <a-radio value="d7">
                  {{ t('gameReport.search.last7') }}
                </a-radio>
                <a-radio value="d30">
                  {{ t('gameReport.search.last30') }}
                </a-radio>
              </a-radio-group>
              <a-range-picker
                v-model="form.dateRange"
                class="range-picker"
                value-format="YYYY-MM-DD"
                allow-clear
                :placeholder="[
                  t('gameReport.search.start'),
                  t('gameReport.search.end'),
                ]"
                @change="onRangeChange"
              />
            </a-space>
          </a-form-item>
          <a-form-item>
            <a-checkbox v-model="form.excludeTest">
              {{ t('gameReport.search.excludeTest') }}
            </a-checkbox>
          </a-form-item>
          <a-form-item class="search-actions">
            <a-space :size="12">
              <a-button type="primary" @click="handleSearch">
                <template #icon><icon-search /></template>
                {{ t('front.common.search.search') }}
              </a-button>
              <a-button @click="handleReset">
                <template #icon><icon-refresh /></template>
                {{ t('front.common.search.reset') }}
              </a-button>
            </a-space>
          </a-form-item>
        </div>
      </a-form>
    </a-card>
    <a-card class="general-card table-card" :bordered="false">
      <a-table
        ref="tableRef"
        row-key="date"
        :columns="columns"
        :data="displayList"
        :pagination="pagination"
        :bordered="true"
        :stripe="true"
        :scroll="{ x: 980 }"
        @page-change="onPageChange"
        @page-size-change="onPageSizeChange"
        @sorter-change="onSorterChange"
      >
        <template #validBet="{ record }">
          <span class="cell-link">{{ formatAmount(record.validBet) }}</span>
        </template>
        <template #killRate="{ record }">
          {{ formatKillRate(record.memberWinLoss, record.validBet) }}
        </template>
        <template #memberWinLoss="{ record }">
          <span :class="winLossClass(record.memberWinLoss)">
            {{ formatAmount(record.memberWinLoss) }}
          </span>
        </template>
        <template #empty>
          <a-empty :description="t('gameReport.empty')" />
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import dayjs from 'dayjs';
  import { useI18n } from 'vue-i18n';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import { formatAmount, formatKillRate } from '../winOrLoss/format';
  import createGameReportSeed from './mock';
  import {
    buildQuickRange,
    detectQuickRange,
    matchGameDate,
    toGameRow,
  } from './query';
  import type { GameSortField, QuickRange, SortDirection } from './types';

  const { t } = useI18n();
  const records = createGameReportSeed();
  const tableRef = ref<{ clearSorters: () => void }>();
  const quickRange = ref<QuickRange | ''>('');
  const form = reactive({
    dateField: 'date',
    dateRange: [] as string[],
    excludeTest: false,
  });
  const appliedRange = ref<string[]>([]);
  const appliedExcludeTest = ref(false);
  const sortField = ref<GameSortField | ''>('');
  const sortDirection = ref<SortDirection>('');
  const pagination = reactive({
    current: 1,
    pageSize: 15,
    total: records.length,
    showTotal: true,
    showPageSize: true,
    pageSizeOptions: [15, 20, 30, 50, 100],
    pageSizeProps: {
      popupContainer: 'body',
    },
  });

  const sortableFields: GameSortField[] = [
    'betCount',
    'bettorCount',
    'validBet',
    'memberWinLoss',
  ];

  const numberSort = {
    sortDirections: ['ascend', 'descend'] as ('ascend' | 'descend')[],
    sorter: true,
  };

  const columns = computed<TableColumnData[]>(() => [
    {
      title: t('gameReport.column.date'),
      dataIndex: 'date',
      width: 150,
      align: 'center',
    },
    {
      title: t('gameReport.column.currency'),
      dataIndex: 'currency',
      width: 120,
      align: 'center',
    },
    {
      title: t('gameReport.column.betCount'),
      dataIndex: 'betCount',
      align: 'center',
      sortable: numberSort,
    },
    {
      title: t('gameReport.column.bettorCount'),
      dataIndex: 'bettorCount',
      align: 'center',
      sortable: numberSort,
    },
    {
      title: t('gameReport.column.validBet'),
      dataIndex: 'validBet',
      slotName: 'validBet',
      align: 'center',
      sortable: numberSort,
    },
    {
      title: t('gameReport.column.killRate'),
      slotName: 'killRate',
      align: 'center',
    },
    {
      title: t('gameReport.column.memberWinLoss'),
      dataIndex: 'memberWinLoss',
      slotName: 'memberWinLoss',
      align: 'center',
      sortable: numberSort,
    },
  ]);

  const displayList = computed(() => {
    const list = records.flatMap((item) => {
      if (!matchGameDate(item, appliedRange.value)) return [];
      const row = toGameRow(item, appliedExcludeTest.value);
      return row ? [row] : [];
    });
    const field = sortField.value;
    const direction = sortDirection.value;
    if (!field || !direction) return list;
    const factor = direction === 'ascend' ? 1 : -1;
    return [...list].sort((a, b) => (a[field] - b[field]) * factor);
  });

  watch(displayList, (list) => {
    pagination.total = list.length;
    const maxPage = Math.max(
      1,
      Math.ceil(list.length / pagination.pageSize) || 1
    );
    if (pagination.current > maxPage) pagination.current = 1;
  });

  function winLossClass(value: number) {
    if (value < 0) return 'cell-link';
    return 'cell-plain';
  }

  function resolveRange() {
    if (form.dateRange[0] && form.dateRange[1]) return [...form.dateRange];
    if (quickRange.value) return buildQuickRange(quickRange.value);
    return [];
  }

  function handleSearch() {
    appliedRange.value = resolveRange();
    appliedExcludeTest.value = form.excludeTest;
    pagination.current = 1;
  }

  function applyQuickRange(key: QuickRange) {
    quickRange.value = key;
    form.dateRange = [...buildQuickRange(key)];
  }

  function handleReset() {
    form.dateField = 'date';
    form.dateRange = [];
    form.excludeTest = false;
    quickRange.value = '';
    appliedRange.value = [];
    appliedExcludeTest.value = false;
    sortField.value = '';
    sortDirection.value = '';
    pagination.current = 1;
    tableRef.value?.clearSorters();
  }

  function onPageChange(current: number) {
    pagination.current = current;
  }

  function onPageSizeChange(pageSize: number) {
    pagination.pageSize = pageSize;
    pagination.current = 1;
  }

  function onQuickRange(value: string | number | boolean) {
    const key = String(value);
    if (key !== 'today' && key !== 'd7' && key !== 'd30') return;
    applyQuickRange(key);
  }

  function onQuickRangeClick(event: MouseEvent) {
    const label = (event.target as HTMLElement | null)?.closest(
      '.arco-radio-button'
    );
    const key = (label?.querySelector('input') as HTMLInputElement | null)
      ?.value;
    if (key !== 'today' && key !== 'd7' && key !== 'd30') return;
    applyQuickRange(key);
  }

  function onRangeChange(
    value: (string | number | Date | undefined)[] | undefined
  ) {
    const range = (value || []).flatMap((item) => {
      if (item == null || item === '') return [];
      if (item instanceof Date) return [dayjs(item).format('YYYY-MM-DD')];
      return [String(item).slice(0, 10)];
    });
    form.dateRange = range.length === 2 ? [range[0], range[1]] : [];
    if (form.dateRange[0] && form.dateRange[1]) {
      quickRange.value = detectQuickRange(form.dateRange);
    }
  }

  function onSorterChange(dataIndex: string, direction: string) {
    const field = sortableFields.find((item) => item === dataIndex);
    if (!field || (direction !== 'ascend' && direction !== 'descend')) {
      sortField.value = '';
      sortDirection.value = '';
      return;
    }
    sortField.value = field;
    sortDirection.value = direction;
  }
</script>

<style scoped lang="less">
  .container {
    padding: 0 20px 20px;
  }

  .general-card {
    margin-top: 16px;
  }

  .general-card :deep(.arco-card-body) {
    padding: 20px;
  }

  .search-card :deep(.arco-card-body) {
    padding: 16px 20px;
  }

  .search-form {
    display: flex;
    flex-wrap: wrap;
    gap: 12px 16px;
    align-items: center;
    width: 100%;
  }

  .search-form :deep(.arco-form-item) {
    margin: 0;
  }

  .search-fields {
    display: flex;
    flex-wrap: wrap;
    gap: 12px 16px;
    align-items: center;
  }

  .date-type {
    width: 88px;

    :deep(.arco-select-view-single) {
      color: #fff;
      background: rgb(var(--primary-6));
      border-color: rgb(var(--primary-6));
    }

    :deep(.arco-select-view-value),
    :deep(.arco-select-view-suffix),
    :deep(.arco-select-view-icon) {
      color: #fff;
    }

    :deep(.arco-select-view-single:hover),
    :deep(.arco-select-view-single.arco-select-view-focus) {
      color: #fff;
      background: rgb(var(--primary-5));
      border-color: rgb(var(--primary-5));
    }
  }

  .range-picker {
    width: 280px;
  }

  .quick-range {
    :deep(.arco-radio-group-button) {
      box-sizing: border-box;
      height: 32px;
      padding: 2px;
      line-height: 26px;
      background: var(--color-fill-2);
    }

    :deep(.arco-radio-button) {
      height: 26px;
      margin: 0;
      line-height: 26px;
    }

    :deep(.arco-radio-button-content) {
      padding: 0 14px;
      line-height: 26px;
    }
  }

  .cell-link {
    color: rgb(var(--primary-6));
  }

  .cell-plain {
    color: var(--color-text-1);
  }

  :deep(.arco-table-th) {
    color: var(--color-text-1);
    font-weight: 600;
    background-color: var(--color-fill-2);
  }

  :deep(.arco-table-td) {
    color: var(--color-text-2);
  }

  :deep(.arco-pagination-options .arco-select) {
    min-width: 100px;
  }
</style>
