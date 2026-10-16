<template>
  <div>
    <h1>testmenu2（路径不对，请用 /test1/test2menu）</h1>
    <a-button v-if="hasPermission('test2:button')">测试2按钮</a-button>
  </div>
</template>

<script lang="ts" setup>
  import usePermission from '@/hooks/permission';

  const { hasPermission } = usePermission();
</script>
