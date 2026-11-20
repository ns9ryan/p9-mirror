import i18n, { LOCALE_OPTIONS } from '@/language';
import { queryEnabledLangs, queryI18nDict } from '@/api/i18n';
import  SITE_I18N_CODE  from '@/constants/i18n';

export type LocaleOption = { label: string; value: string };

const FALLBACK_LOCALE = 'zh-CN';

/** vue-i18n 默认只声明了 zh-CN / en-US，远端语言需按 string 写入 */
type I18nGlobal = {
  locale: string | { value: string };
  availableLocales: string[];
  setLocaleMessage: (locale: string, messages: Record<string, string>) => void;
  mergeLocaleMessage: (
    locale: string,
    messages: Record<string, string>
  ) => void;
};

const i18nGlobal = i18n.global as unknown as I18nGlobal;

function getStoredLocale() {
  return localStorage.getItem('arco-locale') || FALLBACK_LOCALE;
}

/** 写入本地语言，并同步 vue-i18n（不弹提示、不刷新） */
export function applyLocale(locale: string) {
  localStorage.setItem('arco-locale', locale);
  const current = i18nGlobal.locale;
  if (typeof current === 'string') {
    i18nGlobal.locale = locale;
  } else {
    current.value = locale;
  }
}

/** 从词条接口 data 中取出扁平文案表 */
function normalizeDictItems(
  data: unknown
): Record<string, string> | null {
  if (!data || typeof data !== 'object' || Array.isArray(data)) {
    return null;
  }
  const record = data as Record<string, unknown>;
  if (
    record.items &&
    typeof record.items === 'object' &&
    !Array.isArray(record.items)
  ) {
    return record.items as Record<string, string>;
  }
  return record as Record<string, string>;
}

/**
 * 拉取远端词条并合并进 vue-i18n（扁平 key，如 login.form.title）
 */
export async function loadRemoteDict(lang?: string): Promise<boolean> {
  const locale = lang || getStoredLocale();
  if (!SITE_I18N_CODE) return false;
  try {
    const { data } = await queryI18nDict(
      { i18n_code: SITE_I18N_CODE, lang: locale },
      { silentError: true }
    );
    const items = normalizeDictItems(data);
    if (!items || !Object.keys(items).length) return false;
    if (!i18nGlobal.availableLocales.includes(locale)) {
      i18nGlobal.setLocaleMessage(locale, {});
    }
    i18nGlobal.mergeLocaleMessage(locale, items);
    return true;
  } catch {
    return false;
  }
}

/**
 * 拉取已启用语言；若本地缓存语言已停用，则回退到默认语言（is_default）或列表第一项。
 * 应在登录后、拉取服务端菜单之前调用，保证 X-Lang 正确。
 */
export async function syncLocaleWithEnabledLangs(): Promise<{
  locales: LocaleOption[];
  locale: string;
  changed: boolean;
}> {
  try {
    const { data } = await queryEnabledLangs({ silentError: true });
    const list = [...(data?.list || [])].sort(
      (a, b) => (a.sort_no ?? a.id) - (b.sort_no ?? b.id)
    );
    if (!list.length) {
      return {
        locales: [...LOCALE_OPTIONS],
        locale: getStoredLocale(),
        changed: false,
      };
    }

    const locales: LocaleOption[] = list.map((item) => ({
      label: item.name,
      value: item.lang,
    }));
    const enabled = new Set(list.map((item) => item.lang));
    const current = getStoredLocale();

    if (enabled.has(current)) {
      return { locales, locale: current, changed: false };
    }

    const fallback =
      list.find((item) => item.is_default === 1)?.lang ||
      list[0]?.lang ||
      FALLBACK_LOCALE;
    applyLocale(fallback);
    return { locales, locale: fallback, changed: true };
  } catch {
    return {
      locales: [...LOCALE_OPTIONS],
      locale: getStoredLocale(),
      changed: false,
    };
  }
}

/**
 * 登录页切换语言：写本地、拉词条，不整页刷新（登录页无服务端菜单依赖）
 */
export async function changeLoginLocale(locale: string) {
  applyLocale(locale);
  await loadRemoteDict(locale);
}
