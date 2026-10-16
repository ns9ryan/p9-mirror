<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.gameManage', 'menu.gameManage.manufacturer']" />
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item :label="t('manufacturerManage.search.providerCode')" field="providerCode">
          <a-input
            v-model="searchForm.providerCode"
            :placeholder="t('manufacturerManage.search.providerCodePlaceholder')"
            allow-clear
            class="search-input"
          />
        </a-form-item>
        <a-form-item :label="t('manufacturerManage.search.name')" field="name">
          <a-input
            v-model="searchForm.name"
            :placeholder="t('manufacturerManage.search.namePlaceholder')"
            allow-clear
            class="search-input"
          />
        </a-form-item>
        <a-form-item :label="t('manufacturerManage.search.status')" field="status">
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
        <a-form-item :label="t('manufacturerManage.search.isDeleted')" field="isDeleted">
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
        <div class="table-toolbar-left">
          <a-space>
            <a-button type="primary" @click="handleSync">
              <template #icon><icon-sync /></template>
              {{ t('manufacturerManage.toolbar.sync') }}
            </a-button>
            <a-button
              :type="isSyncProgressDone ? 'outline' : undefined"
              :status="isSyncProgressDone ? 'success' : undefined"
              :loading="syncProgressLoading"
              :disabled="isSyncProgressDone"
              :class="{ 'sync-progress-btn--done': isSyncProgressDone }"
              @click="fetchSyncProgress"
            >
              {{ t('manufacturerManage.toolbar.syncProgress', { progress: syncProgress }) }}
            </a-button>
            <a-button @click="() => fetchData()">
              <template #icon><icon-refresh /></template>
              {{ t('front.common.toolbar.refresh') }}
            </a-button>
            <a-button
              :type="expandAllColumns ? 'outline' : 'secondary'"
              @click="expandAllColumns = !expandAllColumns"
            >
              {{ expandAllColumns ? t('manufacturerManage.toolbar.collapseColumns') : t('manufacturerManage.toolbar.expandAll') }}
            </a-button>
          </a-space>
          <a-alert
            v-if="isFiltering"
            type="warning"
            :show-icon="false"
            class="drag-tip"
          >
            {{ t('manufacturerManage.drag.tipFilter') }}
          </a-alert>
          <a-alert v-else type="info" :show-icon="false" class="drag-tip">
            {{ t('manufacturerManage.drag.tipOk') }}
          </a-alert>
        </div>
        <a-popover
          trigger="click"
          position="br"
          @popup-visible-change="popupVisibleChange"
        >
          <a-button>
            <template #icon><icon-settings /></template>
            {{ t('manufacturerManage.toolbar.columnSettings') }}
          </a-button>
          <template #content>
            <div ref="columnSettingRef" class="column-setting-list">
              <div
                v-for="item in showColumns"
                :key="item.dataIndex"
                class="column-setting-item"
              >
                <icon-drag-arrow class="drag-handle" />
                <a-checkbox
                  v-model="item.checked"
                  :disabled="
                    item.dataIndex === 'operations' ||
                    item.dataIndex === 'dragHandle'
                  "
                  @change="handleColumnCheckedChange"
                >
                  {{ item.title }}
                </a-checkbox>
              </div>
            </div>
          </template>
        </a-popover>
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
        :scroll="tableScroll"
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
        <template #logo="{ record }">
          <a-image
            v-if="record.logoUrl"
            :src="record.logoUrl"
            width="40"
            height="40"
            fit="contain"
            show-loader
          />
          <span v-else>--</span>
        </template>
        <template #sourceLogo="{ record }">
          <a-image
            v-if="record.sourceLogoUrl"
            :src="record.sourceLogoUrl"
            width="40"
            height="40"
            fit="contain"
            show-loader
          />
          <span v-else>--</span>
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
            {{ record.isDeleted ? t('manufacturerManage.search.deleted') : t('manufacturerManage.search.notDeleted') }}
          </a-tag>
        </template>
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <a-link @click="handleEdit(record)">{{ t('front.common.action.edit') }}</a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>
    <ProviderModal
      :visible="modalVisible"
      :provider="editingProvider"
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
  import cloneDeep from 'lodash/cloneDeep';
  import Sortable from 'sortablejs';
  import dayjs from 'dayjs';
  import {
    queryGameProviderList,
    updateGameProvider,
    type GameProviderRecord,
    type GameProviderListParams,
  } from '@/api/gameProvider';
  import { GAME_SYNC_SCOPE } from '@/api/gameSyncCheckpoint';
  import useLoading from '@/hooks/loading';
  import useSyncProgress from '@/hooks/sync-progress';
  import ProviderModal from './components/provider-modal.vue';
  import SyncPreviewModal from './components/sync-preview-modal.vue';

  type ColumnSettingItem = TableColumnData & { checked?: boolean };

  interface GameProviderRow {
    id: number;
    providerCode: string;
    name: string;
    logoUrl: string;
    sortNo: number;
    status: number;
    sourceId: number;
    sourceProviderCode: string;
    sourceName: string;
    sourceLogoUrl: string;
    sourceSortNo: number;
    sourceStatus: number;
    isDeleted: number;
    createdAt: string;
    updatedAt: string;
    raw: GameProviderRecord;
  }

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(true);
  const modalVisible = ref(false);
  const editingProvider = ref<GameProviderRecord | null>(null);
  /** 截图用：取消横向滚动与固定列，一屏展示全部列 */
  const expandAllColumns = ref(false);
  const syncModalVisible = ref(false);
  const {
    syncProgress,
    syncProgressLoading,
    isSyncProgressDone,
    fetchSyncProgress,
    scheduleFetchSyncProgress,
  } = useSyncProgress(GAME_SYNC_SCOPE.PROVIDER);
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
    { label: t('manufacturerManage.search.notDeleted'), value: 0 },
    { label: t('manufacturerManage.search.deleted'), value: 1 },
  ]);

  const searchForm = reactive<{
    providerCode: string;
    name: string;
    status: number | undefined;
    isDeleted: number | undefined;
  }>({
    providerCode: '',
    name: '',
    status: undefined,
    isDeleted: 0,
  });
  const appliedSearch = reactive({
    providerCode: '',
    name: '',
    status: undefined as number | undefined,
    isDeleted: 0 as number | undefined,
  });
  /** 存在筛选条件时，当前页并非连续的排序区间，禁止拖拽排序 */
  const isFiltering = computed(
    () =>
      Boolean(appliedSearch.providerCode) ||
      Boolean(appliedSearch.name) ||
      appliedSearch.status !== undefined
  );

  const providers = ref<GameProviderRow[]>([]);
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

  const mapProvider = (item: GameProviderRecord): GameProviderRow => ({
    id: item.id,
    providerCode: item.provider_code || '--',
    name: parseI18nName(item.name),
    logoUrl: item.logo_url || '',
    sortNo: item.sort_no ?? 0,
    status: item.status,
    sourceId: item.source_id,
    sourceProviderCode: item.source_provider_code || '--',
    sourceName: parseI18nName(item.source_name_i18n),
    sourceLogoUrl: item.source_logo_url || '',
    sourceSortNo: item.source_sort_no ?? 0,
    sourceStatus: item.source_status,
    isDeleted: item.is_deleted,
    createdAt: formatTime(item.created_at),
    updatedAt: formatTime(item.updated_at),
    raw: item,
  });

  const displayList = computed(() => providers.value);

  const buildQueryParams = (
    page = pagination.current,
    pageSize = pagination.pageSize
  ): GameProviderListParams => {
    const params: GameProviderListParams = {
      page,
      page_size: pageSize,
      is_deleted: appliedSearch.isDeleted ?? 0,
    };
    const providerCode = appliedSearch.providerCode.trim();
    const name = appliedSearch.name.trim();
    if (providerCode) params.provider_code = providerCode;
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
      const { data } = await queryGameProviderList(
        buildQueryParams(page, pageSize)
      );
      providers.value = (data?.list || []).map(mapProvider);
      pagination.current = page;
      pagination.pageSize = pageSize;
      pagination.total = data?.total ?? 0;
    } catch {
      providers.value = [];
      pagination.total = 0;
    } finally {
      setLoading(false);
    }
  };

  const columns = computed<TableColumnData[]>(() => [
    {
      title: t('manufacturerManage.column.drag'),
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
      title: t('manufacturerManage.column.providerCode'),
      dataIndex: 'providerCode',
      width: 140,
      align: 'center',
    },
    {
      title: t('manufacturerManage.column.logo'),
      dataIndex: 'logoUrl',
      slotName: 'logo',
      width: 100,
      align: 'center',
    },
    {
      title: t('manufacturerManage.column.name'),
      dataIndex: 'name',
      width: 150,
      align: 'center',
    },
    {
      title: t('manufacturerManage.column.sortNo'),
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
    // {
    //   title: t('manufacturerManage.column.sourceProviderCode'),
    //   dataIndex: 'sourceProviderCode',
    //   width: 150,
    //   align: 'center',
    // },
    // {
    //   title: t('manufacturerManage.column.sourceLogo'),
    //   dataIndex: 'sourceLogoUrl',
    //   slotName: 'sourceLogo',
    //   width: 100,
    //   align: 'center',
    // },
    // {
    //   title: t('manufacturerManage.column.sourceName'),
    //   dataIndex: 'sourceName',
    //   width: 150,
    //   align: 'center',
    // },
    // {
    //   title: t('manufacturerManage.column.sourceSortNo'),
    //   dataIndex: 'sourceSortNo',
    //   width: 100,
    //   align: 'center',
    // },
    // {
    //   title: t('manufacturerManage.column.sourceStatus'),
    //   dataIndex: 'sourceStatus',
    //   slotName: 'sourceStatus',
    //   width: 100,
    //   align: 'center',
    // },
    {
      title: t('manufacturerManage.column.isDeleted'),
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

  const showColumns = ref<ColumnSettingItem[]>([]);
  const columnSettingRef = ref<HTMLElement | null>(null);
  let columnSortable: Sortable | null = null;

  const initColumnSettings = () => {
    showColumns.value = cloneDeep(columns.value).map((item) => ({
      ...item,
      checked: true,
    }));
  };

  const syncVisibleColumns = () => {
    showColumns.value = [...showColumns.value];
  };

  const handleColumnCheckedChange = () => {
    syncVisibleColumns();
  };

  const exchangeArray = (
    array: ColumnSettingItem[],
    beforeIdx: number,
    newIdx: number
  ) => {
    if (beforeIdx < 0 || newIdx < 0 || beforeIdx === newIdx) return;
    const [moved] = array.splice(beforeIdx, 1);
    array.splice(newIdx, 0, moved);
  };

  const popupVisibleChange = (visible: boolean) => {
    if (!visible) return;
    nextTick(() => {
      const el = columnSettingRef.value;
      if (!el) return;
      columnSortable?.destroy();
      columnSortable = new Sortable(el, {
        handle: '.drag-handle',
        animation: 150,
        onEnd(e) {
          const { oldIndex, newIndex } = e;
          if (oldIndex == null || newIndex == null) return;
          exchangeArray(showColumns.value, oldIndex, newIndex);
          syncVisibleColumns();
        },
      });
    });
  };

  initColumnSettings();

  const displayColumns = computed(() => {
    const visible = showColumns.value
      .filter((item) => item.checked)
      .map(({ checked: _checked, ...col }) => col);
    if (!expandAllColumns.value) return visible;
    return visible.map((col) => {
      const next = { ...col };
      delete next.width;
      delete next.fixed;
      return next;
    });
  });

  const tableScroll = computed(() => {
    if (expandAllColumns.value) return undefined;
    const total = displayColumns.value.reduce(
      (sum, col) => sum + (Number(col.width) || 120),
      0
    );
    return { x: Math.max(total, 1200) };
  });

  const onModalVisibleChange = (visible: boolean) => {
    modalVisible.value = visible;
    if (!visible) {
      editingProvider.value = null;
    }
  };

  const handleEdit = (record: GameProviderRow) => {
    editingProvider.value = record.raw;
    modalVisible.value = true;
  };

  const handleSearch = () => {
    appliedSearch.providerCode = searchForm.providerCode;
    appliedSearch.name = searchForm.name;
    appliedSearch.status = searchForm.status;
    appliedSearch.isDeleted = searchForm.isDeleted;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    Object.assign(searchForm, {
      providerCode: '',
      name: '',
      status: undefined,
      isDeleted: 0,
    });
    Object.assign(appliedSearch, {
      providerCode: '',
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
    rows: GameProviderRow[],
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

    providers.value = rows.map((row, index) => {
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
        await updateGameProvider({
          id: row.id,
          name,
          logo_url: row.raw.logo_url || '',
          sort_no: sortNo,
          status: row.raw.status,
        });
      }, Promise.resolve());
      Message.success(t('manufacturerManage.message.reorderSuccess'));
    } catch {
      Message.error(t('manufacturerManage.message.reorderFail'));
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
        const list = [...providers.value];
        if (oldIndex >= list.length || newIndex >= list.length) {
          isRowDragging.value = false;
          return;
        }
        const [moved] = list.splice(oldIndex, 1);
        list.splice(newIndex, 0, moved);
        providers.value = list;
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
      () => providers.value.map((item) => item.id).join(','),
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
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 12px;
    margin-bottom: 14px;
  }
  .table-toolbar-left {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 12px;
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
  .column-setting-list {
    max-height: 360px;
    overflow-y: auto;
  }
  .column-setting-item {
    display: flex;
    align-items: center;
    width: 200px;
    padding: 4px 0;
  }
  .drag-handle {
    margin-right: 8px;
    color: var(--color-text-3);
    cursor: move;
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
