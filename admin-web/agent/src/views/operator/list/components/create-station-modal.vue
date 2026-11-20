<template>
  <a-modal
    :visible="visible"
    :width="1040"
    :footer="false"
    :mask-closable="false"
    :esc-to-close="false"
    unmount-on-close
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel station-create-modal"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    @cancel="handleCancel"
  >
    <div class="wizard">
      <a-steps :current="currentStep" class="wizard-steps" small>
        <a-step v-for="name in stepNames" :key="name">{{ name }}</a-step>
      </a-steps>

      <a-spin :loading="optionsLoading" style="width: 100%">
        <!-- Step 1: 基础信息 -->
        <div v-if="currentStep === 1" class="step-body">
          <div class="step-main">
            <div class="step-title">
              {{ t('operatorStation.create.baseInfo.title') }}
            </div>
            <div class="step-desc">
              {{ t('operatorStation.create.baseInfo.desc') }}
            </div>
            <a-form
              ref="baseInfoFormRef"
              :model="form"
              :rules="baseInfoRules"
              layout="vertical"
              class="step-form"
            >
              <a-row :gutter="16">
                <a-col :span="12">
                  <a-form-item
                    :label="t('operatorStation.create.baseInfo.name')"
                    field="name"
                    required
                  >
                    <a-input
                      v-model="form.name"
                      :placeholder="
                        t('operatorStation.create.baseInfo.namePlaceholder')
                      "
                      allow-clear
                      :max-length="100"
                    />
                  </a-form-item>
                </a-col>
                <a-col :span="12">
                  <a-form-item
                    :label="t('operatorStation.create.baseInfo.code')"
                    field="code"
                  >
                    <a-input
                      v-model="form.code"
                      :placeholder="
                        t('operatorStation.create.baseInfo.codePlaceholder')
                      "
                      :disabled="Boolean(operatorId)"
                      allow-clear
                      :max-length="32"
                    />
                    <template #extra>
                      {{ t('operatorStation.create.baseInfo.codeTip') }}
                    </template>
                  </a-form-item>
                </a-col>
              </a-row>
              <a-row :gutter="16">
                <a-col :span="12">
                  <a-form-item
                    :label="t('operatorStation.create.baseInfo.timezone')"
                    field="timezone_code"
                    required
                  >
                    <a-select
                      v-model="form.timezone_code"
                      :placeholder="
                        t('operatorStation.create.baseInfo.timezonePlaceholder')
                      "
                      :disabled="baseLocked"
                      allow-clear
                      allow-search
                      :options="timezoneOptions"
                    />
                  </a-form-item>
                </a-col>
                <a-col :span="12">
                  <a-form-item
                    :label="t('operatorStation.create.baseInfo.currency')"
                    field="settlement_currency_code"
                    required
                  >
                    <a-select
                      v-model="form.settlement_currency_code"
                      :placeholder="
                        t('operatorStation.create.baseInfo.currencyPlaceholder')
                      "
                      :disabled="baseLocked"
                      allow-clear
                      allow-search
                      :options="currencyOptions"
                    />
                  </a-form-item>
                </a-col>
              </a-row>
              <a-form-item
                :label="t('operatorStation.create.baseInfo.status')"
                field="status"
                required
              >
                <a-select
                  v-model="form.status"
                  :placeholder="
                    t('operatorStation.create.baseInfo.statusPlaceholder')
                  "
                  :options="statusOptions"
                  class="status-select"
                />
              </a-form-item>
              <a-form-item
                :label="t('operatorStation.create.baseInfo.remark')"
                field="remark"
              >
                <a-textarea
                  v-model="form.remark"
                  :placeholder="
                    t('operatorStation.create.baseInfo.remarkPlaceholder')
                  "
                  :max-length="1000"
                  show-word-limit
                  allow-clear
                  :auto-size="{ minRows: 3, maxRows: 5 }"
                />
              </a-form-item>
            </a-form>
          </div>
          <div class="step-side">
            <a-alert type="info" :show-icon="true">
              <template #title>
                {{ t('operatorStation.create.baseInfo.tipTitle') }}
              </template>
              {{ t('operatorStation.create.baseInfo.tipContent') }}
            </a-alert>
          </div>
        </div>

        <!-- Step 2: 档案信息 -->
        <div v-else-if="currentStep === 2" class="step-body step-body--full">
          <div class="step-main">
            <div class="current-station-bar">
              <span>
                {{
                  t('operatorStation.create.profileInfo.statusBar', {
                    code: form.code || '--',
                    name: form.name || '--',
                  })
                }}
              </span>
              <a-tag :color="isCompletedEdit ? 'green' : 'gray'">
                <span class="draft-dot" :class="{ 'draft-dot--done': isCompletedEdit }" />
                {{
                  isCompletedEdit
                    ? t('operatorStation.creationStatus.completed')
                    : t('operatorStation.create.profileInfo.statusDraft')
                }}
              </a-tag>
            </div>
            <a-alert type="info" :show-icon="true" class="step-alert">
              {{ t('operatorStation.create.profileInfo.tip') }}
            </a-alert>
            <a-form
              ref="profileFormRef"
              :model="profileForm"
              :rules="profileRules"
              layout="vertical"
              class="step-form step-form--narrow"
            >
              <a-form-item
                :label="t('operatorStation.create.profileInfo.companyName')"
                field="companyName"
              >
                <a-input
                  v-model="profileForm.companyName"
                  :placeholder="
                    t(
                      'operatorStation.create.profileInfo.companyNamePlaceholder'
                    )
                  "
                  allow-clear
                  :max-length="200"
                />
              </a-form-item>
              <a-form-item
                :label="t('operatorStation.create.profileInfo.contactName')"
                field="contactName"
              >
                <a-input
                  v-model="profileForm.contactName"
                  :placeholder="
                    t(
                      'operatorStation.create.profileInfo.contactNamePlaceholder'
                    )
                  "
                  allow-clear
                  :max-length="100"
                />
              </a-form-item>
              <a-form-item
                :label="t('operatorStation.create.profileInfo.contactEmail')"
                field="contactEmail"
              >
                <a-input
                  v-model="profileForm.contactEmail"
                  :placeholder="
                    t(
                      'operatorStation.create.profileInfo.contactEmailPlaceholder'
                    )
                  "
                  allow-clear
                  :max-length="255"
                />
              </a-form-item>
              <a-form-item
                :label="t('operatorStation.create.profileInfo.remark')"
                field="remark"
              >
                <a-textarea
                  v-model="profileForm.remark"
                  :placeholder="
                    t('operatorStation.create.profileInfo.remarkPlaceholder')
                  "
                  :max-length="1000"
                  show-word-limit
                  allow-clear
                  :auto-size="{ minRows: 3, maxRows: 5 }"
                />
              </a-form-item>
            </a-form>
          </div>
        </div>

        <!-- Step 3: 管理员账号 -->
        <div v-else-if="currentStep === 3" class="step-body">
          <div class="step-main">
            <div class="step-title">
              {{ t('operatorStation.create.adminAccount.title') }}
            </div>
            <div class="step-desc">
              {{ t('operatorStation.create.adminAccount.desc') }}
            </div>
            <a-spin :loading="adminLoading" style="width: 100%">
              <a-form
                ref="adminFormRef"
                :model="adminForm"
                :rules="adminRules"
                layout="vertical"
                class="step-form step-form--narrow"
              >
                <a-form-item
                  :label="t('operatorStation.create.adminAccount.account')"
                  field="account"
                  required
                >
                  <a-input
                    v-model="adminForm.account"
                    :placeholder="
                      t(
                        'operatorStation.create.adminAccount.accountPlaceholder'
                      )
                    "
                    allow-clear
                    :max-length="64"
                  />
                </a-form-item>
                <a-form-item
                  :label="t('operatorStation.create.adminAccount.password')"
                  field="password"
                  :required="!adminId"
                >
                  <a-input-password
                    v-model="adminForm.password"
                    :placeholder="
                      adminId
                        ? t(
                            'operatorStation.create.adminAccount.passwordOptionalPlaceholder'
                          )
                        : t(
                            'operatorStation.create.adminAccount.passwordPlaceholder'
                          )
                    "
                    allow-clear
                    :max-length="32"
                  />
                </a-form-item>
                <a-form-item
                  :label="
                    t('operatorStation.create.adminAccount.confirmPassword')
                  "
                  field="confirmPassword"
                  :required="!adminId || Boolean(adminForm.password)"
                >
                  <a-input-password
                    v-model="adminForm.confirmPassword"
                    :placeholder="
                      adminId
                        ? t(
                            'operatorStation.create.adminAccount.confirmPasswordOptionalPlaceholder'
                          )
                        : t(
                            'operatorStation.create.adminAccount.confirmPasswordPlaceholder'
                          )
                    "
                    allow-clear
                    :max-length="32"
                  />
                </a-form-item>
                <a-form-item
                  :label="t('operatorStation.create.adminAccount.displayName')"
                  field="displayName"
                  required
                >
                  <a-input
                    v-model="adminForm.displayName"
                    :placeholder="
                      t(
                        'operatorStation.create.adminAccount.displayNamePlaceholder'
                      )
                    "
                    allow-clear
                    :max-length="100"
                  />
                </a-form-item>
                <a-form-item
                  :label="t('operatorStation.create.adminAccount.status')"
                  field="status"
                >
                  <a-select
                    v-model="adminForm.status"
                    :options="adminStatusOptions"
                    class="status-select"
                  />
                </a-form-item>
              </a-form>
            </a-spin>
            <a-alert type="info" :show-icon="true" class="step-alert">
              {{ t('operatorStation.create.adminAccount.tip') }}
            </a-alert>
          </div>
          <div class="step-side">
            <div class="side-panel-title">
              {{ t('operatorStation.create.adminAccount.summaryTitle') }}
            </div>
            <a-descriptions :column="1" :label-style="{ width: '80px' }">
              <a-descriptions-item
                :label="t('operatorStation.create.baseInfo.name')"
              >
                {{ form.name || '--' }}
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('operatorStation.create.baseInfo.timezone')"
              >
                {{ timezoneLabel }}
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('operatorStation.create.baseInfo.currency')"
              >
                {{ currencyLabel }}
              </a-descriptions-item>
            </a-descriptions>
          </div>
        </div>

        <!-- Step 4: 域名配置 -->
        <div v-else-if="currentStep === 4" class="step-body">
          <div class="step-main">
            <div class="step-title">
              {{ t('operatorStation.create.domain.title') }}
            </div>
            <a-button
              type="primary"
              class="add-domain-btn"
              @click="openDomainModal()"
            >
              <template #icon><icon-plus /></template>
              {{ t('operatorStation.create.domain.addButton') }}
            </a-button>
            <a-table
              row-key="id"
              :loading="domainLoading"
              :columns="domainColumns"
              :data="domainList"
              :pagination="false"
              :bordered="true"
              size="small"
            >
              <template #status="{ record }">
                <a-tag :color="record.status === 1 ? 'green' : 'gray'">
                  {{
                    record.status === 1
                      ? t('front.common.status.enabled')
                      : t('front.common.status.disabled')
                  }}
                </a-tag>
              </template>
              <template #operations="{ record }">
                <a-space :size="8">
                  <a-link @click="openDomainModal(record)">
                    {{ t('front.common.action.edit') }}
                  </a-link>
                  <a-link status="danger" @click="handleDeleteDomain(record)">
                    {{ t('front.common.action.delete') }}
                  </a-link>
                </a-space>
              </template>
            </a-table>
          </div>
          <div class="step-side">
            <a-alert type="info" :show-icon="true">
              <template #title>
                {{ t('operatorStation.create.domain.tipTitle') }}
              </template>
              {{ t('operatorStation.create.domain.tip') }}
            </a-alert>
          </div>
        </div>

        <!-- Step 5: 基础资源 -->
        <div v-else class="step-body">
          <div class="step-main">
            <div class="step-title">
              {{ t('operatorStation.create.resource.title') }}
            </div>
            <a-tabs v-model:active-key="resourceState.activeTab">
              <a-tab-pane
                v-for="tab in resourceTabs"
                :key="tab.key"
                :title="tab.title"
              />
            </a-tabs>
            <div class="resource-toolbar">
              <a-input
                v-model="resourceState.keyword"
                :placeholder="
                  t('operatorStation.create.resource.searchPlaceholder')
                "
                allow-clear
                class="resource-search"
              >
                <template #prefix><icon-search /></template>
              </a-input>
              <a-radio-group
                v-model="resourceState.filter"
                type="button"
                size="small"
              >
                <a-radio value="all">
                  {{ t('operatorStation.create.resource.filter.all') }}
                </a-radio>
                <a-radio value="assigned">
                  {{ t('operatorStation.create.resource.filter.assigned') }}
                </a-radio>
                <a-radio value="unassigned">
                  {{ t('operatorStation.create.resource.filter.unassigned') }}
                </a-radio>
              </a-radio-group>
            </div>
            <a-table
              row-key="code"
              :loading="resourcesLoading"
              :columns="resourceColumns"
              :data="activeResourceList"
              :pagination="{ pageSize: 10, showTotal: true }"
              :bordered="true"
              size="small"
              :row-selection="{
                type: 'checkbox',
                showCheckedAll: true,
                onlyCurrent: false,
              }"
              :selected-keys="activeResourceSelectedKeys"
              @selection-change="onResourceSelectionChange"
            >
              <template #status="{ record }">
                <a-tag :color="record.enabled ? 'green' : 'gray'">
                  {{
                    record.enabled
                      ? t('front.common.status.enabled')
                      : t('front.common.status.disabled')
                  }}
                </a-tag>
              </template>
              <template #assigned="{ record }">
                {{
                  record.assigned ? t('front.common.yes') : t('front.common.no')
                }}
              </template>
              <template #assignedAt="{ record }">
                {{ record.assignedAt || '--' }}
              </template>
            </a-table>
          </div>
          <div class="step-side">
            <div class="side-panel-title">
              {{ t('operatorStation.create.resource.summaryTitle') }}
            </div>
            <div
              v-for="tab in resourceTabs"
              :key="tab.key"
              class="resource-summary-item"
            >
              <span>{{ tab.title }}</span>
              <a-tag color="arcoblue">
                {{ resourceState.selected[tab.key].length }}
                {{ t('operatorStation.create.resource.unit') }}
              </a-tag>
            </div>
          </div>
        </div>
      </a-spin>

      <div class="wizard-footer">
        <a-space>
          <a-button @click="handleCancel">
            {{ t('operatorStation.create.action.cancel') }}
          </a-button>
          <a-button
            :loading="savingDraft"
            :disabled="submitting"
            @click="handleSaveDraft"
          >
            {{
              isCompletedEdit
                ? t('operatorStation.create.action.save')
                : t('operatorStation.create.action.saveDraft')
            }}
          </a-button>
        </a-space>
        <a-space>
          <a-button v-if="currentStep > 1" @click="handlePrev">
            {{ t('operatorStation.create.action.prev') }}
          </a-button>
          <a-button
            v-if="currentStep < totalSteps"
            type="primary"
            :loading="creatingDraft"
            @click="handleNext"
          >
            {{ t('operatorStation.create.action.next') }}
          </a-button>
          <a-button
            v-else
            type="primary"
            :loading="submitting"
            :disabled="savingDraft"
            @click="handleSubmit"
          >
            {{
              isCompletedEdit
                ? t('operatorStation.create.action.save')
                : t('operatorStation.create.action.finish')
            }}
          </a-button>
        </a-space>
      </div>
    </div>

    <!-- 域名新增/编辑弹窗 -->
    <a-modal
      v-model:visible="domainModalVisible"
      :title="
        editingDomainId
          ? t('operatorStation.create.domain.modal.editTitle')
          : t('operatorStation.create.domain.modal.addTitle')
      "
      :width="480"
      :ok-loading="domainSubmitting"
      unmount-on-close
      @before-ok="handleDomainOk"
      @cancel="domainModalVisible = false"
    >
      <a-form ref="domainFormRef" :model="domainForm" :rules="domainRules">
        <a-form-item
          :label="t('operatorStation.create.domain.field.type')"
          field="type"
          required
        >
          <a-select v-model="domainForm.type" :options="domainTypeOptions" />
        </a-form-item>
        <a-form-item
          :label="t('operatorStation.create.domain.field.domain')"
          field="domain"
          required
        >
          <a-input
            v-model="domainForm.domain"
            :placeholder="
              t('operatorStation.create.domain.field.domainPlaceholder')
            "
            allow-clear
            :max-length="253"
          />
        </a-form-item>
        <a-form-item
          :label="t('operatorStation.create.domain.field.status')"
          field="status"
        >
          <a-radio-group v-model="domainForm.status">
            <a-radio :value="1">{{ t('front.common.status.enabled') }}</a-radio>
            <a-radio :value="2">
              {{ t('front.common.status.disabled') }}
            </a-radio>
          </a-radio-group>
        </a-form-item>
        <a-form-item
          :label="t('operatorStation.create.domain.field.remark')"
          field="remark"
        >
          <a-input
            v-model="domainForm.remark"
            :placeholder="
              t('operatorStation.create.domain.field.remarkPlaceholder')
            "
            allow-clear
            :max-length="1000"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, Modal } from '@arco-design/web-vue';
  import type { FieldRule, FormInstance } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import dayjs from 'dayjs';
  import { queryTimezoneListAll, TIMEZONE_STATUS } from '@/api/timezone';
  import { queryCurrencyListAll, CURRENCY_STATUS } from '@/api/currency';
  import { queryRegionListAll, REGION_STATUS } from '@/api/region';
  import { queryEnabledLangs } from '@/api/i18n';
  import {
    completeOperator,
    createOperator,
    createOperatorAdmin,
    updateOperatorAdmin,
    createOperatorDomain,
    createOperatorProfile,
    deleteOperatorDomain,
    getOperatorProfile,
    queryOperatorAdminList,
    queryOperatorAgentLineAllocation,
    queryOperatorDomainList,
    queryOperatorLanguageAllocation,
    queryOperatorRegionAllocation,
    saveOperatorAgentLineAllocation,
    saveOperatorLanguageAllocation,
    saveOperatorRegionAllocation,
    updateOperator,
    updateOperatorDomain,
    updateOperatorProfile,
    isOperatorBaseCurrencyLocked,
    OPERATOR_ADMIN_STATUS,
    OPERATOR_CREATION_STATUS,
    OPERATOR_DOMAIN_STATUS,
    OPERATOR_DOMAIN_TYPE,
    OPERATOR_STATUS,
    type CreateOperatorParams,
    type OperatorRecord,
    type UpdateOperatorParams,
  } from '@/api/operator';

  const { t } = useI18n();

  const props = defineProps<{
    visible: boolean;
    /** 从列表续编/编辑时传入；有值则进入已有分站的创建向导 */
    resumeRecord?: OperatorRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  /** 列表「编辑」已完成分站：走同一向导，文案与提交逻辑按编辑态处理 */
  const isCompletedEdit = computed(
    () =>
      props.resumeRecord?.creation_status === OPERATOR_CREATION_STATUS.COMPLETED
  );

  /** 发布中或已发布过的分站，时区与结算币种不允许再改 */
  const baseLocked = computed(() =>
    props.resumeRecord
      ? isOperatorBaseCurrencyLocked(props.resumeRecord)
      : false
  );

  /**
   * 5 步向导与接口对照：
   * - Step1 基础信息：create / update
   * - Step2 档案信息：profile get / create / update
   * - Step3 管理员账号：admin list / create / update
   * - Step4 域名配置：domain list / create / update / delete
   * - Step5 基础资源：language|region|agent-line allocation list / save
   * - 「完成创建」：complete
   */
  const totalSteps = 5;
  const currentStep = ref(1);
  const optionsLoading = ref(false);
  const creatingDraft = ref(false);
  const savingDraft = ref(false);
  const submitting = ref(false);
  /** 已在后端创建成功的分站 ID，创建前为 null */
  const operatorId = ref<number | null>(null);
  /** 已创建的档案 ID；有值则后续走 profile/update */
  const profileId = ref<number | null>(null);
  /** 已创建或已知的管理员 ID；有值则后续走 admin/update */
  const adminId = ref<number | null>(null);
  const resourcesLoading = ref(false);
  const adminLoading = ref(false);

  const formatTime = (value?: number | null) => {
    if (!value) return '--';
    const ms = value < 1e12 ? value * 1000 : value;
    return dayjs(ms).format('YYYY-MM-DD HH:mm:ss');
  };

  const baseInfoFormRef = ref<FormInstance>();
  const profileFormRef = ref<FormInstance>();
  const adminFormRef = ref<FormInstance>();
  const domainFormRef = ref<FormInstance>();

  const stepNames = computed(() => [
    t('operatorStation.create.step.baseInfo'),
    t('operatorStation.create.step.profileInfo'),
    t('operatorStation.create.step.adminAccount'),
    t('operatorStation.create.step.domain'),
    t('operatorStation.create.step.resource'),
  ]);

  const timezoneOptions = ref<{ label: string; value: string }[]>([]);
  const currencyOptions = ref<{ label: string; value: string }[]>([]);
  const statusOptions = computed(() => [
    {
      label: t('operatorStation.status.normal'),
      value: OPERATOR_STATUS.NORMAL,
    },
    {
      label: t('operatorStation.status.paused'),
      value: OPERATOR_STATUS.PAUSED,
    },
    {
      label: t('operatorStation.status.closed'),
      value: OPERATOR_STATUS.CLOSED,
    },
  ]);

  // ---------- Step 1 基础信息（真实接口） ----------
  const getDefaultForm = (): CreateOperatorParams & { code: string } => ({
    name: '',
    code: '',
    timezone_code: '',
    settlement_currency_code: '',
    status: OPERATOR_STATUS.NORMAL,
    remark: '',
  });

  const form = reactive(getDefaultForm());

  const baseInfoRules = computed<Record<string, FieldRule | FieldRule[]>>(
    () => ({
      name: [
        {
          required: true,
          message: t('operatorStation.create.rule.nameRequired'),
        },
      ],
      timezone_code: [
        {
          required: true,
          message: t('operatorStation.create.rule.timezoneRequired'),
        },
      ],
      settlement_currency_code: [
        {
          required: true,
          message: t('operatorStation.create.rule.currencyRequired'),
        },
      ],
      status: [
        {
          required: true,
          message: t('operatorStation.create.rule.statusRequired'),
        },
      ],
    })
  );

  const timezoneLabel = computed(
    () =>
      timezoneOptions.value.find((item) => item.value === form.timezone_code)
        ?.label ||
      form.timezone_code ||
      '--'
  );
  const currencyLabel = computed(
    () =>
      currencyOptions.value.find(
        (item) => item.value === form.settlement_currency_code
      )?.label ||
      form.settlement_currency_code ||
      '--'
  );

  const fetchOptions = async () => {
    optionsLoading.value = true;
    try {
      const [timezoneRes, currencyRes] = await Promise.all([
        queryTimezoneListAll({ status: TIMEZONE_STATUS.ENABLED }),
        queryCurrencyListAll({ status: CURRENCY_STATUS.ENABLED }),
      ]);
      timezoneOptions.value = (timezoneRes.data?.list || []).map((item) => ({
        label: item.name ? `${item.name}（${item.code}）` : item.code,
        value: item.code,
      }));
      currencyOptions.value = (currencyRes.data?.list || []).map((item) => ({
        label: item.name ? `${item.name}（${item.code}）` : item.code,
        value: item.code,
      }));
    } catch {
      timezoneOptions.value = [];
      currencyOptions.value = [];
    } finally {
      optionsLoading.value = false;
    }
  };

  const buildCreateParams = (): CreateOperatorParams => {
    const params: CreateOperatorParams = {
      name: form.name.trim(),
      timezone_code: form.timezone_code,
      settlement_currency_code: form.settlement_currency_code,
      status: form.status ?? OPERATOR_STATUS.NORMAL,
    };
    const remark = form.remark?.trim();
    if (remark) params.remark = remark;
    return params;
  };

  /** 首次离开 Step1 时创建草稿；之后再改动 Step1 字段则走 update */
  const persistBaseInfo = async () => {
    if (!operatorId.value) {
      const { data } = await createOperator(buildCreateParams());
      if (data?.id) {
        operatorId.value = data.id;
        // 编码由后端生成，覆盖用户手动填写的本地占位编码
        if (data.code) form.code = data.code;
      }
      return;
    }
    const params: UpdateOperatorParams = {
      id: operatorId.value,
      name: form.name.trim(),
      status: form.status ?? OPERATOR_STATUS.NORMAL,
      remark: form.remark?.trim() || '',
    };
    if (!baseLocked.value) {
      params.timezone_code = form.timezone_code;
      params.settlement_currency_code = form.settlement_currency_code;
    }
    await updateOperator(params);
  };

  // ---------- Step 2 档案信息（profile/create） ----------
  const getDefaultProfileForm = () => ({
    companyName: '',
    contactName: '',
    contactEmail: '',
    remark: '',
  });
  const profileForm = reactive(getDefaultProfileForm());
  /** 接口字段均为选填；邮箱若填写则校验格式 */
  const profileRules = computed<Record<string, FieldRule | FieldRule[]>>(
    () => ({
      contactEmail: [
        {
          validator: (value, callback) => {
            const email = String(value || '').trim();
            if (!email) {
              callback();
              return;
            }
            if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
              callback(t('operatorStation.create.rule.contactEmailInvalid'));
              return;
            }
            callback();
          },
        },
      ],
    })
  );

  const persistProfile = async () => {
    if (!operatorId.value) {
      throw new Error('operator id missing');
    }

    const companyName = profileForm.companyName.trim();
    const contactName = profileForm.contactName.trim();
    const contactEmail = profileForm.contactEmail.trim();
    const remark = profileForm.remark.trim();
    const hasContent = Boolean(
      companyName || contactName || contactEmail || remark
    );

    if (profileId.value) {
      // update 除 operator_id 外至少要有一个字段
      if (!hasContent) return;
      await updateOperatorProfile({
        operator_id: operatorId.value,
        company_name: companyName,
        contact_name: contactName,
        contact_email: contactEmail,
        remark,
      });
      return;
    }

    // 全部为空时不创建档案，允许直接跳过
    if (!hasContent) return;

    const { data } = await createOperatorProfile({
      operator_id: operatorId.value,
      ...(companyName ? { company_name: companyName } : {}),
      ...(contactName ? { contact_name: contactName } : {}),
      ...(contactEmail ? { contact_email: contactEmail } : {}),
      ...(remark ? { remark } : {}),
    });
    if (data?.id) profileId.value = data.id;
  };

  const loadProfile = async () => {
    if (!operatorId.value) return;
    try {
      const { data } = await getOperatorProfile({
        operator_id: operatorId.value,
      });
      const profile = data?.profile;
      if (!profile) {
        profileId.value = null;
        return;
      }
      profileId.value = profile.id;
      Object.assign(profileForm, {
        companyName: profile.company_name || '',
        contactName: profile.contact_name || '',
        contactEmail: profile.contact_email || '',
        remark: profile.remark || '',
      });
    } catch {
      // 全局拦截器已提示
    }
  };

  // ---------- Step 3 管理员账号（admin create / update） ----------
  const getDefaultAdminForm = () => ({
    account: '',
    password: '',
    confirmPassword: '',
    displayName: '',
    status: OPERATOR_ADMIN_STATUS.ENABLED as number,
  });
  const adminForm = reactive(getDefaultAdminForm());
  const adminStatusOptions = computed(() => [
    {
      label: t('front.common.status.enabled'),
      value: OPERATOR_ADMIN_STATUS.ENABLED,
    },
    {
      label: t('front.common.status.disabled'),
      value: OPERATOR_ADMIN_STATUS.DISABLED,
    },
  ]);
  const adminRules = computed<Record<string, FieldRule | FieldRule[]>>(() => ({
    account: [
      {
        required: true,
        message: t('operatorStation.create.rule.accountRequired'),
      },
    ],
    password: [
      ...(adminId.value
        ? []
        : [
            {
              required: true,
              message: t('operatorStation.create.rule.passwordRequired'),
            },
          ]),
      {
        validator: (value, callback) => {
          const password = String(value || '');
          if (!password) {
            callback();
            return;
          }
          if (password.length < 6 || password.length > 32) {
            callback(t('operatorStation.create.rule.passwordLength'));
            return;
          }
          callback();
        },
      },
    ],
    confirmPassword: [
      ...(adminId.value && !adminForm.password
        ? []
        : [
            {
              required: true,
              message: t('operatorStation.create.rule.confirmPasswordRequired'),
            },
          ]),
      {
        validator: (value, callback) => {
          if (!adminForm.password) {
            callback();
            return;
          }
          if (value && value !== adminForm.password) {
            callback(t('operatorStation.create.rule.passwordNotMatch'));
            return;
          }
          callback();
        },
      },
    ],
    displayName: [
      {
        required: true,
        message: t('operatorStation.create.rule.displayNameRequired'),
      },
    ],
  }));

  const loadAdmin = async () => {
    if (!operatorId.value) return;
    adminLoading.value = true;
    try {
      const { data } = await queryOperatorAdminList({
        page: 1,
        page_size: 1,
        operator_id: operatorId.value,
      });
      const admin = data?.list?.[0];
      if (!admin) {
        adminId.value = null;
        Object.assign(adminForm, getDefaultAdminForm());
        return;
      }
      adminId.value = admin.id;
      Object.assign(adminForm, {
        account: admin.username || '',
        password: '',
        confirmPassword: '',
        displayName: admin.display_name || '',
        status: admin.status ?? OPERATOR_ADMIN_STATUS.ENABLED,
      });
    } catch {
      // 全局拦截器已提示
    } finally {
      adminLoading.value = false;
    }
  };

  const persistAdmin = async () => {
    if (!operatorId.value) {
      throw new Error('operator id missing');
    }

    const username = adminForm.account.trim();
    const displayName = adminForm.displayName.trim();
    const status = adminForm.status ?? OPERATOR_ADMIN_STATUS.ENABLED;
    const { password } = adminForm;

    if (adminId.value) {
      await updateOperatorAdmin({
        id: adminId.value,
        username,
        display_name: displayName,
        status,
        ...(password ? { password } : {}),
      });
      return;
    }

    const { data } = await createOperatorAdmin({
      operator_id: operatorId.value,
      username,
      password,
      display_name: displayName,
      status,
    });
    if (data?.id) adminId.value = data.id;
  };

  // ---------- Step 4 域名配置（list / create / update / delete） ----------
  interface DomainRow {
    id: number;
    type: number;
    domain: string;
    status: number;
    remark: string;
  }
  const domainList = ref<DomainRow[]>([]);
  const domainLoading = ref(false);
  const domainSubmitting = ref(false);
  const domainModalVisible = ref(false);
  const editingDomainId = ref<number | null>(null);
  const getDefaultDomainForm = () => ({
    type: OPERATOR_DOMAIN_TYPE.ADMIN as number,
    domain: '',
    status: OPERATOR_DOMAIN_STATUS.ENABLED as number,
    remark: '',
  });
  const domainForm = reactive(getDefaultDomainForm());
  const domainTypeOptions = computed(() => [
    {
      label: t('operatorStation.create.domain.type.admin'),
      value: OPERATOR_DOMAIN_TYPE.ADMIN,
    },
    {
      label: t('operatorStation.create.domain.type.agent'),
      value: OPERATOR_DOMAIN_TYPE.AGENT,
    },
    {
      label: t('operatorStation.create.domain.type.h5'),
      value: OPERATOR_DOMAIN_TYPE.H5,
    },
  ]);
  const domainRules = computed<Record<string, FieldRule | FieldRule[]>>(() => ({
    type: [
      {
        required: true,
        message: t('operatorStation.create.domain.rule.typeRequired'),
      },
    ],
    domain: [
      {
        required: true,
        message: t('operatorStation.create.domain.rule.domainRequired'),
      },
    ],
  }));
  const domainTypeLabel = (type: number) =>
    domainTypeOptions.value.find((item) => item.value === type)?.label ||
    String(type);
  const domainColumns = computed<TableColumnData[]>(() => [
    {
      title: t('operatorStation.create.domain.column.type'),
      dataIndex: 'type',
      render: ({ record }) => domainTypeLabel((record as DomainRow).type),
      width: 130,
      align: 'center',
    },
    {
      title: t('operatorStation.create.domain.column.domain'),
      dataIndex: 'domain',
      align: 'center',
    },
    {
      title: t('operatorStation.create.domain.column.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 100,
      align: 'center',
    },
    {
      title: t('operatorStation.create.domain.column.remark'),
      dataIndex: 'remark',
      align: 'center',
      render: ({ record }) => (record as DomainRow).remark || '--',
    },
    {
      title: t('operatorStation.create.domain.column.operations'),
      dataIndex: 'operations',
      slotName: 'operations',
      width: 120,
      align: 'center',
    },
  ]);

  const loadDomains = async () => {
    if (!operatorId.value) return;
    domainLoading.value = true;
    try {
      const { data } = await queryOperatorDomainList({
        page: 1,
        page_size: 100,
        operator_id: operatorId.value,
      });
      domainList.value = (data?.list || []).map((item) => ({
        id: item.id,
        type: item.domain_type,
        domain: item.domain_name,
        status: item.status,
        remark: item.remark || '',
      }));
    } catch {
      domainList.value = [];
    } finally {
      domainLoading.value = false;
    }
  };

  const openDomainModal = (record?: DomainRow) => {
    if (record) {
      editingDomainId.value = record.id;
      Object.assign(domainForm, {
        type: record.type,
        domain: record.domain,
        status: record.status,
        remark: record.remark,
      });
    } else {
      editingDomainId.value = null;
      Object.assign(domainForm, getDefaultDomainForm());
    }
    domainModalVisible.value = true;
  };

  const handleDomainOk = async (done: (closed: boolean) => void) => {
    const errors = await domainFormRef.value?.validate();
    if (errors || domainSubmitting.value) {
      done(false);
      return;
    }
    if (!operatorId.value) {
      Message.warning(t('operatorStation.create.message.needBaseInfoFirst'));
      done(false);
      return;
    }

    const domainName = domainForm.domain.trim();
    const remark = domainForm.remark.trim();

    domainSubmitting.value = true;
    try {
      if (editingDomainId.value) {
        await updateOperatorDomain({
          id: editingDomainId.value,
          domain_name: domainName,
          domain_type: domainForm.type,
          status: domainForm.status,
          remark,
        });
      } else {
        await createOperatorDomain({
          operator_id: operatorId.value,
          domain_name: domainName,
          domain_type: domainForm.type,
          status: domainForm.status,
          ...(remark ? { remark } : {}),
        });
      }
      await loadDomains();
      domainModalVisible.value = false;
      done(true);
    } catch {
      done(false);
    } finally {
      domainSubmitting.value = false;
    }
  };

  const handleDeleteDomain = (record: DomainRow) => {
    Modal.confirm({
      title: t('front.common.delete.title'),
      content: t('operatorStation.create.domain.delete.content', {
        domain: record.domain,
      }),
      okText: t('front.common.delete.okText'),
      okButtonProps: { status: 'danger' },
      async onOk() {
        await deleteOperatorDomain({ id: record.id });
        Message.success(t('front.common.message.deleteSuccess'));
        await loadDomains();
      },
    });
  };

  // ---------- Step 5 基础资源（allocation list / save） ----------
  interface ResourceItem {
    /** 业务编码，用作表格 row-key */
    code: string;
    name: string;
    enabled: boolean;
    assigned: boolean;
    assignedAt: string | null;
  }
  type ResourceTabKey = 'language' | 'region' | 'proxyLine';

  const resourceTabs = computed<{ key: ResourceTabKey; title: string }[]>(
    () => [
      {
        key: 'language',
        title: t('operatorStation.create.resource.tab.language'),
      },
      {
        key: 'region',
        title: t('operatorStation.create.resource.tab.region'),
      },
      {
        key: 'proxyLine',
        title: t('operatorStation.create.resource.tab.proxyLine'),
      },
    ]
  );

  const emptyResourceItems = (): Record<ResourceTabKey, ResourceItem[]> => ({
    language: [],
    region: [],
    proxyLine: [],
  });

  const resourceState = reactive<{
    activeTab: ResourceTabKey;
    keyword: string;
    filter: 'all' | 'assigned' | 'unassigned';
    items: Record<ResourceTabKey, ResourceItem[]>;
    selected: Record<ResourceTabKey, string[]>;
  }>({
    activeTab: 'language',
    keyword: '',
    filter: 'all',
    items: emptyResourceItems(),
    selected: { language: [], region: [], proxyLine: [] },
  });

  const activeResourceSelectedKeys = computed(
    () => resourceState.selected[resourceState.activeTab]
  );

  const activeResourceList = computed(() => {
    const keyword = resourceState.keyword.trim().toLowerCase();
    const selectedKeys = resourceState.selected[resourceState.activeTab];
    return resourceState.items[resourceState.activeTab]
      .filter(
        (item) =>
          !keyword ||
          item.name.toLowerCase().includes(keyword) ||
          item.code.toLowerCase().includes(keyword)
      )
      .filter((item) => {
        const isAssigned = selectedKeys.includes(item.code);
        if (resourceState.filter === 'assigned') return isAssigned;
        if (resourceState.filter === 'unassigned') return !isAssigned;
        return true;
      });
  });

  const resourceColumns = computed<TableColumnData[]>(() => [
    {
      title: t('operatorStation.create.resource.column.name'),
      dataIndex: 'name',
      align: 'center',
    },
    {
      title: t('operatorStation.create.resource.column.code'),
      dataIndex: 'code',
      align: 'center',
    },
    {
      title: t('operatorStation.create.resource.column.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 100,
      align: 'center',
    },
    {
      title: t('operatorStation.create.resource.column.assigned'),
      dataIndex: 'assigned',
      slotName: 'assigned',
      width: 90,
      align: 'center',
    },
    {
      title: t('operatorStation.create.resource.column.assignedAt'),
      dataIndex: 'assignedAt',
      slotName: 'assignedAt',
      width: 170,
      align: 'center',
    },
  ]);

  const onResourceSelectionChange = (keys: (string | number)[]) => {
    const codes = keys.map((key) => String(key));
    const tab = resourceState.activeTab;
    const nowStr = dayjs().format('YYYY-MM-DD HH:mm:ss');
    resourceState.items[tab].forEach((item) => {
      const nowAssigned = codes.includes(item.code);
      if (nowAssigned && !item.assigned) item.assignedAt = nowStr;
      if (!nowAssigned) item.assignedAt = null;
      item.assigned = nowAssigned;
    });
    resourceState.selected[tab] = codes;
  };

  const loadResources = async () => {
    if (!operatorId.value) return;
    resourcesLoading.value = true;
    try {
      const operatorIdVal = operatorId.value;
      const [langRes, langAllocRes, regionRes, regionAllocRes, agentRes] =
        await Promise.all([
          queryEnabledLangs(),
          queryOperatorLanguageAllocation({ operator_id: operatorIdVal }),
          queryRegionListAll({ status: REGION_STATUS.ENABLED }),
          queryOperatorRegionAllocation({ operator_id: operatorIdVal }),
          queryOperatorAgentLineAllocation({ operator_id: operatorIdVal }),
        ]);

      const langAllocated = new Map(
        (langAllocRes.data?.list || []).map((item) => [
          item.language_code,
          item.allocated_at,
        ])
      );
      const languageItems: ResourceItem[] = (langRes.data?.list || []).map(
        (item) => {
          const allocatedAt = langAllocated.get(item.lang);
          return {
            code: item.lang,
            name: item.i18n_name || item.name || item.lang,
            enabled: !item.disabled,
            assigned: allocatedAt !== undefined,
            assignedAt:
              allocatedAt !== undefined ? formatTime(allocatedAt) : null,
          };
        }
      );

      const regionAllocated = new Map(
        (regionAllocRes.data?.list || []).map((item) => [
          item.region_code.toUpperCase(),
          item.allocated_at,
        ])
      );
      const regionItems: ResourceItem[] = (regionRes.data?.list || []).map(
        (item) => {
          const code = (item.code || '').toUpperCase();
          const allocatedAt = regionAllocated.get(code);
          return {
            code,
            name: item.name || code,
            enabled: item.status === REGION_STATUS.ENABLED,
            assigned: allocatedAt !== undefined,
            assignedAt:
              allocatedAt !== undefined ? formatTime(allocatedAt) : null,
          };
        }
      );

      const agentItems: ResourceItem[] = (agentRes.data?.list || []).map(
        (item) => ({
          code: item.agent_line_code,
          name: item.agent_line_code,
          enabled: true,
          assigned: Boolean(item.allocated),
          assignedAt: item.allocated ? formatTime(item.allocated_at) : null,
        })
      );

      resourceState.items = {
        language: languageItems,
        region: regionItems,
        proxyLine: agentItems,
      };
      resourceState.selected = {
        language: languageItems
          .filter((item) => item.assigned)
          .map((item) => item.code),
        region: regionItems
          .filter((item) => item.assigned)
          .map((item) => item.code),
        proxyLine: agentItems
          .filter((item) => item.assigned)
          .map((item) => item.code),
      };
    } catch {
      resourceState.items = emptyResourceItems();
      resourceState.selected = { language: [], region: [], proxyLine: [] };
    } finally {
      resourcesLoading.value = false;
    }
  };

  /** 三类资源均全量覆盖保存 */
  const persistResources = async () => {
    if (!operatorId.value) return;
    await Promise.all([
      saveOperatorLanguageAllocation({
        operator_id: operatorId.value,
        language_codes: [...resourceState.selected.language],
      }),
      saveOperatorRegionAllocation({
        operator_id: operatorId.value,
        region_codes: [...resourceState.selected.region],
      }),
      saveOperatorAgentLineAllocation({
        operator_id: operatorId.value,
        agent_line_codes: [...resourceState.selected.proxyLine],
      }),
    ]);
  };

  // ---------- 向导整体流程 ----------
  const resetWizard = () => {
    currentStep.value = 1;
    operatorId.value = null;
    profileId.value = null;
    adminId.value = null;
    Object.assign(form, getDefaultForm());
    Object.assign(profileForm, getDefaultProfileForm());
    Object.assign(adminForm, getDefaultAdminForm());
    domainList.value = [];
    Object.assign(resourceState, {
      activeTab: 'language',
      keyword: '',
      filter: 'all',
      items: emptyResourceItems(),
      selected: { language: [], region: [], proxyLine: [] },
    });
    baseInfoFormRef.value?.clearValidate();
    profileFormRef.value?.clearValidate();
    adminFormRef.value?.clearValidate();
  };

  /** 用已有分站回填 Step1；草稿定位到 Step2 续编，已完成从 Step1 起可改全流程 */
  const hydrateFromResume = async (record: OperatorRecord) => {
    operatorId.value = record.id;
    Object.assign(form, {
      name: record.name || '',
      code: record.code || '',
      timezone_code: record.timezone_code || '',
      settlement_currency_code: record.settlement_currency_code || '',
      status: record.status ?? OPERATOR_STATUS.NORMAL,
      remark: record.remark || '',
    });
    currentStep.value =
      record.creation_status === OPERATOR_CREATION_STATUS.COMPLETED ? 1 : 2;
    await Promise.all([loadProfile(), loadDomains()]);
  };

  const close = () => {
    emit('update:visible', false);
  };

  const handleCancel = () => {
    if (operatorId.value) {
      Modal.confirm({
        title: t('operatorStation.create.cancelConfirm.title'),
        content: t('operatorStation.create.cancelConfirm.content', {
          name: form.name || '--',
        }),
        onOk() {
          close();
          resetWizard();
        },
      });
      return;
    }
    close();
    resetWizard();
  };

  const validateStepForms: Record<number, () => Promise<boolean>> = {
    1: async () => !(await baseInfoFormRef.value?.validate()),
    2: async () => !(await profileFormRef.value?.validate()),
    3: async () => !(await adminFormRef.value?.validate()),
  };

  const handleNext = async () => {
    const validator = validateStepForms[currentStep.value];
    if (validator && !(await validator())) return;

    if (currentStep.value === 1) {
      creatingDraft.value = true;
      try {
        await persistBaseInfo();
      } catch {
        creatingDraft.value = false;
        return;
      }
      creatingDraft.value = false;
    }

    if (currentStep.value === 2) {
      creatingDraft.value = true;
      try {
        await persistProfile();
      } catch {
        creatingDraft.value = false;
        return;
      }
      creatingDraft.value = false;
    }

    if (currentStep.value === 3) {
      creatingDraft.value = true;
      try {
        await persistAdmin();
      } catch {
        creatingDraft.value = false;
        return;
      }
      creatingDraft.value = false;
    }

    if (currentStep.value < totalSteps) {
      currentStep.value += 1;
      if (currentStep.value === 2) await loadProfile();
      if (currentStep.value === 3) await loadAdmin();
      if (currentStep.value === 4) await loadDomains();
      if (currentStep.value === 5) await loadResources();
    }
  };

  const handlePrev = async () => {
    if (currentStep.value <= 1) return;
    currentStep.value -= 1;
    if (currentStep.value === 2) await loadProfile();
    if (currentStep.value === 3) await loadAdmin();
  };

  const handleSaveDraft = async () => {
    if (savingDraft.value || submitting.value) return;
    if (currentStep.value === 1) {
      const errors = await baseInfoFormRef.value?.validate();
      if (errors) return;
    }
    if (currentStep.value === 2) {
      const errors = await profileFormRef.value?.validate();
      if (errors) return;
    }
    if (currentStep.value === 3) {
      const errors = await adminFormRef.value?.validate();
      if (errors) return;
    }
    savingDraft.value = true;
    try {
      await persistBaseInfo();
      if (currentStep.value >= 2) {
        await persistProfile();
      }
      if (currentStep.value >= 3) {
        await persistAdmin();
      }
      if (currentStep.value >= 5) {
        await persistResources();
      }
      Message.success(
        isCompletedEdit.value
          ? t('front.common.message.saveSuccess')
          : t('operatorStation.create.message.saveDraftSuccess')
      );
      emit('success');
    } finally {
      savingDraft.value = false;
    }
  };

  const handleSubmit = async () => {
    if (savingDraft.value || submitting.value) return;
    submitting.value = true;
    try {
      if (!operatorId.value) {
        const errors = await baseInfoFormRef.value?.validate();
        if (errors) return;
        await persistBaseInfo();
      }
      if (operatorId.value) {
        try {
          await persistResources();
        } catch {
          return;
        }
        if (isCompletedEdit.value) {
          Message.success(t('front.common.message.saveSuccess'));
        } else {
          try {
            await completeOperator({ id: operatorId.value });
            Message.success(
              t('operatorStation.create.message.createSuccess', {
                name: form.name,
              })
            );
          } catch {
            Message.warning(t('operatorStation.create.message.completeFailed'));
          }
        }
      }
      emit('success');
      close();
      resetWizard();
    } finally {
      submitting.value = false;
    }
  };

  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return;
      resetWizard();
      if (props.resumeRecord) {
        hydrateFromResume(props.resumeRecord);
      }
      fetchOptions();
    }
  );
</script>

<style scoped lang="less">
  .wizard {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  .wizard-steps {
    padding: 4px 0 8px;
  }
  .step-body {
    display: flex;
    gap: 24px;
    align-items: flex-start;
  }
  .step-body--full .step-main {
    max-width: none;
  }
  .step-main {
    flex: 1;
    min-width: 0;
  }
  .step-side {
    width: 260px;
    flex-shrink: 0;
  }
  .step-title {
    font-size: 15px;
    font-weight: 600;
    color: var(--color-text-1);
  }
  .step-desc {
    margin-top: 4px;
    margin-bottom: 16px;
    color: var(--color-text-3);
    font-size: 13px;
  }
  .step-form {
    margin-top: 16px;
    max-width: 640px;
  }
  .step-form--narrow {
    max-width: 480px;
  }
  .status-select {
    max-width: 260px;
  }
  .step-alert {
    margin-top: 16px;
  }
  .current-station-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 16px;
    margin-bottom: 16px;
    background-color: var(--color-fill-2);
    border-radius: 4px;
    font-size: 13px;
    color: var(--color-text-2);
  }
  .draft-dot {
    display: inline-block;
    width: 6px;
    height: 6px;
    margin-right: 4px;
    border-radius: 50%;
    background-color: var(--color-neutral-5);

    &--done {
      background-color: rgb(var(--green-6));
    }
  }
  .side-panel-title {
    margin-bottom: 12px;
    font-size: 14px;
    font-weight: 600;
    color: var(--color-text-1);
  }
  .add-domain-btn {
    margin-bottom: 12px;
  }
  .resource-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin: 12px 0;
  }
  .resource-search {
    max-width: 280px;
  }
  .resource-summary-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 0;
    border-bottom: 1px solid var(--color-neutral-3);
    font-size: 13px;
    color: var(--color-text-2);
  }
  .resource-summary-item:last-child {
    border-bottom: none;
  }
  .wizard-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-top: 8px;
    border-top: 1px solid var(--color-neutral-3);
  }
</style>
