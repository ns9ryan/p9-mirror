<template>
  <a-modal
    :visible="visible"
    :title="$t('currencyModal.title.edit')"
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
      <a-form-item :label="$t('currencyModal.field.code')">
        <a-input :model-value="form.code" disabled />
        <template #extra>{{ $t('currencyModal.field.codeExtra') }}</template>
      </a-form-item>
      <a-form-item :label="$t('currencyModal.field.nameKey')">
        <a-input :model-value="form.nameKey" disabled />
        <template #extra>{{ $t('currencyModal.field.nameKeyExtra') }}</template>
      </a-form-item>
      <a-form-item :label="$t('currencyModal.field.name')">
        <a-input :model-value="form.name" disabled />
      </a-form-item>
      <a-form-item :label="$t('currencyModal.field.currencyType')">
        <a-input :model-value="currencyTypeLabel" disabled />
        <template #extra>{{
          $t('currencyModal.field.currencyTypeExtra')
        }}</template>
      </a-form-item>
      <a-form-item :label="$t('currencyModal.field.amountFactor')">
        <a-input :model-value="String(form.amountFactor)" disabled />
        <template #extra>{{
          $t('currencyModal.field.amountFactorExtra')
        }}</template>
      </a-form-item>
      <a-form-item
        :label="$t('currencyModal.field.symbol')"
        field="symbol"
        required
      >
        <a-input
          v-model="form.symbol"
          :placeholder="$t('currencyModal.field.symbolPlaceholder')"
          allow-clear
          :max-length="16"
        />
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
  import { updateCurrency, type CurrencyRecord } from '@/api/currency';

  const props = defineProps<{
    visible: boolean;
    record?: CurrencyRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  const { t } = useI18n();
  const formRef = ref<FormInstance>();
  const submitting = ref(false);

  interface CurrencyForm {
    code: string;
    nameKey: string;
    name: string;
    currencyType: number;
    amountFactor: number;
    symbol: string;
    status: number;
  }

  const getDefaultForm = (): CurrencyForm => ({
    code: '',
    nameKey: '',
    name: '',
    currencyType: 1,
    amountFactor: 1,
    symbol: '',
    status: 1,
  });

  const form = reactive<CurrencyForm>(getDefaultForm());

  const currencyTypeLabel = computed(() => {
    if (form.currencyType === 1) return t('currencyManage.type.fiat');
    if (form.currencyType === 2) return t('currencyManage.type.virtual');
    return '--';
  });

  const rules = computed<Record<string, FieldRule | FieldRule[]>>(() => ({
    symbol: [
      { required: true, message: t('currencyModal.rule.symbol') },
      {
        validator: (value: string, callback: (error?: string) => void) => {
          if (!String(value || '').trim()) {
            callback(t('currencyModal.rule.symbolRequired'));
            return;
          }
          callback();
        },
      },
    ],
    status: [{ required: true, message: t('currencyModal.rule.status') }],
  }));

  const resetForm = () => {
    Object.assign(form, getDefaultForm());
    formRef.value?.clearValidate();
  };

  const fillForm = (record: CurrencyRecord) => {
    Object.assign(form, {
      code: record.code || '',
      nameKey: record.name_key || '',
      name: record.name || '',
      currencyType: record.currency_type ?? 1,
      amountFactor: record.amount_factor ?? 1,
      symbol: record.symbol || '',
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

    const symbol = form.symbol.trim();
    if (!symbol) {
      Message.warning(t('currencyManage.message.symbolRequired'));
      done(false);
      return;
    }

    submitting.value = true;
    try {
      await updateCurrency({
        id: props.record.id,
        symbol,
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
