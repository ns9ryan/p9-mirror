<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.route.i18n', 'menu.route.i18nLang']" />
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item :label="$t('i18nLangManage.search.lang')" field="lang">
          <a-input
            v-model="searchForm.lang"
            :placeholder="$t('i18nLangManage.search.langPlaceholder')"
            allow-clear
            class="search-input"
          />
        </a-form-item>
        <a-form-item
          :label="$t('i18nLangManage.search.disabled')"
          field="disabled"
        >
          <a-select
            v-model="searchForm.disabled"
            :placeholder="$t('i18nLangManage.search.disabledPlaceholder')"
            allow-clear
            class="search-select"
          >
            <a-option :value="0">
              {{ $t('front.common.status.enabled') }}
            </a-option>
            <a-option :value="1">
              {{ $t('front.common.status.disabled') }}
            </a-option>
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
        <a-space>
          <a-button type="primary" @click="handleCreate">
            <template #icon><icon-plus /></template>
            {{ $t('front.common.toolbar.add') }}
          </a-button>
          <!-- 暂时不支持批量删除 -->
          <!-- <a-button
            status="danger"
            :disabled="!selectedKeys.length"
            @click="handleBatchDelete"
          >
            <template #icon><icon-delete /></template>
            {{ $t('i18nLangManage.toolbar.batchDelete') }}
          </a-button> -->
          <a-button @click="() => fetchData()">
            <template #icon><icon-refresh /></template>
            {{ $t('front.common.toolbar.refresh') }}
          </a-button>
        </a-space>
        <a-alert
          v-if="isFiltering"
          type="warning"
          :show-icon="false"
          class="drag-tip"
        >
          {{ $t('i18nLangManage.drag.disabledTip') }}
        </a-alert>
        <a-alert v-else type="info" :show-icon="false" class="drag-tip">
          {{ $t('i18nLangManage.drag.tipPrefix') }}
          <icon-drag-arrow />
          {{ $t('i18nLangManage.drag.tipSuffix') }}
        </a-alert>
      </div>
      <a-table
        ref="tableRef"
        row-key="id"
        :loading="loading"
        :columns="columns"
        :data="list"
        :pagination="pagination"
        :bordered="true"
        :stripe="true"
        :scroll="{ x: 1400 }"
        :row-selection="rowSelection"
        :selected-keys="selectedKeys"
        :class="{
          'is-row-dragging': isRowDragging || isDragPersisting,
        }"
        @selection-change="onSelectionChange"
        @page-change="onPageChange"
        @page-size-change="onPageSizeChange"
      >
        <template #dragHandle>
          <icon-drag-arrow
            class="row-drag-handle"
            :class="{ 'row-drag-handle--disabled': isFiltering }"
          />
        </template>
        <template #index="{ rowIndex }">
          {{ rowIndex + 1 + (pagination.current - 1) * pagination.pageSize }}
        </template>
        <template #status="{ record }">
          <a-switch
            :model-value="!record.isDisabled"
            :checked-text="$t('front.common.status.enabled')"
            :unchecked-text="$t('front.common.status.disabled')"
            size="small"
            :disabled="Boolean(record.isDefault)"
            :loading="togglingId === record.id"
            @change="
              (checked) => handleToggleDisabled(record, Boolean(checked))
            "
          />
        </template>
        <template #isDefault="{ record }">
          <a-tag v-if="record.isDefault" color="arcoblue">
            {{ $t('i18nLangManage.isDefault.yes') }}
          </a-tag>
          <a-link
            v-else
            :disabled="settingDefaultId === record.id"
            @click="handleSetDefault(record)"
          >
            {{ $t('i18nLangManage.action.setDefault') }}
          </a-link>
        </template>
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <a-link @click="handleEdit(record)">
              {{ $t('front.common.action.edit') }}
            </a-link>
            <a-link
              v-if="hasPermission('i18n:import')"
              :disabled="importingId === record.id"
              @click="handleImport(record)"
            >
              {{ $t('i18nManage.toolbar.import') }}
            </a-link>
            <a-link
              v-if="hasPermission('i18n:export')"
              :disabled="exportingId === record.id"
              @click="handleExport(record)"
            >
              {{ $t('i18nManage.toolbar.export') }}
            </a-link>
            <a-link
              status="danger"
              :disabled="
                deletingId === record.id || Boolean(record.isDefault)
              "
              @click="handleDelete(record)"
            >
              {{ $t('front.common.action.delete') }}
            </a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>
    <LangModal
      :visible="modalVisible"
      :record="editingRecord"
      @update:visible="onModalVisibleChange"
      @success="fetchData"
    />
    <input
      ref="importInputRef"
      type="file"
      accept=".json,application/json"
      class="hidden-file-input"
      @change="onImportFileChange"
    />
  </div>
</template>

<script lang="ts" setup>
  import {
    computed,
    nextTick,
    onBeforeUnmount,
    reactive,
    ref,
    watch,
  } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, Modal } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import dayjs from 'dayjs';
  import Sortable from 'sortablejs';
  import {
    deleteI18nLang,
    exportI18n,
    importI18n,
    queryI18nLangList,
    reorderI18nLang,
    updateI18nLang,
    type I18nLangListParams,
    type I18nLangRecord,
  } from '@/api/i18n';
  import useLoading from '@/hooks/loading';
  import usePermission from '@/hooks/permission';
  import LangModal from './components/lang-modal.vue';

  const { t } = useI18n();
  const { hasPermission } = usePermission();

  interface LangRow {
    id: number;
    lang: string;
    name: string;
    i18nKey: string;
    i18nName: string;
    isDisabled: number;
    isDefault: number;
    createdAt: string;
    updatedAt: string;
    raw: I18nLangRecord;
  }

  const { loading, setLoading } = useLoading(true);
  const modalVisible = ref(false);
  const editingRecord = ref<I18nLangRecord | null>(null);
  const deletingId = ref<number | null>(null);
  const togglingId = ref<number | null>(null);
  const settingDefaultId = ref<number | null>(null);
  const importingId = ref<number | null>(null);
  const importingLang = ref('');
  const exportingId = ref<number | null>(null);
  const importInputRef = ref<HTMLInputElement | null>(null);
  const selectedKeys = ref<(string | number)[]>([]);
  const tableRef = ref();
  let rowSortable: Sortable | null = null;
  /** 拖拽落库中，禁止分页跳转 */
  const isDragPersisting = ref(false);
  /** 行拖拽进行中（防止松手误点到底部分页） */
  const isRowDragging = ref(false);

  const searchForm = reactive<{
    lang: string;
    disabled: number | undefined;
  }>({
    lang: '',
    disabled: undefined,
  });
  const appliedSearch = reactive<{
    lang: string;
    disabled: number | undefined;
  }>({
    lang: '',
    disabled: undefined,
  });
  /** 存在筛选条件时，当前页并非连续的排序区间，禁止拖拽排序 */
  const isFiltering = computed(
    () =>
      Boolean(appliedSearch.lang.trim()) ||
      appliedSearch.disabled !== undefined
  );

  const list = ref<LangRow[]>([]);
  const pagination = reactive({
    current: 1,
    pageSize: 15,
    total: 0,
    showTotal: true,
    showPageSize: true,
    pageSizeOptions: [15, 20, 30, 50, 100],
    pageSizeProps: {
      popupContainer: 'body',
    },
  });

  const rowSelection = {
    type: 'checkbox' as const,
    showCheckedAll: true,
    onlyCurrent: false,
  };

  const formatTime = (value?: number) => {
    if (!value) return '--';
    const ms = value < 1e12 ? value * 1000 : value;
    return dayjs(ms).format('YYYY-MM-DD HH:mm:ss');
  };

  const mapRow = (item: I18nLangRecord): LangRow => ({
    id: item.id,
    lang: item.lang || '--',
    name: item.name || '--',
    i18nKey: item.i18n_key || '--',
    i18nName: item.i18n_name || '--',
    isDisabled: item.disabled ?? 0,
    isDefault: item.is_default ?? 0,
    createdAt: formatTime(item.created_at),
    updatedAt: formatTime(item.updated_at),
    raw: item,
  });

  const buildQueryParams = (
    page = pagination.current,
    pageSize = pagination.pageSize
  ): I18nLangListParams => {
    const params: I18nLangListParams = {
      page,
      page_size: pageSize,
    };
    const lang = appliedSearch.lang.trim();
    if (lang) params.lang = lang;
    if (appliedSearch.disabled === 0 || appliedSearch.disabled === 1) {
      params.disabled = appliedSearch.disabled;
    }
    return params;
  };

  const fetchData = async (
    page = pagination.current,
    pageSize = pagination.pageSize
  ) => {
    setLoading(true);
    try {
      const { data } = await queryI18nLangList(
        buildQueryParams(page, pageSize)
      );
      list.value = (data?.list || []).map(mapRow);
      pagination.current = page;
      pagination.pageSize = pageSize;
      pagination.total = data?.total ?? 0;
      selectedKeys.value = [];
    } catch {
      list.value = [];
      pagination.total = 0;
      selectedKeys.value = [];
    } finally {
      setLoading(false);
    }
  };

  const columns = computed<TableColumnData[]>(() => [
    {
      title: t('i18nLangManage.column.drag'),
      dataIndex: 'dragHandle',
      slotName: 'dragHandle',
      width: 70,
      align: 'center',
      fixed: 'left',
    },
    {
      title: t('front.common.column.index'),
      dataIndex: 'index',
      slotName: 'index',
      width: 80,
      align: 'center',
      fixed: 'left',
    },
    {
      title: t('i18nLangManage.column.id'),
      dataIndex: 'id',
      width: 90,
      align: 'center',
    },
    {
      title: t('i18nLangManage.column.lang'),
      dataIndex: 'lang',
      width: 120,
      align: 'center',
    },
    {
      title: t('i18nLangManage.column.name'),
      dataIndex: 'name',
      width: 160,
      align: 'center',
    },
    {
      title: t('i18nLangManage.column.i18n_key'),
      dataIndex: 'i18nKey',
      width: 180,
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('i18nLangManage.column.i18n_name'),
      dataIndex: 'i18nName',
      width: 160,
      align: 'center',
    },
    {
      title: t('i18nLangManage.column.disabled'),
      dataIndex: 'isDisabled',
      slotName: 'status',
      width: 120,
      align: 'center',
    },
    // 暂时没有这个字段
    // {
    //   title: t('i18nLangManage.column.isDefault'),
    //   dataIndex: 'isDefault',
    //   slotName: 'isDefault',
    //   width: 120,
    //   align: 'center',
    // },
    {
      title: t('front.common.column.createdAt'),
      dataIndex: 'createdAt',
      width: 180,
      align: 'center',
    },
    {
      title: t('front.common.column.updatedAt'),
      dataIndex: 'updatedAt',
      width: 180,
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
    appliedSearch.lang = searchForm.lang;
    appliedSearch.disabled = searchForm.disabled;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    searchForm.lang = '';
    searchForm.disabled = undefined;
    appliedSearch.lang = '';
    appliedSearch.disabled = undefined;
    fetchData(1, pagination.pageSize);
  };

  const onPageChange = (current: number) => {
    if (isDragPersisting.value || isRowDragging.value) return;
    fetchData(current, pagination.pageSize);
  };

  const onPageSizeChange = (pageSize: number) => {
    if (isDragPersisting.value || isRowDragging.value) return;
    fetchData(1, pageSize);
  };

  const onSelectionChange = (keys: (string | number)[]) => {
    selectedKeys.value = keys;
  };

  const onModalVisibleChange = (visible: boolean) => {
    modalVisible.value = visible;
    if (!visible) editingRecord.value = null;
  };

  const handleCreate = () => {
    editingRecord.value = null;
    modalVisible.value = true;
  };

  const handleEdit = (record: LangRow) => {
    editingRecord.value = record.raw;
    modalVisible.value = true;
  };

  const parseFilename = (disposition?: string, lang = 'export') => {
    if (!disposition) return `i18n-${lang}.json`;
    const utf8Match = disposition.match(/filename\*=UTF-8''([^;]+)/i);
    if (utf8Match?.[1]) {
      try {
        return decodeURIComponent(utf8Match[1]);
      } catch {
        return utf8Match[1];
      }
    }
    const plainMatch = disposition.match(/filename="?([^";]+)"?/i);
    return plainMatch?.[1] || `i18n-${lang}.json`;
  };

  const downloadTextFile = (content: string, filename: string) => {
    const blob = new Blob([content], {
      type: 'application/json;charset=utf-8',
    });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    link.click();
    URL.revokeObjectURL(url);
  };

  const rawToText = async (raw: unknown) => {
    if (raw instanceof Blob) return raw.text();
    if (raw instanceof ArrayBuffer) {
      return new TextDecoder('utf-8').decode(raw);
    }
    if (typeof raw === 'string') return raw;
    if (raw == null) return '';
    return JSON.stringify(raw);
  };

  const handleExport = async (record: LangRow) => {
    const lang = record.lang?.trim();
    if (!lang || lang === '--') {
      Message.warning(t('i18nManage.message.exportNeedLang'));
      return;
    }
    exportingId.value = record.id;
    try {
      const res = (await exportI18n({ lang })) as {
        data?: unknown;
        headers?: Record<string, string>;
      };
      const text = await rawToText(res?.data);
      if (!text) {
        Message.error(t('i18nManage.message.exportFailed'));
        return;
      }
      try {
        const json = JSON.parse(text) as {
          code?: number;
          msg?: string;
          i18n_items?: unknown;
        };
        if (
          typeof json.code === 'number' &&
          json.code !== 0 &&
          !Array.isArray(json.i18n_items)
        ) {
          Message.error(json.msg || t('i18nManage.message.exportFailed'));
          return;
        }
      } catch {
        // 非 JSON 仍按文件下载
      }
      const filename = parseFilename(
        res?.headers?.['content-disposition'],
        lang
      );
      downloadTextFile(text, filename);
      Message.success(t('i18nManage.message.exportSuccess'));
    } catch (error: any) {
      const data = error?.response?.data;
      if (data instanceof Blob || data instanceof ArrayBuffer) {
        try {
          const text = await rawToText(data);
          const json = JSON.parse(text) as { msg?: string };
          Message.error(json.msg || t('i18nManage.message.exportFailed'));
        } catch {
          Message.error(t('i18nManage.message.exportFailed'));
        }
      } else {
        Message.error(
          error?.response?.data?.msg ||
            error?.message ||
            t('i18nManage.message.exportFailed')
        );
      }
    } finally {
      exportingId.value = null;
    }
  };

  const handleImport = (record: LangRow) => {
    const input = importInputRef.value;
    if (!input) return;
    importingId.value = record.id;
    importingLang.value = record.raw.lang || '';
    input.value = '';
    input.click();
    // 取消选文件时 change 不会触发，聚焦回来后清掉 loading 态
    const clearIfCancelled = () => {
      window.setTimeout(() => {
        if (!input.files?.length && importingId.value === record.id) {
          importingId.value = null;
          importingLang.value = '';
        }
      }, 300);
      window.removeEventListener('focus', clearIfCancelled);
    };
    window.addEventListener('focus', clearIfCancelled);
  };

  const onImportFileChange = async (event: Event) => {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) {
      importingId.value = null;
      importingLang.value = '';
      return;
    }
    if (!file.name.toLowerCase().endsWith('.json')) {
      Message.warning(t('i18nManage.message.importNeedJson'));
      input.value = '';
      importingId.value = null;
      importingLang.value = '';
      return;
    }
    if (file.size > 8 * 1024 * 1024) {
      Message.warning(t('i18nManage.message.importTooLarge'));
      input.value = '';
      importingId.value = null;
      importingLang.value = '';
      return;
    }
    try {
      const formData = new FormData();
      formData.append('file', file);
      formData.append('lang', importingLang.value);
      const { data } = await importI18n(formData);
      Message.success(
        t('i18nManage.message.importSuccess', {
          created: data?.created ?? 0,
          updated: data?.updated ?? 0,
          skipped: data?.skipped ?? 0,
        })
      );
    } finally {
      importingId.value = null;
      importingLang.value = '';
      input.value = '';
    }
  };

  const handleToggleDisabled = async (record: LangRow, enabled: boolean) => {
    if (record.isDefault && !enabled) {
      Message.warning(t('i18nLangManage.message.cannotDisableDefault'));
      return;
    }
    togglingId.value = record.id;
    try {
      await updateI18nLang({
        id: record.id,
        disabled: enabled ? 0 : 1,
      });
      Message.success(
        enabled
          ? t('i18nLangManage.message.enableSuccess')
          : t('i18nLangManage.message.disableSuccess')
      );
      await fetchData(pagination.current, pagination.pageSize);
    } finally {
      togglingId.value = null;
    }
  };

  const handleSetDefault = (record: LangRow) => {
    Modal.confirm({
      title: t('i18nLangManage.setDefault.title'),
      content: t('i18nLangManage.setDefault.content', {
        name: record.name,
        lang: record.lang,
      }),
      okText: t('i18nLangManage.setDefault.okText'),
      async onOk() {
        settingDefaultId.value = record.id;
        try {
          await updateI18nLang({
            id: record.id,
            is_default: 1,
          });
          Message.success(t('i18nLangManage.message.setDefaultSuccess'));
          await fetchData(pagination.current, pagination.pageSize);
        } finally {
          settingDefaultId.value = null;
        }
      },
    });
  };

  const handleDelete = (record: LangRow) => {
    if (record.isDefault) {
      Message.warning(t('i18nLangManage.message.cannotDeleteDefault'));
      return;
    }
    Modal.confirm({
      title: t('front.common.delete.title'),
      content: t('i18nLangManage.delete.content', {
        name: record.name,
        lang: record.lang,
      }),
      okText: t('front.common.delete.okText'),
      okButtonProps: { status: 'danger' },
      async onOk() {
        deletingId.value = record.id;
        try {
          await deleteI18nLang({ ids: [record.id] });
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

  const handleBatchDelete = () => {
    const ids = selectedKeys.value
      .map((key) => Number(key))
      .filter((id) => Number.isFinite(id));
    if (!ids.length) return;

    const hasDefault = list.value.some(
      (item) => ids.includes(item.id) && item.isDefault
    );
    if (hasDefault) {
      Message.warning(t('i18nLangManage.message.cannotDeleteDefault'));
      return;
    }

    Modal.confirm({
      title: t('i18nLangManage.batchDelete.title'),
      content: t('i18nLangManage.batchDelete.content', { count: ids.length }),
      okText: t('front.common.delete.okText'),
      okButtonProps: { status: 'danger' },
      async onOk() {
        await deleteI18nLang({ ids });
        Message.success(t('front.common.message.deleteSuccess'));
        const nextPage =
          list.value.length <= ids.length && pagination.current > 1
            ? pagination.current - 1
            : pagination.current;
        await fetchData(nextPage, pagination.pageSize);
      },
    });
  };

  /**
   * 语言排序由后端 /i18n/lang/reorder 接口维护，仅需传入被移动的语言 id
   * 与目标位置原本所在的语言 target_id，接口会自动处理排序值。
   */
  const persistDragOrder = async (
    movedId: number,
    targetId: number,
    rows: LangRow[]
  ) => {
    const { current: page, pageSize } = pagination;
    isDragPersisting.value = true;
    setLoading(true);
    try {
      await reorderI18nLang({ id: movedId, target_id: targetId });
      Message.success(t('i18nLangManage.message.reorderSuccess'));
    } catch {
      list.value = rows;
      Message.error(t('i18nLangManage.message.reorderFailed'));
    } finally {
      try {
        await fetchData(page, pageSize);
        await nextTick();
      } finally {
        isDragPersisting.value = false;
      }
    }
  };

  const destroyRowSortable = () => {
    rowSortable?.destroy();
    rowSortable = null;
  };

  const initRowSortable = async () => {
    destroyRowSortable();
    if (isFiltering.value || loading.value) return;
    await nextTick();
    const root = tableRef.value?.$el as HTMLElement | undefined;
    const tbody = root?.querySelector(
      '.arco-table-body tbody, .arco-table-content-inner tbody, tbody'
    ) as HTMLElement | null;
    if (!tbody) return;

    rowSortable = new Sortable(tbody, {
      handle: '.row-drag-handle',
      animation: 150,
      draggable: '.arco-table-tr',
      filter: '.arco-table-tr-empty',
      preventOnFilter: true,
      onMove: () => !isFiltering.value,
      onStart: () => {
        isRowDragging.value = true;
      },
      onEnd: (evt) => {
        const { oldIndex, newIndex } = evt;
        if (
          oldIndex == null ||
          newIndex == null ||
          oldIndex === newIndex ||
          isFiltering.value
        ) {
          isRowDragging.value = false;
          return;
        }
        const original = [...list.value];
        if (oldIndex >= original.length || newIndex >= original.length) {
          isRowDragging.value = false;
          return;
        }
        const movedId = original[oldIndex].id;
        const targetId = original[newIndex].id;

        const nextList = [...original];
        const [moved] = nextList.splice(oldIndex, 1);
        nextList.splice(newIndex, 0, moved);
        list.value = nextList;

        persistDragOrder(movedId, targetId, original).finally(() => {
          isRowDragging.value = false;
        });
      },
    });
  };

  watch(
    [
      loading,
      isFiltering,
      () => list.value.map((item) => item.id).join(','),
    ],
    ([isLoading]) => {
      if (isLoading) {
        destroyRowSortable();
        return;
      }
      initRowSortable();
    }
  );

  onBeforeUnmount(() => {
    destroyRowSortable();
  });

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
  .search-input {
    width: 200px;
  }
  .search-select {
    width: 160px;
  }
  .table-toolbar {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 12px;
    margin-bottom: 14px;
  }
  .drag-tip {
    padding: 4px 12px;
  }
  .row-drag-handle {
    color: var(--color-text-3);
    cursor: grab;
    font-size: 16px;
  }
  .row-drag-handle:active {
    cursor: grabbing;
  }
  .row-drag-handle--disabled {
    cursor: not-allowed;
    opacity: 0.4;
  }
  .is-row-dragging :deep(.arco-pagination) {
    pointer-events: none;
  }
  .operation-links {
    white-space: nowrap;
  }
  .hidden-file-input {
    display: none;
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
