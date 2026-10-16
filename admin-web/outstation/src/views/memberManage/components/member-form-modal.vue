<template>
  <a-modal
    :visible="visible"
    :title="
      mode === 'create'
        ? t('memberManage.modal.create')
        : t('memberManage.modal.edit')
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
    @cancel="close"
  >
    <a-form :model="form" layout="vertical">
      <a-row :gutter="16">
        <a-col :span="12">
          <a-form-item :label="t('memberManage.card.account')">
            <a-input
              v-model="form.account"
              :placeholder="t('memberManage.modal.accountPlaceholder')"
              allow-clear
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="t('memberManage.card.phone')">
            <a-input
              v-model="form.phone"
              :placeholder="t('memberManage.modal.phonePlaceholder')"
              allow-clear
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="t('memberManage.card.email')">
            <a-input
              v-model="form.email"
              :placeholder="t('memberManage.modal.emailPlaceholder')"
              allow-clear
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="t('memberManage.card.registerType')">
            <a-select v-model="form.registerType">
              <a-option value="email">
                {{ t('memberManage.registerType.email') }}
              </a-option>
              <a-option value="username">
                {{ t('memberManage.registerType.username') }}
              </a-option>
              <a-option value="phone">
                {{ t('memberManage.registerType.phone') }}
              </a-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="t('memberManage.modal.vip')">
            <a-input-number
              v-model="form.vipLevel"
              :min="0"
              :max="20"
              :precision="0"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="t('memberManage.card.status')">
            <a-radio-group v-model="form.status">
              <a-radio value="normal">
                {{ t('memberManage.status.normal') }}
              </a-radio>
              <a-radio value="disabled">
                {{ t('memberManage.status.disabled') }}
              </a-radio>
            </a-radio-group>
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="t('memberManage.modal.levelKind')">
            <a-radio-group v-model="form.levelKind" @change="onLevelKindChange">
              <a-radio value="auto">
                {{ t('memberManage.levelKind.auto') }}
              </a-radio>
              <a-radio value="fixed">
                {{ t('memberManage.levelKind.fixed') }}
              </a-radio>
            </a-radio-group>
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="t('memberManage.card.level')">
            <a-select v-model="form.level">
              <a-option v-for="item in levelOptions" :key="item" :value="item">
                {{ item }}
              </a-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="t('memberManage.card.source')">
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
          <a-form-item :label="t('memberManage.card.channel')">
            <a-select
              v-model="form.channelSource"
              allow-create
              allow-clear
              :placeholder="t('memberManage.modal.channelPlaceholder')"
            >
              <a-option
                v-for="item in CHANNEL_PRESETS"
                :key="item"
                :value="item"
              >
                {{ item }}
              </a-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="t('memberManage.modal.parent')">
            <a-input v-model="form.parentAgent" allow-clear />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="t('memberManage.modal.parentId')">
            <a-input v-model="form.parentAgentId" allow-clear />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="t('memberManage.card.referral')">
            <a-select v-model="form.referralStatus">
              <a-option value="normal">
                {{ t('memberManage.referral.normal') }}
              </a-option>
              <a-option value="none">{{ t('memberManage.empty') }}</a-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="t('memberManage.search.betBanned')">
            <a-switch v-model="form.betBanned" />
          </a-form-item>
        </a-col>
      </a-row>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { CHANNEL_PRESETS, REGISTER_SOURCES, levelsOf } from '../options';
  import type { LevelKind, MemberFormValue, MemberRecord } from '../types';

  const props = defineProps<{
    visible: boolean;
    mode: 'create' | 'edit';
    record: MemberRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'submit', value: MemberFormValue): void;
  }>();

  const { t } = useI18n();

  function createForm(): MemberFormValue {
    return {
      account: '',
      phone: '',
      email: '',
      vipLevel: 2,
      registerType: 'email',
      level: '默认层级',
      levelKind: 'auto',
      registerSource: '官网(PC注册)',
      channelSource: '',
      parentAgent: '',
      parentAgentId: '',
      referralStatus: 'normal',
      status: 'normal',
      betBanned: false,
    };
  }

  const form = reactive<MemberFormValue>(createForm());

  const levelOptions = computed(() => levelsOf(form.levelKind));

  function fill(record: MemberRecord) {
    form.account = record.account;
    form.phone = record.phone;
    form.email = record.email;
    form.vipLevel = record.vipLevel;
    form.registerType = record.registerType;
    form.level = record.level;
    form.levelKind = record.levelKind;
    form.registerSource = record.registerSource;
    form.channelSource = record.channelSource;
    form.parentAgent = record.parentAgent;
    form.parentAgentId = record.parentAgentId;
    form.referralStatus = record.referralStatus;
    form.status = record.status;
    form.betBanned = record.betBanned;
  }

  watch(
    () => [props.visible, props.mode, props.record?.id] as const,
    ([open]) => {
      if (!open) return;
      if (props.mode === 'edit' && props.record) fill(props.record);
      else Object.assign(form, createForm());
    }
  );

  function onLevelKindChange(value: string | number | boolean) {
    const kind = value as LevelKind;
    const options = levelsOf(kind);
    const [firstLevel] = options;
    if (!options.includes(form.level) && firstLevel) form.level = firstLevel;
  }

  function close() {
    emit('update:visible', false);
  }

  function handleBeforeOk() {
    const account = (form.account || '').trim();
    const phone = (form.phone || '').trim();
    const email = (form.email || '').trim();
    if (!account && !phone && !email) {
      Message.warning(t('memberManage.modal.identityRequired'));
      return false;
    }
    if (email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      Message.warning(t('memberManage.modal.emailInvalid'));
      return false;
    }
    emit('submit', {
      ...form,
      account,
      phone,
      email,
      parentAgent: (form.parentAgent || '').trim(),
      parentAgentId: (form.parentAgentId || '').trim(),
      channelSource: (form.channelSource || '').trim(),
      vipLevel: form.vipLevel ?? 0,
    });
    return true;
  }
</script>
