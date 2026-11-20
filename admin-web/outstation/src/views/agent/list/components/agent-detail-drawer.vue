<template>
  <a-drawer
    :visible="visible"
    :width="480"
    :title="t('agentList.detail.title')"
    :footer="false"
    :body-style="{ paddingTop: 0 }"
    unmount-on-close
    @cancel="emit('update:visible', false)"
  >
    <div v-if="record" ref="rootRef">
      <div class="drawer-profile">
        <span class="level-orb" :style="{ background: color }">
          {{ record.level }}
        </span>
        <div>
          <strong>{{ record.account }}</strong>
          <p>ID：{{ record.id }}</p>
        </div>
        <em :class="record.status">
          {{ t(`agentList.status.${record.status}`) }}
        </em>
      </div>
      <div ref="navRef" class="anchor-nav">
        <a-tabs
          :active-key="activeKey"
          :header-padding="false"
          hide-content
          @tab-click="scrollToSection"
        >
          <a-tab-pane
            v-for="section in sections"
            :key="section.key"
            :title="section.title"
          />
        </a-tabs>
      </div>
      <div class="section-stack">
        <section
          v-for="section in sections"
          :key="section.key"
          :ref="(el) => setSectionRef(section.key, el)"
        >
          <h4>
            <i class="dot" :style="{ background: section.color }"></i>
            {{ section.title }}
          </h4>
          <a-descriptions :data="section.items" :column="1" size="medium" />
        </section>
      </div>
      <div class="drawer-actions">
        <a-button @click="emit('update:visible', false)">
          {{ t('agentList.detail.close') }}
        </a-button>
        <a-button type="primary" @click="emit('edit')">
          {{ t('agentList.card.edit') }}
        </a-button>
      </div>
    </div>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { displayText, formatAmount, formatRatio } from '../format';
  import { buildPromoLink, levelColor } from '../options';
  import type { AgentRecord } from '../types';

  const props = defineProps<{
    visible: boolean;
    record: AgentRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'edit'): void;
  }>();

  const { t } = useI18n();

  const rootRef = ref<HTMLElement | null>(null);
  const navRef = ref<HTMLElement | null>(null);
  const sectionRefs = new Map<string, HTMLElement>();
  const activeKey = ref('basic');
  let container: HTMLElement | null = null;
  let scrollingByClick = false;
  let scrollEndTimer: number | undefined;

  const color = computed(() =>
    props.record
      ? levelColor(props.record.level, props.record.layer)
      : '#64748b'
  );

  const promoLink = computed(() =>
    props.record
      ? buildPromoLink(props.record.promoHost, props.record.promoCode)
      : ''
  );

  const sections = computed(() => {
    const record = props.record;
    if (!record) return [];
    return [
      {
        key: 'basic',
        title: t('agentList.card.basic'),
        color: '#14c9c9',
        items: [
          { label: t('agentList.card.level'), value: record.level },
          {
            label: t('agentList.card.parent'),
            value: displayText(record.parentAgent),
          },
          { label: t('agentList.card.region'), value: t('agentList.empty') },
          { label: t('agentList.card.source'), value: record.registerSource },
          {
            label: t('agentList.card.registeredAt'),
            value: record.registeredAt,
          },
          {
            label: t('agentList.card.lastLogin'),
            value: displayText(record.lastLoginAt),
          },
          {
            label: t('agentList.card.remark'),
            value: displayText(record.remark),
          },
        ],
      },
      {
        key: 'team',
        title: t('agentList.card.team'),
        color: '#36a3f7',
        items: [
          {
            label: t('agentList.card.directMembers'),
            value: record.directMembers,
          },
          {
            label: t('agentList.card.directAgents'),
            value: record.directAgents,
          },
          { label: t('agentList.card.allMembers'), value: record.allMembers },
          { label: t('agentList.card.allAgents'), value: record.allAgents },
          {
            label: t('agentList.card.share'),
            value: formatRatio(record.shareRatio),
          },
        ],
      },
      {
        key: 'funds',
        title: t('agentList.card.funds'),
        color: '#3b6cff',
        items: [
          {
            label: t('agentList.card.balance'),
            value: `¥${formatAmount(record.balance)}`,
          },
          {
            label: t('agentList.card.commission'),
            value: `¥${formatAmount(record.totalCommission)}`,
          },
          {
            label: t('agentList.card.unpaid'),
            value: `¥${formatAmount(record.unpaidCommission)}`,
          },
          {
            label: t('agentList.card.withdrawLimit'),
            value: t(`agentList.withdraw.${record.withdrawLimit}`),
          },
        ],
      },
      {
        key: 'promo',
        title: t('agentList.card.promo'),
        color: '#f59a23',
        items: [
          { label: t('agentList.card.link'), value: promoLink.value },
          { label: t('agentList.card.uv'), value: record.visitUv },
          {
            label: t('agentList.card.lastLogin'),
            value: displayText(record.lastLoginAt),
          },
        ],
      },
      {
        key: 'status',
        title: t('agentList.card.status'),
        color: '#7c6bf2',
        items: [
          {
            label: t('agentList.column.status'),
            value: t(`agentList.status.${record.status}`),
          },
          { label: t('agentList.card.online'), value: t('agentList.empty') },
          { label: t('agentList.card.device'), value: t('agentList.empty') },
        ],
      },
    ];
  });

  function setSectionRef(key: string, el: unknown) {
    if (el instanceof HTMLElement) sectionRefs.set(key, el);
    else sectionRefs.delete(key);
  }

  function sectionOffset(el: HTMLElement) {
    if (!container) return 0;
    const navHeight = navRef.value?.offsetHeight ?? 0;
    return (
      el.getBoundingClientRect().top -
      container.getBoundingClientRect().top +
      container.scrollTop -
      navHeight
    );
  }

  function syncActiveFromScroll() {
    if (!container) return;
    if (scrollingByClick) {
      window.clearTimeout(scrollEndTimer);
      scrollEndTimer = window.setTimeout(() => {
        scrollingByClick = false;
      }, 120);
      return;
    }
    const atBottom =
      container.scrollTop + container.clientHeight >=
      container.scrollHeight - 2;
    if (atBottom) {
      activeKey.value = sections.value[sections.value.length - 1].key;
      return;
    }
    let current = sections.value[0]?.key ?? 'basic';
    sections.value.forEach(({ key }) => {
      const el = sectionRefs.get(key);
      if (el && sectionOffset(el) <= container!.scrollTop + 1) current = key;
    });
    activeKey.value = current;
  }

  function scrollToSection(key: string | number) {
    const el = sectionRefs.get(String(key));
    if (!container || !el) return;
    activeKey.value = String(key);
    scrollingByClick = true;
    container.scrollTo({ top: sectionOffset(el), behavior: 'smooth' });
    window.clearTimeout(scrollEndTimer);
    scrollEndTimer = window.setTimeout(() => {
      scrollingByClick = false;
    }, 600);
  }

  function unbindContainer() {
    container?.removeEventListener('scroll', syncActiveFromScroll);
    container = null;
  }

  watch(rootRef, (el) => {
    unbindContainer();
    container = el?.closest<HTMLElement>('.arco-drawer-body') ?? null;
    container?.addEventListener('scroll', syncActiveFromScroll, {
      passive: true,
    });
  });

  watch(
    () => props.record?.id,
    async () => {
      activeKey.value = 'basic';
      await nextTick();
      container?.scrollTo({ top: 0 });
    }
  );

  onBeforeUnmount(() => {
    window.clearTimeout(scrollEndTimer);
    unbindContainer();
  });
</script>

<style scoped lang="less">
  .drawer-profile {
    display: flex;
    gap: 12px;
    align-items: center;
    margin: 12px 0;
  }

  .drawer-profile strong {
    color: var(--color-text-1);
    font-size: 16px;
  }

  .drawer-profile p {
    margin: 4px 0 0;
    color: var(--color-text-3);
    font-size: 12px;
  }

  .drawer-profile em {
    margin-left: auto;
    font-style: normal;
  }

  .drawer-profile em.normal {
    color: rgb(var(--green-6));
  }

  .drawer-profile em.disabled {
    color: rgb(var(--red-6));
  }

  .level-orb {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 42px;
    height: 42px;
    color: #fff;
    font-weight: 700;
    border-radius: 50%;
  }

  .anchor-nav {
    position: sticky;
    top: 0;
    z-index: 2;
    margin: 0 -16px 12px;
    padding: 0 16px;
    background: var(--color-bg-3);

    :deep(.arco-tabs-nav-tab),
    :deep(.arco-tabs-nav-tab-list) {
      flex: 1;
    }

    :deep(.arco-tabs-nav-tab-list) {
      display: flex;
    }

    :deep(.arco-tabs-tab) {
      flex: 1;
      justify-content: center;
      margin: 0;
    }
  }

  .section-stack section {
    padding: 12px 14px 4px;
    border: 1px solid var(--color-border-1);
    border-radius: 8px;
  }

  .section-stack section + section {
    margin-top: 12px;
  }

  .section-stack h4 {
    display: flex;
    gap: 8px;
    align-items: center;
    margin: 0 0 10px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 14px;
  }

  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }

  .drawer-actions {
    display: flex;
    gap: 8px;
    justify-content: flex-end;
    margin-top: 16px;
  }
</style>
