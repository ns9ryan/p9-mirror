<template>
  <a-modal
    :visible="visible"
    :title="t('operatorStation.detail.title')"
    :footer="false"
    unmount-on-close
    :width="1220"
    :body-style="{ maxHeight: '78vh', overflow: 'auto', padding: '16px 20px' }"
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel station-detail-modal"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    @cancel="handleCancel"
  >
    <a-spin :loading="loading" style="width: 100%">
      <div v-if="detail" class="detail-page">
        <!-- 基础信息 + 档案信息 -->
        <div class="detail-row">
          <div class="detail-card">
            <div class="detail-card__title">
              <span class="section-icon section-icon--blue">
                <icon-idcard />
              </span>
              {{ t('operatorStation.detail.section.baseInfo') }}
            </div>
            <a-descriptions
              :column="2"
              size="small"
              :label-style="{ width: '110px', color: 'var(--color-text-3)' }"
            >
              <a-descriptions-item :label="t('operatorStation.column.id')">
                {{ detail.id }}
              </a-descriptions-item>
              <a-descriptions-item :label="t('operatorStation.column.code')">
                {{ detail.code || '--' }}
              </a-descriptions-item>
              <a-descriptions-item :label="t('operatorStation.column.name')">
                {{ detail.name || '--' }}
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('operatorStation.column.timezone')"
              >
                {{ timezoneLabel }}
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('operatorStation.column.currency')"
              >
                {{ currencyLabel }}
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('operatorStation.column.creationStatus')"
              >
                <span class="status-inline">
                  <span
                    class="status-dot"
                    :class="creationStatusMeta(detail.creation_status).dotClass"
                  />
                  {{ creationStatusMeta(detail.creation_status).label }}
                </span>
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('operatorStation.column.publishStatus')"
              >
                <span class="status-inline">
                  <span
                    class="status-dot"
                    :class="publishStatusMeta(detail.publish_status).dotClass"
                  />
                  {{ publishStatusMeta(detail.publish_status).label }}
                </span>
              </a-descriptions-item>
              <a-descriptions-item :label="t('operatorStation.column.status')">
                <span class="status-inline">
                  <span
                    class="status-dot"
                    :class="statusMeta(detail.status).dotClass"
                  />
                  {{ statusMeta(detail.status).label }}
                </span>
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('operatorStation.column.remark')"
                :span="2"
              >
                {{ detail.remark || '--' }}
              </a-descriptions-item>
              <a-descriptions-item :label="t('front.common.column.createdAt')">
                {{ formatTime(detail.created_at) }}
              </a-descriptions-item>
            </a-descriptions>
          </div>

          <div class="detail-card">
            <div class="detail-card__title">
              <span class="section-icon section-icon--purple">
                <icon-folder />
              </span>
              {{ t('operatorStation.detail.section.profile') }}
            </div>
            <a-empty
              v-if="!profile"
              :description="t('operatorStation.detail.profileEmpty')"
            />
            <a-descriptions
              v-else
              :column="1"
              size="small"
              :label-style="{ width: '120px', color: 'var(--color-text-3)' }"
            >
              <a-descriptions-item
                :label="t('operatorStation.create.profileInfo.companyName')"
              >
                {{ profile.company_name || '--' }}
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('operatorStation.create.profileInfo.contactName')"
              >
                {{ profile.contact_name || '--' }}
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('operatorStation.create.profileInfo.contactEmail')"
              >
                {{ profile.contact_email || '--' }}
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('operatorStation.create.profileInfo.remark')"
              >
                {{ profile.remark || '--' }}
              </a-descriptions-item>
            </a-descriptions>
          </div>
        </div>

        <!-- 管理员 + 域名 -->
        <div class="detail-row">
          <div class="detail-card">
            <div class="detail-card__title">
              <span class="section-icon section-icon--orange">
                <icon-user-group />
              </span>
              {{ t('operatorStation.detail.section.admin') }}
            </div>
            <a-table
              row-key="id"
              size="small"
              :bordered="false"
              :pagination="false"
              :columns="adminColumns"
              :data="adminList"
              :scroll="{ y: 220 }"
            >
              <template #status="{ record }">
                <span class="status-inline">
                  <span
                    class="status-dot"
                    :class="
                      record.status === OPERATOR_ADMIN_STATUS.ENABLED
                        ? 'status-dot--success'
                        : 'status-dot--muted'
                    "
                  />
                  {{
                    record.status === OPERATOR_ADMIN_STATUS.ENABLED
                      ? t('front.common.status.enabled')
                      : t('front.common.status.disabled')
                  }}
                </span>
              </template>
            </a-table>
          </div>

          <div class="detail-card">
            <div class="detail-card__title">
              <span class="section-icon section-icon--cyan">
                <icon-public />
              </span>
              {{ t('operatorStation.detail.section.domain') }}
            </div>
            <a-table
              row-key="id"
              size="small"
              :bordered="false"
              :pagination="false"
              :columns="domainColumns"
              :data="domainList"
              :scroll="{ y: 220 }"
            >
              <template #type="{ record }">
                {{ domainTypeLabel(record.domain_type) }}
              </template>
              <template #status="{ record }">
                <span class="status-inline">
                  <span
                    class="status-dot"
                    :class="
                      record.status === OPERATOR_DOMAIN_STATUS.ENABLED
                        ? 'status-dot--success'
                        : 'status-dot--muted'
                    "
                  />
                  {{
                    record.status === OPERATOR_DOMAIN_STATUS.ENABLED
                      ? t('front.common.status.enabled')
                      : t('front.common.status.disabled')
                  }}
                </span>
              </template>
            </a-table>
          </div>
        </div>

        <!-- 基础资源摘要 -->
        <div class="detail-card detail-card--full">
          <div class="detail-card__title">
            <span class="section-icon section-icon--cyan">
              <icon-storage />
            </span>
            {{ t('operatorStation.detail.section.resource') }}
            <span class="detail-card__title-extra">
              {{
                t('operatorStation.detail.resourceCategoryTotal', {
                  count: resourceSummaries.length,
                })
              }}
            </span>
          </div>
          <div class="resource-grid">
            <div
              v-for="item in resourceSummaries"
              :key="item.key"
              class="resource-block"
              :class="{ 'resource-block--clickable': item.tags.length }"
              @click="openResourceDrawer(item.key, item.tags.length)"
            >
              <div class="resource-block__head">
                <span class="resource-block__icon" :class="item.colorClass">
                  <component :is="item.icon" />
                </span>
                <div class="resource-block__meta">
                  <div class="resource-block__title">{{ item.title }}</div>
                  <div class="resource-block__count">
                    {{
                      t('operatorStation.detail.resourceAssigned', {
                        count: item.tags.length,
                      })
                    }}
                  </div>
                </div>
                <icon-right
                  v-if="item.tags.length"
                  class="resource-block__arrow"
                />
              </div>
              <div class="resource-block__tags">
                <template v-if="item.tags.length">
                  <a-tag
                    v-for="(tag, index) in item.tags.slice(
                      0,
                      RESOURCE_PREVIEW_COUNT
                    )"
                    :key="`${tag}-${index}`"
                    size="small"
                    color="arcoblue"
                    :title="tag"
                  >
                    <span class="resource-block__tag-text">{{ tag }}</span>
                  </a-tag>
                  <a-tag
                    v-if="item.tags.length > RESOURCE_PREVIEW_COUNT"
                    size="small"
                    class="resource-block__more"
                  >
                    +{{ item.tags.length - RESOURCE_PREVIEW_COUNT }}
                  </a-tag>
                </template>
                <span v-else class="resource-block__empty">--</span>
              </div>
              <div v-if="item.tags.length" class="resource-block__footer">
                <span class="resource-block__link">
                  {{ t('operatorStation.detail.viewAll') }}
                  <icon-right />
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- 游戏资源摘要 -->
        <div class="detail-card detail-card--full">
          <div class="detail-card__title">
            <span class="section-icon section-icon--green">
              <icon-apps />
            </span>
            {{ t('operatorStation.detail.section.gameResource') }}
            <span class="detail-card__title-tip">
              {{ t('operatorStation.detail.gameResource.mockTip') }}
            </span>
          </div>
          <div class="game-resource-grid">
            <div
              v-for="item in gameResourceSummaries"
              :key="item.key"
              class="game-resource-block"
            >
              <span
                class="game-resource-block__icon"
                :class="item.colorClass"
              >
                <component :is="item.icon" />
              </span>
              <div class="game-resource-block__main">
                <div class="game-resource-block__label">{{ item.title }}</div>
                <div class="game-resource-block__value">
                  {{ item.count
                  }}<span
                    v-if="item.unit"
                    class="game-resource-block__unit"
                  >
                    {{ item.unit }}</span
                  >
                </div>
              </div>
              <div v-if="item.tags.length" class="game-resource-block__tags">
                <a-tag
                  v-for="tag in item.tags.slice(0, 3)"
                  :key="tag"
                  size="small"
                >
                  {{ tag }}
                </a-tag>
                <span
                  v-if="item.tags.length > 3"
                  class="game-resource-block__more"
                  >...</span
                >
              </div>
            </div>
          </div>
        </div>

        <!-- 发布信息 -->
        <div class="detail-card detail-card--full">
          <div class="detail-card__title">
            <span class="section-icon section-icon--pink">
              <icon-send />
            </span>
            {{ t('operatorStation.detail.section.publish') }}
          </div>
          <a-descriptions
            :column="3"
            size="small"
            :label-style="{ width: '110px', color: 'var(--color-text-3)' }"
          >
            <a-descriptions-item
              :label="t('operatorStation.column.publishedAt')"
            >
              {{ formatTime(detail.published_at) }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('front.common.column.updatedAt')">
              {{ formatTime(detail.updated_at) }}
            </a-descriptions-item>
            <a-descriptions-item
              :label="t('operatorStation.detail.publishResult')"
            >
              <span class="status-inline">
                <span
                  class="status-dot"
                  :class="publishStatusMeta(detail.publish_status).dotClass"
                />
                {{ publishStatusMeta(detail.publish_status).label }}
              </span>
            </a-descriptions-item>
            <a-descriptions-item
              :label="t('operatorStation.detail.publishNote')"
              :span="3"
            >
              {{ detail.remark || '--' }}
            </a-descriptions-item>
          </a-descriptions>
        </div>
      </div>
    </a-spin>

    <a-drawer
      v-model:visible="resourceDrawerVisible"
      :width="420"
      unmount-on-close
      class="resource-drawer"
    >
      <template #title>
        <div v-if="activeResource" class="resource-drawer__title">
          <span class="resource-block__icon" :class="activeResource.colorClass">
            <component :is="activeResource.icon" />
          </span>
          {{
            t('operatorStation.detail.resourceDrawerTitle', {
              name: activeResource.title,
            })
          }}
          <a-tag size="small" color="arcoblue">
            {{ activeResource.tags.length }}
          </a-tag>
        </div>
      </template>
      <div v-if="activeResource" class="resource-drawer__grid">
        <div
          v-for="(tag, index) in activeResource.tags"
          :key="`${tag}-${index}`"
          class="resource-drawer__item"
          :title="tag"
        >
          <icon-check-circle-fill class="resource-drawer__check" />
          <span class="resource-drawer__text">{{ tag }}</span>
        </div>
      </div>
      <template #footer>
        <div v-if="activeResource" class="resource-drawer__footer">
          {{
            t('operatorStation.detail.resourceTotal', {
              count: activeResource.tags.length,
            })
          }}
        </div>
      </template>
    </a-drawer>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import dayjs from 'dayjs';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import IconIdcard from '@arco-design/web-vue/es/icon/icon-idcard';
  import IconFolder from '@arco-design/web-vue/es/icon/icon-folder';
  import IconUserGroup from '@arco-design/web-vue/es/icon/icon-user-group';
  import IconPublic from '@arco-design/web-vue/es/icon/icon-public';
  import IconStorage from '@arco-design/web-vue/es/icon/icon-storage';
  import IconApps from '@arco-design/web-vue/es/icon/icon-apps';
  import IconSend from '@arco-design/web-vue/es/icon/icon-send';
  import IconLanguage from '@arco-design/web-vue/es/icon/icon-language';
  import IconLocation from '@arco-design/web-vue/es/icon/icon-location';
  import IconLink from '@arco-design/web-vue/es/icon/icon-link';
  import IconRight from '@arco-design/web-vue/es/icon/icon-right';
  import IconCheckCircleFill from '@arco-design/web-vue/es/icon/icon-check-circle-fill';
  import { queryTimezoneListAll, TIMEZONE_STATUS } from '@/api/timezone';
  import { queryCurrencyListAll, CURRENCY_STATUS } from '@/api/currency';
  import { queryRegionListAll, REGION_STATUS } from '@/api/region';
  import { queryEnabledLangs } from '@/api/i18n';
  import {
    getOperator,
    getOperatorProfile,
    queryOperatorAdminList,
    queryOperatorAgentLineAllocation,
    queryOperatorDomainList,
    queryOperatorLanguageAllocation,
    queryOperatorRegionAllocation,
    OPERATOR_ADMIN_STATUS,
    OPERATOR_CREATION_STATUS,
    OPERATOR_DOMAIN_STATUS,
    OPERATOR_DOMAIN_TYPE,
    OPERATOR_PUBLISH_STATUS,
    OPERATOR_STATUS,
    type OperatorAdminRecord,
    type OperatorDomainRecord,
    type OperatorProfileRecord,
    type OperatorRecord,
  } from '@/api/operator';

  const { t } = useI18n();

  const props = defineProps<{
    visible: boolean;
    stationId?: number | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
  }>();

  const loading = ref(false);
  const detail = ref<OperatorRecord | null>(null);
  const profile = ref<OperatorProfileRecord | null>(null);
  const adminList = ref<OperatorAdminRecord[]>([]);
  const domainList = ref<OperatorDomainRecord[]>([]);
  const timezoneMap = ref<Record<string, string>>({});
  const currencyMap = ref<Record<string, string>>({});
  const languageTags = ref<string[]>([]);
  const regionTags = ref<string[]>([]);
  const agentLineTags = ref<string[]>([]);

  const AGENT_LINE_LABEL_KEYS: Record<string, string> = {
    CASH_PRODUCTION: 'operatorStation.detail.agentLine.CASH_PRODUCTION',
    CASH_DEMO: 'operatorStation.detail.agentLine.CASH_DEMO',
    CASH_TEST: 'operatorStation.detail.agentLine.CASH_TEST',
    CREDIT_DEMO: 'operatorStation.detail.agentLine.CREDIT_DEMO',
    CREDIT_TEST: 'operatorStation.detail.agentLine.CREDIT_TEST',
  };

  const agentLineLabel = (code: string) => {
    const key = AGENT_LINE_LABEL_KEYS[code];
    return key ? t(key) : code;
  };

  const formatTime = (value?: number | null) => {
    if (!value) return '--';
    const ms = value < 1e12 ? value * 1000 : value;
    return dayjs(ms).format('YYYY-MM-DD HH:mm:ss');
  };

  const timezoneLabel = computed(() => {
    const code = detail.value?.timezone_code;
    if (!code) return '--';
    return timezoneMap.value[code] || code;
  });

  const currencyLabel = computed(() => {
    const code = detail.value?.settlement_currency_code;
    if (!code) return '--';
    return currencyMap.value[code] || code;
  });

  const resourceSummaries = computed(() => [
    {
      key: 'language',
      icon: IconLanguage,
      colorClass: 'section-icon--blue',
      title: t('operatorStation.create.resource.tab.language'),
      tags: languageTags.value,
    },
    {
      key: 'region',
      icon: IconLocation,
      colorClass: 'section-icon--orange',
      title: t('operatorStation.create.resource.tab.region'),
      tags: regionTags.value,
    },
    {
      key: 'proxyLine',
      icon: IconLink,
      colorClass: 'section-icon--cyan',
      title: t('operatorStation.create.resource.tab.proxyLine'),
      tags: agentLineTags.value,
    },
  ]);

  const RESOURCE_PREVIEW_COUNT = 3;

  const resourceDrawerVisible = ref(false);
  const activeResourceKey = ref('');

  const activeResource = computed(() =>
    resourceSummaries.value.find((item) => item.key === activeResourceKey.value)
  );

  const openResourceDrawer = (key: string, count: number) => {
    if (!count) return;
    activeResourceKey.value = key;
    resourceDrawerVisible.value = true;
  };

  // 游戏资源摘要：当前暂无「分站-游戏」分配相关接口，先用示例数据搭建展示样式，
  // 待后端提供分站维度的游戏分类/厂商/渠道/数量统计接口后，替换为真实数据源。
  const gameResourceSummaries = computed(() => [
    {
      key: 'category',
      icon: IconStorage,
      colorClass: 'section-icon--cyan',
      title: t('operatorStation.detail.gameResource.category'),
      count: 12,
      tags: ['电子', '体育', '棋牌'],
    },
    {
      key: 'provider',
      icon: IconUserGroup,
      colorClass: 'section-icon--purple',
      title: t('operatorStation.detail.gameResource.provider'),
      count: 8,
      tags: ['PG', 'JILI', 'PP'],
    },
    {
      key: 'channel',
      icon: IconLink,
      colorClass: 'section-icon--blue',
      title: t('operatorStation.detail.gameResource.channel'),
      count: 4,
      tags: ['官网', 'H5', 'APP'],
    },
    {
      key: 'count',
      icon: IconApps,
      colorClass: 'section-icon--pink',
      title: t('operatorStation.detail.gameResource.count'),
      count: 156,
      unit: t('operatorStation.detail.gameResource.unit'),
      tags: ['热门', '新游', '经典'],
    },
  ]);

  const adminColumns = computed<TableColumnData[]>(() => [
    {
      title: t('operatorStation.create.adminAccount.account'),
      dataIndex: 'username',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('operatorStation.create.adminAccount.displayName'),
      dataIndex: 'display_name',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('operatorStation.create.adminAccount.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 90,
    },
    {
      title: t('front.common.column.updatedAt'),
      dataIndex: 'updated_at',
      width: 160,
      render: ({ record }) => formatTime(record.updated_at),
    },
  ]);

  const domainColumns = computed<TableColumnData[]>(() => [
    {
      title: t('operatorStation.create.domain.column.type'),
      dataIndex: 'domain_type',
      slotName: 'type',
      width: 110,
    },
    {
      title: t('operatorStation.create.domain.column.domain'),
      dataIndex: 'domain_name',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('operatorStation.create.domain.column.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 90,
    },
    {
      title: t('operatorStation.create.domain.column.remark'),
      dataIndex: 'remark',
      ellipsis: true,
      tooltip: true,
      render: ({ record }) => record.remark || '--',
    },
  ]);

  const creationStatusMeta = (status: number) => {
    if (status === OPERATOR_CREATION_STATUS.COMPLETED) {
      return {
        label: t('operatorStation.creationStatus.completed'),
        dotClass: 'status-dot--success',
      };
    }
    return {
      label: t('operatorStation.creationStatus.draft'),
      dotClass: 'status-dot--muted',
    };
  };

  const publishStatusMeta = (status: number) => {
    switch (status) {
      case OPERATOR_PUBLISH_STATUS.PUBLISHING:
        return {
          label: t('operatorStation.publishStatus.publishing'),
          dotClass: 'status-dot--info',
        };
      case OPERATOR_PUBLISH_STATUS.PUBLISHED:
        return {
          label: t('operatorStation.publishStatus.published'),
          dotClass: 'status-dot--success',
        };
      case OPERATOR_PUBLISH_STATUS.FAILED:
        return {
          label: t('operatorStation.publishStatus.failed'),
          dotClass: 'status-dot--danger',
        };
      default:
        return {
          label: t('operatorStation.publishStatus.unpublished'),
          dotClass: 'status-dot--muted',
        };
    }
  };

  const statusMeta = (status: number) => {
    switch (status) {
      case OPERATOR_STATUS.PAUSED:
        return {
          label: t('operatorStation.status.paused'),
          dotClass: 'status-dot--warning',
        };
      case OPERATOR_STATUS.CLOSED:
        return {
          label: t('operatorStation.status.closed'),
          dotClass: 'status-dot--danger',
        };
      default:
        return {
          label: t('operatorStation.status.normal'),
          dotClass: 'status-dot--success',
        };
    }
  };

  const domainTypeLabel = (type: number) => {
    if (type === OPERATOR_DOMAIN_TYPE.AGENT) {
      return t('operatorStation.create.domain.type.agent');
    }
    if (type === OPERATOR_DOMAIN_TYPE.H5) {
      return t('operatorStation.create.domain.type.h5');
    }
    return t('operatorStation.create.domain.type.admin');
  };

  const resetState = () => {
    detail.value = null;
    profile.value = null;
    adminList.value = [];
    domainList.value = [];
    languageTags.value = [];
    regionTags.value = [];
    agentLineTags.value = [];
    resourceDrawerVisible.value = false;
  };

  const fetchDetail = async (id: number) => {
    loading.value = true;
    resetState();
    try {
      const operatorRes = await getOperator({ id });
      detail.value = operatorRes.data ?? null;
      const operatorCode = operatorRes.data?.code || '';

      const [
        profileRes,
        adminRes,
        domainRes,
        langAllocRes,
        regionAllocRes,
        agentAllocRes,
        timezoneRes,
        currencyRes,
        langRes,
        regionRes,
      ] = await Promise.all([
        getOperatorProfile({ operator_id: id }),
        queryOperatorAdminList({ page: 1, page_size: 100, operator_id: id }),
        queryOperatorDomainList({ page: 1, page_size: 100, operator_id: id }),
        queryOperatorLanguageAllocation({ operator_id: id }),
        queryOperatorRegionAllocation({ operator_id: id }),
        queryOperatorAgentLineAllocation({ operator_id: id }),
        queryTimezoneListAll({ status: TIMEZONE_STATUS.ENABLED }),
        queryCurrencyListAll({ status: CURRENCY_STATUS.ENABLED }),
        queryEnabledLangs({ params: { operator_code: operatorCode } }),
        queryRegionListAll({ status: REGION_STATUS.ENABLED }),
      ]);
      profile.value = profileRes.data?.profile ?? null;
      adminList.value = adminRes.data?.list || [];
      domainList.value = domainRes.data?.list || [];

      const nextTimezoneMap: Record<string, string> = {};
      (timezoneRes.data?.list || []).forEach((item) => {
        nextTimezoneMap[item.code] = item.name
          ? `${item.name}（${item.code}）`
          : item.code;
      });
      timezoneMap.value = nextTimezoneMap;

      const nextCurrencyMap: Record<string, string> = {};
      (currencyRes.data?.list || []).forEach((item) => {
        nextCurrencyMap[item.code] = item.name
          ? `${item.name}（${item.code}）`
          : item.code;
      });
      currencyMap.value = nextCurrencyMap;

      const langNameMap = new Map(
        (langRes.data?.list || []).map((item) => [
          item.lang,
          item.name || item.lang,
        ])
      );
      languageTags.value = (langAllocRes.data?.list || []).map(
        (item) => langNameMap.get(item.language_code) || item.language_code
      );

      const regionNameMap = new Map(
        (regionRes.data?.list || []).map((item) => [
          (item.code || '').toUpperCase(),
          item.name || item.code,
        ])
      );
      regionTags.value = (regionAllocRes.data?.list || []).map((item) => {
        const code = (item.region_code || '').toUpperCase();
        return regionNameMap.get(code) || item.region_code;
      });

      agentLineTags.value = (agentAllocRes.data?.list || [])
        .filter((item) => item.allocated)
        .map((item) => agentLineLabel(item.agent_line_code));
    } catch {
      resetState();
    } finally {
      loading.value = false;
    }
  };

  const handleCancel = () => {
    resourceDrawerVisible.value = false;
    emit('update:visible', false);
  };

  watch(
    () => props.visible,
    (visible) => {
      if (visible && props.stationId) {
        fetchDetail(props.stationId);
      }
    }
  );
</script>

<style scoped lang="less">
  .detail-page {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .detail-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }

  .detail-card {
    background: var(--color-fill-1);
    border: 1px solid var(--color-border-2);
    border-radius: 8px;
    padding: 14px 16px;
    min-width: 0;

    &--full {
      width: 100%;
    }

    &__title {
      display: flex;
      align-items: center;
      font-size: 14px;
      font-weight: 600;
      color: var(--color-text-1);
      margin-bottom: 12px;

      &-tip {
        margin-left: 8px;
        font-size: 12px;
        font-weight: 400;
        color: var(--color-text-3);
      }
    }
  }

  .section-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    height: 20px;
    margin-right: 8px;
    border-radius: 6px;
    color: #fff;
    font-size: 12px;
    flex-shrink: 0;

    &--blue {
      background: rgb(var(--arcoblue-6));
    }
    &--purple {
      background: rgb(var(--purple-6));
    }
    &--orange {
      background: rgb(var(--orange-6));
    }
    &--cyan {
      background: rgb(var(--cyan-6));
    }
    &--green {
      background: rgb(var(--green-6));
    }
    &--pink {
      background: rgb(var(--pinkpurple-6));
    }
  }

  .status-inline {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--color-text-4);

    &--success {
      background: rgb(var(--green-6));
    }
    &--info {
      background: rgb(var(--arcoblue-6));
    }
    &--warning {
      background: rgb(var(--orange-6));
    }
    &--danger {
      background: rgb(var(--red-6));
    }
    &--muted {
      background: var(--color-text-4);
    }
  }

  .resource-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 12px;
  }

  .detail-card__title-extra {
    margin-left: auto;
    font-size: 12px;
    font-weight: 400;
    color: var(--color-text-3);
  }

  .resource-block {
    display: flex;
    flex-direction: column;
    min-width: 0;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
    padding: 12px;
    transition: border-color 0.2s, box-shadow 0.2s;

    &--clickable {
      cursor: pointer;

      &:hover {
        border-color: rgb(var(--arcoblue-4));
        box-shadow: 0 2px 8px rgba(var(--arcoblue-6), 0.12);

        .resource-block__arrow,
        .resource-block__link {
          color: rgb(var(--arcoblue-6));
        }
      }
    }

    &__head {
      display: flex;
      align-items: center;
      gap: 10px;
      margin-bottom: 10px;
    }

    &__icon {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      width: 28px;
      height: 28px;
      border-radius: 7px;
      color: #fff;
      font-size: 15px;
      flex-shrink: 0;
    }

    &__meta {
      flex: 1;
      min-width: 0;
    }

    &__title {
      font-weight: 500;
      color: var(--color-text-1);
      line-height: 1.4;
    }

    &__count {
      color: var(--color-text-3);
      font-size: 12px;
      line-height: 1.4;
    }

    &__arrow {
      color: var(--color-text-3);
      transition: color 0.2s;
    }

    &__tags {
      display: flex;
      flex-wrap: nowrap;
      align-items: center;
      gap: 6px;
      min-height: 24px;
      overflow: hidden;

      :deep(.arco-tag) {
        max-width: 110px;
        flex-shrink: 0;
      }
    }

    &__tag-text {
      min-width: 0;
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
    }

    &__more {
      color: var(--color-text-2);
    }

    &__empty {
      color: var(--color-text-4);
    }

    &__footer {
      display: flex;
      justify-content: flex-end;
      margin-top: auto;
      padding-top: 8px;
    }

    &__link {
      display: inline-flex;
      align-items: center;
      gap: 2px;
      font-size: 12px;
      color: rgb(var(--arcoblue-6));
      transition: color 0.2s;
    }
  }

  .resource-drawer {
    &__title {
      display: flex;
      align-items: center;
      gap: 8px;

      .resource-block__icon {
        width: 22px;
        height: 22px;
        font-size: 13px;
        border-radius: 6px;
      }
    }

    &__grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(110px, 1fr));
      gap: 8px;
    }

    &__item {
      display: flex;
      align-items: center;
      gap: 6px;
      min-width: 0;
      padding: 6px 10px;
      border-radius: 6px;
      background: var(--color-fill-2);
      color: var(--color-text-1);
      font-size: 13px;
    }

    &__check {
      flex-shrink: 0;
      color: rgb(var(--green-6));
    }

    &__text {
      min-width: 0;
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
    }

    &__footer {
      text-align: left;
      color: var(--color-text-3);
      font-size: 12px;
    }
  }

  .game-resource-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 12px;
  }

  .game-resource-block {
    display: flex;
    align-items: center;
    gap: 10px;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
    padding: 12px;
    min-width: 0;

    &__icon {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      width: 32px;
      height: 32px;
      border-radius: 8px;
      color: #fff;
      font-size: 16px;
      flex-shrink: 0;
    }

    &__main {
      flex-shrink: 0;
    }

    &__label {
      color: var(--color-text-3);
      font-size: 12px;
      white-space: nowrap;
    }

    &__value {
      font-size: 18px;
      font-weight: 600;
      color: var(--color-text-1);
      line-height: 1.3;
    }

    &__unit {
      font-size: 12px;
      font-weight: 400;
      color: var(--color-text-3);
      margin-left: 2px;
    }

    &__tags {
      display: flex;
      flex-wrap: wrap;
      align-items: center;
      gap: 4px;
      margin-left: auto;
      min-width: 0;
      justify-content: flex-end;
    }

    &__more {
      color: var(--color-text-3);
      font-size: 12px;
    }
  }

  @media (max-width: 900px) {
    .detail-row,
    .resource-grid {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 1100px) {
    .game-resource-grid {
      grid-template-columns: repeat(2, 1fr);
    }
  }
</style>
