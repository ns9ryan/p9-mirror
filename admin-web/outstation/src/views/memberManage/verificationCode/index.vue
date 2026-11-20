<template>
  <div class="container list-page">
    <a-card class="general-card search-card" :bordered="false">
      <a-form :model="form" layout="inline" class="search-form">
        <div class="search-fields">
          <a-form-item>
            <a-space :size="12" align="center" wrap>
              <a-select v-model="form.timeField" class="time-type">
                <a-option value="createdAt">
                  {{ t('verificationCode.search.createdAt') }}
                </a-option>
              </a-select>
              <a-radio-group
                v-model="quickRange"
                class="quick-range"
                type="button"
                @change="onQuickChange"
              >
                <a-radio value="today">
                  {{ t('verificationCode.search.today') }}
                </a-radio>
                <a-radio value="d7">
                  {{ t('verificationCode.search.last7') }}
                </a-radio>
                <a-radio value="d30">
                  {{ t('verificationCode.search.last30') }}
                </a-radio>
              </a-radio-group>
              <a-range-picker
                v-model="form.range"
                class="range-picker"
                value-format="YYYY-MM-DD"
                allow-clear
                :placeholder="[
                  t('verificationCode.search.start'),
                  t('verificationCode.search.end'),
                ]"
                @change="onRangeChange"
              />
            </a-space>
          </a-form-item>
          <a-form-item :label="t('verificationCode.search.memberId')">
            <a-input-number
              v-model="form.memberId"
              class="member-input"
              :min="1"
              hide-button
              allow-clear
              :placeholder="t('verificationCode.search.pleaseInput')"
              @press-enter="handleSearch"
            />
          </a-form-item>
          <a-form-item :label="t('verificationCode.search.sendType')">
            <a-select
              v-model="form.channel"
              class="search-select"
              allow-clear
              :placeholder="t('verificationCode.search.pleaseSelect')"
            >
              <a-option :value="CONTACT_CHANNEL.SMS">
                {{ t('verificationCode.channel.sms') }}
              </a-option>
              <a-option :value="CONTACT_CHANNEL.EMAIL">
                {{ t('verificationCode.channel.email') }}
              </a-option>
            </a-select>
          </a-form-item>
          <a-form-item :label="t('verificationCode.search.status')">
            <a-select
              v-model="form.status"
              class="search-select status-select"
              allow-clear
              :placeholder="t('verificationCode.search.pleaseSelect')"
            >
              <a-option
                v-for="value in STATUS_VALUES"
                :key="value"
                :value="value"
              >
                {{ statusLabel(value) }}
              </a-option>
            </a-select>
          </a-form-item>
          <a-form-item class="search-actions">
            <a-space :size="12">
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
        </div>
      </a-form>
    </a-card>

    <a-card class="general-card table-card" :bordered="false">
      <a-table
        row-key="id"
        :loading="loading"
        :data="rows"
        :pagination="false"
        :bordered="{ cell: true }"
        :stripe="true"
        :scroll="{ x: 1280 }"
      >
        <template #columns>
          <a-table-column
            :title="t('verificationCode.column.id')"
            data-index="id"
            :width="100"
          />
          <a-table-column
            :title="t('verificationCode.column.code')"
            :width="140"
          >
            <template #cell="{ record }">
              {{ record.verification_code || '--' }}
            </template>
          </a-table-column>
          <a-table-column
            :title="t('verificationCode.column.memberId')"
            data-index="member_id"
            :width="140"
          />
          <a-table-column
            :title="t('verificationCode.column.channel')"
            :width="120"
          >
            <template #cell="{ record }">
              {{ channelLabel(record.channel) }}
            </template>
          </a-table-column>
          <a-table-column
            :title="t('verificationCode.column.status')"
            :width="130"
          >
            <template #cell="{ record }">
              <a-tag :color="statusColor(record.status)">
                {{ statusLabel(record.status) }}
              </a-tag>
            </template>
          </a-table-column>
          <a-table-column
            :title="t('verificationCode.column.failedAttempts')"
            data-index="failed_attempts"
            :width="110"
          />
          <a-table-column
            :title="t('verificationCode.column.sentAt')"
            :width="190"
          >
            <template #cell="{ record }">
              {{ formatTime(record.sent_at) }}
            </template>
          </a-table-column>
          <a-table-column
            :title="t('verificationCode.column.expiresAt')"
            :width="190"
          >
            <template #cell="{ record }">
              {{ formatTime(record.expires_at) }}
            </template>
          </a-table-column>
          <a-table-column
            :title="t('verificationCode.column.usedAt')"
            :width="190"
          >
            <template #cell="{ record }">
              {{ formatTime(record.used_at) }}
            </template>
          </a-table-column>
        </template>
        <template #empty>
          <a-empty :description="t('verificationCode.empty')" />
        </template>
      </a-table>
      <div class="table-pagination">
        <a-pagination
          :current="pageNo"
          :page-size="pageSize"
          :total="total"
          :page-size-options="[15, 20, 30, 50, 100]"
          show-total
          show-page-size
          show-jumper
          @change="onPageChange"
          @page-size-change="onPageSizeChange"
        />
      </div>
    </a-card>
  </div>
</template>

<script lang="ts" setup>
  import { onMounted, reactive, ref } from 'vue';
  import dayjs from 'dayjs';
  import { useI18n } from 'vue-i18n';
  import {
    CONTACT_BINDING_CODE_STATUS,
    CONTACT_CHANNEL,
    queryContactBindingCodeList,
    type ContactBindingCodeListParams,
    type ContactBindingCodeRecord,
  } from '@/api/member';
  import {
    ensureSiteTimezone,
    formatSiteTime,
    parseSiteTime,
    siteNow,
  } from '@/utils/site-time';

  const { t } = useI18n();

  type QuickKey = 'today' | 'd7' | 'd30';

  const quickOffset: Record<QuickKey, number> = {
    today: 0,
    d7: 6,
    d30: 29,
  };

  /** 渠道文案 key，与接口的 1 短信、2 邮件对应 */
  const CHANNEL_KEYS: Record<number, 'sms' | 'email'> = {
    [CONTACT_CHANNEL.SMS]: 'sms',
    [CONTACT_CHANNEL.EMAIL]: 'email',
  };

  const STATUS_META: Record<number, { key: string; color: string }> = {
    [CONTACT_BINDING_CODE_STATUS.PENDING]: { key: 'pending', color: 'arcoblue' },
    [CONTACT_BINDING_CODE_STATUS.SENT]: { key: 'sent', color: 'green' },
    [CONTACT_BINDING_CODE_STATUS.USED]: { key: 'used', color: 'cyan' },
    [CONTACT_BINDING_CODE_STATUS.VOIDED]: { key: 'voided', color: 'gray' },
    [CONTACT_BINDING_CODE_STATUS.SEND_FAILED]: { key: 'failed', color: 'red' },
  };
  const STATUS_VALUES = Object.keys(STATUS_META).map(Number);

  function channelLabel(channel: number) {
    const key = CHANNEL_KEYS[channel];
    return key ? t(`verificationCode.channel.${key}`) : '--';
  }

  function statusLabel(status: number) {
    const meta = STATUS_META[status];
    return meta ? t(`verificationCode.status.${meta.key}`) : '--';
  }

  function statusColor(status: number) {
    return STATUS_META[status]?.color ?? 'gray';
  }

  /** 按厅时区显示；0 表示未发送、未核销等没有时间的情况，显示 '--' */
  function formatTime(value: number) {
    return formatSiteTime(value);
  }

  /** 快捷日期按厅时区的"今天"计算 */
  function buildQuickRange(key: QuickKey): [string, string] {
    return [
      siteNow().subtract(quickOffset[key], 'day').format('YYYY-MM-DD'),
      siteNow().format('YYYY-MM-DD'),
    ];
  }

  /** 厅时区某一天 00:00 的 Unix 毫秒（parseSiteTime 返回秒，需换算） */
  function dayStartMs(date: string) {
    return parseSiteTime(`${date} 00:00:00`) * 1000;
  }

  const quickRange = ref<QuickKey | ''>('');
  const form = reactive({
    timeField: 'createdAt',
    range: [] as string[],
    memberId: undefined as number | undefined,
    channel: undefined as number | undefined,
    status: undefined as number | undefined,
  });

  /** 已应用的查询条件，点击搜索后才生效 */
  interface QueryState {
    createdFrom?: number;
    createdBefore?: number;
    memberId?: number;
    channel?: number;
    status?: number;
  }
  const query = ref<QueryState>({});

  const rows = ref<ContactBindingCodeRecord[]>([]);
  const total = ref(0);
  const loading = ref(false);
  const pageNo = ref(1);
  const pageSize = ref(15);
  let ticket = 0;

  function toQueryState(): QueryState {
    const [start, end] = form.range;
    return {
      // 创建时间左闭右开：开始日 00:00 起，结束日的次日 00:00 之前（厅时区）
      createdFrom: start ? dayStartMs(start) : undefined,
      createdBefore: end
        ? dayStartMs(dayjs(end).add(1, 'day').format('YYYY-MM-DD'))
        : undefined,
      memberId: form.memberId || undefined,
      channel: form.channel,
      status: form.status,
    };
  }

  async function fetchList() {
    ticket += 1;
    const current = ticket;
    loading.value = true;
    const q = query.value;
    const params: ContactBindingCodeListParams = {
      page: pageNo.value,
      page_size: pageSize.value,
      member_id: q.memberId,
      channel: q.channel,
      status: q.status,
      created_from: q.createdFrom,
      created_before: q.createdBefore,
    };
    try {
      const { data } = await queryContactBindingCodeList(params);
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

  function handleSearch() {
    query.value = toQueryState();
    pageNo.value = 1;
    fetchList();
  }

  function handleReset() {
    form.timeField = 'createdAt';
    form.range = [];
    form.memberId = undefined;
    form.channel = undefined;
    form.status = undefined;
    quickRange.value = '';
    handleSearch();
  }

  /** 快捷时间立即查询 */
  function onQuickChange(value: unknown) {
    if (value !== 'today' && value !== 'd7' && value !== 'd30') return;
    form.range = [...buildQuickRange(value)];
    handleSearch();
  }

  /** 手动选择日期时同步快捷按钮高亮，不自动查询 */
  function onRangeChange() {
    const range = form.range || [];
    const keys: QuickKey[] = ['today', 'd7', 'd30'];
    quickRange.value =
      keys.find((key) => {
        const [start, end] = buildQuickRange(key);
        return range[0] === start && range[1] === end;
      }) || '';
  }

  function onPageChange(page: number) {
    pageNo.value = page;
    fetchList();
  }

  function onPageSizeChange(size: number) {
    pageSize.value = size;
    pageNo.value = 1;
    fetchList();
  }

  onMounted(async () => {
    // 厅时区未加载时，日期筛选和显示会不准确，先确保拿到再查询
    await ensureSiteTimezone();
    fetchList();
  });
</script>

<style scoped lang="less">
  .container {
    padding: 0 20px 20px;
  }

  .general-card {
    margin-top: 16px;
  }

  .search-card :deep(.arco-card-body) {
    padding: 16px 20px;
  }

  .table-card :deep(.arco-card-body) {
    padding: 20px;
  }

  .search-form {
    width: 100%;
  }

  .search-form :deep(.arco-form-item) {
    margin: 0;
  }

  .search-form :deep(.arco-form-item-label-col) {
    padding-right: 8px;
    line-height: 32px;
  }

  .search-fields {
    display: flex;
    flex-wrap: wrap;
    gap: 12px 16px;
    align-items: center;
  }

  .time-type {
    width: 112px;
  }

  .time-type {
    :deep(.arco-select-view-single) {
      color: #fff;
      background: rgb(var(--primary-6));
      border-color: rgb(var(--primary-6));
    }

    :deep(.arco-select-view-value),
    :deep(.arco-select-view-suffix),
    :deep(.arco-select-view-icon) {
      color: #fff;
    }

    :deep(.arco-select-view-single:hover),
    :deep(.arco-select-view-single.arco-select-view-focus) {
      color: #fff;
      background: rgb(var(--primary-5));
      border-color: rgb(var(--primary-5));
    }
  }

  .member-input {
    width: 180px;
  }

  .search-select {
    width: 148px;
  }

  .range-picker {
    width: 280px;
  }

  .quick-range {
    :deep(.arco-radio-group-button) {
      box-sizing: border-box;
      height: 32px;
      padding: 2px;
      line-height: 26px;
      background: var(--color-fill-2);
    }

    :deep(.arco-radio-button) {
      height: 26px;
      margin: 0;
      line-height: 26px;
    }

    :deep(.arco-radio-button-content) {
      padding: 0 14px;
      line-height: 26px;
    }
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
