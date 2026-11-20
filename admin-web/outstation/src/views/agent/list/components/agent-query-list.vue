<template>
  <div class="query-list">
    <a-form :model="form" layout="inline" class="search-form">
      <div class="search-row">
        <a-form-item>
          <a-space :size="12" align="center" wrap>
            <a-select v-model="form.timeFilterType" class="field-type">
              <a-option :value="1">
                {{ t('agentList.search.timeRegister') }}
              </a-option>
              <a-option :value="2">
                {{ t('agentList.search.timeLogin') }}
              </a-option>
            </a-select>
            <a-radio-group
              v-model="quickRange"
              class="quick-range"
              type="button"
              @change="onQuickRange"
            >
              <a-radio value="today">{{ t('agentList.search.today') }}</a-radio>
              <a-radio value="d7">{{ t('agentList.search.last7') }}</a-radio>
              <a-radio value="d30">{{ t('agentList.search.last30') }}</a-radio>
            </a-radio-group>
            <a-range-picker
              v-model="form.range"
              value-format="YYYY-MM-DD"
              allow-clear
              class="range-picker"
              :placeholder="[
                t('agentList.search.start'),
                t('agentList.search.end'),
              ]"
              @change="onRangeChange"
            />
          </a-space>
        </a-form-item>
        <a-form-item>
          <a-input-group compact class="keyword-search">
            <a-select v-model="form.agentFilterType" class="search-type">
              <a-option :value="1">
                {{ t('agentList.query.filter.self') }}
              </a-option>
              <a-option :value="2">
                {{ t('agentList.query.filter.parent') }}
              </a-option>
            </a-select>
            <a-input
              v-model="form.agentKeyword"
              class="search-input"
              allow-clear
              :placeholder="t('agentList.search.keywordAny')"
              @press-enter="handleSearch"
            />
          </a-input-group>
        </a-form-item>
        <a-form-item :label="t('agentList.search.level')">
          <a-select
            v-model="form.level"
            class="search-select"
            allow-clear
            :placeholder="t('agentList.search.all')"
          >
            <a-option
              v-for="item in OPEN_AGENT_LEVELS"
              :key="item.layer"
              :value="item.layer"
            >
              {{ item.name }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item :label="t('agentList.search.status')">
          <a-select
            v-model="form.status"
            class="search-select"
            allow-clear
            :placeholder="t('agentList.search.all')"
          >
            <a-option :value="AGENT_STATUS.NORMAL">
              {{ t('agentList.status.normal') }}
            </a-option>
            <a-option :value="AGENT_STATUS.DISABLED">
              {{ t('agentList.status.disabled') }}
            </a-option>
          </a-select>
        </a-form-item>
      </div>

      <div class="search-row">
        <a-form-item :label="t('agentList.search.withdrawLimit')">
          <a-select
            v-model="form.payoutRestricted"
            class="search-select"
            allow-clear
            :placeholder="t('agentList.search.all')"
          >
            <a-option :value="AGENT_PAYOUT.NONE">
              {{ t('agentList.withdraw.none') }}
            </a-option>
            <a-option :value="AGENT_PAYOUT.FORBIDDEN">
              {{ t('agentList.withdraw.forbidden') }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item>
          <a-space :size="8" align="center" wrap>
            <a-select
              v-model="form.numberFilterType"
              class="field-type number-type"
              allow-clear
              :placeholder="t('agentList.query.numberFilter')"
            >
              <a-option
                v-for="item in numberFilterOptions"
                :key="item.value"
                :value="item.value"
              >
                {{ t(item.label) }}
              </a-option>
            </a-select>
            <a-input-number
              v-model="form.numberMin"
              class="range-number"
              :min="0"
              :placeholder="t('agentList.search.min')"
            />
            <span class="range-sep">-</span>
            <a-input-number
              v-model="form.numberMax"
              class="range-number"
              :min="0"
              :placeholder="t('agentList.search.max')"
            />
          </a-space>
        </a-form-item>
        <a-form-item>
          <a-space :size="8">
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

    <a-table
      row-key="id"
      :data="rows"
      :loading="loading"
      :pagination="false"
      :bordered="{ cell: true }"
      :scroll="{ x: 2300 }"
    >
      <template #columns>
        <a-table-column
          :title="t('agentList.card.agentId')"
          data-index="id"
          :width="90"
          fixed="left"
        />
        <a-table-column
          :title="t('agentList.card.account')"
          data-index="account"
          :width="130"
          fixed="left"
        >
          <template #cell="{ record }">
            <a-link @click="emit('view', record)">{{ record.account }}</a-link>
          </template>
        </a-table-column>
        <!-- <a-table-column
          :title="t('agentList.card.level')"
          data-index="level"
          :width="100"
        /> -->
        <a-table-column
          :title="t('agentList.card.parent')"
          data-index="parentAgent"
          :width="120"
        >
          <template #cell="{ record }">
            {{ record.parentAgent || '--' }}
          </template>
        </a-table-column>
        <!-- <a-table-column
          :title="t('agentList.card.top')"
          data-index="topAgent"
          :width="120"
        >
          <template #cell="{ record }">
            {{ record.topAgent || '--' }}
          </template>
        </a-table-column> -->
        <a-table-column
          :title="t('agentList.card.directMembers')"
          data-index="directMembers"
          :width="100"
        />
        <a-table-column
          :title="t('agentList.card.directAgents')"
          data-index="directAgents"
          :width="100"
        />
        <a-table-column
          :title="t('agentList.card.allMembers')"
          data-index="allMembers"
          :width="100"
        />
        <a-table-column
          :title="t('agentList.card.allAgents')"
          data-index="allAgents"
          :width="100"
        />
        <a-table-column
          :title="t('agentList.card.share')"
          data-index="shareRatio"
          :width="100"
        >
          <template #cell="{ record }">
            {{ formatRatio(record.shareRatio) }}
          </template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.card.balance')"
          data-index="balance"
          :width="120"
        >
          <template #cell="{ record }">
            {{ formatAmount(record.balance) }}
          </template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.card.commission')"
          data-index="totalCommission"
          :width="120"
        >
          <template #cell="{ record }">
            {{ formatAmount(record.totalCommission) }}
          </template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.card.unpaid')"
          data-index="unpaidCommission"
          :width="120"
        >
          <template #cell="{ record }">
            {{ formatAmount(record.unpaidCommission) }}
          </template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.card.status')"
          data-index="status"
          :width="90"
        >
          <template #cell="{ record }">
            {{ t(`agentList.status.${record.status}`) }}
          </template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.card.withdrawLimit')"
          data-index="withdrawLimit"
          :width="120"
        >
          <template #cell="{ record }">
            {{ t(`agentList.withdraw.${record.withdrawLimit}`) }}
          </template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.card.uv')"
          data-index="visitUv"
          :width="100"
        />
        <a-table-column
          :title="t('agentList.card.registeredAt')"
          data-index="registeredAt"
          :width="170"
        >
          <template #cell="{ record }">
            {{ record.registeredAt || '--' }}
          </template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.card.lastLogin')"
          data-index="lastLoginAt"
          :width="170"
        >
          <template #cell="{ record }">
            {{ record.lastLoginAt || '--' }}
          </template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.card.remark')"
          data-index="remark"
          :width="140"
        >
          <template #cell="{ record }">
            {{ record.remark || '--' }}
          </template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.column.action')"
          :width="90"
          fixed="right"
        >
          <template #cell="{ record }">
            <a-button type="text" size="small" @click="emit('view', record)">
              {{ t('agentList.card.view') }}
            </a-button>
          </template>
        </a-table-column>
      </template>
      <template #empty>
        <a-empty :description="t('agentList.emptyList')" />
      </template>
    </a-table>

    <div class="query-pagination">
      <a-pagination
        :total="total"
        :current="pagination.current"
        :page-size="pagination.pageSize"
        :page-size-options="pagination.pageSizeOptions"
        show-total
        show-page-size
        @change="onPageChange"
        @page-size-change="onPageSizeChange"
      />
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { onMounted, reactive, ref } from 'vue';
  import {
    siteDayEndSec,
    siteDayStartSec,
    siteNow,
  } from '@/utils/site-time';
  import { useI18n } from 'vue-i18n';
  import {
    AGENT_PAYOUT,
    AGENT_STATUS,
    listAgents,
    type AgentListParams,
  } from '@/api/agent';
  import { formatAmount, formatRatio } from '../format';
  import { mapAgentListItem } from '../mapper';
  import { OPEN_AGENT_LEVELS } from '../options';
  import type { AgentRecord } from '../types';

  type QuickRange = 'today' | 'd7' | 'd30';

  const emit = defineEmits<{
    (e: 'view', record: AgentRecord): void;
  }>();

  const { t } = useI18n();
  const quickOffset: Record<QuickRange, number> = { today: 0, d7: 6, d30: 29 };

  const numberFilterOptions = [
    { value: 1, label: 'agentList.query.number.directMembers' },
    { value: 2, label: 'agentList.query.number.directAgents' },
    { value: 3, label: 'agentList.query.number.allAgents' },
    { value: 4, label: 'agentList.query.number.allMembers' },
    { value: 5, label: 'agentList.query.number.share' },
    { value: 6, label: 'agentList.query.number.commission' },
    { value: 7, label: 'agentList.query.number.withdrawn' },
  ] as const;

  function createForm() {
    return {
      timeFilterType: 1 as number,
      range: [] as string[],
      agentFilterType: 1 as number,
      agentKeyword: '',
      level: undefined as number | undefined,
      status: undefined as number | undefined,
      payoutRestricted: undefined as number | undefined,
      numberFilterType: undefined as number | undefined,
      numberMin: undefined as number | undefined,
      numberMax: undefined as number | undefined,
    };
  }

  const form = reactive(createForm());
  const quickRange = ref<QuickRange | ''>('');
  const loading = ref(false);
  const rows = ref<AgentRecord[]>([]);
  const total = ref(0);
  const pagination = reactive({
    current: 1,
    pageSize: 20,
    pageSizeOptions: [10, 20, 50, 100],
  });
  let ticket = 0;

  function buildParams(): AgentListParams {
    const params: AgentListParams = {
      page: pagination.current,
      page_size: pagination.pageSize,
      payout_restricted: form.payoutRestricted || 0,
    };
    if (form.level) params.level = form.level;
    if (form.status) params.status = form.status;
    const keyword = form.agentKeyword.trim();
    if (keyword) {
      params.agent_filter_type = form.agentFilterType || 1;
      params.agent_keyword = keyword;
    }
    if (form.range?.length === 2) {
      params.time_filter_type = form.timeFilterType || 1;
      params.time_start = siteDayStartSec(form.range[0]);
      params.time_end = siteDayEndSec(form.range[1]);
    }
    if (form.numberFilterType) {
      params.number_filter_type = form.numberFilterType;
      if (form.numberMin != null && form.numberMin > 0) {
        params.number_min = form.numberMin;
      }
      if (form.numberMax != null && form.numberMax > 0) {
        params.number_max = form.numberMax;
      }
    }
    return params;
  }

  async function loadList() {
    ticket += 1;
    const current = ticket;
    loading.value = true;
    try {
      const { data } = await listAgents(buildParams());
      if (current !== ticket) return;
      rows.value = (data?.list || []).map(mapAgentListItem);
      total.value = data?.total || 0;
    } catch {
      if (current !== ticket) return;
      rows.value = [];
      total.value = 0;
    } finally {
      if (current === ticket) loading.value = false;
    }
  }

  function handleSearch() {
    pagination.current = 1;
    loadList();
  }

  function handleReset() {
    Object.assign(form, createForm());
    quickRange.value = '';
    handleSearch();
  }

  function onQuickRange(value: string | number | boolean) {
    const key = String(value) as QuickRange;
    if (key !== 'today' && key !== 'd7' && key !== 'd30') return;
    const end = siteNow();
    form.range = [
      end.subtract(quickOffset[key], 'day').format('YYYY-MM-DD'),
      end.format('YYYY-MM-DD'),
    ];
    quickRange.value = key;
    handleSearch();
  }

  function onRangeChange() {
    const end = siteNow().format('YYYY-MM-DD');
    const keys: QuickRange[] = ['today', 'd7', 'd30'];
    quickRange.value =
      keys.find((key) => {
        const start = siteNow()
          .subtract(quickOffset[key], 'day')
          .format('YYYY-MM-DD');
        return form.range?.[0] === start && form.range?.[1] === end;
      }) || '';
  }

  function onPageChange(page: number) {
    pagination.current = page;
    loadList();
  }

  function onPageSizeChange(size: number) {
    pagination.pageSize = size;
    pagination.current = 1;
    loadList();
  }

  onMounted(loadList);
</script>

<style scoped lang="less">
  .query-list {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-width: 0;
    min-height: 0;
    padding-top: 16px;
    overflow: auto;
  }

  .search-form {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-bottom: 16px;
  }

  .search-form :deep(.arco-form-item) {
    flex: none;
    margin: 0;
  }

  .search-form :deep(.arco-form-item-label-col) {
    padding-right: 8px;
    line-height: 32px;
  }

  .search-row {
    display: flex;
    flex-wrap: wrap;
    gap: 12px 16px;
    align-items: center;
  }

  .field-type {
    width: 112px;

    :deep(.arco-select-view-single) {
      color: #fff;
      background: rgb(var(--primary-6));
      border-color: rgb(var(--primary-6));
    }

    :deep(.arco-select-view-value),
    :deep(.arco-select-view-suffix),
    :deep(.arco-select-view-icon) {
      color: #fff;
    }

    :deep(.arco-select-view-single:hover),
    :deep(.arco-select-view-single.arco-select-view-focus) {
      color: #fff;
      background: rgb(var(--primary-5));
      border-color: rgb(var(--primary-5));
    }
  }

  .number-type {
    width: 140px;
  }

  .quick-range {
    :deep(.arco-radio-group-button) {
      box-sizing: border-box;
      height: 32px;
      padding: 2px;
      line-height: 26px;
      background: var(--color-fill-2);
    }

    :deep(.arco-radio-button) {
      height: 26px;
      margin: 0;
      line-height: 26px;
    }

    :deep(.arco-radio-button-content) {
      padding: 0 14px;
      line-height: 26px;
    }
  }

  .range-picker {
    width: 280px;
  }

  .keyword-search {
    display: inline-flex;
    flex-wrap: nowrap;
    width: auto;
  }

  .search-type {
    flex: none;
    width: 88px;

    :deep(.arco-select-view) {
      min-width: 88px;
    }
  }

  .search-input {
    flex: none;
    width: 200px;
  }

  .search-select {
    width: 140px;
  }

  .range-number {
    width: 110px;
  }

  .range-sep {
    color: var(--color-text-3);
  }

  .query-list :deep(.arco-table-th),
  .query-list :deep(.arco-table-td) {
    white-space: nowrap;
  }

  .query-pagination {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }
</style>
