<template>
  <a-modal
    :visible="visible"
    :title="t('categoryModal.title.edit')"
    :ok-loading="submitting"
    :ok-button-props="{ disabled: detailLoading || !detailCategory }"
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
    <a-spin :loading="detailLoading" style="width: 100%">
      <a-form
        ref="formRef"
        :model="form"
        label-align="right"
        :label-col-props="{ span: 6 }"
        :wrapper-col-props="{ span: 17 }"
      >
        <a-form-item
          :label="t('categoryModal.field.categoryCode')"
          field="categoryCode"
        >
          <a-input v-model="form.categoryCode" disabled />
        </a-form-item>
        <a-form-item :label="t('categoryModal.field.name')" field="name">
          <a-input v-model="form.name" disabled />
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
            <a-radio :value="1">
              {{ t('categoryManage.status.enabled') }}
            </a-radio>
            <a-radio :value="2">
              {{ t('categoryManage.status.disabled') }}
            </a-radio>
          </a-radio-group>
        </a-form-item>
      </a-form>
    </a-spin>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FormInstance } from '@arco-design/web-vue';
  import {
    getGameCategory,
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
  const detailLoading = ref(false);
  const detailCategory = ref<GameCategoryRecord | null>(null);
  const isEdit = computed(() => Boolean(props.category?.id));

  const getDefaultForm = (): CategoryForm => ({
    categoryCode: '',
    name: '',
    sort_no: 1,
    status: 1,
  });

  const form = reactive<CategoryForm>(getDefaultForm());

  const resetForm = () => {
    Object.assign(form, getDefaultForm());
    formRef.value?.clearValidate();
  };

  const fillForm = (category: GameCategoryRecord) => {
    Object.assign(form, getDefaultForm(), {
      categoryCode: category.category_code || '--',
      name: category.name || '--',
      sort_no: category.sort_no && category.sort_no > 0 ? category.sort_no : 1,
      status: category.status ?? 1,
    });
  };

  const fetchDetail = async (id: number) => {
    detailLoading.value = true;
    detailCategory.value = null;
    resetForm();
    try {
      const { data } = await getGameCategory(id);
      detailCategory.value = data ?? null;
      if (data) {
        fillForm(data);
      }
    } catch {
      detailCategory.value = null;
    } finally {
      detailLoading.value = false;
    }
  };

  const close = () => {
    emit('update:visible', false);
  };

  const handleCancel = () => {
    close();
    detailCategory.value = null;
    resetForm();
  };

  const buildPayload = (): UpdateGameCategoryParams | null => {
    const id = detailCategory.value?.id;
    if (!id) return null;
    return {
      id,
      sort_no: form.sort_no && form.sort_no > 0 ? form.sort_no : 1,
      status: form.status ?? 1,
    };
  };

  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    const payload = buildPayload();
    if (!payload || submitting.value || detailLoading.value || !isEdit.value) {
      done(false);
      return;
    }

    submitting.value = true;
    try {
      await updateGameCategory(payload);
      Message.success(t('front.common.message.saveSuccess'));
      detailCategory.value = null;
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
      if (props.category?.id) {
        fetchDetail(props.category.id);
      } else {
        detailCategory.value = null;
        resetForm();
      }
    }
  );
</script>
