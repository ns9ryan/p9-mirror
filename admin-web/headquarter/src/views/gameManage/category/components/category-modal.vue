<template>
  <a-modal
    :visible="visible"
    :title="t('categoryModal.title.edit')"
    :ok-loading="submitting"
    :ok-text="t('front.common.okText.save')"
    :mask-closable="false"
    :esc-to-close="false"
    unmount-on-close
    :width="680"
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    @before-ok="handleBeforeOk"
    @cancel="handleCancel"
  >
    <a-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-align="right"
      :label-col-props="{ span: 6 }"
      :wrapper-col-props="{ span: 17 }"
    >
      <a-form-item :label="t('categoryModal.field.categoryCode')" field="categoryCode">
        <a-input v-model="form.categoryCode" disabled />
      </a-form-item>
      <a-form-item :label="t('categoryModal.field.name')" field="name" required>
        <a-textarea
          v-model="form.name"
          :placeholder="t('categoryModal.field.namePlaceholder')"
          allow-clear
          :auto-size="{ minRows: 4, maxRows: 8 }"
        />
      </a-form-item>
      <a-form-item :label="t('categoryModal.field.sortNo')" field="sort_no">
        <a-input-number
          v-model="form.sort_no"
          :min="1"
          :max="9999"
          :placeholder="t('categoryModal.field.sortNoPlaceholder')"
          style="width: 100%"
        />
      </a-form-item>
      <a-form-item :label="t('categoryModal.field.status')" field="status">
        <a-radio-group v-model="form.status">
          <a-radio :value="1">{{ t('front.common.status.enabled') }}</a-radio>
          <a-radio :value="2">{{ t('front.common.status.disabled') }}<span style="color: red">{{ t('categoryModal.field.forceLogoutHint') }}</span></a-radio>
        </a-radio-group>
      </a-form-item>
      <a-form-item :label="t('categoryModal.field.note')">
        <a-alert type="info" :show-icon="false">
          {{ t('categoryModal.field.noteContent') }}
        </a-alert>
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FieldRule, FormInstance } from '@arco-design/web-vue';
  import {
    updateGameCategory,
    type GameCategoryRecord,
    type UpdateGameCategoryParams,
  } from '@/api/gameCategory';

  const { t } = useI18n();

  const props = defineProps<{
    visible: boolean;
    category?: GameCategoryRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  interface CategoryForm {
    categoryCode: string;
    name: string;
    sort_no: number;
    status: number;
  }

  const formRef = ref<FormInstance>();
  const submitting = ref(false);
  const isEdit = computed(() => Boolean(props.category?.id));

  const getDefaultForm = (): CategoryForm => ({
    categoryCode: '',
    name: '',
    sort_no: 1,
    status: 1,
  });

  const form = reactive<CategoryForm>(getDefaultForm());

  const rules = computed<Record<string, FieldRule | FieldRule[]>>(() => ({
    name: [
      { required: true, message: t('categoryModal.rule.nameRequired') },
      {
        validator: (value: string, callback: (error?: string) => void) => {
          try {
            const parsed = JSON.parse(value || '');
            if (!parsed || typeof parsed !== 'object') {
              callback(t('categoryModal.rule.nameObject'));
              return;
            }
            callback();
          } catch {
            callback(t('categoryModal.rule.nameJson'));
          }
        },
      },
    ],
  }));

  const prettyJson = (value: string) => {
    try {
      return JSON.stringify(JSON.parse(value), null, 2);
    } catch {
      return value;
    }
  };

  const resetForm = () => {
    Object.assign(form, getDefaultForm());
    formRef.value?.clearValidate();
  };

  const fillForm = (category: GameCategoryRecord) => {
    Object.assign(form, getDefaultForm(), {
      categoryCode: category.category_code || '--',
      name: prettyJson(category.name || ''),
      sort_no: category.sort_no && category.sort_no > 0 ? category.sort_no : 1,
      status: category.status ?? 1,
    });
  };

  const close = () => {
    emit('update:visible', false);
  };

  const handleCancel = () => {
    close();
    resetForm();
  };

  const buildPayload = (): UpdateGameCategoryParams => ({
    id: props.category!.id,
    name: form.name.trim(),
    sort_no: form.sort_no && form.sort_no > 0 ? form.sort_no : 1,
    status: form.status ?? 1,
    force_logout: form.status === 2,
  });

  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    const errors = await formRef.value?.validate();
    if (errors || submitting.value || !isEdit.value || !props.category?.id) {
      done(false);
      return;
    }

    submitting.value = true;
    try {
      await updateGameCategory(buildPayload());
      Message.success(t('front.common.message.saveSuccess'));
      resetForm();
      emit('success');
      done(true);
      close();
    } catch {
      done(false);
    } finally {
      submitting.value = false;
    }
  };

  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return;
      resetForm();
      if (props.category?.id) {
        fillForm(props.category);
      }
    }
  );
</script>