<template>
  <a-modal
    :visible="visible"
    :title="$t('i18nByKeyModal.title')"
    :ok-loading="submitting"
    :ok-text="$t('i18nByKeyModal.okText')"
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
          :label="$t('i18nByKeyModal.field.transKey')"
          field="trans_key"
          required
        >
          <a-input
            v-model="form.trans_key"
            :placeholder="$t('i18nByKeyModal.field.transKeyPlaceholder')"
            allow-clear
            :max-length="128"
            :disabled="Boolean(transKey)"
          />
        </a-form-item>
        <a-form-item
          :label="$t('i18nByKeyModal.field.langs')"
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
                v-model="item.lang"
                :placeholder="$t('i18nByKeyModal.field.langPlaceholder')"
                allow-clear
                :max-length="32"
                class="lang-input"
              />
              <a-textarea
                v-model="item.value"
                :placeholder="$t('i18nByKeyModal.field.valuePlaceholder')"
                :auto-size="{ minRows: 1, maxRows: 3 }"
                allow-clear
                :max-length="2000"
                class="value-input"
              />
            </div>
            <!-- <a-button type="dashed" long @click="addItem">
              <template #icon><icon-plus /></template>
              {{ $t('i18nByKeyModal.action.addLang') }}
            </a-button> -->
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
  import { queryI18nLangList, queryI18nList, updateI18nByKey } from '@/api/i18n';

  interface LangItem {
    uid: number;
    lang: string;
    value: string;
  }

  const { t } = useI18n();

  const props = defineProps<{
    visible: boolean;
    /** 从列表行带入时预填并锁定 */
    transKey?: string;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  let uidSeed = 1;
  const createItem = (lang = '', value = ''): LangItem => {
    uidSeed += 1;
    return {
      uid: uidSeed,
      lang,
      value,
    };
  };

  const formRef = ref<FormInstance>();
  const loading = ref(false);
  const submitting = ref(false);
  const form = reactive<{
    trans_key: string;
    items: LangItem[];
  }>({
    trans_key: '',
    items: [createItem()],
  });

  const rules = computed<Record<string, FieldRule | FieldRule[]>>(() => ({
    trans_key: [{ required: true, message: t('i18nByKeyModal.rule.transKey') }],
  }));

  const resetForm = () => {
    form.trans_key = '';
    form.items = [createItem()];
    formRef.value?.clearValidate();
  };

  const fillForm = (
    transKey: string,
    langs: string[],
    valueMap: Record<string, string>
  ) => {
    form.trans_key = transKey;
    if (langs.length) {
      form.items = langs.map((lang) =>
        createItem(lang, valueMap[lang] || '')
      );
      return;
    }
    const entries = Object.entries(valueMap);
    form.items = entries.length
      ? entries.map(([lang, value]) => createItem(lang, value))
      : [createItem()];
  };

  const loadByKey = async (transKey: string) => {
    loading.value = true;
    try {
      const [langResult, entryResult] = await Promise.allSettled([
        queryI18nLangList({ page: 1, page_size: 200 }),
        queryI18nList({
          trans_key: transKey,
          page: 1,
          page_size: 100,
        }),
      ]);

      const valueMap: Record<string, string> = {};
      if (entryResult.status === 'fulfilled') {
        (entryResult.value.data?.list || []).forEach((item) => {
          if (!item.lang) return;
          valueMap[item.lang] = item.value || '';
        });
      }

      const langs =
        langResult.status === 'fulfilled'
          ? (langResult.value.data?.list || [])
              .map((item) => item.lang)
              .filter(Boolean)
          : [];

      fillForm(transKey, langs, valueMap);
    } catch {
      fillForm(transKey, [], {});
    } finally {
      loading.value = false;
    }
  };

  const addItem = () => {
    form.items.push(createItem());
  };

  const buildDataMap = () => {
    const data: Record<string, string> = {};
    const langs = new Set<string>();
    form.items.forEach((item) => {
      const lang = item.lang.trim();
      if (!lang) return;
      if (langs.has(lang)) {
        throw new Error('DUPLICATE_LANG');
      }
      langs.add(lang);
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

    let data: Record<string, string>;
    try {
      data = buildDataMap();
    } catch {
      Message.warning(t('i18nByKeyModal.message.duplicateLang'));
      done(false);
      return;
    }

    if (!Object.keys(data).length) {
      Message.warning(t('i18nByKeyModal.message.emptyLang'));
      done(false);
      return;
    }

    submitting.value = true;
    try {
      await updateI18nByKey({
        trans_key: form.trans_key.trim(),
        data,
      });
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
    async (visible) => {
      if (!visible) return;
      resetForm();
      const key = (props.transKey || '').trim();
      if (key) {
        await loadByKey(key);
      }
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
    width: 110px;
    flex-shrink: 0;
  }
  .value-input {
    flex: 1;
  }
</style>
