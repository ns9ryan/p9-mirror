import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { Message } from '@arco-design/web-vue';

export default function useLocale() {
  const i18 = useI18n();
  const currentLocale = computed(() => {
    return i18.locale.value;
  });
  const changeLocale = (value: string) => {
    if (i18.locale.value === value) {
      return;
    }
    i18.locale.value = value;
    localStorage.setItem('arco-locale', value);
    Message.success(i18.t('navbar.action.locale'));
    // 后端菜单标题/角色名/接口描述等内容由 X-Lang 请求头驱动翻译，
    // 刷新以携带新的 X-Lang 重新拉取菜单、权限等服务端数据
    setTimeout(() => window.location.reload(), 300);
  };
  return {
    currentLocale,
    changeLocale,
  };
}
