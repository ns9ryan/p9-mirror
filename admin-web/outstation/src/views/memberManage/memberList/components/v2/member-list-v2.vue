<template>
  <div class="container member-v2">
    <div class="v2-layout" :class="{ 'has-panel': !!activeRecord }">
      <div class="v2-main">
        <a-card class="general-card" :bordered="false">
          <MemberFilterBar ref="filterRef" @search="onSearch" />
        </a-card>

        <a-card class="general-card table-card" :bordered="false">
          <div class="table-toolbar">
            <a-space wrap>
              <a-button type="primary" @click="createVisible = true">
                <template #icon><icon-plus /></template>
                {{ t('memberManage.v2.create') }}
              </a-button>
              <a-button @click="batchCreateVisible = true">
                <template #icon><icon-plus /></template>
                {{ t('memberManage.toolbar.batchCreate') }}
              </a-button>
              <a-dropdown @select="onBatch">
                <a-button>
                  {{ t('memberManage.toolbar.batch') }}
                  <icon-down />
                </a-button>
                <template #content>
                  <a-doption
                    v-for="command in BATCH_COMMAND_KEYS"
                    :key="command"
                    :value="command"
                  >
                    {{ t(BATCH_COMMANDS[command].label) }}
                  </a-doption>
                </template>
              </a-dropdown>
              <a-button @click="exportVisible = true">
                <template #icon><icon-download /></template>
                {{ t('memberManage.v2.toolbar.export') }}
              </a-button>
              <a-button :loading="loading" @click="refresh">
                <template #icon><icon-refresh /></template>
                {{ t('memberManage.v2.toolbar.refresh') }}
              </a-button>
              <span class="total">
                {{ t('memberManage.v2.toolbar.total', { count: totalText }) }}
              </span>
              <span v-if="selectedKeys.length" class="selected-count">
                {{
                  t('memberManage.toolbar.selected', {
                    count: selectedKeys.length,
                  })
                }}
              </span>
            </a-space>
            <div class="sort-control">
              <span>{{ t('memberManage.v2.toolbar.sort') }}：</span>
              <a-dropdown @select="onSortKey">
                <a-link>{{ t(`memberManage.v2.sort.${sortKey}`) }}</a-link>
                <template #content>
                  <a-doption
                    v-for="key in SORT_KEYS"
                    :key="key"
                    :value="key"
                    :class="{ 'is-current': key === sortKey }"
                  >
                    {{ t(`memberManage.v2.sort.${key}`) }}
                  </a-doption>
                </template>
              </a-dropdown>
              <a-button
                type="text"
                size="mini"
                :title="t('memberManage.v2.toolbar.toggleOrder')"
                :disabled="sortKey === 'registeredAt'"
                @click="toggleOrder"
              >
                <template #icon>
                  <icon-arrow-down
                    v-if="sortKey === 'registeredAt' || sortOrder === 'desc'"
                  />
                  <icon-arrow-up v-else />
                </template>
              </a-button>
              <LayoutSwitch class="layout-switch" />
            </div>
          </div>

          <MemberTable
            class="table-scroll"
            :data="members"
            :loading="loading"
            :selected-keys="selectedKeys"
            :active-id="activeId"
            :compact="!!activeRecord"
            @update:selected-keys="selectedKeys = $event"
            @select="openPanel"
            @detail="openPanel"
            @command="runCommand"
          />

          <div class="table-footer">
            <div class="page-size">
              <span>{{ t('memberManage.v2.pageSize') }}</span>
              <a-select
                :model-value="pagination.pageSize"
                size="small"
                class="page-size__select"
                :popup-container="'body'"
                @change="onPageSizeChange"
              >
                <a-option
                  v-for="size in PAGE_SIZES"
                  :key="size"
                  :value="size"
                >
                  {{ t('memberManage.v2.pageSizeOption', { size }) }}
                </a-option>
              </a-select>
            </div>
            <a-pagination
              :total="total"
              :current="pagination.current"
              :page-size="pagination.pageSize"
              show-total
              show-jumper
              @change="onPageChange"
            />
          </div>
        </a-card>
      </div>

      <MemberDetailPanel
        v-if="activeRecord"
        class="v2-panel"
        :record="activeRecord"
        :detail="activeDetail"
        :avatar-url="activeAvatarUrl"
        :loading="detailLoading"
        @close="activeId = null"
        @edit="onPanelCommand('edit')"
        @command="onPanelCommand"
      />
    </div>

    <MemberCreateModal
      :visible="createVisible"
      @update:visible="createVisible = $event"
      @created="onCreated"
    />
    <MemberBatchCreateModal
      :visible="batchCreateVisible"
      @update:visible="batchCreateVisible = $event"
      @refresh="onBatchCreated"
    />
    <MemberFormModal
      :visible="formVisible"
      :record="editingRecord"
      @update:visible="formVisible = $event"
      @saved="onSaved"
    />
    <MemberExportModal
      :visible="exportVisible"
      @update:visible="exportVisible = $event"
    />
    <MemberOperations ref="operationsRef" />
  </div>
</template>

<script lang="ts" setup>
  import { computed, onMounted, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, Modal } from '@arco-design/web-vue';
  import LayoutSwitch from '../layout-switch.vue';
  import MemberCreateModal from '../member-create-modal.vue';
  import MemberBatchCreateModal from '../member-batch-create-modal.vue';
  import MemberExportModal from '../member-export-modal.vue';
  import MemberFormModal from '../member-form-modal.vue';
  import MemberOperations from '../operations/member-operations.vue';
  import { isOperationCommand } from '../operations/commands';
  import MemberDetailPanel from './member-detail-panel.vue';
  import MemberFilterBar from './member-filter-bar.vue';
  import MemberTable from './member-table.vue';
  import {
    buildMemberBriefParams,
    buildSortParams,
    createMemberBriefQuery,
  } from '../../query';
  import { useMemberStore } from '../../store';
  import type {
    MemberBriefQuery,
    MemberRecord,
    MemberSortKey,
    SortOrder,
  } from '../../types';
  import {
    BATCH_COMMANDS,
    BATCH_COMMAND_KEYS,
    STATUS_COMMANDS,
    type BatchCommand,
    type MemberCommand,
  } from './actions';

  const SORT_KEYS: MemberSortKey[] = [
    'registeredAt',
    'totalDeposit',
    'depositCount',
    'totalWithdraw',
    'withdrawCount',
  ];
  const PAGE_SIZES = [10, 20, 50];

  const { t } = useI18n();
  const {
    members,
    total,
    loading,
    fetchMembers,
    detail,
    detailLoading,
    fetchDetail,
    findMember,
    updateStatus,
    images,
    loadImages,
  } = useMemberStore();

  const filterRef = ref<InstanceType<typeof MemberFilterBar>>();
  const operationsRef = ref<InstanceType<typeof MemberOperations>>();
  const applied = ref<MemberBriefQuery>(createMemberBriefQuery());
  const sortKey = ref<MemberSortKey>('registeredAt');
  const sortOrder = ref<SortOrder>('desc');
  const selectedKeys = ref<number[]>([]);
  const pagination = reactive({ current: 1, pageSize: PAGE_SIZES[0] });

  const createVisible = ref(false);
  const batchCreateVisible = ref(false);
  const exportVisible = ref(false);
  const formVisible = ref(false);
  const editingId = ref<number | null>(null);

  const totalText = computed(() => total.value.toLocaleString());

  const activeId = ref<number | null>(null);
  const activeDetail = computed(() =>
    detail.value?.id === activeId.value ? detail.value : null
  );
  const activeRecord = computed(
    () => activeDetail.value || findMember(activeId.value)
  );
  const editingRecord = computed(() => findMember(editingId.value));

  const activeAvatarUrl = computed(() => {
    const avatarId = activeDetail.value?.avatarId;
    if (!avatarId) return '';
    return (
      images.value.avatars.find((item) => item.id === avatarId)?.image_url || ''
    );
  });

  watch(activeId, (id) => {
    fetchDetail(id);
    if (id != null) loadImages();
  });

  async function loadMembers() {
    selectedKeys.value = [];
    await fetchMembers({
      page: pagination.current,
      page_size: pagination.pageSize,
      ...buildMemberBriefParams(applied.value),
      ...buildSortParams(sortKey.value, sortOrder.value),
    });
  }

  function reloadFirstPage() {
    pagination.current = 1;
    loadMembers();
  }

  onMounted(loadMembers);

  function onSearch(query: MemberBriefQuery) {
    applied.value = query;
    reloadFirstPage();
  }

  function onSortKey(
    value: string | number | Record<string, unknown> | undefined
  ) {
    sortKey.value = value as MemberSortKey;
    reloadFirstPage();
  }

  function toggleOrder() {
    sortOrder.value = sortOrder.value === 'desc' ? 'asc' : 'desc';
    reloadFirstPage();
  }

  function onPageChange(current: number) {
    pagination.current = current;
    loadMembers();
  }

  function onPageSizeChange(value: unknown) {
    pagination.pageSize = Number(value);
    reloadFirstPage();
  }

  function openPanel(record: MemberRecord) {
    activeId.value = record.id;
  }

  function openEdit(record: MemberRecord) {
    editingId.value = record.id;
    formVisible.value = true;
  }

  function runCommand(record: MemberRecord, command: MemberCommand) {
    if (command === 'edit') {
      openEdit(record);
      return;
    }
    if (isOperationCommand(command)) {
      operationsRef.value?.open(command, record);
      return;
    }
    const meta = STATUS_COMMANDS[command];
    Modal.confirm({
      title: t(meta.confirmTitle),
      content: t(meta.confirmContent, { id: record.code }),
      okText: t('memberManage.confirm.ok'),
      cancelText: t('memberManage.confirm.cancel'),
      onBeforeOk: async () => {
        const { changed } = await updateStatus([record.id], meta.target);
        if (changed) Message.success(t(meta.success));
        return true;
      },
    });
  }

  function onPanelCommand(command: MemberCommand) {
    if (activeRecord.value) runCommand(activeRecord.value, command);
  }

  function onBatch(
    value: string | number | Record<string, unknown> | undefined
  ) {
    if (!selectedKeys.value.length) {
      Message.warning(t('memberManage.message.selectFirst'));
      return;
    }
    const meta = BATCH_COMMANDS[value as BatchCommand];
    const ids = [...selectedKeys.value];
    Modal.confirm({
      title: t(meta.confirmTitle),
      content: t(meta.confirmContent, { count: ids.length }),
      okText: t('memberManage.confirm.ok'),
      cancelText: t('memberManage.confirm.cancel'),
      onBeforeOk: async () => {
        const { changed, failed } = await updateStatus(ids, meta.target);
        if (failed) {
          Message.warning(
            t('memberManage.message.batchPartial', { changed, failed })
          );
        } else if (!changed) {
          Message.info(t('memberManage.message.already'));
        } else {
          Message.success(t(meta.success, { count: changed }));
        }
        return true;
      },
    });
  }

  async function refresh() {
    await Promise.all([
      loadMembers(),
      activeId.value != null && fetchDetail(activeId.value),
    ]);
    Message.success(t('memberManage.v2.message.refreshed'));
  }

  /** 回到默认条件第一页，新会员按注册时间倒序排在最前 */
  /** 批量创建成功或超时后刷新：按注册时间倒序回到第一页，便于看到新会员 */
  async function onBatchCreated() {
    sortKey.value = 'registeredAt';
    sortOrder.value = 'desc';
    applied.value = createMemberBriefQuery();
    filterRef.value?.clear();
    pagination.current = 1;
    await loadMembers();
  }

  async function onCreated(result: { id: number }) {
    Message.success(t('memberManage.message.created'));
    sortKey.value = 'registeredAt';
    sortOrder.value = 'desc';
    applied.value = createMemberBriefQuery();
    filterRef.value?.clear();
    pagination.current = 1;
    await loadMembers();
    activeId.value = result.id;
  }

  function onSaved() {
    Message.success(t('memberManage.message.updated'));
  }
</script>

<style scoped lang="less">
  .container {
    padding: 0 20px 20px;
  }

  // 整页限定在一屏内，列表和详情各自滚动：
  // 导航栏 60px + 页签栏 40px + 内容区上边距 12px + 页脚 30px
  @layout-offset: 142px;
  @page-min-height: 560px;

  .member-v2 {
    box-sizing: border-box;
    height: ~'max(@{page-min-height}, calc(100vh - @{layout-offset}))';
    padding-bottom: 16px;
  }

  .v2-layout {
    display: grid;
    grid-template-rows: minmax(0, 1fr);
    grid-template-columns: minmax(0, 1fr);
    gap: 16px;
    height: 100%;

    &.has-panel {
      grid-template-columns: minmax(0, 1fr) clamp(420px, 25vw, 480px);
    }
  }

  .v2-main {
    display: flex;
    flex-direction: column;
    min-width: 0;
    min-height: 0;
  }

  .general-card {
    flex: none;
    margin-top: 16px;
  }

  .general-card :deep(.arco-card-body) {
    padding: 20px;
  }

  .table-card {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;

    :deep(.arco-card-body) {
      display: flex;
      flex: 1;
      flex-direction: column;
      min-height: 0;
    }
  }

  .table-scroll {
    flex: 1;
    min-height: 0;
  }

  .v2-panel {
    min-height: 0;
    margin-top: 16px;
  }

  .table-toolbar {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 14px;
  }

  .total {
    color: var(--color-text-1);
    font-weight: 500;
  }

  .selected-count {
    color: var(--color-text-2);
    font-size: 13px;
  }

  .sort-control {
    display: inline-flex;
    gap: 2px;
    align-items: center;
    color: var(--color-text-2);
    font-size: 13px;
  }

  .layout-switch {
    margin-left: 12px;
  }

  .is-current {
    color: rgb(var(--primary-6));
  }

  .table-footer {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    align-items: center;
    justify-content: space-between;
    padding-top: 16px;
  }

  .page-size {
    display: inline-flex;
    flex-shrink: 0;
    gap: 8px;
    align-items: center;
    color: var(--color-text-2);
    font-size: 13px;
    white-space: nowrap;
  }

  .page-size__select {
    flex-shrink: 0;
    width: 100px;
  }

  // 窄屏上下堆叠时恢复整页滚动，列表按内容撑开，详情限高后内部滚动
  @media (max-width: 1280px) {
    .member-v2 {
      height: auto;
    }

    .v2-layout {
      grid-template-rows: auto;
      height: auto;

      &.has-panel {
        grid-template-columns: minmax(0, 1fr);
      }
    }

    .v2-panel {
      height: 720px;
      margin-top: 0;
    }
  }
</style>
