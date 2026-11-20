<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.route.baseData', 'menu.route.currency']" />
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item :label="$t('currencyManage.search.status')" field="status">
          <a-select
            v-model="searchForm.status"
            :placeholder="$t('front.common.placeholder.select')"
            allow-clear
            class="search-select"
          >
            <a-option
              v-for="item in statusOptions"
              :key="item.value"
              :value="item.value"
            >
              {{ item.label }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item>
          <a-space>
            <a-button type="primary" @click="handleSearch">
              <template #icon><icon-search /></template>
              {{ $t('front.common.search.search') }}
            </a-button>
            <a-button @click="handleReset">
              <template #icon><icon-refresh /></template>
              {{ $t('front.common.search.reset') }}
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </a-card>
    <a-card class="general-card table-card" :bordered="false">
      <div class="table-toolbar">
        <a-space>
          <!-- 新的后端接口不再支持新建货币，故「新建货币」按钮已移除。 -->
          <!-- <a-button type="primary" @click="handleCreate">
            <template #icon><icon-plus /></template>
            {{ $t('front.common.toolbar.add') }}
          </a-button> -->
          <a-button @click="() => fetchData()">
            <template #icon><icon-refresh /></template>
            {{ $t('front.common.toolbar.refresh') }}
          </a-button>
        </a-space>
        <a-alert
          v-if="isFiltering"
          type="warning"
          :show-icon="false"
          class="drag-tip"
        >
          {{ $t('currencyManage.drag.tipFilter') }}
        </a-alert>
        <a-alert v-else type="info" :show-icon="false" class="drag-tip">
          {{ $t('currencyManage.drag.tipOk') }}
          <icon-drag-arrow />
        </a-alert>
      </div>
      <a-table
        ref="tableRef"
        row-key="id"
        :loading="loading"
        :columns="columns"
        :data="displayList"
        :pagination="pagination"
        :bordered="true"
        :stripe="true"
        :scroll="{ x: 1300 }"
        :class="{
          'is-row-dragging': isRowDragging || isDragPersisting,
        }"
        @page-change="onPageChange"
        @page-size-change="onPageSizeChange"
      >
        <template #dragHandle>
          <icon-drag-arrow
            class="row-drag-handle"
            :class="{ 'row-drag-handle--disabled': isFiltering }"
          />
        </template>
        <template #index="{ rowIndex }">
          {{ rowIndex + 1 + (pagination.current - 1) * pagination.pageSize }}
        </template>
        <template #currencyType="{ record }">
          {{
            record.currencyType === 1
              ? $t('currencyManage.type.fiat')
              : record.currencyType === 2
              ? $t('currencyManage.type.virtual')
              : '--'
          }}
        </template>
        <template #status="{ record }">
          <a-switch
            :model-value="record.status === 1"
            :checked-text="$t('front.common.status.enabled')"
            :unchecked-text="$t('front.common.status.disabled')"
            size="small"
            :loading="statusUpdatingId === record.id"
            @change="(value) => handleStatusChange(record, Boolean(value))"
          />
        </template>
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <a-link @click="handleEdit(record)">{{
              $t('front.common.action.edit')
            }}</a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>
    <CurrencyModal
      :visible="modalVisible"
      :record="editingCurrency"
      @update:visible="onModalVisibleChange"
      @success="fetchData"
    />
  </div>
</template>

<script lang="ts" setup>
  import {
    computed,
    nextTick,
    onBeforeUnmount,
    reactive,
    ref,
    watch,
  } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import { Message } from '@arco-design/web-vue';
  import Sortable from 'sortablejs';
  import {
    getCurrency,
    queryCurrencyList,
    reorderCurrency,
    updateCurrency,
    type CurrencyRecord,
    type CurrencyListParams,
  } from '@/api/currency';
  import useLoading from '@/hooks/loading';
  import CurrencyModal from './components/currency-modal.vue';

  interface CurrencyRow {
    id: number;
    code: string;
    nameKey: string;
    name: string;
    currencyType: number;
    symbol: string;
    amountFactor: number;
    sortNo: number;
    status: number;
    raw: CurrencyRecord;
  }

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(true);
  const modalVisible = ref(false);
  const editingCurrency = ref<CurrencyRecord | null>(null);
  const statusUpdatingId = ref<number | null>(null);
  const tableRef = ref();
  let rowSortable: Sortable | null = null;
  /** 拖拽落库中，禁止分页跳转 */
  const isDragPersisting = ref(false);
  /** 行拖拽进行中（防止松手误点到底部分页） */
  const isRowDragging = ref(false);

  const statusOptions = computed(() => [
    { label: t('front.common.status.enabled'), value: 1 },
    { label: t('front.common.status.disabled'), value: 2 },
  ]);

  const searchForm = reactive<{
    status: number | undefined;
  }>({
    status: undefined,
  });
  const appliedSearch = reactive({
    status: undefined as number | undefined,
  });
  /** 存在筛选条件时，当前页并非连续的排序区间，禁止拖拽排序 */
  const isFiltering = computed(() => appliedSearch.status !== undefined);

  const currencies = ref<CurrencyRow[]>([]);
  const pagination = reactive({
    current: 1,
    pageSize: 20,
    total: 0,
    showTotal: true,
    showPageSize: true,
    pageSizeOptions: [10, 20, 30, 50, 100],
    pageSizeProps: {
      popupContainer: 'body',
    },
  });

  const mapCurrency = (item: CurrencyRecord): CurrencyRow => ({
    id: item.id,
    code: item.code || '--',
    nameKey: item.name_key || '--',
    name: item.name || '--',
    currencyType: item.currency_type,
    symbol: item.symbol || '',
    amountFactor: item.amount_factor ?? 0,
    sortNo: item.sort_no ?? 0,
    status: item.status,
    raw: item,
  });

  const displayList = computed(() => currencies.value);

  const buildQueryParams = (
    page = pagination.current,
    pageSize = pagination.pageSize
  ): CurrencyListParams => {
    const params: CurrencyListParams = {
      page,
      page_size: pageSize,
    };
    if (appliedSearch.status !== undefined)
      params.status = appliedSearch.status;
    return params;
  };

  const fetchData = async (
    page = pagination.current,
    pageSize = pagination.pageSize
  ) => {
    setLoading(true);
    try {
      const { data } = await queryCurrencyList(
        buildQueryParams(page, pageSize)
      );
      currencies.value = (data?.list || []).map(mapCurrency);
      pagination.current = page;
      pagination.pageSize = pageSize;
      pagination.total = data?.total ?? 0;
    } catch {
      currencies.value = [];
      pagination.total = 0;
    } finally {
      setLoading(false);
    }
  };

  const columns = computed<TableColumnData[]>(() => [
    {
      title: t('currencyManage.column.drag'),
      dataIndex: 'dragHandle',
      slotName: 'dragHandle',
      width: 70,
      align: 'center',
      fixed: 'left',
    },
    {
      title: t('front.common.column.index'),
      dataIndex: 'index',
      slotName: 'index',
      width: 80,
      align: 'center',
      fixed: 'left',
    },
    {
      title: t('currencyManage.column.code'),
      dataIndex: 'code',
      width: 120,
      align: 'center',
    },
    {
      title: t('currencyManage.column.name'),
      dataIndex: 'name',
      width: 140,
      align: 'center',
    },
    // 这列暂时没有用到，先隐藏
    // {
    //   title: '名称 Key',
    //   dataIndex: 'nameKey',
    //   width: 200,
    //   align: 'center',
    //   ellipsis: true,
    //   tooltip: true,
    // },
    {
      title: t('currencyManage.column.currencyType'),
      dataIndex: 'currencyType',
      slotName: 'currencyType',
      width: 120,
      align: 'center',
    },
    {
      title: t('currencyManage.column.symbol'),
      dataIndex: 'symbol',
      width: 140,
      align: 'center',
    },
    {
      title: t('currencyManage.column.amountFactor'),
      dataIndex: 'amountFactor',
      width: 110,
      align: 'center',
    },
    // 这列暂时没有用到，先隐藏
    // {
    //   title: t('currencyManage.column.sortNo'),
    //   dataIndex: 'sortNo',
    //   width: 90,
    //   align: 'center',
    // },
    {
      title: t('front.common.column.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 120,
      align: 'center',
    },
    // 操作列暂时隐藏
    // {
    //   title: t('front.common.column.operations'),
    //   dataIndex: 'operations',
    //   slotName: 'operations',
    //   width: 100,
    //   align: 'center',
    //   fixed: 'right',
    // },
  ]);

  const onModalVisibleChange = (visible: boolean) => {
    modalVisible.value = visible;
    if (!visible) {
      editingCurrency.value = null;
    }
  };

  const handleEdit = async (record: CurrencyRow) => {
    setLoading(true);
    try {
      const { data } = await getCurrency({ id: record.id });
      editingCurrency.value = data?.id ? data : record.raw;
      modalVisible.value = true;
    } catch {
      // 全局拦截器已提示错误
    } finally {
      setLoading(false);
    }
  };

  const handleStatusChange = async (record: CurrencyRow, enabled: boolean) => {
    if (statusUpdatingId.value === record.id) return;
    const nextStatus = enabled ? 1 : 2;
    if (record.status === nextStatus) return;
    statusUpdatingId.value = record.id;
    try {
      await updateCurrency({ id: record.id, status: nextStatus });
      record.status = nextStatus;
      record.raw.status = nextStatus;
      Message.success(
        enabled
          ? t('front.common.message.enabled')
          : t('currencyManage.message.disabled')
      );
    } catch {
      // 保持原状态；错误提示由拦截器处理
    } finally {
      statusUpdatingId.value = null;
    }
  };

  const handleSearch = () => {
    appliedSearch.status = searchForm.status;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    searchForm.status = undefined;
    appliedSearch.status = undefined;
    fetchData(1, pagination.pageSize);
  };

  const onPageChange = (current: number) => {
    if (isDragPersisting.value || isRowDragging.value) return;
    fetchData(current, pagination.pageSize);
  };

  const onPageSizeChange = (pageSize: number) => {
    if (isDragPersisting.value || isRowDragging.value) return;
    fetchData(1, pageSize);
  };

  /**
   * 货币排序由后端 /currency/reorder 接口维护，仅需传入被移动的货币 id
   * 与目标位置原本所在的货币 target_id，接口会自动处理排序值。
   */
  const persistDragOrder = async (
    movedId: number,
    targetId: number,
    rows: CurrencyRow[]
  ) => {
    const { current: page, pageSize } = pagination;
    isDragPersisting.value = true;
    setLoading(true);
    try {
      await reorderCurrency({ id: movedId, target_id: targetId });
      Message.success(t('currencyManage.message.reorderSuccess'));
    } catch {
      currencies.value = rows;
      Message.error(t('currencyManage.message.reorderFail'));
    } finally {
      try {
        await fetchData(page, pageSize);
        await nextTick();
      } finally {
        isDragPersisting.value = false;
      }
    }
  };

  const destroyRowSortable = () => {
    rowSortable?.destroy();
    rowSortable = null;
  };

  const initRowSortable = async () => {
    destroyRowSortable();
    if (isFiltering.value || loading.value) return;
    await nextTick();
    const root = tableRef.value?.$el as HTMLElement | undefined;
    const tbody = root?.querySelector(
      '.arco-table-body tbody, .arco-table-content-inner tbody, tbody'
    ) as HTMLElement | null;
    if (!tbody) return;

    rowSortable = new Sortable(tbody, {
      handle: '.row-drag-handle',
      animation: 150,
      draggable: '.arco-table-tr',
      filter: '.arco-table-tr-empty',
      preventOnFilter: true,
      onMove: () => !isFiltering.value,
      onStart: () => {
        isRowDragging.value = true;
      },
      onEnd: (evt) => {
        const { oldIndex, newIndex } = evt;
        if (
          oldIndex == null ||
          newIndex == null ||
          oldIndex === newIndex ||
          isFiltering.value
        ) {
          isRowDragging.value = false;
          return;
        }
        const original = [...currencies.value];
        if (oldIndex >= original.length || newIndex >= original.length) {
          isRowDragging.value = false;
          return;
        }
        const movedId = original[oldIndex].id;
        const targetId = original[newIndex].id;

        const list = [...original];
        const [moved] = list.splice(oldIndex, 1);
        list.splice(newIndex, 0, moved);
        currencies.value = list;

        persistDragOrder(movedId, targetId, original).finally(() => {
          isRowDragging.value = false;
        });
      },
    });
  };

  watch(
    [
      loading,
      isFiltering,
      () => currencies.value.map((item) => item.id).join(','),
    ],
    ([isLoading]) => {
      if (isLoading) {
        destroyRowSortable();
        return;
      }
      initRowSortable();
    }
  );

  onBeforeUnmount(() => {
    destroyRowSortable();
  });

  fetchData();
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
  .search-form :deep(.arco-form-item) {
    margin-bottom: 0;
  }
  .search-select :deep(.arco-select-view) {
    min-width: 130px;
  }
  .table-toolbar {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 12px;
    margin-bottom: 14px;
  }
  .drag-tip {
    padding: 4px 12px;
  }
  .row-drag-handle {
    color: var(--color-text-3);
    cursor: grab;
    font-size: 16px;
  }
  .row-drag-handle:active {
    cursor: grabbing;
  }
  .row-drag-handle--disabled {
    cursor: not-allowed;
    opacity: 0.4;
  }
  .is-row-dragging :deep(.arco-pagination) {
    pointer-events: none;
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
  .search-form {
    row-gap: 16px;
    column-gap: 16px;
  }
</style>
