<template>
  <div class="container list-page">
    <a-card class="general-card" :bordered="false">
      <a-form :model="form" layout="inline" class="search-form">
        <a-form-item>
          <a-space wrap>
            <a-select v-model="form.timeType" class="time-type">
              <a-option value="register">
                {{ t('memberManage.search.timeRegister') }}
              </a-option>
              <a-option value="login">
                {{ t('memberManage.search.timeLogin') }}
              </a-option>
            </a-select>
            <a-radio-group
              v-model="quickRange"
              type="button"
              size="small"
              @change="onQuickRange"
            >
              <a-radio value="today">
                {{ t('memberManage.search.today') }}
              </a-radio>
              <a-radio value="d7">{{ t('memberManage.search.last7') }}</a-radio>
              <a-radio value="d30">
                {{ t('memberManage.search.last30') }}
              </a-radio>
            </a-radio-group>
            <a-range-picker
              v-model="form.dateRange"
              value-format="YYYY-MM-DD"
              allow-clear
              class="range-picker"
              :placeholder="[
                t('memberManage.search.start'),
                t('memberManage.search.end'),
              ]"
              @change="onRangeChange"
            />
          </a-space>
        </a-form-item>
        <a-form-item>
          <a-input
            v-model="form.keyword"
            class="search-input"
            allow-clear
            :placeholder="t('memberManage.search.keywordPlaceholder')"
            @press-enter="handleSearch"
          >
            <template #prefix><icon-search /></template>
          </a-input>
        </a-form-item>
        <a-form-item :label="t('memberManage.search.autoLevel')">
          <a-select v-model="form.autoLevelId" class="search-select">
            <a-option value="">{{ t('memberManage.search.all') }}</a-option>
            <a-option
              v-for="item in autoLevels"
              :key="item.id"
              :value="item.id"
            >
              {{ item.name }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item :label="t('memberManage.search.fixedLevel')">
          <a-select v-model="form.fixedLevelId" class="search-select">
            <a-option value="">{{ t('memberManage.search.all') }}</a-option>
            <a-option
              v-for="item in fixedLevels"
              :key="item.id"
              :value="item.id"
            >
              {{ item.name }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item :label="t('memberManage.search.registerSource')">
          <a-select
            v-model="form.registerChannel"
            class="search-select"
            allow-create
          >
            <a-option value="">{{ t('memberManage.search.all') }}</a-option>
            <a-option
              v-for="item in REGISTER_CHANNELS"
              :key="item"
              :value="item"
            >
              {{ registerChannelText(item, t) }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item :label="t('memberManage.search.status')">
          <a-select v-model="form.status" class="search-select">
            <a-option value="">{{ t('memberManage.search.all') }}</a-option>
            <a-option value="normal">
              {{ t('memberManage.status.normal') }}
            </a-option>
            <a-option value="frozen">
              {{ t('memberManage.status.frozen') }}
            </a-option>
            <a-option value="disabled">
              {{ t('memberManage.status.disabled') }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item :label="t('memberManage.search.registerMethod')">
          <a-select
            v-model="form.registerMethod"
            class="search-select"
            allow-create
          >
            <a-option value="">{{ t('memberManage.search.all') }}</a-option>
            <a-option
              v-for="item in REGISTER_METHODS"
              :key="item"
              :value="item"
            >
              {{ registerMethodText(item, t) }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item class="flag-item">
          <a-space wrap>
            <a-checkbox v-model="form.betBanned">
              {{ t('memberManage.search.betBanned') }}
            </a-checkbox>
            <a-checkbox v-model="form.noPhone">
              {{ t('memberManage.search.noPhone') }}
            </a-checkbox>
            <a-checkbox v-model="form.noEmail">
              {{ t('memberManage.search.noEmail') }}
            </a-checkbox>
            <a-checkbox
              v-model="form.deposited"
              @change="onDepositFlag('deposited')"
            >
              {{ t('memberManage.search.deposited') }}
            </a-checkbox>
            <a-checkbox
              v-model="form.notDeposited"
              @change="onDepositFlag('notDeposited')"
            >
              {{ t('memberManage.search.notDeposited') }}
            </a-checkbox>
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
          <a-button type="primary" @click="openCreate">
            <template #icon><icon-plus /></template>
            {{ t('memberManage.toolbar.create') }}
          </a-button>
          <a-button @click="batchCreateVisible = true">
            <template #icon><icon-plus /></template>
            {{ t('memberManage.toolbar.batchCreate') }}
          </a-button>
          <a-dropdown @select="onBatch">
            <a-button>
              {{ t('memberManage.toolbar.batch') }}
              <icon-down />
            </a-button>
            <template #content>
              <a-doption value="disable">
                {{ t('memberManage.toolbar.batchDisable') }}
              </a-doption>
              <a-doption value="enable">
                {{ t('memberManage.toolbar.batchEnable') }}
              </a-doption>
            </template>
          </a-dropdown>
          <a-button @click="exportVisible = true">
            <template #icon><icon-download /></template>
            {{ t('memberManage.toolbar.export') }}
          </a-button>
          <a-button :disabled="!pageData.length" @click="togglePageExpanded">
            <template #icon>
              <icon-up v-if="pageAllExpanded" />
              <icon-down v-else />
            </template>
            {{
              pageAllExpanded
                ? t('memberManage.toolbar.collapsePage')
                : t('memberManage.toolbar.expandPage')
            }}
          </a-button>
          <a-checkbox
            :model-value="pageAllSelected"
            :indeterminate="pageIndeterminate"
            @change="togglePage"
          >
            {{ t('memberManage.toolbar.selectPage') }}
          </a-checkbox>
          <span v-if="selectedIds.size" class="selected-count">
            {{
              t('memberManage.toolbar.selected', { count: selectedIds.size })
            }}
          </span>
        </a-space>
        <LayoutSwitch />
      </div>

      <a-spin :loading="loading" class="member-table-wrap">
        <div class="member-panel">
          <div class="member-scroll">
            <div class="member-head">
              <span>
                <a-checkbox
                  :model-value="pageAllSelected"
                  :indeterminate="pageIndeterminate"
                  :disabled="!pageData.length"
                  @change="togglePage"
                />
              </span>
              <span></span>
              <span>{{ t('memberManage.column.memberId') }}</span>
              <span>{{ t('memberManage.column.account') }}</span>
              <span>{{ t('memberManage.column.agent') }}</span>
              <span>{{ t('memberManage.column.vip') }}</span>
              <span>{{ t('memberManage.column.status') }}</span>
              <span class="head-stack is-num">
                {{ t('memberManage.column.balance') }}
                <small>{{ t('memberManage.column.unsettled') }}</small>
              </span>
              <span class="head-stack is-num">
                {{ t('memberManage.column.todayDeposit') }}
                <small
                  class="sort-trigger"
                  :class="{ 'is-active': sortState.key === 'totalDeposit' }"
                  @click="toggleSort('totalDeposit')"
                >
                  {{ t('memberManage.column.totalDeposit') }}
                  <span class="sort-icons">
                    <icon-caret-up
                      :class="{
                        'is-on': sortOrderOf('totalDeposit') === 'asc',
                      }"
                    />
                    <icon-caret-down
                      :class="{
                        'is-on': sortOrderOf('totalDeposit') === 'desc',
                      }"
                    />
                  </span>
                </small>
              </span>
              <span class="head-stack is-num">
                {{ t('memberManage.column.todayWithdraw') }}
                <small
                  class="sort-trigger"
                  :class="{ 'is-active': sortState.key === 'totalWithdraw' }"
                  @click="toggleSort('totalWithdraw')"
                >
                  {{ t('memberManage.column.totalWithdraw') }}
                  <span class="sort-icons">
                    <icon-caret-up
                      :class="{
                        'is-on': sortOrderOf('totalWithdraw') === 'asc',
                      }"
                    />
                    <icon-caret-down
                      :class="{
                        'is-on': sortOrderOf('totalWithdraw') === 'desc',
                      }"
                    />
                  </span>
                </small>
              </span>
              <span class="head-stack is-num">
                {{ t('memberManage.column.todayBet') }}
                <small>{{ t('memberManage.column.validBet') }}</small>
              </span>
              <span class="head-stack is-num">
                {{ t('memberManage.column.todayWin') }}
                <small>{{ t('memberManage.column.totalWin') }}</small>
              </span>
              <span>{{ t('memberManage.column.source') }}</span>
              <span class="head-stack">
                {{ t('memberManage.column.registeredAt') }}
                <small>{{ t('memberManage.column.lastLogin') }}</small>
              </span>
              <span class="head-action">
                {{ t('memberManage.column.action') }}
              </span>
            </div>
            <MemberCard
              v-for="record in pageData"
              :key="record.id"
              class="member-line"
              :record="record"
              :selected="selectedIds.has(record.id)"
              :expanded="isExpanded(record.id)"
              @update:selected="(checked) => setSelected(record.id, checked)"
              @update:expanded="(value) => setExpanded(record.id, value)"
              @detail="openDetail(record)"
              @edit="openEdit(record)"
              @status="(command) => confirmStatus(record, command)"
              @operation="(command) => operationsRef?.open(command, record)"
            />
          </div>
          <a-empty
            v-if="!pageData.length"
            class="member-empty"
            :description="t('memberManage.emptyList')"
          />
        </div>

        <div class="table-pagination">
          <a-pagination
            :total="total"
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
      </a-spin>
    </a-card>

    <MemberDetailDrawer
      :visible="detailVisible"
      :record="detailRecord"
      :detail="activeDetail"
      :loading="detailLoading"
      @update:visible="detailVisible = $event"
      @edit="editFromDetail"
      @operation="operationFromDetail"
    />
    <MemberCreateModal
      :visible="createVisible"
      @update:visible="createVisible = $event"
      @created="onCreated"
    />
    <MemberBatchCreateModal
      :visible="batchCreateVisible"
      @update:visible="batchCreateVisible = $event"
      @refresh="onBatchCreated"
    />
    <MemberFormModal
      :visible="formVisible"
      :record="editingRecord"
      @update:visible="formVisible = $event"
      @saved="onSaved"
    />
    <MemberExportModal
      :visible="exportVisible"
      @update:visible="exportVisible = $event"
    />
    <MemberOperations ref="operationsRef" />
  </div>
</template>

<script lang="ts" setup>
  import { computed, onMounted, reactive, ref } from 'vue';
  import dayjs from 'dayjs';
  import { useI18n } from 'vue-i18n';
  import { Message, Modal } from '@arco-design/web-vue';
  import LayoutSwitch from './layout-switch.vue';
  import MemberCard from './member-card.vue';
  import MemberCreateModal from './member-create-modal.vue';
  import MemberBatchCreateModal from './member-batch-create-modal.vue';
  import MemberDetailDrawer from './member-detail-drawer.vue';
  import MemberExportModal from './member-export-modal.vue';
  import MemberFormModal from './member-form-modal.vue';
  import MemberOperations from './operations/member-operations.vue';
  import type { OperationCommand } from './operations/commands';
  import { STATUS_COMMANDS, type StatusCommand } from './v2/actions';
  import {
    REGISTER_CHANNELS,
    REGISTER_METHODS,
    registerChannelText,
    registerMethodText,
  } from '../format';
  import { useLevelOptions } from '../lookups';
  import {
    buildMemberParams,
    buildQuickRange,
    buildSortParams,
    createMemberQuery,
    detectQuickRange,
  } from '../query';
  import { useMemberStore } from '../store';
  import type {
    MemberQuery,
    MemberRecord,
    QuickRange,
    SortOrder,
  } from '../types';

  const { t } = useI18n();

  const {
    members: pageData,
    total,
    loading,
    fetchMembers,
    detail,
    detailLoading,
    fetchDetail,
    findMember,
    updateStatus,
  } = useMemberStore();
  const operationsRef = ref<InstanceType<typeof MemberOperations>>();
  const { autoLevels, fixedLevels, load: loadLevels } = useLevelOptions();
  const form = reactive<MemberQuery>(createMemberQuery());
  const applied = ref<MemberQuery>(createMemberQuery());
  const quickRange = ref<QuickRange | ''>('');
  const selectedIds = ref<Set<number>>(new Set());
  const expandedIds = ref<Set<number>>(new Set());
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

  const detailVisible = ref(false);
  const detailId = ref<number | null>(null);
  const createVisible = ref(false);
  const batchCreateVisible = ref(false);
  const exportVisible = ref(false);
  const formVisible = ref(false);
  const editingId = ref<number | null>(null);

  type SortKey = 'totalDeposit' | 'totalWithdraw';

  /** 未选排序时按注册时间倒序 */
  const sortState = reactive<{ key: SortKey | ''; order: SortOrder | '' }>({
    key: '',
    order: '',
  });

  async function loadMembers() {
    const { key, order } = sortState;
    selectedIds.value = new Set();
    await fetchMembers({
      page: pagination.current,
      page_size: pagination.pageSize,
      ...buildMemberParams(applied.value),
      ...(key && order ? buildSortParams(key, order) : {}),
    });
  }

  function reloadFirstPage() {
    pagination.current = 1;
    loadMembers();
  }

  onMounted(() => {
    loadLevels();
    loadMembers();
  });

  function sortOrderOf(key: SortKey) {
    return sortState.key === key ? sortState.order : '';
  }

  function toggleSort(key: SortKey) {
    const current = sortOrderOf(key);
    if (!current) {
      sortState.key = key;
      sortState.order = 'desc';
    } else if (current === 'desc') {
      sortState.order = 'asc';
    } else {
      sortState.key = '';
      sortState.order = '';
    }
    reloadFirstPage();
  }

  const pageAllExpanded = computed(
    () =>
      pageData.value.length > 0 &&
      pageData.value.every((item) => expandedIds.value.has(item.id))
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

  const activeDetail = computed(() =>
    detail.value?.id === detailId.value ? detail.value : null
  );
  const detailRecord = computed(
    () => activeDetail.value || findMember(detailId.value)
  );

  const editingRecord = computed(() => findMember(editingId.value));

  function snapshotQuery(): MemberQuery {
    return { ...form, dateRange: [...(form.dateRange || [])] };
  }

  function handleSearch() {
    applied.value = snapshotQuery();
    reloadFirstPage();
  }

  /** 接口不允许“充值过”和“未充值”同时为 true */
  function onDepositFlag(flag: 'deposited' | 'notDeposited') {
    if (!form[flag]) return;
    if (flag === 'deposited') form.notDeposited = false;
    else form.deposited = false;
  }

  function handleReset() {
    Object.assign(form, createMemberQuery());
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
    loadMembers();
  }

  function onPageSizeChange(size: number) {
    pagination.pageSize = size;
    reloadFirstPage();
  }

  function isExpanded(id: number) {
    return expandedIds.value.has(id);
  }

  function setExpanded(id: number, expanded: boolean) {
    const next = new Set(expandedIds.value);
    if (expanded) next.add(id);
    else next.delete(id);
    expandedIds.value = next;
  }

  function togglePageExpanded() {
    const next = new Set(expandedIds.value);
    const collapse = pageAllExpanded.value;
    pageData.value.forEach((item) => {
      if (collapse) next.delete(item.id);
      else next.add(item.id);
    });
    expandedIds.value = next;
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

  function confirmStatus(record: MemberRecord, command: StatusCommand) {
    const meta = STATUS_COMMANDS[command];
    Modal.confirm({
      title: t(meta.confirmTitle),
      content: t(meta.confirmContent, { id: record.code }),
      okText: t('memberManage.confirm.ok'),
      cancelText: t('memberManage.confirm.cancel'),
      onBeforeOk: async () => {
        const { changed } = await updateStatus([record.id], meta.target);
        if (changed) Message.success(t(meta.success));
        return true;
      },
    });
  }

  function onBatch(
    value: string | number | Record<string, unknown> | undefined
  ) {
    const action = String(value);
    if (!selectedIds.value.size) {
      Message.warning(t('memberManage.message.selectFirst'));
      return;
    }
    const disabling = action === 'disable';
    const ids = [...selectedIds.value];
    Modal.confirm({
      title: t(
        disabling
          ? 'memberManage.confirm.batchDisableTitle'
          : 'memberManage.confirm.batchEnableTitle'
      ),
      content: t(
        disabling
          ? 'memberManage.confirm.batchDisableContent'
          : 'memberManage.confirm.batchEnableContent',
        { count: ids.length }
      ),
      okText: t('memberManage.confirm.ok'),
      cancelText: t('memberManage.confirm.cancel'),
      onBeforeOk: async () => {
        const { changed, failed } = await updateStatus(
          ids,
          disabling ? 'disabled' : 'normal'
        );
        if (failed) {
          Message.warning(
            t('memberManage.message.batchPartial', { changed, failed })
          );
        } else if (!changed) {
          Message.info(t('memberManage.message.already'));
        } else {
          Message.success(
            t(
              disabling
                ? 'memberManage.message.batchDisabled'
                : 'memberManage.message.batchEnabled',
              { count: changed }
            )
          );
        }
        return true;
      },
    });
  }

  function openDetail(record: MemberRecord) {
    detailId.value = record.id;
    detailVisible.value = true;
    fetchDetail(record.id);
  }

  function openEdit(record: MemberRecord) {
    editingId.value = record.id;
    formVisible.value = true;
  }

  function openCreate() {
    createVisible.value = true;
  }

  function editFromDetail() {
    if (!detailRecord.value) return;
    detailVisible.value = false;
    openEdit(detailRecord.value);
  }

  function operationFromDetail(command: OperationCommand) {
    if (!detailRecord.value) return;
    operationsRef.value?.open(command, detailRecord.value);
  }

  /** 回到默认条件第一页，新会员按注册时间倒序排在最前 */
  function onCreated() {
    Message.success(t('memberManage.message.created'));
    Object.assign(form, createMemberQuery());
    quickRange.value = '';
    applied.value = createMemberQuery();
    sortState.key = '';
    sortState.order = '';
    reloadFirstPage();
  }

  /** 批量创建成功或超时后刷新：回到默认条件第一页，与单个创建保持一致 */
  function onBatchCreated() {
    Object.assign(form, createMemberQuery());
    quickRange.value = '';
    applied.value = createMemberQuery();
    sortState.key = '';
    sortState.order = '';
    reloadFirstPage();
  }

  function onSaved() {
    Message.success(t('memberManage.message.updated'));
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

  .time-type {
    width: 120px;
  }

  .search-input {
    width: 240px;
  }

  .range-picker {
    width: 260px;
  }

  .table-toolbar {
    display: flex;
    gap: 12px;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 14px;
  }

  .selected-count {
    color: var(--color-text-2);
    font-size: 13px;
  }

  .member-table-wrap {
    display: block;
    width: 100%;
  }

  .member-panel {
    --member-cols: 24px 28px minmax(96px, 0.9fr) minmax(120px, 1.1fr)
      minmax(110px, 1fr) 72px 64px minmax(110px, 1fr) minmax(100px, 0.9fr)
      minmax(100px, 0.9fr) minmax(100px, 0.9fr) minmax(100px, 0.9fr)
      minmax(90px, 0.8fr) minmax(136px, 1.1fr) 104px;

    overflow-x: auto;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 10px;
  }

  .member-scroll {
    min-width: 1500px;
  }

  .member-head {
    display: grid;
    grid-template-columns: var(--member-cols);
    gap: 8px;
    align-items: center;
    min-height: 52px;
    padding: 6px 16px;
    color: var(--color-text-2);
    font-weight: 500;
    font-size: 13px;
    background: var(--color-fill-2);
    border-bottom: 1px solid var(--color-border-2);
  }

  .head-stack {
    display: flex;
    flex-direction: column;
    line-height: 1.4;
  }

  .head-stack small {
    color: var(--color-text-3);
    font-weight: normal;
    font-size: 12px;
  }

  .head-stack.is-num {
    align-items: flex-end;
    text-align: right;
  }

  .sort-trigger {
    display: inline-flex;
    gap: 4px;
    align-items: center;
    cursor: pointer;
    user-select: none;

    &:hover,
    &.is-active {
      color: rgb(var(--primary-6));
    }
  }

  .sort-icons {
    display: inline-flex;
    flex-direction: column;
    color: var(--color-text-4);
    font-size: 10px;
    line-height: 0;

    .arco-icon + .arco-icon {
      margin-top: -3px;
    }

    .is-on {
      color: rgb(var(--primary-6));
    }
  }

  .head-action {
    text-align: center;
  }

  .member-line + .member-line {
    border-top: 1px solid var(--color-border-1);
  }

  .member-empty {
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
