<template>
  <a-modal
    :visible="visible"
    :title="
      created ? t('nodeList.create.result.title') : t('nodeList.create.title')
    "
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
    <div class="wizard">
      <a-steps
        v-if="!created"
        :current="currentStep"
        class="wizard-steps"
        small
      >
        <a-step>{{ t('nodeList.create.step.base') }}</a-step>
        <a-step>{{ t('nodeList.create.step.confirm') }}</a-step>
      </a-steps>

      <div v-if="created" class="result-body">
        <a-alert type="warning" class="secret-alert">
          {{ t('nodeList.create.result.warning') }}
        </a-alert>
        <a-descriptions
          :column="1"
          size="large"
          :label-style="{ width: '120px', color: 'var(--color-text-3)' }"
        >
          <a-descriptions-item :label="t('nodeList.column.id')">
            {{ created.id }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('nodeList.create.field.name')">
            {{ form.name }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('nodeList.column.code')">
            <div class="copy-row">
              <span class="mono">{{ created.code }}</span>
              <a-button size="small" @click="handleCopyCode">
                {{ t('nodeList.create.action.copyCode') }}
              </a-button>
            </div>
          </a-descriptions-item>
          <a-descriptions-item :label="t('nodeList.create.result.secretLabel')">
            <div class="copy-row">
              <span class="mono">{{ created.auth_secret }}</span>
              <a-button type="primary" size="small" @click="handleCopySecret">
                {{ t('nodeList.create.action.copySecret') }}
              </a-button>
            </div>
          </a-descriptions-item>
        </a-descriptions>
      </div>

      <template v-else>
        <a-form
          v-show="currentStep === 1"
          ref="formRef"
          :model="form"
          :rules="rules"
          layout="vertical"
        >
          <a-form-item
            :label="t('nodeList.create.field.name')"
            field="name"
            required
          >
            <a-input
              v-model="form.name"
              :placeholder="t('nodeList.create.field.namePlaceholder')"
              allow-clear
              :max-length="100"
              show-word-limit
            />
          </a-form-item>
          <a-form-item
            :label="t('nodeList.create.field.status')"
            field="status"
            required
          >
            <a-select v-model="form.status" :options="statusOptions" />
          </a-form-item>
          <a-form-item
            :label="t('nodeList.create.field.remark')"
            field="remark"
            required
          >
            <a-textarea
              v-model="form.remark"
              :placeholder="t('nodeList.create.field.remarkPlaceholder')"
              :max-length="1000"
              show-word-limit
              allow-clear
              :auto-size="{ minRows: 3, maxRows: 5 }"
            />
          </a-form-item>
        </a-form>

        <a-descriptions
          v-if="currentStep === 2"
          :column="1"
          size="large"
          :label-style="{ width: '120px', color: 'var(--color-text-3)' }"
        >
          <a-descriptions-item :label="t('nodeList.create.field.name')">
            {{ form.name }}
          </a-descriptions-item>
          <a-descriptions-item :label="t('nodeList.create.field.status')">
            <a-tag :color="statusMeta.color">{{ statusMeta.label }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item :label="t('nodeList.create.field.remark')">
            <div class="remark-text">{{ form.remark }}</div>
          </a-descriptions-item>
        </a-descriptions>
      </template>

      <div class="wizard-footer">
        <template v-if="created">
          <span />
          <a-button type="primary" @click="handleCancel">
            {{ t('nodeList.create.action.close') }}
          </a-button>
        </template>
        <template v-else-if="currentStep === 1">
          <a-button @click="handleCancel">
            {{ t('nodeList.create.action.cancel') }}
          </a-button>
          <a-button type="primary" @click="handleNext">
            {{ t('nodeList.create.action.next') }}
          </a-button>
        </template>
        <template v-else>
          <a-button :disabled="submitting" @click="currentStep = 1">
            {{ t('nodeList.create.action.prev') }}
          </a-button>
          <a-button type="primary" :loading="submitting" @click="handleSubmit">
            {{ t('nodeList.create.action.submit') }}
          </a-button>
        </template>
      </div>
    </div>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, Modal } from '@arco-design/web-vue';
  import type { FieldRule, FormInstance } from '@arco-design/web-vue';
  import { NODE_STATUS, createNode, type CreateNodeRes } from '@/api/node';

  const { t } = useI18n();

  const props = defineProps<{
    visible: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  const formRef = ref<FormInstance>();
  const currentStep = ref(1);
  const submitting = ref(false);
  const created = ref<CreateNodeRes | null>(null);
  const secretCopied = ref(false);

  const getDefaultForm = () => ({
    name: '',
    status: NODE_STATUS.ENABLED as number,
    remark: '',
  });

  const form = reactive(getDefaultForm());

  const statusOptions = computed(() => [
    { label: t('nodeList.status.enabled'), value: NODE_STATUS.ENABLED },
    { label: t('nodeList.status.disabled'), value: NODE_STATUS.DISABLED },
  ]);

  const statusMeta = computed(() => {
    if (form.status === NODE_STATUS.DISABLED) {
      return { label: t('nodeList.status.disabled'), color: 'red' };
    }
    return { label: t('nodeList.status.enabled'), color: 'green' };
  });

  const rules = computed<Record<string, FieldRule | FieldRule[]>>(() => ({
    name: [
      {
        validator: (value, callback) => {
          if (!String(value || '').trim()) {
            callback(t('nodeList.create.rule.nameRequired'));
            return;
          }
          callback();
        },
      },
    ],
    status: [
      {
        required: true,
        message: t('nodeList.create.rule.statusRequired'),
      },
    ],
    remark: [
      {
        validator: (value, callback) => {
          if (!String(value || '').trim()) {
            callback(t('nodeList.create.rule.remarkRequired'));
            return;
          }
          callback();
        },
      },
    ],
  }));

  const resetState = () => {
    currentStep.value = 1;
    submitting.value = false;
    created.value = null;
    secretCopied.value = false;
    Object.assign(form, getDefaultForm());
    formRef.value?.clearValidate();
  };

  const close = () => {
    created.value = null;
    secretCopied.value = false;
    emit('update:visible', false);
  };

  const handleCancel = () => {
    if (created.value && !secretCopied.value) {
      Modal.confirm({
        title: t('nodeList.create.message.closeSecretTitle'),
        content: t('nodeList.create.message.closeSecretContent'),
        onOk() {
          close();
        },
      });
      return;
    }
    close();
  };

  const handleNext = async () => {
    const errors = await formRef.value?.validate();
    if (errors) return;
    form.name = form.name.trim();
    form.remark = form.remark.trim();
    currentStep.value = 2;
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
    if (!created.value?.code) return;
    const ok = await copyText(created.value.code);
    if (!ok) {
      Message.error(t('nodeList.create.message.copyFailed'));
      return;
    }
    Message.success(t('nodeList.create.message.codeCopied'));
  };

  const handleCopySecret = async () => {
    if (!created.value?.auth_secret) return;
    const ok = await copyText(created.value.auth_secret);
    if (!ok) {
      Message.error(t('nodeList.create.message.copyFailed'));
      return;
    }
    secretCopied.value = true;
    Message.success(t('nodeList.create.message.copied'));
  };

  const handleSubmit = async () => {
    if (submitting.value) return;
    submitting.value = true;
    try {
      const { data } = await createNode({
        name: form.name.trim(),
        status: form.status,
        remark: form.remark.trim(),
      });
      if (!data?.auth_secret) {
        Message.error(t('nodeList.create.message.secretMissing'));
        emit('success');
        close();
        return;
      }
      created.value = {
        id: data.id,
        code: data.code,
        auth_secret: data.auth_secret,
      };
      secretCopied.value = false;
      emit('success');
    } catch {
      // 业务错误由拦截器提示
    } finally {
      submitting.value = false;
    }
  };

  watch(
    () => props.visible,
    (visible) => {
      if (visible) resetState();
    }
  );
</script>

<style scoped lang="less">
  .wizard-steps {
    width: 320px;
    margin: 0 auto 24px;
  }
  .secret-alert {
    margin-bottom: 16px;
  }
  .result-body :deep(.arco-descriptions) {
    width: 100%;
  }
  .result-body :deep(.arco-descriptions-table) {
    width: 100%;
    table-layout: fixed;
  }
  .result-body :deep(.arco-descriptions-item-value) {
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
  .remark-text {
    white-space: pre-wrap;
    word-break: break-word;
  }
  .wizard-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 8px;
  }
</style>
