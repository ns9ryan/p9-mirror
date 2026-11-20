<template>
  <a-modal
    :visible="visible"
    :title="t('discountCenter.taskCenter.config.title')"
    :width="760"
    :mask-closable="false"
    unmount-on-close
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    @before-ok="handleBeforeOk"
    @cancel="close"
  >
    <a-spin :loading="loading" class="config-spin">
      <a-form
        ref="formRef"
        :model="form"
        :rules="rules"
        layout="horizontal"
        label-align="right"
        :label-col-props="{ flex: '112px' }"
        :wrapper-col-props="{ flex: '1' }"
      >
        <a-form-item :label="t('discountCenter.taskCenter.config.enabled')">
          <a-switch
            v-model="form.enabled"
            :checked-text="t('discountCenter.taskCenter.switch.on')"
            :unchecked-text="t('discountCenter.taskCenter.switch.off')"
          />
        </a-form-item>

        <a-form-item
          :label="t('discountCenter.taskCenter.config.dispatchMode')"
          field="dispatchMode"
          required
        >
          <a-radio-group v-model="form.dispatchMode">
            <a-radio value="manual">
              {{ t('discountCenter.taskCenter.config.dispatch.manual') }}
            </a-radio>
            <a-radio value="auto">
              {{ t('discountCenter.taskCenter.config.dispatch.auto') }}
            </a-radio>
          </a-radio-group>
        </a-form-item>

        <a-form-item
          v-for="group in levelGroups"
          :key="group.field"
          :label="t(group.label)"
        >
          <div class="check-panel">
            <template v-if="group.options.length">
              <a-checkbox
                class="check-all"
                :model-value="isAll(form[group.field], group.options)"
                :indeterminate="isPartial(form[group.field], group.options)"
                @change="
                  (value) => toggleAll(group.field, group.options, value)
                "
              >
                {{ t('discountCenter.taskCenter.config.all') }}
              </a-checkbox>
              <a-checkbox-group v-model="form[group.field]" class="check-grid">
                <a-checkbox
                  v-for="item in group.options"
                  :key="item.value"
                  :value="item.value"
                >
                  {{ item.label }}
                </a-checkbox>
              </a-checkbox-group>
            </template>
            <span v-else class="check-empty">
              {{ t('discountCenter.taskCenter.config.noLevels') }}
            </span>
          </div>
        </a-form-item>

        <a-form-item :label="t('discountCenter.taskCenter.config.washLimit')">
          <a-input-number
            v-model="form.washMultiplier"
            :min="0"
            :precision="0"
            hide-button
            :placeholder="t('discountCenter.taskCenter.config.washPlaceholder')"
          >
            <template #suffix>
              {{ t('discountCenter.taskCenter.config.washSuffix') }}
            </template>
          </a-input-number>
        </a-form-item>

        <a-form-item
          :label="t('discountCenter.taskCenter.config.clients')"
          field="clients"
          required
        >
          <div class="check-panel">
            <a-checkbox
              class="check-all"
              :model-value="isAll(form.clients, clientOptions)"
              :indeterminate="isPartial(form.clients, clientOptions)"
              @change="(value) => toggleAll('clients', clientOptions, value)"
            >
              {{ t('discountCenter.taskCenter.config.all') }}
            </a-checkbox>
            <a-checkbox-group v-model="form.clients" class="check-grid">
              <a-checkbox
                v-for="item in clientOptions"
                :key="item.value"
                :value="item.value"
              >
                {{ item.label }}
              </a-checkbox>
            </a-checkbox-group>
          </div>
        </a-form-item>

        <a-form-item :label="t('discountCenter.taskCenter.config.claimLimit')">
          <a-checkbox v-model="form.oncePerDevice">
            {{ t('discountCenter.taskCenter.config.oncePerDevice') }}
          </a-checkbox>
        </a-form-item>

        <a-form-item
          :label="t('discountCenter.taskCenter.config.effect')"
          field="effect"
          required
        >
          <a-radio-group v-model="form.effect">
            <a-radio value="immediate">
              {{ t('discountCenter.taskCenter.config.effect.immediate') }}
            </a-radio>
          </a-radio-group>
        </a-form-item>
      </a-form>
    </a-spin>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, nextTick, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { FormInstance } from '@arco-design/web-vue/es/form';
  import type { FieldRule } from '@arco-design/web-vue/es/form/interface';
  import { useLevelOptions } from '@/views/memberManage/memberList/lookups';
  import type { TaskClient, TaskConfig } from '../types';

  type CheckField = 'autoLevelIds' | 'fixedLevelIds' | 'clients';
  type CheckValue = number | string;
  interface CheckOption {
    label: string;
    value: CheckValue;
  }

  const CLIENTS: TaskClient[] = ['web', 'mobile', 'android', 'ios'];

  const props = defineProps<{
    visible: boolean;
    config: TaskConfig | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'submit', value: TaskConfig): void;
  }>();

  const { t } = useI18n();
  const formRef = ref<FormInstance>();
  const loading = ref(false);
  const { autoLevels, fixedLevels, load: loadLevels } = useLevelOptions();

  const form = reactive<TaskConfig>({
    enabled: true,
    dispatchMode: 'manual',
    autoLevelIds: [],
    fixedLevelIds: [],
    washMultiplier: 1,
    clients: [...CLIENTS],
    oncePerDevice: false,
    effect: 'immediate',
  });

  const clientOptions = computed<CheckOption[]>(() =>
    CLIENTS.map((value) => ({
      label: t(`discountCenter.taskCenter.config.client.${value}`),
      value,
    }))
  );

  const levelGroups = computed(() => [
    {
      field: 'autoLevelIds' as const,
      label: 'discountCenter.taskCenter.config.autoLevels',
      options: autoLevels.value.map((item) => ({
        label: item.name,
        value: item.id,
      })),
    },
    {
      field: 'fixedLevelIds' as const,
      label: 'discountCenter.taskCenter.config.fixedLevels',
      options: fixedLevels.value.map((item) => ({
        label: item.name,
        value: item.id,
      })),
    },
  ]);

  const rules = computed<Record<string, FieldRule[]>>(() => ({
    clients: [
      {
        type: 'array',
        minLength: 1,
        message: t('discountCenter.taskCenter.config.clientsRequired'),
      },
    ],
  }));

  function isAll(selected: CheckValue[], options: CheckOption[]) {
    return (
      options.length > 0 &&
      options.every((item) => selected.includes(item.value))
    );
  }

  function isPartial(selected: CheckValue[], options: CheckOption[]) {
    return selected.length > 0 && !isAll(selected, options);
  }

  function toggleAll(
    field: CheckField,
    options: CheckOption[],
    checked: boolean | (string | number | boolean)[]
  ) {
    const values = checked === true ? options.map((item) => item.value) : [];
    if (field === 'clients') form.clients = values as TaskClient[];
    else form[field] = values as number[];
  }

  async function fill() {
    loading.value = true;
    await loadLevels();
    loading.value = false;
    const saved = props.config;
    Object.assign(form, {
      enabled: saved?.enabled ?? true,
      dispatchMode: saved?.dispatchMode ?? 'manual',
      autoLevelIds: saved
        ? [...saved.autoLevelIds]
        : autoLevels.value.map((item) => item.id),
      fixedLevelIds: saved
        ? [...saved.fixedLevelIds]
        : fixedLevels.value.map((item) => item.id),
      washMultiplier: saved ? saved.washMultiplier : 1,
      clients: saved ? [...saved.clients] : [...CLIENTS],
      oncePerDevice: saved?.oncePerDevice ?? false,
      effect: 'immediate',
    });
    await nextTick();
    formRef.value?.clearValidate();
  }

  watch(
    () => props.visible,
    (open) => {
      if (open) fill();
    }
  );

  function close() {
    emit('update:visible', false);
  }

  async function handleBeforeOk() {
    const errors = await formRef.value?.validate();
    if (errors) return false;
    emit('submit', {
      ...form,
      autoLevelIds: [...form.autoLevelIds],
      fixedLevelIds: [...form.fixedLevelIds],
      clients: [...form.clients],
    });
    close();
    return true;
  }
</script>

<style scoped lang="less">
  .config-spin {
    display: block;
  }

  :deep(.arco-input-number) {
    width: 100%;
  }

  :deep(.arco-input-number .arco-input-suffix) {
    color: var(--color-text-3);
  }

  .check-panel {
    width: 100%;
    padding: 10px 16px 12px;
    background: var(--color-fill-1);
    border-radius: 6px;
  }

  .check-all {
    margin-bottom: 8px;
    font-weight: 500;
  }

  .check-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
    gap: 10px 16px;
    padding-top: 10px;
    border-top: 1px dashed var(--color-border-2);

    :deep(.arco-checkbox) {
      margin-right: 0;
    }
  }

  .check-empty {
    color: var(--color-text-3);
    font-size: 13px;
  }
</style>
