<template>
  <a-modal
    :visible="visible"
    :title="isEdit ? $t('i18nModal.title.edit') : $t('i18nModal.title.create')"
    :ok-loading="submitting"
    :ok-text="
      isEdit ? $t('front.common.okText.save') : $t('front.common.okText.create')
    "
    :mask-closable="false"
    :esc-to-close="false"
    unmount-on-close
    :width="640"
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    @before-ok="handleBeforeOk"
    @cancel="handleCancel"
  >
    <a-spin :loading="loading" style="width: 100%">
      <a-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-align="right"
        :label-col-props="{ span: 5 }"
        :wrapper-col-props="{ span: 18 }"
      >
        <a-form-item
          :label="$t('i18nModal.field.group')"
          field="i18n_group"
          required
        >
          <a-input
            v-model="form.i18n_group"
            :placeholder="$t('i18nModal.field.groupPlaceholder')"
            allow-clear
            :max-length="64"
          />
        </a-form-item>
        <a-form-item
          :label="$t('i18nModal.field.transKey')"
          field="trans_key"
          required
        >
          <a-input
            v-model="form.trans_key"
            :placeholder="$t('i18nModal.field.transKeyPlaceholder')"
            allow-clear
            :max-length="128"
          />
        </a-form-item>
        <a-form-item
          :label="$t('i18nModal.field.langs')"
          field="items"
          required
        >
          <div class="lang-list">
            <div
              v-for="item in form.items"
              :key="item.uid"
              class="lang-row"
            >
              <a-input
                :model-value="item.label"
                disabled
                class="lang-input"
              />
              <a-textarea
                v-model="item.value"
                :placeholder="$t('i18nModal.field.valuePlaceholder')"
                :auto-size="{ minRows: 1, maxRows: 3 }"
                allow-clear
                :max-length="2000"
                class="value-input"
              />
            </div>
          </div>
        </a-form-item>
      </a-form>
    </a-spin>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FormInstance, FieldRule } from '@arco-design/web-vue';
  import {
    createI18n,
    queryI18nLangList,
    updateI18n,
    type I18nRecord,
  } from '@/api/i18n';

  interface LangItem {
    uid: number;
    lang: string;
    label: string;
    value: string;
  }

  const { t } = useI18n();

  const props = defineProps<{
    visible: boolean;
    record?: I18nRecord | null;
    /** 当前站点，新建时写入 */
    i18nCode?: string;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  let uidSeed = 1;
  const createItem = (lang = '', label = '', value = ''): LangItem => {
    uidSeed += 1;
    return {
      uid: uidSeed,
      lang,
      label: label || lang,
      value,
    };
  };

  const formRef = ref<FormInstance>();
  const loading = ref(false);
  const submitting = ref(false);

  const isEdit = computed(() => Boolean(props.record?.id));

  const form = reactive<{
    i18n_group: string;
    trans_key: string;
    items: LangItem[];
  }>({
    i18n_group: '',
    trans_key: '',
    items: [],
  });

  const rules = computed<Record<string, FieldRule | FieldRule[]>>(() => ({
    i18n_group: [{ required: true, message: t('i18nModal.rule.group') }],
    trans_key: [{ required: true, message: t('i18nModal.rule.transKey') }],
  }));

  const resetForm = () => {
    form.i18n_group = '';
    form.trans_key = '';
    form.items = [];
    formRef.value?.clearValidate();
  };

  const buildLangLabel = (lang: string, name?: string) =>
    name ? `${name} (${lang})` : lang;

  const loadLangItems = async (valueMap: Record<string, string> = {}) => {
    const { data } = await queryI18nLangList({ page: 1, page_size: 200 });
    form.items = (data?.list || [])
      .filter((item) => item.lang)
      .map((item) =>
        createItem(
          item.lang,
          buildLangLabel(item.lang, item.name),
          valueMap[item.lang] || ''
        )
      );
  };

  const loadFormData = async () => {
    loading.value = true;
    try {
      if (isEdit.value && props.record) {
        const lang = (props.record.lang || '').trim();
        form.i18n_group = props.record.i18n_group || '';
        form.trans_key = (props.record.trans_key || '').trim();
        let label = lang;
        try {
          const { data } = await queryI18nLangList({ page: 1, page_size: 200 });
          const matched = (data?.list || []).find((item) => item.lang === lang);
          if (matched) {
            label = buildLangLabel(matched.lang, matched.name);
          }
        } catch {
          // keep lang as label
        }
        form.items = [createItem(lang, label, props.record.value || '')];
      } else {
        await loadLangItems();
      }
    } catch {
      form.items = [];
    } finally {
      loading.value = false;
    }
  };

  const buildDataMap = () => {
    const data: Record<string, string> = {};
    form.items.forEach((item) => {
      const lang = item.lang.trim();
      if (!lang) return;
      data[lang] = item.value.trim();
    });
    return data;
  };

  const close = () => {
    emit('update:visible', false);
  };

  const handleCancel = () => {
    close();
    resetForm();
  };

  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    const errors = await formRef.value?.validate();
    if (errors || submitting.value) {
      done(false);
      return;
    }

    if (!form.items.length) {
      Message.warning(t('i18nModal.message.emptyLang'));
      done(false);
      return;
    }

    const data = buildDataMap();
    const filledData = Object.fromEntries(
      Object.entries(data).filter(([, value]) => value)
    );

    if (!Object.keys(filledData).length) {
      Message.warning(t('i18nModal.message.emptyLang'));
      done(false);
      return;
    }

    const i18nGroup = form.i18n_group.trim();
    const transKey = form.trans_key.trim();

    submitting.value = true;
    try {
      if (isEdit.value && props.record?.id) {
        const currentLang = (props.record.lang || form.items[0]?.lang || '').trim();
        await updateI18n({
          id: props.record.id,
          i18n_group: i18nGroup,
          lang: currentLang,
          trans_key: transKey,
          value: data[currentLang] ?? form.items[0]?.value?.trim() ?? '',
        });
        Message.success(t('front.common.message.saveSuccess'));
      } else {
        await Promise.all(
          Object.entries(filledData).map(([lang, value]) =>
            createI18n({
              i18n_code: props.i18nCode || props.record?.i18n_code,
              i18n_group: i18nGroup,
              lang,
              trans_key: transKey,
              value,
            })
          )
        );
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
    async (visible) => {
      if (!visible) return;
      resetForm();
      await loadFormData();
    }
  );
</script>

<style scoped lang="less">
  .lang-list {
    width: 100%;
  }
  .lang-row {
    display: flex;
    gap: 8px;
    align-items: flex-start;
    margin-bottom: 10px;
  }
  .lang-input {
    width: 140px;
    flex-shrink: 0;
  }
  .value-input {
    flex: 1;
  }
</style>
