/// <reference types="vite/client" />

declare module '*.vue' {
  import { DefineComponent } from 'vue';
  // eslint-disable-next-line @typescript-eslint/no-explicit-any, @typescript-eslint/ban-types
  const component: DefineComponent<{}, {}, any>;
  export default component;
}
interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string;
  /** 站点 i18n 编码，用于词条下发 /i18n/dict */
  readonly VITE_I18N_CODE: string;
  /** 顶栏预览跳转的分站前台地址 */
  readonly VITE_OPERATOR_PREVIEW_URL: string;
}
