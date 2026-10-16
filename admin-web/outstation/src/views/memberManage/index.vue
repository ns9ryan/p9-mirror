<template>
  <div class="container list-page">
    <Breadcrumb :items="['menu.memberManage', 'menu.memberManage.list']" />
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
          <a-input-group compact>
            <a-select v-model="form.keywordType" class="search-type">
              <a-option value="id">
                {{ t('memberManage.search.keyword.id') }}
              </a-option>
              <a-option value="account">
                {{ t('memberManage.search.keyword.account') }}
              </a-option>
              <a-option value="phone">
                {{ t('memberManage.search.keyword.phone') }}
              </a-option>
              <a-option value="email">
                {{ t('memberManage.search.keyword.email') }}
              </a-option>
            </a-select>
            <a-input
              v-model="form.keyword"
              class="search-input"
              allow-clear
              :placeholder="t('memberManage.search.keywordPlaceholder')"
              @press-enter="handleSearch"
            />
          </a-input-group>
        </a-form-item>
        <a-form-item :label="t('memberManage.search.autoLevel')">
          <a-select v-model="form.autoLevel" class="search-select">
            <a-option value="">{{ t('memberManage.search.all') }}</a-option>
            <a-option
              v-for="item in autoLevelOptions"
              :key="item"
              :value="item"
            >
              {{ item }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item :label="t('memberManage.search.fixedLevel')">
          <a-select v-model="form.fixedLevel" class="search-select">
            <a-option value="">{{ t('memberManage.search.all') }}</a-option>
            <a-option
              v-for="item in fixedLevelOptions"
              :key="item"
              :value="item"
            >
              {{ item }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item :label="t('memberManage.search.registerSource')">
          <a-select v-model="form.registerSource" class="search-select">
            <a-option value="">{{ t('memberManage.search.all') }}</a-option>
            <a-option
              v-for="item in REGISTER_SOURCES"
              :key="item"
              :value="item"
            >
              {{ item }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item :label="t('memberManage.search.status')">
          <a-select v-model="form.status" class="search-select">
            <a-option value="">{{ t('memberManage.search.all') }}</a-option>
            <a-option value="normal">
              {{ t('memberManage.status.normal') }}
            </a-option>
            <a-option value="disabled">
              {{ t('memberManage.status.disabled') }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item :label="t('memberManage.search.channel')">
          <a-select v-model="form.channelSource" class="search-select">
            <a-option value="">{{ t('memberManage.search.all') }}</a-option>
            <a-option v-for="item in channelOptions" :key="item" :value="item">
              {{ item }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item :label="t('memberManage.search.loginState')">
          <a-select v-model="form.loginState" class="search-select">
            <a-option value="">{{ t('memberManage.search.all') }}</a-option>
            <a-option value="today">
              {{ t('memberManage.search.loginToday') }}
            </a-option>
            <a-option value="d7">{{
              t('memberManage.search.login7')
            }}</a-option>
            <a-option value="d30">
              {{ t('memberManage.search.login30') }}
            </a-option>
            <a-option value="over30">
              {{ t('memberManage.search.loginOver30') }}
            </a-option>
            <a-option value="never">
              {{ t('memberManage.search.loginNever') }}
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
            <a-checkbox v-model="form.deposited">
              {{ t('memberManage.search.deposited') }}
            </a-checkbox>
            <a-checkbox v-model="form.notDeposited">
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
          <a-button :disabled="!filtered.length" @click="exportCsv">
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
      </div>

      <div class="member-table-wrap">
        <a-table
          row-key="id"
          class="member-table"
          :data="pageData"
          :pagination="false"
          :bordered="{ cell: true }"
          :show-header="false"
        >
          <template #columns>
            <a-table-column data-index="id">
              <template #cell="{ record }">
                <MemberCard
                  :record="record"
                  :selected="selectedIds.has(record.id)"
                  :expanded="isExpanded(record.id)"
                  @update:selected="
                    (checked) => setSelected(record.id, checked)
                  "
                  @update:expanded="(value) => setExpanded(record.id, value)"
                  @detail="openDetail(record)"
                  @edit="openEdit(record)"
                  @toggle-status="confirmToggle(record)"
                />
              </template>
            </a-table-column>
          </template>
          <template #empty>
            <a-empty :description="t('memberManage.emptyList')" />
          </template>
        </a-table>

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

    <MemberDetailDrawer
      :visible="detailVisible"
      :record="detailRecord"
      @update:visible="detailVisible = $event"
      @edit="editFromDetail"
    />
    <MemberFormModal
      :visible="formVisible"
      :mode="formMode"
      :record="editingRecord"
      @update:visible="formVisible = $event"
      @submit="onSubmit"
    />
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import dayjs from 'dayjs';
  import { useI18n } from 'vue-i18n';
  import { Message, Modal } from '@arco-design/web-vue';
  import MemberCard from './components/member-card.vue';
  import MemberDetailDrawer from './components/member-detail-drawer.vue';
  import MemberFormModal from './components/member-form-modal.vue';
  import { formatAmount, formatAmountCount } from './format';
  import createMemberSeed from './mock';
  import { AUTO_LEVELS, FIXED_LEVELS, REGISTER_SOURCES } from './options';
  import {
    buildQuickRange,
    createMemberQuery,
    detectQuickRange,
    matchMember,
  } from './query';
  import type {
    AccountStatus,
    MemberFormValue,
    MemberQuery,
    MemberRecord,
    QuickRange,
  } from './types';

  const { t } = useI18n();

  const members = ref<MemberRecord[]>(createMemberSeed());
  const form = reactive<MemberQuery>(createMemberQuery());
  const applied = ref<MemberQuery>(createMemberQuery());
  const quickRange = ref<QuickRange | ''>('');
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

  const detailVisible = ref(false);
  const detailId = ref<number | null>(null);
  const formVisible = ref(false);
  const formMode = ref<'create' | 'edit'>('create');
  const editingId = ref<number | null>(null);

  const autoLevelOptions = computed(() => {
    const values = new Set<string>(AUTO_LEVELS);
    members.value.forEach((item) => {
      if (item.levelKind === 'auto' && item.level) values.add(item.level);
    });
    return [...values];
  });

  const fixedLevelOptions = computed(() => {
    const values = new Set<string>(FIXED_LEVELS);
    members.value.forEach((item) => {
      if (item.levelKind === 'fixed' && item.level) values.add(item.level);
    });
    return [...values];
  });

  const channelOptions = computed(() => {
    const values = new Set<string>();
    members.value.forEach((item) => {
      if (item.channelSource) values.add(item.channelSource);
    });
    return [...values];
  });

  const filtered = computed(() =>
    members.value.filter((item) => matchMember(item, applied.value))
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

  const detailRecord = computed(
    () => members.value.find((item) => item.id === detailId.value) || null
  );

  const editingRecord = computed(
    () => members.value.find((item) => item.id === editingId.value) || null
  );

  watch(filtered, (list) => {
    const maxPage = Math.max(1, Math.ceil(list.length / pagination.pageSize));
    if (pagination.current > maxPage) pagination.current = maxPage;
  });

  function snapshotQuery(): MemberQuery {
    return { ...form, dateRange: [...(form.dateRange || [])] };
  }

  function handleSearch() {
    applied.value = snapshotQuery();
    pagination.current = 1;
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

  function applyStatus(ids: number[], status: AccountStatus) {
    const idSet = new Set(ids);
    let count = 0;
    members.value = members.value.map((item) => {
      if (!idSet.has(item.id) || item.status === status) return item;
      count += 1;
      return { ...item, status };
    });
    return count;
  }

  function confirmToggle(record: MemberRecord) {
    const disabling = record.status === 'normal';
    Modal.confirm({
      title: t(
        disabling
          ? 'memberManage.confirm.disableTitle'
          : 'memberManage.confirm.enableTitle'
      ),
      content: t(
        disabling
          ? 'memberManage.confirm.disableContent'
          : 'memberManage.confirm.enableContent',
        { id: record.id }
      ),
      okText: t('memberManage.confirm.ok'),
      cancelText: t('memberManage.confirm.cancel'),
      onOk: () => {
        applyStatus([record.id], disabling ? 'disabled' : 'normal');
        Message.success(
          t(
            disabling
              ? 'memberManage.message.disabled'
              : 'memberManage.message.enabled'
          )
        );
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
      onOk: () => {
        const count = applyStatus(ids, disabling ? 'disabled' : 'normal');
        if (!count) {
          Message.info(t('memberManage.message.already'));
          return;
        }
        Message.success(
          t(
            disabling
              ? 'memberManage.message.batchDisabled'
              : 'memberManage.message.batchEnabled',
            { count }
          )
        );
      },
    });
  }

  function openDetail(record: MemberRecord) {
    detailId.value = record.id;
    detailVisible.value = true;
  }

  function openEdit(record: MemberRecord) {
    formMode.value = 'edit';
    editingId.value = record.id;
    formVisible.value = true;
  }

  function openCreate() {
    formMode.value = 'create';
    editingId.value = null;
    formVisible.value = true;
  }

  function editFromDetail() {
    if (!detailRecord.value) return;
    detailVisible.value = false;
    openEdit(detailRecord.value);
  }

  function nextMemberId() {
    return members.value.reduce((max, item) => Math.max(max, item.id), 0) + 1;
  }

  function onSubmit(value: MemberFormValue) {
    if (formMode.value === 'create') {
      const now = dayjs().format('YYYY-MM-DD HH:mm:ss');
      Object.assign(form, createMemberQuery());
      quickRange.value = '';
      applied.value = createMemberQuery();
      members.value = [
        {
          ...value,
          id: nextMemberId(),
          balance: 0,
          unsettled: 0,
          totalDepositAmount: 0,
          totalDepositCount: 0,
          totalWithdrawAmount: 0,
          totalWithdrawCount: 0,
          depositWithdrawDiff: 0,
          firstDepositAmount: 0,
          todayDeposit: 0,
          todayWithdraw: 0,
          todayValidBet: 0,
          todayWinLoss: 0,
          todayBonus: 0,
          registeredAt: now,
          lastLoginAt: '',
          lastLoginIp: '',
        },
        ...members.value,
      ];
      pagination.current = 1;
      Message.success(t('memberManage.message.created'));
      return;
    }
    members.value = members.value.map((item) =>
      item.id === editingId.value ? { ...item, ...value } : item
    );
    Message.success(t('memberManage.message.updated'));
  }

  function csvCell(value: string | number) {
    const text = String(value ?? '');
    if (/[",\n]/.test(text)) return `"${text.replace(/"/g, '""')}"`;
    return text;
  }

  function exportCsv() {
    const headers = [
      t('memberManage.card.memberId'),
      t('memberManage.card.account'),
      t('memberManage.card.phone'),
      t('memberManage.card.email'),
      t('memberManage.card.vip'),
      t('memberManage.card.registerType'),
      t('memberManage.card.level'),
      t('memberManage.card.source'),
      t('memberManage.card.channel'),
      t('memberManage.modal.parent'),
      t('memberManage.modal.parentId'),
      t('memberManage.card.balance'),
      t('memberManage.card.unsettled'),
      t('memberManage.detail.totalDeposit'),
      t('memberManage.detail.totalWithdraw'),
      t('memberManage.detail.diff'),
      t('memberManage.card.firstDeposit'),
      t('memberManage.card.todayDeposit'),
      t('memberManage.card.todayWithdraw'),
      t('memberManage.detail.todayBet'),
      t('memberManage.card.todayWin'),
      t('memberManage.card.todayBonus'),
      t('memberManage.card.referral'),
      t('memberManage.card.status'),
      t('memberManage.card.registeredAt'),
      t('memberManage.card.lastLogin'),
      t('memberManage.card.loginIp'),
    ];
    const rows = filtered.value.map((item) => [
      item.id,
      item.account,
      item.phone,
      item.email,
      item.vipLevel,
      t(`memberManage.registerType.${item.registerType}`),
      item.level,
      item.registerSource,
      item.channelSource,
      item.parentAgent,
      item.parentAgentId,
      formatAmount(item.balance),
      formatAmount(item.unsettled),
      formatAmountCount(item.totalDepositAmount, item.totalDepositCount),
      formatAmountCount(item.totalWithdrawAmount, item.totalWithdrawCount),
      formatAmount(item.depositWithdrawDiff),
      formatAmount(item.firstDepositAmount),
      formatAmount(item.todayDeposit),
      formatAmount(item.todayWithdraw),
      formatAmount(item.todayValidBet),
      formatAmount(item.todayWinLoss),
      formatAmount(item.todayBonus),
      item.referralStatus === 'normal' ? t('memberManage.referral.normal') : '',
      t(`memberManage.status.${item.status}`),
      item.registeredAt,
      item.lastLoginAt,
      item.lastLoginIp,
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
    link.download = `members-${dayjs().format('YYYYMMDD-HHmmss')}.csv`;
    link.click();
    URL.revokeObjectURL(url);
    Message.success(t('memberManage.message.exported'));
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

  .table-toolbar {
    margin-bottom: 14px;
  }

  .selected-count {
    color: var(--color-text-2);
    font-size: 13px;
  }

  .member-table-wrap {
    width: 100%;
  }

  .member-table {
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
