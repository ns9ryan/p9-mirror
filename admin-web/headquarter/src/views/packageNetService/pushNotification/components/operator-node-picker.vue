<template>
  <div class="picker">
    <div class="picker-toolbar">
      <a-input
        v-model="keyword"
        :placeholder="t('pushNotification.picker.searchPlaceholder')"
        allow-clear
        :max-length="100"
        class="picker-search"
      >
        <template #prefix><icon-search /></template>
      </a-input>
      <div class="picker-actions">
        <label class="only-selected">
          <a-switch
            v-model="onlySelected"
            size="small"
            :disabled="!selectedIds.size"
          />
          {{ t('pushNotification.picker.onlySelected') }}
        </label>
        <a-divider direction="vertical" class="picker-divider" />
        <span class="picker-count">
          {{
            t('pushNotification.picker.selected', {
              selected: selectedIds.size,
              total: operators.length,
            })
          }}
        </span>
        <a-button
          size="small"
          :disabled="!visibleOperators.length"
          @click="selectVisible"
        >
          {{ t('pushNotification.picker.selectAll') }}
        </a-button>
        <a-button size="small" :disabled="!selectedIds.size" @click="clearAll">
          {{ t('pushNotification.picker.clear') }}
        </a-button>
        <a-tooltip :content="t('front.common.toolbar.refresh')">
          <a-button size="small" :loading="loading" @click="loadData(true)">
            <template #icon><icon-refresh /></template>
          </a-button>
        </a-tooltip>
      </div>
    </div>

    <a-spin :loading="loading" class="picker-spin">
      <div class="picker-body">
        <div class="node-nav">
          <div
            class="node-nav-item"
            :class="{ 'node-nav-item--active': activeNodeKey === ALL_KEY }"
            @click="activeNodeKey = ALL_KEY"
          >
            <span class="node-nav-name">
              {{ t('pushNotification.picker.allNodes') }}
            </span>
            <span class="node-nav-count">
              {{ countText(operators, filteredOperators) }}
            </span>
          </div>
          <div
            v-for="group in groups"
            :key="group.key"
            class="node-nav-item"
            :class="{
              'node-nav-item--active': activeNodeKey === group.key,
              'node-nav-item--muted':
                !matchedByGroup(group).length && filtering,
            }"
            :title="group.code || group.name"
            @click="activeNodeKey = group.key"
          >
            <span class="status-dot" :class="nodeDotClass(group)" />
            <span class="node-nav-name">{{ group.name }}</span>
            <span class="node-nav-count">
              {{ countText(group.operators, matchedByGroup(group)) }}
            </span>
          </div>
        </div>

        <div class="site-pane">
          <template v-for="group in visibleGroups" :key="group.key">
            <div class="group-header">
              <a-checkbox
                :model-value="groupCheckState(group).checked"
                :indeterminate="groupCheckState(group).indeterminate"
                :disabled="!matchedByGroup(group).length"
                @change="(value) => toggleGroup(group, Boolean(value))"
              >
                <span class="group-name">{{ group.name }}</span>
              </a-checkbox>
              <span class="group-meta">
                <span class="status-dot" :class="nodeDotClass(group)" />
                {{ nodeStatusText(group) }}
              </span>
              <span class="group-count">
                {{ groupCheckState(group).selected }} /
                {{ matchedByGroup(group).length }}
              </span>
            </div>
            <div v-if="matchedByGroup(group).length" class="site-grid">
              <div
                v-for="item in matchedByGroup(group)"
                :key="item.id"
                class="site-item"
                :class="{ 'site-item--checked': selectedIds.has(item.id) }"
                :title="`${item.name}（${item.code}）`"
                @click="toggleOperator(item.id)"
              >
                <a-checkbox
                  :model-value="selectedIds.has(item.id)"
                  class="site-checkbox"
                  @click.stop
                  @change="toggleOperator(item.id)"
                />
                <span class="site-name">
                  <template
                    v-for="(part, index) in highlightParts(item.name)"
                    :key="index"
                  >
                    <mark v-if="part.match" class="site-mark">{{
                      part.text
                    }}</mark>
                    <template v-else>{{ part.text }}</template>
                  </template>
                </span>
              </div>
            </div>
            <div v-else class="group-empty">
              {{ t('pushNotification.picker.emptyNode') }}
            </div>
          </template>
          <a-empty
            v-if="!visibleGroups.length"
            class="site-empty"
            :description="
              loadFailed
                ? t('pushNotification.picker.loadFailed')
                : t('pushNotification.picker.emptySearch')
            "
          />
        </div>
      </div>
    </a-spin>
  </div>
</template>

<script lang="ts" setup>
  import { computed, onMounted, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { NODE_STATUS, queryNodeList, type NodeRecord } from '@/api/node';
  import {
    getOperatorNode,
    queryOperatorList,
    type OperatorRecord,
  } from '@/api/operator';

  interface PickerOperator {
    id: number;
    code: string;
    name: string;
  }

  interface NodeGroup {
    key: string;
    code: string;
    name: string;
    online: boolean;
    enabled: boolean;
    unassigned: boolean;
    operators: PickerOperator[];
  }

  const props = defineProps<{
    modelValue: number[];
  }>();

  const emit = defineEmits<{
    (e: 'update:modelValue', value: number[]): void;
  }>();

  const { t } = useI18n();

  const ALL_KEY = '__all__';
  const UNASSIGNED_KEY = '__unassigned__';
  const PAGE_SIZE = 100;
  const NODE_QUERY_CONCURRENCY = 8;

  const loading = ref(false);
  const loadFailed = ref(false);
  const keyword = ref('');
  const activeNodeKey = ref(ALL_KEY);
  const operators = ref<PickerOperator[]>([]);
  const groups = ref<NodeGroup[]>([]);

  const selectedIds = computed(() => new Set(props.modelValue));

  const onlySelected = ref(false);

  const normalizedKeyword = computed(() => keyword.value.trim().toLowerCase());

  const filtering = computed(
    () => Boolean(normalizedKeyword.value) || onlySelected.value
  );

  const matchOperator = (item: PickerOperator) => {
    if (onlySelected.value && !selectedIds.value.has(item.id)) return false;
    const kw = normalizedKeyword.value;
    if (!kw) return true;
    return (
      item.name.toLowerCase().includes(kw) ||
      item.code.toLowerCase().includes(kw)
    );
  };

  const filteredOperators = computed(() =>
    operators.value.filter(matchOperator)
  );

  const matchedCache = computed(() => {
    const map = new Map<string, PickerOperator[]>();
    groups.value.forEach((group) => {
      map.set(group.key, group.operators.filter(matchOperator));
    });
    return map;
  });

  const matchedByGroup = (group: NodeGroup) =>
    matchedCache.value.get(group.key) || [];

  const visibleGroups = computed(() => {
    const scoped =
      activeNodeKey.value === ALL_KEY
        ? groups.value
        : groups.value.filter((group) => group.key === activeNodeKey.value);
    if (activeNodeKey.value !== ALL_KEY) return scoped;
    return scoped.filter((group) =>
      filtering.value
        ? matchedByGroup(group).length > 0
        : group.operators.length > 0
    );
  });

  const visibleOperators = computed(() =>
    visibleGroups.value.flatMap((group) => matchedByGroup(group))
  );

  // 搜索时切回全部节点，避免关键字命中的分站因所选节点不同而被隐藏
  watch(normalizedKeyword, (value) => {
    if (value) activeNodeKey.value = ALL_KEY;
  });

  watch(
    () => selectedIds.value.size,
    (size) => {
      if (!size) onlySelected.value = false;
    }
  );

  const countText = (all: PickerOperator[], matched: PickerOperator[]) => {
    if (filtering.value) return `${matched.length}`;
    const selected = all.filter((item) => selectedIds.value.has(item.id));
    return selected.length
      ? `${selected.length}/${all.length}`
      : `${all.length}`;
  };

  const groupCheckState = (group: NodeGroup) => {
    const matched = matchedByGroup(group);
    const selected = matched.filter((item) =>
      selectedIds.value.has(item.id)
    ).length;
    return {
      selected,
      checked: matched.length > 0 && selected === matched.length,
      indeterminate: selected > 0 && selected < matched.length,
    };
  };

  const nodeDotClass = (group: NodeGroup) => {
    if (group.unassigned || !group.enabled) return 'status-dot--muted';
    return group.online ? 'status-dot--online' : 'status-dot--offline';
  };

  const nodeStatusText = (group: NodeGroup) => {
    if (group.unassigned) return t('pushNotification.picker.unassigned');
    if (!group.enabled) return t('pushNotification.picker.disabled');
    return group.online
      ? t('pushNotification.picker.online')
      : t('pushNotification.picker.offline');
  };

  const highlightParts = (text: string) => {
    const kw = normalizedKeyword.value;
    if (!kw) return [{ text, match: false }];
    const index = text.toLowerCase().indexOf(kw);
    if (index < 0) return [{ text, match: false }];
    return [
      { text: text.slice(0, index), match: false },
      { text: text.slice(index, index + kw.length), match: true },
      { text: text.slice(index + kw.length), match: false },
    ].filter((part) => part.text);
  };

  const emitSelection = (ids: Set<number>) => {
    emit('update:modelValue', [...ids]);
  };

  const toggleOperator = (id: number) => {
    const next = new Set(selectedIds.value);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    emitSelection(next);
  };

  const toggleGroup = (group: NodeGroup, checked: boolean) => {
    const next = new Set(selectedIds.value);
    matchedByGroup(group).forEach((item) => {
      if (checked) next.add(item.id);
      else next.delete(item.id);
    });
    emitSelection(next);
  };

  const selectVisible = () => {
    const next = new Set(selectedIds.value);
    visibleOperators.value.forEach((item) => next.add(item.id));
    emitSelection(next);
  };

  const clearAll = () => {
    emitSelection(new Set());
  };

  async function fetchAll<T>(
    fetchPage: (page: number) => Promise<{ list: T[]; total: number }>
  ) {
    const result: T[] = [];
    let page = 1;
    // eslint-disable-next-line no-constant-condition
    while (true) {
      // eslint-disable-next-line no-await-in-loop
      const { list, total } = await fetchPage(page);
      result.push(...list);
      if (!list.length || result.length >= total) break;
      page += 1;
    }
    return result;
  }

  async function mapWithConcurrency<T, R>(
    items: T[],
    limit: number,
    worker: (item: T) => Promise<R>
  ) {
    const results: R[] = new Array(items.length);
    let cursor = 0;
    const run = async () => {
      while (cursor < items.length) {
        const index = cursor;
        cursor += 1;
        // eslint-disable-next-line no-await-in-loop
        results[index] = await worker(items[index]);
      }
    };
    await Promise.all(
      Array.from({ length: Math.min(limit, items.length) }, run)
    );
    return results;
  }

  const buildGroups = (
    nodes: NodeRecord[],
    operatorNodes: Array<{
      operator: OperatorRecord;
      nodeCode: string;
      nodeName: string;
    }>
  ) => {
    const groupMap = new Map<string, NodeGroup>();
    nodes.forEach((node) => {
      groupMap.set(node.code, {
        key: node.code,
        code: node.code,
        name: node.name || node.code,
        online: node.online,
        enabled: node.status === NODE_STATUS.ENABLED,
        unassigned: false,
        operators: [],
      });
    });
    const unassigned: NodeGroup = {
      key: UNASSIGNED_KEY,
      code: '',
      name: t('pushNotification.picker.unassigned'),
      online: false,
      enabled: false,
      unassigned: true,
      operators: [],
    };
    operatorNodes.forEach(({ operator, nodeCode, nodeName }) => {
      const item: PickerOperator = {
        id: operator.id,
        code: operator.code || '',
        name: operator.name || operator.code || `#${operator.id}`,
      };
      if (!nodeCode) {
        unassigned.operators.push(item);
        return;
      }
      let group = groupMap.get(nodeCode);
      if (!group) {
        group = {
          key: nodeCode,
          code: nodeCode,
          name: nodeName || nodeCode,
          online: false,
          enabled: true,
          unassigned: false,
          operators: [],
        };
        groupMap.set(nodeCode, group);
      }
      group.operators.push(item);
    });
    const result = [...groupMap.values()];
    if (unassigned.operators.length) result.push(unassigned);
    return result;
  };

  const loadData = async (force = false) => {
    if (force) operatorNodeCache.clear();
    loading.value = true;
    loadFailed.value = false;
    try {
      const [nodes, operatorList] = await Promise.all([
        fetchAll<NodeRecord>(async (page) => {
          const { data } = await queryNodeList({ page, page_size: PAGE_SIZE });
          return { list: data?.list || [], total: data?.total ?? 0 };
        }),
        fetchAll<OperatorRecord>(async (page) => {
          const { data } = await queryOperatorList({
            page,
            page_size: PAGE_SIZE,
          });
          return { list: data?.list || [], total: data?.total ?? 0 };
        }),
      ]);
      const operatorNodes = await mapWithConcurrency(
        operatorList,
        NODE_QUERY_CONCURRENCY,
        async (operator) => {
          if ('node_code' in operator) {
            return {
              operator,
              nodeCode: operator.node_code || '',
              nodeName: operator.node_name || '',
            };
          }
          const cached = operatorNodeCache.get(operator.id);
          if (cached) return { operator, ...cached };
          try {
            const { data } = await getOperatorNode({
              operator_id: operator.id,
            });
            const node = data?.operator_node;
            const resolved = {
              nodeCode: node?.node_code || '',
              nodeName: node?.node_name || '',
            };
            operatorNodeCache.set(operator.id, resolved);
            return { operator, ...resolved };
          } catch {
            return { operator, nodeCode: '', nodeName: '' };
          }
        }
      );
      groups.value = buildGroups(nodes, operatorNodes);
      operators.value = groups.value.flatMap((group) => group.operators);
      if (
        activeNodeKey.value !== ALL_KEY &&
        !groups.value.some((group) => group.key === activeNodeKey.value)
      ) {
        activeNodeKey.value = ALL_KEY;
      }
    } catch {
      loadFailed.value = true;
      groups.value = [];
      operators.value = [];
    } finally {
      loading.value = false;
    }
  };

  onMounted(() => loadData());
</script>

<script lang="ts">
  // 跨弹窗打开复用，点击刷新时清空
  const operatorNodeCache = new Map<
    number,
    { nodeCode: string; nodeName: string }
  >();
</script>

<style scoped lang="less">
  .picker {
    display: flex;
    flex-direction: column;
    width: 100%;
    overflow: hidden;
    border: 1px solid var(--color-border-2);
    border-radius: 8px;
  }

  .picker-toolbar {
    display: flex;
    gap: 12px;
    align-items: center;
    justify-content: space-between;
    padding: 10px 12px;
    background-color: var(--color-fill-1);
    border-bottom: 1px solid var(--color-border-2);
  }

  .picker-search {
    flex: 1;
    max-width: 280px;
  }

  .picker-actions {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .only-selected {
    display: inline-flex;
    gap: 6px;
    align-items: center;
    color: var(--color-text-2);
    font-size: 13px;
    white-space: nowrap;
    cursor: pointer;
  }

  .picker-divider {
    margin: 0 4px;
  }

  .picker-count {
    margin-right: 4px;
    color: var(--color-text-3);
    font-size: 13px;
    white-space: nowrap;
  }

  .picker-spin {
    display: block;
  }

  .picker-body {
    display: flex;
    height: 340px;
  }

  .node-nav {
    flex-shrink: 0;
    width: 190px;
    padding: 6px;
    overflow-y: auto;
    border-right: 1px solid var(--color-border-2);
  }

  .node-nav-item {
    display: flex;
    gap: 8px;
    align-items: center;
    height: 34px;
    padding: 0 10px;
    color: var(--color-text-2);
    font-size: 13px;
    border-radius: 6px;
    cursor: pointer;
    transition: background-color 0.15s;

    &:hover {
      background-color: var(--color-fill-2);
    }

    &--active {
      color: rgb(var(--primary-6));
      font-weight: 500;
      background-color: rgb(var(--primary-1));

      &:hover {
        background-color: rgb(var(--primary-1));
      }
    }

    &--muted {
      opacity: 0.45;
    }
  }

  .node-nav-name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .node-nav-count {
    flex-shrink: 0;
    color: var(--color-text-3);
    font-size: 12px;
  }

  .site-pane {
    flex: 1;
    min-width: 0;
    padding: 4px 14px 12px;
    overflow-y: auto;
  }

  .group-header {
    position: sticky;
    top: 0;
    z-index: 1;
    display: flex;
    gap: 10px;
    align-items: center;
    padding: 10px 0 8px;
    background-color: var(--color-bg-2);
    border-bottom: 1px dashed var(--color-border-2);
  }

  .group-name {
    color: var(--color-text-1);
    font-weight: 600;
  }

  .group-meta {
    display: inline-flex;
    gap: 6px;
    align-items: center;
    color: var(--color-text-3);
    font-size: 12px;
  }

  .group-count {
    margin-left: auto;
    color: var(--color-text-3);
    font-size: 12px;
  }

  .group-empty {
    padding: 12px 0;
    color: var(--color-text-3);
    font-size: 13px;
  }

  .site-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    gap: 8px;
    padding: 10px 0 6px;
  }

  .site-item {
    display: flex;
    gap: 6px;
    align-items: center;
    min-width: 0;
    height: 32px;
    padding: 0 10px;
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
    cursor: pointer;
    transition: border-color 0.15s, background-color 0.15s;

    &:hover {
      border-color: rgb(var(--primary-4));
    }

    &--checked {
      background-color: rgb(var(--primary-1));
      border-color: rgb(var(--primary-5));
    }
  }

  .site-checkbox {
    flex-shrink: 0;
  }

  .site-name {
    overflow: hidden;
    color: var(--color-text-1);
    font-size: 13px;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .site-mark {
    padding: 0;
    color: rgb(var(--primary-6));
    font-weight: 600;
    background: transparent;
  }

  .site-empty {
    padding-top: 80px;
  }

  .status-dot {
    display: inline-block;
    flex-shrink: 0;
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }

  .status-dot--online {
    background-color: rgb(var(--success-6));
  }

  .status-dot--offline {
    background-color: rgb(var(--danger-6));
  }

  .status-dot--muted {
    background-color: rgb(var(--gray-5));
  }
</style>
