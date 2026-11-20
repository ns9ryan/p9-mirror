<template>
  <a-modal
    :visible="visible"
    :title="
      isEdit ? $t('i18nLangModal.title.edit') : $t('i18nLangModal.title.create')
    "
    :ok-loading="submitting"
    :ok-text="
      isEdit ? $t('front.common.okText.save') : $t('front.common.okText.create')
    "
    :mask-closable="false"
    :esc-to-close="false"
    unmount-on-close
    :width="520"
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
      <a-form-item
        :label="$t('i18nLangModal.field.lang')"
        field="lang"
        required
      >
        <a-input
          v-model="form.lang"
          :placeholder="$t('i18nLangModal.field.langPlaceholder')"
          allow-clear
          :max-length="32"
        />
      </a-form-item>
      <a-form-item
        :label="$t('i18nLangModal.field.name')"
        field="name"
        required
      >
        <a-input
          v-model="form.name"
          :placeholder="$t('i18nLangModal.field.namePlaceholder')"
          allow-clear
          :max-length="64"
        />
      </a-form-item>
      <a-form-item :label="$t('i18nLangModal.field.i18nKey')" field="i18n_key">
        <a-input
          v-model="form.i18n_key"
          :placeholder="$t('i18nLangModal.field.i18nKeyPlaceholder')"
          allow-clear
          :max-length="128"
        />
      </a-form-item>
      <a-form-item :label="$t('i18nLangModal.field.disabled')" field="disabled">
        <a-radio-group
          v-model="form.disabled"
          :disabled="isEdit && Boolean(record?.is_default)"
        >
          <a-radio :value="0">
            {{ $t('i18nLangModal.disabled.enabled') }}
          </a-radio>
          <a-radio :value="1">
            {{ $t('i18nLangModal.disabled.disabled') }}
          </a-radio>
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
    createI18nLang,
    updateI18nLang,
    type CreateI18nLangParams,
    type I18nLangRecord,
  } from '@/api/i18n';

  const { t } = useI18n();

  const props = defineProps<{
    visible: boolean;
    record?: I18nLangRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  const formRef = ref<FormInstance>();
  const submitting = ref(false);

  const isEdit = computed(() => Boolean(props.record?.id));

  const getDefaultForm = (): CreateI18nLangParams => ({
    lang: '',
    name: '',
    i18n_key: '',
    disabled: 0,
    is_default: 0,
  });

  const form = reactive<CreateI18nLangParams>(getDefaultForm());

  const rules = computed<Record<string, FieldRule | FieldRule[]>>(() => ({
    lang: [{ required: true, message: t('i18nLangModal.rule.lang') }],
    name: [{ required: true, message: t('i18nLangModal.rule.name') }],
  }));

  const resetForm = () => {
    Object.assign(form, getDefaultForm());
    formRef.value?.clearValidate();
  };

  const fillForm = (record: I18nLangRecord) => {
    Object.assign(form, {
      lang: record.lang || '',
      name: record.name || '',
      i18n_key: record.i18n_key || '',
      disabled: record.disabled ?? 0,
      is_default: record.is_default ?? 0,
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
    if (errors || submitting.value) {
      done(false);
      return;
    }

    submitting.value = true;
    try {
      const lang = form.lang.trim();
      const name = form.name.trim();
      if (!lang || !name) {
        done(false);
        return;
      }

      const i18nKey = form.i18n_key?.trim() || '';

      if (isEdit.value && props.record?.id) {
        await updateI18nLang({
          id: props.record.id,
          lang,
          name,
          i18n_key: i18nKey,
          disabled: form.disabled ?? 0,
          is_default: form.is_default ?? 0,
        });
        Message.success(t('front.common.message.saveSuccess'));
      } else {
        await createI18nLang({
          lang,
          name,
          i18n_key: i18nKey,
          disabled: form.disabled ?? 0,
          is_default: form.is_default ?? 0,
        });
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
      if (isEdit.value && props.record) {
        fillForm(props.record);
      }
    }
  );
</script>
