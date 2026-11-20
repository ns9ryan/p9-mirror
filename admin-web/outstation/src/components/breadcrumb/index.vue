<template>
  <a-breadcrumb class="container-breadcrumb">
    <a-breadcrumb-item v-for="item in breadcrumbItems" :key="item">
      {{ $t(item) }}
    </a-breadcrumb-item>
  </a-breadcrumb>
</template>

<script lang="ts" setup>
  import { PropType, computed } from 'vue';
  import { useRoute } from 'vue-router';

  const props = defineProps({
    items: {
      type: Array as PropType<string[]>,
      default: undefined,
    },
  });

  const route = useRoute();

  const breadcrumbItems = computed(() => {
    if (props.items) return props.items;
    return route.matched
      .map((record) => record.meta?.locale as string | undefined)
      .filter((locale): locale is string => Boolean(locale));
  });
</script>

<style scoped lang="less">
  .container-breadcrumb {
    margin: 16px 0;
    :deep(.arco-breadcrumb-item) {
      color: var(--color-text-3);
      &:last-child {
        color: var(--color-text-1);
        font-weight: 600;
        font-size: 15px;
      }
    }

    :deep(.arco-breadcrumb-item-separator) {
      color: var(--color-text-4);
    }
  }
</style>
