<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.route.system', 'menu.route.api']" />
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item :label="$t('apiManage.search.path')" field="keyword">
          <a-input
            v-model="searchForm.keyword"
            :placeholder="$t('apiManage.search.placeholder')"
            allow-clear
            class="search-input"
          />
        </a-form-item>
        <a-form-item :label="$t('apiManage.search.group')" field="apiGroup">
          <a-input
            v-model="searchForm.apiGroup"
            :placeholder="$t('apiManage.search.groupPlaceholder')"
            allow-clear
            class="search-input"
          />
        </a-form-item>
        <a-form-item :label="$t('apiManage.search.method')" field="method">
          <a-select
            v-model="searchForm.method"
            :placeholder="$t('apiManage.search.methodPlaceholder')"
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
        <a-form-item :label="$t('apiManage.search.serviceName')" field="serviceName">
          <a-input
            v-model="searchForm.serviceName"
            :placeholder="$t('apiManage.search.serviceNamePlaceholder')"
            allow-clear
            class="search-input"
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
          <a-button type="primary" @click="handleCreate">
            <template #icon><icon-plus /></template>
            {{ $t('front.common.toolbar.add') }}
          </a-button>
          <a-button @click="() => fetchData()">
            <template #icon><icon-refresh /></template>
            {{ $t('front.common.toolbar.refresh') }}
          </a-button>
          <a-button
            :type="expandAllColumns ? 'outline' : 'secondary'"
            @click="expandAllColumns = !expandAllColumns"
          >
            {{ expandAllColumns ? $t('apiManage.toolbar.collapseColumns') : $t('apiManage.toolbar.expandColumns') }}
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
        :scroll="expandAllColumns ? undefined : { x: 1280 }"
        :class="{ 'table-expand-all': expandAllColumns }"
        @page-change="onPageChange"
        @page-size-change="onPageSizeChange"
      >
        <template #index="{ rowIndex }">
          {{ rowIndex + 1 + (pagination.current - 1) * pagination.pageSize }}
        </template>
        <template #method="{ record }">
          <a-tag :color="methodColor(record.method)">{{ record.method }}</a-tag>
        </template>
        <template #isRequired="{ record }">
          <a-tag :color="record.isRequired ? 'orangered' : 'gray'">
            {{ record.isRequired ? $t('apiManage.isRequired.yes') : $t('apiManage.isRequired.no') }}
          </a-tag>
        </template>
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <a-link @click="handleEdit(record)">{{ $t('front.common.action.edit') }}</a-link>
            <a-link
              status="danger"
              :disabled="deletingId === record.id"
              @click="handleDelete(record)"
            >
              {{ $t('front.common.action.delete') }}
            </a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>
    <ApiModal
      :visible="modalVisible"
      :api="editingApi"
      @update:visible="onModalVisibleChange"
      @success="fetchData"
    />
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, Modal } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import dayjs from 'dayjs';
  import {
    deleteApi,
    queryApiList,
    type ApiListParams,
    type ApiListRes,
    type ApiRecord,
  } from '@/api/apiManage';
  import useLoading from '@/hooks/loading';
  import ApiModal from './components/api-modal.vue';

  const { t } = useI18n();

  interface ApiRow {
    id: number;
    trans_description: string;
    apiGroup: string;
    method: string;
    path: string;
    isRequired: boolean;
    serviceName: string;
    createdAt: string;
    updatedAt: string;
    raw: ApiRecord;
  }

  const { loading, setLoading } = useLoading(true);
  const modalVisible = ref(false);
  const editingApi = ref<ApiRecord | null>(null);
  const deletingId = ref<number | null>(null);
  /** 截图用：取消横向滚动与固定列，一屏展示全部列 */
  const expandAllColumns = ref(false);
  const searchForm = reactive<{
    keyword: string;
    apiGroup: string;
    method: string;
    serviceName: string;
  }>({
    keyword: '',
    apiGroup: '',
    method: '',
    serviceName: '',
  });
  const appliedSearch = reactive({
    keyword: '',
    apiGroup: '',
    method: '',
    serviceName: '',
  });

  const apis = ref<ApiRow[]>([]);
  const methodOptions = ['DELETE', 'GET', 'PATCH', 'POST', 'PUT'];
  const pagination = reactive({
    current: 1,
    pageSize: 10,
    total: 0,
    showTotal: true,
    showPageSize: true,
    pageSizeOptions: [10, 15, 20, 30, 50, 100],
    pageSizeProps: {
      popupContainer: 'body',
    },
  });

  const formatTime = (value?: number) => {
    if (!value) return '--';
    const ms = value < 1e12 ? value * 1000 : value;
    return dayjs(ms).format('YYYY-MM-DD HH:mm:ss');
  };

  const mapApi = (item: ApiRecord): ApiRow => ({
    id: item.id,
    trans_description: item.trans_description || item.description || '--',
    apiGroup: item.api_group || '--',
    method: (item.method || '--').toUpperCase(),
    path: item.path || '--',
    isRequired: Boolean(item.is_required),
    serviceName: item.service_name || '--',
    createdAt: formatTime(item.created_at),
    updatedAt: formatTime(item.updated_at),
    raw: item,
  });

  const normalizeList = (data?: ApiListRes | ApiRecord[]) => {
    if (!data) return { list: [] as ApiRecord[], total: 0 };
    if (Array.isArray(data)) {
      return { list: data, total: data.length };
    }
    const list = data.list || [];
    return { list, total: data.total ?? list.length };
  };

  const buildQueryParams = (
    page = pagination.current,
    pageSize = pagination.pageSize
  ): ApiListParams => {
    const params: ApiListParams = {
      page,
      page_size: pageSize,
    };
    const keyword = appliedSearch.keyword.trim();
    const apiGroup = appliedSearch.apiGroup.trim();
    const serviceName = appliedSearch.serviceName.trim();
    if (keyword) params.path = keyword;
    if (apiGroup) params.api_group = apiGroup;
    if (appliedSearch.method) params.method = appliedSearch.method;
    if (serviceName) params.service_name = serviceName;
    return params;
  };

  /** 接口筛选参数若暂未生效，对当前页做本地兜底过滤 */
  const displayList = computed(() =>
    apis.value.filter((api) => {
      const keyword = appliedSearch.keyword.trim().toLowerCase();
      const group = appliedSearch.apiGroup.trim().toLowerCase();
      const service = appliedSearch.serviceName.trim().toLowerCase();
      const matchedKeyword =
        !keyword || api.path.toLowerCase().includes(keyword);
      const matchedGroup =
        !group || api.apiGroup.toLowerCase().includes(group);
      const matchedMethod =
        !appliedSearch.method || api.method === appliedSearch.method;
      const matchedService =
        !service || api.serviceName.toLowerCase().includes(service);
      return (
        matchedKeyword && matchedGroup && matchedMethod && matchedService
      );
    })
  );

  const methodColor = (method: string) => {
    const colors: Record<string, string> = {
      GET: 'arcoblue',
      POST: 'green',
      PUT: 'orange',
      DELETE: 'red',
      PATCH: 'purple',
    };
    return colors[method] || 'gray';
  };

  const fetchData = async (
    page = pagination.current,
    pageSize = pagination.pageSize
  ) => {
    setLoading(true);
    try {
      const { data } = await queryApiList(buildQueryParams(page, pageSize));
      const { list, total } = normalizeList(data);
      apis.value = list.map(mapApi);
      pagination.current = page;
      pagination.pageSize = pageSize;
      pagination.total = total;
    } catch {
      apis.value = [];
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
      width: 90,
      align: 'center',
      fixed: 'left',
    },
    {
      title: t('apiManage.column.description'),
      dataIndex: 'trans_description',
      width: 160,
      align: 'center',
    },
    {
      title: t('apiManage.column.group'),
      dataIndex: 'apiGroup',
      width: 120,
      align: 'center',
    },
    {
      title: t('apiManage.column.method'),
      dataIndex: 'method',
      slotName: 'method',
      width: 100,
      align: 'center',
    },
    {
      title: t('apiManage.column.path'),
      dataIndex: 'path',
      width: 200,
      align: 'center',
    },
    {
      title: t('apiManage.column.isRequired'),
      dataIndex: 'isRequired',
      slotName: 'isRequired',
      width: 90,
      align: 'center',
    },
    {
      title: t('apiManage.column.serviceName'),
      dataIndex: 'serviceName',
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
      title: t('front.common.column.updatedAt'),
      dataIndex: 'updatedAt',
      width: 170,
      align: 'center',
    },
    {
      title: t('front.common.column.operations'),
      dataIndex: 'operations',
      slotName: 'operations',
      width: 140,
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
    appliedSearch.keyword = searchForm.keyword;
    appliedSearch.apiGroup = searchForm.apiGroup;
    appliedSearch.method = searchForm.method;
    appliedSearch.serviceName = searchForm.serviceName;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    Object.assign(searchForm, {
      keyword: '',
      apiGroup: '',
      method: '',
      serviceName: '',
    });
    Object.assign(appliedSearch, {
      keyword: '',
      apiGroup: '',
      method: '',
      serviceName: '',
    });
    fetchData(1, pagination.pageSize);
  };

  const onPageChange = (current: number) => {
    fetchData(current, pagination.pageSize);
  };

  const onPageSizeChange = (pageSize: number) => {
    fetchData(1, pageSize);
  };

  const onModalVisibleChange = (visible: boolean) => {
    modalVisible.value = visible;
    if (!visible) editingApi.value = null;
  };

  const handleCreate = () => {
    editingApi.value = null;
    modalVisible.value = true;
  };

  const handleEdit = (record: ApiRow) => {
    editingApi.value = record.raw;
    modalVisible.value = true;
  };

  const handleDelete = (record: ApiRow) => {
    Modal.confirm({
      title: t('front.common.delete.title'),
      content: t('apiManage.delete.content', {
        description: record.trans_description,
        method: record.method,
        path: record.path,
      }),
      okText: t('front.common.delete.okText'),
      okButtonProps: { status: 'danger' },
      async onOk() {
        deletingId.value = record.id;
        try {
          await deleteApi({ ids: [record.id] });
          Message.success(t('front.common.message.deleteSuccess'));
          const nextPage =
            apis.value.length <= 1 && pagination.current > 1
              ? pagination.current - 1
              : pagination.current;
          await fetchData(nextPage, pagination.pageSize);
        } finally {
          deletingId.value = null;
        }
      },
    });
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
  :deep(.arco-pagination-options .arco-select) {
    min-width: 100px;
  }
</style>
