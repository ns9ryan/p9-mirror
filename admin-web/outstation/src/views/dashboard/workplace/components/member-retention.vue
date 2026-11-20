<template>
  <a-card class="general-card" :bordered="false">
    <template #title>
      {{ t('dashboardHome.retention.title') }}
      <a-tooltip :content="t('dashboardHome.retention.tip')">
        <icon-question-circle class="title-tip" />
      </a-tooltip>
    </template>
    <div class="toolbar">
      <RangeFilter
        v-model="range"
        :quick-keys="QUICK_KEYS"
        :timezone="timezone"
      />
      <a-radio-group v-model="mode" type="button" size="small">
        <a-radio value="rate">{{ t('dashboardHome.retention.rate') }}</a-radio>
        <a-radio value="count">
          {{ t('dashboardHome.retention.count') }}
        </a-radio>
      </a-radio-group>
    </div>
    <a-spin :loading="loading" style="width: 100%">
      <Chart height="300px" :options="chartOption" />
      <a-table
        class="cohort-table"
        row-key="date"
        size="small"
        :columns="columns"
        :data="rows"
        :pagination="false"
        :bordered="{ cell: true }"
        :scroll="{ x: 900, y: 420 }"
        :summary="summary"
      >
        <template #newCount="{ record }">
          {{ formatCount(record.newCount) }}
        </template>
        <template #retention="{ record, column }">
          <div class="heat-cell" :style="heatStyle(record, column.dataIndex)">
            {{ formatValue(record[column.dataIndex]) }}
          </div>
        </template>
        <template #summary-cell="{ record, column }">
          <template v-if="column.dataIndex === 'date'">
            {{ record.date }}
          </template>
          <template v-else-if="column.dataIndex === 'newCount'">
            {{ formatCount(record.newCount) }}
          </template>
          <template v-else>
            {{ formatValue(record[column.dataIndex]) }}
          </template>
        </template>
        <template #empty>
          <a-empty :description="t('dashboardHome.empty')" />
        </template>
      </a-table>
    </a-spin>
  </a-card>
</template>

<script lang="ts" setup>
  import { computed, ref, watch } from 'vue';
  import dayjs from 'dayjs';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import {
    getMemberGrowthTrend,
    MemberTrendType,
    RETENTION_DAYS,
    RetentionDay,
  } from '@/api/report';
  import useLoading from '@/hooks/loading';
  import useChartOption from '@/hooks/chart-option';
  import RangeFilter from './range-filter.vue';
  import {
    buildBuckets,
    buildQuickRange,
    DATE_FMT,
    DateRange,
    detectQuickRange,
    formatCount,
    formatRate,
    nowIn,
    QuickRange,
    toCountMap,
  } from '../utils';

  type Mode = 'rate' | 'count';
  type DayKey = `d${RetentionDay}`;
  type CohortRow = { date: string; newCount: number } & Record<
    DayKey,
    number | null
  >;

  const LINE_COLORS = [
    '#165DFF',
    '#14C9C9',
    '#F7BA1E',
    '#722ED1',
    '#F77234',
    '#3491FA',
    '#9FDB1D',
  ];
  const DEFAULT_VISIBLE: RetentionDay[] = [1, 7, 30];

  const QUICK_KEYS: QuickRange[] = ['d7', 'd30'];

  const props = defineProps({
    /** 查询时区（IANA），为空则不传 timezone 参数 */
    timezone: {
      type: String,
      default: '',
    },
  });

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(false);

  const range = ref<DateRange>(buildQuickRange('d30', nowIn(props.timezone)));
  const mode = ref<Mode>('rate');
  const loadedMode = ref<Mode>('rate');
  const dates = ref<string[]>([]);
  const newCounts = ref<Map<string, number>>(new Map());
  const values = ref<Partial<Record<RetentionDay, Map<string, number>>>>({});

  const dayKey = (day: RetentionDay): DayKey => `d${day}`;

  function dayLabel(day: RetentionDay) {
    return day === 1
      ? t('dashboardHome.retention.day1')
      : t('dashboardHome.retention.dayN', { day });
  }

  function isMatured(date: string, day: RetentionDay) {
    // 按日期文本比较，以所选时区的“今天”为准
    return (
      dayjs(date).add(day, 'day').format(DATE_FMT) <=
      nowIn(props.timezone).format(DATE_FMT)
    );
  }

  function formatValue(value?: number | null) {
    return loadedMode.value === 'rate' ? formatRate(value) : formatCount(value);
  }

  const cohorts = computed<CohortRow[]>(() =>
    dates.value.map((date) => {
      const row = {
        date,
        newCount: newCounts.value.get(date) || 0,
      } as CohortRow;
      RETENTION_DAYS.forEach((day) => {
        row[dayKey(day)] = isMatured(date, day)
          ? values.value[day]?.get(date) ?? 0
          : null;
      });
      return row;
    })
  );

  const rows = computed(() => [...cohorts.value].reverse());

  const columns = computed<TableColumnData[]>(() => [
    {
      title: t('dashboardHome.retention.colDate'),
      dataIndex: 'date',
      width: 120,
      fixed: 'left',
      align: 'center',
    },
    {
      title: t('dashboardHome.retention.colNew'),
      dataIndex: 'newCount',
      slotName: 'newCount',
      width: 100,
      align: 'center',
    },
    ...RETENTION_DAYS.map((day) => ({
      title: dayLabel(day),
      dataIndex: dayKey(day),
      slotName: 'retention',
      align: 'center' as const,
      bodyCellClass: 'heat-td',
    })),
  ]);

  function summary() {
    const list = cohorts.value;
    const isRate = loadedMode.value === 'rate';
    const row: Record<string, string | number | null> = {
      date: t(
        isRate
          ? 'dashboardHome.retention.average'
          : 'dashboardHome.retention.sum'
      ),
      newCount: list.reduce((acc, item) => acc + item.newCount, 0),
    };
    RETENTION_DAYS.forEach((day) => {
      const matured = list
        .map((item) => item[dayKey(day)])
        .filter((value): value is number => value != null);
      const total = matured.reduce((acc, value) => acc + value, 0);
      if (!matured.length) row[dayKey(day)] = null;
      else row[dayKey(day)] = isRate ? total / matured.length : total;
    });
    return list.length ? [row] : [];
  }

  function heatStyle(record: CohortRow, key: DayKey) {
    const value = record[key];
    if (value == null) return { color: 'var(--color-text-4)' };
    const ratio =
      loadedMode.value === 'rate'
        ? value / 100
        : value / Math.max(record.newCount, 1);
    const alpha = Math.min(Math.max(ratio, 0), 1) * 0.85 + 0.06;
    return {
      background: `rgba(22, 93, 255, ${alpha.toFixed(2)})`,
      color: alpha > 0.45 ? '#fff' : 'var(--color-text-1)',
    };
  }

  const { chartOption } = useChartOption((isDark) => {
    const axisColor = isDark ? 'rgba(255,255,255,0.6)' : '#86909C';
    const splitColor = isDark ? '#2E2E30' : '#E5E8EF';
    const isRate = loadedMode.value === 'rate';
    const newName = t('dashboardHome.retention.colNew');
    const selected: Record<string, boolean> = { [newName]: true };
    RETENTION_DAYS.forEach((day) => {
      selected[dayLabel(day)] = DEFAULT_VISIBLE.includes(day);
    });
    return {
      grid: { left: 8, right: 8, top: 48, bottom: 8, containLabel: true },
      legend: {
        top: 0,
        icon: 'circle',
        itemWidth: 8,
        itemHeight: 8,
        selected,
        textStyle: { color: axisColor },
      },
      tooltip: { trigger: 'axis' },
      xAxis: {
        type: 'category',
        data: dates.value.map((date) => date.slice(5)),
        axisLine: { lineStyle: { color: splitColor } },
        axisTick: { show: false },
        axisLabel: { color: axisColor },
      },
      yAxis: [
        {
          type: 'value',
          name: t(
            isRate
              ? 'dashboardHome.retention.rate'
              : 'dashboardHome.retention.count'
          ),
          nameTextStyle: { color: axisColor },
          axisLabel: {
            color: axisColor,
            formatter: isRate ? '{value}%' : '{value}',
          },
          splitLine: { lineStyle: { type: 'dashed', color: splitColor } },
        },
        {
          type: 'value',
          name: newName,
          minInterval: 1,
          nameTextStyle: { color: axisColor },
          axisLabel: { color: axisColor },
          splitLine: { show: false },
        },
      ],
      series: [
        {
          name: newName,
          type: 'bar',
          yAxisIndex: 1,
          barMaxWidth: 16,
          itemStyle: {
            color: isDark ? 'rgba(22,93,255,0.35)' : '#E8F3FF',
            borderRadius: [2, 2, 0, 0],
          },
          data: cohorts.value.map((row) => row.newCount),
        },
        ...RETENTION_DAYS.map((day, idx) => ({
          name: dayLabel(day),
          type: 'line' as const,
          smooth: true,
          symbolSize: 6,
          connectNulls: false,
          tooltip: {
            valueFormatter: (value: unknown) => formatValue(value as number),
          },
          itemStyle: { color: LINE_COLORS[idx] },
          lineStyle: { width: 2 },
          data: cohorts.value.map((row) => row[dayKey(day)]),
        })),
      ],
    };
  });

  let requestSeq = 0;

  async function fetchData() {
    requestSeq += 1;
    const seq = requestSeq;
    const [start, end] = range.value;
    const currentMode = mode.value;
    const tz = props.timezone;
    const query = (type: MemberTrendType) =>
      getMemberGrowthTrend({
        start_date: start,
        end_date: end,
        granularity: 'day',
        type,
        timezone: tz,
      }).then((list) => toCountMap(list, 'day'));
    setLoading(true);
    try {
      const [register, ...retained] = await Promise.all([
        query('register'),
        ...RETENTION_DAYS.map((day) =>
          query(
            currentMode === 'rate' ? `retention${day}dRate` : `retention${day}d`
          )
        ),
      ]);
      if (seq !== requestSeq) return;
      const next: Partial<Record<RetentionDay, Map<string, number>>> = {};
      RETENTION_DAYS.forEach((day, idx) => {
        next[day] = retained[idx];
      });
      dates.value = buildBuckets([start, end], 'day');
      newCounts.value = register;
      values.value = next;
      loadedMode.value = currentMode;
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
      const quick = detectQuickRange(range.value, QUICK_KEYS, nowIn(oldTz));
      range.value = quick
        ? buildQuickRange(quick, nowIn(tz))
        : ([...range.value] as DateRange);
    }
  );

  watch([range, mode], fetchData, { immediate: true });
</script>

<style scoped lang="less">
  .title-tip {
    margin-left: 6px;
    color: var(--color-text-3);
    font-size: 14px;
    cursor: help;
  }

  .toolbar {
    display: flex;
    flex-wrap: wrap;
    gap: 12px 16px;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
  }

  .cohort-table {
    margin-top: 16px;

    :deep(.arco-table-th) {
      color: var(--color-text-1);
      font-weight: 600;
      background-color: var(--color-fill-2);
    }

    :deep(.heat-td .arco-table-cell) {
      padding: 0;
    }

    :deep(.arco-table-tr-summary .arco-table-td) {
      color: var(--color-text-1);
      font-weight: 600;
      background-color: var(--color-fill-1);
    }
  }

  .heat-cell {
    padding: 6px 8px;
    font-variant-numeric: tabular-nums;
    transition: background 0.2s;
  }
</style>
