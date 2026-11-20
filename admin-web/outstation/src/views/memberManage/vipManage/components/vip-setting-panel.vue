<template>
  <div class="setting-panel">
    <section class="setting-section">
      <div class="section-title">{{ t('vipManage.setting.section.master') }}</div>
      <div class="setting-row">
        <div class="row-label">{{ t('vipManage.setting.enable') }}</div>
        <div class="row-content">
          <div class="switch-line">
            <a-switch v-model="draft.levelEnabled" />
            <span>{{ switchText(draft.levelEnabled) }}</span>
          </div>
          <div class="row-desc">{{ t('vipManage.setting.enableDesc') }}</div>
        </div>
      </div>
    </section>

    <section class="setting-section">
      <div class="section-title">
        {{ t('vipManage.setting.section.feature') }}
      </div>
      <div class="setting-row">
        <div class="row-label">{{ t('vipManage.setting.reward') }}</div>
        <div class="row-content">
          <div class="switch-line">
            <a-switch v-model="draft.rewardEnabled" />
            <span>{{ switchText(draft.rewardEnabled) }}</span>
          </div>
          <div class="row-desc">{{ t('vipManage.setting.rewardDesc') }}</div>
        </div>
      </div>
      <div class="setting-row">
        <div class="row-label">{{ t('vipManage.setting.keep') }}</div>
        <div class="row-content">
          <div class="switch-line">
            <a-switch v-model="draft.keepEnabled" />
            <span>{{ switchText(draft.keepEnabled) }}</span>
          </div>
          <div class="row-desc">
            <div>{{ t('vipManage.setting.keepDescOn') }}</div>
            <div>{{ t('vipManage.setting.keepDescOff') }}</div>
          </div>
        </div>
      </div>
    </section>

    <section class="setting-section">
      <div class="section-title">
        {{ t('vipManage.setting.section.dispatch') }}
      </div>
      <div class="setting-row">
        <div class="row-label is-required">
          {{ t('vipManage.setting.dispatch') }}
        </div>
        <div class="row-content">
          <a-radio-group v-model="draft.dispatchMode" class="mode-group">
            <a-radio
              v-for="mode in dispatchModes"
              :key="mode"
              :value="mode"
              class="mode-option"
              :class="{ 'is-active': draft.dispatchMode === mode }"
            >
              <div class="mode-name">{{ t(`vipManage.dispatch.${mode}`) }}</div>
              <div class="row-desc">
                {{ t(`vipManage.setting.dispatchDesc.${mode}`) }}
              </div>
            </a-radio>
          </a-radio-group>
        </div>
      </div>
    </section>

    <section class="setting-section">
      <div class="section-title">
        {{ t('vipManage.setting.section.withdraw') }}
      </div>
      <div class="setting-row">
        <div class="row-label">{{ t('vipManage.setting.washMultiple') }}</div>
        <div class="row-content wash-content">
          <a-input-number
            v-model="draft.washMultiple"
            class="multiple-input"
            :min="0"
            :precision="2"
            hide-button
          >
            <template #append>{{ t('vipManage.setting.times') }}</template>
          </a-input-number>
          <div class="row-desc">
            <div>{{ t('vipManage.setting.washDesc') }}</div>
            <div>{{ t('vipManage.setting.washAccumulate') }}</div>
          </div>
        </div>
      </div>
    </section>

    <div class="panel-footer">
      <a-button type="primary" :loading="saving" @click="save">
        {{ t('vipManage.setting.save') }}
      </a-button>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { reactive, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { DispatchMode, VipSettings } from '../types';

  const props = defineProps<{
    settings: VipSettings;
    saving?: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'save', value: VipSettings): void;
  }>();

  const { t } = useI18n();
  const dispatchModes: DispatchMode[] = ['manual', 'auto'];
  const draft = reactive<VipSettings>({ ...props.settings });

  watch(
    () => props.settings,
    (value) => Object.assign(draft, value),
    { deep: true }
  );

  function switchText(enabled: boolean) {
    return t(enabled ? 'vipManage.setting.on' : 'vipManage.setting.off');
  }

  function save() {
    emit('save', { ...draft, washMultiple: draft.washMultiple ?? 0 });
  }
</script>

<style scoped lang="less">
  .setting-panel {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .setting-section {
    overflow: hidden;
    border: 1px solid var(--color-neutral-3);
    border-radius: 6px;
  }

  .section-title {
    padding: 10px 16px;
    color: var(--color-text-1);
    font-weight: 600;
    background: var(--color-fill-2);
  }

  .setting-row {
    display: flex;
    gap: 16px;
    align-items: flex-start;
    padding: 16px;

    & + & {
      border-top: 1px solid var(--color-neutral-3);
    }
  }

  .row-label {
    flex: none;
    width: 140px;
    padding-top: 2px;
    color: var(--color-text-1);
    font-weight: 500;

    &.is-required::after {
      margin-left: 2px;
      color: rgb(var(--danger-6));
      content: '*';
    }
  }

  .row-content {
    flex: 1;
    min-width: 0;
  }

  .switch-line {
    display: flex;
    gap: 10px;
    align-items: center;
    color: var(--color-text-1);
  }

  .row-desc {
    margin-top: 6px;
    color: var(--color-text-3);
    font-size: 12px;
    line-height: 20px;
  }

  .mode-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
    width: 100%;
  }

  .mode-option {
    display: flex;
    align-items: flex-start;
    box-sizing: border-box;
    width: 100%;
    margin: 0;
    padding: 10px 16px;
    border: 1px solid var(--color-neutral-3);
    border-radius: 6px;
    transition: border-color 0.2s, background-color 0.2s;

    &:hover {
      border-color: rgb(var(--primary-4));
    }

    &.is-active {
      background: rgb(var(--primary-1));
      border-color: rgb(var(--primary-5));
    }

    :deep(.arco-radio-icon-hover) {
      margin-top: 2px;
    }

    :deep(.arco-radio-label) {
      flex: 1;
    }

    .row-desc {
      margin-top: 2px;
    }
  }

  .mode-name {
    color: var(--color-text-1);
    font-weight: 500;
  }

  .wash-content {
    display: flex;
    gap: 16px;
    align-items: flex-start;

    .row-desc {
      margin-top: 0;
    }
  }

  .multiple-input {
    flex: none;
    width: 160px;
  }

  .panel-footer {
    display: flex;
    justify-content: center;
    padding-bottom: 4px;

    .arco-btn {
      min-width: 120px;
    }
  }
</style>
