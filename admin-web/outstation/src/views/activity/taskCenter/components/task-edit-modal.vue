<template>
  <a-modal
    :visible="visible"
    :title="t('discountCenter.taskCenter.edit.title')"
    :width="560"
    :mask-closable="false"
    unmount-on-close
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    @before-ok="handleBeforeOk"
    @cancel="close"
  >
    <a-form
      ref="formRef"
      :model="form"
      :rules="rules"
      layout="horizontal"
      label-align="left"
      :label-col-props="{ flex: '112px' }"
      :wrapper-col-props="{ flex: '1' }"
    >
      <template v-if="record">
        <a-form-item :label="t('discountCenter.taskCenter.column.taskId')">
          <a-input :model-value="String(record.id)" disabled />
        </a-form-item>
        <a-form-item :label="t('discountCenter.taskCenter.column.taskCode')">
          <a-input :model-value="record.task_code" disabled />
        </a-form-item>
        <a-form-item :label="t('discountCenter.taskCenter.column.name')">
          <a-input :model-value="record.name" disabled />
        </a-form-item>
        <a-form-item :label="t('discountCenter.taskCenter.column.description')">
          <a-textarea
            :model-value="record.description"
            disabled
            :auto-size="{ minRows: 1, maxRows: 4 }"
          />
        </a-form-item>
      </template>
      <a-form-item
        :label="t('discountCenter.taskCenter.column.reward')"
        field="reward_amount"
        :extra="t('discountCenter.taskCenter.edit.minorUnitTip')"
        required
      >
        <a-input-number
          v-model="form.reward_amount"
          :min="0"
          :precision="0"
          hide-button
          :placeholder="t('discountCenter.taskCenter.edit.rewardPlaceholder')"
        />
      </a-form-item>
      <a-form-item
        :label="t('discountCenter.taskCenter.column.completionAmount')"
        field="completion_amount"
        required
      >
        <a-input-number
          v-model="form.completion_amount"
          :min="0"
          :precision="0"
          hide-button
          :placeholder="t('discountCenter.taskCenter.edit.inputPlaceholder')"
        >
          <template #suffix>USD</template>
        </a-input-number>
      </a-form-item>
      <a-form-item
        :label="t('discountCenter.taskCenter.column.payoutMode')"
        field="payout_mode"
        required
      >
        <a-radio-group v-model="form.payout_mode">
          <a-radio v-for="mode in TASK_PAYOUT_MODES" :key="mode" :value="mode">
            {{ t(`discountCenter.taskCenter.payoutMode.${mode}`) }}
          </a-radio>
        </a-radio-group>
      </a-form-item>
      <a-form-item :label="t('discountCenter.taskCenter.column.timezone')">
        <a-input :model-value="form.timezone" disabled />
      </a-form-item>
      <a-form-item
        :label="t('discountCenter.taskCenter.column.sortValue')"
        field="sort"
        :extra="t('discountCenter.taskCenter.edit.sortTip')"
        required
      >
        <a-input-number
          v-model="form.sort"
          :precision="0"
          :placeholder="t('discountCenter.taskCenter.edit.inputPlaceholder')"
        />
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, nextTick, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FormInstance } from '@arco-design/web-vue/es/form';
  import type { FieldRule } from '@arco-design/web-vue/es/form/interface';
  import {
    TASK_PAYOUT_MODES,
    updateTaskConfig,
    type TaskConfigItem,
    type TaskPayoutMode,
  } from '@/api/task';
  import type { TaskEditValue } from '../types';

  const props = defineProps<{
    visible: boolean;
    record: TaskConfigItem | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'saved', value: TaskConfigItem): void;
  }>();

  const { t } = useI18n();
  const formRef = ref<FormInstance>();
  const form = reactive<{
    reward_amount: number | undefined;
    completion_amount: number | undefined;
    payout_mode: TaskPayoutMode;
    timezone: string;
    sort: number | undefined;
  }>({
    reward_amount: undefined,
    completion_amount: undefined,
    payout_mode: 'MANUAL',
    timezone: '',
    sort: undefined,
  });

  function requiredRule(message: string): FieldRule[] {
    return [{ required: true, message }];
  }

  const rules = computed<Record<string, FieldRule[]>>(() => ({
    reward_amount: requiredRule(
      t('discountCenter.taskCenter.edit.rewardRequired')
    ),
    completion_amount: requiredRule(
      t('discountCenter.taskCenter.edit.inputRequired')
    ),
    payout_mode: requiredRule(
      t('discountCenter.taskCenter.edit.inputRequired')
    ),
    sort: requiredRule(t('discountCenter.taskCenter.edit.inputRequired')),
  }));

  watch(
    () => props.visible,
    async (open) => {
      if (!open) return;
      const { record } = props;
      Object.assign(form, {
        reward_amount: record?.reward_amount,
        completion_amount: record?.completion_amount,
        payout_mode: record?.payout_mode || 'MANUAL',
        timezone: record?.timezone || '',
        sort: record?.sort,
      });
      await nextTick();
      formRef.value?.clearValidate();
    }
  );

  function close() {
    emit('update:visible', false);
  }

  async function handleBeforeOk() {
    const { record } = props;
    if (!record) return false;
    const errors = await formRef.value?.validate();
    if (errors) return false;
    const value: TaskEditValue = {
      reward_amount: form.reward_amount ?? 0,
      completion_amount: form.completion_amount ?? 0,
      payout_mode: form.payout_mode,
      timezone: form.timezone.trim(),
      sort: form.sort ?? 0,
    };
    try {
      const { data } = await updateTaskConfig({ id: record.id, ...value });
      emit('saved', data?.task_config ?? { ...record, ...value });
      Message.success(t('discountCenter.taskCenter.message.saved'));
      close();
      return true;
    } catch {
      return false;
    }
  }
</script>

<style scoped lang="less">
  :deep(.arco-input-number) {
    width: 100%;
  }
</style>
