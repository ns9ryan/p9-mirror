<template>
  <a-dropdown
    trigger="contextMenu"
    :popup-max-height="false"
    @select="actionSelect"
  >
    <div
      class="tab-item"
      :class="{ 'tab-item--active': itemData.fullPath === $route.fullPath }"
      @click="goto(itemData)"
    >
      <span class="tab-item-title">
        {{ $t(itemData.title) }}
      </span>
      <span
        v-if="index !== 0"
        class="tab-item-close"
        @click.stop="tagClose(itemData, index)"
      >
        <icon-close />
      </span>
    </div>
    <template #content>
      <a-doption :disabled="disabledReload" :value="Eaction.reload">
        <icon-refresh />
        <span>重新加载</span>
      </a-doption>
      <a-doption
        class="sperate-line"
        :disabled="disabledCurrent"
        :value="Eaction.current"
      >
        <icon-close />
        <span>关闭当前标签页</span>
      </a-doption>
      <a-doption :disabled="disabledLeft" :value="Eaction.left">
        <icon-to-left />
        <span>关闭左侧标签页</span>
      </a-doption>
      <a-doption
        class="sperate-line"
        :disabled="disabledRight"
        :value="Eaction.right"
      >
        <icon-to-right />
        <span>关闭右侧标签页</span>
      </a-doption>
      <a-doption :value="Eaction.others">
        <icon-swap />
        <span>关闭其它标签页</span>
      </a-doption>
      <a-doption :value="Eaction.all">
        <icon-folder-delete />
        <span>关闭全部标签页</span>
      </a-doption>
    </template>
  </a-dropdown>
</template>

<script lang="ts" setup>
  import { PropType, computed } from 'vue';
  import { useRouter, useRoute } from 'vue-router';
  import { useTabBarStore } from '@/store';
  import type { TagProps } from '@/store/modules/tab-bar/types';
  import { REDIRECT_ROUTE_NAME } from '@/router/constants';

  // eslint-disable-next-line no-shadow
  enum Eaction {
    reload = 'reload',
    current = 'current',
    left = 'left',
    right = 'right',
    others = 'others',
    all = 'all',
  }

  const props = defineProps({
    itemData: {
      type: Object as PropType<TagProps>,
      default() {
        return {};
      },
    },
    index: {
      type: Number,
      default: 0,
    },
  });

  const router = useRouter();
  const route = useRoute();
  const tabBarStore = useTabBarStore();

  const goto = (tag: TagProps) => {
    router.push({ ...tag });
  };
  const tagList = computed(() => {
    return tabBarStore.getTabList;
  });

  const disabledReload = computed(() => {
    return props.itemData.fullPath !== route.fullPath;
  });

  const disabledCurrent = computed(() => {
    return props.index === 0;
  });

  const disabledLeft = computed(() => {
    return [0, 1].includes(props.index);
  });

  const disabledRight = computed(() => {
    return props.index === tagList.value.length - 1;
  });

  const tagClose = (tag: TagProps, idx: number) => {
    tabBarStore.deleteTag(idx, tag);
    if (props.itemData.fullPath === route.fullPath) {
      const latest = tagList.value[idx - 1]; // 获取队列的前一个tab
      router.push({ name: latest.name });
    }
  };

  const findCurrentRouteIndex = () => {
    return tagList.value.findIndex((el) => el.fullPath === route.fullPath);
  };
  const actionSelect = async (value: any) => {
    const { itemData, index } = props;
    const copyTagList = [...tagList.value];
    if (value === Eaction.current) {
      tagClose(itemData, index);
    } else if (value === Eaction.left) {
      const currentRouteIdx = findCurrentRouteIndex();
      copyTagList.splice(1, props.index - 1);

      tabBarStore.freshTabList(copyTagList);
      if (currentRouteIdx < index) {
        router.push({ name: itemData.name });
      }
    } else if (value === Eaction.right) {
      const currentRouteIdx = findCurrentRouteIndex();
      copyTagList.splice(props.index + 1);

      tabBarStore.freshTabList(copyTagList);
      if (currentRouteIdx > index) {
        router.push({ name: itemData.name });
      }
    } else if (value === Eaction.others) {
      const filterList = tagList.value.filter((el, idx) => {
        return idx === 0 || idx === props.index;
      });
      tabBarStore.freshTabList(filterList);
      router.push({ name: itemData.name });
    } else if (value === Eaction.reload) {
      tabBarStore.deleteCache(itemData);
      await router.push({
        name: REDIRECT_ROUTE_NAME,
        params: {
          path: route.fullPath,
        },
      });
      tabBarStore.addCache(itemData.name);
    } else {
      tabBarStore.resetTabList();
      router.push({ name: tabBarStore.homeTag.name });
    }
  };
</script>

<style scoped lang="less">
  .tab-item {
    position: relative;
    display: inline-flex;
    flex-shrink: 0;
    align-items: center;
    height: 30px;
    padding: 0 14px;
    color: var(--color-text-2);
    font-size: 13px;
    border-radius: 4px;
    cursor: pointer;
    user-select: none;
    transition: color 0.2s, background-color 0.2s;

    &:hover {
      color: var(--color-text-1);
      background-color: var(--color-fill-2);
    }

    &--active,
    &--active:hover {
      color: rgb(var(--primary-6));
      font-weight: 500;
      background-color: rgb(var(--primary-1));
    }

    &--active::after {
      position: absolute;
      right: 14px;
      bottom: -4px;
      left: 14px;
      height: 2px;
      background-color: rgb(var(--primary-6));
      border-radius: 2px;
      content: '';
    }
  }

  .tab-item-close {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    margin-right: -6px;
    margin-left: 6px;
    color: var(--color-text-3);
    font-size: 12px;
    border-radius: 50%;
    transition: color 0.2s, background-color 0.2s;

    &:hover {
      color: var(--color-white);
      background-color: rgb(var(--gray-5));
    }

    .tab-item--active & {
      color: rgb(var(--primary-6));

      &:hover {
        color: var(--color-white);
        background-color: rgb(var(--primary-6));
      }
    }
  }

  :deep(.arco-dropdown-option-content) {
    span {
      margin-left: 10px;
    }
  }

  .tab-item.arco-dropdown-open {
    color: rgb(var(--danger-6));
  }
  .sperate-line {
    border-bottom: 1px solid var(--color-neutral-3);
  }
</style>
