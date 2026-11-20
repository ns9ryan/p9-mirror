<template>
  <div class="agent-init-page">
    <div class="agent-init-card">
      <div class="agent-init-head">
        <div>
          <h2 class="agent-init-title">{{ t('agentInit.title') }}</h2>
          <p class="agent-init-desc">{{ t('agentInit.desc') }}</p>
        </div>
        <a-button type="text" @click="handleLogout">
          {{ t('agentInit.logout') }}
        </a-button>
      </div>

      <a-spin :loading="loading" class="agent-init-spin">
        <a-result
          v-if="loadFailed"
          status="error"
          :title="t('agentInit.loadFailed')"
        >
          <template #extra>
            <a-button type="primary" @click="refresh">
              {{ t('agentInit.retry') }}
            </a-button>
          </template>
        </a-result>

        <template v-else-if="status">
          <div v-if="lines.length" class="line-progress">
            <span class="line-progress__label">
              {{ t('agentInit.progress') }}
            </span>
            <a-tag
              v-for="line in lines"
              :key="line.line_code"
              :color="lineTagColor(line)"
            >
              {{ lineLabel(line.line_code) }} · {{ lineTagText(line) }}
            </a-tag>
          </div>

          <a-steps v-if="currentLine" :current="step + 1" class="init-steps">
            <a-step>{{ t('agentInit.step.levels') }}</a-step>
            <a-step>{{ t('agentInit.step.agent') }}</a-step>
          </a-steps>

          <a-form
            v-if="step === 0"
            :model="levelForm"
            layout="horizontal"
            :label-col-props="{ span: 5 }"
            :wrapper-col-props="{ span: 19 }"
          >
            <a-form-item :label="t('agentInit.levelCount')" required>
              <a-radio-group v-model="levelForm.count" type="button">
                <a-radio v-for="n in LEVEL_COUNT_OPTIONS" :key="n" :value="n">
                  {{ t('agentList.settings.layer', { n }) }}
                </a-radio>
              </a-radio-group>
            </a-form-item>
            <a-form-item :label="t('agentList.settings.openLevels')" required>
              <div class="level-list">
                <div
                  v-for="item in openLevels"
                  :key="item.layer"
                  class="level-row"
                >
                  <span class="level-layer">
                    {{ t('agentList.settings.layer', { n: item.layer }) }}
                  </span>
                  <a-input
                    v-model="item.name"
                    class="level-name"
                    :max-length="LEVEL_NAME_MAX"
                  />
                </div>
              </div>
            </a-form-item>
          </a-form>

          <a-form
            v-else
            :model="agent"
            layout="horizontal"
            :label-col-props="{ span: 5 }"
            :wrapper-col-props="{ span: 19 }"
          >
            <a-form-item :label="t('agentList.modal.line')">
              <span>{{ lineLabel(currentLine?.line_code || '') }}</span>
            </a-form-item>
            <a-form-item :label="t('agentList.card.account')" required>
              <a-input
                v-model="agent.account"
                :placeholder="t('agentList.modal.accountPlaceholder')"
                allow-clear
              />
            </a-form-item>
            <a-form-item
              :label="t('agentList.modal.password')"
              :extra="defaultTip(agent.password, DEFAULT_PASSWORD)"
              required
            >
              <a-input-password
                v-model="agent.password"
                :placeholder="t('agentList.modal.passwordPlaceholder')"
                autocomplete="new-password"
                allow-clear
              />
            </a-form-item>
            <a-form-item :label="t('agentList.modal.confirmPassword')" required>
              <a-input-password
                v-model="agent.confirmPassword"
                :placeholder="t('agentList.modal.confirmPasswordPlaceholder')"
                allow-clear
              />
            </a-form-item>
            <a-form-item
              :label="t('agentList.modal.withdrawPassword')"
              :extra="
                defaultTip(agent.withdrawPassword, DEFAULT_WITHDRAW_PASSWORD)
              "
              required
            >
              <a-input-password
                v-model="agent.withdrawPassword"
                :placeholder="t('agentList.modal.withdrawPasswordPlaceholder')"
                :max-length="WITHDRAW_PASSWORD_LENGTH"
                allow-clear
              />
            </a-form-item>
            <a-form-item :label="t('agentList.modal.agentMode')">
              <span>{{ t('agentList.modal.agentModeShare') }}</span>
            </a-form-item>
            <a-form-item :label="t('agentList.card.share')" required>
              <a-input-number
                v-model="agent.shareRatio"
                :min="0"
                :max="SHARE_RATIO_MAX"
                :precision="0"
                :step="1"
                :placeholder="t('agentList.modal.sharePlaceholder')"
                hide-button
              >
                <template #suffix>
                  <span class="share-range">(0%~{{ SHARE_RATIO_MAX }}%)</span>
                </template>
              </a-input-number>
            </a-form-item>
            <a-form-item :label="t('agentList.modal.remark')">
              <a-textarea
                v-model="agent.remark"
                :placeholder="t('agentList.modal.remarkPlaceholder')"
                :auto-size="{ minRows: 4, maxRows: 6 }"
              />
            </a-form-item>
          </a-form>

          <a-alert
            v-if="submitError"
            type="error"
            :title="t('agentInit.result.failTitle')"
            closable
            class="submit-error"
            @close="submitError = ''"
          >
            {{ submitError }}
          </a-alert>

          <div class="agent-init-footer">
            <template v-if="step === 0">
              <a-button type="primary" :loading="saving" @click="handleNext">
                {{ currentLine ? t('agentInit.next') : t('agentInit.finish') }}
              </a-button>
            </template>
            <template v-else>
              <a-button
                v-if="!status.levels_ready"
                :disabled="saving"
                @click="backToLevels"
              >
                {{ t('agentInit.prev') }}
              </a-button>
              <a-button type="primary" :loading="saving" @click="handleSubmit">
                {{ t('agentInit.submit') }}
              </a-button>
            </template>
          </div>
        </template>
      </a-spin>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { computed, onMounted, reactive, ref } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import type { RouteRecordRaw } from 'vue-router';
  import { useI18n } from 'vue-i18n';
  import { Message, Notification } from '@arco-design/web-vue';
  import {
    AGENT_TYPE,
    createAgent,
    getAgentInitStatus,
    updateAgentLevels,
  } from '@/api/agent';
  import type { AgentInitLine, AgentInitStatus } from '@/api/agent';
  import { useAppStore, useUserStore } from '@/store';
  import {
    getRedirectQuery,
    resolvePostLoginTarget,
  } from '@/router/login-redirect';
  import { markAgentInitReady } from '@/utils/agent-init';
  import { OPEN_AGENT_LEVELS } from '@/views/agent/list/options';

  const LEVEL_NAME_MAX = 16;
  const LEVEL_COUNT_OPTIONS = [3, 4, 5, 6, 7];
  const WITHDRAW_PASSWORD_LENGTH = 6;
  const SHARE_RATIO_MAX = 100;
  const DEFAULT_PASSWORD = 'Admin@123';
  const DEFAULT_WITHDRAW_PASSWORD = '123456';

  const { t } = useI18n();
  const route = useRoute();
  const router = useRouter();
  const appStore = useAppStore();
  const userStore = useUserStore();

  const status = ref<AgentInitStatus | null>(null);
  const loading = ref(false);
  const loadFailed = ref(false);
  const saving = ref(false);
  const submitError = ref('');
  /** 本页内实际完成的初始化，用于进入系统后提示结果 */
  const levelsSaved = ref(false);
  const createdLineCount = ref(0);
  const step = ref(0);
  const levelForm = reactive({
    count: LEVEL_COUNT_OPTIONS[LEVEL_COUNT_OPTIONS.length - 1],
    levels: OPEN_AGENT_LEVELS.map((item) => ({ ...item })),
  });
  const openLevels = computed(() => levelForm.levels.slice(0, levelForm.count));

  function createAgentForm() {
    return {
      account: '',
      password: DEFAULT_PASSWORD,
      confirmPassword: DEFAULT_PASSWORD,
      withdrawPassword: DEFAULT_WITHDRAW_PASSWORD,
      shareRatio: SHARE_RATIO_MAX as number | undefined,
      remark: '',
    };
  }

  const agent = reactive(createAgentForm());

  /** 密码框是掩码，仍为默认值时把默认密码展示出来 */
  function defaultTip(value: string, defaultValue: string) {
    return value === defaultValue
      ? t('agentInit.defaultPasswordTip', { password: defaultValue })
      : '';
  }

  const lines = computed(() => status.value?.lines || []);
  const currentLine = computed(
    () => lines.value.find((line) => !line.initialized) || null
  );

  function lineLabel(lineCode: string) {
    const key = `agentList.line.${String(lineCode || '').toUpperCase()}`;
    const text = t(key);
    return text === key ? lineCode : text;
  }

  function lineTagColor(line: AgentInitLine) {
    if (line.initialized) return 'green';
    return line.line_code === currentLine.value?.line_code
      ? 'arcoblue'
      : 'gray';
  }

  function lineTagText(line: AgentInitLine) {
    if (line.initialized) return t('agentInit.lineDone');
    return line.line_code === currentLine.value?.line_code
      ? t('agentInit.lineCurrent')
      : t('agentInit.linePending');
  }

  async function finish() {
    markAgentInitReady();
    const target = resolvePostLoginTarget(
      router,
      getRedirectQuery(route.query),
      appStore.appAsyncMenus as unknown as RouteRecordRaw[]
    );
    await router.replace(target);
    if (!levelsSaved.value && !createdLineCount.value) return;
    Notification.success({
      title: t('agentInit.result.successTitle'),
      content: createdLineCount.value
        ? t('agentInit.result.successLines', { count: createdLineCount.value })
        : t('agentInit.result.successLevels'),
      duration: 6000,
      closable: true,
    });
  }

  function errorText(err: unknown) {
    const error = err as {
      message?: string;
      response?: { data?: { msg?: string } };
    };
    const reason =
      error?.response?.data?.msg ||
      error?.message ||
      t('agentInit.result.unknownReason');
    return t('agentInit.result.failDesc', { reason });
  }

  function backToLevels() {
    submitError.value = '';
    step.value = 0;
  }

  async function refresh() {
    loading.value = true;
    loadFailed.value = false;
    try {
      const { data } = await getAgentInitStatus();
      status.value = data || null;
    } catch {
      loadFailed.value = true;
    } finally {
      loading.value = false;
    }
    if (status.value?.ready) {
      await finish();
      return;
    }
    // 层级全站只设置一次，已就绪后每条线直接填顶层代理
    step.value = status.value?.levels_ready && currentLine.value ? 1 : 0;
  }

  function levelNames() {
    const names = openLevels.value.map((item) => item.name.trim());
    if (names.some((name) => !name)) {
      Message.warning(t('agentList.settings.levelRequired'));
      return null;
    }
    if (new Set(names).size !== names.length) {
      Message.warning(t('agentList.settings.levelDuplicate'));
      return null;
    }
    return names;
  }

  async function saveLevels(names: string[]) {
    try {
      await updateAgentLevels(
        openLevels.value.map((item, index) => ({
          level: item.layer,
          name: names[index],
          alias: names[index],
        }))
      );
      levelsSaved.value = true;
      return true;
    } catch (err) {
      submitError.value = errorText(err);
      return false;
    }
  }

  async function handleNext() {
    const names = levelNames();
    if (!names) return;
    submitError.value = '';
    saving.value = true;
    const ok = await saveLevels(names);
    saving.value = false;
    if (!ok) return;
    if (currentLine.value) step.value = 1;
    else await refresh();
  }

  function agentFormError() {
    if (!agent.account.trim()) return 'agentList.modal.accountRequired';
    if (!agent.password) return 'agentList.modal.passwordRequired';
    if (!agent.confirmPassword) {
      return 'agentList.modal.confirmPasswordRequired';
    }
    if (agent.confirmPassword !== agent.password) {
      return 'agentList.modal.confirmPasswordMismatch';
    }
    if (!/^\d{6}$/.test(agent.withdrawPassword)) {
      return 'agentList.modal.withdrawPasswordInvalid';
    }
    if (agent.shareRatio == null) return 'agentList.modal.shareRequired';
    return '';
  }

  async function handleSubmit() {
    const line = currentLine.value;
    if (!line) return;
    const error = agentFormError();
    if (error) {
      Message.warning(t(error));
      return;
    }
    const account = agent.account.trim();
    const remark = agent.remark.trim();
    submitError.value = '';
    saving.value = true;
    try {
      await createAgent({
        agent: {
          account,
          password: agent.password,
          withdraw_password: agent.withdrawPassword,
          agent_type: AGENT_TYPE.DEFAULT,
          agent_line_code: line.line_code,
          share_rate: Math.min(
            SHARE_RATIO_MAX,
            Math.round(agent.shareRatio || 0)
          ),
          ...(remark ? { remark } : {}),
        },
        is_init: true,
        parent_id: 0,
      });
    } catch (err) {
      saving.value = false;
      submitError.value = errorText(err);
      return;
    }
    saving.value = false;
    createdLineCount.value += 1;
    Message.success(
      t('agentInit.lineCreated', { line: lineLabel(line.line_code) })
    );
    Object.assign(agent, createAgentForm());
    await refresh();
  }

  async function handleLogout() {
    await userStore.logout();
    await router.replace({ name: 'login' });
  }

  onMounted(refresh);
</script>

<style scoped lang="less">
  .agent-init-page {
    display: flex;
    align-items: flex-start;
    justify-content: center;
    box-sizing: border-box;
    min-height: 100vh;
    padding: 48px 16px;
    background: var(--color-fill-2);
  }

  .agent-init-card {
    box-sizing: border-box;
    width: min(100%, 760px);
    padding: 28px 32px;
    background: var(--color-bg-2);
    border-radius: 12px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.06);
  }

  .agent-init-head {
    display: flex;
    gap: 16px;
    align-items: flex-start;
    justify-content: space-between;
    margin-bottom: 20px;
  }

  .agent-init-title {
    margin: 0 0 6px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 20px;
  }

  .agent-init-desc {
    margin: 0;
    color: var(--color-text-3);
    font-size: 13px;
  }

  .agent-init-spin {
    display: block;
    min-height: 200px;
  }

  .line-progress {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
    margin-bottom: 24px;
    padding: 12px;
    background: var(--color-fill-1);
    border-radius: 8px;
  }

  .line-progress__label {
    color: var(--color-text-2);
    font-size: 13px;
  }

  .init-steps {
    margin-bottom: 28px;
  }

  .level-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
    width: 100%;
  }

  .level-row {
    display: flex;
    align-items: stretch;
    height: 32px;
    overflow: hidden;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 2px;

    &:focus-within {
      border-color: rgb(var(--primary-6));
    }
  }

  .level-layer {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    width: 72px;
    color: var(--color-text-2);
    background: var(--color-fill-2);
    border-right: 1px solid var(--color-border-2);
  }

  .level-name {
    flex: 1;
    min-width: 0;
    background: transparent;
    border: none;

    :deep(.arco-input-wrapper) {
      background: transparent;
      border: none;
      box-shadow: none;
    }
  }

  .share-range {
    color: var(--color-text-3);
  }

  .submit-error {
    margin-bottom: 16px;
  }

  .agent-init-footer {
    display: flex;
    gap: 12px;
    justify-content: center;
    margin-top: 8px;
  }
</style>
