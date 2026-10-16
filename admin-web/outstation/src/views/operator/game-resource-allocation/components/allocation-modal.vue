<template>
  <a-modal
    :visible="visible"
    :title="t('operatorGameResource.modal.title')"
    :width="1350"
    :footer="false"
    unmount-on-close
    modal-class="game-allocation-modal"
    @cancel="handleCancel"
  >
    <div class="allocation-body">
      <div class="allocation-main">
        <a-form
          :model="operatorPickerForm"
          layout="inline"
          class="operator-picker"
        >
          <a-form-item
            :label="t('operatorGameResource.modal.selectOperator')"
            field="operatorId"
          >
            <a-select
              v-model="operatorPickerForm.operatorId"
              :placeholder="
                t('operatorGameResource.modal.selectOperatorPlaceholder')
              "
              :loading="operatorOptionsLoading"
              allow-search
              :filter-option="false"
              class="operator-select"
              :options="operatorOptions"
              @search="handleOperatorSearch"
              @change="handleOperatorChange"
            />
          </a-form-item>
        </a-form>

        <div v-if="operatorInfo" class="operator-info">
          <div class="operator-info__item">
            <span class="operator-info__label">
              {{ t('operatorGameResource.modal.info.code') }}
            </span>
            <span class="operator-info__value">{{ operatorInfo.code }}</span>
          </div>
          <div class="operator-info__item">
            <span class="operator-info__label">
              {{ t('operatorGameResource.modal.info.timezone') }}
            </span>
            <span class="operator-info__value">
              {{ operatorInfo.timezone_code || '--' }}
            </span>
          </div>
          <div class="operator-info__item">
            <span class="operator-info__label">
              {{ t('operatorGameResource.modal.info.domain') }}
            </span>
            <span class="operator-info__value">{{
              operatorDomain || '--'
            }}</span>
          </div>
          <div class="operator-info__item">
            <span class="operator-info__label">
              {{ t('operatorGameResource.modal.info.status') }}
            </span>
            <span class="operator-info__value status-dot-text">
              <span class="status-dot" :class="operatorStatusMeta.dotClass" />
              {{ operatorStatusMeta.label }}
            </span>
          </div>
        </div>

        <a-empty
          v-if="!operatorPickerForm.operatorId"
          :description="
            t('operatorGameResource.modal.message.selectOperatorFirst')
          "
          class="operator-empty"
        />

        <template v-else>
          <a-tabs v-model:active-key="resourceState.activeTab">
            <a-tab-pane
              v-for="tab in resourceTabs"
              :key="tab.key"
              :title="tab.title"
            />
          </a-tabs>
          <div class="resource-toolbar">
            <a-input
              v-model="resourceState.keyword"
              :placeholder="keywordPlaceholder"
              allow-clear
              class="resource-search"
              @press-enter="handleSearch"
            >
              <template #prefix><icon-search /></template>
            </a-input>
            <!-- 暂时用不到 -->
            <!-- <a-select
              v-model="resourceState.statusFilter"
              class="resource-filter-select"
              :options="statusFilterOptions"
            /> -->
            <a-select
              v-model="resourceState.allocatedFilter"
              class="resource-filter-select"
              :options="allocatedFilterOptions"
            />
            <a-space>
              <a-button type="primary" @click="handleSearch">
                <template #icon><icon-search /></template>
                {{ t('front.common.search.search') }}
              </a-button>
              <a-button @click="handleResetFilter">
                <template #icon><icon-refresh /></template>
                {{ t('front.common.search.reset') }}
              </a-button>
              <a-dropdown
                :disabled="!activeSelectedIds.length"
                @select="handleBatchStatusSelect"
              >
                <a-button :loading="batchStatusLoading">
                  {{ t('operatorGameResource.modal.action.batchStatus') }}
                  <icon-down />
                </a-button>
                <template #content>
                  <a-doption :value="OPERATOR_GAME_ITEM_STATUS.ENABLED">
                    {{ t('operatorGameResource.modal.action.batchEnable') }}
                  </a-doption>
                  <a-doption :value="OPERATOR_GAME_ITEM_STATUS.DISABLED">
                    {{ t('operatorGameResource.modal.action.batchDisable') }}
                  </a-doption>
                </template>
              </a-dropdown>
            </a-space>
          </div>
          <a-table
            row-key="code"
            :loading="resourcesLoading"
            :columns="resourceColumns"
            :data="pagedResourceList"
            :pagination="{
              current: resourcePagination.current,
              pageSize: resourcePagination.pageSize,
              total: filteredResourceList.length,
              showTotal: true,
              pageSizeOptions: [10, 20, 50],
            }"
            :bordered="true"
            :row-selection="{
              type: 'checkbox',
              showCheckedAll: true,
              selectedRowKeys: activeResourceSelectedKeys,
            }"
            size="small"
            :scroll="{ x: 980 }"
            @select="onSelectChange"
            @select-all="onSelectAllChange"
            @page-change="onResourcePageChange"
            @page-size-change="onResourcePageSizeChange"
          >
            <template #status="{ record }">
              <span class="status-dot-text">
                <span
                  class="status-dot"
                  :class="
                    (record as ResourceItem).enabled
                      ? 'status-dot--enabled'
                      : 'status-dot--disabled'
                  "
                />
                {{
                  (record as ResourceItem).enabled
                    ? t('operatorGameResource.modal.filter.enabled')
                    : t('operatorGameResource.modal.filter.disabled')
                }}
              </span>
            </template>
            <template #allocated="{ record }">
              <a-tag
                :color="(record as ResourceItem).assigned ? 'green' : 'gray'"
              >
                {{
                  (record as ResourceItem).assigned
                    ? t('operatorGameResource.modal.allocated.yes')
                    : t('operatorGameResource.modal.allocated.no')
                }}
              </a-tag>
            </template>
            <template #allocatedAt="{ record }">
              {{ (record as ResourceItem).assignedAt || '--' }}
            </template>
            <template #remark="{ record }">
              <span class="remark-cell">
                {{ (record as ResourceItem).remark || '--' }}
              </span>
            </template>
            <template #operations="{ record }">
              <a-link @click="handleShowDetail(record as ResourceItem)">
                {{ t('operatorGameResource.modal.action.detail') }}
              </a-link>
            </template>
          </a-table>
        </template>
      </div>

      <div class="allocation-side">
        <div class="side-card">
          <div class="side-card__title">
            {{ t('operatorGameResource.modal.summary.title') }}
          </div>
          <div class="summary-item">
            <icon-apps class="summary-item__icon" />
            <div class="summary-item__body">
              <div class="summary-item__label">
                {{ t('operatorGameResource.modal.summary.category') }}
              </div>
              <div class="summary-item__value">
                {{ resourceState.selected.category.length }}
                {{ t('operatorGameResource.modal.summary.unit') }}
              </div>
            </div>
          </div>
          <div class="summary-item">
            <icon-user-group class="summary-item__icon" />
            <div class="summary-item__body">
              <div class="summary-item__label">
                {{ t('operatorGameResource.modal.summary.provider') }}
              </div>
              <div class="summary-item__value">
                {{ resourceState.selected.provider.length }}
                {{ t('operatorGameResource.modal.summary.unit') }}
              </div>
            </div>
          </div>
          <div class="summary-item">
            <icon-share-alt class="summary-item__icon" />
            <div class="summary-item__body">
              <div class="summary-item__label">
                {{ t('operatorGameResource.modal.summary.channel') }}
              </div>
              <div class="summary-item__value">
                {{ resourceState.selected.channel.length }}
                {{ t('operatorGameResource.modal.summary.unit') }}
              </div>
            </div>
          </div>
          <div class="summary-item">
            <icon-desktop class="summary-item__icon" />
            <div class="summary-item__body">
              <div class="summary-item__label">
                {{ t('operatorGameResource.modal.summary.game') }}
              </div>
              <div class="summary-item__value">
                {{ resourceState.selected.game.length }}
                {{ t('operatorGameResource.modal.summary.unit') }}
              </div>
            </div>
          </div>
        </div>

        <a-button
          type="primary"
          long
          class="save-btn"
          :loading="saving"
          :disabled="!operatorPickerForm.operatorId"
          @click="handleSave"
        >
          <template #icon><icon-save /></template>
          {{ t('operatorGameResource.modal.action.save') }}
        </a-button>
        <a-button
          long
          class="reset-btn"
          :disabled="!operatorPickerForm.operatorId"
          @click="handleResetSelection"
        >
          <template #icon><icon-refresh /></template>
          {{ t('operatorGameResource.modal.action.resetSelection') }}
        </a-button>

        <a-alert type="info" :show-icon="true" class="side-tip">
          <template #title>
            {{ t('operatorGameResource.modal.tip.title') }}
          </template>
          <div>{{ t('operatorGameResource.modal.tip.content') }}</div>
        </a-alert>
      </div>
    </div>

    <a-modal
      :visible="detailVisible"
      :title="t('operatorGameResource.modal.detail.title')"
      :width="520"
      :footer="false"
      unmount-on-close
      @cancel="detailVisible = false"
    >
      <a-descriptions v-if="detailRecord" :column="1" bordered size="medium">
        <a-descriptions-item
          :label="t('operatorGameResource.modal.column.name')"
        >
          {{ detailRecord.name }}
        </a-descriptions-item>
        <a-descriptions-item
          :label="t('operatorGameResource.modal.column.code')"
        >
          {{ detailRecord.code }}
        </a-descriptions-item>
        <a-descriptions-item
          :label="t('operatorGameResource.modal.column.status')"
        >
          {{
            detailRecord.enabled
              ? t('operatorGameResource.modal.filter.enabled')
              : t('operatorGameResource.modal.filter.disabled')
          }}
        </a-descriptions-item>
        <a-descriptions-item
          :label="t('operatorGameResource.modal.column.allocated')"
        >
          {{
            detailRecord.assigned
              ? t('operatorGameResource.modal.allocated.yes')
              : t('operatorGameResource.modal.allocated.no')
          }}
        </a-descriptions-item>
        <a-descriptions-item
          :label="t('operatorGameResource.modal.column.allocatedAt')"
        >
          {{ detailRecord.assignedAt || '--' }}
        </a-descriptions-item>
        <a-descriptions-item
          :label="t('operatorGameResource.modal.column.remark')"
        >
          {{ detailRecord.remark || '--' }}
        </a-descriptions-item>
      </a-descriptions>
    </a-modal>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import dayjs from 'dayjs';
  import {
    batchCreateOperatorGameCategory,
    batchCreateOperatorGameChannel,
    batchCreateOperatorGameItem,
    batchCreateOperatorGameProvider,
    batchUpdateOperatorGameCategoryStatus,
    batchUpdateOperatorGameChannelStatus,
    batchUpdateOperatorGameItemStatus,
    batchUpdateOperatorGameProviderStatus,
    getOperator,
    OPERATOR_GAME_CHECK_STATUS,
    OPERATOR_GAME_ITEM_STATUS,
    OPERATOR_STATUS,
    queryOperatorDomainList,
    queryOperatorGameCategoryList,
    queryOperatorGameChannelList,
    queryOperatorGameItemList,
    queryOperatorGameProviderList,
    queryOperatorList,
    type OperatorGameCategoryRecord,
    type OperatorGameChannelRecord,
    type OperatorGameItemRecord,
    type OperatorGameProviderRecord,
    type OperatorRecord,
  } from '@/api/operator';

  const props = defineProps<{
    visible: boolean;
    opCode: string | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  const { t } = useI18n();

  interface ResourceItem {
    id: number;
    code: string;
    name: string;
    enabled: boolean;
    assigned: boolean;
    assignedAt: string | null;
    remark: string;
  }

  type ResourceTabKey = 'category' | 'provider' | 'channel' | 'game';

  const resourceTabs = computed<{ key: ResourceTabKey; title: string }[]>(
    () => [
      {
        key: 'category',
        title: t('operatorGameResource.modal.tab.category'),
      },
      {
        key: 'provider',
        title: t('operatorGameResource.modal.tab.provider'),
      },
      {
        key: 'channel',
        title: t('operatorGameResource.modal.tab.channel'),
      },
      { key: 'game', title: t('operatorGameResource.modal.tab.game') },
    ]
  );

  const statusFilterOptions = computed(() => [
    { label: t('operatorGameResource.modal.filter.all'), value: 'all' },
    {
      label: t('operatorGameResource.modal.filter.enabled'),
      value: 'enabled',
    },
    {
      label: t('operatorGameResource.modal.filter.disabled'),
      value: 'disabled',
    },
  ]);

  const allocatedFilterOptions = computed(() => [
    { label: t('operatorGameResource.modal.filter.all'), value: 'all' },
    {
      label: t('operatorGameResource.modal.filter.allocated'),
      value: 'allocated',
    },
    {
      label: t('operatorGameResource.modal.filter.unallocated'),
      value: 'unallocated',
    },
  ]);

  const emptyResourceItems = (): Record<ResourceTabKey, ResourceItem[]> => ({
    category: [],
    provider: [],
    channel: [],
    game: [],
  });

  const resourceState = reactive<{
    activeTab: ResourceTabKey;
    keyword: string;
    appliedKeyword: string;
    statusFilter: 'all' | 'enabled' | 'disabled';
    allocatedFilter: 'all' | 'allocated' | 'unallocated';
    items: Record<ResourceTabKey, ResourceItem[]>;
    selected: Record<ResourceTabKey, string[]>;
  }>({
    activeTab: 'category',
    keyword: '',
    appliedKeyword: '',
    statusFilter: 'all',
    allocatedFilter: 'all',
    items: emptyResourceItems(),
    selected: { category: [], provider: [], channel: [], game: [] },
  });

  const keywordPlaceholder = computed(() => {
    const map: Record<ResourceTabKey, string> = {
      category: t(
        'operatorGameResource.modal.search.keywordPlaceholder.category'
      ),
      provider: t(
        'operatorGameResource.modal.search.keywordPlaceholder.provider'
      ),
      channel: t(
        'operatorGameResource.modal.search.keywordPlaceholder.channel'
      ),
      game: t('operatorGameResource.modal.search.keywordPlaceholder.game'),
    };
    return map[resourceState.activeTab];
  });

  const resourcePagination = reactive({ current: 1, pageSize: 10 });

  watch(
    () => [
      resourceState.activeTab,
      resourceState.appliedKeyword,
      resourceState.statusFilter,
      resourceState.allocatedFilter,
    ],
    () => {
      resourcePagination.current = 1;
    }
  );

  const activeResourceSelectedKeys = computed(
    () => resourceState.selected[resourceState.activeTab]
  );

  const filteredResourceList = computed(() => {
    const keyword = resourceState.appliedKeyword.trim().toLowerCase();
    return resourceState.items[resourceState.activeTab].filter((item) => {
      const matchKeyword =
        !keyword ||
        item.name.toLowerCase().includes(keyword) ||
        item.code.toLowerCase().includes(keyword);
      const matchStatus =
        resourceState.statusFilter === 'all' ||
        (resourceState.statusFilter === 'enabled' && item.enabled) ||
        (resourceState.statusFilter === 'disabled' && !item.enabled);
      const matchAllocated =
        resourceState.allocatedFilter === 'all' ||
        (resourceState.allocatedFilter === 'allocated' && item.assigned) ||
        (resourceState.allocatedFilter === 'unallocated' && !item.assigned);
      return matchKeyword && matchStatus && matchAllocated;
    });
  });

  const pagedResourceList = computed(() => {
    const start =
      (resourcePagination.current - 1) * resourcePagination.pageSize;
    return filteredResourceList.value.slice(
      start,
      start + resourcePagination.pageSize
    );
  });

  const onResourcePageChange = (current: number) => {
    resourcePagination.current = current;
  };
  const onResourcePageSizeChange = (pageSize: number) => {
    resourcePagination.current = 1;
    resourcePagination.pageSize = pageSize;
  };

  const resourceColumns = computed<TableColumnData[]>(() => [
    {
      title: t('operatorGameResource.modal.column.name'),
      dataIndex: 'name',
      width: 140,
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('operatorGameResource.modal.column.code'),
      dataIndex: 'code',
      width: 120,
      align: 'center',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('operatorGameResource.modal.column.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 100,
      align: 'center',
    },
    {
      title: t('operatorGameResource.modal.column.allocated'),
      dataIndex: 'allocated',
      slotName: 'allocated',
      width: 100,
      align: 'center',
    },
    {
      title: t('operatorGameResource.modal.column.allocatedAt'),
      dataIndex: 'allocatedAt',
      slotName: 'allocatedAt',
      width: 170,
      align: 'center',
    },
    // 暂时去掉
    // {
    //   title: t('operatorGameResource.modal.column.remark'),
    //   dataIndex: 'remark',
    //   slotName: 'remark',
    //   width: 180,
    //   align: 'center',
    //   ellipsis: true,
    //   tooltip: true,
    // },
    {
      title: t('operatorGameResource.modal.column.operations'),
      dataIndex: 'operations',
      slotName: 'operations',
      width: 80,
      align: 'center',
      fixed: 'right',
    },
  ]);

  const formatTime = (value?: number | null) => {
    if (!value) return null;
    const ms = value < 1e12 ? value * 1000 : value;
    return dayjs(ms).format('YYYY-MM-DD HH:mm:ss');
  };

  const isCheckedByStatus = (checkStatus: number) => checkStatus === 1;

  const mapCategoryItem = (item: OperatorGameCategoryRecord): ResourceItem => {
    const assigned = isCheckedByStatus(item.check_status);
    return {
      id: item.id,
      code: item.category_code,
      name: item.name || '--',
      enabled: item.status === OPERATOR_GAME_ITEM_STATUS.ENABLED,
      assigned,
      assignedAt: assigned ? formatTime(item.updated_at) : null,
      remark: item.remark || '',
    };
  };

  const mapProviderItem = (item: OperatorGameProviderRecord): ResourceItem => {
    const assigned = isCheckedByStatus(item.check_status);
    return {
      id: item.id,
      code: item.provider_code,
      name: item.name || '--',
      enabled: item.status === OPERATOR_GAME_ITEM_STATUS.ENABLED,
      assigned,
      assignedAt: assigned ? formatTime(item.updated_at) : null,
      remark: item.remark || '',
    };
  };

  const mapChannelItem = (item: OperatorGameChannelRecord): ResourceItem => {
    const assigned = isCheckedByStatus(item.check_status);
    return {
      id: item.id,
      code: item.channel_code,
      name: item.name || '--',
      enabled: item.status === OPERATOR_GAME_ITEM_STATUS.ENABLED,
      assigned,
      assignedAt: assigned ? formatTime(item.updated_at) : null,
      remark: item.remark || '',
    };
  };

  const mapGameItem = (item: OperatorGameItemRecord): ResourceItem => {
    const assigned = isCheckedByStatus(item.check_status);
    return {
      id: item.id,
      code: item.game_code,
      name: item.name || '--',
      enabled: item.status === OPERATOR_GAME_ITEM_STATUS.ENABLED,
      assigned,
      assignedAt: assigned ? formatTime(item.updated_at) : null,
      remark: item.remark || '',
    };
  };

  const setSelection = (tab: ResourceTabKey, codes: string[]) => {
    const nowStr = dayjs().format('YYYY-MM-DD HH:mm:ss');
    resourceState.items[tab].forEach((item) => {
      const nowAssigned = codes.includes(item.code);
      if (nowAssigned && !item.assigned) item.assignedAt = nowStr;
      if (!nowAssigned) item.assignedAt = null;
      item.assigned = nowAssigned;
    });
    resourceState.selected[tab] = codes;
  };

  const onSelectChange = (rowKeys: (string | number)[]) => {
    setSelection(
      resourceState.activeTab,
      rowKeys.map((key) => String(key))
    );
  };

  const onSelectAllChange = (checked: boolean) => {
    const tab = resourceState.activeTab;
    if (checked) {
      const visibleCodes = filteredResourceList.value.map((item) => item.code);
      const merged = new Set([...resourceState.selected[tab], ...visibleCodes]);
      setSelection(tab, [...merged]);
    } else {
      const visibleCodes = new Set(
        filteredResourceList.value.map((item) => item.code)
      );
      setSelection(
        tab,
        resourceState.selected[tab].filter((code) => !visibleCodes.has(code))
      );
    }
  };

  type GameResourcePageFetcher<T> = (params: {
    page: number;
    page_size: number;
    op_code: string;
    status?: number;
  }) => Promise<{ data?: { items?: T[]; total?: number } }>;

  async function fetchAllItems<T>(
    fetchPage: GameResourcePageFetcher<T>,
    opCode: string,
    status?: number
  ) {
    const pageSize = 100;
    const buildParams = (page: number) => ({
      page,
      page_size: pageSize,
      op_code: opCode,
      ...(status !== undefined ? { status } : {}),
    });
    const firstRes = await fetchPage(buildParams(1));
    const firstItems = firstRes.data?.items || [];
    const total = firstRes.data?.total ?? firstItems.length;
    if (firstItems.length >= total) return firstItems;

    const totalPages = Math.min(Math.ceil(total / pageSize), 50);
    const restPages = Array.from(
      { length: totalPages - 1 },
      (_, index) => index + 2
    );
    const restResults = await Promise.all(
      restPages.map((page) => fetchPage(buildParams(page)))
    );
    const restItems = restResults.flatMap((res) => res.data?.items || []);
    return [...firstItems, ...restItems];
  }

  const operatorPickerForm = reactive<{ operatorId?: number }>({
    operatorId: undefined,
  });
  const operatorInfo = ref<OperatorRecord | null>(null);
  const operatorDomain = ref('');
  const operatorOptions = ref<{ label: string; value: number }[]>([]);
  const operatorOptionsLoading = ref(false);
  const resourcesLoading = ref(false);
  const batchStatusLoading = ref(false);

  const activeSelectedIds = computed(() => {
    const selectedCodes = new Set(
      resourceState.selected[resourceState.activeTab]
    );
    return resourceState.items[resourceState.activeTab]
      .filter((item) => selectedCodes.has(item.code) && item.id > 0)
      .map((item) => item.id);
  });

  const operatorStatusMeta = computed(() => {
    const status = operatorInfo.value?.status;
    if (status === OPERATOR_STATUS.PAUSED) {
      return {
        label: t('operatorStation.status.paused'),
        dotClass: 'status-dot--paused',
      };
    }
    if (status === OPERATOR_STATUS.CLOSED) {
      return {
        label: t('operatorStation.status.closed'),
        dotClass: 'status-dot--disabled',
      };
    }
    return {
      label: t('operatorStation.status.normal'),
      dotClass: 'status-dot--enabled',
    };
  });

  const loadOperatorOptions = async (keyword?: string) => {
    operatorOptionsLoading.value = true;
    try {
      const { data } = await queryOperatorList({
        page: 1,
        page_size: 100,
        ...(keyword ? { keyword } : {}),
      });
      operatorOptions.value = (data?.list || []).map(
        (item: OperatorRecord) => ({
          label: `${item.name}（${item.code}）`,
          value: item.id,
        })
      );
    } catch {
      operatorOptions.value = [];
    } finally {
      operatorOptionsLoading.value = false;
    }
  };

  let operatorSearchTimer: ReturnType<typeof setTimeout> | null = null;
  const handleOperatorSearch = (keyword: string) => {
    if (operatorSearchTimer) clearTimeout(operatorSearchTimer);
    operatorSearchTimer = setTimeout(() => {
      loadOperatorOptions(keyword.trim() || undefined);
    }, 300);
  };

  const loadOperatorInfo = async (id: number) => {
    try {
      const [operatorRes, domainRes] = await Promise.all([
        getOperator({ id }),
        queryOperatorDomainList({ page: 1, page_size: 10, operator_id: id }),
      ]);
      operatorInfo.value = operatorRes.data || null;
      operatorDomain.value =
        (domainRes.data?.list || [])
          .map((item) => item.domain_name)
          .filter(Boolean)
          .join('、') || '';
    } catch {
      operatorInfo.value = null;
      operatorDomain.value = '';
    }
  };

  const buildSelectedMap = (itemsMap: Record<ResourceTabKey, ResourceItem[]>) =>
    ({
      category: itemsMap.category
        .filter((item) => item.assigned)
        .map((item) => item.code),
      provider: itemsMap.provider
        .filter((item) => item.assigned)
        .map((item) => item.code),
      channel: itemsMap.channel
        .filter((item) => item.assigned)
        .map((item) => item.code),
      game: itemsMap.game
        .filter((item) => item.assigned)
        .map((item) => item.code),
    } as Record<ResourceTabKey, string[]>);

  const loadResources = async (opCode: string) => {
    resourcesLoading.value = true;
    try {
      const [categoryItems, providerItems, channelItems, gameItems] =
        await Promise.all([
          fetchAllItems(queryOperatorGameCategoryList, opCode),
          fetchAllItems(queryOperatorGameProviderList, opCode),
          fetchAllItems(queryOperatorGameChannelList, opCode),
          fetchAllItems(queryOperatorGameItemList, opCode),
        ]);

      const itemsMap: Record<ResourceTabKey, ResourceItem[]> = {
        category: categoryItems.map(mapCategoryItem),
        provider: providerItems.map(mapProviderItem),
        channel: channelItems.map(mapChannelItem),
        game: gameItems.map(mapGameItem),
      };

      resourceState.items = itemsMap;
      resourceState.selected = buildSelectedMap(itemsMap);
    } catch {
      resourceState.items = emptyResourceItems();
      resourceState.selected = {
        category: [],
        provider: [],
        channel: [],
        game: [],
      };
    } finally {
      resourcesLoading.value = false;
    }
  };

  const loadAllForOperator = async (id: number) => {
    await loadOperatorInfo(id);
    const opCode = operatorInfo.value?.code;
    if (!opCode) return;
    await loadResources(opCode);
  };

  const handleOperatorChange = (value: unknown) => {
    const id = value as number | undefined;
    if (id) loadAllForOperator(id);
  };

  const handleSearch = () => {
    resourceState.appliedKeyword = resourceState.keyword;
  };

  const handleResetFilter = () => {
    resourceState.keyword = '';
    resourceState.appliedKeyword = '';
    resourceState.statusFilter = 'all';
    resourceState.allocatedFilter = 'all';
  };

  const handleResetSelection = () => {
    const opCode = operatorInfo.value?.code;
    if (!opCode) return;
    loadResources(opCode);
  };

  const saving = ref(false);
  const handleSave = async () => {
    const opCode = operatorInfo.value?.code;
    if (!opCode) {
      Message.warning(
        t('operatorGameResource.modal.message.selectOperatorFirst')
      );
      return;
    }

    const tab = resourceState.activeTab;
    let saveTask: Promise<unknown> | null = null;

    if (tab === 'category') {
      const selectedCategoryCodes = new Set(resourceState.selected.category);
      const categoryItems = resourceState.items.category.map((item) => ({
        code: item.code,
        check_status: selectedCategoryCodes.has(item.code)
          ? OPERATOR_GAME_CHECK_STATUS.CHECKED
          : OPERATOR_GAME_CHECK_STATUS.UNCHECKED,
      }));
      if (!categoryItems.length) {
        Message.warning(t('operatorGameResource.modal.message.saveEmpty'));
        return;
      }
      saveTask = batchCreateOperatorGameCategory({
        op_code: opCode,
        items: categoryItems,
      });
    } else if (tab === 'provider') {
      const selectedProviderCodes = new Set(resourceState.selected.provider);
      const providerItems = resourceState.items.provider.map((item) => ({
        code: item.code,
        check_status: selectedProviderCodes.has(item.code)
          ? OPERATOR_GAME_CHECK_STATUS.CHECKED
          : OPERATOR_GAME_CHECK_STATUS.UNCHECKED,
      }));
      if (!providerItems.length) {
        Message.warning(t('operatorGameResource.modal.message.saveEmpty'));
        return;
      }
      saveTask = batchCreateOperatorGameProvider({
        op_code: opCode,
        items: providerItems,
      });
    } else if (tab === 'channel') {
      const selectedChannelCodes = new Set(resourceState.selected.channel);
      const channelItems = resourceState.items.channel.map((item) => ({
        code: item.code,
        check_status: selectedChannelCodes.has(item.code)
          ? OPERATOR_GAME_CHECK_STATUS.CHECKED
          : OPERATOR_GAME_CHECK_STATUS.UNCHECKED,
      }));
      if (!channelItems.length) {
        Message.warning(t('operatorGameResource.modal.message.saveEmpty'));
        return;
      }
      saveTask = batchCreateOperatorGameChannel({
        op_code: opCode,
        items: channelItems,
      });
    } else {
      const selectedGameCodes = new Set(resourceState.selected.game);
      const gameItems = resourceState.items.game.map((item) => ({
        code: item.code,
        check_status: selectedGameCodes.has(item.code)
          ? OPERATOR_GAME_CHECK_STATUS.CHECKED
          : OPERATOR_GAME_CHECK_STATUS.UNCHECKED,
      }));
      if (!gameItems.length) {
        Message.warning(t('operatorGameResource.modal.message.saveEmpty'));
        return;
      }
      saveTask = batchCreateOperatorGameItem({
        op_code: opCode,
        items: gameItems,
      });
    }

    saving.value = true;
    try {
      await saveTask;
      Message.success(t('operatorGameResource.modal.message.saveSuccess'));
      await loadResources(opCode);
      emit('success');
    } finally {
      saving.value = false;
    }
  };

  const handleBatchStatusSelect = async (
    value: string | number | Record<string, unknown> | undefined
  ) => {
    if (value === undefined || typeof value === 'object') return;
    const status = Number(value);
    const ids = activeSelectedIds.value;
    if (!ids.length) {
      Message.warning(t('operatorGameResource.modal.message.selectItemsFirst'));
      return;
    }
    const opCode = operatorInfo.value?.code;
    if (!opCode) {
      Message.warning(
        t('operatorGameResource.modal.message.selectOperatorFirst')
      );
      return;
    }

    const tab = resourceState.activeTab;
    const updateMap = {
      category: batchUpdateOperatorGameCategoryStatus,
      provider: batchUpdateOperatorGameProviderStatus,
      channel: batchUpdateOperatorGameChannelStatus,
      game: batchUpdateOperatorGameItemStatus,
    } as const;

    batchStatusLoading.value = true;
    try {
      await updateMap[tab]({ ids, status });
      Message.success(
        t('operatorGameResource.modal.message.batchStatusSuccess')
      );
      await loadResources(opCode);
    } finally {
      batchStatusLoading.value = false;
    }
  };

  const handleCancel = () => {
    emit('update:visible', false);
  };

  const detailVisible = ref(false);
  const detailRecord = ref<ResourceItem | null>(null);
  const handleShowDetail = (record: ResourceItem) => {
    detailRecord.value = record;
    detailVisible.value = true;
  };

  const resetLocalState = () => {
    operatorPickerForm.operatorId = undefined;
    operatorInfo.value = null;
    operatorDomain.value = '';
    resourceState.activeTab = 'category';
    resourceState.keyword = '';
    resourceState.appliedKeyword = '';
    resourceState.statusFilter = 'all';
    resourceState.allocatedFilter = 'all';
    resourceState.items = emptyResourceItems();
    resourceState.selected = {
      category: [],
      provider: [],
      channel: [],
      game: [],
    };
    resourcePagination.current = 1;
    detailVisible.value = false;
    detailRecord.value = null;
  };

  const resolveOperatorByCode = async (code: string) => {
    const { data } = await queryOperatorList({
      page: 1,
      page_size: 20,
      keyword: code,
    });
    const matched = (data?.list || []).find((item) => item.code === code);
    if (matched) {
      operatorPickerForm.operatorId = matched.id;
      await loadAllForOperator(matched.id);
    }
  };

  watch(
    () => props.visible,
    async (visible) => {
      if (visible) {
        resetLocalState();
        await loadOperatorOptions();
        if (props.opCode) {
          await resolveOperatorByCode(props.opCode);
        }
      }
    }
  );
</script>

<style scoped lang="less">
  .allocation-body {
    display: flex;
    gap: 20px;
    align-items: flex-start;
  }
  .allocation-main {
    flex: 1;
    min-width: 0;
  }
  .allocation-side {
    width: 260px;
    flex-shrink: 0;
  }
  .operator-picker {
    margin-bottom: 12px;
  }
  .operator-select {
    width: 320px;
  }
  .operator-info {
    display: flex;
    flex-wrap: wrap;
    gap: 12px 24px;
    padding: 12px 16px;
    margin-bottom: 16px;
    background-color: var(--color-fill-2);
    border-radius: 4px;
  }
  .operator-info__item {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
  }
  .operator-info__label {
    color: var(--color-text-3);
  }
  .operator-info__value {
    color: var(--color-text-1);
    font-weight: 500;
  }
  .operator-empty {
    padding: 60px 0;
  }
  .resource-toolbar {
    display: flex;
    align-items: center;
    flex-wrap: nowrap;
    gap: 12px;
    margin: 12px 0;
  }
  .resource-search {
    flex: 0 0 220px;
    width: 220px;
  }
  .resource-filter-select {
    flex: 0 0 140px;
    width: 140px;
  }
  .resource-toolbar :deep(.arco-select),
  .resource-toolbar :deep(.arco-input-wrapper) {
    width: 100%;
  }
  .remark-cell {
    display: inline-block;
    max-width: 160px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .side-card {
    padding: 16px;
    margin-bottom: 16px;
    background-color: var(--color-fill-1);
    border-radius: 4px;
  }
  .side-card__title {
    margin-bottom: 12px;
    font-size: 14px;
    font-weight: 600;
    color: var(--color-text-1);
  }
  .summary-item {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    padding: 10px 0;
    & + .summary-item {
      border-top: 1px solid var(--color-border-2);
    }
  }
  .summary-item__icon {
    margin-top: 2px;
    font-size: 18px;
    color: rgb(var(--arcoblue-6));
  }
  .summary-item__label {
    font-size: 13px;
    color: var(--color-text-2);
  }
  .summary-item__value {
    font-size: 18px;
    font-weight: 600;
    color: var(--color-text-1);
  }
  .save-btn {
    margin-bottom: 10px;
  }
  .reset-btn {
    margin-bottom: 16px;
  }
  .side-tip {
    font-size: 12px;
  }
  .status-dot-text {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .status-dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }
  .status-dot--enabled {
    background-color: rgb(var(--success-6));
  }
  .status-dot--paused {
    background-color: rgb(var(--warning-6));
  }
  .status-dot--disabled {
    background-color: rgb(var(--gray-5));
  }
</style>
