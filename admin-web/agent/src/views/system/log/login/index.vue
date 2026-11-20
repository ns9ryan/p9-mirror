<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.route.log', 'menu.route.loginLog']" />
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item :label="$t('loginLog.search.username')" field="username">
          <a-input
            v-model="searchForm.username"
            :placeholder="$t('loginLog.search.usernamePlaceholder')"
            allow-clear
            class="search-input"
          />
        </a-form-item>
        <a-form-item :label="$t('loginLog.search.result')" field="loginResult">
          <a-select
            v-model="searchForm.loginResult"
            :placeholder="$t('loginLog.search.resultPlaceholder')"
            allow-clear
            class="search-select"
          >
            <a-option :value="1">{{ $t('loginLog.result.success') }}</a-option>
            <a-option :value="2">{{ $t('loginLog.result.failed') }}</a-option>
          </a-select>
        </a-form-item>
        <a-form-item :label="$t('loginLog.search.loginAt')" field="loginRange">
          <a-range-picker
            v-model="searchForm.loginRange"
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
        </a-space>
      </div>
      <a-table
        row-key="id"
        :loading="loading"
        :columns="columns"
        :data="displayList"
        :pagination="pagination"
        :bordered="true"
        :stripe="true"
        :scroll="{ x: 1100 }"
        @page-change="onPageChange"
        @page-size-change="onPageSizeChange"
      >
        <template #index="{ rowIndex }">
          {{ rowIndex + 1 + (pagination.current - 1) * pagination.pageSize }}
        </template>
        <template #loginResult="{ record }">
          <a-tag :color="resultColor(record.loginResult)">
            {{ record.loginResult || '--' }}
          </a-tag>
        </template>
        <template #failureReason="{ record }">
          <a-tooltip :content="record.failureReason">
            <span class="ellipsis-text">{{ record.failureReason || '--' }}</span>
          </a-tooltip>
        </template>
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <a-link @click="handleViewDetail(record)">
              {{ $t('loginLog.action.detail') }}
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
    queryLoginLogList,
    type LoginLogListParams,
    type LoginLogRecord,
  } from '@/api/log';
  import useLoading from '@/hooks/loading';
  import LogDetailModal from './components/log-detail-modal.vue';

  const { t } = useI18n();

  interface LoginLogRow {
    id: number;
    userId: number;
    username: string;
    loginIp: string;
    deviceId: number;
    loginResult: string;
    failureReason: string;
    loginAt: string;
    raw: LoginLogRecord;
  }

  const { loading, setLoading } = useLoading(true);
  const detailModalVisible = ref(false);
  const detailRecord = ref<LoginLogRecord | null>(null);

  const searchForm = reactive<{
    username: string;
    loginResult: number | undefined;
    loginRange: string[];
  }>({
    username: '',
    loginResult: undefined,
    loginRange: [],
  });
  const appliedSearch = reactive({
    username: '',
    loginResult: undefined as number | undefined,
    loginAtFrom: undefined as number | undefined,
    loginAtTo: undefined as number | undefined,
  });

  const logs = ref<LoginLogRow[]>([]);
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

  /** 登录结果：后端未提供明确枚举，按常见语义 success/fail 着色 */
  const resultColor = (result?: string) => {
    const value = (result || '').toLowerCase();
    if (['success', 'ok', '1', '成功'].includes(value)) return 'green';
    if (['fail', 'failed', 'error', '2', '失败'].includes(value)) return 'red';
    return 'gray';
  };

  const mapLog = (item: LoginLogRecord): LoginLogRow => ({
    id: item.id,
    userId: item.user_id,
    username: item.username || '--',
    loginIp: item.login_ip || '--',
    deviceId: item.device_id,
    loginResult: item.login_result || '--',
    failureReason: item.failure_reason || '--',
    loginAt: formatTime(item.login_at),
    raw: item,
  });

  const displayList = computed(() => logs.value);

  const buildQueryParams = (
    page = pagination.current,
    pageSize = pagination.pageSize
  ): LoginLogListParams => {
    const params: LoginLogListParams = {
      page,
      page_size: pageSize,
    };
    const username = appliedSearch.username.trim();
    if (username) params.username = username;
    if (appliedSearch.loginResult !== undefined)
      params.login_result = appliedSearch.loginResult;
    if (appliedSearch.loginAtFrom !== undefined)
      params.login_at_from = appliedSearch.loginAtFrom;
    if (appliedSearch.loginAtTo !== undefined)
      params.login_at_to = appliedSearch.loginAtTo;
    return params;
  };

  const fetchData = async (
    page = pagination.current,
    pageSize = pagination.pageSize
  ) => {
    setLoading(true);
    try {
      const { data } = await queryLoginLogList(
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
      title: t('loginLog.column.username'),
      dataIndex: 'username',
      width: 120,
      align: 'center',
    },
    {
      title: t('loginLog.column.userId'),
      dataIndex: 'userId',
      width: 90,
      align: 'center',
    },
    {
      title: t('loginLog.column.loginIp'),
      dataIndex: 'loginIp',
      width: 130,
      align: 'center',
    },
    {
      title: t('loginLog.column.deviceId'),
      dataIndex: 'deviceId',
      width: 100,
      align: 'center',
    },
    {
      title: t('loginLog.column.result'),
      dataIndex: 'loginResult',
      slotName: 'loginResult',
      width: 100,
      align: 'center',
    },
    {
      title: t('loginLog.column.failureReason'),
      dataIndex: 'failureReason',
      slotName: 'failureReason',
      width: 200,
      align: 'center',
    },
    {
      title: t('loginLog.column.loginAt'),
      dataIndex: 'loginAt',
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

  const handleSearch = () => {
    appliedSearch.username = searchForm.username;
    appliedSearch.loginResult = searchForm.loginResult;
    const [start, end] = searchForm.loginRange || [];
    appliedSearch.loginAtFrom = start
      ? Math.floor(dayjs(start).valueOf() / 1000)
      : undefined;
    appliedSearch.loginAtTo = end
      ? Math.floor(dayjs(end).valueOf() / 1000)
      : undefined;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    Object.assign(searchForm, {
      username: '',
      loginResult: undefined,
      loginRange: [],
    });
    Object.assign(appliedSearch, {
      username: '',
      loginResult: undefined,
      loginAtFrom: undefined,
      loginAtTo: undefined,
    });
    fetchData(1, pagination.pageSize);
  };

  const onPageChange = (current: number) => {
    fetchData(current, pagination.pageSize);
  };

  const onPageSizeChange = (pageSize: number) => {
    fetchData(1, pageSize);
  };

  const handleViewDetail = (record: LoginLogRow) => {
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