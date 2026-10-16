<template>
  <a-modal
    :visible="visible"
    :title="$t('log.detail.title')"
    :footer="false"
    unmount-on-close
    :width="720"
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    @cancel="handleCancel"
  >
    <a-descriptions
      v-if="record"
      :column="2"
      bordered
      :label-style="{ width: '120px' }"
    >
      <a-descriptions-item :label="$t('log.detail.id')">
        {{ record.id }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('log.detail.username')">
        {{ record.username || '--' }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('log.detail.userId')">
        {{ record.user_id }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('log.detail.clientIp')">
        {{ record.client_ip || '--' }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('log.detail.requestMethod')">
        <a-tag :color="methodColor(record.request_method)">
          {{ record.request_method || '--' }}
        </a-tag>
      </a-descriptions-item>
      <a-descriptions-item :label="$t('log.detail.responseStatus')">
        <a-tag :color="statusColor(record.response_status)">
          {{ record.response_status }}
        </a-tag>
      </a-descriptions-item>
      <a-descriptions-item :label="$t('log.detail.requestPath')" :span="2">
        {{ record.request_path || '--' }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('log.detail.requestQuery')" :span="2">
        <pre class="detail-pre">{{ record.request_query || '--' }}</pre>
      </a-descriptions-item>
      <a-descriptions-item :label="$t('log.detail.requestBody')" :span="2">
        <pre class="detail-pre">{{ record.request_body || '--' }}</pre>
      </a-descriptions-item>
      <a-descriptions-item :label="$t('log.detail.responseBody')" :span="2">
        <pre class="detail-pre">{{ record.response_body || '--' }}</pre>
      </a-descriptions-item>
      <a-descriptions-item :label="$t('log.detail.actionResult')">
        <a-tag :color="resultColor(record.action_result)">
          {{ record.action_result || '--' }}
        </a-tag>
      </a-descriptions-item>
      <a-descriptions-item :label="$t('log.detail.durationMs')">
        {{ record.duration_ms }} ms
      </a-descriptions-item>
      <a-descriptions-item :label="$t('log.detail.userAgent')" :span="2">
        <pre class="detail-pre">{{ record.user_agent || '--' }}</pre>
      </a-descriptions-item>
      <a-descriptions-item :label="$t('log.detail.createdAt')" :span="2">
        {{ formatTime(record.created_at) }}
      </a-descriptions-item>
    </a-descriptions>
  </a-modal>
</template>

<script lang="ts" setup>
  import dayjs from 'dayjs';
  import type { OperationLogRecord } from '@/api/log';

  defineProps<{
    visible: boolean;
    record?: OperationLogRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
  }>();

  const formatTime = (value?: number) => {
    if (!value) return '--';
    const ms = value < 1e12 ? value * 1000 : value;
    return dayjs(ms).format('YYYY-MM-DD HH:mm:ss');
  };

  const methodColor = (method?: string) => {
    const colors: Record<string, string> = {
      GET: 'arcoblue',
      POST: 'green',
      PUT: 'orange',
      DELETE: 'red',
      PATCH: 'purple',
    };
    return colors[(method || '').toUpperCase()] || 'gray';
  };

  const statusColor = (status?: number) => {
    if (!status) return 'gray';
    if (status >= 200 && status < 300) return 'green';
    if (status >= 400) return 'red';
    return 'orange';
  };

  const resultColor = (result?: string) => {
    const value = (result || '').toLowerCase();
    if (['success', 'ok', '1', '成功'].includes(value)) return 'green';
    if (['fail', 'failed', 'error', '2', '失败'].includes(value)) return 'red';
    return 'gray';
  };

  const handleCancel = () => {
    emit('update:visible', false);
  };
</script>

<style scoped lang="less">
  .detail-pre {
    max-height: 220px;
    margin: 0;
    padding: 8px;
    overflow: auto;
    white-space: pre-wrap;
    word-break: break-all;
    background-color: var(--color-fill-1);
    border-radius: 4px;
    font-size: 12px;
    line-height: 1.6;
  }
</style>
