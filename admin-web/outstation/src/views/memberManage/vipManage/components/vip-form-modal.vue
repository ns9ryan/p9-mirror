<template>
  <a-modal
    :visible="visible"
    :title="
      mode === 'create' ? t('vipManage.form.create') : t('vipManage.form.edit')
    "
    :width="600"
    :mask-closable="false"
    unmount-on-close
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    :ok-text="mode === 'create' ? undefined : t('front.common.okText.save')"
    :ok-button-props="{ disabled: isDefault }"
    @before-ok="handleBeforeOk"
    @cancel="close"
  >
    <a-alert v-if="mode === 'create'" class="form-tip">
      <ul class="tip-list">
        <li>{{ t('vipManage.form.tip.sequence') }}</li>
        <li>{{ t('vipManage.form.tip.default') }}</li>
      </ul>
    </a-alert>
    <a-alert v-else-if="isDefault" class="form-tip">
      {{ t('vipManage.form.zeroTip') }}
    </a-alert>
    <a-alert v-else-if="thresholdLocked" type="warning" class="form-tip">
      {{ t('vipManage.form.memberLockTip') }}
    </a-alert>
    <a-form
      ref="formRef"
      :model="form"
      :rules="rules"
      :label-col-props="{ flex: '104px' }"
      :wrapper-col-props="{ flex: '1' }"
    >
      <a-form-item
        class="level-item"
        :label="t('vipManage.form.level')"
        :extra="
          mode === 'create'
            ? t('vipManage.form.levelAuto')
            : t('vipManage.form.levelFixed')
        "
      >
        <a-input :model-value="levelText" disabled />
      </a-form-item>

      <section class="form-section">
        <div class="section-title">{{ t('vipManage.form.promote') }}</div>
        <div class="section-body">
          <a-form-item
            :label="t('vipManage.form.deposit')"
            field="promoteDeposit"
            :extra="t('vipManage.form.help.promoteDeposit')"
          >
            <a-input-number
              v-model="form.promoteDeposit"
              :min="0"
              :precision="0"
              :disabled="thresholdLocked"
              hide-button
              :placeholder="t('vipManage.form.amountPlaceholder')"
            />
          </a-form-item>
          <a-form-item
            :label="t('vipManage.form.wash')"
            field="promoteWash"
            :extra="t('vipManage.form.help.promoteWash')"
          >
            <a-input-number
              v-model="form.promoteWash"
              :min="0"
              :precision="0"
              :disabled="thresholdLocked"
              hide-button
              :placeholder="t('vipManage.form.amountPlaceholder')"
            />
          </a-form-item>
        </div>
      </section>

      <section class="form-section">
        <div class="section-title">{{ t('vipManage.form.keep') }}</div>
        <div class="section-body">
          <a-form-item
            :label="t('vipManage.form.deposit')"
            field="keepDeposit"
            :extra="t('vipManage.form.help.keepDeposit')"
          >
            <a-input-number
              v-model="form.keepDeposit"
              :min="0"
              :precision="0"
              :disabled="thresholdLocked"
              hide-button
              :placeholder="t('vipManage.form.amountPlaceholder')"
            />
          </a-form-item>
          <a-form-item
            :label="t('vipManage.form.wash')"
            field="keepWash"
            :extra="t('vipManage.form.help.keepWash')"
          >
            <a-input-number
              v-model="form.keepWash"
              :min="0"
              :precision="0"
              :disabled="thresholdLocked"
              hide-button
              :placeholder="t('vipManage.form.amountPlaceholder')"
            />
          </a-form-item>
        </div>
      </section>

      <section class="form-section">
        <div class="section-title">{{ t('vipManage.form.reward') }}</div>
        <div class="section-body">
          <a-form-item
            :label="t('vipManage.form.rewardAmount')"
            field="reward"
            :extra="t('vipManage.form.help.reward')"
          >
            <a-input-number
              v-model="form.reward"
              :min="0"
              :precision="0"
              :disabled="isDefault"
              hide-button
              :placeholder="t('vipManage.form.amountPlaceholder')"
            />
          </a-form-item>
        </div>
      </section>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, nextTick, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { FormInstance } from '@arco-design/web-vue/es/form';
  import type { FieldRule } from '@arco-design/web-vue/es/form/interface';
  import type { VipLevelFormValue, VipLevelRecord } from '../types';

  const props = defineProps<{
    visible: boolean;
    mode: 'create' | 'edit';
    record: VipLevelRecord | null;
    levels: number[];
    /** 新增时预览的等级编号；不传则按 levels 最大值 + 1 */
    nextLevel?: number;
    /** 以 @submit 绑定；返回 false 或抛错时弹窗保持打开 */
    onSubmit: (
      value: VipLevelFormValue
    ) => boolean | void | Promise<boolean | void>;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
  }>();

  const { t } = useI18n();
  const formRef = ref<FormInstance>();
  const level = ref(0);
  const form = reactive<{
    reward: number | undefined;
    promoteDeposit: number | undefined;
    promoteWash: number | undefined;
    keepDeposit: number | undefined;
    keepWash: number | undefined;
  }>({
    reward: undefined,
    promoteDeposit: undefined,
    promoteWash: undefined,
    keepDeposit: undefined,
    keepWash: undefined,
  });

  const isDefault = computed(
    () => props.mode === 'edit' && props.record?.level === 0
  );
  /** 已有会员的等级不可修改晋级/保级门槛，只能改奖励 */
  const thresholdLocked = computed(
    () =>
      isDefault.value ||
      (props.mode === 'edit' && (props.record?.memberCount ?? 0) > 0)
  );
  const levelText = computed(() => `VIP${level.value}`);

  function requiredRule(message: string): FieldRule[] {
    return [
      {
        required: true,
        validator: (value, callback) => {
          if (value == null || value === '') {
            callback(message);
            return;
          }
          callback();
        },
      },
    ];
  }

  const rules = computed<Record<string, FieldRule[]>>(() => ({
    promoteDeposit: requiredRule(t('vipManage.form.depositRequired')),
    promoteWash: requiredRule(t('vipManage.form.washRequired')),
    keepDeposit: requiredRule(t('vipManage.form.depositRequired')),
    keepWash: requiredRule(t('vipManage.form.washRequired')),
    reward: requiredRule(t('vipManage.form.rewardRequired')),
  }));

  function fill(record: VipLevelRecord) {
    level.value = record.level;
    form.reward = record.reward;
    form.promoteDeposit = record.promoteDeposit;
    form.promoteWash = record.promoteWash;
    form.keepDeposit = record.keepDeposit;
    form.keepWash = record.keepWash;
  }

  function reset() {
    level.value =
      props.nextLevel ??
      props.levels.reduce((max, item) => (item > max ? item : max), 0) + 1;
    form.reward = undefined;
    form.promoteDeposit = undefined;
    form.promoteWash = undefined;
    form.keepDeposit = undefined;
    form.keepWash = undefined;
  }

  watch(
    () => [props.visible, props.mode, props.record?.level] as const,
    async ([open]) => {
      if (!open) return;
      if (props.mode === 'edit' && props.record) fill(props.record);
      else reset();
      await nextTick();
      formRef.value?.clearValidate();
    }
  );

  function close() {
    emit('update:visible', false);
  }

  async function handleBeforeOk() {
    const errors = await formRef.value?.validate();
    if (errors) return false;
    try {
      const ok =
        (await props.onSubmit({
          level: level.value,
          reward: form.reward ?? 0,
          promoteDeposit: form.promoteDeposit ?? 0,
          promoteWash: form.promoteWash ?? 0,
          keepDeposit: form.keepDeposit ?? 0,
          keepWash: form.keepWash ?? 0,
        })) !== false;
      if (ok) close();
      return ok;
    } catch {
      return false;
    }
  }
</script>

<style scoped lang="less">
  .form-tip {
    margin-bottom: 16px;
    border-radius: 6px;
  }

  .tip-list {
    margin: 0;
    padding-left: 16px;
    color: var(--color-text-2);
    line-height: 22px;
  }

  .level-item {
    margin-bottom: 16px;
  }

  .form-section {
    margin-bottom: 12px;
    overflow: hidden;
    border: 1px solid var(--color-neutral-3);
    border-radius: 6px;

    &:last-child {
      margin-bottom: 0;
    }
  }

  .section-title {
    padding: 8px 16px;
    color: var(--color-text-1);
    font-weight: 600;
    background: var(--color-fill-2);
  }

  .section-body {
    padding: 12px 16px 0;

    :deep(.arco-form-item) {
      margin-bottom: 12px;
    }
  }

  :deep(.arco-form-item-extra) {
    color: var(--color-text-3);
    font-size: 12px;
  }

  :deep(.arco-input-number) {
    width: 100%;
  }
</style>
