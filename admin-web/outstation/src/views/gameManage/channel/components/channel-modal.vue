<template>
  <a-modal
    :visible="visible"
    :title="t('channelModal.title.edit')"
    :ok-loading="submitting"
    :ok-button-props="{ disabled: detailLoading || !detailChannel }"
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
    <a-spin :loading="detailLoading" style="width: 100%">
      <a-form
        ref="formRef"
        :model="form"
        label-align="right"
        :label-col-props="{ span: 6 }"
        :wrapper-col-props="{ span: 17 }"
      >
        <a-form-item
          :label="t('channelModal.field.channelCode')"
          field="channelCode"
        >
          <a-input v-model="form.channelCode" disabled />
        </a-form-item>
        <a-form-item :label="t('channelModal.field.name')" field="name">
          <a-input v-model="form.name" disabled />
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
            <a-radio :value="1">
              {{ t('channelManage.status.enabled') }}
            </a-radio>
            <a-radio :value="2">
              {{ t('channelManage.status.disabled') }}
            </a-radio>
          </a-radio-group>
        </a-form-item>
      </a-form>
    </a-spin>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FormInstance } from '@arco-design/web-vue';
  import {
    getGameChannel,
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
  const detailLoading = ref(false);
  const detailChannel = ref<GameChannelRecord | null>(null);
  const isEdit = computed(() => Boolean(props.channel?.id));

  const getDefaultForm = (): ChannelForm => ({
    channelCode: '',
    name: '',
    sort_no: 1,
    status: 1,
  });

  const form = reactive<ChannelForm>(getDefaultForm());

  const resetForm = () => {
    Object.assign(form, getDefaultForm());
    formRef.value?.clearValidate();
  };

  const fillForm = (channel: GameChannelRecord) => {
    Object.assign(form, getDefaultForm(), {
      channelCode: channel.channel_code || '--',
      name: channel.name || '--',
      sort_no: channel.sort_no && channel.sort_no > 0 ? channel.sort_no : 1,
      status: channel.status ?? 1,
    });
  };

  const fetchDetail = async (id: number) => {
    detailLoading.value = true;
    detailChannel.value = null;
    resetForm();
    try {
      const { data } = await getGameChannel(id);
      detailChannel.value = data ?? null;
      if (data) {
        fillForm(data);
      }
    } catch {
      detailChannel.value = null;
    } finally {
      detailLoading.value = false;
    }
  };

  const close = () => {
    emit('update:visible', false);
  };

  const handleCancel = () => {
    close();
    detailChannel.value = null;
    resetForm();
  };

  const buildPayload = (): UpdateGameChannelParams | null => {
    const id = detailChannel.value?.id;
    if (!id) return null;
    return {
      id,
      sort_no: form.sort_no && form.sort_no > 0 ? form.sort_no : 1,
      status: form.status ?? 1,
    };
  };

  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    const payload = buildPayload();
    if (!payload || submitting.value || detailLoading.value || !isEdit.value) {
      done(false);
      return;
    }

    submitting.value = true;
    try {
      await updateGameChannel(payload);
      Message.success(t('front.common.message.saveSuccess'));
      detailChannel.value = null;
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
      if (props.channel?.id) {
        fetchDetail(props.channel.id);
      } else {
        detailChannel.value = null;
        resetForm();
      }
    }
  );
</script>
