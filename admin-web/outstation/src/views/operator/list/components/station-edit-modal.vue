<template>
  <a-modal
    :visible="visible"
    :title="t('operatorStation.edit.title')"
    :ok-loading="submitting"
    :ok-text="t('front.common.okText.save')"
    :mask-closable="false"
    :esc-to-close="false"
    unmount-on-close
    :width="600"
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    @before-ok="handleBeforeOk"
    @cancel="handleCancel"
  >
    <a-spin :loading="optionsLoading" style="width: 100%">
      <a-alert v-if="baseLocked" type="warning" class="lock-tip">
        {{ t('operatorStation.edit.lockTip') }}
      </a-alert>
      <a-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-align="right"
        :label-col-props="{ span: 7 }"
        :wrapper-col-props="{ span: 16 }"
      >
        <a-form-item :label="t('operatorStation.column.code')">
          <span class="readonly-text">{{ record?.code || '--' }}</span>
        </a-form-item>
        <a-form-item
          :label="t('operatorStation.create.baseInfo.name')"
          field="name"
          required
        >
          <a-input
            v-model="form.name"
            :placeholder="t('operatorStation.create.baseInfo.namePlaceholder')"
            allow-clear
            :max-length="100"
            show-word-limit
          />
        </a-form-item>
        <a-form-item
          :label="t('operatorStation.create.baseInfo.timezone')"
          field="timezone_code"
          required
        >
          <a-select
            v-model="form.timezone_code"
            :placeholder="
              t('operatorStation.create.baseInfo.timezonePlaceholder')
            "
            :disabled="baseLocked"
            allow-search
            :options="timezoneOptions"
          />
        </a-form-item>
        <a-form-item
          :label="t('operatorStation.create.baseInfo.currency')"
          field="settlement_currency_code"
          required
        >
          <a-select
            v-model="form.settlement_currency_code"
            :placeholder="
              t('operatorStation.create.baseInfo.currencyPlaceholder')
            "
            :disabled="baseLocked"
            allow-search
            :options="currencyOptions"
          />
        </a-form-item>
        <a-form-item
          :label="t('operatorStation.create.baseInfo.status')"
          field="status"
          required
        >
          <a-select v-model="form.status" :options="statusOptions" />
        </a-form-item>
        <a-form-item
          :label="t('operatorStation.create.baseInfo.remark')"
          field="remark"
        >
          <a-textarea
            v-model="form.remark"
            :placeholder="
              t('operatorStation.create.baseInfo.remarkPlaceholder')
            "
            :max-length="1000"
            show-word-limit
            allow-clear
            :auto-size="{ minRows: 3, maxRows: 5 }"
          />
        </a-form-item>
      </a-form>
    </a-spin>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FieldRule, FormInstance } from '@arco-design/web-vue';
  import { queryTimezoneListAll, TIMEZONE_STATUS } from '@/api/timezone';
  import { queryCurrencyListAll, CURRENCY_STATUS } from '@/api/currency';
  import {
    isOperatorBaseCurrencyLocked,
    OPERATOR_STATUS,
    updateOperator,
    type OperatorRecord,
    type UpdateOperatorParams,
  } from '@/api/operator';

  const { t } = useI18n();

  const props = defineProps<{
    visible: boolean;
    record?: OperatorRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  const formRef = ref<FormInstance>();
  const optionsLoading = ref(false);
  const submitting = ref(false);

  const timezoneOptions = ref<{ label: string; value: string }[]>([]);
  const currencyOptions = ref<{ label: string; value: string }[]>([]);
  const statusOptions = computed(() => [
    {
      label: t('operatorStation.status.normal'),
      value: OPERATOR_STATUS.NORMAL,
    },
    {
      label: t('operatorStation.status.paused'),
      value: OPERATOR_STATUS.PAUSED,
    },
    {
      label: t('operatorStation.status.closed'),
      value: OPERATOR_STATUS.CLOSED,
    },
  ]);

  /** 发布中或已发布过的分站，时区与结算币种不允许再修�? */
  const baseLocked = computed(() =>
    props.record ? isOperatorBaseCurrencyLocked(props.record) : false
  );

  const getDefaultForm = () => ({
    name: '',
    timezone_code: '',
    settlement_currency_code: '',
    status: OPERATOR_STATUS.NORMAL as number,
    remark: '',
  });

  const form = reactive(getDefaultForm());

  const rules = computed<Record<string, FieldRule | FieldRule[]>>(() => ({
    name: [
      {
        required: true,
        message: t('operatorStation.create.rule.nameRequired'),
      },
    ],
    timezone_code: [
      {
        required: true,
        message: t('operatorStation.create.rule.timezoneRequired'),
      },
    ],
    settlement_currency_code: [
      {
        required: true,
        message: t('operatorStation.create.rule.currencyRequired'),
      },
    ],
    status: [
      {
        required: true,
        message: t('operatorStation.create.rule.statusRequired'),
      },
    ],
  }));

  /**
   * 选项接口只返回启用中的时�?币种，分站上存量的编码可能已被停用，
   * 补一条同值选项，避免下拉框把已有值显示成空�?   */
  const ensureOption = (
    options: { label: string; value: string }[],
    code?: string
  ) => {
    if (!code || options.some((item) => item.value === code)) return options;
    return [...options, { label: code, value: code }];
  };

  const fetchOptions = async () => {
    optionsLoading.value = true;
    try {
      const [timezoneRes, currencyRes] = await Promise.all([
        queryTimezoneListAll({ status: TIMEZONE_STATUS.ENABLED }),
        queryCurrencyListAll({ status: CURRENCY_STATUS.ENABLED }),
      ]);
      timezoneOptions.value = ensureOption(
        (timezoneRes.data?.list || []).map((item) => ({
          label: item.name ? `${item.name}（${item.code}）` : item.code,
          value: item.code,
        })),
        props.record?.timezone_code
      );
      currencyOptions.value = ensureOption(
        (currencyRes.data?.list || []).map((item) => ({
          label: item.name ? `${item.name}（${item.code}）` : item.code,
          value: item.code,
        })),
        props.record?.settlement_currency_code
      );
    } catch {
      timezoneOptions.value = ensureOption([], props.record?.timezone_code);
      currencyOptions.value = ensureOption(
        [],
        props.record?.settlement_currency_code
      );
    } finally {
      optionsLoading.value = false;
    }
  };

  const fillForm = (record: OperatorRecord) => {
    Object.assign(form, {
      name: record.name || '',
      timezone_code: record.timezone_code || '',
      settlement_currency_code: record.settlement_currency_code || '',
      status: record.status ?? OPERATOR_STATUS.NORMAL,
      remark: record.remark || '',
    });
  };

  const close = () => {
    emit('update:visible', false);
  };

  const handleCancel = () => {
    close();
    Object.assign(form, getDefaultForm());
    formRef.value?.clearValidate();
  };

  /** 接口要求�?id 外至少传一个字段，这里只提交实际发生变化的字段 */
  const buildUpdateParams = (record: OperatorRecord) => {
    const params: UpdateOperatorParams = { id: record.id };
    const name = form.name.trim();
    const remark = form.remark.trim();
    if (name !== record.name) params.name = name;
    if (form.status !== record.status) params.status = form.status;
    if (remark !== (record.remark || '')) params.remark = remark;
    if (!baseLocked.value) {
      if (form.timezone_code !== record.timezone_code) {
        params.timezone_code = form.timezone_code;
      }
      if (form.settlement_currency_code !== record.settlement_currency_code) {
        params.settlement_currency_code = form.settlement_currency_code;
      }
    }
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
      Message.info(t('operatorStation.edit.message.noChange'));
      done(false);
      return;
    }

    submitting.value = true;
    try {
      await updateOperator(params);
      Message.success(t('front.common.message.saveSuccess'));
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
      Object.assign(form, getDefaultForm());
      formRef.value?.clearValidate();
      if (props.record) fillForm(props.record);
      fetchOptions();
    }
  );
</script>

<style scoped lang="less">
  .lock-tip {
    margin-bottom: 16px;
  }
  .readonly-text {
    color: var(--color-text-2);
  }
</style>
