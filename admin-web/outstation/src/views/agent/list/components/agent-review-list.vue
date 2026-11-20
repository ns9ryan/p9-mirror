<template>
  <div class="review-list">
    <a-form :model="form" layout="inline" class="search-form">
      <a-form-item :label="t('agentList.search.timeRegister')">
        <a-space wrap>
          <a-radio-group
            v-model="quickRange"
            type="button"
            size="small"
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
        <a-input-group compact>
          <a-select v-model="form.accountType" class="search-type">
            <a-option :value="AGENT_APPLY_ACCOUNT_TYPE.AGENT">
              {{ t('agentList.review.keywordType.account') }}
            </a-option>
            <a-option :value="AGENT_APPLY_ACCOUNT_TYPE.MEMBER">
              {{ t('agentList.review.keywordType.member') }}
            </a-option>
          </a-select>
          <a-input
            v-model="form.keyword"
            class="search-input"
            allow-clear
            :placeholder="t('agentList.search.keywordPlaceholder')"
            @press-enter="handleSearch"
          />
        </a-input-group>
      </a-form-item>
      <a-form-item :label="t('agentList.review.status')">
        <a-select v-model="form.auditStatus" class="search-select">
          <a-option :value="0">{{ t('agentList.search.all') }}</a-option>
          <a-option
            v-for="(name, status) in statusKeyMap"
            :key="status"
            :value="Number(status)"
          >
            {{ t(`agentList.review.status.${name}`) }}
          </a-option>
        </a-select>
      </a-form-item>
      <a-form-item>
        <a-space>
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
    </a-form>
    <div class="review-toolbar">
      <a-button type="primary" @click="settingsVisible = true">
        {{ t('agentList.review.applySettings') }}
      </a-button>
    </div>
    <AgentApplySettingsModal
      :visible="settingsVisible"
      @update:visible="settingsVisible = $event"
    />

    <a-table
      row-key="id"
      :data="rows"
      :loading="loading"
      :pagination="false"
      :bordered="{ cell: true }"
      :scroll="{ x: 1450 }"
    >
      <template #columns>
        <a-table-column
          :title="t('agentList.review.col.account')"
          data-index="account"
          :width="130"
        />
        <a-table-column
          :title="t('agentList.review.col.realName')"
          data-index="realname"
          :width="120"
        >
          <template #cell="{ record }">{{ record.realname || '--' }}</template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.review.col.phone')"
          data-index="mobile"
          :width="140"
        >
          <template #cell="{ record }">{{ record.mobile || '--' }}</template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.review.col.source')"
          data-index="register_source"
          :width="150"
        >
          <template #cell="{ record }">
            {{
              record.register_source
                ? t(`agentList.edit.source.${record.register_source}`)
                : '--'
            }}
          </template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.review.col.ratio')"
          data-index="apply_share_rate"
          :width="120"
        >
          <template #cell="{ record }">
            {{ Number(record.apply_share_rate || 0).toFixed(2) }}%
          </template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.review.col.member')"
          data-index="member_account"
          :width="130"
        >
          <template #cell="{ record }">
            {{ record.member_account || '--' }}
          </template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.review.col.status')"
          data-index="audit_status"
          :width="100"
        >
          <template #cell="{ record }">
            {{ statusText(record.audit_status) }}
          </template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.review.col.remark')"
          data-index="apply_remark"
          :width="160"
        >
          <template #cell="{ record }">
            {{ record.apply_remark || '--' }}
          </template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.review.col.registerTime')"
          data-index="register_at"
          :width="170"
        >
          <template #cell="{ record }">
            {{ formatUnix(record.register_at) || '--' }}
          </template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.review.col.operateTime')"
          data-index="audited_at"
          :width="170"
        >
          <template #cell="{ record }">
            {{ formatUnix(record.audited_at) || '--' }}
          </template>
        </a-table-column>
        <a-table-column
          :title="t('agentList.review.col.action')"
          :width="120"
          fixed="right"
        >
          <template #cell="{ record }">
            <template v-if="record.audit_status === AGENT_APPLY_AUDIT.PENDING">
              <a-link @click="confirmApprove(record)">
                {{ t('agentList.review.approve') }}
              </a-link>
              <a-link status="danger" @click="openReject(record)">
                {{ t('agentList.review.reject') }}
              </a-link>
            </template>
            <a-popover
              v-else-if="isRejected(record.audit_status)"
              trigger="click"
              :content="record.reject_reason || '--'"
            >
              <a-link>{{ t('agentList.review.detail') }}</a-link>
            </a-popover>
            <span v-else>--</span>
          </template>
        </a-table-column>
      </template>
    </a-table>

    <div class="review-pagination">
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

    <a-modal
      v-model:visible="rejectForm.visible"
      :title="t('agentList.review.rejectTitle')"
      :width="440"
      unmount-on-close
      modal-class="mac-modal-panel"
      :ok-text="t('agentList.confirm.ok')"
      :cancel-text="t('agentList.confirm.cancel')"
      :on-before-ok="submitReject"
    >
      <a-form :model="rejectForm" layout="vertical">
        <a-form-item :label="t('agentList.review.col.account')">
          {{ rejectForm.account }}
        </a-form-item>
        <a-form-item :label="t('agentList.review.rejectType')">
          <a-radio-group v-model="rejectForm.action">
            <a-radio :value="AGENT_APPLY_ACTION.REJECT">
              {{ t('agentList.review.rejectManual') }}
            </a-radio>
            <a-radio :value="AGENT_APPLY_ACTION.RISK_REJECT">
              {{ t('agentList.review.rejectRisk') }}
            </a-radio>
          </a-radio-group>
        </a-form-item>
        <a-form-item :label="t('agentList.review.rejectReason')" required>
          <a-textarea
            v-model="rejectForm.reason"
            :max-length="200"
            show-word-limit
            :auto-size="{ minRows: 3, maxRows: 6 }"
            :placeholder="t('agentList.review.rejectReasonPlaceholder')"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal
      v-model:visible="passwordResult.visible"
      :title="t('agentList.review.passwordTitle')"
      :width="440"
      :mask-closable="false"
      :closable="false"
      hide-cancel
      unmount-on-close
      modal-class="mac-modal-panel"
      :ok-text="t('agentList.edit.reset.done')"
      @close="passwordResult.password = ''"
    >
      <div class="password-result">
        <a-alert type="warning" :show-icon="true">
          {{ t('agentList.review.passwordTip') }}
        </a-alert>
        <div class="password-label">
          {{
            t('agentList.review.passwordLabel', {
              account: passwordResult.account,
            })
          }}
        </div>
        <div class="password-box">
          <span>{{ passwordResult.password }}</span>
          <a-button
            type="primary"
            size="small"
            @click="copy(passwordResult.password)"
          >
            <template #icon><icon-copy /></template>
            {{ t('agentList.edit.reset.copy') }}
          </a-button>
        </div>
      </div>
    </a-modal>
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
  import { Message, Modal } from '@arco-design/web-vue';
  import {
    AGENT_APPLY_ACCOUNT_TYPE,
    AGENT_APPLY_ACTION,
    AGENT_APPLY_AUDIT,
    listAgentApplies,
    reviewAgentApply,
    type AgentApplyItem,
    type AgentApplyListParams,
  } from '@/api/agent';
  import { formatUnix } from '../mapper';
  import AgentApplySettingsModal from './agent-apply-settings-modal.vue';

  type QuickRange = 'today' | 'd7' | 'd30';

  const emit = defineEmits<{ (e: 'pendingChange', n: number): void }>();
  const settingsVisible = ref(false);
  const { t } = useI18n();
  const statusKeyMap: Record<number, string> = {
    [AGENT_APPLY_AUDIT.PENDING]: 'pending',
    [AGENT_APPLY_AUDIT.APPROVED]: 'approved',
    [AGENT_APPLY_AUDIT.REJECTED]: 'rejected',
    [AGENT_APPLY_AUDIT.RISK_REJECTED]: 'riskRejected',
  };
  const quickOffset: Record<QuickRange, number> = { today: 0, d7: 6, d30: 29 };

  function statusText(status: number) {
    const name = statusKeyMap[status];
    return name ? t(`agentList.review.status.${name}`) : '--';
  }

  function createForm() {
    return {
      accountType: AGENT_APPLY_ACCOUNT_TYPE.AGENT as number,
      keyword: '',
      auditStatus: 0,
      range: [] as string[],
    };
  }

  const form = reactive(createForm());
  const quickRange = ref<QuickRange | ''>('');
  const loading = ref(false);
  const rows = ref<AgentApplyItem[]>([]);
  const total = ref(0);
  const pagination = reactive({
    current: 1,
    pageSize: 20,
    pageSizeOptions: [10, 20, 50, 100],
  });
  let ticket = 0;

  function buildParams(): AgentApplyListParams {
    const params: AgentApplyListParams = {
      page: pagination.current,
      page_size: pagination.pageSize,
      audit_status: form.auditStatus || 0,
    };
    const keyword = form.keyword.trim();
    if (keyword) {
      params.account_type = form.accountType;
      params.account = keyword;
    }
    if (form.range?.length === 2) {
      params.time_start = siteDayStartSec(form.range[0]);
      params.time_end = siteDayEndSec(form.range[1]);
    }
    return params;
  }

  async function loadList() {
    ticket += 1;
    const current = ticket;
    loading.value = true;
    try {
      const { data } = await listAgentApplies(buildParams());
      if (current !== ticket) return;
      rows.value = data?.list || [];
      total.value = data?.total || 0;
    } catch {
      if (current !== ticket) return;
      rows.value = [];
      total.value = 0;
    } finally {
      if (current === ticket) loading.value = false;
    }
  }

  function loadPendingCount() {
    listAgentApplies(
      { page: 1, page_size: 1, audit_status: AGENT_APPLY_AUDIT.PENDING },
      true
    )
      .then(({ data }) => emit('pendingChange', data?.total || 0))
      .catch(() => undefined);
  }

  function handleSearch() {
    pagination.current = 1;
    loadList();
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
  function handleReset() {
    Object.assign(form, createForm());
    quickRange.value = '';
    handleSearch();
  }
  function isRejected(status: number) {
    return (
      status === AGENT_APPLY_AUDIT.REJECTED ||
      status === AGENT_APPLY_AUDIT.RISK_REJECTED
    );
  }

  const rejectForm = reactive({
    visible: false,
    id: 0,
    account: '',
    action: AGENT_APPLY_ACTION.REJECT as number,
    reason: '',
  });
  const passwordResult = reactive({
    visible: false,
    account: '',
    password: '',
  });

  function afterReview() {
    Message.success(t('agentList.review.done'));
    loadList();
    loadPendingCount();
  }

  function confirmApprove(record: AgentApplyItem) {
    Modal.warning({
      title: t('agentList.review.approveTitle'),
      content: t('agentList.review.approveContent', {
        account: record.account,
      }),
      hideCancel: false,
      okText: t('agentList.confirm.ok'),
      cancelText: t('agentList.confirm.cancel'),
      modalClass: 'mac-modal-panel',
      onBeforeOk: async () => {
        try {
          const { data } = await reviewAgentApply({
            id: record.id,
            action: AGENT_APPLY_ACTION.APPROVE,
          });
          if (data?.password) {
            passwordResult.account = data.apply?.account || record.account;
            passwordResult.password = data.password;
            passwordResult.visible = true;
          }
          afterReview();
          return true;
        } catch {
          return false;
        }
      },
    });
  }

  function openReject(record: AgentApplyItem) {
    rejectForm.id = record.id;
    rejectForm.account = record.account;
    rejectForm.action = AGENT_APPLY_ACTION.REJECT;
    rejectForm.reason = '';
    rejectForm.visible = true;
  }

  async function submitReject() {
    const reason = rejectForm.reason.trim();
    if (!reason) {
      Message.warning(t('agentList.review.rejectReasonRequired'));
      return false;
    }
    try {
      await reviewAgentApply({
        id: rejectForm.id,
        action: rejectForm.action,
        reject_reason: reason,
      });
      afterReview();
      return true;
    } catch {
      return false;
    }
  }

  async function copy(text: string) {
    try {
      await navigator.clipboard.writeText(text);
      Message.success(t('agentList.edit.copied'));
    } catch {
      Message.warning(t('agentList.message.copyFail'));
    }
  }

  onMounted(() => {
    loadList();
    loadPendingCount();
  });
</script>

<style scoped lang="less">
  .review-list {
    flex: 1;
    min-width: 0;
    min-height: 0;
    padding-top: 16px;
    overflow: auto;
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

  .search-form :deep(.range-picker) {
    width: 300px;
  }

  .search-form :deep(.range-picker .arco-picker-input),
  .search-form :deep(.range-picker input) {
    min-width: 0;
  }

  .review-toolbar {
    margin-bottom: 14px;
  }

  .review-list :deep(.arco-table-th),
  .review-list :deep(.arco-table-td) {
    white-space: nowrap;
  }

  .password-result {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .password-label {
    color: var(--color-text-2);
    font-size: 13px;
  }

  .password-box {
    display: flex;
    gap: 12px;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    background: var(--color-fill-1);
    border: 1px dashed var(--color-neutral-4);
    border-radius: 8px;

    span {
      color: var(--color-text-1);
      font-weight: 600;
      font-size: 18px;
      font-family: Consolas, Menlo, monospace;
      word-break: break-all;
    }
  }

  .review-pagination {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }
</style>
