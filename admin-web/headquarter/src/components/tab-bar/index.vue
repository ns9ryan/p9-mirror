<template>
  <div class="tab-bar-container">
    <a-affix ref="affixRef" :offset-top="offsetTop">
      <div class="tab-bar-box">
        <div ref="scrollRef" class="tab-bar-scroll" @wheel.prevent="onWheel">
          <tab-item
            v-for="(tag, index) in tagList"
            :key="tag.fullPath"
            :index="index"
            :item-data="tag"
          />
        </div>
      </div>
    </a-affix>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, watch, nextTick, onUnmounted } from 'vue';
  import type { RouteLocationNormalized, RouteRecordRaw } from 'vue-router';
  import {
    listenerRouteChange,
    removeRouteListener,
  } from '@/utils/route-listener';
  import { useAppStore, useTabBarStore } from '@/store';
  import tabItem from './tab-item.vue';

  const appStore = useAppStore();
  const tabBarStore = useTabBarStore();

  const affixRef = ref();
  const scrollRef = ref<HTMLElement>();
  const tagList = computed(() => {
    return tabBarStore.getTabList;
  });
  const offsetTop = computed(() => {
    return appStore.navbar ? 60 : 0;
  });

  const onWheel = (event: WheelEvent) => {
    if (!scrollRef.value) return;
    scrollRef.value.scrollLeft += event.deltaY || event.deltaX;
  };

  const scrollActiveIntoView = () => {
    nextTick(() => {
      const active = scrollRef.value?.querySelector('.tab-item--active');
      active?.scrollIntoView({ block: 'nearest', inline: 'nearest' });
    });
  };

  watch(
    () => appStore.navbar,
    () => {
      affixRef.value.updatePosition();
    }
  );
  watch(
    () => appStore.serverMenu,
    (menus) => {
      if (appStore.menuFromServer && menus?.length) {
        tabBarStore.syncHomeTag(menus as unknown as RouteRecordRaw[]);
      }
    },
    { immediate: true }
  );
  listenerRouteChange((route: RouteLocationNormalized) => {
    if (
      !route.meta.noAffix &&
      !tagList.value.some((tag) => tag.fullPath === route.fullPath)
    ) {
      tabBarStore.updateTabList(route);
    }
    scrollActiveIntoView();
  }, true);

  onUnmounted(() => {
    removeRouteListener();
  });
</script>

<style scoped lang="less">
  .tab-bar-container {
    position: relative;
    background-color: var(--color-bg-2);

    .tab-bar-box {
      display: flex;
      align-items: center;
      height: 40px;
      padding: 0 12px;
      background-color: var(--color-bg-2);
      border-bottom: 1px solid var(--color-border);
      box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
    }

    .tab-bar-scroll {
      display: flex;
      flex: 1;
      align-items: center;
      gap: 4px;
      height: 100%;
      overflow-x: auto;
      overflow-y: hidden;
      white-space: nowrap;
      scrollbar-width: none;

      &::-webkit-scrollbar {
        display: none;
      }
    }
  }
</style>
