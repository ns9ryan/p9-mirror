<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.outstation', 'operatorGameResource.title']" />
    <div class="page-header">
      <div class="page-title">{{ t('operatorGameResource.title') }}</div>
      <div class="page-subtitle">{{ t('operatorGameResource.subtitle') }}</div>
    </div>
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item
          :label="t('operatorGameResource.search.opCode')"
          field="opCode"
        >
          <a-input
            v-model="searchForm.opCode"
            :placeholder="t('operatorGameResource.search.opCodePlaceholder')"
            allow-clear
            :max-length="64"
            class="search-input"
            @press-enter="handleSearch"
          />
        </a-form-item>
        <!-- 这个是没有用到的状态查询 -->
        <!-- <a-form-item
          :label="t('operatorGameResource.search.status')"
          field="status"
        >
          <a-select
            v-model="searchForm.status"
            :placeholder="t('front.common.all')"
            allow-clear
            class="search-select"
            :options="statusOptions"
          />
        </a-form-item> -->
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
        <template #gameCount="{ record }">
          {{ (record as AllocationRow).gameCount }}
          {{ t('operatorGameResource.column.gameUnit') }}
        </template>
        <template #gameCategoryCount="{ record }">
          {{ (record as AllocationRow).gameCategoryCount }}
          {{ t('operatorGameResource.column.unit') }}
        </template>
        <template #gameProviderCount="{ record }">
          {{ (record as AllocationRow).gameProviderCount }}
          {{ t('operatorGameResource.column.unit') }}
        </template>
        <template #gameChannelCount="{ record }">
          {{ (record as AllocationRow).gameChannelCount }}
          {{ t('operatorGameResource.column.unit') }}
        </template>
        <template #updatedAt="{ record }">
          {{ (record as AllocationRow).updatedAt }}
        </template>
        <template #operations="{ record }">
          <a-space>
            <a-link @click="handleConfigure(record as AllocationRow)">
              {{ t('operatorGameResource.action.configure') }}
            </a-link>
            <a-link
              :disabled="
                publishingOpCode === (record as AllocationRow).opCode
              "
              @click="handlePublish(record as AllocationRow)"
            >
              {{ t('operatorGameResource.action.publish') }}
            </a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>

    <AllocationModal
      v-model:visible="modalVisible"
      :op-code="configuringOpCode"
      @success="handleAllocationSuccess"
    />
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import dayjs from 'dayjs';
  import { Message, Modal } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import {
    OPERATOR_STATUS,
    publishGameResourceAllocation,
    queryGameResourceAllocationList,
    type GameResourceAllocationListParams,
    type GameResourceAllocationRecord,
  } from '@/api/operator';
  import useLoading from '@/hooks/loading';
  import AllocationModal from './components/allocation-modal.vue';

  interface AllocationRow {
    id: number;
    opCode: string;
    opName: string;
    gameCount: number;
    gameCategoryCount: number;
    gameProviderCount: number;
    gameChannelCount: number;
    updatedAt: string;
    raw: GameResourceAllocationRecord;
  }

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(true);

  const searchForm = reactive<{ opCode: string; status?: number }>({
    opCode: '',
    status: undefined,
  });
  const appliedSearch = reactive<{ opCode: string; status?: number }>({
    opCode: '',
    status: undefined,
  });

  const statusOptions = computed(() => [
    {
      label: t('operatorStation.status.normal'),
      value: OPERATOR_STATUS.NORMAL,
    },
    {
      label: t('operatorStation.status.paused'),
      value: OPERATOR_STATUS.PAUSED,
    },
    {
      label: t('operatorStation.status.closed'),
      value: OPERATOR_STATUS.CLOSED,
    },
  ]);

  const list = ref<AllocationRow[]>([]);

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

  const formatTime = (ms?: number | null) => {
    if (!ms) return '--';
    return dayjs(ms).format('YYYY-MM-DD HH:mm:ss');
  };

  const mapRow = (item: GameResourceAllocationRecord): AllocationRow => ({
    id: item.id,
    opCode: item.op_code || '--',
    opName: item.op_name || '--',
    gameCount: item.game_count ?? 0,
    gameCategoryCount: item.game_category_count ?? 0,
    gameProviderCount: item.game_provider_count ?? 0,
    gameChannelCount: item.game_channel_count ?? 0,
    updatedAt: formatTime(item.updated_at),
    raw: item,
  });

  const buildQueryParams = (
    page = pagination.current,
    pageSize = pagination.pageSize
  ): GameResourceAllocationListParams => {
    const params: GameResourceAllocationListParams = {
      page,
      page_size: pageSize,
    };
    const opCode = appliedSearch.opCode.trim();
    if (opCode) params.op_code = opCode;
    if (appliedSearch.status !== undefined) params.status = appliedSearch.status;
    return params;
  };

  const fetchData = async (
    page = pagination.current,
    pageSize = pagination.pageSize
  ) => {
    setLoading(true);
    try {
      const { data } = await queryGameResourceAllocationList(
        buildQueryParams(page, pageSize)
      );
      list.value = (data?.items || []).map(mapRow);
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
      title: t('operatorGameResource.column.id'),
      dataIndex: 'id',
      width: 100,
      align: 'center',
    },
    {
      title: t('operatorGameResource.column.opCode'),
      dataIndex: 'opCode',
      width: 280,
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('operatorGameResource.column.opName'),
      dataIndex: 'opName',
      width: 160,
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('operatorGameResource.column.gameCount'),
      dataIndex: 'gameCount',
      slotName: 'gameCount',
      width: 140,
      align: 'center',
    },
    {
      title: t('operatorGameResource.column.gameCategoryCount'),
      dataIndex: 'gameCategoryCount',
      slotName: 'gameCategoryCount',
      width: 140,
      align: 'center',
    },
    {
      title: t('operatorGameResource.column.gameProviderCount'),
      dataIndex: 'gameProviderCount',
      slotName: 'gameProviderCount',
      width: 140,
      align: 'center',
    },
    {
      title: t('operatorGameResource.column.gameChannelCount'),
      dataIndex: 'gameChannelCount',
      slotName: 'gameChannelCount',
      width: 140,
      align: 'center',
    },
    {
      title: t('operatorGameResource.column.updatedAt'),
      dataIndex: 'updatedAt',
      slotName: 'updatedAt',
      width: 180,
      align: 'center',
    },
    {
      title: t('operatorGameResource.column.operations'),
      dataIndex: 'operations',
      slotName: 'operations',
      width: 140,
      align: 'center',
      fixed: 'right',
    },
  ]);

  const modalVisible = ref(false);
  const configuringOpCode = ref<string | null>(null);
  const publishingOpCode = ref<string | null>(null);

  const handleConfigure = (record: AllocationRow) => {
    configuringOpCode.value = record.opCode === '--' ? null : record.opCode;
    modalVisible.value = true;
  };

  const handlePublish = (record: AllocationRow) => {
    if (!record.opCode || record.opCode === '--') return;
    Modal.confirm({
      title: t('operatorGameResource.message.publishConfirmTitle'),
      content: t('operatorGameResource.message.publishConfirmContent', {
        name: record.opName,
      }),
      async onOk() {
        publishingOpCode.value = record.opCode;
        try {
          await publishGameResourceAllocation({ op_code: record.opCode });
          Message.success(t('operatorGameResource.message.publishSuccess'));
          await fetchData(pagination.current, pagination.pageSize);
        } finally {
          publishingOpCode.value = null;
        }
      },
    });
  };

  const handleAllocationSuccess = () => {
    fetchData(pagination.current, pagination.pageSize);
  };

  const handleSearch = () => {
    appliedSearch.opCode = searchForm.opCode;
    appliedSearch.status = searchForm.status;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    searchForm.opCode = '';
    searchForm.status = undefined;
    appliedSearch.opCode = '';
    appliedSearch.status = undefined;
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
    width: 280px;
  }
  .search-select {
    width: 180px;
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
