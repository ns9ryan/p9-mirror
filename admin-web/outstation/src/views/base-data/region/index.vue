<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.route.baseData', 'menu.route.region']" />
    <a-card class="general-card" :bordered="false">
      <a-form :model="searchForm" layout="inline" class="search-form">
        <a-form-item :label="$t('regionManage.search.status')" field="status">
          <a-select
            v-model="searchForm.status"
            :placeholder="$t('front.common.placeholder.select')"
            allow-clear
            class="search-select"
          >
            <a-option
              v-for="item in statusOptions"
              :key="item.value"
              :value="item.value"
            >
              {{ item.label }}
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
          <!-- 新的后端接口不再支持新建国家地区，故「新建国家地区」按钮已移除。 -->
          <!-- <a-button type="primary" @click="handleCreate">
            <template #icon><icon-plus /></template>
            {{ $t('front.common.toolbar.add') }}
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
          {{ $t('regionManage.drag.tipFilter') }}
        </a-alert>
        <a-alert v-else type="info" :show-icon="false" class="drag-tip">
          {{ $t('regionManage.drag.tipOk') }}
          <icon-drag-arrow />
        </a-alert>
      </div>
      <a-table
        ref="tableRef"
        row-key="id"
        :loading="loading"
        :columns="columns"
        :data="displayList"
        :pagination="pagination"
        :bordered="true"
        :stripe="true"
        :scroll="{ x: 1000 }"
        :class="{
          'is-row-dragging': isRowDragging || isDragPersisting,
        }"
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
        <template #callingCode="{ record }">
          <a-input
            v-model="record.callingCode"
            size="small"
            :placeholder="$t('regionManage.column.callingCodePlaceholder')"
            :max-length="3"
            :disabled="callingCodeUpdatingId === record.id"
            class="calling-code-input"
            @blur="() => handleCallingCodeChange(record)"
            @press-enter="handleCallingCodePressEnter"
          >
            <template #prefix>+</template>
          </a-input>
        </template>
        <template #status="{ record }">
          <a-switch
            :model-value="record.status === 1"
            :checked-text="$t('front.common.status.enabled')"
            :unchecked-text="$t('front.common.status.disabled')"
            size="small"
            :loading="statusUpdatingId === record.id"
            @change="(value) => handleStatusChange(record, Boolean(value))"
          />
        </template>
        <!-- 操作列暂时隐藏
        <template #operations="{ record }">
          <a-space :size="12" class="operation-links">
            <a-link @click="handleEdit(record)">{{ $t('front.common.action.edit') }}</a-link>
          </a-space>
        </template>
        -->
      </a-table>
    </a-card>
    <!-- 操作列隐藏后，编辑弹窗暂不挂载
    <RegionModal
      :visible="modalVisible"
      :record="editingRegion"
      @update:visible="onModalVisibleChange"
      @success="fetchData"
    />
    -->
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
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import { Message } from '@arco-design/web-vue';
  import Sortable from 'sortablejs';
  import {
    // getRegion,
    queryRegionList,
    reorderRegion,
    updateRegion,
    type RegionRecord,
    type RegionListParams,
  } from '@/api/region';
  import useLoading from '@/hooks/loading';
  // import RegionModal from './components/region-modal.vue';

  interface RegionRow {
    id: number;
    code: string;
    callingCode: string;
    nameKey: string;
    name: string;
    sortNo: number;
    status: number;
    raw: RegionRecord;
  }

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(true);
  // const modalVisible = ref(false);
  // const editingRegion = ref<RegionRecord | null>(null);
  const statusUpdatingId = ref<number | null>(null);
  const callingCodeUpdatingId = ref<number | null>(null);
  const tableRef = ref();
  let rowSortable: Sortable | null = null;
  /** 拖拽落库中，禁止分页跳转 */
  const isDragPersisting = ref(false);
  /** 行拖拽进行中（防止松手误点到底部分页） */
  const isRowDragging = ref(false);

  const statusOptions = computed(() => [
    { label: t('front.common.status.enabled'), value: 1 },
    { label: t('front.common.status.disabled'), value: 2 },
  ]);

  const searchForm = reactive<{
    status: number | undefined;
  }>({
    status: undefined,
  });
  const appliedSearch = reactive({
    status: undefined as number | undefined,
  });
  /** 存在筛选条件时，当前页并非连续的排序区间，禁止拖拽排序 */
  const isFiltering = computed(() => appliedSearch.status !== undefined);

  const regions = ref<RegionRow[]>([]);
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

  const mapRegion = (item: RegionRecord): RegionRow => ({
    id: item.id,
    code: item.code || '--',
    callingCode: item.calling_code || '',
    nameKey: item.name_key || '--',
    name: item.name || '--',
    sortNo: item.sort_no ?? 0,
    status: item.status,
    raw: item,
  });

  const displayList = computed(() => regions.value);

  const buildQueryParams = (
    page = pagination.current,
    pageSize = pagination.pageSize
  ): RegionListParams => {
    const params: RegionListParams = {
      page,
      page_size: pageSize,
    };
    if (appliedSearch.status !== undefined)
      params.status = appliedSearch.status;
    return params;
  };

  const fetchData = async (
    page = pagination.current,
    pageSize = pagination.pageSize
  ) => {
    setLoading(true);
    try {
      const { data } = await queryRegionList(buildQueryParams(page, pageSize));
      regions.value = (data?.list || []).map(mapRegion);
      pagination.current = page;
      pagination.pageSize = pageSize;
      pagination.total = data?.total ?? 0;
    } catch {
      regions.value = [];
      pagination.total = 0;
    } finally {
      setLoading(false);
    }
  };

  const columns = computed<TableColumnData[]>(() => [
    {
      title: t('regionManage.column.drag'),
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
      title: t('regionManage.column.code'),
      dataIndex: 'code',
      width: 110,
      align: 'center',
    },
    {
      title: t('regionManage.column.callingCode'),
      dataIndex: 'callingCode',
      slotName: 'callingCode',
      width: 120,
      align: 'center',
    },
    {
      title: t('regionManage.column.name'),
      dataIndex: 'name',
      width: 160,
      align: 'center',
    },
    // 这列暂时没有用到，先隐藏
    // {
    //   title: '名称 Key',
    //   dataIndex: 'nameKey',
    //   width: 200,
    //   align: 'center',
    //   ellipsis: true,
    //   tooltip: true,
    // },
    // 这列暂时没有用到，先隐藏
    // {
    //   title: t('regionManage.column.sortNo'),
    //   dataIndex: 'sortNo',
    //   width: 90,
    //   align: 'center',
    // },
    {
      title: t('front.common.column.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 120,
      align: 'center',
    },
    // 操作列暂时隐藏
    // {
    //   title: t('front.common.column.operations'),
    //   dataIndex: 'operations',
    //   slotName: 'operations',
    //   width: 100,
    //   align: 'center',
    //   fixed: 'right',
    // },
  ]);

  // const onModalVisibleChange = (visible: boolean) => {
  //   modalVisible.value = visible;
  //   if (!visible) {
  //     editingRegion.value = null;
  //   }
  // };

  // const handleEdit = async (record: RegionRow) => {
  //   setLoading(true);
  //   try {
  //     const { data } = await getRegion({ id: record.id });
  //     editingRegion.value = data?.id ? data : record.raw;
  //     modalVisible.value = true;
  //   } catch {
  //     // 全局拦截器已提示错误
  //   } finally {
  //     setLoading(false);
  //   }
  // };

  /** 仅保留数字，去掉 + 等符号 */
  const normalizeCallingCode = (value: string) =>
    String(value || '')
      .trim()
      .replace(/^\+/, '')
      .replace(/\D/g, '')
      .slice(0, 3);

  const handleCallingCodePressEnter = (e: KeyboardEvent) => {
    (e.target as HTMLInputElement)?.blur();
  };

  const handleCallingCodeChange = async (record: RegionRow) => {
    if (callingCodeUpdatingId.value === record.id) return;
    const nextCode = normalizeCallingCode(record.callingCode);
    const prevCode = record.raw.calling_code || '';
    if (!nextCode) {
      record.callingCode = prevCode;
      Message.warning(t('regionManage.message.callingCodeRequired'));
      return;
    }
    if (nextCode === prevCode) {
      record.callingCode = prevCode;
      return;
    }
    callingCodeUpdatingId.value = record.id;
    try {
      await updateRegion({ id: record.id, calling_code: nextCode });
      record.callingCode = nextCode;
      record.raw.calling_code = nextCode;
      Message.success(t('regionManage.message.callingCodeUpdated'));
    } catch {
      record.callingCode = prevCode;
    } finally {
      callingCodeUpdatingId.value = null;
    }
  };

  const handleStatusChange = async (record: RegionRow, enabled: boolean) => {
    if (statusUpdatingId.value === record.id) return;
    const nextStatus = enabled ? 1 : 2;
    if (record.status === nextStatus) return;
    statusUpdatingId.value = record.id;
    try {
      await updateRegion({ id: record.id, status: nextStatus });
      record.status = nextStatus;
      record.raw.status = nextStatus;
      Message.success(
        enabled
          ? t('front.common.message.enabled')
          : t('regionManage.message.disabled')
      );
    } catch {
      // 保持原状态；错误提示由拦截器处理
    } finally {
      statusUpdatingId.value = null;
    }
  };

  const handleSearch = () => {
    appliedSearch.status = searchForm.status;
    fetchData(1, pagination.pageSize);
  };

  const handleReset = () => {
    searchForm.status = undefined;
    appliedSearch.status = undefined;
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

  /**
   * 国家地区排序由后端 /region/reorder 接口维护，仅需传入被移动的 id
   * 与目标位置原本所在的 target_id，接口会自动处理排序值。
   */
  const persistDragOrder = async (
    movedId: number,
    targetId: number,
    rows: RegionRow[]
  ) => {
    const { current: page, pageSize } = pagination;
    isDragPersisting.value = true;
    setLoading(true);
    try {
      await reorderRegion({ id: movedId, target_id: targetId });
      Message.success(t('regionManage.message.reorderSuccess'));
    } catch {
      regions.value = rows;
      Message.error(t('regionManage.message.reorderFail'));
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
        const original = [...regions.value];
        if (oldIndex >= original.length || newIndex >= original.length) {
          isRowDragging.value = false;
          return;
        }
        const movedId = original[oldIndex].id;
        const targetId = original[newIndex].id;

        const list = [...original];
        const [moved] = list.splice(oldIndex, 1);
        list.splice(newIndex, 0, moved);
        regions.value = list;

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
      () => regions.value.map((item) => item.id).join(','),
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
  .search-select :deep(.arco-select-view) {
    min-width: 130px;
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
  .search-form {
    row-gap: 16px;
    column-gap: 16px;
  }
  .calling-code-input {
    max-width: 110px;
  }
</style>
