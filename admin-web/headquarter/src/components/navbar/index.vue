<template>
  <div class="navbar">
    <div class="left-side">
      <a-space>
        <img
          alt="logo"
          src="//p3-armor.byteimg.com/tos-cn-i-49unhts6dw/dfdba5317c0c20ce20e64fac803d52bc.svg~tplv-49unhts6dw-image.image"
        />
        <a-typography-title
          :style="{ margin: 0, fontSize: '18px' }"
          :heading="5"
        >
          总网平台
        </a-typography-title>
        <icon-menu-fold
          v-if="!topMenu && appStore.device === 'mobile'"
          style="font-size: 22px; cursor: pointer"
          @click="toggleDrawerMenu"
        />
      </a-space>
    </div>
    <div class="center-side">
      <Menu v-if="topMenu" />
    </div>
    <ul class="right-side aaa">
      <!-- 暂时没有这个功能，先隐藏 -->
      <!-- <li>
        <a-tooltip :content="$t('settings.search')">
          <a-button class="nav-btn" type="outline" :shape="'circle'">
            <template #icon>
              <icon-search />
            </template>
          </a-button>
        </a-tooltip>
      </li> -->
      <li>
        <a-tooltip :content="$t('settings.preview')">
          <a-button
            class="nav-btn"
            type="outline"
            :shape="'circle'"
            @click="handlePreview"
          >
            <template #icon>
              <icon-eye />
            </template>
          </a-button>
        </a-tooltip>
        <a-dropdown trigger="click" @select="goOutstation">
          <div ref="triggerBtnPreview" class="trigger-btn"></div>
          <template #content>
            <a-doption
              v-for="item in previewLocales"
              :key="item.value"
              :value="item.value"
            >
              <a-tooltip :content="item.code" position="left">
                <span>{{ item.label }}</span>
              </a-tooltip>
            </a-doption>
          </template>
        </a-dropdown>
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
          <a-avatar
            :size="32"
            :style="{ marginRight: '8px', cursor: 'pointer' }"
          >
            <img v-if="avatar" alt="avatar" :src="avatar" />
            <span v-else-if="avatarFallback">{{ avatarFallback }}</span>
            <icon-user v-else />
          </a-avatar>
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
  import { computed, ref, inject, onMounted } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { useDark, useToggle, useFullscreen } from '@vueuse/core';
  import { useAppStore, useUserStore } from '@/store';
  import { queryOperatorList } from '@/api/operator';
  import { previewToken, readPreviewLogin } from '@/api/user';
  import { openOutstationWithSession } from '@/utils/preview-handoff';
  import { LOCALE_OPTIONS } from '@/language';
  import { syncLocaleWithEnabledLangs } from '@/utils/locale';
  import useLocale from '@/hooks/locale';
  import useUser from '@/hooks/user';
  import Menu from '@/components/menu/index.vue';
  import MessageBox from '../message-box/index.vue';
  import ChangePasswordModal from './change-password-modal.vue';

  const appStore = useAppStore();
  const userStore = useUserStore();
  const { logout } = useUser();
  const { changeLocale, currentLocale } = useLocale();
  const { isFullscreen, toggle: toggleFullScreen } = useFullscreen();
  const locales = ref([...LOCALE_OPTIONS]);
  const previewLocales = ref<{ label: string; value: number; code: string }[]>(
    []
  );
  const changePasswordVisible = ref(false);

  const fetchEnabledLocales = async () => {
    const { locales: list, changed } = await syncLocaleWithEnabledLangs();
    locales.value = list;
    // 语言已回退但页面仍可能带着旧 X-Lang 的菜单数据，刷新一次对齐
    if (changed) {
      window.location.reload();
    }
  };

  const fetchPreviewStations = async () => {
    try {
      const { data } = await queryOperatorList({ page: 1, page_size: 100 });
      previewLocales.value = (data?.list || []).map((item) => ({
        label: item.name,
        value: item.id,
        code: item.code,
      }));
    } catch {
      previewLocales.value = [];
    }
  };

  onMounted(() => {
    fetchEnabledLocales();
    fetchPreviewStations();
  });
  const avatar = computed(() => userStore.avatar);
  const avatarFallback = computed(() => {
    const name = userStore.name?.trim();
    return name ? name.charAt(0).toUpperCase() : '';
  });
  const theme = computed(() => {
    return appStore.theme;
  });
  const topMenu = computed(() => appStore.topMenu && appStore.menu);
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
  const triggerBtnPreview = ref();
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
  const handlePreview = () => {
    const event = new MouseEvent('click', {
      view: window,
      bubbles: true,
      cancelable: true,
    });
    triggerBtnPreview.value.dispatchEvent(event);
  };
  const previewOpening = ref(false);
  const goOutstation = async (
    id: string | number | Record<string, any> | undefined
  ) => {
    const station = previewLocales.value.find(
      (item) => item.value === Number(id)
    );
    if (!station || previewOpening.value) return;
    previewOpening.value = true;
    try {
      const res = await previewToken({ operator_code: station.code });
      const session = readPreviewLogin(res);
      if (!session.token.access_token) {
        Message.error('换票未返回登录 token');
        return;
      }
      const opened = openOutstationWithSession(
        session,
        import.meta.env.VITE_OPERATOR_PREVIEW_URL
      );
      if (!opened) Message.error('浏览器拦截了新窗口');
    } catch {
      // 接口错误由拦截器提示
    } finally {
      previewOpening.value = false;
    }
  };
  const switchRoles = async () => {
    const res = await userStore.switchRoles();
    Message.success(res as string);
  };
  const toggleDrawerMenu = inject('toggleDrawerMenu') as () => void;
</script>

<style scoped lang="less">
  .navbar {
    display: flex;
    justify-content: space-between;
    height: 100%;
    background-color: var(--color-bg-2);
    border-bottom: 1px solid var(--color-border);
  }

  .left-side {
    display: flex;
    align-items: center;
    padding-left: 20px;
  }

  .center-side {
    flex: 1;
  }

  .right-side {
    display: flex;
    padding-right: 20px;
    list-style: none;
    :deep(.locale-select) {
      border-radius: 20px;
    }
    li {
      display: flex;
      align-items: center;
      padding: 0 10px;
    }

    a {
      color: var(--color-text-1);
      text-decoration: none;
    }
    .nav-btn {
      border-color: rgb(var(--gray-2));
      color: rgb(var(--gray-8));
      font-size: 16px;
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
</style>

<style lang="less">
  .message-popover {
    .arco-popover-content {
      margin-top: 0;
    }
  }
</style>
