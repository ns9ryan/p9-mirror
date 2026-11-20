<template>
  <div class="container list-page">
    <Breadcrumb />
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item
          :label="t('nicknameManage.search.keyword')"
          field="keyword"
        >
          <a-input
            v-model="searchForm.keyword"
            :placeholder="t('nicknameManage.search.keywordPlaceholder')"
            :max-length="NICKNAME_MAX_LENGTH"
            allow-clear
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
          <a-button
            v-if="hasPermission(PERMISSION.create)"
            type="primary"
            @click="handleCreate"
          >
            <template #icon><icon-plus /></template>
            {{ t('nicknameManage.toolbar.create') }}
          </a-button>
        </a-space>
        <a-alert type="info" :show-icon="false" class="table-tip">
          {{ t('nicknameManage.tip') }}
        </a-alert>
      </div>
      <a-table
        row-key="id"
        :loading="loading"
        :columns="columns"
        :data="list"
        :pagination="pagination"
        :bordered="true"
        :stripe="true"
        @page-change="onPageChange"
        @page-size-change="onPageSizeChange"
      >
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <a-link
              v-if="hasPermission(PERMISSION.update)"
              @click="handleEdit(record as NicknameRecord)"
            >
              {{ t('front.common.action.edit') }}
            </a-link>
            <a-link
              v-if="hasPermission(PERMISSION.delete)"
              status="danger"
              :disabled="deletingId === (record as NicknameRecord).id"
              @click="handleDelete(record as NicknameRecord)"
            >
              {{ t('front.common.action.delete') }}
            </a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>

    <a-modal
      v-model:visible="modalVisible"
      :title="
        editingId
          ? t('nicknameManage.modal.editTitle')
          : t('nicknameManage.modal.createTitle')
      "
      :width="480"
      :ok-loading="submitting"
      :ok-text="
        editingId
          ? t('front.common.okText.save')
          : t('front.common.okText.create')
      "
      :mask-closable="false"
      unmount-on-close
      @before-ok="handleModalOk"
      @cancel="modalVisible = false"
    >
      <a-form ref="formRef" :model="form" :rules="rules">
        <a-form-item
          :label="t('nicknameManage.field.name')"
          field="name"
          required
        >
          <a-input
            v-model="form.name"
            :placeholder="t('nicknameManage.field.namePlaceholder')"
            :max-length="NICKNAME_MAX_LENGTH"
            :word-length="countChars"
            :word-slice="sliceChars"
            show-word-limit
            allow-clear
            @press-enter="submitByEnter"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, Modal } from '@arco-design/web-vue';
  import type { FieldRule, FormInstance } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import {
    createNickname,
    deleteNickname,
    getNickname,
    NICKNAME_MAX_LENGTH,
    queryNicknameList,
    updateNickname,
    type NicknameListParams,
    type NicknameRecord,
  } from '@/api/nickname';
  import useLoading from '@/hooks/loading';
  import usePermission from '@/hooks/permission';

  const PERMISSION = {
    create: 'nickname:create',
    update: 'nickname:update',
    delete: 'nickname:delete',
  } as const;

  const { t } = useI18n();
  const { hasPermission } = usePermission();
  const { loading, setLoading } = useLoading(true);

  /** 按 Unicode 码点计数，与后端 100 字符上限一致（emoji 记 1 个） */
  const countChars = (value: string) => Array.from(value).length;
  const sliceChars = (value: string, max: number) =>
    Array.from(value).slice(0, max).join('');

  const searchForm = reactive({ keyword: '' });
  const appliedSearch = reactive({ keyword: '' });

  const list = ref<NicknameRecord[]>([]);
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

  const fetchData = async (
    page = pagination.current,
    pageSize = pagination.pageSize
  ) => {
    setLoading(true);
    try {
      const params: NicknameListParams = { page, page_size: pageSize };
      const keyword = appliedSearch.keyword.trim();
      if (keyword) params.keyword = keyword;
      const { data } = await queryNicknameList(params);
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
      title: t('nicknameManage.column.id'),
      dataIndex: 'id',
      width: 100,
      align: 'center',
    },
    {
      title: t('nicknameManage.column.name'),
      dataIndex: 'name',
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('front.common.column.operations'),
      dataIndex: 'operations',
      slotName: 'operations',
      width: 160,
      align: 'center',
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

  // ---------------- 新增 / 编辑弹窗 ----------------
  const modalVisible = ref(false);
  const submitting = ref(false);
  const editingId = ref<number | null>(null);
  const formRef = ref<FormInstance>();
  const form = reactive({ name: '' });

  const rules = computed<Record<string, FieldRule | FieldRule[]>>(() => ({
    name: [
      {
        validator: (
          value: string | undefined,
          callback: (msg?: string) => void
        ) => {
          const name = (value || '').trim();
          if (!name) {
            callback(t('nicknameManage.rule.nameRequired'));
            return;
          }
          if (countChars(name) > NICKNAME_MAX_LENGTH) {
            callback(
              t('nicknameManage.rule.nameTooLong', { max: NICKNAME_MAX_LENGTH })
            );
            return;
          }
          callback();
        },
      },
    ],
  }));

  const handleCreate = () => {
    editingId.value = null;
    form.name = '';
    modalVisible.value = true;
  };

  const handleEdit = async (record: NicknameRecord) => {
    editingId.value = record.id;
    form.name = record.name;
    modalVisible.value = true;
    try {
      const { data } = await getNickname({ id: record.id });
      if (data?.nickname && editingId.value === record.id) {
        form.name = data.nickname.name;
      }
    } catch {
      // 详情拉取失败时保留列表行数据，仍可编辑
    }
  };

  const submit = async () => {
    const errors = await formRef.value?.validate();
    if (errors || submitting.value) return false;

    const name = form.name.trim();
    submitting.value = true;
    try {
      if (editingId.value) {
        await updateNickname({ id: editingId.value, name });
        Message.success(t('front.common.message.updateSuccess'));
      } else {
        await createNickname({ name });
        Message.success(t('front.common.message.createSuccess'));
      }
      await fetchData(pagination.current, pagination.pageSize);
      return true;
    } catch {
      return false;
    } finally {
      submitting.value = false;
    }
  };

  const handleModalOk = async (done: (closed: boolean) => void) => {
    done(await submit());
  };

  const submitByEnter = async () => {
    if (await submit()) modalVisible.value = false;
  };

  const handleDelete = (record: NicknameRecord) => {
    Modal.confirm({
      title: t('front.common.delete.title'),
      content: t('nicknameManage.message.deleteConfirmContent', {
        name: record.name,
      }),
      okText: t('front.common.delete.okText'),
      okButtonProps: { status: 'danger' },
      async onOk() {
        deletingId.value = record.id;
        try {
          await deleteNickname({ id: record.id });
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
  .table-toolbar {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 12px;
    margin-bottom: 14px;
  }
  .table-tip {
    width: auto;
    padding: 4px 12px;
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
