<template>
  <a-modal
    :visible="visible"
    :title="$t('loginLog.detail.title')"
    :footer="false"
    unmount-on-close
    :width="640"
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
      <a-descriptions-item :label="$t('loginLog.detail.id')">
        {{ record.id }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('loginLog.detail.username')">
        {{ record.username || '--' }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('loginLog.detail.userId')">
        {{ record.user_id }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('loginLog.detail.loginIp')">
        {{ record.login_ip || '--' }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('loginLog.detail.deviceId')">
        {{ record.device_id }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('loginLog.detail.result')">
        <a-tag :color="resultColor(record.login_result)">
          {{ record.login_result || '--' }}
        </a-tag>
      </a-descriptions-item>
      <a-descriptions-item :label="$t('loginLog.detail.failureReason')" :span="2">
        {{ record.failure_reason || '--' }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('loginLog.detail.userAgent')" :span="2">
        <pre class="detail-pre">{{ record.user_agent || '--' }}</pre>
      </a-descriptions-item>
      <a-descriptions-item :label="$t('loginLog.detail.loginAt')" :span="2">
        {{ formatTime(record.login_at) }}
      </a-descriptions-item>
    </a-descriptions>
  </a-modal>
</template>

<script lang="ts" setup>
  import dayjs from 'dayjs';
  import type { LoginLogRecord } from '@/api/log';

  defineProps<{
    visible: boolean;
    record?: LoginLogRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
  }>();

  const formatTime = (value?: number) => {
    if (!value) return '--';
    const ms = value < 1e12 ? value * 1000 : value;
    return dayjs(ms).format('YYYY-MM-DD HH:mm:ss');
  };

  /** 登录结果：后端未提供明确枚举，按常见语义 success/fail 着色 */
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
    max-height: 200px;
    margin: 0;
    overflow: auto;
    white-space: pre-wrap;
    word-break: break-all;
  }
</style>
