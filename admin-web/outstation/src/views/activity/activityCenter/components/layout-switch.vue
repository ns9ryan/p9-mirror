<template>
  <a-radio-group v-model="layout" type="button" size="small">
    <a-radio value="classic">
      {{ t('discountCenter.activityCenter.layout.classic') }}
    </a-radio>
    <a-radio value="v2">{{
      t('discountCenter.activityCenter.layout.v2')
    }}</a-radio>
  </a-radio-group>
</template>

<script lang="ts" setup>
  import { useI18n } from 'vue-i18n';
  import { useActivityLayout } from '../layout';

  const { t } = useI18n();
  const layout = useActivityLayout();
</script>
