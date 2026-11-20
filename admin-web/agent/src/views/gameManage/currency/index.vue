<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.gameManage', 'menu.gameManage.currency']" />
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item :label="t('gameCurrencyManage.search.gameId')" field="gameId">
          <a-input-number
            v-model="searchForm.gameId"
            :placeholder="t('gameCurrencyManage.search.gameIdPlaceholder')"
            :min="1"
            allow-clear
            class="search-input"
          />
        </a-form-item>
        <a-form-item :label="t('gameCurrencyManage.search.status')" field="status">
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
        <a-form-item :label="t('gameCurrencyManage.search.isDeleted')" field="isDeleted">
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
            {{ t('gameCurrencyManage.toolbar.sync') }}
          </a-button>
          <a-button
            :type="isSyncProgressDone ? 'outline' : undefined"
            :status="isSyncProgressDone ? 'success' : undefined"
            :loading="syncProgressLoading"
            :disabled="isSyncProgressDone"
            :class="{ 'sync-progress-btn--done': isSyncProgressDone }"
            @click="fetchSyncProgress"
          >
            {{ t('gameCurrencyManage.toolbar.syncProgress', { progress: syncProgress }) }}
          </a-button>
          <a-button @click="() => fetchData()">
            <template #icon><icon-refresh /></template>
            {{ t('front.common.toolbar.refresh') }}
          </a-button>
          <a-button
            :type="expandAllColumns ? 'outline' : 'secondary'"
            @click="expandAllColumns = !expandAllColumns"
          >
            {{ expandAllColumns ? t('gameCurrencyManage.toolbar.collapseColumns') : t('gameCurrencyManage.toolbar.expandAll') }}
          </a-button>
        </a-space>
      </div>
      <a-table
        row-key="id"
        :loading="loading"
        :columns="displayColumns"
        :data="displayList"
        :pagination="pagination"
        :bordered="true"
        :stripe="true"
        :scroll="expandAllColumns ? undefined : { x: 1400 }"
        :class="{ 'table-expand-all': expandAllColumns }"
        @page-change="onPageChange"
        @page-size-change="onPageSizeChange"
      >
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
            {{ record.isDeleted ? t('gameCurrencyManage.search.deleted') : t('gameCurrencyManage.search.notDeleted') }}
          </a-tag>
        </template>
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <a-link @click="handleEdit(record)">{{ t('front.common.action.edit') }}</a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>
    <CurrencyModal
      :visible="modalVisible"
      :currency="editingCurrency"
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
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import dayjs from 'dayjs';
  import {
    queryGameCurrencyList,
    type GameCurrencyRecord,
    type GameCurrencyListParams,
  } from '@/api/gameCurrency';
  import { GAME_SYNC_SCOPE } from '@/api/gameSyncCheckpoint';
  import useLoading from '@/hooks/loading';
  import useSyncProgress from '@/hooks/sync-progress';
  import CurrencyModal from './components/currency-modal.vue';
  import SyncPreviewModal from './components/sync-preview-modal.vue';

  interface GameCurrencyRow {
    id: number;
    currencyCode: string;
    currencyName: string;
    gameCode: string;
    gameName: string;
    status: number;
    sourceStatus: number;
    isDeleted: number;
    createdAt: string;
    updatedAt: string;
    raw: GameCurrencyRecord;
  }

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(true);
  const modalVisible = ref(false);
  const editingCurrency = ref<GameCurrencyRecord | null>(null);
  /** 截图用：取消横向滚动与固定列，一屏展示全部列 */
  const expandAllColumns = ref(false);
  const syncModalVisible = ref(false);
  const {
    syncProgress,
    syncProgressLoading,
    isSyncProgressDone,
    fetchSyncProgress,
    scheduleFetchSyncProgress,
  } = useSyncProgress(GAME_SYNC_SCOPE.CURRENCY);

  const statusOptions = computed(() => [
    { label: t('front.common.status.enabled'), value: 1 },
    { label: t('front.common.status.disabled'), value: 2 },
  ]);
  const deletedOptions = computed(() => [
    { label: t('gameCurrencyManage.search.notDeleted'), value: 0 },
    { label: t('gameCurrencyManage.search.deleted'), value: 1 },
  ]);

  const searchForm = reactive<{
    gameId: number | undefined;
    status: number | undefined;
    isDeleted: number | undefined;
  }>({
    gameId: undefined,
    status: undefined,
    isDeleted: 0,
  });
  const appliedSearch = reactive({
    gameId: undefined as number | undefined,
    status: undefined as number | undefined,
    isDeleted: 0 as number | undefined,
  });

  const currencies = ref<GameCurrencyRow[]>([]);
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

  /** name_i18n 为多语言 JSON 字符串，优先展示 default 字段 */
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

  const mapCurrency = (item: GameCurrencyRecord): GameCurrencyRow => ({
    id: item.id,
    currencyCode: item.currency_code || '--',
    currencyName: parseI18nName(item.currency_name),
    gameCode: item.game_code || '--',
    gameName: parseI18nName(item.game_name),
    status: item.status,
    sourceStatus: item.source_status,
    isDeleted: item.is_deleted,
    createdAt: formatTime(item.created_at),
    updatedAt: formatTime(item.updated_at),
    raw: item,
  });

  const displayList = computed(() => currencies.value);

  const buildQueryParams = (
    page = pagination.current,
    pageSize = pagination.pageSize
  ): GameCurrencyListParams => {
    const params: GameCurrencyListParams = {
      page,
      page_size: pageSize,
      is_deleted: appliedSearch.isDeleted ?? 0,
    };
    if (appliedSearch.gameId !== undefined) params.game_id = appliedSearch.gameId;
    if (appliedSearch.status !== undefined) params.status = appliedSearch.status;
    return params;
  };

  const fetchData = async (
    page = pagination.current,
    pageSize = pagination.pageSize
  ) => {
    setLoading(true);
    try {
      const { data } = await queryGameCurrencyList(
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
      title: t('front.common.column.index'),
      dataIndex: 'index',
      slotName: 'index',
      width: 80,
      align: 'center',
      fixed: 'left',
    },
    {
      title: t('gameCurrencyManage.column.currencyCode'),
      dataIndex: 'currencyCode',
      width: 140,
      align: 'center',
    },
    {
      title: t('gameCurrencyManage.column.currencyName'),
      dataIndex: 'currencyName',
      width: 140,
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('gameCurrencyManage.column.gameCode'),
      dataIndex: 'gameCode',
      width: 140,
      align: 'center',
    },
    {
      title: t('gameCurrencyManage.column.gameName'),
      dataIndex: 'gameName',
      width: 160,
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('front.common.column.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 90,
      align: 'center',
    },
    {
      title: t('gameCurrencyManage.column.sourceStatus'),
      dataIndex: 'sourceStatus',
      slotName: 'sourceStatus',
      width: 100,
      align: 'center',
    },
    {
      title: t('gameCurrencyManage.column.isDeleted'),
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
      editingCurrency.value = null;
    }
  };

  const handleEdit = (record: GameCurrencyRow) => {
    editingCurrency.value = record.raw;
    modalVisible.value = true;
  };

  const handleSearch = () => {
    appliedSearch.gameId = searchForm.gameId;
    appliedSearch.status = searchForm.status;
    appliedSearch.isDeleted = searchForm.isDeleted;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    Object.assign(searchForm, {
      gameId: undefined,
      status: undefined,
      isDeleted: 0,
    });
    Object.assign(appliedSearch, {
      gameId: undefined,
      status: undefined,
      isDeleted: 0,
    });
    fetchData(1, pagination.pageSize);
  };

  const onPageChange = (current: number) => {
    fetchData(current, pagination.pageSize);
  };

  const onPageSizeChange = (pageSize: number) => {
    fetchData(1, pageSize);
  };

  const handleSync = () => {
    syncModalVisible.value = true;
  };

  const onSyncSuccess = () => {
    fetchData();
    scheduleFetchSyncProgress(1000);
  };

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
    // display: flex;
    // justify-content: flex-end;
    margin-bottom: 14px;
  }
  .search-form {
    row-gap: 16px;
    column-gap: 16px;
  }
</style>