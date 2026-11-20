<template>
  <div class="container list-page">
    <a-card class="general-card" :bordered="false">
      <a-form :model="form" layout="inline" class="search-form">
        <div class="search-row">
          <a-form-item :label="t('relatedSearch.search.loginTime')">
            <a-space wrap>
              <a-radio-group v-model="quick" type="button" size="small">
                <a-radio value="today">
                  {{ t('relatedSearch.search.today') }}
                </a-radio>
                <a-radio value="d7">{{
                  t('relatedSearch.search.last7')
                }}</a-radio>
                <a-radio value="d30">
                  {{ t('relatedSearch.search.last30') }}
                </a-radio>
              </a-radio-group>
              <a-range-picker
                v-model="form.range"
                value-format="YYYY-MM-DD"
                allow-clear
                class="range-picker"
                :placeholder="[
                  t('relatedSearch.search.start'),
                  t('relatedSearch.search.end'),
                ]"
                @change="onRangeChange"
              />
            </a-space>
          </a-form-item>
        </div>
        <div class="search-row">
          <a-form-item>
            <a-input-group compact>
              <a-select v-model="form.agentField" class="search-type">
                <a-option value="agentId">
                  {{ t('relatedSearch.search.agentId') }}
                </a-option>
              </a-select>
              <a-input
                v-model="form.agentKeyword"
                class="search-input"
                allow-clear
                :placeholder="t('relatedSearch.search.pleaseInput')"
                @press-enter="handleSearch"
              />
            </a-input-group>
          </a-form-item>
          <a-form-item>
            <a-input-group compact>
              <a-select v-model="form.keywordType" class="search-type">
                <a-option
                  v-for="item in KEYWORD_TYPE_OPTIONS"
                  :key="item.value"
                  :value="item.value"
                >
                  {{ t(item.label) }}
                </a-option>
              </a-select>
              <a-input
                v-model="form.keyword"
                class="search-input"
                allow-clear
                :placeholder="t('relatedSearch.search.pleaseInput')"
                @press-enter="handleSearch"
              />
            </a-input-group>
          </a-form-item>
          <a-form-item :label="t('relatedSearch.search.device')">
            <a-select v-model="form.deviceType" class="search-select">
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
              <a-button type="primary" :loading="loading" @click="handleSearch">
                <template #icon><icon-search /></template>
                {{ t('front.common.search.search') }}
              </a-button>
              <a-button @click="handleReset">
                <template #icon><icon-refresh /></template>
                {{ t('front.common.search.reset') }}
              </a-button>
            </a-space>
          </a-form-item>
        </div>
      </a-form>

      <div class="table-toolbar">
        <a-space>
          <a-button type="outline" @click="batchDisable">
            <template #icon><icon-stop /></template>
            {{ t('relatedSearch.toolbar.batchDisable') }}
          </a-button>
          <a-button type="outline" @click="downloadReport">
            <template #icon><icon-download /></template>
            {{ t('relatedSearch.toolbar.download') }}
          </a-button>
        </a-space>
      </div>

      <a-table
        row-key="id"
        :data="rows"
        :loading="loading"
        :pagination="false"
        :bordered="{ cell: true }"
        :row-selection="rowSelection"
        :selected-keys="selectedKeys"
        :scroll="{ x: 1100 }"
        @selection-change="onSelectionChange"
      >
        <template #columns>
          <a-table-column
            :title="t('relatedSearch.col.loginTime')"
            :width="180"
          >
            <template #cell="{ record }">
              {{ formatUnix(record.login_at) || '--' }}
            </template>
          </a-table-column>
          <a-table-column :title="t('relatedSearch.col.agentId')" :width="140">
            <template #cell="{ record }">
              <div>{{ record.agent_id }}</div>
              <div v-if="record.account" class="cell-sub">
                {{ record.account }}
              </div>
            </template>
          </a-table-column>
          <a-table-column :title="t('relatedSearch.col.loginIp')" :width="160">
            <template #cell="{ record }">
              <div>{{ record.login_ip || '--' }}</div>
              <div v-if="record.login_location" class="cell-sub">
                {{ record.login_location }}
              </div>
            </template>
          </a-table-column>
          <a-table-column
            :title="t('relatedSearch.col.sameIp')"
            data-index="same_ip_agent_count"
            :width="120"
          />
          <a-table-column
            :title="t('relatedSearch.col.deviceCode')"
            :width="140"
          >
            <template #cell="{ record }">
              {{ record.udid || '--' }}
            </template>
          </a-table-column>
          <a-table-column :title="t('relatedSearch.col.device')" :width="180">
            <template #cell="{ record }">
              <div>{{ record.login_device || '--' }}</div>
              <div v-if="record.device_type_label" class="cell-sub">
                {{ record.device_type_label }}
              </div>
            </template>
          </a-table-column>
          <a-table-column
            :title="t('relatedSearch.col.sameDevice')"
            data-index="same_device_agent_count"
            :width="130"
          />
        </template>
      </a-table>
      <div class="table-pagination">
        <a-pagination
          :current="current"
          :page-size="pageSize"
          :total="total"
          :page-size-options="[10, 20, 50, 100]"
          show-total
          show-page-size
          @change="onPageChange"
          @page-size-change="onPageSizeChange"
        />
      </div>
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
  import { Message } from '@arco-design/web-vue';
  import {
    AGENT_ASSOC_KEYWORD,
    AGENT_LOGIN_DEVICE,
    listAgentLoginAssociations,
    type AgentLoginAssocItem,
    type AgentLoginAssocParams,
  } from '@/api/agent';
  import { formatUnix } from '../list/mapper';

  type QuickKey = 'today' | 'd7' | 'd30';

  const KEYWORD_TYPE_OPTIONS = [
    { value: AGENT_ASSOC_KEYWORD.IP, label: 'relatedSearch.search.loginIp' },
    {
      value: AGENT_ASSOC_KEYWORD.UDID,
      label: 'relatedSearch.search.deviceCode',
    },
  ];

  const DEVICE_TYPE_OPTIONS = [
    {
      value: AGENT_LOGIN_DEVICE.ALL,
      label: 'relatedSearch.search.pleaseSelect',
    },
    { value: AGENT_LOGIN_DEVICE.PC, label: 'relatedSearch.device.pc' },
    { value: AGENT_LOGIN_DEVICE.H5, label: 'relatedSearch.device.h5' },
    {
      value: AGENT_LOGIN_DEVICE.ANDROID,
      label: 'relatedSearch.device.android',
    },
    { value: AGENT_LOGIN_DEVICE.IOS, label: 'relatedSearch.device.ios' },
    {
      value: AGENT_LOGIN_DEVICE.PC_CLIENT,
      label: 'relatedSearch.device.pcClient',
    },
  ];

  function createForm() {
    return {
      agentField: 'agentId',
      range: [] as string[],
      agentKeyword: '',
      keywordType: AGENT_ASSOC_KEYWORD.IP as number,
      keyword: '',
      deviceType: AGENT_LOGIN_DEVICE.ALL as number,
    };
  }

  const { t } = useI18n();
  const quickKeys: QuickKey[] = ['today', 'd7', 'd30'];
  const quickOffset: Record<QuickKey, number> = {
    today: 0,
    d7: 6,
    d30: 29,
  };
  const form = reactive(createForm());
  const quick = ref<QuickKey | ''>('');
  const query = ref(createForm());
  const current = ref(1);
  const pageSize = ref(20);
  const rows = ref<AgentLoginAssocItem[]>([]);
  const total = ref(0);
  const loading = ref(false);
  let loadTicket = 0;
  const selectedKeys = ref<(string | number)[]>([]);
  const rowSelection = {
    type: 'checkbox' as const,
    showCheckedAll: true,
  };

  function buildParams(): AgentLoginAssocParams {
    const q = query.value;
    const params: AgentLoginAssocParams = {
      page: current.value,
      page_size: pageSize.value,
      device_type: q.deviceType,
    };
    if (q.range.length === 2) {
      params.time_start = siteDayStartSec(q.range[0]);
      params.time_end = siteDayEndSec(q.range[1]);
    }
    const agentId = q.agentKeyword.trim();
    if (agentId) params.agent_id = Number(agentId);
    const keyword = q.keyword.trim();
    if (keyword) {
      params.keyword_type = q.keywordType;
      params.keyword = keyword;
    }
    return params;
  }

  async function loadList() {
    loadTicket += 1;
    const ticket = loadTicket;
    loading.value = true;
    selectedKeys.value = [];
    try {
      const { data } = await listAgentLoginAssociations(buildParams());
      if (ticket !== loadTicket) return;
      rows.value = data?.list || [];
      total.value = data?.total || 0;
    } catch {
      if (ticket !== loadTicket) return;
      rows.value = [];
      total.value = 0;
    } finally {
      if (ticket === loadTicket) loading.value = false;
    }
  }

  function onPageChange(page: number) {
    current.value = page;
    loadList();
  }
  function onPageSizeChange(size: number) {
    pageSize.value = size;
    current.value = 1;
    loadList();
  }

  function handleSearch() {
    const agentId = form.agentKeyword.trim();
    if (agentId && !/^\d+$/.test(agentId)) {
      Message.warning(t('relatedSearch.message.invalidAgentId'));
      return;
    }
    query.value = { ...form, range: [...form.range] };
    current.value = 1;
    loadList();
  }
  function handleReset() {
    Object.assign(form, createForm());
    quick.value = '';
    handleSearch();
  }
  function onRangeChange() {
    const end = siteNow().format('YYYY-MM-DD');
    quick.value =
      quickKeys.find((key) => {
        const start = siteNow()
          .subtract(quickOffset[key], 'day')
          .format('YYYY-MM-DD');
        return form.range[0] === start && form.range[1] === end;
      }) || '';
  }
  function onSelectionChange(keys: (string | number)[]) {
    selectedKeys.value = keys;
  }
  function batchDisable() {
    if (!selectedKeys.value.length) {
      Message.warning(t('relatedSearch.message.noSelection'));
      return;
    }
    Message.success(t('relatedSearch.message.disabled'));
  }
  function downloadReport() {
    if (!rows.value.length) {
      Message.warning(t('relatedSearch.message.noData'));
    }
  }

  watch(quick, (key) => {
    if (key !== 'today' && key !== 'd7' && key !== 'd30') return;
    form.range = [
      siteNow().subtract(quickOffset[key], 'day').format('YYYY-MM-DD'),
      siteNow().format('YYYY-MM-DD'),
    ];
  });

  onMounted(loadList);
</script>

<style scoped lang="less">
  .container {
    padding: 0 20px 20px;
  }

  .general-card {
    margin-top: 16px;
  }

  .general-card :deep(.arco-card-body) {
    padding-top: 20px;
  }

  .search-form {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .search-row {
    display: flex;
    flex-wrap: wrap;
    gap: 8px 16px;
    align-items: center;
  }

  .search-row :deep(.arco-form-item) {
    margin: 0 0 8px;
  }

  .search-type :deep(.arco-select-view) {
    width: 112px;
    min-width: 112px;
  }

  .search-select {
    width: 160px;
  }

  .search-input {
    width: 180px;
  }

  .search-form :deep(.range-picker) {
    flex: none;
    width: 300px;
  }

  .table-toolbar {
    margin-bottom: 12px;
  }

  .cell-sub {
    color: var(--color-text-3);
    font-size: 12px;
    line-height: 18px;
  }

  .table-pagination {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }
</style>
