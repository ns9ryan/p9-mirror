<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.agentManage', 'menu.agentManage.allAgent']" />
    <a-card class="general-card" :bordered="false">
      <a-form :model="form" layout="inline" class="search-form">
        <a-form-item :label="t('allAgent.search.registerTime')">
          <a-space wrap>
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
        <a-form-item>
          <a-space>
            <a-button type="primary">
              {{ t('allAgent.search.subCount') }}
            </a-button>
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
      </div>

      <div class="agent-table-wrap">
        <a-table
          row-key="id"
          class="agent-table"
          :data="pageData"
          :pagination="false"
          :bordered="{ cell: true }"
          :show-header="false"
        >
          <template #columns>
            <a-table-column data-index="id">
              <template #cell="{ record }">
                <AllAgentCard
                  :record="record"
                  :expanded="isExpanded(record.id)"
                  @update:expanded="(value) => setExpanded(record.id, value)"
                  @bet="(action) => confirmBet(record, action)"
                  @remark="openRemark(record)"
                />
              </template>
            </a-table-column>
          </template>
          <template #empty>
            <a-empty :description="t('allAgent.emptyList')" />
          </template>
        </a-table>
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
  import createAllAgentSeed from './mock';
  import { REGISTER_SOURCES } from './options';
  import {
    buildQuickRange,
    createAllAgentQuery,
    detectQuickRange,
    matchAllAgent,
  } from './query';
  import type { AllAgentQuery, AllAgentRecord, QuickRange } from './types';

  type BetAction = 'limitSelf' | 'limitLine' | 'liftSelf' | 'liftLine';

  const { t } = useI18n();

  const agents = ref<AllAgentRecord[]>(createAllAgentSeed());
  const form = reactive<AllAgentQuery>(createAllAgentQuery());
  const applied = ref<AllAgentQuery>(createAllAgentQuery());
  const quickRange = ref<QuickRange | ''>('');
  const keywordType = ref('id');
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

  function patchAgent(
    id: number,
    patch: Partial<AllAgentRecord>
  ): AllAgentRecord | undefined {
    let next: AllAgentRecord | undefined;
    agents.value = agents.value.map((item) => {
      if (item.id !== id) return item;
      next = { ...item, ...patch };
      return next;
    });
    return next;
  }

  function confirmBet(record: AllAgentRecord, action: BetAction) {
    const already =
      (action === 'limitSelf' && record.selfBetBanned) ||
      (action === 'liftSelf' && !record.selfBetBanned) ||
      (action === 'limitLine' && record.lineBetBanned) ||
      (action === 'liftLine' && !record.lineBetBanned);
    if (already) {
      Message.info(t('allAgent.message.already'));
      return;
    }
    const titleKey = {
      limitSelf: 'allAgent.confirm.limitSelfTitle',
      limitLine: 'allAgent.confirm.limitLineTitle',
      liftSelf: 'allAgent.confirm.liftSelfTitle',
      liftLine: 'allAgent.confirm.liftLineTitle',
    }[action];
    const contentKey = {
      limitSelf: 'allAgent.confirm.limitSelfContent',
      limitLine: 'allAgent.confirm.limitLineContent',
      liftSelf: 'allAgent.confirm.liftSelfContent',
      liftLine: 'allAgent.confirm.liftLineContent',
    }[action];
    Modal.confirm({
      title: t(titleKey),
      content: t(contentKey, { id: record.id }),
      okText: t('allAgent.confirm.ok'),
      cancelText: t('allAgent.confirm.cancel'),
      onOk: () => {
        if (action === 'limitSelf' || action === 'liftSelf') {
          patchAgent(record.id, { selfBetBanned: action === 'limitSelf' });
        } else {
          patchAgent(record.id, { lineBetBanned: action === 'limitLine' });
        }
        const messageKey = {
          limitSelf: 'allAgent.message.limitSelf',
          limitLine: 'allAgent.message.limitLine',
          liftSelf: 'allAgent.message.liftSelf',
          liftLine: 'allAgent.message.liftLine',
        }[action];
        Message.success(t(messageKey, { id: record.id }));
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
    patchAgent(remarkId.value, { remark: remarkDraft.value.trim() });
    remarkVisible.value = false;
    Message.success(t('allAgent.message.remarkSaved'));
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

  .search-input {
    width: 170px;
  }

  .range-picker {
    width: 260px;
  }

  .range-number {
    width: 110px;
  }

  .range-sep {
    margin: 0 8px;
    color: var(--color-text-3);
  }

  .table-toolbar {
    margin-bottom: 14px;
  }

  .agent-table-wrap {
    width: 100%;
  }

  .agent-table {
    width: 100%;

    :deep(.arco-table),
    :deep(.arco-table-container),
    :deep(.arco-table-element) {
      width: 100%;
    }

    :deep(.arco-table-td) {
      padding: 0;
      vertical-align: top;
    }

    :deep(.arco-table-tr) {
      background-color: var(--color-bg-2);
    }

    :deep(.arco-table-tr + .arco-table-tr .arco-table-td) {
      border-top: 12px solid var(--color-fill-1);
    }
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
