<template>
  <a-modal
    :visible="visible"
    :title="t('nodeList.detail.title')"
    :width="1180"
    :footer="false"
    unmount-on-close
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    @cancel="close"
  >
    <a-tabs v-model:active-key="activeTab">
      <a-tab-pane key="base" :title="t('nodeList.detail.tab.base')">
        <a-spin :loading="loading" style="width: 100%">
          <div v-if="detail" class="info-list">
            <div class="info-row">
              <div class="info-label">{{ t('nodeList.column.id') }}</div>
              <div class="info-value">{{ detail.id }}</div>
            </div>
            <div class="info-row">
              <div class="info-label">{{ t('nodeList.column.code') }}</div>
              <div class="info-value">
                <span class="code-row">
                  <span>{{ detail.code || '--' }}</span>
                  <a-button
                    v-if="detail.code"
                    type="text"
                    size="mini"
                    @click="handleCopyCode"
                  >
                    <template #icon><icon-copy /></template>
                  </a-button>
                </span>
              </div>
            </div>
            <div class="info-row">
              <div class="info-label">{{ t('nodeList.column.name') }}</div>
              <div class="info-value">{{ detail.name || '--' }}</div>
            </div>
            <div class="info-row">
              <div class="info-label">{{ t('nodeList.column.status') }}</div>
              <div class="info-value">
                <a-tag :color="statusMeta.color">{{ statusMeta.label }}</a-tag>
              </div>
            </div>
            <div class="info-row">
              <div class="info-label">
                {{ t('nodeList.column.onlineStatus') }}
              </div>
              <div class="info-value">
                <span class="status-dot-text">
                  <span class="status-dot" :class="onlineMeta.dotClass" />
                  {{ onlineMeta.label }}
                </span>
              </div>
            </div>
            <div class="info-row">
              <div class="info-label">
                {{ t('nodeList.column.lastActiveAt') }}
              </div>
              <div class="info-value">{{
                formatTime(detail.last_seen_at)
              }}</div>
            </div>
            <div class="info-row">
              <div class="info-label">
                {{ t('nodeList.create.field.remark') }}
              </div>
              <div class="info-value remark-text">
                {{ detail.remark || '--' }}
              </div>
            </div>
            <div class="info-row">
              <div class="info-label">
                {{ t('front.common.column.createdAt') }}
              </div>
              <div class="info-value">{{ formatTime(detail.created_at) }}</div>
            </div>
            <div class="info-row">
              <div class="info-label">
                {{ t('front.common.column.updatedAt') }}
              </div>
              <div class="info-value">{{ formatTime(detail.updated_at) }}</div>
            </div>
          </div>
        </a-spin>
      </a-tab-pane>
      <a-tab-pane key="tasks" :title="t('nodeList.detail.tab.tasks')">
        <a-form :model="taskSearch" layout="inline" class="task-search">
          <a-form-item
            :label="t('nodeList.detail.task.keyword')"
            field="keyword"
          >
            <a-input
              v-model="taskSearch.keyword"
              :placeholder="t('nodeList.detail.task.keywordPlaceholder')"
              allow-clear
              class="task-keyword"
              @press-enter="handleTaskSearch"
            />
          </a-form-item>
          <a-form-item :label="t('nodeList.detail.task.status')" field="status">
            <a-select
              v-model="taskSearch.status"
              :placeholder="t('front.common.all')"
              allow-clear
              class="task-select"
              :options="taskStatusOptions"
            />
          </a-form-item>
          <a-form-item :label="t('nodeList.detail.task.type')" field="type">
            <a-select
              v-model="taskSearch.type"
              :placeholder="t('front.common.all')"
              allow-clear
              class="task-select"
              :options="taskTypeOptions"
            />
          </a-form-item>
          <a-form-item>
            <a-space>
              <a-button type="primary" @click="handleTaskSearch">
                <template #icon><icon-search /></template>
                {{ t('front.common.search.search') }}
              </a-button>
              <a-button @click="handleTaskReset">
                <template #icon><icon-refresh /></template>
                {{ t('front.common.search.reset') }}
              </a-button>
            </a-space>
          </a-form-item>
        </a-form>
        <a-table
          row-key="rowKey"
          :loading="taskLoading"
          :columns="taskColumns"
          :data="taskList"
          :pagination="taskPagination"
          :bordered="true"
          :stripe="true"
          :scroll="{ x: 1080 }"
          @page-change="onTaskPageChange"
          @page-size-change="onTaskPageSizeChange"
        >
          <template #status="{ record }">
            <a-tag :color="taskStatusMeta(record.status).color">
              {{ taskStatusMeta(record.status).label }}
            </a-tag>
          </template>
          <template #finishedAt="{ record }">
            {{ record.finishedAt || '-' }}
          </template>
          <template #operations>
            <a-link>{{ t('nodeList.action.detail') }}</a-link>
          </template>
        </a-table>
      </a-tab-pane>
    </a-tabs>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import dayjs from 'dayjs';
  import { NODE_STATUS, queryNodeDetail, type NodeDetail } from '@/api/node';
  import {
    DISPATCH_TASK_STATUS,
    DISPATCH_TASK_TYPE,
    queryDispatchTaskList,
    type DispatchTaskListParams,
    type DispatchTaskRecord,
  } from '@/api/dispatch-task';

  interface NodeTaskRow {
    rowKey: string;
    taskNo: string;
    requestNo: string;
    type: string;
    status: number | string;
    startedAt: string;
    finishedAt: string;
  }

  const { t } = useI18n();

  const props = defineProps<{
    visible: boolean;
    nodeId?: number | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
  }>();

  const activeTab = ref('base');
  const loading = ref(false);
  const detail = ref<NodeDetail | null>(null);
  let detailRequestId = 0;

  const taskSearch = reactive({
    keyword: '',
    status: undefined as number | undefined,
    type: undefined as string | undefined,
  });
  const appliedTaskSearch = reactive({
    keyword: '',
    status: undefined as number | undefined,
    type: undefined as string | undefined,
  });
  const taskLoading = ref(false);
  const taskList = ref<NodeTaskRow[]>([]);
  const taskPagination = reactive({
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
  let taskRequestId = 0;

  const taskStatusOptions = computed(() => [
    {
      label: t('nodeList.detail.task.status.pending'),
      value: DISPATCH_TASK_STATUS.PENDING,
    },
    {
      label: t('nodeList.detail.task.status.running'),
      value: DISPATCH_TASK_STATUS.RUNNING,
    },
    {
      label: t('nodeList.detail.task.status.success'),
      value: DISPATCH_TASK_STATUS.SUCCESS,
    },
    {
      label: t('nodeList.detail.task.status.failed'),
      value: DISPATCH_TASK_STATUS.FAILED,
    },
  ]);
  const taskTypeOptions = computed(() => [
    {
      label: t('nodeList.detail.task.type.createOperator'),
      value: DISPATCH_TASK_TYPE.CREATE_OPERATOR,
    },
    {
      label: t('nodeList.detail.task.type.updateOperator'),
      value: DISPATCH_TASK_TYPE.UPDATE_OPERATOR,
    },
    {
      label: t('nodeList.detail.task.type.deleteOperator'),
      value: DISPATCH_TASK_TYPE.DELETE_OPERATOR,
    },
  ]);

  const taskColumns = computed<TableColumnData[]>(() => [
    {
      title: t('nodeList.detail.task.column.taskNo'),
      dataIndex: 'taskNo',
      width: 190,
    },
    {
      title: t('nodeList.detail.task.column.requestNo'),
      dataIndex: 'requestNo',
      width: 190,
    },
    {
      title: t('nodeList.detail.task.column.type'),
      dataIndex: 'type',
      width: 170,
    },
    {
      title: t('nodeList.detail.task.column.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 110,
    },
    {
      title: t('nodeList.detail.task.column.startedAt'),
      dataIndex: 'startedAt',
      width: 180,
    },
    {
      title: t('nodeList.detail.task.column.finishedAt'),
      dataIndex: 'finishedAt',
      slotName: 'finishedAt',
      width: 180,
    },
    {
      title: t('nodeList.column.operations'),
      dataIndex: 'operations',
      slotName: 'operations',
      width: 80,
    },
  ]);

  const statusMeta = computed(() => {
    if (detail.value?.status === NODE_STATUS.DISABLED) {
      return { label: t('nodeList.status.disabled'), color: 'red' };
    }
    return { label: t('nodeList.status.enabled'), color: 'gold' };
  });

  const onlineMeta = computed(() => {
    if (detail.value?.online) {
      return {
        label: t('nodeList.onlineStatus.online'),
        dotClass: 'status-dot--online',
      };
    }
    return {
      label: t('nodeList.onlineStatus.offline'),
      dotClass: 'status-dot--offline',
    };
  });

  const taskStatusMeta = (status: number | string) => {
    const value = Number(status);
    if (value === DISPATCH_TASK_STATUS.PENDING) {
      return {
        label: t('nodeList.detail.task.status.pending'),
        color: 'orange',
      };
    }
    if (value === DISPATCH_TASK_STATUS.RUNNING) {
      return {
        label: t('nodeList.detail.task.status.running'),
        color: 'arcoblue',
      };
    }
    if (value === DISPATCH_TASK_STATUS.SUCCESS) {
      return {
        label: t('nodeList.detail.task.status.success'),
        color: 'green',
      };
    }
    if (value === DISPATCH_TASK_STATUS.FAILED) {
      return { label: t('nodeList.detail.task.status.failed'), color: 'red' };
    }
    return { label: status ? String(status) : '--', color: 'gray' };
  };

  const formatTime = (value?: number | null) => {
    if (!value) return '--';
    const ms = value < 1e12 ? value * 1000 : value;
    return dayjs(ms).format('YYYY-MM-DD HH:mm:ss');
  };

  const close = () => {
    emit('update:visible', false);
  };

  const copyText = async (text: string) => {
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(text);
        return true;
      }
    } catch {
      // 非安全上下文下 clipboard API 会失败，改用选区复制
    }
    try {
      const input = document.createElement('textarea');
      input.value = text;
      input.setAttribute('readonly', 'true');
      input.style.position = 'fixed';
      input.style.left = '-9999px';
      document.body.appendChild(input);
      input.select();
      const ok = document.execCommand('copy');
      document.body.removeChild(input);
      return ok;
    } catch {
      return false;
    }
  };

  const handleCopyCode = async () => {
    if (!detail.value?.code) return;
    const ok = await copyText(detail.value.code);
    if (!ok) {
      Message.error(t('nodeList.create.message.copyFailed'));
      return;
    }
    Message.success(t('nodeList.create.message.codeCopied'));
  };

  const loadDetail = async (id: number) => {
    detailRequestId += 1;
    const requestId = detailRequestId;
    loading.value = true;
    detail.value = null;
    try {
      const { data } = await queryNodeDetail(id);
      if (requestId !== detailRequestId) return;
      detail.value = data;
    } catch {
      if (requestId !== detailRequestId) return;
      close();
    } finally {
      if (requestId === detailRequestId) loading.value = false;
    }
  };

  const textOf = (item: DispatchTaskRecord, keys: string[]) => {
    const raw = item as Record<string, unknown>;
    const matched = keys
      .map((key) => raw[key])
      .find((value) => value !== undefined && value !== null && value !== '');
    return matched === undefined ? '' : String(matched);
  };

  const timeOf = (item: DispatchTaskRecord, keys: string[]) => {
    const raw = item as Record<string, unknown>;
    const matched = keys
      .map((key) => raw[key])
      .find((value) => value !== undefined && value !== null && value !== '');
    if (typeof matched === 'number') return formatTime(matched);
    if (typeof matched === 'string') {
      const num = Number(matched);
      if (matched && !Number.isNaN(num) && num > 0) return formatTime(num);
      return matched;
    }
    return '';
  };

  const mapTaskRow = (item: DispatchTaskRecord, index: number): NodeTaskRow => {
    const taskNo = textOf(item, ['task_no', 'task_code', 'code']);
    const id = Number(textOf(item, ['id']) || 0);
    return {
      rowKey: id ? String(id) : taskNo || String(index),
      taskNo: taskNo || '--',
      requestNo: textOf(item, ['request_no', 'request_code', 'req_no']) || '--',
      type: textOf(item, ['task_type', 'type']) || '--',
      status: item.status ?? '',
      startedAt: timeOf(item, ['started_at', 'start_at', 'begin_at']) || '--',
      finishedAt: timeOf(item, [
        'finished_at',
        'ended_at',
        'end_at',
        'finish_at',
      ]),
    };
  };

  const fetchTasks = async (
    page = taskPagination.current,
    pageSize = taskPagination.pageSize
  ) => {
    const nodeCode = detail.value?.code;
    if (!nodeCode) return;
    taskRequestId += 1;
    const requestId = taskRequestId;
    taskLoading.value = true;
    try {
      const params: DispatchTaskListParams = {
        page,
        page_size: pageSize,
        node_code: nodeCode,
      };
      const keyword = appliedTaskSearch.keyword.trim();
      if (keyword) params.keyword = keyword;
      if (appliedTaskSearch.status !== undefined) {
        params.status = appliedTaskSearch.status;
      }
      if (appliedTaskSearch.type) params.task_type = appliedTaskSearch.type;
      const { data } = await queryDispatchTaskList(params);
      if (requestId !== taskRequestId) return;
      taskList.value = (data?.list || []).map(mapTaskRow);
      taskPagination.current = page;
      taskPagination.pageSize = pageSize;
      taskPagination.total = data?.total ?? 0;
    } catch {
      if (requestId !== taskRequestId) return;
      taskList.value = [];
      taskPagination.total = 0;
    } finally {
      if (requestId === taskRequestId) taskLoading.value = false;
    }
  };

  const handleTaskSearch = () => {
    appliedTaskSearch.keyword = taskSearch.keyword;
    appliedTaskSearch.status = taskSearch.status;
    appliedTaskSearch.type = taskSearch.type;
    fetchTasks(1, taskPagination.pageSize);
  };

  const clearTaskQuery = () => {
    taskSearch.keyword = '';
    taskSearch.status = undefined;
    taskSearch.type = undefined;
    appliedTaskSearch.keyword = '';
    appliedTaskSearch.status = undefined;
    appliedTaskSearch.type = undefined;
    taskList.value = [];
    taskPagination.current = 1;
    taskPagination.total = 0;
  };

  const handleTaskReset = () => {
    taskSearch.keyword = '';
    taskSearch.status = undefined;
    taskSearch.type = undefined;
    appliedTaskSearch.keyword = '';
    appliedTaskSearch.status = undefined;
    appliedTaskSearch.type = undefined;
    fetchTasks(1, taskPagination.pageSize);
  };

  const onTaskPageChange = (current: number) => {
    fetchTasks(current, taskPagination.pageSize);
  };

  const onTaskPageSizeChange = (pageSize: number) => {
    fetchTasks(1, pageSize);
  };

  watch(
    () => props.visible,
    (visible) => {
      if (!visible) {
        detailRequestId += 1;
        taskRequestId += 1;
        detail.value = null;
        loading.value = false;
        taskLoading.value = false;
        activeTab.value = 'base';
        clearTaskQuery();
        return;
      }
      if (props.nodeId) loadDetail(props.nodeId);
    }
  );

  watch(
    () =>
      props.visible && activeTab.value === 'tasks'
        ? detail.value?.code || ''
        : '',
    (code) => {
      if (!code) return;
      fetchTasks(1, taskPagination.pageSize);
    }
  );
</script>

<style scoped lang="less">
  .info-list {
    padding: 8px 8px 0;
  }
  .info-row {
    display: flex;
    align-items: flex-start;
    padding: 10px 0;
    line-height: 22px;
  }
  .info-label {
    flex-shrink: 0;
    width: 120px;
    color: var(--color-text-3);
  }
  .info-value {
    flex: 1;
    min-width: 0;
    color: var(--color-text-1);
  }
  .code-row {
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }
  .remark-text {
    white-space: pre-wrap;
    word-break: break-word;
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
  .status-dot--online {
    background-color: rgb(var(--success-6));
  }
  .status-dot--offline {
    background-color: rgb(var(--gray-5));
  }
  .task-search {
    margin-bottom: 16px;
    row-gap: 12px;
  }
  .task-search :deep(.arco-form-item) {
    margin-bottom: 0;
  }
  .task-keyword {
    width: 260px;
  }
  .task-select {
    width: 160px;
  }
  :deep(.arco-table-th) {
    color: var(--color-text-1);
    font-weight: 600;
    background-color: var(--color-fill-2);
  }
  :deep(.arco-pagination-options .arco-select) {
    min-width: 100px;
  }
</style>
