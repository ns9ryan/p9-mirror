<template>
  <a-modal
    :visible="visible"
    :title="t('manufacturerSync.title')"
    :width="760"
    :mask-closable="false"
    unmount-on-close
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    @cancel="handleCancel"
  >
    <template #footer>
      <template v-if="!applyResult">
        <a-button @click="handleCancel">{{ t('manufacturerSync.action.cancel') }}</a-button>
        <a-button
          type="primary"
          :loading="applying"
          :disabled="previewLoading || !previewResult"
          @click="handleConfirm"
        >
          {{ t('manufacturerSync.action.confirm') }}
        </a-button>
      </template>
      <template v-else>
        <a-button type="primary" @click="handleCancel">{{
          t('manufacturerSync.action.close')
        }}</a-button>
      </template>
    </template>

    <a-spin :loading="previewLoading" style="width: 100%">
      <template v-if="applyResult">
        <a-descriptions
          :title="t('manufacturerSync.apply.title')"
          :column="3"
          bordered
          :label-style="{ width: '90px' }"
        >
          <a-descriptions-item :label="t('manufacturerSync.apply.created')">
            {{ applyResult.created }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('manufacturerSync.apply.updated')">
            {{ applyResult.updated }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('manufacturerSync.apply.deleted')">
            {{ applyResult.deleted }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('manufacturerSync.apply.skipped')">
            {{ applyResult.skipped }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('manufacturerSync.apply.failed')">
            {{ applyResult.failed }}
          </a-descriptions-item>
        </a-descriptions>
      </template>
      <template v-else-if="previewResult">
        <a-descriptions
          :title="t('manufacturerSync.preview.title')"
          :column="3"
          bordered
          :label-style="{ width: '90px' }"
        >
          <a-descriptions-item :label="t('manufacturerSync.preview.localTotal')">
            {{ previewResult.stats.local_total }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('manufacturerSync.preview.remoteTotal')">
            {{ previewResult.stats.remote_total }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('manufacturerSync.preview.created')">
            {{ previewResult.stats.create_total }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('manufacturerSync.preview.updated')">
            {{ previewResult.stats.update_total }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('manufacturerSync.preview.deleted')">
            {{ previewResult.stats.delete_total }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('manufacturerSync.preview.noop')">
            {{ previewResult.stats.noop_total }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('manufacturerSync.preview.conflict')">
            {{ previewResult.stats.conflict_total }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('manufacturerSync.preview.error')">
            {{ previewResult.stats.error_total }}
          </a-descriptions-item>
        </a-descriptions>

        <a-table
          v-if="pagination.total > 0"
          class="diff-table"
          row-key="object_id"
          :data="previewResult.diffs"
          :columns="diffColumns"
          :pagination="pagination"
          :bordered="true"
          size="small"
          @page-change="onPageChange"
          @page-size-change="onPageSizeChange"
        />
        <a-empty
          v-else
          class="diff-empty"
          :description="t('manufacturerSync.diff.empty')"
        />

        <div class="sync-cols">
          <div class="sync-cols__label">
            {{ t('manufacturerSync.cols.label') }}
            <span class="sync-cols__tip">
              {{ t('manufacturerSync.cols.tip') }}
            </span>
          </div>
          <a-checkbox-group v-model="syncCols" :options="syncColOptions" />
        </div>
      </template>
    </a-spin>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import {
    previewGameProviderSync,
    runGameProviderSync,
    type GameProviderPreviewResult,
    type GameProviderSyncApplyResult,
    type GameProviderSyncCol,
  } from '@/api/gameProvider';

  const { t } = useI18n();

  const props = defineProps<{
    visible: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  const previewLoading = ref(false);
  const applying = ref(false);
  const previewResult = ref<GameProviderPreviewResult | null>(null);
  const applyResult = ref<GameProviderSyncApplyResult | null>(null);
  const syncCols = ref<GameProviderSyncCol[]>([]);

  const syncColOptions = computed(() => [
    { label: t('manufacturerSync.cols.nameI18n'), value: 'name' as const },
    { label: t('manufacturerSync.cols.sortNo'), value: 'sort_no' as const },
    {
      label: t('manufacturerSync.cols.sourceProviderCode'),
      value: 'source_provider_code' as const,
    },
    { label: t('manufacturerSync.cols.sourceId'), value: 'source_id' as const },
    {
      label: t('manufacturerSync.cols.sourceNameI18n'),
      value: 'source_name_i18n' as const,
    },
    {
      label: t('manufacturerSync.cols.sourceStatus'),
      value: 'source_status' as const,
    },
    { label: t('manufacturerSync.cols.status'), value: 'status' as const },
  ]);

  const pagination = reactive({
    current: 1,
    pageSize: 10,
    total: 0,
    showTotal: true,
    showPageSize: true,
    pageSizeOptions: [10, 20, 30, 50, 100],
  });

  const diffColumns = computed<TableColumnData[]>(() => [
    {
      title: t('manufacturerSync.column.action'),
      dataIndex: 'action',
      width: 90,
      align: 'center',
    },
    {
      title: t('manufacturerSync.column.objectCode'),
      dataIndex: 'object_code',
      width: 130,
      align: 'center',
    },
    {
      title: t('manufacturerSync.column.remoteCode'),
      dataIndex: 'remote_code',
      width: 130,
      align: 'center',
    },
    {
      title: t('manufacturerSync.column.conflictType'),
      dataIndex: 'conflict_type',
      width: 110,
      align: 'center',
    },
    {
      title: t('manufacturerSync.column.reason'),
      dataIndex: 'reason',
      width: 150,
      align: 'center',
    },
    {
      title: t('manufacturerSync.column.details'),
      dataIndex: 'details',
      ellipsis: true,
      tooltip: true,
    },
  ]);

  const fetchPreview = async (
    page = pagination.current,
    pageSize = pagination.pageSize
  ) => {
    previewLoading.value = true;
    try {
      const { data } = await previewGameProviderSync({
        page,
        page_size: pageSize,
        is_skip: true,
      });
      previewResult.value = data ?? null;
      pagination.current = data?.page ?? page;
      pagination.pageSize = data?.page_size ?? pageSize;
      pagination.total = data?.total_count ?? 0;
    } catch {
      previewResult.value = null;
      pagination.total = 0;
      Message.error(t('manufacturerSync.message.previewFail'));
    } finally {
      previewLoading.value = false;
    }
  };

  const onPageChange = (page: number) => {
    fetchPreview(page, pagination.pageSize);
  };

  const onPageSizeChange = (pageSize: number) => {
    fetchPreview(1, pageSize);
  };

  const handleConfirm = async () => {
    applying.value = true;
    try {
      const payload =
        syncCols.value.length > 0 ? { sync_cols: [...syncCols.value] } : {};
      const { data } = await runGameProviderSync(payload);
      if (data?.apply) {
        applyResult.value = data.apply;
      }
      Message.success(t('manufacturerSync.message.started'));
      emit('success');
      emit('update:visible', false);
    } catch {
      Message.error(t('manufacturerSync.message.fail'));
    } finally {
      applying.value = false;
    }
  };

  const handleCancel = () => {
    emit('update:visible', false);
  };

  watch(
    () => props.visible,
    (visible) => {
      if (visible) {
        previewResult.value = null;
        applyResult.value = null;
        syncCols.value = [];
        pagination.current = 1;
        pagination.total = 0;
        fetchPreview(1, pagination.pageSize);
      }
    }
  );
</script>

<style scoped lang="less">
  .sync-cols {
    margin-top: 16px;

    &__label {
      margin-bottom: 8px;
      color: var(--color-text-2);
      font-size: 13px;
    }

    &__tip {
      margin-left: 8px;
      color: rgb(var(--red-6));
      font-size: 12px;
    }
  }
  .diff-table {
    margin-top: 16px;
  }
  .diff-empty {
    margin-top: 16px;
  }
</style>
