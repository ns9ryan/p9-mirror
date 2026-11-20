<template>
  <div class="container">
    <div class="chat-page">
      <SessionSidebar
        class="chat-page__sidebar"
        :scenarios="scenarios"
        :sessions="sessions"
        :current-scene="currentScene"
        :active-code="activeCode"
        :scenarios-loading="scenariosLoading"
        :sessions-loading="sessionsLoading"
        @select-scene="selectScene"
        @select-session="openSession"
        @new-chat="startDraft"
      />

      <div class="chat">
        <div class="chat__header">
          <div class="chat__header-icon">
            <component :is="sceneIcon(currentScenario?.icon)" />
          </div>
          <div class="chat__header-body">
            <div class="chat__title">
              {{ activeSession?.title || '新对话' }}
              <a-tag v-if="currentScene" size="small" color="arcoblue">
                {{ currentScene }}
              </a-tag>
              <a-tag v-if="currentScenario?.model_name" size="small">
                {{ currentScenario.model_name }}
              </a-tag>
            </div>
            <div class="chat__subtitle">
              {{
                activeSession?.context_summary ||
                currentScenario?.description ||
                '选择场景后输入问题，开始对话'
              }}
            </div>
          </div>
          <a-button
            v-if="activeCode"
            size="small"
            :loading="messagesLoading"
            @click="loadMessages"
          >
            <template #icon><icon-refresh /></template>
            刷新
          </a-button>
        </div>

        <div ref="scrollRef" class="chat__messages">
          <a-spin
            :loading="messagesLoading"
            tip="加载消息中..."
            class="chat__spin"
          >
            <div v-if="hasMore" class="chat__more">
              <a-button
                type="text"
                size="mini"
                :loading="loadingMore"
                @click="loadMore"
              >
                加载更早的消息
              </a-button>
            </div>

            <div
              v-if="!messages.length && !messagesLoading && !sending"
              class="welcome"
            >
              <div class="welcome__icon">
                <component :is="sceneIcon(currentScenario?.icon)" />
              </div>
              <div class="welcome__title">
                {{ currentScenario?.name || 'AI 聊天助手' }}
              </div>
              <div class="welcome__desc">
                {{
                  currentScenario?.description ||
                  '请先在左侧选择一个场景，然后在下方输入你的问题'
                }}
              </div>
              <div
                v-if="currentScenario?.available_tools?.length"
                class="welcome__tools"
              >
                <span>可用工具：</span>
                <a-tag
                  v-for="tool in currentScenario.available_tools"
                  :key="tool"
                  size="small"
                >
                  {{ tool }}
                </a-tag>
              </div>
            </div>

            <MessageItem
              v-for="msg in messages"
              :key="msg.key"
              :message="msg"
              :actionable="msg.key === actionableKey"
              :busy="sending"
              @confirm="send(CONFIRM_TEXT)"
              @cancel="dismissConfirm(msg.key)"
              @retry="retry(msg)"
            />

            <div v-if="sending" class="pending">
              <a-avatar :size="36" class="pending__avatar">
                <icon-robot />
              </a-avatar>
              <div class="pending__bubble">
                <icon-loading />
                {{ pendingText }}
              </div>
            </div>
          </a-spin>
        </div>

        <div class="chat__input">
          <a-textarea
            v-model="draft"
            class="chat__textarea"
            :placeholder="
              currentScene
                ? '请输入您的问题，Enter 发送，Shift + Enter 换行'
                : '请先选择场景'
            "
            :disabled="!currentScene"
            :auto-size="{ minRows: 1, maxRows: 6 }"
            :max-length="4000"
            @keydown="onKeydown"
          />
          <a-button
            type="primary"
            :loading="sending"
            :disabled="!draft.trim() || !currentScene"
            @click="send(draft)"
          >
            <template #icon><icon-send /></template>
            发送
          </a-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed, nextTick, onMounted, reactive, ref } from 'vue';
  import dayjs from 'dayjs';
  import {
    createChatSession,
    queryChatMessages,
    queryChatScenarios,
    queryChatSessions,
    sendChatMessage,
  } from '@/api/aichat';
  import type {
    ChatHistoryMessage,
    ChatScenario,
    ChatSession,
  } from '@/api/aichat';
  import SessionSidebar from './components/session-sidebar.vue';
  import MessageItem from './components/message-item.vue';
  import type { ChatUiMessage } from './utils';
  import { createRequestId, sceneIcon } from './utils';

  const PAGE_SIZE = 50;
  const CONFIRM_TEXT = '确认';
  /** 后端识别为“确认执行 SQL”的关键词 */
  const CONFIRM_KEYWORDS = [
    '确认',
    '同意',
    '可以',
    'ok',
    '没问题',
    '好的',
    '是的',
    '对的',
    '采用',
    '使用',
    '行',
  ];

  const scenarios = ref<ChatScenario[]>([]);
  const sessions = ref<ChatSession[]>([]);
  const scenariosLoading = ref(false);
  const sessionsLoading = ref(false);

  const currentScene = ref('');
  /** 为空表示尚未创建会话的新对话，首次发送时再创建 */
  const activeCode = ref('');

  const messages = ref<ChatUiMessage[]>([]);
  const total = ref(0);
  const messagesLoading = ref(false);
  const loadingMore = ref(false);
  const sending = ref(false);
  const pendingContent = ref('');
  const draft = ref('');
  const dismissedKeys = ref<string[]>([]);
  const scrollRef = ref<HTMLElement>();

  const currentScenario = computed(() =>
    scenarios.value.find((s) => s.name === currentScene.value)
  );
  const activeSession = computed(() =>
    sessions.value.find((s) => s.session_code === activeCode.value)
  );

  const serverMessageCount = computed(
    () => messages.value.filter((m) => !m.status).length
  );
  const hasMore = computed(
    () => !!activeCode.value && serverMessageCount.value < total.value
  );

  /** 只有最后一条、且未被取消的待确认 AI 回复可以操作 */
  const actionableKey = computed(() => {
    const last = messages.value[messages.value.length - 1];
    if (
      last?.role === 'assistant' &&
      last.isNeedConfirm &&
      !dismissedKeys.value.includes(last.key)
    ) {
      return last.key;
    }
    return '';
  });

  const pendingText = computed(() =>
    CONFIRM_KEYWORDS.includes(pendingContent.value.trim().toLowerCase())
      ? '正在执行 SQL 查询，请稍候…'
      : 'AI 正在思考，请稍候…'
  );

  function scrollToBottom() {
    nextTick(() => {
      const el = scrollRef.value;
      if (el) el.scrollTop = el.scrollHeight;
    });
  }

  function fromHistory(m: ChatHistoryMessage): ChatUiMessage {
    return {
      key: `h_${m.id}`,
      role: m.role,
      content: m.content,
      time: dayjs(m.createdAt).valueOf(),
      isNeedConfirm: m.isNeedConfirm,
    };
  }

  function sortSessions(list: ChatSession[]) {
    return [...list].sort(
      (a, b) => (b.updated_at || b.created_at) - (a.updated_at || a.created_at)
    );
  }

  async function loadScenarios() {
    scenariosLoading.value = true;
    try {
      const { data } = await queryChatScenarios();
      scenarios.value = data.scenarios || [];
    } finally {
      scenariosLoading.value = false;
    }
  }

  async function loadSessions() {
    sessionsLoading.value = true;
    try {
      const { data } = await queryChatSessions();
      sessions.value = sortSessions(data.sessions || []);
    } finally {
      sessionsLoading.value = false;
    }
  }

  async function loadMessages() {
    const code = activeCode.value;
    if (!code) return;
    messagesLoading.value = true;
    try {
      const { data } = await queryChatMessages({
        session_code: code,
        limit: PAGE_SIZE,
        offset: 0,
        order: 'desc',
      });
      if (code !== activeCode.value) return;
      messages.value = (data.messages || []).map(fromHistory).reverse();
      total.value = data.total || 0;
      dismissedKeys.value = [];
      scrollToBottom();
    } finally {
      if (code === activeCode.value) messagesLoading.value = false;
    }
  }

  async function loadMore() {
    const code = activeCode.value;
    const el = scrollRef.value;
    if (!code || loadingMore.value) return;
    loadingMore.value = true;
    try {
      const { data } = await queryChatMessages({
        session_code: code,
        limit: PAGE_SIZE,
        offset: serverMessageCount.value,
        order: 'desc',
      });
      if (code !== activeCode.value) return;
      const prevHeight = el?.scrollHeight ?? 0;
      const known = new Set(messages.value.map((m) => m.key));
      const older = (data.messages || [])
        .map(fromHistory)
        .filter((m) => !known.has(m.key))
        .reverse();
      messages.value = [...older, ...messages.value];
      total.value = data.total || total.value;
      await nextTick();
      if (el) el.scrollTop += el.scrollHeight - prevHeight;
    } finally {
      loadingMore.value = false;
    }
  }

  function resetChat() {
    messages.value = [];
    total.value = 0;
    dismissedKeys.value = [];
    messagesLoading.value = false;
  }

  function openSession(session: ChatSession) {
    if (sending.value || session.session_code === activeCode.value) return;
    activeCode.value = session.session_code;
    currentScene.value = session.scene;
    resetChat();
    loadMessages();
  }

  function startDraft() {
    if (sending.value) return;
    activeCode.value = '';
    resetChat();
  }

  function selectScene(name: string) {
    if (sending.value || name === currentScene.value) return;
    currentScene.value = name;
    if (activeSession.value?.scene !== name) startDraft();
  }

  function dismissConfirm(key: string) {
    dismissedKeys.value.push(key);
  }

  async function ensureSession(firstMessage: string) {
    if (activeCode.value) return activeCode.value;
    const { data } = await createChatSession({
      scene: currentScene.value,
      title: firstMessage.replace(/\s+/g, ' ').slice(0, 30),
    });
    sessions.value = sortSessions([data, ...sessions.value]);
    activeCode.value = data.session_code;
    total.value = 0;
    return data.session_code;
  }

  function touchSession(code: string, time: number) {
    const target = sessions.value.find((s) => s.session_code === code);
    if (!target) return;
    target.updated_at = time;
    target.message_count = (target.message_count || 0) + 2;
    sessions.value = sortSessions(sessions.value);
  }

  async function send(text: string, retryOf?: ChatUiMessage) {
    const content = text.trim();
    if (!content || sending.value || !currentScene.value) return;

    sending.value = true;
    pendingContent.value = content;
    const fromInput = !retryOf && text === draft.value;
    if (fromInput) draft.value = '';

    let code: string;
    try {
      code = await ensureSession(content);
    } catch {
      if (fromInput) draft.value = text;
      sending.value = false;
      return;
    }

    const userMsg: ChatUiMessage =
      retryOf ||
      reactive<ChatUiMessage>({
        key: `local_${Date.now()}`,
        role: 'user',
        content,
        time: Date.now(),
        isNeedConfirm: false,
        requestId: createRequestId(),
      });
    userMsg.status = 'sending';
    if (!retryOf) messages.value.push(userMsg);
    scrollToBottom();

    try {
      const { data } = await sendChatMessage({
        session_code: code,
        content,
        content_type: 'text',
        request_id: userMsg.requestId as string,
      });
      if (code !== activeCode.value) return;
      userMsg.status = undefined;
      messages.value.push({
        key: `s_${data.id}`,
        role: data.role,
        content: data.content,
        contentType: data.content_type,
        time: data.created_at || Date.now(),
        isNeedConfirm: data.is_need_confirm,
        latencyMs: data.latency_ms,
        modelName: data.model_name,
        totalTokens: data.total_tokens,
      });
      total.value += 2;
      touchSession(code, data.created_at || Date.now());
    } catch {
      userMsg.status = 'failed';
    } finally {
      sending.value = false;
      scrollToBottom();
    }
  }

  function retry(msg: ChatUiMessage) {
    // 失败消息需保持在末尾，后续回复才能接在它后面
    messages.value = [...messages.value.filter((m) => m !== msg), msg];
    send(msg.content, msg);
  }

  function onKeydown(e: KeyboardEvent) {
    if (e.key !== 'Enter' || e.shiftKey || e.isComposing) return;
    e.preventDefault();
    send(draft.value);
  }

  onMounted(async () => {
    await Promise.allSettled([loadScenarios(), loadSessions()]);
    const [first] = sessions.value;
    if (first) {
      openSession(first);
    } else if (scenarios.value.length) {
      currentScene.value = scenarios.value[0].name;
    }
  });
</script>

<style scoped lang="less">
  .container {
    padding: 0 20px 20px;
  }

  // 导航栏 60px + 页签栏 40px + 内容区上边距 12px + 页脚 30px
  @layout-offset: 142px;
  @page-min-height: 600px;

  .chat-page {
    display: grid;
    grid-template-columns: 320px minmax(0, 1fr);
    gap: 16px;
    box-sizing: border-box;
    height: ~'max(@{page-min-height}, calc(100vh - @{layout-offset}))';
    padding-top: 16px;

    &__sidebar {
      min-height: 0;
    }
  }

  .chat {
    display: flex;
    flex-direction: column;
    min-width: 0;
    min-height: 0;
    overflow: hidden;
    background: var(--color-bg-2);
    border-radius: 8px;

    &__header {
      display: flex;
      flex: none;
      gap: 12px;
      align-items: center;
      padding: 14px 20px;
      border-bottom: 1px solid var(--color-border-1);
    }

    &__header-icon {
      display: flex;
      flex: none;
      align-items: center;
      justify-content: center;
      width: 40px;
      height: 40px;
      color: rgb(var(--primary-6));
      font-size: 20px;
      background: rgb(var(--primary-1));
      border-radius: 10px;
    }

    &__header-body {
      flex: 1;
      min-width: 0;
    }

    &__title {
      display: flex;
      gap: 8px;
      align-items: center;
      overflow: hidden;
      color: var(--color-text-1);
      font-weight: 600;
      font-size: 16px;
      white-space: nowrap;
    }

    &__subtitle {
      margin-top: 2px;
      overflow: hidden;
      color: var(--color-text-3);
      font-size: 12px;
      white-space: nowrap;
      text-overflow: ellipsis;
    }

    &__messages {
      flex: 1;
      min-height: 0;
      padding: 20px 24px;
      overflow-y: auto;
      background: var(--color-fill-1);
    }

    &__spin {
      display: block;
      min-height: 100%;
    }

    &__more {
      margin-bottom: 12px;
      text-align: center;
    }

    &__input {
      display: flex;
      flex: none;
      gap: 12px;
      align-items: flex-end;
      padding: 14px 20px;
      border-top: 1px solid var(--color-border-1);
    }

    &__textarea {
      flex: 1;
    }
  }

  .welcome {
    display: flex;
    flex-direction: column;
    gap: 8px;
    align-items: center;
    padding: 80px 24px 24px;
    text-align: center;

    &__icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 64px;
      height: 64px;
      margin-bottom: 8px;
      color: rgb(var(--primary-6));
      font-size: 32px;
      background: rgb(var(--primary-1));
      border-radius: 16px;
    }

    &__title {
      color: var(--color-text-1);
      font-weight: 600;
      font-size: 18px;
    }

    &__desc {
      max-width: 480px;
      color: var(--color-text-3);
    }

    &__tools {
      display: flex;
      flex-wrap: wrap;
      gap: 6px;
      align-items: center;
      justify-content: center;
      margin-top: 4px;
      color: var(--color-text-3);
      font-size: 12px;
    }
  }

  .pending {
    display: flex;
    gap: 12px;
    align-items: center;
    margin-top: 20px;

    &__avatar {
      flex: none;
      color: rgb(var(--primary-6));
      background: rgb(var(--primary-1));
    }

    &__bubble {
      display: inline-flex;
      gap: 8px;
      align-items: center;
      padding: 10px 16px;
      color: var(--color-text-2);
      background: var(--color-bg-2);
      border: 1px solid var(--color-border-2);
      border-radius: 12px;
      border-top-left-radius: 4px;
    }
  }

  @media (max-width: 1100px) {
    .chat-page {
      grid-template-columns: 260px minmax(0, 1fr);
    }
  }
</style>
