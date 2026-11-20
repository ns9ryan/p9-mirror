<template>
  <div class="record-panel">
    <div class="record-filter">
      <a-space :size="12" wrap>
        <div class="time-group">
          <span class="time-label">
            <icon-clock-circle />
            {{
              kind === 'wallet'
                ? t('agentList.edit.record.walletTime')
                : t('agentList.edit.record.logTime')
            }}
          </span>
          <a-range-picker
            v-model="filter.range"
            class="range-picker"
            value-format="YYYY-MM-DD"
            :placeholder="[
              t('agentList.search.start'),
              t('agentList.search.end'),
            ]"
            allow-clear
            @change="onRangeChange"
          />
        </div>
        <a-radio-group
          v-model="quickRange"
          class="quick-range"
          type="button"
          size="small"
          @change="onQuickRange"
        >
          <a-radio value="today">{{ t('agentList.search.today') }}</a-radio>
          <a-radio value="d7">{{ t('agentList.search.last7') }}</a-radio>
          <a-radio value="d30">{{ t('agentList.search.last30') }}</a-radio>
        </a-radio-group>
        <template v-if="kind === 'wallet'">
          <a-select
            v-model="filter.mainType"
            class="type-select"
            allow-clear
            :placeholder="t('agentList.edit.record.mainType')"
            @change="filter.subType = undefined"
          >
            <a-option
              v-for="item in WALLET_TYPES"
              :key="item.value"
              :value="item.value"
            >
              {{ t(`agentList.edit.wallet.type.${item.value}`) }}
            </a-option>
          </a-select>
          <a-select
            v-model="filter.subType"
            class="type-select"
            allow-clear
            :disabled="!filter.mainType"
            :placeholder="t('agentList.edit.record.subType')"
          >
            <a-option v-for="sub in subTypes" :key="sub" :value="sub">
              {{ t(`agentList.edit.wallet.sub.${sub}`) }}
            </a-option>
          </a-select>
        </template>
        <a-select
          v-else
          v-model="filter.result"
          class="type-select"
          :placeholder="t('agentList.edit.log.col.result')"
        >
          <a-option
            v-for="item in RESULT_OPTIONS"
            :key="item.value"
            :value="item.value"
          >
            {{ t(item.label) }}
          </a-option>
        </a-select>
      </a-space>
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
    </div>

    <a-table
      class="record-table"
      row-key="id"
      :data="rows"
      :columns="columns"
      :loading="loading"
      :pagination="false"
      :scroll="{ x: 1280 }"
    >
      <template #opTime="{ record }">
        {{ formatUnix(record.operated_at) || '--' }}
      </template>
      <template #opOperator="{ record }">
        <div>{{ record.admin_account || '--' }}</div>
        <div v-if="operatorSub(record)" class="cell-sub">
          {{ operatorSub(record) }}
        </div>
      </template>
      <template #opIp="{ record }">
        <div>{{ record.client_ip || '--' }}</div>
        <div v-if="record.client_ip_location" class="cell-sub">
          {{ record.client_ip_location }}
        </div>
      </template>
      <template #opItem="{ record }">
        {{ record.action_label || record.action || '--' }}
      </template>
      <template #opDevice="{ record }">
        <span v-if="!record.login_device && !record.udid">--</span>
        <a-tooltip
          v-else
          position="tl"
          :content-style="{ maxWidth: '420px', wordBreak: 'break-all' }"
        >
          <template #content>
            <div v-if="record.login_device">
              UA：{{ record.login_device }}
            </div>
            <div v-if="record.udid">UDID：{{ record.udid }}</div>
          </template>
          <div class="device-cell">
            <span class="device-icon">
              <icon-mobile v-if="deviceOf(record).mobile" />
              <icon-desktop v-else />
            </span>
            <div class="device-text">
              <div class="device-name">
                {{ deviceLabel(record) }}
              </div>
              <div v-if="record.udid" class="cell-sub device-udid">
                {{ shortUdid(record.udid as string) }}
              </div>
            </div>
          </div>
        </a-tooltip>
      </template>
      <template #opResult="{ record }">
        <span
          :class="
            record.result === AGENT_OP_RESULT.SUCCESS
              ? 'result-success'
              : 'result-fail'
          "
        >
          {{ record.result_label || '--' }}
        </span>
        <div v-if="record.fail_reason" class="cell-sub">
          {{ record.fail_reason }}
        </div>
      </template>
      <template #empty>
        <a-empty :description="t('agentList.edit.empty')" />
      </template>
    </a-table>

    <div class="record-pagination">
      <a-pagination
        :total="total"
        :current="pagination.current"
        :page-size="pagination.pageSize"
        :page-size-options="[10, 15, 30, 50]"
        show-total
        show-page-size
        @change="onPageChange"
        @page-size-change="onPageSizeChange"
      />
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { computed, onMounted, reactive, ref, watch } from 'vue';
  import { siteDayEndSec, siteDayStartSec, siteNow } from '@/utils/site-time';
  import { useI18n } from 'vue-i18n';
  import type { TableColumnData } from '@arco-design/web-vue';
  import {
    AGENT_OP_RESULT,
    AGENT_OP_TARGET_TYPE,
    listAgentOpLogs,
  } from '@/api/agent';
  import { parseUserAgent } from '@/utils/device';
  import { formatUnix } from '../mapper';
  import { WALLET_TYPES } from '../options';
  import type { QuickRange } from '../types';

  const props = defineProps<{
    kind: 'wallet' | 'log';
    /** 操作记录按操作目标（代理）查询 */
    agentId?: number;
  }>();

  const RESULT_OPTIONS = [
    { value: AGENT_OP_RESULT.ALL, label: 'agentList.edit.log.result.all' },
    {
      value: AGENT_OP_RESULT.SUCCESS,
      label: 'agentList.edit.log.result.success',
    },
    { value: AGENT_OP_RESULT.FAIL, label: 'agentList.edit.log.result.fail' },
  ];

  const { t } = useI18n();
  const quickOffset: Record<QuickRange, number> = { today: 0, d7: 6, d30: 29 };
  const QUICK_KEYS: QuickRange[] = ['today', 'd7', 'd30'];
  const DEFAULT_QUICK: QuickRange = 'd7';

  function quickDates(key: QuickRange) {
    const end = siteNow();
    return [
      end.subtract(quickOffset[key], 'day').format('YYYY-MM-DD'),
      end.format('YYYY-MM-DD'),
    ];
  }

  function createFilter() {
    return {
      range: quickDates(DEFAULT_QUICK),
      mainType: undefined as string | undefined,
      subType: undefined as string | undefined,
      result: AGENT_OP_RESULT.ALL as number,
    };
  }

  const filter = reactive(createFilter());
  const quickRange = ref<QuickRange | ''>(DEFAULT_QUICK);
  const rows = ref<Record<string, unknown>[]>([]);
  const total = ref(0);
  const loading = ref(false);
  let loadSeq = 0;
  const pagination = reactive({
    current: 1,
    pageSize: props.kind === 'wallet' ? 15 : 10,
  });

  const subTypes = computed(
    () =>
      WALLET_TYPES.find((item) => item.value === filter.mainType)?.subs || []
  );

  function deviceOf(record: Record<string, unknown>) {
    return parseUserAgent(record.login_device as string | undefined);
  }

  function deviceLabel(record: Record<string, unknown>) {
    const { browser, os } = deviceOf(record);
    const label = [browser, os].filter(Boolean).join(' · ');
    return label || (record.login_device as string) || '--';
  }

  function operatorSub(record: Record<string, unknown>) {
    const id = record.admin_id ? `ID：${record.admin_id}` : '';
    return [record.admin_type_label, id].filter(Boolean).join(' · ');
  }

  function shortUdid(udid: string) {
    return udid.length > 13 ? `${udid.slice(0, 8)}…${udid.slice(-4)}` : udid;
  }

  const sortable: TableColumnData['sortable'] = {
    sortDirections: ['ascend', 'descend'],
  };

  const columns = computed<TableColumnData[]>(() => {
    if (props.kind === 'wallet') {
      return [
        {
          title: t('agentList.edit.wallet.col.orderNo'),
          dataIndex: 'orderNo',
          width: 180,
        },
        {
          title: t('agentList.edit.wallet.col.time'),
          dataIndex: 'time',
          width: 170,
        },
        {
          title: t('agentList.edit.wallet.col.agentId'),
          dataIndex: 'agentId',
          width: 100,
        },
        {
          title: t('agentList.edit.wallet.col.account'),
          dataIndex: 'account',
          width: 130,
        },
        {
          title: t('agentList.edit.wallet.col.currency'),
          dataIndex: 'currency',
          width: 90,
        },
        {
          title: t('agentList.edit.wallet.col.mainType'),
          dataIndex: 'mainType',
          width: 110,
        },
        {
          title: t('agentList.edit.wallet.col.subType'),
          dataIndex: 'subType',
          width: 110,
        },
        {
          title: t('agentList.edit.wallet.col.amount'),
          dataIndex: 'amount',
          width: 120,
          align: 'right',
          sortable,
        },
        {
          title: t('agentList.edit.wallet.col.before'),
          dataIndex: 'before',
          width: 130,
          align: 'right',
          sortable,
        },
        {
          title: t('agentList.edit.wallet.col.after'),
          dataIndex: 'after',
          width: 130,
          align: 'right',
          sortable,
        },
        { title: t('agentList.edit.wallet.col.remark'), dataIndex: 'remark' },
      ];
    }
    return [
      { title: t('agentList.edit.log.col.id'), dataIndex: 'id', width: 90 },
      {
        title: t('agentList.edit.log.col.time'),
        slotName: 'opTime',
        width: 170,
      },
      {
        title: t('agentList.edit.log.col.operator'),
        slotName: 'opOperator',
        width: 130,
      },
      {
        title: t('agentList.edit.log.col.item'),
        slotName: 'opItem',
        width: 140,
      },
      {
        title: t('agentList.edit.log.col.before'),
        dataIndex: 'before_value',
        width: 160,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: t('agentList.edit.log.col.after'),
        dataIndex: 'after_value',
        width: 160,
        ellipsis: true,
        tooltip: true,
      },
      {
        title: t('agentList.edit.log.col.device'),
        slotName: 'opDevice',
        width: 220,
      },
      {
        title: t('agentList.edit.log.col.ip'),
        slotName: 'opIp',
        width: 150,
      },
      { title: t('agentList.edit.log.col.result'), slotName: 'opResult' },
    ];
  });

  function rangeParams() {
    const [start, end] = filter.range || [];
    if (!start || !end) return {};
    return {
      time_start: siteDayStartSec(start),
      time_end: siteDayEndSec(end),
    };
  }

  /** 钱包记录接口未提供，先保持空列表 */
  async function loadList() {
    loadSeq += 1;
    const seq = loadSeq;
    if (props.kind !== 'log' || !props.agentId) {
      rows.value = [];
      total.value = 0;
      return;
    }
    loading.value = true;
    try {
      const { data } = await listAgentOpLogs({
        target_id: props.agentId,
        target_type: AGENT_OP_TARGET_TYPE.AGENT,
        page: pagination.current,
        page_size: pagination.pageSize,
        result: filter.result,
        ...rangeParams(),
      });
      if (seq !== loadSeq) return;
      rows.value = (data?.list || []) as unknown as Record<string, unknown>[];
      total.value = data?.total || 0;
    } catch {
      if (seq !== loadSeq) return;
      rows.value = [];
      total.value = 0;
    } finally {
      if (seq === loadSeq) loading.value = false;
    }
  }

  function handleSearch() {
    pagination.current = 1;
    loadList();
  }

  function handleReset() {
    Object.assign(filter, createFilter());
    quickRange.value = DEFAULT_QUICK;
    handleSearch();
  }

  function onQuickRange(value: string | number | boolean) {
    const key = String(value) as QuickRange;
    if (!QUICK_KEYS.includes(key)) return;
    filter.range = quickDates(key);
    quickRange.value = key;
    handleSearch();
  }

  function onRangeChange() {
    quickRange.value =
      QUICK_KEYS.find((key) => {
        const [start, end] = quickDates(key);
        return filter.range?.[0] === start && filter.range?.[1] === end;
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

  watch(() => props.agentId, handleReset);

  onMounted(loadList);
</script>

<style scoped lang="less">
  .record-panel {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .record-filter {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    align-items: center;
    justify-content: space-between;
    padding: 14px 16px;
    background: var(--color-fill-1);
    border-radius: 8px;
  }

  .time-group {
    display: inline-flex;
    align-items: stretch;
  }

  .time-label {
    display: inline-flex;
    gap: 6px;
    align-items: center;
    padding: 0 12px;
    color: #fff;
    font-size: 13px;
    white-space: nowrap;
    background: rgb(var(--primary-6));
    border-radius: var(--border-radius-small) 0 0 var(--border-radius-small);
  }

  .range-picker {
    width: 260px;
    background: var(--color-bg-2);
    border-top-left-radius: 0;
    border-bottom-left-radius: 0;
  }

  .quick-range {
    :deep(.arco-radio-button-content) {
      padding: 0 12px;
    }
  }

  .type-select {
    width: 150px;

    :deep(.arco-select-view) {
      background: var(--color-bg-2);
    }
  }

  .record-table {
    :deep(.arco-table-th) {
      font-weight: 600;
      background: var(--color-fill-2);
    }

    :deep(.arco-table-th),
    :deep(.arco-table-td) {
      white-space: nowrap;
    }

    :deep(.arco-table-container) {
      overflow: hidden;
      border: 1px solid var(--color-neutral-3);
      border-radius: 8px;
    }
  }

  .cell-sub {
    color: var(--color-text-3);
    font-size: 12px;
    line-height: 18px;
  }

  .device-cell {
    display: inline-flex;
    gap: 8px;
    align-items: center;
    max-width: 100%;
    cursor: default;
  }

  .device-icon {
    display: inline-flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    color: rgb(var(--primary-6));
    font-size: 16px;
    background: rgb(var(--primary-1));
    border-radius: 6px;
  }

  .device-text {
    min-width: 0;
  }

  .device-name {
    overflow: hidden;
    color: var(--color-text-1);
    text-overflow: ellipsis;
  }

  .device-udid {
    font-family: Consolas, Menlo, monospace;
  }

  .result-success {
    color: rgb(var(--success-6));
  }

  .result-fail {
    color: rgb(var(--danger-6));
  }

  .record-pagination {
    display: flex;
    justify-content: flex-end;
  }
</style>
