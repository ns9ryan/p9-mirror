<template>
  <a-card class="general-card" :bordered="false">
    <template #title>
      {{ t('dashboardHome.overview.title') }}
      <span class="title-date">{{ todayText }}</span>
    </template>
    <template #extra>
      <a-button size="small" :loading="loading" @click="fetchData">
        <template #icon><icon-refresh /></template>
      </a-button>
    </template>
    <a-spin :loading="loading" style="width: 100%">
      <a-grid :cols="{ xs: 1, sm: 2, lg: 4 }" :col-gap="16" :row-gap="16">
        <a-grid-item v-for="card in cards" :key="card.key">
          <div class="stat-card">
            <div class="stat-head">
              <span class="stat-icon" :style="{ color: card.color }">
                <component :is="card.icon" />
              </span>
              <span class="stat-title">{{ card.title }}</span>
              <a-tooltip :content="card.tip">
                <icon-question-circle class="stat-tip" />
              </a-tooltip>
            </div>
            <div class="stat-value">
              {{ card.value }}
              <span v-if="card.unit" class="stat-unit">{{ card.unit }}</span>
            </div>
            <div class="stat-compare">
              <span>{{ card.compare.label }}</span>
              <span class="stat-badge" :class="badgeClass(card.compare.rate)">
                {{ formatChange(card.compare.rate) }}
              </span>
            </div>
            <div v-if="card.rows.length" class="stat-rows">
              <div v-for="row in card.rows" :key="row.label" class="stat-row">
                <span>{{ row.label }}</span>
                <span class="stat-row-value">{{ row.value }}</span>
              </div>
            </div>
          </div>
        </a-grid-item>
      </a-grid>
    </a-spin>
  </a-card>
</template>

<script lang="ts" setup>
  import { computed, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { getMemberSummary, MemberSummary } from '@/api/report';
  import useLoading from '@/hooks/loading';
  import { DATE_FMT, formatChange, formatCount, nowIn } from '../utils';

  const props = defineProps({
    /** 查询时区（IANA），为空则不传 timezone 参数 */
    timezone: {
      type: String,
      default: '',
    },
  });

  interface OverviewCard {
    key: string;
    icon: string;
    color: string;
    title: string;
    tip: string;
    value: string;
    unit?: string;
    compare: { label: string; rate: number | null };
    rows: { label: string; value: string }[];
  }

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(false);

  const summary = ref<Partial<MemberSummary>>({});

  const todayText = computed(() => nowIn(props.timezone).format(DATE_FMT));

  function formatAmount(value?: number | null) {
    if (value == null || Number.isNaN(value)) return '-';
    return Number(value).toLocaleString(undefined, {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    });
  }

  const cards = computed<OverviewCard[]>(() => {
    const s = summary.value;
    return [
      {
        key: 'total',
        icon: 'icon-user-group',
        color: 'rgb(var(--arcoblue-6))',
        title: t('dashboardHome.overview.totalMembers'),
        tip: t('dashboardHome.overview.totalTip'),
        value: formatCount(s.total_members),
        compare: {
          label: t('dashboardHome.overview.vsYesterday'),
          rate: s.members_growth_rate ?? null,
        },
        rows: [
          {
            label: t('dashboardHome.overview.activeTotal'),
            value: formatCount(s.active_members),
          },
        ],
      },
      {
        key: 'new',
        icon: 'icon-user-add',
        color: 'rgb(var(--cyan-6))',
        title: t('dashboardHome.overview.newMembers'),
        tip: t('dashboardHome.overview.newTip'),
        value: formatCount(s.today_new_members),
        compare: {
          label: t('dashboardHome.overview.vsLast24h'),
          rate: s.last_24h_new_growth_rate ?? null,
        },
        rows: [
          {
            label: t('dashboardHome.overview.last24hNew'),
            value: formatCount(s.last_24h_new_members),
          },
        ],
      },
      {
        key: 'active',
        icon: 'icon-thunderbolt',
        color: 'rgb(var(--purple-6))',
        title: t('dashboardHome.overview.activeMembers'),
        tip: t('dashboardHome.overview.activeTip'),
        value: formatCount(s.today_active_members),
        compare: {
          label: t('dashboardHome.overview.vsLast24h'),
          rate: s.last_24h_active_growth_rate ?? null,
        },
        rows: [
          {
            label: t('dashboardHome.overview.last24hActive'),
            value: formatCount(s.last_24h_active_members),
          },
        ],
      },
      {
        key: 'deposit',
        icon: 'icon-gift',
        color: 'rgb(var(--orange-6))',
        title: t('dashboardHome.overview.deposit'),
        tip: t('dashboardHome.overview.depositTip'),
        value: formatAmount(s.today_deposit_amount),
        unit: t('dashboardHome.overview.yuan'),
        compare: {
          label: t('dashboardHome.overview.vsYesterday'),
          rate: s.deposit_growth_rate ?? null,
        },
        rows: [],
      },
    ];
  });

  function badgeClass(change: number | null) {
    if (!change) return 'is-flat';
    return change > 0 ? 'is-up' : 'is-down';
  }

  let requestSeq = 0;

  async function fetchData() {
    requestSeq += 1;
    const seq = requestSeq;
    setLoading(true);
    try {
      const data = await getMemberSummary({ timezone: props.timezone });
      // 切换时区过快时，丢弃过期的响应
      if (seq === requestSeq) summary.value = data || {};
    } catch (err) {
      if (seq === requestSeq) Message.error(t('dashboardHome.loadFailed'));
    } finally {
      if (seq === requestSeq) setLoading(false);
    }
  }

  watch(() => props.timezone, fetchData, { immediate: true });
</script>

<style scoped lang="less">
  .title-date {
    margin-left: 8px;
    color: var(--color-text-3);
    font-weight: 400;
    font-size: 13px;
  }

  .stat-card {
    display: flex;
    flex-direction: column;
    height: 100%;
    padding: 16px 20px;
    background: var(--color-fill-1);
    border-radius: 6px;
    transition: box-shadow 0.2s;

    &:hover {
      box-shadow: 0 2px 12px rgb(0 0 0 / 8%);
    }
  }

  .stat-head {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .stat-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    font-size: 16px;
    background: var(--color-bg-2);
    border-radius: 50%;
  }

  .stat-title {
    color: var(--color-text-2);
    font-size: 13px;
  }

  .stat-tip {
    color: var(--color-text-4);
    font-size: 13px;
    cursor: help;
  }

  .stat-value {
    margin: 12px 0 8px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 28px;
    line-height: 36px;
  }

  .stat-unit {
    margin-left: 2px;
    color: var(--color-text-3);
    font-weight: 400;
    font-size: 13px;
  }

  .stat-compare {
    display: flex;
    align-items: center;
    gap: 6px;
    color: var(--color-text-3);
    font-size: 12px;
  }

  .stat-badge {
    padding: 0 6px;
    font-size: 12px;
    line-height: 20px;
    border-radius: 4px;

    &.is-up {
      color: rgb(var(--red-6));
      background: rgb(var(--red-1));
    }

    &.is-down {
      color: rgb(var(--green-6));
      background: rgb(var(--green-1));
    }

    &.is-flat {
      color: var(--color-text-3);
      background: var(--color-fill-3);
    }
  }

  .stat-rows {
    margin-top: auto;
    padding-top: 12px;
  }

  .stat-row {
    display: flex;
    justify-content: space-between;
    padding-top: 10px;
    color: var(--color-text-3);
    font-size: 12px;
    border-top: 1px dashed var(--color-border-2);
  }

  .stat-row-value {
    color: var(--color-text-1);
    font-weight: 500;
    font-size: 13px;
  }
</style>
