<template>
  <div class="level-table">
    <div class="level-head">
      <span></span>
      <span>{{ t('agentList.section.children') }}</span>
      <span>{{ t('agentList.column.level') }}</span>
      <span>{{ t('agentList.card.directDown') }}</span>
      <span>{{ t('agentList.card.teamSize') }}</span>
      <span>{{ t('agentList.column.status') }}</span>
      <span>{{ t('agentList.column.action') }}</span>
    </div>
    <div
      v-for="node in nodes"
      :key="node.id"
      class="level-item"
      :class="{ 'is-active': node.id === activeId }"
    >
      <div class="level-row">
        <button
          type="button"
          class="fold"
          :aria-expanded="expandedIds.has(node.id)"
          :aria-label="
            expandedIds.has(node.id)
              ? t('agentList.card.collapse')
              : t('agentList.card.expand')
          "
          @click="toggle(node.id)"
        >
          <icon-down :class="{ 'is-open': expandedIds.has(node.id) }" />
        </button>
        <button type="button" class="agent-cell" @click="emit('enter', node)">
          <span
            class="level-orb"
            :style="{ background: levelColor(node.level) }"
          >
            {{ node.level }}
          </span>
          <span>
            <strong>{{ node.account }}</strong>
            <small>ID：{{ node.id }}</small>
          </span>
        </button>
        <span>{{ node.level }}</span>
        <span>{{ node.children?.length ?? 0 }}</span>
        <span>{{ node.allMembers }}</span>
        <span :class="['state', node.status]">
          {{ t(`agentList.status.${node.status}`) }}
        </span>
        <span class="ops" @click.stop>
          <a-button type="text" size="small" @click="emit('view', node)">
            {{ t('agentList.card.view') }}
          </a-button>
          <a-dropdown trigger="click" @select="(value) => onMore(node, value)">
            <a-button type="text" size="small">
              {{ t('agentList.card.more') }}
            </a-button>
            <template #content>
              <a-doption value="edit">{{ t('agentList.card.edit') }}</a-doption>
              <a-doption value="toggle">
                {{
                  node.status === 'normal'
                    ? t('agentList.card.disable')
                    : t('agentList.card.enable')
                }}
              </a-doption>
            </template>
          </a-dropdown>
        </span>
      </div>
      <AgentDetailGrid
        v-if="expandedIds.has(node.id)"
        :record="node"
        :child-count="node.children?.length ?? 0"
      />
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import AgentDetailGrid from './agent-detail-grid.vue';
  import { levelColor } from '../options';
  import type { AgentTreeNode } from '../types';

  defineProps<{
    nodes: AgentTreeNode[];
    activeId: number | null;
  }>();

  const emit = defineEmits<{
    (e: 'enter', node: AgentTreeNode): void;
    (e: 'view', node: AgentTreeNode): void;
    (e: 'edit', node: AgentTreeNode): void;
    (e: 'toggleStatus', node: AgentTreeNode): void;
  }>();

  const { t } = useI18n();
  const expandedIds = ref(new Set<number>());

  function toggle(id: number) {
    const next = new Set(expandedIds.value);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    expandedIds.value = next;
  }

  function onMore(
    node: AgentTreeNode,
    value: string | number | Record<string, unknown> | undefined
  ) {
    if (value === 'edit') emit('edit', node);
    if (value === 'toggle') emit('toggleStatus', node);
  }
</script>

<style scoped lang="less">
  .level-head,
  .level-row {
    display: grid;
    grid-template-columns: 36px minmax(180px, 1.6fr) 72px 88px 88px 72px 112px;
    gap: 8px;
    align-items: center;
    padding: 0 16px;
  }

  .level-head {
    height: 44px;
    color: var(--color-text-3);
    font-size: 13px;
    background: #f7f8fa;
  }

  .level-item + .level-item {
    border-top: 1px solid #f2f3f5;
  }

  .level-row {
    min-height: 64px;
  }

  .level-item.is-active {
    background: #f5f8ff;
  }

  .fold {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    padding: 0;
    color: var(--color-text-3);
    background: transparent;
    border: none;
    cursor: pointer;
  }

  .fold :deep(svg) {
    transform: rotate(-90deg);
    transition: transform 0.2s ease;
  }

  .fold :deep(.is-open) {
    transform: rotate(0deg);
  }

  .agent-cell {
    display: flex;
    gap: 10px;
    align-items: center;
    min-width: 0;
    padding: 0;
    text-align: left;
    background: transparent;
    border: none;
    cursor: pointer;
  }

  .level-orb {
    display: inline-flex;
    flex: none;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    color: #fff;
    font-weight: 700;
    font-size: 12px;
    border-radius: 50%;
  }

  .agent-cell strong,
  .agent-cell small {
    display: block;
  }

  .agent-cell strong {
    color: var(--color-text-1);
    font-size: 14px;
  }

  .agent-cell small {
    margin-top: 2px;
    color: var(--color-text-3);
    font-size: 12px;
  }

  .state.normal {
    color: #00b42a;
  }

  .state.disabled {
    color: #f53f3f;
  }

  .ops {
    display: flex;
    gap: 4px;
  }
</style>
