<template>
  <div class="container-breadcrumb page-header">
    <div class="page-header-bg" />
    <div
      class="page-header-deco"
      :class="`page-header-deco--${appStore.theme}`"
      :style="{ backgroundImage: `url(${decoImage})` }"
    />
    <span class="page-header-icon">
      <component :is="iconName" />
    </span>
    <div class="page-header-text">
      <div class="page-header-title">{{ title }}</div>
      <a-breadcrumb class="page-header-trail">
        <a-breadcrumb-item v-for="(item, index) in breadcrumbItems" :key="index">
          {{ $t(item) }}
        </a-breadcrumb-item>
      </a-breadcrumb>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRoute } from 'vue-router';
  import { useAppStore } from '@/store';
  import decoLight from '@/assets/images/theme/header-deco-light.jpg';
  import decoDark from '@/assets/images/theme/header-deco-dark.jpg';

  const { t } = useI18n();
  const route = useRoute();
  const appStore = useAppStore();
  const decoImage = computed(() =>
    appStore.theme === 'dark' ? decoDark : decoLight
  );

  // 面包屑由当前路由的 matched 链路生成：每一级取 meta.locale（与侧边栏使用的菜单标题一致）
  const breadcrumbItems = computed<string[]>(() =>
    route.matched
      .map((record) => record.meta?.locale)
      .filter(
        (locale): locale is string => typeof locale === 'string' && locale !== ''
      )
  );

  const title = computed(() => {
    const items = breadcrumbItems.value;
    const last = items[items.length - 1];
    return last ? t(last) : '';
  });

  const iconName = computed(() => {
    const matched = [...route.matched]
      .reverse()
      .find((item) => typeof item.meta?.icon === 'string' && item.meta.icon);
    return (matched?.meta?.icon as string) || 'icon-apps';
  });
</script>

<style scoped lang="less">
  .page-header {
    position: relative;
    display: flex;
    gap: 12px;
    align-items: center;
    height: 56px;
    margin: 12px 0 0;
    padding: 0 18px;
    border: 1px solid var(--app-banner-border);
    border-radius: 14px;
    box-shadow: var(--app-banner-shadow);
  }

  .page-header-bg {
    position: absolute;
    inset: 0;
    overflow: hidden;
    background: var(--app-banner-glow), var(--app-banner-bg);
    border-radius: 13px;

    &::before {
      position: absolute;
      top: 0;
      right: 0;
      bottom: 0;
      width: 50%;
      background: repeating-linear-gradient(
        115deg,
        var(--app-banner-line) 0 1px,
        transparent 1px 12px
      );
      mask-image: linear-gradient(90deg, transparent, #000 70%);
      content: '';
    }

    &::after {
      position: absolute;
      top: -120px;
      right: 200px;
      width: 180px;
      height: 180px;
      border: 1px solid var(--app-banner-ring);
      border-radius: 50%;
      outline: 1px solid var(--app-banner-ring);
      outline-offset: 22px;
      content: '';
    }
  }

  // 顶部越出 12px 正好贴到内容区顶边，形成“探出”横幅的立体感
  .page-header-deco {
    position: absolute;
    right: 32px;
    bottom: 0;
    width: 136px;
    height: 68px;
    background-repeat: no-repeat;
    background-position: center 46%;
    background-size: cover;
    pointer-events: none;
    mask-image: radial-gradient(closest-side, #000 60%, transparent 100%);

    @media (max-width: 768px) {
      display: none;
    }

    &--light {
      mix-blend-mode: multiply;
    }

    &--dark {
      mix-blend-mode: screen;
    }
  }

  .page-header-icon {
    position: relative;
    display: inline-flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    color: #fff;
    font-size: 19px;
    background: var(--app-brand-gradient);
    border-radius: 10px;
    box-shadow: var(--app-brand-shadow);
  }

  .page-header-text {
    position: relative;
    min-width: 0;
  }

  .page-header-title {
    color: var(--color-text-1);
    font-weight: 700;
    font-size: 17px;
    line-height: 24px;
    letter-spacing: 0.5px;
  }

  .page-header-trail {
    font-size: 12px;
    line-height: 16px;

    :deep(.arco-breadcrumb-item) {
      padding: 0;
      color: var(--color-text-3);
      font-size: 12px;

      &:last-child {
        color: var(--color-text-2);
      }
    }

    :deep(.arco-breadcrumb-item-separator) {
      margin: 0 6px;
      color: var(--color-text-4);
    }
  }
</style>
