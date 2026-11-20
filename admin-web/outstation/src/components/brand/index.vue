<template>
  <div class="brand" :class="{ 'brand--collapsed': collapsed }">
    <span class="brand-logo">
      <icon-branch />
    </span>
    <span v-if="!collapsed" class="brand-name">分站平台</span>
  </div>
</template>

<script lang="ts" setup>
  defineProps({
    collapsed: {
      type: Boolean,
      default: false,
    },
  });
</script>

<style scoped lang="less">
  .brand {
    display: flex;
    gap: 10px;
    align-items: center;
    min-width: 0;
    height: 100%;
    white-space: nowrap;
    user-select: none;

    &--collapsed {
      justify-content: center;
    }
  }

  .brand-logo {
    display: inline-flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    color: #fff;
    font-size: 18px;
    background: var(--app-brand-gradient);
    border-radius: 10px;
    box-shadow: var(--app-brand-shadow);
  }

  .brand-name {
    overflow: hidden;
    color: var(--color-text-1);
    font-weight: 700;
    font-size: 17px;
    letter-spacing: 0.5px;
    text-overflow: ellipsis;
  }
</style>
