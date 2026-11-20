<template>
  <a-modal
    :visible="visible"
    :title="t('agentList.review.applySettings')"
    :width="640"
    :mask-closable="false"
    unmount-on-close
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel apply-settings-modal"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    :ok-text="t('agentList.settings.ok')"
    :cancel-text="t('agentList.settings.cancel')"
    @before-ok="handleBeforeOk"
    @cancel="emit('update:visible', false)"
  >
    <a-form
      :model="form"
      layout="horizontal"
      class="apply-form"
      :label-col-props="{ flex: '96px' }"
      :wrapper-col-props="{ flex: '1' }"
      label-align="right"
    >
      <a-form-item :label="t('agentList.review.apply.mode')">
        <a-radio-group v-model="form.mode">
          <a-radio value="cash">{{ t('agentList.review.apply.cash') }}</a-radio>
        </a-radio-group>
      </a-form-item>
      <a-form-item :label="t('agentList.review.apply.ratio')">
        <div class="ratio-list">
          <div class="ratio-field">
            <span class="ratio-prefix">{{
              t('agentList.review.apply.min')
            }}</span>
            <a-input-number
              v-model="form.minRatio"
              :min="0"
              :max="100"
              :precision="2"
              hide-button
              class="ratio-input"
            />
            <span class="ratio-suffix">%</span>
          </div>
          <div class="ratio-field">
            <span class="ratio-prefix">{{
              t('agentList.review.apply.max')
            }}</span>
            <a-input-number
              v-model="form.maxRatio"
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
      <a-form-item :label="t('agentList.review.apply.limit')">
        <a-space size="large">
          <a-checkbox v-model="form.limitIp">
            {{ t('agentList.review.apply.limitIp') }}
          </a-checkbox>
          <a-checkbox v-model="form.limitDevice">
            {{ t('agentList.review.apply.limitDevice') }}
          </a-checkbox>
          <a-checkbox v-model="form.limitId">
            {{ t('agentList.review.apply.limitId') }}
          </a-checkbox>
        </a-space>
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { reactive, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';

  interface ApplySettings {
    mode: 'cash';
    minRatio: number;
    maxRatio: number;
    limitIp: boolean;
    limitDevice: boolean;
    limitId: boolean;
  }

  const props = defineProps<{ visible: boolean }>();
  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
  }>();

  const { t } = useI18n();

  function createSettings(): ApplySettings {
    return {
      mode: 'cash',
      minRatio: 10,
      maxRatio: 90,
      limitIp: false,
      limitDevice: false,
      limitId: false,
    };
  }

  const saved = reactive<ApplySettings>(createSettings());
  const form = reactive<ApplySettings>(createSettings());

  watch(
    () => props.visible,
    (open) => {
      if (!open) return;
      form.mode = saved.mode;
      form.minRatio = saved.minRatio;
      form.maxRatio = saved.maxRatio;
      form.limitIp = saved.limitIp;
      form.limitDevice = saved.limitDevice;
      form.limitId = saved.limitId;
    }
  );

  function handleBeforeOk() {
    if (form.minRatio == null || form.maxRatio == null) {
      Message.warning(t('agentList.review.apply.ratioRequired'));
      return false;
    }
    if (form.minRatio > form.maxRatio) {
      Message.warning(t('agentList.review.apply.ratioInvalid'));
      return false;
    }
    saved.mode = form.mode;
    saved.minRatio = form.minRatio;
    saved.maxRatio = form.maxRatio;
    saved.limitIp = form.limitIp;
    saved.limitDevice = form.limitDevice;
    saved.limitId = form.limitId;
    Message.success(t('agentList.review.done'));
    return true;
  }
</script>

<style lang="less" scoped>
  .apply-form {
    :deep(.arco-form-item) {
      align-items: flex-start;
      margin-bottom: 20px;
    }

    :deep(.arco-form-item-label-col) {
      line-height: 32px;
      padding-right: 12px;
    }
  }

  .ratio-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
    width: 280px;
  }

  .ratio-field {
    display: flex;
    align-items: center;
    height: 32px;
    overflow: hidden;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 2px;

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
    background: transparent;
    border: none;

    :deep(.arco-input-wrapper) {
      background: transparent;
      border: none;
      box-shadow: none;
    }
  }
</style>

<style lang="less">
  .apply-settings-modal .arco-modal-footer {
    text-align: center;
  }
</style>
