<template>
  <a-modal
    :visible="visible"
    :title="t('taskList.detail.title')"
    :width="720"
    :footer="false"
    unmount-on-close
    @cancel="close"
  >
    <a-descriptions
      v-if="record"
      :column="1"
      bordered
      size="large"
      :label-style="{ width: '120px' }"
    >
      <a-descriptions-item :label="t('taskList.column.taskNo')">
        {{ record.taskNo }}
      </a-descriptions-item>
      <a-descriptions-item :label="t('taskList.column.requestNo')">
        {{ record.requestNo }}
      </a-descriptions-item>
      <a-descriptions-item :label="t('taskList.column.target')">
        {{ record.target }}
      </a-descriptions-item>
      <a-descriptions-item :label="t('taskList.column.taskType')">
        {{ record.taskType }}
      </a-descriptions-item>
      <a-descriptions-item :label="t('taskList.column.nodeCode')">
        {{ record.nodeCode }}
      </a-descriptions-item>
      <a-descriptions-item :label="t('taskList.column.status')">
        <a-tag :color="statusMeta.color">{{ statusMeta.label }}</a-tag>
      </a-descriptions-item>
      <a-descriptions-item :label="t('taskList.column.params')">
        <pre class="json-block">{{ record.paramsText }}</pre>
      </a-descriptions-item>
      <a-descriptions-item :label="t('taskList.column.result')">
        <pre class="json-block">{{ record.resultText }}</pre>
      </a-descriptions-item>
      <a-descriptions-item :label="t('taskList.column.errorMessage')">
        {{ record.errorMessage }}
      </a-descriptions-item>
      <a-descriptions-item :label="t('taskList.column.startedAt')">
        {{ record.startedAt }}
      </a-descriptions-item>
      <a-descriptions-item :label="t('taskList.column.finishedAt')">
        {{ record.finishedAt }}
      </a-descriptions-item>
      <a-descriptions-item :label="t('taskList.column.createdAt')">
        {{ record.createdAt }}
      </a-descriptions-item>
      <a-descriptions-item :label="t('taskList.column.updatedAt')">
        {{ record.updatedAt }}
      </a-descriptions-item>
    </a-descriptions>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { DISPATCH_TASK_STATUS } from '@/api/dispatch-task';

  export interface TaskRow {
    rowKey: string;
    taskNo: string;
    requestNo: string;
    target: string;
    taskType: string;
    paramsText: string;
    nodeCode: string;
    status: number | string;
    resultText: string;
    errorMessage: string;
    startedAt: string;
    finishedAt: string;
    createdAt: string;
    updatedAt: string;
  }

  const props = defineProps<{
    visible: boolean;
    record?: TaskRow | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
  }>();

  const { t } = useI18n();

  const statusMeta = computed(() => {
    const value = Number(props.record?.status);
    if (value === DISPATCH_TASK_STATUS.PENDING) {
      return { label: t('taskList.status.pending'), color: 'orange' };
    }
    if (value === DISPATCH_TASK_STATUS.RUNNING) {
      return { label: t('taskList.status.running'), color: 'arcoblue' };
    }
    if (value === DISPATCH_TASK_STATUS.SUCCESS) {
      return { label: t('taskList.status.success'), color: 'green' };
    }
    if (value === DISPATCH_TASK_STATUS.FAILED) {
      return { label: t('taskList.status.failed'), color: 'red' };
    }
    return {
      label: props.record?.status ? String(props.record.status) : '--',
      color: 'gray',
    };
  });

  const close = () => {
    emit('update:visible', false);
  };
</script>

<style scoped lang="less">
  .json-block {
    margin: 0;
    white-space: pre-wrap;
    word-break: break-word;
    font-family: inherit;
    line-height: 22px;
  }
</style>
