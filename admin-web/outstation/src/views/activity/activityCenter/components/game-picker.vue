<template>
  <div class="game-picker">
    <div v-if="!categories.length" class="picker-empty">
      暂无可选游戏，请先在游戏管理中启用游戏
    </div>

    <template v-else>
      <div class="category-tabs">
        <button
          v-for="category in categories"
          :key="category"
          type="button"
          class="category-tab"
          :class="{ active: activeCategory === category }"
          @click="activeCategory = category"
        >
          {{ category }}
          <span v-if="categorySelectedCount(category)" class="tab-count">
            {{ categorySelectedCount(category) }}
          </span>
        </button>
      </div>

      <div class="picker-body">
        <!-- 左栏：厂商，数量多时内部滚动 -->
        <aside class="provider-pane">
          <a-input
            v-model="providerKeyword"
            size="small"
            allow-clear
            placeholder="搜索平台"
            class="pane-search"
          >
            <template #prefix><icon-search /></template>
          </a-input>
          <div class="provider-list">
            <div
              v-for="provider in filteredProviders"
              :key="provider.name"
              class="provider-item"
              :class="{
                'active': !gameKeyword && activeProvider === provider.name,
                'has-selected': selectedCount(provider.games) > 0,
              }"
              @click="selectProvider(provider.name)"
            >
              <a-checkbox
                :model-value="isAllSelected(provider.games)"
                :indeterminate="isPartSelected(provider.games)"
                @click.stop
                @change="(value) => setSelected(provider.games, !!value)"
              />
              <span class="provider-name" :title="provider.name">
                {{ provider.name }}
              </span>
              <span class="provider-count">
                {{ selectedCount(provider.games) }}/{{ provider.games.length }}
              </span>
            </div>
            <div v-if="!filteredProviders.length" class="pane-empty">
              没有匹配的平台
            </div>
          </div>
        </aside>

        <!-- 右栏：当前平台的游戏；输入游戏名时跨平台搜索当前分类 -->
        <section class="game-pane">
          <div class="game-toolbar">
            <a-input
              v-model="gameKeyword"
              size="small"
              allow-clear
              :placeholder="`在「${activeCategory}」中搜索游戏`"
              class="pane-search game-search"
            >
              <template #prefix><icon-search /></template>
            </a-input>
            <a-button
              type="text"
              size="mini"
              class="select-all"
              :disabled="!visibleGames.length"
              @click="setSelected(visibleGames, !isAllSelected(visibleGames))"
            >
              {{ isAllSelected(visibleGames) ? '取消全选' : '全选' }}
            </a-button>
          </div>
          <div class="game-scope">
            {{
              gameKeyword
                ? `搜索结果 ${visibleGames.length} 个`
                : `${activeProvider} · ${visibleGames.length} 个游戏`
            }}
          </div>
          <div class="game-grid">
            <label
              v-for="game in visibleGames"
              :key="game.id"
              class="game-option"
              :class="{ checked: selected.has(game.id) }"
            >
              <a-checkbox
                :model-value="selected.has(game.id)"
                @change="(value) => setSelected([game], !!value)"
              />
              <span class="game-name" :title="game.name">{{ game.name }}</span>
              <small v-if="gameKeyword" class="game-provider">
                {{ game.provider }}
              </small>
            </label>
            <div v-if="!visibleGames.length" class="pane-empty">
              没有匹配的游戏
            </div>
          </div>
        </section>
      </div>

      <div class="picker-footer">
        <span class="footer-summary">
          已选 <b>{{ selectedProviderCount }}</b> 个平台，<b>{{
            modelValue.length
          }}</b>
          个游戏
        </span>
        <div v-if="modelValue.length" class="selected-tags">
          <a-tag
            v-for="game in selectedPreview"
            :key="game.id"
            size="small"
            closable
            @close="setSelected([game], false)"
          >
            {{ game.name }}
          </a-tag>
          <span v-if="modelValue.length > PREVIEW_LIMIT" class="more-hint">
            等 {{ modelValue.length }} 个
          </span>
          <a-link class="clear-all" @click="emit('update:modelValue', [])">
            清空
          </a-link>
        </div>
      </div>
    </template>
  </div>
</template>

<script lang="ts" setup>
  import { computed, ref, watch } from 'vue';
  import { useActivityStore } from '../store';

  interface PickerGame {
    id: number;
    name: string;
    provider: string;
  }

  interface ProviderGroup {
    name: string;
    games: PickerGame[];
  }

  const props = defineProps<{
    /** 选中的游戏 ID */
    modelValue: number[];
  }>();

  const emit = defineEmits<{
    (e: 'update:modelValue', value: number[]): void;
  }>();

  /** 底部最多展示的已选标签数，避免选了一整类时撑出很高 */
  const PREVIEW_LIMIT = 20;

  const { games, loadGames } = useActivityStore();
  loadGames();

  /** 分类 → 厂商 → 游戏，保持接口返回顺序（按 sort_no） */
  const grouped = computed(() => {
    const result = new Map<string, Map<string, PickerGame[]>>();
    games.value.forEach((game) => {
      const category =
        game.category_name_i18n ||
        game.category_name ||
        game.category_code ||
        '其他';
      const provider =
        game.provider_name_i18n ||
        game.provider_name ||
        game.provider_code ||
        '其他';
      const providerMap =
        result.get(category) ?? new Map<string, PickerGame[]>();
      const list = providerMap.get(provider) ?? [];
      list.push({ id: game.id, name: game.name, provider });
      providerMap.set(provider, list);
      result.set(category, providerMap);
    });
    return result;
  });

  const categories = computed(() => Array.from(grouped.value.keys()));

  const activeCategory = ref('');
  const activeProvider = ref('');
  const providerKeyword = ref('');
  const gameKeyword = ref('');

  const providers = computed<ProviderGroup[]>(() =>
    Array.from(grouped.value.get(activeCategory.value) || []).map(
      ([name, list]) => ({ name, games: list })
    )
  );

  const filteredProviders = computed(() => {
    const keyword = providerKeyword.value.trim().toLowerCase();
    if (!keyword) return providers.value;
    return providers.value.filter((item) =>
      item.name.toLowerCase().includes(keyword)
    );
  });

  watch(
    categories,
    (list) => {
      if (!list.includes(activeCategory.value)) {
        activeCategory.value = list[0] ?? '';
      }
    },
    { immediate: true }
  );

  watch(
    activeCategory,
    () => {
      providerKeyword.value = '';
      gameKeyword.value = '';
      activeProvider.value = providers.value[0]?.name ?? '';
    },
    { immediate: true }
  );

  // 游戏数据异步加载完成后补上默认平台
  watch(providers, (list) => {
    if (!list.some((item) => item.name === activeProvider.value)) {
      activeProvider.value = list[0]?.name ?? '';
    }
  });

  function selectProvider(name: string) {
    activeProvider.value = name;
    gameKeyword.value = '';
  }

  const visibleGames = computed(() => {
    const keyword = gameKeyword.value.trim().toLowerCase();
    if (keyword) {
      return providers.value
        .flatMap((item) => item.games)
        .filter((game) => game.name.toLowerCase().includes(keyword));
    }
    return (
      providers.value.find((item) => item.name === activeProvider.value)
        ?.games ?? []
    );
  });

  const selected = computed(() => new Set(props.modelValue));

  const selectedCount = (list: PickerGame[]) =>
    list.filter((game) => selected.value.has(game.id)).length;

  const isAllSelected = (list: PickerGame[]) =>
    list.length > 0 && selectedCount(list) === list.length;

  const isPartSelected = (list: PickerGame[]) => {
    const count = selectedCount(list);
    return count > 0 && count < list.length;
  };

  function setSelected(list: PickerGame[], value: boolean) {
    const next = new Set(props.modelValue);
    list.forEach((game) => {
      if (value) next.add(game.id);
      else next.delete(game.id);
    });
    emit('update:modelValue', Array.from(next));
  }

  function categorySelectedCount(category: string) {
    const providerMap = grouped.value.get(category);
    if (!providerMap) return 0;
    return Array.from(providerMap.values()).reduce(
      (total, list) => total + selectedCount(list),
      0
    );
  }

  const allGames = computed(() =>
    Array.from(grouped.value.values()).flatMap((providerMap) =>
      Array.from(providerMap.values()).flat()
    )
  );

  const selectedProviderCount = computed(() => {
    const names = new Set<string>();
    allGames.value.forEach((game) => {
      if (selected.value.has(game.id)) names.add(game.provider);
    });
    return names.size;
  });

  const selectedPreview = computed(() =>
    allGames.value
      .filter((game) => selected.value.has(game.id))
      .slice(0, PREVIEW_LIMIT)
  );
</script>

<style scoped lang="less">
  .game-picker {
    padding: 12px;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 10px;
  }

  .picker-empty {
    padding: 24px 0;
    color: var(--color-text-3);
    text-align: center;
  }

  .category-tabs {
    display: flex;
    gap: 4px;
    margin-bottom: 10px;
    border-bottom: 1px solid var(--color-border-1);
  }

  .category-tab {
    display: inline-flex;
    gap: 6px;
    align-items: center;
    margin-bottom: -1px;
    padding: 8px 14px;
    color: var(--color-text-2);
    font-size: 13px;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    cursor: pointer;

    &.active {
      color: rgb(var(--primary-6));
      font-weight: 600;
      border-bottom-color: rgb(var(--primary-6));
    }
  }

  .tab-count {
    min-width: 18px;
    padding: 0 5px;
    color: #fff;
    font-weight: 500;
    font-size: 11px;
    line-height: 16px;
    text-align: center;
    background: rgb(var(--primary-6));
    border-radius: 8px;
  }

  .picker-body {
    display: grid;
    grid-template-columns: 180px minmax(0, 1fr);
    height: 340px;
    overflow: hidden;
    border: 1px solid var(--color-border-1);
    border-radius: 8px;
  }

  .provider-pane {
    display: flex;
    flex-direction: column;
    min-height: 0;
    background: var(--color-fill-1);
    border-right: 1px solid var(--color-border-1);
  }

  .pane-search {
    flex: none;
    margin: 8px;
    width: auto;
  }

  .provider-list {
    flex: 1;
    min-height: 0;
    padding: 0 6px 6px;
    overflow-y: auto;
  }

  .provider-item {
    display: flex;
    gap: 6px;
    align-items: center;
    padding: 6px 8px;
    font-size: 13px;
    border-radius: 6px;
    cursor: pointer;

    &:hover {
      background: var(--color-fill-2);
    }

    &.active {
      color: rgb(var(--primary-6));
      background: rgb(var(--primary-1));
    }

    &.has-selected .provider-count {
      color: rgb(var(--primary-6));
      font-weight: 600;
    }
  }

  .provider-name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .provider-count {
    flex: none;
    color: var(--color-text-3);
    font-size: 12px;
    font-variant-numeric: tabular-nums;
  }

  .game-pane {
    display: flex;
    flex-direction: column;
    min-width: 0;
    min-height: 0;
  }

  .game-toolbar {
    display: flex;
    flex: none;
    gap: 4px;
    align-items: center;
    padding-right: 6px;
  }

  .game-search {
    flex: 1;
    min-width: 0;
  }

  .select-all {
    flex: none;
  }

  .game-scope {
    flex: none;
    padding: 0 10px 6px;
    overflow: hidden;
    color: var(--color-text-3);
    font-size: 12px;
    white-space: nowrap;
    text-overflow: ellipsis;
    border-bottom: 1px solid var(--color-border-1);
  }

  .game-grid {
    display: grid;
    flex: 1;
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    grid-auto-rows: min-content;
    gap: 8px;
    min-height: 0;
    padding: 10px;
    overflow-y: auto;
  }

  .game-option {
    display: flex;
    gap: 6px;
    align-items: center;
    min-width: 0;
    padding: 6px 8px;
    font-size: 13px;
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
    cursor: pointer;

    &.checked {
      background: rgb(var(--primary-1));
      border-color: rgb(var(--primary-3));
    }
  }

  .game-name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .game-provider {
    flex: none;
    color: var(--color-text-3);
    font-size: 11px;
  }

  .pane-empty {
    grid-column: 1 / -1;
    padding: 24px 0;
    color: var(--color-text-3);
    font-size: 12px;
    text-align: center;
  }

  .picker-footer {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-top: 10px;
  }

  .footer-summary {
    color: var(--color-text-2);
    font-size: 13px;

    b {
      color: rgb(var(--primary-6));
    }
  }

  .selected-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    align-items: center;
  }

  .more-hint {
    color: var(--color-text-3);
    font-size: 12px;
  }

  .clear-all {
    font-size: 12px;
  }
</style>
