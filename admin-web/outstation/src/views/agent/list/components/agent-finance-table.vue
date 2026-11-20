<template>
  <div class="finance-panel">
    <div class="finance-filter">
      <a-space :size="12" wrap>
        <span class="filter-label">
          {{ t('agentList.edit.finance.accountType') }}
        </span>
        <a-select
          v-model="filterType"
          class="type-select"
          allow-clear
          :placeholder="t('agentList.search.all')"
        >
          <a-option
            v-for="item in FINANCE_ACCOUNT_TYPES"
            :key="item.value"
            :value="item.value"
          >
            {{ t(`agentList.edit.finance.type.${item.value}`) }}
          </a-option>
        </a-select>
        <a-button type="primary" @click="handleSearch">
          <template #icon><icon-search /></template>
          {{ t('front.common.search.search') }}
        </a-button>
        <a-button @click="handleReset">
          <template #icon><icon-refresh /></template>
          {{ t('front.common.search.reset') }}
        </a-button>
      </a-space>
      <a-button type="primary" @click="openCreate">
        <template #icon><icon-plus /></template>
        {{ t('agentList.edit.finance.add') }}
      </a-button>
    </div>

    <a-table
      class="finance-table"
      row-key="id"
      :data="pageRows"
      :pagination="false"
      :scroll="{ x: 980 }"
    >
      <template #columns>
        <a-table-column
          :title="t('agentList.edit.finance.currency')"
          data-index="currency"
          :width="90"
        />
        <a-table-column
          :title="t('agentList.edit.finance.category')"
          data-index="type"
          :width="110"
        >
          <template #cell="{ record }">
            <a-tag size="small" :color="categoryColor(record.type)">
              {{
                t(`agentList.edit.finance.category.${categoryOf(record.type)}`)
              }}
            </a-tag>
          </template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.edit.finance.accountType')"
          data-index="type"
          :width="130"
        >
          <template #cell="{ record }">
            {{ t(`agentList.edit.finance.type.${record.type}`) }}
          </template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.edit.finance.account')"
          data-index="account"
          :min-width="220"
          ellipsis
          tooltip
        />
        <a-table-column
          :title="t('agentList.edit.finance.bankCodeType')"
          data-index="bankCodeType"
          :width="120"
        >
          <template #cell="{ record }">{{
            record.bankCodeType || '--'
          }}</template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.edit.finance.bankCode')"
          data-index="bankCode"
          :width="120"
        >
          <template #cell="{ record }">{{ record.bankCode || '--' }}</template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.edit.finance.bankBranch')"
          data-index="bankBranch"
          :width="140"
        >
          <template #cell="{ record }">{{
            record.bankBranch || '--'
          }}</template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.edit.finance.action')"
          :width="100"
          align="center"
          fixed="right"
        >
          <template #cell="{ record }">
            <a-popconfirm
              :content="t('agentList.edit.finance.deleteConfirm')"
              @ok="removeRow(record.id)"
            >
              <a-button type="text" size="mini" status="danger">
                {{ t('agentList.edit.finance.delete') }}
              </a-button>
            </a-popconfirm>
          </template>
        </a-table-column>
      </template>
      <template #empty>
        <a-empty :description="t('agentList.edit.empty')" />
      </template>
    </a-table>

    <div class="finance-pagination">
      <a-pagination
        :total="filteredRows.length"
        :current="pagination.current"
        :page-size="pagination.pageSize"
        :page-size-options="[10, 20, 50]"
        show-total
        show-page-size
        @change="pagination.current = $event"
        @page-size-change="onPageSizeChange"
      />
    </div>

    <a-modal
      v-model:visible="createVisible"
      :title="t('agentList.edit.finance.addTitle')"
      :width="480"
      :mask-closable="false"
      unmount-on-close
      modal-class="mac-modal-panel"
      :ok-text="t('front.common.okText.save')"
      @before-ok="handleCreate"
    >
      <a-form :model="draft" layout="vertical" class="finance-form">
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item
              :label="t('agentList.edit.finance.accountType')"
              required
            >
              <a-select
                v-model="draft.type"
                :placeholder="t('agentList.search.all')"
              >
                <a-option
                  v-for="item in FINANCE_ACCOUNT_TYPES"
                  :key="item.value"
                  :value="item.value"
                >
                  {{ t(`agentList.edit.finance.type.${item.value}`) }}
                </a-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item :label="t('agentList.edit.finance.currency')" required>
              <a-input v-model="draft.currency" allow-clear />
            </a-form-item>
          </a-col>
          <a-col :span="24">
            <a-form-item :label="t('agentList.edit.finance.account')" required>
              <a-input v-model="draft.account" allow-clear />
            </a-form-item>
          </a-col>
          <template v-if="draft.type === 'bank'">
            <a-col :span="12">
              <a-form-item :label="t('agentList.edit.finance.bankCodeType')">
                <a-select v-model="draft.bankCodeType" allow-clear>
                  <a-option
                    v-for="item in BANK_CODE_TYPES"
                    :key="item"
                    :value="item"
                  >
                    {{ item }}
                  </a-option>
                </a-select>
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item :label="t('agentList.edit.finance.bankCode')">
                <a-input v-model="draft.bankCode" allow-clear />
              </a-form-item>
            </a-col>
            <a-col :span="24">
              <a-form-item :label="t('agentList.edit.finance.bankBranch')">
                <a-input v-model="draft.bankBranch" allow-clear />
              </a-form-item>
            </a-col>
          </template>
        </a-row>
      </a-form>
    </a-modal>
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { BANK_CODE_TYPES, FINANCE_ACCOUNT_TYPES } from '../options';

  type FinanceType = (typeof FINANCE_ACCOUNT_TYPES)[number]['value'];

  interface FinanceRow {
    id: number;
    type: FinanceType;
    currency: string;
    account: string;
    bankCodeType: string;
    bankCode: string;
    bankBranch: string;
  }

  const CATEGORY_COLOR: Record<string, string> = {
    bank: 'arcoblue',
    crypto: 'orangered',
    ewallet: 'green',
  };

  const props = defineProps<{
    defaultCurrency?: string;
  }>();

  const { t } = useI18n();

  const rows = ref<FinanceRow[]>([]);
  const filterType = ref<FinanceType | undefined>();
  const appliedType = ref<FinanceType | undefined>();
  const pagination = reactive({ current: 1, pageSize: 10 });
  const createVisible = ref(false);
  let rowSeed = 0;

  function createDraft() {
    return {
      type: undefined as FinanceType | undefined,
      currency: props.defaultCurrency || '',
      account: '',
      bankCodeType: undefined as string | undefined,
      bankCode: '',
      bankBranch: '',
    };
  }

  const draft = reactive(createDraft());

  const filteredRows = computed(() =>
    appliedType.value
      ? rows.value.filter((item) => item.type === appliedType.value)
      : rows.value
  );

  const pageRows = computed(() => {
    const start = (pagination.current - 1) * pagination.pageSize;
    return filteredRows.value.slice(start, start + pagination.pageSize);
  });

  function categoryOf(type: FinanceType) {
    return (
      FINANCE_ACCOUNT_TYPES.find((item) => item.value === type)?.category ||
      'bank'
    );
  }

  function categoryColor(type: FinanceType) {
    return CATEGORY_COLOR[categoryOf(type)];
  }

  function handleSearch() {
    appliedType.value = filterType.value;
    pagination.current = 1;
  }

  function handleReset() {
    filterType.value = undefined;
    handleSearch();
  }

  function onPageSizeChange(size: number) {
    pagination.pageSize = size;
    pagination.current = 1;
  }

  function openCreate() {
    Object.assign(draft, createDraft());
    createVisible.value = true;
  }

  function handleCreate(done: (closed: boolean) => void) {
    if (!draft.type) {
      Message.warning(t('agentList.edit.finance.typeRequired'));
      done(false);
      return;
    }
    const currency = draft.currency.trim();
    const account = draft.account.trim();
    if (!currency) {
      Message.warning(t('agentList.edit.finance.currencyRequired'));
      done(false);
      return;
    }
    if (!account) {
      Message.warning(t('agentList.edit.finance.accountRequired'));
      done(false);
      return;
    }
    const isBank = draft.type === 'bank';
    rowSeed += 1;
    rows.value = [
      {
        id: rowSeed,
        type: draft.type,
        currency: currency.toUpperCase(),
        account,
        bankCodeType: isBank ? draft.bankCodeType || '' : '',
        bankCode: isBank ? draft.bankCode.trim() : '',
        bankBranch: isBank ? draft.bankBranch.trim() : '',
      },
      ...rows.value,
    ];
    Message.success(t('agentList.edit.finance.added'));
    done(true);
  }

  function removeRow(id: number) {
    rows.value = rows.value.filter((item) => item.id !== id);
    const maxPage = Math.max(
      1,
      Math.ceil(filteredRows.value.length / pagination.pageSize)
    );
    if (pagination.current > maxPage) pagination.current = maxPage;
    Message.success(t('agentList.edit.finance.deleted'));
  }
</script>

<style scoped lang="less">
  .finance-panel {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .finance-filter {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    align-items: center;
    justify-content: space-between;
    padding: 14px 16px;
    background: var(--color-fill-1);
    border-radius: 8px;
  }

  .filter-label {
    color: var(--color-text-2);
    font-size: 13px;
  }

  .type-select {
    width: 180px;

    :deep(.arco-select-view) {
      background: var(--color-bg-2);
    }
  }

  .finance-table {
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

  .finance-pagination {
    display: flex;
    justify-content: flex-end;
  }

  .finance-form :deep(.arco-form-item) {
    margin-bottom: 16px;
  }
</style>
