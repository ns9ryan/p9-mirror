<template>
  <a-table
    class="member-table"
    row-key="id"
    :data="data"
    :loading="loading"
    :pagination="false"
    :bordered="{ cell: false }"
    :scroll="{ x: compact ? 960 : 1120, y: '100%' }"
    :scrollbar="false"
    :row-selection="ROW_SELECTION"
    :selected-keys="selectedKeys"
    :row-class="rowClass"
    @selection-change="onSelectionChange"
    @row-click="onRowClick"
  >
    <template #empty>
      <a-empty :description="t('memberManage.emptyList')" />
    </template>
    <template #columns>
      <a-table-column
        :title="t('memberManage.column.memberId')"
        data-index="code"
        :width="130"
      >
        <template #cell="{ record }">
          <span class="num member-id">{{ displayText(record.code) }}</span>
        </template>
      </a-table-column>
      <a-table-column :title="t('memberManage.column.account')" :width="150">
        <template #cell="{ record }">
          <div class="stack">
            <span class="account-line">
              <strong class="ellipsis" :title="record.account">
                {{ displayText(record.account) }}
              </strong>
              <a-tooltip v-if="record.remark" :content="record.remark">
                <icon-message class="remark-icon" @click.stop />
              </a-tooltip>
            </span>
            <small class="ellipsis" :title="record.nickname">
              {{ displayText(record.nickname) }}
            </small>
          </div>
        </template>
      </a-table-column>
      <a-table-column :title="t('memberManage.column.agent')" :width="120">
        <template #cell="{ record }">
          <span class="ellipsis" :title="record.parentAgent">
            {{ displayText(record.parentAgent) }}
          </span>
        </template>
      </a-table-column>
      <a-table-column :title="t('memberManage.column.vip')" :width="80">
        <template #cell="{ record }">
          <a-tag size="small" :color="vipColor(record.vipLevel)">
            VIP{{ record.vipLevel }}
          </a-tag>
        </template>
      </a-table-column>
      <a-table-column :title="t('memberManage.column.status')" :width="80">
        <template #cell="{ record }">
          <a-tag size="small" :color="STATUS_COLORS[asRecord(record).status]">
            {{ t(`memberManage.status.${record.status}`) }}
          </a-tag>
        </template>
      </a-table-column>
      <a-table-column
        v-for="column in AMOUNT_COLUMNS"
        :key="column.key"
        :title="t(column.title)"
        align="right"
        :width="110"
      >
        <template #cell="{ record }">
          <span
            class="num"
            :class="column.signed && winLossClass(asRecord(record)[column.key])"
          >
            {{ formatAmount(asRecord(record)[column.key]) }}
          </span>
        </template>
      </a-table-column>
      <a-table-column
        v-if="!compact"
        :title="t('memberManage.column.registeredAt')"
        data-index="registeredAt"
        :width="160"
      >
        <template #cell="{ record }">
          <span class="num muted">{{ record.registeredAt }}</span>
        </template>
      </a-table-column>
      <a-table-column
        :title="t('memberManage.column.action')"
        :width="140"
        fixed="right"
        align="center"
      >
        <template #cell="{ record }">
          <div class="ops" @click.stop>
            <a-button size="mini" @click="emit('detail', asRecord(record))">
              {{ t('memberManage.v2.action.detail') }}
            </a-button>
            <a-dropdown
              position="br"
              :popup-max-height="false"
              @select="(value) => onCommand(asRecord(record), value)"
            >
              <a-button size="mini">
                <template #icon><icon-more /></template>
              </a-button>
              <template #content>
                <a-doption value="edit">
                  {{ t('memberManage.v2.action.edit') }}
                </a-doption>
                <a-doption
                  v-for="item in OPERATION_COMMANDS"
                  :key="item.key"
                  :value="item.key"
                >
                  {{ t(item.label) }}
                </a-doption>
                <a-doption
                  v-for="command in statusCommandsOf(record.status)"
                  :key="command"
                  :value="command"
                >
                  {{ t(STATUS_COMMANDS[command].label) }}
                </a-doption>
              </template>
            </a-dropdown>
          </div>
        </template>
      </a-table-column>
    </template>
  </a-table>
</template>

<script lang="ts" setup>
  import { useI18n } from 'vue-i18n';
  import type { TableData, TableRowSelection } from '@arco-design/web-vue';
  import {
    STATUS_COLORS,
    displayText,
    formatAmount,
    vipColor,
  } from '../../format';
  import type { MemberRecord } from '../../types';
  import { OPERATION_COMMANDS } from '../operations/commands';
  import {
    STATUS_COMMANDS,
    statusCommandsOf,
    type MemberCommand,
  } from './actions';

  const props = defineProps<{
    data: MemberRecord[];
    selectedKeys: number[];
    activeId: number | null;
    loading?: boolean;
    /** 详情面板打开、列表变窄时隐藏次要列 */
    compact?: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'update:selectedKeys', value: number[]): void;
    (e: 'select', record: MemberRecord): void;
    (e: 'detail', record: MemberRecord): void;
    (e: 'command', record: MemberRecord, command: MemberCommand): void;
  }>();

  const { t } = useI18n();

  type AmountKey =
    | 'periodDeposit'
    | 'periodWithdraw'
    | 'periodBet'
    | 'periodWinLoss';

  interface AmountColumn {
    key: AmountKey;
    title: string;
    signed?: boolean;
  }

  const AMOUNT_COLUMNS: AmountColumn[] = [
    { key: 'periodDeposit', title: 'memberManage.column.todayDeposit' },
    { key: 'periodWithdraw', title: 'memberManage.column.todayWithdraw' },
    { key: 'periodBet', title: 'memberManage.column.todayBet' },
    {
      key: 'periodWinLoss',
      title: 'memberManage.v2.column.todayWinLoss',
      signed: true,
    },
  ];

  const ROW_SELECTION: TableRowSelection = {
    type: 'checkbox',
    showCheckedAll: true,
  };

  function asRecord(record: TableData) {
    return record as MemberRecord;
  }

  function rowClass(record: TableData) {
    return asRecord(record).id === props.activeId ? 'is-active' : '';
  }

  function winLossClass(value: number) {
    if (value > 0) return 'is-win';
    if (value < 0) return 'is-loss';
    return '';
  }

  function onSelectionChange(keys: (string | number)[]) {
    emit('update:selectedKeys', keys.map(Number));
  }

  function onRowClick(record: TableData) {
    emit('select', asRecord(record));
  }

  function onCommand(
    record: MemberRecord,
    value: string | number | Record<string, unknown> | undefined
  ) {
    emit('command', record, value as MemberCommand);
  }
</script>

<style scoped lang="less">
  // scroll.y = '100%' 时表头固定，表体占满父容器剩余高度并独立滚动
  .member-table :deep(.arco-table-container) {
    flex: 1;
    min-height: 0;
  }

  .member-table :deep(.arco-table-body) {
    flex: 1;
    min-height: 0;
  }

  .member-table :deep(.arco-table-tr) {
    cursor: pointer;
  }

  .member-table :deep(.arco-table-tr.is-active .arco-table-td) {
    background: rgb(var(--primary-1));
  }

  .member-table :deep(.arco-table-th) {
    color: var(--color-text-2);
    font-weight: 500;
    background: var(--color-fill-2);
  }

  .member-id {
    font-weight: 600;
  }

  .stack {
    display: flex;
    flex-direction: column;
    min-width: 0;
    line-height: 1.5;
  }

  .stack strong {
    font-weight: 600;
  }

  .account-line {
    display: flex;
    gap: 4px;
    align-items: center;
    min-width: 0;
  }

  .remark-icon {
    flex: none;
    color: rgb(var(--primary-6));
    font-size: 13px;
    cursor: help;
  }

  .stack small {
    color: var(--color-text-3);
    font-size: 12px;
  }

  .ellipsis {
    display: block;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .muted {
    color: var(--color-text-3);
  }

  .num {
    font-variant-numeric: tabular-nums;
  }

  .is-win {
    color: rgb(var(--red-6));
  }

  .is-loss {
    color: rgb(var(--green-6));
  }

  .ops {
    display: inline-flex;
    gap: 6px;
    align-items: center;
  }
</style>
