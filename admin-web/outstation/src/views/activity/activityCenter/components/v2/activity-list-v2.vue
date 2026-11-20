<template>
  <div
    ref="rootRef"
    class="activity-v2"
    :class="{ 'has-panel': !!activeRecord, 'is-resizing': resizing }"
    :style="panelWidth ? { '--panel-width': `${panelWidth}px` } : undefined"
  >
    <div class="v2-main">
      <a-card class="general-card" :bordered="false">
        <ActivityFilterBar :hide-status="closedOnly" @search="onSearch" />
      </a-card>

      <a-card class="general-card table-card" :bordered="false">
        <div class="table-toolbar">
          <a-space wrap>
            <a-button type="primary" @click="openDialog(null)">
              <template #icon><icon-plus /></template>
              {{ t('discountCenter.activityCenter.operation.create') }}
            </a-button>
            <span class="total">
              {{
                t('discountCenter.activityCenter.v2.toolbar.total', {
                  count: total,
                })
              }}
            </span>
          </a-space>
          <LayoutSwitch />
        </div>

        <ActivityTable
          class="table-scroll"
          :data="activities"
          :loading="loading"
          :active-id="activeId"
          :image-src="activityImageSrc"
          :compact="!!activeRecord"
          @select="openPanel"
          @action="onAction"
        />

        <div class="table-footer">
          <div class="page-size">
            <span>{{ t('discountCenter.activityCenter.v2.pageSize') }}</span>
            <a-select
              :model-value="pagination.pageSize"
              size="small"
              class="page-size__select"
              :popup-container="'body'"
              @change="onPageSizeChange"
            >
              <a-option v-for="size in PAGE_SIZES" :key="size" :value="size">
                {{
                  t('discountCenter.activityCenter.v2.pageSizeOption', { size })
                }}
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

    <div v-if="activeRecord" class="v2-panel">
      <div
        class="panel-resizer"
        @pointerdown="startResize"
        @dblclick="resetPanelWidth"
      />
      <ActivityDetailPanel
        class="v2-panel__body"
        :record="activeRecord"
        :detail="detail"
        :loading="detailLoading"
        :image-src="activityImageSrc"
        @close="closePanel"
        @action="onPanelAction"
      />
    </div>

    <CreateActivityDialog
      v-model:visible="createVisible"
      :activity-id="editingId"
      @saved="refresh"
    />
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { ACTIVITY_STATUS } from '@/api/activity';
  import activityImageSrc from '@/assets/images/a.jpg';
  import CreateActivityDialog from '../create-activity-dialog.vue';
  import LayoutSwitch from '../layout-switch.vue';
  import ActivityDetailPanel from './activity-detail-panel.vue';
  import ActivityFilterBar from './activity-filter-bar.vue';
  import ActivityTable from './activity-table.vue';
  import useActivityActions from '../../actions';
  import { useActivityStore } from '../../store';
  import type {
    ActivityAction,
    ActivityQuery,
    ActivityRecord,
  } from '../../types';

  const PAGE_SIZES = [10, 20, 50];

  const props = defineProps<{
    /** 关闭列表只看已结束的活动 */
    closedOnly?: boolean;
  }>();

  const { t } = useI18n();
  const {
    activities,
    total,
    loading,
    fetchActivities,
    detail,
    detailLoading,
    fetchDetail,
    findActivity,
    refresh,
    loadLevels,
  } = useActivityStore();
  const { confirmPublish, confirmClose } = useActivityActions();

  const createVisible = ref(false);
  /** 为 null 时弹窗用于新增 */
  const editingId = ref<number | null>(null);

  function openDialog(id: number | null) {
    editingId.value = id;
    createVisible.value = true;
  }
  const pagination = reactive({ current: 1, pageSize: PAGE_SIZES[0] });

  const activeId = ref<number | null>(null);
  const activeRecord = computed(() => findActivity(activeId.value));

  const query = ref<ActivityQuery>({});

  function load() {
    fetchActivities(
      {
        ...query.value,
        status: props.closedOnly ? ACTIVITY_STATUS.ENDED : query.value.status,
      },
      { current: pagination.current, pageSize: pagination.pageSize }
    );
  }

  function onSearch(value: ActivityQuery) {
    query.value = value;
    pagination.current = 1;
    load();
  }

  function onPageChange(current: number) {
    pagination.current = current;
    load();
  }

  function onPageSizeChange(value: unknown) {
    pagination.pageSize = Number(value);
    pagination.current = 1;
    load();
  }

  function openPanel(record: ActivityRecord) {
    activeId.value = record.id;
    fetchDetail(record.id);
  }

  function closePanel() {
    activeId.value = null;
    fetchDetail(null);
  }

  function onAction(record: ActivityRecord, action: ActivityAction) {
    if (action === 'detail') openPanel(record);
    else if (action === 'edit') openDialog(record.id);
    else if (action === 'publish') confirmPublish(record);
    else if (action === 'close') confirmClose(record);
  }

  function onPanelAction(action: ActivityAction) {
    if (activeRecord.value) onAction(activeRecord.value, action);
  }

  const PANEL_WIDTH_KEY = 'activityCenter.v2.panelWidth';
  const PANEL_MIN_WIDTH = 400;
  const PANEL_MAX_WIDTH = 800;
  /** 拖拽时给左侧列表至少保留的宽度 */
  const MAIN_MIN_WIDTH = 560;

  const rootRef = ref<HTMLElement>();
  /** 为 null 时使用默认的响应式宽度 */
  const panelWidth = ref<number | null>(
    Number(localStorage.getItem(PANEL_WIDTH_KEY)) || null
  );
  const resizing = ref(false);

  function startResize(event: PointerEvent) {
    const panel = (event.currentTarget as HTMLElement).parentElement;
    const root = rootRef.value;
    if (!panel || !root) return;
    event.preventDefault();
    const startX = event.clientX;
    const startWidth = panel.offsetWidth;
    const maxWidth = Math.max(
      PANEL_MIN_WIDTH,
      Math.min(PANEL_MAX_WIDTH, root.clientWidth - MAIN_MIN_WIDTH)
    );
    resizing.value = true;

    const onMove = (e: PointerEvent) => {
      const width = startWidth + startX - e.clientX;
      panelWidth.value = Math.round(
        Math.min(maxWidth, Math.max(PANEL_MIN_WIDTH, width))
      );
    };
    const onUp = () => {
      resizing.value = false;
      window.removeEventListener('pointermove', onMove);
      window.removeEventListener('pointerup', onUp);
      if (panelWidth.value) {
        localStorage.setItem(PANEL_WIDTH_KEY, String(panelWidth.value));
      }
    };
    window.addEventListener('pointermove', onMove);
    window.addEventListener('pointerup', onUp);
  }

  function resetPanelWidth() {
    panelWidth.value = null;
    localStorage.removeItem(PANEL_WIDTH_KEY);
  }

  loadLevels();
  load();
</script>

<style scoped lang="less">
  .activity-v2 {
    display: grid;
    grid-template-rows: minmax(0, 1fr);
    grid-template-columns: minmax(0, 1fr);
    gap: 16px;

    &.has-panel {
      // 窗口缩小后保存的宽度可能过大，始终给列表留出 560px
      grid-template-columns: ~'minmax(0, 1fr) max(400px, min(var(--panel-width, clamp(400px, 25vw, 460px)), calc(100% - 576px)))';
    }

    &.is-resizing {
      cursor: col-resize;
      user-select: none;

      .v2-main,
      .v2-panel__body {
        pointer-events: none;
      }
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
    position: relative;
    display: flex;
    flex-direction: column;
    min-height: 0;
    margin-top: 16px;
  }

  .v2-panel__body {
    flex: 1;
  }

  .panel-resizer {
    position: absolute;
    top: 0;
    bottom: 0;
    left: -12px;
    z-index: 1;
    width: 8px;
    cursor: col-resize;

    &::after {
      position: absolute;
      top: 50%;
      left: 3px;
      width: 2px;
      height: 32px;
      background: var(--color-neutral-4);
      border-radius: 1px;
      transform: translateY(-50%);
      transition: height 0.2s, background-color 0.2s;
      content: '';
    }

    &:hover::after,
    .is-resizing &::after {
      height: 100%;
      background: rgb(var(--primary-6));
    }
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
    .activity-v2 {
      grid-template-rows: auto;

      &.has-panel {
        grid-template-columns: minmax(0, 1fr);
      }
    }

    .v2-panel {
      height: 720px;
      margin-top: 0;
    }

    .panel-resizer {
      display: none;
    }
  }
</style>
