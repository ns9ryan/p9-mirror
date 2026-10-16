<template>
  <main class="login-page">
    <div
      class="login-bg"
      aria-hidden="true"
      :style="{ backgroundImage: `url(${robotBg})` }"
    />
    <div class="login-overlay" aria-hidden="true" />
    <div class="login-locale">
      <a-dropdown trigger="click" @select="onSelectLocale">
        <a-button class="login-locale-btn" type="text">
          <template #icon><icon-language /></template>
          {{ currentLocaleLabel }}
        </a-button>
        <template #content>
          <a-doption
            v-for="item in locales"
            :key="item.value"
            :value="item.value"
          >
            <template #icon>
              <icon-check v-show="item.value === currentLocale" />
            </template>
            {{ item.label }}
          </a-doption>
        </template>
      </a-dropdown>
    </div>
    <section class="login-panel">
      <div class="login-brand">
        <h2 class="login-brand-title">{{ $t('front.login.brand.title') }}</h2>
        <p class="login-brand-subtitle">{{ $t('front.login.brand.subtitle') }}</p>
      </div>
      <LoginForm />
    </section>
    <p class="login-version">v{{ appVersion }}</p>
  </main>
</template>

<script lang="ts" setup>
  import { computed, onMounted, ref } from 'vue';
  import appVersion from '@/constants/version';
  import robotBg from '@/assets/images/robot_bg.webp';
  import { LOCALE_OPTIONS } from '@/language';
  import {
    changeLoginLocale,
    loadRemoteDict,
    syncLocaleWithEnabledLangs,
    type LocaleOption,
  } from '@/utils/locale';
  import useLocale from '@/hooks/locale';
  import LoginForm from './components/login-form.vue';

  const { currentLocale } = useLocale();
  const locales = ref<LocaleOption[]>([...LOCALE_OPTIONS]);

  const currentLocaleLabel = computed(() => {
    return (
      locales.value.find((item) => item.value === currentLocale.value)?.label ||
      currentLocale.value
    );
  });

  const onSelectLocale = async (value: string | number | Record<string, any> | undefined) => {
    const locale = String(value ?? '');
    if (!locale || locale === currentLocale.value) return;
    await changeLoginLocale(locale);
  };

  onMounted(async () => {
    const { locales: list } = await syncLocaleWithEnabledLangs();
    locales.value = list;
    await loadRemoteDict();
  });
</script>

<style lang="less" scoped>
  .login-page {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    min-width: 1180px;
    min-height: 100vh;
    overflow: hidden;
    background: #040b18;
  }

  .login-bg {
    position: absolute;
    inset: 0;
    background-position: left center;
    background-size: cover;
    background-repeat: no-repeat;
  }

  .login-overlay {
    position: absolute;
    inset: 0;
    background: linear-gradient(
      100deg,
      rgba(4, 11, 24, 0.15) 0%,
      rgba(4, 11, 24, 0.35) 38%,
      rgba(4, 11, 24, 0.82) 52%,
      rgba(4, 11, 24, 0.96) 62%,
      #040b18 72%
    );
    pointer-events: none;
  }

  .login-locale {
    position: absolute;
    top: 20px;
    right: 24px;
    z-index: 3;
  }

  .login-locale-btn {
    color: rgba(174, 198, 230, 0.85);
    font-size: 14px;

    &:hover {
      color: #5ec8ff;
      background: rgba(64, 160, 255, 0.08);
    }
  }

  .login-panel {
    position: relative;
    z-index: 2;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    width: 46%;
    min-height: 100vh;
    padding: 48px 6% 48px 2%;
  }

  .login-brand {
    width: min(100%, 440px);
    margin-bottom: 28px;
    text-align: center;
  }

  .login-brand-title {
    margin: 0;
    color: #f0f6ff;
    font-size: 34px;
    font-weight: 700;
    line-height: 1.3;
    letter-spacing: 4px;
    text-shadow: 0 0 24px rgba(64, 160, 255, 0.35);
  }

  .login-brand-subtitle {
    margin: 10px 0 0;
    color: rgba(174, 198, 230, 0.55);
    font-size: 13px;
    letter-spacing: 6px;
  }

  @media (max-width: 1180px) {
    .login-page {
      min-width: 0;
    }

    .login-panel {
      width: 50%;
      padding-right: 5%;
    }
  }

  .login-version {
    position: absolute;
    right: 24px;
    bottom: 16px;
    z-index: 2;
    margin: 0;
    color: rgba(174, 198, 230, 0.45);
    font-size: 12px;
    letter-spacing: 1px;
  }

  @media (max-width: 860px) {
    .login-page {
      justify-content: center;
      min-width: 0;
    }

    .login-bg {
      background-position: 30% center;
    }

    .login-overlay {
      background: linear-gradient(
        180deg,
        rgba(4, 11, 24, 0.55) 0%,
        rgba(4, 11, 24, 0.92) 45%,
        #040b18 70%
      );
    }

    .login-panel {
      width: 100%;
      min-height: 100vh;
      padding: 36px 20px;
    }

    .login-brand-title {
      font-size: 26px;
      letter-spacing: 2px;
    }

    .login-brand-subtitle {
      font-size: 11px;
      letter-spacing: 4px;
    }
  }
</style>
