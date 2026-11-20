<template>
  <div class="msg" :class="`msg--${isUser ? 'user' : 'assistant'}`">
    <a-avatar :size="36" class="msg__avatar">
      <icon-user v-if="isUser" />
      <icon-robot v-else />
    </a-avatar>

    <div class="msg__main">
      <div class="msg__bubble">
        <template v-for="(block, idx) in bodyBlocks" :key="idx">
          <div v-if="block.type === 'text'" class="msg__text">{{
            block.text
          }}</div>
          <ResultTable v-else-if="block.type === 'table'" :table="block" />
          <pre v-else class="msg__code"><code>{{ block.code }}</code></pre>
        </template>

        <div v-if="message.isNeedConfirm" class="confirm">
          <div class="confirm__title">
            <icon-exclamation-circle-fill class="confirm__icon" />
            {{
              actionable
                ? '包含可执行的 SQL 语句，是否确认执行？'
                : '包含可执行的 SQL 语句'
            }}
          </div>
          <pre
            v-for="(block, idx) in sqlBlocks"
            :key="idx"
            class="confirm__sql"
          ><code>{{ block.code }}</code></pre>
          <div v-if="actionable" class="confirm__actions">
            <a-button
              type="primary"
              size="small"
              :loading="busy"
              @click="emit('confirm')"
            >
              <template #icon><icon-check /></template>
              确认
            </a-button>
            <a-button size="small" :disabled="busy" @click="emit('cancel')">
              取消
            </a-button>
          </div>
        </div>
      </div>

      <div class="msg__meta">
        <template v-if="message.status === 'failed'">
          <span class="msg__failed" @click="emit('retry')">
            <icon-exclamation-circle-fill /> 发送失败，点击重试
          </span>
        </template>
        <template v-else>
          <span>{{ formatMessageTime(message.time) }}</span>
          <span v-if="message.latencyMs">
            耗时 {{ (message.latencyMs / 1000).toFixed(1) }}s
          </span>
          <span v-if="message.totalTokens"
            >{{ message.totalTokens }} tokens</span
          >
          <icon-loading v-if="message.status === 'sending'" />
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue';
  import type { ChatUiMessage, ContentBlock } from '../utils';
  import { formatMessageTime, parseContent } from '../utils';
  import ResultTable from './result-table.vue';

  const props = defineProps<{
    message: ChatUiMessage;
    /** 是否展示确认/取消按钮 */
    actionable?: boolean;
    busy?: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'confirm'): void;
    (e: 'cancel'): void;
    (e: 'retry'): void;
  }>();

  const isUser = computed(() => props.message.role === 'user');

  const blocks = computed<ContentBlock[]>(() =>
    isUser.value
      ? [{ type: 'text', text: props.message.content }]
      : parseContent(props.message.content, props.message.contentType)
  );

  const isSql = (b: ContentBlock) => b.type === 'code' && b.lang === 'sql';

  // 需确认的消息把 SQL 代码块移到确认卡片里展示
  const sqlBlocks = computed(() =>
    props.message.isNeedConfirm
      ? (blocks.value.filter(isSql) as Extract<
          ContentBlock,
          { type: 'code' }
        >[])
      : []
  );

  const bodyBlocks = computed(() =>
    props.message.isNeedConfirm
      ? blocks.value.filter((b) => !isSql(b))
      : blocks.value
  );
</script>

<style scoped lang="less">
  .msg {
    display: flex;
    gap: 12px;
    align-items: flex-start;

    & + & {
      margin-top: 20px;
    }

    &__avatar {
      flex: none;
    }

    &__main {
      display: flex;
      flex-direction: column;
      min-width: 0;
      max-width: min(820px, calc(100% - 96px));
    }

    &__bubble {
      padding: 12px 16px;
      color: var(--color-text-1);
      line-height: 1.7;
      word-break: break-word;
      border-radius: 12px;
    }

    &__text {
      white-space: pre-wrap;

      & + & {
        margin-top: 8px;
      }
    }

    &__code {
      margin: 8px 0;
      padding: 10px 12px;
      overflow-x: auto;
      font-size: 12px;
      font-family: Menlo, Consolas, monospace;
      line-height: 1.6;
      white-space: pre;
      background: var(--color-fill-2);
      border-radius: 6px;
    }

    &__meta {
      display: flex;
      gap: 10px;
      align-items: center;
      margin-top: 6px;
      color: var(--color-text-3);
      font-size: 12px;
    }

    &__failed {
      color: rgb(var(--danger-6));
      cursor: pointer;
    }

    &--assistant {
      .msg__avatar {
        color: rgb(var(--primary-6));
        background: rgb(var(--primary-1));
      }

      .msg__bubble {
        background: var(--color-bg-2);
        border: 1px solid var(--color-border-2);
        border-top-left-radius: 4px;
      }
    }

    &--user {
      flex-direction: row-reverse;

      .msg__avatar {
        background: rgb(var(--primary-6));
      }

      .msg__main {
        align-items: flex-end;
      }

      .msg__bubble {
        background: rgb(var(--primary-1));
        border-top-right-radius: 4px;
      }
    }
  }

  .confirm {
    margin-top: 12px;
    padding: 12px 14px;
    background: rgb(var(--orange-1));
    border: 1px solid rgb(var(--orange-2));
    border-radius: 8px;

    &__title {
      display: flex;
      gap: 8px;
      align-items: center;
      font-weight: 500;
    }

    &__icon {
      color: rgb(var(--orange-6));
      font-size: 16px;
    }

    &__sql {
      margin: 10px 0 0;
      padding: 8px 12px;
      overflow-x: auto;
      font-size: 12px;
      font-family: Menlo, Consolas, monospace;
      line-height: 1.6;
      white-space: pre;
      background: var(--color-bg-2);
      border-left: 3px solid rgb(var(--orange-5));
      border-radius: 4px;
    }

    &__actions {
      display: flex;
      gap: 8px;
      margin-top: 12px;
    }
  }
</style>
