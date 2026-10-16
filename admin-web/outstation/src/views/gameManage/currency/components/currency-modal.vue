<template>
  <a-modal
    :visible="visible"
    :title="t('gameCurrencyModal.title.edit')"
    :ok-loading="submitting"
    :ok-text="t('front.common.okText.save')"
    :mask-closable="false"
    :esc-to-close="false"
    unmount-on-close
    :width="480"
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
      label-align="right"
      :label-col-props="{ span: 6 }"
      :wrapper-col-props="{ span: 17 }"
    >
      <a-form-item
        :label="t('gameCurrencyModal.field.currencyCode')"
        field="currencyCode"
      >
        <a-input v-model="form.currencyCode" disabled />
      </a-form-item>
      <a-form-item :label="t('gameCurrencyModal.field.status')" field="status">
        <a-radio-group v-model="form.status">
          <a-radio :value="1">{{ t('front.common.status.enabled') }}</a-radio>
          <a-radio :value="2">{{ t('front.common.status.disabled') }}</a-radio>
        </a-radio-group>
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FormInstance } from '@arco-design/web-vue';
  import {
    updateGameCurrency,
    type GameCurrencyRecord,
    type UpdateGameCurrencyParams,
  } from '@/api/gameCurrency';

  const { t } = useI18n();

  const props = defineProps<{
    visible: boolean;
    currency?: GameCurrencyRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  interface CurrencyForm {
    currencyCode: string;
    status: number;
  }

  const formRef = ref<FormInstance>();
  const submitting = ref(false);
  const isEdit = computed(() => Boolean(props.currency?.id));

  const getDefaultForm = (): CurrencyForm => ({
    currencyCode: '',
    status: 1,
  });

  const form = reactive<CurrencyForm>(getDefaultForm());

  const resetForm = () => {
    Object.assign(form, getDefaultForm());
    formRef.value?.clearValidate();
  };

  const fillForm = (currency: GameCurrencyRecord) => {
    Object.assign(form, getDefaultForm(), {
      currencyCode: currency.currency_code || '--',
      status: currency.status ?? 1,
    });
  };

  const close = () => {
    emit('update:visible', false);
  };

  const handleCancel = () => {
    close();
    resetForm();
  };

  const buildPayload = (): UpdateGameCurrencyParams => ({
    id: props.currency!.id,
    status: form.status ?? 1,
  });

  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    const errors = await formRef.value?.validate();
    if (errors || submitting.value || !isEdit.value || !props.currency?.id) {
      done(false);
      return;
    }

    submitting.value = true;
    try {
      await updateGameCurrency(buildPayload());
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
      if (props.currency?.id) {
        fillForm(props.currency);
      }
    }
  );
</script>
