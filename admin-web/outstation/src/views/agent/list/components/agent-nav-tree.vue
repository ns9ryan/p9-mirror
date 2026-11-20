<template>
  <div class="nav-tree">
    <div v-for="node in displayNodes" :key="node.id" class="nav-node">
      <div
        class="nav-row"
        :class="{ 'is-active': node.id === activeId, 'is-root': isRoot(node) }"
        :style="{ paddingLeft: `${8 + depth * 16}px` }"
      >
        <a-tooltip
          :content="lineText(node)"
          :disabled="!node.lineCode"
          position="tl"
          mini
        >
          <button
            type="button"
            class="nav-label"
            @click="onNodeClick(node)"
          >
            <span class="nav-icon" aria-hidden="true">
              <svg
                v-if="isRoot(node)"
                class="crown"
                viewBox="0 0 24 24"
                width="14"
                height="14"
              >
                <path
                  fill="currentColor"
                  d="M5 16.5 3.2 7.8l4.7 3.2L12 5.5l4.1 5.5 4.7-3.2L19 16.5H5Zm0 1.5h14v2H5v-2Z"
                />
              </svg>
              <svg v-else class="iconfont-svg" aria-hidden="true">
                <use href="#icon-kehuguanli" />
              </svg>
            </span>
            <i v-if="showLayers && !isRoot(node)" class="nav-layer">
              {{ t('agentList.settings.layer', { n: node.layer }) }}
            </i>
            <span>{{ node.account }}</span>
            <em>({{ node.directAgents }})</em>
          </button>
        </a-tooltip>
      </div>
      <AgentNavTree
        v-if="showChildren(node)"
        :nodes="node.children || []"
        :depth="depth + 1"        :active-id="activeId"
        :expanded-ids="expandedIds"
        :show-layers="showLayers"
        :reveal-ids="revealIds"
        @enter="emit('enter', $event)"
        @toggle="emit('toggle', $event)"
      />
    </div>
    <button
      v-if="hiddenCount > 0"
      type="button"
      class="nav-more"
      :style="{ paddingLeft: `${8 + depth * 16}px` }"
      @click="loadMore"
    >
      {{ t('agentList.nav.loadMore', { count: hiddenCount }) }}
    </button>
  </div>
</template>

<script lang="ts" setup>
  import { computed, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { AgentTreeNode } from '../types';

  /** 每层默认先渲染条数，点击「加载更多」再增加 */
  const TREE_PAGE_SIZE = 50;

  const props = withDefaults(
    defineProps<{
      nodes: AgentTreeNode[];
      depth?: number;
      activeId: number | null;
      expandedIds: Set<number>;
      /** 是否展示「n层」标签 */
      showLayers?: boolean;
      /** 必须露出的节点（当前选中 / 展开路径），避免被截断挡住 */
      revealIds?: Set<number>;
    }>(),
    {
      depth: 0,
      showLayers: true,
      revealIds: () => new Set(),
    }
  );

  const emit = defineEmits<{
    (e: 'enter', id: number): void;
    (e: 'toggle', id: number): void;
  }>();

  const { t } = useI18n();
  const visibleCount = ref(TREE_PAGE_SIZE);

  const displayNodes = computed(() =>
    props.nodes.slice(0, visibleCount.value)
  );

  const hiddenCount = computed(() =>
    Math.max(0, props.nodes.length - visibleCount.value)
  );

  function ensureRevealVisible() {
    let maxIdx = -1;
    props.nodes.forEach((node, index) => {
      if (node.id === props.activeId || props.revealIds.has(node.id)) {
        maxIdx = Math.max(maxIdx, index);
      }
    });
    if (maxIdx < 0) return;
    if (maxIdx < visibleCount.value) return;
    const pages = Math.ceil((maxIdx + 1) / TREE_PAGE_SIZE);
    visibleCount.value = Math.min(
      props.nodes.length,
      pages * TREE_PAGE_SIZE
    );
  }

  watch(
    () => props.nodes,
    () => {
      visibleCount.value = TREE_PAGE_SIZE;
      ensureRevealVisible();
    }
  );

  watch(
    () => [props.activeId, props.revealIds, props.nodes.length] as const,
    () => ensureRevealVisible(),
    { immediate: true }
  );

  function loadMore() {
    visibleCount.value = Math.min(
      props.nodes.length,
      visibleCount.value + TREE_PAGE_SIZE
    );
  }

  /** 顶层 root 代理（无「全部代理」虚拟节点） */
  function isRoot(node: AgentTreeNode) {
    return node.parentId === 0;
  }

  function lineText(node: AgentTreeNode) {
    const lineCode = node.lineCode || '';
    const key = `agentList.line.${lineCode.toUpperCase()}`;
    const text = t(key);
    return t('agentList.nav.line', { line: text === key ? lineCode : text });
  }

  function canExpand(node: AgentTreeNode) {
    return node.directAgents > 0 || Boolean(node.children?.length);
  }

  function showChildren(node: AgentTreeNode) {
    if (!node.children?.length) return false;
    return props.expandedIds.has(node.id);
  }

  /** 无箭头时：已展开且当前选中 → 再点收起；否则进入并展开 */
  function onNodeClick(node: AgentTreeNode) {
    if (
      canExpand(node) &&
      props.expandedIds.has(node.id) &&
      node.id === props.activeId
    ) {
      emit('toggle', node.id);
      return;
    }
    emit('enter', node.id);
  }
</script>

<style scoped lang="less">
  .nav-row {
    display: flex;
    gap: 2px;
    align-items: center;
    min-height: 34px;
    padding-right: 8px;
    border-radius: 6px;
  }

  .nav-row.is-active {
    background: rgb(var(--primary-1));
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

  .nav-icon {
    display: inline-flex;
    flex: none;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    color: var(--color-text-2);
    font-size: 22px;
  }

  .nav-icon .crown {
    display: block;
    width: 20px;
    height: 20px;
    color: #637eba;
  }

  .nav-icon .iconfont-svg {
    display: block;
    width: 24px;
    height: 24px;
    overflow: hidden;
  }

  .nav-layer {
    display: inline-flex;
    flex: none;
    align-items: center;
    justify-content: center;
    min-width: 28px;
    height: 18px;
    padding: 0 5px;
    color: #fff;
    font-size: 11px;
    font-style: normal;
    line-height: 1;
    background: rgb(var(--primary-6));
    border-radius: 4px;
  }

  .nav-label > span:not(.nav-icon) {
    overflow: hidden;
    font-size: 13px;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .nav-row.is-root .nav-label > span:not(.nav-icon) {
    color: var(--color-text-1);
    font-weight: 600;
  }

  .nav-label em {
    flex: none;
    color: var(--color-text-3);
    font-size: 12px;
    font-style: normal;
  }

  .nav-more {
    display: block;
    width: 100%;
    margin: 2px 0 6px;
    padding: 6px 8px;
    color: rgb(var(--primary-6));
    font-size: 12px;
    text-align: left;
    background: transparent;
    border: none;
    cursor: pointer;
  }

  .nav-more:hover {
    color: rgb(var(--primary-5));
  }
</style>
