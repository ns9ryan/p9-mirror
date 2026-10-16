<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.route.log', 'menu.route.errorLog']" />
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item
          :label="$t('errorLog.search.serviceName')"
          field="serviceName"
        >
          <a-input
            v-model="searchForm.serviceName"
            :placeholder="$t('errorLog.search.serviceNamePlaceholder')"
            allow-clear
            class="search-input"
          />
        </a-form-item>
        <a-form-item :label="$t('errorLog.search.path')" field="requestPath">
          <a-input
            v-model="searchForm.requestPath"
            :placeholder="$t('errorLog.search.pathPlaceholder')"
            allow-clear
            class="search-input"
          />
        </a-form-item>
        <a-form-item :label="$t('errorLog.search.status')" field="responseStatus">
          <a-input-number
            v-model="searchForm.responseStatus"
            :placeholder="$t('errorLog.search.statusPlaceholder')"
            :min="100"
            :max="599"
            hide-button
            allow-clear
            class="search-input"
          />
        </a-form-item>
        <a-form-item :label="$t('errorLog.search.userId')" field="userId">
          <a-input-number
            v-model="searchForm.userId"
            :placeholder="$t('errorLog.search.userIdPlaceholder')"
            :min="1"
            hide-button
            allow-clear
            class="search-input"
          />
        </a-form-item>
        <a-form-item :label="$t('errorLog.search.createdAt')" field="createdRange">
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
                ? $t('errorLog.toolbar.collapseColumns')
                : $t('errorLog.toolbar.expandColumns')
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
        <template #requestPath="{ record }">
          <a-tooltip :content="record.requestPath">
            <span class="ellipsis-text">{{ record.requestPath || '--' }}</span>
          </a-tooltip>
        </template>
        <template #subject="{ record }">
          <a-tooltip :content="record.subject">
            <span class="ellipsis-text">{{ record.subject || '--' }}</span>
          </a-tooltip>
        </template>
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <a-link @click="handleViewDetail(record)">
              {{ $t('errorLog.action.detail') }}
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
    queryErrorLogList,
    type ErrorLogListParams,
    type ErrorLogRecord,
  } from '@/api/log';
  import useLoading from '@/hooks/loading';
  import LogDetailModal from './components/log-detail-modal.vue';

  const { t } = useI18n();

  interface ErrorLogRow {
    id: number;
    serviceName: string;
    userId: number;
    username: string;
    clientIp: string;
    requestMethod: string;
    requestPath: string;
    responseStatus: number;
    subject: string;
    durationMs: number;
    createdAt: string;
    raw: ErrorLogRecord;
  }

  const { loading, setLoading } = useLoading(true);
  const expandAllColumns = ref(false);
  const detailModalVisible = ref(false);
  const detailRecord = ref<ErrorLogRecord | null>(null);

  const searchForm = reactive<{
    serviceName: string;
    requestPath: string;
    responseStatus: number | undefined;
    userId: number | undefined;
    createdRange: string[];
  }>({
    serviceName: '',
    requestPath: '',
    responseStatus: undefined,
    userId: undefined,
    createdRange: [],
  });
  const appliedSearch = reactive({
    serviceName: '',
    requestPath: '',
    responseStatus: undefined as number | undefined,
    userId: undefined as number | undefined,
    createdAtFrom: undefined as number | undefined,
    createdAtTo: undefined as number | undefined,
  });

  const logs = ref<ErrorLogRow[]>([]);
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

  const mapLog = (item: ErrorLogRecord): ErrorLogRow => ({
    id: item.id,
    serviceName: item.service_name || '--',
    userId: item.user_id,
    username: item.username || '--',
    clientIp: item.client_ip || '--',
    requestMethod: (item.request_method || '--').toUpperCase(),
    requestPath: item.request_path || '--',
    responseStatus: item.response_status,
    subject: item.subject || '--',
    durationMs: item.duration_ms ?? 0,
    createdAt: formatTime(item.created_at),
    raw: item,
  });

  const displayList = computed(() => logs.value);

  const buildQueryParams = (
    page = pagination.current,
    pageSize = pagination.pageSize
  ): ErrorLogListParams => {
    const params: ErrorLogListParams = {
      page,
      page_size: pageSize,
    };
    const serviceName = appliedSearch.serviceName.trim();
    const requestPath = appliedSearch.requestPath.trim();
    if (serviceName) params.service_name = serviceName;
    if (requestPath) params.request_path = requestPath;
    if (appliedSearch.responseStatus !== undefined)
      params.response_status = appliedSearch.responseStatus;
    if (appliedSearch.userId !== undefined) params.user_id = appliedSearch.userId;
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
      const { data } = await queryErrorLogList(buildQueryParams(page, pageSize));
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
      title: t('errorLog.column.serviceName'),
      dataIndex: 'serviceName',
      width: 120,
      align: 'center',
    },
    {
      title: t('errorLog.column.username'),
      dataIndex: 'username',
      width: 120,
      align: 'center',
    },
    {
      title: t('errorLog.column.userId'),
      dataIndex: 'userId',
      width: 90,
      align: 'center',
    },
    {
      title: t('errorLog.column.method'),
      dataIndex: 'requestMethod',
      slotName: 'requestMethod',
      width: 100,
      align: 'center',
    },
    {
      title: t('errorLog.column.path'),
      dataIndex: 'requestPath',
      slotName: 'requestPath',
      width: 220,
      align: 'center',
    },
    {
      title: t('front.common.column.status'),
      dataIndex: 'responseStatus',
      slotName: 'responseStatus',
      width: 110,
      align: 'center',
    },
    {
      title: t('errorLog.column.subject'),
      dataIndex: 'subject',
      slotName: 'subject',
      width: 180,
      align: 'center',
    },
    {
      title: t('errorLog.column.duration'),
      dataIndex: 'durationMs',
      width: 100,
      align: 'center',
    },
    {
      title: t('errorLog.column.clientIp'),
      dataIndex: 'clientIp',
      width: 130,
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
    appliedSearch.serviceName = searchForm.serviceName;
    appliedSearch.requestPath = searchForm.requestPath;
    appliedSearch.responseStatus = searchForm.responseStatus;
    appliedSearch.userId = searchForm.userId;
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
      serviceName: '',
      requestPath: '',
      responseStatus: undefined,
      userId: undefined,
      createdRange: [],
    });
    Object.assign(appliedSearch, {
      serviceName: '',
      requestPath: '',
      responseStatus: undefined,
      userId: undefined,
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

  const handleViewDetail = (record: ErrorLogRow) => {
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
