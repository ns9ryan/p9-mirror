<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.gameManage', 'menu.gameManage.category']" />
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item :label="t('categoryManage.search.categoryCode')" field="categoryCode">
          <a-input
            v-model="searchForm.categoryCode"
            :placeholder="t('categoryManage.search.categoryCodePlaceholder')"
            allow-clear
            class="search-input"
          />
        </a-form-item>
        <a-form-item :label="t('categoryManage.search.status')" field="status">
          <a-select
            v-model="searchForm.status"
            :placeholder="t('front.common.placeholder.select')"
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
        <a-form-item :label="t('categoryManage.search.isDeleted')" field="isDeleted">
          <a-select
            v-model="searchForm.isDeleted"
            :placeholder="t('front.common.placeholder.select')"
            allow-clear
            class="search-select"
          >
            <a-option
              v-for="item in deletedOptions"
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
              {{ t('front.common.search.search') }}
            </a-button>
            <a-button @click="handleReset">
              <template #icon><icon-refresh /></template>
              {{ t('front.common.search.reset') }}
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </a-card>
    <a-card class="general-card table-card" :bordered="false">
      <div class="table-toolbar">
        <a-space>
          <a-button type="primary" @click="handleSync">
            <template #icon><icon-sync /></template>
            {{ t('categoryManage.toolbar.sync') }}
          </a-button>
          <a-button
            :type="isSyncProgressDone ? 'outline' : undefined"
            :status="isSyncProgressDone ? 'success' : undefined"
            :loading="syncProgressLoading"
            :disabled="isSyncProgressDone"
            :class="{ 'sync-progress-btn--done': isSyncProgressDone }"
            @click="fetchSyncProgress"
          >
            {{ t('categoryManage.toolbar.syncProgress', { progress: syncProgress }) }}
          </a-button>
          <a-button @click="() => fetchData()">
            <template #icon><icon-refresh /></template>
            {{ t('front.common.toolbar.refresh') }}
          </a-button>
          <a-button
            :type="expandAllColumns ? 'outline' : 'secondary'"
            @click="expandAllColumns = !expandAllColumns"
          >
            {{ expandAllColumns ? t('categoryManage.toolbar.collapseColumns') : t('categoryManage.toolbar.expandAll') }}
          </a-button>
        </a-space>
        <a-alert
          v-if="isFiltering"
          type="warning"
          :show-icon="false"
          class="drag-tip"
        >
          {{ t('categoryManage.drag.tipFilter') }}
        </a-alert>
        <a-alert v-else type="info" :show-icon="false" class="drag-tip">
          {{ t('categoryManage.drag.tipOk') }}
        </a-alert>
      </div>
      <a-table
        ref="tableRef"
        row-key="id"
        :loading="loading"
        :columns="displayColumns"
        :data="displayList"
        :pagination="pagination"
        :bordered="true"
        :stripe="true"
        :scroll="expandAllColumns ? undefined : { x: 1400 }"
        :class="{
          'table-expand-all': expandAllColumns,
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
        <template #status="{ record }">
          <a-tag :color="record.status === 1 ? 'green' : 'gray'">
            {{ record.status === 1 ? t('front.common.status.enabled') : t('front.common.status.disabled') }}
          </a-tag>
        </template>
        <template #sourceStatus="{ record }">
          <a-tag :color="record.sourceStatus === 1 ? 'green' : 'gray'">
            {{ record.sourceStatus === 1 ? t('front.common.status.enabled') : t('front.common.status.disabled') }}
          </a-tag>
        </template>
        <template #isDeleted="{ record }">
          <a-tag :color="record.isDeleted ? 'red' : 'arcoblue'">
            {{ record.isDeleted ? t('categoryManage.search.deleted') : t('categoryManage.search.notDeleted') }}
          </a-tag>
        </template>
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <a-link @click="handleEdit(record)">{{ t('front.common.action.edit') }}</a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>
    <CategoryModal
      :visible="modalVisible"
      :category="editingCategory"
      @update:visible="onModalVisibleChange"
      @success="fetchData"
    />
    <SyncPreviewModal
      v-model:visible="syncModalVisible"
      @success="onSyncSuccess"
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
  import dayjs from 'dayjs';
  import Sortable from 'sortablejs';
  import {
    queryGameCategoryList,
    updateGameCategory,
    type GameCategoryRecord,
    type GameCategoryListParams,
  } from '@/api/gameCategory';
  import { GAME_SYNC_SCOPE } from '@/api/gameSyncCheckpoint';
  import useLoading from '@/hooks/loading';
  import useSyncProgress from '@/hooks/sync-progress';
  import CategoryModal from './components/category-modal.vue';
  import SyncPreviewModal from './components/sync-preview-modal.vue';

  interface GameCategoryRow {
    id: number;
    categoryCode: string;
    name: string;
    sortNo: number;
    status: number;
    sourceId: number;
    sourceCategoryCode: string;
    sourceName: string;
    sourceSortNo: number;
    sourceStatus: number;
    isDeleted: number;
    createdAt: string;
    updatedAt: string;
    raw: GameCategoryRecord;
  }

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(true);
  const modalVisible = ref(false);
  const editingCategory = ref<GameCategoryRecord | null>(null);
  /** 截图用：取消横向滚动与固定列，一屏展示全部列 */
  const expandAllColumns = ref(false);
  const syncModalVisible = ref(false);
  const {
    syncProgress,
    syncProgressLoading,
    isSyncProgressDone,
    fetchSyncProgress,
    scheduleFetchSyncProgress,
  } = useSyncProgress(GAME_SYNC_SCOPE.CATEGORY);
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
  const deletedOptions = computed(() => [
    { label: t('categoryManage.search.notDeleted'), value: 0 },
    { label: t('categoryManage.search.deleted'), value: 1 },
  ]);

  const searchForm = reactive<{
    categoryCode: string;
    name: string;
    status: number | undefined;
    isDeleted: number | undefined;
  }>({
    categoryCode: '',
    name: '',
    status: undefined,
    isDeleted: 0,
  });
  const appliedSearch = reactive({
    categoryCode: '',
    name: '',
    status: undefined as number | undefined,
    isDeleted: 0 as number | undefined,
  });
  /** 存在筛选条件时，当前页并非连续的排序区间，禁止拖拽排序 */
  const isFiltering = computed(
    () =>
      Boolean(appliedSearch.categoryCode) ||
      Boolean(appliedSearch.name) ||
      appliedSearch.status !== undefined
  );

  const categories = ref<GameCategoryRow[]>([]);
  const pagination = reactive({
    current: 1,
    pageSize: 10,
    total: 0,
    showTotal: true,
    showPageSize: true,
    pageSizeOptions: [10, 20, 30, 50, 100],
    pageSizeProps: {
      popupContainer: 'body',
    },
  });

  const formatTime = (value?: number) => {
    if (!value) return '--';
    const ms = value < 1e12 ? value * 1000 : value;
    return dayjs(ms).format('YYYY-MM-DD HH:mm:ss');
  };

  /** name 为多语言 JSON 字符串，优先展示 default 字段 */
  const parseI18nName = (value?: string) => {
    if (!value) return '--';
    try {
      const parsed = JSON.parse(value);
      if (parsed && typeof parsed === 'object') {
        return parsed.default || Object.values(parsed)[0] || '--';
      }
      return value;
    } catch {
      return value;
    }
  };

  const mapCategory = (item: GameCategoryRecord): GameCategoryRow => ({
    id: item.id,
    categoryCode: item.category_code || '--',
    name: parseI18nName(item.name),
    sortNo: item.sort_no ?? 0,
    status: item.status,
    sourceId: item.source_id,
    sourceCategoryCode: item.source_category_code || '--',
    sourceName: parseI18nName(item.source_name_i18n),
    sourceSortNo: item.source_sort_no ?? 0,
    sourceStatus: item.source_status,
    isDeleted: item.is_deleted,
    createdAt: formatTime(item.created_at),
    updatedAt: formatTime(item.updated_at),
    raw: item,
  });

  const displayList = computed(() => categories.value);

  const buildQueryParams = (
    page = pagination.current,
    pageSize = pagination.pageSize
  ): GameCategoryListParams => {
    const params: GameCategoryListParams = {
      page,
      page_size: pageSize,
      is_deleted: appliedSearch.isDeleted ?? 0,
    };
    const categoryCode = appliedSearch.categoryCode.trim();
    const name = appliedSearch.name.trim();
    if (categoryCode) params.category_code = categoryCode;
    if (name) params.name = name;
    if (appliedSearch.status !== undefined)
      params.status = appliedSearch.status;
    params.sort_by = 'sort_no';
    params.sort_order = 'asc';
    return params;
  };

  const fetchData = async (
    page = pagination.current,
    pageSize = pagination.pageSize
  ) => {
    setLoading(true);
    try {
      const { data } = await queryGameCategoryList(
        buildQueryParams(page, pageSize)
      );
      categories.value = (data?.list || []).map(mapCategory);
      pagination.current = page;
      pagination.pageSize = pageSize;
      pagination.total = data?.total ?? 0;
    } catch {
      categories.value = [];
      pagination.total = 0;
    } finally {
      setLoading(false);
    }
  };

  const columns = computed<TableColumnData[]>(() => [
    {
      title: t('categoryManage.column.drag'),
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
      title: t('categoryManage.column.categoryCode'),
      dataIndex: 'categoryCode',
      width: 140,
      align: 'center',
    },
    {
      title: t('categoryManage.column.name'),
      dataIndex: 'name',
      width: 150,
      align: 'center',
    },
    {
      title: t('categoryManage.column.sortNo'),
      dataIndex: 'sortNo',
      width: 90,
      align: 'center',
    },
    {
      title: t('front.common.column.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 90,
      align: 'center',
    },
    {
      title: t('categoryManage.column.isDeleted'),
      dataIndex: 'isDeleted',
      slotName: 'isDeleted',
      width: 100,
      align: 'center',
    },
    {
      title: t('front.common.column.createdAt'),
      dataIndex: 'createdAt',
      width: 170,
      align: 'center',
    },
    {
      title: t('front.common.column.updatedAt'),
      dataIndex: 'updatedAt',
      width: 170,
      align: 'center',
    },
    {
      title: t('front.common.column.operations'),
      dataIndex: 'operations',
      slotName: 'operations',
      width: 100,
      align: 'center',
      fixed: 'right',
    },
  ]);

  const displayColumns = computed(() => {
    if (!expandAllColumns.value) return columns.value;
    return columns.value.map((col) => {
      const next = { ...col };
      delete next.width;
      delete next.fixed;
      return next;
    });
  });

  const onModalVisibleChange = (visible: boolean) => {
    modalVisible.value = visible;
    if (!visible) {
      editingCategory.value = null;
    }
  };

  const handleEdit = (record: GameCategoryRow) => {
    editingCategory.value = record.raw;
    modalVisible.value = true;
  };

  const handleSearch = () => {
    appliedSearch.categoryCode = searchForm.categoryCode;
    appliedSearch.name = searchForm.name;
    appliedSearch.status = searchForm.status;
    appliedSearch.isDeleted = searchForm.isDeleted;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    Object.assign(searchForm, {
      categoryCode: '',
      name: '',
      status: undefined,
      isDeleted: 0,
    });
    Object.assign(appliedSearch, {
      categoryCode: '',
      name: '',
      status: undefined,
      isDeleted: 0,
    });
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

  const handleSync = () => {
    syncModalVisible.value = true;
  };

  const onSyncSuccess = () => {
    fetchData();
    scheduleFetchSyncProgress(1000);
  };

  /**
   * 仅重排拖拽起点到终点这段区间内的 sort_no，区间外的行不受影响、不调接口。
   * 区间内原 sort_no 若已是互不相同的正数，直接复用这些值按新顺序重新分配；
   * 否则（存在重复/非正数）才按页内位置从 1 起连续编号兜底。
   * 落库期间锁定页码，避免误点分页或表格内部改页导致跳到其他页。
   */
  const persistDragOrder = async (
    rows: GameCategoryRow[],
    oldIndex: number,
    newIndex: number
  ) => {
    const { current: page, pageSize } = pagination;
    const from = Math.min(oldIndex, newIndex);
    const to = Math.max(oldIndex, newIndex);
    const affected = rows.slice(from, to + 1);
    const reused = affected.map((row) => row.sortNo).sort((a, b) => a - b);
    const canReuse =
      reused.every((n) => n > 0) && new Set(reused).size === reused.length;
    const base = (page - 1) * pageSize;
    const targets = canReuse
      ? reused
      : affected.map((_, i) => base + from + i + 1);

    const changed = affected
      .map((row, i) => ({ row, sortNo: targets[i] }))
      .filter(({ row, sortNo }) => row.sortNo !== sortNo);
    if (!changed.length) return;

    categories.value = rows.map((row, index) => {
      if (index < from || index > to) return row;
      return { ...row, sortNo: targets[index - from] };
    });

    isDragPersisting.value = true;
    setLoading(true);
    try {
      await changed.reduce(async (prevPromise, { row, sortNo }) => {
        await prevPromise;
        const name =
          typeof row.raw.name === 'string'
            ? row.raw.name
            : JSON.stringify(row.raw.name ?? {});
        await updateGameCategory({
          id: row.id,
          name,
          sort_no: sortNo,
          status: row.raw.status,
        });
      }, Promise.resolve());
      Message.success(t('categoryManage.message.reorderSuccess'));
    } catch {
      Message.error(t('categoryManage.message.reorderFail'));
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
        const list = [...categories.value];
        if (oldIndex >= list.length || newIndex >= list.length) {
          isRowDragging.value = false;
          return;
        }
        const [moved] = list.splice(oldIndex, 1);
        list.splice(newIndex, 0, moved);
        categories.value = list;
        persistDragOrder(list, oldIndex, newIndex).finally(() => {
          isRowDragging.value = false;
        });
      },
    });
  };

  watch(
    [
      loading,
      isFiltering,
      () => categories.value.map((item) => item.id).join(','),
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
  fetchSyncProgress();
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
  .search-form :deep(.arco-select-view) {
    min-width: 160px;
  }
  .search-input {
    width: 170px;
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
  .table-expand-all :deep(.arco-table-td),
  .table-expand-all :deep(.arco-table-th) {
    white-space: normal;
    word-break: break-all;
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
