<template>
  <a-modal
    :visible="visible"
    :title="title"
    :width="420"
    :mask-closable="false"
    unmount-on-close
    modal-class="mac-modal-panel"
    :ok-text="t('front.common.okText.save')"
    @before-ok="handleBeforeOk"
    @cancel="emit('update:visible', false)"
  >
    <a-form :model="draft" layout="vertical" class="field-editor">
      <template v-if="isPassword">
        <a-form-item :label="t('agentList.edit.editor.newPassword')" required>
          <a-input-password
            v-model="draft.value"
            :max-length="20"
            :placeholder="t('agentList.modal.withdrawPasswordPlaceholder')"
            autocomplete="new-password"
            allow-clear
          />
        </a-form-item>
        <a-form-item
          :label="t('agentList.edit.editor.confirmPassword')"
          required
        >
          <a-input-password
            v-model="draft.confirm"
            :max-length="20"
            :placeholder="t('agentList.modal.confirmPasswordPlaceholder')"
            autocomplete="new-password"
            allow-clear
          />
        </a-form-item>
      </template>
      <a-form-item v-else-if="kind === 'share'" :label="label" required>
        <a-input-number
          v-model="draft.share"
          :min="0"
          :max="max"
          :precision="0"
          :step="1"
          :placeholder="t('agentList.modal.sharePlaceholder')"
        >
          <template #suffix>%</template>
        </a-input-number>
        <template #extra>
          {{ t('agentList.edit.editor.shareRange', { max }) }}
        </template>
      </a-form-item>
      <a-form-item v-else-if="kind === 'textarea'" :label="label">
        <a-textarea
          v-model="draft.value"
          :placeholder="t('agentList.edit.editor.placeholder', { label })"
          :max-length="200"
          show-word-limit
          :auto-size="{ minRows: 4, maxRows: 6 }"
        />
      </a-form-item>
      <a-form-item v-else-if="kind === 'phone'" :label="label">
        <a-input-group class="phone-group">
          <a-select
            v-model="draft.region"
            class="phone-code"
            :options="callingCodeOptions"
            :loading="callingCodesLoading"
            :format-label="formatCodeLabel"
            :placeholder="t('agentList.edit.editor.callingCode')"
            :trigger-props="{ autoFitPopupMinWidth: true }"
            allow-search
            allow-clear
          />
          <a-input
            v-model="draft.value"
            :placeholder="t('agentList.edit.editor.placeholder', { label })"
            allow-clear
          />
        </a-input-group>
      </a-form-item>
      <a-form-item v-else :label="label">
        <a-input
          v-model="draft.value"
          :placeholder="t('agentList.edit.editor.placeholder', { label })"
          allow-clear
        />
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, type SelectOptionData } from '@arco-design/web-vue';
  import { getPhoneCallingCodes, type PhoneCallingCode } from '@/api/agent';
  import type { FieldEditorKind } from '../types';

  let callingCodesRequest: Promise<PhoneCallingCode[]> | null = null;

  function loadCallingCodes() {
    if (!callingCodesRequest) {
      callingCodesRequest = getPhoneCallingCodes()
        .then(({ data }) => data?.list ?? [])
        .catch((error) => {
          callingCodesRequest = null;
          throw error;
        });
    }
    return callingCodesRequest;
  }

  const props = withDefaults(
    defineProps<{
      visible: boolean;
      kind: FieldEditorKind;
      label: string;
      value?: string | number;
      max?: number;
      /** 提交到接口，返回 false 时保持弹窗不关闭 */
      beforeSave?: (value: string | number) => Promise<boolean> | boolean;
    }>(),
    { value: '', max: 100, beforeSave: undefined }
  );

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'save', value: string | number): void;
  }>();

  const { t } = useI18n();

  const draft = reactive({
    value: '',
    confirm: '',
    share: undefined as number | undefined,
    region: '',
  });

  const isPassword = computed(() => props.kind === 'withdrawPassword');

  const callingCodes = ref<PhoneCallingCode[]>([]);
  const callingCodesLoading = ref(false);

  const callingCodeOptions = computed(() =>
    callingCodes.value.map((item) => ({
      value: item.region,
      label: `+${item.code} (${item.region})`,
    }))
  );

  function regionCode(region: string) {
    return callingCodes.value.find((item) => item.region === region)?.code;
  }

  function formatCodeLabel(option: SelectOptionData) {
    return `+${regionCode(String(option.value)) ?? ''}`;
  }

  /** 已有号码按 "区号-号码" 存储，也兼容 "区号 号码"、"+区号号码" 的写法 */
  function splitPhone(raw: string) {
    const text = raw.trim();
    const spaced = text.match(/^\+?(\d{1,4})[\s-]+(.+)$/);
    const codes = callingCodes.value;
    if (spaced) {
      const found = codes.find((item) => item.code === spaced[1]);
      if (found) return { region: found.region, number: spaced[2].trim() };
    }
    if (text.startsWith('+')) {
      const digits = text.slice(1).replace(/[\s-]/g, '');
      const found = [...codes]
        .sort((a, b) => b.code.length - a.code.length)
        .find((item) => digits.startsWith(item.code));
      if (found) {
        return {
          region: found.region,
          number: digits.slice(found.code.length),
        };
      }
    }
    return { region: '', number: text };
  }

  async function initPhone(raw: string) {
    draft.region = '';
    draft.value = raw;
    callingCodesLoading.value = true;
    try {
      callingCodes.value = await loadCallingCodes();
    } catch {
      return;
    } finally {
      callingCodesLoading.value = false;
    }
    if (!props.visible || draft.value !== raw) return;
    const { region, number } = splitPhone(raw);
    draft.region = region;
    draft.value = number;
  }

  const title = computed(() =>
    t('agentList.edit.editor.title', { label: props.label })
  );

  watch(
    () => props.visible,
    (open) => {
      if (!open) return;
      draft.confirm = '';
      if (props.kind === 'share') {
        draft.value = '';
        draft.share = props.value === '' ? undefined : Number(props.value);
        return;
      }
      if (props.kind === 'phone') {
        initPhone(String(props.value ?? ''));
        return;
      }
      draft.value = isPassword.value ? '' : String(props.value ?? '');
    }
  );

  function validate() {
    if (props.kind === 'share') {
      if (draft.share == null) return t('agentList.modal.shareRequired');
      return '';
    }
    if (props.kind === 'phone') {
      const number = draft.value.replace(/[\s-]/g, '');
      if (!number) return '';
      if (!/^\d+$/.test(number)) return t('agentList.edit.editor.phoneInvalid');
      if (!regionCode(draft.region)) {
        return t('agentList.edit.editor.callingCodeRequired');
      }
      return '';
    }
    if (!isPassword.value) return '';
    if (!draft.value) return t('agentList.edit.editor.passwordRequired');
    if (
      props.kind === 'withdrawPassword' &&
      !/^[\x21-\x7E]{6,20}$/.test(draft.value)
    ) {
      return t('agentList.modal.withdrawPasswordInvalid');
    }
    if (draft.value !== draft.confirm) {
      return t('agentList.modal.confirmPasswordMismatch');
    }
    return '';
  }

  function draftValue() {
    if (props.kind === 'share') return Number(draft.share);
    if (props.kind === 'phone') {
      const number = draft.value.replace(/[\s-]/g, '');
      return number ? `${regionCode(draft.region)}-${number}` : '';
    }
    return isPassword.value ? draft.value : draft.value.trim();
  }

  async function handleBeforeOk(done: (closed: boolean) => void) {
    const error = validate();
    if (error) {
      Message.warning(error);
      done(false);
      return;
    }
    const value = draftValue();
    if (props.beforeSave) {
      let ok = false;
      try {
        ok = await props.beforeSave(value);
      } catch {
        ok = false;
      }
      if (!ok) {
        done(false);
        return;
      }
    }
    emit('save', value);
    done(true);
    emit('update:visible', false);
  }
</script>

<style scoped lang="less">
  .field-editor :deep(.arco-form-item:last-child) {
    margin-bottom: 0;
  }

  .phone-group {
    display: flex;
    width: 100%;

    .phone-code {
      flex: 0 0 110px;
    }
  }
</style>
