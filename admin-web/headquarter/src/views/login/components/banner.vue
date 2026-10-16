<template>
  <aside class="banner" :aria-label="$t('front.login.brand.title')">
    <img class="banner-background" :src="bannerBackground" alt="" />
  </aside>
</template>

<script lang="ts" setup>
  import bannerBackground from '@/assets/images/robot_bg.webp';
</script>

<style lang="less" scoped>
  .banner { position: relative; flex: 0 0 55.34%; min-height: 100vh; overflow: hidden; }
  .banner-background {
    position: absolute;
    top: 0;
    left: 0;
    display: block;
    width: 100%;
    height: 100%;
    max-width: none;
    object-fit: cover;
    object-position: left center;
    user-select: none;
    pointer-events: none;
  }
  @media (max-width: 860px) { .banner { display: none; } }
</style>
