<template>
  <a-spin :loading="loading" class="edit-panel">
    <div class="edit-hero">
      <div class="hero-main">
        <span class="hero-orb" :style="{ background: color }">
          {{ record.level || '--' }}
        </span>
        <div class="hero-info">
          <div class="hero-name">
            <strong>{{ record.account }}</strong>
            <a-tag size="small" :color="STATUS_STYLE[statusKey].tag">
              {{ t(`agentList.edit.status.${statusKey}`) }}
            </a-tag>
            <a-tag v-if="agentTypeText" size="small" color="arcoblue">
              {{ agentTypeText }}
            </a-tag>
          </div>
          <div class="hero-meta">
            <span>ID：{{ record.id }}</span>
            <a-divider direction="vertical" />
            <span>
              {{ t('agentList.edit.field.parent') }}：{{
                displayText(parentAccount)
              }}
            </span>
            <a-divider direction="vertical" />
            <span>
              {{ t('agentList.edit.field.openedAt') }}：{{
                displayText(openedAt)
              }}
            </span>
          </div>
        </div>
      </div>
      <div class="hero-stats">
        <div v-for="item in heroStats" :key="item.label" class="hero-stat">
          <span>{{ item.label }}</span>
          <strong>{{ item.value }}</strong>
        </div>
      </div>
    </div>

    <a-tabs v-model:active-key="activeTab" class="edit-tabs" lazy-load>
      <a-tab-pane key="basic">
        <template #title>
          <icon-user />
          {{ t('agentList.edit.tab.basic') }}
        </template>
        <AgentInfoGrid
          :title="t('agentList.edit.section.account')"
          :items="accountItems"
          @action="openEditor"
        >
          <template #status>
            <a-badge
              :status="STATUS_STYLE[statusKey].badge"
              :text="t(`agentList.edit.status.${statusKey}`)"
            />
          </template>
          <template #password>
            <span class="secret">••••••</span>
          </template>
          <template #withdrawPassword>
            <span class="secret">••••••</span>
          </template>
        </AgentInfoGrid>

        <AgentInfoGrid
          :title="t('agentList.edit.section.promo')"
          :items="promoItems"
          @action="openEditor"
        >
          <template #promoCode>
            <template v-if="promoCode">
              <span class="code-text">{{ promoCode }}</span>
              <a-button
                class="copy-btn"
                type="text"
                size="mini"
                @click="copy(promoCode)"
              >
                <template #icon><icon-copy /></template>
              </a-button>
            </template>
            <span v-else class="muted">--</span>
          </template>
          <template #domain>
            <template v-if="promoLink">
              <a-link :href="promoLink" target="_blank" class="domain-link">
                {{ promoLink }}
              </a-link>
              <a-button
                class="copy-btn"
                type="text"
                size="mini"
                @click="copy(promoLink)"
              >
                <template #icon><icon-copy /></template>
              </a-button>
            </template>
            <span v-else class="muted">--</span>
          </template>
          <template #wallet>
            <div class="wallet-pills">
              <span class="wallet-pill">
                <em>{{ t('agentList.edit.walletBalance') }}</em>
                {{ formatAmount(record.balance) }}
              </span>
              <span class="wallet-pill frozen">
                <em>{{ t('agentList.edit.walletFrozen') }}</em>
                {{ formatAmount(0) }}
              </span>
            </div>
          </template>
          <template #share>
            <span class="share-value">{{ shareText }}</span>
          </template>
          <template #withdrawLimit>
            <a-select
              class="withdraw-select"
              size="small"
              :model-value="payoutValue"
              :loading="payoutSaving"
              :disabled="payoutSaving"
              :placeholder="t('agentList.search.all')"
              @change="onPayoutChange"
            >
              <a-option
                v-for="value in PAYOUT_OPTIONS"
                :key="value"
                :value="value"
              >
                {{ t(`agentList.edit.payout.${value}`) }}
              </a-option>
            </a-select>
          </template>
        </AgentInfoGrid>

        <AgentInfoGrid
          :title="t('agentList.edit.section.login')"
          :items="loginItems"
        />
      </a-tab-pane>

      <a-tab-pane key="contact">
        <template #title>
          <icon-phone />
          {{ t('agentList.edit.tab.contact') }}
        </template>
        <section class="profile-section">
          <div class="avatar-card">
            <a-upload
              accept="image/*"
              :auto-upload="false"
              :show-file-list="false"
              :disabled="avatarUploading"
              @change="onAvatarChange"
            >
              <template #upload-button>
                <a-spin :loading="avatarUploading" class="avatar-trigger">
                  <img v-if="contact.avatar" :src="contact.avatar" alt="" />
                  <span v-else class="avatar-initial">
                    {{ avatarInitial }}
                  </span>
                  <div class="avatar-mask">
                    <icon-camera />
                    <span>
                      {{
                        contact.avatar
                          ? t('agentList.edit.action.change')
                          : t('agentList.edit.action.upload')
                      }}
                    </span>
                  </div>
                </a-spin>
              </template>
            </a-upload>
            <strong>{{ contact.nickname || record.account }}</strong>
            <span>{{ t('agentList.edit.field.avatar') }}</span>
          </div>
          <AgentInfoGrid
            class="profile-grid"
            :title="t('agentList.edit.section.profile')"
            :items="profileItems"
            @action="openEditor"
          />
        </section>

        <AgentInfoGrid
          :title="t('agentList.edit.section.social')"
          :items="socialItems"
          @action="openEditor"
        >
          <template
            v-for="item in SOCIAL_CHANNELS"
            :key="item.key"
            #[`label-${item.key}`]
          >
            <span class="social-label">
              <span class="social-badge" :style="{ background: item.color }">
                {{ item.label.charAt(0) }}
              </span>
              {{ item.label }}
            </span>
          </template>
        </AgentInfoGrid>
      </a-tab-pane>

      <a-tab-pane key="finance">
        <template #title>
          <icon-storage />
          {{ t('agentList.edit.tab.finance') }}
        </template>
        <AgentFinanceTable />
      </a-tab-pane>

      <a-tab-pane key="wallet">
        <template #title>
          <icon-list />
          {{ t('agentList.edit.tab.wallet') }}
        </template>
        <AgentRecordTable kind="wallet" />
      </a-tab-pane>

      <a-tab-pane key="log">
        <template #title>
          <icon-history />
          {{ t('agentList.edit.tab.log') }}
        </template>
        <AgentRecordTable kind="log" :agent-id="record.id" />
      </a-tab-pane>
    </a-tabs>

    <AgentFieldEditor
      v-model:visible="editor.visible"
      :kind="editor.kind"
      :label="editor.label"
      :value="editor.value"
      :max="shareMax"
      :before-save="submitField"
      @save="onEditorSave"
    />

    <a-modal
      v-model:visible="resetResult.visible"
      :title="t('agentList.edit.reset.resultTitle')"
      :width="440"
      :mask-closable="false"
      :closable="false"
      hide-cancel
      unmount-on-close
      modal-class="mac-modal-panel"
      :ok-text="t('agentList.edit.reset.done')"
      @close="resetResult.password = ''"
    >
      <div class="reset-result">
        <a-alert type="warning" :show-icon="true">
          {{ t('agentList.edit.reset.resultTip') }}
        </a-alert>
        <div class="reset-label">
          {{
            t('agentList.edit.reset.newPassword', { account: record.account })
          }}
        </div>
        <div class="reset-password">
          <span>{{ resetResult.password }}</span>
          <a-button
            type="primary"
            size="small"
            @click="copy(resetResult.password)"
          >
            <template #icon><icon-copy /></template>
            {{ t('agentList.edit.reset.copy') }}
          </a-button>
        </div>
      </div>
    </a-modal>
  </a-spin>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, Modal, type FileItem } from '@arco-design/web-vue';
  import {
    AGENT_PROFILE_FIELD,
    AGENT_PROFILE_PAYOUT,
    AGENT_PROFILE_STATUS,
    resetAgentPassword,
    resetAgentWithdrawPassword,
    updateAgentPayoutRestricted,
    updateAgentProfile,
    updateAgentRemark,
    updateAgentShareRate,
    type AgentProfile,
    type AgentProfileFields,
  } from '@/api/agent';
  import { uploadOssFile } from '@/api/oss';
  import AgentFieldEditor from './agent-field-editor.vue';
  import AgentFinanceTable from './agent-finance-table.vue';
  import AgentInfoGrid from './agent-info-grid.vue';
  import AgentRecordTable from './agent-record-table.vue';
  import { displayText, formatAmount, formatRatio } from '../format';
  import { formatUnix } from '../mapper';
  import { buildPromoLink, levelColor } from '../options';
  import type {
    AgentFormValue,
    AgentRecord,
    FieldEditorKind,
    InfoItem,
  } from '../types';

  type StatusKey = 'normal' | 'frozen' | 'disabled';

  const STATUS_STYLE: Record<
    StatusKey,
    { tag: string; badge: 'success' | 'warning' | 'danger' }
  > = {
    normal: { tag: 'green', badge: 'success' },
    frozen: { tag: 'orange', badge: 'warning' },
    disabled: { tag: 'red', badge: 'danger' },
  };

  const PAYOUT_OPTIONS: number[] = Object.values(AGENT_PROFILE_PAYOUT);
  const REGISTER_SOURCE_KEYS = [1, 2, 3, 4];
  const AGENT_TYPE_CASH = 1;
  const AGENT_TYPE_CREDIT = 2;

  type ProfileKey =
    | 'nickname'
    | 'realName'
    | 'idNumber'
    | 'phone'
    | 'email'
    | 'telegram'
    | 'zalo'
    | 'facebook'
    | 'twitter'
    | 'whatsapp'
    | 'line';

  const SOCIAL_CHANNELS: { key: ProfileKey; label: string; color: string }[] = [
    { key: 'telegram', label: 'Telegram', color: '#229ed9' },
    { key: 'zalo', label: 'Zalo', color: '#0068ff' },
    { key: 'facebook', label: 'Facebook', color: '#1877f2' },
    { key: 'twitter', label: 'Twitter', color: '#1d9bf0' },
    { key: 'whatsapp', label: 'Whatsapp', color: '#25d366' },
    { key: 'line', label: 'Line', color: '#06c755' },
  ];

  const PROFILE_UPDATE_FIELDS: Record<
    ProfileKey,
    { type: number; name: keyof AgentProfileFields }
  > = {
    nickname: { type: AGENT_PROFILE_FIELD.NICKNAME, name: 'nickname' },
    realName: { type: AGENT_PROFILE_FIELD.REALNAME, name: 'realname' },
    idNumber: {
      type: AGENT_PROFILE_FIELD.ID_DOCUMENT_NO,
      name: 'id_document_no',
    },
    phone: { type: AGENT_PROFILE_FIELD.MOBILE, name: 'mobile' },
    email: { type: AGENT_PROFILE_FIELD.EMAIL, name: 'email' },
    telegram: { type: AGENT_PROFILE_FIELD.TELEGRAM, name: 'telegram' },
    zalo: { type: AGENT_PROFILE_FIELD.ZALO, name: 'zalo' },
    facebook: { type: AGENT_PROFILE_FIELD.FACEBOOK, name: 'facebook' },
    twitter: { type: AGENT_PROFILE_FIELD.TWITTER, name: 'twitter' },
    whatsapp: { type: AGENT_PROFILE_FIELD.WHATSAPP, name: 'whatsapp' },
    line: { type: AGENT_PROFILE_FIELD.LINE, name: 'line' },
  };

  const PROFILE_FIELDS: ProfileKey[] = [
    'nickname',
    'realName',
    'idNumber',
    'phone',
    'email',
  ];

  const props = defineProps<{
    record: AgentRecord;
    lineCode: string;
    shareMax: number;
    /** GET /agent/profile 的结果，加载中或失败时为 null，此时回退到列表数据 */
    profile: AgentProfile | null;
    loading: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'update', value: Partial<AgentFormValue>): void;
    (e: 'payout', value: number): void;
  }>();

  const { t } = useI18n();

  const activeTab = ref('basic');

  function createContact() {
    return {
      avatar: '',
      nickname: '',
      realName: '',
      idNumber: '',
      phone: '',
      email: '',
      telegram: '',
      zalo: '',
      facebook: '',
      twitter: '',
      whatsapp: '',
      line: '',
    } as Record<ProfileKey | 'avatar', string>;
  }

  const contact = reactive(createContact());

  watch(
    () => props.record.id,
    () => {
      activeTab.value = 'basic';
    }
  );

  watch(
    () => props.profile,
    (data) => {
      Object.assign(contact, createContact());
      if (!data) return;
      Object.assign(contact, {
        avatar: data.avatar_url || '',
        nickname: data.nickname || '',
        realName: data.realname || '',
        idNumber: data.id_document_no || '',
        phone: data.mobile || '',
        email: data.email || '',
        telegram: data.telegram || '',
        zalo: data.zalo || '',
        facebook: data.facebook || '',
        twitter: data.twitter || '',
        whatsapp: data.whatsapp || '',
        line: data.line || '',
      });
    },
    { immediate: true }
  );

  const color = computed(() =>
    levelColor(props.record.level, props.record.layer)
  );

  const statusKey = computed<StatusKey>(() => {
    const status = props.profile?.status;
    if (status === AGENT_PROFILE_STATUS.NORMAL) return 'normal';
    if (status === AGENT_PROFILE_STATUS.FROZEN) return 'frozen';
    if (status === AGENT_PROFILE_STATUS.DISABLED) return 'disabled';
    return props.record.status;
  });

  const agentTypeText = computed(() => {
    const type = props.profile?.agent_type;
    if (type === AGENT_TYPE_CASH) return t('agentList.edit.agentType.cash');
    if (type === AGENT_TYPE_CREDIT) return t('agentList.edit.agentType.credit');
    const line = (props.record.lineCode || props.lineCode || '').toUpperCase();
    if (line.startsWith('CREDIT')) return t('agentList.edit.agentType.credit');
    if (line.startsWith('CASH')) return t('agentList.edit.agentType.cash');
    return '';
  });

  const sourceText = computed(() => {
    const source = props.profile?.register_source;
    if (source == null) return props.record.registerSource;
    return REGISTER_SOURCE_KEYS.includes(source)
      ? t(`agentList.edit.source.${source}`)
      : '';
  });

  const regionText = computed(() =>
    [props.profile?.country, props.profile?.currency]
      .map((v) => (v || '').trim())
      .filter(Boolean)
      .join(' / ')
  );

  const parentAccount = computed(
    () => props.profile?.parent_account ?? props.record.parentAgent
  );

  const openedAt = computed(() =>
    props.profile
      ? formatUnix(props.profile.register_at)
      : props.record.registeredAt
  );

  const remark = computed(() => props.profile?.remark ?? props.record.remark);

  const shareRatio = computed(
    () => props.profile?.share_rate ?? props.record.shareRatio ?? 0
  );

  const rawPromoCode = computed(() =>
    (props.profile?.promo_code ?? props.record.promoCode).trim()
  );

  const promoLink = computed(() => {
    const link = buildPromoLink(props.record.promoHost, rawPromoCode.value);
    return /^https?:\/\//i.test(link) ? link : '';
  });

  const promoCode = computed(() =>
    /^https?:\/\//i.test(rawPromoCode.value) ? '' : rawPromoCode.value
  );

  const shareText = computed(() => `${Number(shareRatio.value)}%`);

  const payoutValue = computed(() => {
    const value = props.profile?.payout_restricted;
    if (value && PAYOUT_OPTIONS.includes(value)) return value;
    if (props.record.withdrawLimit === 'none') return AGENT_PROFILE_PAYOUT.NONE;
    if (props.record.withdrawLimit === 'forbidden') {
      return AGENT_PROFILE_PAYOUT.SELF;
    }
    return undefined;
  });

  const avatarInitial = computed(() =>
    (contact.nickname || props.record.account || '?').charAt(0).toUpperCase()
  );

  const heroStats = computed(() => [
    {
      label: t('agentList.card.balance'),
      value: formatAmount(props.record.balance),
    },
    {
      label: t('agentList.edit.field.share'),
      value: formatRatio(shareRatio.value),
    },
    {
      label: t('agentList.card.directMembers'),
      value: props.record.directMembers,
    },
    {
      label: t('agentList.card.directAgents'),
      value: props.record.directAgents,
    },
    {
      label: t('agentList.card.commission'),
      value: formatAmount(props.record.totalCommission),
    },
  ]);

  const editText = computed(() => t('agentList.edit.action.edit'));

  const accountItems = computed<InfoItem[]>(() => [
    { key: 'status', label: t('agentList.edit.field.status') },
    {
      key: 'agentType',
      label: t('agentList.edit.field.agentType'),
      value: agentTypeText.value,
    },
    {
      key: 'region',
      label: t('agentList.edit.field.region'),
      value: regionText.value,
    },
    {
      key: 'source',
      label: t('agentList.edit.field.source'),
      value: sourceText.value,
    },
    {
      key: 'parent',
      label: t('agentList.edit.field.parent'),
      value: parentAccount.value,
    },
    {
      key: 'openedAt',
      label: t('agentList.edit.field.openedAt'),
      value: openedAt.value,
    },
    {
      key: 'password',
      label: t('agentList.edit.field.password'),
      action: t('agentList.edit.action.resetPassword'),
    },
    {
      key: 'withdrawPassword',
      label: t('agentList.edit.field.withdrawPassword'),
      action: editText.value,
    },
    {
      key: 'remark',
      label: t('agentList.edit.field.remark'),
      value: remark.value,
      action: editText.value,
      full: true,
    },
  ]);

  const promoItems = computed<InfoItem[]>(() => [
    { key: 'promoCode', label: t('agentList.edit.field.promoCode') },
    { key: 'domain', label: t('agentList.edit.field.domain') },
    { key: 'wallet', label: t('agentList.edit.field.wallet') },
    {
      key: 'invited',
      label: t('agentList.edit.field.invited'),
      value: String(
        props.profile?.invite_member_count ?? props.record.directMembers ?? 0
      ),
    },
    {
      key: 'share',
      label: t('agentList.edit.field.share'),
      action: editText.value,
    },
    { key: 'withdrawLimit', label: t('agentList.edit.field.withdrawLimit') },
  ]);

  const loginItems = computed<InfoItem[]>(() => {
    const data = props.profile;
    return [
      {
        key: 'loginCount',
        label: t('agentList.edit.field.loginCount'),
        value: data
          ? t('agentList.edit.loginTimes', { n: data.login_count || 0 })
          : '',
      },
      {
        key: 'lastLoginAt',
        label: t('agentList.edit.field.lastLoginAt'),
        value: data ? formatUnix(data.last_login_at) : props.record.lastLoginAt,
      },
      {
        key: 'lastLoginIp',
        label: t('agentList.edit.field.lastLoginIp'),
        value: data?.last_login_ip,
      },
      {
        key: 'lastDevice',
        label: t('agentList.edit.field.lastDevice'),
        value: data?.last_login_udid,
      },
    ];
  });

  const profileItems = computed<InfoItem[]>(() =>
    PROFILE_FIELDS.map((key) => ({
      key,
      label: t(`agentList.edit.field.${key}`),
      value: key === 'phone' ? contact.phone.replace('-', ' - ') : contact[key],
      action: editText.value,
    }))
  );

  const socialItems = computed<InfoItem[]>(() =>
    SOCIAL_CHANNELS.map((item) => ({
      key: item.key,
      label: item.label,
      value: contact[item.key],
      action: editText.value,
    }))
  );

  const editor = reactive({
    visible: false,
    key: '',
    kind: 'text' as FieldEditorKind,
    label: '',
    value: '' as string | number,
  });

  const resetResult = reactive({ visible: false, password: '' });

  function fieldLabel(key: string) {
    const channel = SOCIAL_CHANNELS.find((item) => item.key === key);
    return channel ? channel.label : t(`agentList.edit.field.${key}`);
  }

  function editorKind(key: string): FieldEditorKind {
    if (key === 'withdrawPassword') return 'withdrawPassword';
    if (key === 'share') return 'share';
    if (key === 'remark') return 'textarea';
    if (key === 'phone') return 'phone';
    return 'text';
  }

  function editorValue(key: string) {
    if (key === 'share') return Math.round(shareRatio.value);
    if (key === 'remark') return remark.value;
    if (key in contact) return contact[key as ProfileKey];
    return '';
  }

  function confirmResetPassword() {
    const agentId = props.record.id;
    Modal.warning({
      title: t('agentList.edit.reset.confirmTitle'),
      content: t('agentList.edit.reset.confirmContent', {
        account: props.record.account,
      }),
      hideCancel: false,
      okText: t('agentList.confirm.ok'),
      cancelText: t('agentList.confirm.cancel'),
      modalClass: 'mac-modal-panel',
      onBeforeOk: async () => {
        try {
          const { data } = await resetAgentPassword(agentId);
          if (!data?.password) return false;
          resetResult.password = data.password;
          resetResult.visible = true;
          return true;
        } catch {
          return false;
        }
      },
    });
  }

  function openEditor(key: string) {
    if (key === 'password') {
      confirmResetPassword();
      return;
    }
    editor.key = key;
    editor.kind = editorKind(key);
    editor.label = fieldLabel(key);
    editor.value = editorValue(key);
    editor.visible = true;
  }

  async function submitField(value: string | number) {
    if (editor.key === 'share') {
      await updateAgentShareRate(props.record.id, Math.round(Number(value)));
    }
    if (editor.key === 'remark') {
      await updateAgentRemark(props.record.id, String(value));
    }
    if (editor.key === 'withdrawPassword') {
      await resetAgentWithdrawPassword(props.record.id, String(value));
    }
    const profileField = PROFILE_UPDATE_FIELDS[editor.key as ProfileKey];
    if (profileField) {
      await updateAgentProfile({
        agent_id: props.record.id,
        field_type: profileField.type,
        [profileField.name]: String(value),
      });
    }
    return true;
  }

  function onEditorSave(value: string | number) {
    const { key } = editor;
    if (key === 'share') {
      emit('update', { shareRatio: Number(value) });
      return;
    }
    if (key === 'remark') {
      emit('update', { remark: String(value) });
      return;
    }
    if (key in contact) contact[key as ProfileKey] = String(value);
    Message.success(t('agentList.message.updated'));
  }

  const payoutSaving = ref(false);

  async function onPayoutChange(value: unknown) {
    const next = Number(value);
    if (!next || next === payoutValue.value || payoutSaving.value) return;
    payoutSaving.value = true;
    try {
      await updateAgentPayoutRestricted(props.record.id, next);
      emit('payout', next);
    } catch {
      // 失败时 payoutValue 未变，下拉框会回到原值
    } finally {
      payoutSaving.value = false;
    }
  }

  const avatarUploading = ref(false);

  async function onAvatarChange(_: FileItem[], fileItem: FileItem) {
    if (!fileItem.file || avatarUploading.value) return;
    const agentId = props.record.id;
    avatarUploading.value = true;
    try {
      const form = new FormData();
      form.append('file', fileItem.file);
      form.append('scene', 'public');
      const { data: file } = await uploadOssFile(form);
      await updateAgentProfile({
        agent_id: agentId,
        field_type: AGENT_PROFILE_FIELD.AVATAR_URL,
        avatar_url: file.url,
      });
      if (props.record.id !== agentId) return;
      contact.avatar = file.url;
      Message.success(t('agentList.message.updated'));
    } catch {
      // 错误提示由请求拦截器统一弹出，头像保持原样
    } finally {
      avatarUploading.value = false;
    }
  }

  async function copy(text: string) {
    try {
      await navigator.clipboard.writeText(text);
      Message.success(t('agentList.edit.copied'));
    } catch {
      Message.warning(t('agentList.message.copyFail'));
    }
  }
</script>

<style scoped lang="less">
  .edit-panel {
    display: block;
    width: 100%;
  }

  .reset-result {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .reset-label {
    color: var(--color-text-2);
    font-size: 13px;
  }

  .reset-password {
    display: flex;
    gap: 12px;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    background: var(--color-fill-1);
    border: 1px dashed var(--color-neutral-4);
    border-radius: 8px;

    span {
      color: var(--color-text-1);
      font-weight: 600;
      font-size: 18px;
      font-family: Consolas, Menlo, monospace;
      letter-spacing: 1px;
      word-break: break-all;
      user-select: all;
    }
  }

  .edit-hero {
    display: flex;
    flex-wrap: wrap;
    gap: 16px 24px;
    align-items: center;
    justify-content: space-between;
    padding: 18px 20px;
    background: linear-gradient(
      120deg,
      rgb(var(--primary-1)) 0%,
      var(--color-bg-2) 70%
    );
    border: 1px solid var(--color-neutral-3);
    border-radius: 10px;
  }

  .hero-main {
    display: flex;
    gap: 14px;
    align-items: center;
    min-width: 0;
  }

  .hero-orb {
    display: inline-flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    width: 52px;
    height: 52px;
    color: #fff;
    font-weight: 700;
    font-size: 16px;
    border-radius: 50%;
    box-shadow: 0 6px 16px rgba(0, 0, 0, 0.12);
  }

  .hero-info {
    min-width: 0;
  }

  .hero-name {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;

    strong {
      color: var(--color-text-1);
      font-size: 18px;
    }
  }

  .hero-meta {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    margin-top: 6px;
    color: var(--color-text-3);
    font-size: 12px;
  }

  .hero-stats {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .hero-stat {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 96px;
    padding: 8px 14px;
    background: var(--color-bg-2);
    border: 1px solid var(--color-neutral-3);
    border-radius: 8px;

    span {
      color: var(--color-text-3);
      font-size: 12px;
    }

    strong {
      color: var(--color-text-1);
      font-size: 16px;
      font-variant-numeric: tabular-nums;
    }
  }

  .edit-tabs {
    margin-top: 12px;

    :deep(.arco-tabs-nav) {
      position: sticky;
      top: 0;
      z-index: 2;
      background: var(--color-bg-2);
    }

    :deep(.arco-tabs-tab-title) {
      display: inline-flex;
      gap: 6px;
      align-items: center;
    }

    :deep(.arco-tabs-content) {
      min-height: 420px;
      padding-top: 16px;
    }
  }

  .secret {
    color: var(--color-text-2);
    letter-spacing: 2px;
  }

  .muted {
    color: var(--color-text-4);
  }

  .code-text {
    padding: 1px 8px;
    font-family: Consolas, Menlo, monospace;
    background: var(--color-fill-2);
    border-radius: 4px;
  }

  .domain-link {
    min-width: 0;
    padding: 0;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .copy-btn {
    flex-shrink: 0;
    color: var(--color-text-3);

    &:hover {
      color: rgb(var(--primary-6));
    }
  }

  .wallet-pills {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .wallet-pill {
    display: inline-flex;
    gap: 6px;
    align-items: center;
    padding: 2px 10px;
    color: rgb(var(--success-6));
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    background: rgb(var(--success-1));
    border-radius: 12px;

    em {
      color: var(--color-text-3);
      font-weight: 400;
      font-style: normal;
      font-size: 12px;
    }

    &.frozen {
      color: rgb(var(--warning-6));
      background: rgb(var(--warning-1));
    }
  }

  .share-value {
    color: rgb(var(--primary-6));
    font-weight: 600;
  }

  .withdraw-select {
    width: 180px;
  }

  .profile-section {
    display: flex;
    gap: 16px;
    align-items: stretch;
    margin-bottom: 20px;
  }

  .avatar-card {
    display: flex;
    flex-direction: column;
    flex-shrink: 0;
    gap: 6px;
    align-items: center;
    justify-content: center;
    width: 200px;
    padding: 20px 16px;
    background: var(--color-fill-1);
    border: 1px solid var(--color-neutral-3);
    border-radius: 8px;

    strong {
      max-width: 100%;
      margin-top: 6px;
      overflow: hidden;
      color: var(--color-text-1);
      white-space: nowrap;
      text-overflow: ellipsis;
    }

    > span {
      color: var(--color-text-3);
      font-size: 12px;
    }
  }

  .avatar-trigger {
    position: relative;
    width: 88px;
    height: 88px;
    overflow: hidden;
    background: rgb(var(--primary-6));
    border: 3px solid var(--color-bg-2);
    border-radius: 50%;
    box-shadow: 0 4px 14px rgba(0, 0, 0, 0.12);
    cursor: pointer;

    img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }

    &:hover .avatar-mask {
      opacity: 1;
    }
  }

  .avatar-initial {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: #fff;
    font-weight: 700;
    font-size: 32px;
  }

  .avatar-mask {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
    align-items: center;
    justify-content: center;
    color: #fff;
    font-size: 12px;
    background: rgba(0, 0, 0, 0.45);
    opacity: 0;
    transition: opacity 0.2s;

    :deep(.arco-icon) {
      font-size: 18px;
    }
  }

  .profile-grid {
    flex: 1;
    min-width: 0;
  }

  .social-label {
    display: inline-flex;
    gap: 8px;
    align-items: center;
  }

  .social-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    height: 20px;
    color: #fff;
    font-weight: 700;
    font-size: 11px;
    border-radius: 6px;
  }
</style>
