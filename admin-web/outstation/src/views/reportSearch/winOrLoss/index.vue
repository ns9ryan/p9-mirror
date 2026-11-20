<template>
  <div class="container list-page">
    <a-card class="general-card search-card" :bordered="false">
      <a-form :model="form" layout="inline" class="search-form">
        <div class="search-fields">
          <a-form-item :label="t('winOrLoss.search.date')">
            <a-space :size="12" align="center">
              <a-radio-group
                v-model="quickRange"
                class="quick-range"
                type="button"
                @change="onQuickRange"
                @click="onQuickRangeClick"
              >
                <a-radio value="today">
                  {{ t('winOrLoss.search.today') }}
                </a-radio>
                <a-radio value="d7">{{ t('winOrLoss.search.last7') }}</a-radio>
                <a-radio value="d30">
                  {{ t('winOrLoss.search.last30') }}
                </a-radio>
              </a-radio-group>
              <a-range-picker
                v-model="form.dateRange"
                class="range-picker"
                value-format="YYYY-MM-DD"
                allow-clear
                :placeholder="[
                  t('winOrLoss.search.start'),
                  t('winOrLoss.search.end'),
                ]"
                @change="onRangeChange"
              />
            </a-space>
          </a-form-item>
          <a-form-item>
            <a-input-group class="id-search">
              <a-select v-model="form.idType" class="id-type">
                <a-option value="memberId">
                  {{ t('winOrLoss.search.memberId') }}
                </a-option>
                <a-option value="agentId">
                  {{ t('winOrLoss.search.agentId') }}
                </a-option>
              </a-select>
              <a-input
                v-model="form.keyword"
                class="search-input"
                allow-clear
                :placeholder="t('winOrLoss.search.keywordPlaceholder')"
                @press-enter="handleSearch"
              />
            </a-input-group>
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
      <div class="table-toolbar">
        <a-space>
          <a-button type="primary" @click="downloadReport">
            <template #icon><icon-download /></template>
            {{ t('winOrLoss.toolbar.download') }}
          </a-button>
        </a-space>
      </div>
      <a-table
        ref="tableRef"
        row-key="date"
        :columns="columns"
        :data="displayList"
        :pagination="pagination"
        :bordered="true"
        :stripe="true"
        :scroll="{ x: 1100 }"
        :summary="summary"
        @page-change="onPageChange"
        @page-size-change="onPageSizeChange"
        @sorter-change="onSorterChange"
      >
        <template #validBet="{ record }">
          {{ formatAmount(record.validBet) }}
        </template>
        <template #killRate="{ record }">
          {{ formatKillRate(record.memberWinLoss, record.validBet) }}
        </template>
        <template #memberWinLoss="{ record }">
          {{ formatAmount(record.memberWinLoss) }}
        </template>
        <template #costFee="{ record }">
          {{ formatAmount(record.costFee) }}
        </template>
        <template #summary-cell="{ record, column }">
          <template v-if="column.dataIndex === 'date'">
            {{ record.date }}
          </template>
          <template v-else-if="column.dataIndex === 'betCount'">
            {{ formatCount(record.betCount) }}
          </template>
          <template v-else-if="column.dataIndex === 'validBet'">
            {{ formatAmount(record.validBet) }}
          </template>
          <template v-else-if="column.dataIndex === 'memberWinLoss'">
            {{ formatAmount(record.memberWinLoss) }}
          </template>
          <template v-else-if="column.dataIndex === 'costFee'">
            {{ formatAmount(record.costFee) }}
          </template>
        </template>
        <template #empty>
          <a-empty :description="t('winOrLoss.empty')" />
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import dayjs from 'dayjs';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import { formatAmount, formatCount, formatKillRate } from './format';
  import createWinLossSeed from './mock';
  import { buildQuickRange, detectQuickRange, matchWinLoss } from './query';
  import type {
    IdType,
    QuickRange,
    SortDirection,
    WinLossRecord,
    WinLossSortField,
  } from './types';

  const { t } = useI18n();
  const records = createWinLossSeed();
  const tableRef = ref<{ clearSorters: () => void }>();
  const quickRange = ref<QuickRange | ''>('');
  const form = reactive({
    dateRange: [] as string[],
    idType: 'memberId' as IdType,
    keyword: '',
  });
  const appliedRange = ref<string[]>([]);
  const appliedIdType = ref<IdType>('memberId');
  const appliedKeyword = ref('');
  const sortField = ref<WinLossSortField | ''>('');
  const sortDirection = ref<SortDirection>('');
  const pagination = reactive({
    current: 1,
    pageSize: 10,
    total: records.length,
    showTotal: true,
    showPageSize: true,
    pageSizeOptions: [10, 15, 20, 30, 50, 100],
    pageSizeProps: {
      popupContainer: 'body',
    },
  });

  const sortableFields: WinLossSortField[] = [
    'betCount',
    'bettorCount',
    'validBet',
    'memberWinLoss',
    'costFee',
  ];

  const numberSort = {
    sortDirections: ['ascend', 'descend'] as ('ascend' | 'descend')[],
    sorter: true,
  };

  const columns = computed<TableColumnData[]>(() => [
    {
      title: t('winOrLoss.column.date'),
      dataIndex: 'date',
      width: 150,
      align: 'center',
    },
    {
      title: t('winOrLoss.column.currency'),
      dataIndex: 'currency',
      width: 120,
      align: 'center',
    },
    {
      title: t('winOrLoss.column.betCount'),
      dataIndex: 'betCount',
      align: 'center',
      sortable: numberSort,
    },
    {
      title: t('winOrLoss.column.bettorCount'),
      dataIndex: 'bettorCount',
      align: 'center',
      sortable: numberSort,
    },
    {
      title: t('winOrLoss.column.validBet'),
      dataIndex: 'validBet',
      slotName: 'validBet',
      align: 'center',
      sortable: numberSort,
    },
    {
      title: t('winOrLoss.column.killRate'),
      slotName: 'killRate',
      align: 'center',
    },
    {
      title: t('winOrLoss.column.memberWinLoss'),
      dataIndex: 'memberWinLoss',
      slotName: 'memberWinLoss',
      align: 'center',
      sortable: numberSort,
    },
    {
      title: t('winOrLoss.column.costFee'),
      dataIndex: 'costFee',
      slotName: 'costFee',
      align: 'center',
      sortable: numberSort,
    },
  ]);

  const displayList = computed(() => {
    const list = records.filter((item) =>
      matchWinLoss(
        item,
        appliedRange.value,
        appliedIdType.value,
        appliedKeyword.value
      )
    );
    const field = sortField.value;
    const direction = sortDirection.value;
    if (!field || !direction) return list;
    const factor = direction === 'ascend' ? 1 : -1;
    return [...list].sort((a, b) => (a[field] - b[field]) * factor);
  });

  const summaryRow = computed(() => {
    const list = displayList.value;
    return {
      date: t('winOrLoss.summary.total'),
      betCount: list.reduce((sum, item) => sum + item.betCount, 0),
      validBet: list.reduce((sum, item) => sum + item.validBet, 0),
      memberWinLoss: list.reduce((sum, item) => sum + item.memberWinLoss, 0),
      costFee: list.reduce((sum, item) => sum + item.costFee, 0),
    };
  });

  function summary() {
    return [summaryRow.value];
  }

  watch(displayList, (list) => {
    pagination.total = list.length;
    const maxPage = Math.max(
      1,
      Math.ceil(list.length / pagination.pageSize) || 1
    );
    if (pagination.current > maxPage) pagination.current = 1;
  });

  function resolveRange() {
    if (form.dateRange[0] && form.dateRange[1]) return [...form.dateRange];
    if (quickRange.value) return buildQuickRange(quickRange.value);
    return [];
  }

  function handleSearch() {
    appliedRange.value = resolveRange();
    appliedIdType.value = form.idType;
    appliedKeyword.value = form.keyword.trim();
    pagination.current = 1;
  }

  function applyQuickRange(key: QuickRange) {
    quickRange.value = key;
    form.dateRange = [...buildQuickRange(key)];
  }

  function handleReset() {
    form.keyword = '';
    form.idType = 'memberId';
    form.dateRange = [];
    quickRange.value = '';
    appliedRange.value = [];
    appliedIdType.value = 'memberId';
    appliedKeyword.value = '';
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

  function csvCell(value: string | number) {
    const text = String(value ?? '');
    if (/[",\n]/.test(text)) return `"${text.replace(/"/g, '""')}"`;
    return text;
  }

  function rowCells(item: WinLossRecord) {
    return [
      item.date,
      item.currency,
      item.betCount,
      item.bettorCount,
      formatAmount(item.validBet),
      formatKillRate(item.memberWinLoss, item.validBet),
      formatAmount(item.memberWinLoss),
      formatAmount(item.costFee),
    ];
  }

  function downloadReport() {
    if (!displayList.value.length) {
      Message.warning(t('winOrLoss.message.exportEmpty'));
      return;
    }
    const headers = [
      t('winOrLoss.column.date'),
      t('winOrLoss.column.currency'),
      t('winOrLoss.column.betCount'),
      t('winOrLoss.column.bettorCount'),
      t('winOrLoss.column.validBet'),
      t('winOrLoss.column.killRate'),
      t('winOrLoss.column.memberWinLoss'),
      t('winOrLoss.column.costFee'),
    ];
    const total = summaryRow.value;
    const csv = [
      headers,
      ...displayList.value.map(rowCells),
      [
        total.date,
        '',
        formatCount(total.betCount),
        '',
        formatAmount(total.validBet),
        '',
        formatAmount(total.memberWinLoss),
        formatAmount(total.costFee),
      ],
    ]
      .map((cols) => cols.map((col) => csvCell(col)).join(','))
      .join('\n');
    const blob = new Blob([`\uFEFF${csv}`], {
      type: 'text/csv;charset=utf-8;',
    });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `win-loss-${dayjs().format('YYYYMMDD-HHmmss')}.csv`;
    link.click();
    URL.revokeObjectURL(url);
    Message.success(t('winOrLoss.message.exported'));
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

  .search-form :deep(.arco-form-item-label-col) {
    padding-right: 8px;
    line-height: 32px;
  }

  .search-fields {
    display: flex;
    flex-wrap: wrap;
    gap: 12px 16px;
    align-items: center;
  }

  .id-search {
    display: inline-flex;
  }

  .id-type {
    width: 112px;

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

  .search-input {
    width: 280px;
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

  .table-toolbar {
    margin-bottom: 16px;
  }

  :deep(.arco-table-th) {
    color: var(--color-text-1);
    font-weight: 600;
    background-color: var(--color-fill-2);
  }

  :deep(.arco-table-td) {
    color: var(--color-text-2);
  }

  :deep(.arco-table-tr-summary .arco-table-td) {
    color: var(--color-text-1);
    font-weight: 600;
    background-color: var(--color-fill-1);
  }

  :deep(.arco-pagination-options .arco-select) {
    min-width: 100px;
  }
</style>
