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
      <!-- <a-form-item field="operator_code" label="分站代码">
        <a-input
          v-model="userInfo.operator_code"
          placeholder="分站模式请输入分站代码"
          allow-clear
          ><template #prefix><icon-branch /></template
        ></a-input>
      </a-form-item> -->
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
        :loading="loading"
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
  import { ref, reactive } from 'vue';
  import { useRouter } from 'vue-router';
  import type { RouteRecordRaw } from 'vue-router';
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

  const { t } = useI18n();
  const router = useRouter();
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
    operator_code: loginConfig.value.operator_code,
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
      if (operatorCode) loginData.operator_code = operatorCode;
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
      if (typeof target === 'string') {
        await router.push(target);
      } else {
        const extraQuery =
          target && typeof target === 'object' && 'query' in target
            ? target.query
            : undefined;
        await router.push({
          ...target,
          query: {
            ...othersQuery,
            ...extraQuery,
          },
        });
      }
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
  .login-form-wrapper {
    position: relative;
    box-sizing: border-box;
    width: min(100%, 440px);
    padding: 32px 36px 28px;
    color: #b8c8de;
    background: linear-gradient(
      160deg,
      rgba(8, 22, 48, 0.72),
      rgba(5, 14, 32, 0.82)
    );
    border: 1px solid rgba(64, 180, 255, 0.35);
    border-radius: 12px;
    backdrop-filter: blur(16px);
    box-shadow:
      0 0 0 1px rgba(64, 180, 255, 0.08) inset,
      0 0 40px rgba(32, 120, 220, 0.12),
      0 24px 48px rgba(0, 0, 0, 0.35);

    &::before,
    &::after {
      position: absolute;
      width: 18px;
      height: 18px;
      border-color: rgba(64, 200, 255, 0.7);
      border-style: solid;
      content: '';
      pointer-events: none;
    }

    &::before {
      top: -1px;
      left: -1px;
      border-width: 2px 0 0 2px;
    }

    &::after {
      right: -1px;
      bottom: -1px;
      border-width: 0 2px 2px 0;
    }
  }

  .login-form-title {
    margin: 0 0 24px;
    color: #5ec8ff;
    font-size: 22px;
    font-weight: 600;
    line-height: 1.4;
    text-align: center;
    letter-spacing: 6px;
    text-shadow: 0 0 16px rgba(64, 200, 255, 0.4);
  }
  .login-form-error-msg {
    height: 18px;
    color: #f56c6c;
    font-size: 12px;
  }
  .login-form-wrapper .arco-form-item {
    margin-bottom: 21px;
  }
  .login-form-wrapper .arco-form-item-label-col {
    margin-bottom: 7px;
  }
  .login-form-wrapper .arco-form-item-label {
    color: #9eb4cc;
    font-size: 14px;
  }
  .login-form-wrapper .arco-input-wrapper {
    height: 50px;
    padding: 0 16px;
    color: #e2edfb;
    background: rgba(4, 12, 28, 0.55) !important;
    border: 1px solid rgba(64, 160, 255, 0.28) !important;
    border-radius: 6px;
    box-shadow: none;
    transition: border-color 0.2s, box-shadow 0.2s;
  }
  .login-form-wrapper .arco-input-wrapper:hover,
  .login-form-wrapper .arco-input-wrapper-focus {
    background: rgba(4, 12, 28, 0.7) !important;
    border-color: rgba(64, 200, 255, 0.65) !important;
    box-shadow: 0 0 12px rgba(64, 180, 255, 0.15);
  }
  .login-form-wrapper .arco-input-prefix {
    margin-right: 12px;
    color: #5ec8ff;
    font-size: 18px;
  }
  .login-form-wrapper .arco-input {
    color: #dde7f7;
    font-size: 15px;
  }
  .login-form-wrapper .arco-input::placeholder {
    color: #7f899c;
  }
  .login-form-password-actions {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin: 3px 0 25px;
  }
  .login-form-wrapper .arco-checkbox-label {
    color: #aeb9cc;
  }
  .login-form-wrapper .arco-link {
    color: #5ec8ff;
  }
  .login-form-wrapper .arco-checkbox-checked .arco-checkbox-icon {
    background-color: #1a7fff;
    border-color: #1a7fff;
  }
  .login-button {
    height: 50px;
    font-size: 17px;
    font-weight: 600;
    letter-spacing: 8px;
    background: linear-gradient(180deg, #2b8cff, #1260e0);
    border: 0;
    border-radius: 6px;
    box-shadow: 0 4px 20px rgba(26, 120, 255, 0.4);
    transition: box-shadow 0.2s, transform 0.15s;

    &:hover {
      box-shadow: 0 6px 28px rgba(26, 120, 255, 0.55);
      transform: translateY(-1px);
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
</style>
