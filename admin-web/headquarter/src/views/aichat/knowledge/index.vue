<template>
  <div class="container list-page">
    <div class="page-header">
      <div class="page-title">{{ $t('aichatKnowledge.title') }}</div>
      <div class="page-subtitle">
        {{ $t('aichatKnowledge.subtitle') }}
      </div>
    </div>
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item :label="$t('aichatKnowledge.search.keyword')" field="keyword">
          <a-input
            v-model="searchForm.keyword"
            :placeholder="$t('aichatKnowledge.search.keywordPlaceholder')"
            allow-clear
            :max-length="100"
            class="search-input"
            @press-enter="handleSearch"
          />
        </a-form-item>
        <a-form-item
          :label="$t('aichatKnowledge.column.documentType')"
          field="document_type"
        >
          <a-select
            v-model="searchForm.document_type"
            :placeholder="$t('front.common.all')"
            allow-clear
            class="search-select"
            :options="documentTypeOptions"
          />
        </a-form-item>
        <a-form-item
          :label="$t('aichatKnowledge.column.sourceTable')"
          field="source_table"
        >
          <a-input
            v-model="searchForm.source_table"
            :placeholder="$t('aichatKnowledge.search.sourceTablePlaceholder')"
            allow-clear
            class="search-input search-input--short"
            @press-enter="handleSearch"
          />
        </a-form-item>
        <a-form-item :label="$t('front.common.column.status')" field="status">
          <a-select
            v-model="searchForm.status"
            :placeholder="$t('front.common.all')"
            allow-clear
            class="search-select"
            :options="statusFilterOptions"
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
          <a-button type="primary" @click="createModalVisible = true">
            <template #icon><icon-plus /></template>
            {{ $t('aichatKnowledge.toolbar.create') }}
          </a-button>
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
        :data="list"
        :pagination="pagination"
        :bordered="true"
        :stripe="true"
        :scroll="{ x: 1500 }"
        @page-change="onPageChange"
        @page-size-change="onPageSizeChange"
      >
        <template #title="{ record }">
          <a-link class="title-link" @click="openDetail(record)">
            {{ record.title || '--' }}
          </a-link>
        </template>
        <template #documentType="{ record }">
          <a-tag :color="documentTypeMeta(record.document_type).color">
            {{ documentTypeMeta(record.document_type).label }}
          </a-tag>
        </template>
        <template #status="{ record }">
          <a-tag :color="statusMeta(record.status).color">
            {{ statusMeta(record.status).label }}
          </a-tag>
        </template>
        <template #embed="{ record }">
          <span v-if="record.embed_model">
            {{ record.embed_model }}
            <span v-if="record.embed_dim" class="text-muted">
              / {{ record.embed_dim }}
            </span>
          </span>
          <span v-else>--</span>
        </template>
        <template #createdAt="{ record }">
          {{ formatTime(record.created_at) }}
        </template>
        <template #updatedAt="{ record }">
          {{ formatTime(record.updated_at) }}
        </template>
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <a-link @click="openDetail(record)">
              {{ $t('aichatKnowledge.action.detail') }}
            </a-link>
            <a-link status="danger" @click="handleDelete(record)">
              {{ $t('front.common.action.delete') }}
            </a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>

    <CreateDocumentModal
      v-model:visible="createModalVisible"
      @success="() => fetchData(1, pagination.pageSize)"
    />

    <a-drawer
      v-model:visible="detailVisible"
      :width="720"
      :footer="false"
      :title="$t('aichatKnowledge.detail.title')"
      unmount-on-close
    >
      <template v-if="detailRecord">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="ID">
            {{ detailRecord.id }}
          </a-descriptions-item>
          <a-descriptions-item :label="$t('aichatKnowledge.detail.documentCode')">
            {{ detailRecord.document_code || '--' }}
          </a-descriptions-item>
          <a-descriptions-item
            :label="$t('aichatKnowledge.column.title')"
            :span="2"
          >
            {{ detailRecord.title || '--' }}
          </a-descriptions-item>
          <a-descriptions-item :label="$t('aichatKnowledge.column.documentType')">
            <a-tag :color="documentTypeMeta(detailRecord.document_type).color">
              {{ documentTypeMeta(detailRecord.document_type).label }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item :label="$t('front.common.column.status')">
            <a-tag :color="statusMeta(detailRecord.status).color">
              {{ statusMeta(detailRecord.status).label }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item :label="$t('aichatKnowledge.column.sourceTable')">
            {{ detailRecord.source_table || '--' }}
          </a-descriptions-item>
          <a-descriptions-item :label="$t('aichatKnowledge.column.version')">
            {{ detailRecord.version ?? '--' }}
          </a-descriptions-item>
          <a-descriptions-item :label="$t('aichatKnowledge.column.chunkCount')">
            {{ detailRecord.chunk_count ?? '--' }}
          </a-descriptions-item>
          <a-descriptions-item :label="$t('aichatKnowledge.column.embed')">
            {{ detailRecord.embed_model || '--' }}
            <template v-if="detailRecord.embed_dim">
              / {{ detailRecord.embed_dim }}
            </template>
          </a-descriptions-item>
          <a-descriptions-item
            :label="$t('aichatKnowledge.column.description')"
            :span="2"
          >
            {{ detailRecord.description || '--' }}
          </a-descriptions-item>
          <a-descriptions-item
            :label="$t('aichatKnowledge.detail.contentHash')"
            :span="2"
          >
            <span class="hash-text">{{ detailRecord.content_hash || '--' }}</span>
          </a-descriptions-item>
          <a-descriptions-item :label="$t('front.common.column.createdAt')">
            {{ formatTime(detailRecord.created_at) }}
          </a-descriptions-item>
          <a-descriptions-item :label="$t('front.common.column.updatedAt')">
            {{ formatTime(detailRecord.updated_at) }}
          </a-descriptions-item>
        </a-descriptions>
        <div class="detail-content-title">
          {{ $t('aichatKnowledge.detail.content') }}
        </div>
        <pre class="detail-content">{{ detailRecord.content || '--' }}</pre>
      </template>
    </a-drawer>
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, Modal } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import dayjs from 'dayjs';
  import {
    deleteKnowledgeDocument,
    queryKnowledgeDocuments,
    type KnowledgeDocument,
    type KnowledgeDocumentStatus,
    type KnowledgeDocumentType,
    type KnowledgeDocumentsParams,
  } from '@/api/aichat';
  import useLoading from '@/hooks/loading';
  import CreateDocumentModal from './components/create-document-modal.vue';
  import {
    documentTypeMeta,
    getDocumentTypeOptions,
    getStatusFilterOptions,
    statusMeta,
  } from './constants';

  interface SearchState {
    keyword: string;
    document_type: KnowledgeDocumentType | undefined;
    source_table: string;
    status: KnowledgeDocumentStatus | undefined;
  }

  const createSearchState = (): SearchState => ({
    keyword: '',
    document_type: undefined,
    source_table: '',
    status: undefined,
  });

  const { t } = useI18n();

  const documentTypeOptions = computed(() => getDocumentTypeOptions());
  const statusFilterOptions = computed(() => getStatusFilterOptions());

  const { loading, setLoading } = useLoading(true);

  const searchForm = reactive<SearchState>(createSearchState());
  const appliedSearch = reactive<SearchState>(createSearchState());

  const list = ref<KnowledgeDocument[]>([]);

  const pagination = reactive({
    current: 1,
    pageSize: 20,
    total: 0,
    showTotal: true,
    showPageSize: true,
    pageSizeOptions: [10, 20, 50, 100],
    pageSizeProps: {
      popupContainer: 'body',
    },
  });

  const formatTime = (value?: number | null) => {
    if (!value) return '--';
    const ms = value < 1e12 ? value * 1000 : value;
    return dayjs(ms).format('YYYY-MM-DD HH:mm:ss');
  };

  const buildQueryParams = (
    page: number,
    pageSize: number
  ): KnowledgeDocumentsParams => {
    const params: KnowledgeDocumentsParams = {
      limit: pageSize,
      offset: (page - 1) * pageSize,
    };
    const keyword = appliedSearch.keyword.trim();
    const sourceTable = appliedSearch.source_table.trim();
    if (keyword) params.keyword = keyword;
    if (sourceTable) params.source_table = sourceTable;
    if (appliedSearch.document_type) {
      params.document_type = appliedSearch.document_type;
    }
    if (appliedSearch.status) params.status = appliedSearch.status;
    return params;
  };

  const fetchData = async (
    page = pagination.current,
    pageSize = pagination.pageSize
  ) => {
    setLoading(true);
    try {
      const { data } = await queryKnowledgeDocuments(
        buildQueryParams(page, pageSize)
      );
      list.value = data?.documents || [];
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
    { title: 'ID', dataIndex: 'id', width: 90, align: 'center' },
    {
      title: t('aichatKnowledge.column.title'),
      dataIndex: 'title',
      slotName: 'title',
      width: 220,
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('aichatKnowledge.column.documentType'),
      dataIndex: 'document_type',
      slotName: 'documentType',
      width: 120,
      align: 'center',
    },
    {
      title: t('aichatKnowledge.column.sourceTable'),
      dataIndex: 'source_table',
      width: 140,
      align: 'center',
      ellipsis: true,
      tooltip: true,
      render: ({ record }) => record.source_table || '--',
    },
    {
      title: t('aichatKnowledge.column.description'),
      dataIndex: 'description',
      width: 220,
      ellipsis: true,
      tooltip: true,
      render: ({ record }) => record.description || '--',
    },
    {
      title: t('front.common.column.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 100,
      align: 'center',
    },
    {
      title: t('aichatKnowledge.column.version'),
      dataIndex: 'version',
      width: 80,
      align: 'center',
    },
    {
      title: t('aichatKnowledge.column.chunkCount'),
      dataIndex: 'chunk_count',
      width: 90,
      align: 'center',
    },
    {
      title: t('aichatKnowledge.column.embed'),
      dataIndex: 'embed_model',
      slotName: 'embed',
      width: 220,
      align: 'center',
    },
    {
      title: t('front.common.column.createdAt'),
      dataIndex: 'created_at',
      slotName: 'createdAt',
      width: 180,
      align: 'center',
    },
    {
      title: t('front.common.column.updatedAt'),
      dataIndex: 'updated_at',
      slotName: 'updatedAt',
      width: 180,
      align: 'center',
    },
    {
      title: t('front.common.column.operations'),
      dataIndex: 'operations',
      slotName: 'operations',
      width: 130,
      align: 'center',
      fixed: 'right',
    },
  ]);

  const handleSearch = () => {
    Object.assign(appliedSearch, searchForm);
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    Object.assign(searchForm, createSearchState());
    Object.assign(appliedSearch, createSearchState());
    fetchData(1, pagination.pageSize);
  };

  const onPageChange = (current: number) => {
    fetchData(current, pagination.pageSize);
  };

  const onPageSizeChange = (pageSize: number) => {
    fetchData(1, pageSize);
  };

  const createModalVisible = ref(false);

  const detailVisible = ref(false);
  const detailRecord = ref<KnowledgeDocument | null>(null);

  const openDetail = (record: KnowledgeDocument) => {
    detailRecord.value = record;
    detailVisible.value = true;
  };

  const handleDelete = (record: KnowledgeDocument) => {
    Modal.confirm({
      title: t('aichatKnowledge.delete.title'),
      content: t('aichatKnowledge.delete.content', {
        name: record.title || record.id,
      }),
      okText: t('front.common.delete.okText'),
      okButtonProps: { status: 'danger' },
      async onOk() {
        const { data } = await deleteKnowledgeDocument(record.id);
        if (data && data.deleted === false) {
          Message.error(data.message || t('aichatKnowledge.message.deleteFail'));
          return;
        }
        Message.success(t('front.common.message.deleteSuccess'));
        const isLastOnPage = list.value.length === 1 && pagination.current > 1;
        fetchData(
          isLastOnPage ? pagination.current - 1 : pagination.current,
          pagination.pageSize
        );
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
  .search-input--short {
    width: 160px;
  }
  .search-select :deep(.arco-select-view) {
    min-width: 140px;
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
  .operation-links {
    white-space: nowrap;
  }
  .title-link {
    max-width: 100%;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
    display: inline-block;
  }
  .text-muted {
    color: var(--color-text-3);
  }
  .hash-text {
    font-family: Consolas, Monaco, monospace;
    word-break: break-all;
  }
  .detail-content-title {
    margin: 20px 0 8px;
    font-weight: 600;
    color: var(--color-text-1);
  }
  .detail-content {
    margin: 0;
    padding: 12px 14px;
    max-height: 480px;
    overflow: auto;
    font-family: Consolas, Monaco, monospace;
    font-size: 13px;
    line-height: 1.6;
    color: var(--color-text-2);
    white-space: pre-wrap;
    word-break: break-word;
    background-color: var(--color-fill-2);
    border-radius: 4px;
  }
</style>
