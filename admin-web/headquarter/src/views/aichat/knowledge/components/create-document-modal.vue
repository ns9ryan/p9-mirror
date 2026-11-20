<template>
  <a-modal
    :visible="visible"
    :title="$t('aichatKnowledgeModal.title.create')"
    :width="760"
    :ok-loading="submitting"
    :mask-closable="false"
    :ok-text="$t('aichatKnowledgeModal.okText')"
    unmount-on-close
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-form ref="formRef" :model="form" :rules="rules" layout="vertical">
      <a-row :gutter="16">
        <a-col :span="12">
          <a-form-item
            :label="$t('aichatKnowledgeModal.field.title')"
            field="title"
          >
            <a-input
              v-model="form.title"
              :placeholder="$t('aichatKnowledgeModal.field.titlePlaceholder')"
              :max-length="255"
              show-word-limit
              allow-clear
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item
            :label="$t('aichatKnowledgeModal.field.documentType')"
            field="document_type"
          >
            <a-select
              v-model="form.document_type"
              :placeholder="
                $t('aichatKnowledgeModal.field.documentTypePlaceholder')
              "
              :options="documentTypeOptions"
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item
            :label="$t('aichatKnowledgeModal.field.sourceTable')"
            field="source_table"
          >
            <a-input
              v-model="form.source_table"
              :placeholder="
                $t('aichatKnowledgeModal.field.sourceTablePlaceholder')
              "
              allow-clear
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item :label="$t('aichatKnowledgeModal.field.tags')" field="tags">
            <a-input-tag
              v-model="form.tags"
              :placeholder="$t('aichatKnowledgeModal.field.tagsPlaceholder')"
              allow-clear
              unique-value
            />
          </a-form-item>
        </a-col>
      </a-row>
      <a-form-item
        :label="$t('aichatKnowledgeModal.field.description')"
        field="description"
      >
        <a-input
          v-model="form.description"
          :placeholder="$t('aichatKnowledgeModal.field.descriptionPlaceholder')"
          allow-clear
        />
      </a-form-item>
      <a-form-item
        :label="$t('aichatKnowledgeModal.field.content')"
        field="content"
        :extra="$t('aichatKnowledgeModal.field.contentExtra')"
      >
        <a-textarea
          v-model="form.content"
          :placeholder="$t('aichatKnowledgeModal.field.contentPlaceholder')"
          :auto-size="{ minRows: 8, maxRows: 16 }"
        />
      </a-form-item>
      <a-form-item
        :label="$t('aichatKnowledgeModal.field.metadata')"
        field="metadata"
      >
        <a-textarea
          v-model="form.metadata"
          :placeholder="$t('aichatKnowledgeModal.field.metadataPlaceholder')"
          :auto-size="{ minRows: 2, maxRows: 6 }"
        />
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FormInstance } from '@arco-design/web-vue/es/form';
  import type { FieldRule } from '@arco-design/web-vue/es/form/interface';
  import {
    createKnowledgeDocument,
    type CreateKnowledgeDocumentParams,
    type KnowledgeDocumentType,
  } from '@/api/aichat';
  import { getDocumentTypeOptions } from '../constants';

  const props = defineProps<{ visible: boolean }>();
  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  const { t } = useI18n();

  const documentTypeOptions = computed(() => getDocumentTypeOptions());

  const createDefaultForm = () => ({
    title: '',
    content: '',
    document_type: undefined as KnowledgeDocumentType | undefined,
    source_table: '',
    description: '',
    tags: [] as string[],
    metadata: '',
  });

  const formRef = ref<FormInstance>();
  const form = reactive(createDefaultForm());
  const submitting = ref(false);

  const rules = computed<Record<string, FieldRule[]>>(() => ({
    title: [{ required: true, message: t('aichatKnowledgeModal.rule.title') }],
    document_type: [
      { required: true, message: t('aichatKnowledgeModal.rule.documentType') },
    ],
    content: [
      { required: true, message: t('aichatKnowledgeModal.rule.content') },
    ],
    metadata: [
      {
        validator: (value: string | undefined, callback) => {
          if (!value?.trim()) {
            callback();
            return;
          }
          try {
            JSON.parse(value);
            callback();
          } catch {
            callback(t('aichatKnowledgeModal.rule.metadataInvalid'));
          }
        },
      },
    ],
  }));

  watch(
    () => props.visible,
    (visible) => {
      if (visible) Object.assign(form, createDefaultForm());
    }
  );

  const buildParams = (): CreateKnowledgeDocumentParams => {
    const params: CreateKnowledgeDocumentParams = {
      title: form.title.trim(),
      content: form.content,
      document_type: form.document_type as KnowledgeDocumentType,
    };
    const sourceTable = form.source_table.trim();
    const description = form.description.trim();
    const metadata = form.metadata.trim();
    if (sourceTable) params.source_table = sourceTable;
    if (description) params.description = description;
    if (form.tags.length) params.tags = form.tags;
    if (metadata) params.metadata = metadata;
    return params;
  };

  const handleOk = async () => {
    const errors = await formRef.value?.validate();
    if (errors) return;
    submitting.value = true;
    try {
      const { data } = await createKnowledgeDocument(buildParams());
      Message.success(
        data?.chunk_count
          ? t('aichatKnowledgeModal.message.createSuccessWithChunks', {
              count: data.chunk_count,
            })
          : t('front.common.message.createSuccess')
      );
      emit('update:visible', false);
      emit('success');
    } catch {
      // 错误提示由全局拦截器处理
    } finally {
      submitting.value = false;
    }
  };

  const handleCancel = () => {
    emit('update:visible', false);
  };
</script>
