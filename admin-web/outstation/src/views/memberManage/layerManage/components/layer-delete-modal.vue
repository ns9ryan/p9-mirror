<template>
  <a-modal
    :visible="visible"
    :width="460"
    hide-title
    unmount-on-close
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel layer-delete-modal"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    @cancel="close"
  >
    <template v-if="record">
      <div class="delete-header">
        <icon-exclamation-circle-fill
          class="delete-icon"
          :class="blockReason ? 'is-danger' : 'is-warning'"
        />
        <div class="delete-heading">
          <div class="delete-title">{{ title }}</div>
          <div class="delete-subtitle">{{ subtitle }}</div>
        </div>
        <icon-close class="delete-close" @click="close" />
      </div>

      <div class="delete-info">
        <div v-for="row in infoRows" :key="row.label" class="info-row">
          <span class="info-label">{{ row.label }}</span>
          <span class="info-value">{{ row.value }}</span>
        </div>
      </div>

      <div v-if="blockReason" class="blocked-reason">
        <div class="reason-title">
          <icon-info-circle-fill class="reason-icon" />
          {{ t('layerManage.deleteBlocked.reasonTitle') }}
        </div>
        <ul class="reason-list">
          <li v-for="item in reasons" :key="item">{{ item }}</li>
        </ul>
      </div>
      <div v-else class="delete-tip">
        <icon-info-circle-fill class="tip-icon" />
        <span>{{ confirmTip }}</span>
      </div>
    </template>
    <template #footer>
      <a-button v-if="blockReason" @click="close">
        {{ t('layerManage.deleteBlocked.ok') }}
      </a-button>
      <a-space v-else :size="12">
        <a-button @click="close">
          {{ t('memberManage.confirm.cancel') }}
        </a-button>
        <a-button
          type="primary"
          status="danger"
          :loading="submitting"
          @click="handleConfirm"
        >
          {{ t('layerManage.deleteConfirm.ok') }}
        </a-button>
      </a-space>
    </template>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { formatAmount } from '../../memberList/format';
  import type { LayerRecord } from '../types';

  const props = defineProps<{
    visible: boolean;
    record: LayerRecord | null;
    /** 以 @confirm 绑定；返回 false 或抛错时弹窗保持打开 */
    onConfirm: (
      value: LayerRecord
    ) => boolean | void | Promise<boolean | void>;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
  }>();

  const { t } = useI18n();
  const submitting = ref(false);

  const blockReason = computed(() => {
    const record = props.record;
    if (!record) return null;
    if (record.isDefault) return 'default';
    if (record.isSystem) return 'system';
    return null;
  });

  const title = computed(() => {
    if (blockReason.value) {
      return t(`layerManage.deleteBlocked.${blockReason.value}.title`);
    }
    return t(`layerManage.deleteConfirm.${props.record?.kind}.title`);
  });

  const subtitle = computed(() => {
    if (blockReason.value) {
      return t(`layerManage.deleteBlocked.${blockReason.value}.subtitle`);
    }
    return t('layerManage.deleteConfirm.subtitle');
  });

  const infoRows = computed(() => {
    const record = props.record;
    if (!record) return [];
    const rows = [
      { label: t('layerManage.column.name'), value: record.name },
      {
        label: t(
          record.kind === 'fixed'
            ? 'layerManage.column.linkedCount'
            : 'layerManage.column.count'
        ),
        value: formatCount(record.memberCount),
      },
    ];
    if (record.kind === 'fixed') {
      const withdrawLimit = {
        label: t('layerManage.column.withdrawLimit'),
        value:
          record.withdrawLimit == null
            ? t('layerManage.withdraw.unlimited')
            : formatAmount(record.withdrawLimit),
      };
      if (!blockReason.value) return [...rows, withdrawLimit];
      return [
        ...rows,
        {
          label: t('layerManage.column.description'),
          value: record.description || '--',
        },
        withdrawLimit,
      ];
    }
    return [
      ...rows,
      {
        label: t('layerManage.form.deposit'),
        value: formatAmount(record.deposit ?? 0),
      },
      {
        label: t('layerManage.form.wash'),
        value: formatAmount(record.wash ?? 0),
      },
    ];
  });

  const reasons = computed(() => {
    const record = props.record;
    if (!record || !blockReason.value) return [];
    const count = formatCount(record.memberCount);
    const prefix = `layerManage.deleteBlocked.${blockReason.value}`;
    return [t(`${prefix}.reason`, { count }), t(`${prefix}.hint`)];
  });

  const confirmTip = computed(() => {
    const record = props.record;
    if (!record) return '';
    if (record.memberCount > 0) {
      return t(`layerManage.deleteConfirm.${record.kind}.tipWithMembers`, {
        count: formatCount(record.memberCount),
      });
    }
    return t(`layerManage.deleteConfirm.${record.kind}.tip`);
  });

  function formatCount(value: number) {
    return value.toLocaleString('en-US');
  }

  function close() {
    emit('update:visible', false);
  }

  async function handleConfirm() {
    if (!props.record || submitting.value) return;
    submitting.value = true;
    try {
      if ((await props.onConfirm(props.record)) !== false) close();
    } catch {
      // 错误提示由请求拦截器统一弹出，这里保持弹窗打开
    } finally {
      submitting.value = false;
    }
  }
</script>

<style scoped lang="less">
  .delete-header {
    display: flex;
    gap: 12px;
    align-items: flex-start;
    margin-bottom: 16px;
  }

  .delete-icon {
    flex-shrink: 0;
    font-size: 24px;

    &.is-danger {
      color: rgb(var(--danger-6));
    }

    &.is-warning {
      color: rgb(var(--warning-6));
    }
  }

  .delete-heading {
    flex: 1;
    min-width: 0;
  }

  .delete-title {
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 16px;
    line-height: 24px;
  }

  .delete-subtitle {
    margin-top: 4px;
    color: var(--color-text-3);
    font-size: 12px;
  }

  .delete-close {
    flex-shrink: 0;
    color: var(--color-text-2);
    font-size: 14px;
    cursor: pointer;

    &:hover {
      color: var(--color-text-1);
    }
  }

  .delete-info {
    padding: 12px 16px;
    background: var(--color-fill-2);
    border-radius: 6px;
  }

  .info-row {
    display: flex;
    gap: 16px;
    font-size: 12px;
    line-height: 20px;

    & + & {
      margin-top: 6px;
    }
  }

  .info-label {
    flex-shrink: 0;
    width: 84px;
    color: var(--color-text-3);
  }

  .info-value {
    flex: 1;
    min-width: 0;
    color: var(--color-text-1);
    word-break: break-all;
  }

  .blocked-reason {
    margin-top: 16px;
    padding: 12px 16px;
    background: rgb(var(--primary-1));
    border-radius: 6px;
  }

  .reason-title {
    display: flex;
    gap: 8px;
    align-items: center;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 13px;
  }

  .reason-icon {
    color: rgb(var(--primary-6));
    font-size: 16px;
  }

  .reason-list {
    margin: 8px 0 0;
    padding-left: 40px;
    color: var(--color-text-2);
    font-size: 12px;
    line-height: 20px;

    li + li {
      margin-top: 4px;
    }
  }

  .delete-tip {
    display: flex;
    gap: 6px;
    align-items: flex-start;
    margin-top: 12px;
    color: var(--color-text-3);
    font-size: 12px;
    line-height: 18px;
  }

  .tip-icon {
    flex-shrink: 0;
    margin-top: 2px;
    color: rgb(var(--primary-6));
    font-size: 14px;
  }
</style>
