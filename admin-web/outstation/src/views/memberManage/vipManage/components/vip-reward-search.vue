<template>
  <a-form
    :model="form"
    class="reward-search"
    :label-col-props="labelColProps"
    :wrapper-col-props="{ flex: '1' }"
  >
    <a-row :gutter="24">
      <a-col :span="6">
        <a-form-item :label="t('vipManage.record.account')">
          <a-input
            v-model="form.account"
            allow-clear
            :placeholder="t('vipManage.record.accountPlaceholder')"
            @press-enter="search"
          />
        </a-form-item>
      </a-col>
      <a-col :span="6">
        <a-form-item :label="t('vipManage.record.nickname')">
          <a-input
            v-model="form.nickname"
            allow-clear
            :placeholder="t('vipManage.record.nicknamePlaceholder')"
            @press-enter="search"
          />
        </a-form-item>
      </a-col>
      <a-col :span="6">
        <a-form-item :label="t('vipManage.record.level')">
          <a-select
            v-model="form.level"
            allow-clear
            :placeholder="t('vipManage.record.all')"
          >
            <a-option v-for="level in levels" :key="level" :value="level">
              VIP{{ level }}
            </a-option>
          </a-select>
        </a-form-item>
      </a-col>
      <a-col :span="6">
        <a-form-item :label="t('vipManage.record.status')">
          <a-select
            v-model="form.status"
            allow-clear
            :placeholder="t('vipManage.record.all')"
          >
            <a-option v-for="status in statuses" :key="status" :value="status">
              {{ t(`vipManage.record.statusText.${status}`) }}
            </a-option>
          </a-select>
        </a-form-item>
      </a-col>
      <a-col :span="6">
        <a-form-item :label="t('vipManage.record.mode')">
          <a-select
            v-model="form.mode"
            allow-clear
            :placeholder="t('vipManage.record.all')"
          >
            <a-option v-for="mode in modes" :key="mode" :value="mode">
              {{ t(`vipManage.record.modeText.${mode}`) }}
            </a-option>
          </a-select>
        </a-form-item>
      </a-col>
      <a-col :span="12">
        <a-form-item :label="t('vipManage.record.time')">
          <a-range-picker
            v-model="form.dateRange"
            class="range-picker"
            show-time
            allow-clear
            value-format="YYYY-MM-DD HH:mm:ss"
            format="YYYY-MM-DD HH:mm:ss"
            :placeholder="[
              t('memberManage.search.start'),
              t('memberManage.search.end'),
            ]"
          />
        </a-form-item>
      </a-col>
      <a-col :span="6" class="search-actions">
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
      </a-col>
    </a-row>
  </a-form>
</template>

<script lang="ts" setup>
  import { computed, reactive, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { createRewardQuery } from '../mock';
  import type { DispatchMode, RewardStatus, VipRewardQuery } from '../types';

  const props = defineProps<{
    query: VipRewardQuery;
    levels: number[];
  }>();

  const emit = defineEmits<{
    (e: 'search', query: VipRewardQuery): void;
  }>();

  const { t, locale } = useI18n();
  const labelColProps = computed(() => ({
    flex: locale.value.startsWith('zh') ? '80px' : '120px',
  }));
  const statuses: RewardStatus[] = ['claimed', 'dispatched', 'pending'];
  const modes: DispatchMode[] = ['manual', 'auto'];
  const form = reactive<VipRewardQuery>(createRewardQuery());

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
    Object.assign(form, createRewardQuery());
    search();
  }
</script>

<style scoped lang="less">
  .reward-search {
    margin-top: 16px;

    :deep(.arco-form-item) {
      margin-bottom: 16px;
    }
  }

  .range-picker {
    width: 100%;
  }

  .search-actions {
    display: flex;
    justify-content: flex-end;
    margin-bottom: 16px;
  }
</style>
