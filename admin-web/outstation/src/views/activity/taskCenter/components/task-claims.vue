<template>
  <a-card class="general-card" :bordered="false">
    <a-form :model="form" layout="inline" class="search-form">
      <a-form-item
        :label="t('discountCenter.taskCenter.claims.search.createdAt')"
      >
        <a-space>
          <a-radio-group v-model="quickRange" type="button">
            <a-radio value="today">
              {{ t('discountCenter.taskCenter.claims.search.today') }}
            </a-radio>
            <a-radio value="d7">
              {{ t('discountCenter.taskCenter.claims.search.last7') }}
            </a-radio>
            <a-radio value="d30">
              {{ t('discountCenter.taskCenter.claims.search.last30') }}
            </a-radio>
          </a-radio-group>
          <a-range-picker
            v-model="form.range"
            class="range-picker"
            value-format="YYYY-MM-DD"
            allow-clear
            :placeholder="[
              t('discountCenter.taskCenter.claims.search.start'),
              t('discountCenter.taskCenter.claims.search.end'),
            ]"
            @change="onRangeChange"
          />
        </a-space>
      </a-form-item>
      <a-form-item>
        <a-input-group compact>
          <a-select v-model="form.keywordType" class="search-type">
            <a-option value="memberId">
              {{ t('discountCenter.taskCenter.claims.search.memberId') }}
            </a-option>
            <a-option value="orderNo">
              {{ t('discountCenter.taskCenter.claims.search.orderNo') }}
            </a-option>
          </a-select>
          <a-input
            v-model="form.keyword"
            class="search-input"
            allow-clear
            :placeholder="
              t('discountCenter.taskCenter.claims.search.inputPlaceholder')
            "
            @press-enter="handleSearch"
          />
        </a-input-group>
      </a-form-item>
      <a-form-item :label="t('discountCenter.taskCenter.claims.search.task')">
        <a-select
          v-model="form.taskCode"
          :options="taskOptions"
          allow-clear
          allow-search
          :placeholder="t('discountCenter.taskCenter.search.placeholder')"
        />
      </a-form-item>
      <a-form-item :label="t('discountCenter.taskCenter.search.category')">
        <a-select
          v-model="form.category"
          :options="categoryOptions"
          allow-clear
          :placeholder="t('discountCenter.taskCenter.search.placeholder')"
        />
      </a-form-item>
      <a-form-item :label="t('discountCenter.taskCenter.claims.column.status')">
        <a-select
          v-model="form.status"
          :options="statusOptions"
          allow-clear
          :placeholder="t('discountCenter.taskCenter.search.placeholder')"
        />
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
    <div class="table-header">
      <a-button
        type="primary"
        :disabled="!total"
        :loading="exporting"
        @click="handleExport"
      >
        <template #icon><icon-download /></template>
        {{ t('discountCenter.taskCenter.toolbar.export') }}
      </a-button>
    </div>
    <a-table
      row-key="id"
      :columns="columns"
      :data="records"
      :loading="loading"
      :pagination="false"
      :bordered="{ wrapper: true, headerCell: true }"
    >
      <template #orderNo="{ record }">
        <a-tooltip :content="record.order_no">
          <a-link class="order-link" @click="openDetail(record)">
            {{ record.order_no }}
          </a-link>
        </a-tooltip>
      </template>
      <template #taskName="{ record }">
        {{ taskNames.get(record.task_code) || record.task_code }}
      </template>
      <template #category="{ record }">
        <a-tag :color="CATEGORY_COLORS[record.category as TaskCategory]">
          {{ t(`discountCenter.taskCenter.category.${record.category}`) }}
        </a-tag>
      </template>
      <template #amount="{ record }">
        <span class="amount">{{ formatAmount(record.amount) }}</span>
      </template>
      <template #payoutMode="{ record }">
        <a-tag :color="PAYOUT_MODE_COLORS[record.payout_mode as TaskPayoutMode]">
          {{ t(`discountCenter.taskCenter.payoutMode.${record.payout_mode}`) }}
        </a-tag>
      </template>
      <template #status="{ record }">
        <a-tag :color="REWARD_STATUS_COLORS[record.status]">
          {{ t(`discountCenter.taskCenter.claims.status.${record.status}`) }}
        </a-tag>
      </template>
      <template #createdAt="{ record }">
        {{ formatTime(record.created_at) }}
      </template>
      <template #operations="{ record }">
        <a-space :size="12">
          <a-link class="operation-link" @click="openDetail(record)">
            {{ t('discountCenter.taskCenter.claims.action.detail') }}
          </a-link>
          <a-link
            v-if="record.status === TASK_REWARD_STATUS.PENDING_CLAIM"
            class="operation-link"
            :disabled="dispatching.has(record.id)"
            @click="confirmDispatch(record)"
          >
            {{ t('discountCenter.taskCenter.claims.action.dispatch') }}
          </a-link>
        </a-space>
      </template>
      <template #empty>
        <a-empty :description="t('discountCenter.taskCenter.claims.empty')" />
      </template>
    </a-table>
    <div class="table-pagination">
      <a-pagination
        v-model:current="pageNo"
        v-model:page-size="pageSize"
        :total="total"
        :page-size-options="[15, 30, 50, 100]"
        :page-size-props="{ popupContainer: 'body' }"
        show-total
        show-page-size
      />
    </div>
  </a-card>

  <ClaimDetailDrawer
    :record="detailing"
    :task-name="detailTaskName"
    @close="detailing = null"
  />
</template>

<script lang="ts" setup>
  import { computed, onMounted, reactive, ref, watch } from 'vue';
  import {
    siteDayEndSec,
    siteDayStartSec,
    siteNow,
  } from '@/utils/site-time';
  import { useI18n } from 'vue-i18n';
  import { Message, Modal } from '@arco-design/web-vue';
  import type { SelectOptionData } from '@arco-design/web-vue/es/select/interface';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import {
    queryTaskConfigList,
    queryTaskRewardList,
    TASK_CATEGORIES,
    TASK_REWARD_STATUS,
    updateTaskRewardStatus,
    type TaskCategory,
    type TaskConfigItem,
    type TaskPayoutMode,
    type TaskRewardItem,
    type TaskRewardListParams,
  } from '@/api/task';
  import { exportClaimsCsv, fetchAllPages } from '../export';
  import {
    CATEGORY_COLORS,
    PAYOUT_MODE_COLORS,
    REWARD_STATUS_COLORS,
    formatAmount,
    formatTime,
  } from '../format';
  import ClaimDetailDrawer from './claim-detail-drawer.vue';

  type QuickKey = 'today' | 'd7' | 'd30';
  type KeywordType = 'memberId' | 'orderNo';
  type FilterParams = Omit<TaskRewardListParams, 'page' | 'page_size'>;

  const QUICK_OFFSET: Record<QuickKey, number> = {
    today: 0,
    d7: 6,
    d30: 29,
  };

  const { t } = useI18n();
  const records = ref<TaskRewardItem[]>([]);
  const total = ref(0);
  const loading = ref(false);
  const exporting = ref(false);
  const dispatching = reactive(new Set<number>());
  const tasks = ref<TaskConfigItem[]>([]);
  const detailing = ref<TaskRewardItem | null>(null);
  const quickRange = ref<QuickKey | ''>('');
  const form = reactive({
    range: [] as string[],
    keywordType: 'memberId' as KeywordType,
    keyword: '',
    taskCode: '',
    category: '' as TaskCategory | '',
    status: '' as number | '',
  });
  const query = ref<FilterParams>({});
  const pageNo = ref(1);
  const pageSize = ref(15);

  const taskNames = computed(
    () => new Map(tasks.value.map((item) => [item.task_code, item.name]))
  );

  const detailTaskName = computed(() => {
    const code = detailing.value?.task_code;
    if (!code) return '';
    return taskNames.value.get(code) || code;
  });

  const taskOptions = computed<SelectOptionData[]>(() =>
    tasks.value.map((item) => ({ label: item.name, value: item.task_code }))
  );

  const categoryOptions = computed<SelectOptionData[]>(() =>
    TASK_CATEGORIES.map((value) => ({
      label: t(`discountCenter.taskCenter.category.${value}`),
      value,
    }))
  );

  const statusOptions = computed<SelectOptionData[]>(() =>
    Object.values(TASK_REWARD_STATUS).map((value) => ({
      label: t(`discountCenter.taskCenter.claims.status.${value}`),
      value,
    }))
  );

  const columns = computed<TableColumnData[]>(() => [
    {
      title: t('discountCenter.taskCenter.claims.column.orderNo'),
      slotName: 'orderNo',
      width: 160,
    },
    {
      title: t('discountCenter.taskCenter.claims.column.memberId'),
      dataIndex: 'member_id',
      width: 100,
      align: 'center',
    },
    {
      title: t('discountCenter.taskCenter.column.name'),
      slotName: 'taskName',
      width: 150,
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('discountCenter.taskCenter.column.category'),
      slotName: 'category',
      width: 100,
      align: 'center',
    },
    {
      title: t('discountCenter.taskCenter.claims.column.periodKey'),
      dataIndex: 'period_key',
      width: 90,
      align: 'center',
    },
    {
      title: t('discountCenter.taskCenter.claims.column.amount'),
      slotName: 'amount',
      width: 110,
      align: 'center',
    },
    {
      title: t('discountCenter.taskCenter.column.payoutMode'),
      slotName: 'payoutMode',
      width: 100,
      align: 'center',
    },
    {
      title: t('discountCenter.taskCenter.claims.column.status'),
      slotName: 'status',
      width: 110,
      align: 'center',
    },
    {
      title: t('discountCenter.taskCenter.claims.column.createdAt'),
      slotName: 'createdAt',
      width: 170,
      align: 'center',
    },
    {
      title: t('discountCenter.taskCenter.claims.column.operations'),
      slotName: 'operations',
      width: 150,
      align: 'center',
    },
  ]);

  let ticket = 0;
  async function load() {
    ticket += 1;
    const current = ticket;
    loading.value = true;
    try {
      const { data } = await queryTaskRewardList({
        page: pageNo.value,
        page_size: pageSize.value,
        ...query.value,
      });
      if (current !== ticket) return;
      records.value = data?.list || [];
      total.value = data?.total || 0;
      const maxPage = Math.max(1, Math.ceil(total.value / pageSize.value));
      if (pageNo.value > maxPage) pageNo.value = maxPage;
    } catch {
      if (current !== ticket) return;
      records.value = [];
      total.value = 0;
    } finally {
      if (current === ticket) loading.value = false;
    }
  }

  /** 任务名称只在配置接口里有，订单只带 task_code */
  async function loadTasks() {
    try {
      const list = await fetchAllPages(async (page, size) => {
        const { data } = await queryTaskConfigList({ page, page_size: size });
        return { total: data?.total || 0, list: data?.list || [] };
      });
      tasks.value = list;
    } catch {
      tasks.value = [];
    }
  }

  watch([pageNo, pageSize], load);
  onMounted(() => {
    loadTasks();
    load();
  });

  function buildQuickRange(key: QuickKey): [string, string] {
    return [
      siteNow().subtract(QUICK_OFFSET[key], 'day').format('YYYY-MM-DD'),
      siteNow().format('YYYY-MM-DD'),
    ];
  }

  function buildQuery(): FilterParams | null {
    const params: FilterParams = {
      task_code: form.taskCode || undefined,
      category: form.category || undefined,
      status: form.status || undefined,
    };
    const keyword = form.keyword.trim();
    if (keyword && form.keywordType === 'memberId') {
      if (!/^\d+$/.test(keyword)) {
        Message.warning(t('discountCenter.taskCenter.claims.memberIdInvalid'));
        return null;
      }
      params.member_id = Number(keyword);
    }
    if (keyword && form.keywordType === 'orderNo') params.order_no = keyword;
    const range = form.range || [];
    if (range.length === 2) {
      params.created_at_start = siteDayStartSec(range[0]);
      params.created_at_end = siteDayEndSec(range[1]);
    }
    return params;
  }

  function handleSearch() {
    const params = buildQuery();
    if (!params) return;
    query.value = params;
    if (pageNo.value === 1) load();
    else pageNo.value = 1;
  }

  function handleReset() {
    form.range = [];
    form.keywordType = 'memberId';
    form.keyword = '';
    form.taskCode = '';
    form.category = '';
    form.status = '';
    quickRange.value = '';
    handleSearch();
  }

  watch(quickRange, (key) => {
    if (!key) return;
    form.range = [...buildQuickRange(key)];
    handleSearch();
  });

  function onRangeChange() {
    const range = form.range || [];
    const keys: QuickKey[] = ['today', 'd7', 'd30'];
    quickRange.value =
      keys.find((key) => {
        const [start, end] = buildQuickRange(key);
        return range[0] === start && range[1] === end;
      }) || '';
  }

  async function handleExport() {
    exporting.value = true;
    try {
      const params = query.value;
      const list = await fetchAllPages(async (page, size) => {
        const { data } = await queryTaskRewardList({
          page,
          page_size: size,
          ...params,
        });
        return { total: data?.total || 0, list: data?.list || [] };
      });
      exportClaimsCsv(list, taskNames.value, t);
    } catch {
      // 错误已由拦截器提示
    } finally {
      exporting.value = false;
    }
  }

  function openDetail(record: TaskRewardItem) {
    detailing.value = record;
  }

  /** 接口只允许 待领取(1) → 待派发(2) */
  function confirmDispatch(record: TaskRewardItem) {
    if (
      record.status !== TASK_REWARD_STATUS.PENDING_CLAIM ||
      dispatching.has(record.id)
    ) {
      return;
    }
    Modal.confirm({
      title: t('discountCenter.taskCenter.claims.confirm.dispatchTitle'),
      content: t('discountCenter.taskCenter.claims.confirm.dispatchContent', {
        orderNo: record.order_no,
      }),
      okText: t('discountCenter.taskCenter.claims.confirm.ok'),
      cancelText: t('discountCenter.taskCenter.claims.confirm.cancel'),
      onBeforeOk: async () => {
        dispatching.add(record.id);
        try {
          const { data } = await updateTaskRewardStatus({
            operator_code: record.operator_code,
            id: record.id,
            status: TASK_REWARD_STATUS.PENDING_PAYOUT,
          });
          const next = data?.member_task_reward ?? {
            ...record,
            status: TASK_REWARD_STATUS.PENDING_PAYOUT,
          };
          records.value = records.value.map((row) =>
            row.id === record.id ? next : row
          );
          if (detailing.value?.id === record.id) detailing.value = next;
          Message.success(
            t('discountCenter.taskCenter.claims.message.dispatched')
          );
          return true;
        } catch {
          return false;
        } finally {
          dispatching.delete(record.id);
        }
      },
    });
  }
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

  .table-header {
    display: flex;
    align-items: center;
    margin-bottom: 14px;
  }

  .amount {
    color: rgb(var(--primary-6));
    font-variant-numeric: tabular-nums;
  }

  .order-link {
    display: block;
    max-width: 100%;
    padding: 0;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .operation-link {
    padding: 0;
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
