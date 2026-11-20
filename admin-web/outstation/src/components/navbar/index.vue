<template>
  <div class="navbar">
    <div class="left-side">
      <a-space :size="16">
        <Brand v-if="!siderVisible" />
        <span v-if="!topMenu" class="menu-fold-btn" @click="handleMenuFold">
          <icon-menu-unfold v-if="siderVisible && appStore.menuCollapse" />
          <icon-menu-fold v-else />
        </span>
        <Breadcrumb v-if="!topMenu" class="navbar-breadcrumb" />
      </a-space>
    </div>
    <div class="center-side">
      <Menu v-if="topMenu" />
    </div>
    <ul class="right-side aaa">
      <li>
        <a-tooltip :content="$t('settings.search')">
          <a-button class="nav-btn" type="outline" :shape="'circle'">
            <template #icon>
              <icon-search />
            </template>
          </a-button>
        </a-tooltip>
      </li>
      <li>
        <a-tooltip :content="$t('settings.language')">
          <a-button
            class="nav-btn"
            type="outline"
            :shape="'circle'"
            @click="setDropDownVisible"
          >
            <template #icon>
              <icon-language />
            </template>
          </a-button>
        </a-tooltip>
        <a-dropdown trigger="click" @select="changeLocale as any">
          <div ref="triggerBtn" class="trigger-btn"></div>
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
      </li>
      <li>
        <a-tooltip
          :content="
            theme === 'light'
              ? $t('settings.navbar.theme.toDark')
              : $t('settings.navbar.theme.toLight')
          "
        >
          <a-button
            class="nav-btn"
            type="outline"
            :shape="'circle'"
            @click="handleToggleTheme"
          >
            <template #icon>
              <icon-moon-fill v-if="theme === 'dark'" />
              <icon-sun-fill v-else />
            </template>
          </a-button>
        </a-tooltip>
      </li>
      <li>
        <a-tooltip :content="$t('settings.navbar.alerts')">
          <div class="message-box-trigger">
            <a-badge :count="9" dot>
              <a-button
                class="nav-btn"
                type="outline"
                :shape="'circle'"
                @click="setPopoverVisible"
              >
                <icon-notification />
              </a-button>
            </a-badge>
          </div>
        </a-tooltip>
        <a-popover
          trigger="click"
          :arrow-style="{ display: 'none' }"
          :content-style="{ padding: 0, minWidth: '400px' }"
          content-class="message-popover"
        >
          <div ref="refBtn" class="ref-btn"></div>
          <template #content>
            <message-box />
          </template>
        </a-popover>
      </li>
      <li>
        <a-tooltip
          :content="
            isFullscreen
              ? $t('settings.navbar.screen.toExit')
              : $t('settings.navbar.screen.toFull')
          "
        >
          <a-button
            class="nav-btn"
            type="outline"
            :shape="'circle'"
            @click="toggleFullScreen"
          >
            <template #icon>
              <icon-fullscreen-exit v-if="isFullscreen" />
              <icon-fullscreen v-else />
            </template>
          </a-button>
        </a-tooltip>
      </li>
      <li>
        <a-tooltip :content="$t('settings.title')">
          <a-button
            class="nav-btn"
            type="outline"
            :shape="'circle'"
            @click="setVisible"
          >
            <template #icon>
              <icon-settings />
            </template>
          </a-button>
        </a-tooltip>
      </li>
      <li>
        <a-dropdown trigger="click">
          <div class="user-trigger">
            <a-avatar :size="32" class="user-avatar">
              <img v-if="avatar" alt="avatar" :src="avatar" />
              <span v-else-if="avatarFallback">{{ avatarFallback }}</span>
              <icon-user v-else />
            </a-avatar>
            <div class="user-meta">
              <span v-if="userName" class="user-name">{{ userName }}</span>
              <span class="user-timezone">{{ timezoneClock }}</span>
            </div>
            <icon-down class="user-arrow" />
          </div>
          <template #content>
            <a-doption>
              <a-space @click="switchRoles">
                <icon-tag />
                <span>
                  {{ $t('messageBox.switchRoles') }}
                </span>
              </a-space>
            </a-doption>
            <a-doption>
              <a-space @click="$router.push({ name: 'Info' })">
                <icon-user />
                <span>
                  {{ $t('messageBox.userCenter') }}
                </span>
              </a-space>
            </a-doption>
            <a-doption>
              <a-space @click="$router.push({ name: 'Setting' })">
                <icon-settings />
                <span>
                  {{ $t('messageBox.userSettings') }}
                </span>
              </a-space>
            </a-doption>
            <a-doption>
              <a-space @click="openChangePassword">
                <icon-lock />
                <span>
                  {{ $t('messageBox.changePassword') }}
                </span>
              </a-space>
            </a-doption>
            <a-doption>
              <a-space @click="handleLogout">
                <icon-export />
                <span>
                  {{ $t('messageBox.logout') }}
                </span>
              </a-space>
            </a-doption>
          </template>
        </a-dropdown>
      </li>
    </ul>
    <ChangePasswordModal v-model:visible="changePasswordVisible" />
  </div>
</template>

<script lang="ts" setup>
  import { computed, ref, inject, onMounted, onUnmounted } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { useDark, useToggle, useFullscreen } from '@vueuse/core';
  import { useAppStore, useUserStore } from '@/store';
  import { LOCALE_OPTIONS } from '@/language';
  import { syncLocaleWithEnabledLangs } from '@/utils/locale';
  import { ensureSiteTimezone, siteNow } from '@/utils/site-time';
  import useLocale from '@/hooks/locale';
  import useUser from '@/hooks/user';
  import Menu from '@/components/menu/index.vue';
  import Brand from '@/components/brand/index.vue';
  import MessageBox from '../message-box/index.vue';
  import ChangePasswordModal from './change-password-modal.vue';

  /** 展示本站配置时区（如 Asia/Tokyo），不是浏览器本地时区 */
  const formatTimezoneClock = () => {
    const now = siteNow();
    return `(UTC${now.format('Z')}) ${now.format('YYYY-MM-DD HH:mm:ss')}`;
  };

  const timezoneClock = ref(formatTimezoneClock());
  let timezoneTimer: ReturnType<typeof setInterval> | undefined;

  const appStore = useAppStore();
  const userStore = useUserStore();
  const { logout } = useUser();
  const { changeLocale, currentLocale } = useLocale();
  const { isFullscreen, toggle: toggleFullScreen } = useFullscreen();
  const locales = ref([...LOCALE_OPTIONS]);
  const changePasswordVisible = ref(false);

  const fetchEnabledLocales = async () => {
    const { locales: list, changed } = await syncLocaleWithEnabledLangs();
    locales.value = list;
    // 语言已回退但页面仍可能带着旧 X-Lang 的菜单数据，刷新一次对齐
    if (changed) {
      window.location.reload();
    }
  };

  onMounted(async () => {
    fetchEnabledLocales();
    await ensureSiteTimezone();
    timezoneClock.value = formatTimezoneClock();
    timezoneTimer = setInterval(() => {
      timezoneClock.value = formatTimezoneClock();
    }, 1000);
  });

  onUnmounted(() => {
    if (timezoneTimer) clearInterval(timezoneTimer);
  });
  const userName = computed(() => userStore.name?.trim() || '');
  const avatar = computed(() => userStore.avatar);
  const avatarFallback = computed(() => {
    const name = userStore.name?.trim();
    return name ? name.charAt(0).toUpperCase() : '';
  });
  const theme = computed(() => {
    return appStore.theme;
  });
  const topMenu = computed(() => appStore.topMenu && appStore.menu);
  const siderVisible = computed(
    () => appStore.menu && !appStore.topMenu && !appStore.hideMenu
  );
  const isDark = useDark({
    selector: 'body',
    attribute: 'arco-theme',
    valueDark: 'dark',
    valueLight: 'light',
    storageKey: 'arco-theme',
    onChanged(dark: boolean) {
      // overridden default behavior
      appStore.toggleTheme(dark);
    },
  });
  const toggleTheme = useToggle(isDark);
  const handleToggleTheme = () => {
    toggleTheme();
  };
  const setVisible = () => {
    appStore.updateSettings({ globalSettings: true });
  };
  const refBtn = ref();
  const triggerBtn = ref();
  const setPopoverVisible = () => {
    const event = new MouseEvent('click', {
      view: window,
      bubbles: true,
      cancelable: true,
    });
    refBtn.value.dispatchEvent(event);
  };
  const handleLogout = () => {
    logout();
  };
  const openChangePassword = () => {
    changePasswordVisible.value = true;
  };
  const setDropDownVisible = () => {
    const event = new MouseEvent('click', {
      view: window,
      bubbles: true,
      cancelable: true,
    });
    triggerBtn.value.dispatchEvent(event);
  };
  const switchRoles = async () => {
    const res = await userStore.switchRoles();
    Message.success(res as string);
  };
  const toggleDrawerMenu = inject('toggleDrawerMenu') as () => void;
  const handleMenuFold = () => {
    if (siderVisible.value) {
      appStore.updateSettings({ menuCollapse: !appStore.menuCollapse });
    } else {
      toggleDrawerMenu();
    }
  };
</script>

<style scoped lang="less">
  .navbar {
    display: flex;
    justify-content: space-between;
    height: 100%;
    background: var(--app-navbar-bg);
    border-bottom: 1px solid var(--app-divider);
    backdrop-filter: blur(16px) saturate(140%);
  }

  .left-side {
    display: flex;
    align-items: center;
    padding-left: 16px;
    min-width: 0;
  }

  .menu-fold-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    color: var(--color-text-2);
    font-size: 18px;
    border-radius: 10px;
    cursor: pointer;
    transition: background 0.2s, color 0.2s;

    &:hover {
      color: rgb(var(--primary-6));
      background: var(--color-fill-2);
    }
  }

  .left-side .navbar-breadcrumb {
    margin: 0;
  }

  .center-side {
    flex: 1;
  }

  .right-side {
    display: flex;
    margin: 0;
    padding-right: 16px;
    list-style: none;
    :deep(.locale-select) {
      border-radius: 20px;
    }
    li {
      display: flex;
      align-items: center;
      padding: 0 3px;
    }

    a {
      color: var(--color-text-1);
      text-decoration: none;
    }
    .nav-btn.arco-btn {
      width: 36px;
      height: 36px;
      color: var(--color-text-2);
      font-size: 17px;
      background: transparent;
      border-color: transparent;
      border-radius: 10px;

      &:hover {
        color: rgb(var(--primary-6));
        background: var(--color-fill-2);
        border-color: transparent;
      }
    }
    .trigger-btn,
    .ref-btn {
      position: absolute;
      bottom: 14px;
    }
    .trigger-btn {
      margin-left: 14px;
    }
  }

  .user-trigger {
    display: flex;
    gap: 8px;
    align-items: center;
    margin-left: 8px;
    padding: 4px 10px 4px 4px;
    border-radius: 20px;
    cursor: pointer;
    transition: background 0.2s;

    &:hover {
      background: var(--color-fill-2);
    }
  }

  .user-avatar {
    background: var(--app-brand-gradient);
    box-shadow: 0 0 0 2px var(--color-bg-2);
  }

  .user-meta {
    display: flex;
    flex-direction: column;
    gap: 1px;
    min-width: 0;
    line-height: 1.2;
  }

  .user-name {
    max-width: 160px;
    overflow: hidden;
    color: var(--color-text-1);
    font-size: 13px;
    font-weight: 500;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .user-timezone {
    color: var(--color-text-3);
    font-size: 12px;
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
  }

  .user-arrow {
    color: var(--color-text-3);
    font-size: 12px;
  }
</style>

<style lang="less">
  .message-popover {
    .arco-popover-content {
      margin-top: 0;
    }
  }
</style>
