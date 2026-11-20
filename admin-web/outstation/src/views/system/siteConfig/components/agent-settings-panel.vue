<template>
  <a-spin :loading="loading" class="agent-settings-panel">
    <a-card :bordered="false" class="section-card">
      <div class="section-head">
        <span class="section-icon purple">
          <icon-user-group />
        </span>
        <div>
          <h2>{{ t('agentList.settings.title') }}</h2>
          <p>配置开放代理层级、占成比例与现金代理归属</p>
        </div>
      </div>

      <div class="section-body">
        <a-form
          :model="form"
          layout="horizontal"
          class="agent-settings-form"
          :label-col-props="{ flex: '128px' }"
          :wrapper-col-props="{ flex: '1' }"
          label-align="left"
        >
          <a-form-item :label="t('agentList.settings.openLevels')" required>
            <div class="level-list">
              <div
                v-for="item in form.levels"
                :key="item.layer"
                class="level-row"
              >
                <span class="level-layer">
                  {{ t('agentList.settings.layer', { n: item.layer }) }}
                </span>
                <a-input
                  v-model="item.name"
                  class="level-name"
                  :max-length="16"
                />
              </div>
            </div>
          </a-form-item>
          <a-form-item :label="t('agentList.settings.maxShare')" required>
            <div class="ratio-field">
              <a-input-number
                v-model="form.maxShareRatio"
                :min="0"
                :max="100"
                :precision="0"
                :step="1"
                hide-button
                class="ratio-input"
              />
              <span class="ratio-suffix">%</span>
            </div>
          </a-form-item>
          <a-form-item :label="t('agentList.settings.applyShare')" required>
            <div class="apply-share">
              <div class="ratio-field">
                <span class="ratio-prefix">
                  {{ t('agentList.settings.highest') }}
                </span>
                <a-input-number
                  v-model="form.applyShareMax"
                  :min="0"
                  :max="100"
                  :precision="0"
                  :step="1"
                  hide-button
                  class="ratio-input"
                />
                <span class="ratio-suffix">%</span>
              </div>
              <div class="ratio-field">
                <span class="ratio-prefix">
                  {{ t('agentList.settings.lowest') }}
                </span>
                <a-input-number
                  v-model="form.applyShareMin"
                  :min="0"
                  :max="100"
                  :precision="0"
                  :step="1"
                  hide-button
                  class="ratio-input"
                />
                <span class="ratio-suffix">%</span>
              </div>
            </div>
          </a-form-item>
        </a-form>

        <div class="settings-note">
          <span class="settings-note-mark">*</span>
          {{ t('agentList.settings.hint') }}
        </div>

        <div class="divider" />

        <div class="cash-grid">
          <div class="cash-field">
            <div class="field-label">代理最大会员数(现金)</div>
            <a-input-number
              v-model="form.cashMaxMemberCount"
              :min="0"
              :precision="0"
              :step="1"
              hide-button
              class="cash-input"
            />
          </div>

          <div class="cash-field">
            <div class="field-label">IP最大可注册代理数(现金)</div>
            <a-input-number
              v-model="form.cashMaxAgentPerIp"
              :min="0"
              :precision="0"
              :step="1"
              hide-button
              class="cash-input"
            />
          </div>

          <div class="cash-field">
            <div class="field-label">现金代理指定直属代理</div>
            <a-select
              v-model="form.cashProductionAgentId"
              allow-search
              allow-clear
              :loading="agentLoading"
              :filter-option="filterAgentOption"
              placeholder="请选择"
              :format-label="formatAgentLabel"
            >
              <a-option
                v-for="item in productionAgents"
                :key="item.id"
                :value="item.id"
                :label="agentOptionLabel(item)"
              >
                {{ agentOptionLabel(item) }}
              </a-option>
            </a-select>
            <div class="field-hint">开设代理账号指定归属（现金正式线）</div>
          </div>

          <div class="cash-field">
            <div class="field-label">现金代理指定测试代理</div>
            <a-select
              v-model="form.cashTestAgentId"
              allow-search
              allow-clear
              :loading="agentLoading"
              :filter-option="filterAgentOption"
              placeholder="请选择"
              :format-label="formatAgentLabel"
            >
              <a-option
                v-for="item in testAgents"
                :key="item.id"
                :value="item.id"
                :label="agentOptionLabel(item)"
              >
                {{ agentOptionLabel(item) }}
              </a-option>
            </a-select>
            <div class="field-hint">测试账号数据不参与财务结算处理</div>
          </div>

          <div class="cash-field">
            <div class="field-label">现金代理指定试玩代理</div>
            <a-select
              v-model="form.cashDemoAgentId"
              allow-search
              allow-clear
              :loading="agentLoading"
              :filter-option="filterAgentOption"
              placeholder="请选择"
              :format-label="formatAgentLabel"
            >
              <a-option
                v-for="item in demoAgents"
                :key="item.id"
                :value="item.id"
                :label="agentOptionLabel(item)"
              >
                {{ agentOptionLabel(item) }}
              </a-option>
            </a-select>
            <div class="field-hint">内部账号数据可在账务处理排除</div>
          </div>
        </div>

        <div class="panel-actions">
          <a-button :disabled="!dirty || saving" @click="onReset">
            取消修改
          </a-button>
          <a-button
            type="primary"
            :loading="saving"
            :disabled="!dirty || loading"
            @click="onSave"
          >
            保存
          </a-button>
        </div>
      </div>
    </a-card>
  </a-spin>
</template>

<script lang="ts" setup>
  import { computed, onMounted, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import {
    AGENT_LINE,
    getAgentConfig,
    listAgentLevels,
    listOperatorAgents,
    searchAgents,
    updateAgentConfig,
    updateAgentLevels,
    type AgentConfig,
    type OperatorAgentItem,
  } from '@/api/agent';
  import { createDefaultAgentSettings } from '@/views/agent/list/options';
  import type { AgentLevelSetting } from '@/views/agent/list/types';

  const { t } = useI18n();
  /** 线路默认代理选项：GET /agent/operator/list，业务常用 level=2 */
  const OPERATOR_AGENT_LEVEL = 2;

  interface PanelForm {
    levels: AgentLevelSetting[];
    maxShareRatio: number;
    applyShareMax: number;
    applyShareMin: number;
    defaultLevel: string;
    defaultShareRatio: number;
    cashMaxMemberCount: number;
    cashMaxAgentPerIp: number;
    cashProductionAgentId: number | undefined;
    cashTestAgentId: number | undefined;
    cashDemoAgentId: number | undefined;
  }

  function createForm(): PanelForm {
    const defaults = createDefaultAgentSettings();
    return {
      levels: defaults.levels.map((item) => ({ ...item })),
      maxShareRatio: defaults.maxShareRatio,
      applyShareMax: defaults.applyShareMax,
      applyShareMin: defaults.applyShareMin,
      defaultLevel: defaults.defaultLevel,
      defaultShareRatio: defaults.defaultShareRatio,
      cashMaxMemberCount: 0,
      cashMaxAgentPerIp: 0,
      cashProductionAgentId: undefined,
      cashTestAgentId: undefined,
      cashDemoAgentId: undefined,
    };
  }

  function cloneForm(source: PanelForm): PanelForm {
    return {
      ...source,
      levels: source.levels.map((item) => ({ ...item })),
    };
  }

  function toAgentId(value: number | undefined | null) {
    return value && value > 0 ? value : undefined;
  }

  function toSaveAgentId(value: number | undefined) {
    return value && value > 0 ? value : 0;
  }

  const form = reactive<PanelForm>(createForm());
  const saved = ref<PanelForm>(cloneForm(form));
  const loading = ref(false);
  const saving = ref(false);
  const agentLoading = ref(false);
  const loadedConfig = ref<AgentConfig | null>(null);
  const agentGroups = ref(
    new Map<string, OperatorAgentItem[]>()
  );
  const extraAgents = ref(new Map<number, OperatorAgentItem>());

  const dirty = computed(
    () => JSON.stringify(form) !== JSON.stringify(saved.value)
  );

  function agentsForLine(lineCode: string, selectedId?: number) {
    const list = [...(agentGroups.value.get(lineCode) || [])];
    if (selectedId && selectedId > 0 && !list.some((item) => item.id === selectedId)) {
      const extra = extraAgents.value.get(selectedId);
      if (extra) list.unshift(extra);
      else {
        list.unshift({
          id: selectedId,
          account: String(selectedId),
          level: 0,
          level_alias: '',
          line_code: lineCode,
        });
      }
    }
    return list;
  }

  const productionAgents = computed(() =>
    agentsForLine(AGENT_LINE.CASH_PRODUCTION, form.cashProductionAgentId)
  );
  const testAgents = computed(() =>
    agentsForLine(AGENT_LINE.CASH_TEST, form.cashTestAgentId)
  );
  const demoAgents = computed(() =>
    agentsForLine(AGENT_LINE.CASH_DEMO, form.cashDemoAgentId)
  );

  function agentOptionLabel(item: OperatorAgentItem) {
    return item.level_alias
      ? `${item.account}（${item.level_alias}）`
      : item.account;
  }

  const agentLabelMap = computed(() => {
    const map = new Map<number, string>();
    agentGroups.value.forEach((list) => {
      list.forEach((item) => map.set(item.id, agentOptionLabel(item)));
    });
    extraAgents.value.forEach((item) =>
      map.set(item.id, agentOptionLabel(item))
    );
    return map;
  });

  function formatAgentLabel(option: { label?: string; value?: number }) {
    if (option.label) return option.label;
    const id = Number(option.value);
    return agentLabelMap.value.get(id) || String(id || '');
  }

  function filterAgentOption(input: string, option: { label?: string }) {
    const keyword = input.trim().toLowerCase();
    if (!keyword) return true;
    return String(option.label || '')
      .toLowerCase()
      .includes(keyword);
  }

  function rememberExtra(list: OperatorAgentItem[]) {
    if (!list.length) return;
    const next = new Map(extraAgents.value);
    list.forEach((item) => next.set(item.id, item));
    extraAgents.value = next;
  }

  async function resolveMissingAgents(ids: number[]) {
    const known = new Set<number>();
    agentGroups.value.forEach((list) => {
      list.forEach((item) => known.add(item.id));
    });
    const pending = [...new Set(ids.filter((id) => id > 0 && !known.has(id)))];
    if (!pending.length) return;
    const results = await Promise.allSettled(
      pending.map((id) => searchAgents({ keyword: String(id), limit: 5 }, true))
    );
    const found: OperatorAgentItem[] = [];
    results.forEach((result, index) => {
      if (result.status !== 'fulfilled') return;
      const matched = (result.value.data?.list || []).find(
        (item) => item.id === pending[index]
      );
      if (!matched) return;
      found.push({
        id: matched.id,
        account: matched.account,
        level: matched.level,
        level_alias: matched.level_alias,
        line_code: matched.line_code,
        share_rate: matched.share_rate,
      });
    });
    rememberExtra(found);
  }

  async function loadOperatorAgents() {
    agentLoading.value = true;
    try {
      const { data } = await listOperatorAgents(OPERATOR_AGENT_LEVEL);
      const next = new Map<string, OperatorAgentItem[]>();
      (data?.groups || []).forEach((group) => {
        next.set(group.line_code, group.agents || []);
      });
      agentGroups.value = next;
    } catch {
      agentGroups.value = new Map();
    } finally {
      agentLoading.value = false;
    }
  }

  function applyConfig(data: AgentConfig) {
    form.maxShareRatio = data.max_share_rate || 0;
    form.applyShareMax = data.apply_max_share_rate || 0;
    form.applyShareMin = data.apply_min_share_rate || 0;
    form.cashMaxMemberCount = data.cash_max_member_count || 0;
    form.cashMaxAgentPerIp = data.cash_max_agent_per_ip || 0;
    form.cashProductionAgentId = toAgentId(data.cash_production_agent_id);
    form.cashTestAgentId = toAgentId(data.cash_test_agent_id);
    form.cashDemoAgentId = toAgentId(data.cash_demo_agent_id);
  }

  function applyLevels(
    items: { level: number; alias?: string; name: string }[]
  ) {
    if (!items.length) return;
    form.levels = [...items]
      .sort((a, b) => a.level - b.level)
      .map((item) => ({
        layer: item.level,
        name: item.alias || item.name,
      }));
  }

  function snapshot() {
    saved.value = cloneForm(form);
  }

  async function loadConfig() {
    loading.value = true;
    try {
      const [{ data }, { data: levelData }] = await Promise.all([
        getAgentConfig(),
        listAgentLevels(),
        loadOperatorAgents(),
      ]);
      loadedConfig.value = data;
      applyConfig(data);
      applyLevels(levelData?.list || []);
      await resolveMissingAgents([
        data.cash_production_agent_id || 0,
        data.cash_test_agent_id || 0,
        data.cash_demo_agent_id || 0,
      ]);
      snapshot();
    } catch {
      // 错误提示由请求拦截器统一弹出
    } finally {
      loading.value = false;
    }
  }

  function onReset() {
    const next = cloneForm(saved.value);
    form.levels = next.levels;
    form.maxShareRatio = next.maxShareRatio;
    form.applyShareMax = next.applyShareMax;
    form.applyShareMin = next.applyShareMin;
    form.defaultLevel = next.defaultLevel;
    form.defaultShareRatio = next.defaultShareRatio;
    form.cashMaxMemberCount = next.cashMaxMemberCount;
    form.cashMaxAgentPerIp = next.cashMaxAgentPerIp;
    form.cashProductionAgentId = next.cashProductionAgentId;
    form.cashTestAgentId = next.cashTestAgentId;
    form.cashDemoAgentId = next.cashDemoAgentId;
  }

  async function onSave() {
    const levels = form.levels.map((item) => ({
      layer: item.layer,
      name: item.name.trim(),
    }));
    if (levels.some((item) => !item.name)) {
      Message.warning(t('agentList.settings.levelRequired'));
      return;
    }
    const uniqueNames = new Set(levels.map((item) => item.name));
    if (uniqueNames.size !== levels.length) {
      Message.warning(t('agentList.settings.levelDuplicate'));
      return;
    }
    const maxShare = form.maxShareRatio;
    const applyMax = form.applyShareMax;
    const applyMin = form.applyShareMin;
    if (maxShare == null || applyMax == null || applyMin == null) {
      Message.warning(t('agentList.settings.required'));
      return;
    }
    if (applyMax > maxShare) {
      Message.warning(t('agentList.settings.applyMaxInvalid'));
      return;
    }
    if (applyMin > applyMax) {
      Message.warning(t('agentList.settings.applyMinInvalid'));
      return;
    }
    if (form.cashMaxMemberCount == null) {
      Message.warning('请填写代理最大会员数(现金)');
      return;
    }
    if (form.cashMaxAgentPerIp == null) {
      Message.warning('请填写IP最大可注册代理数(现金)');
      return;
    }

    const current = loadedConfig.value;
    if (!current) {
      Message.warning(t('agentList.settings.required'));
      return;
    }

    saving.value = true;
    try {
      const payload: AgentConfig = {
        max_share_rate: Math.round(maxShare),
        apply_min_share_rate: Math.round(applyMin),
        apply_max_share_rate: Math.round(applyMax),
        cash_max_member_count: form.cashMaxMemberCount || 0,
        cash_max_agent_per_ip: form.cashMaxAgentPerIp || 0,
        cash_production_agent_id: toSaveAgentId(form.cashProductionAgentId),
        cash_demo_agent_id: toSaveAgentId(form.cashDemoAgentId),
        cash_test_agent_id: toSaveAgentId(form.cashTestAgentId),
        credit_demo_agent_id: current.credit_demo_agent_id || 0,
        credit_test_agent_id: current.credit_test_agent_id || 0,
      };
      await Promise.all([
        updateAgentLevels(
          levels.map((item) => ({
            level: item.layer,
            name: item.name,
            alias: item.name,
          }))
        ),
        updateAgentConfig(payload),
      ]);
      loadedConfig.value = { ...current, ...payload };
      form.levels = levels;
      snapshot();
      Message.success(t('agentList.message.settingsSaved'));
    } catch {
      // 错误提示由请求拦截器统一弹出
    } finally {
      saving.value = false;
    }
  }

  onMounted(loadConfig);
</script>

<style lang="less" scoped>
  .agent-settings-panel {
    display: block;
    width: 100%;
  }

  .section-card {
    border-radius: 10px;
    box-shadow: 0 1px 3px rgba(15, 23, 42, 0.04);

    :deep(.arco-card-body) {
      padding: 0;
    }
  }

  .section-head {
    display: flex;
    gap: 12px;
    align-items: center;
    padding: 18px 24px;
    border-bottom: 1px solid var(--color-border-1);

    h2 {
      margin: 0 0 2px;
      font-size: 15px;
      font-weight: 600;
      color: var(--color-text-1);
    }

    p {
      margin: 0;
      font-size: 12px;
      color: var(--color-text-3);
    }
  }

  .section-icon {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    font-size: 18px;
    border-radius: 9px;
  }

  .purple {
    color: rgb(var(--purple-6));
    background: rgb(var(--purple-1));
  }

  .section-body {
    max-width: 960px;
    padding: 22px 24px 24px;
  }

  .agent-settings-form {
    :deep(.arco-form-item) {
      align-items: flex-start;
      margin-bottom: 16px;
    }

    :deep(.arco-form-item-label-col) {
      line-height: 32px;
      padding-right: 8px;
    }

    :deep(.arco-form-item-label) {
      color: var(--color-text-1);
    }
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
    border: 1px solid var(--color-border-2);
    border-radius: 2px;
    background: var(--color-bg-2);
    overflow: hidden;

    &:focus-within {
      border-color: rgb(var(--primary-6));
    }
  }

  .level-layer {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 72px;
    flex-shrink: 0;
    border-right: 1px solid var(--color-border-2);
    background: var(--color-fill-2);
    color: var(--color-text-2);
  }

  .level-name {
    flex: 1;
    min-width: 0;
    border: none;
    background: transparent;

    :deep(.arco-input-wrapper) {
      border: none;
      background: transparent;
      box-shadow: none;
    }
  }

  .apply-share {
    display: flex;
    gap: 16px;
    width: 100%;
  }

  .ratio-field {
    display: flex;
    align-items: center;
    flex: 1;
    min-width: 0;
    height: 32px;
    border: 1px solid var(--color-border-2);
    border-radius: 2px;
    background: var(--color-bg-2);
    overflow: hidden;

    &:focus-within {
      border-color: rgb(var(--primary-6));
    }
  }

  .ratio-prefix,
  .ratio-suffix {
    flex-shrink: 0;
    color: var(--color-text-2);
  }

  .ratio-prefix {
    padding-left: 12px;
  }

  .ratio-suffix {
    padding-right: 12px;
  }

  .ratio-input {
    flex: 1;
    min-width: 0;
    border: none;
    background: transparent;

    :deep(.arco-input-wrapper) {
      border: none;
      background: transparent;
      box-shadow: none;
    }

    :deep(.arco-input) {
      text-align: left;
    }
  }

  .settings-note {
    margin: 4px 0 8px 128px;
    color: var(--color-text-1);
    font-size: 14px;
    line-height: 22px;
  }

  .settings-note-mark {
    margin-right: 4px;
    color: rgb(var(--danger-6));
  }

  .divider {
    height: 1px;
    margin: 20px 0 24px;
    background: var(--color-border-1);
  }

  .cash-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 20px 24px;
  }

  .cash-field {
    min-width: 0;
  }

  .field-label {
    margin-bottom: 8px;
    font-size: 13px;
    font-weight: 500;
    color: var(--color-text-1);
  }

  .field-hint {
    margin-top: 6px;
    font-size: 12px;
    color: var(--color-text-3);
  }

  .cash-input {
    width: 100%;
  }

  .panel-actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    margin-top: 24px;
  }

  @media (max-width: 768px) {
    .cash-grid {
      grid-template-columns: 1fr;
    }

    .settings-note {
      margin-left: 0;
    }
  }
</style>
