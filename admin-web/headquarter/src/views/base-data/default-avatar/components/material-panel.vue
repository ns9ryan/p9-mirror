<template>
  <section>
    <section-header
      :title="t(`defaultAvatar.section.${type}`)"
      :desc="t(`defaultAvatar.desc.${type}`)"
      :specs="specs"
    >
      <template #title-extra>
        <a-tag
          v-if="list.length"
          size="small"
          :color="configuredCount === list.length ? 'green' : 'orange'"
        >
          {{
            t('defaultAvatar.configured', {
              done: configuredCount,
              total: list.length,
            })
          }}
        </a-tag>
      </template>
    </section-header>

    <a-spin :loading="loading" class="panel-spin">
      <div v-if="list.length" class="slot-grid">
        <div
          v-for="(item, index) in list"
          :key="item.id"
          class="slot-card"
          :class="{
            'is-uploading': uploadingId === item.id,
            'is-locked': uploadingId != null && uploadingId !== item.id,
          }"
        >
          <a-upload
            :show-file-list="false"
            accept=".png,image/png"
            :disabled="uploadingId != null"
            :custom-request="(option) => onUpload(item, index, option)"
          >
            <template #upload-button>
              <div class="preview" :class="{ 'is-frame': isFrame }">
                <img
                  v-if="isFrame && sampleAvatar"
                  class="sample-avatar"
                  :src="sampleAvatar"
                  alt=""
                />
                <img
                  v-if="item.image_url"
                  class="material"
                  :src="item.image_url"
                  alt=""
                />
                <div v-else class="empty-placeholder">
                  <icon-plus />
                  <span>{{ t('defaultAvatar.clickUpload') }}</span>
                </div>
                <div v-if="item.image_url" class="preview-mask">
                  <icon-loading v-if="uploadingId === item.id" />
                  <icon-upload v-else />
                  <span>
                    {{
                      uploadingId === item.id
                        ? t('defaultAvatar.uploading')
                        : t('defaultAvatar.replace')
                    }}
                  </span>
                </div>
              </div>
            </template>
          </a-upload>
          <div class="slot-meta">
            <span class="slot-name">{{ slotName(index) }}</span>
            <span class="slot-id">#{{ item.id }}</span>
          </div>
        </div>
      </div>
      <a-empty v-else-if="!loading" :description="t('defaultAvatar.noData')" />
    </a-spin>

    <a-modal
      v-model:visible="confirmVisible"
      :title="
        t('defaultAvatar.confirm.title', {
          name: pending ? slotName(pending.index) : '',
        })
      "
      :ok-text="t('defaultAvatar.confirm.okText')"
      :on-before-ok="confirmUpdate"
      :mask-closable="false"
      @close="pending = null"
    >
      <template v-if="pending">
        <div class="compare">
          <div class="compare-item">
            <div class="compare-label">{{
              t('defaultAvatar.confirm.current')
            }}</div>
            <div class="preview is-large" :class="{ 'is-frame': isFrame }">
              <img
                v-if="isFrame && sampleAvatar"
                class="sample-avatar"
                :src="sampleAvatar"
                alt=""
              />
              <img
                v-if="pending.record.image_url"
                class="material"
                :src="pending.record.image_url"
                alt=""
              />
              <span v-else class="empty-text">
                {{ t('defaultAvatar.empty') }}
              </span>
            </div>
          </div>
          <icon-arrow-right class="compare-arrow" />
          <div class="compare-item">
            <div class="compare-label is-next">{{
              t('defaultAvatar.confirm.next')
            }}</div>
            <div
              class="preview is-large is-next"
              :class="{ 'is-frame': isFrame }"
            >
              <img
                v-if="isFrame && sampleAvatar"
                class="sample-avatar"
                :src="sampleAvatar"
                alt=""
              />
              <img class="material" :src="pending.url" alt="" />
            </div>
          </div>
        </div>
        <div class="confirm-tip">{{ t('defaultAvatar.confirm.tip') }}</div>
      </template>
    </a-modal>
  </section>
</template>

<script lang="ts" setup>
  import { computed, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { RequestOption } from '@arco-design/web-vue/es/upload/interfaces';
  import {
    queryAvatarMaterialList,
    updateAvatarMaterial,
    uploadAvatarMaterial,
    type AvatarMaterialRecord,
    type AvatarMaterialType,
  } from '@/api/defaultAvatar';
  import { checkPngFile, type PngRule } from '@/utils/image-check';
  import useLoading from '@/hooks/loading';
  import SectionHeader from './section-header.vue';

  interface PendingReplace {
    record: AvatarMaterialRecord;
    index: number;
    url: string;
    key: string;
  }

  const props = defineProps<{ type: AvatarMaterialType }>();
  const emit = defineEmits<{ (e: 'updated'): void }>();

  const RULES: Record<AvatarMaterialType, PngRule> = {
    avatar: { width: 200, height: 200, maxSize: 204800 },
    frame: {
      width: 145,
      height: 145,
      maxSize: 204800,
      transparentCenter: true,
    },
  };

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(false);
  const list = ref<AvatarMaterialRecord[]>([]);
  /** 头像框预览时垫在下面的示例头像 */
  const sampleAvatar = ref('');
  const uploadingId = ref<number | null>(null);
  const pending = ref<PendingReplace | null>(null);
  const confirmVisible = ref(false);

  const isFrame = computed(() => props.type === 'frame');
  const configuredCount = computed(
    () => list.value.filter((item) => item.image_url).length
  );
  const specs = computed(() => {
    const rule = RULES[props.type];
    const result = [
      t(
        isFrame.value
          ? 'defaultAvatar.spec.transparentPng'
          : 'defaultAvatar.spec.png'
      ),
      t('defaultAvatar.spec.dimension', {
        width: rule.width,
        height: rule.height,
      }),
      t('defaultAvatar.spec.maxSize', { size: rule.maxSize / 1024 }),
    ];
    if (rule.transparentCenter) {
      result.push(t('defaultAvatar.spec.transparentCenter'));
    }
    return result;
  });

  function slotName(index: number) {
    return t(`defaultAvatar.slot.${props.type}`, { no: index + 1 });
  }

  async function fetchData() {
    setLoading(true);
    try {
      const { data } = await queryAvatarMaterialList(props.type);
      list.value = data?.list || [];
    } catch {
      list.value = [];
    } finally {
      setLoading(false);
    }
  }

  async function fetchSampleAvatar() {
    try {
      const { data } = await queryAvatarMaterialList('avatar', {
        silentError: true,
      });
      sampleAvatar.value =
        (data?.list || []).find((item) => item.image_url)?.image_url || '';
    } catch {
      sampleAvatar.value = '';
    }
  }

  async function handleFile(
    record: AvatarMaterialRecord,
    index: number,
    file: File
  ) {
    const rule = RULES[props.type];
    uploadingId.value = record.id;
    try {
      const error = await checkPngFile(file, rule);
      if (error) {
        Message.warning(
          t(`imageCheck.${error}`, { width: rule.width, height: rule.height })
        );
        return;
      }
      const { data } = await uploadAvatarMaterial(props.type, file);
      pending.value = { record, index, url: data.url, key: data.key };
      confirmVisible.value = true;
    } catch {
      // 错误提示已由全局拦截器处理
    } finally {
      uploadingId.value = null;
    }
  }

  function onUpload(
    record: AvatarMaterialRecord,
    index: number,
    options: RequestOption
  ) {
    const { fileItem, onSuccess } = options;
    if (fileItem.file) handleFile(record, index, fileItem.file);
    onSuccess({});
    return {};
  }

  async function confirmUpdate() {
    if (!pending.value) return true;
    const { record, key, url } = pending.value;
    try {
      await updateAvatarMaterial(props.type, {
        id: record.id,
        image_url: key || url,
      });
      Message.success(t('defaultAvatar.message.updated'));
      fetchData();
      if (isFrame.value) fetchSampleAvatar();
      emit('updated');
      return true;
    } catch {
      return false;
    }
  }

  defineExpose({ refreshSample: fetchSampleAvatar });

  fetchData();
  if (isFrame.value) fetchSampleAvatar();
</script>

<style scoped lang="less">
  .panel-spin {
    display: block;
    min-height: 160px;
  }

  .slot-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    gap: 16px;
  }

  .slot-card {
    display: flex;
    flex-direction: column;
    gap: 12px;
    align-items: center;
    padding: 16px 12px 12px;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 10px;
    transition: border-color 0.2s, box-shadow 0.2s, transform 0.2s;

    &:hover {
      border-color: rgb(var(--primary-3));
      box-shadow: 0 6px 16px rgb(0 0 0 / 8%);
      transform: translateY(-2px);
    }

    &:hover .preview-mask,
    &.is-uploading .preview-mask {
      opacity: 1;
    }

    &.is-locked {
      opacity: 0.6;

      .preview {
        cursor: not-allowed;
      }
    }
  }

  .preview {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 112px;
    height: 112px;
    overflow: hidden;
    background: var(--color-fill-2);
    border-radius: 12px;
    cursor: pointer;

    &.is-large {
      width: 160px;
      height: 160px;
      cursor: default;
    }

    &.is-next {
      box-shadow: 0 0 0 2px rgb(var(--primary-5));
    }

    &.is-frame {
      background-color: #fff;
      background-image: linear-gradient(45deg, #e5e6eb 25%, transparent 25%),
        linear-gradient(-45deg, #e5e6eb 25%, transparent 25%),
        linear-gradient(45deg, transparent 75%, #e5e6eb 75%),
        linear-gradient(-45deg, transparent 75%, #e5e6eb 75%);
      background-position: 0 0, 0 6px, 6px -6px, -6px 0;
      background-size: 12px 12px;
    }

    .material {
      position: relative;
      width: 100%;
      height: 100%;
      object-fit: contain;
    }

    .sample-avatar {
      position: absolute;
      top: 50%;
      left: 50%;
      width: 68%;
      height: 68%;
      object-fit: cover;
      border-radius: 50%;
      transform: translate(-50%, -50%);
    }

    .empty-text {
      position: relative;
      color: var(--color-text-3);
      font-size: 12px;
    }
  }

  .empty-placeholder {
    position: relative;
    display: flex;
    flex-direction: column;
    gap: 6px;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
    color: var(--color-text-3);
    font-size: 12px;
    border: 1px dashed var(--color-border-3);
    border-radius: 12px;
    transition: color 0.2s, border-color 0.2s;

    .arco-icon {
      font-size: 22px;
    }

    &:hover {
      color: rgb(var(--primary-6));
      border-color: rgb(var(--primary-6));
    }
  }

  .preview-mask {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
    align-items: center;
    justify-content: center;
    color: #fff;
    font-size: 13px;
    background: rgb(0 0 0 / 45%);
    opacity: 0;
    transition: opacity 0.2s;

    .arco-icon {
      font-size: 20px;
    }
  }

  .slot-meta {
    display: flex;
    gap: 6px;
    align-items: baseline;
    justify-content: center;
    width: 100%;
  }

  .slot-name {
    overflow: hidden;
    color: var(--color-text-1);
    font-weight: 500;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .slot-id {
    flex-shrink: 0;
    color: var(--color-text-4);
    font-size: 12px;
  }

  .compare {
    display: flex;
    gap: 24px;
    align-items: center;
    justify-content: center;
  }

  .compare-item {
    display: flex;
    flex-direction: column;
    gap: 8px;
    align-items: center;
  }

  .compare-label {
    color: var(--color-text-2);

    &.is-next {
      color: rgb(var(--primary-6));
      font-weight: 500;
    }
  }

  .compare-arrow {
    color: var(--color-text-3);
    font-size: 20px;
  }

  .confirm-tip {
    margin-top: 16px;
    color: var(--color-text-3);
    font-size: 12px;
    text-align: center;
  }
</style>
