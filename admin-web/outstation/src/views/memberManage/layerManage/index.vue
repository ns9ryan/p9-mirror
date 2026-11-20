<template>
  <div class="container list-page">
    <a-card class="general-card search-card" :bordered="false">
      <a-tabs
        v-model:active-key="activeKind"
        class="layer-tabs"
        hide-content
        :header-padding="false"
      >
        <a-tab-pane key="auto" :title="t('memberManage.levelKind.auto')" />
        <a-tab-pane key="fixed" :title="t('memberManage.levelKind.fixed')" />
      </a-tabs>
      <a-alert
        v-if="activeKind === 'auto'"
        class="layer-tip"
        closable
        :title="t('layerManage.tip.autoTitle')"
      >
        {{ t('layerManage.tip.autoDesc') }}
      </a-alert>
      <div class="search-bar">
        <a-form :model="form" layout="inline" class="search-form">
          <a-form-item :label="t('layerManage.search.name')">
            <a-input
              v-model="form.name"
              class="search-input"
              :max-length="100"
              allow-clear
              :placeholder="t('layerManage.form.namePlaceholder')"
              @press-enter="handleSearch"
            />
          </a-form-item>
          <a-form-item>
            <a-space :size="12">
              <a-button type="primary" @click="handleSearch">
                <template #icon><icon-search /></template>
                {{ t('front.common.search.search') }}
              </a-button>
              <a-button @click="handleReset">
                <template #icon><icon-refresh /></template>
                {{ t('front.common.search.reset') }}
              </a-button>
            </a-space>
          </a-form-item>
        </a-form>
        <a-button type="primary" @click="openCreate">
          <template #icon><icon-plus /></template>
          {{ createText }}
        </a-button>
      </div>
    </a-card>

    <a-card class="general-card table-card" :bordered="false">
      <div class="list-title">{{ listTitle }}</div>
      <a-table
        row-key="id"
        :columns="columns"
        :data="layers"
        :loading="loading"
        :pagination="false"
        :bordered="false"
        :scroll="{ x: 1080 }"
      >
        <template #index="{ rowIndex }">
          {{ (pageNo - 1) * pageSize + rowIndex + 1 }}
        </template>
        <template #deposit="{ record }">
          {{ formatAmount(record.deposit ?? 0) }}
        </template>
        <template #wash="{ record }">
          {{ formatAmount(record.wash ?? 0) }}
        </template>
        <template #description="{ record }">
          {{ record.description || '--' }}
        </template>
        <template #withdrawLimit="{ record }">
          {{
            record.withdrawLimit == null
              ? t('layerManage.withdraw.unlimited')
              : formatAmount(record.withdrawLimit)
          }}
        </template>
        <template #count="{ record }">
          <a-link class="count-link" @click="openMembers(record)">
            {{ formatCount(record.memberCount) }}
          </a-link>
        </template>
        <template #operations="{ record }">
          <span v-if="record.isDefault">--</span>
          <a-space v-else :size="20" class="operation-links">
            <a-link @click="openEdit(record)">
              {{ t('front.common.action.edit') }}
            </a-link>
            <a-link v-if="record.kind === 'fixed'" @click="openMembers(record)">
              {{ t('layerManage.action.members') }}
            </a-link>
            <a-link status="danger" @click="confirmDelete(record)">
              {{ t('front.common.action.delete') }}
            </a-link>
          </a-space>
        </template>
        <template #empty>
          <a-empty :description="t('layerManage.empty')" />
        </template>
      </a-table>
      <div class="table-pagination">
        <a-pagination
          v-model:current="pageNo"
          v-model:page-size="pageSize"
          :total="total"
          :page-size-options="[15, 30, 50, 100]"
          :page-size-props="{ popupContainer: 'body' }"
          show-total
          show-page-size
        />
      </div>
    </a-card>

    <LayerFormModal
      v-model:visible="formVisible"
      :mode="formMode"
      :kind="activeKind"
      :record="editing"
      :names="takenNames"
      :auto-layers="autoLayers"
      @submit="onSubmit"
    />

    <LayerDeleteModal
      v-model:visible="deleteVisible"
      :record="deleting"
      @confirm="onDelete"
    />

    <a-drawer
      :visible="drawerRecord !== null"
      :width="640"
      :footer="false"
      unmount-on-close
      :title="drawerTitle"
      @cancel="closeMembers"
    >
      <template v-if="drawerRecord">
        <div class="drawer-summary">
          {{
            t('layerManage.drawer.count', { count: drawerRecord.memberCount })
          }}
        </div>
        <div v-if="drawerMembers.length" class="drawer-note">
          {{ t('layerManage.drawer.sample') }}
        </div>
        <a-table
          v-if="drawerMembers.length"
          row-key="id"
          :columns="memberColumns"
          :data="drawerMembers"
          :pagination="false"
          :bordered="true"
        >
          <template #account="{ record }">
            {{ memberAccount(record) }}
          </template>
        </a-table>
        <a-empty v-else :description="t('layerManage.drawer.empty')" />
      </template>
    </a-drawer>
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { formatTime, siteNow } from '@/utils/site-time';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import {
    createAutoLevel,
    createFixedLevel,
    deleteFixedLevel,
    queryAutoLevelList,
    queryFixedLevelList,
    updateFixedLevel,
  } from '@/api/member';
  import type { AutoLevelRecord, FixedLevelRecord } from '@/api/member';
  import { fixedLevelsText, formatAmount } from '../memberList/format';
  import createMemberSeed from '../memberList/mock';
  import type { MemberRecord } from '../memberList/types';
  import LayerFormModal from './components/layer-form-modal.vue';
  import LayerDeleteModal from './components/layer-delete-modal.vue';
  import type { LayerFormValue, LayerKind, LayerRecord } from './types';

  const { t } = useI18n();
  const activeKind = ref<LayerKind>('auto');
  /** 当前 Tab 当前页的数据 */
  const layers = ref<LayerRecord[]>([]);
  const memberSeed = createMemberSeed();
  const total = ref(0);
  const loading = ref(false);
  let loadTicket = 0;

  const form = reactive({ name: '' });
  const keyword = ref('');
  const pageNo = ref(1);
  const pageSize = ref(15);

  const formVisible = ref(false);
  const formMode = ref<'create' | 'edit'>('create');
  const editing = ref<LayerRecord | null>(null);
  const drawerId = ref<number | null>(null);
  const deleteVisible = ref(false);
  const deleting = ref<LayerRecord | null>(null);

  const createText = computed(() =>
    t(
      activeKind.value === 'auto'
        ? 'layerManage.toolbar.createAuto'
        : 'layerManage.toolbar.createFixed'
    )
  );
  const listTitle = computed(() => t(`layerManage.list.${activeKind.value}`));

  const autoLayers = computed(() =>
    layers.value.filter((item) => item.kind === 'auto')
  );

  watch(activeKind, () => {
    layers.value = [];
    total.value = 0;
    pageNo.value = 1;
  });

  function toAutoLayer(item: AutoLevelRecord): LayerRecord {
    return {
      id: item.id,
      kind: 'auto',
      name: item.name,
      isDefault: !item.min_deposit && !item.min_water,
      isSystem: false,
      description: '',
      deposit: item.min_deposit,
      wash: item.min_water,
      withdrawLimit: null,
      memberCount: item.current_member_count,
      createdAt: formatTime(item.created_at),
      updatedAt: formatTime(item.updated_at),
    };
  }

  function toFixedLayer(item: FixedLevelRecord): LayerRecord {
    return {
      id: item.id,
      kind: 'fixed',
      name: item.name,
      isDefault: false,
      isSystem: false,
      description: item.description,
      deposit: null,
      wash: null,
      withdrawLimit: item.withdrawal_limit,
      memberCount: item.current_member_count,
      createdAt: formatTime(item.created_at),
      updatedAt: formatTime(item.updated_at),
    };
  }

  async function fetchLayers(kind: LayerKind) {
    const params = {
      page: pageNo.value,
      page_size: pageSize.value,
      name: keyword.value || undefined,
    };
    if (kind === 'auto') {
      const { data } = await queryAutoLevelList(params);
      return { list: (data?.list || []).map(toAutoLayer), total: data?.total };
    }
    const { data } = await queryFixedLevelList(params);
    return { list: (data?.list || []).map(toFixedLayer), total: data?.total };
  }

  async function loadLayers() {
    loadTicket += 1;
    const ticket = loadTicket;
    loading.value = true;
    try {
      const res = await fetchLayers(activeKind.value);
      if (ticket !== loadTicket) return;
      layers.value = res.list;
      total.value = res.total || 0;
    } catch {
      if (ticket !== loadTicket) return;
      layers.value = [];
      total.value = 0;
    } finally {
      if (ticket === loadTicket) loading.value = false;
    }
  }

  watch(
    () => [activeKind.value, pageNo.value, pageSize.value] as const,
    () => loadLayers(),
    { immediate: true }
  );

  const takenNames = computed(() =>
    layers.value
      .filter(
        (item) =>
          item.kind === activeKind.value && item.id !== editing.value?.id
      )
      .map((item) => item.name)
  );

  const columns = computed<TableColumnData[]>(() => {
    const head: TableColumnData[] = [
      {
        title: t('layerManage.column.index'),
        slotName: 'index',
        width: 80,
        align: 'center',
      },
      {
        title: t('layerManage.column.name'),
        dataIndex: 'name',
        width: 140,
      },
    ];
    const updatedAt: TableColumnData = {
      title: t('layerManage.column.updatedAt'),
      dataIndex: 'updatedAt',
      width: 170,
    };
    if (activeKind.value === 'fixed') {
      return [
        ...head,
        {
          title: t('layerManage.column.description'),
          slotName: 'description',
          width: 240,
          ellipsis: true,
          tooltip: true,
        },
        {
          title: t('layerManage.column.linkedCount'),
          slotName: 'count',
          width: 120,
          align: 'center',
        },
        {
          title: t('layerManage.column.withdrawLimit'),
          slotName: 'withdrawLimit',
          width: 140,
        },
        updatedAt,
        {
          title: t('layerManage.column.operations'),
          slotName: 'operations',
          width: 190,
        },
      ];
    }
    return [
      ...head,
      {
        title: t('layerManage.column.deposit'),
        slotName: 'deposit',
        width: 160,
      },
      {
        title: t('layerManage.column.wash'),
        slotName: 'wash',
        width: 160,
      },
      {
        title: t('layerManage.column.count'),
        slotName: 'count',
        width: 120,
        align: 'center',
      },
      {
        title: t('layerManage.column.createdAt'),
        dataIndex: 'createdAt',
        width: 170,
      },
      updatedAt,
      {
        title: t('layerManage.column.operations'),
        slotName: 'operations',
        width: 120,
      },
    ];
  });

  const drawerRecord = computed(
    () => layers.value.find((item) => item.id === drawerId.value) ?? null
  );

  const memberColumns = computed<TableColumnData[]>(() => [
    {
      title: t('memberManage.card.memberId'),
      dataIndex: 'id',
      width: 140,
    },
    {
      title: t('memberManage.card.account'),
      slotName: 'account',
    },
    {
      title: t('memberManage.card.level'),
      width: 140,
      render: ({ record }) => {
        const member = record as MemberRecord;
        return drawerRecord.value?.kind === 'fixed'
          ? fixedLevelsText(member)
          : member.autoLevelName;
      },
    },
  ]);

  const drawerTitle = computed(() => {
    if (!drawerRecord.value) return t('layerManage.drawer.title');
    return `${t('layerManage.drawer.title')} · ${drawerRecord.value.name}`;
  });

  const drawerMembers = computed(() => {
    const layer = drawerRecord.value;
    if (!layer) return [];
    return memberSeed.filter((item) =>
      layer.kind === 'auto'
        ? item.autoLevelName === layer.name
        : item.fixedLevelNames.includes(layer.name)
    );
  });

  function formatCount(value: number) {
    return value.toLocaleString('en-US');
  }

  function memberAccount(record: MemberRecord) {
    return record.account || record.email || record.phone || '--';
  }

  function handleSearch() {
    keyword.value = form.name.trim();
    if (pageNo.value !== 1) {
      pageNo.value = 1;
      return;
    }
    loadLayers();
  }

  function handleReset() {
    form.name = '';
    handleSearch();
  }

  function openCreate() {
    formMode.value = 'create';
    editing.value = null;
    formVisible.value = true;
  }

  function openEdit(record: LayerRecord) {
    formMode.value = 'edit';
    editing.value = record;
    formVisible.value = true;
  }

  async function onSubmit(value: LayerFormValue) {
    const now = siteNow().format('YYYY-MM-DD HH:mm:ss');
    if (formMode.value === 'create' && activeKind.value === 'auto') {
      await createAutoLevel({
        name: value.name,
        min_deposit: value.deposit ?? 0,
        min_water: value.wash ?? 0,
      });
      Message.success(t('layerManage.message.created'));
      loadLayers();
      return;
    }
    if (formMode.value === 'create') {
      await createFixedLevel({
        name: value.name,
        description: value.description ?? '',
        withdrawal_limit: value.withdrawLimit ?? null,
      });
      Message.success(t('layerManage.message.created'));
      loadLayers();
      return;
    }
    const current = editing.value;
    if (!current) return;
    if (current.kind === 'fixed') {
      await updateFixedLevel({
        id: current.id,
        name: value.name,
        description: value.description ?? '',
        withdrawal_limit: value.withdrawLimit ?? null,
      });
      Message.success(t('layerManage.message.updated'));
      loadLayers();
      return;
    }
    layers.value = layers.value.map((item) => {
      if (item.id !== current.id) return item;
      return {
        ...item,
        name: value.name,
        deposit: item.isDefault ? item.deposit : value.deposit ?? 0,
        wash: item.isDefault ? item.wash : value.wash ?? 0,
        updatedAt: now,
      };
    });
    Message.success(t('layerManage.message.updated'));
  }

  function confirmDelete(record: LayerRecord) {
    deleting.value = record;
    deleteVisible.value = true;
  }

  async function onDelete(record: LayerRecord) {
    if (record.kind === 'fixed') {
      await deleteFixedLevel(record.id);
      if (drawerId.value === record.id) drawerId.value = null;
      Message.success(t('layerManage.message.deleted'));
      if (layers.value.length === 1 && pageNo.value > 1) pageNo.value -= 1;
      else loadLayers();
      return;
    }
    layers.value = layers.value.filter((item) => item.id !== record.id);
    total.value = Math.max(0, total.value - 1);
    if (drawerId.value === record.id) drawerId.value = null;
    Message.success(t('layerManage.message.deleted'));
  }

  function openMembers(record: LayerRecord) {
    drawerId.value = record.id;
  }

  function closeMembers() {
    drawerId.value = null;
  }
</script>

<style scoped lang="less">
  .general-card {
    margin-top: 16px;
  }

  .search-card :deep(.arco-card-body) {
    padding: 12px 20px 20px;
  }

  .table-card :deep(.arco-card-body) {
    padding: 20px;
  }

  .layer-tabs {
    :deep(.arco-tabs-nav-type-line .arco-tabs-tab) {
      margin: 0 28px 0 0;
      padding: 0 0 8px;
      line-height: 22px;
    }

    :deep(.arco-tabs-nav::before) {
      right: 0;
      left: 0;
    }
  }

  .layer-tip {
    margin-top: 16px;
    border-radius: 6px;

    :deep(.arco-alert-title) {
      font-weight: 600;
    }

    :deep(.arco-alert-content) {
      color: var(--color-text-2);
      font-size: 13px;
    }
  }

  .search-bar {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    align-items: center;
    justify-content: space-between;
    margin-top: 16px;
    padding: 16px 20px;
    background: var(--color-fill-1);
    border-radius: 6px;
  }

  .search-form {
    flex: 1;
    min-width: 0;

    :deep(.arco-form-item) {
      margin-bottom: 0;
    }

    :deep(.arco-form-item-label-col) {
      padding-right: 8px;
    }
  }

  .search-input {
    width: 200px;
  }

  .list-title {
    margin-bottom: 16px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 16px;
  }

  .count-link {
    padding: 0;
  }

  .operation-links :deep(.arco-link) {
    padding: 0;
  }

  .table-pagination {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }

  :deep(.arco-pagination-options .arco-select) {
    min-width: 100px;
  }

  .drawer-summary {
    margin-bottom: 8px;
    color: var(--color-text-1);
    font-weight: 500;
  }

  .drawer-note {
    margin-bottom: 12px;
    color: var(--color-text-3);
    font-size: 12px;
  }
</style>
