<template>
  <div class="agent-tree">
    <div
      v-for="(node, index) in nodes"
      :key="node.id"
      class="agent-tree-item"
      :class="{
        'is-child': depth > 0,
        'is-last': index === nodes.length - 1,
      }"
    >
      <div class="agent-tree-row">
        <button
          type="button"
          class="tree-switch"
          :aria-expanded="expandedIds.has(node.id)"
          :aria-label="
            expandedIds.has(node.id)
              ? t('agentList.card.collapse')
              : t('agentList.card.expand')
          "
          @click="emit('expand', node.id, !expandedIds.has(node.id))"
        >
          <icon-down
            class="tree-switch-icon"
            :class="{ 'is-open': expandedIds.has(node.id) }"
          />
        </button>
        <AgentCard
          class="agent-tree-card"
          :record="node"
          :selected="selectedIds.has(node.id)"
          :expanded="expandedIds.has(node.id)"
          @update:selected="(checked) => emit('select', node.id, checked)"
          @update:expanded="(value) => emit('expand', node.id, value)"
          @detail="emit('detail', node)"
          @edit="emit('edit', node)"
          @toggle-status="emit('toggleStatus', node)"
        />
      </div>
      <div v-if="node.children?.length" class="agent-tree-children">
        <AgentTree
          :nodes="node.children"
          :depth="depth + 1"
          :selected-ids="selectedIds"
          :expanded-ids="expandedIds"
          @select="(id, checked) => emit('select', id, checked)"
          @expand="(id, value) => emit('expand', id, value)"
          @detail="(record) => emit('detail', record)"
          @edit="(record) => emit('edit', record)"
          @toggle-status="(record) => emit('toggleStatus', record)"
        />
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { useI18n } from 'vue-i18n';
  import AgentCard from './agent-card.vue';
  import type { AgentRecord, AgentTreeNode } from '../types';

  withDefaults(
    defineProps<{
      nodes: AgentTreeNode[];
      depth?: number;
      selectedIds: Set<number>;
      expandedIds: Set<number>;
    }>(),
    {
      depth: 0,
    }
  );

  const emit = defineEmits<{
    (e: 'select', id: number, checked: boolean): void;
    (e: 'expand', id: number, expanded: boolean): void;
    (e: 'detail', record: AgentRecord): void;
    (e: 'edit', record: AgentRecord): void;
    (e: 'toggleStatus', record: AgentRecord): void;
  }>();

  const { t } = useI18n();
</script>

<style scoped lang="less">
  .agent-tree-item + .agent-tree-item {
    margin-top: 10px;
  }

  .agent-tree-item.is-child {
    position: relative;
  }

  .agent-tree-item.is-child::before {
    position: absolute;
    top: 21px;
    left: -22px;
    width: 14px;
    border-top: 1px solid #d5d7e5;
    content: '';
  }

  .agent-tree-item.is-child.is-last::after {
    position: absolute;
    top: 22px;
    bottom: 0;
    left: -24px;
    width: 4px;
    background: var(--color-bg-2);
    content: '';
  }

  .agent-tree-row {
    display: flex;
    gap: 8px;
    align-items: flex-start;
  }

  .agent-tree-card {
    flex: 1;
    min-width: 0;
  }

  .tree-switch {
    display: inline-flex;
    flex: none;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    margin-top: 11px;
    padding: 0;
    color: #4e5969;
    background: transparent;
    border: none;
    border-radius: 4px;
    cursor: pointer;
  }

  .tree-switch:hover {
    background: var(--color-fill-2);
  }

  .tree-switch-icon {
    transform: rotate(-90deg);
    transition: transform 0.2s ease;
  }

  .tree-switch-icon.is-open {
    transform: rotate(0deg);
  }

  .agent-tree-children {
    margin-left: 10px;
    padding-top: 10px;
    padding-left: 22px;
    border-left: 1px solid #d5d7e5;
  }
</style>
