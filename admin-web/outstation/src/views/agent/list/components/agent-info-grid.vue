<template>
  <section class="info-section">
    <header v-if="title" class="info-section-head">
      <span class="info-section-bar" />
      <span class="info-section-title">{{ title }}</span>
    </header>
    <div class="info-grid">
      <div
        v-for="item in items"
        :key="item.key"
        class="info-cell"
        :class="{ full: item.full }"
      >
        <div class="info-label">
          <slot :name="`label-${item.key}`">{{ item.label }}</slot>
        </div>
        <div class="info-value">
          <div class="info-text">
            <slot :name="item.key">
              <span :class="{ muted: !item.value }">
                {{ item.value || '--' }}
              </span>
            </slot>
          </div>
          <a-button
            v-if="item.action"
            class="info-action"
            type="text"
            size="mini"
            @click="emit('action', item.key)"
          >
            <template #icon><icon-edit /></template>
            {{ item.action }}
          </a-button>
        </div>
      </div>
    </div>
  </section>
</template>

<script lang="ts" setup>
  import type { InfoItem } from '../types';

  defineProps<{
    title?: string;
    items: InfoItem[];
  }>();

  const emit = defineEmits<{
    (e: 'action', key: string): void;
  }>();
</script>

<style scoped lang="less">
  .info-section + .info-section {
    margin-top: 20px;
  }

  .info-section-head {
    display: flex;
    gap: 8px;
    align-items: center;
    margin-bottom: 10px;
  }

  .info-section-bar {
    width: 3px;
    height: 14px;
    background: rgb(var(--primary-6));
    border-radius: 2px;
  }

  .info-section-title {
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 14px;
  }

  .info-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    overflow: hidden;
    border: 1px solid var(--color-neutral-3);
    border-radius: 8px;
  }

  .info-cell {
    display: grid;
    grid-template-columns: 132px minmax(0, 1fr);
    min-height: 48px;
    border-bottom: 1px solid var(--color-neutral-3);

    &:nth-child(odd):not(.full) {
      border-right: 1px solid var(--color-neutral-3);
    }

    &.full {
      grid-column: 1 / -1;
    }
  }

  /* 最后一行不要底边，奇数个时最后一个单元格独占一行 */
  .info-cell:last-child,
  .info-cell:nth-last-child(2):nth-child(odd):not(.full) {
    border-bottom: none;
  }

  .info-cell:last-child:nth-child(odd):not(.full) {
    grid-column: 1 / -1;
    border-right: none;
  }

  .info-label {
    display: flex;
    align-items: center;
    padding: 0 16px;
    color: var(--color-text-3);
    font-size: 13px;
    background: var(--color-fill-1);
  }

  .info-value {
    display: flex;
    gap: 8px;
    align-items: center;
    min-width: 0;
    padding: 8px 16px;
    color: var(--color-text-1);
    font-size: 13px;
  }

  .info-text {
    display: flex;
    flex: 1;
    gap: 8px;
    align-items: center;
    min-width: 0;
    word-break: break-all;
  }

  .muted {
    color: var(--color-text-4);
  }

  .info-action {
    flex-shrink: 0;
    padding: 0 6px;
    color: rgb(var(--primary-6));
    opacity: 0.85;
    transition: opacity 0.2s;

    &:hover {
      opacity: 1;
    }
  }
</style>
