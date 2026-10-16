<template>
  <a-modal
    :visible="visible"
    :title="t('manufacturerModal.title.edit')"
    :ok-loading="submitting"
    :ok-text="t('front.common.okText.save')"
    :mask-closable="false"
    :esc-to-close="false"
    unmount-on-close
    :width="680"
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
        :label="t('manufacturerModal.field.providerCode')"
        field="providerCode"
      >
        <a-input v-model="form.providerCode" disabled />
      </a-form-item>
      <a-form-item
        :label="t('manufacturerModal.field.name')"
        field="name"
        required
      >
        <a-textarea
          v-model="form.name"
          :placeholder="t('manufacturerModal.field.namePlaceholder')"
          allow-clear
          :auto-size="{ minRows: 4, maxRows: 8 }"
        />
      </a-form-item>
      <a-form-item :label="t('manufacturerModal.field.logoUrl')" field="logo_url">
        <a-input
          v-model="form.logo_url"
          :placeholder="t('manufacturerModal.field.logoUrlPlaceholder')"
          allow-clear
        />
      </a-form-item>
      <a-form-item
        v-if="form.logo_url"
        :label="t('manufacturerModal.field.logoPreview')"
      >
        <a-image :src="form.logo_url" width="60" height="60" fit="contain" />
      </a-form-item>
      <a-form-item :label="t('manufacturerModal.field.sortNo')" field="sort_no">
        <a-input-number
          v-model="form.sort_no"
          :min="1"
          :max="9999"
          :placeholder="t('manufacturerModal.field.sortNoPlaceholder')"
          style="width: 100%"
        />
      </a-form-item>
      <a-form-item :label="t('manufacturerModal.field.status')" field="status">
        <a-radio-group v-model="form.status">
          <a-radio :value="1">{{ t('front.common.status.enabled') }}</a-radio>
          <a-radio :value="2"
            >{{ t('front.common.status.disabled')
            }}<span style="color: red">{{
              t('manufacturerModal.field.forceLogoutHint')
            }}</span></a-radio
          >
        </a-radio-group>
      </a-form-item>
      <a-form-item :label="t('manufacturerModal.field.note')">
        <a-alert type="info" :show-icon="false">
          {{ t('manufacturerModal.field.noteContent') }}
        </a-alert>
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FieldRule, FormInstance } from '@arco-design/web-vue';
  import {
    updateGameProvider,
    type GameProviderRecord,
    type UpdateGameProviderParams,
  } from '@/api/gameProvider';

  const { t } = useI18n();

  const props = defineProps<{
    visible: boolean;
    provider?: GameProviderRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  interface ProviderForm {
    providerCode: string;
    name: string;
    logo_url: string;
    sort_no: number;
    status: number;
  }

  const formRef = ref<FormInstance>();
  const submitting = ref(false);
  const isEdit = computed(() => Boolean(props.provider?.id));

  const getDefaultForm = (): ProviderForm => ({
    providerCode: '',
    name: '',
    logo_url: '',
    sort_no: 1,
    status: 1,
  });

  const form = reactive<ProviderForm>(getDefaultForm());

  const rules = computed<Record<string, FieldRule | FieldRule[]>>(() => ({
    name: [
      { required: true, message: t('manufacturerModal.rule.nameRequired') },
      {
        validator: (value: string, callback: (error?: string) => void) => {
          try {
            const parsed = JSON.parse(value || '');
            if (!parsed || typeof parsed !== 'object') {
              callback(t('manufacturerModal.rule.nameObject'));
              return;
            }
            callback();
          } catch {
            callback(t('manufacturerModal.rule.nameJson'));
          }
        },
      },
    ],
  }));

  const prettyJson = (value: string) => {
    try {
      return JSON.stringify(JSON.parse(value), null, 2);
    } catch {
      return value;
    }
  };

  const resetForm = () => {
    Object.assign(form, getDefaultForm());
    formRef.value?.clearValidate();
  };

  const fillForm = (provider: GameProviderRecord) => {
    Object.assign(form, getDefaultForm(), {
      providerCode: provider.provider_code || '--',
      name: prettyJson(provider.name || ''),
      logo_url: provider.logo_url || '',
      sort_no: provider.sort_no && provider.sort_no > 0 ? provider.sort_no : 1,
      status: provider.status ?? 1,
    });
  };

  const close = () => {
    emit('update:visible', false);
  };

  const handleCancel = () => {
    close();
    resetForm();
  };

  const buildPayload = (): UpdateGameProviderParams => ({
    id: props.provider!.id,
    name: form.name.trim(),
    logo_url: form.logo_url.trim(),
    sort_no: form.sort_no && form.sort_no > 0 ? form.sort_no : 1,
    status: form.status ?? 1,
    force_logout: form.status === 2,
  });

  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    const errors = await formRef.value?.validate();
    if (errors || submitting.value || !isEdit.value || !props.provider?.id) {
      done(false);
      return;
    }

    submitting.value = true;
    try {
      await updateGameProvider(buildPayload());
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
      if (props.provider?.id) {
        fillForm(props.provider);
      }
    }
  );
</script>
