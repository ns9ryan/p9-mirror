<template>
  <div class="container">
    <a-card class="general-card" title="OSS 前端对接 Demo">
      <template #extra>
        <a-space>
          <a-tag color="arcoblue">/p9/oss</a-tag>
          <a-button size="small" @click="loadProvider">刷新生效配置</a-button>
        </a-space>
      </template>

      <a-alert type="info" style="margin-bottom: 16px">
        业务表只存 <b>key</b>；url 仅展示。只传
        <b>scene</b>，不要传桶名。大文件：presign → PUT put_url → complete。
      </a-alert>

      <a-form :model="form" layout="inline" style="margin-bottom: 16px">
        <a-form-item label="scene">
          <a-input
            v-model="form.scene"
            allow-clear
            placeholder="空则走 default_scene"
            style="width: 200px"
          />
        </a-form-item>
      </a-form>

      <a-tabs default-active-key="small">
        <!-- 小文件上传 -->
        <a-tab-pane key="small" title="小文件上传">
          <a-upload
            :custom-request="onSmallUpload"
            :show-file-list="false"
            tip="POST /file/upload · multipart field=file"
          >
            <template #upload-button>
              <a-button type="primary" :loading="loading.small">
                选择文件并上传
              </a-button>
            </template>
          </a-upload>
        </a-tab-pane>

        <!-- 大文件直传 -->
        <a-tab-pane key="large" title="大文件直传">
          <a-upload
            :custom-request="onLargeUpload"
            :show-file-list="false"
            tip="presign → PUT 存储 → complete（生产流程）"
          >
            <template #upload-button>
              <a-button type="primary" :loading="loading.large">
                选择文件并直传
              </a-button>
            </template>
          </a-upload>
          <a-progress
            v-if="largeProgress > 0"
            :percent="largeProgress"
            style="margin-top: 12px; max-width: 360px"
          />
          <a-steps
            v-if="largeSteps.length"
            :current="largeStepCurrent"
            size="small"
            style="margin-top: 16px; max-width: 640px"
          >
            <a-step
              v-for="item in largeSteps"
              :key="item.title"
              :title="item.title"
              :description="item.desc"
              :status="item.status"
            />
          </a-steps>
        </a-tab-pane>

        <!-- key 换链 -->
        <a-tab-pane key="url" title="key 换链">
          <a-space>
            <a-input
              v-model="form.key"
              allow-clear
              placeholder="已入库的 object key"
              style="width: 420px"
            />
            <a-button
              type="primary"
              :loading="loading.url"
              @click="onRefreshUrl"
            >
              GET /url/detail
            </a-button>
          </a-space>
        </a-tab-pane>

        <!-- 档案列表 -->
        <a-tab-pane key="list" title="档案列表">
          <a-space style="margin-bottom: 12px">
            <a-button type="primary" :loading="loading.list" @click="onList">
              POST /file/list
            </a-button>
            <a-tag>total: {{ fileTotal }}</a-tag>
          </a-space>
          <a-table
            :data="fileList"
            :pagination="false"
            :bordered="{ cell: true }"
            row-key="id"
            size="small"
          >
            <template #columns>
              <a-table-column title="id" data-index="id" :width="70" />
              <a-table-column title="name" data-index="name" :width="140" />
              <a-table-column title="scene" data-index="scene" :width="100" />
              <a-table-column title="key" data-index="key" :ellipsis="true" />
              <a-table-column title="size" data-index="size" :width="90" />
              <a-table-column
                title="visibility"
                data-index="visibility"
                :width="100"
              />
              <a-table-column title="操作" :width="160">
                <template #cell="{ record }">
                  <a-space>
                    <a-link @click="previewFromRecord(record)">预览</a-link>
                    <a-popconfirm
                      content="确认删除对象与档案？"
                      @ok="onDelete(record.key)"
                    >
                      <a-link status="danger">删除</a-link>
                    </a-popconfirm>
                  </a-space>
                </template>
              </a-table-column>
            </template>
          </a-table>
        </a-tab-pane>
      </a-tabs>

      <a-divider />

      <a-row :gutter="16">
        <a-col :span="12">
          <a-typography-title :heading="6">最近一次结果</a-typography-title>
          <a-descriptions
            v-if="lastResult"
            :column="1"
            size="small"
            bordered
            style="margin-bottom: 12px"
          >
            <a-descriptions-item label="key">
              {{ lastResult.key }}
              <a-button
                type="text"
                size="mini"
                @click="form.key = lastResult!.key"
              >
                填入换链
              </a-button>
            </a-descriptions-item>
            <a-descriptions-item label="scene">
              {{ lastResult.scene }}
            </a-descriptions-item>
            <a-descriptions-item label="visibility">
              {{ lastResult.visibility }}
            </a-descriptions-item>
            <a-descriptions-item
              v-if="lastResult.url_expires_in"
              label="url_expires_in"
            >
              {{ lastResult.url_expires_in }}s
            </a-descriptions-item>
            <a-descriptions-item label="url">
              <a :href="lastResult.url" target="_blank" rel="noopener">
                {{ lastResult.url }}
              </a>
            </a-descriptions-item>
          </a-descriptions>
          <a-empty v-else description="尚无结果" />
          <div v-if="previewSrc" class="preview">
            <div class="preview-label"
              >预览（优先本地文件，直链失败时仍可看）</div
            >
            <img :src="previewSrc" alt="preview" @error="onPreviewError" />
            <a-alert v-if="previewError" type="warning" style="margin-top: 8px">
              MinIO 直链加载失败：{{
                lastResult?.url
              }}。上传接口本身可能已成功，请点开 url 或看右侧 JSON。
            </a-alert>
          </div>
        </a-col>
        <a-col :span="12">
          <a-typography-title :heading="6">原始 JSON</a-typography-title>
          <pre class="json-box">{{ lastJson }}</pre>
          <a-typography-title :heading="6" style="margin-top: 16px">
            provider/active
          </a-typography-title>
          <pre class="json-box">{{ providerJson }}</pre>
        </a-col>
      </a-row>
    </a-card>
  </div>
</template>

<script lang="ts" setup>
  import { onMounted, reactive, ref } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import type { RequestOption } from '@arco-design/web-vue/es/upload/interfaces';
  import {
    completeOssUpload,
    deleteOssFile,
    getOssProviderActive,
    getOssUrlDetail,
    presignOssUpload,
    putOssObject,
    queryOssFileList,
    uploadOssFile,
    type OssFileRecord,
    type OssFileResult,
    type OssProvider,
  } from '@/api/oss';

  const form = reactive({
    scene: 'public',
    key: '',
  });

  const loading = reactive({
    small: false,
    large: false,
    url: false,
    list: false,
  });

  const largeProgress = ref(0);
  const largeStepCurrent = ref(0);
  const largeSteps = ref<
    {
      title: string;
      desc: string;
      status: 'wait' | 'process' | 'finish' | 'error';
    }[]
  >([]);
  const lastResult = ref<OssFileResult | null>(null);
  const lastJson = ref('{}');
  const providerJson = ref('尚未加载');
  const fileList = ref<OssFileRecord[]>([]);
  const fileTotal = ref(0);
  const previewSrc = ref('');
  const previewError = ref(false);
  let localObjectUrl = '';

  const isImageUrl = (url: string) =>
    /\.(png|jpe?g|gif|webp|bmp|svg)(\?|$)/i.test(url);

  const revokeLocalPreview = () => {
    if (localObjectUrl) {
      URL.revokeObjectURL(localObjectUrl);
      localObjectUrl = '';
    }
  };

  const setLocalPreview = (file: File) => {
    revokeLocalPreview();
    previewError.value = false;
    if (!file.type.startsWith('image/')) {
      previewSrc.value = '';
      return;
    }
    localObjectUrl = URL.createObjectURL(file);
    previewSrc.value = localObjectUrl;
  };

  const setRemotePreview = (url?: string) => {
    // 已有本地预览时不覆盖，避免 MinIO 直链偶发裂图
    if (localObjectUrl || !url || !isImageUrl(url)) return;
    previewError.value = false;
    previewSrc.value = url;
  };

  const onPreviewError = () => {
    previewError.value = true;
  };

  const setResult = (data: unknown) => {
    lastJson.value = JSON.stringify(data, null, 2);
    if (data && typeof data === 'object' && 'key' in data) {
      lastResult.value = data as OssFileResult;
      form.key = (data as OssFileResult).key || form.key;
      setRemotePreview((data as OssFileResult).url);
    }
  };

  const scenePayload = () => {
    const scene = form.scene.trim();
    return scene || undefined;
  };

  const resetLargeSteps = () => {
    largeStepCurrent.value = 0;
    largeSteps.value = [
      { title: '1. presign', desc: '签发 put_url', status: 'process' },
      { title: '2. PUT', desc: '直传 MinIO/OSS', status: 'wait' },
      { title: '3. complete', desc: '确认入库', status: 'wait' },
    ];
  };

  const markLargeStep = (
    index: number,
    status: 'wait' | 'process' | 'finish' | 'error',
    desc?: string
  ) => {
    const step = largeSteps.value[index];
    if (!step) return;
    step.status = status;
    if (desc) step.desc = desc;
    largeStepCurrent.value = index;
  };

  const onSmallUpload = (options: RequestOption) => {
    const controller = new AbortController();
    (async () => {
      const { onError, onSuccess, fileItem } = options;
      const file = fileItem.file as File;
      if (!file) {
        onError(new Error('no file'));
        return;
      }
      loading.small = true;
      setLocalPreview(file);
      try {
        const fd = new FormData();
        fd.append('file', file);
        const scene = scenePayload();
        if (scene) fd.append('scene', scene);
        const res = await uploadOssFile(fd);
        setResult(res.data);
        Message.success('小文件上传成功');
        onSuccess(res);
      } catch (e) {
        onError(e);
      } finally {
        loading.small = false;
      }
    })();
    return {
      abort() {
        controller.abort();
      },
    };
  };

  const onLargeUpload = (options: RequestOption) => {
    const controller = new AbortController();
    (async () => {
      const { onError, onSuccess, fileItem } = options;
      const file = fileItem.file as File;
      if (!file) {
        onError(new Error('no file'));
        return;
      }
      loading.large = true;
      largeProgress.value = 5;
      resetLargeSteps();
      setLocalPreview(file);
      try {
        const { data: presign } = await presignOssUpload({
          filename: file.name,
          size: file.size,
          scene: scenePayload(),
        });
        largeProgress.value = 20;
        markLargeStep(0, 'finish', `key=${presign.key}`);
        markLargeStep(1, 'process');
        setResult(presign);

        await putOssObject(presign.put_url, file, presign.headers || {});
        largeProgress.value = 70;
        markLargeStep(1, 'finish', '存储返回 200/204');
        markLargeStep(2, 'process');

        const { data: done } = await completeOssUpload({
          key: presign.key,
          scene: scenePayload(),
        });
        largeProgress.value = 100;
        markLargeStep(2, 'finish', '拿到最终 url/key');
        setResult(done);
        Message.success('大文件直传完成');
        onSuccess(done);
      } catch (e) {
        const cur = largeSteps.value.findIndex((s) => s.status === 'process');
        if (cur >= 0) {
          markLargeStep(cur, 'error', (e as Error)?.message || '失败');
        }
        onError(e);
        Message.error((e as Error)?.message || '大文件直传失败');
      } finally {
        loading.large = false;
      }
    })();
    return {
      abort() {
        controller.abort();
      },
    };
  };

  const onRefreshUrl = async () => {
    const key = form.key.trim();
    if (!key) {
      Message.warning('请填写 key');
      return;
    }
    loading.url = true;
    try {
      const { data } = await getOssUrlDetail({ key });
      revokeLocalPreview();
      setResult(data);
      Message.success('换链成功');
    } finally {
      loading.url = false;
    }
  };

  const onList = async () => {
    loading.list = true;
    try {
      const { data } = await queryOssFileList({
        scene: scenePayload(),
        page: 1,
        page_size: 20,
      });
      fileList.value = data?.list || [];
      fileTotal.value = data?.total || 0;
      setResult(data);
    } finally {
      loading.list = false;
    }
  };

  const previewFromRecord = (record: OssFileRecord) => {
    revokeLocalPreview();
    setResult(record);
    form.key = record.key;
  };

  const onDelete = async (key: string) => {
    const { data } = await deleteOssFile({ key });
    setResult(data);
    Message.success('已删除');
    await onList();
  };

  const loadProvider = async () => {
    try {
      const { data } = await getOssProviderActive();
      providerJson.value = JSON.stringify(data as OssProvider, null, 2);
    } catch (e) {
      providerJson.value = (e as Error)?.message || '加载失败';
    }
  };

  onMounted(() => {
    loadProvider();
  });
</script>

<style scoped lang="less">
  .container {
    padding: 0 20px 20px;
  }

  .json-box {
    margin: 0;
    padding: 12px;
    max-height: 320px;
    overflow: auto;
    font-size: 12px;
    line-height: 1.5;
    background: var(--color-fill-2);
    border-radius: 4px;
    white-space: pre-wrap;
    word-break: break-all;
  }

  .preview {
    margin-top: 12px;

    .preview-label {
      margin-bottom: 8px;
      font-size: 12px;
      color: var(--color-text-3);
    }

    img {
      max-width: 100%;
      max-height: 240px;
      border-radius: 4px;
      border: 1px solid var(--color-border-2);
    }
  }
</style>
