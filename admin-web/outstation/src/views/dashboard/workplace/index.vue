<template>
  <div class="container">
    <div class="timezone-bar">
      <div class="timezone-title">
        <icon-clock-circle class="timezone-icon" />
        <span>{{ t('dashboardHome.timezone.label') }}</span>
        <a-tooltip :content="t('dashboardHome.timezone.tip')">
          <icon-question-circle class="timezone-tip" />
        </a-tooltip>
      </div>
      <a-radio-group
        v-model="timezoneMode"
        class="timezone-group"
        type="button"
        size="small"
      >
        <a-radio value="site">
          {{ t('dashboardHome.timezone.site') }}
          <span class="tz-code">{{ siteTimezone || '--' }}</span>
        </a-radio>
        <a-radio value="local">
          {{ t('dashboardHome.timezone.local') }}
          <span class="tz-code">{{ localTimezone }}</span>
        </a-radio>
      </a-radio-group>
    </div>
    <template v-if="ready">
      <MemberOverview :timezone="timezone" />
      <MemberTrend :timezone="timezone" />
      <MemberRetention :timezone="timezone" />
    </template>
  </div>
</template>

<script lang="ts" setup>
  import { computed, onMounted, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { ensureSiteTimezone, getSiteTimezone } from '@/utils/site-time';
  import MemberOverview from './components/member-overview.vue';
  import MemberTrend from './components/member-trend.vue';
  import MemberRetention from './components/member-retention.vue';

  type TimezoneMode = 'site' | 'local';

  const { t } = useI18n();

  /** 浏览器本机时区（IANA） */
  const localTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone;

  /** 选中的是分站时区还是本地时区，默认分站时区 */
  const timezoneMode = ref<TimezoneMode>('site');
  /** 分站时区就绪后再渲染图表，避免先按错误时区查一遍再切换 */
  const ready = ref(false);

  const siteTimezone = computed(() => getSiteTimezone());

  /** 传给各模块的查询时区（IANA） */
  const timezone = computed(() =>
    timezoneMode.value === 'local' ? localTimezone : siteTimezone.value
  );

  onMounted(async () => {
    await ensureSiteTimezone();
    ready.value = true;
  });
</script>

<script lang="ts">
  export default {
    name: 'Dashboard', // If you want the include property of keep-alive to take effect, you must name the component
  };
</script>

<style lang="less" scoped>
  .container {
    padding: 0 20px 20px;

    :deep(.general-card) {
      margin-top: 16px;
    }
  }

  .timezone-bar {
    display: flex;
    flex-wrap: wrap;
    gap: 12px 16px;
    align-items: center;
    justify-content: space-between;
    margin-top: 16px;
    padding: 12px 16px;
    background: var(--color-bg-2);
    border-radius: 4px;
  }

  .timezone-title {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    color: var(--color-text-1);
    font-weight: 500;
    font-size: 14px;
    white-space: nowrap;
  }

  .timezone-icon {
    margin-right: 6px;
    color: rgb(var(--arcoblue-6));
    font-size: 16px;
  }

  .timezone-tip {
    margin-left: 6px;
    color: var(--color-text-3);
    cursor: help;
  }

  .timezone-group {
    flex-wrap: wrap;
  }

  .tz-code {
    margin-left: 6px;
    color: var(--color-text-3);
    font-size: 12px;
    font-family: var(--font-family-mono, monospace);
  }

  :deep(.arco-radio-button.arco-radio-checked) .tz-code {
    color: inherit;
    opacity: 0.8;
  }
</style>
