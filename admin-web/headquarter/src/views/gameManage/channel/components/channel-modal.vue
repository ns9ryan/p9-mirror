<template>
  <a-modal
    :visible="visible"
    :title="t('channelModal.title.edit')"
    :ok-loading="submitting"
    :ok-text="t('front.common.okText.save')"
    :mask-closable="false"
    :esc-to-close="false"
    unmount-on-close
    :width="680"
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    @before-ok="handleBeforeOk"
    @cancel="handleCancel"
  >
    <a-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-align="right"
      :label-col-props="{ span: 6 }"
      :wrapper-col-props="{ span: 17 }"
    >
      <a-form-item :label="t('channelModal.field.channelCode')" field="channelCode">
        <a-input v-model="form.channelCode" disabled />
      </a-form-item>
      <a-form-item :label="t('channelModal.field.name')" field="name" required>
        <a-textarea
          v-model="form.name"
          :placeholder="t('channelModal.field.namePlaceholder')"
          allow-clear
          :auto-size="{ minRows: 4, maxRows: 8 }"
        />
      </a-form-item>
      <a-form-item :label="t('channelModal.field.sortNo')" field="sort_no">
        <a-input-number
          v-model="form.sort_no"
          :min="1"
          :max="9999"
          :placeholder="t('channelModal.field.sortNoPlaceholder')"
          style="width: 100%"
        />
      </a-form-item>
      <a-form-item :label="t('channelModal.field.status')" field="status">
        <a-radio-group v-model="form.status">
          <a-radio :value="1">{{ t('front.common.status.enabled') }}</a-radio>
          <a-radio :value="2"
            >{{ t('front.common.status.disabled')
            }}<span style="color: red">{{
              t('channelModal.field.forceLogoutHint')
            }}</span></a-radio
          >
        </a-radio-group>
      </a-form-item>
      <a-form-item :label="t('channelModal.field.note')">
        <a-alert type="info" :show-icon="false">
          {{ t('channelModal.field.noteContent') }}
        </a-alert>
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FieldRule, FormInstance } from '@arco-design/web-vue';
  import {
    updateGameChannel,
    type GameChannelRecord,
    type UpdateGameChannelParams,
  } from '@/api/gameChannel';

  const { t } = useI18n();

  const props = defineProps<{
    visible: boolean;
    channel?: GameChannelRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  interface ChannelForm {
    channelCode: string;
    name: string;
    sort_no: number;
    status: number;
  }

  const formRef = ref<FormInstance>();
  const submitting = ref(false);
  const isEdit = computed(() => Boolean(props.channel?.id));

  const getDefaultForm = (): ChannelForm => ({
    channelCode: '',
    name: '',
    sort_no: 1,
    status: 1,
  });

  const form = reactive<ChannelForm>(getDefaultForm());

  const rules = computed<Record<string, FieldRule | FieldRule[]>>(() => ({
    name: [
      { required: true, message: t('channelModal.rule.nameRequired') },
      {
        validator: (value: string, callback: (error?: string) => void) => {
          try {
            const parsed = JSON.parse(value || '');
            if (!parsed || typeof parsed !== 'object') {
              callback(t('channelModal.rule.nameObject'));
              return;
            }
            callback();
          } catch {
            callback(t('channelModal.rule.nameJson'));
          }
        },
      },
    ],
  }));

  const prettyJson = (value: string) => {
    try {
      return JSON.stringify(JSON.parse(value), null, 2);
    } catch {
      return value;
    }
  };

  const resetForm = () => {
    Object.assign(form, getDefaultForm());
    formRef.value?.clearValidate();
  };

  const fillForm = (channel: GameChannelRecord) => {
    Object.assign(form, getDefaultForm(), {
      channelCode: channel.channel_code || '--',
      name: prettyJson(channel.name || ''),
      sort_no: channel.sort_no && channel.sort_no > 0 ? channel.sort_no : 1,
      status: channel.status ?? 1,
    });
  };

  const close = () => {
    emit('update:visible', false);
  };

  const handleCancel = () => {
    close();
    resetForm();
  };

  const buildPayload = (): UpdateGameChannelParams => ({
    id: props.channel!.id,
    name: form.name.trim(),
    sort_no: form.sort_no && form.sort_no > 0 ? form.sort_no : 1,
    status: form.status ?? 1,
    force_logout: form.status === 2,
  });

  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    const errors = await formRef.value?.validate();
    if (errors || submitting.value || !isEdit.value || !props.channel?.id) {
      done(false);
      return;
    }

    submitting.value = true;
    try {
      await updateGameChannel(buildPayload());
      Message.success(t('front.common.message.saveSuccess'));
      resetForm();
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
      resetForm();
      if (props.channel?.id) {
        fillForm(props.channel);
      }
    }
  );
</script>
