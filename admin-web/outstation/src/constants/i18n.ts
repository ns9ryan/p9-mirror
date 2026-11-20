/** 站点编码，对应 GET /admin/i18n/dict 的 i18n_code */
const SITE_I18N_CODE =
  import.meta.env.VITE_I18N_CODE || 'operator';
export default SITE_I18N_CODE;