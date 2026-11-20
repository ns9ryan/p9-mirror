<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.outstation', 'operatorAdmin.title']" />
    <div class="page-header">
      <div class="page-title">{{ t('operatorAdmin.title') }}</div>
      <div class="page-subtitle">{{ t('operatorAdmin.subtitle') }}</div>
    </div>
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item
          :label="t('operatorAdmin.search.operator')"
          field="operatorId"
        >
          <a-select
            v-model="searchForm.operatorId"
            :placeholder="t('operatorAdmin.search.operatorPlaceholder')"
            :loading="operatorOptionsLoading"
            allow-clear
            allow-search
            :filter-option="false"
            class="search-select"
            :options="operatorOptions"
            @search="handleOperatorSearch"
          />
        </a-form-item>
        <a-form-item :label="t('operatorAdmin.search.keyword')" field="keyword">
          <a-input
            v-model="searchForm.keyword"
            :placeholder="t('operatorAdmin.search.keywordPlaceholder')"
            allow-clear
            :max-length="100"
            class="search-input"
            @press-enter="handleSearch"
          />
        </a-form-item>
        <a-form-item :label="t('operatorAdmin.search.status')" field="status">
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
            v-if="hasPermission('operatorAdmin:create')"
            type="primary"
            @click="handleCreate"
          >
            <template #icon><icon-plus /></template>
            {{ t('operatorAdmin.toolbar.create') }}
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
        :scroll="{ x: 1300 }"
        @page-change="onPageChange"
        @page-size-change="onPageSizeChange"
      >
        <template #status="{ record }">
          <span class="status-dot-text">
            <span
              class="status-dot"
              :class="
                (record as AdminRow).status === OPERATOR_ADMIN_STATUS.ENABLED
                  ? 'status-dot--enabled'
                  : 'status-dot--disabled'
              "
            />
            {{
              (record as AdminRow).status === OPERATOR_ADMIN_STATUS.ENABLED
                ? t('front.common.status.enabled')
                : t('front.common.status.disabled')
            }}
          </span>
        </template>
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <a-link
              v-if="hasPermission('operatorAdmin:update')"
              @click="handleEdit(record as AdminRow)"
            >
              {{ t('operatorAdmin.action.edit') }}
            </a-link>
            <a-link
              v-if="hasPermission('operatorAdmin:resetPassword')"
              @click="handleResetPassword(record as AdminRow)"
            >
              {{ t('operatorAdmin.action.resetPassword') }}
            </a-link>
            <a-link
              v-if="hasPermission('operatorAdmin:updateStatus')"
              :disabled="togglingId === (record as AdminRow).id"
              @click="handleToggleStatus(record as AdminRow)"
            >
              {{
                (record as AdminRow).status === OPERATOR_ADMIN_STATUS.ENABLED
                  ? t('operatorAdmin.action.disable')
                  : t('operatorAdmin.action.enable')
              }}
            </a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>

    <a-modal
      v-model:visible="modalVisible"
      :title="
        editingId
          ? t('operatorAdmin.modal.editTitle')
          : t('operatorAdmin.modal.createTitle')
      "
      :width="480"
      :ok-loading="submitting"
      unmount-on-close
      @before-ok="handleModalOk"
      @cancel="modalVisible = false"
    >
      <a-form ref="formRef" :model="form" :rules="rules">
        <a-form-item
          :label="t('operatorAdmin.field.operator')"
          field="operatorId"
          required
        >
          <a-select
            v-model="form.operatorId"
            :placeholder="t('operatorAdmin.field.operatorPlaceholder')"
            :disabled="!!editingId"
            :loading="operatorOptionsLoading"
            allow-search
            :filter-option="false"
            :options="operatorOptions"
            @search="handleOperatorSearch"
          />
        </a-form-item>
        <a-form-item
          :label="t('operatorAdmin.field.username')"
          field="username"
          required
        >
          <a-input
            v-model="form.username"
            :placeholder="t('operatorAdmin.field.usernamePlaceholder')"
            allow-clear
            :max-length="64"
          />
        </a-form-item>
        <a-form-item
          v-if="!editingId"
          :label="t('operatorAdmin.field.password')"
          field="password"
          required
        >
          <a-input-password
            v-model="form.password"
            :placeholder="t('operatorAdmin.field.passwordPlaceholder')"
            allow-clear
            :max-length="32"
          />
        </a-form-item>
        <a-form-item
          v-if="!editingId"
          :label="t('operatorAdmin.field.confirmPassword')"
          field="confirmPassword"
          required
        >
          <a-input-password
            v-model="form.confirmPassword"
            :placeholder="t('operatorAdmin.field.confirmPasswordPlaceholder')"
            allow-clear
            :max-length="32"
          />
        </a-form-item>
        <a-form-item
          :label="t('operatorAdmin.field.displayName')"
          field="displayName"
          required
        >
          <a-input
            v-model="form.displayName"
            :placeholder="t('operatorAdmin.field.displayNamePlaceholder')"
            allow-clear
            :max-length="100"
          />
        </a-form-item>
        <a-form-item :label="t('operatorAdmin.field.status')" field="status">
          <a-radio-group v-model="form.status">
            <a-radio :value="OPERATOR_ADMIN_STATUS.ENABLED">
              {{ t('front.common.status.enabled') }}
            </a-radio>
            <a-radio :value="OPERATOR_ADMIN_STATUS.DISABLED">
              {{ t('front.common.status.disabled') }}
            </a-radio>
          </a-radio-group>
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal
      v-model:visible="resetPasswordVisible"
      :title="t('operatorAdmin.modal.resetPasswordTitle')"
      :width="480"
      :ok-loading="resetPasswordSubmitting"
      unmount-on-close
      @before-ok="handleResetPasswordOk"
      @cancel="resetPasswordVisible = false"
    >
      <a-alert type="info" :show-icon="true" class="reset-password-tip">
        {{
          t('operatorAdmin.modal.resetPasswordTip', {
            username: resetPasswordTarget?.username || '--',
          })
        }}
      </a-alert>
      <a-form
        ref="resetPasswordFormRef"
        :model="resetPasswordForm"
        :rules="resetPasswordRules"
      >
        <a-form-item
          :label="t('operatorAdmin.field.password')"
          field="password"
          required
        >
          <a-input-password
            v-model="resetPasswordForm.password"
            :placeholder="t('operatorAdmin.field.passwordPlaceholder')"
            allow-clear
            :max-length="32"
          />
        </a-form-item>
        <a-form-item
          :label="t('operatorAdmin.field.confirmPassword')"
          field="confirmPassword"
          required
        >
          <a-input-password
            v-model="resetPasswordForm.confirmPassword"
            :placeholder="t('operatorAdmin.field.confirmPasswordPlaceholder')"
            allow-clear
            :max-length="32"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FieldRule, FormInstance } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import dayjs from 'dayjs';
  import {
    createOperatorAdmin,
    OPERATOR_ADMIN_STATUS,
    queryOperatorAdminList,
    queryOperatorList,
    resetOperatorAdminPassword,
    updateOperatorAdmin,
    updateOperatorAdminStatus,
    type OperatorAdminListParams,
    type OperatorAdminRecord,
    type OperatorRecord,
  } from '@/api/operator';
  import useLoading from '@/hooks/loading';
  import usePermission from '@/hooks/permission';

  interface AdminRow {
    id: number;
    operatorId: number;
    operatorName: string;
    username: string;
    displayName: string;
    status: number;
    createdAt: string;
    updatedAt: string;
    raw: OperatorAdminRecord;
  }

  const { t } = useI18n();
  const { hasPermission } = usePermission();
  const { loading, setLoading } = useLoading(true);

  const statusOptions = computed(() => [
    {
      label: t('front.common.status.enabled'),
      value: OPERATOR_ADMIN_STATUS.ENABLED,
    },
    {
      label: t('front.common.status.disabled'),
      value: OPERATOR_ADMIN_STATUS.DISABLED,
    },
  ]);

  const operatorOptions = ref<{ label: string; value: number }[]>([]);
  const operatorOptionsLoading = ref(false);

  const loadOperatorOptions = async (keyword?: string) => {
    operatorOptionsLoading.value = true;
    try {
      const { data } = await queryOperatorList({
        page: 1,
        page_size: 100,
        ...(keyword ? { keyword } : {}),
      });
      operatorOptions.value = (data?.list || []).map(
        (item: OperatorRecord) => ({
          label: `${item.name}（${item.code}）`,
          value: item.id,
        })
      );
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

  const searchForm = reactive<{
    operatorId: number | undefined;
    keyword: string;
    status: number | undefined;
  }>({
    operatorId: undefined,
    keyword: '',
    status: undefined,
  });
  const appliedSearch = reactive<{
    operatorId: number | undefined;
    keyword: string;
    status: number | undefined;
  }>({
    operatorId: undefined,
    keyword: '',
    status: undefined,
  });

  const list = ref<AdminRow[]>([]);
  const togglingId = ref<number | null>(null);

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

  const mapRow = (item: OperatorAdminRecord): AdminRow => ({
    id: item.id,
    operatorId: item.operator_id,
    operatorName: item.operator_name || `--（#${item.operator_id}）`,
    username: item.username || '--',
    displayName: item.display_name || '--',
    status: item.status,
    createdAt: formatTime(item.created_at),
    updatedAt: formatTime(item.updated_at),
    raw: item,
  });

  const buildQueryParams = (
    page = pagination.current,
    pageSize = pagination.pageSize
  ): OperatorAdminListParams => {
    const params: OperatorAdminListParams = {
      page,
      page_size: pageSize,
    };
    if (appliedSearch.operatorId !== undefined) {
      params.operator_id = appliedSearch.operatorId;
    }
    const keyword = appliedSearch.keyword.trim();
    if (keyword) params.keyword = keyword;
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
      const { data } = await queryOperatorAdminList(
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
      title: t('operatorAdmin.column.id'),
      dataIndex: 'id',
      width: 100,
      align: 'center',
    },
    {
      title: t('operatorAdmin.column.operator'),
      dataIndex: 'operatorName',
      width: 180,
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('operatorAdmin.column.username'),
      dataIndex: 'username',
      width: 160,
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('operatorAdmin.column.displayName'),
      dataIndex: 'displayName',
      width: 160,
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('operatorAdmin.column.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 110,
      align: 'center',
    },
    {
      title: t('front.common.column.createdAt'),
      dataIndex: 'createdAt',
      width: 170,
      align: 'center',
    },
    {
      title: t('operatorAdmin.column.updatedAt'),
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
    appliedSearch.status = searchForm.status;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    Object.assign(searchForm, {
      operatorId: undefined,
      keyword: '',
      status: undefined,
    });
    Object.assign(appliedSearch, {
      operatorId: undefined,
      keyword: '',
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
    username: '',
    password: '',
    confirmPassword: '',
    displayName: '',
    status: OPERATOR_ADMIN_STATUS.ENABLED as number,
  });
  const form = reactive(getDefaultForm());

  const rules = computed<Record<string, FieldRule | FieldRule[]>>(() => ({
    operatorId: [
      {
        required: true,
        message: t('operatorAdmin.rule.operatorRequired'),
      },
    ],
    username: [
      {
        required: true,
        message: t('operatorAdmin.rule.usernameRequired'),
      },
    ],
    password: [
      {
        validator: (value, callback) => {
          if (editingId.value) {
            callback();
            return;
          }
          const password = String(value || '');
          if (!password) {
            callback(t('operatorAdmin.rule.passwordRequired'));
            return;
          }
          if (password.length < 6 || password.length > 32) {
            callback(t('operatorAdmin.rule.passwordLength'));
            return;
          }
          callback();
        },
      },
    ],
    confirmPassword: [
      {
        validator: (value, callback) => {
          if (editingId.value) {
            callback();
            return;
          }
          const confirmPassword = String(value || '');
          if (!confirmPassword) {
            callback(t('operatorAdmin.rule.confirmPasswordRequired'));
            return;
          }
          if (confirmPassword !== form.password) {
            callback(t('operatorAdmin.rule.passwordNotMatch'));
            return;
          }
          callback();
        },
      },
    ],
    displayName: [
      {
        required: true,
        message: t('operatorAdmin.rule.displayNameRequired'),
      },
    ],
  }));

  const handleCreate = () => {
    editingId.value = null;
    Object.assign(form, getDefaultForm());
    modalVisible.value = true;
  };

  const handleEdit = (record: AdminRow) => {
    editingId.value = record.id;
    Object.assign(form, {
      operatorId: record.operatorId,
      username: record.username === '--' ? '' : record.username,
      password: '',
      confirmPassword: '',
      displayName: record.displayName === '--' ? '' : record.displayName,
      status: record.status,
    });
    modalVisible.value = true;
  };

  const handleModalOk = async (done: (closed: boolean) => void) => {
    const errors = await formRef.value?.validate();
    if (errors || submitting.value) {
      done(false);
      return;
    }

    const { username: rawUsername, displayName: rawDisplayName, password } =
      form;
    const username = rawUsername.trim();
    const displayName = rawDisplayName.trim();

    submitting.value = true;
    try {
      if (editingId.value) {
        await updateOperatorAdmin({
          id: editingId.value,
          username,
          display_name: displayName,
          status: form.status,
        });
        Message.success(t('operatorAdmin.message.updateSuccess'));
        await fetchData(pagination.current, pagination.pageSize);
      } else {
        await createOperatorAdmin({
          operator_id: form.operatorId as number,
          username,
          password,
          display_name: displayName,
          status: form.status,
        });
        Message.success(t('operatorAdmin.message.createSuccess'));
        await fetchData(1, pagination.pageSize);
      }
      done(true);
    } catch {
      done(false);
    } finally {
      submitting.value = false;
    }
  };

  const handleToggleStatus = (record: AdminRow) => {
    const nextStatus =
      record.status === OPERATOR_ADMIN_STATUS.ENABLED
        ? OPERATOR_ADMIN_STATUS.DISABLED
        : OPERATOR_ADMIN_STATUS.ENABLED;
    togglingId.value = record.id;
    updateOperatorAdminStatus({ id: record.id, status: nextStatus })
      .then(() => {
        Message.success(t('operatorAdmin.message.statusUpdateSuccess'));
        return fetchData(pagination.current, pagination.pageSize);
      })
      .finally(() => {
        togglingId.value = null;
      });
  };

  // ---------------- 重置密码弹窗 ----------------
  const resetPasswordVisible = ref(false);
  const resetPasswordSubmitting = ref(false);
  const resetPasswordTarget = ref<AdminRow | null>(null);
  const resetPasswordFormRef = ref<FormInstance>();
  const resetPasswordForm = reactive({
    password: '',
    confirmPassword: '',
  });

  const resetPasswordRules = computed<
    Record<string, FieldRule | FieldRule[]>
  >(() => ({
    password: [
      {
        validator: (value, callback) => {
          const password = String(value || '');
          if (!password) {
            callback(t('operatorAdmin.rule.passwordRequired'));
            return;
          }
          if (password.length < 6 || password.length > 32) {
            callback(t('operatorAdmin.rule.passwordLength'));
            return;
          }
          callback();
        },
      },
    ],
    confirmPassword: [
      {
        validator: (value, callback) => {
          const confirmPassword = String(value || '');
          if (!confirmPassword) {
            callback(t('operatorAdmin.rule.confirmPasswordRequired'));
            return;
          }
          if (confirmPassword !== resetPasswordForm.password) {
            callback(t('operatorAdmin.rule.passwordNotMatch'));
            return;
          }
          callback();
        },
      },
    ],
  }));

  const handleResetPassword = (record: AdminRow) => {
    resetPasswordTarget.value = record;
    Object.assign(resetPasswordForm, {
      password: '',
      confirmPassword: '',
    });
    resetPasswordVisible.value = true;
  };

  const handleResetPasswordOk = async (done: (closed: boolean) => void) => {
    const errors = await resetPasswordFormRef.value?.validate();
    if (errors || resetPasswordSubmitting.value || !resetPasswordTarget.value) {
      done(false);
      return;
    }

    const { password } = resetPasswordForm;
    resetPasswordSubmitting.value = true;
    try {
      await resetOperatorAdminPassword({
        id: resetPasswordTarget.value.id,
        password,
      });
      Message.success(t('operatorAdmin.message.resetPasswordSuccess'));
      done(true);
    } catch {
      done(false);
    } finally {
      resetPasswordSubmitting.value = false;
    }
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
    width: 240px;
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
  .reset-password-tip {
    margin-bottom: 16px;
  }
</style>
