<template>
  <div class="sidebar">
    <div class="section-title">
      选择场景
      <a-tooltip content="新建对话将使用所选场景">
        <icon-question-circle class="section-tip" />
      </a-tooltip>
    </div>
    <a-spin :loading="scenariosLoading" class="scene-spin">
      <div v-if="scenarios.length" class="scene-grid">
        <a-tooltip
          v-for="item in scenarios"
          :key="item.name"
          position="bottom"
          :content-style="{ maxWidth: '260px' }"
        >
          <template #content>
            <div>{{ item.description || item.name }}</div>
            <div class="scene-tip-model">模型：{{ item.model_name }}</div>
          </template>
          <div
            class="scene-card"
            :class="{ 'is-active': item.name === currentScene }"
            @click="emit('selectScene', item.name)"
          >
            <component :is="sceneIcon(item.icon)" class="scene-card__icon" />
            <span class="scene-card__name">{{ item.name }}</span>
          </div>
        </a-tooltip>
      </div>
      <div v-else-if="!scenariosLoading" class="scene-empty">暂无可用场景</div>
    </a-spin>

    <div class="section-title section-title--list">会话列表</div>
    <div class="list-toolbar">
      <a-input
        v-model="keyword"
        allow-clear
        placeholder="搜索会话标题..."
        class="list-search"
      >
        <template #prefix><icon-search /></template>
      </a-input>
      <a-button
        type="primary"
        :disabled="!currentScene"
        @click="emit('newChat')"
      >
        <template #icon><icon-plus /></template>
        新建对话
      </a-button>
    </div>

    <a-spin :loading="sessionsLoading" class="session-spin">
      <div class="session-list">
        <div
          v-for="item in filteredSessions"
          :key="item.session_code"
          class="session-item"
          :class="{ 'is-active': item.session_code === activeCode }"
          @click="emit('selectSession', item)"
        >
          <component
            :is="sceneIcon(sceneMap[item.scene]?.icon)"
            class="session-item__icon"
          />
          <div class="session-item__body">
            <div class="session-item__head">
              <span class="session-item__title">{{ item.title }}</span>
              <span class="session-item__time">
                {{ formatListTime(item.updated_at || item.created_at) }}
              </span>
            </div>
            <div v-if="item.context_summary" class="session-item__summary">
              {{ item.context_summary }}
            </div>
            <div class="session-item__foot">
              <a-tag size="small" :color="sceneColor(item.scene)">
                {{ item.scene }}
              </a-tag>
              <a-tag v-if="item.status === 'CLOSED'" size="small">已关闭</a-tag>
              <span
                v-if="item.message_count != null"
                class="session-item__count"
              >
                <icon-message /> {{ item.message_count }}
              </span>
            </div>
          </div>
        </div>

        <div
          v-if="!sessionsLoading && !filteredSessions.length"
          class="session-empty"
        >
          <icon-robot class="session-empty__icon" />
          <div class="session-empty__title">
            {{ keyword ? '没有匹配的会话' : '开始新的对话吧' }}
          </div>
          <div class="session-empty__desc">选择一个场景，开启智能对话之旅</div>
          <a-button
            v-if="!keyword"
            type="primary"
            size="small"
            :disabled="!currentScene"
            @click="emit('newChat')"
          >
            新建对话
          </a-button>
        </div>
      </div>
    </a-spin>
  </div>
</template>

<script setup lang="ts">
  import { computed, ref } from 'vue';
  import type { ChatScenario, ChatSession } from '@/api/aichat';
  import { formatListTime, sceneIcon } from '../utils';

  const props = defineProps<{
    scenarios: ChatScenario[];
    sessions: ChatSession[];
    currentScene: string;
    activeCode: string;
    scenariosLoading: boolean;
    sessionsLoading: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'selectScene', name: string): void;
    (e: 'selectSession', session: ChatSession): void;
    (e: 'newChat'): void;
  }>();

  const SCENE_COLORS = ['arcoblue', 'cyan', 'green', 'purple', 'orange'];

  const keyword = ref('');

  const sceneMap = computed(() =>
    Object.fromEntries(props.scenarios.map((s) => [s.name, s]))
  );

  const filteredSessions = computed(() => {
    const kw = keyword.value.trim().toLowerCase();
    if (!kw) return props.sessions;
    return props.sessions.filter(
      (s) =>
        s.title?.toLowerCase().includes(kw) ||
        s.context_summary?.toLowerCase().includes(kw)
    );
  });

  function sceneColor(scene: string) {
    const idx = props.scenarios.findIndex((s) => s.name === scene);
    return idx < 0 ? 'gray' : SCENE_COLORS[idx % SCENE_COLORS.length];
  }
</script>

<style scoped lang="less">
  .sidebar {
    display: flex;
    flex-direction: column;
    height: 100%;
    padding: 20px 16px 16px;
    background: var(--color-bg-2);
    border-radius: 8px;
  }

  .section-title {
    display: flex;
    gap: 6px;
    align-items: center;
    margin-bottom: 12px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 15px;

    &--list {
      margin-top: 20px;
    }
  }

  .section-tip {
    color: var(--color-text-3);
    font-size: 14px;
    cursor: help;
  }

  .scene-spin {
    display: block;
    min-height: 40px;
  }

  .scene-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(76px, 1fr));
    gap: 8px;
  }

  .scene-card {
    display: flex;
    flex-direction: column;
    gap: 6px;
    align-items: center;
    justify-content: center;
    min-height: 64px;
    padding: 8px 4px;
    color: var(--color-text-2);
    border: 1px solid var(--color-border-2);
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s;

    &:hover {
      color: rgb(var(--primary-6));
      border-color: rgb(var(--primary-3));
    }

    &.is-active {
      color: rgb(var(--primary-6));
      background: rgb(var(--primary-1));
      border-color: rgb(var(--primary-6));
    }

    &__icon {
      font-size: 20px;
    }

    &__name {
      max-width: 100%;
      overflow: hidden;
      font-size: 12px;
      white-space: nowrap;
      text-overflow: ellipsis;
    }
  }

  .scene-tip-model {
    margin-top: 4px;
    opacity: 0.75;
  }

  .scene-empty {
    color: var(--color-text-3);
    font-size: 13px;
  }

  .list-toolbar {
    display: flex;
    gap: 8px;
    margin-bottom: 12px;
  }

  .list-search {
    flex: 1;
  }

  .session-spin {
    display: block;
    flex: 1;
    min-height: 0;
  }

  .session-list {
    height: 100%;
    margin: 0 -8px;
    padding: 0 8px;
    overflow-y: auto;
  }

  .session-item {
    position: relative;
    display: flex;
    gap: 10px;
    padding: 12px 10px;
    border-radius: 8px;
    cursor: pointer;
    transition: background 0.2s;

    & + & {
      margin-top: 2px;
    }

    &:hover {
      background: var(--color-fill-1);
    }

    &.is-active {
      background: rgb(var(--primary-1));

      &::before {
        position: absolute;
        top: 10px;
        bottom: 10px;
        left: 0;
        width: 3px;
        background: rgb(var(--primary-6));
        border-radius: 2px;
        content: '';
      }
    }

    &__icon {
      flex: none;
      margin-top: 2px;
      color: rgb(var(--primary-6));
      font-size: 18px;
    }

    &__body {
      flex: 1;
      min-width: 0;
    }

    &__head {
      display: flex;
      gap: 8px;
      align-items: center;
      justify-content: space-between;
    }

    &__title {
      overflow: hidden;
      color: var(--color-text-1);
      font-weight: 500;
      white-space: nowrap;
      text-overflow: ellipsis;
    }

    &__time {
      flex: none;
      color: var(--color-text-3);
      font-size: 12px;
    }

    &__summary {
      margin-top: 4px;
      overflow: hidden;
      color: var(--color-text-3);
      font-size: 12px;
      white-space: nowrap;
      text-overflow: ellipsis;
    }

    &__foot {
      display: flex;
      gap: 6px;
      align-items: center;
      margin-top: 6px;
    }

    &__count {
      display: inline-flex;
      gap: 4px;
      align-items: center;
      margin-left: auto;
      color: var(--color-text-3);
      font-size: 12px;
    }
  }

  .session-empty {
    display: flex;
    flex-direction: column;
    gap: 6px;
    align-items: center;
    margin-top: 16px;
    padding: 24px 16px;
    text-align: center;
    background: linear-gradient(180deg, rgb(var(--primary-1)), transparent);
    border-radius: 8px;

    &__icon {
      color: rgb(var(--primary-5));
      font-size: 36px;
    }

    &__title {
      color: var(--color-text-1);
      font-weight: 500;
    }

    &__desc {
      margin-bottom: 8px;
      color: var(--color-text-3);
      font-size: 12px;
    }
  }
</style>
