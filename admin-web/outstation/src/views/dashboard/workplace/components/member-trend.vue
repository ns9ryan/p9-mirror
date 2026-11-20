<template>
  <a-card class="general-card" :bordered="false">
    <template #title>{{ t('dashboardHome.trend.title') }}</template>
    <div class="toolbar">
      <RangeFilter v-model="range" :timezone="timezone" />
      <a-radio-group v-model="metric" type="button" size="small">
        <a-radio value="register">
          {{ t('dashboardHome.trend.register') }}
        </a-radio>
        <a-radio value="login">{{ t('dashboardHome.trend.login') }}</a-radio>
      </a-radio-group>
    </div>
    <a-spin :loading="loading" style="width: 100%">
      <div class="summary">
        <div class="summary-item">
          <div class="summary-label">{{ labels.current }}</div>
          <div class="summary-value">{{ formatCount(currentTotal) }}</div>
        </div>
        <div class="summary-item">
          <div class="summary-label">{{ labels.previous }}</div>
          <div class="summary-value">{{ formatCount(previousTotal) }}</div>
        </div>
        <div class="summary-item">
          <div class="summary-label">{{ t('dashboardHome.trend.change') }}</div>
          <div class="summary-value" :class="changeClass">
            {{ formatChange(totalChange) }}
          </div>
        </div>
        <div class="summary-item">
          <div class="summary-label">{{ t('dashboardHome.trend.peak') }}</div>
          <div class="summary-value">
            {{ formatCount(peak.value) }}
            <span v-if="peak.label" class="summary-sub">{{ peak.label }}</span>
          </div>
        </div>
      </div>
      <Chart height="320px" :options="chartOption" />
    </a-spin>
  </a-card>
</template>

<script lang="ts" setup>
  import { computed, ref, watch } from 'vue';
  import { graphic } from 'echarts';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { getMemberGrowthTrend, TrendGranularity } from '@/api/report';
  import useLoading from '@/hooks/loading';
  import useChartOption from '@/hooks/chart-option';
  import { ToolTipFormatterParams } from '@/types/echarts';
  import RangeFilter from './range-filter.vue';
  import {
    bucketLabel,
    buildBuckets,
    buildQuickRange,
    changeRate,
    DateRange,
    formatChange,
    formatCount,
    granularityOf,
    detectQuickRange,
    isFutureBucket,
    nowIn,
    previousRange,
    sum,
    toCountMap,
  } from '../utils';

  type Metric = 'register' | 'login';

  const props = defineProps({
    /** 查询时区（IANA），为空则不传 timezone 参数 */
    timezone: {
      type: String,
      default: '',
    },
  });

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(false);

  const range = ref<DateRange>(buildQuickRange('today', nowIn(props.timezone)));
  const metric = ref<Metric>('register');

  const granularity = ref<TrendGranularity>('hour');
  const currentBuckets = ref<string[]>([]);
  const previousBuckets = ref<string[]>([]);
  const currentData = ref<(number | null)[]>([]);
  const previousData = ref<number[]>([]);

  const labels = computed(() => {
    const today = range.value[0] === range.value[1];
    return {
      current: t(
        today ? 'dashboardHome.trend.today' : 'dashboardHome.trend.current'
      ),
      previous: t(
        today ? 'dashboardHome.trend.yesterday' : 'dashboardHome.trend.previous'
      ),
    };
  });

  const currentTotal = computed(() => sum(currentData.value));
  const previousTotal = computed(() => sum(previousData.value));
  const totalChange = computed(() =>
    changeRate(currentTotal.value, previousTotal.value)
  );
  const changeClass = computed(() => {
    if (!totalChange.value) return '';
    return totalChange.value > 0 ? 'is-up' : 'is-down';
  });

  const peak = computed(() => {
    let index = -1;
    currentData.value.forEach((value, idx) => {
      if (
        value != null &&
        (index < 0 || value > (currentData.value[index] || 0))
      ) {
        index = idx;
      }
    });
    if (index < 0) return { value: null, label: '' };
    return {
      value: currentData.value[index],
      label: bucketLabel(currentBuckets.value[index], granularity.value),
    };
  });

  const { chartOption } = useChartOption((isDark) => {
    const axisColor = isDark ? 'rgba(255,255,255,0.6)' : '#86909C';
    const splitColor = isDark ? '#2E2E30' : '#E5E8EF';
    const xLabels = currentBuckets.value.map((bucket) =>
      granularity.value === 'hour'
        ? bucketLabel(bucket, 'hour')
        : bucket.slice(5)
    );
    return {
      grid: { left: 8, right: 16, top: 40, bottom: 8, containLabel: true },
      legend: {
        top: 0,
        right: 0,
        icon: 'circle',
        itemWidth: 8,
        itemHeight: 8,
        textStyle: { color: axisColor },
      },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: xLabels,
        axisLine: { lineStyle: { color: splitColor } },
        axisTick: { show: false },
        axisLabel: { color: axisColor },
      },
      yAxis: {
        type: 'value',
        minInterval: 1,
        axisLabel: { color: axisColor },
        splitLine: { lineStyle: { type: 'dashed', color: splitColor } },
      },
      tooltip: {
        trigger: 'axis',
        className: 'echarts-tooltip-diy',
        formatter(params) {
          const list = params as ToolTipFormatterParams[];
          const idx = list[0]?.dataIndex ?? 0;
          const g = granularity.value;
          const rows = [
            {
              name: labels.value.current,
              bucket: currentBuckets.value[idx],
              value: currentData.value[idx],
              color: '#165DFF',
            },
            {
              name: labels.value.previous,
              bucket: previousBuckets.value[idx],
              value: previousData.value[idx],
              color: '#C9CDD4',
            },
          ];
          const title =
            g === 'hour' ? bucketLabel(currentBuckets.value[idx], g) : '';
          return `<div>
            ${title ? `<p class="tooltip-title">${title}</p>` : ''}
            ${rows
              .map(
                (row) => `<div class="content-panel">
                  <p><span style="background-color:${
                    row.color
                  }" class="tooltip-item-icon"></span>
                  <span>${row.name}${
                  g === 'day' && row.bucket ? ` (${row.bucket})` : ''
                }</span></p>
                  <span class="tooltip-value">${formatCount(row.value)}</span>
                </div>`
              )
              .join('')}
          </div>`;
        },
      },
      series: [
        {
          name: labels.value.current,
          type: 'line',
          smooth: true,
          showSymbol: currentBuckets.value.length <= 31,
          symbolSize: 6,
          data: currentData.value,
          lineStyle: { width: 3, color: '#165DFF' },
          itemStyle: { color: '#165DFF' },
          areaStyle: {
            color: new graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: 'rgba(22, 93, 255, 0.2)' },
              { offset: 1, color: 'rgba(22, 93, 255, 0)' },
            ]),
          },
        },
        {
          name: labels.value.previous,
          type: 'line',
          smooth: true,
          showSymbol: false,
          data: previousData.value,
          lineStyle: { width: 2, type: 'dashed', color: '#C9CDD4' },
          itemStyle: { color: '#C9CDD4' },
        },
      ],
    };
  });

  let requestSeq = 0;

  async function fetchData() {
    requestSeq += 1;
    const seq = requestSeq;
    const current = range.value;
    const previous = previousRange(current);
    const g = granularityOf(current);
    const tz = props.timezone;
    setLoading(true);
    try {
      const params = { granularity: g, type: metric.value, timezone: tz };
      const [currentList, previousList] = await Promise.all([
        getMemberGrowthTrend({
          ...params,
          start_date: current[0],
          end_date: current[1],
        }),
        getMemberGrowthTrend({
          ...params,
          start_date: previous[0],
          end_date: previous[1],
        }),
      ]);
      if (seq !== requestSeq) return;
      const currentMap = toCountMap(currentList, g);
      const previousMap = toCountMap(previousList, g);
      const now = nowIn(tz);
      granularity.value = g;
      currentBuckets.value = buildBuckets(current, g);
      previousBuckets.value = buildBuckets(previous, g);
      currentData.value = currentBuckets.value.map((bucket) =>
        isFutureBucket(bucket, g, now) ? null : currentMap.get(bucket) || 0
      );
      previousData.value = previousBuckets.value.map(
        (bucket) => previousMap.get(bucket) || 0
      );
    } catch (err) {
      if (seq === requestSeq) Message.error(t('dashboardHome.loadFailed'));
    } finally {
      if (seq === requestSeq) setLoading(false);
    }
  }

  // 切换时区后“今天”可能变化：快捷区间按新时区重算，自定义区间保持不变；
  // 两种情况都会赋一个新数组，从而触发重新查询
  watch(
    () => props.timezone,
    (tz, oldTz) => {
      const quick = detectQuickRange(
        range.value,
        ['today', 'd7', 'd30'],
        nowIn(oldTz)
      );
      range.value = quick
        ? buildQuickRange(quick, nowIn(tz))
        : ([...range.value] as DateRange);
    }
  );

  watch([range, metric], fetchData, { immediate: true });
</script>

<style scoped lang="less">
  .toolbar {
    display: flex;
    flex-wrap: wrap;
    gap: 12px 16px;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
  }

  .summary {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    margin-bottom: 16px;
    padding: 12px 0;
    background: var(--color-fill-1);
    border-radius: 6px;
  }

  .summary-item {
    padding: 0 20px;

    & + & {
      border-left: 1px solid var(--color-border-2);
    }
  }

  .summary-label {
    color: var(--color-text-3);
    font-size: 12px;
  }

  .summary-value {
    margin-top: 4px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 20px;

    &.is-up {
      color: rgb(var(--red-6));
    }

    &.is-down {
      color: rgb(var(--green-6));
    }
  }

  .summary-sub {
    margin-left: 6px;
    color: var(--color-text-3);
    font-weight: 400;
    font-size: 12px;
  }

  @media (max-width: 768px) {
    .summary {
      grid-template-columns: repeat(2, minmax(0, 1fr));
      row-gap: 12px;
    }
  }
</style>
