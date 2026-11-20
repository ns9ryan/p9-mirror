<template>
  <div class="panel-header">
    <div class="list-title">{{ t('vipManage.icon.title') }}</div>
    <a-button
      :loading="loading"
      :disabled="busy != null"
      @click="emit('refresh')"
    >
      <template #icon><icon-refresh /></template>
      {{ t('front.common.toolbar.refresh') }}
    </a-button>
  </div>
  <a-table
    row-key="no"
    :columns="columns"
    :data="groups"
    :loading="loading"
    :pagination="false"
    :bordered="{ cell: true }"
    :scroll="{ x: 1040 }"
  >
    <template #range="{ record }">{{ rangeText(record) }}</template>
    <template #defaultIcon="{ record }">
      <div class="icon-box">
        <img
          v-if="showImage(record.defaultIcon)"
          :src="record.defaultIcon"
          alt=""
          @error="markBroken(record.defaultIcon)"
        />
        <span v-else class="icon-empty">{{ t('vipManage.icon.empty') }}</span>
      </div>
    </template>
    <template #current="{ record }">
      <div class="current-cell">
        <div class="icon-box" :class="{ 'is-draft': drafts[record.no] }">
          <img
            v-if="showImage(displayIcon(record))"
            :src="displayIcon(record)"
            alt=""
            @error="markBroken(displayIcon(record))"
          />
          <span v-else class="icon-empty">{{ t('vipManage.icon.empty') }}</span>
        </div>
        <a-tag v-if="drafts[record.no]" size="small" color="orangered">
          {{ t('vipManage.icon.pending') }}
        </a-tag>
        <div class="current-note">
          {{ t('vipManage.icon.currentNote', { range: rangeText(record) }) }}
        </div>
      </div>
    </template>
    <template #operations="{ record }">
      <a-space :size="8" wrap>
        <a-upload
          :show-file-list="false"
          accept=".png,image/png"
          :disabled="busy != null || validatingNo != null"
          :custom-request="(option) => onUpload(record, option)"
        >
          <template #upload-button>
            <a-button
              :type="drafts[record.no] ? 'secondary' : 'primary'"
              :loading="isBusy(record, 'upload') || validatingNo === record.no"
              :disabled="isLocked(record, 'upload') || validatingNo != null"
            >
              <template #icon><icon-upload /></template>
              {{
                drafts[record.no]
                  ? t('vipManage.icon.reupload')
                  : t('vipManage.icon.upload')
              }}
            </a-button>
          </template>
        </a-upload>
        <template v-if="drafts[record.no]">
          <a-button
            type="primary"
            :loading="isBusy(record, 'save')"
            :disabled="isLocked(record, 'save')"
            @click="emit('save', record)"
          >
            <template #icon><icon-save /></template>
            {{ t('vipManage.icon.save') }}
          </a-button>
          <a-button :disabled="busy != null" @click="emit('discard', record)">
            {{ t('vipManage.icon.discard') }}
          </a-button>
        </template>
        <a-popconfirm
          :content="t('vipManage.icon.restoreConfirm')"
          @ok="emit('restore', record)"
        >
          <a-button
            :loading="isBusy(record, 'restore')"
            :disabled="isLocked(record, 'restore')"
          >
            <template #icon><icon-undo /></template>
            {{ t('vipManage.icon.restore') }}
          </a-button>
        </a-popconfirm>
      </a-space>
    </template>
  </a-table>
</template>

<script lang="ts" setup>
  import { computed, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import type { RequestOption } from '@arco-design/web-vue/es/upload/interfaces';
  import {
    VIP_ICON_HEIGHT,
    VIP_ICON_MAX_SIZE,
    VIP_ICON_WIDTH,
  } from '@/api/member';
  import type {
    VipIconAction,
    VipIconBusy,
    VipIconDraft,
    VipIconGroup,
  } from '../types';

  const props = defineProps<{
    groups: VipIconGroup[];
    loading?: boolean;
    /** 已上传未保存的图标，按 group_no 索引 */
    drafts: Record<number, VipIconDraft>;
    /** 正在执行的操作；同一时间只允许一个，保证保存 / 恢复串行 */
    busy?: VipIconBusy | null;
  }>();

  const emit = defineEmits<{
    (e: 'upload', group: VipIconGroup, file: File): void;
    (e: 'save', group: VipIconGroup): void;
    (e: 'discard', group: VipIconGroup): void;
    (e: 'restore', group: VipIconGroup): void;
    (e: 'refresh'): void;
  }>();

  const PNG_SIGNATURE = [0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a];
  const { t } = useI18n();
  const validatingNo = ref<number | null>(null);
  const brokenUrls = ref(new Set<string>());

  const columns = computed<TableColumnData[]>(() => [
    {
      title: t('vipManage.icon.column.no'),
      dataIndex: 'no',
      width: 110,
      align: 'center',
    },
    {
      title: t('vipManage.icon.column.range'),
      slotName: 'range',
      width: 140,
    },
    {
      title: t('vipManage.icon.column.default'),
      slotName: 'defaultIcon',
      width: 160,
      align: 'center',
    },
    {
      title: t('vipManage.icon.column.current'),
      slotName: 'current',
      width: 240,
      align: 'center',
    },
    {
      title: t('vipManage.icon.column.updatedAt'),
      dataIndex: 'updatedAt',
      width: 180,
    },
    {
      title: t('vipManage.icon.column.operations'),
      slotName: 'operations',
      width: 340,
    },
  ]);

  function rangeText(group: VipIconGroup) {
    return `VIP${group.startLevel} - VIP${group.endLevel}`;
  }

  function displayIcon(group: VipIconGroup) {
    return props.drafts[group.no]?.url || group.icon;
  }

  function showImage(url: string) {
    return Boolean(url) && !brokenUrls.value.has(url);
  }

  function markBroken(url: string) {
    brokenUrls.value = new Set(brokenUrls.value).add(url);
  }

  function isBusy(group: VipIconGroup, action: VipIconAction) {
    return props.busy?.no === group.no && props.busy.action === action;
  }

  /** 有操作进行中时锁住其余按钮，正在执行的那个只显示 loading */
  function isLocked(group: VipIconGroup, action: VipIconAction) {
    return props.busy != null && !isBusy(group, action);
  }

  async function isPngFile(file: File) {
    const head = new Uint8Array(await file.slice(0, 8).arrayBuffer());
    return PNG_SIGNATURE.every((byte, index) => head[index] === byte);
  }

  function readImageSize(file: File) {
    return new Promise<{ width: number; height: number }>((resolve, reject) => {
      const url = URL.createObjectURL(file);
      const img = new Image();
      img.onload = () => {
        resolve({ width: img.naturalWidth, height: img.naturalHeight });
        URL.revokeObjectURL(url);
      };
      img.onerror = () => {
        reject(new Error('decode'));
        URL.revokeObjectURL(url);
      };
      img.src = url;
    });
  }

  async function validateIcon(file: File) {
    if (!/\.png$/i.test(file.name) || !(await isPngFile(file))) {
      return t('vipManage.icon.invalidType');
    }
    if (file.size > VIP_ICON_MAX_SIZE) return t('vipManage.icon.tooLarge');
    try {
      const { width, height } = await readImageSize(file);
      if (width !== VIP_ICON_WIDTH || height !== VIP_ICON_HEIGHT) {
        return t('vipManage.icon.invalidSize', { width, height });
      }
    } catch {
      return t('vipManage.icon.invalidType');
    }
    return '';
  }

  function onUpload(group: VipIconGroup, options: RequestOption) {
    const { fileItem, onError, onSuccess } = options;
    const { file } = fileItem;
    if (!file) {
      onError(new Error('file'));
      return {};
    }
    validatingNo.value = group.no;
    validateIcon(file)
      .then((reason) => {
        if (reason) {
          Message.warning(reason);
          onError(new Error(reason));
          return;
        }
        emit('upload', group, file);
        onSuccess({});
      })
      .finally(() => {
        validatingNo.value = null;
      });
    return {};
  }
</script>

<style scoped lang="less">
  .panel-header {
    display: flex;
    gap: 12px;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
  }

  .list-title {
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 16px;
  }

  .icon-box {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 72px;
    height: 96px;
    background: var(--color-fill-2);
    border: 1px dashed transparent;
    border-radius: 8px;

    &.is-draft {
      border-color: rgb(var(--orangered-5));
    }

    img {
      width: 58px;
      height: 83px;
      object-fit: contain;
    }
  }

  .icon-empty {
    color: var(--color-text-4);
    font-size: 12px;
  }

  .current-cell {
    display: flex;
    flex-direction: column;
    gap: 6px;
    align-items: center;
    padding: 8px 0;
  }

  .current-note {
    max-width: 200px;
    color: var(--color-text-3);
    font-size: 12px;
    line-height: 18px;
  }
</style>
