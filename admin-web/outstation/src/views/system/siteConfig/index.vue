<template>
  <div class="site-config">
    <!-- header 暂时不需要，先注释掉 -->
    <!-- <header class="page-hero">
      <a-breadcrumb class="page-crumb">
        <a-breadcrumb-item>站点管理</a-breadcrumb-item>
        <a-breadcrumb-item>基础配置</a-breadcrumb-item>
      </a-breadcrumb>
      <h1>站点设置</h1>
      <p>管理品牌信息、访问验证和站点经营范围</p>
    </header> -->

    <a-tabs v-model:active-key="activeTab" class="site-tabs">
      <a-tab-pane key="basic" title="基础设置" />
      <a-tab-pane key="agent" :title="t('agentList.settings.title')" />
    </a-tabs>

    <AgentSettingsPanel v-if="activeTab === 'agent'" />

    <a-spin v-show="activeTab === 'basic'" :loading="loading" class="layout-spin">
      <div class="layout">
        <main class="main">
          <a-card
            :id="sections.basic.id"
            :bordered="false"
            class="section-card"
          >
            <div class="section-head">
              <span class="section-icon" :class="sections.basic.tone">
                <icon-settings />
              </span>
              <div>
                <h2>{{ sections.basic.title }}</h2>
                <p>{{ sections.basic.desc }}</p>
              </div>
            </div>

            <div class="section-body">
              <div class="two-col">
                <div>
                  <div class="field-label">网站标题</div>
                  <a-input
                    v-model="form.title"
                    size="large"
                    :max-length="64"
                    placeholder="请输入网站标题"
                    allow-clear
                  />
                  <div class="field-hint">显示在网站页面与浏览器标签中</div>
                </div>

                <div>
                  <div class="field-label">网站 Logo</div>
                  <div class="logo-field">
                    <a-spin :loading="logoUploading" class="logo-preview">
                      <img v-if="form.logo" :src="form.logo" alt="logo" />
                      <span v-else class="logo-initial">{{ logoInitial }}</span>
                    </a-spin>
                    <div>
                      <a-space :size="8">
                        <a-upload
                          accept=".jpg,.png"
                          :auto-upload="false"
                          :show-file-list="false"
                          :disabled="logoUploading"
                          @change="onLogoChange"
                        >
                          <template #upload-button>
                            <a-button>
                              <template #icon><icon-upload /></template>
                              上传 Logo
                            </a-button>
                          </template>
                        </a-upload>
                        <a-button
                          v-if="form.logo"
                          type="text"
                          status="danger"
                          :disabled="logoUploading"
                          @click="form.logo = ''"
                        >
                          移除
                        </a-button>
                      </a-space>
                      <div class="field-hint">
                        支持 .jpg / .png，不超过 32 MB，边长不超过 4096
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <div class="divider" />

              <div class="sub-head">
                <div>
                  <h3>开放语言</h3>
                  <p>用户可在前台切换以下语言</p>
                </div>
                <div class="sub-actions">
                  <span class="count">
                    {{ form.languages.length }} / {{ languageOptions.length }}
                  </span>
                  <a-button size="small" @click="toggleAllLanguages">
                    {{ allLanguagesChecked ? '取消全选' : '全选' }}
                  </a-button>
                </div>
              </div>
              <a-checkbox-group v-model="form.languages" class="option-grid">
                <a-checkbox
                  v-for="lang in languageOptions"
                  :key="lang.code"
                  :value="lang.code"
                >
                  <span class="option-name">{{ lang.name }}</span>
                  <span
                    v-if="!lang.available"
                    class="option-off"
                    title="基础资料已禁用或缺失，勾选后暂不生效"
                  >
                    不可用
                  </span>
                  <span v-else class="option-code">{{ lang.code }}</span>
                </a-checkbox>
              </a-checkbox-group>
            </div>
          </a-card>

          <a-card
            :id="sections.login.id"
            :bordered="false"
            class="section-card"
          >
            <div class="section-head">
              <span class="section-icon" :class="sections.login.tone">
                <icon-lock />
              </span>
              <div>
                <h2>{{ sections.login.title }}</h2>
                <p>{{ sections.login.desc }}</p>
              </div>
            </div>

            <div class="section-body">
              <div class="limit-grid">
                <div v-for="item in limitFields" :key="item.key">
                  <div class="field-label">{{ item.label }}</div>
                  <a-input-number
                    v-model="form[item.key]"
                    size="large"
                    :min="0"
                    :max="100"
                    :precision="0"
                    hide-button
                  >
                    <template #suffix>{{ item.unit }}</template>
                  </a-input-number>
                </div>
              </div>
              <div class="notice">
                <icon-info-circle />
                数值范围为 0–100，设置为 0 时表示不限制。
              </div>

              <div class="divider" />

              <div class="two-col">
                <div v-for="item in verifyFields" :key="item.key">
                  <div class="field-label">{{ item.label }}</div>
                  <a-radio-group
                    v-model="form[item.key]"
                    type="button"
                    size="large"
                    class="verify-group"
                  >
                    <a-radio
                      v-for="opt in verifyOptions"
                      :key="opt.value"
                      :value="opt.value"
                    >
                      {{ opt.label }}
                    </a-radio>
                  </a-radio-group>
                </div>
              </div>
            </div>
          </a-card>

          <a-card
            :id="sections.region.id"
            :bordered="false"
            class="section-card"
          >
            <div class="section-head">
              <span class="section-icon" :class="sections.region.tone">
                <icon-location />
              </span>
              <div>
                <h2>{{ sections.region.title }}</h2>
                <p>{{ sections.region.desc }}</p>
              </div>
              <a-checkbox
                class="head-extra"
                :model-value="allRegionsChecked"
                :indeterminate="regionIndeterminate"
                @change="toggleAllRegions"
              >
                全选
              </a-checkbox>
            </div>

            <div class="section-body">
              <a-checkbox-group v-model="regionForm" class="option-grid">
                <a-checkbox
                  v-for="region in regionOptions"
                  :key="region.code"
                  :value="region.code"
                >
                  <span class="option-name">{{ region.name }}</span>
                  <span
                    v-if="!region.available"
                    class="option-off"
                    title="基础资料已关闭或缺失，勾选后暂不生效"
                  >
                    不可用
                  </span>
                  <span v-else class="option-code">{{ region.code }}</span>
                </a-checkbox>
              </a-checkbox-group>
            </div>
          </a-card>
        </main>

        <aside class="aside">
          <div class="aside-inner">
            <a-card :bordered="false" class="aside-card overview">
              <div class="overview-brand">
                <div class="brand-logo">
                  <img v-if="form.logo" :src="form.logo" alt="logo" />
                  <span v-else>{{ logoInitial }}</span>
                </div>
                <div class="brand-text">
                  <strong>{{ form.title || '未命名站点' }}</strong>
                  <span>{{ siteInfo[0].value }}</span>
                </div>
              </div>

              <div class="overview-stats">
                <div>
                  <strong>{{ form.languages.length }}</strong>
                  <span>开放语言</span>
                </div>
                <div>
                  <strong>{{ regionForm.length }}</strong>
                  <span>经营地区</span>
                </div>
              </div>

              <ul class="overview-list">
                <li v-for="item in siteInfo.slice(1)" :key="item.label">
                  <span>{{ item.label }}</span>
                  <strong>{{ item.value }}</strong>
                </li>
                <li>
                  <span>登录验证</span>
                  <strong>{{ verifyLabel(form.loginVerify) }}</strong>
                </li>
                <li>
                  <span>注册验证</span>
                  <strong>{{ verifyLabel(form.registerVerify) }}</strong>
                </li>
              </ul>
            </a-card>

            <a-card :bordered="false" class="aside-card">
              <div class="aside-title">配置导航</div>
              <a
                v-for="item in sectionList"
                :key="item.id"
                class="nav-item"
                @click="scrollTo(item.id)"
              >
                <span class="nav-mark" :class="item.tone" />
                <span class="nav-text">{{ item.title }}</span>
                <span v-if="sectionDirty[item.key]" class="nav-tag"
                  >未保存</span
                >
                <icon-right v-else class="nav-arrow" />
              </a>
            </a-card>

            <a-card :bordered="false" class="aside-card">
              <div class="save-state" :class="{ dirty }">
                <span class="state-dot" />
                {{ dirty ? '有未保存的修改' : '全部设置均已保存' }}
              </div>
              <div class="aside-actions">
                <a-button long :disabled="!dirty || saving" @click="onReset">
                  取消修改
                </a-button>
                <a-button
                  type="primary"
                  long
                  :loading="saving"
                  :disabled="!dirty"
                  @click="onSave"
                >
                  保存
                </a-button>
              </div>
            </a-card>
          </div>
        </aside>
      </div>
    </a-spin>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FileItem } from '@arco-design/web-vue';
  import {
    getSiteInfo,
    updateSiteLanguages,
    updateSiteRegions,
    updateSiteSettings,
    uploadSiteLogo,
    SITE_CAPTCHA_MODE,
    SITE_OPTION_STATUS,
  } from '@/api/siteConfig';
  import type { SiteInfoRes, SiteRegionBase } from '@/api/siteConfig';
  import AgentSettingsPanel from './components/agent-settings-panel.vue';

  const activeTab = ref<'basic' | 'agent'>('basic');

  const LOGO_MAX_SIZE = 32 * 1024 * 1024;
  const LOGO_MAX_EDGE = 4096;
  /** 后端目前不接受 .jpeg 扩展名 */
  const LOGO_EXT = /\.(jpg|png)$/i;

  const sections = {
    basic: {
      key: 'basic',
      id: 'site-basic',
      title: '网站基础配置',
      desc: '设置站点名称、品牌标识与前台展示语言',
      tone: 'blue',
    },
    login: {
      key: 'login',
      id: 'site-login',
      title: '登录与注册设置',
      desc: '配置访问频率限制与身份验证方式',
      tone: 'orange',
    },
    region: {
      key: 'region',
      id: 'site-region',
      title: '经营地区管理',
      desc: '选择站点当前提供服务的国家与地区',
      tone: 'green',
    },
  } as const;
  const sectionList = Object.values(sections);

  const { t, te, locale } = useI18n();

  interface SiteOption {
    code: string;
    name: string;
    /** 基础资料存在且未被禁用；本站勾选后才会实际生效 */
    available: boolean;
  }

  const site = ref({
    operatorCode: '',
    domain: '',
    currency: '',
    timezone: '',
  });
  const siteInfo = computed(() => [
    { label: '后台域名', value: site.value.domain || '-' },
    { label: '结算货币', value: site.value.currency || '-' },
    { label: '所在时区', value: site.value.timezone || '-' },
    { label: '站点编码', value: site.value.operatorCode || '-' },
  ]);

  const languageOptions = ref<SiteOption[]>([]);
  const regionOptions = ref<SiteOption[]>([]);

  const verifyOptions = [
    { value: SITE_CAPTCHA_MODE.NONE, label: '不验证' },
    { value: SITE_CAPTCHA_MODE.DRAG, label: '拖拽验证' },
    { value: SITE_CAPTCHA_MODE.IMAGE, label: '图形验证' },
  ];
  const verifyLabel = (value: number) =>
    verifyOptions.find((o) => o.value === value)?.label ?? '-';

  const getDefault = () => ({
    title: '',
    logo: '',
    languages: [] as string[],
    emailLoginLimit: 0,
    smsLoginLimit: 0,
    pwdErrorLimit: 0,
    ipRegisterLimit: 0,
    loginVerify: SITE_CAPTCHA_MODE.NONE as number,
    registerVerify: SITE_CAPTCHA_MODE.NONE as number,
  });

  type FormState = ReturnType<typeof getDefault>;
  type FormKey = keyof FormState;

  const limitFields: {
    key:
      | 'emailLoginLimit'
      | 'smsLoginLimit'
      | 'pwdErrorLimit'
      | 'ipRegisterLimit';
    label: string;
    unit: string;
  }[] = [
    { key: 'emailLoginLimit', label: '邮箱登录阈值', unit: '次' },
    { key: 'smsLoginLimit', label: '短信登录阈值', unit: '次' },
    { key: 'pwdErrorLimit', label: '会员密码错误次数上限', unit: '次' },
    { key: 'ipRegisterLimit', label: '同一 IP 最多注册会员数', unit: '人' },
  ];

  const verifyFields: {
    key: 'loginVerify' | 'registerVerify';
    label: string;
  }[] = [
    { key: 'loginVerify', label: '登录页智能验证' },
    { key: 'registerVerify', label: '注册页智能验证' },
  ];

  const basicKeys: FormKey[] = ['title', 'logo', 'languages'];
  const loginKeys: FormKey[] = [
    ...limitFields.map((f) => f.key),
    ...verifyFields.map((f) => f.key),
  ];
  /** PUT /operator/settings 覆盖的字段（不含开放语言） */
  const settingsKeys: FormKey[] = ['title', 'logo', ...loginKeys];

  function clone<T>(v: T): T {
    return JSON.parse(JSON.stringify(v));
  }

  const form = reactive<FormState>(getDefault());
  const saved = ref<FormState>(clone(form));
  const isChanged = (key: FormKey) =>
    JSON.stringify(form[key]) !== JSON.stringify(saved.value[key]);
  const settingsDirty = computed(() => settingsKeys.some(isChanged));
  const languagesDirty = computed(() => isChanged('languages'));

  const regionForm = ref<string[]>([]);
  const savedRegions = ref<string[]>([]);
  const regionDirty = computed(
    () =>
      regionForm.value.length !== savedRegions.value.length ||
      regionForm.value.some((code) => !savedRegions.value.includes(code))
  );
  const dirty = computed(
    () =>
      settingsDirty.value || languagesDirty.value || regionDirty.value
  );

  const logoInitial = computed(
    () => form.title.trim().charAt(0).toUpperCase() || 'S'
  );

  const allLanguagesChecked = computed(
    () =>
      languageOptions.value.length > 0 &&
      form.languages.length === languageOptions.value.length
  );
  function toggleAllLanguages() {
    form.languages = allLanguagesChecked.value
      ? []
      : languageOptions.value.map((l) => l.code);
  }

  const allRegionsChecked = computed(
    () =>
      regionOptions.value.length > 0 &&
      regionForm.value.length === regionOptions.value.length
  );
  const regionIndeterminate = computed(
    () => regionForm.value.length > 0 && !allRegionsChecked.value
  );
  function toggleAllRegions(checked: boolean | (string | number | boolean)[]) {
    regionForm.value = checked ? regionOptions.value.map((r) => r.code) : [];
  }

  function readImageSize(file: File) {
    return new Promise<{ width: number; height: number }>((resolve, reject) => {
      const url = URL.createObjectURL(file);
      const img = new Image();
      img.onload = () => {
        URL.revokeObjectURL(url);
        resolve({ width: img.naturalWidth, height: img.naturalHeight });
      };
      img.onerror = () => {
        URL.revokeObjectURL(url);
        reject(new Error('invalid image'));
      };
      img.src = url;
    });
  }

  const logoUploading = ref(false);
  async function onLogoChange(_: FileItem[], fileItem: FileItem) {
    const { file } = fileItem;
    if (!file || logoUploading.value) return;
    if (!LOGO_EXT.test(file.name)) {
      Message.warning('仅支持 .jpg / .png（暂不支持 .jpeg）');
      return;
    }
    if (file.size > LOGO_MAX_SIZE) {
      Message.warning('Logo 大小不能超过 32 MB');
      return;
    }
    logoUploading.value = true;
    try {
      const { width, height } = await readImageSize(file);
      if (width > LOGO_MAX_EDGE || height > LOGO_MAX_EDGE) {
        Message.warning('Logo 宽高不能超过 4096 像素');
        return;
      }
      const { data } = await uploadSiteLogo(file);
      form.logo = data.url;
    } catch {
      // 错误提示由请求拦截器统一弹出
    } finally {
      logoUploading.value = false;
    }
  }

  function onReset() {
    Object.assign(form, clone(saved.value));
    regionForm.value = [...savedRegions.value];
  }

  const saving = ref(false);
  async function onSave() {
    const needSettings = settingsDirty.value;
    const needLanguages = languagesDirty.value;
    const needRegions = regionDirty.value;
    if (!needSettings && !needLanguages && !needRegions) return;

    if (needSettings && !form.title.trim()) {
      Message.warning('请输入网站标题');
      return;
    }
    // 接口允许 []，但全关会影响业务，前端先拦住
    if (needLanguages && !form.languages.length) {
      Message.warning('请至少选择一种开放语言');
      return;
    }
    if (needRegions && !regionForm.value.length) {
      Message.warning('请至少选择一个经营地区');
      return;
    }
    if (saving.value) return;
    saving.value = true;
    try {
      const tasks: Promise<void>[] = [];

      if (needSettings) {
        tasks.push(
          updateSiteSettings({
            site_title: form.title.trim(),
            site_logo_url: form.logo || '',
            email_login_limit: form.emailLoginLimit ?? 0,
            sms_login_limit: form.smsLoginLimit ?? 0,
            member_password_error_limit: form.pwdErrorLimit ?? 0,
            ip_member_register_limit: form.ipRegisterLimit ?? 0,
            register_captcha_mode: form.registerVerify,
            login_captcha_mode: form.loginVerify,
          }).then(({ data }) => {
            form.title = data.site_title || '';
            form.logo = data.site_logo_url || '';
            form.emailLoginLimit = data.email_login_limit;
            form.smsLoginLimit = data.sms_login_limit;
            form.pwdErrorLimit = data.member_password_error_limit;
            form.ipRegisterLimit = data.ip_member_register_limit;
            form.registerVerify = data.register_captcha_mode;
            form.loginVerify = data.login_captcha_mode;
            Object.assign(saved.value, {
              title: form.title,
              logo: form.logo,
              emailLoginLimit: form.emailLoginLimit,
              smsLoginLimit: form.smsLoginLimit,
              pwdErrorLimit: form.pwdErrorLimit,
              ipRegisterLimit: form.ipRegisterLimit,
              registerVerify: form.registerVerify,
              loginVerify: form.loginVerify,
            });
          })
        );
      }

      if (needLanguages) {
        const codes = [...form.languages];
        tasks.push(
          updateSiteLanguages({ language_codes: codes }).then(({ data }) => {
            form.languages = data.language_codes || codes;
            saved.value.languages = [...form.languages];
          })
        );
      }

      if (needRegions) {
        const codes = [...regionForm.value];
        tasks.push(
          updateSiteRegions({ region_codes: codes }).then(({ data }) => {
            regionForm.value = data.region_codes || codes;
            savedRegions.value = [...regionForm.value];
          })
        );
      }

      await Promise.all(tasks);
      Message.success('保存成功');
    } catch {
      // 错误提示由请求拦截器统一弹出；部分成功时已写入的 saved 会保留
    } finally {
      saving.value = false;
    }
  }

  const sectionDirty = computed<Record<string, boolean>>(() => ({
    basic: basicKeys.some(isChanged),
    login: loginKeys.some(isChanged),
    region: regionDirty.value,
  }));

  function scrollTo(id: string) {
    document
      .getElementById(id)
      ?.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }

  function regionName(region: SiteRegionBase) {
    if (region.name_key && te(region.name_key)) return t(region.name_key);
    try {
      return (
        new Intl.DisplayNames([locale.value], { type: 'region' }).of(
          region.code
        ) || region.code
      );
    } catch {
      return region.code;
    }
  }

  function applySiteInfo(data: SiteInfoRes) {
    const { site_config: config, currency, timezone } = data;
    const settings = data.site_settings;
    const currencyCode = currency?.code || config.settlement_currency_code;

    site.value = {
      operatorCode: config.operator_code || settings.operator_code,
      domain: config.domain_name,
      currency:
        currencyCode && currency?.symbol
          ? `${currencyCode}（${currency.symbol}）`
          : currencyCode,
      timezone: timezone?.code || config.timezone_code,
    };

    const languages = [...(data.languages || [])].sort(
      (a, b) => a.sort_no - b.sort_no
    );
    languageOptions.value = languages.map((l) => ({
      code: l.lang,
      name: l.name || l.lang,
      available: l.id > 0 && l.disabled === 0,
    }));

    const regions = [...(data.region_options || [])].sort(
      (a, b) => a.region.sort_no - b.region.sort_no
    );
    regionOptions.value = regions.map(({ region }) => ({
      code: region.code,
      name: regionName(region),
      available: region.id > 0 && region.status === SITE_OPTION_STATUS.ON,
    }));
    regionForm.value = regions
      .filter((r) => r.status === SITE_OPTION_STATUS.ON)
      .map((r) => r.region.code);
    savedRegions.value = [...regionForm.value];

    Object.assign(form, {
      title: settings.site_title || '',
      logo: settings.site_logo_url || '',
      languages: languages
        .filter((l) => l.status === SITE_OPTION_STATUS.ON)
        .map((l) => l.lang),
      emailLoginLimit: settings.email_login_limit,
      smsLoginLimit: settings.sms_login_limit,
      pwdErrorLimit: settings.member_password_error_limit,
      ipRegisterLimit: settings.ip_member_register_limit,
      loginVerify: settings.login_captcha_mode,
      registerVerify: settings.register_captcha_mode,
    });
    saved.value = clone(form);
  }

  const loading = ref(false);
  async function loadSiteInfo() {
    loading.value = true;
    try {
      const { data } = await getSiteInfo();
      applySiteInfo(data);
    } catch {
      // 错误提示由请求拦截器统一弹出
    } finally {
      loading.value = false;
    }
  }

  onMounted(loadSiteInfo);
</script>

<style scoped lang="less">
  .site-config {
    padding: 0 20px 24px;
  }

  .site-tabs {
    margin-bottom: 16px;

    :deep(.arco-tabs-nav) {
      margin-bottom: 0;
    }

    :deep(.arco-tabs-content) {
      display: none;
    }
  }

  .page-hero {
    margin-bottom: 18px;

    h1 {
      margin: 10px 0 4px;
      font-size: 24px;
      font-weight: 600;
      line-height: 1.3;
      color: var(--color-text-1);
    }

    p {
      margin: 0;
      font-size: 13px;
      color: var(--color-text-3);
    }
  }

  .page-crumb {
    font-size: 12px;
  }

  .layout-spin {
    display: block;
    width: 100%;
  }

  .layout {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 300px;
    gap: 16px;
    align-items: start;
  }

  .main {
    display: flex;
    flex-direction: column;
    gap: 16px;
    min-width: 0;
  }

  .section-card,
  .aside-card {
    border-radius: 10px;
    box-shadow: 0 1px 3px rgba(15, 23, 42, 0.04);
  }

  .section-card {
    scroll-margin-top: 16px;

    :deep(.arco-card-body) {
      padding: 0;
    }
  }

  .section-head {
    display: flex;
    gap: 12px;
    align-items: center;
    padding: 18px 24px;
    border-bottom: 1px solid var(--color-border-1);

    h2 {
      margin: 0 0 2px;
      font-size: 15px;
      font-weight: 600;
      color: var(--color-text-1);
    }

    p {
      margin: 0;
      font-size: 12px;
      color: var(--color-text-3);
    }
  }

  .head-extra {
    margin-left: auto;
  }

  .section-icon {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    font-size: 18px;
    border-radius: 9px;
  }

  .blue {
    color: rgb(var(--primary-6));
    background: rgb(var(--primary-1));
  }

  .orange {
    color: rgb(var(--orange-6));
    background: rgb(var(--orange-1));
  }

  .green {
    color: rgb(var(--green-6));
    background: rgb(var(--green-1));
  }

  .section-body {
    padding: 22px 24px 24px;
  }

  .two-col,
  .limit-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 20px 24px;
  }

  .divider {
    height: 1px;
    margin: 24px 0;
    background: var(--color-border-1);
  }

  .field-label {
    margin-bottom: 8px;
    font-size: 13px;
    font-weight: 500;
    color: var(--color-text-1);
  }

  .field-hint {
    margin-top: 6px;
    font-size: 12px;
    color: var(--color-text-3);
  }

  .logo-field {
    display: flex;
    gap: 14px;
    align-items: center;
  }

  .logo-preview {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    width: 56px;
    height: 56px;
    overflow: hidden;
    background: var(--color-fill-2);
    border: 1px dashed var(--color-border-3);
    border-radius: 10px;

    img {
      width: 100%;
      height: 100%;
      object-fit: contain;
    }
  }

  .logo-initial {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 34px;
    height: 34px;
    font-size: 18px;
    font-weight: 700;
    color: #fff;
    background: rgb(var(--primary-6));
    border-radius: 8px;
  }

  .sub-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 14px;

    h3 {
      margin: 0 0 2px;
      font-size: 13px;
      font-weight: 600;
      color: var(--color-text-1);
    }

    p {
      margin: 0;
      font-size: 12px;
      color: var(--color-text-3);
    }
  }

  .sub-actions {
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .count {
    font-size: 13px;
    font-weight: 600;
    color: rgb(var(--primary-6));
  }

  .option-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 10px;
    width: 100%;

    :deep(.arco-checkbox) {
      display: flex;
      align-items: center;
      height: 40px;
      margin-right: 0;
      padding: 0 12px;
      border: 1px solid var(--color-border-2);
      border-radius: 8px;
      transition: border-color 0.2s, background-color 0.2s;

      &:hover {
        border-color: rgb(var(--primary-4));
      }

      &.arco-checkbox-checked {
        background: rgb(var(--primary-1));
        border-color: rgb(var(--primary-4));
      }
    }

    :deep(.arco-checkbox-label) {
      display: flex;
      flex: 1;
      align-items: center;
      justify-content: space-between;
      min-width: 0;
      font-size: 13px;
    }
  }

  .option-name {
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .option-code {
    flex-shrink: 0;
    margin-left: 8px;
    font-size: 12px;
    color: var(--color-text-4);
  }

  .option-off {
    flex-shrink: 0;
    margin-left: 8px;
    padding: 0 6px;
    font-size: 12px;
    line-height: 18px;
    color: rgb(var(--orange-6));
    background: rgb(var(--orange-1));
    border-radius: 4px;
  }

  .limit-grid :deep(.arco-input-number) {
    width: 100%;
  }

  .notice {
    display: flex;
    gap: 6px;
    align-items: center;
    margin-top: 14px;
    padding: 9px 12px;
    font-size: 12px;
    color: var(--color-text-3);
    background: var(--color-fill-1);
    border-radius: 6px;
  }

  .verify-group {
    display: flex;
    width: 100%;

    :deep(.arco-radio-button) {
      flex: 1;
      text-align: center;
    }
  }

  .aside {
    position: sticky;
    top: 16px;
  }

  .aside-inner {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .aside-card :deep(.arco-card-body) {
    padding: 18px 20px;
  }

  .aside-title {
    margin-bottom: 10px;
    font-size: 13px;
    font-weight: 600;
    color: var(--color-text-1);
  }

  .overview-brand {
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .brand-logo {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    width: 44px;
    height: 44px;
    overflow: hidden;
    font-size: 20px;
    font-weight: 700;
    color: #fff;
    background: linear-gradient(
      135deg,
      rgb(var(--primary-5)),
      rgb(var(--primary-7))
    );
    border-radius: 10px;

    img {
      width: 100%;
      height: 100%;
      object-fit: contain;
      background: var(--color-bg-2);
    }
  }

  .brand-text {
    display: flex;
    flex-direction: column;
    min-width: 0;

    strong {
      overflow: hidden;
      font-size: 15px;
      color: var(--color-text-1);
      white-space: nowrap;
      text-overflow: ellipsis;
    }

    span {
      font-size: 12px;
      color: var(--color-text-3);
    }
  }

  .overview-stats {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
    margin: 16px 0 6px;

    div {
      padding: 10px 12px;
      background: var(--color-fill-1);
      border-radius: 8px;
    }

    strong {
      display: block;
      font-size: 20px;
      line-height: 1.3;
      color: var(--color-text-1);
    }

    span {
      font-size: 12px;
      color: var(--color-text-3);
    }
  }

  .overview-list {
    margin: 0;
    padding: 0;
    list-style: none;

    li {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 9px 0;
      font-size: 13px;

      & + li {
        border-top: 1px dashed var(--color-border-2);
      }
    }

    span {
      color: var(--color-text-3);
    }

    strong {
      font-weight: 500;
      color: var(--color-text-1);
    }
  }

  .nav-item {
    display: flex;
    gap: 10px;
    align-items: center;
    margin: 0 -10px;
    padding: 9px 10px;
    font-size: 13px;
    color: var(--color-text-2);
    border-radius: 6px;
    cursor: pointer;
    transition: background-color 0.2s;

    &:hover {
      background: var(--color-fill-2);
    }
  }

  .nav-mark {
    width: 8px;
    height: 8px;
    border-radius: 50%;

    &.blue {
      background: rgb(var(--primary-6));
    }

    &.orange {
      background: rgb(var(--orange-6));
    }

    &.green {
      background: rgb(var(--green-6));
    }
  }

  .nav-text {
    flex: 1;
  }

  .nav-tag {
    padding: 0 6px;
    font-size: 12px;
    line-height: 20px;
    color: rgb(var(--orange-6));
    background: rgb(var(--orange-1));
    border-radius: 4px;
  }

  .nav-arrow {
    color: var(--color-text-4);
  }

  .save-state {
    display: flex;
    gap: 8px;
    align-items: center;
    font-size: 13px;
    color: var(--color-text-3);

    &.dirty {
      color: rgb(var(--orange-6));
    }
  }

  .state-dot {
    width: 6px;
    height: 6px;
    background: rgb(var(--green-6));
    border-radius: 50%;

    .dirty & {
      background: rgb(var(--orange-6));
    }
  }

  .aside-actions {
    display: flex;
    gap: 10px;
    margin-top: 14px;
  }

  @media (max-width: 1200px) {
    .layout {
      grid-template-columns: 1fr;
    }

    .aside {
      position: static;
      grid-row: 1;
    }

    .aside-inner {
      display: grid;
      grid-template-columns: repeat(3, minmax(0, 1fr));
    }
  }

  @media (max-width: 768px) {
    .site-config {
      padding: 0 12px 16px;
    }

    .aside-inner,
    .two-col,
    .limit-grid {
      grid-template-columns: 1fr;
    }

    .section-head,
    .section-body {
      padding-right: 16px;
      padding-left: 16px;
    }
  }
</style>
