<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.route.i18n', 'menu.route.i18nEntry']" />
    <a-card class="general-card table-card" :bordered="false">
      <a-tabs
        v-model:active-key="activeI18nCode"
        class="i18n-tabs"
        hide-content
        @change="handleTabChange"
      >
        <a-tab-pane
          v-for="tab in i18nCodeTabs"
          :key="tab.value"
          :title="t(tab.labelKey)"
        />
      </a-tabs>
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item :label="$t('i18nManage.search.group')" field="i18nGroup">
          <a-input
            v-model="searchForm.i18nGroup"
            :placeholder="$t('i18nManage.search.groupPlaceholder')"
            allow-clear
            class="search-input"
            @press-enter="handleSearch"
          />
        </a-form-item>
        <a-form-item :label="$t('i18nManage.search.lang')" field="lang">
          <a-select
            v-model="searchForm.lang"
            :placeholder="$t('i18nManage.search.langPlaceholder')"
            allow-clear
            :loading="langLoading"
            :options="langOptions"
            class="search-input"
          />
        </a-form-item>
        <a-form-item
          :label="$t('i18nManage.search.transKey')"
          field="transKey"
        >
          <a-input
            v-model="searchForm.transKey"
            :placeholder="$t('i18nManage.search.transKeyPlaceholder')"
            allow-clear
            class="search-input"
            @press-enter="handleSearch"
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
      <div class="table-toolbar">
        <a-space>
          <a-button v-if="hasPermission('i18n:create')" type="primary" @click="handleCreate">
            <template #icon><icon-plus /></template>
            {{ $t('front.common.toolbar.add') }}
          </a-button>
          <!-- 暂时没有这个功能 -->
          <!-- <a-button @click="handleUpdateByKey()">
            <template #icon><icon-edit /></template>
            {{ $t('i18nManage.toolbar.updateByKey') }}
          </a-button> -->
          <a-button @click="() => fetchData()">
            <template #icon><icon-refresh /></template>
            {{ $t('front.common.toolbar.refresh') }}
          </a-button>
          <a-button
            v-if="hasPermission('i18n:import')"
            :loading="importing"
            @click="handleImport"
          >
            <template #icon><icon-upload /></template>
            {{ $t('i18nManage.toolbar.import') }}
          </a-button>
          <a-button
            v-if="hasPermission('i18n:export')"
            :loading="exporting"
            @click="handleExport"
          >
            <template #icon><icon-download /></template>
            {{ $t('i18nManage.toolbar.export') }}
          </a-button>
          <input
            ref="importInputRef"
            type="file"
            accept=".json,application/json"
            class="hidden-file-input"
            @change="onImportFileChange"
          />
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
        <template #index="{ rowIndex }">
          {{ rowIndex + 1 + (pagination.current - 1) * pagination.pageSize }}
        </template>
        <template #transKey="{ record }">
          <a-tooltip :content="record.transKey">
            <span class="ellipsis-text">{{ record.transKey || '--' }}</span>
          </a-tooltip>
        </template>
        <template #value="{ record }">
          <a-tooltip :content="record.value">
            <span class="ellipsis-text">{{ record.value || '--' }}</span>
          </a-tooltip>
        </template>
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <a-link @click="handleEdit(record)">
              {{ $t('front.common.action.edit') }}
            </a-link>
            <a-link @click="handleUpdateByKey(record)">
              {{ $t('i18nManage.action.updateByKey') }}
            </a-link>
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
    <I18nModal
      :visible="modalVisible"
      :record="editingRecord"
      :i18n-code="activeI18nCode"
      @update:visible="onModalVisibleChange"
      @success="fetchData"
    />
    <I18nByKeyModal
      :visible="byKeyModalVisible"
      :trans-key="byKeyTransKey"
      @update:visible="onByKeyModalVisibleChange"
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
    deleteI18nByKey,
    exportI18n,
    importI18n,
    queryI18nLangList,
    queryI18nList,
    type I18nListParams,
    type I18nRecord,
  } from '@/api/i18n';
  import useLoading from '@/hooks/loading';
  import useLocale from '@/hooks/locale';
  import usePermission from '@/hooks/permission';
  import I18nModal from './components/i18n-modal.vue';
  import I18nByKeyModal from './components/i18n-by-key-modal.vue';

  const { t } = useI18n();
  const { currentLocale } = useLocale();
  const { hasPermission } = usePermission();

  type I18nCode = 'platform' | 'operator' | 'agent' | 'user';

  const i18nCodeTabs: { value: I18nCode; labelKey: string }[] = [
    { value: 'platform', labelKey: 'i18nManage.tab.platform' },
    { value: 'operator', labelKey: 'i18nManage.tab.operator' },
    { value: 'agent', labelKey: 'i18nManage.tab.agent' },
    { value: 'user', labelKey: 'i18nManage.tab.user' },
  ];

  interface I18nRow {
    id: number;
    i18nGroup: string;
    lang: string;
    transKey: string;
    value: string;
    createdAt: string;
    updatedAt: string;
    raw: I18nRecord;
  }

  const { loading, setLoading } = useLoading(true);
  const activeI18nCode = ref<I18nCode>('platform');
  const modalVisible = ref(false);
  const editingRecord = ref<I18nRecord | null>(null);
  const byKeyModalVisible = ref(false);
  const byKeyTransKey = ref('');
  const deletingId = ref<number | null>(null);
  const langLoading = ref(false);
  const langOptions = ref<{ label: string; value: string }[]>([]);
  const importing = ref(false);
  const exporting = ref(false);
  const importInputRef = ref<HTMLInputElement | null>(null);

  const searchForm = reactive({
    i18nGroup: '',
    lang: currentLocale.value,
    transKey: '',
  });
  const appliedSearch = reactive({
    i18nGroup: '',
    lang: currentLocale.value,
    transKey: '',
  });

  const list = ref<I18nRow[]>([]);
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

  const formatTime = (value?: number) => {
    if (!value) return '--';
    const ms = value < 1e12 ? value * 1000 : value;
    return dayjs(ms).format('YYYY-MM-DD HH:mm:ss');
  };

  const mapRow = (item: I18nRecord): I18nRow => ({
    id: item.id,
    i18nGroup: item.i18n_group || '--',
    lang: item.lang || '--',
    transKey: item.trans_key || '--',
    value: item.value || '--',
    createdAt: formatTime(item.created_at),
    updatedAt: formatTime(item.updated_at),
    raw: item,
  });

  const buildQueryParams = (
    page = pagination.current,
    pageSize = pagination.pageSize
  ): I18nListParams => {
    const params: I18nListParams = {
      page,
      page_size: pageSize,
      i18n_code: activeI18nCode.value,
    };
    const i18nGroup = appliedSearch.i18nGroup.trim();
    const lang = appliedSearch.lang.trim();
    const transKey = appliedSearch.transKey.trim();
    if (i18nGroup) params.i18n_group = i18nGroup;
    if (lang) params.lang = lang;
    if (transKey) params.trans_key = transKey;
    return params;
  };

  const fetchLangOptions = async () => {
    langLoading.value = true;
    try {
      const { data } = await queryI18nLangList({ page: 1, page_size: 200 });
      langOptions.value = (data?.list || []).map((item) => ({
        label: item.name ? `${item.name} (${item.lang})` : item.lang,
        value: item.lang,
      }));
    } catch {
      langOptions.value = [];
    } finally {
      langLoading.value = false;
    }
  };

  const fetchData = async (
    page = pagination.current,
    pageSize = pagination.pageSize
  ) => {
    setLoading(true);
    try {
      const { data } = await queryI18nList(buildQueryParams(page, pageSize));
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
      title: t('front.common.column.index'),
      dataIndex: 'index',
      slotName: 'index',
      width: 80,
      align: 'center',
      fixed: 'left',
    },
    {
      title: t('i18nManage.column.id'),
      dataIndex: 'id',
      width: 90,
      align: 'center',
    },
    {
      title: t('i18nManage.column.group'),
      dataIndex: 'i18nGroup',
      width: 140,
      align: 'center',
    },
    {
      title: t('i18nManage.column.lang'),
      dataIndex: 'lang',
      width: 100,
      align: 'center',
    },
    {
      title: t('i18nManage.column.transKey'),
      dataIndex: 'transKey',
      slotName: 'transKey',
      width: 220,
      align: 'center',
    },
    {
      title: t('i18nManage.column.value'),
      dataIndex: 'value',
      slotName: 'value',
      width: 260,
      align: 'center',
    },
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
      width: 220,
      align: 'center',
      fixed: 'right',
    },
  ]);

  const handleTabChange = (key: string | number) => {
    activeI18nCode.value = key as I18nCode;
    fetchData(1, pagination.pageSize);
  };

  const handleSearch = () => {
    appliedSearch.i18nGroup = searchForm.i18nGroup;
    appliedSearch.lang = searchForm.lang;
    appliedSearch.transKey = searchForm.transKey;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    Object.assign(searchForm, {
      i18nGroup: '',
      lang: currentLocale.value,
      transKey: '',
    });
    Object.assign(appliedSearch, {
      i18nGroup: '',
      lang: currentLocale.value,
      transKey: '',
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
    if (!visible) editingRecord.value = null;
  };

  const onByKeyModalVisibleChange = (visible: boolean) => {
    byKeyModalVisible.value = visible;
    if (!visible) byKeyTransKey.value = '';
  };

  const handleCreate = () => {
    editingRecord.value = null;
    modalVisible.value = true;
  };

  const handleEdit = (record: I18nRow) => {
    editingRecord.value = record.raw;
    modalVisible.value = true;
  };

  const handleUpdateByKey = (record?: I18nRow) => {
    byKeyTransKey.value = record?.transKey || '';
    byKeyModalVisible.value = true;
  };

  const handleDelete = (record: I18nRow) => {
    Modal.confirm({
      title: t('front.common.delete.title'),
      content: t('i18nManage.delete.content', { key: record.transKey }),
      okText: t('front.common.delete.okText'),
      okButtonProps: { status: 'danger' },
      async onOk() {
        deletingId.value = record.id;
        try {
          await deleteI18nByKey({
            i18n_code: record.raw.i18n_code || activeI18nCode.value,
            i18n_group: record.raw.i18n_group,
            trans_key: record.raw.trans_key,
          });
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

  const handleExport = async () => {
    const lang = appliedSearch.lang.trim();
    if (!lang) {
      Message.warning(t('i18nManage.message.exportNeedLang'));
      return;
    }
    exporting.value = true;
    try {
      const payload: {
        lang: string;
        i18n_code?: string;
        i18n_group?: string;
      } = { lang };
      if (activeI18nCode.value) payload.i18n_code = activeI18nCode.value;
      const i18nGroup = appliedSearch.i18nGroup.trim();
      if (i18nGroup) payload.i18n_group = i18nGroup;

      const res = (await exportI18n(payload)) as {
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
        // 业务错误有时也会以 JSON blob 返回
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
      exporting.value = false;
    }
  };

  const handleImport = () => {
    const input = importInputRef.value;
    if (!input) return;
    input.value = '';
    input.click();
  };

  const onImportFileChange = async (event: Event) => {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;
    if (!file.name.toLowerCase().endsWith('.json')) {
      Message.warning(t('i18nManage.message.importNeedJson'));
      input.value = '';
      return;
    }
    if (file.size > 8 * 1024 * 1024) {
      Message.warning(t('i18nManage.message.importTooLarge'));
      input.value = '';
      return;
    }
    importing.value = true;
    try {
      const formData = new FormData();
      formData.append('file', file);
      const { data } = await importI18n(formData);
      Message.success(
        t('i18nManage.message.importSuccess', {
          created: data?.created ?? 0,
          updated: data?.updated ?? 0,
          skipped: data?.skipped ?? 0,
        })
      );
      await fetchData(1, pagination.pageSize);
    } finally {
      importing.value = false;
      input.value = '';
    }
  };

  fetchLangOptions();
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
  .i18n-tabs {
    flex-shrink: 0;
    margin-bottom: 0;
  }
  .search-form {
    display: flex;
    flex-shrink: 0;
    flex-wrap: nowrap;
    align-items: center;
    margin: 12px 0;
    overflow-x: auto;
  }
  .search-form :deep(.arco-form-item) {
    margin-bottom: 0;
    margin-right: 16px;
    flex-shrink: 0;
  }
  .search-input {
    width: 200px;
  }
  .table-toolbar {
    flex-shrink: 0;
    margin-bottom: 14px;
  }
  .ellipsis-text {
    display: inline-block;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    vertical-align: bottom;
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
