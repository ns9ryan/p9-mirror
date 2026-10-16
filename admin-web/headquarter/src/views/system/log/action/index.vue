<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.route.log', 'menu.route.actionLog']" />
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item :label="$t('log.search.username')" field="username">
          <a-input
            v-model="searchForm.username"
            :placeholder="$t('log.search.usernamePlaceholder')"
            allow-clear
            class="search-input"
          />
        </a-form-item>
        <a-form-item :label="$t('log.search.method')" field="requestMethod">
          <a-select
            v-model="searchForm.requestMethod"
            :placeholder="$t('log.search.methodPlaceholder')"
            allow-clear
            class="search-select"
          >
            <a-option
              v-for="method in methodOptions"
              :key="method"
              :value="method"
            >
              {{ method }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item :label="$t('log.search.path')" field="requestPath">
          <a-input
            v-model="searchForm.requestPath"
            :placeholder="$t('log.search.pathPlaceholder')"
            allow-clear
            class="search-input"
          />
        </a-form-item>
        <a-form-item :label="$t('log.search.result')" field="actionResult">
          <a-select
            v-model="searchForm.actionResult"
            :placeholder="$t('log.search.resultPlaceholder')"
            allow-clear
            class="search-select"
          >
            <a-option :value="1">{{ $t('log.result.success') }}</a-option>
            <a-option :value="2">{{ $t('log.result.failed') }}</a-option>
          </a-select>
        </a-form-item>
        <a-form-item :label="$t('log.search.createdAt')" field="createdRange">
          <a-range-picker
            v-model="searchForm.createdRange"
            show-time
            format="YYYY-MM-DD HH:mm:ss"
            class="search-range"
          />
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
          <a-button @click="() => fetchData()">
            <template #icon><icon-refresh /></template>
            {{ $t('front.common.toolbar.refresh') }}
          </a-button>
          <a-button
            :type="expandAllColumns ? 'outline' : 'secondary'"
            @click="expandAllColumns = !expandAllColumns"
          >
            {{
              expandAllColumns
                ? $t('log.toolbar.collapseColumns')
                : $t('log.toolbar.expandColumns')
            }}
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
        :scroll="expandAllColumns ? undefined : { x: 1600 }"
        :class="{ 'table-expand-all': expandAllColumns }"
        @page-change="onPageChange"
        @page-size-change="onPageSizeChange"
      >
        <template #index="{ rowIndex }">
          {{ rowIndex + 1 + (pagination.current - 1) * pagination.pageSize }}
        </template>
        <template #requestMethod="{ record }">
          <a-tag :color="methodColor(record.requestMethod)">
            {{ record.requestMethod || '--' }}
          </a-tag>
        </template>
        <template #responseStatus="{ record }">
          <a-tag :color="statusColor(record.responseStatus)">
            {{ record.responseStatus }}
          </a-tag>
        </template>
        <template #actionResult="{ record }">
          <a-tag :color="resultColor(record.actionResult)">
            {{ record.actionResult || '--' }}
          </a-tag>
        </template>
        <template #requestPath="{ record }">
          <a-tooltip :content="record.requestPath">
            <span class="ellipsis-text">{{ record.requestPath || '--' }}</span>
          </a-tooltip>
        </template>
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <a-link @click="handleViewDetail(record)">
              {{ $t('log.action.detail') }}
            </a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>
    <LogDetailModal
      v-model:visible="detailModalVisible"
      :record="detailRecord"
    />
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import dayjs from 'dayjs';
  import {
    queryOperationLogList,
    type OperationLogListParams,
    type OperationLogRecord,
  } from '@/api/log';
  import useLoading from '@/hooks/loading';
  import LogDetailModal from './components/log-detail-modal.vue';

  const { t } = useI18n();

  interface OperationLogRow {
    id: number;
    userId: number;
    username: string;
    clientIp: string;
    requestMethod: string;
    requestPath: string;
    responseStatus: number;
    actionResult: string;
    durationMs: number;
    createdAt: string;
    raw: OperationLogRecord;
  }

  const { loading, setLoading } = useLoading(true);
  /** 截图用：取消横向滚动与固定列，一屏展示全部列 */
  const expandAllColumns = ref(false);
  const detailModalVisible = ref(false);
  const detailRecord = ref<OperationLogRecord | null>(null);
  const methodOptions = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'];

  const searchForm = reactive<{
    username: string;
    requestMethod: string;
    requestPath: string;
    actionResult: number | undefined;
    createdRange: string[];
  }>({
    username: '',
    requestMethod: '',
    requestPath: '',
    actionResult: undefined,
    createdRange: [],
  });
  const appliedSearch = reactive({
    username: '',
    requestMethod: '',
    requestPath: '',
    actionResult: undefined as number | undefined,
    createdAtFrom: undefined as number | undefined,
    createdAtTo: undefined as number | undefined,
  });

  const logs = ref<OperationLogRow[]>([]);
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

  const methodColor = (method?: string) => {
    const colors: Record<string, string> = {
      GET: 'arcoblue',
      POST: 'green',
      PUT: 'orange',
      DELETE: 'red',
      PATCH: 'purple',
    };
    return colors[(method || '').toUpperCase()] || 'gray';
  };

  const statusColor = (status?: number) => {
    if (!status) return 'gray';
    if (status >= 200 && status < 300) return 'green';
    if (status >= 400) return 'red';
    return 'orange';
  };

  /** 操作结果：后端未提供明确枚举，按常见语义 success/fail 着色 */
  const resultColor = (result?: string) => {
    const value = (result || '').toLowerCase();
    if (['success', 'ok', '1', '成功'].includes(value)) return 'green';
    if (['fail', 'failed', 'error', '2', '失败'].includes(value)) return 'red';
    return 'gray';
  };

  const mapLog = (item: OperationLogRecord): OperationLogRow => ({
    id: item.id,
    userId: item.user_id,
    username: item.username || '--',
    clientIp: item.client_ip || '--',
    requestMethod: (item.request_method || '--').toUpperCase(),
    requestPath: item.request_path || '--',
    responseStatus: item.response_status,
    actionResult: item.action_result || '--',
    durationMs: item.duration_ms ?? 0,
    createdAt: formatTime(item.created_at),
    raw: item,
  });

  const displayList = computed(() => logs.value);

  const buildQueryParams = (
    page = pagination.current,
    pageSize = pagination.pageSize
  ): OperationLogListParams => {
    const params: OperationLogListParams = {
      page,
      page_size: pageSize,
    };
    const username = appliedSearch.username.trim();
    const requestPath = appliedSearch.requestPath.trim();
    if (username) params.username = username;
    if (appliedSearch.requestMethod)
      params.request_method = appliedSearch.requestMethod;
    if (requestPath) params.request_path = requestPath;
    if (appliedSearch.actionResult !== undefined)
      params.action_result = appliedSearch.actionResult;
    if (appliedSearch.createdAtFrom !== undefined)
      params.created_at_from = appliedSearch.createdAtFrom;
    if (appliedSearch.createdAtTo !== undefined)
      params.created_at_to = appliedSearch.createdAtTo;
    return params;
  };

  const fetchData = async (
    page = pagination.current,
    pageSize = pagination.pageSize
  ) => {
    setLoading(true);
    try {
      const { data } = await queryOperationLogList(
        buildQueryParams(page, pageSize)
      );
      logs.value = (data?.list || []).map(mapLog);
      pagination.current = page;
      pagination.pageSize = pageSize;
      pagination.total = data?.total ?? 0;
    } catch {
      logs.value = [];
      pagination.total = 0;
    } finally {
      setLoading(false);
    }
  };

  const columns: TableColumnData[] = [
    {
      title: t('front.common.column.index'),
      dataIndex: 'index',
      slotName: 'index',
      width: 80,
      align: 'center',
      fixed: 'left',
    },
    {
      title: t('log.column.username'),
      dataIndex: 'username',
      width: 120,
      align: 'center',
    },
    {
      title: t('log.column.userId'),
      dataIndex: 'userId',
      width: 90,
      align: 'center',
    },
    {
      title: t('log.column.clientIp'),
      dataIndex: 'clientIp',
      width: 130,
      align: 'center',
    },
    {
      title: t('log.column.method'),
      dataIndex: 'requestMethod',
      slotName: 'requestMethod',
      width: 100,
      align: 'center',
    },
    {
      title: t('log.column.path'),
      dataIndex: 'requestPath',
      slotName: 'requestPath',
      width: 220,
      align: 'center',
    },
    {
      title: t('front.common.column.status'),
      dataIndex: 'responseStatus',
      slotName: 'responseStatus',
      width: 100,
      align: 'center',
    },
    {
      title: t('log.column.result'),
      dataIndex: 'actionResult',
      slotName: 'actionResult',
      width: 100,
      align: 'center',
    },
    {
      title: t('log.column.duration'),
      dataIndex: 'durationMs',
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
      title: t('front.common.column.operations'),
      dataIndex: 'operations',
      slotName: 'operations',
      width: 90,
      align: 'center',
      fixed: 'right',
    },
  ];

  const displayColumns = computed(() => {
    if (!expandAllColumns.value) return columns;
    return columns.map((col) => {
      const next = { ...col };
      delete next.width;
      delete next.fixed;
      return next;
    });
  });

  const handleSearch = () => {
    appliedSearch.username = searchForm.username;
    appliedSearch.requestMethod = searchForm.requestMethod;
    appliedSearch.requestPath = searchForm.requestPath;
    appliedSearch.actionResult = searchForm.actionResult;
    const [start, end] = searchForm.createdRange || [];
    appliedSearch.createdAtFrom = start
      ? Math.floor(dayjs(start).valueOf() / 1000)
      : undefined;
    appliedSearch.createdAtTo = end
      ? Math.floor(dayjs(end).valueOf() / 1000)
      : undefined;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    Object.assign(searchForm, {
      username: '',
      requestMethod: '',
      requestPath: '',
      actionResult: undefined,
      createdRange: [],
    });
    Object.assign(appliedSearch, {
      username: '',
      requestMethod: '',
      requestPath: '',
      actionResult: undefined,
      createdAtFrom: undefined,
      createdAtTo: undefined,
    });
    fetchData(1, pagination.pageSize);
  };

  const onPageChange = (current: number) => {
    fetchData(current, pagination.pageSize);
  };

  const onPageSizeChange = (pageSize: number) => {
    fetchData(1, pageSize);
  };

  const handleViewDetail = (record: OperationLogRow) => {
    detailRecord.value = record.raw;
    detailModalVisible.value = true;
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
  .search-form {
    row-gap: 16px;
    column-gap: 16px;
  }
  .search-form :deep(.arco-form-item) {
    margin-bottom: 0;
  }
  .search-input {
    width: 170px;
  }
  .search-select :deep(.arco-select-view) {
    min-width: 130px;
  }
  .search-range {
    width: 340px;
  }
  .table-toolbar {
    margin-bottom: 14px;
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
  .operation-links {
    white-space: nowrap;
  }
  .ellipsis-text {
    display: inline-block;
    max-width: 200px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    vertical-align: middle;
  }
  :deep(.arco-pagination-options .arco-select) {
    min-width: 100px;
  }
</style>
