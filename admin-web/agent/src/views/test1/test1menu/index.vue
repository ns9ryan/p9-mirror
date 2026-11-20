<template>
  <div>
    <h1>test1menu</h1>
    <a-button v-if="hasPermission('test1:button')">测试1按钮</a-button>
  </div>
</template>

<script lang="ts" setup>
  import usePermission from '@/hooks/permission';

  const { hasPermission } = usePermission();
</script>
