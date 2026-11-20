<template>
  <a-modal
    :visible="visible"
    :title="
      mode === 'create'
        ? t('agentList.modal.create')
        : t('agentList.modal.edit')
    "
    :width="mode === 'create' ? 620 : 'min(1120px, 94vw)'"
    :footer="mode === 'create'"
    :align-center="mode === 'create'"
    :top="48"
    :body-style="mode === 'create' ? undefined : EDIT_BODY_STYLE"
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
    <a-form
      v-if="mode === 'create'"
      :model="creating"
      layout="horizontal"
      :label-col-props="{ span: 5 }"
      :wrapper-col-props="{ span: 19 }"
    >
      <a-form-item :label="t('agentList.modal.line')" required>
        <div class="line-select-wrap">
          <a-select
            v-model="creating.lineCode"
            class="line-select"
            allow-search
            allow-clear
            :loading="lineLoading"
            :placeholder="t('agentList.modal.linePlaceholder')"
            @change="onLineChange"
          >
            <a-option
              v-for="item in lineOptions"
              :key="item.line_code"
              :value="item.line_code"
              :label="lineLabel(item.line_code)"
            />
          </a-select>
          <div v-if="!lineLoading && !lineOptions.length" class="empty-hint">
            {{ t('agentList.modal.parentEmpty') }}
          </div>
        </div>
      </a-form-item>
      <a-form-item
        v-if="creating.lineCode"
        :label="t('agentList.modal.parentBelong')"
        required
      >
        <a-select
          v-model="creating.parentId"
          class="parent-select"
          allow-search
          allow-clear
          :filter-option="false"
          :loading="parentSearchLoading"
          :placeholder="t('agentList.modal.parentSearchPlaceholder')"
          :options="parentSelectOptions"
          @search="onParentSearch"
        />
      </a-form-item>
      <a-form-item :label="t('agentList.card.account')" required>
        <a-input
          v-model="creating.account"
          :placeholder="t('agentList.modal.accountPlaceholder')"
          allow-clear
        />
      </a-form-item>
      <a-form-item :label="t('agentList.modal.password')" required>
        <a-input-password
          v-model="creating.password"
          :placeholder="t('agentList.modal.passwordPlaceholder')"
          autocomplete="new-password"
          allow-clear
        />
      </a-form-item>
      <a-form-item :label="t('agentList.modal.confirmPassword')" required>
        <a-input-password
          v-model="creating.confirmPassword"
          :placeholder="t('agentList.modal.confirmPasswordPlaceholder')"
          allow-clear
        />
      </a-form-item>
      <a-form-item :label="t('agentList.modal.withdrawPassword')" required>
        <a-input-password
          v-model="creating.withdrawPassword"
          :placeholder="t('agentList.modal.withdrawPasswordPlaceholder')"
          :max-length="WITHDRAW_PASSWORD_MAX_LENGTH"
          allow-clear
        />
      </a-form-item>
      <a-form-item :label="t('agentList.modal.agentMode')">
        <span>{{ t('agentList.modal.agentModeShare') }}</span>
      </a-form-item>
      <a-form-item
        :label="t('agentList.card.share')"
        required
        :validate-status="shareExceeded ? 'error' : undefined"
        :help="
          shareExceeded
            ? t('agentList.modal.shareExceedParent', { max: shareMax })
            : undefined
        "
      >
        <a-input-number
          v-model="creating.shareRatio"
          :min="0"
          :precision="0"
          :step="1"
          :placeholder="t('agentList.modal.sharePlaceholder')"
          model-event="input"
          hide-button
        >
          <template #suffix>
            <span class="share-range"> (0%~{{ shareMax }}%) </span>
          </template>
        </a-input-number>
      </a-form-item>
      <a-form-item :label="t('agentList.modal.remark')">
        <a-textarea
          v-model="creating.remark"
          :placeholder="t('agentList.modal.remarkPlaceholder')"
          :auto-size="{ minRows: 4, maxRows: 6 }"
        />
      </a-form-item>
    </a-form>
    <AgentEditPanel
      v-else-if="record"
      :record="record"
      :line-code="lineCode"
      :share-max="settings.maxShareRatio || SHARE_RATIO_MAX"
      :profile="profile"
      :loading="profileLoading"
      @update="onEditUpdate"
      @payout="onPayoutUpdate"
    />
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { useDebounceFn } from '@vueuse/core';
  import {
    AGENT_PROFILE_PAYOUT,
    AGENT_TYPE,
    createAgent,
    getAgentLineDefaults,
    getAgentProfile,
    searchAgents,
    type AgentLineDefault,
    type AgentProfile,
    type AgentSearchItem,
  } from '@/api/agent';
  import AgentEditPanel from './agent-edit-panel.vue';
  import { REGISTER_SOURCES } from '../options';
  import type { AgentFormValue, AgentRecord, AgentSettings } from '../types';

  const DEFAULT_PARENT_LEVEL = 1;
  const SEARCH_LIMIT = 20;
  const WITHDRAW_PASSWORD_MAX_LENGTH = 20;
  const WITHDRAW_PASSWORD_PATTERN = /^[\x21-\x7E]{6,20}$/;
  const SHARE_RATIO_MAX = 100;
  const EDIT_BODY_STYLE = {
    maxHeight: 'calc(100vh - 150px)',
    padding: '16px 24px 24px',
    overflow: 'auto',
  };

  const props = defineProps<{
    visible: boolean;
    mode: 'create' | 'edit';
    record: AgentRecord | null;
    settings: AgentSettings;
    parentId: number;
    parentAccount: string;
    lineCode: string;
    parentLevel: number;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'submit', value: AgentFormValue): void;
    (e: 'created'): void;
  }>();

  const { t } = useI18n();

  function lineLabel(lineCode: string) {
    const key = `agentList.line.${String(lineCode || '').toUpperCase()}`;
    const text = t(key);
    return text === key ? lineCode : text;
  }

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
  const creating = reactive({
    account: '',
    password: '',
    confirmPassword: '',
    withdrawPassword: '',
    shareRatio: SHARE_RATIO_MAX as number | undefined,
    remark: '',
    lineCode: '',
    parentId: undefined as number | undefined,
  });

  const lineOptions = ref<AgentLineDefault[]>([]);
  const lineLoading = ref(false);
  const parentCandidates = ref<AgentSearchItem[]>([]);
  const parentSearchLoading = ref(false);
  let parentSearchSeq = 0;
  /** 所属上级的搜索层级：取树上选中代理的 level，未选中时用默认层级 */
  const parentSearchLevel = computed(() =>
    props.parentLevel > 0 ? props.parentLevel : DEFAULT_PARENT_LEVEL
  );
  /** 所属上级的占成（百分比），新增代理的占成不能超过它 */
  const parentShareRatio = ref<number | null>(null);

  const shareMax = computed(() =>
    parentShareRatio.value == null
      ? SHARE_RATIO_MAX
      : Math.min(SHARE_RATIO_MAX, Math.floor(parentShareRatio.value))
  );

  const shareExceeded = computed(
    () => creating.shareRatio != null && creating.shareRatio > shareMax.value
  );

  function findParentShare(parentId: number | undefined) {
    const hit = parentCandidates.value.find((item) => item.id === parentId);
    return hit ? Number(hit.share_rate) || 0 : null;
  }

  watch(
    () => creating.parentId,
    (parentId) => {
      parentShareRatio.value = findParentShare(parentId);
    }
  );

  /** 搜索结果晚于选中返回时（如线路默认上级），用结果补上占成 */
  watch(parentCandidates, () => {
    const share = findParentShare(creating.parentId);
    if (share != null) parentShareRatio.value = share;
  });

  watch(shareMax, (max) => {
    if (creating.shareRatio != null && creating.shareRatio > max) {
      creating.shareRatio = max;
    }
  });

  function lineAgentLabel(item: AgentLineDefault) {
    return `${Number(item.agent_id) || 0}`;
  }

  function parentOptionLabel(item: AgentSearchItem) {
    const alias = item.level_alias ? ` · ${item.level_alias}` : '';
    return `${item.account}${alias}`;
  }

  const parentSelectOptions = computed(() => {
    const options = parentCandidates.value.map((item) => ({
      value: item.id,
      label: parentOptionLabel(item),
    }));
    const selectedId = creating.parentId;
    if (
      selectedId != null &&
      !options.some((item) => item.value === selectedId)
    ) {
      const label =
        selectedId === props.parentId && props.parentAccount
          ? props.parentAccount
          : String(selectedId);
      options.unshift({ value: selectedId, label });
    }
    return options;
  });

  function lineDefaultAgentId(lineCode: string) {
    const line = lineOptions.value.find((item) => item.line_code === lineCode);
    return Number(line?.agent_id) || 0;
  }

  function clearParentSelect() {
    creating.parentId = undefined;
    parentCandidates.value = [];
  }

  async function loadParentSearch(keyword: string, keepSelected = false) {
    const text = keyword.trim();
    if (!creating.lineCode.trim() || !text) {
      parentCandidates.value = [];
      return;
    }
    parentSearchSeq += 1;
    const seq = parentSearchSeq;
    parentSearchLoading.value = true;
    try {
      const { data } = await searchAgents(
        {
          keyword: text,
          line_code: creating.lineCode.trim(),
          level: parentSearchLevel.value,
          limit: SEARCH_LIMIT,
        },
        true
      );
      if (seq !== parentSearchSeq) return;
      parentCandidates.value = data?.list || [];
      if (
        !keepSelected &&
        creating.parentId &&
        !parentCandidates.value.some((item) => item.id === creating.parentId)
      ) {
        creating.parentId = undefined;
      }
    } catch {
      if (seq !== parentSearchSeq) return;
      parentCandidates.value = [];
    } finally {
      if (seq === parentSearchSeq) parentSearchLoading.value = false;
    }
  }

  const debouncedParentSearch = useDebounceFn((keyword: string) => {
    const text = keyword.trim();
    if (!text) {
      parentCandidates.value = [];
      return;
    }
    loadParentSearch(text);
  }, 300);

  function onParentSearch(keyword: string) {
    debouncedParentSearch(keyword);
  }

  function onLineChange(value: unknown) {
    const lineCode = String(value || '');
    creating.lineCode = lineCode;
    clearParentSelect();
    loadLineDefaultParent(lineCode);
  }

  async function loadLineDefaultParent(lineCode: string) {
    if (!lineCode) return;
    const agentId = lineDefaultAgentId(lineCode);
    creating.parentId = agentId;
    if (agentId) await loadParentSearch(String(agentId), true);
  }

  /** 在树里选中了某个代理时，默认选它所在线路，并以它为所属上级 */
  function presetFocusParent() {
    const lineCode = props.lineCode.trim().toUpperCase();
    if (props.parentId <= 0 || !lineCode) return;
    const line = lineOptions.value.find(
      (item) => item.line_code.toUpperCase() === lineCode
    );
    if (!line) return;
    creating.lineCode = line.line_code;
    creating.parentId = props.parentId;
    loadParentSearch(String(props.parentId), true);
  }

  async function loadLineOptions() {
    lineLoading.value = true;
    try {
      const { data } = await getAgentLineDefaults();
      lineOptions.value = data?.list || [];
      creating.lineCode = '';
      clearParentSelect();
      presetFocusParent();
    } catch {
      lineOptions.value = [];
      creating.lineCode = '';
      clearParentSelect();
    } finally {
      lineLoading.value = false;
    }
  }

  function resetCreating() {
    creating.account = '';
    creating.password = '';
    creating.confirmPassword = '';
    creating.withdrawPassword = '';
    creating.shareRatio = SHARE_RATIO_MAX;
    creating.remark = '';
    creating.lineCode = '';
    creating.parentId = undefined;
    parentCandidates.value = [];
  }

  function fill(record: AgentRecord) {
    form.account = record.account;
    form.level = record.level;
    form.parentAgent = record.parentAgent;
    form.topAgent = record.topAgent;
    form.registerSource = record.registerSource;
    form.shareRatio =
      record.shareRatio == null
        ? record.shareRatio
        : Math.round(record.shareRatio);
    form.withdrawLimit = record.withdrawLimit;
    form.remark = record.remark;
    form.status = record.status;
    form.promoCode = record.promoCode;
  }

  const profile = ref<AgentProfile | null>(null);
  const profileLoading = ref(false);
  let profileSeq = 0;

  async function loadProfile(agentId: number) {
    profileSeq += 1;
    const seq = profileSeq;
    profile.value = null;
    profileLoading.value = true;
    try {
      const { data } = await getAgentProfile(agentId);
      if (seq === profileSeq) profile.value = data || null;
    } catch {
      if (seq === profileSeq) profile.value = null;
    } finally {
      if (seq === profileSeq) profileLoading.value = false;
    }
  }

  watch(
    () => [props.visible, props.mode, props.record?.id] as const,
    ([open]) => {
      if (!open) return;
      if (props.mode === 'edit' && props.record) {
        fill(props.record);
        loadProfile(props.record.id);
      } else {
        Object.assign(form, createForm());
        resetCreating();
        loadLineOptions();
      }
    }
  );

  function warnIfEmpty(value: string, messageKey: string) {
    if (value) return false;
    Message.warning(t(messageKey));
    return true;
  }

  async function submitCreate() {
    if (!creating.lineCode.trim()) {
      Message.warning(t('agentList.modal.linePickRequired'));
      return false;
    }
    if (!creating.parentId) {
      Message.warning(t('agentList.modal.parentRequired'));
      return false;
    }
    const account = creating.account.trim();
    const lineCode = creating.lineCode.trim();
    if (warnIfEmpty(account, 'agentList.modal.accountRequired')) return false;
    if (warnIfEmpty(creating.password, 'agentList.modal.passwordRequired')) {
      return false;
    }
    if (
      warnIfEmpty(
        creating.confirmPassword,
        'agentList.modal.confirmPasswordRequired'
      )
    ) {
      return false;
    }
    if (creating.confirmPassword !== creating.password) {
      Message.warning(t('agentList.modal.confirmPasswordMismatch'));
      return false;
    }
    if (!WITHDRAW_PASSWORD_PATTERN.test(creating.withdrawPassword)) {
      Message.warning(t('agentList.modal.withdrawPasswordInvalid'));
      return false;
    }
    if (creating.shareRatio == null) {
      Message.warning(t('agentList.modal.shareRequired'));
      return false;
    }
    if (shareExceeded.value) {
      Message.warning(
        t('agentList.modal.shareExceedParent', { max: shareMax.value })
      );
      return false;
    }
    const remark = creating.remark.trim();
    try {
      await createAgent({
        agent: {
          account,
          code: account,
          password: creating.password,
          withdraw_password: creating.withdrawPassword,
          agent_type: AGENT_TYPE.DEFAULT,
          agent_line_code: lineCode,
          share_rate: Math.round(creating.shareRatio),
          ...(remark ? { remark } : {}),
        },
        is_init: false,
        parent_id: creating.parentId,
      });
      Message.success(t('agentList.message.created'));
      return true;
    } catch {
      return false;
    }
  }

  async function handleBeforeOk(done: (closed: boolean) => void) {
    const ok = await submitCreate();
    done(ok);
    if (!ok) return;
    emit('update:visible', false);
    emit('created');
  }

  /** 详情页逐项修改，改一项就提交一次 */
  function onEditUpdate(patch: Partial<AgentFormValue>) {
    Object.assign(form, patch);
    if (profile.value) {
      if (patch.shareRatio != null) profile.value.share_rate = patch.shareRatio;
      if (patch.remark != null) profile.value.remark = patch.remark;
    }
    const account = (form.account || '').trim();
    if (!account) {
      Message.warning(t('agentList.modal.accountRequired'));
      return;
    }
    emit('submit', {
      ...form,
      account,
      parentAgent: (form.parentAgent || '').trim(),
      topAgent: (form.topAgent || '').trim(),
      remark: (form.remark || '').trim(),
      shareRatio: Math.round(form.shareRatio ?? 0),
    });
  }

  /** 列表只区分「不限制 / 限制」，限制本人和限制全线都按限制出款展示 */
  function onPayoutUpdate(value: number) {
    if (profile.value) profile.value.payout_restricted = value;
    onEditUpdate({
      withdrawLimit: value === AGENT_PROFILE_PAYOUT.NONE ? 'none' : 'forbidden',
    });
  }
</script>

<style lang="less" scoped>
  .empty-hint {
    color: var(--color-text-3);
    font-size: 13px;
    line-height: 1.5;
  }

  .parent-select {
    display: block;
    width: 100%;
  }

  .line-select-wrap {
    width: 100%;
  }

  .line-select {
    display: block;
    width: 100%;
  }

  .line-select-wrap .empty-hint {
    margin-top: 6px;
  }

  .share-range {
    color: var(--color-text-3);
  }
</style>
