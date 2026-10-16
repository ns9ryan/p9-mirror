<template>
  <div class="nav-tree">
    <div
      v-for="node in nodes"
      :key="node.id"
      class="nav-node"
      :style="{ paddingLeft: `${depth * 16}px` }"
    >
      <div class="nav-row" :class="{ 'is-active': node.id === activeId }">
        <button
          v-if="node.children?.length"
          type="button"
          class="fold"
          @click="emit('toggle', node.id)"
        >
          <icon-down :class="{ 'is-open': expandedIds.has(node.id) }" />
        </button>
        <span v-else class="fold-gap"></span>
        <button type="button" class="nav-label" @click="emit('enter', node.id)">
          <i :style="{ background: levelColor(node.level) }">{{
            node.level
          }}</i>
          <span>{{ node.account }}</span>
          <em>({{ node.children?.length ?? 0 }})</em>
        </button>
      </div>
      <AgentNavTree
        v-if="node.children?.length && expandedIds.has(node.id)"
        :nodes="node.children"
        :depth="depth + 1"
        :active-id="activeId"
        :expanded-ids="expandedIds"
        @enter="emit('enter', $event)"
        @toggle="emit('toggle', $event)"
      />
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { levelColor } from '../options';
  import type { AgentTreeNode } from '../types';

  withDefaults(
    defineProps<{
      nodes: AgentTreeNode[];
      depth?: number;
      activeId: number | null;
      expandedIds: Set<number>;
    }>(),
    { depth: 0 }
  );

  const emit = defineEmits<{
    (e: 'enter', id: number): void;
    (e: 'toggle', id: number): void;
  }>();
</script>

<style scoped lang="less">
  .nav-row {
    display: flex;
    gap: 2px;
    align-items: center;
    min-height: 34px;
    border-radius: 6px;
  }

  .nav-row.is-active {
    background: #e8f3ff;
  }

  .fold,
  .fold-gap {
    flex: none;
    width: 22px;
    height: 22px;
  }

  .fold {
    display: inline-flex;
    align-items: center;
    justify-content: center;
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

  .nav-label {
    display: flex;
    flex: 1;
    gap: 6px;
    align-items: center;
    min-width: 0;
    padding: 0;
    color: var(--color-text-1);
    text-align: left;
    background: transparent;
    border: none;
    cursor: pointer;
  }

  .nav-label i {
    display: inline-flex;
    flex: none;
    align-items: center;
    justify-content: center;
    min-width: 22px;
    height: 16px;
    padding: 0 4px;
    color: #fff;
    font-size: 10px;
    font-style: normal;
    border-radius: 4px;
  }

  .nav-label span {
    overflow: hidden;
    font-size: 13px;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .nav-label em {
    flex: none;
    color: var(--color-text-3);
    font-size: 12px;
    font-style: normal;
  }
</style>
