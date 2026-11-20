<template>
  <div class="site-settings">
    <a-card :bordered="false" class="settings-card" :body-style="{ padding: 0 }">
      <a-form :model="form" layout="vertical">
        <a-tabs v-model:active-key="activeTab" class="settings-tabs" size="large">
          <a-tab-pane v-for="tab in tabs" :key="tab.key">
            <template #title>
              <span class="tab-title">
                {{ tab.title }}
                <span v-if="changedTabs.has(tab.key)" class="tab-dot" />
              </span>
            </template>

            <div v-for="g in tab.groups" :key="g.key" class="group">
              <div class="group-head">
                <div class="group-title">{{ g.title }}</div>
                <div class="group-desc">{{ g.desc }}</div>
              </div>

              <!-- 站点信息 -->
              <a-row v-if="g.key === 'site'" :gutter="24">
                <a-col v-bind="col">
                  <a-form-item label="后台域名">
                    <a-input v-model="form.domain" readonly />
                  </a-form-item>
                </a-col>
                <a-col v-bind="col">
                  <a-form-item label="站点title">
                    <a-input v-model="form.title" />
                  </a-form-item>
                </a-col>
                <a-col v-bind="col">
                  <a-form-item label="站点所属公司">
                    <a-input v-model="form.company" />
                  </a-form-item>
                </a-col>
                <a-col v-bind="col">
                  <a-form-item label="货币">
                    <a-input v-model="form.currency" />
                  </a-form-item>
                </a-col>
                <a-col v-bind="col">
                  <a-form-item label="所在时区">
                    <a-select v-model="form.timezone">
                      <a-option v-for="t in timezones" :key="t" :value="t">{{ t }}</a-option>
                    </a-select>
                  </a-form-item>
                </a-col>
                <a-col v-bind="col">
                  <a-form-item label="结算周期">
                    <a-input v-model="form.settlement" />
                  </a-form-item>
                </a-col>
                <a-col :span="24">
                  <a-form-item label="开放语言">
                    <a-space wrap>
                      <a-tag v-for="(l, i) in form.languages" :key="i" color="arcoblue">{{ l }}</a-tag>
                    </a-space>
                  </a-form-item>
                </a-col>
              </a-row>

              <!-- 品牌素材 -->
              <a-row v-else-if="g.key === 'brand'" :gutter="24">
                <a-col v-bind="col">
                  <a-form-item label="站点 LOGO" extra="建议尺寸 178×40，支持 jpg / gif / png">
                    <a-upload :show-file-list="false" action="/" />
                  </a-form-item>
                </a-col>
              </a-row>

              <!-- 业务规则 -->
              <a-row v-else-if="g.key === 'business'" :gutter="24">
                <a-col v-bind="col">
                  <a-form-item label="财务日">
                    <a-select v-model="form.financeDay">
                      <a-option value="每月 01 日">每月 01 日</a-option>
                    </a-select>
                  </a-form-item>
                </a-col>
                <a-col v-bind="col">
                  <a-form-item label="对接业务员">
                    <a-input v-model="form.salesman" />
                  </a-form-item>
                </a-col>
                <a-col v-bind="col">
                  <a-form-item label="站点费率">
                    <span class="text-value">免开版费 / 免线路维护费</span>
                  </a-form-item>
                </a-col>
                <a-col v-bind="col">
                  <a-form-item label="允许会员输入密码错误次数" extra="密码错误时需登案第二次再登录">
                    <a-input v-model="form.pwdErrors">
                      <template #suffix>次</template>
                    </a-input>
                  </a-form-item>
                </a-col>
                <a-col v-bind="col">
                  <a-form-item label="IP最大可注册代理数(现金)">
                    <a-input v-model="form.ipMaxAgent" />
                  </a-form-item>
                </a-col>
                <a-col v-bind="col">
                  <a-form-item label="IP最大可注册会员数">
                    <a-input v-model="form.ipMaxMember" />
                  </a-form-item>
                </a-col>
              </a-row>

              <!-- 代理归属 -->
              <a-row v-else-if="g.key === 'agent'" :gutter="24">
                <a-col v-bind="col">
                  <a-form-item label="现金代理指定直属代理">
                    <a-select v-model="form.cashAgent">
                      <a-option value="ruguagent">ruguagent</a-option>
                    </a-select>
                  </a-form-item>
                </a-col>
                <a-col v-bind="col">
                  <a-form-item label="指定代理">
                    <a-select v-model="form.agent1">
                      <a-option value="a623187898">a623187898</a-option>
                    </a-select>
                  </a-form-item>
                </a-col>
                <a-col v-bind="col">
                  <a-form-item label="指定全民代理">
                    <a-select v-model="form.agent2">
                      <a-option value="a64318840">a64318840</a-option>
                    </a-select>
                  </a-form-item>
                </a-col>
                <a-col v-bind="col">
                  <a-form-item label="指定代理签定全民代理">
                    <a-select v-model="form.agent3">
                      <a-option value="a66318882">a66318882</a-option>
                    </a-select>
                  </a-form-item>
                </a-col>
              </a-row>

              <!-- 运营管理 -->
              <template v-else-if="g.key === 'operation'">
                <a-form-item label="运营国家和地区设置(多选)">
                  <a-checkbox-group v-model="form.countries" class="country-group">
                    <a-checkbox v-for="c in countries" :key="c.value" :value="c.value">{{ c.label }}</a-checkbox>
                  </a-checkbox-group>
                </a-form-item>
                <a-row :gutter="24">
                  <a-col v-bind="col">
                    <a-form-item>
                      <template #label>
                        <span>实力奖池金额显示</span>
                        <a-switch v-model="form.prizeShow" size="small" class="ml8" />
                      </template>
                      <a-input v-model="form.prizeAmount" :disabled="!form.prizeShow" />
                    </a-form-item>
                  </a-col>
                  <a-col v-bind="col">
                    <a-form-item label="探活超时">
                      <a-input v-model="form.probeTimeout">
                        <template #suffix>毫秒</template>
                      </a-input>
                    </a-form-item>
                  </a-col>
                  <a-col v-bind="col">
                    <a-form-item label="重试次数">
                      <a-input v-model="form.retry">
                        <template #suffix>次</template>
                      </a-input>
                    </a-form-item>
                  </a-col>
                </a-row>
              </template>

              <!-- 注册与验证 -->
              <a-row v-else-if="g.key === 'security'" :gutter="24">
                <a-col v-bind="col">
                  <a-form-item label="站点注册方式">
                    <a-checkbox-group v-model="form.registerWays">
                      <a-checkbox value="email">邮箱注册</a-checkbox>
                      <a-checkbox value="phone">手机注册</a-checkbox>
                      <a-checkbox value="account">账号注册</a-checkbox>
                    </a-checkbox-group>
                  </a-form-item>
                </a-col>
                <a-col v-bind="col">
                  <a-form-item label="注册页智能验证(跨脚本)">
                    <a-radio-group v-model="form.registerVerify" type="button">
                      <a-radio v-for="v in verifyOptions" :key="v.value" :value="v.value">{{ v.label }}</a-radio>
                    </a-radio-group>
                  </a-form-item>
                </a-col>
                <a-col v-bind="col">
                  <a-form-item label="登录页智能验证(跨脚本)">
                    <a-radio-group v-model="form.loginVerify" type="button">
                      <a-radio v-for="v in verifyOptions" :key="v.value" :value="v.value">{{ v.label }}</a-radio>
                    </a-radio-group>
                  </a-form-item>
                </a-col>
                <a-col v-bind="col">
                  <a-form-item label="试玩模式">
                    <a-switch v-model="form.trial" />
                  </a-form-item>
                </a-col>
              </a-row>

              <!-- 下载方式 -->
              <a-row v-else-if="g.key === 'download'" :gutter="24">
                <a-col v-bind="col">
                  <a-form-item label="下载方式开关">
                    <a-space>
                      <span>PWA下载页</span>
                      <a-switch v-model="form.pwaPage" size="small" />
                      <a-button size="small">配置已启用</a-button>
                    </a-space>
                  </a-form-item>
                </a-col>
                <a-col v-bind="col">
                  <a-form-item label="PWA 下载页">
                    <a-space>
                      <a-switch v-model="form.pwaDownload" size="small" />
                      <a-button size="small">配置已启用</a-button>
                    </a-space>
                  </a-form-item>
                </a-col>
              </a-row>

              <!-- PWA 安装 -->
              <a-row v-else-if="g.key === 'pwaInstall'" :gutter="24">
                <a-col v-bind="col">
                  <a-form-item label="PWA 安装">
                    <a-switch v-model="form.pwaInstall" size="small" />
                  </a-form-item>
                </a-col>
                <a-col v-bind="col">
                  <a-form-item label="探活超时">
                    <a-input v-model="form.pwaTimeout" :disabled="!form.pwaInstall">
                      <template #suffix>毫秒</template>
                    </a-input>
                  </a-form-item>
                </a-col>
                <a-col v-bind="col">
                  <a-form-item label="重试次数">
                    <a-input v-model="form.pwaRetry" :disabled="!form.pwaInstall">
                      <template #suffix>次</template>
                    </a-input>
                  </a-form-item>
                </a-col>
              </a-row>
            </div>
          </a-tab-pane>
        </a-tabs>
      </a-form>

      <div class="settings-footer">
        <span class="footer-tip" :class="{ dirty }">
          <template v-if="dirty">未保存的修改：{{ changedTabTitles }}</template>
          <template v-else>所有修改已保存</template>
        </span>
        <a-space>
          <a-button :disabled="!dirty" @click="onReset">重置</a-button>
          <a-button type="primary" :disabled="!dirty" @click="onSave">保存设置</a-button>
        </a-space>
      </div>
    </a-card>
  </div>
</template>

<script setup lang="ts">
  import { computed, reactive, ref } from 'vue';
  import { Message } from '@arco-design/web-vue';

  const timezones = ['UTC+08:00', 'UTC+07:00', 'UTC+00:00', 'UTC-03:00'];
  const countries = [
    { value: 'ALL', label: '全部' },
    { value: 'KOR', label: 'KOR(大韩民国)' },
    { value: 'MYS', label: 'MYS(马来西亚)' },
    { value: 'USA', label: 'USA(美利坚合众国)' },
    { value: 'BRA', label: 'BRA(巴西)' },
    { value: 'KHM', label: 'KHM(柬埔寨)' },
    { value: 'JPN', label: 'JPN(日本)' },
    { value: 'VNM', label: 'VNM(越南)' },
    { value: 'GBR', label: 'GBR(英国)' },
    { value: 'IND', label: 'IND(印度)' },
    { value: 'THA', label: 'THA(泰国)' },
    { value: 'MEX', label: 'MEX(墨西哥)' },
  ];
  const verifyOptions = [
    { value: 'none', label: '不验证' },
    { value: 'drag', label: '拖拽验证' },
    { value: 'image', label: '图形验证' },
  ];

  const col = { xs: 24, md: 12, xl: 8 };

  const getDefault = () => ({
    domain: '12312.dubet.vip',
    title: 'betatest',
    currency: 'usdt',
    timezone: 'UTC+08:00',
    settlement: '当日 11:00 至 次日 11:00',
    company: 'axis',
    languages: ['English', 'Tiếng Việt', '简体中文', 'English'],
    countries: ['ALL', 'MYS', 'BRA'],
    registerVerify: 'none',
    pwaPage: true,
    prizeShow: true,
    prizeAmount: '35396105.78',
    probeTimeout: '3000',
    retry: '0',
    financeDay: '每月 01 日',
    salesman: 'pretty',
    pwdErrors: '3',
    ipMaxAgent: '0',
    ipMaxMember: '0',
    cashAgent: 'ruguagent',
    agent1: 'a623187898',
    agent2: 'a64318840',
    agent3: 'a66318882',
    loginVerify: 'none',
    trial: false,
    registerWays: ['email', 'phone'],
    pwaDownload: true,
    pwaInstall: true,
    pwaTimeout: '3000',
    pwaRetry: '0',
  });

  type FormState = ReturnType<typeof getDefault>;
  type FieldKey = keyof FormState;

  interface TabGroup {
    key: string;
    title: string;
    desc: string;
    fields: FieldKey[];
  }

  const tabs: { key: string; title: string; groups: TabGroup[] }[] = [
    {
      key: 'basic',
      title: '基础信息',
      groups: [
        {
          key: 'site',
          title: '站点信息',
          desc: '站点标识、货币与时区',
          fields: ['domain', 'title', 'currency', 'timezone', 'settlement', 'company', 'languages'],
        },
        { key: 'brand', title: '品牌素材', desc: '站点 LOGO 等展示素材', fields: [] },
      ],
    },
    {
      key: 'business',
      title: '业务规则',
      groups: [
        {
          key: 'business',
          title: '业务规则',
          desc: '财务结算、费率与注册限制',
          fields: ['financeDay', 'salesman', 'pwdErrors', 'ipMaxAgent', 'ipMaxMember'],
        },
      ],
    },
    {
      key: 'agent',
      title: '代理归属',
      groups: [
        {
          key: 'agent',
          title: '代理归属',
          desc: '新注册用户默认挂靠的代理',
          fields: ['cashAgent', 'agent1', 'agent2', 'agent3'],
        },
      ],
    },
    {
      key: 'operation',
      title: '运营管理',
      groups: [
        {
          key: 'operation',
          title: '运营管理',
          desc: '运营地区、奖池展示与探活策略',
          fields: ['countries', 'prizeShow', 'prizeAmount', 'probeTimeout', 'retry'],
        },
      ],
    },
    {
      key: 'security',
      title: '注册与安全',
      groups: [
        {
          key: 'security',
          title: '注册与验证',
          desc: '注册方式与登录/注册验证',
          fields: ['registerWays', 'registerVerify', 'loginVerify', 'trial'],
        },
      ],
    },
    {
      key: 'pwa',
      title: '下载与 PWA',
      groups: [
        { key: 'download', title: '下载方式', desc: 'PWA 下载页开关', fields: ['pwaPage', 'pwaDownload'] },
        {
          key: 'pwaInstall',
          title: 'PWA 安装',
          desc: '安装开关与探活策略',
          fields: ['pwaInstall', 'pwaTimeout', 'pwaRetry'],
        },
      ],
    },
  ];

  function clone<T>(v: T): T {
    return JSON.parse(JSON.stringify(v));
  }

  const activeTab = ref(tabs[0].key);
  const form = reactive(getDefault());
  const saved = ref<FormState>(clone(form));

  const isChanged = (key: FieldKey) => JSON.stringify(form[key]) !== JSON.stringify(saved.value[key]);
  const changedTabs = computed(
    () =>
      new Set(
        tabs.filter((t) => t.groups.some((g) => g.fields.some(isChanged))).map((t) => t.key)
      )
  );
  const changedTabTitles = computed(() =>
    tabs
      .filter((t) => changedTabs.value.has(t.key))
      .map((t) => t.title)
      .join('、')
  );
  const dirty = computed(() => changedTabs.value.size > 0);

  const onReset = () => Object.assign(form, clone(saved.value));
  const onSave = () => {
    saved.value = clone(form);
    Message.success('保存成功');
  };
</script>

<style scoped lang="less">
  .site-settings {
    padding: 0 20px 16px;
    background: var(--color-fill-2);
  }
  .settings-card {
    border-radius: 8px;
  }
  .settings-tabs {
    :deep(.arco-tabs-nav) {
      padding: 0 24px;
    }
    :deep(.arco-tabs-content) {
      padding: 0 24px;
    }
  }
  .tab-title {
    position: relative;
    display: inline-flex;
    align-items: center;
  }
  .tab-dot {
    position: absolute;
    top: -2px;
    right: -8px;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: rgb(var(--orangered-6));
  }
  .group {
    padding-top: 20px;
    & + & {
      border-top: 1px dashed var(--color-border-2);
    }
  }
  .group-head {
    display: flex;
    align-items: baseline;
    gap: 12px;
    margin-bottom: 16px;
  }
  .group-title {
    position: relative;
    padding-left: 10px;
    font-size: 15px;
    font-weight: 600;
    &::before {
      content: '';
      position: absolute;
      left: 0;
      top: 3px;
      width: 3px;
      height: 14px;
      border-radius: 2px;
      background: rgb(var(--primary-6));
    }
  }
  .group-desc {
    font-size: 12px;
    color: var(--color-text-3);
  }
  .text-value {
    line-height: 32px;
    color: var(--color-text-1);
  }
  .country-group {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    row-gap: 12px;
    width: 100%;
  }
  .ml8 {
    margin-left: 8px;
  }
  .settings-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 24px;
    border-top: 1px solid var(--color-border-2);
  }
  .footer-tip {
    font-size: 13px;
    color: var(--color-text-3);
    &.dirty {
      color: rgb(var(--orangered-6));
    }
  }
</style>
