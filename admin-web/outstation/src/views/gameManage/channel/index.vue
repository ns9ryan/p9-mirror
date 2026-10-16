<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.gameManage', 'menu.gameManage.channel']" />
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item :label="t('channelManage.search.channelCode')" field="channelCode">
          <a-input
            v-model="searchForm.channelCode"
            :placeholder="t('channelManage.search.channelCodePlaceholder')"
            allow-clear
            class="search-input"
          />
        </a-form-item>
        <a-form-item :label="t('channelManage.search.name')" field="name">
          <a-input
            v-model="searchForm.name"
            :placeholder="t('channelManage.search.namePlaceholder')"
            allow-clear
            class="search-input"
          />
        </a-form-item>
        <a-form-item :label="t('channelManage.search.status')" field="status">
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
          <!-- <a-button type="primary" @click="handleSync">
            <template #icon><icon-sync /></template>
            {{ t('channelManage.toolbar.sync') }}
          </a-button> -->
          <!-- <a-button
            :type="isSyncProgressDone ? 'outline' : undefined"
            :status="isSyncProgressDone ? 'success' : undefined"
            :loading="syncProgressLoading"
            :disabled="isSyncProgressDone"
            :class="{ 'sync-progress-btn--done': isSyncProgressDone }"
            @click="fetchSyncProgress"
          >
            {{ t('channelManage.toolbar.syncProgress', { progress: syncProgress }) }}
          </a-button> -->
          <a-button @click="() => fetchData()">
            <template #icon><icon-refresh /></template>
            {{ t('front.common.toolbar.refresh') }}
          </a-button>
          <a-button
            :type="expandAllColumns ? 'outline' : 'secondary'"
            @click="expandAllColumns = !expandAllColumns"
          >
            {{ expandAllColumns ? t('channelManage.toolbar.collapseColumns') : t('channelManage.toolbar.expandAll') }}
          </a-button>
        </a-space>
        <a-alert
          v-if="isFiltering"
          type="warning"
          :show-icon="false"
          class="drag-tip"
        >
          {{ t('channelManage.drag.tipFilter') }}
        </a-alert>
        <a-alert v-else type="info" :show-icon="false" class="drag-tip">
          {{ t('channelManage.drag.tipOk') }}
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
        :scroll="expandAllColumns ? undefined : { x: 1500 }"
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
            {{
              record.status === 1
                ? t('channelManage.status.enabled')
                : t('channelManage.status.disabled')
            }}
          </a-tag>
        </template>
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <a-link @click="handleEdit(record)">{{ t('front.common.action.edit') }}</a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>

    <ChannelModal
      :visible="modalVisible"
      :channel="editingChannel"
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
    queryGameChannelList,
    updateGameChannel,
    type GameChannelRecord,
    type GameChannelListParams,
  } from '@/api/gameChannel';
  import { GAME_SYNC_SCOPE } from '@/api/gameSyncCheckpoint';
  import useLoading from '@/hooks/loading';
  import useSyncProgress from '@/hooks/sync-progress';
  import ChannelModal from './components/channel-modal.vue';
  import SyncPreviewModal from './components/sync-preview-modal.vue';

  interface GameChannelRow {
    id: number;
    channelCode: string;
    providerCount: number;
    gameCount: number;
    name: string;
    sortNo: number;
    loadType: number;
    status: number;
    createdAt: string;
    updatedAt: string;
    raw: GameChannelRecord;
  }

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(true);
  const expandAllColumns = ref(false);
  const modalVisible = ref(false);
  const editingChannel = ref<GameChannelRecord | null>(null);
  const syncModalVisible = ref(false);
  const {
    syncProgress,
    syncProgressLoading,
    isSyncProgressDone,
    fetchSyncProgress,
    scheduleFetchSyncProgress,
  } = useSyncProgress(GAME_SYNC_SCOPE.CHANNEL);
  const tableRef = ref();
  let rowSortable: Sortable | null = null;
  /** 拖拽落库中，禁止分页跳转 */
  const isDragPersisting = ref(false);
  /** 行拖拽进行中（防止松手误点到底部分页） */
  const isRowDragging = ref(false);

  const statusOptions = computed(() => [
    { label: t('channelManage.status.enabled'), value: 1 },
    { label: t('channelManage.status.disabled'), value: 2 },
  ]);

  const searchForm = reactive<{
    channelCode: string;
    name: string;
    status: number | undefined;
  }>({
    channelCode: '',
    name: '',
    status: undefined,
  });

  const appliedSearch = reactive({
    channelCode: '',
    name: '',
    status: undefined as number | undefined,
  });
  /** 存在筛选条件时，当前页并非连续的排序区间，禁止拖拽排序 */
  const isFiltering = computed(
    () =>
      Boolean(appliedSearch.channelCode) ||
      Boolean(appliedSearch.name) ||
      appliedSearch.status !== undefined
  );

  const channels = ref<GameChannelRow[]>([]);
  const pagination = reactive({
    current: 1,
    pageSize: 15,
    total: 0,
    showTotal: true,
    showPageSize: true,
    pageSizeOptions: [10, 15, 20, 30, 50, 100],
    pageSizeProps: {
      popupContainer: 'body',
    },
  });

  const formatTime = (value?: number) => {
    if (!value) return '--';
    const ms = value < 1e12 ? value * 1000 : value;
    return dayjs(ms).format('YYYY-MM-DD HH:mm:ss');
  };

  const mapChannel = (item: GameChannelRecord): GameChannelRow => ({
    id: item.id,
    channelCode: item.channel_code || '--',
    providerCount: item.provider_count ?? 0,
    gameCount: item.game_count ?? 0,
    name: item.name || '--',
    sortNo: item.sort_no ?? 0,
    loadType: item.load_type ?? 0,
    status: item.status,
    createdAt: formatTime(item.created_at),
    updatedAt: formatTime(item.updated_at),
    raw: item,
  });

  const displayList = computed(() => channels.value);

  const buildQueryParams = (
    page = pagination.current,
    pageSize = pagination.pageSize
  ): GameChannelListParams => {
    const params: GameChannelListParams = {
      page,
      page_size: pageSize,
    };

    const channelCode = appliedSearch.channelCode.trim();
    const name = appliedSearch.name.trim();

    if (channelCode) params.channel_code = channelCode;
    if (name) params.name = name;
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
      const { data } = await queryGameChannelList(
        buildQueryParams(page, pageSize)
      );
      channels.value = (data?.list || []).map(mapChannel);
      pagination.current = page;
      pagination.pageSize = pageSize;
      pagination.total = data?.total ?? 0;
    } catch {
      channels.value = [];
      pagination.total = 0;
    } finally {
      setLoading(false);
    }
  };

  const columns = computed<TableColumnData[]>(() => [
    {
      title: t('channelManage.column.drag'),
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
      title: t('channelManage.column.channelCode'),
      dataIndex: 'channelCode',
      width: 140,
      align: 'center',
    },
    {
      title: t('channelManage.column.name'),
      dataIndex: 'name',
      width: 150,
      align: 'center',
    },
    {
      title: t('channelManage.column.providerCount'),
      dataIndex: 'providerCount',
      width: 110,
      align: 'center',
    },
    {
      title: t('channelManage.column.gameCount'),
      dataIndex: 'gameCount',
      width: 110,
      align: 'center',
    },
    {
      title: t('channelManage.column.sortNo'),
      dataIndex: 'sortNo',
      width: 90,
      align: 'center',
    },
    {
      title: t('channelManage.column.loadType'),
      dataIndex: 'loadType',
      width: 100,
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

  const handleSearch = () => {
    appliedSearch.channelCode = searchForm.channelCode;
    appliedSearch.name = searchForm.name;
    appliedSearch.status = searchForm.status;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    Object.assign(searchForm, {
      channelCode: '',
      name: '',
      status: undefined,
    });
    Object.assign(appliedSearch, {
      channelCode: '',
      name: '',
      status: undefined,
    });
    fetchData(1, pagination.pageSize);
  };

  const onModalVisibleChange = (visible: boolean) => {
    modalVisible.value = visible;
    if (!visible) {
      editingChannel.value = null;
    }
  };

  const handleEdit = (record: GameChannelRow) => {
    editingChannel.value = record.raw;
    modalVisible.value = true;
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
    rows: GameChannelRow[],
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

    channels.value = rows.map((row, index) => {
      if (index < from || index > to) return row;
      return { ...row, sortNo: targets[index - from] };
    });

    isDragPersisting.value = true;
    setLoading(true);
    try {
      await changed.reduce(async (prevPromise, { row, sortNo }) => {
        await prevPromise;
        await updateGameChannel({
          id: row.id,
          sort_no: sortNo,
        });
      }, Promise.resolve());
      Message.success(t('channelManage.message.reorderSuccess'));
    } catch {
      Message.error(t('channelManage.message.reorderFail'));
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
        const list = [...channels.value];
        if (oldIndex >= list.length || newIndex >= list.length) {
          isRowDragging.value = false;
          return;
        }
        const [moved] = list.splice(oldIndex, 1);
        list.splice(newIndex, 0, moved);
        channels.value = list;
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
      () => channels.value.map((item) => item.id).join(','),
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
  // fetchSyncProgress();
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
