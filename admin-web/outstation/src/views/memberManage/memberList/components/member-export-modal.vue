<template>
  <a-modal
    :visible="visible"
    :width="480"
    :mask-closable="false"
    :esc-to-close="!submitting"
    :closable="!submitting"
    unmount-on-close
    title-align="start"
    modal-animation-name="mac-modal"
    mask-animation-name="mac-mask"
    modal-class="mac-modal-panel member-export-modal"
    :mask-style="{ backgroundColor: 'rgba(0, 0, 0, 0.35)' }"
    :title="t('memberManage.export.title')"
    :ok-text="t('memberManage.export.submit')"
    :cancel-text="t('memberManage.confirm.cancel')"
    :ok-loading="submitting"
    :cancel-button-props="{ disabled: submitting }"
    @before-ok="handleBeforeOk"
    @cancel="close"
  >
    <div class="export-tip">{{ t('memberManage.export.tip') }}</div>
    <a-checkbox
      :model-value="allChecked"
      :indeterminate="indeterminate"
      :disabled="submitting"
      class="export-all"
      @change="toggleAll"
    >
      {{ t('memberManage.export.selectAll') }}
    </a-checkbox>
    <a-checkbox-group v-model="selected" :disabled="submitting" class="export-cols">
      <a-checkbox
        v-for="column in MEMBER_EXPORT_COLUMNS"
        :key="column.key"
        :value="column.key"
      >
        {{ t(column.label) }}
      </a-checkbox>
    </a-checkbox-group>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import exportMembersFile, { MEMBER_EXPORT_COLUMNS } from '../export';

  const props = defineProps<{
    visible: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
  }>();

  const { t } = useI18n();
  const ALL_KEYS = MEMBER_EXPORT_COLUMNS.map((item) => item.key);

  const selected = ref<string[]>([...ALL_KEYS]);
  const submitting = ref(false);

  const allChecked = computed(() => selected.value.length === ALL_KEYS.length);
  const indeterminate = computed(
    () => selected.value.length > 0 && !allChecked.value
  );

  watch(
    () => props.visible,
    (open) => {
      if (open) selected.value = [...ALL_KEYS];
    }
  );

  function toggleAll(value: boolean | (string | number | boolean)[]) {
    selected.value = value === true ? [...ALL_KEYS] : [];
  }

  function close() {
    if (submitting.value) return;
    emit('update:visible', false);
  }

  async function handleBeforeOk() {
    if (submitting.value) return false;
    if (!selected.value.length) {
      Message.warning(t('memberManage.export.selectFirst'));
      return false;
    }
    // 按固定顺序提交，保证导出文件列顺序稳定
    const columns = ALL_KEYS.filter((key) => selected.value.includes(key));
    submitting.value = true;
    try {
      const { rows } = await exportMembersFile(columns);
      Message.success(t('memberManage.message.exported', { count: rows }));
      emit('update:visible', false);
      return true;
    } catch {
      // 错误提示已由请求拦截器统一弹出
      return false;
    } finally {
      submitting.value = false;
    }
  }
</script>

<style scoped lang="less">
  .export-tip {
    margin-bottom: 12px;
    color: var(--color-text-3);
    font-size: 13px;
  }

  .export-all {
    margin-bottom: 8px;
    padding-bottom: 8px;
    border-bottom: 1px solid var(--color-border-2);
    width: 100%;
  }

  .export-cols {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px 16px;
  }
</style>
