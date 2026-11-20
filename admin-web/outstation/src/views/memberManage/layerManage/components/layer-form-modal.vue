<template>
  <a-modal
    :visible="visible"
    :title="title"
    :width="560"
    :mask-closable="false"
    unmount-on-close
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    :ok-text="t('layerManage.form.ok')"
    @before-ok="handleBeforeOk"
    @ok="close"
    @cancel="close"
  >
    <a-form
      ref="formRef"
      :model="form"
      :rules="rules"
      layout="horizontal"
      label-align="left"
      :label-col-props="{ flex: '128px' }"
      :wrapper-col-props="{ flex: '1' }"
    >
      <a-form-item :label="t('layerManage.column.name')" field="name" required>
        <a-input
          v-model="form.name"
          :max-length="100"
          allow-clear
          :placeholder="t('layerManage.form.nameMaxPlaceholder')"
        />
      </a-form-item>
      <template v-if="kind === 'fixed'">
        <a-form-item :label="t('layerManage.column.description')">
          <a-textarea
            v-model="form.description"
            :max-length="255"
            show-word-limit
            :auto-size="{ minRows: 3, maxRows: 5 }"
            :placeholder="t('layerManage.form.descriptionPlaceholder')"
          />
        </a-form-item>
        <a-form-item :label="t('layerManage.form.limitEnabled')">
          <div class="limit-field">
            <a-switch v-model="form.limitEnabled" @change="onLimitToggle" />
            <div class="limit-extra">
              {{ t('layerManage.form.limitEnabledTip') }}
            </div>
            <div class="limit-panel">
              <a-form-item
                field="withdrawLimit"
                class="limit-item"
                :label="t('layerManage.column.withdrawLimit')"
                required
                :label-col-props="{ flex: '96px' }"
                :wrapper-col-props="{ flex: '1' }"
                :extra="t('layerManage.form.withdrawLimitTip')"
              >
                <a-input-number
                  v-model="form.withdrawLimit"
                  :min="0"
                  :precision="0"
                  hide-button
                  :disabled="!form.limitEnabled"
                  :placeholder="t('layerManage.form.amountPlaceholder')"
                />
              </a-form-item>
            </div>
          </div>
        </a-form-item>
      </template>
      <template v-else-if="showThreshold">
        <a-form-item
          :label="t('layerManage.form.deposit')"
          field="deposit"
          required
        >
          <a-input-number
            v-model="form.deposit"
            :min="0"
            :precision="0"
            hide-button
            :placeholder="t('layerManage.form.amountPlaceholder')"
          />
        </a-form-item>
        <a-form-item :label="t('layerManage.form.wash')" field="wash" required>
          <a-input-number
            v-model="form.wash"
            :min="0"
            :precision="0"
            hide-button
            :placeholder="t('layerManage.form.amountPlaceholder')"
          />
        </a-form-item>
      </template>
    </a-form>
    <div class="form-note">
      <icon-info-circle-fill class="note-icon" />
      <div>
        <div v-for="line in noteLines" :key="line">{{ line }}</div>
      </div>
    </div>
    <LayerThresholdConfirmModal
      v-if="pendingThreshold && record"
      v-model:visible="thresholdConfirmVisible"
      :record="record"
      :name="pendingThreshold.name"
      :after="pendingThreshold"
      :layers="autoLayers"
      @confirm="confirmThreshold"
    />
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, nextTick, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FormInstance } from '@arco-design/web-vue/es/form';
  import type { FieldRule } from '@arco-design/web-vue/es/form/interface';
  import type { LayerFormValue, LayerKind, LayerRecord } from '../types';
  import LayerThresholdConfirmModal from './layer-threshold-confirm-modal.vue';

  const props = defineProps<{
    visible: boolean;
    mode: 'create' | 'edit';
    kind: LayerKind;
    record: LayerRecord | null;
    names: string[];
    autoLayers: LayerRecord[];
    /** 以 @submit 绑定；返回 false 或抛错时弹窗保持打开 */
    onSubmit: (
      value: LayerFormValue
    ) => boolean | void | Promise<boolean | void>;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
  }>();

  const { t } = useI18n();
  const formRef = ref<FormInstance>();
  const form = reactive<{
    name: string;
    description: string;
    deposit: number | undefined;
    wash: number | undefined;
    limitEnabled: boolean;
    withdrawLimit: number | undefined;
  }>({
    name: '',
    description: '',
    deposit: undefined,
    wash: undefined,
    limitEnabled: false,
    withdrawLimit: undefined,
  });
  const thresholdConfirmVisible = ref(false);
  const pendingThreshold = ref<{
    name: string;
    deposit: number;
    wash: number;
  } | null>(null);

  const title = computed(() => {
    const action = props.mode === 'create' ? 'create' : 'edit';
    const kind = props.kind === 'fixed' ? 'Fixed' : 'Auto';
    return t(`layerManage.form.${action}${kind}`);
  });

  const isDefaultLevel = computed(
    () => props.mode === 'edit' && Boolean(props.record?.isDefault)
  );
  const showThreshold = computed(
    () => props.kind === 'auto' && !isDefaultLevel.value
  );

  const noteLines = computed(() => {
    if (props.kind === 'fixed') {
      return [
        t('layerManage.form.fixedNoteAssign'),
        t('layerManage.form.fixedNoteLimit'),
      ];
    }
    if (isDefaultLevel.value) return [t('layerManage.form.defaultTip')];
    return [
      t('layerManage.form.autoNotePromote'),
      t('layerManage.form.autoNoteThreshold'),
    ];
  });

  const rules = computed<Record<string, FieldRule[]>>(() => {
    const nameRules: FieldRule[] = [
      { required: true, message: t('layerManage.form.nameRequired') },
    ];
    if (props.kind === 'fixed') {
      if (!form.limitEnabled) return { name: nameRules };
      return {
        name: nameRules,
        withdrawLimit: [
          {
            required: true,
            message: t('layerManage.form.withdrawLimitRequired'),
          },
          {
            validator: (value: number | undefined, callback) => {
              if (value != null && value <= 0) {
                callback(t('layerManage.form.withdrawLimitPositive'));
                return;
              }
              callback();
            },
          },
        ],
      };
    }
    if (!showThreshold.value) return { name: nameRules };
    return {
      name: nameRules,
      deposit: [
        { required: true, message: t('layerManage.form.depositRequired') },
      ],
      wash: [{ required: true, message: t('layerManage.form.washRequired') }],
    };
  });

  function fill(record: LayerRecord) {
    form.name = record.name;
    form.description = record.description;
    form.deposit = record.deposit ?? undefined;
    form.wash = record.wash ?? undefined;
    form.limitEnabled = record.withdrawLimit != null;
    form.withdrawLimit = record.withdrawLimit ?? undefined;
  }

  function reset() {
    form.name = '';
    form.description = '';
    form.deposit = undefined;
    form.wash = undefined;
    form.limitEnabled = false;
    form.withdrawLimit = undefined;
  }

  watch(
    () => [props.visible, props.mode, props.record?.id] as const,
    async ([open]) => {
      if (!open) return;
      thresholdConfirmVisible.value = false;
      pendingThreshold.value = null;
      if (props.mode === 'edit' && props.record) fill(props.record);
      else reset();
      await nextTick();
      formRef.value?.clearValidate();
    }
  );

  function onLimitToggle(enabled: string | number | boolean) {
    if (!enabled) formRef.value?.clearValidate('withdrawLimit');
  }

  function close() {
    emit('update:visible', false);
  }

  async function submit(value: LayerFormValue) {
    try {
      return (await props.onSubmit(value)) !== false;
    } catch {
      return false;
    }
  }

  async function confirmThreshold() {
    if (!pendingThreshold.value) return;
    if (await submit({ ...pendingThreshold.value })) close();
  }

  async function handleBeforeOk() {
    const errors = await formRef.value?.validate();
    if (errors) return false;
    const name = form.name.trim();
    if (!name) {
      Message.warning(t('layerManage.form.nameRequired'));
      return false;
    }
    if (props.names.includes(name)) {
      Message.warning(t('layerManage.form.nameDuplicate'));
      return false;
    }
    if (showThreshold.value) {
      const deposit = form.deposit ?? 0;
      const wash = form.wash ?? 0;
      if (deposit <= 0 && wash <= 0) {
        Message.warning(t('layerManage.form.thresholdRequired'));
        return false;
      }
      const { record } = props;
      if (
        props.mode === 'edit' &&
        record &&
        (deposit !== (record.deposit ?? 0) || wash !== (record.wash ?? 0))
      ) {
        pendingThreshold.value = { name, deposit, wash };
        thresholdConfirmVisible.value = true;
        return false;
      }
      return submit({ name, deposit, wash });
    }
    if (props.kind === 'fixed') {
      return submit({
        name,
        description: form.description.trim(),
        withdrawLimit: form.limitEnabled ? form.withdrawLimit ?? null : null,
      });
    }
    return submit({ name });
  }
</script>

<style scoped lang="less">
  :deep(.arco-input-number) {
    width: 100%;
  }

  .limit-field {
    width: 100%;
  }

  .limit-extra {
    margin-top: 8px;
    color: var(--color-text-3);
    font-size: 12px;
    line-height: 18px;
  }

  .limit-panel {
    margin-top: 12px;
    padding: 16px 16px 4px;
    background: var(--color-fill-1);
    border-radius: 6px;
  }

  .limit-item {
    margin-bottom: 12px;
    :deep(.arco-form-item-label-col > .arco-form-item-label) {
      color: var(--color-text-2);
      font-size: 13px;
      white-space: nowrap;
    }

    :deep(.arco-form-item-extra) {
      margin-top: 6px;
      color: var(--color-text-3);
      font-size: 12px;
    }
  }

  .form-note {
    display: flex;
    gap: 8px;
    align-items: flex-start;
    padding: 12px 16px;
    color: var(--color-text-2);
    font-size: 12px;
    line-height: 20px;
    background: rgb(var(--primary-1));
    border-radius: 6px;
  }

  .note-icon {
    flex-shrink: 0;
    margin-top: 3px;
    color: rgb(var(--primary-6));
    font-size: 14px;
  }
</style>
