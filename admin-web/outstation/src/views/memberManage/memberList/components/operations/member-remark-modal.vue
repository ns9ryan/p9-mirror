<template>
  <a-modal
    :visible="visible"
    :width="520"
    :mask-closable="false"
    unmount-on-close
    title-align="start"
    :title="t('memberManage.ops.remark.title')"
    :ok-text="t('front.common.okText.save')"
    @before-ok="handleBeforeOk"
    @cancel="close"
  >
    <a-form :model="form" layout="vertical">
      <a-form-item
        :label="t('memberManage.ops.remark.label')"
        :extra="t('memberManage.ops.remark.extra')"
      >
        <a-textarea
          v-model="form.remark"
          :max-length="200"
          show-word-limit
          allow-clear
          :auto-size="{ minRows: 3, maxRows: 6 }"
          :placeholder="t('memberManage.ops.remark.placeholder')"
        />
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { reactive, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { useMemberStore } from '../../store';
  import type { MemberRecord } from '../../types';

  const props = defineProps<{
    visible: boolean;
    record: MemberRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
  }>();

  const { t } = useI18n();
  const { updateRemark } = useMemberStore();

  const form = reactive({ remark: '' });

  watch(
    () => props.visible,
    (open) => {
      if (open) form.remark = props.record?.remark || '';
    }
  );

  function close() {
    emit('update:visible', false);
  }

  async function handleBeforeOk() {
    const { record } = props;
    if (!record) return false;
    const remark = form.remark.trim();
    if (remark === record.remark) {
      Message.info(t('memberManage.ops.noChange'));
      return false;
    }
    try {
      await updateRemark(record.id, remark);
      Message.success(t('memberManage.ops.remark.success'));
      close();
      return true;
    } catch {
      return false;
    }
  }
</script>
