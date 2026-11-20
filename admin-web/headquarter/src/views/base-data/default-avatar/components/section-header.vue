<template>
  <div class="section-header">
    <div class="section-main">
      <div class="section-title">
        <span class="title-bar" />
        <span>{{ title }}</span>
        <slot name="title-extra" />
      </div>
      <div v-if="desc" class="section-desc">{{ desc }}</div>
      <div v-if="specs?.length" class="spec-list">
        <span v-for="spec in specs" :key="spec" class="spec-item">
          {{ spec }}
        </span>
      </div>
    </div>
    <div v-if="$slots.extra" class="section-extra">
      <slot name="extra" />
    </div>
  </div>
</template>

<script lang="ts" setup>
  defineProps<{
    title: string;
    desc?: string;
    specs?: string[];
  }>();
</script>

<style scoped lang="less">
  .section-header {
    display: flex;
    flex-wrap: wrap;
    gap: 12px 24px;
    align-items: flex-start;
    justify-content: space-between;
    margin-bottom: 20px;
  }

  .section-main {
    flex: 1;
    min-width: 280px;
  }

  .section-title {
    display: flex;
    gap: 8px;
    align-items: center;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 16px;
    line-height: 24px;
  }

  .title-bar {
    width: 4px;
    height: 16px;
    background: rgb(var(--primary-6));
    border-radius: 2px;
  }

  .section-desc {
    margin-top: 6px;
    padding-left: 12px;
    color: var(--color-text-3);
    font-size: 13px;
    line-height: 20px;
  }

  .spec-list {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-top: 10px;
    padding-left: 12px;
  }

  .spec-item {
    padding: 0 8px;
    color: rgb(var(--primary-6));
    font-size: 12px;
    line-height: 22px;
    background: rgb(var(--primary-1));
    border-radius: 4px;
  }

  .section-extra {
    display: flex;
    flex-shrink: 0;
    align-items: center;
  }
</style>
