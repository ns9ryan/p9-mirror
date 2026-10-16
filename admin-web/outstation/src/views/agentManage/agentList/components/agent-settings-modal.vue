<template>
  <a-modal
    :visible="visible"
    :title="t('agentList.settings.title')"
    :width="720"
    :mask-closable="false"
    unmount-on-close
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel agent-settings-modal"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    :ok-text="t('agentList.settings.ok')"
    :cancel-text="t('agentList.settings.cancel')"
    @before-ok="handleBeforeOk"
    @cancel="emit('update:visible', false)"
  >
    <a-form
      :model="form"
      layout="horizontal"
      class="agent-settings-form"
      :label-col-props="{ flex: '128px' }"
      :wrapper-col-props="{ flex: '1' }"
      label-align="left"
    >
      <a-form-item :label="t('agentList.settings.openLevels')" required>
        <div class="level-list">
          <div v-for="item in form.levels" :key="item.layer" class="level-row">
            <span class="level-layer">
              {{ t('agentList.settings.layer', { n: item.layer }) }}
            </span>
            <a-input v-model="item.name" class="level-name" :max-length="16" />
          </div>
        </div>
      </a-form-item>
      <a-form-item :label="t('agentList.settings.maxShare')" required>
        <div class="ratio-field">
          <a-input-number
            v-model="form.maxShareRatio"
            :min="0"
            :max="100"
            :precision="2"
            hide-button
            class="ratio-input"
          />
          <span class="ratio-suffix">%</span>
        </div>
      </a-form-item>
      <a-form-item :label="t('agentList.settings.applyShare')" required>
        <div class="apply-share">
          <div class="ratio-field">
            <span class="ratio-prefix">
              {{ t('agentList.settings.highest') }}
            </span>
            <a-input-number
              v-model="form.applyShareMax"
              :min="0"
              :max="100"
              :precision="2"
              hide-button
              class="ratio-input"
            />
            <span class="ratio-suffix">%</span>
          </div>
          <div class="ratio-field">
            <span class="ratio-prefix">
              {{ t('agentList.settings.lowest') }}
            </span>
            <a-input-number
              v-model="form.applyShareMin"
              :min="0"
              :max="100"
              :precision="2"
              hide-button
              class="ratio-input"
            />
            <span class="ratio-suffix">%</span>
          </div>
        </div>
      </a-form-item>
    </a-form>
    <div class="settings-note">
      <span class="settings-note-mark">*</span>
      {{ t('agentList.settings.hint') }}
    </div>
  </a-modal>
</template>

<script lang="ts" setup>
  import { reactive, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { AgentSettings } from '../types';

  const props = defineProps<{
    visible: boolean;
    settings: AgentSettings;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'submit', value: AgentSettings): void;
  }>();

  const { t } = useI18n();

  function cloneSettings(source: AgentSettings): AgentSettings {
    return {
      ...source,
      levels: source.levels.map((item) => ({ ...item })),
    };
  }

  const form = reactive<AgentSettings>(cloneSettings(props.settings));

  function resetForm() {
    const next = cloneSettings(props.settings);
    form.levels = next.levels;
    form.maxShareRatio = next.maxShareRatio;
    form.applyShareMax = next.applyShareMax;
    form.applyShareMin = next.applyShareMin;
    form.defaultLevel = next.defaultLevel;
    form.defaultShareRatio = next.defaultShareRatio;
    form.defaultWithdrawLimit = next.defaultWithdrawLimit;
  }

  watch(
    () => props.visible,
    (open) => {
      if (open) resetForm();
    }
  );

  function handleBeforeOk() {
    const levels = form.levels.map((item) => ({
      layer: item.layer,
      name: item.name.trim(),
    }));
    if (levels.some((item) => !item.name)) {
      Message.warning(t('agentList.settings.levelRequired'));
      return false;
    }
    const uniqueNames = new Set(levels.map((item) => item.name));
    if (uniqueNames.size !== levels.length) {
      Message.warning(t('agentList.settings.levelDuplicate'));
      return false;
    }
    const maxShare = form.maxShareRatio;
    const applyMax = form.applyShareMax;
    const applyMin = form.applyShareMin;
    if (maxShare == null || applyMax == null || applyMin == null) {
      Message.warning(t('agentList.settings.required'));
      return false;
    }
    if (applyMax > maxShare) {
      Message.warning(t('agentList.settings.applyMaxInvalid'));
      return false;
    }
    if (applyMin > applyMax) {
      Message.warning(t('agentList.settings.applyMinInvalid'));
      return false;
    }
    const defaultLayer = props.settings.levels.find(
      (item) => item.name === form.defaultLevel
    )?.layer;
    const defaultLevel =
      levels.find((item) => item.layer === defaultLayer)?.name ||
      form.defaultLevel;
    emit('submit', {
      levels,
      maxShareRatio: maxShare,
      applyShareMax: applyMax,
      applyShareMin: applyMin,
      defaultLevel,
      defaultShareRatio: applyMax,
      defaultWithdrawLimit: form.defaultWithdrawLimit,
    });
    return true;
  }
</script>

<style lang="less" scoped>
  .agent-settings-form {
    :deep(.arco-form-item) {
      align-items: flex-start;
      margin-bottom: 16px;
    }

    :deep(.arco-form-item-label-col) {
      line-height: 32px;
      padding-right: 8px;
    }

    :deep(.arco-form-item-label) {
      color: var(--color-text-1);
    }
  }

  .level-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
    width: 100%;
  }

  .level-row {
    display: flex;
    align-items: stretch;
    height: 32px;
    border: 1px solid var(--color-border-2);
    border-radius: 2px;
    background: var(--color-bg-2);
    overflow: hidden;

    &:focus-within {
      border-color: rgb(var(--primary-6));
    }
  }

  .level-layer {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 72px;
    flex-shrink: 0;
    border-right: 1px solid var(--color-border-2);
    background: var(--color-fill-2);
    color: var(--color-text-2);
  }

  .level-name {
    flex: 1;
    min-width: 0;
    border: none;
    background: transparent;

    :deep(.arco-input-wrapper) {
      border: none;
      background: transparent;
      box-shadow: none;
    }
  }

  .apply-share {
    display: flex;
    gap: 16px;
    width: 100%;
  }

  .ratio-field {
    display: flex;
    align-items: center;
    flex: 1;
    min-width: 0;
    height: 32px;
    border: 1px solid var(--color-border-2);
    border-radius: 2px;
    background: var(--color-bg-2);
    overflow: hidden;

    &:focus-within {
      border-color: rgb(var(--primary-6));
    }
  }

  .ratio-prefix,
  .ratio-suffix {
    flex-shrink: 0;
    color: var(--color-text-2);
  }

  .ratio-prefix {
    padding-left: 12px;
  }

  .ratio-suffix {
    padding-right: 12px;
  }

  .ratio-input {
    flex: 1;
    min-width: 0;
    border: none;
    background: transparent;

    :deep(.arco-input-wrapper) {
      border: none;
      background: transparent;
      box-shadow: none;
    }

    :deep(.arco-input) {
      text-align: left;
    }
  }

  .settings-note {
    margin: 4px 0 8px 128px;
    color: var(--color-text-1);
    font-size: 14px;
    line-height: 22px;
  }

  .settings-note-mark {
    margin-right: 4px;
    color: rgb(var(--danger-6));
  }
</style>

<style lang="less">
  .agent-settings-modal .arco-modal-footer {
    text-align: center;
  }
</style>
