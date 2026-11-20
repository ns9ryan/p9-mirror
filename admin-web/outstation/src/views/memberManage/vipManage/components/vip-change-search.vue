<template>
  <div class="change-search">
    <div class="search-row">
      <label class="search-item">
        <span class="search-label">{{ t('vipManage.record.account') }}</span>
        <a-input
          v-model="form.account"
          class="search-control"
          allow-clear
          :placeholder="t('vipManage.record.accountPlaceholder')"
          @press-enter="search"
        />
      </label>
      <label class="search-item">
        <span class="search-label">{{ t('vipManage.change.type') }}</span>
        <a-select
          v-model="form.type"
          class="search-control"
          allow-clear
          :placeholder="t('front.common.placeholder.select')"
        >
          <a-option v-for="type in types" :key="type" :value="type">
            {{ t(`vipManage.change.typeText.${type}`) }}
          </a-option>
        </a-select>
      </label>
      <label class="search-item">
        <span class="search-label">{{ t('vipManage.change.level') }}</span>
        <a-select
          v-model="form.level"
          class="search-control"
          allow-clear
          :placeholder="t('front.common.placeholder.select')"
        >
          <a-option v-for="level in levels" :key="level" :value="level">
            VIP{{ level }}
          </a-option>
        </a-select>
      </label>
      <label class="search-item">
        <span class="search-label">{{ t('vipManage.change.range') }}</span>
        <a-range-picker
          v-model="form.dateRange"
          class="range-control"
          allow-clear
          value-format="YYYY-MM-DD"
          :placeholder="[
            t('vipManage.change.startDate'),
            t('vipManage.change.endDate'),
          ]"
        />
      </label>
      <a-space :size="12">
        <a-button type="primary" @click="search">
          <template #icon><icon-search /></template>
          {{ t('front.common.search.search') }}
        </a-button>
        <a-button @click="reset">
          <template #icon><icon-refresh /></template>
          {{ t('front.common.search.reset') }}
        </a-button>
      </a-space>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { reactive, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { createChangeQuery } from '../mock';
  import type { LevelChangeType, VipLevelChangeQuery } from '../types';

  const props = defineProps<{
    query: VipLevelChangeQuery;
    levels: number[];
  }>();

  const emit = defineEmits<{
    (e: 'search', query: VipLevelChangeQuery): void;
  }>();

  const { t } = useI18n();
  const types: LevelChangeType[] = ['promote', 'demote', 'manual', 'recalc'];
  const form = reactive<VipLevelChangeQuery>(createChangeQuery());

  watch(
    () => props.query,
    (value) => {
      Object.assign(form, value, { dateRange: [...value.dateRange] });
    },
    { immediate: true }
  );

  function search() {
    emit('search', { ...form, dateRange: [...(form.dateRange || [])] });
  }

  function reset() {
    Object.assign(form, createChangeQuery());
    search();
  }
</script>

<style scoped lang="less">
  .change-search {
    margin-top: 16px;
    padding: 16px 20px;
    background: var(--color-fill-1);
    border-radius: 6px;
  }

  .search-row {
    display: flex;
    flex-wrap: wrap;
    gap: 12px 24px;
    align-items: center;
  }

  .search-item {
    display: inline-flex;
    gap: 8px;
    align-items: center;
  }

  .search-label {
    flex: none;
    color: var(--color-text-1);
    white-space: nowrap;

    &::after {
      content: ':';
    }
  }

  .search-control {
    width: 200px;
  }

  .range-control {
    width: 260px;
  }
</style>
