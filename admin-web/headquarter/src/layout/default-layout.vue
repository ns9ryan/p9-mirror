<template>
  <a-layout class="layout" :class="{ mobile: appStore.hideMenu }">
    <div
      v-if="navbar"
      class="layout-navbar"
      :class="{ 'layout-navbar--resizing': isResizing }"
      :style="{ left: `${siderOffset}px` }"
    >
      <NavBar />
    </div>
    <a-layout>
      <a-layout>
        <a-layout-sider
          v-if="renderMenu"
          v-show="!hideMenu"
          class="layout-sider"
          :class="{ 'layout-sider--resizing': isResizing }"
          breakpoint="xl"
          :collapsed="collapsed"
          :collapsible="true"
          :width="menuWidth"
          :hide-trigger="true"
          @collapse="setCollapsed"
        >
          <div class="sider-brand" :class="{ 'sider-brand--collapsed': collapsed }">
            <Brand :collapsed="collapsed" />
          </div>
          <div class="menu-wrapper">
            <Menu />
          </div>
          <div
            v-if="!collapsed"
            class="sider-resize-handle"
            title="拖拽调整菜单宽度"
            @mousedown.prevent="onResizeStart"
          />
        </a-layout-sider>
        <a-drawer
          v-if="hideMenu"
          :visible="drawerVisible"
          placement="left"
          :footer="false"
          mask-closable
          :closable="false"
          @cancel="drawerCancel"
        >
          <Menu />
        </a-drawer>
        <a-layout
          class="layout-content"
          :class="{ 'layout-content--resizing': isResizing }"
          :style="paddingStyle"
        >
          <TabBar v-if="appStore.tabBar" />
          <a-layout-content class="layout-page">
            <PageLayout />
          </a-layout-content>
          <Footer v-if="footer" />
        </a-layout>
      </a-layout>
    </a-layout>
  </a-layout>
</template>

<script lang="ts" setup>
  import { ref, computed, watch, provide, onMounted, onBeforeUnmount } from 'vue';
  import { useRouter, useRoute } from 'vue-router';
  import { useAppStore, useUserStore } from '@/store';
  import NavBar from '@/components/navbar/index.vue';
  import Menu from '@/components/menu/index.vue';
  import Footer from '@/components/footer/index.vue';
  import TabBar from '@/components/tab-bar/index.vue';
  import Brand from '@/components/brand/index.vue';
  import usePermission from '@/hooks/permission';
  import useResponsive from '@/hooks/responsive';
  import PageLayout from './page-layout.vue';

  const MENU_WIDTH_MIN = 160;
  const MENU_WIDTH_MAX = 420;

  const isInit = ref(false);
  const isResizing = ref(false);
  const appStore = useAppStore();
  const userStore = useUserStore();
  const router = useRouter();
  const route = useRoute();
  const permission = usePermission();
  useResponsive(true);
  const navbarHeight = `60px`;
  const navbar = computed(() => appStore.navbar);
  const renderMenu = computed(() => appStore.menu && !appStore.topMenu);
  const hideMenu = computed(() => appStore.hideMenu);
  const footer = computed(() => appStore.footer);
  const menuWidth = computed(() => {
    return appStore.menuCollapse ? 48 : appStore.menuWidth;
  });
  const collapsed = computed(() => {
    return appStore.menuCollapse;
  });
  const siderOffset = computed(() =>
    renderMenu.value && !hideMenu.value ? menuWidth.value : 0
  );
  const paddingStyle = computed(() => {
    const paddingLeft = siderOffset.value
      ? { paddingLeft: `${siderOffset.value}px` }
      : {};
    const paddingTop = navbar.value ? { paddingTop: navbarHeight } : {};
    return { ...paddingLeft, ...paddingTop };
  });
  const setCollapsed = (val: boolean) => {
    if (!isInit.value) return; // for page initialization menu state problem
    appStore.updateSettings({ menuCollapse: val });
  };

  const clampMenuWidth = (width: number) =>
    Math.min(MENU_WIDTH_MAX, Math.max(MENU_WIDTH_MIN, Math.round(width)));

  const onResizeMove = (event: MouseEvent) => {
    if (!isResizing.value) return;
    appStore.updateSettings({ menuWidth: clampMenuWidth(event.clientX) });
  };

  const onResizeEnd = () => {
    if (!isResizing.value) return;
    isResizing.value = false;
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
    window.removeEventListener('mousemove', onResizeMove);
    window.removeEventListener('mouseup', onResizeEnd);
  };

  const onResizeStart = () => {
    if (collapsed.value) return;
    isResizing.value = true;
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
    window.addEventListener('mousemove', onResizeMove);
    window.addEventListener('mouseup', onResizeEnd);
  };

  watch(
    () => userStore.role,
    (roleValue) => {
      if (roleValue && !permission.accessRouter(route))
        router.push({ name: 'notFound' });
    }
  );
  const drawerVisible = ref(false);
  const drawerCancel = () => {
    drawerVisible.value = false;
  };
  provide('toggleDrawerMenu', () => {
    drawerVisible.value = !drawerVisible.value;
  });
  onMounted(() => {
    isInit.value = true;
  });
  onBeforeUnmount(() => {
    onResizeEnd();
  });
</script>

<style scoped lang="less">
  @nav-size-height: 60px;
  @layout-max-width: 1100px;

  // 独立层叠上下文，让 ::before 纹理层压在背景之上、内容之下
  .layout {
    position: relative;
    z-index: 0;
    width: 100%;
    height: 100%;
    background-color: var(--app-bg);
    background-image: var(--app-bg-image);
    background-attachment: fixed;

    &::before {
      position: fixed;
      inset: 0;
      z-index: -1;
      background-image: var(--app-bg-pattern);
      background-size: var(--app-bg-pattern-size);
      mask-image: radial-gradient(
        ellipse 70% 60% at 80% 0%,
        #000 0%,
        transparent 75%
      );
      pointer-events: none;
      content: '';
    }
  }

  .layout-navbar {
    position: fixed;
    top: 0;
    right: 0;
    z-index: 100;
    height: @nav-size-height;
    transition: left 0.2s cubic-bezier(0.34, 0.69, 0.1, 1);

    &--resizing {
      transition: none;
    }
  }

  .layout-sider {
    position: fixed;
    top: 0;
    left: 0;
    z-index: 99;
    height: 100%;
    background: var(--app-sider-bg);
    box-shadow: none;
    backdrop-filter: blur(16px);
    transition: all 0.2s cubic-bezier(0.34, 0.69, 0.1, 1);
    &::before {
      position: absolute;
      right: 0;
      bottom: 0;
      left: 0;
      height: 260px;
      background: var(--app-sider-glow);
      pointer-events: none;
      content: '';
    }
    &::after {
      position: absolute;
      top: 0;
      right: -1px;
      display: block;
      width: 1px;
      height: 100%;
      background-color: var(--app-divider);
      content: '';
    }

    &--resizing {
      transition: none;
    }

    > :deep(.arco-layout-sider-children) {
      position: relative;
      display: flex;
      flex-direction: column;
      overflow-y: hidden;
    }
  }

  .sider-brand {
    flex-shrink: 0;
    height: @nav-size-height;
    padding: 0 20px;

    &--collapsed {
      padding: 0;
    }
  }

  .sider-resize-handle {
    position: absolute;
    top: 0;
    right: -3px;
    z-index: 10;
    width: 6px;
    height: 100%;
    cursor: col-resize;
    touch-action: none;

    &:hover,
    &:active {
      background-color: rgba(var(--primary-6), 0.35);
    }
  }

  .menu-wrapper {
    flex: 1;
    min-height: 0;
    overflow: auto;
    overflow-x: hidden;
    :deep(.arco-menu) {
      ::-webkit-scrollbar {
        width: 12px;
        height: 4px;
      }

      ::-webkit-scrollbar-thumb {
        border: 4px solid transparent;
        background-clip: padding-box;
        border-radius: 7px;
        background-color: var(--color-text-4);
      }

      ::-webkit-scrollbar-thumb:hover {
        background-color: var(--color-text-3);
      }
    }
  }

  .layout-content {
    min-height: 100vh;
    overflow-y: hidden;
    background: transparent;
    transition: padding 0.2s cubic-bezier(0.34, 0.69, 0.1, 1);

    &--resizing {
      transition: none;
    }
  }

  .layout-page {
    padding-top: 12px;

    :deep(.container > .general-card:first-child) {
      margin-top: 0;
    }
  }
</style>
