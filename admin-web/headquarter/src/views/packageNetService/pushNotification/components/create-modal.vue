<template>
  <a-modal
    :visible="visible"
    :title="t('pushNotification.create.title')"
    :width="960"
    :ok-text="t('pushNotification.create.action.submit')"
    :ok-loading="submitting"
    :mask-closable="false"
    unmount-on-close
    modal-class="push-notification-create-modal"
    @before-ok="handleBeforeOk"
    @cancel="emit('update:visible', false)"
  >
    <div class="section">
      <div class="section-title">
        {{ t('pushNotification.create.section.base') }}
      </div>
      <div class="base-row">
        <div class="base-field">
          <span class="field-label field-label--required">
            {{ t('pushNotification.create.field.type') }}
          </span>
          <a-select
            v-model="form.type"
            :options="typeOptions"
            :placeholder="t('front.common.placeholder.select')"
            class="base-select"
          />
        </div>
        <div class="base-field">
          <span class="field-label">
            {{ t('pushNotification.create.field.validity') }}
          </span>
          <a-radio-group v-model="form.validity" type="button">
            <a-radio :value="NOTIFICATION_VALIDITY.PERMANENT">
              {{ t('pushNotification.validity.permanent') }}
            </a-radio>
          </a-radio-group>
        </div>
      </div>
    </div>

    <div class="section">
      <div class="section-title section-title--required">
        {{ t('pushNotification.create.section.targets') }}
      </div>
      <OperatorNodePicker v-model="form.operatorIds" />
    </div>

    <div class="section">
      <div class="section-title section-title--required">
        {{ t('pushNotification.create.section.content') }}
      </div>
      <div class="i18n-list">
        <div
          v-for="(row, index) in form.entries"
          :key="row.key"
          class="i18n-card"
        >
          <div class="i18n-card-header">
            <div class="i18n-lang">
              <a-select
                v-model="row.languageCode"
                :options="languageOptions(row)"
                :placeholder="t('pushNotification.create.field.language')"
                :loading="languagesLoading"
                allow-search
              />
            </div>
            <a-button
              type="text"
              status="danger"
              shape="circle"
              class="i18n-remove"
              @click="removeRow(index)"
            >
              <template #icon><icon-minus-circle /></template>
            </a-button>
          </div>
          <div class="i18n-field">
            <span class="field-label field-label--required i18n-label">
              {{ t('pushNotification.create.field.title') }}
            </span>
            <div class="i18n-control">
              <a-input
                v-model="row.title"
                :placeholder="
                  t('pushNotification.create.field.titlePlaceholder')
                "
                :max-length="200"
                show-word-limit
                allow-clear
              />
            </div>
          </div>
          <div class="i18n-field">
            <span class="field-label field-label--required i18n-label">
              {{ t('pushNotification.create.field.content') }}
            </span>
            <div class="i18n-control">
              <a-textarea
                v-model="row.content"
                :placeholder="
                  t('pushNotification.create.field.contentPlaceholder')
                "
                :max-length="5000"
                :auto-size="{ minRows: 4, maxRows: 10 }"
                show-word-limit
              />
            </div>
          </div>
        </div>
        <a-button type="dashed" class="i18n-add" @click="addRow">
          <template #icon><icon-plus /></template>
          {{ t('pushNotification.create.action.addLanguage') }}
        </a-button>
      </div>
    </div>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { queryEnabledLangs } from '@/api/i18n';
  import {
    createNotification,
    NOTIFICATION_TYPE,
    NOTIFICATION_VALIDITY,
    type NotificationI18nText,
  } from '@/api/notification';
  import OperatorNodePicker from './operator-node-picker.vue';

  const props = defineProps<{
    visible: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
    (e: 'success'): void;
  }>();

  const { t } = useI18n();

  interface LanguageOption {
    code: string;
    name: string;
    isDefault: boolean;
  }

  const typeOptions = computed(() => [
    {
      label: t('pushNotification.type.official'),
      value: NOTIFICATION_TYPE.OFFICIAL,
    },
  ]);

  interface I18nEntry {
    key: number;
    languageCode: string | undefined;
    title: string;
    content: string;
  }

  const form = reactive<{
    type: number | undefined;
    validity: number;
    operatorIds: number[];
    entries: I18nEntry[];
  }>({
    type: NOTIFICATION_TYPE.OFFICIAL,
    validity: NOTIFICATION_VALIDITY.PERMANENT,
    operatorIds: [],
    entries: [],
  });

  const submitting = ref(false);
  const languages = ref<LanguageOption[]>([]);
  const languagesLoading = ref(false);
  let rowSeed = 0;

  const languageLabel = (code: string) =>
    languages.value.find((item) => item.code === code)?.name || code;

  /** 语言不可重复，当前组自身已选的语言保留 */
  const languageOptions = (current: I18nEntry) => {
    const used = new Set(
      form.entries
        .filter((row) => row !== current && row.languageCode)
        .map((row) => row.languageCode)
    );
    return languages.value
      .filter((item) => !used.has(item.code))
      .map((item) => ({ label: languageLabel(item.code), value: item.code }));
  };

  /** 新增组优先默认语言，其次第一个未使用的语言；语言用完时为空白组 */
  const addRow = () => {
    const used = new Set(form.entries.map((row) => row.languageCode));
    const unused = languages.value.filter((item) => !used.has(item.code));
    rowSeed += 1;
    form.entries.push({
      key: rowSeed,
      languageCode:
        unused.find((item) => item.isDefault)?.code || unused[0]?.code,
      title: '',
      content: '',
    });
  };

  const removeRow = (index: number) => {
    form.entries.splice(index, 1);
  };

  const loadLanguages = async () => {
    languagesLoading.value = true;
    try {
      const { data } = await queryEnabledLangs();
      languages.value = [...(data?.list || [])]
        .filter((item) => !item.disabled)
        .sort((a, b) => (a.sort_no ?? 0) - (b.sort_no ?? 0))
        .map((item) => ({
          code: item.lang,
          name: item.i18n_name || item.name || item.lang,
          isDefault: item.is_default === 1,
        }));
    } catch {
      languages.value = [];
    } finally {
      languagesLoading.value = false;
    }
  };

  const resetForm = () => {
    form.type = NOTIFICATION_TYPE.OFFICIAL;
    form.validity = NOTIFICATION_VALIDITY.PERMANENT;
    form.operatorIds = [];
    form.entries = [];
  };

  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return;
      resetForm();
      loadLanguages();
    }
  );

  const validate = () => {
    if (!form.type) return t('pushNotification.create.rule.typeRequired');
    if (!form.operatorIds.length) {
      return t('pushNotification.create.rule.targetsRequired');
    }
    if (!form.entries.length) {
      return t('pushNotification.create.rule.entriesRequired');
    }
    for (let i = 0; i < form.entries.length; i += 1) {
      const row = form.entries[i];
      if (!row.languageCode) {
        return t('pushNotification.create.rule.languageRequired', {
          index: i + 1,
        });
      }
      if (!row.title.trim() || !row.content.trim()) {
        return t('pushNotification.create.rule.entryIncomplete', {
          language: languageLabel(row.languageCode),
        });
      }
    }
    return '';
  };

  const toI18nText = (field: 'title' | 'content'): NotificationI18nText =>
    form.entries.reduce<NotificationI18nText>((result, row) => {
      if (row.languageCode) result[row.languageCode] = row[field].trim();
      return result;
    }, {});

  const handleBeforeOk = async () => {
    const error = validate();
    if (error) {
      Message.warning(error);
      return false;
    }
    submitting.value = true;
    try {
      await createNotification({
        type: form.type as number,
        validity: form.validity,
        operator_ids: [...form.operatorIds],
        title_i18n: toI18nText('title'),
        content_i18n: toI18nText('content'),
      });
      Message.success(t('pushNotification.create.message.success'));
      emit('success');
      emit('update:visible', false);
      return true;
    } catch {
      return false;
    } finally {
      submitting.value = false;
    }
  };
</script>

<style scoped lang="less">
  .section + .section {
    margin-top: 20px;
  }

  .section-title {
    position: relative;
    margin-bottom: 12px;
    padding-left: 10px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 14px;
    line-height: 20px;

    &::before {
      position: absolute;
      top: 3px;
      left: 0;
      width: 3px;
      height: 14px;
      background-color: rgb(var(--primary-6));
      border-radius: 2px;
      content: '';
    }

    &--required::after {
      margin-left: 4px;
      color: rgb(var(--danger-6));
      content: '*';
    }
  }

  .base-row {
    display: flex;
    flex-wrap: wrap;
    gap: 16px 40px;
  }

  .base-field {
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .base-select {
    width: 240px;
  }

  .field-label {
    flex-shrink: 0;
    color: var(--color-text-2);
    font-size: 14px;

    &--required::before {
      margin-right: 4px;
      color: rgb(var(--danger-6));
      content: '*';
    }
  }

  .i18n-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .i18n-card {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 12px 16px 16px;
    background-color: var(--color-fill-1);
    border: 1px solid var(--color-border-2);
    border-radius: 8px;
  }

  .i18n-card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-bottom: 12px;
    border-bottom: 1px dashed var(--color-border-2);
  }

  .i18n-lang {
    width: 220px;
  }

  .i18n-remove {
    flex-shrink: 0;
    font-size: 18px;
  }

  .i18n-field {
    display: flex;
    gap: 12px;
    align-items: flex-start;
  }

  .i18n-label {
    width: 56px;
    line-height: 32px;
    text-align: right;
  }

  .i18n-control {
    flex: 1;
    min-width: 0;
  }

  .i18n-add {
    align-self: flex-start;
  }
</style>

<style lang="less">
  .push-notification-create-modal .arco-modal-body {
    max-height: calc(100vh - 220px);
    overflow-y: auto;
  }
</style>
