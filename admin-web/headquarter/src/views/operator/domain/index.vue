<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.outstation', 'operatorDomain.title']" />
    <div class="page-header">
      <div class="page-title">{{ t('operatorDomain.title') }}</div>
      <div class="page-subtitle">{{ t('operatorDomain.subtitle') }}</div>
    </div>
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item
          :label="t('operatorDomain.search.operator')"
          field="operatorId"
        >
          <a-select
            v-model="searchForm.operatorId"
            :placeholder="t('operatorDomain.search.operatorPlaceholder')"
            :loading="operatorOptionsLoading"
            allow-clear
            allow-search
            :filter-option="false"
            class="search-select"
            :options="operatorOptions"
            @search="handleOperatorSearch"
          />
        </a-form-item>
        <a-form-item
          :label="t('operatorDomain.search.keyword')"
          field="keyword"
        >
          <a-input
            v-model="searchForm.keyword"
            :placeholder="t('operatorDomain.search.keywordPlaceholder')"
            allow-clear
            :max-length="253"
            class="search-input"
            @press-enter="handleSearch"
          />
        </a-form-item>
        <a-form-item
          :label="t('operatorDomain.search.domainType')"
          field="domainType"
        >
          <a-select
            v-model="searchForm.domainType"
            :placeholder="t('front.common.all')"
            allow-clear
            class="search-select"
            :options="domainTypeOptions"
          />
        </a-form-item>
        <a-form-item :label="t('operatorDomain.search.status')" field="status">
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
    <a-card class="general-card table-card" :bordered="false">
      <div class="table-toolbar">
        <a-space>
          <a-button
            v-if="hasPermission('operatorDomain:create')"
            type="primary"
            @click="handleCreate"
          >
            <template #icon><icon-plus /></template>
            {{ t('operatorDomain.toolbar.create') }}
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
        <template #domainType="{ record }">
          <a-tag>{{ domainTypeLabel((record as DomainRow).domainType) }}</a-tag>
        </template>
        <template #status="{ record }">
          <span class="status-dot-text">
            <span
              class="status-dot"
              :class="
                (record as DomainRow).status === OPERATOR_DOMAIN_STATUS.ENABLED
                  ? 'status-dot--enabled'
                  : 'status-dot--disabled'
              "
            />
            {{
              (record as DomainRow).status === OPERATOR_DOMAIN_STATUS.ENABLED
                ? t('front.common.status.enabled')
                : t('front.common.status.disabled')
            }}
          </span>
        </template>
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <!-- 暂无 detail 接口，先隐藏 -->
            <!-- <a-link @click="handleDetail(record as DomainRow)">
              {{ t('operatorDomain.action.detail') }}
            </a-link> -->
            <a-link
              :disabled="publishingId === (record as DomainRow).id"
              @click="handlePublish(record as DomainRow)"
            >
              {{ t('operatorDomain.action.publish') }}
            </a-link>
            <a-link
              v-if="hasPermission('operatorDomain:update')"
              @click="handleEdit(record as DomainRow)"
            >
              {{ t('operatorDomain.action.edit') }}
            </a-link>
            <!-- 暂无 toggle 接口，先隐藏 -->
            <!-- <a-link
              :disabled="togglingId === (record as DomainRow).id"
              @click="handleToggleStatus(record as DomainRow)"
            >
              {{
                (record as DomainRow).status === OPERATOR_DOMAIN_STATUS.ENABLED
                  ? t('operatorDomain.action.disable')
                  : t('operatorDomain.action.enable')
              }}
            </a-link> -->
            <a-link
              v-if="hasPermission('operatorDomain:delete')"
              status="danger"
              :disabled="deletingId === (record as DomainRow).id"
              @click="handleDelete(record as DomainRow)"
            >
              {{ t('operatorDomain.action.delete') }}
            </a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>

    <a-modal
      v-model:visible="modalVisible"
      :title="
        editingId ? t('operatorDomain.modal.editTitle') : t('operatorDomain.modal.createTitle')
      "
      :width="480"
      :ok-loading="submitting"
      unmount-on-close
      @before-ok="handleModalOk"
      @cancel="modalVisible = false"
    >
      <a-form ref="formRef" :model="form" :rules="rules">
        <a-form-item
          :label="t('operatorDomain.field.operator')"
          field="operatorId"
          required
        >
          <a-select
            v-model="form.operatorId"
            :placeholder="t('operatorDomain.field.operatorPlaceholder')"
            :disabled="!!editingId"
            :loading="operatorOptionsLoading"
            allow-search
            :filter-option="false"
            :options="operatorOptions"
            @search="handleOperatorSearch"
          />
        </a-form-item>
        <a-form-item
          :label="t('operatorDomain.field.domainType')"
          field="domainType"
          required
        >
          <a-select v-model="form.domainType" :options="domainTypeOptions" />
        </a-form-item>
        <a-form-item
          :label="t('operatorDomain.field.domainName')"
          field="domainName"
          required
        >
          <a-input
            v-model="form.domainName"
            :placeholder="t('operatorDomain.field.domainNamePlaceholder')"
            allow-clear
            :max-length="253"
          />
        </a-form-item>
        <a-form-item :label="t('operatorDomain.field.status')" field="status">
          <a-radio-group v-model="form.status">
            <a-radio :value="OPERATOR_DOMAIN_STATUS.ENABLED">
              {{ t('front.common.status.enabled') }}
            </a-radio>
            <a-radio :value="OPERATOR_DOMAIN_STATUS.DISABLED">
              {{ t('front.common.status.disabled') }}
            </a-radio>
          </a-radio-group>
        </a-form-item>
        <a-form-item :label="t('operatorDomain.field.remark')" field="remark">
          <a-textarea
            v-model="form.remark"
            :placeholder="t('operatorDomain.field.remarkPlaceholder')"
            :max-length="1000"
            :auto-size="{ minRows: 2, maxRows: 4 }"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal
      v-model:visible="detailVisible"
      :title="t('operatorDomain.modal.detailTitle')"
      :width="560"
      :footer="false"
      unmount-on-close
      @cancel="detailVisible = false"
    >
      <a-spin :loading="detailLoading" style="width: 100%">
        <a-descriptions
          v-if="detailRecord"
          :column="1"
          bordered
          size="medium"
          table-layout="fixed"
        >
          <a-descriptions-item :label="t('operatorDomain.column.id')">
            {{ detailRecord.id }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('operatorDomain.column.operator')">
            {{ operatorLabel(detailRecord.operator_id) }}
            （ID: {{ detailRecord.operator_id }}）
          </a-descriptions-item>
          <a-descriptions-item :label="t('operatorDomain.column.domainName')">
            {{ detailRecord.domain_name || '--' }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('operatorDomain.column.domainType')">
            {{ domainTypeLabel(detailRecord.domain_type) }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('operatorDomain.column.status')">
            <span class="status-dot-text">
              <span
                class="status-dot"
                :class="
                  detailRecord.status === OPERATOR_DOMAIN_STATUS.ENABLED
                    ? 'status-dot--enabled'
                    : 'status-dot--disabled'
                "
              />
              {{
                detailRecord.status === OPERATOR_DOMAIN_STATUS.ENABLED
                  ? t('front.common.status.enabled')
                  : t('front.common.status.disabled')
              }}
            </span>
          </a-descriptions-item>
          <a-descriptions-item :label="t('operatorDomain.column.remark')">
            {{ detailRecord.remark || '--' }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('front.common.column.createdAt')">
            {{ formatTime(detailRecord.created_at) }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('operatorDomain.column.updatedAt')">
            {{ formatTime(detailRecord.updated_at) }}
          </a-descriptions-item>
        </a-descriptions>
      </a-spin>
    </a-modal>
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, Modal } from '@arco-design/web-vue';
  import type { FieldRule, FormInstance } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import dayjs from 'dayjs';
  import {
    createOperatorDomain,
    deleteOperatorDomain,
    getOperator,
    getOperatorDomain,
    OPERATOR_DOMAIN_STATUS,
    OPERATOR_DOMAIN_TYPE,
    publishOperatorDomain,
    queryOperatorDomainList,
    queryOperatorList,
    updateOperatorDomain,
    type OperatorDomainListParams,
    type OperatorDomainRecord,
    type OperatorRecord,
  } from '@/api/operator';
  import useLoading from '@/hooks/loading';
  import usePermission from '@/hooks/permission';

  interface DomainRow {
    id: number;
    operatorId: number;
    operatorName: string;
    domainName: string;
    domainType: number;
    status: number;
    remark: string;
    createdAt: string;
    updatedAt: string;
    raw: OperatorDomainRecord;
  }

  const { t } = useI18n();
  const { hasPermission } = usePermission();
  const { loading, setLoading } = useLoading(true);

  const domainTypeOptions = computed(() => [
    { label: t('operatorDomain.type.admin'), value: OPERATOR_DOMAIN_TYPE.ADMIN },
    { label: t('operatorDomain.type.agent'), value: OPERATOR_DOMAIN_TYPE.AGENT },
    { label: t('operatorDomain.type.h5'), value: OPERATOR_DOMAIN_TYPE.H5 },
  ]);
  const statusOptions = computed(() => [
    {
      label: t('front.common.status.enabled'),
      value: OPERATOR_DOMAIN_STATUS.ENABLED,
    },
    {
      label: t('front.common.status.disabled'),
      value: OPERATOR_DOMAIN_STATUS.DISABLED,
    },
  ]);
  const domainTypeLabel = (type: number) =>
    domainTypeOptions.value.find((item) => item.value === type)?.label ||
    String(type);

  /** 分站下拉选项，key 为分站名 map，用于列表展示所属分站 */
  const operatorOptions = ref<{ label: string; value: number }[]>([]);
  const operatorNameMap = ref<Map<number, string>>(new Map());
  const operatorOptionsLoading = ref(false);

  const loadOperatorOptions = async (keyword?: string) => {
    operatorOptionsLoading.value = true;
    try {
      const { data } = await queryOperatorList({
        page: 1,
        page_size: 100,
        ...(keyword ? { keyword } : {}),
      });
      const list = data?.list || [];
      operatorOptions.value = list.map((item: OperatorRecord) => ({
        label: `${item.name}（${item.code}）`,
        value: item.id,
      }));
      list.forEach((item: OperatorRecord) => {
        operatorNameMap.value.set(item.id, item.name);
      });
    } catch {
      operatorOptions.value = [];
    } finally {
      operatorOptionsLoading.value = false;
    }
  };

  let operatorSearchTimer: ReturnType<typeof setTimeout> | null = null;
  const handleOperatorSearch = (keyword: string) => {
    if (operatorSearchTimer) clearTimeout(operatorSearchTimer);
    operatorSearchTimer = setTimeout(() => {
      loadOperatorOptions(keyword.trim() || undefined);
    }, 300);
  };

  const operatorLabel = (id: number) =>
    operatorNameMap.value.get(id) || `#${id}`;

  const searchForm = reactive<{
    operatorId: number | undefined;
    keyword: string;
    domainType: number | undefined;
    status: number | undefined;
  }>({
    operatorId: undefined,
    keyword: '',
    domainType: undefined,
    status: undefined,
  });
  const appliedSearch = reactive<{
    operatorId: number | undefined;
    keyword: string;
    domainType: number | undefined;
    status: number | undefined;
  }>({
    operatorId: undefined,
    keyword: '',
    domainType: undefined,
    status: undefined,
  });

  const list = ref<DomainRow[]>([]);
  const togglingId = ref<number | null>(null);
  const deletingId = ref<number | null>(null);
  const publishingId = ref<number | null>(null);

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

  const mapRow = (item: OperatorDomainRecord): DomainRow => ({
    id: item.id,
    operatorId: item.operator_id,
    operatorName: operatorLabel(item.operator_id),
    domainName: item.domain_name || '--',
    domainType: item.domain_type,
    status: item.status,
    remark: item.remark || '--',
    createdAt: formatTime(item.created_at),
    updatedAt: formatTime(item.updated_at),
    raw: item,
  });

  const buildQueryParams = (
    page = pagination.current,
    pageSize = pagination.pageSize
  ): OperatorDomainListParams => {
    const params: OperatorDomainListParams = {
      page,
      page_size: pageSize,
    };
    if (appliedSearch.operatorId !== undefined) {
      params.operator_id = appliedSearch.operatorId;
    }
    const keyword = appliedSearch.keyword.trim();
    if (keyword) params.keyword = keyword;
    if (appliedSearch.domainType !== undefined) {
      params.domain_type = appliedSearch.domainType;
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
      const { data } = await queryOperatorDomainList(
        buildQueryParams(page, pageSize)
      );
      const rawList = data?.list || [];
      // 补齐列表中出现、但尚未加载进下拉选项的分站名称
      const missingIds = Array.from(
        new Set(
          rawList
            .map((item) => item.operator_id)
            .filter((id) => !operatorNameMap.value.has(id))
        )
      );
      if (missingIds.length) {
        await Promise.all(
          missingIds.map(async (id) => {
            try {
              const { data: operatorData } = await getOperator({ id });
              if (operatorData) operatorNameMap.value.set(id, operatorData.name);
            } catch {
              // ignore, fallback to #id display
            }
          })
        );
      }
      list.value = rawList.map(mapRow);
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
      title: t('operatorDomain.column.id'),
      dataIndex: 'id',
      width: 90,
      align: 'center',
    },
    {
      title: t('operatorDomain.column.operator'),
      dataIndex: 'operatorName',
      width: 200,
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('operatorDomain.column.domainName'),
      dataIndex: 'domainName',
      width: 240,
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('operatorDomain.column.domainType'),
      dataIndex: 'domainType',
      slotName: 'domainType',
      width: 130,
      align: 'center',
    },
    {
      title: t('operatorDomain.column.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 110,
      align: 'center',
    },
    {
      title: t('operatorDomain.column.remark'),
      dataIndex: 'remark',
      width: 200,
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('front.common.column.createdAt'),
      dataIndex: 'createdAt',
      width: 170,
      align: 'center',
    },
    {
      title: t('operatorDomain.column.updatedAt'),
      dataIndex: 'updatedAt',
      width: 170,
      align: 'center',
    },
    {
      title: t('front.common.column.operations'),
      dataIndex: 'operations',
      slotName: 'operations',
      width: 260,
      align: 'center',
      fixed: 'right',
    },
  ]);

  const handleSearch = () => {
    appliedSearch.operatorId = searchForm.operatorId;
    appliedSearch.keyword = searchForm.keyword;
    appliedSearch.domainType = searchForm.domainType;
    appliedSearch.status = searchForm.status;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    Object.assign(searchForm, {
      operatorId: undefined,
      keyword: '',
      domainType: undefined,
      status: undefined,
    });
    Object.assign(appliedSearch, {
      operatorId: undefined,
      keyword: '',
      domainType: undefined,
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

  // ---------------- 新增 / 编辑弹窗 ----------------
  const modalVisible = ref(false);
  const submitting = ref(false);
  const editingId = ref<number | null>(null);
  const formRef = ref<FormInstance>();

  const getDefaultForm = () => ({
    operatorId: undefined as number | undefined,
    domainType: OPERATOR_DOMAIN_TYPE.ADMIN as number,
    domainName: '',
    status: OPERATOR_DOMAIN_STATUS.ENABLED as number,
    remark: '',
  });
  const form = reactive(getDefaultForm());

  const rules = computed<Record<string, FieldRule | FieldRule[]>>(() => ({
    operatorId: [
      {
        required: true,
        message: t('operatorDomain.rule.operatorRequired'),
      },
    ],
    domainType: [
      {
        required: true,
        message: t('operatorDomain.rule.domainTypeRequired'),
      },
    ],
    domainName: [
      {
        required: true,
        message: t('operatorDomain.rule.domainNameRequired'),
      },
    ],
  }));

  const handleCreate = () => {
    editingId.value = null;
    Object.assign(form, getDefaultForm());
    modalVisible.value = true;
  };

  const fillFormFromRecord = (record: OperatorDomainRecord) => {
    Object.assign(form, {
      operatorId: record.operator_id,
      domainType: record.domain_type,
      domainName: record.domain_name || '',
      status: record.status,
      remark: record.remark || '',
    });
  };

  const handleEdit = async (record: DomainRow) => {
    editingId.value = record.id;
    fillFormFromRecord(record.raw);
    modalVisible.value = true;
    try {
      const { data } = await getOperatorDomain({ id: record.id });
      if (data) fillFormFromRecord(data);
    } catch {
      // 详情拉取失败时保留列表行数据，仍可编辑
    }
  };

  // ---------------- 详情弹窗 ----------------
  const detailVisible = ref(false);
  const detailLoading = ref(false);
  const detailRecord = ref<OperatorDomainRecord | null>(null);

  const handleDetail = async (record: DomainRow) => {
    detailRecord.value = record.raw;
    detailVisible.value = true;
    detailLoading.value = true;
    try {
      const { data } = await getOperatorDomain({ id: record.id });
      if (data) {
        detailRecord.value = data;
        if (!operatorNameMap.value.has(data.operator_id)) {
          try {
            const { data: operatorData } = await getOperator({
              id: data.operator_id,
            });
            if (operatorData) {
              operatorNameMap.value.set(data.operator_id, operatorData.name);
            }
          } catch {
            // ignore
          }
        }
      }
    } catch {
      // 保留列表行兜底展示
    } finally {
      detailLoading.value = false;
    }
  };

  const handleModalOk = async (done: (closed: boolean) => void) => {
    const errors = await formRef.value?.validate();
    if (errors) {
      done(false);
      return;
    }
    if (submitting.value) {
      done(false);
      return;
    }

    const domainName = form.domainName.trim();
    const remark = form.remark.trim();

    submitting.value = true;
    try {
      if (editingId.value) {
        await updateOperatorDomain({
          id: editingId.value,
          domain_name: domainName,
          domain_type: form.domainType,
          status: form.status,
          remark,
        });
        Message.success(t('operatorDomain.message.updateSuccess'));
        await fetchData(pagination.current, pagination.pageSize);
      } else {
        await createOperatorDomain({
          operator_id: form.operatorId as number,
          domain_name: domainName,
          domain_type: form.domainType,
          status: form.status,
          ...(remark ? { remark } : {}),
        });
        Message.success(t('operatorDomain.message.createSuccess'));
        await fetchData(1, pagination.pageSize);
      }
      done(true);
    } catch {
      done(false);
    } finally {
      submitting.value = false;
    }
  };

  const handleToggleStatus = (record: DomainRow) => {
    const nextStatus =
      record.status === OPERATOR_DOMAIN_STATUS.ENABLED
        ? OPERATOR_DOMAIN_STATUS.DISABLED
        : OPERATOR_DOMAIN_STATUS.ENABLED;
    togglingId.value = record.id;
    updateOperatorDomain({ id: record.id, status: nextStatus })
      .then(() => {
        Message.success(t('operatorDomain.message.statusUpdateSuccess'));
        return fetchData(pagination.current, pagination.pageSize);
      })
      .finally(() => {
        togglingId.value = null;
      });
  };

  const handlePublish = (record: DomainRow) => {
    Modal.confirm({
      title: t('operatorDomain.message.publishConfirmTitle'),
      content: t('operatorDomain.message.publishConfirmContent', {
        domain: record.domainName,
      }),
      async onOk() {
        publishingId.value = record.id;
        try {
          await publishOperatorDomain({ id: record.id });
          Message.success(t('operatorDomain.message.publishSuccess'));
          await fetchData(pagination.current, pagination.pageSize);
        } finally {
          publishingId.value = null;
        }
      },
    });
  };

  const handleDelete = (record: DomainRow) => {
    Modal.confirm({
      title: t('front.common.delete.title'),
      content: t('operatorDomain.message.deleteConfirmContent', {
        domain: record.domainName,
      }),
      okText: t('front.common.delete.okText'),
      okButtonProps: { status: 'danger' },
      async onOk() {
        deletingId.value = record.id;
        try {
          await deleteOperatorDomain({ id: record.id });
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

  loadOperatorOptions();
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
  .search-select {
    min-width: 180px;
  }
  .search-select :deep(.arco-select-view) {
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
  .status-dot--enabled {
    background-color: rgb(var(--success-6));
  }
  .status-dot--disabled {
    background-color: rgb(var(--gray-5));
  }
</style>
