<template>
  <a-modal
    :visible="visible"
    :title="t('manufacturerModal.title.edit')"
    :ok-loading="submitting"
    :ok-button-props="{ disabled: detailLoading || !detailProvider }"
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
          :label="t('manufacturerModal.field.providerCode')"
          field="providerCode"
        >
          <a-input v-model="form.providerCode" disabled />
        </a-form-item>
        <a-form-item
          :label="t('manufacturerModal.field.channelCode')"
          field="channelCode"
        >
          <a-input v-model="form.channelCode" disabled />
        </a-form-item>
        <a-form-item :label="t('manufacturerModal.field.name')" field="name">
          <a-input v-model="form.name" disabled />
        </a-form-item>
        <a-form-item
          :label="t('manufacturerModal.field.channelName')"
          field="channelName"
        >
          <a-input v-model="form.channelName" disabled />
        </a-form-item>
        <a-form-item :label="t('manufacturerModal.field.logoPreview')">
          <a-image
            v-if="form.logo_url"
            :src="form.logo_url"
            width="60"
            height="60"
            fit="contain"
          />
          <span v-else>--</span>
        </a-form-item>
        <a-form-item :label="t('manufacturerModal.field.sortNo')" field="sort_no">
          <a-input-number
            v-model="form.sort_no"
            :min="1"
            :max="9999"
            :placeholder="t('manufacturerModal.field.sortNoPlaceholder')"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item :label="t('manufacturerModal.field.status')" field="status">
          <a-radio-group v-model="form.status">
            <a-radio :value="1">
              {{ t('manufacturerManage.status.enabled') }}
            </a-radio>
            <a-radio :value="2">
              {{ t('manufacturerManage.status.disabled') }}
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
    getGameProvider,
    updateGameProvider,
    type GameProviderRecord,
    type UpdateGameProviderParams,
  } from '@/api/gameProvider';

  const { t } = useI18n();

  const props = defineProps<{
    visible: boolean;
    provider?: GameProviderRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  interface ProviderForm {
    providerCode: string;
    channelCode: string;
    name: string;
    channelName: string;
    logo_url: string;
    sort_no: number;
    status: number;
  }

  const formRef = ref<FormInstance>();
  const submitting = ref(false);
  const detailLoading = ref(false);
  const detailProvider = ref<GameProviderRecord | null>(null);
  const isEdit = computed(() => Boolean(props.provider?.id));

  const getDefaultForm = (): ProviderForm => ({
    providerCode: '',
    channelCode: '',
    name: '',
    channelName: '',
    logo_url: '',
    sort_no: 1,
    status: 1,
  });

  const form = reactive<ProviderForm>(getDefaultForm());

  const resetForm = () => {
    Object.assign(form, getDefaultForm());
    formRef.value?.clearValidate();
  };

  const fillForm = (provider: GameProviderRecord) => {
    Object.assign(form, getDefaultForm(), {
      providerCode: provider.provider_code || '--',
      channelCode: provider.channel_code || '--',
      name: provider.name || '--',
      channelName: provider.channel_name || '--',
      logo_url: provider.logo_url || '',
      sort_no: provider.sort_no && provider.sort_no > 0 ? provider.sort_no : 1,
      status: provider.status ?? 1,
    });
  };

  const fetchDetail = async (id: number) => {
    detailLoading.value = true;
    detailProvider.value = null;
    resetForm();
    try {
      const { data } = await getGameProvider(id);
      detailProvider.value = data ?? null;
      if (data) {
        fillForm(data);
      }
    } catch {
      detailProvider.value = null;
    } finally {
      detailLoading.value = false;
    }
  };

  const close = () => {
    emit('update:visible', false);
  };

  const handleCancel = () => {
    close();
    detailProvider.value = null;
    resetForm();
  };

  const buildPayload = (): UpdateGameProviderParams | null => {
    const id = detailProvider.value?.id;
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
      await updateGameProvider(payload);
      Message.success(t('front.common.message.saveSuccess'));
      detailProvider.value = null;
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
      if (props.provider?.id) {
        fetchDetail(props.provider.id);
      } else {
        detailProvider.value = null;
        resetForm();
      }
    }
  );
</script>
