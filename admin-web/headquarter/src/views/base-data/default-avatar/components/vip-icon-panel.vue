<template>
  <section>
    <section-header
      :title="t('vipDefaultIcon.title')"
      :desc="t('vipDefaultIcon.desc')"
      :specs="specs"
    >
      <template #title-extra>
        <a-tag v-if="dirtyCount" size="small" color="orangered">
          {{ t('vipDefaultIcon.dirtyTag', { count: dirtyCount }) }}
        </a-tag>
      </template>
      <template #extra>
        <a-space wrap>
          <a-button :disabled="saving" @click="fetchData">
            <template #icon><icon-refresh /></template>
            {{ t('front.common.toolbar.refresh') }}
          </a-button>
          <a-popconfirm
            :content="t('vipDefaultIcon.discardConfirm')"
            @ok="pending = {}"
          >
            <a-button :disabled="!dirtyCount || saving">
              <template #icon><icon-undo /></template>
              {{ t('vipDefaultIcon.discard') }}
            </a-button>
          </a-popconfirm>
          <a-button
            type="primary"
            :loading="saving"
            :disabled="!dirtyCount || uploadingLevel != null"
            @click="handleSave"
          >
            <template #icon><icon-save /></template>
            {{ t('vipDefaultIcon.save', { count: dirtyCount }) }}
          </a-button>
        </a-space>
      </template>
    </section-header>

    <a-spin :loading="loading" class="panel-spin">
      <div v-if="groups.length" class="vip-grid">
        <div
          v-for="group in groups"
          :key="group.min_level"
          class="vip-card"
          :class="{ 'is-dirty': pending[group.min_level] }"
        >
          <div class="vip-card-header">
            <span class="vip-range">
              <icon-trophy />
              {{ rangeText(group) }}
            </span>
            <a-tag
              v-if="pending[group.min_level]"
              size="small"
              color="orangered"
            >
              {{
                pending[group.min_level].icon_url
                  ? t('vipDefaultIcon.pendingTag')
                  : t('vipDefaultIcon.pendingClear')
              }}
            </a-tag>
          </div>

          <div class="vip-card-body">
            <template v-if="pending[group.min_level]">
              <div class="icon-col">
                <div class="icon-box is-small">
                  <img v-if="group.icon_url" :src="group.icon_url" alt="" />
                  <span v-else class="empty-text">{{
                    t('vipDefaultIcon.empty')
                  }}</span>
                </div>
                <span class="icon-label">{{
                  t('vipDefaultIcon.current')
                }}</span>
              </div>
              <icon-arrow-right class="icon-arrow" />
              <div class="icon-col">
                <div class="icon-box is-small is-next">
                  <img
                    v-if="pending[group.min_level].preview"
                    :src="pending[group.min_level].preview"
                    alt=""
                  />
                  <span v-else class="empty-text">{{
                    t('vipDefaultIcon.pendingClear')
                  }}</span>
                </div>
                <span class="icon-label is-next">
                  {{ t('vipDefaultIcon.next') }}
                </span>
              </div>
            </template>
            <div v-else class="icon-box">
              <img v-if="group.icon_url" :src="group.icon_url" alt="" />
              <span v-else class="empty-text">
                <icon-image />
                {{ t('vipDefaultIcon.empty') }}
              </span>
            </div>
          </div>

          <div class="vip-card-footer">
            <a-upload
              :show-file-list="false"
              accept=".png,image/png"
              :disabled="busy"
              :custom-request="(option) => onUpload(group, option)"
            >
              <template #upload-button>
                <a-button
                  size="small"
                  type="primary"
                  :loading="uploadingLevel === group.min_level"
                  :disabled="busy && uploadingLevel !== group.min_level"
                >
                  <template #icon><icon-upload /></template>
                  {{ t('vipDefaultIcon.upload') }}
                </a-button>
              </template>
            </a-upload>
            <a-button
              v-if="pending[group.min_level]"
              size="small"
              :disabled="busy"
              @click="revert(group)"
            >
              {{ t('vipDefaultIcon.revert') }}
            </a-button>
            <a-popconfirm
              v-else
              :content="
                t('vipDefaultIcon.clearConfirm', { range: rangeText(group) })
              "
              @ok="markClear(group)"
            >
              <a-button
                size="small"
                status="danger"
                :disabled="busy || !group.icon_url"
              >
                {{ t('vipDefaultIcon.clear') }}
              </a-button>
            </a-popconfirm>
          </div>
        </div>
      </div>
      <a-empty v-else-if="!loading" />
    </a-spin>
  </section>
</template>

<script lang="ts" setup>
  import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { onBeforeRouteLeave } from 'vue-router';
  import { Message, Modal } from '@arco-design/web-vue';
  import type { RequestOption } from '@arco-design/web-vue/es/upload/interfaces';
  import {
    queryVipDefaultIconList,
    saveVipDefaultIcons,
    uploadVipDefaultIcon,
    type VipDefaultIconGroup,
  } from '@/api/vipIcon';
  import { checkPngFile, type PngRule } from '@/utils/image-check';
  import useLoading from '@/hooks/loading';
  import SectionHeader from './section-header.vue';

  interface PendingIcon {
    /** 提交给 save 的值：上传返回的 key，"" 表示清空 */
    icon_url: string;
    /** 预览用的完整地址，清空时为 "" */
    preview: string;
  }

  const ICON_RULE: PngRule = { width: 115, height: 165, maxSize: 204800 };

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(false);
  const groups = ref<VipDefaultIconGroup[]>([]);
  /** 按 min_level 暂存的未保存修改 */
  const pending = ref<Record<number, PendingIcon>>({});
  const uploadingLevel = ref<number | null>(null);
  const saving = ref(false);

  const dirtyCount = computed(() => Object.keys(pending.value).length);
  const busy = computed(() => saving.value || uploadingLevel.value != null);
  const specs = computed(() => [
    t('defaultAvatar.spec.png'),
    t('defaultAvatar.spec.dimension', {
      width: ICON_RULE.width,
      height: ICON_RULE.height,
    }),
    t('defaultAvatar.spec.maxSize', { size: ICON_RULE.maxSize / 1024 }),
  ]);

  function rangeText(group: VipDefaultIconGroup) {
    return `VIP${group.min_level} - VIP${group.max_level}`;
  }

  async function fetchData() {
    setLoading(true);
    try {
      const { data } = await queryVipDefaultIconList();
      groups.value = data?.list || [];
    } catch {
      groups.value = [];
    } finally {
      setLoading(false);
    }
  }

  function setPending(minLevel: number, value: PendingIcon | null) {
    const next = { ...pending.value };
    if (value) next[minLevel] = value;
    else delete next[minLevel];
    pending.value = next;
  }

  async function handleFile(group: VipDefaultIconGroup, file: File) {
    uploadingLevel.value = group.min_level;
    try {
      const error = await checkPngFile(file, ICON_RULE);
      if (error) {
        Message.warning(
          t(`imageCheck.${error}`, {
            width: ICON_RULE.width,
            height: ICON_RULE.height,
          })
        );
        return;
      }
      const { data } = await uploadVipDefaultIcon(group.min_level, file);
      setPending(group.min_level, {
        icon_url: data.key || data.url,
        preview: data.url,
      });
    } catch {
      // 错误提示已由全局拦截器处理
    } finally {
      uploadingLevel.value = null;
    }
  }

  function onUpload(group: VipDefaultIconGroup, options: RequestOption) {
    const { fileItem, onSuccess } = options;
    if (fileItem.file) handleFile(group, fileItem.file);
    onSuccess({});
    return {};
  }

  function markClear(group: VipDefaultIconGroup) {
    setPending(group.min_level, { icon_url: '', preview: '' });
  }

  function revert(group: VipDefaultIconGroup) {
    setPending(group.min_level, null);
  }

  async function handleSave() {
    const list = groups.value
      .filter((group) => pending.value[group.min_level])
      .map((group) => ({
        min_level: group.min_level,
        max_level: group.max_level,
        icon_url: pending.value[group.min_level].icon_url,
      }));
    if (!list.length) return;
    saving.value = true;
    try {
      await saveVipDefaultIcons(list);
      pending.value = {};
      Message.success(t('vipDefaultIcon.message.saved'));
      fetchData();
    } catch {
      // 保存失败保留待保存内容
    } finally {
      saving.value = false;
    }
  }

  onBeforeRouteLeave(() => {
    if (!dirtyCount.value) return true;
    return new Promise<boolean>((resolve) => {
      Modal.confirm({
        title: t('vipDefaultIcon.leave.title'),
        content: t('vipDefaultIcon.leave.content'),
        okText: t('vipDefaultIcon.leave.okText'),
        onOk: () => resolve(true),
        onCancel: () => resolve(false),
      });
    });
  });

  function onBeforeUnload(event: BeforeUnloadEvent) {
    if (!dirtyCount.value) return;
    event.preventDefault();
    event.returnValue = '';
  }

  onMounted(() => window.addEventListener('beforeunload', onBeforeUnload));
  onBeforeUnmount(() =>
    window.removeEventListener('beforeunload', onBeforeUnload)
  );

  fetchData();
</script>

<style scoped lang="less">
  .panel-spin {
    display: block;
    min-height: 200px;
  }

  .vip-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: 16px;
  }

  .vip-card {
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 10px;
    transition: border-color 0.2s, box-shadow 0.2s, transform 0.2s;

    &:hover {
      border-color: rgb(var(--gold-3));
      box-shadow: 0 6px 16px rgb(0 0 0 / 8%);
      transform: translateY(-2px);
    }

    &.is-dirty {
      border-color: rgb(var(--orangered-4));
      box-shadow: 0 0 0 2px rgb(var(--orangered-1));
    }
  }

  .vip-card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    min-height: 44px;
    padding: 0 12px;
    background: linear-gradient(
      135deg,
      rgb(var(--gold-1)) 0%,
      var(--color-bg-2) 100%
    );
    border-bottom: 1px solid var(--color-border-1);
  }

  .vip-range {
    display: inline-flex;
    gap: 6px;
    align-items: center;
    color: rgb(var(--gold-7));
    font-weight: 600;
    font-size: 14px;
  }

  .vip-card-body {
    display: flex;
    gap: 10px;
    align-items: center;
    justify-content: center;
    height: 156px;
    padding: 12px;
  }

  .icon-col {
    display: flex;
    flex-direction: column;
    gap: 6px;
    align-items: center;
  }

  .icon-label {
    color: var(--color-text-3);
    font-size: 12px;

    &.is-next {
      color: rgb(var(--orangered-6));
      font-weight: 500;
    }
  }

  .icon-arrow {
    color: var(--color-text-3);
    font-size: 16px;
  }

  .icon-box {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 92px;
    height: 132px;
    overflow: hidden;
    background: var(--color-fill-2);
    border-radius: 8px;

    &.is-small {
      width: 72px;
      height: 103px;
    }

    &.is-next {
      box-shadow: 0 0 0 2px rgb(var(--orangered-4));
    }

    img {
      width: 100%;
      height: 100%;
      object-fit: contain;
    }
  }

  .empty-text {
    display: flex;
    flex-direction: column;
    gap: 6px;
    align-items: center;
    color: var(--color-text-3);
    font-size: 12px;

    .arco-icon {
      font-size: 22px;
    }
  }

  .vip-card-footer {
    display: flex;
    gap: 8px;
    align-items: center;
    justify-content: center;
    padding: 10px 12px 14px;
  }
</style>
