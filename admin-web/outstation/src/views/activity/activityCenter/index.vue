<template>
  <div class="container list-page" :class="{ 'is-fill': fillScreen }">
    <a-card class="general-card tabs-card" :bordered="false">
      <a-tabs
        v-model:active-key="activeTab"
        class="activity-tabs"
        hide-content
        :header-padding="false"
      >
        <a-tab-pane
          v-for="tab in TABS"
          :key="tab"
          :title="t(`discountCenter.activityCenter.tabs.${tab}`)"
        >
          <template v-if="tab === 'review'" #title>
            {{ t('discountCenter.activityCenter.tabs.review') }}
            <a-badge
              class="tab-badge"
              :count="pendingReviewCount"
              :max-count="99"
            />
          </template>
        </a-tab-pane>
      </a-tabs>
    </a-card>

    <template v-if="isListTab">
      <ActivityListV2
        v-if="layout === 'v2'"
        :key="activeTab"
        class="list-fill"
        :closed-only="activeTab === 'closed'"
      />
      <ActivityListClassic
        v-else
        :key="activeTab"
        :closed-only="activeTab === 'closed'"
      />
    </template>
    <ActivityOrderList
      v-else-if="activeTab === 'review' || activeTab === 'orders'"
      :key="activeTab"
      :mode="activeTab"
      @audited="loadPendingCount"
    />
    <a-card v-else class="general-card" :bordered="false">
      <a-empty
        :description="t('discountCenter.activityCenter.tabs.developing')"
      />
    </a-card>
  </div>
</template>

<script lang="ts" setup>
  import { computed, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRoute } from 'vue-router';
  import { queryActivityAuditList } from '@/api/activity';
  import ActivityListClassic from './components/activity-list-classic.vue';
  import ActivityOrderList from './components/orders/activity-order-list.vue';
  import ActivityListV2 from './components/v2/activity-list-v2.vue';
  import { useActivityLayout } from './layout';
  import { provideActivityStore } from './store';

  const TABS = ['list', 'closed', 'review', 'orders', 'statistics'] as const;
  type ActivityTab = (typeof TABS)[number];

  const { t } = useI18n();
  const route = useRoute();
  const layout = useActivityLayout(route.query.layout);
  // 列表的类型、状态、客户端文案和时间显示的时区都来自枚举接口
  const { loadEnumParams } = provideActivityStore();
  loadEnumParams();

  /** 审核列表页签角标：待审核订单总数 */
  const pendingReviewCount = ref(0);
  async function loadPendingCount() {
    try {
      const { data } = await queryActivityAuditList(
        { page: 1, page_size: 1 },
        true
      );
      pendingReviewCount.value = data?.total || 0;
    } catch {
      pendingReviewCount.value = 0;
    }
  }
  loadPendingCount();

  const activeTab = ref<ActivityTab>('list');
  const isListTab = computed(
    () => activeTab.value === 'list' || activeTab.value === 'closed'
  );
  const fillScreen = computed(() => isListTab.value && layout.value === 'v2');
</script>

<style scoped lang="less">
  .container {
    padding: 0 20px 20px;
  }

  // 新版整页限定在一屏内，列表和详情各自滚动：
  // 导航栏 60px + 页签栏 40px + 内容区上边距 12px + 页脚 30px
  @layout-offset: 142px;
  @page-min-height: 600px;

  .container.is-fill {
    display: flex;
    flex-direction: column;
    box-sizing: border-box;
    height: ~'max(@{page-min-height}, calc(100vh - @{layout-offset}))';
    padding-bottom: 16px;
  }

  .tabs-card {
    flex: none;
  }

  .list-fill {
    flex: 1;
    min-height: 0;
  }

  .general-card {
    margin-top: 16px;
  }

  .tabs-card :deep(.arco-card-body) {
    padding: 4px 20px 0;
  }

  .activity-tabs {
    :deep(.arco-tabs-nav-type-line .arco-tabs-tab) {
      margin: 0 48px 0 0;
      padding: 12px 0;
      font-size: 15px;
      line-height: 22px;
    }

    :deep(.arco-tabs-nav::before) {
      display: none;
    }
  }

  .tab-badge {
    margin-left: 2px;
    vertical-align: top;
    transform: translateY(-8px);
  }

  // 窄屏上下堆叠时恢复整页滚动
  @media (max-width: 1280px) {
    .container.is-fill {
      height: auto;
    }
  }
</style>
