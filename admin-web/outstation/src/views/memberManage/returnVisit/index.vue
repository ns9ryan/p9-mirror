<template>
  <div class="container list-page">
    <a-card class="general-card search-card" :bordered="false">
      <a-form :model="form" layout="inline" class="search-form">
        <div class="search-fields">
          <div class="search-row">
            <a-form-item>
              <a-space :size="12" align="center" wrap>
                <a-select v-model="form.dateField" class="field-type">
                  <a-option value="stat">
                    {{ t('memberReturnVisit.search.statDate') }}
                  </a-option>
                  <a-option value="register">
                    {{ t('memberReturnVisit.search.registerTime') }}
                  </a-option>
                  <a-option value="login">
                    {{ t('memberReturnVisit.search.loginTime') }}
                  </a-option>
                </a-select>
                <a-radio-group
                  v-model="quickRange"
                  class="quick-range"
                  type="button"
                >
                  <a-radio value="today">
                    {{ t('memberReturnVisit.search.today') }}
                  </a-radio>
                  <a-radio value="d7">
                    {{ t('memberReturnVisit.search.last7') }}
                  </a-radio>
                  <a-radio value="d30">
                    {{ t('memberReturnVisit.search.last30') }}
                  </a-radio>
                </a-radio-group>
                <a-range-picker
                  v-model="form.range"
                  class="range-picker"
                  value-format="YYYY-MM-DD"
                  allow-clear
                  :placeholder="[
                    t('memberReturnVisit.search.start'),
                    t('memberReturnVisit.search.end'),
                  ]"
                  @change="onRangeChange"
                />
              </a-space>
            </a-form-item>
            <a-form-item :label="t('memberReturnVisit.search.autoLevel')">
              <a-select
                v-model="form.autoLevel"
                class="search-select"
                allow-clear
                :placeholder="t('memberReturnVisit.search.pleaseSelect')"
              >
                <a-option v-for="item in AUTO_LEVELS" :key="item" :value="item">
                  {{ item }}
                </a-option>
              </a-select>
            </a-form-item>
            <a-form-item :label="t('memberReturnVisit.search.track')">
              <a-select
                v-model="form.track"
                class="search-select"
                allow-clear
                :placeholder="t('memberReturnVisit.search.pleaseSelect')"
              >
                <a-option value="notDeposit">
                  {{ t('memberReturnVisit.track.notDeposit') }}
                </a-option>
                <a-option value="hasBalance">
                  {{ t('memberReturnVisit.track.hasBalance') }}
                </a-option>
                <a-option value="negativeDiff">
                  {{ t('memberReturnVisit.track.negativeDiff') }}
                </a-option>
                <a-option value="positiveDiff">
                  {{ t('memberReturnVisit.track.positiveDiff') }}
                </a-option>
              </a-select>
            </a-form-item>
            <a-form-item :label="t('memberReturnVisit.search.channel')">
              <a-select
                v-model="form.channel"
                class="search-select"
                allow-clear
                :placeholder="t('memberReturnVisit.search.pleaseSelect')"
              >
                <a-option
                  v-for="item in CHANNEL_PRESETS"
                  :key="item"
                  :value="item"
                >
                  {{ item }}
                </a-option>
              </a-select>
            </a-form-item>
          </div>
          <div class="search-row">
            <a-form-item>
              <a-space :size="8" align="center">
                <a-select v-model="form.amountField" class="field-type">
                  <a-option
                    v-for="item in amountFields"
                    :key="item"
                    :value="item"
                  >
                    {{ t(`memberReturnVisit.amount.${item}`) }}
                  </a-option>
                </a-select>
                <a-input
                  v-model="form.amountMin"
                  class="amount-input"
                  allow-clear
                  :placeholder="t('memberReturnVisit.search.min')"
                  @press-enter="handleSearch"
                />
                <span class="amount-sep">-</span>
                <a-input
                  v-model="form.amountMax"
                  class="amount-input"
                  allow-clear
                  :placeholder="t('memberReturnVisit.search.max')"
                  @press-enter="handleSearch"
                />
              </a-space>
            </a-form-item>
            <a-form-item>
              <a-input-group class="keyword-search">
                <a-select v-model="form.keywordField" class="field-type">
                  <a-option value="offline">
                    {{ t('memberReturnVisit.search.offline') }}
                  </a-option>
                  <a-option value="memberId">
                    {{ t('memberReturnVisit.search.memberId') }}
                  </a-option>
                </a-select>
                <a-input
                  v-model="form.keyword"
                  class="search-input"
                  allow-clear
                  :placeholder="t('memberReturnVisit.search.pleaseInput')"
                  @press-enter="handleSearch"
                />
              </a-input-group>
            </a-form-item>
            <a-form-item class="search-actions">
              <a-space :size="12">
                <a-button type="primary" @click="handleSearch">
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
        </div>
      </a-form>
    </a-card>

    <a-card class="general-card table-card" :bordered="false">
      <a-table
        row-key="id"
        :data="pageRows"
        :pagination="false"
        :bordered="{ cell: true }"
        :stripe="true"
        :scroll="{ x: 1680 }"
        @sorter-change="onSorterChange"
      >
        <template #columns>
          <a-table-column
            :title="t('memberReturnVisit.column.memberId')"
            data-index="memberId"
            :width="120"
          />
          <a-table-column
            :title="t('memberReturnVisit.column.level')"
            data-index="level"
            :width="140"
          />
          <a-table-column
            :title="t('memberReturnVisit.column.vip')"
            data-index="vipLevel"
            :width="100"
          />
          <a-table-column
            :title="t('memberReturnVisit.column.channel')"
            :width="120"
          >
            <template #cell="{ record }">
              {{ record.channel || '--' }}
            </template>
          </a-table-column>
          <a-table-column
            :title="t('memberReturnVisit.column.registeredAt')"
            data-index="registeredAt"
            :width="180"
          />
          <a-table-column
            :title="t('memberReturnVisit.column.loginAt')"
            data-index="loginAt"
            :width="180"
          />
          <a-table-column
            v-for="column in moneyColumns"
            :key="column.key"
            :title="t(`memberReturnVisit.column.${column.key}`)"
            :data-index="column.key"
            :width="110"
            :sortable="column.sortable ? numberSort : undefined"
          >
            <template #cell="{ record }">
              <span class="money">{{ formatMoney(record[column.key]) }}</span>
            </template>
          </a-table-column>
        </template>
        <template #empty>
          <a-empty :description="t('memberReturnVisit.empty')" />
        </template>
      </a-table>
      <div class="table-pagination">
        <a-pagination
          v-model:current="pageNo"
          v-model:page-size="pageSize"
          :total="filtered.length"
          :page-size-options="[15, 20, 30, 50, 100]"
          show-total
          show-page-size
          show-jumper
        />
      </div>
    </a-card>
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import dayjs from 'dayjs';
  import {
    siteDayStartMs,
    siteDayjs,
    siteNextDayStartMs,
    siteNow,
  } from '@/utils/site-time';
  import { useI18n } from 'vue-i18n';
  import { AUTO_LEVELS, CHANNEL_PRESETS } from '../memberList/options';

  type DateField = 'stat' | 'register' | 'login';
  type AmountField =
    | 'diff'
    | 'balance'
    | 'manualDeposit'
    | 'orderDeposit'
    | 'gift'
    | 'orderWithdraw'
    | 'manualWithdraw'
    | 'historyDeposit';
  type KeywordField = 'offline' | 'memberId';
  type TrackCondition =
    | 'notDeposit'
    | 'hasBalance'
    | 'negativeDiff'
    | 'positiveDiff';
  type QuickKey = 'today' | 'd7' | 'd30';
  type SortDirection = 'ascend' | 'descend' | '';

  interface VisitRecord {
    id: number;
    memberId: string;
    level: string;
    vipLevel: number;
    channel: string;
    registeredAt: string;
    loginAt: string;
    statAt: string;
    balance: number;
    manualDeposit: number;
    orderDeposit: number;
    gift: number;
    orderWithdraw: number;
    manualWithdraw: number;
    diff: number;
    historyDeposit: number;
  }

  interface QueryState {
    dateField: DateField;
    range: string[];
    autoLevel?: string;
    track?: TrackCondition;
    channel?: string;
    amountField: AmountField;
    amountMin: string;
    amountMax: string;
    keywordField: KeywordField;
    keyword: string;
  }

  const { t } = useI18n();
  const quickOffset: Record<QuickKey, number> = {
    today: 0,
    d7: 6,
    d30: 29,
  };
  const amountFields: AmountField[] = [
    'diff',
    'balance',
    'manualDeposit',
    'orderDeposit',
    'gift',
    'orderWithdraw',
    'manualWithdraw',
    'historyDeposit',
  ];
  const moneyColumns: { key: AmountField; sortable?: boolean }[] = [
    { key: 'balance', sortable: true },
    { key: 'manualDeposit' },
    { key: 'orderDeposit' },
    { key: 'gift' },
    { key: 'orderWithdraw' },
    { key: 'manualWithdraw' },
    { key: 'diff' },
    { key: 'historyDeposit' },
  ];
  const numberSort = {
    sortDirections: ['ascend', 'descend'] as ('ascend' | 'descend')[],
    sorter: true,
  };

  const records = ref<VisitRecord[]>([
    {
      id: 1,
      memberId: '31592197',
      level: '默认层级',
      vipLevel: 0,
      channel: '',
      registeredAt: '2026-08-18 21:22:42',
      loginAt: '2026-08-18 21:22:42',
      statAt: '2026-08-18 21:22:42',
      balance: 0.1,
      manualDeposit: 0,
      orderDeposit: 0,
      gift: 0,
      orderWithdraw: 0,
      manualWithdraw: 0,
      diff: 0,
      historyDeposit: 0,
    },
  ]);

  const quickRange = ref<QuickKey | ''>('');
  const form = reactive({
    dateField: 'stat' as DateField,
    range: [] as string[],
    autoLevel: undefined as string | undefined,
    track: undefined as TrackCondition | undefined,
    channel: undefined as string | undefined,
    amountField: 'diff' as AmountField,
    amountMin: '',
    amountMax: '',
    keywordField: 'offline' as KeywordField,
    keyword: '',
  });
  const query = ref<QueryState>({
    dateField: 'stat',
    range: [],
    autoLevel: undefined,
    track: undefined,
    channel: undefined,
    amountField: 'diff',
    amountMin: '',
    amountMax: '',
    keywordField: 'offline',
    keyword: '',
  });
  const pageNo = ref(1);
  const pageSize = ref(15);
  const sortField = ref<AmountField | ''>('');
  const sortDirection = ref<SortDirection>('');

  function formatMoney(value: number) {
    return value.toFixed(2);
  }

  function rowTime(item: VisitRecord, field: DateField) {
    if (field === 'register') return item.registeredAt;
    if (field === 'login') return item.loginAt;
    return item.statAt;
  }

  function inRange(time: string, range: string[]) {
    if (range.length !== 2 || !range[0] || !range[1]) return true;
    // 按分站时区的整天区间，左闭右开
    const value = dayjs(time).valueOf();
    return (
      value >= siteDayStartMs(range[0]) && value < siteNextDayStartMs(range[1])
    );
  }

  function matchAmount(value: number, minText: string, maxText: string) {
    const min = minText.trim();
    const max = maxText.trim();
    if (min && value < Number(min)) return false;
    if (max && value > Number(max)) return false;
    return true;
  }

  function matchKeyword(item: VisitRecord, current: QueryState) {
    const keyword = current.keyword.trim();
    if (!keyword) return true;
    if (current.keywordField === 'memberId') {
      return item.memberId.includes(keyword);
    }
    const days = Number(keyword);
    if (Number.isNaN(days)) return true;
    return (
      siteNow().startOf('day').diff(siteDayjs(item.loginAt).startOf('day'), 'day') >=
      days
    );
  }

  function matchTrack(item: VisitRecord, track?: TrackCondition) {
    if (track === 'notDeposit') return item.historyDeposit <= 0;
    if (track === 'hasBalance') return item.balance > 0;
    if (track === 'negativeDiff') return item.diff < 0;
    if (track === 'positiveDiff') return item.diff > 0;
    return true;
  }

  const filtered = computed(() => {
    const current = query.value;
    const list = records.value.filter((item) => {
      if (current.autoLevel && item.level !== current.autoLevel) return false;
      if (current.channel && item.channel !== current.channel) return false;
      if (!matchTrack(item, current.track)) return false;
      if (!inRange(rowTime(item, current.dateField), current.range)) {
        return false;
      }
      if (
        !matchAmount(
          item[current.amountField],
          current.amountMin,
          current.amountMax
        )
      ) {
        return false;
      }
      return matchKeyword(item, current);
    });
    if (!sortField.value || !sortDirection.value) return list;
    const factor = sortDirection.value === 'ascend' ? 1 : -1;
    const field = sortField.value;
    return [...list].sort((a, b) => (a[field] - b[field]) * factor);
  });

  const pageRows = computed(() =>
    filtered.value.slice(
      (pageNo.value - 1) * pageSize.value,
      pageNo.value * pageSize.value
    )
  );

  watch([() => filtered.value.length, pageSize], () => {
    const maxPage = Math.max(
      1,
      Math.ceil(filtered.value.length / pageSize.value) || 1
    );
    if (pageNo.value > maxPage) pageNo.value = maxPage;
  });

  function buildQuickRange(key: QuickKey): [string, string] {
    const end = siteNow().format('YYYY-MM-DD');
    return [
      siteNow().subtract(quickOffset[key], 'day').format('YYYY-MM-DD'),
      end,
    ];
  }

  function snapshot(): QueryState {
    return {
      dateField: form.dateField,
      range: [...(form.range || [])],
      autoLevel: form.autoLevel,
      track: form.track,
      channel: form.channel,
      amountField: form.amountField,
      amountMin: form.amountMin,
      amountMax: form.amountMax,
      keywordField: form.keywordField,
      keyword: form.keyword,
    };
  }

  function handleSearch() {
    query.value = snapshot();
    pageNo.value = 1;
  }

  function handleReset() {
    form.dateField = 'stat';
    form.range = [];
    form.autoLevel = undefined;
    form.track = undefined;
    form.channel = undefined;
    form.amountField = 'diff';
    form.amountMin = '';
    form.amountMax = '';
    form.keywordField = 'offline';
    form.keyword = '';
    quickRange.value = '';
    sortField.value = '';
    sortDirection.value = '';
    handleSearch();
  }

  function applyQuick(key: QuickKey | '') {
    if (key !== 'today' && key !== 'd7' && key !== 'd30') return;
    form.range = [...buildQuickRange(key)];
    handleSearch();
  }

  watch(quickRange, (key) => applyQuick(key));

  function onRangeChange() {
    const range = form.range || [];
    const keys: QuickKey[] = ['today', 'd7', 'd30'];
    quickRange.value =
      keys.find((key) => {
        const [start, end] = buildQuickRange(key);
        return range[0] === start && range[1] === end;
      }) || '';
  }

  function onSorterChange(dataIndex: string, direction: string) {
    const field = amountFields.find((item) => item === dataIndex);
    if (!field || (direction !== 'ascend' && direction !== 'descend')) {
      sortField.value = '';
      sortDirection.value = '';
      return;
    }
    sortField.value = field;
    sortDirection.value = direction;
  }
</script>

<style scoped lang="less">
  .container {
    padding: 0 20px 20px;
  }

  .general-card {
    margin-top: 16px;
  }

  .search-card :deep(.arco-card-body) {
    padding: 16px 20px;
  }

  .table-card :deep(.arco-card-body) {
    padding: 20px;
  }

  .search-form {
    width: 100%;
  }

  .search-form :deep(.arco-form-item) {
    margin: 0;
  }

  .search-form :deep(.arco-form-item-label-col) {
    padding-right: 8px;
    line-height: 32px;
  }

  .search-fields {
    display: flex;
    flex-direction: column;
    gap: 12px;
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

  .keyword-search {
    display: inline-flex;
  }

  .search-input {
    width: 180px;
  }

  .amount-input {
    width: 108px;
  }

  .amount-sep {
    color: var(--color-text-3);
  }

  .search-select {
    width: 148px;
  }

  .range-picker {
    width: 280px;
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

  .money {
    color: rgb(var(--primary-6));
  }

  .table-pagination {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }

  :deep(.arco-pagination-options .arco-select) {
    min-width: 100px;
  }
</style>
