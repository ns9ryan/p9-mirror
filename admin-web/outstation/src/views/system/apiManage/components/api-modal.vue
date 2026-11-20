<template>
  <a-modal
    :visible="visible"
    :title="isEdit ? $t('apiModal.title.edit') : $t('apiModal.title.create')"
    :ok-loading="submitting"
    :ok-text="isEdit ? $t('front.common.okText.save') : $t('front.common.okText.create')"
    :mask-closable="false"
    :esc-to-close="false"
    unmount-on-close
    :width="580"
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
      <a-form-item :label="$t('apiModal.field.method')" field="method" required>
        <a-select v-model="form.method" :placeholder="$t('apiModal.field.methodPlaceholder')">
          <a-option v-for="item in methodOptions" :key="item" :value="item">
            {{ item }}
          </a-option>
        </a-select>
      </a-form-item>
      <a-form-item :label="$t('apiModal.field.path')" field="path" required>
        <a-input
          v-model="form.path"
          :placeholder="$t('apiModal.field.pathPlaceholder')"
          allow-clear
          :max-length="255"
        />
      </a-form-item>
      <a-form-item :label="$t('apiModal.field.description')" field="description">
        <a-input
          v-model="form.description"
          :placeholder="$t('apiModal.field.descriptionPlaceholder')"
          allow-clear
          :max-length="128"
        />
      </a-form-item>
      <a-form-item :label="$t('apiModal.field.apiGroup')" field="api_group">
        <a-input
          v-model="form.api_group"
          :placeholder="$t('apiModal.field.apiGroupPlaceholder')"
          allow-clear
          :max-length="64"
        />
      </a-form-item>
      <a-form-item :label="$t('apiModal.field.serviceName')" field="service_name">
        <a-input
          v-model="form.service_name"
          :placeholder="$t('apiModal.field.serviceNamePlaceholder')"
          allow-clear
          :max-length="64"
        />
      </a-form-item>
      <a-form-item :label="$t('apiModal.field.isRequired')" field="is_required">
        <a-radio-group v-model="form.is_required">
          <a-radio :value="0">{{ $t('apiModal.field.isRequiredNo') }}</a-radio>
          <a-radio :value="1">{{ $t('apiModal.field.isRequiredYes') }}</a-radio>
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
  import {
    createApi,
    updateApi,
    type ApiRecord,
    type CreateApiParams,
    type UpdateApiParams,
  } from '@/api/apiManage';

  const { t } = useI18n();

  const props = defineProps<{
    visible: boolean;
    api?: ApiRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  const methodOptions = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'] as const;

  const formRef = ref<FormInstance>();
  const submitting = ref(false);

  const isEdit = computed(() => Boolean(props.api?.id));

  const getDefaultForm = (): CreateApiParams => ({
    method: 'POST',
    path: '',
    description: '',
    api_group: '',
    is_required: 0,
    service_name: '',
  });

  const form = reactive<CreateApiParams>(getDefaultForm());

  const rules: Record<string, FieldRule | FieldRule[]> = {
    method: [{ required: true, message: t('apiModal.rule.method') }],
    path: [
      { required: true, message: t('apiModal.rule.path') },
      {
        match: /^\//,
        message: t('apiModal.rule.pathFormat'),
      },
    ],
  };

  const resetForm = () => {
    Object.assign(form, getDefaultForm());
    formRef.value?.clearValidate();
  };

  const fillForm = (api: ApiRecord) => {
    Object.assign(form, getDefaultForm(), {
      method: (api.method || 'POST').toUpperCase(),
      path: api.path || '',
      description: api.description || '',
      api_group: api.api_group || '',
      is_required: api.is_required ? 1 : 0,
      service_name: api.service_name || '',
    });
  };

  const close = () => {
    emit('update:visible', false);
  };

  const handleCancel = () => {
    close();
    resetForm();
  };

  const buildPayload = (): CreateApiParams => ({
    method: form.method.trim().toUpperCase(),
    path: form.path.trim(),
    description: form.description?.trim() || '',
    api_group: form.api_group?.trim() || '',
    service_name: form.service_name?.trim() || '',
    is_required: form.is_required ? 1 : 0,
  });

  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    const errors = await formRef.value?.validate();
    if (errors || submitting.value) {
      done(false);
      return;
    }

    submitting.value = true;
    try {
      const payload = buildPayload();
      if (isEdit.value && props.api?.id) {
        const updatePayload: UpdateApiParams = {
          id: props.api.id,
          ...payload,
        };
        await updateApi(updatePayload);
        Message.success(t('front.common.message.saveSuccess'));
      } else {
        await createApi(payload);
        Message.success(t('front.common.message.createSuccess'));
      }
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
      if (isEdit.value && props.api) {
        fillForm(props.api);
      }
    }
  );
</script>
