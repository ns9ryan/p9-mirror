<template>
  <main class="login-page">
    <div
      class="login-bg"
      aria-hidden="true"
      :style="{ backgroundImage: `url(${loginBg})` }"
    />
    <div class="login-overlay" aria-hidden="true" />
    <div class="login-particles" aria-hidden="true" />
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
        <span class="login-brand-logo"><icon-branch /></span>
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
  import loginBg from '@/assets/images/login-bg.jpg';
  import { LOCALE_OPTIONS } from '@/language';
  import {
    changeLoginLocale,
    loadRemoteDict,
    syncLocaleWithEnabledLangs,
    type LocaleOption,
  } from '@/utils/locale';
  import { resolveOperatorCodeFromDomain } from '@/utils/operator-site';
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
    // 未登录时 enabled 接口要求带 operator_code，先按域名解析
    await resolveOperatorCodeFromDomain();
    const { locales: list } = await syncLocaleWithEnabledLangs();
    locales.value = list;
    await loadRemoteDict();
  });
</script>

<style lang="less" scoped>
  @bg: #060d22;

  .login-page {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    min-width: 1180px;
    min-height: 100vh;
    overflow: hidden;
    background: @bg;
  }

  .login-bg {
    position: absolute;
    inset: 0;
    background-position: left center;
    background-size: cover;
    background-repeat: no-repeat;
    animation: login-bg-zoom 30s ease-in-out infinite alternate;
  }

  .login-overlay {
    position: absolute;
    inset: 0;
    background: radial-gradient(
        620px 520px at 78% 50%,
        rgba(40, 110, 255, 0.22),
        transparent 70%
      ),
      linear-gradient(
        90deg,
        rgba(6, 13, 34, 0) 0%,
        rgba(6, 13, 34, 0) 45%,
        rgba(6, 13, 34, 0.5) 60%,
        rgba(6, 13, 34, 0.78) 78%
      ),
      linear-gradient(0deg, rgba(6, 13, 34, 0.7) 0%, transparent 22%);
    pointer-events: none;
  }

  // 粒子层高出一个图块，平移恰好一个图块高度，循环无缝
  .login-particles {
    position: absolute;
    top: 0;
    right: 0;
    bottom: -320px;
    left: 0;
    background-image: radial-gradient(1.5px 1.5px at 30px 40px, rgba(140, 220, 255, 0.9), transparent),
      radial-gradient(1px 1px at 140px 90px, rgba(255, 255, 255, 0.8), transparent),
      radial-gradient(2px 2px at 250px 30px, rgba(56, 200, 255, 0.9), transparent),
      radial-gradient(1px 1px at 280px 200px, rgba(255, 255, 255, 0.7), transparent),
      radial-gradient(1.5px 1.5px at 70px 250px, rgba(120, 160, 255, 0.9), transparent),
      radial-gradient(2px 2px at 190px 290px, rgba(56, 200, 255, 0.7), transparent);
    background-size: 320px 320px;
    animation: login-particles-rise 36s linear infinite;
    pointer-events: none;
    will-change: transform;
  }

  @keyframes login-particles-rise {
    from {
      transform: translate3d(0, 0, 0);
    }
    to {
      transform: translate3d(0, -320px, 0);
    }
  }

  @keyframes login-bg-zoom {
    from {
      transform: scale(1);
    }
    to {
      transform: scale(1.05);
    }
  }

  .login-locale {
    position: absolute;
    top: 20px;
    right: 24px;
    z-index: 3;
  }

  .login-locale-btn {
    color: rgba(190, 212, 240, 0.85);
    font-size: 14px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(120, 170, 255, 0.18);
    border-radius: 18px;
    backdrop-filter: blur(8px);

    &:hover {
      color: #6fd6ff;
      background: rgba(64, 160, 255, 0.12);
      border-color: rgba(56, 200, 255, 0.45);
    }
  }

  .login-panel {
    position: relative;
    z-index: 2;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    width: 42%;
    min-height: 100vh;
    padding: 48px 6% 48px 2%;
  }

  .login-brand {
    display: flex;
    flex-direction: column;
    align-items: center;
    width: min(100%, 420px);
    margin-bottom: 28px;
    text-align: center;
  }

  .login-brand-logo {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 56px;
    height: 56px;
    margin-bottom: 16px;
    color: #fff;
    font-size: 30px;
    background: linear-gradient(135deg, #b08cff 0%, #6a4bff 50%, #22c8ff 100%);
    border-radius: 16px;
    box-shadow: 0 0 0 6px rgba(130, 100, 255, 0.1), 0 0 30px rgba(130, 100, 255, 0.55);
  }

  .login-brand-title {
    margin: 0;
    color: transparent;
    font-weight: 700;
    font-size: 34px;
    line-height: 1.3;
    letter-spacing: 4px;
    background: linear-gradient(90deg, #ffffff 30%, #7fd8ff 100%);
    -webkit-background-clip: text;
    background-clip: text;
    filter: drop-shadow(0 0 16px rgba(56, 160, 255, 0.35));
  }

  .login-brand-subtitle {
    margin: 10px 0 0;
    color: rgba(174, 198, 230, 0.55);
    font-size: 12px;
    letter-spacing: 6px;
    text-transform: uppercase;
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

  @media (max-width: 1180px) {
    .login-page {
      min-width: 0;
    }

    .login-panel {
      width: 50%;
      padding-right: 5%;
    }
  }

  @media (max-width: 860px) {
    .login-page {
      justify-content: center;
      min-width: 0;
    }

    .login-bg {
      background-position: 25% center;
    }

    .login-overlay {
      background: linear-gradient(
        180deg,
        rgba(6, 13, 34, 0.5) 0%,
        rgba(6, 13, 34, 0.9) 45%,
        @bg 70%
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

  @media (prefers-reduced-motion: reduce) {
    .login-bg,
    .login-particles {
      animation: none;
    }
  }
</style>
