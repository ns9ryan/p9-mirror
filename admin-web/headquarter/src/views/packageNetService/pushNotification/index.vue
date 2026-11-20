<template>
  <div class="container list-page">
    <Breadcrumb />
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item field="keyword">
          <a-input-group class="keyword-group">
            <a-select
              v-model="searchForm.keywordField"
              :options="keywordFieldOptions"
              class="keyword-field"
            />
            <a-input
              v-model="searchForm.keyword"
              :placeholder="t('pushNotification.search.keywordPlaceholder')"
              allow-clear
              :max-length="100"
              class="keyword-input"
              @press-enter="handleSearch"
            />
          </a-input-group>
        </a-form-item>
        <a-form-item :label="t('pushNotification.search.type')" field="type">
          <a-select
            v-model="searchForm.type"
            :placeholder="t('front.common.all')"
            allow-clear
            class="search-select"
            :options="typeOptions"
          />
        </a-form-item>
        <a-form-item
          :label="t('pushNotification.search.status')"
          field="status"
        >
          <a-select
            v-model="searchForm.status"
            :placeholder="t('front.common.all')"
            allow-clear
            class="search-select"
            :options="statusOptions"
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

    <a-card class="general-card" :bordered="false">
      <div class="table-toolbar">
        <a-button type="primary" @click="createVisible = true">
          <template #icon><icon-plus /></template>
          {{ t('pushNotification.toolbar.create') }}
        </a-button>
      </div>
      <a-table
        row-key="id"
        :loading="loading"
        :columns="columns"
        :data="list"
        :pagination="pagination"
        :bordered="true"
        :stripe="true"
        :table-layout-fixed="true"
        :scroll="{ x: 1500 }"
        @page-change="onPageChange"
        @page-size-change="onPageSizeChange"
      >
        <template #titleCell="{ record }">
          <a-tooltip
            :content="record.title"
            content-class="push-notification-tooltip"
          >
            <span class="cell-ellipsis">{{ record.title || '--' }}</span>
          </a-tooltip>
        </template>
        <template #contentCell="{ record }">
          <a-tooltip
            :content="record.content"
            content-class="push-notification-tooltip"
          >
            <span
              class="cell-ellipsis cell-ellipsis--link"
              @click="openDetail(record)"
            >
              {{ record.content || '--' }}
            </span>
          </a-tooltip>
        </template>
        <template #type="{ record }">
          {{ typeLabel(record.type) }}
        </template>
        <template #targets="{ record }">
          <div v-if="record.target_operators?.length" class="target-cell">
            <a-tag
              v-for="item in record.target_operators.slice(0, TARGET_PREVIEW)"
              :key="item.id"
              size="small"
              class="target-tag"
            >
              {{ item.name }}
            </a-tag>
            <a-popover
              v-if="record.target_operators.length > TARGET_PREVIEW"
              :title="
                t('pushNotification.column.targetsCount', {
                  count: record.target_operators.length,
                })
              "
              content-class="push-notification-targets-popover"
            >
              <a-tag size="small" color="arcoblue" class="target-more">
                +{{ record.target_operators.length - TARGET_PREVIEW }}
              </a-tag>
              <template #content>
                <div class="targets-popover-list">
                  <a-tag
                    v-for="item in record.target_operators"
                    :key="item.id"
                    size="small"
                  >
                    {{ item.name }}
                  </a-tag>
                </div>
              </template>
            </a-popover>
          </div>
          <span v-else>--</span>
        </template>
        <template #createdAt="{ record }">
          {{ formatTime(record.created_at) }}
        </template>
        <template #updatedAt="{ record }">
          {{ formatTime(record.updated_at) }}
        </template>
        <template #updatedBy="{ record }">
          {{ record.updated_by || '--' }}
        </template>
        <template #status="{ record }">
          <span class="status-dot-text">
            <span class="status-dot" :class="statusMeta(record.status).dot" />
            {{ statusMeta(record.status).label }}
          </span>
        </template>
        <template #operations="{ record }">
          <a-link @click="openDetail(record)">
            {{ t('pushNotification.action.detail') }}
          </a-link>
        </template>
      </a-table>
    </a-card>

    <CreateModal v-model:visible="createVisible" @success="fetchData(1)" />

    <a-modal
      v-model:visible="detailVisible"
      :title="t('pushNotification.detail.title')"
      :width="720"
      :footer="false"
      unmount-on-close
    >
      <template v-if="detailRecord">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item :label="t('pushNotification.column.id')">
            {{ detailRecord.id }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('pushNotification.column.type')">
            {{ typeLabel(detailRecord.type) }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('pushNotification.column.status')">
            <span class="status-dot-text">
              <span
                class="status-dot"
                :class="statusMeta(detailRecord.status).dot"
              />
              {{ statusMeta(detailRecord.status).label }}
            </span>
          </a-descriptions-item>
          <a-descriptions-item :label="t('pushNotification.detail.validity')">
            {{ t('pushNotification.validity.permanent') }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('pushNotification.column.createdAt')">
            {{ formatTime(detailRecord.created_at) }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('pushNotification.column.updatedAt')">
            {{ formatTime(detailRecord.updated_at) }}
          </a-descriptions-item>
          <a-descriptions-item
            :label="t('pushNotification.column.targets')"
            :span="2"
          >
            <div class="detail-targets">
              <a-tag
                v-for="item in detailRecord.target_operators || []"
                :key="item.id"
                size="small"
              >
                {{ item.name }}
              </a-tag>
            </div>
          </a-descriptions-item>
        </a-descriptions>

        <a-tabs
          v-if="detailContents.length > 1"
          type="card-gutter"
          class="detail-tabs"
        >
          <a-tab-pane
            v-for="item in detailContents"
            :key="item.language_code"
            :title="item.language_code"
          >
            <div class="detail-title">{{ item.title }}</div>
            <div class="detail-content">{{ item.content }}</div>
          </a-tab-pane>
        </a-tabs>
        <div v-else class="detail-block">
          <div class="detail-title">{{ detailContents[0]?.title }}</div>
          <div class="detail-content">{{ detailContents[0]?.content }}</div>
        </div>
      </template>
    </a-modal>
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import dayjs from 'dayjs';
  import {
    NOTIFICATION_STATUS,
    NOTIFICATION_TYPE,
    queryNotificationList,
    type NotificationKeywordField,
    type NotificationListParams,
    type NotificationRecord,
  } from '@/api/notification';
  import useLoading from '@/hooks/loading';
  import CreateModal from './components/create-modal.vue';

  const TARGET_PREVIEW = 2;

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(true);

  const keywordFieldOptions = computed(() => [
    {
      label: t('pushNotification.search.keywordField.title'),
      value: 'title',
    },
    {
      label: t('pushNotification.search.keywordField.content'),
      value: 'content',
    },
  ]);

  const typeOptions = computed(() => [
    {
      label: t('pushNotification.type.official'),
      value: NOTIFICATION_TYPE.OFFICIAL,
    },
  ]);

  const statusOptions = computed(() => [
    {
      label: t('pushNotification.status.pending'),
      value: NOTIFICATION_STATUS.PENDING,
    },
    {
      label: t('pushNotification.status.sent'),
      value: NOTIFICATION_STATUS.SENT,
    },
    {
      label: t('pushNotification.status.failed'),
      value: NOTIFICATION_STATUS.FAILED,
    },
  ]);

  const typeLabel = (type: number) =>
    typeOptions.value.find((item) => item.value === type)?.label || '--';

  const statusMeta = (status: number) => {
    switch (status) {
      case NOTIFICATION_STATUS.SENT:
        return {
          label: t('pushNotification.status.sent'),
          dot: 'status-dot--sent',
        };
      case NOTIFICATION_STATUS.FAILED:
        return {
          label: t('pushNotification.status.failed'),
          dot: 'status-dot--failed',
        };
      default:
        return {
          label: t('pushNotification.status.pending'),
          dot: 'status-dot--pending',
        };
    }
  };

  interface SearchState {
    keywordField: NotificationKeywordField;
    keyword: string;
    type: number | undefined;
    status: number | undefined;
  }

  const createSearchState = (): SearchState => ({
    keywordField: 'title',
    keyword: '',
    type: undefined,
    status: undefined,
  });

  const searchForm = reactive<SearchState>(createSearchState());
  const appliedSearch = reactive<SearchState>(createSearchState());

  const list = ref<NotificationRecord[]>([]);
  const pagination = reactive({
    current: 1,
    pageSize: 10,
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

  const buildQueryParams = (
    page: number,
    pageSize: number
  ): NotificationListParams => {
    const params: NotificationListParams = { page, page_size: pageSize };
    const keyword = appliedSearch.keyword.trim();
    if (keyword) {
      params.keyword = keyword;
      params.keyword_field = appliedSearch.keywordField;
    }
    if (appliedSearch.type !== undefined) params.type = appliedSearch.type;
    if (appliedSearch.status !== undefined) {
      params.status = appliedSearch.status;
    }
    return params;
  };

  const fetchData = async (
    page = pagination.current,
    pageSize = pagination.pageSize
  ) => {
    setLoading(true);
    try {
      const { data } = await queryNotificationList(
        buildQueryParams(page, pageSize)
      );
      list.value = data?.list || [];
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
      title: t('pushNotification.column.id'),
      dataIndex: 'id',
      width: 80,
      align: 'center',
    },
    {
      title: t('pushNotification.column.title'),
      dataIndex: 'title',
      slotName: 'titleCell',
      width: 200,
    },
    {
      title: t('pushNotification.column.content'),
      dataIndex: 'content',
      slotName: 'contentCell',
      width: 280,
    },
    {
      title: t('pushNotification.column.type'),
      dataIndex: 'type',
      slotName: 'type',
      width: 110,
      align: 'center',
    },
    {
      title: t('pushNotification.column.targets'),
      dataIndex: 'target_operators',
      slotName: 'targets',
      width: 240,
    },
    {
      title: t('pushNotification.column.createdAt'),
      dataIndex: 'created_at',
      slotName: 'createdAt',
      width: 170,
      align: 'center',
    },
    {
      title: t('pushNotification.column.updatedAt'),
      dataIndex: 'updated_at',
      slotName: 'updatedAt',
      width: 170,
      align: 'center',
    },
    {
      title: t('pushNotification.column.updatedBy'),
      dataIndex: 'updated_by',
      slotName: 'updatedBy',
      width: 120,
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('pushNotification.column.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 110,
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
  ]);

  const handleSearch = () => {
    Object.assign(appliedSearch, { ...searchForm });
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

  const createVisible = ref(false);

  const detailVisible = ref(false);
  const detailRecord = ref<NotificationRecord | null>(null);
  const detailContents = computed(() => {
    const record = detailRecord.value;
    if (!record) return [];
    const titles = record.title_i18n || {};
    const contents = record.content_i18n || {};
    const codes = [
      ...new Set([...Object.keys(titles), ...Object.keys(contents)]),
    ];
    if (!codes.length) {
      return [
        { language_code: '', title: record.title, content: record.content },
      ];
    }
    return codes.map((code) => ({
      language_code: code,
      title: titles[code] || '--',
      content: contents[code] || '--',
    }));
  });

  const openDetail = (record: NotificationRecord) => {
    detailRecord.value = record;
    detailVisible.value = true;
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

  .keyword-group {
    display: flex;
  }

  .keyword-field {
    width: 96px;
  }

  .keyword-input {
    width: 220px;
  }

  .search-select :deep(.arco-select-view) {
    min-width: 130px;
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

  .cell-ellipsis {
    display: block;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;

    &--link {
      cursor: pointer;

      &:hover {
        color: rgb(var(--primary-6));
      }
    }
  }

  .target-cell {
    display: flex;
    flex-wrap: nowrap;
    gap: 4px;
    align-items: center;
    min-width: 0;
  }

  .target-tag {
    flex-shrink: 1;
    min-width: 0;
    max-width: 90px;

    :deep(.arco-tag-content) {
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
    }
  }

  .target-more {
    flex-shrink: 0;
    cursor: pointer;
  }

  .status-dot-text {
    display: inline-flex;
    gap: 6px;
    align-items: center;
  }

  .status-dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }

  .status-dot--sent {
    background-color: rgb(var(--success-6));
  }

  .status-dot--pending {
    background-color: rgb(var(--warning-6));
  }

  .status-dot--failed {
    background-color: rgb(var(--danger-6));
  }

  :deep(.arco-pagination-options .arco-select) {
    min-width: 100px;
  }

  .detail-targets {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    max-height: 120px;
    overflow-y: auto;
  }

  .detail-tabs,
  .detail-block {
    margin-top: 16px;
  }

  .detail-tabs :deep(.arco-tabs-content) {
    padding-top: 12px;
  }

  .detail-title {
    margin-bottom: 8px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 15px;
    word-break: break-word;
  }

  .detail-content {
    max-height: 320px;
    padding: 12px;
    overflow-y: auto;
    color: var(--color-text-2);
    line-height: 1.7;
    white-space: pre-wrap;
    word-break: break-word;
    background-color: var(--color-fill-1);
    border-radius: 6px;
  }
</style>

<style lang="less">
  .push-notification-tooltip {
    max-width: 420px;
    max-height: 240px;
    overflow-y: auto;
    white-space: pre-wrap;
    word-break: break-word;
  }

  .push-notification-targets-popover {
    max-width: 360px;
  }

  .push-notification-targets-popover .targets-popover-list {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    max-height: 200px;
    overflow-y: auto;
  }
</style>
