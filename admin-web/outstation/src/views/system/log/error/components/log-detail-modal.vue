<template>
  <a-modal
    :visible="visible"
    :title="$t('errorLog.detail.title')"
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
      <a-descriptions-item :label="$t('errorLog.detail.id')">
        {{ record.id }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('errorLog.detail.serviceName')">
        {{ record.service_name || '--' }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('errorLog.detail.username')">
        {{ record.username || '--' }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('errorLog.detail.userId')">
        {{ record.user_id }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('errorLog.detail.operatorId')">
        {{ record.operator_id ?? '--' }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('errorLog.detail.clientIp')">
        {{ record.client_ip || '--' }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('errorLog.detail.requestMethod')">
        <a-tag :color="methodColor(record.request_method)">
          {{ record.request_method || '--' }}
        </a-tag>
      </a-descriptions-item>
      <a-descriptions-item :label="$t('errorLog.detail.responseStatus')">
        <a-tag :color="statusColor(record.response_status)">
          {{ record.response_status }}
        </a-tag>
      </a-descriptions-item>
      <a-descriptions-item :label="$t('errorLog.detail.requestPath')" :span="2">
        {{ record.request_path || '--' }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('errorLog.detail.subject')" :span="2">
        {{ record.subject || '--' }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('errorLog.detail.detail')" :span="2">
        <pre class="detail-pre">{{ record.detail || '--' }}</pre>
      </a-descriptions-item>
      <a-descriptions-item :label="$t('errorLog.detail.requestQuery')" :span="2">
        <pre class="detail-pre">{{ record.request_query || '--' }}</pre>
      </a-descriptions-item>
      <a-descriptions-item :label="$t('errorLog.detail.requestBody')" :span="2">
        <pre class="detail-pre">{{ record.request_body || '--' }}</pre>
      </a-descriptions-item>
      <a-descriptions-item :label="$t('errorLog.detail.responseBody')" :span="2">
        <pre class="detail-pre">{{ record.response_body || '--' }}</pre>
      </a-descriptions-item>
      <a-descriptions-item :label="$t('errorLog.detail.durationMs')">
        {{ record.duration_ms }} ms
      </a-descriptions-item>
      <a-descriptions-item :label="$t('errorLog.detail.createdAt')">
        {{ formatTime(record.created_at) }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('errorLog.detail.userAgent')" :span="2">
        <pre class="detail-pre">{{ record.user_agent || '--' }}</pre>
      </a-descriptions-item>
    </a-descriptions>
  </a-modal>
</template>

<script lang="ts" setup>
  import dayjs from 'dayjs';
  import type { ErrorLogRecord } from '@/api/log';

  defineProps<{
    visible: boolean;
    record?: ErrorLogRecord | null;
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
