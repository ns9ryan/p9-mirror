<template>
  <a-modal
    :visible="visible"
    :width="520"
    :mask-closable="false"
    unmount-on-close
    title-align="start"
    :title="t('memberManage.ops.level.title')"
    :ok-text="t('front.common.okText.save')"
    @before-ok="handleBeforeOk"
    @cancel="close"
  >
    <a-form :model="form" layout="vertical">
      <a-form-item :label="t('memberManage.ops.level.auto')">
        <a-select
          v-model="form.autoLevelId"
          :placeholder="t('memberManage.ops.level.autoPlaceholder')"
          allow-search
        >
          <a-option
            v-for="item in autoLevels"
            :key="item.id"
            :value="item.id"
            :label="item.name"
          />
        </a-select>
      </a-form-item>
      <a-form-item
        :label="t('memberManage.ops.level.fixed')"
        :extra="t('memberManage.ops.level.fixedExtra')"
      >
        <a-select
          v-model="form.fixedLevelIds"
          :placeholder="t('memberManage.ops.level.fixedPlaceholder')"
          multiple
          allow-clear
          allow-search
          :limit="FIXED_LEVEL_LIMIT"
        >
          <a-option
            v-for="item in fixedLevels"
            :key="item.id"
            :value="item.id"
            :label="item.name"
          />
        </a-select>
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { reactive, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { useLevelOptions } from '../../lookups';
  import { useMemberStore, type LevelChange } from '../../store';
  import type { MemberRecord } from '../../types';

  const FIXED_LEVEL_LIMIT = 100;

  const props = defineProps<{
    visible: boolean;
    record: MemberRecord | null;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void;
  }>();

  const { t } = useI18n();
  const { updateLevels } = useMemberStore();
  const { autoLevels, fixedLevels, load: loadLevels } = useLevelOptions();

  const form = reactive({
    autoLevelId: undefined as number | undefined,
    fixedLevelIds: [] as number[],
  });

  watch(
    () => props.visible,
    (open) => {
      if (!open || !props.record) return;
      form.autoLevelId = props.record.autoLevelId || undefined;
      form.fixedLevelIds = [...props.record.fixedLevelIds];
      loadLevels();
    }
  );

  function sameIds(a: number[], b: number[]) {
    if (a.length !== b.length) return false;
    const set = new Set(a);
    return b.every((id) => set.has(id));
  }

  function close() {
    emit('update:visible', false);
  }

  async function handleBeforeOk() {
    const { record } = props;
    if (!record) return false;
    const change: LevelChange = {};
    if (form.autoLevelId && form.autoLevelId !== record.autoLevelId) {
      change.autoLevelId = form.autoLevelId;
    }
    if (!sameIds(form.fixedLevelIds, record.fixedLevelIds)) {
      change.fixedLevelIds = [...form.fixedLevelIds];
    }
    if (change.autoLevelId == null && !change.fixedLevelIds) {
      Message.info(t('memberManage.ops.noChange'));
      return false;
    }
    try {
      await updateLevels(record.id, change);
      Message.success(t('memberManage.ops.level.success'));
      close();
      return true;
    } catch {
      return false;
    }
  }
</script>
