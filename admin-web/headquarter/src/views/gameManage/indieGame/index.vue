<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.gameManage', 'menu.gameManage.indieGame']" />
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item :label="t('indieGameManage.search.gameCode')" field="gameCode">
          <a-input
            v-model="searchForm.gameCode"
            :placeholder="t('indieGameManage.search.gameCodePlaceholder')"
            allow-clear
            class="search-input"
          />
        </a-form-item>
        <a-form-item :label="t('indieGameManage.search.name')" field="name">
          <a-input
            v-model="searchForm.name"
            :placeholder="t('indieGameManage.search.namePlaceholder')"
            allow-clear
            class="search-input"
          />
        </a-form-item>
        <a-form-item :label="t('indieGameManage.search.category')" field="categoryId">
          <a-select
            v-model="searchForm.categoryId"
            :placeholder="t('front.common.placeholder.select')"
            allow-clear
            allow-search
            class="search-select"
          >
            <a-option
              v-for="item in categoryOptions"
              :key="item.value"
              :value="item.value"
            >
              {{ item.label }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item :label="t('indieGameManage.search.provider')" field="providerId">
          <a-select
            v-model="searchForm.providerId"
            :placeholder="t('front.common.placeholder.select')"
            allow-clear
            allow-search
            class="search-select"
          >
            <a-option
              v-for="item in providerOptions"
              :key="item.value"
              :value="item.value"
            >
              {{ item.label }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item :label="t('indieGameManage.search.channel')" field="channelId">
          <a-select
            v-model="searchForm.channelId"
            :placeholder="t('front.common.placeholder.select')"
            allow-clear
            allow-search
            class="search-select"
          >
            <a-option
              v-for="item in channelOptions"
              :key="item.value"
              :value="item.value"
            >
              {{ item.label }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item :label="t('indieGameManage.search.status')" field="status">
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
        <a-form-item :label="t('indieGameManage.search.isDeleted')" field="isDeleted">
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
              {{ t('indieGameManage.toolbar.sync') }}
            </a-button>
            <a-button
              :type="isSyncProgressDone ? 'outline' : undefined"
              :status="isSyncProgressDone ? 'success' : undefined"
              :loading="syncProgressLoading"
              :disabled="isSyncProgressDone"
              :class="{ 'sync-progress-btn--done': isSyncProgressDone }"
              @click="fetchSyncProgress"
            >
              {{ t('indieGameManage.toolbar.syncProgress', { progress: syncProgress }) }}
            </a-button>
            <a-button @click="() => fetchData()">
              <template #icon><icon-refresh /></template>
              {{ t('front.common.toolbar.refresh') }}
            </a-button>
            <a-button
              :type="expandAllColumns ? 'outline' : 'secondary'"
              @click="expandAllColumns = !expandAllColumns"
            >
              {{ expandAllColumns ? t('indieGameManage.toolbar.collapseColumns') : t('indieGameManage.toolbar.expandAll') }}
            </a-button>
          </a-space>
          <a-alert
            v-if="isFiltering"
            type="warning"
            :show-icon="false"
            class="drag-tip"
          >
            {{ t('indieGameManage.drag.tipFilter') }}
          </a-alert>
          <a-alert v-else type="info" :show-icon="false" class="drag-tip">
            {{ t('indieGameManage.drag.tipOk') }}
          </a-alert>
        </div>
        <a-popover
          trigger="click"
          position="br"
          @popup-visible-change="popupVisibleChange"
        >
          <a-button>
            <template #icon><icon-settings /></template>
            {{ t('indieGameManage.toolbar.columnSettings') }}
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
        <template #image="{ record }">
          <a-image
            v-if="record.imageUrl"
            :src="record.imageUrl"
            width="40"
            height="40"
            fit="contain"
            show-loader
          />
          <span v-else>--</span>
        </template>
        <template #sourceImage="{ record }">
          <a-image
            v-if="record.sourceImageUrl"
            :src="record.sourceImageUrl"
            width="40"
            height="40"
            fit="contain"
            show-loader
          />
          <span v-else>--</span>
        </template>
        <template #currencies="{ record }">
          <a-space v-if="record.currencyNames.length" wrap :size="4">
            <a-tag v-for="name in record.currencyNames" :key="name">
              {{ name }}
            </a-tag>
          </a-space>
          <span v-else>--</span>
        </template>
        <template #supportsEmbed="{ record }">
          <a-tag :color="record.supportsEmbed ? 'green' : 'gray'">
            {{ record.supportsEmbed ? t('indieGameManage.status.supported') : t('indieGameManage.status.unsupported') }}
          </a-tag>
        </template>
        <template #supportsRedirect="{ record }">
          <a-tag :color="record.supportsRedirect ? 'green' : 'gray'">
            {{ record.supportsRedirect ? t('indieGameManage.status.supported') : t('indieGameManage.status.unsupported') }}
          </a-tag>
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
            {{ record.isDeleted ? t('indieGameManage.search.deleted') : t('indieGameManage.search.notDeleted') }}
          </a-tag>
        </template>
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <a-link @click="handleEdit(record)">{{ t('front.common.action.edit') }}</a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>
    <GameModal
      :visible="modalVisible"
      :game="editingGame"
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
    onMounted,
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
    queryGameList,
    updateGame,
    type GameRecord,
    type GameListParams,
  } from '@/api/game';
  import { queryGameCategoryList } from '@/api/gameCategory';
  import { queryGameProviderList } from '@/api/gameProvider';
  import { queryGameChannelList } from '@/api/gameChannel';
  import { GAME_SYNC_SCOPE } from '@/api/gameSyncCheckpoint';
  import useLoading from '@/hooks/loading';
  import useSyncProgress from '@/hooks/sync-progress';
  import GameModal from './components/game-modal.vue';
  import SyncPreviewModal from './components/sync-preview-modal.vue';

  type ColumnSettingItem = TableColumnData & { checked?: boolean };

  interface GameRow {
    id: number;
    gameCode: string;
    name: string;
    imageUrl: string;
    categoryId: number;
    categoryName: string;
    providerId: number;
    providerKey: string;
    providerName: string;
    channelId: number;
    channelName: string;
    currencyNames: string[];
    sortNo: number;
    supportsEmbed: boolean;
    supportsRedirect: boolean;
    status: number;
    sourceId: number;
    sourceGameCode: string;
    sourceProviderKey: string;
    sourceName: string;
    sourceImageUrl: string;
    sourceSortNo: number;
    sourceStatus: number;
    isDeleted: number;
    createdAt: string;
    updatedAt: string;
    raw: GameRecord;
  }

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(true);
  const modalVisible = ref(false);
  const editingGame = ref<GameRecord | null>(null);
  /** 截图用：取消横向滚动与固定列，一屏展示全部列 */
  const expandAllColumns = ref(false);
  const syncModalVisible = ref(false);
  const {
    syncProgress,
    syncProgressLoading,
    isSyncProgressDone,
    fetchSyncProgress,
    scheduleFetchSyncProgress,
  } = useSyncProgress(GAME_SYNC_SCOPE.GAME);
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
    { label: t('indieGameManage.search.notDeleted'), value: 0 },
    { label: t('indieGameManage.search.deleted'), value: 1 },
  ]);

  const categoryOptions = ref<{ label: string; value: number }[]>([]);
  const providerOptions = ref<{ label: string; value: number }[]>([]);
  const channelOptions = ref<{ label: string; value: number }[]>([]);

  const searchForm = reactive<{
    gameCode: string;
    name: string;
    categoryId: number | undefined;
    providerId: number | undefined;
    channelId: number | undefined;
    status: number | undefined;
    isDeleted: number | undefined;
  }>({
    gameCode: '',
    name: '',
    categoryId: undefined,
    providerId: undefined,
    channelId: undefined,
    status: undefined,
    isDeleted: 0,
  });
  const appliedSearch = reactive({
    gameCode: '',
    name: '',
    categoryId: undefined as number | undefined,
    providerId: undefined as number | undefined,
    channelId: undefined as number | undefined,
    status: undefined as number | undefined,
    isDeleted: 0 as number | undefined,
  });
  /** 存在筛选条件时，当前页并非连续的排序区间，禁止拖拽排序 */
  const isFiltering = computed(
    () =>
      Boolean(appliedSearch.gameCode) ||
      Boolean(appliedSearch.name) ||
      appliedSearch.categoryId !== undefined ||
      appliedSearch.providerId !== undefined ||
      appliedSearch.channelId !== undefined ||
      appliedSearch.status !== undefined
  );

  const games = ref<GameRow[]>([]);
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

  const mapGame = (item: GameRecord): GameRow => ({
    id: item.id,
    gameCode: item.game_code || '--',
    name: item.name || '--',
    imageUrl: item.image_url || '',
    categoryId: item.category_id,
    categoryName: parseI18nName(item.category_name_i18n),
    providerId: item.provider_id,
    providerKey: item.provider_key || '--',
    providerName: parseI18nName(item.provider_name_i18n),
    channelId: item.channel_id,
    channelName: parseI18nName(item.channel_name_i18n),
    currencyNames: (item.game_currency_info || []).map(
      (c) => c.currency_name || '--'
    ),
    sortNo: item.sort_no ?? 0,
    supportsEmbed: Boolean(item.supports_embed),
    supportsRedirect: Boolean(item.supports_redirect),
    status: item.status,
    sourceId: item.source_id,
    sourceGameCode: item.source_game_code || '--',
    sourceProviderKey: item.source_provider_key || '--',
    sourceName: parseI18nName(item.source_name_i18n),
    sourceImageUrl: item.source_image_url || '',
    sourceSortNo: item.source_sort_no ?? 0,
    sourceStatus: item.source_status,
    isDeleted: item.is_deleted,
    createdAt: formatTime(item.created_at),
    updatedAt: formatTime(item.updated_at),
    raw: item,
  });

  const displayList = computed(() => games.value);

  const buildQueryParams = (
    page = pagination.current,
    pageSize = pagination.pageSize
  ): GameListParams => {
    const params: GameListParams = {
      page,
      page_size: pageSize,
      is_deleted: appliedSearch.isDeleted ?? 0,
    };
    const gameCode = appliedSearch.gameCode.trim();
    const name = appliedSearch.name.trim();
    if (gameCode) params.game_code = gameCode;
    if (name) params.name = name;
    if (appliedSearch.categoryId !== undefined)
      params.category_id = appliedSearch.categoryId;
    if (appliedSearch.providerId !== undefined)
      params.provider_id = appliedSearch.providerId;
    if (appliedSearch.channelId !== undefined)
      params.channel_id = appliedSearch.channelId;
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
      const { data } = await queryGameList(buildQueryParams(page, pageSize));
      games.value = (data?.list || []).map(mapGame);
      pagination.current = page;
      pagination.pageSize = pageSize;
      pagination.total = data?.total ?? 0;
    } catch {
      games.value = [];
      pagination.total = 0;
    } finally {
      setLoading(false);
    }
  };

  const fetchFilterOptions = async () => {
    try {
      const [categoryRes, providerRes, channelRes] = await Promise.all([
        queryGameCategoryList({ page: 1, page_size: 100, is_deleted: 0 }),
        queryGameProviderList({ page: 1, page_size: 100, is_deleted: 0 }),
        queryGameChannelList({ page: 1, page_size: 100, is_deleted: 0 }),
      ]);
      categoryOptions.value = (categoryRes.data?.list || []).map((item) => ({
        label: `${parseI18nName(item.name)}（${item.category_code}）`,
        value: item.source_id,
      }));
      providerOptions.value = (providerRes.data?.list || []).map((item) => ({
        label: `${parseI18nName(item.name)}（${item.provider_code}）`,
        value: item.source_id,
      }));
      channelOptions.value = (channelRes.data?.list || []).map((item) => ({
        label: `${parseI18nName(item.name)}（${item.channel_code}）`,
        value: item.source_id,
      }));
    } catch {
      categoryOptions.value = [];
      providerOptions.value = [];
      channelOptions.value = [];
    }
  };

  const columns = computed<TableColumnData[]>(() => [
    {
      title: t('indieGameManage.column.drag'),
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
      title: t('indieGameManage.column.gameCode'),
      dataIndex: 'gameCode',
      width: 140,
      align: 'center',
    },
    {
      title: t('indieGameManage.column.image'),
      dataIndex: 'imageUrl',
      slotName: 'image',
      width: 90,
      align: 'center',
    },
    {
      title: t('indieGameManage.column.name'),
      dataIndex: 'name',
      width: 150,
      align: 'center',
    },
    {
      title: t('indieGameManage.column.category'),
      dataIndex: 'categoryName',
      width: 130,
      align: 'center',
    },
    {
      title: t('indieGameManage.column.provider'),
      dataIndex: 'providerName',
      width: 130,
      align: 'center',
    },
    {
      title: t('indieGameManage.column.providerKey'),
      dataIndex: 'providerKey',
      width: 130,
      align: 'center',
    },
    {
      title: t('indieGameManage.column.channel'),
      dataIndex: 'channelName',
      width: 130,
      align: 'center',
    },
    {
      title: t('indieGameManage.column.currencies'),
      dataIndex: 'currencyNames',
      slotName: 'currencies',
      width: 180,
      align: 'center',
    },
    {
      title: t('indieGameManage.column.sortNo'),
      dataIndex: 'sortNo',
      width: 90,
      align: 'center',
    },
    {
      title: t('indieGameManage.column.supportsEmbed'),
      dataIndex: 'supportsEmbed',
      slotName: 'supportsEmbed',
      width: 100,
      align: 'center',
    },
    {
      title: t('indieGameManage.column.supportsRedirect'),
      dataIndex: 'supportsRedirect',
      slotName: 'supportsRedirect',
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
    // {
    //   title: '来源游戏编码',
    //   dataIndex: 'sourceGameCode',
    //   width: 150,
    //   align: 'center',
    // },
    // {
    //   title: '来源厂商Key',
    //   dataIndex: 'sourceProviderKey',
    //   width: 140,
    //   align: 'center',
    // },
    // {
    //   title: '来源游戏名称',
    //   dataIndex: 'sourceName',
    //   width: 150,
    //   align: 'center',
    // },
    // {
    //   title: '来源图片',
    //   dataIndex: 'sourceImageUrl',
    //   slotName: 'sourceImage',
    //   width: 100,
    //   align: 'center',
    // },
    // {
    //   title: '来源排序',
    //   dataIndex: 'sourceSortNo',
    //   width: 100,
    //   align: 'center',
    // },
    // {
    //   title: '来源状态',
    //   dataIndex: 'sourceStatus',
    //   slotName: 'sourceStatus',
    //   width: 100,
    //   align: 'center',
    // },
    {
      title: t('indieGameManage.column.isDeleted'),
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

  /** 默认隐藏的来源相关列，可通过列设置打开 */
  const defaultHiddenColumns = new Set([
    'sourceGameCode',
    'sourceProviderKey',
    'sourceName',
    'sourceImageUrl',
    'sourceSortNo',
    'sourceStatus',
  ]);

  const showColumns = ref<ColumnSettingItem[]>([]);
  const columnSettingRef = ref<HTMLElement | null>(null);
  let columnSortable: Sortable | null = null;

  const initColumnSettings = () => {
    showColumns.value = cloneDeep(columns.value).map((item) => ({
      ...item,
      checked: !defaultHiddenColumns.has(String(item.dataIndex)),
    }));
  };

  const syncVisibleColumns = () => {
    // checked 变更后触发 displayColumns 重算
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
      editingGame.value = null;
    }
  };

  const handleEdit = (record: GameRow) => {
    editingGame.value = record.raw;
    modalVisible.value = true;
  };

  const handleSearch = () => {
    appliedSearch.gameCode = searchForm.gameCode;
    appliedSearch.name = searchForm.name;
    appliedSearch.categoryId = searchForm.categoryId;
    appliedSearch.providerId = searchForm.providerId;
    appliedSearch.channelId = searchForm.channelId;
    appliedSearch.status = searchForm.status;
    appliedSearch.isDeleted = searchForm.isDeleted;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    Object.assign(searchForm, {
      gameCode: '',
      name: '',
      categoryId: undefined,
      providerId: undefined,
      channelId: undefined,
      status: undefined,
      isDeleted: 0,
    });
    Object.assign(appliedSearch, {
      gameCode: '',
      name: '',
      categoryId: undefined,
      providerId: undefined,
      channelId: undefined,
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

  /**
   * 仅重排拖拽起点到终点这段区间内的 sort_no，区间外的行不受影响、不调接口。
   * 区间内原 sort_no 若已是互不相同的正数，直接复用这些值按新顺序重新分配；
   * 否则（存在重复/非正数）才按页内位置从 1 起连续编号兜底。
   * 落库期间锁定页码，避免误点分页或表格内部改页导致跳到其他页。
   */
  const persistDragOrder = async (
    rows: GameRow[],
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

    games.value = rows.map((row, index) => {
      if (index < from || index > to) return row;
      return { ...row, sortNo: targets[index - from] };
    });

    isDragPersisting.value = true;
    setLoading(true);
    try {
      await changed.reduce(async (prevPromise, { row, sortNo }) => {
        await prevPromise;
        await updateGame({
          id: row.id,
          name: row.raw.name || '',
          image_url: row.raw.image_url || '',
          provider_key: row.raw.provider_key || '',
          sort_no: sortNo,
          supports_embed: Boolean(row.raw.supports_embed),
          supports_redirect: Boolean(row.raw.supports_redirect),
          status: row.raw.status,
        });
      }, Promise.resolve());
      Message.success(t('indieGameManage.message.reorderSuccess'));
    } catch {
      Message.error(t('indieGameManage.message.reorderFail'));
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
        const list = [...games.value];
        if (oldIndex >= list.length || newIndex >= list.length) {
          isRowDragging.value = false;
          return;
        }
        const [moved] = list.splice(oldIndex, 1);
        list.splice(newIndex, 0, moved);
        games.value = list;
        persistDragOrder(list, oldIndex, newIndex).finally(() => {
          isRowDragging.value = false;
        });
      },
    });
  };

  watch(
    [loading, isFiltering, () => games.value.map((item) => item.id).join(',')],
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

  const handleSync = () => {
    syncModalVisible.value = true;
  };

  const onSyncSuccess = () => {
    fetchData();
    scheduleFetchSyncProgress(1000);
  };

  onMounted(() => {
    fetchFilterOptions();
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
  .search-form {
    row-gap: 16px;
    column-gap: 16px;
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
  .search-select {
    width: 180px;
  }
  .search-select :deep(.arco-select-view) {
    min-width: 180px;
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
</style>
