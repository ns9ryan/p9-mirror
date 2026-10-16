<template>
  <a-modal
    :visible="visible"
    :title="$t('passwordModal.title')"
    :ok-loading="submitting"
    unmount-on-close
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
      <a-form-item :label="$t('passwordModal.field.password')" field="password" required>
        <a-input-password
          v-model="form.password"
          :placeholder="$t('passwordModal.field.passwordPlaceholder')"
          allow-clear
          :max-length="64"
        />
      </a-form-item>
      <a-form-item :label="$t('passwordModal.field.confirm')" field="confirm_password" required>
        <a-input-password
          v-model="form.confirm_password"
          :placeholder="$t('passwordModal.field.passwordPlaceholder')"
          allow-clear
          :max-length="64"
        />
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FormInstance } from '@arco-design/web-vue';
  import { changeUserPassword } from '@/api/user';

  const { t } = useI18n();

  const props = defineProps<{
    visible: boolean;
    userId?: number | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  const formRef = ref<FormInstance>();
  const submitting = ref(false);

  const form = reactive({
    password: '',
    confirm_password: '',
  });

  const resetForm = () => {
    form.password = '';
    form.confirm_password = '';
    formRef.value?.clearValidate();
  };

  watch(
    () => props.visible,
    (visible) => {
      if (visible) resetForm();
    }
  );

  const rules = {
    password: [{ required: true, message: t('passwordModal.rule.password') }],
    confirm_password: [
      { required: true, message: t('passwordModal.rule.confirm') },
      {
        validator: (value: string, callback: (error?: string) => void) => {
          if (value && value !== form.password) {
            callback(t('passwordModal.rule.mismatch'));
            return;
          }
          callback();
        },
      },
    ],
  };

  const handleCancel = () => {
    emit('update:visible', false);
  };

  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    if (!props.userId) {
      done(false);
      return;
    }
    const errors = await formRef.value?.validate();
    if (errors || submitting.value) {
      done(false);
      return;
    }

    submitting.value = true;
    try {
      await changeUserPassword({
        user_id: props.userId,
        password: form.password,
      });
      Message.success(t('passwordModal.message.success'));
      resetForm();
      emit('success');
      emit('update:visible', false);
      done(true);
    } catch {
      done(false);
    } finally {
      submitting.value = false;
    }
  };
</script>
