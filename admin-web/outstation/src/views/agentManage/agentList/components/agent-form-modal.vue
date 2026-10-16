<template>
  <a-modal
    :visible="visible"
    :title="
      mode === 'create'
        ? t('agentList.modal.create')
        : t('agentList.modal.edit')
    "
    :width="720"
    :mask-closable="false"
    unmount-on-close
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    :ok-text="t('front.common.okText.save')"
    @before-ok="handleBeforeOk"
    @cancel="emit('update:visible', false)"
  >
    <a-form :model="form" layout="vertical">
      <a-row :gutter="16">
        <a-col :span="12">
          <a-form-item :label="t('agentList.card.account')">
            <a-input v-model="form.account" allow-clear />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="t('agentList.card.level')">
            <a-select v-model="form.level">
              <a-option
                v-for="item in settings.levels"
                :key="item.layer"
                :value="item.name"
              >
                {{ item.name }}
              </a-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="t('agentList.card.parent')">
            <a-input v-model="form.parentAgent" allow-clear />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="t('agentList.card.top')">
            <a-input v-model="form.topAgent" allow-clear />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="t('agentList.card.source')">
            <a-select v-model="form.registerSource">
              <a-option
                v-for="item in REGISTER_SOURCES"
                :key="item"
                :value="item"
              >
                {{ item }}
              </a-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="t('agentList.card.share')">
            <a-input-number
              v-model="form.shareRatio"
              :min="0"
              :max="100"
              :precision="2"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="t('agentList.card.withdrawLimit')">
            <a-select v-model="form.withdrawLimit">
              <a-option value="none">
                {{ t('agentList.withdraw.none') }}
              </a-option>
              <a-option value="forbidden">
                {{ t('agentList.withdraw.forbidden') }}
              </a-option>
              <a-option value="unset">
                {{ t('agentList.withdraw.unset') }}
              </a-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="t('agentList.card.status')">
            <a-radio-group v-model="form.status">
              <a-radio value="normal">
                {{ t('agentList.status.normal') }}
              </a-radio>
              <a-radio value="disabled">
                {{ t('agentList.status.disabled') }}
              </a-radio>
            </a-radio-group>
          </a-form-item>
        </a-col>
        <a-col :span="24">
          <a-form-item :label="t('agentList.card.remark')">
            <a-input v-model="form.remark" allow-clear />
          </a-form-item>
        </a-col>
      </a-row>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { reactive, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { REGISTER_SOURCES } from '../options';
  import type { AgentFormValue, AgentRecord, AgentSettings } from '../types';

  const props = defineProps<{
    visible: boolean;
    mode: 'create' | 'edit';
    record: AgentRecord | null;
    settings: AgentSettings;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'submit', value: AgentFormValue): void;
  }>();

  const { t } = useI18n();

  function createForm(): AgentFormValue {
    return {
      account: '',
      level: props.settings.defaultLevel,
      parentAgent: '',
      topAgent: 'cash11',
      registerSource: REGISTER_SOURCES[0],
      shareRatio: props.settings.defaultShareRatio,
      withdrawLimit: props.settings.defaultWithdrawLimit,
      remark: '',
      status: 'normal',
      promoCode: '',
    };
  }

  const form = reactive<AgentFormValue>(createForm());

  function fill(record: AgentRecord) {
    form.account = record.account;
    form.level = record.level;
    form.parentAgent = record.parentAgent;
    form.topAgent = record.topAgent;
    form.registerSource = record.registerSource;
    form.shareRatio = record.shareRatio;
    form.withdrawLimit = record.withdrawLimit;
    form.remark = record.remark;
    form.status = record.status;
    form.promoCode = record.promoCode;
  }

  watch(
    () => [props.visible, props.mode, props.record?.id] as const,
    ([open]) => {
      if (!open) return;
      if (props.mode === 'edit' && props.record) fill(props.record);
      else Object.assign(form, createForm());
    }
  );

  function handleBeforeOk() {
    const account = (form.account || '').trim();
    if (!account) {
      Message.warning(t('agentList.modal.accountRequired'));
      return false;
    }
    emit('submit', {
      ...form,
      account,
      parentAgent: (form.parentAgent || '').trim(),
      topAgent: (form.topAgent || '').trim(),
      remark: (form.remark || '').trim(),
      shareRatio: form.shareRatio ?? 0,
    });
    return true;
  }
</script>
