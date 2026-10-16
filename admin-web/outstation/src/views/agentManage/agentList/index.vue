<template>
  <div class="container list-page agent-list-page">
    <Breadcrumb :items="['menu.agentManage', 'menu.agentManage.list']" />
    <a-card class="general-card table-card" :bordered="false">
      <div class="filter-bar">
        <div class="filter-main">
          <a-input
            v-model="form.keyword"
            class="keyword-input"
            allow-clear
            :placeholder="t('agentList.search.keywordAny')"
            @press-enter="handleSearch"
          >
            <template #prefix><icon-search /></template>
          </a-input>
          <a-select
            v-model="form.level"
            class="filter-select"
            allow-clear
            :placeholder="t('agentList.search.levelPlaceholder')"
          >
            <a-option
              v-for="item in settings.levels"
              :key="item.layer"
              :value="item.name"
            >
              {{ item.name }}
            </a-option>
          </a-select>
          <a-select
            v-model="form.status"
            class="filter-select"
            allow-clear
            :placeholder="t('agentList.search.statusAll')"
          >
            <a-option value="normal">{{
              t('agentList.status.normal')
            }}</a-option>
            <a-option value="disabled">
              {{ t('agentList.status.disabled') }}
            </a-option>
          </a-select>
          <a-button @click="handleReset">
            <template #icon><icon-refresh /></template>
            {{ t('front.common.search.reset') }}
          </a-button>
          <a-button type="primary" @click="handleSearch">
            <template #icon><icon-search /></template>
            {{ t('front.common.search.search') }}
          </a-button>
        </div>
        <div class="filter-actions">
          <a-button type="primary" @click="openCreate">
            <template #icon><icon-plus /></template>
            {{ t('agentList.toolbar.create') }}
          </a-button>
          <a-button @click="settingsVisible = true">
            <template #icon><icon-settings /></template>
            {{ t('agentList.toolbar.settings') }}
          </a-button>
          <a-button :disabled="!tree.length" @click="exportCsv">
            <template #icon><icon-download /></template>
            {{ t('agentList.toolbar.export') }}
          </a-button>
        </div>
      </div>

      <div class="list-scroll">
        <div class="level-nav">
          <a-button v-if="focusNode" type="text" @click="goBack">
            <template #icon><icon-left /></template>
            {{ t('agentList.nav.back') }}
          </a-button>
          <a-breadcrumb>
            <a-breadcrumb-item>
              <a v-if="focusNode" class="crumb" @click="goRoot">
                {{ t('menu.agentManage.list') }}
              </a>
              <span v-else>{{ t('menu.agentManage.list') }}</span>
            </a-breadcrumb-item>
            <a-breadcrumb-item v-for="(item, index) in crumbs" :key="item.id">
              <a
                v-if="index < crumbs.length - 1"
                class="crumb"
                @click="enterLevel(item.id)"
              >
                {{ item.account }}
              </a>
              <span v-else>{{ item.account }}</span>
            </a-breadcrumb-item>
          </a-breadcrumb>
        </div>

        <div class="level-body">
          <aside class="nav-pane">
            <button
              type="button"
              class="nav-all"
              :class="{ 'is-active': !focusNode }"
              @click="goRoot"
            >
              <span>{{ t('agentList.nav.all') }}</span>
              <em>({{ allCount }})</em>
            </button>
            <AgentNavTree
              :nodes="tree"
              :active-id="focusId"
              :expanded-ids="expandedIds"
              @enter="enterLevel"
              @toggle="toggleExpanded"
            />
          </aside>

          <section class="level-main">
            <h3 class="section-title">
              {{ levelTitle }}
              <em>({{ filteredList.length }})</em>
            </h3>
            <a-empty
              v-if="!filteredList.length"
              :description="t('agentList.emptyList')"
            />
            <AgentLevelTable
              v-else
              :nodes="pageData"
              :active-id="detailId"
              @enter="enterLevel($event.id)"
              @view="openDetail"
              @edit="openEdit"
              @toggle-status="confirmToggle"
            />
            <div v-if="filteredList.length" class="table-pagination">
              <a-pagination
                :total="filteredList.length"
                :current="pagination.current"
                :page-size="pagination.pageSize"
                :page-size-options="pagination.pageSizeOptions"
                :show-total="pagination.showTotal"
                :show-page-size="pagination.showPageSize"
                :page-size-props="pagination.pageSizeProps"
                @change="onPageChange"
                @page-size-change="onPageSizeChange"
              />
            </div>
          </section>
        </div>
      </div>
    </a-card>

    <AgentDetailDrawer
      :visible="detailVisible"
      :record="detailRecord"
      @update:visible="detailVisible = $event"
      @edit="editFromDetail"
    />
    <AgentFormModal
      :visible="formVisible"
      :mode="formMode"
      :record="editingRecord"
      :settings="settings"
      @update:visible="formVisible = $event"
      @submit="onSubmit"
    />
    <AgentSettingsModal
      :visible="settingsVisible"
      :settings="settings"
      @update:visible="settingsVisible = $event"
      @submit="onSaveSettings"
    />
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import dayjs from 'dayjs';
  import { useI18n } from 'vue-i18n';
  import { Message, Modal } from '@arco-design/web-vue';
  import AgentDetailDrawer from './components/agent-detail-drawer.vue';
  import AgentFormModal from './components/agent-form-modal.vue';
  import AgentLevelTable from './components/agent-level-table.vue';
  import AgentNavTree from './components/agent-nav-tree.vue';
  import AgentSettingsModal from './components/agent-settings-modal.vue';
  import { formatAmount, formatRatio } from './format';
  import createAgentSeed from './mock';
  import { PROMO_HOSTS, createDefaultAgentSettings } from './options';
  import { buildAgentTree, createAgentQuery, flattenAgentTree } from './query';
  import type {
    AccountStatus,
    AgentFormValue,
    AgentQuery,
    AgentRecord,
    AgentSettings,
    AgentTreeNode,
  } from './types';

  const { t } = useI18n();

  const agents = ref<AgentRecord[]>(createAgentSeed());
  const form = reactive<AgentQuery>(createAgentQuery());
  const applied = ref<AgentQuery>(createAgentQuery());
  const expandedIds = ref<Set<number>>(new Set());
  const focusId = ref<number | null>(null);
  const pagination = reactive({
    current: 1,
    pageSize: 10,
    showTotal: true,
    showPageSize: true,
    pageSizeProps: {
      popupContainer: 'body',
    },
    pageSizeOptions: [10, 15, 30, 50],
  });
  const settings = reactive<AgentSettings>(createDefaultAgentSettings());

  const detailVisible = ref(false);
  const detailId = ref<number | null>(null);
  const formVisible = ref(false);
  const formMode = ref<'create' | 'edit'>('create');
  const editingId = ref<number | null>(null);
  const settingsVisible = ref(false);

  const tree = computed(() =>
    buildAgentTree(
      agents.value.filter(
        (item) =>
          !applied.value.clientHost ||
          item.promoHost === applied.value.clientHost
      )
    )
  );

  const nodeById = computed(() => {
    const map = new Map<number, AgentTreeNode>();
    const visit = (nodes: AgentTreeNode[]) => {
      nodes.forEach((node) => {
        map.set(node.id, node);
        if (node.children?.length) visit(node.children);
      });
    };
    visit(tree.value);
    return map;
  });

  const focusNode = computed(() =>
    focusId.value == null ? null : nodeById.value.get(focusId.value) || null
  );

  const crumbs = computed(() => {
    const chain: AgentTreeNode[] = [];
    const guard = new Set<number>();
    let current = focusNode.value;
    while (current && !guard.has(current.id)) {
      guard.add(current.id);
      chain.unshift(current);
      const parentName = current.parentAgent.trim().toLowerCase();
      const parent = parentName
        ? [...nodeById.value.values()].find(
            (item) => item.account.trim().toLowerCase() === parentName
          )
        : undefined;
      current = parent || null;
    }
    return chain;
  });

  watch(crumbs, (chain) => {
    if (!chain.length) return;
    const missing = chain.some((item) => !expandedIds.value.has(item.id));
    if (!missing) return;
    const next = new Set(expandedIds.value);
    chain.forEach((item) => next.add(item.id));
    expandedIds.value = next;
  });

  const currentNodes = computed(() =>
    focusNode.value ? focusNode.value.children ?? [] : tree.value
  );

  const filteredList = computed(() =>
    currentNodes.value.filter((item) => {
      const keyword = applied.value.keyword.trim().toLowerCase();
      if (keyword) {
        const hit =
          String(item.id).includes(keyword) ||
          item.account.toLowerCase().includes(keyword);
        if (!hit) return false;
      }
      if (applied.value.level && item.level !== applied.value.level) {
        return false;
      }
      if (applied.value.status && item.status !== applied.value.status) {
        return false;
      }
      return true;
    })
  );

  const pageData = computed(() => {
    const start = (pagination.current - 1) * pagination.pageSize;
    return filteredList.value.slice(start, start + pagination.pageSize);
  });

  const allCount = computed(() => flattenAgentTree(tree.value).length);

  const levelTitle = computed(() =>
    focusNode.value
      ? `${focusNode.value.account} · ${t('agentList.card.directDown')}`
      : t('agentList.nav.all')
  );

  const detailRecord = computed(
    () => agents.value.find((item) => item.id === detailId.value) || null
  );

  const editingRecord = computed(
    () => agents.value.find((item) => item.id === editingId.value) || null
  );

  watch(filteredList, (list) => {
    const maxPage = Math.max(1, Math.ceil(list.length / pagination.pageSize));
    if (pagination.current > maxPage) pagination.current = maxPage;
  });

  watch(focusNode, (node) => {
    if (focusId.value != null && !node) focusId.value = null;
  });

  function snapshotQuery(): AgentQuery {
    return {
      ...form,
      level: form.level || '',
      status: form.status || '',
      dateRange: [...(form.dateRange || [])],
    };
  }

  function handleSearch() {
    applied.value = snapshotQuery();
    pagination.current = 1;
  }

  function handleReset() {
    Object.assign(form, createAgentQuery());
    handleSearch();
  }

  function onPageChange(current: number) {
    pagination.current = current;
  }

  function onPageSizeChange(size: number) {
    pagination.pageSize = size;
    pagination.current = 1;
  }

  function toggleExpanded(id: number) {
    const next = new Set(expandedIds.value);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    expandedIds.value = next;
  }

  function enterLevel(id: number) {
    focusId.value = id;
    pagination.current = 1;
    const next = new Set(expandedIds.value);
    next.add(id);
    expandedIds.value = next;
  }

  function goRoot() {
    focusId.value = null;
    pagination.current = 1;
  }

  function goBack() {
    const chain = crumbs.value;
    focusId.value = chain.length > 1 ? chain[chain.length - 2].id : null;
    pagination.current = 1;
  }

  function applyStatus(ids: number[], status: AccountStatus) {
    const idSet = new Set(ids);
    let count = 0;
    agents.value = agents.value.map((item) => {
      if (!idSet.has(item.id) || item.status === status) return item;
      count += 1;
      return { ...item, status };
    });
    return count;
  }

  function confirmToggle(record: AgentRecord) {
    const disabling = record.status === 'normal';
    Modal.confirm({
      title: t(
        disabling
          ? 'agentList.confirm.disableTitle'
          : 'agentList.confirm.enableTitle'
      ),
      content: t(
        disabling
          ? 'agentList.confirm.disableContent'
          : 'agentList.confirm.enableContent',
        { id: record.id }
      ),
      okText: t('agentList.confirm.ok'),
      cancelText: t('agentList.confirm.cancel'),
      onOk: () => {
        applyStatus([record.id], disabling ? 'disabled' : 'normal');
        Message.success(
          t(
            disabling
              ? 'agentList.message.disabled'
              : 'agentList.message.enabled'
          )
        );
      },
    });
  }

  function openDetail(record: AgentRecord) {
    detailId.value = record.id;
    detailVisible.value = true;
  }

  function openEdit(record: AgentRecord) {
    formMode.value = 'edit';
    editingId.value = record.id;
    formVisible.value = true;
  }

  function openCreate() {
    formMode.value = 'create';
    editingId.value = null;
    formVisible.value = true;
  }

  function editFromDetail() {
    if (!detailRecord.value) return;
    detailVisible.value = false;
    openEdit(detailRecord.value);
  }

  function nextAgentId() {
    return agents.value.reduce((max, item) => Math.max(max, item.id), 0) + 1;
  }

  function createPromoCode() {
    return Math.random().toString(36).slice(2, 12).toUpperCase();
  }

  function onSubmit(value: AgentFormValue) {
    if (formMode.value === 'create') {
      const now = dayjs().format('YYYY-MM-DD HH:mm:ss');
      const host = applied.value.clientHost || PROMO_HOSTS[0];
      agents.value = [
        {
          ...value,
          id: nextAgentId(),
          promoHost: host,
          promoCode: createPromoCode(),
          directMembers: 0,
          directAgents: 0,
          allMembers: 0,
          allAgents: 0,
          balance: 0,
          totalCommission: 0,
          unpaidCommission: 0,
          visitUv: 0,
          registeredAt: now,
          lastLoginAt: '',
        },
        ...agents.value,
      ];
      pagination.current = 1;
      Message.success(t('agentList.message.created'));
      return;
    }
    agents.value = agents.value.map((item) =>
      item.id === editingId.value ? { ...item, ...value } : item
    );
    Message.success(t('agentList.message.updated'));
  }

  function onSaveSettings(value: AgentSettings) {
    Object.assign(settings, value);
    Message.success(t('agentList.message.settingsSaved'));
  }

  function csvCell(value: string | number) {
    const text = String(value ?? '');
    if (/[",\n]/.test(text)) return `"${text.replace(/"/g, '""')}"`;
    return text;
  }

  function exportCsv() {
    const headers = [
      t('agentList.card.agentId'),
      t('agentList.card.account'),
      t('agentList.card.level'),
      t('agentList.card.parent'),
      t('agentList.card.top'),
      t('agentList.card.source'),
      t('agentList.card.directMembers'),
      t('agentList.card.directAgents'),
      t('agentList.card.allMembers'),
      t('agentList.card.allAgents'),
      t('agentList.card.share'),
      t('agentList.card.balance'),
      t('agentList.card.commission'),
      t('agentList.card.unpaid'),
      t('agentList.card.status'),
      t('agentList.card.withdrawLimit'),
      t('agentList.card.remark'),
      t('agentList.card.link'),
      t('agentList.card.uv'),
      t('agentList.card.registeredAt'),
      t('agentList.card.lastLogin'),
    ];
    const rows = flattenAgentTree(tree.value).map((item) => [
      item.id,
      item.account,
      item.level,
      item.parentAgent,
      item.topAgent,
      item.registerSource,
      item.directMembers,
      item.directAgents,
      item.allMembers,
      item.allAgents,
      formatRatio(item.shareRatio),
      formatAmount(item.balance),
      formatAmount(item.totalCommission),
      formatAmount(item.unpaidCommission),
      t(`agentList.status.${item.status}`),
      t(`agentList.withdraw.${item.withdrawLimit}`),
      item.remark,
      `https://${item.promoHost}?ref=${item.promoCode}`,
      item.visitUv,
      item.registeredAt,
      item.lastLoginAt,
    ]);
    const csv = [headers, ...rows]
      .map((cols) => cols.map((col) => csvCell(col)).join(','))
      .join('\n');
    const blob = new Blob([`\uFEFF${csv}`], {
      type: 'text/csv;charset=utf-8;',
    });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `agents-${dayjs().format('YYYYMMDD-HHmmss')}.csv`;
    link.click();
    URL.revokeObjectURL(url);
    Message.success(t('agentList.message.exported'));
  }
</script>

<style scoped lang="less">
  .container {
    padding: 0 20px 20px;
  }

  .general-card {
    margin-top: 16px;
  }

  .filter-bar {
    display: flex;
    flex-shrink: 0;
    flex-wrap: wrap;
    gap: 12px 16px;
    align-items: center;
    justify-content: space-between;
    margin: 16px 0;
    padding: 12px;
    background: #f7f8fa;
    border-radius: 10px;
  }

  .filter-main,
  .filter-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
  }

  .filter-actions {
    margin-left: auto;
  }

  .keyword-input {
    flex: 0 0 240px;
    width: 240px;
  }

  .filter-bar :deep(.arco-input-wrapper),
  .filter-bar :deep(.arco-select-view-single),
  .filter-bar :deep(.arco-btn-secondary:not(.arco-btn-disabled)) {
    background: #fff;
    border: 1px solid #e5e6eb;
  }

  .filter-bar :deep(.arco-input-wrapper:hover),
  .filter-bar :deep(.arco-select-view-single:hover),
  .filter-bar :deep(.arco-input-wrapper.arco-input-focus),
  .filter-bar :deep(.arco-select-view-single.arco-select-view-focus),
  .filter-bar :deep(.arco-btn-secondary:not(.arco-btn-disabled):hover) {
    background: #fff;
    border-color: rgb(var(--primary-6));
  }

  .filter-main :deep(.arco-select-view-single) {
    flex: 0 0 148px;
    width: 148px;
    max-width: 148px;
  }

  .list-page.container :deep(.table-card.arco-card) {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
    overflow: hidden;
  }

  .list-page.container :deep(.table-card > .arco-card-body) {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
    overflow: hidden;
  }

  .list-scroll {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
    overflow: hidden;
  }

  .level-nav {
    display: flex;
    flex-shrink: 0;
    gap: 8px;
    align-items: center;
    margin-bottom: 12px;
  }

  .crumb {
    color: rgb(var(--primary-6));
    cursor: pointer;
  }

  .level-body {
    display: flex;
    flex: 1;
    gap: 16px;
    min-height: 0;
  }

  .nav-pane,
  .level-main {
    min-height: 0;
    overflow: auto;
    background: #fff;
    border: 1px solid #e5e6eb;
    border-radius: 10px;
  }

  .nav-pane {
    flex: none;
    width: 280px;
    padding: 8px;
  }

  .nav-all {
    display: flex;
    gap: 6px;
    align-items: center;
    width: 100%;
    height: 36px;
    padding: 0 8px;
    color: var(--color-text-1);
    font-weight: 600;
    text-align: left;
    background: transparent;
    border: none;
    border-radius: 6px;
    cursor: pointer;
  }

  .nav-all em {
    color: var(--color-text-3);
    font-weight: 400;
    font-style: normal;
  }

  .nav-all.is-active {
    background: #e8f3ff;
  }

  .level-main {
    flex: 1;
  }

  .section-title {
    display: flex;
    gap: 6px;
    align-items: center;
    margin: 0;
    padding: 14px 16px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 16px;
  }

  .section-title em {
    color: var(--color-text-3);
    font-weight: 400;
    font-style: normal;
  }

  .table-pagination {
    display: flex;
    justify-content: flex-end;
    padding-top: 16px;
  }

  :deep(.arco-pagination-options .arco-select) {
    min-width: 100px;
  }
</style>

<style lang="less">
  .layout-content:has(.tab-bar-container) .agent-list-page.list-page {
    height: calc(100vh - 122px);
  }
</style>
