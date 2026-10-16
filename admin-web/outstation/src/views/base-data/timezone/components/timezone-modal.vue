<template>
  <a-modal
    :visible="visible"
    :title="$t('timezoneModal.title.edit')"
    :ok-loading="submitting"
    :ok-text="$t('front.common.okText.save')"
    :mask-closable="false"
    :esc-to-close="false"
    unmount-on-close
    :width="560"
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    @before-ok="handleBeforeOk"
    @cancel="handleCancel"
  >
    <a-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-align="right"
      :label-col-props="{ span: 6 }"
      :wrapper-col-props="{ span: 17 }"
    >
      <a-form-item :label="$t('timezoneModal.field.code')">
        <a-input :model-value="form.code" disabled />
        <template #extra>{{ $t('timezoneModal.field.codeExtra') }}</template>
      </a-form-item>
      <a-form-item :label="$t('timezoneModal.field.nameKey')">
        <a-input :model-value="form.nameKey" disabled />
        <template #extra>{{ $t('timezoneModal.field.nameKeyExtra') }}</template>
      </a-form-item>
      <a-form-item :label="$t('timezoneModal.field.name')">
        <a-input :model-value="form.name" disabled />
      </a-form-item>
      <a-form-item
        :label="$t('front.common.column.status')"
        field="status"
        required
      >
        <a-radio-group v-model="form.status">
          <a-radio :value="1">{{ $t('front.common.status.enabled') }}</a-radio>
          <a-radio :value="2">{{ $t('front.common.status.disabled') }}</a-radio>
        </a-radio-group>
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FormInstance, FieldRule } from '@arco-design/web-vue';
  import { updateTimezone, type TimezoneRecord } from '@/api/timezone';

  const props = defineProps<{
    visible: boolean;
    record?: TimezoneRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  const { t } = useI18n();
  const formRef = ref<FormInstance>();
  const submitting = ref(false);

  interface TimezoneForm {
    code: string;
    nameKey: string;
    name: string;
    status: number;
  }

  const getDefaultForm = (): TimezoneForm => ({
    code: '',
    nameKey: '',
    name: '',
    status: 1,
  });

  const form = reactive<TimezoneForm>(getDefaultForm());

  const rules = computed<Record<string, FieldRule | FieldRule[]>>(() => ({
    status: [{ required: true, message: t('timezoneModal.rule.status') }],
  }));

  const resetForm = () => {
    Object.assign(form, getDefaultForm());
    formRef.value?.clearValidate();
  };

  const fillForm = (record: TimezoneRecord) => {
    Object.assign(form, {
      code: record.code || '',
      nameKey: record.name_key || '',
      name: record.name || '',
      status: record.status ?? 1,
    });
  };

  const close = () => {
    emit('update:visible', false);
  };

  const handleCancel = () => {
    close();
    resetForm();
  };

  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    const errors = await formRef.value?.validate();
    if (errors || submitting.value || !props.record?.id) {
      done(false);
      return;
    }

    submitting.value = true;
    try {
      await updateTimezone({
        id: props.record.id,
        status: form.status,
      });
      Message.success(t('front.common.message.saveSuccess'));
      resetForm();
      emit('success');
      done(true);
      close();
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
      resetForm();
      if (props.record) {
        fillForm(props.record);
      }
    }
  );
</script>
