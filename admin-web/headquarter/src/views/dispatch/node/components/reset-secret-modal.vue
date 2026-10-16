<template>
  <a-modal
    :visible="visible"
    :title="t('nodeList.reset.title')"
    :width="640"
    :footer="false"
    :mask-closable="false"
    :esc-to-close="false"
    unmount-on-close
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    @cancel="handleCancel"
  >
    <a-alert type="warning" class="secret-alert">
      {{ t('nodeList.reset.warning') }}
    </a-alert>
    <a-descriptions
      v-if="snapshot"
      :column="1"
      size="large"
      :label-style="{ width: '120px', color: 'var(--color-text-3)' }"
    >
      <a-descriptions-item :label="t('nodeList.column.id')">
        {{ snapshot.id }}
      </a-descriptions-item>
      <a-descriptions-item :label="t('nodeList.create.field.name')">
        {{ snapshot.name || '--' }}
      </a-descriptions-item>
      <a-descriptions-item :label="t('nodeList.column.code')">
        <span class="mono">{{ snapshot.code || '--' }}</span>
      </a-descriptions-item>
      <a-descriptions-item :label="t('nodeList.create.result.secretLabel')">
        <div class="copy-row">
          <span class="mono">{{ snapshot.authSecret }}</span>
          <a-button type="primary" size="small" @click="handleCopySecret">
            {{ t('nodeList.create.action.copySecret') }}
          </a-button>
        </div>
      </a-descriptions-item>
    </a-descriptions>
    <div class="modal-footer">
      <span />
      <a-button type="primary" @click="handleCancel">
        {{ t('nodeList.create.action.close') }}
      </a-button>
    </div>
  </a-modal>
</template>

<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, Modal } from '@arco-design/web-vue';

  export interface ResetSecretResult {
    id: number;
    name: string;
    code: string;
    authSecret: string;
  }

  const { t } = useI18n();

  const props = defineProps<{
    visible: boolean;
    record?: ResetSecretResult | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
  }>();

  const snapshot = ref<ResetSecretResult | null>(null);
  const secretCopied = ref(false);

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

  const close = () => {
    snapshot.value = null;
    secretCopied.value = false;
    emit('update:visible', false);
  };

  const handleCancel = () => {
    if (snapshot.value && !secretCopied.value) {
      Modal.confirm({
        title: t('nodeList.reset.message.closeTitle'),
        content: t('nodeList.reset.message.closeContent'),
        onOk() {
          close();
        },
      });
      return;
    }
    close();
  };

  const handleCopySecret = async () => {
    if (!snapshot.value?.authSecret) return;
    const ok = await copyText(snapshot.value.authSecret);
    if (!ok) {
      Message.error(t('nodeList.create.message.copyFailed'));
      return;
    }
    secretCopied.value = true;
    Message.success(t('nodeList.create.message.copied'));
  };

  watch(
    () => props.visible,
    (visible) => {
      if (!visible || !props.record?.authSecret) return;
      snapshot.value = { ...props.record };
      secretCopied.value = false;
    }
  );
</script>

<style scoped lang="less">
  .secret-alert {
    margin-bottom: 16px;
  }
  :deep(.arco-descriptions) {
    width: 100%;
  }
  :deep(.arco-descriptions-table) {
    width: 100%;
    table-layout: fixed;
  }
  :deep(.arco-descriptions-item-value) {
    word-break: break-all;
  }
  .copy-row {
    display: flex;
    align-items: flex-start;
    gap: 12px;
  }
  .copy-row :deep(.arco-btn) {
    flex-shrink: 0;
  }
  .mono {
    flex: 1;
    min-width: 0;
    color: var(--color-text-1);
    font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
    line-height: 1.6;
    word-break: break-all;
  }
  .modal-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 16px;
  }
</style>
