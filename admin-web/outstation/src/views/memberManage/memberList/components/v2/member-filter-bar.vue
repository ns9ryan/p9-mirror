<template>
  <div class="filter-bar">
    <div class="filter-grid">
      <div class="filter-field is-keyword">
        <a-input
          v-model="form.keyword"
          allow-clear
          :placeholder="t('memberManage.v2.search.keywordPlaceholder')"
          @press-enter="submit"
        >
          <template #prefix><icon-search /></template>
        </a-input>
      </div>
      <div class="filter-field">
        <span class="filter-label">{{ t('memberManage.v2.search.status') }}</span>
        <a-select v-model="form.status">
          <a-option value="">{{ t('memberManage.v2.search.all') }}</a-option>
          <a-option v-for="item in STATUSES" :key="item" :value="item">
            {{ t(`memberManage.status.${item}`) }}
          </a-option>
        </a-select>
      </div>
      <div class="filter-field">
        <span class="filter-label">{{ t('memberManage.v2.search.agent') }}</span>
        <a-select
          v-model="form.agentId"
          allow-search
          :loading="agentLoading"
          :filter-option="false"
          @search="searchAgents"
          @change="onAgentChange"
          @popup-visible-change="onAgentPopup"
        >
          <a-option value="">{{ t('memberManage.v2.search.all') }}</a-option>
          <a-option
            v-for="item in agentOptions"
            :key="item.id"
            :value="item.id"
          >
            {{ item.account }}
          </a-option>
        </a-select>
      </div>
      <div class="filter-field">
        <span class="filter-label">
          {{ t('memberManage.v2.search.registerSource') }}
        </span>
        <a-select v-model="form.registerChannel" allow-create>
          <a-option value="">{{ t('memberManage.v2.search.all') }}</a-option>
          <a-option v-for="item in REGISTER_CHANNELS" :key="item" :value="item">
            {{ registerChannelText(item, t) }}
          </a-option>
        </a-select>
      </div>
      <div class="filter-field">
        <span class="filter-label">
          {{ t('memberManage.v2.search.registerType') }}
        </span>
        <a-select v-model="form.registerMethod" allow-create>
          <a-option value="">{{ t('memberManage.v2.search.all') }}</a-option>
          <a-option v-for="item in REGISTER_METHODS" :key="item" :value="item">
            {{ registerMethodText(item, t) }}
          </a-option>
        </a-select>
      </div>
      <div class="filter-actions">
        <a-button type="primary" @click="submit">
          <template #icon><icon-search /></template>
          {{ t('memberManage.v2.search.submit') }}
        </a-button>
        <a-button @click="reset">
          {{ t('memberManage.v2.search.reset') }}
        </a-button>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { AgentListItem } from '@/api/agent';
  import {
    REGISTER_CHANNELS,
    REGISTER_METHODS,
    registerChannelText,
    registerMethodText,
  } from '../../format';
  import { useAgentSearch } from '../../lookups';
  import { createMemberBriefQuery } from '../../query';
  import type { AccountStatus, MemberBriefQuery } from '../../types';

  const emit = defineEmits<{
    (e: 'search', value: MemberBriefQuery): void;
  }>();

  const STATUSES: AccountStatus[] = ['normal', 'frozen', 'disabled'];

  const { t } = useI18n();
  const form = reactive<MemberBriefQuery>(createMemberBriefQuery());
  const {
    agents,
    loading: agentLoading,
    search: searchAgents,
  } = useAgentSearch();
  const agentsLoaded = ref(false);
  /** 搜索结果会被下一次搜索替换，选中项单独保存并始终保留在选项里 */
  const selectedAgent = ref<AgentListItem | null>(null);

  const agentOptions = computed(() => {
    const selected = selectedAgent.value;
    if (!selected) return agents.value;
    return [
      selected,
      ...agents.value.filter((item) => item.id !== selected.id),
    ];
  });

  function onAgentChange(value: unknown) {
    selectedAgent.value =
      agentOptions.value.find((item) => item.id === value) || null;
  }

  function onAgentPopup(visible: boolean) {
    if (visible && !agentsLoaded.value) {
      agentsLoaded.value = true;
      searchAgents();
    }
  }

  function submit() {
    emit('search', { ...form });
  }

  function clear() {
    Object.assign(form, createMemberBriefQuery());
  }

  function reset() {
    clear();
    submit();
  }

  defineExpose({ clear });
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
