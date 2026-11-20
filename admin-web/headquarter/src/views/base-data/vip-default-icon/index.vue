<template>
  <div class="container">
    <Breadcrumb />
    <a-card class="general-card" :bordered="false">
      <vip-icon-panel />
    </a-card>
  </div>
</template>

<script lang="ts" setup>
  import VipIconPanel from '../default-avatar/components/vip-icon-panel.vue';
</script>

<style scoped lang="less">
  .container {
    padding: 0 20px 20px;
  }

  .general-card {
    margin-top: 16px;
  }

  .general-card :deep(.arco-card-body) {
    padding: 24px;
  }
</style>
