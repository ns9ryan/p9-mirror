<template>
  <a-modal
    :visible="visible"
    :title="t('operatorBasicResource.modal.title')"
    :width="1240"
    :footer="false"
    unmount-on-close
    modal-class="allocation-modal"
    @cancel="handleCancel"
  >
    <div class="allocation-body">
      <div class="allocation-main">
        <a-form :model="{}" layout="inline" class="operator-picker">
          <a-form-item :label="t('operatorBasicResource.modal.selectOperator')">
            <a-select
              v-model="selectedOperatorId"
              :placeholder="
                t('operatorBasicResource.modal.selectOperatorPlaceholder')
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
              {{ t('operatorBasicResource.modal.info.id') }}
            </span>
            <span class="operator-info__value">{{ operatorInfo.id }}</span>
          </div>
          <div class="operator-info__item">
            <span class="operator-info__label">
              {{ t('operatorBasicResource.modal.info.code') }}
            </span>
            <span class="operator-info__value">{{ operatorInfo.code }}</span>
          </div>
          <div class="operator-info__item">
            <span class="operator-info__label">
              {{ t('operatorBasicResource.modal.info.name') }}
            </span>
            <span class="operator-info__value">{{ operatorInfo.name }}</span>
          </div>
          <div class="operator-info__item">
            <span class="operator-info__label">
              {{ t('operatorBasicResource.modal.info.timezone') }}
            </span>
            <span class="operator-info__value">
              {{ operatorInfo.timezone_code || '--' }}
            </span>
          </div>
          <div class="operator-info__item">
            <span class="operator-info__label">
              {{ t('operatorBasicResource.modal.info.currency') }}
            </span>
            <span class="operator-info__value">
              {{ operatorInfo.settlement_currency_code || '--' }}
            </span>
          </div>
        </div>

        <a-empty
          v-if="!selectedOperatorId"
          :description="t('operatorBasicResource.modal.message.selectOperatorFirst')"
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
              :placeholder="t('operatorBasicResource.modal.searchPlaceholder')"
              allow-clear
              class="resource-search"
            >
              <template #prefix><icon-search /></template>
            </a-input>
            <a-select
              v-model="resourceState.statusFilter"
              class="resource-status-select"
              :options="statusFilterOptions"
            />
            <a-button type="primary" @click="() => {}">
              <template #icon><icon-search /></template>
              {{ t('front.common.search.search') }}
            </a-button>
            <a-button @click="handleResetFilter">
              <template #icon><icon-refresh /></template>
              {{ t('front.common.search.reset') }}
            </a-button>
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
                    ? t('operatorBasicResource.modal.filter.enabled')
                    : t('operatorBasicResource.modal.filter.disabled')
                }}
              </span>
            </template>
            <template #assigned="{ record }">
              <a-tag :color="(record as ResourceItem).assigned ? 'green' : 'gray'">
                {{
                  (record as ResourceItem).assigned
                    ? t('operatorBasicResource.modal.column.assigned')
                    : '--'
                }}
              </a-tag>
            </template>
            <template #assignedAt="{ record }">
              {{ (record as ResourceItem).assignedAt || '--' }}
            </template>
          </a-table>
        </template>
      </div>

      <div class="allocation-side">
        <div class="side-card">
          <div class="side-card__title">
            {{ t('operatorBasicResource.modal.summary.title') }}
          </div>
          <div class="summary-item">
            <icon-language class="summary-item__icon" />
            <div class="summary-item__body">
              <div class="summary-item__label">
                {{ t('operatorBasicResource.modal.summary.language') }}
              </div>
              <div class="summary-item__value">
                {{ resourceState.selected.language.length }} /
                {{ resourceState.items.language.length }}
              </div>
              <div class="summary-item__desc">
                {{ t('operatorBasicResource.modal.summary.desc') }}
              </div>
            </div>
          </div>
          <div class="summary-item">
            <icon-location class="summary-item__icon" />
            <div class="summary-item__body">
              <div class="summary-item__label">
                {{ t('operatorBasicResource.modal.summary.region') }}
              </div>
              <div class="summary-item__value">
                {{ resourceState.selected.region.length }} /
                {{ resourceState.items.region.length }}
              </div>
              <div class="summary-item__desc">
                {{ t('operatorBasicResource.modal.summary.desc') }}
              </div>
            </div>
          </div>
          <div class="summary-item">
            <icon-share-alt class="summary-item__icon" />
            <div class="summary-item__body">
              <div class="summary-item__label">
                {{ t('operatorBasicResource.modal.summary.proxyLine') }}
              </div>
              <div class="summary-item__value">
                {{ resourceState.selected.proxyLine.length }} /
                {{ resourceState.items.proxyLine.length }}
              </div>
              <div class="summary-item__desc">
                {{ t('operatorBasicResource.modal.summary.desc') }}
              </div>
            </div>
          </div>
        </div>

        <a-button
          type="primary"
          long
          class="save-btn"
          :loading="saving"
          :disabled="!selectedOperatorId"
          @click="handleSave"
        >
          <template #icon><icon-save /></template>
          {{ t('operatorBasicResource.modal.action.save') }}
        </a-button>
        <a-button
          long
          class="reset-btn"
          :disabled="!selectedOperatorId"
          @click="handleResetSelection"
        >
          <template #icon><icon-refresh /></template>
          {{ t('operatorBasicResource.modal.action.resetSelection') }}
        </a-button>

        <a-alert type="info" :show-icon="true" class="side-tip">
          <template #title>
            {{ t('operatorBasicResource.modal.tip.title') }}
          </template>
          <div>{{ t('operatorBasicResource.modal.tip.line1') }}</div>
          <div>{{ t('operatorBasicResource.modal.tip.line2') }}</div>
        </a-alert>
      </div>
    </div>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import dayjs from 'dayjs';
  import {
    getOperator,
    queryOperatorAgentLineAllocation,
    queryOperatorLanguageAllocation,
    queryOperatorList,
    queryOperatorRegionAllocation,
    saveOperatorAgentLineAllocation,
    saveOperatorLanguageAllocation,
    saveOperatorRegionAllocation,
    type OperatorRecord,
  } from '@/api/operator';
  import { queryEnabledLangs } from '@/api/i18n';
  import { queryRegionListAll, REGION_STATUS } from '@/api/region';

  const props = defineProps<{
    visible: boolean;
    operatorId: number | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  const { t } = useI18n();

  interface ResourceItem {
    code: string;
    name: string;
    enabled: boolean;
    assigned: boolean;
    assignedAt: string | null;
  }
  type ResourceTabKey = 'language' | 'region' | 'proxyLine';

  const resourceTabs = computed<{ key: ResourceTabKey; title: string }[]>(
    () => [
      { key: 'language', title: t('operatorBasicResource.modal.tab.language') },
      { key: 'region', title: t('operatorBasicResource.modal.tab.region') },
      {
        key: 'proxyLine',
        title: t('operatorBasicResource.modal.tab.proxyLine'),
      },
    ]
  );

  const statusFilterOptions = computed(() => [
    { label: t('operatorBasicResource.modal.filter.all'), value: 'all' },
    {
      label: t('operatorBasicResource.modal.filter.enabled'),
      value: 'enabled',
    },
    {
      label: t('operatorBasicResource.modal.filter.disabled'),
      value: 'disabled',
    },
  ]);

  const emptyResourceItems = (): Record<ResourceTabKey, ResourceItem[]> => ({
    language: [],
    region: [],
    proxyLine: [],
  });

  const resourceState = reactive<{
    activeTab: ResourceTabKey;
    keyword: string;
    statusFilter: 'all' | 'enabled' | 'disabled';
    items: Record<ResourceTabKey, ResourceItem[]>;
    selected: Record<ResourceTabKey, string[]>;
  }>({
    activeTab: 'language',
    keyword: '',
    statusFilter: 'all',
    items: emptyResourceItems(),
    selected: { language: [], region: [], proxyLine: [] },
  });

  const resourcePagination = reactive({ current: 1, pageSize: 10 });

  watch(
    () => [resourceState.activeTab, resourceState.keyword, resourceState.statusFilter],
    () => {
      resourcePagination.current = 1;
    }
  );

  const activeResourceSelectedKeys = computed(
    () => resourceState.selected[resourceState.activeTab]
  );

  const filteredResourceList = computed(() => {
    const keyword = resourceState.keyword.trim().toLowerCase();
    return resourceState.items[resourceState.activeTab].filter((item) => {
      const matchKeyword =
        !keyword ||
        item.name.toLowerCase().includes(keyword) ||
        item.code.toLowerCase().includes(keyword);
      const matchStatus =
        resourceState.statusFilter === 'all' ||
        (resourceState.statusFilter === 'enabled' && item.enabled) ||
        (resourceState.statusFilter === 'disabled' && !item.enabled);
      return matchKeyword && matchStatus;
    });
  });

  const pagedResourceList = computed(() => {
    const start = (resourcePagination.current - 1) * resourcePagination.pageSize;
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
      title: t('operatorBasicResource.modal.column.name'),
      dataIndex: 'name',
      align: 'center',
    },
    {
      title: t('operatorBasicResource.modal.column.code'),
      dataIndex: 'code',
      align: 'center',
    },
    {
      title: t('operatorBasicResource.modal.column.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 100,
      align: 'center',
    },
    {
      title: t('operatorBasicResource.modal.column.assigned'),
      dataIndex: 'assigned',
      slotName: 'assigned',
      width: 100,
      align: 'center',
    },
    {
      title: t('operatorBasicResource.modal.column.assignedAt'),
      dataIndex: 'assignedAt',
      slotName: 'assignedAt',
      width: 170,
      align: 'center',
    },
  ]);

  const formatTime = (value?: number | null) => {
    if (!value) return '--';
    const ms = value < 1e12 ? value * 1000 : value;
    return dayjs(ms).format('YYYY-MM-DD HH:mm:ss');
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

  // ---------------- 分站选择 ----------------
  const selectedOperatorId = ref<number | undefined>(undefined);
  const operatorInfo = ref<OperatorRecord | null>(null);
  const operatorOptions = ref<{ label: string; value: number }[]>([]);
  const operatorOptionsLoading = ref(false);

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

  const resourcesLoading = ref(false);

  const loadOperatorInfo = async (id: number) => {
    try {
      const { data } = await getOperator({ id });
      operatorInfo.value = data || null;
    } catch {
      operatorInfo.value = null;
    }
  };

  const loadResources = async (id: number) => {
    resourcesLoading.value = true;
    try {
      const [langRes, langAllocRes, regionRes, regionAllocRes, agentRes] =
        await Promise.all([
          queryEnabledLangs(),
          queryOperatorLanguageAllocation({ operator_id: id }),
          queryRegionListAll({ status: REGION_STATUS.ENABLED }),
          queryOperatorRegionAllocation({ operator_id: id }),
          queryOperatorAgentLineAllocation({ operator_id: id }),
        ]);

      const langAllocated = new Map(
        (langAllocRes.data?.list || []).map((item) => [
          item.language_code,
          item.allocated_at,
        ])
      );
      const languageItems: ResourceItem[] = (langRes.data?.list || []).map(
        (item) => {
          const allocatedAt = langAllocated.get(item.lang);
          return {
            code: item.lang,
            name: item.i18n_name || item.name || item.lang,
            enabled: !item.disabled,
            assigned: allocatedAt !== undefined,
            assignedAt:
              allocatedAt !== undefined ? formatTime(allocatedAt) : null,
          };
        }
      );

      const regionAllocated = new Map(
        (regionAllocRes.data?.list || []).map((item) => [
          item.region_code.toUpperCase(),
          item.allocated_at,
        ])
      );
      const regionItems: ResourceItem[] = (regionRes.data?.list || []).map(
        (item) => {
          const code = (item.code || '').toUpperCase();
          const allocatedAt = regionAllocated.get(code);
          return {
            code,
            name: item.name || code,
            enabled: item.status === REGION_STATUS.ENABLED,
            assigned: allocatedAt !== undefined,
            assignedAt:
              allocatedAt !== undefined ? formatTime(allocatedAt) : null,
          };
        }
      );

      const agentItems: ResourceItem[] = (agentRes.data?.list || []).map(
        (item) => ({
          code: item.agent_line_code,
          name: item.agent_line_code,
          enabled: true,
          assigned: Boolean(item.allocated),
          assignedAt: item.allocated ? formatTime(item.allocated_at) : null,
        })
      );

      resourceState.items = {
        language: languageItems,
        region: regionItems,
        proxyLine: agentItems,
      };
      resourceState.selected = {
        language: languageItems
          .filter((item) => item.assigned)
          .map((item) => item.code),
        region: regionItems
          .filter((item) => item.assigned)
          .map((item) => item.code),
        proxyLine: agentItems
          .filter((item) => item.assigned)
          .map((item) => item.code),
      };
    } catch {
      resourceState.items = emptyResourceItems();
      resourceState.selected = { language: [], region: [], proxyLine: [] };
    } finally {
      resourcesLoading.value = false;
    }
  };

  const loadAllForOperator = (id: number) => {
    loadOperatorInfo(id);
    loadResources(id);
  };

  const handleOperatorChange = (value: unknown) => {
    const id = value as number | undefined;
    if (id) loadAllForOperator(id);
  };

  const handleResetFilter = () => {
    resourceState.keyword = '';
    resourceState.statusFilter = 'all';
  };

  const handleResetSelection = () => {
    if (!selectedOperatorId.value) return;
    loadResources(selectedOperatorId.value);
  };

  const saving = ref(false);
  const handleSave = async () => {
    if (!selectedOperatorId.value) {
      Message.warning(t('operatorBasicResource.modal.message.selectOperatorFirst'));
      return;
    }
    saving.value = true;
    try {
      await Promise.all([
        saveOperatorLanguageAllocation({
          operator_id: selectedOperatorId.value,
          language_codes: [...resourceState.selected.language],
        }),
        saveOperatorRegionAllocation({
          operator_id: selectedOperatorId.value,
          region_codes: [...resourceState.selected.region],
        }),
        saveOperatorAgentLineAllocation({
          operator_id: selectedOperatorId.value,
          agent_line_codes: [...resourceState.selected.proxyLine],
        }),
      ]);
      Message.success(t('operatorBasicResource.modal.message.saveSuccess'));
      emit('success');
    } finally {
      saving.value = false;
    }
  };

  const handleCancel = () => {
    emit('update:visible', false);
  };

  const resetLocalState = () => {
    selectedOperatorId.value = undefined;
    operatorInfo.value = null;
    resourceState.activeTab = 'language';
    resourceState.keyword = '';
    resourceState.statusFilter = 'all';
    resourceState.items = emptyResourceItems();
    resourceState.selected = { language: [], region: [], proxyLine: [] };
    resourcePagination.current = 1;
  };

  watch(
    () => props.visible,
    (visible) => {
      if (visible) {
        resetLocalState();
        loadOperatorOptions();
        if (props.operatorId) {
          selectedOperatorId.value = props.operatorId;
          loadAllForOperator(props.operatorId);
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
    width: 280px;
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
    gap: 12px;
    margin: 12px 0;
  }
  .resource-search {
    width: 220px;
  }
  .resource-status-select {
    width: 140px;
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
  .summary-item__desc {
    font-size: 12px;
    color: var(--color-text-3);
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
  .status-dot--disabled {
    background-color: rgb(var(--gray-5));
  }
</style>
