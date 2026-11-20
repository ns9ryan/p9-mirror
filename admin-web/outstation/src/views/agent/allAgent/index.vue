<template>
  <div class="container list-page">
    <a-card class="general-card" :bordered="false">
      <a-form :model="form" layout="inline" class="search-form">
        <a-form-item>
          <a-space wrap>
            <a-select v-model="form.timeType" class="time-type">
              <a-option value="register">
                {{ t('allAgent.search.registerTime') }}
              </a-option>
              <a-option value="login">
                {{ t('allAgent.search.loginTime') }}
              </a-option>
            </a-select>
            <a-radio-group
              v-model="quickRange"
              type="button"
              size="small"
              @change="onQuickRange"
            >
              <a-radio value="today">{{ t('allAgent.search.today') }}</a-radio>
              <a-radio value="d7">{{ t('allAgent.search.last7') }}</a-radio>
              <a-radio value="d30">{{ t('allAgent.search.last30') }}</a-radio>
            </a-radio-group>
            <a-range-picker
              v-model="form.dateRange"
              value-format="YYYY-MM-DD"
              allow-clear
              class="range-picker"
              :placeholder="[
                t('allAgent.search.start'),
                t('allAgent.search.end'),
              ]"
              @change="onRangeChange"
            />
          </a-space>
        </a-form-item>
        <a-form-item>
          <a-input-group compact>
            <a-select v-model="keywordType" class="search-type">
              <a-option value="id">
                {{ t('allAgent.search.memberId') }}
              </a-option>
            </a-select>
            <a-input
              v-model="form.keyword"
              class="search-input"
              allow-clear
              :placeholder="t('allAgent.search.keywordPlaceholder')"
              @press-enter="handleSearch"
            />
          </a-input-group>
        </a-form-item>
        <a-form-item :label="t('allAgent.search.registerSource')">
          <a-select v-model="form.registerSource" class="search-select">
            <a-option value="">{{ t('allAgent.search.all') }}</a-option>
            <a-option
              v-for="item in REGISTER_SOURCES"
              :key="item"
              :value="item"
            >
              {{ item }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item :label="t('allAgent.search.withdrawLimit')">
          <a-select v-model="form.withdrawLimit" class="search-select">
            <a-option value="">{{ t('allAgent.search.all') }}</a-option>
            <a-option value="none">
              {{ t('allAgent.withdraw.none') }}
            </a-option>
            <a-option value="limited">
              {{ t('allAgent.withdraw.limited') }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item :label="t('allAgent.search.status')">
          <a-select v-model="form.status" class="search-select">
            <a-option value="">{{ t('allAgent.search.all') }}</a-option>
            <a-option value="normal">
              {{ t('allAgent.status.normal') }}
            </a-option>
            <a-option value="unset">
              {{ t('allAgent.status.unset') }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item :label="t('allAgent.search.subCount')">
          <a-space :size="0">
            <a-input-number
              v-model="form.subCountMin"
              class="range-number"
              :min="0"
              :precision="0"
              :placeholder="t('allAgent.search.min')"
            />
            <span class="range-sep">-</span>
            <a-input-number
              v-model="form.subCountMax"
              class="range-number"
              :min="0"
              :precision="0"
              :placeholder="t('allAgent.search.max')"
            />
          </a-space>
        </a-form-item>
        <a-form-item>
          <a-space>
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
      </a-form>
    </a-card>

    <a-card class="general-card table-card" :bordered="false">
      <div class="table-toolbar">
        <a-space>
          <a-dropdown @select="onBatch">
            <a-button>
              {{ t('allAgent.toolbar.batch') }}
              <icon-down />
            </a-button>
            <template #content>
              <a-doption
                v-for="action in BET_ACTIONS"
                :key="action"
                :value="action"
              >
                {{ t(`allAgent.action.${action}`) }}
              </a-doption>
            </template>
          </a-dropdown>
          <a-button :disabled="!filtered.length" @click="exportCsv">
            <template #icon><icon-download /></template>
            {{ t('allAgent.toolbar.export') }}
          </a-button>
          <a-button :disabled="!pageData.length" @click="togglePageExpanded">
            <template #icon>
              <icon-up v-if="pageAllExpanded" />
              <icon-down v-else />
            </template>
            {{
              pageAllExpanded
                ? t('allAgent.toolbar.collapsePage')
                : t('allAgent.toolbar.expandPage')
            }}
          </a-button>
          <a-checkbox
            :model-value="pageAllSelected"
            :indeterminate="pageIndeterminate"
            @change="togglePage"
          >
            {{ t('allAgent.toolbar.selectPage') }}
          </a-checkbox>
          <span v-if="selectedIds.size" class="selected-count">
            {{ t('allAgent.toolbar.selected', { count: selectedIds.size }) }}
          </span>
        </a-space>
      </div>

      <div class="agent-table-wrap">
        <div class="agent-board">
          <div class="agent-scroll">
            <div class="agent-head">
              <span>
                <a-checkbox
                  :model-value="pageAllSelected"
                  :indeterminate="pageIndeterminate"
                  :disabled="!pageData.length"
                  @change="togglePage"
                />
              </span>
              <span></span>
              <span>{{ t('allAgent.column.agent') }}</span>
              <span>{{ t('allAgent.column.subCount') }}</span>
              <span>{{ t('allAgent.column.subTotal') }}</span>
              <span>{{ t('allAgent.column.commission') }}</span>
              <span>{{ t('allAgent.column.unclaimed') }}</span>
              <span>{{ t('allAgent.column.withdrawLimit') }}</span>
              <span>{{ t('allAgent.column.status') }}</span>
              <span>{{ t('allAgent.column.registeredAt') }}</span>
              <span>{{ t('allAgent.column.operate') }}</span>
            </div>
            <AllAgentCard
              v-for="record in pageData"
              :key="record.id"
              class="agent-line"
              :record="record"
              :selected="selectedIds.has(record.id)"
              :expanded="isExpanded(record.id)"
              @update:selected="(checked) => setSelected(record.id, checked)"
              @update:expanded="(value) => setExpanded(record.id, value)"
              @bet="(action) => confirmBet(record, action)"
              @remark="openRemark(record)"
            />
          </div>
          <a-empty
            v-if="!pageData.length"
            class="agent-empty"
            :description="t('allAgent.emptyList')"
          />
        </div>

        <div class="table-pagination">
          <a-pagination
            :total="filtered.length"
            :current="pagination.current"
            :page-size="pagination.pageSize"
            :page-size-options="pagination.pageSizeOptions"
            :show-total="pagination.showTotal"
            :show-page-size="pagination.showPageSize"
            :page-size-props="pagination.pageSizeProps"
            @change="onPageChange"
            @page-size-change="onPageSizeChange"
          />
        </div>
      </div>
    </a-card>

    <a-modal
      :visible="remarkVisible"
      :title="t('allAgent.remark.title')"
      :ok-text="t('allAgent.confirm.ok')"
      :cancel-text="t('allAgent.confirm.cancel')"
      @ok="saveRemark"
      @cancel="remarkVisible = false"
    >
      <a-textarea
        v-model="remarkDraft"
        :max-length="200"
        show-word-limit
        :auto-size="{ minRows: 3, maxRows: 6 }"
        :placeholder="t('allAgent.remark.placeholder')"
      />
    </a-modal>
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import dayjs from 'dayjs';
  import { useI18n } from 'vue-i18n';
  import { Message, Modal } from '@arco-design/web-vue';
  import AllAgentCard from './components/all-agent-card.vue';
  import { buildPromoLink, REGISTER_SOURCES } from './options';
  import { displayText, formatAmount } from './format';
  import createAllAgentSeed from './mock';
  import {
    buildQuickRange,
    createAllAgentQuery,
    detectQuickRange,
    matchAllAgent,
  } from './query';
  import type {
    AllAgentQuery,
    AllAgentRecord,
    BetAction,
    QuickRange,
  } from './types';

  const BET_ACTIONS: BetAction[] = [
    'limitSelf',
    'limitLine',
    'liftSelf',
    'liftLine',
  ];

  const { t } = useI18n();

  const agents = ref<AllAgentRecord[]>(createAllAgentSeed());
  const form = reactive<AllAgentQuery>(createAllAgentQuery());
  const applied = ref<AllAgentQuery>(createAllAgentQuery());
  const quickRange = ref<QuickRange | ''>('');
  const keywordType = ref('id');
  const selectedIds = ref<Set<number>>(new Set());
  const collapsedIds = ref<Set<number>>(new Set());
  const pagination = reactive({
    current: 1,
    pageSize: 10,
    showTotal: true,
    showPageSize: true,
    pageSizeProps: {
      popupContainer: 'body',
    },
    pageSizeOptions: [10, 15, 30, 50],
  });
  const remarkVisible = ref(false);
  const remarkId = ref<number | null>(null);
  const remarkDraft = ref('');

  const filtered = computed(() =>
    agents.value.filter((item) => matchAllAgent(item, applied.value))
  );

  const pageData = computed(() => {
    const start = (pagination.current - 1) * pagination.pageSize;
    return filtered.value.slice(start, start + pagination.pageSize);
  });

  const pageAllExpanded = computed(
    () =>
      pageData.value.length > 0 &&
      pageData.value.every((item) => !collapsedIds.value.has(item.id))
  );

  const pageAllSelected = computed(
    () =>
      pageData.value.length > 0 &&
      pageData.value.every((item) => selectedIds.value.has(item.id))
  );

  const pageIndeterminate = computed(
    () =>
      !pageAllSelected.value &&
      pageData.value.some((item) => selectedIds.value.has(item.id))
  );

  watch(filtered, (list) => {
    const maxPage = Math.max(1, Math.ceil(list.length / pagination.pageSize));
    if (pagination.current > maxPage) pagination.current = maxPage;
  });

  function snapshotQuery(): AllAgentQuery {
    return {
      ...form,
      dateRange: [...(form.dateRange || [])],
    };
  }

  function handleSearch() {
    applied.value = snapshotQuery();
    pagination.current = 1;
  }

  function handleReset() {
    Object.assign(form, createAllAgentQuery());
    quickRange.value = '';
    handleSearch();
  }

  function onQuickRange(value: string | number | boolean) {
    const key = String(value) as QuickRange;
    if (key !== 'today' && key !== 'd7' && key !== 'd30') return;
    form.dateRange = buildQuickRange(key);
    quickRange.value = key;
    handleSearch();
  }

  function onRangeChange(
    value: (string | number | Date | undefined)[] | undefined
  ) {
    const range = (value || []).flatMap((item) => {
      if (item == null || item === '') return [];
      if (item instanceof Date) return [dayjs(item).format('YYYY-MM-DD')];
      return [String(item).slice(0, 10)];
    });
    quickRange.value = detectQuickRange(range);
  }

  function onPageChange(current: number) {
    pagination.current = current;
  }

  function onPageSizeChange(size: number) {
    pagination.pageSize = size;
    pagination.current = 1;
  }

  function isExpanded(id: number) {
    return !collapsedIds.value.has(id);
  }

  function setExpanded(id: number, expanded: boolean) {
    const next = new Set(collapsedIds.value);
    if (expanded) next.delete(id);
    else next.add(id);
    collapsedIds.value = next;
  }

  function togglePageExpanded() {
    const next = new Set(collapsedIds.value);
    const collapse = pageAllExpanded.value;
    pageData.value.forEach((item) => {
      if (collapse) next.add(item.id);
      else next.delete(item.id);
    });
    collapsedIds.value = next;
  }

  function setSelected(id: number, checked: boolean) {
    const next = new Set(selectedIds.value);
    if (checked) next.add(id);
    else next.delete(id);
    selectedIds.value = next;
  }

  function togglePage(value: boolean | (string | number | boolean)[]) {
    const checked = value === true;
    const next = new Set(selectedIds.value);
    pageData.value.forEach((item) => {
      if (checked) next.add(item.id);
      else next.delete(item.id);
    });
    selectedIds.value = next;
  }

  function betPatch(action: BetAction): Partial<AllAgentRecord> {
    if (action === 'limitSelf' || action === 'liftSelf') {
      return { selfBetBanned: action === 'limitSelf' };
    }
    return { lineBetBanned: action === 'limitLine' };
  }

  function isBetApplied(record: AllAgentRecord, action: BetAction) {
    return (
      (action === 'limitSelf' && record.selfBetBanned) ||
      (action === 'liftSelf' && !record.selfBetBanned) ||
      (action === 'limitLine' && record.lineBetBanned) ||
      (action === 'liftLine' && !record.lineBetBanned)
    );
  }

  function applyBet(ids: number[], action: BetAction) {
    const idSet = new Set(ids);
    const patch = betPatch(action);
    let count = 0;
    agents.value = agents.value.map((item) => {
      if (!idSet.has(item.id) || isBetApplied(item, action)) return item;
      count += 1;
      return { ...item, ...patch };
    });
    return count;
  }

  function confirmBet(record: AllAgentRecord, action: BetAction) {
    if (isBetApplied(record, action)) {
      Message.info(t('allAgent.message.already'));
      return;
    }
    Modal.confirm({
      title: t(`allAgent.confirm.${action}Title`),
      content: t(`allAgent.confirm.${action}Content`, { id: record.id }),
      okText: t('allAgent.confirm.ok'),
      cancelText: t('allAgent.confirm.cancel'),
      onOk: () => {
        applyBet([record.id], action);
        Message.success(t(`allAgent.message.${action}`, { id: record.id }));
      },
    });
  }

  function onBatch(
    value: string | number | Record<string, unknown> | undefined
  ) {
    const action = BET_ACTIONS.find((item) => item === value);
    if (!action) return;
    if (!selectedIds.value.size) {
      Message.warning(t('allAgent.message.selectFirst'));
      return;
    }
    const ids = [...selectedIds.value];
    Modal.confirm({
      title: t(`allAgent.confirm.${action}Title`),
      content: t('allAgent.confirm.batchContent', {
        count: ids.length,
        action: t(`allAgent.action.${action}`),
      }),
      okText: t('allAgent.confirm.ok'),
      cancelText: t('allAgent.confirm.cancel'),
      onOk: () => {
        const count = applyBet(ids, action);
        if (!count) {
          Message.info(t('allAgent.message.already'));
          return;
        }
        Message.success(t('allAgent.message.batchDone', { count }));
      },
    });
  }

  function openRemark(record: AllAgentRecord) {
    remarkId.value = record.id;
    remarkDraft.value = record.remark;
    remarkVisible.value = true;
  }

  function saveRemark() {
    if (remarkId.value == null) return;
    const id = remarkId.value;
    const remark = remarkDraft.value.trim();
    agents.value = agents.value.map((item) =>
      item.id === id ? { ...item, remark } : item
    );
    remarkVisible.value = false;
    Message.success(t('allAgent.message.remarkSaved'));
  }

  function csvCell(value: string | number) {
    const text = String(value ?? '');
    if (/[",\n]/.test(text)) return `"${text.replace(/"/g, '""')}"`;
    return text;
  }

  function exportCsv() {
    const headers = [
      t('allAgent.column.memberId'),
      t('allAgent.column.parentId'),
      t('allAgent.column.subCount'),
      t('allAgent.column.subTotal'),
      t('allAgent.column.betLevel1'),
      t('allAgent.column.betLevel2'),
      t('allAgent.column.betLevel3'),
      t('allAgent.column.claimed'),
      t('allAgent.column.pending'),
      t('allAgent.column.commission'),
      t('allAgent.column.frozen'),
      t('allAgent.column.cancelled'),
      t('allAgent.column.unclaimed'),
      t('allAgent.column.withdrawLimit'),
      t('allAgent.column.remark'),
      t('allAgent.search.registerSource'),
      t('allAgent.column.link'),
      t('allAgent.column.uv'),
      t('allAgent.column.registeredAt'),
      t('allAgent.column.lastLogin'),
      t('allAgent.column.status'),
    ];
    const rows = filtered.value.map((item) => [
      item.id,
      item.parentId,
      item.subCount,
      item.subTotal,
      formatAmount(item.betLevel1),
      formatAmount(item.betLevel2),
      formatAmount(item.betLevel3),
      formatAmount(item.claimedCommission),
      formatAmount(item.pendingCommission),
      formatAmount(item.commission),
      formatAmount(item.frozenCommission),
      formatAmount(item.cancelledCommission),
      formatAmount(item.unclaimedCommission),
      t(`allAgent.withdraw.${item.withdrawLimit}`),
      item.remark,
      item.registerSource,
      item.promoCode ? buildPromoLink(item.promoHost, item.promoCode) : '',
      item.visitUv,
      displayText(item.registeredAt),
      displayText(item.lastLoginAt),
      t(`allAgent.status.${item.status}`),
    ]);
    const csv = [headers, ...rows]
      .map((cols) => cols.map((col) => csvCell(col)).join(','))
      .join('\n');
    const blob = new Blob([`\uFEFF${csv}`], {
      type: 'text/csv;charset=utf-8;',
    });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `agents-${dayjs().format('YYYYMMDD-HHmmss')}.csv`;
    link.click();
    URL.revokeObjectURL(url);
    Message.success(t('allAgent.message.exported'));
  }
</script>

<style scoped lang="less">
  .container {
    padding: 0 20px 20px;
  }

  .general-card {
    margin-top: 16px;
  }

  .general-card :deep(.arco-card-body) {
    padding: 20px;
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

  .time-type {
    width: 120px;
  }

  .search-input {
    width: 170px;
  }

  .range-picker {
    width: 260px;
  }

  .range-number {
    width: 100px;
  }

  .range-sep {
    margin: 0 8px;
    color: var(--color-text-3);
  }

  .table-toolbar {
    margin-bottom: 14px;
  }

  .selected-count {
    color: var(--color-text-2);
    font-size: 13px;
  }

  .agent-table-wrap {
    width: 100%;
  }

  .agent-board {
    --agent-cols: 24px 28px minmax(150px, 1.3fr) minmax(64px, 0.6fr)
      minmax(72px, 0.6fr) minmax(88px, 0.8fr) minmax(88px, 0.8fr)
      minmax(88px, 0.8fr) minmax(64px, 0.6fr) minmax(140px, 1fr)
      minmax(150px, 1fr);

    overflow-x: auto;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 10px;
  }

  .agent-scroll {
    min-width: 1180px;
  }

  .agent-head {
    display: grid;
    grid-template-columns: var(--agent-cols);
    gap: 8px;
    align-items: center;
    height: 44px;
    padding: 0 16px;
    color: var(--color-text-3);
    font-size: 13px;
    background: var(--color-fill-2);
    border-bottom: 1px solid var(--color-border-2);
  }

  .agent-line + .agent-line {
    border-top: 1px solid var(--color-border-1);
  }

  .agent-empty {
    padding: 32px 0;
  }

  .table-pagination {
    display: flex;
    justify-content: flex-end;
    padding-top: 16px;
  }

  :deep(.arco-pagination-options .arco-select) {
    min-width: 100px;
  }
</style>
