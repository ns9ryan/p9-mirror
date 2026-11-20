<template>
  <div class="container list-page">
    <Breadcrumb
      :items="['menu.accountManagement', 'menu.route.user']"
    />
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item field="account">
          <a-input-group compact>
            <a-select v-model="searchType" class="search-type">
              <a-option value="account">{{ $t('userManage.search.account') }}</a-option>
              <a-option value="nickname">{{ $t('userManage.search.nickname') }}</a-option>
            </a-select>
            <a-input
              v-model="searchForm.account"
              :placeholder="$t('userManage.search.placeholder')"
              allow-clear
              class="search-input"
            />
          </a-input-group>
        </a-form-item>
        <a-form-item :label="$t('userManage.search.role')" field="role_id">
          <a-select
            v-model="searchForm.role_id"
            :placeholder="$t('userManage.search.rolePlaceholder')"
            allow-clear
            :loading="roleLoading"
            :options="roleOptions"
          />
        </a-form-item>
        <a-form-item :label="$t('userManage.search.status')" field="status">
          <a-select
            v-model="searchForm.status"
            :placeholder="$t('userManage.search.statusPlaceholder')"
            allow-clear
          >
            <a-option :value="1">{{ $t('front.common.status.enabled') }}</a-option>
            <a-option :value="2">{{ $t('front.common.status.disabled') }}</a-option>
          </a-select>
        </a-form-item>
        <a-form-item
          ><a-space
            ><a-button type="primary" @click="handleSearch"
              ><template #icon><icon-search /></template>{{ $t('front.common.search.search') }}</a-button
            ><a-button @click="handleReset"
              ><template #icon><icon-refresh /></template>{{ $t('front.common.search.reset') }}</a-button
            ></a-space
          ></a-form-item
        >
      </a-form>
    </a-card>
    <a-card class="general-card table-card" :bordered="false">
      <div class="table-toolbar"
        ><a-space
          ><a-button
            v-if="hasPermission('user:create')"
            type="primary"
            @click="handleCreate"
            ><template #icon><icon-plus /></template>{{ $t('front.common.toolbar.add') }}</a-button
          ><a-button type="primary" @click="notify($t('userManage.toolbar.securitySettings'))"
            >{{ $t('userManage.toolbar.securitySettings') }}</a-button
          ><span class="switch-label">{{ $t('userManage.toolbar.googleAuthenticator') }}</span
          ><a-switch
            v-model="googleAuthenticator"
            :checked-text="$t('front.common.switch.on')"
            :unchecked-text="$t('front.common.switch.off')" /></a-space
      ></div>
      <a-table
        row-key="id"
        :loading="loading"
        :columns="columns"
        :data="displayUsers"
        :pagination="pagination"
        :bordered="true"
        :scroll="tableScroll"
        :stripe="true"
        @page-change="onPageChange"
        @page-size-change="onPageSizeChange"
      >
        <template #index="{ rowIndex }">
          {{ rowIndex + 1 + (pagination.current - 1) * pagination.pageSize }}
        </template>
        <template #groups="{ record }"
          ><a-space wrap :size="4"
            ><a-tag
              v-for="group in record.groups"
              :key="group"
              color="arcoblue"
              size="small"
              >{{ group }}</a-tag
            ></a-space
          ></template
        >
        <template #lastLogin="{ record }"
          ><span v-if="record.lastLogin === '--'">--</span
          ><template v-else
            ><div>{{ record.lastLogin }}</div
            ><div>{{ record.ip }}</div></template
          ></template
        >
        <template #resetGoogle="{ record }"
          ><a-link @click="handleResetGoogle(record)"
            >{{ $t('userManage.action.reset') }}</a-link
          ></template
        >
        <template #status="{ record }"
          ><a-switch
            :model-value="record.enabled"
            :checked-text="$t('front.common.switch.on')"
            :unchecked-text="$t('front.common.switch.off')"
            size="small"
            :disabled="record.isSystem"
            :loading="statusUpdatingId === record.id"
            @change="(value) => handleStatusChange(record, Boolean(value))"
        /></template>
        <template #operations="{ record }"
          ><a-space :size="12" class="operation-links"
            ><a-link @click="handleIpManage(record)"
              ><span class="status-dot-text"
                ><span
                  class="status-dot"
                  :class="
                    record.ipWhitelistEnabled
                      ? 'status-dot--on'
                      : 'status-dot--off'
                  "
                />{{ $t('userManage.action.ipManage') }}</span
              ></a-link
            ><a-link @click="handleChangePassword(record)">{{ $t('userManage.action.changePassword') }}</a-link
            ><a-link @click="handleEdit(record)">{{ $t('front.common.action.edit') }}</a-link
            ><a-link
              status="danger"
              :disabled="deletingId === record.id"
              @click="handleDelete(record)"
              >{{ $t('front.common.action.delete') }}</a-link
            ></a-space
          ></template
        >
      </a-table>
    </a-card>
    <CreateUserModal
      :visible="createVisible"
      :user-id="editingUserId"
      @update:visible="onModalVisibleChange"
      @success="onCreateSuccess"
    />
    <ChangeUserPasswordModal
      :visible="passwordVisible"
      :user-id="passwordUserId"
      @update:visible="onPasswordModalVisibleChange"
    />
    <IpManageModal
      :visible="ipManageVisible"
      :user-id="ipManageUserId"
      @update:visible="onIpManageModalVisibleChange"
      @success="onIpManageSuccess"
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
    deleteUser,
    queryRoleList,
    queryUserList,
    updateUser,
    type AdminUserRecord,
    type RoleRecord,
  } from '@/api/user';
  import useLoading from '@/hooks/loading';
  import usePermission from '@/hooks/permission';
  import CreateUserModal from './components/create-user-modal.vue';
  import ChangeUserPasswordModal from './components/change-user-password-modal.vue';
  import IpManageModal from './components/ip-manage-modal.vue';

  const { t } = useI18n();
  const { hasPermission } = usePermission();

  interface UserRecord {
    id: number;
    account: string;
    nickname: string;
    groups: string[];
    createdAt: string;
    lastLogin: string;
    ip: string;
    enabled: boolean;
    status: number;
    isSystem: boolean;
    /** IP 白名单开关：true 开 / false 关 */
    ipWhitelistEnabled: boolean;
  }

  const { loading, setLoading } = useLoading(true);
  const createVisible = ref(false);
  const editingUserId = ref<number | null>(null);
  const passwordVisible = ref(false);
  const passwordUserId = ref<number | null>(null);
  const ipManageVisible = ref(false);
  const ipManageUserId = ref<number | null>(null);
  const deletingId = ref<number | null>(null);
  const statusUpdatingId = ref<number | null>(null);
  const searchType = ref<'account' | 'nickname'>('account');
  const googleAuthenticator = ref(false);
  const roleLoading = ref(false);
  const roleOptions = ref<{ label: string; value: number }[]>([]);
  const searchForm = reactive<{
    account: string;
    role_id: number | '';
    status: number | '';
  }>({
    account: '',
    role_id: '',
    status: '',
  });
  const appliedSearch = reactive({
    type: 'account' as 'account' | 'nickname',
    account: '',
    role_id: '' as number | '',
    status: '' as number | '',
  });
  const users = ref<UserRecord[]>([]);
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

  const roleLabel = (item: RoleRecord) =>
    item.name ||
    item.role_name ||
    item.display_name ||
    item.role_code ||
    item.code ||
    t('userModal.roleFallback', { id: item.id });

  const fetchRoles = async () => {
    roleLoading.value = true;
    try {
      const { data } = await queryRoleList(
        { page: 1, page_size: 200 },
        { silentError: true }
      );
      roleOptions.value = (data?.list || []).map((item) => ({
        label: roleLabel(item),
        value: item.id,
      }));
    } catch {
      roleOptions.value = [];
    } finally {
      roleLoading.value = false;
    }
  };

  const formatTime = (value?: number) => {
    if (!value) return '--';
    const ms = value < 1e12 ? value * 1000 : value;
    return dayjs(ms).format('YYYY-MM-DD HH:mm:ss');
  };

  const mapUser = (item: AdminUserRecord): UserRecord => ({
    id: item.id,
    account: item.username,
    nickname: item.display_name,
    groups: item.role_names?.length
      ? item.role_names
      : item.role_codes || [],
    createdAt: formatTime(item.created_at),
    lastLogin: formatTime(item.last_login_at),
    ip: '',
    enabled: item.status === 1,
    status: item.status,
    isSystem: Boolean(item.is_super_admin),
    ipWhitelistEnabled: item.ip_whitelist_enabled === 1,
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
        username?: string;
        display_name?: string;
        role_ids?: number[];
        status?: number;
      } = {
        page,
        page_size: pageSize,
      };
      const keyword = appliedSearch.account.trim();
      if (keyword) {
        if (appliedSearch.type === 'nickname') params.display_name = keyword;
        else params.username = keyword;
      }
      if (appliedSearch.role_id !== '') params.role_ids = [appliedSearch.role_id];
      if (appliedSearch.status !== '') params.status = appliedSearch.status;

      const { data } = await queryUserList(params);
      users.value = (data?.list || []).map(mapUser);
      pagination.current = page;
      pagination.pageSize = pageSize;
      pagination.total = data?.total ?? 0;
    } catch (err) {
      users.value = [];
      pagination.total = 0;
    } finally {
      setLoading(false);
    }
  };

  /** 接口筛选参数若暂未生效，对当前页做本地兜底过滤 */
  const displayUsers = computed(() =>
    users.value.filter((user) => {
      const keyword = appliedSearch.account.trim().toLowerCase();
      const matchedKeyword =
        !keyword ||
        (appliedSearch.type === 'account' ? user.account : user.nickname)
          .toLowerCase()
          .includes(keyword);
      const matchedStatus =
        appliedSearch.status === '' || user.status === appliedSearch.status;
      return matchedKeyword && matchedStatus;
    })
  );

  const columns = computed<TableColumnData[]>(() => [
    {
      title: t('front.common.column.index'),
      dataIndex: 'index',
      slotName: 'index',
      width: 80,
      align: 'center',
    },
    {
      title: t('userManage.column.account'),
      dataIndex: 'account',
      width: 130,
      align: 'center',
    },
    {
      title: t('userManage.column.nickname'),
      dataIndex: 'nickname',
      width: 130,
      align: 'center',
    },
    {
      title: t('userManage.column.groups'),
      dataIndex: 'groups',
      slotName: 'groups',
      minWidth: 150,
      align: 'center',
    },
    {
      title: t('front.common.column.createdAt'),
      dataIndex: 'createdAt',
      minWidth: 150,
      align: 'center',
    },
    {
      title: t('userManage.column.lastLogin'),
      dataIndex: 'lastLogin',
      slotName: 'lastLogin',
      minWidth: 150,
      align: 'center',
    },
    {
      title: t('userManage.column.resetGoogle'),
      dataIndex: 'resetGoogle',
      slotName: 'resetGoogle',
      width: 170,
      align: 'center',
    },
    {
      title: t('front.common.column.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 90,
      align: 'center',
    },
    {
      title: t('front.common.column.operations'),
      dataIndex: 'operations',
      slotName: 'operations',
      width: 300,
      align: 'center',
      fixed: 'right',
    },
  ]);

  /** scroll.x 按列宽合计，无 width 的列用 minWidth 参与，避免右侧留白 */
  const tableScroll = computed(() => {
    const total = columns.value.reduce((sum, col) => {
      const w = Number(col.width) || Number(col.minWidth) || 0;
      return sum + w;
    }, 0);
    return { x: total };
  });

  const handleSearch = () => {
    appliedSearch.type = searchType.value;
    appliedSearch.account = searchForm.account;
    appliedSearch.role_id = searchForm.role_id;
    appliedSearch.status = searchForm.status;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    searchType.value = 'account';
    Object.assign(searchForm, { account: '', role_id: '', status: '' });
    Object.assign(appliedSearch, {
      type: 'account',
      account: '',
      role_id: '',
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

  const onCreateSuccess = () => {
    fetchData(1, pagination.pageSize);
  };

  const onModalVisibleChange = (visible: boolean) => {
    createVisible.value = visible;
    if (!visible) editingUserId.value = null;
  };

  const handleCreate = () => {
    editingUserId.value = null;
    createVisible.value = true;
  };

  const handleEdit = (user: UserRecord) => {
    editingUserId.value = user.id;
    createVisible.value = true;
  };

  const onPasswordModalVisibleChange = (visible: boolean) => {
    passwordVisible.value = visible;
    if (!visible) passwordUserId.value = null;
  };

  const handleChangePassword = (user: UserRecord) => {
    passwordUserId.value = user.id;
    passwordVisible.value = true;
  };

  const onIpManageModalVisibleChange = (visible: boolean) => {
    ipManageVisible.value = visible;
    if (!visible) ipManageUserId.value = null;
  };

  const onIpManageSuccess = () => {
    fetchData(pagination.current, pagination.pageSize);
  };

  const handleIpManage = (user: UserRecord) => {
    ipManageUserId.value = user.id;
    ipManageVisible.value = true;
  };

  const handleDelete = (user: UserRecord) => {
    Modal.confirm({
      title: t('front.common.delete.title'),
      content: t('userManage.delete.content', { account: user.account }),
      okText: t('front.common.delete.okText'),
      okButtonProps: { status: 'danger' },
      async onOk() {
        deletingId.value = user.id;
        try {
          await deleteUser({ ids: [user.id] });
          Message.success(t('front.common.message.deleteSuccess'));
          const nextPage =
            users.value.length <= 1 && pagination.current > 1
              ? pagination.current - 1
              : pagination.current;
          await fetchData(nextPage, pagination.pageSize);
        } finally {
          deletingId.value = null;
        }
      },
    });
  };

  const handleStatusChange = async (user: UserRecord, enabled: boolean) => {
    if (user.isSystem) {
      Message.warning(t('userManage.message.disableSystemUser'));
      return;
    }
    if (statusUpdatingId.value === user.id) return;
    const nextStatus = enabled ? 1 : 2;
    statusUpdatingId.value = user.id;
    try {
      await updateUser({ id: user.id, status: nextStatus });
      user.enabled = enabled;
      user.status = nextStatus;
      Message.success(
        enabled
          ? t('front.common.message.enabled')
          : t('userManage.message.disabled')
      );
    } catch {
      // 保持原状态；错误提示由拦截器处理
    } finally {
      statusUpdatingId.value = null;
    }
  };

  const notify = (action: string) => Message.info(action);

  const handleResetGoogle = (user: UserRecord) =>
    Message.info(
      t('userManage.message.resetGoogleDone', { account: user.account })
    );

  fetchRoles();
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
  .search-type :deep(.arco-select-view) {
    min-width: 112px;
  }
  .search-input {
    width: 170px;
  }
  .table-toolbar {
    margin-bottom: 14px;
  }
  .switch-label {
    color: var(--color-text-1);
    font-weight: 500;
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
  .status-dot--on {
    background-color: rgb(var(--success-6));
  }
  .status-dot--off {
    background-color: rgb(var(--gray-5));
  }
  :deep(.arco-pagination-options .arco-select) {
    min-width: 100px;
  }
</style>
