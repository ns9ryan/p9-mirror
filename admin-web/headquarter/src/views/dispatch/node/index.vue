<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.nodeManage', 'menu.nodeManage.nodeList']" />
    <div class="page-header">
      <div class="page-title">{{ t('nodeList.title') }}</div>
      <div class="page-subtitle">{{ t('nodeList.subtitle') }}</div>
    </div>
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item :label="t('nodeList.search.keyword')" field="keyword">
          <a-input
            v-model="searchForm.keyword"
            :placeholder="t('nodeList.search.keywordPlaceholder')"
            allow-clear
            :max-length="100"
            class="search-input"
            @press-enter="handleSearch"
          />
        </a-form-item>
        <a-form-item :label="t('nodeList.search.status')" field="status">
          <a-select
            v-model="searchForm.status"
            :placeholder="t('front.common.all')"
            allow-clear
            class="search-select"
            :options="statusOptions"
          />
        </a-form-item>
        <a-form-item
          :label="t('nodeList.search.onlineStatus')"
          field="onlineStatus"
        >
          <a-select
            v-model="searchForm.onlineStatus"
            :placeholder="t('front.common.all')"
            allow-clear
            class="search-select"
            :options="onlineStatusOptions"
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
          <a-button type="primary" @click="handleCreate">
            <template #icon><icon-plus /></template>
            {{ t('nodeList.toolbar.create') }}
          </a-button>
          <a-button @click="() => fetchData()">
            <template #icon><icon-refresh /></template>
            {{ t('front.common.toolbar.refresh') }}
          </a-button>
        </a-space>
      </div>
      <a-table
        row-key="id"
        :loading="loading"
        :columns="columns"
        :data="list"
        :pagination="pagination"
        :bordered="true"
        :stripe="true"
        :scroll="{ x: 1400 }"
        @page-change="onPageChange"
        @page-size-change="onPageSizeChange"
      >
        <template #name="{ record }">
          {{ record.name || '--' }}
        </template>
        <template #status="{ record }">
          <a-tag :color="statusMeta(record.status).color">
            {{ statusMeta(record.status).label }}
          </a-tag>
        </template>
        <template #onlineStatus="{ record }">
          <span class="status-dot-text">
            <span
              class="status-dot"
              :class="onlineStatusMeta(record.online).dotClass"
            />
            {{ onlineStatusMeta(record.online).label }}
          </span>
        </template>
        <template #assignedOperatorCount="{ record }">
          {{ record.assignedOperatorCount }}{{ t('nodeList.column.unit') }}
        </template>
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <a-link @click="handleDetail(record)">
              {{ t('nodeList.action.detail') }}
            </a-link>
            <a-link @click="handleEdit(record)">
              {{ t('nodeList.action.edit') }}
            </a-link>
            <a-link status="danger" @click="handleResetSecret(record)">
              {{ t('nodeList.action.resetSecret') }}
            </a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>

    <CreateNodeModal
      v-model:visible="createModalVisible"
      @success="() => fetchData(1, pagination.pageSize)"
    />

    <EditNodeModal
      v-model:visible="editModalVisible"
      :record="editRecord"
      @success="() => fetchData()"
    />

    <ResetSecretModal
      v-model:visible="resetModalVisible"
      :record="resetRecord"
    />

    <NodeDetailModal
      v-model:visible="detailModalVisible"
      :node-id="detailNodeId"
    />
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, Modal } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import dayjs from 'dayjs';
  import {
    NODE_ONLINE_STATUS,
    NODE_STATUS,
    queryNodeList,
    resetNodeAuthSecret,
    type NodeListParams,
    type NodeRecord,
  } from '@/api/node';
  import useLoading from '@/hooks/loading';
  import CreateNodeModal from './components/create-node-modal.vue';
  import EditNodeModal from './components/edit-node-modal.vue';
  import NodeDetailModal from './components/node-detail-modal.vue';
  import ResetSecretModal from './components/reset-secret-modal.vue';

  interface NodeRow {
    id: number;
    code: string;
    name: string;
    status: number;
    online: boolean;
    assignedOperatorCount: number;
    lastActiveAt: string;
    remark: string;
  }

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(true);

  const statusOptions = computed(() => [
    { label: t('nodeList.status.enabled'), value: NODE_STATUS.ENABLED },
    { label: t('nodeList.status.disabled'), value: NODE_STATUS.DISABLED },
  ]);
  const onlineStatusOptions = computed(() => [
    {
      label: t('nodeList.onlineStatus.online'),
      value: NODE_ONLINE_STATUS.ONLINE,
    },
    {
      label: t('nodeList.onlineStatus.offline'),
      value: NODE_ONLINE_STATUS.OFFLINE,
    },
  ]);

  const statusMeta = (status: number) => {
    if (status === NODE_STATUS.DISABLED) {
      return { label: t('nodeList.status.disabled'), color: 'red' };
    }
    return { label: t('nodeList.status.enabled'), color: 'green' };
  };

  const onlineStatusMeta = (online: boolean) => {
    if (online === NODE_ONLINE_STATUS.ONLINE) {
      return {
        label: t('nodeList.onlineStatus.online'),
        dotClass: 'status-dot--online',
      };
    }
    return {
      label: t('nodeList.onlineStatus.offline'),
      dotClass: 'status-dot--offline',
    };
  };

  const searchForm = reactive<{
    keyword: string;
    status: number | undefined;
    onlineStatus: boolean | undefined;
  }>({
    keyword: '',
    status: undefined,
    onlineStatus: undefined,
  });
  const appliedSearch = reactive<{
    keyword: string;
    status: number | undefined;
    onlineStatus: boolean | undefined;
  }>({
    keyword: '',
    status: undefined,
    onlineStatus: undefined,
  });

  const list = ref<NodeRow[]>([]);

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

  const formatTime = (value?: number | null) => {
    if (!value) return '--';
    const ms = value < 1e12 ? value * 1000 : value;
    return dayjs(ms).format('YYYY-MM-DD HH:mm:ss');
  };

  const mapRow = (item: NodeRecord): NodeRow => ({
    id: item.id,
    code: item.code || '--',
    name: item.name || '',
    status: item.status,
    online: item.online,
    assignedOperatorCount: item.assigned_operator_count ?? 0,
    lastActiveAt: formatTime(item.last_active_at),
    remark: item.remark || '',
  });

  /** 前端本地过滤：接口暂未提供 status / online 筛选参数 */
  const buildQueryParams = (
    page = pagination.current,
    pageSize = pagination.pageSize
  ): NodeListParams => {
    const params: NodeListParams = {
      page,
      page_size: pageSize,
    };
    const keyword = appliedSearch.keyword.trim();
    if (keyword) params.keyword = keyword;
    return params;
  };

  const applyLocalFilter = (rows: NodeRow[]) =>
    rows.filter((row) => {
      if (
        appliedSearch.status !== undefined &&
        row.status !== appliedSearch.status
      ) {
        return false;
      }
      if (
        appliedSearch.onlineStatus !== undefined &&
        row.online !== appliedSearch.onlineStatus
      ) {
        return false;
      }
      return true;
    });

  const fetchData = async (
    page = pagination.current,
    pageSize = pagination.pageSize
  ) => {
    setLoading(true);
    try {
      const { data } = await queryNodeList(buildQueryParams(page, pageSize));
      const rows = (data?.list || []).map(mapRow);
      list.value = applyLocalFilter(rows);
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
      title: t('nodeList.column.id'),
      dataIndex: 'id',
      width: 90,
      align: 'center',
    },
    {
      title: t('nodeList.column.code'),
      dataIndex: 'code',
      width: 150,
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('nodeList.column.name'),
      dataIndex: 'name',
      slotName: 'name',
      width: 150,
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('nodeList.column.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 110,
      align: 'center',
    },
    {
      title: t('nodeList.column.onlineStatus'),
      dataIndex: 'online',
      slotName: 'onlineStatus',
      width: 110,
      align: 'center',
    },
    {
      title: t('nodeList.column.lastActiveAt'),
      dataIndex: 'lastActiveAt',
      width: 180,
      align: 'center',
    },
    {
      title: t('nodeList.column.assignedOperatorCount'),
      dataIndex: 'assignedOperatorCount',
      slotName: 'assignedOperatorCount',
      width: 120,
      align: 'center',
    },
    {
      title: t('nodeList.column.remark'),
      dataIndex: 'remark',
      width: 160,
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('nodeList.column.operations'),
      dataIndex: 'operations',
      slotName: 'operations',
      width: 200,
      align: 'center',
      fixed: 'right',
    },
  ]);

  const handleSearch = () => {
    appliedSearch.keyword = searchForm.keyword;
    appliedSearch.status = searchForm.status;
    appliedSearch.onlineStatus = searchForm.onlineStatus;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    Object.assign(searchForm, {
      keyword: '',
      status: undefined,
      onlineStatus: undefined,
    });
    Object.assign(appliedSearch, {
      keyword: '',
      status: undefined,
      onlineStatus: undefined,
    });
    fetchData(1, pagination.pageSize);
  };

  const onPageChange = (current: number) => {
    fetchData(current, pagination.pageSize);
  };

  const onPageSizeChange = (pageSize: number) => {
    fetchData(1, pageSize);
  };

  const createModalVisible = ref(false);

  const handleCreate = () => {
    createModalVisible.value = true;
  };

  const editModalVisible = ref(false);
  const editRecord = ref<NodeRow | null>(null);

  const resetModalVisible = ref(false);
  const resetRecord = ref<{
    id: number;
    name: string;
    code: string;
    authSecret: string;
  } | null>(null);

  watch(resetModalVisible, (visible) => {
    if (!visible) resetRecord.value = null;
  });

  const detailModalVisible = ref(false);
  const detailNodeId = ref<number | null>(null);

  const handleDetail = (record: NodeRow) => {
    detailNodeId.value = record.id;
    detailModalVisible.value = true;
  };

  const handleEdit = (record: NodeRow) => {
    editRecord.value = record;
    editModalVisible.value = true;
  };

  const handleResetSecret = (record: NodeRow) => {
    Modal.confirm({
      title: t('nodeList.message.resetSecretConfirmTitle'),
      content: t('nodeList.message.resetSecretConfirmContent', {
        name: record.name || '--',
      }),
      okText: t('nodeList.reset.okText'),
      okButtonProps: { status: 'danger' },
      async onOk() {
        const { data } = await resetNodeAuthSecret({ id: record.id });
        if (!data?.auth_secret) {
          Message.error(t('nodeList.reset.message.secretMissing'));
          return;
        }
        resetRecord.value = {
          id: record.id,
          name: record.name,
          code: record.code,
          authSecret: data.auth_secret,
        };
        resetModalVisible.value = true;
      },
    });
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
    width: 220px;
  }
  .search-select :deep(.arco-select-view) {
    min-width: 130px;
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
  :deep(.arco-pagination-options .arco-select) {
    min-width: 100px;
  }
  .status-dot-text {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .status-dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }
  .status-dot--online {
    background-color: rgb(var(--success-6));
  }
  .status-dot--offline {
    background-color: rgb(var(--gray-5));
  }
</style>
