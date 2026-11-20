<template>
  <a-table
    class="activity-table"
    row-key="id"
    :data="data"
    :loading="loading"
    :pagination="false"
    :bordered="{ cell: false }"
    :scroll="{ x: compact ? 1180 : 1440, y: '100%' }"
    :scrollbar="false"
    :row-class="rowClass"
    @row-click="onRowClick"
  >
    <template #empty>
      <a-empty :description="t('discountCenter.activityCenter.emptyList')" />
    </template>
    <template #columns>
      <a-table-column
        :title="t('discountCenter.activityCenter.columns.sort')"
        data-index="sort_weight"
        :width="64"
      >
        <template #cell="{ record }">
          <span class="num muted">{{ record.sort_weight }}</span>
        </template>
      </a-table-column>
      <a-table-column
        :title="t('discountCenter.activityCenter.columns.activityId')"
        data-index="id"
        :width="80"
      >
        <template #cell="{ record }">
          <span class="num activity-id">{{ record.id }}</span>
        </template>
      </a-table-column>
      <a-table-column
        :title="t('discountCenter.activityCenter.columns.activityName')"
        :width="240"
      >
        <template #cell="{ record }">
          <div class="name-cell">
            <a-image
              class="thumb"
              :src="record.image || imageSrc"
              :width="36"
              :height="36"
              :preview="false"
              fit="cover"
            />
            <div class="stack">
              <strong class="ellipsis" :title="record.name">
                {{ displayText(record.name) }}
              </strong>
              <small class="ellipsis">{{ typeText(record.type, enumParams) }}</small>
            </div>
          </div>
        </template>
      </a-table-column>
      <a-table-column
        v-if="!compact"
        :title="t('discountCenter.activityCenter.columns.listed')"
        :width="96"
      >
        <template #cell="{ record }">
          <a-tag size="small" :color="record.show_in_list ? 'green' : 'gray'">
            {{
              t(
                `discountCenter.activityCenter.${
                  record.show_in_list ? 'yes' : 'no'
                }`
              )
            }}
          </a-tag>
        </template>
      </a-table-column>
      <a-table-column
        :title="t('discountCenter.activityCenter.columns.autoReviewConfig')"
        :width="130"
      >
        <template #cell="{ record }">
          {{ autoAuditText(asRecord(record), t) }}
        </template>
      </a-table-column>
      <a-table-column
        :title="t('discountCenter.activityCenter.columns.washMultiplier')"
        :width="110"
      >
        <template #cell="{ record }">
          <span :class="{ 'is-wash': !!record.wagering_multiplier }">
            {{ displayText(wageringText(asRecord(record), t)) }}
          </span>
        </template>
      </a-table-column>
      <a-table-column
        :title="t('discountCenter.activityCenter.card.participantLevel')"
        :width="150"
      >
        <template #cell="{ record }">
          <div class="stack">
            <span class="ellipsis" :title="levelText(asRecord(record))">
              {{ levelText(asRecord(record)) }}
            </span>
            <small
              class="ellipsis"
              :title="clientsText(record.allowed_clients, t, enumParams)"
            >
              {{ clientsText(record.allowed_clients, t, enumParams) }}
            </small>
          </div>
        </template>
      </a-table-column>
      <a-table-column
        :title="t('discountCenter.activityCenter.columns.activityTime')"
        :width="180"
      >
        <template #cell="{ record }">
          <span class="num">
            {{ displayText(activityTimeText(asRecord(record), t)) }}
          </span>
        </template>
      </a-table-column>
      <a-table-column
        v-if="!compact"
        :title="t('discountCenter.activityCenter.card.createdAt')"
        :width="164"
      >
        <template #cell="{ record }">
          <div class="stack">
            <span class="num muted">
              {{ displayText(formatTime(record.created_at_sec)) }}
            </span>
            <small class="ellipsis">
              {{ displayText(uidText(record.create_uid)) }}
            </small>
          </div>
        </template>
      </a-table-column>
      <a-table-column
        :title="t('discountCenter.activityCenter.columns.status')"
        :width="88"
      >
        <template #cell="{ record }">
          <a-tag size="small" :color="statusColor(record.status)">
            {{ statusText(record.status, enumParams) }}
          </a-tag>
        </template>
      </a-table-column>
      <a-table-column
        :title="t('discountCenter.activityCenter.columns.operations')"
        :width="120"
        fixed="right"
        align="center"
      >
        <template #cell="{ record }">
          <div class="ops" @click.stop>
            <a-button
              size="mini"
              @click="emit('action', asRecord(record), 'detail')"
            >
              {{ t('discountCenter.activityCenter.card.detail') }}
            </a-button>
            <a-dropdown
              position="br"
              @select="(value) => onAction(asRecord(record), value)"
            >
              <a-button
                size="mini"
                :disabled="!moreActions(asRecord(record)).length"
              >
                <template #icon><icon-more /></template>
              </a-button>
              <template #content>
                <a-doption
                  v-for="action in moreActions(asRecord(record))"
                  :key="action"
                  :value="action"
                  :class="{ 'danger-option': action === 'close' }"
                >
                  {{ t(`discountCenter.activityCenter.card.${action}`) }}
                </a-doption>
              </template>
            </a-dropdown>
          </div>
        </template>
      </a-table-column>
    </template>
  </a-table>
</template>

<script lang="ts" setup>
  import { useI18n } from 'vue-i18n';
  import type { TableData } from '@arco-design/web-vue';
  import {
    activityTimeText,
    autoAuditText,
    canClose,
    canEdit,
    canPublish,
    clientsText,
    displayText,
    formatTime,
    statusColor,
    statusText,
    typeText,
    uidText,
    wageringText,
  } from '../../options';
  import { useActivityStore } from '../../store';
  import type { ActivityAction, ActivityRecord } from '../../types';

  const props = defineProps<{
    data: ActivityRecord[];
    loading?: boolean;
    activeId: number | null;
    /** 活动没有图片时的占位图 */
    imageSrc: string;
    /** 详情面板打开、列表变窄时隐藏次要列 */
    compact?: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'select', record: ActivityRecord): void;
    (e: 'action', record: ActivityRecord, action: ActivityAction): void;
  }>();

  const { t } = useI18n();
  const { levelNames, enumParams } = useActivityStore();

  function asRecord(record: TableData) {
    return record as ActivityRecord;
  }

  function levelText(record: ActivityRecord) {
    const names = levelNames(record);
    return names.length
      ? names.join('、')
      : t('discountCenter.activityCenter.all');
  }

  function moreActions(record: ActivityRecord) {
    const list: ActivityAction[] = [];
    if (canEdit(record)) list.push('edit');
    if (canPublish(record)) list.push('publish');
    if (canClose(record)) list.push('close');
    return list;
  }

  function rowClass(record: TableData) {
    return asRecord(record).id === props.activeId ? 'is-active' : '';
  }

  function onRowClick(record: TableData) {
    emit('select', asRecord(record));
  }

  function onAction(
    record: ActivityRecord,
    value: string | number | Record<string, unknown> | undefined
  ) {
    emit('action', record, value as ActivityAction);
  }
</script>

<style scoped lang="less">
  // scroll.y = '100%' 时表头固定，表体占满父容器剩余高度并独立滚动
  .activity-table :deep(.arco-table-container) {
    flex: 1;
    min-height: 0;
  }

  .activity-table :deep(.arco-table-body) {
    flex: 1;
    min-height: 0;
  }

  .activity-table :deep(.arco-table-tr) {
    cursor: pointer;
  }

  .activity-table :deep(.arco-table-tr.is-active .arco-table-td) {
    background: rgb(var(--primary-1));
  }

  .activity-table :deep(.arco-table-th) {
    color: var(--color-text-2);
    font-weight: 500;
    background: var(--color-fill-2);
  }

  .activity-id {
    font-weight: 600;
  }

  .name-cell {
    display: flex;
    gap: 10px;
    align-items: center;
    min-width: 0;
  }

  .name-cell :deep(.thumb) {
    flex: none;
    overflow: hidden;
    border-radius: 6px;

    .arco-image-img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
  }

  .stack {
    display: flex;
    flex-direction: column;
    min-width: 0;
    line-height: 1.5;
  }

  .stack strong {
    font-weight: 600;
  }

  .stack small {
    color: var(--color-text-3);
    font-size: 12px;
  }

  .ellipsis {
    display: block;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .muted {
    color: var(--color-text-3);
  }

  .num {
    font-variant-numeric: tabular-nums;
  }

  .is-wash {
    color: rgb(var(--red-6));
    font-weight: 600;
  }

  .ops {
    display: inline-flex;
    gap: 6px;
    align-items: center;
  }

  .danger-option {
    color: rgb(var(--red-6));
  }
</style>
