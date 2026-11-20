<template>
  <a-modal
    :visible="visible"
    :title="t('syncListModal.title')"
    :footer="false"
    unmount-on-close
    :width="680"
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    @cancel="handleCancel"
  >
    <a-spin :loading="loading" style="width: 100%">
      <a-descriptions
        v-if="detail"
        :column="2"
        bordered
        :label-style="{ width: '130px' }"
      >
        <a-descriptions-item :label="t('syncListModal.field.id')">
          {{ detail.id }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('syncListModal.field.syncScope')">
          {{ detail.sync_scope || '--' }}
        </a-descriptions-item>
        <a-descriptions-item
          :label="t('syncListModal.field.checkpointValue')"
          :span="2"
        >
          {{ detail.checkpoint_value || '--' }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('syncListModal.field.progress')">
          {{ detail.progress ?? 0 }}%
        </a-descriptions-item>
        <a-descriptions-item :label="t('syncListModal.field.localTotal')">
          {{ detail.local_total ?? 0 }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('syncListModal.field.remoteTotal')">
          {{ detail.remote_total ?? 0 }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('syncListModal.field.createdCount')">
          {{ detail.created_count ?? 0 }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('syncListModal.field.updatedCount')">
          {{ detail.updated_count ?? 0 }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('syncListModal.field.deletedCount')">
          {{ detail.deleted_count ?? 0 }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('syncListModal.field.failedCount')">
          {{ detail.failed_count ?? 0 }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('syncListModal.field.syncStatus')">
          <a-tag :color="syncStatusMeta(detail.sync_status).color">
            {{ syncStatusMeta(detail.sync_status).label }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item
          :label="t('syncListModal.field.lastError')"
          :span="2"
        >
          {{ detail.last_error_message || '--' }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('syncListModal.field.lastSyncAt')">
          {{ formatTime(detail.last_sync_at) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('syncListModal.field.lastSuccessAt')">
          {{ formatTime(detail.last_success_at) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('front.common.column.createdAt')">
          {{ formatTime(detail.created_at) }}
        </a-descriptions-item>
        <a-descriptions-item :label="t('front.common.column.updatedAt')">
          {{ formatTime(detail.updated_at) }}
        </a-descriptions-item>
      </a-descriptions>
    </a-spin>
  </a-modal>
</template>

<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import dayjs from 'dayjs';
  import {
    getGameSyncCheckpointDetail,
    type GameSyncCheckpointRecord,
  } from '@/api/gameSyncCheckpoint';

  const { t } = useI18n();

  const props = defineProps<{
    visible: boolean;
    syncScope?: string | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
  }>();

  const loading = ref(false);
  const detail = ref<GameSyncCheckpointRecord | null>(null);

  const formatTime = (value?: number) => {
    if (!value) return '--';
    const ms = value < 1e12 ? value * 1000 : value;
    return dayjs(ms).format('YYYY-MM-DD HH:mm:ss');
  };

  const syncStatusMeta = (status: number) => {
    switch (status) {
      case 1:
        return { label: t('syncListManage.status.success'), color: 'green' };
      case 2:
        return { label: t('syncListManage.status.failed'), color: 'red' };
      default:
        return { label: t('syncListManage.status.pending'), color: 'gray' };
    }
  };

  const fetchDetail = async (syncScope: string) => {
    loading.value = true;
    detail.value = null;
    try {
      const { data } = await getGameSyncCheckpointDetail(syncScope);
      detail.value = data ?? null;
    } catch {
      detail.value = null;
    } finally {
      loading.value = false;
    }
  };

  const handleCancel = () => {
    emit('update:visible', false);
  };

  watch(
    () => props.visible,
    (visible) => {
      if (visible && props.syncScope) {
        fetchDetail(props.syncScope);
      }
    }
  );
</script>
