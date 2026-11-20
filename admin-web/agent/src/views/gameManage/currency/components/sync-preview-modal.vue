<template>
  <a-modal
    :visible="visible"
    :title="t('gameCurrencySync.title')"
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
        <a-button @click="handleCancel">{{ t('gameCurrencySync.action.cancel') }}</a-button>
        <a-button
          type="primary"
          :loading="applying"
          :disabled="previewLoading || !previewResult"
          @click="handleConfirm"
        >
          {{ t('gameCurrencySync.action.confirm') }}
        </a-button>
      </template>
      <template v-else>
        <a-button type="primary" @click="handleCancel">{{
          t('gameCurrencySync.action.close')
        }}</a-button>
      </template>
    </template>

    <a-spin :loading="previewLoading" style="width: 100%">
      <template v-if="applyResult">
        <a-descriptions
          :title="t('gameCurrencySync.apply.title')"
          :column="3"
          bordered
          :label-style="{ width: '90px' }"
        >
          <a-descriptions-item :label="t('gameCurrencySync.apply.created')">
            {{ applyResult.created }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('gameCurrencySync.apply.updated')">
            {{ applyResult.updated }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('gameCurrencySync.apply.deleted')">
            {{ applyResult.deleted }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('gameCurrencySync.apply.skipped')">
            {{ applyResult.skipped }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('gameCurrencySync.apply.failed')">
            {{ applyResult.failed }}
          </a-descriptions-item>
        </a-descriptions>
      </template>
      <template v-else-if="previewResult">
        <a-descriptions
          :title="t('gameCurrencySync.preview.title')"
          :column="3"
          bordered
          :label-style="{ width: '90px' }"
        >
          <a-descriptions-item :label="t('gameCurrencySync.preview.localTotal')">
            {{ previewResult.stats.local_total }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('gameCurrencySync.preview.remoteTotal')">
            {{ previewResult.stats.remote_total }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('gameCurrencySync.preview.created')">
            {{ previewResult.stats.create_total }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('gameCurrencySync.preview.updated')">
            {{ previewResult.stats.update_total }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('gameCurrencySync.preview.deleted')">
            {{ previewResult.stats.delete_total }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('gameCurrencySync.preview.noop')">
            {{ previewResult.stats.noop_total }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('gameCurrencySync.preview.conflict')">
            {{ previewResult.stats.conflict_total }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('gameCurrencySync.preview.error')">
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
          :description="t('gameCurrencySync.diff.empty')"
        />

        <div class="sync-cols">
          <div class="sync-cols__label">
            {{ t('gameCurrencySync.cols.label') }}
            <span class="sync-cols__tip">
              {{ t('gameCurrencySync.cols.tip') }}
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
    previewGameCurrencySync,
    runGameCurrencySync,
    type GameCurrencyPreviewResult,
    type GameCurrencySyncApplyResult,
    type GameCurrencySyncCol,
  } from '@/api/gameCurrency';

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
  const previewResult = ref<GameCurrencyPreviewResult | null>(null);
  const applyResult = ref<GameCurrencySyncApplyResult | null>(null);
  const syncCols = ref<GameCurrencySyncCol[]>([]);

  const syncColOptions = computed(() => [
    {
      label: t('gameCurrencySync.cols.sourceCurrencyCode'),
      value: 'source_currency_code' as const,
    },
    { label: t('gameCurrencySync.cols.sourceId'), value: 'source_id' as const },
    {
      label: t('gameCurrencySync.cols.sourceStatus'),
      value: 'source_status' as const,
    },
    { label: t('gameCurrencySync.cols.status'), value: 'status' as const },
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
      title: t('gameCurrencySync.column.action'),
      dataIndex: 'action',
      width: 90,
      align: 'center',
    },
    {
      title: t('gameCurrencySync.column.objectCode'),
      dataIndex: 'object_code',
      width: 130,
      align: 'center',
    },
    {
      title: t('gameCurrencySync.column.remoteCode'),
      dataIndex: 'remote_code',
      width: 130,
      align: 'center',
    },
    {
      title: t('gameCurrencySync.column.conflictType'),
      dataIndex: 'conflict_type',
      width: 110,
      align: 'center',
    },
    {
      title: t('gameCurrencySync.column.reason'),
      dataIndex: 'reason',
      width: 150,
      align: 'center',
    },
    {
      title: t('gameCurrencySync.column.details'),
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
      const { data } = await previewGameCurrencySync({
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
      Message.error(t('gameCurrencySync.message.previewFail'));
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
      const { data } = await runGameCurrencySync(payload);
      if (data?.apply) {
        applyResult.value = data.apply;
      }
      Message.success(t('gameCurrencySync.message.started'));
      emit('success');
      emit('update:visible', false);
    } catch {
      Message.error(t('gameCurrencySync.message.fail'));
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
