<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.outstation', 'menu.outstation.stationManage']" />
    <div class="page-header">
      <div class="page-title">{{ t('operatorStation.title') }}</div>
      <div class="page-subtitle">{{ t('operatorStation.subtitle') }}</div>
    </div>
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item
          :label="t('operatorStation.search.keyword')"
          field="keyword"
        >
          <a-input
            v-model="searchForm.keyword"
            :placeholder="t('operatorStation.search.keywordPlaceholder')"
            allow-clear
            :max-length="100"
            class="search-input"
            @press-enter="handleSearch"
          />
        </a-form-item>
        <a-form-item
          :label="t('operatorStation.search.creationStatus')"
          field="creationStatus"
        >
          <a-select
            v-model="searchForm.creationStatus"
            :placeholder="t('front.common.all')"
            allow-clear
            class="search-select"
            :options="creationStatusOptions"
          />
        </a-form-item>
        <a-form-item
          :label="t('operatorStation.search.publishStatus')"
          field="publishStatus"
        >
          <a-select
            v-model="searchForm.publishStatus"
            :placeholder="t('front.common.all')"
            allow-clear
            class="search-select"
            :options="publishStatusOptions"
          />
        </a-form-item>
        <a-form-item :label="t('operatorStation.search.status')" field="status">
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
      <a-alert type="info" :show-icon="true" class="search-tip">
        {{ t('operatorStation.tip.title')
        }}{{ t('operatorStation.tip.content') }}
      </a-alert>
    </a-card>
    <a-card class="general-card table-card" :bordered="false">
      <div class="table-toolbar">
        <a-space>
          <a-button
            v-if="hasPermission('operator:create')"
            type="primary"
            @click="handleCreate"
          >
            <template #icon><icon-plus /></template>
            {{ t('operatorStation.toolbar.create') }}
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
        :scroll="{ x: 1660 }"
        @page-change="onPageChange"
        @page-size-change="onPageSizeChange"
      >
        <template #creationStatus="{ record }">
          <a-tag :color="creationStatusMeta(record.creationStatus).color">
            {{ creationStatusMeta(record.creationStatus).label }}
          </a-tag>
        </template>
        <template #status="{ record }">
          <span class="status-dot-text">
            <span
              class="status-dot"
              :class="statusMeta(record.status).dotClass"
            />
            {{ statusMeta(record.status).label }}
          </span>
        </template>
        <template #publishStatus="{ record }">
          <span class="status-dot-text">
            <span
              class="status-dot"
              :class="publishStatusMeta(record.publishStatus).dotClass"
            />
            {{ publishStatusMeta(record.publishStatus).label }}
          </span>
        </template>
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <a-link @click="handleDetail(record)">
              {{ t('operatorStation.action.detail') }}
            </a-link>
            <a-link
              v-if="hasPermission('operator:update')"
              @click="handleEdit(record)"
            >
              {{ t('operatorStation.action.edit') }}
            </a-link>
            <!-- 暂无 complete 接口，先隐藏 -->
            <!-- <a-link
              v-if="record.creationStatus === OPERATOR_CREATION_STATUS.DRAFT"
              :disabled="completingId === record.id"
              @click="handleComplete(record)"
            >
              {{ t('operatorStation.action.complete') }}
            </a-link> -->
            <a-link
              v-if="hasPermission('operator:publish') && canPublish(record)"
              :disabled="publishingId === record.id"
              @click="handlePublish(record)"
            >
              {{ t('operatorStation.action.publish') }}
            </a-link>
            <a-link
              v-if="
                hasPermission('operator:publish') &&
                record.publishStatus === OPERATOR_PUBLISH_STATUS.PUBLISHING
              "
              :disabled="syncingPublishId === record.id"
              @click="handleSyncPublishStatus(record)"
            >
              {{ t('operatorStation.action.syncPublishStatus') }}
            </a-link>
            <a-link
              v-if="hasPermission('operator:delete') && record.deletable"
              status="danger"
              :disabled="deletingId === record.id"
              @click="handleDelete(record)"
            >
              {{ t('operatorStation.action.delete') }}
            </a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>
    <CreateStationModal
      v-model:visible="createModalVisible"
      :resume-record="resumeRecord"
      @success="onMutationSuccess"
      @update:visible="onCreateModalVisibleChange"
    />
    <StationDetailModal
      v-model:visible="detailModalVisible"
      :station-id="detailStationId"
    />
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, Modal } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import dayjs from 'dayjs';
  import { NODE_STATUS } from '@/api/node';
  import {
    completeOperator,
    deleteOperator,
    getOperatorNode,
    isOperatorDeletable,
    OPERATOR_CREATION_STATUS,
    OPERATOR_DOMAIN_TYPE,
    OPERATOR_PUBLISH_STATUS,
    OPERATOR_STATUS,
    publishOperator,
    queryOperatorAdminList,
    queryOperatorAgentLineAllocation,
    queryOperatorDomainList,
    queryOperatorLanguageAllocation,
    queryOperatorList,
    queryOperatorRegionAllocation,
    syncOperatorPublishStatus,
    type OperatorListParams,
    type OperatorRecord,
  } from '@/api/operator';
  import useLoading from '@/hooks/loading';
  import usePermission from '@/hooks/permission';
  import CreateStationModal from './components/create-station-modal.vue';
  import StationDetailModal from './components/station-detail-modal.vue';

  interface StationRow {
    id: number;
    code: string;
    name: string;
    timezone: string;
    currency: string;
    creationStatus: number;
    publishStatus: number;
    status: number;
    createdAt: string;
    deletable: boolean;
    raw: OperatorRecord;
  }

  const { t, locale } = useI18n();
  const { hasPermission } = usePermission();
  const { loading, setLoading } = useLoading(true);

  const creationStatusOptions = computed(() => [
    {
      label: t('operatorStation.creationStatus.draft'),
      value: OPERATOR_CREATION_STATUS.DRAFT,
    },
    {
      label: t('operatorStation.creationStatus.completed'),
      value: OPERATOR_CREATION_STATUS.COMPLETED,
    },
  ]);
  const publishStatusOptions = computed(() => [
    {
      label: t('operatorStation.publishStatus.unpublished'),
      value: OPERATOR_PUBLISH_STATUS.UNPUBLISHED,
    },
    {
      label: t('operatorStation.publishStatus.publishing'),
      value: OPERATOR_PUBLISH_STATUS.PUBLISHING,
    },
    {
      label: t('operatorStation.publishStatus.published'),
      value: OPERATOR_PUBLISH_STATUS.PUBLISHED,
    },
    {
      label: t('operatorStation.publishStatus.failed'),
      value: OPERATOR_PUBLISH_STATUS.FAILED,
    },
  ]);
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

  const creationStatusMeta = (status: number) => {
    if (status === OPERATOR_CREATION_STATUS.COMPLETED) {
      return {
        label: t('operatorStation.creationStatus.completed'),
        color: 'green',
      };
    }
    return { label: t('operatorStation.creationStatus.draft'), color: 'gray' };
  };

  const publishStatusMeta = (status: number) => {
    switch (status) {
      case OPERATOR_PUBLISH_STATUS.PUBLISHING:
        return {
          label: t('operatorStation.publishStatus.publishing'),
          dotClass: 'publish-dot--publishing',
        };
      case OPERATOR_PUBLISH_STATUS.PUBLISHED:
        return {
          label: t('operatorStation.publishStatus.published'),
          dotClass: 'publish-dot--published',
        };
      case OPERATOR_PUBLISH_STATUS.FAILED:
        return {
          label: t('operatorStation.publishStatus.failed'),
          dotClass: 'publish-dot--failed',
        };
      default:
        return {
          label: t('operatorStation.publishStatus.unpublished'),
          dotClass: 'publish-dot--unpublished',
        };
    }
  };

  const statusMeta = (status: number) => {
    switch (status) {
      case OPERATOR_STATUS.PAUSED:
        return {
          label: t('operatorStation.status.paused'),
          dotClass: 'status-dot--paused',
        };
      case OPERATOR_STATUS.CLOSED:
        return {
          label: t('operatorStation.status.closed'),
          dotClass: 'status-dot--closed',
        };
      default:
        return {
          label: t('operatorStation.status.normal'),
          dotClass: 'status-dot--normal',
        };
    }
  };

  /** 草稿状态先完成创建，发布中 / 已发布的分站不再重复发布 */
  const canPublish = (record: StationRow) =>
    record.creationStatus === OPERATOR_CREATION_STATUS.COMPLETED &&
    (record.publishStatus === OPERATOR_PUBLISH_STATUS.UNPUBLISHED ||
      record.publishStatus === OPERATOR_PUBLISH_STATUS.FAILED);

  const searchForm = reactive<{
    keyword: string;
    creationStatus: number | undefined;
    publishStatus: number | undefined;
    status: number | undefined;
  }>({
    keyword: '',
    creationStatus: undefined,
    publishStatus: undefined,
    status: undefined,
  });
  const appliedSearch = reactive<{
    keyword: string;
    creationStatus: number | undefined;
    publishStatus: number | undefined;
    status: number | undefined;
  }>({
    keyword: '',
    creationStatus: undefined,
    publishStatus: undefined,
    status: undefined,
  });

  const list = ref<StationRow[]>([]);
  const completingId = ref<number | null>(null);
  const publishingId = ref<number | null>(null);
  const syncingPublishId = ref<number | null>(null);
  const deletingId = ref<number | null>(null);

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

  const mapRow = (item: OperatorRecord): StationRow => ({
    id: item.id,
    code: item.code || '--',
    name: item.name || '--',
    timezone: item.timezone_code || '--',
    currency: item.settlement_currency_code || '--',
    creationStatus: item.creation_status,
    publishStatus: item.publish_status,
    status: item.status,
    createdAt: formatTime(item.created_at),
    deletable: isOperatorDeletable(item),
    raw: item,
  });

  const buildQueryParams = (
    page = pagination.current,
    pageSize = pagination.pageSize
  ): OperatorListParams => {
    const params: OperatorListParams = {
      page,
      page_size: pageSize,
    };
    const keyword = appliedSearch.keyword.trim();
    if (keyword) params.keyword = keyword;
    if (appliedSearch.creationStatus !== undefined) {
      params.creation_status = appliedSearch.creationStatus;
    }
    if (appliedSearch.publishStatus !== undefined) {
      params.publish_status = appliedSearch.publishStatus;
    }
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
      const { data } = await queryOperatorList(
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
      title: t('operatorStation.column.id'),
      dataIndex: 'id',
      width: 100,
      align: 'center',
    },
    {
      title: t('operatorStation.column.code'),
      dataIndex: 'code',
      width: 260,
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('operatorStation.column.name'),
      dataIndex: 'name',
      width: 140,
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('operatorStation.column.timezone'),
      dataIndex: 'timezone',
      width: 160,
      align: 'center',
    },
    {
      title: t('operatorStation.column.currency'),
      dataIndex: 'currency',
      width: 110,
      align: 'center',
    },
    {
      title: t('operatorStation.column.creationStatus'),
      dataIndex: 'creationStatus',
      slotName: 'creationStatus',
      width: 110,
      align: 'center',
    },
    {
      title: t('operatorStation.column.publishStatus'),
      dataIndex: 'publishStatus',
      slotName: 'publishStatus',
      width: 120,
      align: 'center',
    },
    {
      title: t('operatorStation.column.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 120,
      align: 'center',
    },
    {
      title: t('front.common.column.createdAt'),
      dataIndex: 'createdAt',
      width: 180,
      align: 'center',
    },
    {
      title: t('front.common.column.operations'),
      dataIndex: 'operations',
      slotName: 'operations',
      width: 320,
      align: 'center',
      fixed: 'right',
    },
  ]);

  const handleSearch = () => {
    appliedSearch.keyword = searchForm.keyword;
    appliedSearch.creationStatus = searchForm.creationStatus;
    appliedSearch.publishStatus = searchForm.publishStatus;
    appliedSearch.status = searchForm.status;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    Object.assign(searchForm, {
      keyword: '',
      creationStatus: undefined,
      publishStatus: undefined,
      status: undefined,
    });
    Object.assign(appliedSearch, {
      keyword: '',
      creationStatus: undefined,
      publishStatus: undefined,
      status: undefined,
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
  /** 草稿续编 / 已完成编辑时传入创建向导；新建时为 null */
  const resumeRecord = ref<OperatorRecord | null>(null);
  const detailModalVisible = ref(false);
  const detailStationId = ref<number | null>(null);

  const handleCreate = () => {
    resumeRecord.value = null;
    createModalVisible.value = true;
  };

  /** 新建分站会落在列表首位（按创建时间倒序），回到第一页查看 */
  const onMutationSuccess = () => {
    fetchData(1, pagination.pageSize);
  };

  const onCreateModalVisibleChange = (visible: boolean) => {
    createModalVisible.value = visible;
    if (!visible) resumeRecord.value = null;
  };

  const handleDetail = (record: StationRow) => {
    detailStationId.value = record.id;
    detailModalVisible.value = true;
  };

  /** 草稿与已完成均进入多步向导，可编辑全部步骤 */
  const handleEdit = (record: StationRow) => {
    resumeRecord.value = record.raw;
    createModalVisible.value = true;
  };

  const handleComplete = (record: StationRow) => {
    Modal.confirm({
      title: t('operatorStation.message.completeConfirmTitle'),
      content: t('operatorStation.message.completeConfirmContent', {
        name: record.name,
      }),
      async onOk() {
        completingId.value = record.id;
        try {
          await completeOperator({ id: record.id });
          Message.success(t('operatorStation.message.completeSuccess'));
          await fetchData(pagination.current, pagination.pageSize);
        } finally {
          completingId.value = null;    
        }
      },
    });
  };

  /**
   * 发布前核对创建向导里的硬性条件：管理员、三类域名、三类基础资源、
   * 以及已启用且在线的部署节点。缺任一项都不调用发布接口。
   */
  const collectPublishGaps = async (record: StationRow) => {
    const gaps: string[] = [];
    const operatorId = record.id;
    const domainQuery = (domainType: number) =>
      queryOperatorDomainList({
        page: 1,
        page_size: 1,
        operator_id: operatorId,
        domain_type: domainType,
      });
    const [
      adminRes,
      languageRes,
      regionRes,
      agentRes,
      nodeRes,
      h5DomainRes,
      agentDomainRes,
      adminDomainRes,
    ] = await Promise.all([
      queryOperatorAdminList({
        page: 1,
        page_size: 1,
        operator_id: operatorId,
      }),
      queryOperatorLanguageAllocation({ operator_id: operatorId }),
      queryOperatorRegionAllocation({ operator_id: operatorId }),
      queryOperatorAgentLineAllocation({ operator_id: operatorId }),
      getOperatorNode({ operator_id: operatorId }),
      domainQuery(OPERATOR_DOMAIN_TYPE.H5),
      domainQuery(OPERATOR_DOMAIN_TYPE.AGENT),
      domainQuery(OPERATOR_DOMAIN_TYPE.ADMIN),
    ]);

    if (!adminRes.data?.list?.length) {
      gaps.push(t('operatorStation.message.publishGap.admin'));
    }
    const hasDomain = (total?: number, listLength?: number) =>
      (total ?? 0) > 0 || (listLength ?? 0) > 0;
    if (!hasDomain(h5DomainRes.data?.total, h5DomainRes.data?.list?.length)) {
      gaps.push(t('operatorStation.message.publishGap.domainH5'));
    }
    if (
      !hasDomain(agentDomainRes.data?.total, agentDomainRes.data?.list?.length)
    ) {
      gaps.push(t('operatorStation.message.publishGap.domainAgent'));
    }
    if (
      !hasDomain(adminDomainRes.data?.total, adminDomainRes.data?.list?.length)
    ) {
      gaps.push(t('operatorStation.message.publishGap.domainAdmin'));
    }
    if (!languageRes.data?.list?.length) {
      gaps.push(t('operatorStation.message.publishGap.language'));
    }
    if (!regionRes.data?.list?.length) {
      gaps.push(t('operatorStation.message.publishGap.region'));
    }
    const hasAgentLine = (agentRes.data?.list || []).some(
      (item) => item.allocated
    );
    if (!hasAgentLine) {
      gaps.push(t('operatorStation.message.publishGap.proxyLine'));
    }
    const node = nodeRes.data?.operator_node;
    const nodeReady =
      Boolean(node?.node_code) &&
      node?.node_status === NODE_STATUS.ENABLED &&
      node?.node_online === true;
    if (!nodeReady) {
      gaps.push(t('operatorStation.message.publishGap.deployNode'));
    }
    return gaps;
  };

  const handlePublish = async (record: StationRow) => {
    if (publishingId.value != null) return;
    publishingId.value = record.id;
    let gaps: string[] = [];
    try {
      gaps = await collectPublishGaps(record);
    } catch {
      return;
    } finally {
      publishingId.value = null;
    }

    if (gaps.length) {
      const separator = String(locale.value).toLowerCase().startsWith('zh')
        ? '、'
        : ', ';
      Modal.warning({
        title: t('operatorStation.message.publishNotReadyTitle'),
        content: t('operatorStation.message.publishNotReadyContent', {
          name: record.name,
          items: gaps.join(separator),
        }),
        okText: t('operatorStation.message.publishNotReadyAction'),
        hideCancel: false,
        onOk() {
          handleEdit(record);
        },
      });
      return;
    }

    Modal.confirm({
      title: t('operatorStation.message.publishConfirmTitle'),
      content: `${t('operatorStation.message.publishConfirmContent')}`,
      async onOk() {
        publishingId.value = record.id;
        try {
          await publishOperator({ id: record.id });
          // 后端发布逻辑仍是 TODO，不能据接口返回判断发布成功
          Message.info(t('operatorStation.message.publishSubmitted'));
          await fetchData(pagination.current, pagination.pageSize);
        } finally {
          publishingId.value = null;
        }
      },
    });
  };

  const handleSyncPublishStatus = async (record: StationRow) => {
    syncingPublishId.value = record.id;
    try {
      const { data } = await syncOperatorPublishStatus({ id: record.id });
      const status = data?.publish_status ?? record.publishStatus;
      Message.success(
        t('operatorStation.message.syncPublishStatusSuccess', {
          status: publishStatusMeta(status).label,
        })
      );
      await fetchData(pagination.current, pagination.pageSize);
    } finally {
      syncingPublishId.value = null;
    }
  };

  const handleDelete = (record: StationRow) => {
    if (!record.deletable) {
      Message.warning(t('operatorStation.message.deleteNotAllowed'));
      return;
    }
    Modal.confirm({
      title: t('front.common.delete.title'),
      content: t('operatorStation.message.deleteConfirmContent', {
        name: record.name,
      }),
      okText: t('front.common.delete.okText'),
      okButtonProps: { status: 'danger' },
      async onOk() {
        deletingId.value = record.id;
        try {
          await deleteOperator({ id: record.id });
          Message.success(t('front.common.message.deleteSuccess'));
          const nextPage =
            list.value.length <= 1 && pagination.current > 1
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
  .search-tip {
    margin-top: 12px;
  }
  .search-tip :deep(.arco-alert-body) {
    padding-top: 0;
    padding-bottom: 0;
  }
  .search-tip :deep(.arco-alert-content) {
    font-size: 13px;
    line-height: 1.5;
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
  .status-dot--normal {
    background-color: rgb(var(--success-6));
  }
  .status-dot--paused {
    background-color: rgb(var(--warning-6));
  }
  .status-dot--closed {
    background-color: rgb(var(--gray-5));
  }
  .publish-dot--unpublished {
    background-color: rgb(var(--gray-5));
  }
  .publish-dot--publishing {
    background-color: rgb(var(--arcoblue-6));
  }
  .publish-dot--published {
    background-color: rgb(var(--success-6));
  }
  .publish-dot--failed {
    background-color: rgb(var(--danger-6));
  }
</style>
