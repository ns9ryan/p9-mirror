<template>
  <div class="container list-page">
    <a-card class="general-card search-card" :bordered="false">
      <a-form :model="form" layout="inline" class="search-form">
        <div class="search-fields">
          <a-form-item>
            <a-space :size="12" align="center" wrap>
              <a-select v-model="form.timeField" class="time-type">
                <a-option value="loginAt">
                  {{ t('memberLoginLog.search.loginTime') }}
                </a-option>
              </a-select>
              <a-radio-group
                v-model="quickRange"
                class="quick-range"
                type="button"
              >
                <a-radio value="today">
                  {{ t('memberLoginLog.search.today') }}
                </a-radio>
                <a-radio value="d7">
                  {{ t('memberLoginLog.search.last7') }}
                </a-radio>
                <a-radio value="d30">
                  {{ t('memberLoginLog.search.last30') }}
                </a-radio>
              </a-radio-group>
              <a-range-picker
                v-model="form.range"
                class="range-picker"
                value-format="YYYY-MM-DD"
                allow-clear
                :placeholder="[
                  t('memberLoginLog.search.start'),
                  t('memberLoginLog.search.end'),
                ]"
                @change="onRangeChange"
              />
            </a-space>
          </a-form-item>
          <a-form-item :label="t('memberLoginLog.search.loginType')">
            <a-select
              v-model="form.loginType"
              class="search-select"
              allow-clear
              :placeholder="t('memberLoginLog.search.pleaseSelect')"
            >
              <a-option value="login">
                {{ t('memberLoginLog.loginType.login') }}
              </a-option>
            </a-select>
          </a-form-item>
          <a-form-item>
            <a-input-group class="id-search">
              <a-select v-model="form.keywordType" class="id-type">
                <a-option value="memberId">
                  {{ t('memberLoginLog.search.memberId') }}
                </a-option>
              </a-select>
              <a-input
                v-model="form.keyword"
                class="search-input"
                allow-clear
                :placeholder="t('memberLoginLog.search.pleaseInput')"
                @press-enter="handleSearch"
              />
            </a-input-group>
          </a-form-item>
          <a-form-item :label="t('memberLoginLog.search.device')">
            <a-select
              v-model="form.device"
              class="search-select"
              allow-clear
              :placeholder="t('memberLoginLog.search.pleaseSelect')"
            >
              <a-option value="h5">
                {{ t('memberLoginLog.device.h5') }}
              </a-option>
              <a-option value="pc">
                {{ t('memberLoginLog.device.pc') }}
              </a-option>
              <a-option value="app">
                {{ t('memberLoginLog.device.app') }}
              </a-option>
            </a-select>
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
      </a-form>
    </a-card>

    <a-card class="general-card table-card" :bordered="false">
      <a-table
        row-key="id"
        :data="pageRows"
        :pagination="false"
        :bordered="{ cell: true }"
        :stripe="true"
        :scroll="{ x: 1280 }"
      >
        <template #columns>
          <a-table-column
            :title="t('memberLoginLog.column.memberId')"
            data-index="memberId"
            :width="120"
          />
          <a-table-column
            :title="t('memberLoginLog.column.loginType')"
            :width="110"
            align="center"
          >
            <template #cell="{ record }">
              {{ t(`memberLoginLog.loginType.${record.loginType}`) }}
            </template>
          </a-table-column>
          <a-table-column :title="t('memberLoginLog.column.ip')" :width="200">
            <template #cell="{ record }">
              <div class="stack-cell">
                <div>{{ record.ip }}</div>
                <div class="hit-count">({{ record.ipCount }})</div>
              </div>
            </template>
          </a-table-column>
          <a-table-column
            :title="t('memberLoginLog.column.deviceCode')"
            :width="260"
          >
            <template #cell="{ record }">
              <div class="stack-cell device-code">
                <div>{{ record.deviceCode }}</div>
                <div class="hit-count">({{ record.deviceCount }})</div>
              </div>
            </template>
          </a-table-column>
          <a-table-column
            :title="t('memberLoginLog.column.browser')"
            :width="120"
          >
            <template #cell="{ record }">
              <div class="stack-cell">
                <div>{{ record.browser }}</div>
                <div>{{ record.version }}</div>
              </div>
            </template>
          </a-table-column>
          <a-table-column
            :title="t('memberLoginLog.column.device')"
            :width="120"
            align="center"
          >
            <template #cell="{ record }">
              {{ t(`memberLoginLog.device.${record.device}`) }}
            </template>
          </a-table-column>
          <a-table-column
            :title="t('memberLoginLog.column.loginAt')"
            data-index="loginAt"
            :width="180"
          />
          <a-table-column
            :title="t('memberLoginLog.column.result')"
            :width="140"
            align="center"
          >
            <template #cell="{ record }">
              <div
                :class="
                  record.result === 'success'
                    ? 'result-success'
                    : 'result-failed'
                "
              >
                <div>{{ t(`memberLoginLog.result.${record.result}`) }}</div>
                <div v-if="record.failReason">{{ record.failReason }}</div>
              </div>
            </template>
          </a-table-column>
        </template>
        <template #empty>
          <a-empty :description="t('memberLoginLog.empty')" />
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
    siteNextDayStartMs,
    siteNow,
  } from '@/utils/site-time';
  import { useI18n } from 'vue-i18n';

  type LoginType = 'login';
  type DeviceType = 'h5' | 'pc' | 'app';
  type LoginResult = 'success' | 'failed';
  type QuickKey = 'today' | 'd7' | 'd30';

  interface LoginRecord {
    id: number;
    memberId: string;
    loginType: LoginType;
    ip: string;
    ipCount: number;
    deviceCode: string;
    deviceCount: number;
    browser: string;
    version: string;
    device: DeviceType;
    loginAt: string;
    result: LoginResult;
    failReason?: string;
  }

  interface QueryState {
    range: string[];
    keyword: string;
    loginType?: LoginType;
    device?: DeviceType;
  }

  const { t } = useI18n();
  const quickOffset: Record<QuickKey, number> = {
    today: 0,
    d7: 6,
    d30: 29,
  };

  const records = ref<LoginRecord[]>([
    {
      id: 8,
      memberId: '60569646',
      loginType: 'login',
      ip: '38.54.14.128',
      ipCount: 252,
      deviceCode:
        'f93a83d914022cfc83fe9e53570cba1f320c63850c84a2614497a2e05a496d1',
      deviceCount: 15,
      browser: 'Chrome',
      version: '155.0.0',
      device: 'h5',
      loginAt: '2026-10-14 13:42:50',
      result: 'success',
    },
    {
      id: 7,
      memberId: '80404226',
      loginType: 'login',
      ip: '38.54.14.128',
      ipCount: 252,
      deviceCode:
        '5189a76445a5267c0631d6178a71fb7693878a1cf56764be58512b0415de7f',
      deviceCount: 1,
      browser: 'Chrome',
      version: '155.0.0',
      device: 'h5',
      loginAt: '2026-10-14 13:07:41',
      result: 'success',
    },
    {
      id: 6,
      memberId: '73494799',
      loginType: 'login',
      ip: '38.54.14.128',
      ipCount: 252,
      deviceCode:
        '860aa9c69e52a373c225c708933e0e948645801e381a269689077b7c150f8830',
      deviceCount: 2,
      browser: 'Chrome',
      version: '154.0.0',
      device: 'h5',
      loginAt: '2026-10-14 11:37:21',
      result: 'success',
    },
    {
      id: 5,
      memberId: '23508609',
      loginType: 'login',
      ip: '38.54.14.128',
      ipCount: 252,
      deviceCode:
        '860aa9c69e52a373c225c708933e0e948645801e381a269689077b7c150f8830',
      deviceCount: 2,
      browser: 'Chrome',
      version: '154.0.0',
      device: 'h5',
      loginAt: '2026-10-14 11:17:21',
      result: 'success',
    },
    {
      id: 4,
      memberId: '54449040',
      loginType: 'login',
      ip: '2001:f40:944:47a5:d144:32db:6c06:7c11',
      ipCount: 7,
      deviceCode:
        '3058e488229b9d3940cb7a8ae769c3ac8562c70a23a9d041ab0907d3ce6931',
      deviceCount: 1,
      browser: 'Safari',
      version: '26.5.1',
      device: 'h5',
      loginAt: '2026-10-14 11:16:37',
      result: 'success',
    },
    {
      id: 3,
      memberId: '22495618',
      loginType: 'login',
      ip: '38.54.14.128',
      ipCount: 252,
      deviceCode:
        '860aa9c69e52a373c225c708933e0e948645801e381a269689077b7c150f8830',
      deviceCount: 2,
      browser: 'Chrome',
      version: '154.0.0',
      device: 'h5',
      loginAt: '2026-10-14 11:16:29',
      result: 'success',
    },
    {
      id: 2,
      memberId: '46424302',
      loginType: 'login',
      ip: '103.117.102.220',
      ipCount: 71,
      deviceCode:
        '863a5c601867adde74efe0309765ead5fa50ccd375ea77038096ae70b961468',
      deviceCount: 1,
      browser: 'Chrome',
      version: '155.0.0',
      device: 'h5',
      loginAt: '2026-10-14 11:16:02',
      result: 'success',
    },
    {
      id: 1,
      memberId: '-1',
      loginType: 'login',
      ip: '38.54.14.128',
      ipCount: 252,
      deviceCode:
        'cb29fca824cdc82b51f0256facfb31895d597014bc844900c5779fe70221f141e7',
      deviceCount: 1,
      browser: 'Safari',
      version: '18.5',
      device: 'h5',
      loginAt: '2026-10-14 11:13:42',
      result: 'failed',
      failReason: 'Unknown (142)',
    },
  ]);

  const quickRange = ref<QuickKey | ''>('');
  const form = reactive({
    timeField: 'loginAt',
    range: [] as string[],
    loginType: undefined as LoginType | undefined,
    keywordType: 'memberId',
    keyword: '',
    device: undefined as DeviceType | undefined,
  });
  const query = ref<QueryState>({
    range: [],
    keyword: '',
    loginType: undefined,
    device: undefined,
  });
  const pageNo = ref(1);
  const pageSize = ref(15);

  function inRange(time: string, range: string[]) {
    if (range.length !== 2) return true;
    // 按分站时区的整天区间，左闭右开
    const value = dayjs(time).valueOf();
    return (
      value >= siteDayStartMs(range[0]) && value < siteNextDayStartMs(range[1])
    );
  }

  const filtered = computed(() => {
    const current = query.value;
    const keyword = current.keyword.trim();
    return records.value.filter((item) => {
      if (current.loginType && item.loginType !== current.loginType) {
        return false;
      }
      if (current.device && item.device !== current.device) return false;
      if (!inRange(item.loginAt, current.range)) return false;
      if (!keyword) return true;
      return item.memberId.includes(keyword);
    });
  });

  const pageRows = computed(() =>
    filtered.value.slice(
      (pageNo.value - 1) * pageSize.value,
      pageNo.value * pageSize.value
    )
  );

  watch(
    () => filtered.value.length,
    (total) => {
      const maxPage = Math.max(1, Math.ceil(total / pageSize.value) || 1);
      if (pageNo.value > maxPage) pageNo.value = maxPage;
    }
  );

  function buildQuickRange(key: QuickKey): [string, string] {
    const end = siteNow().format('YYYY-MM-DD');
    return [
      siteNow().subtract(quickOffset[key], 'day').format('YYYY-MM-DD'),
      end,
    ];
  }

  function handleSearch() {
    query.value = {
      range: [...form.range],
      keyword: form.keyword,
      loginType: form.loginType,
      device: form.device,
    };
    pageNo.value = 1;
  }

  function handleReset() {
    form.timeField = 'loginAt';
    form.range = [];
    form.loginType = undefined;
    form.keywordType = 'memberId';
    form.keyword = '';
    form.device = undefined;
    quickRange.value = '';
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
    flex-wrap: wrap;
    gap: 12px 16px;
    align-items: center;
  }

  .time-type {
    width: 112px;
  }

  .time-type,
  .id-type {
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

  .id-search {
    display: inline-flex;
  }

  .id-type {
    width: 112px;
  }

  .search-input {
    width: 180px;
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

  .stack-cell {
    line-height: 20px;
    word-break: break-all;
  }

  .device-code {
    font-size: 12px;
  }

  .hit-count {
    color: rgb(var(--primary-6));
    text-align: center;
  }

  .result-success {
    color: rgb(var(--success-6));
  }

  .result-failed {
    color: rgb(var(--danger-6));
    line-height: 20px;
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
