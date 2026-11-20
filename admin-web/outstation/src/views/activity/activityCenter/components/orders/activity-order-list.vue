<template>
  <div class="order-list">
    <a-card class="general-card" :bordered="false">
      <a-form :model="form" layout="inline" class="search-form">
        <a-form-item :label="t(`${PREFIX}.order.search.orderNo`)">
          <a-input
            v-model="form.orderNo"
            class="search-input"
            allow-clear
            :placeholder="t(`${PREFIX}.order.search.orderNoPlaceholder`)"
            @press-enter="handleSearch"
          />
        </a-form-item>
        <a-form-item :label="t(`${PREFIX}.order.search.memberId`)">
          <a-input-number
            v-model="form.memberId"
            class="search-input"
            hide-button
            allow-clear
            :min="1"
            :precision="0"
            :placeholder="t(`${PREFIX}.order.search.memberIdPlaceholder`)"
            @press-enter="handleSearch"
          />
        </a-form-item>
        <a-form-item :label="t(`${PREFIX}.order.search.activityId`)">
          <a-input-number
            v-model="form.activityId"
            class="search-input"
            hide-button
            allow-clear
            :min="1"
            :precision="0"
            :placeholder="t(`${PREFIX}.order.search.activityIdPlaceholder`)"
            @press-enter="handleSearch"
          />
        </a-form-item>
        <a-form-item :label="t(`${PREFIX}.form.activityType`)">
          <a-select
            v-model="form.type"
            class="search-select"
            allow-clear
            :placeholder="t(`${PREFIX}.form.selectDefault`)"
          >
            <a-option
              v-for="item in enumOptions(enumParams, 'type')"
              :key="item.value"
              :value="item.value"
            >
              {{ item.label }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item
          v-if="!isReview"
          :label="t(`${PREFIX}.order.search.status`)"
        >
          <a-select
            v-model="form.status"
            class="search-select"
            allow-clear
            :placeholder="t(`${PREFIX}.form.selectDefault`)"
          >
            <a-option
              v-for="value in ORDER_STATUS_VALUES"
              :key="value"
              :value="value"
            >
              {{ t(`${PREFIX}.order.status.${value}`) }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item
          v-else
          :label="t(`${PREFIX}.order.search.auditStatus`)"
        >
          <a-select
            v-model="form.auditStatus"
            class="search-select"
            allow-clear
            :placeholder="t(`${PREFIX}.form.selectDefault`)"
          >
            <a-option
              v-for="value in AUDIT_STATUS_VALUES"
              :key="value"
              :value="value"
            >
              {{ t(`${PREFIX}.order.auditStatus.${value}`) }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item :label="t(`${PREFIX}.order.search.time`)">
          <a-space>
            <a-select v-model="form.timeField" class="time-field-select">
              <a-option
                v-for="value in TIME_FIELD_VALUES"
                :key="value"
                :value="value"
              >
                {{ t(`${PREFIX}.order.timeField.${value}`) }}
              </a-option>
            </a-select>
            <a-range-picker
              v-model="form.range"
              class="range-picker"
              value-format="YYYY-MM-DD"
              allow-clear
              :placeholder="[
                t(`${PREFIX}.order.search.start`),
                t(`${PREFIX}.order.search.end`),
              ]"
            />
          </a-space>
        </a-form-item>
        <a-form-item>
          <a-space>
            <a-button type="primary" @click="handleSearch">
              <template #icon><icon-search /></template>
              {{ t(`${PREFIX}.form.search`) }}
            </a-button>
            <a-button @click="handleReset">
              <template #icon><icon-refresh /></template>
              {{ t(`${PREFIX}.form.reset`) }}
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </a-card>

    <a-card class="general-card" :bordered="false">
      <div class="table-header">
        <div class="list-title">
          {{ t(`${PREFIX}.order.${isReview ? 'reviewTitle' : 'title'}`) }}
          <span class="total">
            {{ t(`${PREFIX}.order.total`, { count: total }) }}
          </span>
        </div>
        <a-button :loading="loading" @click="load">
          <template #icon><icon-refresh /></template>
          {{ t(`${PREFIX}.order.refresh`) }}
        </a-button>
      </div>
      <a-table
        class="order-table"
        row-key="order_no"
        :columns="columns"
        :data="orders"
        :loading="loading"
        :pagination="false"
        :bordered="false"
      >
        <template #orderNo="{ record }">
          <a-typography-text copyable class="nowrap mono">
            {{ record.order_no }}
          </a-typography-text>
        </template>
        <template #activity="{ record }">
          <div class="stack">
            <span class="ellipsis strong" :title="record.activity_name">
              {{ displayText(record.activity_name) }}
            </span>
            <small class="muted">
              {{ typeText(record.type, enumParams) }} · ID {{ record.activity_id }}
            </small>
          </div>
        </template>
        <template #memberId="{ record }">
          <span class="mono">{{ record.member_id }}</span>
        </template>
        <template #reward="{ record }">
          <span class="amount">{{ formatAmount(record.rewards_total) }}</span>
        </template>
        <template #payout="{ record }">
          <div class="stack">
            <span>
              {{ enumLabel(enumParams, 'payout_mode', record.payout_mode) }}
            </span>
            <small class="muted">
              {{ enumLabel(enumParams, 'client_type', record.channel_source) }}
            </small>
          </div>
        </template>
        <template #status="{ record }">
          <a-tag size="small" :color="orderStatusColor(record.status)">
            {{ orderStatusText(record.status) }}
          </a-tag>
        </template>
        <template #appliedAt="{ record }">
          <span class="mono nowrap">
            {{ displayText(formatTime(record.applied_at_sec)) }}
          </span>
        </template>
        <template #approvedAt="{ record }">
          <div class="stack">
            <span class="mono nowrap">
              {{ displayText(formatTime(record.approved_at_sec)) }}
            </span>
            <small v-if="record.approver_uid" class="muted">
              {{ uidText(record.approver_uid) }}
            </small>
          </div>
        </template>
        <template #payoutAt="{ record }">
          <div class="stack">
            <span class="mono nowrap">
              {{ displayText(formatTime(record.payout_at_sec)) }}
            </span>
            <small
              v-if="record.claim_expire_at_sec"
              class="muted mono nowrap"
            >
              {{
                t(`${PREFIX}.order.expireAt`, {
                  time: formatTime(record.claim_expire_at_sec),
                })
              }}
            </small>
          </div>
        </template>
        <template #operations="{ record }">
          <a-space :size="8">
            <a-button
              type="primary"
              size="mini"
              @click="openAudit(record, ACTIVITY_AUDIT_ACTION.APPROVE)"
            >
              {{ t(`${PREFIX}.order.action.approve`) }}
            </a-button>
            <a-button
              size="mini"
              status="danger"
              @click="openAudit(record, ACTIVITY_AUDIT_ACTION.REJECT)"
            >
              {{ t(`${PREFIX}.order.action.reject`) }}
            </a-button>
          </a-space>
        </template>
        <template #empty>
          <a-empty :description="t(`${PREFIX}.order.empty`)" />
        </template>
      </a-table>
      <div class="table-pagination">
        <a-pagination
          v-model:current="pageNo"
          v-model:page-size="pageSize"
          :total="total"
          :page-size-options="[10, 20, 50, 100]"
          :page-size-props="{ popupContainer: 'body' }"
          show-total
          show-page-size
        />
      </div>
    </a-card>

    <a-modal
      v-model:visible="auditVisible"
      :title="t(`${PREFIX}.order.audit.title.${auditAction}`)"
      :ok-text="t(`${PREFIX}.confirm.ok`)"
      :cancel-text="t(`${PREFIX}.confirm.cancel`)"
      :ok-button-props="
        auditAction === ACTIVITY_AUDIT_ACTION.REJECT
          ? { status: 'danger' }
          : undefined
      "
      :on-before-ok="submitAudit"
      unmount-on-close
      @close="auditTarget = null"
    >
      <a-form :model="{ remark: auditRemark }" layout="vertical">
        <a-form-item :label="t(`${PREFIX}.order.columns.orderNo`)">
          <span class="mono">{{ auditTarget?.order_no }}</span>
        </a-form-item>
        <a-form-item :label="t(`${PREFIX}.order.columns.rewardsTotal`)">
          <span class="amount">
            {{ formatAmount(auditTarget?.rewards_total) }}
          </span>
        </a-form-item>
        <a-form-item
          :label="t(`${PREFIX}.order.audit.remark`)"
          class="remark-item"
        >
          <a-textarea
            v-model="auditRemark"
            :max-length="200"
            show-word-limit
            :auto-size="{ minRows: 3, maxRows: 5 }"
            :placeholder="
              t(`${PREFIX}.order.audit.remarkPlaceholder.${auditAction}`)
            "
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script lang="ts" setup>
  import { computed, onMounted, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import {
    ACTIVITY_AUDIT_ACTION,
    ACTIVITY_AUDIT_STATUS,
    ACTIVITY_ORDER_STATUS,
    ACTIVITY_TIME_FIELD,
    auditActivityOrder,
    queryActivityAuditList,
    queryActivityOrderList,
    type ActivityAuditListParams,
    type ActivityOrderItem,
    type ActivityOrderListParams,
  } from '@/api/activity';
  import { siteDayEndSec, siteDayStartSec } from '@/utils/site-time';
  import {
    displayText,
    enumLabel,
    enumOptions,
    formatAmount,
    formatTime,
    typeText,
    uidText,
  } from '../../options';
  import { useActivityStore } from '../../store';

  const PREFIX = 'discountCenter.activityCenter';

  const ORDER_STATUS_VALUES = Object.values(ACTIVITY_ORDER_STATUS);
  const AUDIT_STATUS_VALUES = Object.values(ACTIVITY_AUDIT_STATUS);
  const TIME_FIELD_VALUES = Object.values(ACTIVITY_TIME_FIELD);

  const ORDER_STATUS_COLORS: Record<number, string> = {
    [ACTIVITY_ORDER_STATUS.PENDING]: 'orange',
    [ACTIVITY_ORDER_STATUS.APPROVED]: 'arcoblue',
    [ACTIVITY_ORDER_STATUS.REJECTED]: 'red',
    [ACTIVITY_ORDER_STATUS.TO_CLAIM]: 'cyan',
    [ACTIVITY_ORDER_STATUS.CLAIMED]: 'green',
    [ACTIVITY_ORDER_STATUS.EXPIRED]: 'gray',
    [ACTIVITY_ORDER_STATUS.TO_PAYOUT]: 'purple',
    [ACTIVITY_ORDER_STATUS.PAID]: 'green',
  };

  interface FilterForm {
    orderNo: string;
    memberId?: number;
    activityId?: number;
    type?: number;
    status?: number;
    auditStatus?: number;
    timeField: number;
    range: string[];
  }

  type QueryState = Omit<ActivityOrderListParams, 'page' | 'page_size'> &
    Pick<ActivityAuditListParams, 'audit_status'>;

  const props = defineProps<{
    /** review 只看待审核订单，并提供审核操作 */
    mode: 'orders' | 'review';
  }>();

  const emit = defineEmits<{
    (e: 'audited'): void;
  }>();

  const { t } = useI18n();
  const { enumParams } = useActivityStore();
  const isReview = computed(() => props.mode === 'review');

  const createForm = (): FilterForm => ({
    orderNo: '',
    memberId: undefined,
    activityId: undefined,
    type: undefined,
    status: undefined,
    auditStatus: undefined,
    timeField: ACTIVITY_TIME_FIELD.APPLIED,
    range: [],
  });

  const form = reactive<FilterForm>(createForm());
  const query = ref<QueryState>({});
  const orders = ref<ActivityOrderItem[]>([]);
  const total = ref(0);
  const loading = ref(false);
  const pageNo = ref(1);
  const pageSize = ref(20);

  function orderStatusText(status: number) {
    const key = `${PREFIX}.order.status.${status}`;
    return (ORDER_STATUS_VALUES as number[]).includes(status)
      ? t(key)
      : `#${status}`;
  }

  function orderStatusColor(status: number) {
    return ORDER_STATUS_COLORS[status] || 'gray';
  }

  let ticket = 0;
  async function load() {
    ticket += 1;
    const current = ticket;
    loading.value = true;
    try {
      const params = {
        page: pageNo.value,
        page_size: pageSize.value,
        ...query.value,
      };
      const { data } = isReview.value
        ? await queryActivityAuditList(params)
        : await queryActivityOrderList(params);
      if (current !== ticket) return;
      orders.value = data?.list || [];
      total.value = data?.total || 0;
      const maxPage = Math.max(1, Math.ceil(total.value / pageSize.value));
      if (pageNo.value > maxPage) pageNo.value = maxPage;
    } catch {
      if (current !== ticket) return;
      orders.value = [];
      total.value = 0;
    } finally {
      if (current === ticket) loading.value = false;
    }
  }

  watch([pageNo, pageSize], load);

  function buildQuery(): QueryState {
    const state: QueryState = {
      order_no: form.orderNo.trim() || undefined,
      member_id: form.memberId || undefined,
      activity_id: form.activityId || undefined,
      type: form.type || undefined,
    };
    if (isReview.value) state.audit_status = form.auditStatus || undefined;
    else state.status = form.status || undefined;
    // 传了起止时间时必须带 time_field
    if (form.range?.length === 2) {
      state.time_field = form.timeField;
      state.start_sec = siteDayStartSec(form.range[0]);
      state.end_sec = siteDayEndSec(form.range[1]);
    }
    return state;
  }

  function handleSearch() {
    query.value = buildQuery();
    if (pageNo.value === 1) load();
    else pageNo.value = 1;
  }

  function handleReset() {
    Object.assign(form, createForm());
    handleSearch();
  }

  /** 列宽按内容预留，时间和订单号不换行，活动名超长时省略 */
  const columns = computed<TableColumnData[]>(() => {
    const col = (
      key: string,
      extra: Partial<TableColumnData>
    ): TableColumnData => ({
      title: t(`${PREFIX}.order.columns.${key}`),
      ...extra,
    });
    const common: TableColumnData[] = [
      col('orderNo', { slotName: 'orderNo', width: 230 }),
      col('activity', { slotName: 'activity', width: 240 }),
      col('memberId', { slotName: 'memberId', width: 100 }),
      col('rewardsTotal', { slotName: 'reward', width: 130, align: 'right' }),
      col('payoutMode', { slotName: 'payout', width: 200 }),
    ];
    if (isReview.value) {
      return [
        ...common,
        col('appliedAt', { slotName: 'appliedAt', width: 190 }),
        col('operations', {
          slotName: 'operations',
          width: 150,
          align: 'center',
        }),
      ];
    }
    return [
      ...common,
      col('status', { slotName: 'status', width: 110 }),
      col('appliedAt', { slotName: 'appliedAt', width: 190 }),
      col('approvedAt', { slotName: 'approvedAt', width: 190 }),
      col('payoutAt', { slotName: 'payoutAt', width: 200 }),
    ];
  });

  const auditVisible = ref(false);
  const auditAction = ref<number>(ACTIVITY_AUDIT_ACTION.APPROVE);
  const auditTarget = ref<ActivityOrderItem | null>(null);
  const auditRemark = ref('');

  function openAudit(record: ActivityOrderItem, action: number) {
    auditTarget.value = record;
    auditAction.value = action;
    auditRemark.value = '';
    auditVisible.value = true;
  }

  /** 失败时保留弹窗，错误由全局拦截器提示 */
  async function submitAudit() {
    const target = auditTarget.value;
    if (!target) return false;
    try {
      await auditActivityOrder({
        order_no: target.order_no,
        action: auditAction.value,
        remark: auditRemark.value.trim() || undefined,
      });
    } catch {
      return false;
    }
    Message.success(
      t(
        `${PREFIX}.order.message.${
          auditAction.value === ACTIVITY_AUDIT_ACTION.APPROVE
            ? 'approved'
            : 'rejected'
        }`
      )
    );
    emit('audited');
    load();
    return true;
  }

  onMounted(load);
</script>

<style scoped lang="less">
  .general-card {
    margin-top: 16px;
  }

  .general-card :deep(.arco-card-body) {
    padding: 20px;
  }

  .search-form :deep(.arco-form-item) {
    margin-bottom: 12px;
  }

  .search-input {
    width: 180px;
  }

  .search-select {
    width: 160px;
  }

  .time-field-select {
    width: 110px;
  }

  .range-picker {
    width: 260px;
  }

  .table-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
  }

  .list-title {
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 16px;
  }

  .total {
    margin-left: 12px;
    color: var(--color-text-3);
    font-weight: 400;
    font-size: 13px;
  }

  .order-table :deep(.arco-table-th) {
    white-space: nowrap;
  }

  .order-table :deep(.arco-table-td) {
    vertical-align: middle;
  }

  .nowrap {
    white-space: nowrap;
  }

  .mono {
    font-variant-numeric: tabular-nums;
  }

  .stack {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
    line-height: 1.5;
  }

  .stack small {
    font-size: 12px;
  }

  .strong {
    color: var(--color-text-1);
    font-weight: 500;
  }

  .muted {
    color: var(--color-text-3);
  }

  .ellipsis {
    display: block;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .amount {
    color: rgb(var(--primary-6));
    font-weight: 500;
    font-variant-numeric: tabular-nums;
  }

  .remark-item {
    margin-bottom: 0;
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
