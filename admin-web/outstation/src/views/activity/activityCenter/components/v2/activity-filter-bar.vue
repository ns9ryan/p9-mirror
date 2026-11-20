<template>
  <div class="filter-grid">
    <div class="filter-field is-keyword">
      <a-input
        v-model="form.keyword"
        allow-clear
        :placeholder="t('discountCenter.activityCenter.v2.search.keyword')"
        @press-enter="submit"
      >
        <template #prefix><icon-search /></template>
      </a-input>
    </div>
    <div class="filter-field">
      <span class="filter-label">
        {{ t('discountCenter.activityCenter.form.activityType') }}
      </span>
      <a-select v-model="form.activityType">
        <a-option value="">{{ allText }}</a-option>
        <a-option
          v-for="item in enumOptions(enumParams, 'type')"
          :key="item.value"
          :value="item.value"
        >
          {{ item.label }}
        </a-option>
      </a-select>
    </div>
    <div class="filter-field">
      <span class="filter-label">
        {{ t('discountCenter.activityCenter.form.client') }}
      </span>
      <a-select v-model="form.client">
        <a-option value="">{{ allText }}</a-option>
        <a-option
          v-for="item in enumOptions(enumParams, 'client_type')"
          :key="item.value"
          :value="item.value"
        >
          {{ item.label }}
        </a-option>
      </a-select>
    </div>
    <div v-if="!hideStatus" class="filter-field">
      <span class="filter-label">
        {{ t('discountCenter.activityCenter.form.status') }}
      </span>
      <a-select v-model="form.status">
        <a-option value="">{{ allText }}</a-option>
        <a-option
          v-for="item in enumOptions(enumParams, 'status')"
          :key="item.value"
          :value="item.value"
        >
          {{ item.label }}
        </a-option>
      </a-select>
    </div>
    <div class="filter-actions">
      <a-button type="primary" @click="submit">
        <template #icon><icon-search /></template>
        {{ t('discountCenter.activityCenter.form.search') }}
      </a-button>
      <a-button @click="reset">
        {{ t('discountCenter.activityCenter.form.reset') }}
      </a-button>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { enumOptions } from '../../options';
  import { useActivityStore } from '../../store';
  import type { ActivityQuery } from '../../types';

  /** 下拉值为接口枚举编码，空字符串表示全部 */
  interface FilterForm {
    keyword: string;
    activityType: number | '';
    client: number | '';
    status: number | '';
  }

  defineProps<{
    hideStatus?: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'search', value: ActivityQuery): void;
  }>();

  const { t } = useI18n();
  const { enumParams } = useActivityStore();
  const allText = computed(() =>
    t('discountCenter.activityCenter.form.selectDefault')
  );

  const createForm = (): FilterForm => ({
    keyword: '',
    activityType: '',
    client: '',
    status: '',
  });

  const form = reactive<FilterForm>(createForm());

  function submit() {
    emit('search', {
      name: form.keyword.trim() || undefined,
      type: form.activityType || undefined,
      client: form.client || undefined,
      status: form.status || undefined,
    });
  }

  function reset() {
    Object.assign(form, createForm());
    submit();
  }
</script>

<style scoped lang="less">
  .filter-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    gap: 12px;
    align-items: end;
  }

  .is-keyword {
    grid-column: span 2;
  }

  .filter-field {
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-width: 0;
  }

  .filter-label {
    color: var(--color-text-2);
    font-size: 13px;
  }

  .filter-actions {
    display: flex;
    gap: 8px;
    white-space: nowrap;
  }
</style>
