<template>
  <div class="activity-list">
    <a-card class="general-card" :bordered="false">
      <a-form :model="formModel" layout="inline" class="search-form">
        <a-form-item field="keyword">
          <a-input-group compact>
            <a-select
              v-model="formModel.keywordField"
              class="search-type"
              :options="keywordFieldOptions"
            />
            <a-input
              v-model="formModel.keyword"
              class="search-input"
              allow-clear
              :placeholder="
                t('discountCenter.activityCenter.form.keyword.placeholder')
              "
              @press-enter="search"
            />
          </a-input-group>
        </a-form-item>
        <a-form-item
          :label="t('discountCenter.activityCenter.form.activityType')"
          field="activityType"
        >
          <a-select
            v-model="formModel.activityType"
            class="search-select"
            :options="activityTypeOptions"
            allow-clear
            :placeholder="t('discountCenter.activityCenter.form.selectDefault')"
          />
        </a-form-item>
        <a-form-item
          :label="t('discountCenter.activityCenter.form.client')"
          field="client"
        >
          <a-select
            v-model="formModel.client"
            class="search-select"
            :options="clientOptions"
            allow-clear
            :placeholder="t('discountCenter.activityCenter.form.selectDefault')"
          />
        </a-form-item>
        <a-form-item
          v-if="!closedOnly"
          :label="t('discountCenter.activityCenter.form.status')"
          field="status"
        >
          <a-select
            v-model="formModel.status"
            class="search-select"
            :options="statusOptions"
            allow-clear
            :placeholder="t('discountCenter.activityCenter.form.selectDefault')"
          />
        </a-form-item>
        <a-form-item>
          <a-space>
            <a-button type="primary" @click="search">
              <template #icon><icon-search /></template>
              {{ t('discountCenter.activityCenter.form.search') }}
            </a-button>
            <a-button @click="reset">
              <template #icon><icon-refresh /></template>
              {{ t('discountCenter.activityCenter.form.reset') }}
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </a-card>

    <a-card class="general-card table-card" :bordered="false">
      <div class="table-toolbar">
        <a-space>
          <a-button type="primary" @click="openDialog(null)">
            <template #icon><icon-plus /></template>
            {{ t('discountCenter.activityCenter.operation.create') }}
          </a-button>
          <a-button :disabled="!activities.length" @click="togglePageExpanded">
            <template #icon>
              <icon-up v-if="pageAllExpanded" />
              <icon-down v-else />
            </template>
            {{
              pageAllExpanded
                ? t('discountCenter.activityCenter.toolbar.collapsePage')
                : t('discountCenter.activityCenter.toolbar.expandPage')
            }}
          </a-button>
        </a-space>
        <LayoutSwitch />
      </div>

      <a-spin :loading="loading" class="activity-table-wrap">
        <div class="activity-panel">
          <div class="activity-scroll">
            <div class="activity-head">
              <span></span>
              <span v-for="key in HEAD_COLUMNS" :key="key">
                {{ t(`discountCenter.activityCenter.columns.${key}`) }}
              </span>
              <span class="head-action">
                {{ t('discountCenter.activityCenter.columns.operations') }}
              </span>
            </div>
            <ActivityCard
              v-for="record in activities"
              :key="record.id"
              class="activity-line"
              :record="record"
              :image-src="activityImageSrc"
              :expanded="isExpanded(record.id)"
              @update:expanded="(value) => setExpanded(record.id, value)"
              @action="(action) => onAction(record, action)"
            />
          </div>
          <a-empty
            v-if="!activities.length"
            class="activity-empty"
            :description="t('discountCenter.activityCenter.emptyList')"
          />
        </div>

        <div class="table-pagination">
          <a-pagination
            :total="total"
            :current="pagination.current"
            :page-size="pagination.pageSize"
            :page-size-options="pagination.pageSizeOptions"
            :show-total="pagination.showTotal"
            :show-page-size="pagination.showPageSize"
            :page-size-props="pagination.pageSizeProps"
            @change="handlePageChange"
            @page-size-change="handlePageSizeChange"
          />
        </div>
      </a-spin>
    </a-card>

    <CreateActivityDialog
      v-model:visible="createDialogVisible"
      :activity-id="editingId"
      @saved="refresh"
    />
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { SelectOptionData } from '@arco-design/web-vue/es/select/interface';
  import { ACTIVITY_STATUS } from '@/api/activity';
  import activityImageSrc from '@/assets/images/a.jpg';
  import ActivityCard from './activity-card.vue';
  import CreateActivityDialog from './create-activity-dialog.vue';
  import LayoutSwitch from './layout-switch.vue';
  import useActivityActions from '../actions';
  import { enumOptions } from '../options';
  import { useActivityStore } from '../store';
  import type { ActivityAction, ActivityQuery, ActivityRecord } from '../types';

  const props = defineProps<{
    /** 关闭列表只看已结束的活动 */
    closedOnly?: boolean;
  }>();

  const HEAD_COLUMNS = [
    'activity',
    'status',
    'activityTime',
    'scope',
    'reward',
    'display',
    'created',
  ];

  const { t } = useI18n();
  const {
    activities,
    total,
    loading,
    fetchActivities,
    refresh,
    loadLevels,
    enumParams,
  } = useActivityStore();
  const { confirmPublish, confirmClose } = useActivityActions();

  const generateFormModel = () => {
    return {
      keywordField: 'activityName' as 'activityName' | 'activityId',
      keyword: '',
      /** 下拉值为接口枚举编码，空字符串表示全部 */
      activityType: '' as number | '',
      client: '' as number | '',
      status: '' as number | '',
    };
  };

  const formModel = ref(generateFormModel());
  const expandedIds = ref<Set<number>>(new Set());
  const createDialogVisible = ref(false);
  /** 为 null 时弹窗用于新增 */
  const editingId = ref<number | null>(null);

  function openDialog(id: number | null) {
    editingId.value = id;
    createDialogVisible.value = true;
  }

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

  const keywordFieldOptions = computed<SelectOptionData[]>(() => [
    {
      label: t('discountCenter.activityCenter.form.keyword.field.activityName'),
      value: 'activityName',
    },
    {
      label: t('discountCenter.activityCenter.form.keyword.field.activityId'),
      value: 'activityId',
    },
  ]);

  const activityTypeOptions = computed<SelectOptionData[]>(() =>
    enumOptions(enumParams.value, 'type')
  );

  const clientOptions = computed<SelectOptionData[]>(() =>
    enumOptions(enumParams.value, 'client_type')
  );

  const statusOptions = computed<SelectOptionData[]>(() =>
    enumOptions(enumParams.value, 'status')
  );

  /** 按 ID 查找时 ID 无效返回 -1，由详情接口返回无结果 */
  function buildQuery(): ActivityQuery {
    const form = formModel.value;
    const keyword = form.keyword.trim();
    if (keyword && form.keywordField === 'activityId') {
      return { id: /^\d+$/.test(keyword) ? Number(keyword) : -1 };
    }
    return {
      name: keyword || undefined,
      type: form.activityType || undefined,
      client: form.client || undefined,
      status: props.closedOnly ? ACTIVITY_STATUS.ENDED : form.status || undefined,
    };
  }

  const appliedQuery = ref<ActivityQuery>(buildQuery());

  function load() {
    fetchActivities(appliedQuery.value, {
      current: pagination.current,
      pageSize: pagination.pageSize,
    });
  }

  const pageAllExpanded = computed(
    () =>
      activities.value.length > 0 &&
      activities.value.every((item) => expandedIds.value.has(item.id))
  );

  const search = () => {
    appliedQuery.value = buildQuery();
    pagination.current = 1;
    load();
  };

  const reset = () => {
    formModel.value = generateFormModel();
    search();
  };

  const handlePageChange = (current: number) => {
    pagination.current = current;
    load();
  };

  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.current = 1;
    load();
  };

  function isExpanded(id: number) {
    return expandedIds.value.has(id);
  }

  function setExpanded(id: number, expanded: boolean) {
    const next = new Set(expandedIds.value);
    if (expanded) next.add(id);
    else next.delete(id);
    expandedIds.value = next;
  }

  function togglePageExpanded() {
    const next = new Set(expandedIds.value);
    const collapse = pageAllExpanded.value;
    activities.value.forEach((item) => {
      if (collapse) next.delete(item.id);
      else next.add(item.id);
    });
    expandedIds.value = next;
  }

  function onAction(record: ActivityRecord, action: ActivityAction) {
    if (action === 'edit') openDialog(record.id);
    else if (action === 'publish') confirmPublish(record);
    else if (action === 'close') confirmClose(record);
  }

  loadLevels();
  load();
</script>

<style scoped lang="less">
  .general-card {
    margin-top: 16px;
  }

  .general-card :deep(.arco-card-body) {
    padding: 20px;
  }

  .search-form :deep(.arco-form-item) {
    margin-bottom: 12px;
  }

  .search-form :deep(.arco-select-view) {
    min-width: 160px;
  }

  .search-type :deep(.arco-select-view) {
    min-width: 112px;
  }

  .search-input {
    width: 170px;
  }

  // 外层 .list-page 固定一屏高且 overflow: hidden（见 global.less），
  // 这里逐层撑满剩余高度，由表格区域自己滚动
  .activity-list,
  .table-card,
  .table-card > :deep(.arco-card-body),
  .activity-table-wrap {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
  }

  .table-toolbar {
    display: flex;
    flex: none;
    flex-wrap: wrap;
    gap: 12px;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 14px;
  }

  .activity-table-wrap {
    width: 100%;
  }

  .activity-panel {
    // 名称内容短，多出的宽度优先给容易被截断的参与范围（客户端列表）
    --activity-cols: 28px minmax(200px, 1.1fr) 80px minmax(180px, 1fr)
      minmax(150px, 1.6fr) minmax(150px, 1fr) 84px minmax(130px, 0.9fr) 156px;

    flex: 0 1 auto;
    min-height: 0;
    overflow: auto;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 10px;
  }

  .activity-scroll {
    min-width: 1180px;
  }

  .activity-head {
    position: sticky;
    top: 0;
    z-index: 2;
    display: grid;
    grid-template-columns: var(--activity-cols);
    gap: 16px;
    align-items: center;
    min-height: 44px;
    padding: 0 20px 0 12px;
    color: var(--color-text-3);
    font-weight: 500;
    font-size: 12px;
    // 吸顶时下面的行会滚过表头，fill 色在暗色主题半透明，需垫一层不透明底色
    background: linear-gradient(var(--color-fill-2), var(--color-fill-2)),
      var(--color-bg-2);
    border-bottom: 1px solid var(--color-border-2);
  }

  .head-action {
    text-align: right;
  }

  .activity-line + .activity-line {
    border-top: 1px solid var(--color-border-1);
  }

  .activity-empty {
    padding: 32px 0;
  }

  .table-pagination {
    display: flex;
    flex: none;
    justify-content: flex-end;
    padding-top: 16px;
  }

  :deep(.arco-pagination-options .arco-select) {
    min-width: 100px;
  }
</style>
