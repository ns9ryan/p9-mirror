<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.gameManage', 'menu.gameManage.syncList']" />
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item :label="t('syncListManage.search.syncScope')" field="syncScope">
          <a-input
            v-model="searchForm.syncScope"
            :placeholder="t('syncListManage.search.syncScopePlaceholder')"
            allow-clear
            class="search-input"
          />
        </a-form-item>
        <a-form-item :label="t('syncListManage.search.successRange')" field="successRange">
          <a-range-picker
            v-model="searchForm.successRange"
            show-time
            format="YYYY-MM-DD HH:mm:ss"
            class="search-range"
          />
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
          <a-button @click="() => fetchData()">
            <template #icon><icon-refresh /></template>
            {{ t('front.common.toolbar.refresh') }}
          </a-button>
          <a-button
            :type="expandAllColumns ? 'outline' : 'secondary'"
            @click="expandAllColumns = !expandAllColumns"
          >
            {{ expandAllColumns ? t('syncListManage.toolbar.collapseColumns') : t('syncListManage.toolbar.expandAll') }}
          </a-button>
        </a-space>
        <a-popover
          trigger="click"
          position="br"
          @popup-visible-change="popupVisibleChange"
        >
          <a-button>
            <template #icon><icon-settings /></template>
            {{ t('syncListManage.toolbar.columnSettings') }}
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
                  :disabled="item.dataIndex === 'operations'"
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
        row-key="id"
        :loading="loading"
        :columns="displayColumns"
        :data="displayList"
        :pagination="pagination"
        :bordered="true"
        :stripe="true"
        :scroll="tableScroll"
        :class="{ 'table-expand-all': expandAllColumns }"
        @page-change="onPageChange"
        @page-size-change="onPageSizeChange"
      >
        <template #index="{ rowIndex }">
          {{ rowIndex + 1 + (pagination.current - 1) * pagination.pageSize }}
        </template>
        <template #syncStatus="{ record }">
          <a-tag :color="syncStatusMeta(record.syncStatus).color">
            {{ syncStatusMeta(record.syncStatus).label }}
          </a-tag>
        </template>
        <template #lastErrorMessage="{ record }">
          <a-tooltip
            v-if="record.lastErrorMessage && record.lastErrorMessage !== '--'"
            :content="record.lastErrorMessage"
          >
            <span class="error-text">{{ record.lastErrorMessage }}</span>
          </a-tooltip>
          <span v-else>--</span>
        </template>
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <a-link @click="handleViewDetail(record)">{{ t('syncListManage.action.detail') }}</a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>
    <CheckpointDetailModal
      v-model:visible="detailModalVisible"
      :sync-scope="detailSyncScope"
    />
  </div>
</template>

<script lang="ts" setup>
  import { computed, nextTick, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import cloneDeep from 'lodash/cloneDeep';
  import Sortable from 'sortablejs';
  import dayjs from 'dayjs';
  import {
    queryGameSyncCheckpointList,
    type GameSyncCheckpointRecord,
    type GameSyncCheckpointListParams,
  } from '@/api/gameSyncCheckpoint';
  import useLoading from '@/hooks/loading';
  import CheckpointDetailModal from './components/checkpoint-detail-modal.vue';

  type ColumnSettingItem = TableColumnData & { checked?: boolean };

  interface GameSyncCheckpointRow {
    id: number;
    syncScope: string;
    checkpointValue: string;
    localTotal: number;
    remoteTotal: number;
    createdCount: number;
    updatedCount: number;
    deletedCount: number;
    failedCount: number;
    syncStatus: number;
    lastErrorMessage: string;
    lastSyncAt: string;
    lastSuccessAt: string;
    createdAt: string;
    updatedAt: string;
    raw: GameSyncCheckpointRecord;
  }

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(true);
  /** 截图用：取消横向滚动与固定列，一屏展示全部列 */
  const expandAllColumns = ref(false);
  const detailModalVisible = ref(false);
  const detailSyncScope = ref<string | null>(null);

  /** 同步状态：后端未提供明确枚举，按常见语义展示 0 待同步 / 1 成功 / 2 失败 */
  const syncStatusMeta = (status: number) => {
    switch (status) {
      case 1:
        return { label: t('syncListManage.status.success'), color: 'green' };
      case 2:
        return { label: t('syncListManage.status.failed'), color: 'red' };
      default:
        return { label: t('syncListManage.status.pending'), color: 'gray' };
    }
  };

  const searchForm = reactive<{
    syncScope: string;
    successRange: string[];
  }>({
    syncScope: '',
    successRange: [],
  });
  const appliedSearch = reactive({
    syncScope: '',
    startTime: undefined as number | undefined,
    endTime: undefined as number | undefined,
  });

  const checkpoints = ref<GameSyncCheckpointRow[]>([]);
  const pagination = reactive({
    current: 1,
    pageSize: 15,
    total: 0,
    showTotal: true,
    showPageSize: true,
    pageSizeOptions: [15, 20, 30, 50, 100],
    pageSizeProps: {
      popupContainer: 'body',
    },
  });

  const formatTime = (value?: number) => {
    if (!value) return '--';
    const ms = value < 1e12 ? value * 1000 : value;
    return dayjs(ms).format('YYYY-MM-DD HH:mm:ss');
  };

  const mapCheckpoint = (
    item: GameSyncCheckpointRecord
  ): GameSyncCheckpointRow => ({
    id: item.id,
    syncScope: item.sync_scope || '--',
    checkpointValue: item.checkpoint_value || '--',
    localTotal: item.local_total ?? 0,
    remoteTotal: item.remote_total ?? 0,
    createdCount: item.created_count ?? 0,
    updatedCount: item.updated_count ?? 0,
    deletedCount: item.deleted_count ?? 0,
    failedCount: item.failed_count ?? 0,
    syncStatus: item.sync_status,
    lastErrorMessage: item.last_error_message || '--',
    lastSyncAt: formatTime(item.last_sync_at),
    lastSuccessAt: formatTime(item.last_success_at),
    createdAt: formatTime(item.created_at),
    updatedAt: formatTime(item.updated_at),
    raw: item,
  });

  const displayList = computed(() => checkpoints.value);

  const buildQueryParams = (
    page = pagination.current,
    pageSize = pagination.pageSize
  ): GameSyncCheckpointListParams => {
    const params: GameSyncCheckpointListParams = {
      page,
      page_size: pageSize,
    };
    const syncScope = appliedSearch.syncScope.trim();
    if (syncScope) params.sync_scope = syncScope;
    if (appliedSearch.startTime !== undefined)
      params.start_time = appliedSearch.startTime;
    if (appliedSearch.endTime !== undefined)
      params.end_time = appliedSearch.endTime;
    return params;
  };

  const fetchData = async (
    page = pagination.current,
    pageSize = pagination.pageSize
  ) => {
    setLoading(true);
    try {
      const { data } = await queryGameSyncCheckpointList(
        buildQueryParams(page, pageSize)
      );
      checkpoints.value = (data?.list || []).map(mapCheckpoint);
      pagination.current = page;
      pagination.pageSize = pageSize;
      pagination.total = data?.total ?? 0;
    } catch {
      checkpoints.value = [];
      pagination.total = 0;
    } finally {
      setLoading(false);
    }
  };

  const columns = computed<TableColumnData[]>(() => [
    {
      title: t('front.common.column.index'),
      dataIndex: 'index',
      slotName: 'index',
      width: 80,
      align: 'center',
      fixed: 'left',
    },
    {
      title: t('syncListManage.column.syncScope'),
      dataIndex: 'syncScope',
      width: 140,
      align: 'center',
    },
    {
      title: t('syncListManage.column.checkpointValue'),
      dataIndex: 'checkpointValue',
      width: 160,
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('syncListManage.column.localTotal'),
      dataIndex: 'localTotal',
      width: 100,
      align: 'center',
    },
    {
      title: t('syncListManage.column.remoteTotal'),
      dataIndex: 'remoteTotal',
      width: 100,
      align: 'center',
    },
    {
      title: t('syncListManage.column.createdCount'),
      dataIndex: 'createdCount',
      width: 90,
      align: 'center',
    },
    {
      title: t('syncListManage.column.updatedCount'),
      dataIndex: 'updatedCount',
      width: 90,
      align: 'center',
    },
    {
      title: t('syncListManage.column.deletedCount'),
      dataIndex: 'deletedCount',
      width: 90,
      align: 'center',
    },
    {
      title: t('syncListManage.column.failedCount'),
      dataIndex: 'failedCount',
      width: 90,
      align: 'center',
    },
    {
      title: t('syncListManage.column.syncStatus'),
      dataIndex: 'syncStatus',
      slotName: 'syncStatus',
      width: 100,
      align: 'center',
    },
    {
      title: t('syncListManage.column.lastError'),
      dataIndex: 'lastErrorMessage',
      slotName: 'lastErrorMessage',
      width: 180,
      align: 'center',
    },
    {
      title: t('syncListManage.column.lastSyncAt'),
      dataIndex: 'lastSyncAt',
      width: 170,
      align: 'center',
    },
    {
      title: t('syncListManage.column.lastSuccessAt'),
      dataIndex: 'lastSuccessAt',
      width: 170,
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
      width: 90,
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

  const handleViewDetail = (record: GameSyncCheckpointRow) => {
    detailSyncScope.value = record.syncScope;
    detailModalVisible.value = true;
  };

  const handleSearch = () => {
    appliedSearch.syncScope = searchForm.syncScope;
    const [start, end] = searchForm.successRange || [];
    appliedSearch.startTime = start
      ? Math.floor(dayjs(start).valueOf() / 1000)
      : undefined;
    appliedSearch.endTime = end
      ? Math.floor(dayjs(end).valueOf() / 1000)
      : undefined;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    Object.assign(searchForm, {
      syncScope: '',
      successRange: [],
    });
    Object.assign(appliedSearch, {
      syncScope: '',
      startTime: undefined,
      endTime: undefined,
    });
    fetchData(1, pagination.pageSize);
  };

  const onPageChange = (current: number) => {
    fetchData(current, pagination.pageSize);
  };

  const onPageSizeChange = (pageSize: number) => {
    fetchData(1, pageSize);
  };

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
  .search-input {
    width: 170px;
  }
  .search-range {
    width: 340px;
  }
  .table-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 14px;
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
  .error-text {
    display: inline-block;
    max-width: 160px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    vertical-align: middle;
  }
  .search-form {
    row-gap: 16px;
    column-gap: 16px;
  }
</style>
