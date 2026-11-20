<template>
  <a-space :size="12" wrap>
    <a-radio-group
      class="quick-range"
      type="button"
      size="small"
      :model-value="quick"
      @change="onQuickChange"
    >
      <a-radio v-for="key in quickKeys" :key="key" :value="key">
        {{ t(QUICK_LABELS[key]) }}
      </a-radio>
    </a-radio-group>
    <a-range-picker
      class="range-picker"
      size="small"
      value-format="YYYY-MM-DD"
      :allow-clear="false"
      :model-value="modelValue"
      :disabled-date="isFutureDate"
      :placeholder="[
        t('dashboardHome.range.start'),
        t('dashboardHome.range.end'),
      ]"
      @change="onRangeChange"
    />
  </a-space>
</template>

<script lang="ts" setup>
  import { computed, PropType } from 'vue';
  import dayjs from 'dayjs';
  import { useI18n } from 'vue-i18n';
  import {
    buildQuickRange,
    DATE_FMT,
    DateRange,
    detectQuickRange,
    nowIn,
    QuickRange,
  } from '../utils';

  const QUICK_LABELS: Record<QuickRange, string> = {
    today: 'dashboardHome.range.today',
    d7: 'dashboardHome.range.last7',
    d30: 'dashboardHome.range.last30',
  };

  const props = defineProps({
    modelValue: {
      type: Array as unknown as PropType<DateRange>,
      required: true,
    },
    quickKeys: {
      type: Array as PropType<QuickRange[]>,
      default: () => ['today', 'd7', 'd30'],
    },
    /** IANA 时区，决定“今天”以及快捷区间的基准 */
    timezone: {
      type: String,
      default: '',
    },
  });
  const emit = defineEmits<{
    (e: 'update:modelValue', value: DateRange): void;
  }>();

  const { t } = useI18n();

  const quick = computed(() =>
    detectQuickRange(props.modelValue, props.quickKeys, nowIn(props.timezone))
  );

  function isFutureDate(current?: Date) {
    // 按日期文本比较，避免本机时区与所选时区的“今天”不一致
    return (
      Boolean(current) &&
      dayjs(current).format(DATE_FMT) > nowIn(props.timezone).format(DATE_FMT)
    );
  }

  function onQuickChange(value: string | number | boolean) {
    emit(
      'update:modelValue',
      buildQuickRange(value as QuickRange, nowIn(props.timezone))
    );
  }

  function onRangeChange(value?: (string | number | Date | undefined)[]) {
    const [start, end] = (value || []).map((item) =>
      item ? dayjs(item).format(DATE_FMT) : ''
    );
    if (start && end) emit('update:modelValue', [start, end]);
  }
</script>

<style scoped lang="less">
  .range-picker {
    width: 240px;
  }
</style>
