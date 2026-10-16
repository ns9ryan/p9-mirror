<template>
  <a-layout-footer class="footer">
    分站管理平台
    <span class="version">v{{ appVersion }}</span>
  </a-layout-footer>
</template>

<script lang="ts" setup>
  import appVersion from '@/constants/version';
</script>

<style lang="less" scoped>
  .footer {
    display: flex;
    gap: 8px;
    align-items: center;
    justify-content: center;
    height: 30px;
    color: var(--color-text-2);
    text-align: center;
  }

  .version {
    color: var(--color-text-3);
    font-size: 12px;
  }
</style>
