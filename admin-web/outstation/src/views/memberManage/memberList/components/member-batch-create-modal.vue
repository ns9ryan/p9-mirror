<template>
  <a-modal
    :visible="visible"
    :width="640"
    :mask-closable="false"
    :esc-to-close="!submitting"
    :closable="!submitting"
    unmount-on-close
    title-align="start"
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel member-batch-create-modal"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    :ok-text="t('memberManage.batchCreate.submit')"
    :cancel-text="t('memberManage.batchCreate.cancel')"
    :ok-loading="submitting"
    :cancel-button-props="{ disabled: submitting }"
    @before-ok="handleBeforeOk"
    @cancel="close"
  >
    <template #title>
      <div class="batch-title">
        <div class="batch-title-main">{{ t('memberManage.batchCreate.title') }}</div>
        <div class="batch-title-sub">
          {{ t('memberManage.batchCreate.subtitle') }}
        </div>
      </div>
    </template>

    <a-form
      ref="formRef"
      :model="form"
      :rules="rules"
      layout="vertical"
      asterisk-position="end"
    >
      <a-row :gutter="16">
        <a-col :span="12">
          <a-form-item
            field="prefix"
            :label="t('memberManage.batchCreate.prefix')"
          >
            <a-input
              v-model="form.prefix"
              :max-length="16"
              :disabled="submitting"
              :placeholder="t('memberManage.batchCreate.affixPlaceholder')"
              allow-clear
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item
            field="suffix"
            :label="t('memberManage.batchCreate.suffix')"
          >
            <a-input
              v-model="form.suffix"
              :max-length="16"
              :disabled="submitting"
              :placeholder="t('memberManage.batchCreate.affixPlaceholder')"
              allow-clear
            />
          </a-form-item>
        </a-col>
      </a-row>
      <div class="field-extra">{{ t('memberManage.batchCreate.affixExtra') }}</div>

      <a-row :gutter="16">
        <a-col :span="12">
          <a-form-item
            field="count"
            :label="t('memberManage.batchCreate.count')"
            :extra="t('memberManage.batchCreate.countExtra')"
          >
            <a-input-number
              v-model="form.count"
              :min="1"
              :max="2000"
              :precision="0"
              :disabled="submitting"
              :placeholder="t('memberManage.batchCreate.countPlaceholder')"
              mode="button"
              class="full-width"
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item
            field="agentId"
            :label="t('memberManage.batchCreate.agent')"
          >
            <a-select
              v-model="form.agentId"
              allow-search
              allow-clear
              :loading="agentLoading"
              :disabled="submitting"
              :filter-option="false"
              :placeholder="t('memberManage.batchCreate.agentPlaceholder')"
              :format-label="formatAgentLabel"
              @search="searchAgents"
              @change="onAgentChange"
            >
              <a-option
                v-for="item in agentOptions"
                :key="item.id"
                :value="item.id"
                :label="item.account"
              >
                {{ item.account }} / {{ item.id }}
              </a-option>
            </a-select>
          </a-form-item>
        </a-col>
      </a-row>

      <div class="field-extra">{{ t('memberManage.batchCreate.defaultsTip') }}</div>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import axios from 'axios';
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type {
    FieldRule,
    FormInstance,
    SelectOptionData,
  } from '@arco-design/web-vue';
  import type { AgentListItem } from '@/api/agent';
  import { useAgentSearch } from '../lookups';
  import { useMemberStore } from '../store';

  const props = defineProps<{
    visible: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    /** 创建成功或请求超时（结果未知）时触发，用于刷新列表 */
    (e: 'refresh'): void;
  }>();

  const { t } = useI18n();
  const { batchCreateMembers } = useMemberStore();
  const {
    agents,
    loading: agentLoading,
    search: searchAgents,
  } = useAgentSearch(true);

  /** 前缀或后缀单独允许的字符 */
  const AFFIX_PATTERN = /^[A-Za-z0-9_-]*$/;
  /** 前缀与后缀合计上限 */
  const AFFIX_MAX_LENGTH = 16;
  const COUNT_MIN = 1;
  const COUNT_MAX = 2000;

  function createForm() {
    return {
      prefix: '',
      suffix: '',
      count: undefined as number | undefined,
      agentId: undefined as number | undefined,
    };
  }

  const formRef = ref<FormInstance>();
  const form = reactive(createForm());
  const submitting = ref(false);

  /** 搜索结果会被下一次搜索替换，选中项单独保存并始终保留在选项里 */
  const selectedAgent = ref<AgentListItem | null>(null);
  const agentOptions = computed(() => {
    const selected = selectedAgent.value;
    if (!selected) return agents.value;
    return [
      selected,
      ...agents.value.filter((item) => item.id !== selected.id),
    ];
  });

  function onAgentChange(value: unknown) {
    selectedAgent.value =
      agentOptions.value.find((item) => item.id === value) || null;
  }

  function formatAgentLabel(data: SelectOptionData) {
    const agent = selectedAgent.value;
    if (!agent || agent.id !== data.value) return String(data.label ?? '');
    return `${agent.account} / ${agent.id}`;
  }

  function validateAffix(_value: string, callback: (error?: string) => void) {
    const { prefix, suffix } = form;
    if (!prefix && !suffix) {
      callback(t('memberManage.batchCreate.affixRequired'));
      return;
    }
    if (!AFFIX_PATTERN.test(prefix) || !AFFIX_PATTERN.test(suffix)) {
      callback(t('memberManage.batchCreate.affixInvalid'));
      return;
    }
    if (prefix.length + suffix.length > AFFIX_MAX_LENGTH) {
      callback(t('memberManage.batchCreate.affixTooLong'));
      return;
    }
    callback();
  }

  const rules = computed<Record<string, FieldRule[]>>(() => ({
    prefix: [{ validator: validateAffix }],
    suffix: [{ validator: validateAffix }],
    count: [
      { required: true, message: t('memberManage.batchCreate.countRequired') },
      {
        validator: (value: number, callback: (error?: string) => void) => {
          if (
            !Number.isInteger(value) ||
            value < COUNT_MIN ||
            value > COUNT_MAX
          ) {
            callback(t('memberManage.batchCreate.countInvalid'));
            return;
          }
          callback();
        },
      },
    ],
    agentId: [
      { required: true, message: t('memberManage.batchCreate.agentRequired') },
    ],
  }));

  watch(
    () => props.visible,
    (open) => {
      if (!open) return;
      Object.assign(form, createForm());
      selectedAgent.value = null;
      searchAgents();
    }
  );

  function close() {
    if (submitting.value) return;
    emit('update:visible', false);
  }

  /** 超时或断线时服务端结果未知，需先查询列表确认 */
  function isTimeoutError(error: unknown) {
    return (
      axios.isAxiosError(error) &&
      (error.code === 'ECONNABORTED' || error.code === 'ETIMEDOUT')
    );
  }

  async function handleBeforeOk() {
    if (submitting.value) return false;
    const errors = await formRef.value?.validate();
    if (errors || form.agentId == null || form.count == null) return false;

    const count = form.count;
    submitting.value = true;
    try {
      await batchCreateMembers({
        account_prefix: form.prefix || undefined,
        account_suffix: form.suffix || undefined,
        count,
        agent_id: form.agentId,
      });
      Message.success(t('memberManage.batchCreate.success', { count }));
      emit('refresh');
      emit('update:visible', false);
      return true;
    } catch (error) {
      if (isTimeoutError(error)) {
        Message.warning(t('memberManage.batchCreate.timeout'));
        emit('refresh');
        emit('update:visible', false);
        return true;
      }
      return false;
    } finally {
      submitting.value = false;
    }
  }
</script>

<style scoped lang="less">
  .batch-title-main {
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 18px;
    line-height: 26px;
  }

  .batch-title-sub {
    margin-top: 2px;
    color: var(--color-text-3);
    font-weight: 400;
    font-size: 13px;
    line-height: 20px;
  }

  .full-width {
    width: 100%;
  }

  .field-extra {
    margin: -4px 0 14px;
    color: var(--color-text-3);
    font-size: 12px;
    line-height: 18px;
  }

  :deep(.arco-form-item-label) {
    color: var(--color-text-1);
    font-weight: 500;
  }
</style>

<style lang="less">
  .member-batch-create-modal {
    .arco-modal-header {
      height: auto;
      padding: 18px 24px 14px;
      border-bottom: none;
    }

    .arco-modal-body {
      padding: 0 24px 16px;
    }

    .arco-modal-footer {
      padding: 14px 24px;
    }
  }
</style>
