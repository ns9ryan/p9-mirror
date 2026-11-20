<template>
  <div class="login-form-wrapper">
    <h1 class="login-form-title">{{ t('front.login.form.title') }}</h1>
    <div class="login-form-error-msg">{{ errorMessage }}</div>
    <a-form
      ref="loginForm"
      :model="userInfo"
      class="login-form"
      layout="vertical"
      @submit="handleSubmit"
    >
      <a-form-item
        field="username"
        :label="t('front.login.form.userName')"
        :rules="[
          { required: true, message: t('front.login.form.userName.errMsg') },
        ]"
        :validate-trigger="['change', 'blur']"
      >
        <a-input
          v-model="userInfo.username"
          :placeholder="t('front.login.form.userName.placeholder')"
          allow-clear
          ><template #prefix><icon-user /></template
        ></a-input>
      </a-form-item>
      <a-form-item
        field="password"
        :label="t('front.login.form.password')"
        :rules="[
          { required: true, message: t('front.login.form.password.errMsg') },
        ]"
        :validate-trigger="['change', 'blur']"
      >
        <a-input-password
          v-model="userInfo.password"
          :placeholder="t('front.login.form.password.placeholder')"
          allow-clear
          ><template #prefix><icon-lock /></template
        ></a-input-password>
      </a-form-item>
      <a-form-item
        v-if="!operatorCodeFromDomain"
        field="operator_code"
        label="分站代码"
        :rules="[{ required: true, message: '请输入分站代码' }]"
        :validate-trigger="['change', 'blur']"
      >
        <a-input
          v-model="userInfo.operator_code"
          placeholder="请输入分站代码"
          allow-clear
          ><template #prefix><icon-branch /></template
        ></a-input>
      </a-form-item>
      <div class="login-form-password-actions">
        <a-checkbox
          :model-value="loginConfig.rememberPassword"
          @change="setRememberPassword as any"
          >{{ t('front.login.form.rememberPassword') }}</a-checkbox
        >
        <a-link>{{ t('front.login.form.forgetPassword') }}</a-link>
      </div>
      <a-button
        class="login-button"
        type="primary"
        html-type="submit"
        long
        :loading="loading || resolvingOperator"
        >{{ t('front.login.form.login') }}</a-button
      >
    </a-form>
    <!-- <p class="browser-tip"
      ><icon-public /> 推荐使用 Chrome / Edge 浏览器，最佳分辨率 1920x1080
      及以上</p
    > -->
  </div>
</template>

<script lang="ts" setup>
  import { onMounted, ref, reactive } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import type { NavigationFailure, RouteRecordRaw } from 'vue-router';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { ValidatedError } from '@arco-design/web-vue/es/form/interface';
  import { useStorage } from '@vueuse/core';
  import { useAppStore, useUserStore } from '@/store';
  import useLoading from '@/hooks/loading';
  import type { LoginData } from '@/api/user';
  import {
    getRedirectQuery,
    resolvePostLoginTarget,
  } from '@/router/login-redirect';
  import { syncLocaleWithEnabledLangs } from '@/utils/locale';
  import {
    resolveOperatorCodeFromDomain,
    setCachedOperatorCode,
  } from '@/utils/operator-site';

  const { t } = useI18n();
  const route = useRoute();
  const router = useRouter();
  const operatorCodeFromQuery =
    typeof route.query.operator_code === 'string'
      ? route.query.operator_code
      : '';
  const errorMessage = ref('');
  const { loading, setLoading } = useLoading();
  const appStore = useAppStore();
  const userStore = useUserStore();
  const loginConfig = useStorage('login-config', {
    rememberPassword: true,
    username: '',
    password: '',
    operator_code: '',
  });
  const isLegacyMockCredential =
    loginConfig.value.username === 'admin' &&
    loginConfig.value.password === 'admin';
  const userInfo = reactive({
    username: loginConfig.value.username,
    password: isLegacyMockCredential ? 'Admin@123' : loginConfig.value.password,
    operator_code: operatorCodeFromQuery || loginConfig.value.operator_code,
  });

  /** 按地址栏域名查到的站点编码；查不到（未登记、IP、localhost）时回退为手动输入 */
  const operatorCodeFromDomain = ref('');
  const resolvingOperator = ref(false);
  onMounted(async () => {
    resolvingOperator.value = true;
    try {
      const code = await resolveOperatorCodeFromDomain();
      if (code) {
        operatorCodeFromDomain.value = code;
        userInfo.operator_code = code;
      }
    } finally {
      resolvingOperator.value = false;
    }
  });

  const persistLoginConfig = (
    rememberPassword: boolean,
    username: string,
    password: string,
    operatorCode: string
  ) => {
    loginConfig.value.username = rememberPassword ? username : '';
    loginConfig.value.password = rememberPassword ? password : '';
    loginConfig.value.operator_code = rememberPassword ? operatorCode : '';
  };

  const handleSubmit = async ({
    errors,
    values,
  }: {
    errors: Record<string, ValidatedError> | undefined;
    values: Record<string, any>;
  }) => {
    if (loading.value || errors) return;
    setLoading(true);
    try {
      const operatorCode = values.operator_code;
      const loginData: LoginData = {
        username: values.username,
        password: values.password,
      };
      if (operatorCode) {
        loginData.operator_code = operatorCode;
        setCachedOperatorCode(operatorCode);
      }
      await userStore.login(loginData);
      // 登录成功后立刻写入，保证退出后能回填本次账号密码
      persistLoginConfig(
        loginConfig.value.rememberPassword,
        values.username,
        values.password,
        operatorCode || ''
      );
      // 本地缓存语言若已停用，回退到默认语言后再拉菜单（X-Lang）
      await syncLocaleWithEnabledLangs();
      const redirect = getRedirectQuery(router.currentRoute.value.query);
      const { redirect: _ignored, ...othersQuery } =
        router.currentRoute.value.query;
      let menus: RouteRecordRaw[] = [];
      if (appStore.menuFromServer) {
        menus = await appStore.fetchServerMenuConfig();
        menus.forEach((menu) => {
          if (typeof menu.name === 'string' && router.hasRoute(menu.name)) {
            router.removeRoute(menu.name);
          }
          router.addRoute('serverMenuRoot', menu);
        });
      }
      const target = resolvePostLoginTarget(router, redirect, menus);
      let failure: NavigationFailure | void | undefined;
      if (typeof target === 'string') {
        failure = await router.push(target);
      } else {
        const extraQuery =
          target && typeof target === 'object' && 'query' in target
            ? target.query
            : undefined;
        failure = await router.push({
          ...target,
          query: {
            ...othersQuery,
            ...extraQuery,
          },
        });
      }
      // 代理初始化提示里选择退出登录时导航会被取消
      if (failure) return;
      Message.success(t('front.login.form.login.success'));
    } catch (err) {
      errorMessage.value = (err as Error).message;
    } finally {
      setLoading(false);
    }
  };
  const setRememberPassword = (value: boolean) => {
    loginConfig.value.rememberPassword = value;
    if (!value) {
      persistLoginConfig(false, '', '', '');
    } else {
      persistLoginConfig(
        true,
        userInfo.username,
        userInfo.password,
        userInfo.operator_code || ''
      );
    }
  };
</script>

<style lang="less">
  @neon: 56, 200, 255;

  .hud-corner(@c, @len: 18px, @w: 2px) {
    background: linear-gradient(@c, @c) top left / @len @w no-repeat,
      linear-gradient(@c, @c) top left / @w @len no-repeat,
      linear-gradient(@c, @c) top right / @len @w no-repeat,
      linear-gradient(@c, @c) top right / @w @len no-repeat,
      linear-gradient(@c, @c) bottom left / @len @w no-repeat,
      linear-gradient(@c, @c) bottom left / @w @len no-repeat,
      linear-gradient(@c, @c) bottom right / @len @w no-repeat,
      linear-gradient(@c, @c) bottom right / @w @len no-repeat;
  }

  // 背景宽 30% 时，0% ~ 100% 即从左端走到右端
  @keyframes login-edge-flow {
    from {
      background-position: 0% 0;
    }
    to {
      background-position: 100% 0;
    }
  }

  @keyframes login-btn-sheen {
    from {
      transform: translateX(-120%) skewX(-20deg);
    }
    to {
      transform: translateX(320%) skewX(-20deg);
    }
  }

  .login-form-wrapper {
    position: relative;
    box-sizing: border-box;
    width: min(100%, 420px);
    padding: 34px 36px 30px;
    color: #b8c8de;
    background: linear-gradient(
          160deg,
          rgba(16, 32, 72, 0.7),
          rgba(7, 16, 40, 0.82)
        )
        padding-box,
      linear-gradient(
          135deg,
          rgba(@neon, 0.75),
          rgba(60, 126, 255, 0.18) 32%,
          rgba(60, 126, 255, 0.1) 68%,
          rgba(140, 100, 255, 0.65)
        )
        border-box;
    border: 1px solid transparent;
    border-radius: 16px;
    backdrop-filter: blur(18px);
    box-shadow: inset 0 0 60px rgba(40, 110, 255, 0.1),
      0 0 50px rgba(30, 110, 255, 0.18), 0 30px 60px rgba(0, 0, 0, 0.45);

    &::before {
      position: absolute;
      inset: -1px;
      .hud-corner(rgba(@neon, 0.95));
      border-radius: inherit;
      filter: drop-shadow(0 0 4px rgba(@neon, 0.9));
      pointer-events: none;
      content: '';
    }

    &::after {
      position: absolute;
      top: -1px;
      right: 20px;
      left: 20px;
      height: 2px;
      background: linear-gradient(
          90deg,
          transparent,
          rgba(@neon, 0.25) 30%,
          #fff 50%,
          rgba(@neon, 0.25) 70%,
          transparent
        )
        no-repeat;
      background-size: 30% 100%;
      filter: drop-shadow(0 0 6px rgb(@neon));
      animation: login-edge-flow 4s ease-in-out infinite alternate;
      pointer-events: none;
      content: '';
    }
  }

  .login-form-title {
    position: relative;
    margin: 0 0 22px;
    padding-bottom: 14px;
    color: #eaf4ff;
    font-weight: 600;
    font-size: 20px;
    line-height: 1.4;
    text-align: center;
    letter-spacing: 6px;

    &::after {
      position: absolute;
      bottom: 0;
      left: 50%;
      width: 48px;
      height: 3px;
      background: linear-gradient(90deg, #2f6bff, #22c8ff);
      border-radius: 3px;
      box-shadow: 0 0 10px rgba(@neon, 0.7);
      transform: translateX(-50%);
      content: '';
    }
  }

  .login-form-error-msg {
    height: 18px;
    color: #ff7b8a;
    font-size: 12px;
  }

  .login-form-wrapper .arco-form-item {
    margin-bottom: 20px;
  }

  .login-form-wrapper .arco-form-item-label-col {
    margin-bottom: 7px;
  }

  .login-form-wrapper .arco-form-item-label {
    color: #a9bdd6;
    font-size: 14px;
  }

  .login-form-wrapper .arco-input-wrapper {
    height: 48px;
    padding: 0 16px;
    color: #e2edfb;
    background: rgba(4, 12, 30, 0.6) !important;
    border: 1px solid rgba(80, 140, 255, 0.26) !important;
    border-radius: 10px;
    box-shadow: none;
    transition: border-color 0.2s, box-shadow 0.2s, background 0.2s;
  }

  .login-form-wrapper .arco-input-wrapper:hover {
    background: rgba(6, 16, 40, 0.7) !important;
    border-color: rgba(@neon, 0.5) !important;
  }

  .login-form-wrapper .arco-input-wrapper.arco-input-focus {
    background: rgba(6, 16, 40, 0.8) !important;
    border-color: rgb(@neon) !important;
    box-shadow: 0 0 0 3px rgba(@neon, 0.14), 0 0 16px rgba(@neon, 0.25);
  }

  .login-form-wrapper .arco-input-prefix {
    margin-right: 12px;
    color: #6fd6ff;
    font-size: 18px;
  }

  .login-form-wrapper .arco-input {
    color: #e6efff;
    font-size: 15px;
  }

  .login-form-wrapper .arco-input::placeholder {
    color: #6f7d96;
  }

  .login-form-wrapper .arco-input-suffix,
  .login-form-wrapper .arco-icon-hover {
    color: #8ea3c2;
  }

  .login-form-password-actions {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin: 2px 0 24px;
  }

  .login-form-wrapper .arco-checkbox-label {
    color: #aeb9cc;
  }

  .login-form-wrapper .arco-link {
    color: #6fd6ff;
  }

  .login-form-wrapper .arco-checkbox-checked .arco-checkbox-icon {
    background-color: #1f7aff;
    border-color: #1f7aff;
  }

  // 需要压过 theme.less 中 body 级主按钮渐变规则
  body .login-form-wrapper .arco-btn.arco-btn-primary.arco-btn-status-normal.login-button {
    position: relative;
    height: 48px;
    overflow: hidden;
    font-weight: 600;
    font-size: 16px;
    letter-spacing: 8px;
    background: linear-gradient(90deg, #2f6bff 0%, #1fb6ff 100%);
    border: 0;
    border-radius: 10px;
    box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.14),
      0 8px 24px rgba(26, 120, 255, 0.45);
    transition: box-shadow 0.2s, transform 0.15s, filter 0.2s;

    &::after {
      position: absolute;
      top: 0;
      bottom: 0;
      left: 0;
      width: 30%;
      background: linear-gradient(
        90deg,
        transparent,
        rgba(255, 255, 255, 0.35),
        transparent
      );
      transform: translateX(-120%) skewX(-20deg);
      pointer-events: none;
      content: '';
    }

    &:hover {
      filter: brightness(1.08);
      box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.2),
        0 10px 32px rgba(26, 140, 255, 0.6);
      transform: translateY(-1px);

      &::after {
        animation: login-btn-sheen 0.9s ease;
      }
    }
  }

  .browser-tip {
    display: flex;
    gap: 8px;
    align-items: center;
    margin: 61px 0 0;
    color: #8894a8;
    font-size: 14px;
    white-space: nowrap;
  }

  @media (max-width: 860px) {
    .login-form-wrapper {
      padding: 28px 24px;
    }

    .browser-tip {
      margin-top: 35px;
      font-size: 12px;
      white-space: normal;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .login-form-wrapper::after {
      animation: none;
    }
  }
</style>
