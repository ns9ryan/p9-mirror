<template>
  <div class="container list-page">
    <a-card class="general-card tabs-card" :bordered="false">
      <a-tabs
        v-model:active-key="activeTab"
        class="task-tabs"
        hide-content
        :header-padding="false"
      >
        <a-tab-pane
          key="list"
          :title="t('discountCenter.taskCenter.tabs.list')"
        />
        <a-tab-pane
          key="claims"
          :title="t('discountCenter.taskCenter.tabs.claims')"
        />
      </a-tabs>
    </a-card>

    <a-card v-if="activeTab === 'list'" class="general-card" :bordered="false">
      <a-form :model="form" layout="inline" class="search-form">
        <a-form-item :label="t('discountCenter.taskCenter.search.category')">
          <a-select
            v-model="form.category"
            class="search-select"
            :options="categoryOptions"
            allow-clear
            :placeholder="t('discountCenter.taskCenter.search.placeholder')"
          />
        </a-form-item>
        <a-form-item :label="t('discountCenter.taskCenter.search.status')">
          <a-select
            v-model="form.status"
            class="search-select"
            :options="statusOptions"
            allow-clear
            :placeholder="t('discountCenter.taskCenter.search.placeholder')"
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

    <a-card
      v-if="activeTab === 'list'"
      class="general-card table-card"
      :bordered="false"
    >
      <div class="table-header">
        <div class="list-title">
          {{ t('discountCenter.taskCenter.toolbar.title') }}
        </div>
        <a-space :size="12">
          <a-button type="primary" @click="configVisible = true">
            <template #icon><icon-settings /></template>
            {{ t('discountCenter.taskCenter.toolbar.config') }}
          </a-button>
          <a-button
            :disabled="!total"
            :loading="exporting"
            @click="handleExport"
          >
            <template #icon><icon-download /></template>
            {{ t('discountCenter.taskCenter.toolbar.export') }}
          </a-button>
        </a-space>
      </div>
      <a-table
        class="task-table"
        row-key="id"
        :columns="columns"
        :data="tasks"
        :loading="loading || sorting"
        :pagination="false"
        :bordered="false"
        :scroll="{ x: 1500 }"
        :draggable="{ type: 'row' }"
        @mousedown.capture="dragArmed = false"
        @dragstart.capture="guardDragStart"
        @change="onTableChange"
      >
        <template #sort="{ record }">
          <div class="sort-actions">
            <span
              class="drag-handle"
              :title="t('discountCenter.taskCenter.sort.drag')"
              @mousedown="dragArmed = true"
            >
              <icon-drag-dot-vertical />
            </span>
            <a-tooltip :content="t('discountCenter.taskCenter.sort.top')">
              <a-button
                size="mini"
                type="text"
                shape="circle"
                :disabled="isFirst(record)"
                @click="moveTop(record)"
              >
                <template #icon><icon-to-top /></template>
              </a-button>
            </a-tooltip>
          </div>
        </template>
        <template #name="{ record }">
          <span class="task-name">{{ record.name }}</span>
        </template>
        <template #reward="{ record }">
          <span class="amount">{{ formatAmount(record.reward_amount) }}</span>
        </template>
        <template #completion="{ record }">
          {{ formatAmount(record.completion_amount) }}
        </template>
        <template #category="{ record }">
          <a-tag :color="CATEGORY_COLORS[record.category as TaskCategory]">
            {{ t(`discountCenter.taskCenter.category.${record.category}`) }}
          </a-tag>
        </template>
        <template #payoutMode="{ record }">
          <a-tag :color="PAYOUT_MODE_COLORS[record.payout_mode as TaskPayoutMode]">
            {{ t(`discountCenter.taskCenter.payoutMode.${record.payout_mode}`) }}
          </a-tag>
        </template>
        <template #enabled="{ record }">
          <a-switch
            :model-value="record.status === TASK_STATUS.ENABLED"
            :loading="toggling.has(record.id)"
            :checked-text="t('discountCenter.taskCenter.switch.on')"
            :unchecked-text="t('discountCenter.taskCenter.switch.off')"
            @change="(value) => toggleEnabled(record, value === true)"
          />
        </template>
        <template #updatedAt="{ record }">
          {{ formatTime(record.updated_at) }}
        </template>
        <template #operations="{ record }">
          <a-link class="operation-link" @click="openEdit(record)">
            {{ t('discountCenter.taskCenter.action.edit') }}
          </a-link>
        </template>
        <template #empty>
          <a-empty :description="t('discountCenter.taskCenter.empty')" />
        </template>
      </a-table>
      <div class="table-pagination">
        <a-pagination
          v-model:current="pageNo"
          v-model:page-size="pageSize"
          :total="total"
          :page-size-options="[10,15, 30, 50, 100]"
          :page-size-props="{ popupContainer: 'body' }"
          show-total
          show-page-size
        />
      </div>
    </a-card>

    <TaskClaims v-else />

    <TaskConfigModal
      v-model:visible="configVisible"
      :config="taskConfig"
      @submit="onConfigSubmit"
    />
    <TaskEditModal
      v-model:visible="editVisible"
      :record="editing"
      @saved="onEditSaved"
    />
  </div>
</template>

<script lang="ts" setup>
  import { computed, onMounted, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { SelectOptionData } from '@arco-design/web-vue/es/select/interface';
  import type {
    TableChangeExtra,
    TableColumnData,
    TableData,
  } from '@arco-design/web-vue/es/table/interface';
  import {
    queryTaskConfigList,
    TASK_CATEGORIES,
    TASK_STATUS,
    updateTaskConfig,
    type TaskCategory,
    type TaskConfigItem,
    type TaskConfigListParams,
    type TaskPayoutMode,
  } from '@/api/task';
  import TaskClaims from './components/task-claims.vue';
  import TaskConfigModal from './components/task-config-modal.vue';
  import TaskEditModal from './components/task-edit-modal.vue';
  import exportTasksCsv, { fetchAllPages } from './export';
  import {
    CATEGORY_COLORS,
    PAYOUT_MODE_COLORS,
    formatAmount,
    formatTime,
  } from './format';
  import type { TaskConfig } from './types';

  interface QueryState {
    category: TaskCategory | '';
    status: number | '';
  }

  const { t } = useI18n();
  const activeTab = ref<'list' | 'claims'>('list');
  const tasks = ref<TaskConfigItem[]>([]);
  const total = ref(0);
  const loading = ref(false);
  const sorting = ref(false);
  const exporting = ref(false);
  const toggling = reactive(new Set<number>());
  const form = reactive<QueryState>({ category: '', status: '' });
  const applied = ref<QueryState>({ ...form });
  const pageNo = ref(1);
  const pageSize = ref(10);
  const dragArmed = ref(false);

  const configVisible = ref(false);
  const taskConfig = ref<TaskConfig | null>(null);
  const editVisible = ref(false);
  const editing = ref<TaskConfigItem | null>(null);

  const categoryOptions = computed<SelectOptionData[]>(() =>
    TASK_CATEGORIES.map((value) => ({
      label: t(`discountCenter.taskCenter.category.${value}`),
      value,
    }))
  );

  const statusOptions = computed<SelectOptionData[]>(() =>
    [TASK_STATUS.ENABLED, TASK_STATUS.DISABLED].map((value) => ({
      label: t(`discountCenter.taskCenter.status.${value}`),
      value,
    }))
  );

  function filterParams(): Omit<TaskConfigListParams, 'page' | 'page_size'> {
    const { category, status } = applied.value;
    return {
      category: category || undefined,
      status: status || undefined,
    };
  }

  let ticket = 0;
  async function load() {
    ticket += 1;
    const current = ticket;
    loading.value = true;
    try {
      const { data } = await queryTaskConfigList({
        page: pageNo.value,
        page_size: pageSize.value,
        ...filterParams(),
      });
      if (current !== ticket) return;
      tasks.value = data?.list || [];
      total.value = data?.total || 0;
      const maxPage = Math.max(1, Math.ceil(total.value / pageSize.value));
      if (pageNo.value > maxPage) pageNo.value = maxPage;
    } catch {
      if (current !== ticket) return;
      tasks.value = [];
      total.value = 0;
    } finally {
      if (current === ticket) loading.value = false;
    }
  }

  watch([pageNo, pageSize], load);
  onMounted(load);

  const columns = computed<TableColumnData[]>(() => [
    {
      title: t('discountCenter.taskCenter.column.sort'),
      slotName: 'sort',
      width: 96,
      align: 'center',
    },
    {
      title: t('discountCenter.taskCenter.column.taskId'),
      dataIndex: 'id',
      width: 80,
      align: 'center',
    },
    {
      title: t('discountCenter.taskCenter.column.name'),
      slotName: 'name',
      width: 160,
    },
    {
      title: t('discountCenter.taskCenter.column.reward'),
      slotName: 'reward',
      width: 120,
      align: 'right',
    },
    {
      title: t('discountCenter.taskCenter.column.completionAmount'),
      slotName: 'completion',
      width: 140,
      align: 'right',
    },
    {
      title: t('discountCenter.taskCenter.column.category'),
      slotName: 'category',
      width: 100,
      align: 'center',
    },
    {
      title: t('discountCenter.taskCenter.column.description'),
      dataIndex: 'description',
      width: 180,
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('discountCenter.taskCenter.column.payoutMode'),
      slotName: 'payoutMode',
      width: 100,
      align: 'center',
    },
    {
      title: t('discountCenter.taskCenter.column.timezone'),
      dataIndex: 'timezone',
      width: 140,
    },
    {
      title: t('discountCenter.taskCenter.column.sortValue'),
      dataIndex: 'sort',
      width: 80,
      align: 'center',
    },
    {
      title: t('discountCenter.taskCenter.column.enabled'),
      slotName: 'enabled',
      width: 100,
      align: 'center',
    },
    {
      title: t('discountCenter.taskCenter.column.updatedAt'),
      slotName: 'updatedAt',
      width: 170,
    },
    {
      title: t('discountCenter.taskCenter.column.operations'),
      slotName: 'operations',
      width: 80,
      align: 'center',
      fixed: 'right',
    },
  ]);

  function handleSearch() {
    applied.value = { ...form };
    if (pageNo.value === 1) load();
    else pageNo.value = 1;
  }

  function handleReset() {
    form.category = '';
    form.status = '';
    handleSearch();
  }

  /** 无筛选时第一页第一行就是全局 sort 最小的任务 */
  function isFirst(record: TaskConfigItem) {
    const { category, status } = applied.value;
    return (
      !category &&
      !status &&
      pageNo.value === 1 &&
      tasks.value[0]?.id === record.id
    );
  }

  function replaceTask(item: TaskConfigItem) {
    tasks.value = tasks.value.map((row) => (row.id === item.id ? item : row));
  }

  /** 逐条提交 sort 变更；接口不支持批量，任意一条失败后以服务端数据为准重新加载 */
  async function saveSorts(changes: Array<{ id: number; sort: number }>) {
    if (!changes.length) return;
    sorting.value = true;
    try {
      await Promise.all(changes.map((item) => updateTaskConfig(item)));
      Message.success(t('discountCenter.taskCenter.message.sorted'));
    } catch {
      // 错误已由拦截器提示
    } finally {
      sorting.value = false;
      load();
    }
  }

  /** 只允许从拖拽手柄发起拖动，避免在其他单元格误拖整行 */
  function guardDragStart(ev: DragEvent) {
    if (dragArmed.value) return;
    ev.preventDefault();
    ev.stopPropagation();
  }

  /**
   * 拖拽只在当前页内换位：把本页原有的 sort 值按升序依次分给新顺序的行。
   * 本页 sort 有重复时无法靠互换区分先后，改为从最小值起逐个 +1。
   */
  function onTableChange(data: TableData[], extra: TableChangeExtra) {
    dragArmed.value = false;
    if (extra.type !== 'drag') return;
    const ordered = data as TaskConfigItem[];
    let slots = tasks.value.map((item) => item.sort).sort((a, b) => a - b);
    if (new Set(slots).size !== slots.length) {
      slots = slots.map((_, index) => slots[0] + index);
    }
    const changes = ordered
      .map((item, index) => ({ id: item.id, sort: slots[index] }))
      .filter((item, index) => ordered[index].sort !== item.sort);
    tasks.value = ordered.map((item, index) => ({
      ...item,
      sort: slots[index],
    }));
    saveSorts(changes);
  }

  /** 取全局最小 sort 再减 1，排序升序所以会排到最前 */
  async function moveTop(record: TaskConfigItem) {
    sorting.value = true;
    try {
      const { data } = await queryTaskConfigList({ page: 1, page_size: 1 });
      const first = data?.list?.[0];
      sorting.value = false;
      if (!first || first.id === record.id) return;
      saveSorts([{ id: record.id, sort: first.sort - 1 }]);
    } catch {
      sorting.value = false;
    }
  }

  async function toggleEnabled(record: TaskConfigItem, enabled: boolean) {
    if (toggling.has(record.id)) return;
    toggling.add(record.id);
    const status = enabled ? TASK_STATUS.ENABLED : TASK_STATUS.DISABLED;
    try {
      const { data } = await updateTaskConfig({ id: record.id, status });
      replaceTask(data?.task_config ?? { ...record, status });
      Message.success(
        t(
          enabled
            ? 'discountCenter.taskCenter.message.enabled'
            : 'discountCenter.taskCenter.message.disabled'
        )
      );
    } catch {
      // 错误已由拦截器提示，开关保持原状态
    } finally {
      toggling.delete(record.id);
    }
  }

  function openEdit(record: TaskConfigItem) {
    editing.value = record;
    editVisible.value = true;
  }

  /** 改了 sort 会影响行顺序，直接重新拉取当前页 */
  function onEditSaved(item: TaskConfigItem) {
    if (editing.value && editing.value.sort !== item.sort) load();
    else replaceTask(item);
  }

  function onConfigSubmit(value: TaskConfig) {
    taskConfig.value = value;
    Message.success(t('discountCenter.taskCenter.message.saved'));
  }

  async function handleExport() {
    exporting.value = true;
    try {
      const params = filterParams();
      const list = await fetchAllPages(async (page, size) => {
        const { data } = await queryTaskConfigList({
          page,
          page_size: size,
          ...params,
        });
        return { total: data?.total || 0, list: data?.list || [] };
      });
      exportTasksCsv(list, t);
    } catch {
      // 错误已由拦截器提示
    } finally {
      exporting.value = false;
    }
  }
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

  .tabs-card :deep(.arco-card-body) {
    padding: 4px 20px 0;
  }

  .task-tabs {
    :deep(.arco-tabs-nav-type-line .arco-tabs-tab) {
      margin: 0 48px 0 0;
      padding: 12px 0;
      font-size: 15px;
      line-height: 22px;
    }

    :deep(.arco-tabs-nav::before) {
      display: none;
    }
  }

  .search-form :deep(.arco-form-item) {
    margin-bottom: 12px;
  }

  .search-form :deep(.arco-select-view) {
    min-width: 160px;
  }

  .table-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
  }

  .list-title {
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 16px;
  }

  .task-table :deep(.arco-table-tr-draggable) {
    cursor: default;
  }

  .sort-actions {
    display: inline-flex;
    gap: 4px;
    align-items: center;
    color: var(--color-text-3);
  }

  .drag-handle {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    color: var(--color-text-3);
    font-size: 16px;
    border-radius: 50%;
    cursor: grab;
    transition: color 0.15s, background-color 0.15s;

    &:hover {
      color: rgb(var(--primary-6));
      background: var(--color-fill-2);
    }

    &:active {
      cursor: grabbing;
    }
  }

  .task-name {
    color: var(--color-text-1);
    font-weight: 500;
  }

  .amount {
    color: rgb(var(--primary-6));
    font-variant-numeric: tabular-nums;
  }

  .operation-link {
    padding: 0;
  }

  .table-pagination {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }

  :deep(.arco-pagination-options .arco-select) {
    min-width: 100px;
  }
</style>
