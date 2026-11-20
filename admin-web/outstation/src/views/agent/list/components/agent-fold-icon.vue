<template>
  <span class="fold-icon" :class="{ 'is-open': open }" aria-hidden="true">
    <svg class="fold-user">
      <use href="#icon-yonghu1" />
    </svg>
    <!-- 未展开：纯用户图（表示可展开）；已展开：用户 + 上箭头（表示可收起） -->
    <icon-up v-if="open" class="fold-caret" />
  </span>
</template>

<script lang="ts" setup>
  defineProps<{
    open?: boolean;
  }>();
</script>

<style scoped lang="less">
  .fold-icon {
    position: relative;
    display: inline-block;
    width: 22px;
    height: 22px;
  }

  .fold-user {
    display: block;
    width: 20px;
    height: 20px;
    overflow: hidden;
  }

  .fold-icon.is-open .fold-user {
    width: 16px;
    height: 16px;
  }

  .fold-caret {
    position: absolute;
    right: -1px;
    bottom: -1px;
    color: #1677fc;
    font-size: 10px;
  }
</style>
