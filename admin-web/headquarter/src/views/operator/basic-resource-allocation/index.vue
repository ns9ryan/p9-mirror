<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.outstation', 'operatorBasicResource.title']" />
    <div class="page-header">
      <div class="page-title">{{ t('operatorBasicResource.title') }}</div>
      <div class="page-subtitle">{{ t('operatorBasicResource.subtitle') }}</div>
    </div>
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item
          :label="t('operatorBasicResource.search.keyword')"
          field="keyword"
        >
          <a-input
            v-model="searchForm.keyword"
            :placeholder="t('operatorBasicResource.search.keywordPlaceholder')"
            allow-clear
            :max-length="100"
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
        <a-space>
          <a-button @click="() => fetchData()">
            <template #icon><icon-refresh /></template>
            {{ t('front.common.toolbar.refresh') }}
          </a-button>
        </a-space>
      </div>
      <a-table
        row-key="operatorId"
        :loading="loading"
        :columns="columns"
        :data="list"
        :pagination="pagination"
        :bordered="true"
        :stripe="true"
        :scroll="{ x: 1200 }"
        @page-change="onPageChange"
        @page-size-change="onPageSizeChange"
      >
        <template #languageCount="{ record }">
          {{ (record as AllocationRow).languageCount }}
          {{ t('operatorBasicResource.column.unit') }}
        </template>
        <template #regionCount="{ record }">
          {{ (record as AllocationRow).regionCount }}
          {{ t('operatorBasicResource.column.unit') }}
        </template>
        <template #agentLineCount="{ record }">
          {{ (record as AllocationRow).agentLineCount }}
          {{ t('operatorBasicResource.column.unit') }}
        </template>
        <template #operations="{ record }">
          <a-space>
            <a-link @click="handleConfigure(record as AllocationRow)">
              {{ t('operatorBasicResource.action.configure') }}
            </a-link>
            <a-link
              :disabled="
                publishingOpCode === (record as AllocationRow).op_code
              "
              @click="handlePublish(record as AllocationRow)"
            >
              {{ t('operatorBasicResource.action.publish') }}
            </a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>

    <AllocationModal
      v-model:visible="modalVisible"
      :operator-id="configuringOperatorId"
      @success="handleAllocationSuccess"
    />
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, Modal } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import {
    publishBasicResourceAllocation,
    queryBasicResourceAllocationList,
    type BasicResourceAllocationListParams,
    type BasicResourceAllocationRecord,
  } from '@/api/operator';
  import useLoading from '@/hooks/loading';
  import AllocationModal from './components/allocation-modal.vue';

  interface AllocationRow {
    operatorId: number;
    op_code: string;
    operatorName: string;
    languageCount: number;
    regionCount: number;
    agentLineCount: number;
    raw: BasicResourceAllocationRecord;
  }

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(true);

  const searchForm = reactive<{ keyword: string }>({
    keyword: '',
  });
  const appliedSearch = reactive<{ keyword: string }>({
    keyword: '',
  });

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

  const mapRow = (item: BasicResourceAllocationRecord): AllocationRow => ({
    operatorId: item.operator_id,
    op_code: item.op_code || '--',
    operatorName: item.operator_name || '--',
    languageCount: item.language_count ?? 0,
    regionCount: item.region_count ?? 0,
    agentLineCount: item.agent_line_count ?? 0,
    raw: item,
  });

  const buildQueryParams = (
    page = pagination.current,
    pageSize = pagination.pageSize
  ): BasicResourceAllocationListParams => {
    const params: BasicResourceAllocationListParams = {
      page,
      page_size: pageSize,
    };
    const keyword = appliedSearch.keyword.trim();
    if (keyword) params.keyword = keyword;
    return params;
  };

  const fetchData = async (
    page = pagination.current,
    pageSize = pagination.pageSize
  ) => {
    setLoading(true);
    try {
      const { data } = await queryBasicResourceAllocationList(
        buildQueryParams(page, pageSize)
      );
      list.value = (data?.list || []).map(mapRow);
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
      title: t('operatorBasicResource.column.operatorId'),
      dataIndex: 'operatorId',
      width: 100,
      align: 'center',
    },
    {
      title: t('operatorBasicResource.column.operatorCode'),
      dataIndex: 'op_code',
      width: 280,
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('operatorBasicResource.column.operatorName'),
      dataIndex: 'operatorName',
      width: 160,
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('operatorBasicResource.column.languageCount'),
      dataIndex: 'languageCount',
      slotName: 'languageCount',
      width: 140,
      align: 'center',
    },
    {
      title: t('operatorBasicResource.column.regionCount'),
      dataIndex: 'regionCount',
      slotName: 'regionCount',
      width: 160,
      align: 'center',
    },
    {
      title: t('operatorBasicResource.column.agentLineCount'),
      dataIndex: 'agentLineCount',
      slotName: 'agentLineCount',
      width: 180,
      align: 'center',
    },
    {
      title: t('operatorBasicResource.column.operations'),
      dataIndex: 'operations',
      slotName: 'operations',
      width: 140,
      align: 'center',
      fixed: 'right',
    },
  ]);

  const handleSearch = () => {
    appliedSearch.keyword = searchForm.keyword;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    searchForm.keyword = '';
    appliedSearch.keyword = '';
    fetchData(1, pagination.pageSize);
  };

  const onPageChange = (current: number) => {
    fetchData(current, pagination.pageSize);
  };

  const onPageSizeChange = (pageSize: number) => {
    fetchData(1, pageSize);
  };

  const modalVisible = ref(false);
  const configuringOperatorId = ref<number | null>(null);
  const publishingOpCode = ref<string | null>(null);

  const handleConfigure = (record: AllocationRow) => {
    configuringOperatorId.value = record.operatorId;
    modalVisible.value = true;
  };

  const handlePublish = (record: AllocationRow) => {
    if (!record.op_code || record.op_code === '--') return;
    Modal.confirm({
      title: t('operatorBasicResource.message.publishConfirmTitle'),
      content: t('operatorBasicResource.message.publishConfirmContent', {
        name: record.operatorName,
      }),
      async onOk() {
        publishingOpCode.value = record.op_code;
        try {
          await publishBasicResourceAllocation({
            op_code: record.op_code,
          });
          Message.success(t('operatorBasicResource.message.publishSuccess'));
          await fetchData(pagination.current, pagination.pageSize);
        } finally {
          publishingOpCode.value = null;
        }
      },
    });
  };

  const handleAllocationSuccess = () => {
    modalVisible.value = false;
    fetchData(pagination.current, pagination.pageSize);
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
