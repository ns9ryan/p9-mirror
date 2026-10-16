<template>
  <article
    class="summary-card"
    :class="{ 'is-plain': plain, 'is-open': expanded }"
  >
    <div class="summary-head">
      <button type="button" class="identity" @click="onEnter">
        <span class="level-orb" :style="{ background: color }">
          {{ record.level }}
        </span>
        <span class="identity-text">
          <span class="name-line">
            <strong>{{ record.account }}</strong>
            <em v-if="isTop" class="top-tag">{{
              t('agentList.card.topTag')
            }}</em>
          </span>
          <span class="meta">
            ID：{{ record.id }}
            <i></i>
            {{ t('agentList.card.level') }}：{{ record.level }}
            <i></i>
            {{ t('agentList.card.registeredAt') }}：{{ record.registeredAt }}
          </span>
        </span>
      </button>
      <div class="stats">
        <div>
          <span>{{ t('agentList.card.directDown') }}</span>
          <strong>{{ childCount }}</strong>
        </div>
        <div>
          <span>{{ t('agentList.card.teamSize') }}</span>
          <strong>{{ record.allMembers }}</strong>
        </div>
        <div>
          <span>{{ t('agentList.card.balance') }}</span>
          <strong>¥{{ formatAmount(record.balance) }}</strong>
        </div>
        <div>
          <span>{{ t('agentList.card.commission') }}</span>
          <strong>¥{{ formatAmount(record.totalCommission) }}</strong>
        </div>
        <em class="state" :class="record.status">
          {{ t(`agentList.status.${record.status}`) }}
        </em>
        <button
          type="button"
          class="fold"
          :aria-expanded="expanded"
          @click="emit('toggle')"
        >
          <icon-down :class="{ 'is-open': expanded }" />
        </button>
      </div>
    </div>
    <AgentDetailGrid
      v-if="expanded"
      :record="record"
      :child-count="childCount"
    />
  </article>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import AgentDetailGrid from './agent-detail-grid.vue';
  import { formatAmount } from '../format';
  import { levelColor } from '../options';
  import type { AgentRecord } from '../types';

  const props = defineProps<{
    record: AgentRecord;
    childCount: number;
    expanded?: boolean;
    plain?: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'toggle'): void;
    (e: 'enter'): void;
  }>();

  const { t } = useI18n();
  const color = computed(() => levelColor(props.record.level));
  const isTop = computed(() => !props.record.parentAgent.trim());

  function onEnter() {
    if (!props.plain) emit('enter');
  }
</script>

<style scoped lang="less">
  .summary-card {
    overflow: hidden;
    background: #fff;
    border: 1px solid #e5e6eb;
    border-radius: 10px;
  }

  .summary-card + .summary-card {
    margin-top: 12px;
  }

  .is-plain {
    background: #fff;
  }

  .state {
    font-size: 13px;
    font-style: normal;
  }

  .state.normal {
    color: #00b42a;
  }

  .state.disabled {
    color: #f53f3f;
  }

  .summary-head {
    display: flex;
    gap: 16px;
    align-items: center;
    justify-content: space-between;
    padding: 14px 16px;
  }

  .identity {
    display: flex;
    flex: 1;
    gap: 12px;
    align-items: center;
    min-width: 0;
    padding: 0;
    text-align: left;
    background: transparent;
    border: none;
    cursor: pointer;
  }

  .is-plain .identity {
    cursor: default;
  }

  .level-orb {
    display: inline-flex;
    flex: none;
    align-items: center;
    justify-content: center;
    width: 42px;
    height: 42px;
    color: #fff;
    font-weight: 700;
    font-size: 13px;
    border-radius: 50%;
  }

  .name-line {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .name-line strong {
    color: var(--color-text-1);
    font-size: 16px;
  }

  .top-tag {
    padding: 0 8px;
    color: #00b42a;
    font-size: 12px;
    font-style: normal;
    line-height: 22px;
    background: #e8ffea;
    border-radius: 4px;
  }

  .meta {
    display: flex;
    gap: 8px;
    align-items: center;
    margin-top: 4px;
    color: var(--color-text-3);
    font-size: 12px;
  }

  .meta i {
    width: 1px;
    height: 10px;
    background: #e5e6eb;
  }

  .stats {
    display: flex;
    flex: none;
    gap: 28px;
    align-items: center;
  }

  .stats div {
    display: flex;
    flex-direction: column;
    gap: 2px;
    align-items: flex-end;
  }

  .stats span {
    color: var(--color-text-3);
    font-size: 12px;
  }

  .stats strong {
    color: var(--color-text-1);
    font-size: 16px;
  }

  .fold {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    color: var(--color-text-3);
    background: transparent;
    border: none;
    border-radius: 6px;
    cursor: pointer;
  }

  .fold:hover {
    background: #fff;
  }

  .fold :deep(svg) {
    transform: rotate(-90deg);
    transition: transform 0.2s ease;
  }

  .fold :deep(.is-open) {
    transform: rotate(0deg);
  }
</style>
