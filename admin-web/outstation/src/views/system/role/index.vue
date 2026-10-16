<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.route.system', 'menu.route.role']" />
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item :label="$t('roleManage.search.name')" field="keyword">
          <a-input
            v-model="searchForm.keyword"
            :placeholder="$t('roleManage.search.placeholder')"
            allow-clear
            class="search-input"
          />
        </a-form-item>
        <a-form-item :label="$t('roleManage.search.status')" field="status">
          <a-select
            v-model="searchForm.status"
            :placeholder="$t('roleManage.search.statusPlaceholder')"
            allow-clear
          >
            <a-option :value="1">{{ $t('front.common.status.enabled') }}</a-option>
            <a-option :value="2">{{ $t('front.common.status.disabled') }}</a-option>
          </a-select>
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
        <a-button
          v-if="hasPermission('role:create')"
          type="primary"
          @click="handleCreate"
        >
          <template #icon><icon-plus /></template>
          {{ $t('front.common.toolbar.add') }}
        </a-button>
      </div>
      <a-table
        row-key="id"
        :loading="loading"
        :columns="columns"
        :data="displayRoles"
        :pagination="pagination"
        :bordered="true"
        :scroll="{ x: 1100 }"
        :stripe="true"
        @page-change="onPageChange"
        @page-size-change="onPageSizeChange"
      >
        <template #index="{ rowIndex }">
          {{ rowIndex + 1 + (pagination.current - 1) * pagination.pageSize }}
        </template>
        <template #updatedAt="{ record }">
          <!-- <div>{{ record.updatedBy || '--' }}</div> -->
          <div>{{ record.updatedAt }}</div>
        </template>
        <template #status="{ record }">
          <a-switch
            :model-value="record.enabled"
            :checked-text="$t('front.common.switch.on')"
            :unchecked-text="$t('front.common.switch.off')"
            size="small"
            :disabled="record.isSystem"
            :loading="statusUpdatingId === record.id"
            @change="(value) => handleStatusChange(record, Boolean(value))"
          />
        </template>
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <a-link @click="handleEdit(record)">{{ $t('front.common.action.edit') }}</a-link>
            <a-link @click="handlePermission(record)">{{ $t('roleManage.action.permission') }}</a-link>
            <a-link
              status="danger"
              :disabled="
                deletingId === record.id || record.isSystem
              "
              @click="handleDelete(record)"
            >
              {{ $t('front.common.action.delete') }}
            </a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>
    <RoleModal
      :visible="modalVisible"
      :role="editingRole"
      @update:visible="onModalVisibleChange"
      @success="fetchData"
    />
    <PermissionModal
      :visible="permissionVisible"
      :role="permissionRole"
      @update:visible="onPermissionVisibleChange"
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
    deleteRole,
    queryRoleList,
    updateRole,
    type RoleRecord,
  } from '@/api/role';
  import useLoading from '@/hooks/loading';
  import usePermission from '@/hooks/permission';
  import RoleModal from './components/role-modal.vue';
  import PermissionModal from './components/permission-modal.vue';

  const { t } = useI18n();
  const { hasPermission } = usePermission();

  interface RoleRow {
    id: number;
    roleName: string;
    roleCode: string;
    description: string;
    updatedBy: string;
    updatedAt: string;
    enabled: boolean;
    status: number;
    sortNo: number;
    isSystem: boolean;
    raw: RoleRecord;
  }

  const { loading, setLoading } = useLoading(true);
  const modalVisible = ref(false);
  const editingRole = ref<RoleRecord | null>(null);
  const permissionVisible = ref(false);
  const permissionRole = ref<RoleRecord | null>(null);
  const deletingId = ref<number | null>(null);
  const statusUpdatingId = ref<number | null>(null);
  const searchForm = reactive<{
    keyword: string;
    status: number | '';
  }>({
    keyword: '',
    status: '',
  });
  const appliedSearch = reactive({
    keyword: '',
    status: '' as number | '',
  });
  const roles = ref<RoleRow[]>([]);
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

  const mapRole = (item: RoleRecord): RoleRow => ({
    id: item.id,
    roleName:
      item.role_name ||
      item.name ||
      item.display_name ||
      t('roleManage.roleFallback', { id: item.id }),
    roleCode: item.role_code || item.code || '--',
    description: item.description || '--',
    updatedBy: item.updated_by || '--',
    updatedAt: formatTime(item.updated_at),
    enabled: (item.status ?? 1) === 1,
    status: item.status ?? 1,
    sortNo: item.sort_no ?? 0,
    isSystem: Boolean(item.is_system),
    raw: item,
  });

  const fetchData = async (
    page = pagination.current,
    pageSize = pagination.pageSize
  ) => {
    setLoading(true);
    try {
      const params: {
        page: number;
        page_size: number;
        role_name?: string;
        status?: number;
      } = {
        page,
        page_size: pageSize,
      };
      const keyword = appliedSearch.keyword.trim();
      if (keyword) params.role_name = keyword;
      if (appliedSearch.status !== '') params.status = appliedSearch.status;

      const { data } = await queryRoleList(params);
      roles.value = (data?.list || []).map(mapRole);
      pagination.current = page;
      pagination.pageSize = pageSize;
      pagination.total = data?.total ?? 0;
    } catch {
      roles.value = [];
      pagination.total = 0;
    } finally {
      setLoading(false);
    }
  };

  /** 接口筛选参数若暂未生效，对当前页做本地兜底过滤 */
  const displayRoles = computed(() =>
    roles.value.filter((role) => {
      const keyword = appliedSearch.keyword.trim().toLowerCase();
      const matchedKeyword =
        !keyword || role.roleName.toLowerCase().includes(keyword);
      const matchedStatus =
        appliedSearch.status === '' || role.status === appliedSearch.status;
      return matchedKeyword && matchedStatus;
    })
  );

  const columns = computed<TableColumnData[]>(() => [
    {
      title: t('front.common.column.index'),
      dataIndex: 'index',
      slotName: 'index',
      width: 90,
      align: 'center',
    },
    {
      title: t('roleManage.column.roleName'),
      dataIndex: 'roleName',
      width: 180,
      align: 'center',
    },
    {
      title: t('roleManage.column.description'),
      dataIndex: 'description',
      width: 180,
      align: 'center',
    },
    {
      title: t('front.common.column.updatedAt'),
      dataIndex: 'updatedAt',
      slotName: 'updatedAt',
      width: 200,
      align: 'center',
    },
    {
      title: t('front.common.column.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 100,
      align: 'center',
    },
    {
      title: t('front.common.column.operations'),
      dataIndex: 'operations',
      slotName: 'operations',
      width: 220,
      align: 'center',
      fixed: 'right',
    },
  ]);

  const handleSearch = () => {
    appliedSearch.keyword = searchForm.keyword;
    appliedSearch.status = searchForm.status;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    Object.assign(searchForm, { keyword: '', status: '' });
    Object.assign(appliedSearch, {
      keyword: '',
      status: '',
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
    if (!visible) editingRole.value = null;
  };

  const onPermissionVisibleChange = (visible: boolean) => {
    permissionVisible.value = visible;
    if (!visible) permissionRole.value = null;
  };

  const handleCreate = () => {
    editingRole.value = null;
    modalVisible.value = true;
  };

  const handleEdit = (record: RoleRow) => {
    editingRole.value = record.raw;
    modalVisible.value = true;
  };

  const handlePermission = (record: RoleRow) => {
    permissionRole.value = record.raw;
    permissionVisible.value = true;
  };

  const handleStatusChange = async (record: RoleRow, enabled: boolean) => {
    if (record.isSystem) {
      Message.warning(t('roleManage.message.disableSystemRole'));
      return;
    }
    if (statusUpdatingId.value === record.id) return;
    const nextStatus = enabled ? 1 : 2;
    statusUpdatingId.value = record.id;
    try {
      await updateRole({ id: record.id, status: nextStatus });
      record.enabled = enabled;
      record.status = nextStatus;
      record.raw.status = nextStatus;
      Message.success(
        enabled
          ? t('front.common.message.enabled')
          : t('roleManage.message.disabled')
      );
    } catch {
      // 保持原状态；错误提示由拦截器处理
    } finally {
      statusUpdatingId.value = null;
    }
  };

  const handleDelete = (record: RoleRow) => {
    if (record.isSystem) {
      Message.warning(t('roleManage.message.deleteSystemRole'));
      return;
    }
    Modal.confirm({
      title: t('front.common.delete.title'),
      content: t('roleManage.delete.content', { name: record.roleName }),
      okText: t('front.common.delete.okText'),
      okButtonProps: { status: 'danger' },
      async onOk() {
        deletingId.value = record.id;
        try {
          await deleteRole({ ids: [record.id] });
          Message.success(t('front.common.message.deleteSuccess'));
          const nextPage =
            roles.value.length <= 1 && pagination.current > 1
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
    width: 200px;
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
  .operation-links {
    white-space: nowrap;
  }
  :deep(.arco-pagination-options .arco-select) {
    min-width: 100px;
  }
</style>
