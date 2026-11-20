<template>
  <div class="container">
    <Breadcrumb />
    <a-card class="general-card" :bordered="false">
      <a-tabs v-model:active-key="activeTab" lazy-load size="large">
        <a-tab-pane key="material">
          <template #title>
            <span class="tab-title">
              <icon-user />
              {{ t('defaultAvatar.tab.material') }}
            </span>
          </template>
          <div class="tab-body">
            <material-panel
              type="avatar"
              @updated="frameRef?.refreshSample()"
            />
            <a-divider class="section-divider" />
            <material-panel ref="frameRef" type="frame" />
          </div>
        </a-tab-pane>
        <a-tab-pane key="vip">
          <template #title>
            <span class="tab-title">
              <icon-trophy />
              {{ t('defaultAvatar.tab.vipIcon') }}
            </span>
          </template>
          <div class="tab-body">
            <vip-icon-panel />
          </div>
        </a-tab-pane>
      </a-tabs>
    </a-card>
  </div>
</template>

<script lang="ts" setup>
  import { computed, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRoute, useRouter } from 'vue-router';
  import MaterialPanel from './components/material-panel.vue';
  import VipIconPanel from './components/vip-icon-panel.vue';

  type TabKey = 'material' | 'vip';

  const { t } = useI18n();
  const route = useRoute();
  const router = useRouter();
  const frameRef = ref<InstanceType<typeof MaterialPanel> | null>(null);

  const activeTab = computed<TabKey>({
    get: () => (route.query.tab === 'vip' ? 'vip' : 'material'),
    set: (tab) => {
      router.replace({ query: { ...route.query, tab } });
    },
  });
</script>

<style scoped lang="less">
  .container {
    padding: 0 20px 20px;
  }

  .general-card {
    margin-top: 16px;
  }

  .general-card :deep(.arco-card-body) {
    padding: 8px 24px 24px;
  }

  .tab-title {
    display: inline-flex;
    gap: 6px;
    align-items: center;
  }

  .tab-body {
    padding-top: 12px;
  }

  .section-divider {
    margin: 32px 0 28px;
  }
</style>
