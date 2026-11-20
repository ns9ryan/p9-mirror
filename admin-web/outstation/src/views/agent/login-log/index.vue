<template>
  <div class="container list-page">
    <a-card class="general-card" :bordered="false">
      <a-tabs v-model:active-key="activeTab">
        <a-tab-pane key="login" :title="t('agentLog.tab.login')">
          <a-form :model="loginForm" layout="inline" class="search-form">
            <a-form-item :label="t('agentLog.search.loginTime')">
              <a-space wrap>
                <a-radio-group v-model="loginQuick" type="button" size="small">
                  <a-radio value="today">
                    {{ t('agentLog.search.today') }}
                  </a-radio>
                  <a-radio value="d7">{{ t('agentLog.search.last7') }}</a-radio>
                  <a-radio value="d30">
                    {{ t('agentLog.search.last30') }}
                  </a-radio>
                </a-radio-group>
                <a-range-picker
                  v-model="loginForm.range"
                  value-format="YYYY-MM-DD"
                  allow-clear
                  class="range-picker"
                  :placeholder="[
                    t('agentLog.search.start'),
                    t('agentLog.search.end'),
                  ]"
                  @change="onRangeChange('login')"
                />
              </a-space>
            </a-form-item>
            <a-form-item :label="t('agentLog.search.loginType')">
              <a-select v-model="loginForm.loginType" class="search-select">
                <a-option
                  v-for="item in LOGIN_TYPE_OPTIONS"
                  :key="item.value"
                  :value="item.value"
                >
                  {{ t(item.label) }}
                </a-option>
              </a-select>
            </a-form-item>
            <a-form-item>
              <a-input-group compact>
                <a-select v-model="loginForm.keywordType" class="search-type">
                  <a-option
                    v-for="item in KEYWORD_TYPE_OPTIONS"
                    :key="item.value"
                    :value="item.value"
                  >
                    {{ t(item.label) }}
                  </a-option>
                </a-select>
                <a-input
                  v-model="loginForm.keyword"
                  class="search-input"
                  allow-clear
                  :placeholder="t('agentLog.search.pleaseInput')"
                  @press-enter="searchLogin"
                />
              </a-input-group>
            </a-form-item>
            <a-form-item :label="t('agentLog.search.deviceType')">
              <a-select v-model="loginForm.deviceType" class="search-select">
                <a-option
                  v-for="item in DEVICE_TYPE_OPTIONS"
                  :key="item.value"
                  :value="item.value"
                >
                  {{ t(item.label) }}
                </a-option>
              </a-select>
            </a-form-item>
            <a-form-item>
              <a-space>
                <a-button
                  type="primary"
                  :loading="loginLoading"
                  @click="searchLogin"
                >
                  <template #icon><icon-search /></template>
                  {{ t('front.common.search.search') }}
                </a-button>
                <a-button @click="resetLogin">
                  <template #icon><icon-refresh /></template>
                  {{ t('front.common.search.reset') }}
                </a-button>
              </a-space>
            </a-form-item>
          </a-form>

          <a-table
            row-key="id"
            :data="loginRows"
            :loading="loginLoading"
            :pagination="false"
            :bordered="{ cell: true }"
            :scroll="{ x: 1200 }"
          >
            <template #columns>
              <a-table-column :title="t('agentLog.col.agentId')" :width="140">
                <template #cell="{ record }">
                  <div>{{ record.agent_id }}</div>
                  <div v-if="record.account" class="cell-sub">
                    {{ record.account }}
                  </div>
                </template>
              </a-table-column>
              <a-table-column :title="t('agentLog.col.loginType')" :width="100">
                <template #cell="{ record }">
                  {{ record.login_type_label || '--' }}
                </template>
              </a-table-column>
              <a-table-column :title="t('agentLog.col.loginIp')" :width="160">
                <template #cell="{ record }">
                  <div>{{ record.login_ip || '--' }}</div>
                  <div v-if="record.login_location" class="cell-sub">
                    {{ record.login_location }}
                  </div>
                </template>
              </a-table-column>
              <a-table-column
                :title="t('agentLog.col.deviceCode')"
                :width="140"
              >
                <template #cell="{ record }">
                  {{ record.udid || '--' }}
                </template>
              </a-table-column>
              <a-table-column
                :title="t('agentLog.col.loginDevice')"
                :width="180"
              >
                <template #cell="{ record }">
                  {{ record.login_device || '--' }}
                </template>
              </a-table-column>
              <a-table-column
                :title="t('agentLog.col.deviceType')"
                :width="110"
              >
                <template #cell="{ record }">
                  {{ record.device_type_label || '--' }}
                </template>
              </a-table-column>
              <a-table-column :title="t('agentLog.col.loginTime')" :width="180">
                <template #cell="{ record }">
                  {{ formatUnix(record.login_at) || '--' }}
                </template>
              </a-table-column>
              <a-table-column :title="t('agentLog.col.result')" :width="110">
                <template #cell="{ record }">
                  <span
                    :class="
                      record.result === AGENT_LOGIN_RESULT_SUCCESS
                        ? 'result-success'
                        : 'result-invalid'
                    "
                  >
                    {{ record.result_label || '--' }}
                  </span>
                </template>
              </a-table-column>
            </template>
          </a-table>
          <div class="table-pagination">
            <a-pagination
              :current="loginPageNo"
              :page-size="loginPageSize"
              :page-size-options="[10, 20, 50, 100]"
              :total="loginTotal"
              show-total
              show-page-size
              @change="onLoginPageChange"
              @page-size-change="onLoginPageSizeChange"
            />
          </div>
        </a-tab-pane>

        <a-tab-pane key="action" :title="t('agentLog.tab.action')">
          <a-form :model="actionForm" layout="inline" class="search-form">
            <a-form-item :label="t('agentLog.search.targetType')">
              <a-select v-model="actionForm.targetType" class="search-select">
                <a-option
                  v-for="item in TARGET_TYPE_OPTIONS"
                  :key="item.value"
                  :value="item.value"
                >
                  {{ t(item.label) }}
                </a-option>
              </a-select>
            </a-form-item>
            <a-form-item :label="t('agentLog.search.targetId')">
              <a-input-number
                v-model="actionForm.targetId"
                class="search-input"
                hide-button
                allow-clear
                :min="1"
                :precision="0"
                model-event="input"
                :placeholder="t('agentLog.search.pleaseInput')"
                @keydown.enter="searchAction"
              />
            </a-form-item>
            <a-form-item :label="t('agentLog.search.actionResult')">
              <a-select v-model="actionForm.result" class="search-select">
                <a-option
                  v-for="item in ACTION_RESULT_OPTIONS"
                  :key="item.value"
                  :value="item.value"
                >
                  {{ t(item.label) }}
                </a-option>
              </a-select>
            </a-form-item>
            <a-form-item :label="t('agentLog.search.actionTime')">
              <a-space wrap>
                <a-radio-group v-model="actionQuick" type="button" size="small">
                  <a-radio value="today">
                    {{ t('agentLog.search.today') }}
                  </a-radio>
                  <a-radio value="d7">{{ t('agentLog.search.last7') }}</a-radio>
                  <a-radio value="d30">
                    {{ t('agentLog.search.last30') }}
                  </a-radio>
                </a-radio-group>
                <a-range-picker
                  v-model="actionForm.range"
                  value-format="YYYY-MM-DD"
                  allow-clear
                  class="range-picker"
                  :placeholder="[
                    t('agentLog.search.start'),
                    t('agentLog.search.end'),
                  ]"
                  @change="onRangeChange('action')"
                />
              </a-space>
            </a-form-item>
            <a-form-item>
              <a-space>
                <a-button
                  type="primary"
                  :loading="actionLoading"
                  @click="searchAction"
                >
                  <template #icon><icon-search /></template>
                  {{ t('front.common.search.search') }}
                </a-button>
                <a-button @click="resetAction">
                  <template #icon><icon-refresh /></template>
                  {{ t('front.common.search.reset') }}
                </a-button>
              </a-space>
            </a-form-item>
          </a-form>

          <a-table
            row-key="id"
            :data="actionRows"
            :loading="actionLoading"
            :pagination="false"
            :bordered="{ cell: true }"
            :scroll="{ x: 1500 }"
          >
            <template #columns>
              <a-table-column
                :title="t('agentLog.col.id')"
                data-index="id"
                :width="80"
              />
              <a-table-column :title="t('agentLog.col.operator')" :width="180">
                <template #cell="{ record }">
                  <div>{{ record.admin_account || '--' }}</div>
                  <div v-if="operatorSub(record)" class="cell-sub">
                    {{ operatorSub(record) }}
                  </div>
                </template>
              </a-table-column>
              <a-table-column :title="t('agentLog.col.action')" :width="180">
                <template #cell="{ record }">
                  {{ record.action_label || record.action || '--' }}
                </template>
              </a-table-column>
              <a-table-column :title="t('agentLog.col.target')" :width="180">
                <template #cell="{ record }">
                  <div>{{ record.op_target_account || '--' }}</div>
                  <div v-if="targetSub(record)" class="cell-sub">
                    {{ targetSub(record) }}
                  </div>
                </template>
              </a-table-column>
              <a-table-column
                :title="t('agentLog.col.before')"
                data-index="before_value"
                :width="160"
                ellipsis
                tooltip
              />
              <a-table-column
                :title="t('agentLog.col.after')"
                data-index="after_value"
                :width="160"
                ellipsis
                tooltip
              />
              <a-table-column
                :title="t('agentLog.col.actionTime')"
                :width="180"
              >
                <template #cell="{ record }">
                  {{ formatUnix(record.operated_at) || '--' }}
                </template>
              </a-table-column>
              <a-table-column :title="t('agentLog.col.actionIp')" :width="180">
                <template #cell="{ record }">
                  <div>{{ record.client_ip || '--' }}</div>
                  <div v-if="record.client_ip_location" class="cell-sub">
                    {{ record.client_ip_location }}
                  </div>
                </template>
              </a-table-column>
              <a-table-column :title="t('agentLog.col.actionResult')">
                <template #cell="{ record }">
                  <span
                    :class="
                      record.result === AGENT_OP_RESULT.SUCCESS
                        ? 'result-success'
                        : 'result-invalid'
                    "
                  >
                    {{ record.result_label || '--' }}
                  </span>
                  <div v-if="record.fail_reason" class="cell-sub">
                    {{ record.fail_reason }}
                  </div>
                </template>
              </a-table-column>
            </template>
          </a-table>
          <div class="table-pagination">
            <a-pagination
              :current="actionPageNo"
              :page-size="actionPageSize"
              :page-size-options="[10, 20, 50, 100]"
              :total="actionTotal"
              show-total
              show-page-size
              @change="onActionPageChange"
              @page-size-change="onActionPageSizeChange"
            />
          </div>
        </a-tab-pane>
      </a-tabs>
    </a-card>
  </div>
</template>

<script lang="ts" setup>
  import { onMounted, reactive, ref, watch } from 'vue';
  import {
    siteDayEndSec,
    siteDayStartSec,
    siteNow,
  } from '@/utils/site-time';
  import { useI18n } from 'vue-i18n';
  import {
    AGENT_LOGIN_DEVICE,
    AGENT_LOGIN_KEYWORD,
    AGENT_LOGIN_RESULT_SUCCESS,
    AGENT_LOGIN_TYPE,
    AGENT_OP_RESULT,
    AGENT_OP_TARGET_TYPE,
    listAgentLoginLogs,
    listAgentOpLogs,
    type AgentLoginLogItem,
    type AgentLoginLogParams,
    type AgentOpLogItem,
    type AgentOpLogParams,
  } from '@/api/agent';
  import { formatUnix } from '../list/mapper';

  type QuickKey = 'today' | 'd7' | 'd30';

  const LOGIN_TYPE_OPTIONS = [
    { value: AGENT_LOGIN_TYPE.ALL, label: 'agentLog.search.pleaseSelect' },
    { value: AGENT_LOGIN_TYPE.LOGIN, label: 'agentLog.loginType.login' },
    { value: AGENT_LOGIN_TYPE.REGISTER, label: 'agentLog.loginType.register' },
  ];

  const DEVICE_TYPE_OPTIONS = [
    { value: AGENT_LOGIN_DEVICE.ALL, label: 'agentLog.search.pleaseSelect' },
    { value: AGENT_LOGIN_DEVICE.PC, label: 'agentLog.device.pc' },
    { value: AGENT_LOGIN_DEVICE.H5, label: 'agentLog.device.h5' },
    { value: AGENT_LOGIN_DEVICE.ANDROID, label: 'agentLog.device.android' },
    { value: AGENT_LOGIN_DEVICE.IOS, label: 'agentLog.device.ios' },
    { value: AGENT_LOGIN_DEVICE.PC_CLIENT, label: 'agentLog.device.pcClient' },
  ];

  const KEYWORD_TYPE_OPTIONS = [
    { value: AGENT_LOGIN_KEYWORD.AGENT_ID, label: 'agentLog.search.agentId' },
    { value: AGENT_LOGIN_KEYWORD.IP, label: 'agentLog.search.loginIp' },
    { value: AGENT_LOGIN_KEYWORD.UDID, label: 'agentLog.search.deviceCode' },
    {
      value: AGENT_LOGIN_KEYWORD.LOCATION,
      label: 'agentLog.search.loginLocation',
    },
  ];

  const TARGET_TYPE_OPTIONS = [
    { value: AGENT_OP_TARGET_TYPE.ALL, label: 'agentLog.search.pleaseSelect' },
    { value: AGENT_OP_TARGET_TYPE.AGENT, label: 'agentLog.targetType.agent' },
    { value: AGENT_OP_TARGET_TYPE.MEMBER, label: 'agentLog.targetType.member' },
  ];

  const ACTION_RESULT_OPTIONS = [
    { value: AGENT_OP_RESULT.ALL, label: 'agentLog.search.pleaseSelect' },
    { value: AGENT_OP_RESULT.SUCCESS, label: 'agentLog.actionResult.success' },
    { value: AGENT_OP_RESULT.FAIL, label: 'agentLog.actionResult.fail' },
  ];

  const { t } = useI18n();
  const activeTab = ref('login');
  const quickOffset: Record<QuickKey, number> = {
    today: 0,
    d7: 6,
    d30: 29,
  };

  function createLoginForm() {
    return {
      range: [] as string[],
      loginType: AGENT_LOGIN_TYPE.ALL as number,
      keywordType: AGENT_LOGIN_KEYWORD.AGENT_ID as number,
      keyword: '',
      deviceType: AGENT_LOGIN_DEVICE.ALL as number,
    };
  }

  const loginForm = reactive(createLoginForm());
  const loginQuick = ref<QuickKey | ''>('');
  const loginQuery = ref(createLoginForm());
  const loginPageNo = ref(1);
  const loginPageSize = ref(20);
  const loginRows = ref<AgentLoginLogItem[]>([]);
  const loginTotal = ref(0);
  const loginLoading = ref(false);
  let loginTicket = 0;

  function createActionForm() {
    return {
      range: [] as string[],
      targetType: AGENT_OP_TARGET_TYPE.ALL as number,
      targetId: undefined as number | undefined,
      result: AGENT_OP_RESULT.ALL as number,
    };
  }

  const actionForm = reactive(createActionForm());
  const actionQuick = ref<QuickKey | ''>('');
  const actionQuery = ref(createActionForm());
  const actionPageNo = ref(1);
  const actionPageSize = ref(20);
  const actionRows = ref<AgentOpLogItem[]>([]);
  const actionTotal = ref(0);
  const actionLoading = ref(false);
  const actionLoaded = ref(false);
  let actionTicket = 0;

  function operatorSub(record: AgentOpLogItem) {
    const id = record.admin_id ? `ID：${record.admin_id}` : '';
    return [record.admin_type_label, id].filter(Boolean).join(' · ');
  }

  function targetSub(record: AgentOpLogItem) {
    const id = record.op_target_id ? `ID：${record.op_target_id}` : '';
    return [record.op_target_type_label, id].filter(Boolean).join(' · ');
  }

  function buildLoginParams(): AgentLoginLogParams {
    const q = loginQuery.value;
    const params: AgentLoginLogParams = {
      page: loginPageNo.value,
      page_size: loginPageSize.value,
      login_type: q.loginType,
      device_type: q.deviceType,
    };
    if (q.range.length === 2) {
      params.time_start = siteDayStartSec(q.range[0]);
      params.time_end = siteDayEndSec(q.range[1]);
    }
    const keyword = q.keyword.trim();
    if (keyword) {
      params.keyword_type = q.keywordType;
      params.keyword = keyword;
    }
    return params;
  }

  async function loadLogin() {
    loginTicket += 1;
    const ticket = loginTicket;
    loginLoading.value = true;
    try {
      const { data } = await listAgentLoginLogs(buildLoginParams());
      if (ticket !== loginTicket) return;
      loginRows.value = data?.list || [];
      loginTotal.value = data?.total || 0;
    } catch {
      if (ticket !== loginTicket) return;
      loginRows.value = [];
      loginTotal.value = 0;
    } finally {
      if (ticket === loginTicket) loginLoading.value = false;
    }
  }

  function onLoginPageChange(page: number) {
    loginPageNo.value = page;
    loadLogin();
  }

  function onLoginPageSizeChange(size: number) {
    loginPageSize.value = size;
    loginPageNo.value = 1;
    loadLogin();
  }

  function buildActionParams(): AgentOpLogParams {
    const q = actionQuery.value;
    const params: AgentOpLogParams = {
      page: actionPageNo.value,
      page_size: actionPageSize.value,
      target_type: q.targetType,
      result: q.result,
    };
    if (q.targetId) params.target_id = q.targetId;
    if (q.range.length === 2) {
      params.time_start = siteDayStartSec(q.range[0]);
      params.time_end = siteDayEndSec(q.range[1]);
    }
    return params;
  }

  async function loadAction() {
    actionTicket += 1;
    const ticket = actionTicket;
    actionLoaded.value = true;
    actionLoading.value = true;
    try {
      const { data } = await listAgentOpLogs(buildActionParams());
      if (ticket !== actionTicket) return;
      actionRows.value = data?.list || [];
      actionTotal.value = data?.total || 0;
    } catch {
      if (ticket !== actionTicket) return;
      actionRows.value = [];
      actionTotal.value = 0;
    } finally {
      if (ticket === actionTicket) actionLoading.value = false;
    }
  }

  function onActionPageChange(page: number) {
    actionPageNo.value = page;
    loadAction();
  }

  function onActionPageSizeChange(size: number) {
    actionPageSize.value = size;
    actionPageNo.value = 1;
    loadAction();
  }

  function searchLogin() {
    loginQuery.value = { ...loginForm, range: [...loginForm.range] };
    loginPageNo.value = 1;
    loadLogin();
  }
  function resetLogin() {
    Object.assign(loginForm, createLoginForm());
    loginQuick.value = '';
    searchLogin();
  }
  function searchAction() {
    actionQuery.value = { ...actionForm, range: [...actionForm.range] };
    actionPageNo.value = 1;
    loadAction();
  }
  function resetAction() {
    Object.assign(actionForm, createActionForm());
    actionQuick.value = '';
    searchAction();
  }
  function applyQuick(kind: 'login' | 'action', key: QuickKey | '') {
    if (key !== 'today' && key !== 'd7' && key !== 'd30') return;
    const form = kind === 'login' ? loginForm : actionForm;
    form.range = [
      siteNow().subtract(quickOffset[key], 'day').format('YYYY-MM-DD'),
      siteNow().format('YYYY-MM-DD'),
    ];
    if (kind === 'login') searchLogin();
    else searchAction();
  }

  watch(loginQuick, (key) => applyQuick('login', key));
  watch(actionQuick, (key) => applyQuick('action', key));
  watch(activeTab, (tab) => {
    if (tab === 'action' && !actionLoaded.value) loadAction();
  });
  function onRangeChange(kind: 'login' | 'action') {
    const form = kind === 'login' ? loginForm : actionForm;
    const quick = kind === 'login' ? loginQuick : actionQuick;
    const end = siteNow().format('YYYY-MM-DD');
    const keys: QuickKey[] = ['today', 'd7', 'd30'];
    quick.value =
      keys.find((key) => {
        const start = siteNow()
          .subtract(quickOffset[key], 'day')
          .format('YYYY-MM-DD');
        return form.range[0] === start && form.range[1] === end;
      }) || '';
  }

  onMounted(loadLogin);
</script>

<style scoped lang="less">
  .container {
    padding: 0 20px 20px;
  }

  .general-card {
    margin-top: 16px;
  }

  .search-form :deep(.arco-form-item) {
    margin-bottom: 12px;
  }

  .search-form :deep(.arco-select-view) {
    min-width: 160px;
  }

  .search-type :deep(.arco-select-view) {
    min-width: 112px;
  }

  .search-input {
    width: 170px;
  }

  .search-form :deep(.range-picker) {
    width: 300px;
  }

  .cell-sub {
    color: var(--color-text-3);
    font-size: 12px;
    line-height: 18px;
  }

  .result-success {
    color: rgb(var(--success-6));
  }

  .result-invalid {
    color: rgb(var(--danger-6));
  }

  .table-pagination {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }
</style>
