<template>
  <MemberLevelModal
    :visible="active === 'level'"
    :record="record"
    @update:visible="onVisible"
  />
  <MemberRemarkModal
    :visible="active === 'remark'"
    :record="record"
    @update:visible="onVisible"
  />
  <MemberRestrictionModal
    :visible="active === 'restriction'"
    :record="record"
    @update:visible="onVisible"
  />
</template>

<script lang="ts" setup>
  import { ref } from 'vue';
  import type { MemberRecord } from '../../types';
  import type { OperationCommand } from './commands';
  import MemberLevelModal from './member-level-modal.vue';
  import MemberRemarkModal from './member-remark-modal.vue';
  import MemberRestrictionModal from './member-restriction-modal.vue';

  const active = ref<OperationCommand | null>(null);
  const record = ref<MemberRecord | null>(null);

  function open(command: OperationCommand, target: MemberRecord) {
    record.value = target;
    active.value = command;
  }

  function onVisible(value: boolean) {
    if (!value) active.value = null;
  }

  defineExpose({ open });
</script>
