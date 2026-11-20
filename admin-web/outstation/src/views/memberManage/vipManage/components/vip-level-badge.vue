<template>
  <span class="level-badge">
    <svg viewBox="0 0 16 16" :width="size" :height="size" aria-hidden="true">
      <path d="M2 12 L1 4.5 L5 7.5 L8 2.5 L11 7.5 L15 4.5 L14 12 Z" :fill="color" />
      <rect x="2" y="13" width="12" height="1.6" rx="0.8" :fill="color" />
    </svg>
    VIP{{ level }}
  </span>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';

  const CROWN_COLORS = [
    '#A9AEB8',
    '#23C343',
    '#F7BA1E',
    '#57A9FB',
    '#3C7EFF',
    '#8D4EDA',
    '#FF7D00',
    '#F53F3F',
    '#D91AD9',
    '#1D2129',
  ];

  const props = withDefaults(
    defineProps<{
      level: number;
      size?: number;
    }>(),
    { size: 16 }
  );

  const color = computed(
    () => CROWN_COLORS[Math.min(props.level, CROWN_COLORS.length - 1)]
  );
</script>

<style scoped lang="less">
  .level-badge {
    display: inline-flex;
    gap: 6px;
    align-items: center;
    color: var(--color-text-1);
  }
</style>
