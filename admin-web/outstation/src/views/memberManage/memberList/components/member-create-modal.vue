<template>
  <a-modal
    :visible="visible"
    :width="800"
    :mask-closable="false"
    unmount-on-close
    title-align="start"
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel member-create-modal"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    :ok-text="t('memberManage.create.submit')"
    :cancel-text="t('memberManage.create.cancel')"
    @before-ok="handleBeforeOk"
    @cancel="close"
  >
    <template #title>
      <div class="create-title">
        <div class="create-title-main">{{ t('memberManage.create.title') }}</div>
        <div class="create-title-sub">
          {{ t('memberManage.create.subtitle') }}
        </div>
      </div>
    </template>

    <a-form
      ref="formRef"
      :model="form"
      :rules="rules"
      layout="vertical"
      asterisk-position="end"
      class="create-form"
    >
      <section class="create-section">
        <div class="section-head">
          <span class="section-index">1</span>
          {{ t('memberManage.create.sectionAccount') }}
        </div>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item
              field="account"
              :label="t('memberManage.create.account')"
              :extra="t('memberManage.create.accountExtra')"
            >
              <a-input
                v-model="form.account"
                :max-length="20"
                :placeholder="t('memberManage.create.accountPlaceholder')"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item
              field="password"
              :label="t('memberManage.create.password')"
              :extra="t('memberManage.create.passwordExtra')"
            >
              <a-input-password
                v-model="form.password"
                :max-length="32"
                :placeholder="t('memberManage.create.passwordPlaceholder')"
              />
            </a-form-item>
          </a-col>
          <a-col :span="24">
            <a-form-item
              field="nickname"
              :label="t('memberManage.create.nickname')"
              :extra="t('memberManage.create.nicknameExtra')"
            >
              <a-input
                v-model="form.nickname"
                :max-length="100"
                :placeholder="t('memberManage.create.nicknamePlaceholder')"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item field="phone" :label="t('memberManage.create.phone')">
              <a-input-group class="phone-group">
                <a-select
                  v-model="form.phoneRegion"
                  class="phone-code"
                  :options="callingCodeOptions"
                  :loading="callingCodesLoading"
                  :format-label="formatCodeLabel"
                  :placeholder="t('memberManage.create.callingCode')"
                  :trigger-props="{ autoFitPopupMinWidth: true }"
                  allow-search
                  allow-clear
                />
                <a-input
                  v-model="form.phone"
                  :max-length="20"
                  :placeholder="t('memberManage.create.phonePlaceholder')"
                  allow-clear
                />
              </a-input-group>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item field="email" :label="t('memberManage.create.email')">
              <a-input
                v-model="form.email"
                :max-length="100"
                :placeholder="t('memberManage.create.emailPlaceholder')"
                allow-clear
              />
            </a-form-item>
          </a-col>
        </a-row>
      </section>

      <section class="create-section">
        <div class="section-head">
          <span class="section-index">2</span>
          {{ t('memberManage.create.sectionAgent') }}
        </div>
        <a-form-item
          field="agentId"
          :label="t('memberManage.create.agent')"
          :extra="selectedAgent ? undefined : t('memberManage.create.agentExtra')"
          class="agent-item"
        >
          <a-select
            v-model="form.agentId"
            allow-search
            allow-clear
            :loading="agentLoading"
            :filter-option="false"
            :placeholder="t('memberManage.create.agentPlaceholder')"
            :format-label="formatAgentLabel"
            :trigger-props="{ contentClass: 'member-agent-dropdown' }"
            @search="searchAgents"
            @change="onAgentChange"
          >
            <template v-if="!form.agentId" #prefix>
              <icon-search />
            </template>
            <a-option
              v-for="item in agentOptions"
              :key="item.id"
              :value="item.id"
              :label="item.account"
            >
              <div class="agent-option">
                <span>{{ item.account }}</span>
                <span>{{ item.id }}</span>
                <span>{{ agentLevelText(item) }}</span>
                <span class="agent-option-line">
                  {{ item.parent_account || t('memberManage.create.noParent') }}
                </span>
              </div>
            </a-option>
          </a-select>
        </a-form-item>

        <div v-if="selectedAgent" class="agent-card">
          <div class="agent-card-head">
            <span class="agent-card-icon"><icon-user /></span>
            {{ t('memberManage.create.agentSelected') }}
          </div>
          <div class="agent-card-body">
            <div
              v-for="(column, index) in agentColumns"
              :key="index"
              class="agent-card-col"
            >
              <div v-for="field in column" :key="field.label" class="agent-field">
                <span class="agent-field-label">{{ field.label }}</span>
                <span class="agent-field-value">{{ field.value }}</span>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="agent-empty">
          <icon-info-circle-fill class="agent-empty-icon" />
          <div>
            <div class="agent-empty-title">
              {{ t('memberManage.create.agentEmptyTitle') }}
            </div>
            <div class="agent-empty-desc">
              {{ t('memberManage.create.agentEmptyDesc') }}
            </div>
          </div>
        </div>
        <div v-if="selectedAgent" class="field-extra">
          {{ t('memberManage.create.agentSelectedExtra') }}
        </div>
      </section>

      <section class="create-section">
        <div class="section-head">
          <span class="section-index">3</span>
          {{ t('memberManage.create.sectionSetting') }}
        </div>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item :label="t('memberManage.create.vip')">
              <a-select
                v-model="form.vipLevelId"
                disabled
                :placeholder="t('memberManage.create.vipPlaceholder')"
              >
                <a-option
                  v-for="item in vipLevels"
                  :key="item.id"
                  :value="item.id"
                >
                  VIP{{ item.level }}
                </a-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item :label="t('memberManage.create.status')">
              <a-radio-group v-model="form.status">
                <a-radio v-for="item in STATUSES" :key="item" :value="item">
                  {{ t(`memberManage.status.${item}`) }}
                </a-radio>
              </a-radio-group>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item :label="t('memberManage.create.autoLevel')">
              <a-select
                v-model="form.autoLevelId"
                allow-clear
                :placeholder="t('memberManage.create.autoLevelPlaceholder')"
              >
                <a-option
                  v-for="item in autoLevels"
                  :key="item.id"
                  :value="item.id"
                >
                  {{ item.name }}
                </a-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item :label="t('memberManage.create.fixedLevel')">
              <a-select
                v-model="form.fixedLevelIds"
                multiple
                allow-clear
                :placeholder="t('memberManage.create.fixedLevelPlaceholder')"
              >
                <a-option
                  v-for="item in fixedLevels"
                  :key="item.id"
                  :value="item.id"
                >
                  {{ item.name }}
                </a-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item :label="t('memberManage.create.betBanned')">
              <a-switch v-model="form.betBanned" />
            </a-form-item>
          </a-col>
        </a-row>
      </section>

      <section class="create-section">
        <div class="section-head">
          <span class="section-index">4</span>
          {{ t('memberManage.create.sectionOther') }}
        </div>
        <a-form-item field="remark" :label="t('memberManage.create.remark')">
          <a-textarea
            v-model="form.remark"
            :max-length="200"
            show-word-limit
            :auto-size="{ minRows: 2, maxRows: 4 }"
            :placeholder="t('memberManage.create.remarkPlaceholder')"
          />
        </a-form-item>
      </section>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type {
    FieldRule,
    FormInstance,
    SelectOptionData,
  } from '@arco-design/web-vue';
  import type { AgentListItem } from '@/api/agent';
  import {
    useAgentSearch,
    useCallingCodes,
    useLevelOptions,
    useVipLevelOptions,
  } from '../lookups';
  import { useMemberStore } from '../store';
  import type { AccountStatus } from '../types';

  const props = defineProps<{
    visible: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'created', value: { id: number; code: string }): void;
  }>();

  const { t } = useI18n();
  const { createMember } = useMemberStore();
  const {
    agents,
    loading: agentLoading,
    search: searchAgents,
  } = useAgentSearch(true);
  const { autoLevels, fixedLevels, load: loadLevels } = useLevelOptions();
  const { vipLevels, load: loadVipLevels } = useVipLevelOptions();
  const {
    callingCodes,
    loading: callingCodesLoading,
    load: loadCallingCodes,
  } = useCallingCodes();

  const ACCOUNT_PATTERN = /^[A-Za-z0-9_-]{4,20}$/;
  /** ASCII 可见字符，不含空格 */
  const PASSWORD_PATTERN = /^[\x21-\x7E]{8,32}$/;
  const PHONE_PATTERN = /^\d{5,20}$/;
  const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

  const STATUSES: AccountStatus[] = ['normal', 'frozen', 'disabled'];

  function createForm() {
    return {
      account: '',
      password: '',
      nickname: '',
      phoneRegion: '',
      phone: '',
      email: '',
      agentId: undefined as number | undefined,
      vipLevelId: undefined as number | undefined,
      autoLevelId: undefined as number | undefined,
      fixedLevelIds: [] as number[],
      status: 'normal' as AccountStatus,
      betBanned: false,
      remark: '',
    };
  }

  const formRef = ref<FormInstance>();
  const form = reactive(createForm());

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

  /** 默认 VIP0；等级列表加载完成前不覆盖用户已选的值 */
  function applyDefaultVip() {
    if (form.vipLevelId != null) return;
    form.vipLevelId = vipLevels.value.find((item) => item.level === 0)?.id;
  }
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

  function agentLevelText(agent: AgentListItem) {
    return agent.level_alias || String(agent.level || '--');
  }

  const agentColumns = computed(() => {
    const agent = selectedAgent.value;
    if (!agent) return [];
    return [
      [
        { label: t('memberManage.create.agentAccount'), value: agent.account },
        { label: t('memberManage.create.agentCode'), value: String(agent.id) },
      ],
      [
        {
          label: t('memberManage.create.agentLevel'),
          value: agentLevelText(agent),
        },
        {
          label: t('memberManage.create.agentParent'),
          value: agent.parent_account || t('memberManage.create.noParent'),
        },
      ],
    ];
  });

  const rules = computed<Record<string, FieldRule[]>>(() => ({
    account: [
      { required: true, message: t('memberManage.create.accountRequired') },
      {
        validator: (value: string, callback: (error?: string) => void) => {
          if (!ACCOUNT_PATTERN.test((value || '').trim())) {
            callback(t('memberManage.create.accountInvalid'));
            return;
          }
          callback();
        },
      },
    ],
    password: [
      { required: true, message: t('memberManage.create.passwordRequired') },
      {
        match: PASSWORD_PATTERN,
        message: t('memberManage.create.passwordInvalid'),
      },
    ],
    phone: [
      {
        validator: (value: string, callback: (error?: string) => void) => {
          const phone = (value || '').trim();
          if (!phone) {
            callback();
            return;
          }
          if (!PHONE_PATTERN.test(phone)) {
            callback(t('memberManage.create.phoneInvalid'));
            return;
          }
          if (!regionCode(form.phoneRegion)) {
            callback(t('memberManage.create.callingCodeRequired'));
            return;
          }
          callback();
        },
      },
    ],
    email: [
      {
        validator: (value: string, callback: (error?: string) => void) => {
          const email = (value || '').trim();
          if (email && !EMAIL_PATTERN.test(email)) {
            callback(t('memberManage.create.emailInvalid'));
            return;
          }
          callback();
        },
      },
    ],
    agentId: [
      { required: true, message: t('memberManage.create.agentRequired') },
    ],
  }));

  watch(
    () => props.visible,
    async (open) => {
      if (!open) return;
      Object.assign(form, createForm());
      selectedAgent.value = null;
      searchAgents();
      loadLevels();
      loadCallingCodes();
      await loadVipLevels();
      applyDefaultVip();
    }
  );

  function onAgentChange(value: unknown) {
    selectedAgent.value =
      agentOptions.value.find((item) => item.id === value) || null;
  }

  function formatAgentLabel(data: SelectOptionData) {
    const agent = selectedAgent.value;
    if (!agent || agent.id !== data.value) return String(data.label ?? '');
    return `${agent.account} / ${agent.id}`;
  }

  function close() {
    emit('update:visible', false);
  }

  async function handleBeforeOk() {
    const errors = await formRef.value?.validate();
    const agent = selectedAgent.value;
    if (errors || !agent) return false;
    try {
      const result = await createMember({
        account: form.account.trim(),
        password: form.password,
        nickname: form.nickname.trim() || undefined,
        agent_id: agent.id,
        remark: form.remark.trim() || undefined,
      });
      emit('created', result);
      close();
      return true;
    } catch {
      return false;
    }
  }
</script>

<style scoped lang="less">
  .create-title-main {
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 18px;
    line-height: 26px;
  }

  .create-title-sub {
    margin-top: 2px;
    color: var(--color-text-3);
    font-weight: 400;
    font-size: 13px;
    line-height: 20px;
  }

  .create-section {
    padding: 18px 20px 4px;
    background: var(--color-bg-2);
    border: 1px solid var(--color-neutral-2);
    border-radius: 10px;

    & + & {
      margin-top: 12px;
    }
  }

  .section-head {
    display: flex;
    gap: 10px;
    align-items: center;
    margin-bottom: 14px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 15px;
  }

  .section-index {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    color: #fff;
    font-weight: 600;
    font-size: 13px;
    background: rgb(var(--primary-6));
    border-radius: 50%;
  }

  .create-form :deep(.arco-form-item-label) {
    color: var(--color-text-1);
    font-weight: 500;
  }

  .create-form :deep(.arco-form-item-extra) {
    font-size: 12px;
  }

  .agent-item {
    margin-bottom: 12px;
  }

  .phone-group {
    display: flex;
    width: 100%;

    .phone-code {
      flex: 0 0 110px;
    }
  }

  .agent-option {
    display: grid;
    grid-template-columns: 1.2fr 1fr 1fr 0.9fr;
    gap: 12px;
    width: 100%;
  }

  .agent-option-line {
    color: rgb(var(--primary-6));
  }

  .agent-empty {
    display: flex;
    gap: 10px;
    margin-bottom: 18px;
    padding: 14px 16px;
    background: rgb(var(--primary-1));
    border-radius: 8px;
  }

  .agent-empty-icon {
    flex: none;
    margin-top: 2px;
    color: rgb(var(--primary-6));
    font-size: 16px;
  }

  .agent-empty-title {
    color: var(--color-text-1);
    font-size: 13px;
    line-height: 20px;
  }

  .agent-empty-desc {
    margin-top: 2px;
    color: var(--color-text-3);
    font-size: 12px;
    line-height: 18px;
  }

  .agent-card {
    padding: 14px 16px 6px;
    background: rgb(var(--primary-1));
    border: 1px solid rgb(var(--primary-2));
    border-radius: 8px;
  }

  .agent-card-head {
    display: flex;
    gap: 10px;
    align-items: center;
    margin-bottom: 10px;
    color: var(--color-text-1);
    font-weight: 500;
    font-size: 13px;
  }

  .agent-card-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 26px;
    height: 26px;
    color: rgb(var(--primary-6));
    font-size: 15px;
    background: var(--color-bg-2);
    border-radius: 50%;
  }

  .agent-card-body {
    display: grid;
    grid-template-columns: 1fr 1fr;
    padding-left: 36px;
  }

  .agent-card-col + .agent-card-col {
    padding-left: 24px;
    border-left: 1px solid rgb(var(--primary-2));
  }

  .agent-field {
    display: flex;
    gap: 16px;
    margin-bottom: 8px;
    font-size: 13px;
    line-height: 20px;
  }

  .agent-field-label {
    flex: none;
    width: 64px;
    color: var(--color-text-3);
  }

  .agent-field-value {
    color: var(--color-text-1);
    font-weight: 500;
  }

  .field-extra {
    margin: 8px 0 18px;
    color: var(--color-text-3);
    font-size: 12px;
    line-height: 18px;
  }
</style>

<style lang="less">
  .member-create-modal {
    .arco-modal-header {
      height: auto;
      padding: 18px 24px 14px;
      border-bottom: none;
    }

    .arco-modal-body {
      max-height: 64vh;
      padding: 0 24px 16px;
      overflow: auto;
    }

    .arco-modal-footer {
      padding: 14px 24px;
    }
  }

  .member-agent-dropdown .arco-select-option-content {
    flex: 1;
    min-width: 0;
  }
</style>
