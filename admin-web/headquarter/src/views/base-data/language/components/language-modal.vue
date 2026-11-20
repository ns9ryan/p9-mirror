<template>
  <a-modal
    :visible="visible"
    :title="
      isEdit ? $t('languageModal.title.edit') : $t('languageModal.title.create')
    "
    :ok-loading="submitting"
    :ok-text="
      isEdit
        ? $t('front.common.okText.save')
        : $t('front.common.okText.create')
    "
    :mask-closable="false"
    :esc-to-close="false"
    unmount-on-close
    :width="600"
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
      <a-form-item
        :label="$t('languageModal.field.code')"
        field="code"
        required
      >
        <a-input
          v-model="form.code"
          :placeholder="$t('languageModal.field.codePlaceholder')"
          allow-clear
          :max-length="35"
          :disabled="isEdit"
        />
        <template v-if="isEdit" #extra>{{
          $t('languageModal.field.codeExtra')
        }}</template>
      </a-form-item>
      <a-form-item :label="$t('front.common.column.status')" field="status">
        <a-radio-group v-model="form.status">
          <a-radio :value="1">{{ $t('front.common.status.enabled') }}</a-radio>
          <a-radio :value="2">{{ $t('front.common.status.disabled') }}</a-radio>
        </a-radio-group>
      </a-form-item>
      <a-form-item :label="$t('languageModal.field.nameI18n')" required>
        <div class="name-i18n-list">
          <div
            v-for="(row, index) in nameRows"
            :key="row.key"
            class="name-i18n-row"
          >
            <a-input
              v-model="row.lang"
              :placeholder="$t('languageModal.field.langPlaceholder')"
              class="name-i18n-lang"
              allow-clear
            />
            <a-input
              v-model="row.name"
              :placeholder="$t('languageModal.field.namePlaceholder')"
              class="name-i18n-name"
              allow-clear
            />
            <a-button
              shape="circle"
              size="mini"
              status="danger"
              :disabled="nameRows.length <= 1"
              @click="removeNameRow(index)"
            >
              <template #icon><icon-minus /></template>
            </a-button>
          </div>
          <a-button long dashed @click="addNameRow">
            <template #icon><icon-plus /></template>
            {{ $t('languageModal.action.addName') }}
          </a-button>
        </div>
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FormInstance, FieldRule } from '@arco-design/web-vue';
  import {
    createLanguage,
    updateLanguage,
    type LanguageRecord,
  } from '@/api/language';

  const props = defineProps<{
    visible: boolean;
    record?: LanguageRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  interface NameRow {
    key: number;
    lang: string;
    name: string;
  }

  const { t } = useI18n();
  const formRef = ref<FormInstance>();
  const submitting = ref(false);
  let rowSeq = 0;

  const isEdit = computed(() => Boolean(props.record?.id));

  interface LanguageForm {
    code: string;
    status: number;
  }

  const getDefaultForm = (): LanguageForm => ({
    code: '',
    status: 1,
  });

  const form = reactive<LanguageForm>(getDefaultForm());
  const nameRows = ref<NameRow[]>([]);

  const rules = computed<Record<string, FieldRule | FieldRule[]>>(() => ({
    code: [{ required: true, message: t('languageModal.rule.code') }],
  }));

  const createEmptyRow = (lang = '', name = ''): NameRow => {
    rowSeq += 1;
    return { key: rowSeq, lang, name };
  };

  const addNameRow = () => {
    nameRows.value.push(createEmptyRow());
  };

  const removeNameRow = (index: number) => {
    if (nameRows.value.length <= 1) return;
    nameRows.value.splice(index, 1);
  };

  const resetForm = () => {
    Object.assign(form, getDefaultForm());
    nameRows.value = [createEmptyRow()];
    formRef.value?.clearValidate();
  };

  const fillForm = (record: LanguageRecord) => {
    Object.assign(form, {
      code: record.code || '',
      status: record.status ?? 1,
    });
    const entries = Object.entries(record.name_i18n || {});
    nameRows.value = entries.length
      ? entries.map(([lang, name]) => createEmptyRow(lang, name))
      : [createEmptyRow()];
  };

  const close = () => {
    emit('update:visible', false);
  };

  const handleCancel = () => {
    close();
    resetForm();
  };

  const buildNameI18n = (): Record<string, string> | null => {
    const result = nameRows.value.reduce<Record<string, string>>((acc, row) => {
      const lang = row.lang.trim();
      const name = row.name.trim();
      if (lang && name) acc[lang] = name;
      return acc;
    }, {});
    return Object.keys(result).length ? result : null;
  };

  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    const errors = await formRef.value?.validate();
    if (errors || submitting.value) {
      done(false);
      return;
    }

    const nameI18n = buildNameI18n();
    if (!nameI18n) {
      Message.warning(t('languageModal.message.nameRequired'));
      done(false);
      return;
    }

    submitting.value = true;
    try {
      if (isEdit.value && props.record?.id) {
        await updateLanguage({
          id: props.record.id,
          name_i18n: nameI18n,
          status: form.status,
        });
        Message.success(t('front.common.message.saveSuccess'));
      } else {
        await createLanguage({
          code: form.code.trim(),
          name_i18n: nameI18n,
          status: form.status,
        });
        Message.success(t('front.common.message.createSuccess'));
      }
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
      if (isEdit.value && props.record) {
        fillForm(props.record);
      }
    }
  );
</script>

<style scoped lang="less">
  .name-i18n-list {
    width: 100%;
  }
  .name-i18n-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 10px;
  }
  .name-i18n-lang {
    width: 140px;
    flex-shrink: 0;
  }
  .name-i18n-name {
    flex: 1;
  }
</style>
