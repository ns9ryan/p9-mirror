<template>
  <div class="container list-page agent-list-page">
    <a-card class="general-card table-card" :bordered="false">
      <a-tabs v-model:active-key="activeTab" class="page-tabs">
        <a-tab-pane key="list" :title="t('agentList.tab.tree')" />
        <a-tab-pane key="query" :title="t('agentList.tab.query')" />
        <a-tab-pane key="review">
          <template #title>
            {{ t('agentList.tab.review') }}
            <a-badge
              v-if="reviewPending"
              :count="reviewPending"
              :offset="[4, -8]"
            />
          </template>
        </a-tab-pane>

      </a-tabs>
      <AgentQueryList v-show="activeTab === 'query'" @view="onQueryView" />
      <AgentReviewList
        v-show="activeTab === 'review'"
        @pending-change="reviewPending = $event"
      />
      <div v-show="activeTab === 'list'" class="filter-bar">
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
          <!-- <a-select
            v-model="form.level"
            class="filter-select"
            allow-clear
            :placeholder="t('agentList.search.levelPlaceholder')"
          >
            <a-option
              v-for="item in settings.levels"
              :key="item.layer"
              :value="String(item.layer)"
            >
              {{ item.name }}
            </a-option>
          </a-select> -->
          <!-- <a-select
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
          </a-select> -->
          <a-button @click="handleReset">
            <template #icon><icon-refresh /></template>
            {{ t('front.common.search.reset') }}
          </a-button>
          <a-button
            type="primary"
            :loading="locateLoading"
            @click="handleSearch"
          >
            <template #icon><icon-search /></template>
            {{ t('front.common.search.search') }}
          </a-button>
        </div>
        <div class="filter-actions">
          <a-tooltip
            v-if="createBlocked"
            :content="createBlockedTip"
          >
            <span class="create-blocked">
              <a-button type="primary" disabled>
                <template #icon><icon-plus /></template>
                {{ t('agentList.toolbar.create') }}
              </a-button>
            </span>
          </a-tooltip>
          <a-button v-else type="primary" @click="openCreate">
            <template #icon><icon-plus /></template>
            {{ t('agentList.toolbar.create') }}
          </a-button>
          <a-button :disabled="!total" @click="exportCsv">
            <template #icon><icon-download /></template>
            {{ t('agentList.toolbar.export') }}
          </a-button>
        </div>
      </div>

      <div v-if="activeTab === 'list' && locatePath.length" class="locate-hint">
        <span class="locate-hint__label">
          {{ t('agentList.locate.label') }}
        </span>
        <a-breadcrumb class="locate-hint__path">
          <a-breadcrumb-item
            v-for="(item, index) in locatePath"
            :key="`${item.id}-${index}`"
          >
            <a
              class="locate-hint__link"
              :class="{
                'is-target': index === locatePath.length - 1,
              }"
              @click="enterLevel(item.id)"
            >
              {{ item.account }}
            </a>
          </a-breadcrumb-item>
        </a-breadcrumb>
        <a-button type="text" size="mini" @click="clearLocate">
          {{ t('agentList.locate.close') }}
        </a-button>
      </div>

      <div v-show="activeTab === 'list'" class="list-scroll">
        <div class="level-nav">
          <a-button v-if="crumbs.length > 1" type="text" @click="goBack">
            <template #icon><icon-left /></template>
            {{ t('agentList.nav.back') }}
          </a-button>
          <a-breadcrumb>
            <a-breadcrumb-item>
              <a v-if="!isRootFocus" class="crumb" @click="goRoot">
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
            <div class="layer-control">
              <div class="layer-control__head">
                <span class="layer-control__title">
                  <icon-layers />
                  {{ t('agentList.layerControl.title') }}
                </span>
                <div class="layer-control__actions">
                  <button
                    type="button"
                    :class="{ 'is-active': showLayerBadges }"
                    @click="showLayerBadges = true"
                  >
                    {{ t('agentList.layerControl.showAll') }}
                  </button>
                  <span class="layer-control__divider" />
                  <button
                    type="button"
                    :class="{ 'is-active': !showLayerBadges }"
                    @click="showLayerBadges = false"
                  >
                    {{ t('agentList.layerControl.hideAll') }}
                  </button>
                </div>
              </div>
            </div>
            <div class="nav-tree-wrap">
              <AgentNavTree
                :nodes="navTree"
                :active-id="focusId"
                :expanded-ids="expandedIds"
                :show-layers="showLayerBadges"
                :reveal-ids="treeRevealIds"
                @enter="enterLevel"
                @toggle="toggleExpanded"
              />
            </div>
          </aside>

          <section class="level-main">
            <div v-if="focusNode" class="parent-panel">
              <div class="parent-head">
                <span></span>
                <span>{{ t('agentList.parent.label') }}</span>
                <span>{{ t('agentList.card.directDown') }}</span>
                <span>{{ t('agentList.card.directMembers') }}</span>
                <span>{{ t('agentList.card.teamSize') }}</span>
                <span>{{ t('agentList.card.share') }}</span>
                <span>{{ t('agentList.card.balance') }}</span>
                <span>{{ t('agentList.card.commission') }}</span>
                <span>{{ t('agentList.column.status') }}</span>
                <span>{{ t('agentList.column.action') }}</span>
              </div>
              <div class="parent-panel-body">
                <div class="parent-bar">
                  <button
                    type="button"
                    class="fold"
                    :aria-expanded="parentExpanded"
                    :aria-label="
                      parentExpanded
                        ? t('agentList.card.collapse')
                        : t('agentList.card.expand')
                    "
                    @click="toggleParent"
                  >
                    <AgentFoldIcon :open="parentExpanded" />
                  </button>
                  <span class="parent-main">
                    <span
                      v-if="showLayerBadges"
                      class="level-orb"
                      :style="{
                        background: levelColor(
                          focusNode.level,
                          focusNode.layer
                        ),
                      }"
                    >
                      {{ levelText(focusNode) }}
                    </span>
                    <span class="parent-id">
                      <span class="parent-name">
                        <strong>{{ focusNode.account }}</strong>
                      </span>
                      <small>ID：{{ focusNode.id }}</small>
                    </span>
                  </span>
                  <span>{{ focusNode.directAgents }}</span>
                  <span>{{
                    focusLoaded ? focusNode.directMembers : '—'
                  }}</span>
                  <span>{{ focusLoaded ? focusNode.allMembers : '—' }}</span>
                  <span>{{
                    focusLoaded ? formatRatio(focusNode.shareRatio) : '—'
                  }}</span>
                  <span>{{
                    focusLoaded ? `¥${formatAmount(focusNode.balance)}` : '—'
                  }}</span>
                  <span>{{
                    focusLoaded
                      ? `¥${formatAmount(focusNode.totalCommission)}`
                      : '—'
                  }}</span>
                  <span
                    v-if="focusLoaded"
                    class="parent-state"
                    :class="focusNode.status"
                  >
                    {{ t(`agentList.status.${focusNode.status}`) }}
                  </span>
                  <span v-else>—</span>
                  <span class="parent-ops">
                    <template v-if="focusLoaded && !rootAgentFocus">
                      <a-button type="text" size="small" @click="viewParent">
                        {{ t('agentList.card.view') }}
                      </a-button>
                      <a-dropdown trigger="click" @select="onParentMore">
                        <a-button type="text" size="small">
                          {{ t('agentList.card.more') }}
                        </a-button>
                        <template #content>
                          <a-doption value="edit">{{
                            t('agentList.card.edit')
                          }}</a-doption>
                          <a-doption value="toggle">
                            {{
                              focusNode.status === 'normal'
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
                  v-if="focusLoaded && focusNode && parentExpanded"
                  :record="focusNode"
                  :child-count="focusNode.directAgents"
                />
              </div>
            </div>

            <div class="child-panel">
              <h3 class="section-title">
                {{
                  isRootFocus
                    ? t('agentList.nav.all')
                    : t('agentList.card.directDown')
                }}
                <em>({{ total }})</em>
              </h3>
              <div class="child-panel-body">
                <a-spin :loading="loading" class="level-spin">
                  <a-empty
                    v-if="!loading && !pageData.length"
                    :description="t('agentList.emptyList')"
                  />
                  <AgentLevelTable
                    v-else-if="pageData.length"
                    :nodes="pageData"
                    :active-id="detailId"
                    :show-level="showLayerBadges"
                    :root-ids="rootAgentIds"
                    @enter="enterLevel($event.id)"
                    @view="openDetail"
                    @edit="openEdit"
                    @toggle-status="confirmToggle"
                  />
                  <div v-if="total" class="table-pagination">
                    <a-pagination
                      :total="total"
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
                </a-spin>
              </div>
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
      :parent-id="focusId"
      :parent-account="focusNode?.account || ''"
      :line-code="focusLineCode"
      :parent-level="focusLevel"
      @update:visible="formVisible = $event"
      @submit="onSubmit"
      @created="onAgentCreated"
    />
  </div>
</template>

<script lang="ts" setup>
  import { computed, onMounted, reactive, ref, watch } from 'vue';
  import dayjs from 'dayjs';
  import { useI18n } from 'vue-i18n';
  import { Message, Modal } from '@arco-design/web-vue';
  import {
    AGENT_FILTER,
    AGENT_STATUS,
    listAgentTreeAncestors,
    listAgentTreeNodes,
    listAgents,
  } from '@/api/agent';
  import type { AgentListParams, AgentTreeAncestorItem } from '@/api/agent';
  import AgentDetailDrawer from './components/agent-detail-drawer.vue';
  import AgentDetailGrid from './components/agent-detail-grid.vue';
  import AgentFoldIcon from './components/agent-fold-icon.vue';
  import AgentFormModal from './components/agent-form-modal.vue';
  import AgentLevelTable from './components/agent-level-table.vue';
  import AgentNavTree from './components/agent-nav-tree.vue';
  import AgentQueryList from './components/agent-query-list.vue';
  import AgentReviewList from './components/agent-review-list.vue';
  import { formatAmount, formatRatio } from './format';
  import { createEmptyAgent, mapAgentListItem, mapTreeNode } from './mapper';
  import applyTreeStressPad from './tree-stress';
  import {
    buildPromoLink,
    createDefaultAgentSettings,
    levelColor,
    levelLayer,
  } from './options';
  import { createAgentQuery } from './query';
  import type {
    AccountStatus,
    AgentFormValue,
    AgentQuery,
    AgentRecord,
    AgentSettings,
    AgentTreeNode,
  } from './types';

  const { t } = useI18n();
  const ROOT_ID = 0;

  const form = reactive<AgentQuery>(createAgentQuery());
  const applied = ref<AgentQuery>(createAgentQuery());
  const expandedIds = ref<Set<number>>(new Set());
  const showLayerBadges = ref(false);

  function levelText(node: AgentRecord) {
    return node.level || t('agentList.settings.layer', { n: node.layer });
  }

  const focusId = ref(ROOT_ID);
  /** 搜索定位路径：仅顶部提示，由用户自行点击路径进入 */
  const locatePath = ref<{ id: number; account: string }[]>([]);
  const locateLoading = ref(false);
  const parentExpanded = ref(false);
  const loading = ref(false);
  const total = ref(0);
  const tableRows = ref<AgentRecord[]>([]);
  const focusRecord = ref<AgentRecord | null>(null);
  const recordMap = ref<Record<number, AgentRecord>>({});
  const treeChildren = ref(new Map<number, AgentTreeNode[]>());
  const path = ref<{ id: number; account: string }[]>([]);
  const childRequests = new Map<number, Promise<void>>();
  let tableTicket = 0;
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
  const activeTab = ref('list');
  const reviewPending = ref(0);

  function remember(list: AgentRecord[]) {
    if (!list.length) return;
    recordMap.value = {
      ...recordMap.value,
      ...Object.fromEntries(list.map((item) => [item.id, item])),
    };
  }

  function findTreeNode(id: number) {
    return [...treeChildren.value.values()]
      .flat()
      .find((node) => node.id === id);
  }

  function decorate(node: AgentTreeNode, depth: number): AgentTreeNode {
    const children = treeChildren.value.get(node.id);
    return {
      ...node,
      depth,
      children: children?.map((child) => decorate(child, depth + 1)),
    };
  }

  const navTree = computed((): AgentTreeNode[] =>
    (treeChildren.value.get(ROOT_ID) || []).map((node) => decorate(node, 0))
  );

  const focusLoaded = computed(
    () => focusId.value === ROOT_ID || focusRecord.value?.id === focusId.value
  );

  const focusNode = computed(() => {
    if (focusId.value === ROOT_ID) return null;
    if (focusRecord.value?.id === focusId.value) return focusRecord.value;
    const known = recordMap.value[focusId.value] || findTreeNode(focusId.value);
    if (known) return known;
    const crumb = path.value.find((item) => item.id === focusId.value);
    return createEmptyAgent({
      id: focusId.value,
      account: crumb?.account || '',
    });
  });

  const focusLineCode = computed(
    () => findTreeNode(focusId.value)?.lineCode || ''
  );

  const focusLevel = computed(
    () => findTreeNode(focusId.value)?.layer || focusNode.value?.layer || 0
  );

  const isRootFocus = computed(() => focusId.value === ROOT_ID);
  const rootAgentIds = computed(
    () =>
      new Set((treeChildren.value.get(ROOT_ID) || []).map((node) => node.id))
  );
  /** 选中顶层代理（root）时不能新增：它不能作为新代理的上级 */
  const rootAgentFocus = computed(() => rootAgentIds.value.has(focusId.value));
  /** 选中「全部代理」或 root 代理时不能新增，必须先选一个具体的上级代理 */
  const createBlocked = computed(
    () => isRootFocus.value || rootAgentFocus.value
  );
  const createBlockedTip = computed(() =>
    isRootFocus.value
      ? t('agentList.toolbar.createAllBlocked')
      : t('agentList.toolbar.createRootBlocked')
  );
  const crumbs = computed(() => path.value);
  const pageData = computed(() => tableRows.value);

  /** 左侧树截断时仍要露出：当前节点 + 面包屑路径 + 已展开节点 */
  const treeRevealIds = computed(() => {
    const ids = new Set<number>(expandedIds.value);
    ids.add(focusId.value);
    path.value.forEach((item) => ids.add(item.id));
    return ids;
  });

  const detailRecord = computed(() => {
    if (detailId.value == null) return null;
    return recordMap.value[detailId.value] || null;
  });

  const editingRecord = computed(() => {
    if (editingId.value == null) return null;
    return recordMap.value[editingId.value] || null;
  });

  watch(focusId, () => {
    parentExpanded.value = false;
  });

  function toggleParent() {
    parentExpanded.value = !parentExpanded.value;
  }

  function snapshotQuery(): AgentQuery {
    return {
      ...form,
      level: form.level || '',
      status: form.status || '',
      dateRange: [...(form.dateRange || [])],
    };
  }

  function statusParam(status: AgentQuery['status']) {
    if (status === 'normal') return AGENT_STATUS.NORMAL;
    if (status === 'disabled') return AGENT_STATUS.DISABLED;
    return 0;
  }

  function buildListParams(
    page: number,
    pageSize: number
  ): AgentListParams | null {
    const level = Number(applied.value.level);
    const status = statusParam(applied.value.status);
    const params: AgentListParams = {
      page,
      page_size: pageSize,
    };
    if (level > 0) params.level = level;
    if (status) params.status = status;
    // 关键词搜索只走树定位，列表始终按当前 focus 的直属下级拉取
    if (focusId.value > 0) {
      params.agent_filter_type = AGENT_FILTER.PARENT;
      params.agent_keyword = String(focusId.value);
      return params;
    }
    return null;
  }

  function ensureChildren(parentId: number) {
    if (treeChildren.value.has(parentId)) return Promise.resolve();
    const pending = childRequests.get(parentId);
    if (pending) return pending;
    const task = listAgentTreeNodes(parentId > 0 ? parentId : undefined)
      .then(({ data }) => {
        const nodes = applyTreeStressPad(
          (data || []).map((item) => mapTreeNode(item, parentId)),
          parentId
        );
        treeChildren.value = new Map(treeChildren.value).set(parentId, nodes);
      })
      .finally(() => {
        childRequests.delete(parentId);
      });
    childRequests.set(parentId, task);
    return task;
  }

  async function fetchOne(id: number) {
    const { data } = await listAgents(
      {
        page: 1,
        page_size: 1,
        agent_filter_type: AGENT_FILTER.SELF,
        agent_keyword: String(id),
      },
      true
    );
    const item = (data?.list || []).find((row) => row.id === id);
    if (!item) return null;
    const record = mapAgentListItem(item);
    remember([record]);
    return record;
  }

  async function hydrate(ids: number[]) {
    const rows = await Promise.all(
      ids.map(async (id) => {
        try {
          return await fetchOne(id);
        } catch {
          return null;
        }
      })
    );
    return rows.filter((record): record is AgentRecord => Boolean(record));
  }

  async function loadRootTable(ticket: number) {
    await ensureChildren(ROOT_ID);
    if (ticket !== tableTicket) return;
    let nodes = [...(treeChildren.value.get(ROOT_ID) || [])];
    const level = Number(applied.value.level);
    if (level > 0) nodes = nodes.filter((node) => node.layer === level);
    const { status } = applied.value;
    if (status) {
      const detailed = await hydrate(nodes.map((node) => node.id));
      if (ticket !== tableTicket) return;
      const matched = detailed.filter((item) => item.status === status);
      total.value = matched.length;
      const start = (pagination.current - 1) * pagination.pageSize;
      tableRows.value = matched.slice(start, start + pagination.pageSize);
      return;
    }
    total.value = nodes.length;
    const start = (pagination.current - 1) * pagination.pageSize;
    const slice = nodes.slice(start, start + pagination.pageSize);
    const detailed = await hydrate(slice.map((node) => node.id));
    if (ticket !== tableTicket) return;
    const byId = new Map(detailed.map((item) => [item.id, item]));
    tableRows.value = slice.map((node) => byId.get(node.id) || node);
  }

  async function loadTable() {
    tableTicket += 1;
    const ticket = tableTicket;
    loading.value = true;
    try {
      const params = buildListParams(pagination.current, pagination.pageSize);
      if (!params) {
        await loadRootTable(ticket);
        return;
      }
      const { data } = await listAgents(params);
      if (ticket !== tableTicket) return;
      const rows = (data?.list || []).map(mapAgentListItem);
      tableRows.value = rows;
      total.value = data?.total || 0;
      remember(rows);
    } catch {
      if (ticket !== tableTicket) return;
      tableRows.value = [];
      total.value = 0;
    } finally {
      if (ticket === tableTicket) loading.value = false;
    }
  }

  async function loadFocus(id: number) {
    if (id === ROOT_ID) {
      focusRecord.value = null;
      return;
    }
    try {
      focusRecord.value = (await fetchOne(id)) || focusRecord.value;
    } catch {
      focusRecord.value = null;
    }
  }

  async function appendAncestor(
    current: number,
    guard: Set<number>,
    chain: { id: number; account: string }[]
  ) {
    if (current <= 0 || guard.has(current) || chain.length >= 12) return;
    guard.add(current);
    let record: AgentRecord | undefined =
      recordMap.value[current] || findTreeNode(current);
    if (!record) {
      try {
        record = (await fetchOne(current)) || undefined;
      } catch {
        return;
      }
    }
    if (!record) return;
    chain.unshift({ id: record.id, account: record.account });
    await appendAncestor(record.parentId, guard, chain);
  }

  async function syncPath(id: number) {
    if (id === ROOT_ID) {
      path.value = [];
      return;
    }
    const chain: { id: number; account: string }[] = [];
    await appendAncestor(id, new Set<number>(), chain);
    path.value = chain;
  }

  function clearLocate() {
    locatePath.value = [];
  }

  /** 接口返回当前代理在前、顶层在后，路径展示需要从顶层到当前代理 */
  function normalizeLocatePath(list: AgentTreeAncestorItem[] = []) {
    return list
      .filter((item) => item.id > 0 && item.account)
      .map((item) => ({ id: item.id, account: item.account }))
      .reverse();
  }

  async function fetchLocatePath(keyword: string) {
    const { data } = await listAgentTreeAncestors(keyword);
    return normalizeLocatePath(Array.isArray(data) ? data : []);
  }

  async function handleSearch() {
    if (locateLoading.value) return;
    const keyword = form.keyword.trim();
    if (!keyword) {
      Message.warning(t('agentList.message.locateEmpty'));
      return;
    }
    locateLoading.value = true;
    try {
      const chain = await fetchLocatePath(keyword);
      if (!chain.length) {
        clearLocate();
        Message.warning(t('agentList.message.locateNotFound'));
        return;
      }
      locatePath.value = chain;
    } catch {
      clearLocate();
    } finally {
      locateLoading.value = false;
    }
  }

  function handleReset() {
    Object.assign(form, createAgentQuery());
    applied.value = snapshotQuery();
    clearLocate();
  }

  function onPageChange(current: number) {
    pagination.current = current;
    loadTable();
  }

  function onPageSizeChange(size: number) {
    pagination.pageSize = size;
    pagination.current = 1;
    loadTable();
  }

  async function toggleExpanded(id: number) {
    const next = new Set(expandedIds.value);
    if (next.has(id)) {
      next.delete(id);
      expandedIds.value = next;
      return;
    }
    try {
      await ensureChildren(id);
    } catch {
      return;
    }
    next.add(id);
    expandedIds.value = next;
  }

  async function openLevel(
    id: number,
    knownPath?: { id: number; account: string }[]
  ) {
    focusId.value = id;
    pagination.current = 1;
    const next = new Set(expandedIds.value);
    next.add(id);
    expandedIds.value = next;
    if (knownPath) path.value = knownPath;
    await Promise.all([
      knownPath ? undefined : syncPath(id),
      loadFocus(id),
      loadTable(),
    ]);
    const opened = new Set(expandedIds.value);
    path.value.forEach((item) => opened.add(item.id));
    expandedIds.value = opened;
    await Promise.all(
      [ROOT_ID, ...path.value.map((item) => item.id)].map((itemId) =>
        ensureChildren(itemId).catch(() => undefined)
      )
    );
  }

  async function enterLevel(id: number) {
    if (form.keyword || applied.value.keyword) {
      form.keyword = '';
      applied.value = { ...applied.value, keyword: '' };
    }
    await openLevel(id);
  }

  async function openTop() {
    try {
      await ensureChildren(ROOT_ID);
    } catch {
      await enterLevel(ROOT_ID);
      return;
    }
    const nodes = treeChildren.value.get(ROOT_ID) || [];
    if (nodes.length === 1) {
      await enterLevel(nodes[0].id);
      return;
    }
    await enterLevel(ROOT_ID);
  }

  function goRoot() {
    openTop();
  }

  function goBack() {
    const chain = crumbs.value;
    if (chain.length > 1) {
      enterLevel(chain[chain.length - 2].id);
      return;
    }
    openTop();
  }

  function patchRecord(id: number, patch: Partial<AgentRecord>) {
    const current = recordMap.value[id];
    if (!current) return;
    const next = { ...current, ...patch };
    remember([next]);
    tableRows.value = tableRows.value.map((item) =>
      item.id === id ? next : item
    );
    if (focusRecord.value?.id === id) focusRecord.value = next;
  }

  function applyStatus(ids: number[], status: AccountStatus) {
    const idSet = new Set(ids);
    let count = 0;
    ids.forEach((id) => {
      const current = recordMap.value[id];
      if (!current || !idSet.has(id) || current.status === status) return;
      count += 1;
      patchRecord(id, { status });
    });
    return count;
  }

  function confirmToggle(record: AgentRecord) {
    if (record.id === ROOT_ID) return;
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
    if (record.id === ROOT_ID) return;
    detailId.value = record.id;
    detailVisible.value = true;
  }

  function onQueryView(record: AgentRecord) {
    remember([record]);
    openDetail(record);
  }

  function viewParent() {
    if (!focusNode.value) return;
    openDetail(focusNode.value);
  }

  function onParentMore(
    value: string | number | Record<string, unknown> | undefined
  ) {
    if (!focusNode.value) return;
    if (value === 'edit') openEdit(focusNode.value);
    if (value === 'toggle') confirmToggle(focusNode.value);
  }

  function openEdit(record: AgentRecord) {
    if (record.id === ROOT_ID) return;
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

  function dropChildren(parentId: number) {
    if (!treeChildren.value.has(parentId)) return;
    const next = new Map(treeChildren.value);
    next.delete(parentId);
    treeChildren.value = next;
  }

  async function onAgentCreated() {
    const parentId = focusId.value;
    const upperId = focusNode.value?.parentId ?? ROOT_ID;
    dropChildren(parentId);
    if (upperId !== parentId) dropChildren(upperId);
    await Promise.all([
      ensureChildren(upperId).catch(() => undefined),
      ensureChildren(parentId).catch(() => undefined),
      loadTable(),
    ]);
  }

  function onSubmit(value: AgentFormValue) {
    if (editingId.value == null) return;
    patchRecord(editingId.value, {
      ...value,
      layer: levelLayer(value.level),
    });
    Message.success(t('agentList.message.updated'));
  }

  function csvCell(value: string | number) {
    const text = String(value ?? '');
    if (/[",\n]/.test(text)) return `"${text.replace(/"/g, '""')}"`;
    return text;
  }

  async function collectExportRows() {
    const params = buildListParams(1, 100);
    if (params) {
      const first = await listAgents({ ...params, page: 1, page_size: 100 });
      const items = [...(first.data?.list || [])];
      const count = first.data?.total || items.length;
      const pages = Math.ceil(count / 100);
      const rest = await Promise.all(
        Array.from({ length: Math.max(0, pages - 1) }, (_, index) =>
          listAgents({ ...params, page: index + 2, page_size: 100 })
        )
      );
      rest.forEach((res) => {
        items.push(...(res.data?.list || []));
      });
      return items.map(mapAgentListItem);
    }
    await ensureChildren(ROOT_ID);
    let nodes = [...(treeChildren.value.get(ROOT_ID) || [])];
    const level = Number(applied.value.level);
    if (level > 0) nodes = nodes.filter((node) => node.layer === level);
    const detailed = await hydrate(nodes.map((node) => node.id));
    if (!applied.value.status) {
      const byId = new Map(detailed.map((item) => [item.id, item]));
      return nodes.map((node) => byId.get(node.id) || node);
    }
    return detailed.filter((item) => item.status === applied.value.status);
  }

  async function exportCsv() {
    let source: AgentRecord[] = [];
    try {
      source = await collectExportRows();
    } catch {
      return;
    }
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
    const rows = source.map((item) => [
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
      buildPromoLink(item.promoHost, item.promoCode),
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

  onMounted(() => {
    expandedIds.value = new Set();
    openTop();
  });
</script>

<style scoped lang="less">
  .container {
    padding: 0 20px 20px;
  }

  .general-card {
    margin-top: 16px;
  }

  .page-tabs {
    flex-shrink: 0;
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
    background: var(--color-fill-2);
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

  .create-blocked {
    display: inline-flex;
    cursor: not-allowed;
  }

  .keyword-input {
    flex: 0 0 240px;
    width: 240px;
  }

  .filter-bar :deep(.arco-input-wrapper),
  .filter-bar :deep(.arco-select-view-single),
  .filter-bar :deep(.arco-btn-secondary:not(.arco-btn-disabled)) {
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
  }

  .filter-bar :deep(.arco-input-wrapper:hover),
  .filter-bar :deep(.arco-select-view-single:hover),
  .filter-bar :deep(.arco-input-wrapper.arco-input-focus),
  .filter-bar :deep(.arco-select-view-single.arco-select-view-focus),
  .filter-bar :deep(.arco-btn-secondary:not(.arco-btn-disabled):hover) {
    background: var(--color-bg-2);
    border-color: rgb(var(--primary-6));
  }

  .filter-main :deep(.arco-select-view-single) {
    flex: 0 0 148px;
    width: 148px;
    max-width: 148px;
  }

  .locate-hint {
    display: flex;
    flex-shrink: 0;
    gap: 10px;
    align-items: center;
    margin: -4px 0 12px;
    padding: 8px 12px;
    background: rgb(var(--orange-1));
    border: 1px solid rgb(var(--orange-3));
    border-radius: 6px;
  }

  .locate-hint__label {
    flex: none;
    color: rgb(var(--orange-6));
    font-size: 13px;
    white-space: nowrap;
  }

  .locate-hint__path {
    flex: 1;
    min-width: 0;
  }

  .locate-hint__path :deep(.arco-breadcrumb-item) {
    color: var(--color-text-2);
  }

  .locate-hint__link {
    color: inherit;
    cursor: pointer;
  }

  .locate-hint__link:hover {
    color: rgb(var(--orange-6));
  }

  .locate-hint__path .is-target {
    color: rgb(var(--orange-6));
    font-weight: 600;
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

  .nav-pane {
    display: flex;
    flex: none;
    flex-direction: column;
    width: 300px;
    min-height: 0;
    padding: 8px;
    overflow: hidden;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 10px;
  }

  .layer-control {
    flex-shrink: 0;
    margin-bottom: 8px;
    padding: 10px;
    background: var(--color-fill-2);
    border-radius: 8px;
  }

  .layer-control__head {
    display: flex;
    gap: 8px;
    align-items: center;
    justify-content: space-between;
  }

  .layer-control__title {
    display: inline-flex;
    gap: 4px;
    align-items: center;
    color: var(--color-text-1);
    font-weight: 500;
    font-size: 13px;
  }

  .layer-control__title :deep(.arco-icon) {
    color: rgb(var(--primary-6));
  }

  .layer-control__actions {
    display: inline-flex;
    gap: 6px;
    align-items: center;
  }

  .layer-control__actions button {
    padding: 0;
    color: rgb(var(--primary-6));
    font-size: 12px;
    background: transparent;
    border: none;
    cursor: pointer;
  }

  .layer-control__actions button.is-active {
    color: var(--color-text-3);
    cursor: default;
  }

  .layer-control__divider {
    width: 1px;
    height: 10px;
    background: var(--color-border-3);
  }

  .nav-tree-wrap {
    flex: 1;
    min-height: 0;
    overflow: auto;
  }

  .level-main {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 12px;
    min-height: 0;
    overflow: hidden;
    background: transparent;
    border: none;
    --agent-cols: 36px minmax(140px, 1.2fr) minmax(64px, 0.55fr) minmax(64px, 0.55fr)
      minmax(64px, 0.55fr) minmax(64px, 0.55fr) minmax(88px, 0.8fr)
      minmax(88px, 0.8fr) minmax(56px, 0.5fr) minmax(96px, 0.85fr);
  }

  .parent-panel,
  .child-panel {
    display: flex;
    flex-direction: column;
    min-height: 0;
    overflow: hidden;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 10px;
  }

  .parent-panel {
    flex: none;
    max-height: 42%;
  }

  .parent-panel-body {
    min-height: 0;
    overflow: auto;
  }

  .child-panel {
    flex: 1;
  }

  .child-panel-body {
    flex: 1;
    min-height: 0;
    overflow: auto;
    padding-bottom: 12px;
  }

  .child-panel :deep(.level-head) {
    background: var(--color-fill-2);
    border-bottom: 1px solid var(--color-border-2);
  }

  .level-spin {
    display: block;
    width: 100%;
  }

  .parent-head,
  .parent-bar {
    display: grid;
    grid-template-columns: var(--agent-cols);
    gap: 8px;
    align-items: center;
    padding: 0 16px;
  }

  .parent-head {
    height: 40px;
    color: var(--color-text-3);
    font-size: 13px;
    background: var(--color-fill-2);
    border-bottom: 1px solid var(--color-border-2);
  }

  .parent-bar {
    min-height: 52px;
    padding: 6px 16px;
    color: var(--color-text-1);
    font-size: 14px;
    background: var(--color-bg-2);
    border-bottom: 1px solid var(--color-border-1);
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

  .parent-main {
    display: flex;
    gap: 10px;
    align-items: center;
    min-width: 0;
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

  .level-orb.is-root {
    width: auto;
    min-width: 36px;
    padding: 0 8px;
    color: #fff;
    font-weight: 800;
    font-size: 12px;
    background: #86909c;
    border-radius: 18px;
  }

  .parent-id {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .parent-name {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .parent-id strong {
    color: var(--color-text-1);
    font-size: 14px;
  }

  .parent-id strong.is-root,
  .is-root-text {
    color: var(--color-text-1);
    font-weight: 800;
  }

  .parent-id small {
    color: var(--color-text-3);
    font-size: 12px;
  }

  .top-tag {
    padding: 0 8px;
    color: rgb(var(--green-6));
    font-size: 12px;
    font-style: normal;
    line-height: 22px;
    background: rgb(var(--green-1));
    border-radius: 4px;
  }

  .parent-state {
    font-size: 14px;
  }

  .parent-state.normal {
    color: rgb(var(--green-6));
  }

  .parent-state.disabled {
    color: rgb(var(--red-6));
  }

  .parent-ops {
    display: flex;
    gap: 4px;
  }

  .section-title {
    display: flex;
    flex-shrink: 0;
    gap: 6px;
    align-items: center;
    margin: 0;
    padding: 12px 16px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 14px;
    background: rgb(var(--primary-1));
    border-bottom: 1px solid rgb(var(--primary-2));
  }

  .section-title em {
    color: var(--color-text-3);
    font-weight: 400;
    font-style: normal;
  }

  .table-pagination {
    display: flex;
    justify-content: flex-end;
    padding: 16px 16px 0;
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
