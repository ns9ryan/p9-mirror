<template>
  <div class="level-table">
    <div class="level-head">
      <span></span>
      <span>{{ t('agentList.section.children') }}</span>
      <span>{{ t('agentList.card.directDown') }}</span>
      <span>{{ t('agentList.card.directMembers') }}</span>
      <span>{{ t('agentList.card.teamSize') }}</span>
      <span>{{ t('agentList.card.share') }}</span>
      <span>{{ t('agentList.card.balance') }}</span>
      <span>{{ t('agentList.card.commission') }}</span>
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
          v-if="node.id > 0"
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
          <AgentFoldIcon :open="expandedIds.has(node.id)" />
        </button>
        <span v-else></span>
        <button type="button" class="agent-cell" @click="emit('enter', node)">
          <span
            v-if="showLevel"
            class="level-orb"
            :class="{ 'is-root': node.id < 0 }"
            :style="
              node.id > 0
                ? { background: levelColor(node.level, node.layer) }
                : undefined
            "
          >
            {{ levelText(node) }}
          </span>
          <span>
            <strong>{{ node.account }}</strong>
            <small v-if="node.id > 0">ID：{{ node.id }}</small>
          </span>
        </button>
        <span>{{ node.directAgents }}</span>
        <span>{{
          node.detailReady === false ? '—' : node.directMembers
        }}</span>
        <span>{{ node.detailReady === false ? '—' : node.allMembers }}</span>
        <span>{{
          node.detailReady === false ? '—' : formatRatio(node.shareRatio)
        }}</span>
        <span>{{
          node.detailReady === false
            ? '—'
            : `¥${formatAmount(node.balance)}`
        }}</span>
        <span>{{
          node.detailReady === false
            ? '—'
            : `¥${formatAmount(node.totalCommission)}`
        }}</span>
        <span
          v-if="node.id > 0 && node.detailReady !== false"
          :class="['state', node.status]"
        >
          {{ t(`agentList.status.${node.status}`) }}
        </span>
        <span v-else>—</span>
        <span class="ops" @click.stop>
          <template v-if="node.id > 0 && !rootIds?.has(node.id)">
            <a-button type="text" size="small" @click="emit('view', node)">
              {{ t('agentList.card.view') }}
            </a-button>
            <a-dropdown
              trigger="click"
              @select="(value) => onMore(node, value)"
            >
              <a-button type="text" size="small">
                {{ t('agentList.card.more') }}
              </a-button>
              <template #content>
                <a-doption value="edit">{{
                  t('agentList.card.edit')
                }}</a-doption>
                <a-doption value="toggle">
                  {{
                    node.status === 'normal'
                      ? t('agentList.card.disable')
                      : t('agentList.card.enable')
                  }}
                </a-doption>
              </template>
            </a-dropdown>
          </template>
        </span>
      </div>
      <AgentDetailGrid
        v-if="node.id > 0 && expandedIds.has(node.id)"
        :record="node"
        :child-count="node.directAgents"
      />
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import AgentDetailGrid from './agent-detail-grid.vue';
  import AgentFoldIcon from './agent-fold-icon.vue';
  import { formatAmount, formatRatio } from '../format';
  import { levelColor } from '../options';
  import type { AgentRecord } from '../types';

  withDefaults(
    defineProps<{
      nodes: AgentRecord[];
      activeId: number | null;
      showLevel?: boolean;
      /** 顶层代理（root）不提供查看 / 编辑 / 启停 */
      rootIds?: Set<number>;
    }>(),
    { showLevel: true }
  );

  const emit = defineEmits<{
    (e: 'enter', node: AgentRecord): void;
    (e: 'view', node: AgentRecord): void;
    (e: 'edit', node: AgentRecord): void;
    (e: 'toggleStatus', node: AgentRecord): void;
  }>();

  const { t } = useI18n();
  const expandedIds = ref(new Set<number>());

  function levelText(node: AgentRecord) {
    if (node.level || node.id <= 0) return node.level;
    return t('agentList.settings.layer', { n: node.layer });
  }

  function toggle(id: number) {
    const next = new Set(expandedIds.value);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    expandedIds.value = next;
  }

  function onMore(
    node: AgentRecord,
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
    grid-template-columns: var(
      --agent-cols,
      36px minmax(140px, 1.2fr) minmax(64px, 0.55fr) minmax(64px, 0.55fr)
        minmax(64px, 0.55fr) minmax(64px, 0.55fr) minmax(88px, 0.8fr)
        minmax(88px, 0.8fr) minmax(56px, 0.5fr) minmax(96px, 0.85fr)
    );
    gap: 8px;
    align-items: center;
    padding: 0 16px;
  }

  .level-head {
    height: 40px;
    color: var(--color-text-3);
    font-size: 13px;
    background: var(--color-fill-2);
  }

  .level-item + .level-item {
    border-top: 1px solid var(--color-border-1);
  }

  .level-row {
    min-height: 52px;
  }

  .level-item.is-active {
    background: var(--color-fill-1);
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
    background: #86909c;
    border-radius: 50%;
  }

  .level-orb.is-root {
    width: auto;
    min-width: 36px;
    padding: 0 8px;
    border-radius: 18px;
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
    color: rgb(var(--green-6));
  }

  .state.disabled {
    color: rgb(var(--red-6));
  }

  .ops {
    display: flex;
    gap: 4px;
  }
</style>
