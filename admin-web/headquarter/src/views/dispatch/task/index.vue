<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.nodeManage', 'menu.nodeManage.tasks']" />
    <div class="page-header">
      <div class="page-title">{{ t('taskList.title') }}</div>
      <div class="page-subtitle">{{ t('taskList.subtitle') }}</div>
    </div>
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item :label="t('taskList.search.keyword')" field="keyword">
          <a-input
            v-model="searchForm.keyword"
            :placeholder="t('taskList.search.keywordPlaceholder')"
            allow-clear
            :max-length="64"
            class="search-input"
            @press-enter="handleSearch"
          />
        </a-form-item>
        <a-form-item :label="t('taskList.search.status')" field="status">
          <a-select
            v-model="searchForm.status"
            :placeholder="t('front.common.all')"
            allow-clear
            class="search-select"
            :options="statusOptions"
          />
        </a-form-item>
        <a-form-item :label="t('taskList.search.taskType')" field="taskType">
          <a-select
            v-model="searchForm.taskType"
            :placeholder="t('front.common.all')"
            allow-clear
            class="search-select search-select--wide"
            :options="taskTypeOptions"
          />
        </a-form-item>
        <a-form-item :label="t('taskList.search.nodeCode')" field="nodeCode">
          <a-input
            v-model="searchForm.nodeCode"
            :placeholder="t('taskList.search.nodeCodePlaceholder')"
            allow-clear
            :max-length="64"
            class="search-input"
            @press-enter="handleSearch"
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
        <a-button @click="() => fetchData()">
          <template #icon><icon-refresh /></template>
          {{ t('front.common.toolbar.refresh') }}
        </a-button>
      </div>
      <a-table
        row-key="rowKey"
        :loading="loading"
        :columns="columns"
        :data="list"
        :pagination="pagination"
        :bordered="true"
        :stripe="true"
        :scroll="{ x: 1960 }"
        @page-change="onPageChange"
        @page-size-change="onPageSizeChange"
      >
        <template #status="{ record }">
          <a-tag :color="statusMeta(record.status).color">
            {{ statusMeta(record.status).label }}
          </a-tag>
        </template>
        <template #operations="{ record }">
          <a-link @click="handleDetail(record)">
            {{ t('taskList.action.detail') }}
          </a-link>
        </template>
      </a-table>
    </a-card>

    <TaskDetailModal v-model:visible="detailVisible" :record="detailRecord" />
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import dayjs from 'dayjs';
  import {
    DISPATCH_TASK_STATUS,
    DISPATCH_TASK_TYPE,
    queryDispatchTaskList,
    type DispatchTaskListParams,
    type DispatchTaskRecord,
  } from '@/api/dispatch-task';
  import useLoading from '@/hooks/loading';
  import TaskDetailModal, {
    type TaskRow,
  } from './components/task-detail-modal.vue';

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(true);

  const statusOptions = computed(() => [
    {
      label: t('taskList.status.pending'),
      value: DISPATCH_TASK_STATUS.PENDING,
    },
    {
      label: t('taskList.status.running'),
      value: DISPATCH_TASK_STATUS.RUNNING,
    },
    {
      label: t('taskList.status.success'),
      value: DISPATCH_TASK_STATUS.SUCCESS,
    },
    {
      label: t('taskList.status.failed'),
      value: DISPATCH_TASK_STATUS.FAILED,
    },
  ]);
  const taskTypeOptions = computed(() => [
    {
      label: t('taskList.type.createOperator'),
      value: DISPATCH_TASK_TYPE.CREATE_OPERATOR,
    },
    {
      label: t('taskList.type.updateOperator'),
      value: DISPATCH_TASK_TYPE.UPDATE_OPERATOR,
    },
    {
      label: t('taskList.type.deleteOperator'),
      value: DISPATCH_TASK_TYPE.DELETE_OPERATOR,
    },
  ]);

  const searchForm = reactive<{
    keyword: string;
    status: number | undefined;
    taskType: string | undefined;
    nodeCode: string;
  }>({
    keyword: '',
    status: undefined,
    taskType: undefined,
    nodeCode: '',
  });
  const appliedSearch = reactive<{
    keyword: string;
    status: number | undefined;
    taskType: string | undefined;
    nodeCode: string;
  }>({
    keyword: '',
    status: undefined,
    taskType: undefined,
    nodeCode: '',
  });

  const list = ref<TaskRow[]>([]);
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

  const statusMeta = (status: number | string) => {
    const value = Number(status);
    if (value === DISPATCH_TASK_STATUS.PENDING) {
      return { label: t('taskList.status.pending'), color: 'orange' };
    }
    if (value === DISPATCH_TASK_STATUS.RUNNING) {
      return { label: t('taskList.status.running'), color: 'arcoblue' };
    }
    if (value === DISPATCH_TASK_STATUS.SUCCESS) {
      return { label: t('taskList.status.success'), color: 'green' };
    }
    if (value === DISPATCH_TASK_STATUS.FAILED) {
      return { label: t('taskList.status.failed'), color: 'red' };
    }
    return { label: status ? String(status) : '--', color: 'gray' };
  };

  const formatTime = (value?: number | null) => {
    if (!value) return '--';
    const ms = value < 1e12 ? value * 1000 : value;
    return dayjs(ms).format('YYYY-MM-DD HH:mm:ss');
  };

  const formatJson = (value?: string | null) => {
    if (!value) return '--';
    try {
      return JSON.stringify(JSON.parse(value), null, 2);
    } catch {
      return value;
    }
  };

  const displayText = (value?: string | null) => value || '--';

  const mapRow = (
    item: DispatchTaskRecord,
    index: number,
    page: number
  ): TaskRow => {
    const taskNo = item.task_no || '';
    return {
      rowKey: taskNo || `${page}-${index}`,
      taskNo: displayText(taskNo),
      requestNo: displayText(item.request_no),
      target: displayText(item.target),
      taskType: displayText(item.task_type),
      paramsText: formatJson(item.params),
      nodeCode: displayText(item.node_code),
      status: item.status ?? '',
      resultText: formatJson(item.result),
      errorMessage: displayText(item.error_message),
      startedAt: formatTime(item.started_at),
      finishedAt: formatTime(item.finished_at),
      createdAt: formatTime(item.created_at),
      updatedAt: formatTime(item.updated_at),
    };
  };

  const buildQueryParams = (
    page = pagination.current,
    pageSize = pagination.pageSize
  ): DispatchTaskListParams => {
    const params: DispatchTaskListParams = {
      page,
      page_size: pageSize,
    };
    const keyword = appliedSearch.keyword.trim();
    const nodeCode = appliedSearch.nodeCode.trim();
    if (keyword) params.keyword = keyword;
    if (nodeCode) params.node_code = nodeCode;
    if (appliedSearch.status !== undefined) {
      params.status = appliedSearch.status;
    }
    if (appliedSearch.taskType) params.task_type = appliedSearch.taskType;
    return params;
  };

  const fetchData = async (
    page = pagination.current,
    pageSize = pagination.pageSize
  ) => {
    setLoading(true);
    try {
      const { data } = await queryDispatchTaskList(
        buildQueryParams(page, pageSize)
      );
      list.value = (data?.list || []).map((item, index) =>
        mapRow(item, index, page)
      );
      pagination.current = page;
      pagination.pageSize = pageSize;
      pagination.total = data?.total ?? 0;
    } catch {
      list.value = [];
      pagination.total = 0;
    } finally {
      setLoading(false);
    }
  };

  const columns = computed<TableColumnData[]>(() => [
    {
      title: t('taskList.column.taskNo'),
      dataIndex: 'taskNo',
      width: 160,
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('taskList.column.requestNo'),
      dataIndex: 'requestNo',
      width: 160,
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('taskList.column.target'),
      dataIndex: 'target',
      width: 140,
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('taskList.column.taskType'),
      dataIndex: 'taskType',
      width: 180,
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('taskList.column.nodeCode'),
      dataIndex: 'nodeCode',
      width: 220,
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('taskList.column.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 110,
      align: 'center',
    },
    {
      title: t('taskList.column.errorMessage'),
      dataIndex: 'errorMessage',
      width: 180,
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('taskList.column.startedAt'),
      dataIndex: 'startedAt',
      width: 180,
    },
    {
      title: t('taskList.column.finishedAt'),
      dataIndex: 'finishedAt',
      width: 180,
    },
    {
      title: t('taskList.column.createdAt'),
      dataIndex: 'createdAt',
      width: 180,
    },
    {
      title: t('taskList.column.operations'),
      dataIndex: 'operations',
      slotName: 'operations',
      width: 90,
      align: 'center',
      fixed: 'right',
    },
  ]);

  const handleSearch = () => {
    appliedSearch.keyword = searchForm.keyword;
    appliedSearch.status = searchForm.status;
    appliedSearch.taskType = searchForm.taskType;
    appliedSearch.nodeCode = searchForm.nodeCode;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    Object.assign(searchForm, {
      keyword: '',
      status: undefined,
      taskType: undefined,
      nodeCode: '',
    });
    Object.assign(appliedSearch, {
      keyword: '',
      status: undefined,
      taskType: undefined,
      nodeCode: '',
    });
    fetchData(1, pagination.pageSize);
  };

  const onPageChange = (current: number) => {
    fetchData(current, pagination.pageSize);
  };

  const onPageSizeChange = (pageSize: number) => {
    fetchData(1, pageSize);
  };

  const detailVisible = ref(false);
  const detailRecord = ref<TaskRow | null>(null);

  const handleDetail = (record: TaskRow) => {
    detailRecord.value = record;
    detailVisible.value = true;
  };

  fetchData();
</script>

<style scoped lang="less">
  .container {
    padding: 0 20px 20px;
  }
  .page-header {
    margin-top: 16px;
  }
  .page-title {
    font-size: 20px;
    font-weight: 600;
    color: var(--color-text-1);
  }
  .page-subtitle {
    margin-top: 4px;
    color: var(--color-text-3);
    font-size: 13px;
  }
  .table-toolbar {
    margin-bottom: 14px;
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
    width: 240px;
  }
  .search-select :deep(.arco-select-view) {
    min-width: 130px;
  }
  .search-select--wide :deep(.arco-select-view) {
    min-width: 180px;
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
