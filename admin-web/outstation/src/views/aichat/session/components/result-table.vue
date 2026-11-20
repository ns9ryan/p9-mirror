<template>
  <div class="result">
    <div class="result__head">
      <span class="result__title">
        <icon-bar-chart v-if="view === 'chart'" />
        <icon-storage v-else />
        查询结果
        <span class="result__count">共 {{ table.rows.length }} 条</span>
      </span>
      <a-radio-group v-if="chartMeta" v-model="view" type="button" size="mini">
        <a-radio value="chart">图表</a-radio>
        <a-radio value="table">表格</a-radio>
      </a-radio-group>
    </div>
    <Chart
      v-if="chartMeta && view === 'chart'"
      :options="chartOption"
      height="240px"
    />
    <a-table
      v-else
      :columns="columns"
      :data="data"
      :pagination="data.length > PAGE_SIZE ? pagination : false"
      :bordered="{ cell: true }"
      :scroll="{ x: '100%' }"
      size="small"
    />
  </div>
</template>

<script setup lang="ts">
  import { computed, ref } from 'vue';
  import type { TableData } from '../utils';
  import { chartableColumns, toNumber } from '../utils';

  const props = defineProps<{ table: TableData }>();

  const PAGE_SIZE = 10;
  const pagination = { pageSize: PAGE_SIZE, size: 'mini' as const };

  const chartMeta = computed(() => chartableColumns(props.table));
  const view = ref<'chart' | 'table'>(chartMeta.value ? 'chart' : 'table');

  // 列名可能含 "."，直接作 dataIndex 会被当成嵌套路径
  const columns = computed(() =>
    props.table.columns.map((title, idx) => ({
      title,
      dataIndex: `c${idx}`,
      ellipsis: true,
      tooltip: true,
      minWidth: 100,
    }))
  );

  const data = computed(() =>
    props.table.rows.map((row, rowIdx) => {
      const record: Record<string, string | number> = { key: rowIdx };
      props.table.columns.forEach((c, idx) => {
        record[`c${idx}`] = row[c];
      });
      return record;
    })
  );

  const chartOption = computed(() => {
    const meta = chartMeta.value;
    if (!meta) return {};
    const multi = meta.series.length > 1;
    return {
      grid: {
        left: 8,
        right: 16,
        top: multi ? 36 : 16,
        bottom: 8,
        containLabel: true,
      },
      tooltip: { trigger: 'axis' },
      legend: multi ? { top: 0 } : undefined,
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: props.table.rows.map((r) => String(r[meta.xColumn])),
      },
      yAxis: {
        type: 'value',
        splitLine: { lineStyle: { type: 'dashed' } },
      },
      series: meta.series.map((col) => ({
        name: col,
        type: 'line',
        smooth: true,
        symbolSize: 6,
        areaStyle: multi ? undefined : { opacity: 0.12 },
        data: props.table.rows.map((r) => toNumber(r[col])),
      })),
    };
  });
</script>

<style scoped lang="less">
  .result {
    margin: 8px 0;
    padding: 12px;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 8px;

    &__head {
      display: flex;
      gap: 12px;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 10px;
    }

    &__title {
      display: inline-flex;
      gap: 6px;
      align-items: center;
      color: var(--color-text-1);
      font-weight: 500;

      svg {
        color: rgb(var(--primary-6));
      }
    }

    &__count {
      color: var(--color-text-3);
      font-weight: 400;
      font-size: 12px;
    }
  }
</style>
