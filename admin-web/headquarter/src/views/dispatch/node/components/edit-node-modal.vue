<template>
  <a-modal
    :visible="visible"
    :title="t('nodeList.edit.title')"
    :width="640"
    :ok-loading="submitting"
    :ok-text="t('front.common.okText.save')"
    :mask-closable="false"
    :esc-to-close="false"
    unmount-on-close
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    @before-ok="handleBeforeOk"
    @cancel="handleCancel"
  >
    <a-form ref="formRef" :model="form" :rules="rules" layout="vertical">
      <a-form-item :label="t('nodeList.column.id')">
        <span class="readonly-text">{{ record?.id ?? '--' }}</span>
      </a-form-item>
      <a-form-item :label="t('nodeList.column.code')">
        <span class="readonly-text">{{ record?.code || '--' }}</span>
      </a-form-item>
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
      <a-form-item :label="t('nodeList.create.field.remark')" field="remark">
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
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FieldRule, FormInstance } from '@arco-design/web-vue';
  import { NODE_STATUS, updateNode, type UpdateNodeParams } from '@/api/node';

  export interface NodeEditRecord {
    id: number;
    code: string;
    name: string;
    status: number;
    remark: string;
  }

  const { t } = useI18n();

  const props = defineProps<{
    visible: boolean;
    record?: NodeEditRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  const formRef = ref<FormInstance>();
  const submitting = ref(false);

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
  }));

  const fillForm = (record: NodeEditRecord) => {
    Object.assign(form, {
      name: record.name || '',
      status: record.status ?? NODE_STATUS.ENABLED,
      remark: record.remark || '',
    });
  };

  const close = () => {
    emit('update:visible', false);
  };

  const handleCancel = () => {
    close();
  };

  /** 除 id 外至少传一个发生变化的字段 */
  const buildUpdateParams = (record: NodeEditRecord): UpdateNodeParams => {
    const params: UpdateNodeParams = { id: record.id };
    const name = form.name.trim();
    const remark = form.remark.trim();
    if (name !== (record.name || '')) params.name = name;
    if (form.status !== record.status) params.status = form.status;
    if (remark !== (record.remark || '')) params.remark = remark;
    return params;
  };

  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    const errors = await formRef.value?.validate();
    if (errors || submitting.value || !props.record) {
      done(false);
      return;
    }

    const params = buildUpdateParams(props.record);
    if (Object.keys(params).length <= 1) {
      Message.info(t('nodeList.edit.message.noChange'));
      done(false);
      return;
    }

    submitting.value = true;
    try {
      await updateNode(params);
      Message.success(t('front.common.message.saveSuccess'));
      emit('success');
      done(true);
    } catch {
      done(false);
    } finally {
      submitting.value = false;
    }
  };

  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return;
      Object.assign(form, getDefaultForm());
      formRef.value?.clearValidate();
      if (props.record) fillForm(props.record);
    }
  );
</script>

<style scoped lang="less">
  .readonly-text {
    color: var(--color-text-2);
    word-break: break-all;
  }
</style>
