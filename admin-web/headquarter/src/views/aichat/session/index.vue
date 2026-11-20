<template>
  <div class="aichat-page">
    <aside class="chat-sider">
      <section class="sider-block">
        <div class="block-title">
          选择场景
          <a-tooltip content="新建对话时使用当前选中的场景">
            <icon-question-circle class="block-title-tip" />
          </a-tooltip>
        </div>
        <a-spin :loading="scenariosLoading" class="scene-spin">
          <div v-if="scenarios.length" class="scene-grid">
            <a-tooltip
              v-for="scene in scenarios"
              :key="scene.name"
              :content="scene.description"
              position="bottom"
              mini
            >
              <div
                class="scene-card"
                :class="{ active: scene.name === activeScene }"
                @click="activeScene = scene.name"
              >
                <component
                  :is="sceneIcon(scene.icon)"
                  class="scene-card-icon"
                />
                <span class="scene-card-name">{{ scene.name }}</span>
              </div>
            </a-tooltip>
          </div>
          <div v-else-if="!scenariosLoading" class="scene-empty"
            >暂无可用场景</div
          >
        </a-spin>
      </section>

      <section class="sider-block session-block">
        <div class="block-title">会话列表</div>
        <div class="session-toolbar">
          <a-input
            v-model="sessionKeyword"
            placeholder="搜索会话标题..."
            allow-clear
          >
            <template #prefix><icon-search /></template>
          </a-input>
          <a-button
            type="primary"
            :disabled="!activeScene"
            @click="openCreateModal"
          >
            <template #icon><icon-plus /></template>
            新建对话
          </a-button>
        </div>
        <a-spin :loading="sessionsLoading" class="session-spin">
          <div class="session-list">
            <div
              v-for="session in filteredSessions"
              :key="session.session_code"
              class="session-item"
              :class="{ active: session.session_code === activeSessionCode }"
              @click="selectSession(session)"
            >
              <component
                :is="sceneIcon(sceneOf(session.scene)?.icon)"
                class="session-item-icon"
              />
              <div class="session-item-body">
                <div class="session-item-row">
                  <span class="session-item-title" :title="session.title">
                    {{ session.title || session.scene }}
                  </span>
                  <span class="session-item-time">
                    {{ formatSessionTime(session.updated_at) }}
                  </span>
                </div>
                <div
                  v-if="session.context_summary"
                  class="session-item-summary"
                  :title="session.context_summary"
                >
                  {{ session.context_summary }}
                </div>
                <div class="session-item-row">
                  <a-space :size="4">
                    <a-tag size="small" :color="sceneColor(session.scene)">
                      {{ session.scene }}
                    </a-tag>
                    <a-tag
                      v-if="session.status === 'CLOSED'"
                      size="small"
                      color="gray"
                    >
                      已关闭
                    </a-tag>
                  </a-space>
                  <span class="session-item-count">
                    <icon-message />
                    {{ session.message_count ?? 0 }}
                  </span>
                </div>
              </div>
            </div>
            <div
              v-if="!sessionsLoading && !filteredSessions.length"
              class="session-empty"
            >
              {{ sessionKeyword ? '没有匹配的会话' : '暂无会话' }}
            </div>
          </div>
        </a-spin>
        <div class="new-chat-card">
          <div class="new-chat-card-bars">
            <span
              v-for="h in [18, 30, 24, 36, 22]"
              :key="h"
              :style="{ height: `${h}px` }"
            />
          </div>
          <div class="new-chat-card-title">开始新的对话吧</div>
          <div class="new-chat-card-desc">选择一个场景，开启智能对话之旅</div>
          <a-button
            type="primary"
            size="small"
            :disabled="!activeScene"
            @click="openCreateModal"
          >
            新建对话
          </a-button>
        </div>
      </section>
    </aside>

    <main class="chat-main">
      <header class="chat-header">
        <div class="chat-header-icon">
          <component :is="sceneIcon(currentScenario?.icon)" />
        </div>
        <div class="chat-header-info">
          <div class="chat-header-title">
            {{
              activeSession
                ? activeSession.title || activeSession.scene
                : '新对话'
            }}
          </div>
          <div class="chat-header-desc">
            <template v-if="activeSession">
              {{
                activeSession.context_summary || `场景：${activeSession.scene}`
              }}
            </template>
            <template v-else-if="currentScenario">
              {{ currentScenario.name }} · {{ currentScenario.description }}
            </template>
            <template v-else>请选择一个场景开始对话</template>
          </div>
        </div>
        <a-space>
          <a-tag v-if="currentScenario?.model_name" color="arcoblue">
            {{ currentScenario.model_name }}
          </a-tag>
          <a-button
            v-if="activeSession"
            size="small"
            :loading="messagesLoading"
            @click="reloadMessages"
          >
            <template #icon><icon-refresh /></template>
            刷新
          </a-button>
          <a-button v-if="activeSession" size="small" @click="startDraft">
            <template #icon><icon-plus /></template>
            新对话
          </a-button>
        </a-space>
      </header>

      <div ref="messagesRef" class="chat-messages">
        <a-spin
          :loading="messagesLoading && !messages.length"
          class="messages-spin"
        >
          <div v-if="hasMoreMessages" class="load-more">
            <a-button
              type="text"
              size="small"
              :loading="loadingMore"
              @click="loadMoreMessages"
            >
              加载更早的消息
            </a-button>
          </div>

          <div v-if="!messages.length && !messagesLoading" class="chat-welcome">
            <div class="chat-welcome-avatar"><icon-robot /></div>
            <div class="chat-welcome-title">你好，我是 AI 助手</div>
            <div class="chat-welcome-desc">
              {{
                currentScenario
                  ? `当前场景：${currentScenario.name}，${currentScenario.description}`
                  : '请先在左侧选择一个场景'
              }}
            </div>
            <a-space v-if="currentScenario?.available_tools?.length" wrap>
              <a-tag
                v-for="tool in currentScenario.available_tools"
                :key="tool"
              >
                {{ tool }}
              </a-tag>
            </a-space>
          </div>

          <div
            v-for="(msg, index) in messages"
            :key="msg.key"
            class="message-row"
            :class="msg.role === 'user' ? 'is-user' : 'is-assistant'"
          >
            <a-avatar
              v-if="msg.role === 'user'"
              :size="36"
              class="message-avatar user-avatar"
            >
              <img
                v-if="userStore.avatar"
                :src="userStore.avatar"
                alt="avatar"
              />
              <icon-user v-else />
            </a-avatar>
            <div v-else class="message-avatar bot-avatar">
              <icon-robot />
            </div>

            <div class="message-main">
              <div
                class="message-bubble"
                :class="{ failed: msg.status === 'failed' }"
              >
                <a-tag v-if="msg.role === 'tool'" size="small" class="tool-tag">
                  工具结果
                </a-tag>
                <template v-if="msg.role === 'user'">{{
                  msg.content
                }}</template>
                <message-content v-else :content="msg.content" />

                <div v-if="isConfirmVisible(msg, index)" class="confirm-box">
                  <div class="confirm-box-title">
                    <icon-exclamation-circle-fill />
                    包含可执行的 SQL 语句，是否确认执行？
                  </div>
                  <a-space>
                    <a-button type="primary" size="small" @click="confirmSql">
                      <template #icon><icon-check /></template>
                      确认
                    </a-button>
                    <a-button
                      size="small"
                      @click="dismissedConfirm.add(msg.key)"
                    >
                      取消
                    </a-button>
                  </a-space>
                </div>
              </div>
              <div class="message-meta">
                <template v-if="msg.status === 'sending'">
                  <icon-loading /> 发送中
                </template>
                <template v-else-if="msg.status === 'error'">
                  <span class="meta-error">发送失败</span>
                  <a-link :disabled="sending" @click="retry(msg)">
                    重试
                  </a-link>
                </template>
                <template v-else>
                  <span v-if="msg.status === 'failed'" class="meta-error"
                    >处理失败</span
                  >
                  <span v-if="msg.status === 'cancelled'">已取消</span>
                  <span>{{ formatMessageTime(msg.time) }}</span>
                  <a-tooltip
                    v-if="msg.latencyMs !== undefined"
                    :content="`模型：${msg.modelName || '-'}，Token：${
                      msg.totalTokens ?? '-'
                    }`"
                  >
                    <span>耗时 {{ (msg.latencyMs / 1000).toFixed(1) }}s</span>
                  </a-tooltip>
                </template>
              </div>
            </div>
          </div>

          <div v-if="sending" class="message-row is-assistant">
            <div class="message-avatar bot-avatar"><icon-robot /></div>
            <div class="message-main">
              <div class="message-bubble pending">
                <icon-loading />
                {{ pendingText }}
              </div>
            </div>
          </div>
        </a-spin>
      </div>

      <footer class="chat-input">
        <div class="chat-input-box" :class="{ disabled: inputDisabled }">
          <a-textarea
            v-model="inputText"
            :placeholder="inputPlaceholder"
            :auto-size="{ minRows: 1, maxRows: 6 }"
            :disabled="inputDisabled"
            class="chat-textarea"
            @keydown="handleInputKeydown"
          />
          <a-button
            type="primary"
            :loading="sending"
            :disabled="inputDisabled || !inputText.trim()"
            @click="handleSend"
          >
            <template #icon><icon-send /></template>
            发送
          </a-button>
        </div>
        <div class="chat-input-tip">Enter 发送，Shift + Enter 换行</div>
      </footer>
    </main>

    <a-modal
      v-model:visible="createVisible"
      title="新建对话"
      :ok-loading="creating"
      @before-ok="handleCreate"
    >
      <a-form ref="createFormRef" :model="createForm" layout="vertical">
        <a-form-item
          field="scene"
          label="场景"
          :rules="[{ required: true, message: '请选择场景' }]"
        >
          <a-select v-model="createForm.scene" placeholder="请选择场景">
            <a-option
              v-for="scene in scenarios"
              :key="scene.name"
              :value="scene.name"
              :label="scene.name"
            />
          </a-select>
        </a-form-item>
        <a-form-item field="title" label="会话标题">
          <a-input
            v-model="createForm.title"
            placeholder="不填默认使用场景名称"
            :max-length="50"
            allow-clear
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script lang="ts" setup>
  import { computed, nextTick, onMounted, reactive, ref } from 'vue';
  import dayjs from 'dayjs';
  import type { FormInstance } from '@arco-design/web-vue';
  import { useUserStore } from '@/store';
  import {
    createChatSession,
    queryChatMessages,
    queryChatScenarios,
    queryChatSessions,
    sendChatMessage,
    type ChatHistoryMessage,
    type ChatMessageStatus,
    type ChatRole,
    type ChatScenario,
    type ChatSession,
    type SendChatMessageRes,
  } from '@/api/aichat';
  import MessageContent from './components/message-content.vue';

  interface ChatMessage {
    key: string;
    role: ChatRole;
    content: string;
    /** 毫秒时间戳 */
    time: number;
    isNeedConfirm: boolean;
    /** sending / error 为前端本地发送状态，其余为后端消息处理状态 */
    status?: ChatMessageStatus | 'sending' | 'error';
    /** 发送失败重试时复用，保证幂等 */
    requestId?: string;
    latencyMs?: number;
    totalTokens?: number;
    modelName?: string;
  }

  const PAGE_SIZE = 50;
  const SCENE_ICONS: Record<string, string> = {
    chart: 'icon-bar-chart',
    data: 'icon-bar-chart',
    service: 'icon-customer-service',
    knowledge: 'icon-book',
    book: 'icon-book',
    general: 'icon-robot',
  };
  const SCENE_COLORS = [
    'arcoblue',
    'cyan',
    'green',
    'orange',
    'purple',
    'magenta',
  ];
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

  const userStore = useUserStore();

  const scenarios = ref<ChatScenario[]>([]);
  const scenariosLoading = ref(false);
  const activeScene = ref('');

  const sessions = ref<ChatSession[]>([]);
  const sessionsLoading = ref(false);
  const sessionKeyword = ref('');
  const activeSessionCode = ref('');

  const messages = ref<ChatMessage[]>([]);
  const messagesTotal = ref(0);
  /** 已在服务端落库并加载到列表中的消息数，用作分页 offset */
  const loadedCount = ref(0);
  const messagesLoading = ref(false);
  const loadingMore = ref(false);
  const messagesRef = ref<HTMLElement>();

  const inputText = ref('');
  const sending = ref(false);
  const pendingText = ref('');
  const dismissedConfirm = reactive(new Set<string>());

  const createVisible = ref(false);
  const creating = ref(false);
  const createFormRef = ref<FormInstance>();
  const createForm = reactive({ scene: '', title: '' });

  const sceneOf = (name?: string) =>
    scenarios.value.find((s) => s.name === name);
  const sceneIcon = (icon?: string) =>
    (icon && SCENE_ICONS[icon]) || 'icon-robot';
  const sceneColor = (name: string) => {
    const index = scenarios.value.findIndex((s) => s.name === name);
    return SCENE_COLORS[(index < 0 ? 0 : index) % SCENE_COLORS.length];
  };

  const activeSession = computed(() =>
    sessions.value.find((s) => s.session_code === activeSessionCode.value)
  );
  const currentScenario = computed(() =>
    sceneOf(activeSession.value?.scene ?? activeScene.value)
  );
  const filteredSessions = computed(() => {
    const keyword = sessionKeyword.value.trim().toLowerCase();
    if (!keyword) return sessions.value;
    return sessions.value.filter((s) =>
      `${s.title} ${s.context_summary ?? ''}`.toLowerCase().includes(keyword)
    );
  });
  const hasMoreMessages = computed(
    () => !!activeSessionCode.value && loadedCount.value < messagesTotal.value
  );
  const inputDisabled = computed(
    () =>
      activeSession.value?.status === 'CLOSED' ||
      (!activeSession.value && !activeScene.value)
  );
  const inputPlaceholder = computed(() => {
    if (activeSession.value?.status === 'CLOSED')
      return '该会话已关闭，请新建对话';
    if (!activeSession.value && !activeScene.value) return '请先选择一个场景';
    return '请输入您的问题...';
  });

  const formatSessionTime = (ts: number) => {
    if (!ts) return '';
    const time = dayjs(ts);
    if (time.isSame(dayjs(), 'day')) return time.format('HH:mm');
    if (time.isSame(dayjs(), 'year')) return time.format('MM/DD');
    return time.format('YYYY/MM/DD');
  };
  const formatMessageTime = (ts: number) => {
    if (!ts) return '';
    const time = dayjs(ts);
    return time.isSame(dayjs(), 'day')
      ? time.format('HH:mm')
      : time.format('MM-DD HH:mm');
  };

  const createRequestId = () => {
    if (typeof crypto.randomUUID === 'function') return crypto.randomUUID();
    // 非安全上下文（http + 局域网 IP）下没有 randomUUID，手动拼 UUID v4
    const bytes = crypto.getRandomValues(new Uint8Array(16));
    bytes[6] = (bytes[6] % 16) + 0x40;
    bytes[8] = (bytes[8] % 64) + 0x80;
    const hex = Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join(
      ''
    );
    return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(
      12,
      16
    )}-${hex.slice(16, 20)}-${hex.slice(20)}`;
  };

  const fromHistory = (msg: ChatHistoryMessage): ChatMessage => ({
    key: msg.id,
    role: msg.role,
    content: msg.content,
    time: dayjs(msg.createdAt).valueOf(),
    isNeedConfirm: msg.isNeedConfirm,
  });
  const fromSendResult = (msg: SendChatMessageRes): ChatMessage => ({
    key: String(msg.id),
    role: msg.role,
    content: msg.content,
    time: msg.created_at,
    isNeedConfirm: msg.is_need_confirm,
    status: msg.status,
    latencyMs: msg.latency_ms,
    totalTokens: msg.total_tokens,
    modelName: msg.model_name,
  });

  const scrollToBottom = async () => {
    await nextTick();
    const el = messagesRef.value;
    if (el) el.scrollTop = el.scrollHeight;
  };

  const isConfirmVisible = (msg: ChatMessage, index: number) =>
    msg.role !== 'user' &&
    msg.isNeedConfirm &&
    index === messages.value.length - 1 &&
    !sending.value &&
    !dismissedConfirm.has(msg.key);

  const fetchScenarios = async () => {
    scenariosLoading.value = true;
    try {
      const { data } = await queryChatScenarios();
      scenarios.value = data?.scenarios ?? [];
      if (!activeScene.value && scenarios.value.length) {
        activeScene.value = scenarios.value[0].name;
      }
    } finally {
      scenariosLoading.value = false;
    }
  };

  const fetchSessions = async () => {
    sessionsLoading.value = true;
    try {
      const { data } = await queryChatSessions();
      sessions.value = [...(data?.sessions ?? [])].sort(
        (a, b) => b.updated_at - a.updated_at
      );
    } finally {
      sessionsLoading.value = false;
    }
  };

  const fetchMessages = async (sessionCode: string, append = false) => {
    const { data } = await queryChatMessages({
      session_code: sessionCode,
      limit: PAGE_SIZE,
      offset: append ? loadedCount.value : 0,
      order: 'desc',
    });
    if (sessionCode !== activeSessionCode.value) return;
    const list = (data?.messages ?? []).map(fromHistory).reverse();
    messages.value = append ? [...list, ...messages.value] : list;
    loadedCount.value = (append ? loadedCount.value : 0) + list.length;
    messagesTotal.value = data?.total ?? loadedCount.value;
  };

  const reloadMessages = async () => {
    const code = activeSessionCode.value;
    if (!code) return;
    messagesLoading.value = true;
    try {
      await fetchMessages(code);
      scrollToBottom();
    } finally {
      if (code === activeSessionCode.value) messagesLoading.value = false;
    }
  };

  const loadMoreMessages = async () => {
    const code = activeSessionCode.value;
    const el = messagesRef.value;
    const prevHeight = el?.scrollHeight ?? 0;
    loadingMore.value = true;
    try {
      await fetchMessages(code, true);
      await nextTick();
      if (el) el.scrollTop = el.scrollHeight - prevHeight;
    } finally {
      loadingMore.value = false;
    }
  };

  const resetMessages = () => {
    messages.value = [];
    messagesTotal.value = 0;
    loadedCount.value = 0;
    dismissedConfirm.clear();
  };

  const selectSession = (session: ChatSession) => {
    if (sending.value || session.session_code === activeSessionCode.value)
      return;
    activeSessionCode.value = session.session_code;
    activeScene.value = session.scene;
    resetMessages();
    reloadMessages();
  };

  const startDraft = () => {
    if (sending.value) return;
    activeSessionCode.value = '';
    resetMessages();
  };

  const touchSession = (code: string, addCount: number) => {
    const index = sessions.value.findIndex((s) => s.session_code === code);
    if (index < 0) return;
    const [session] = sessions.value.splice(index, 1);
    session.message_count = (session.message_count ?? 0) + addCount;
    session.updated_at = Date.now();
    sessions.value.unshift(session);
  };

  const createSession = async (scene: string, title?: string) => {
    const { data } = await createChatSession({
      scene,
      title: title || undefined,
    });
    sessions.value.unshift({ message_count: 0, ...data });
    activeSessionCode.value = data.session_code;
    activeScene.value = data.scene;
    resetMessages();
    return data;
  };

  const openCreateModal = () => {
    createForm.scene = activeScene.value;
    createForm.title = '';
    createVisible.value = true;
  };

  const handleCreate = async (done: (closed: boolean) => void) => {
    const errors = await createFormRef.value?.validate();
    if (errors) {
      done(false);
      return;
    }
    creating.value = true;
    try {
      await createSession(createForm.scene, createForm.title.trim());
      done(true);
    } catch {
      done(false);
    } finally {
      creating.value = false;
    }
  };

  const deliver = async (userMsg: ChatMessage) => {
    sending.value = true;
    const isConfirm = CONFIRM_KEYWORDS.includes(
      userMsg.content.trim().toLowerCase()
    );
    pendingText.value = isConfirm
      ? '正在执行 SQL 查询，请稍候...'
      : '正在思考中，请稍候...';
    userMsg.status = 'sending';
    scrollToBottom();
    try {
      let code = activeSessionCode.value;
      if (!code) {
        const session = await createSession(
          activeScene.value,
          userMsg.content.slice(0, 20)
        );
        code = session.session_code;
        messages.value.push(userMsg);
      }
      const { data } = await sendChatMessage({
        session_code: code,
        content: userMsg.content,
        content_type: 'text',
        request_id: userMsg.requestId as string,
      });
      userMsg.status = undefined;
      if (code === activeSessionCode.value) {
        messages.value.push(fromSendResult(data));
        loadedCount.value += 2;
        messagesTotal.value += 2;
      }
      touchSession(code, 2);
    } catch {
      userMsg.status = 'error';
      // 自动建会话失败时消息还没进列表，把内容还给输入框
      if (!messages.value.includes(userMsg)) inputText.value = userMsg.content;
    } finally {
      sending.value = false;
      scrollToBottom();
    }
  };

  const send = (content: string) => {
    const text = content.trim();
    if (!text || sending.value || inputDisabled.value) return;
    const userMsg = reactive<ChatMessage>({
      key: `local_${Date.now()}`,
      role: 'user',
      content: text,
      time: Date.now(),
      isNeedConfirm: false,
      requestId: createRequestId(),
    });
    if (activeSessionCode.value) messages.value.push(userMsg);
    deliver(userMsg);
  };

  const handleSend = () => {
    const text = inputText.value;
    if (!text.trim() || sending.value) return;
    inputText.value = '';
    send(text);
  };

  const handleInputKeydown = (e: KeyboardEvent) => {
    if (e.key !== 'Enter' || e.shiftKey || e.isComposing) return;
    e.preventDefault();
    handleSend();
  };

  const confirmSql = () => send('确认');

  const retry = (msg: ChatMessage) => {
    if (sending.value) return;
    deliver(msg);
  };

  onMounted(() => {
    fetchScenarios();
    fetchSessions();
  });
</script>

<style scoped lang="less">
  .aichat-page {
    display: flex;
    gap: 16px;
    height: calc(100vh - 60px - 40px - 32px);
    min-height: 560px;
    padding: 16px 20px;
  }

  .chat-sider {
    display: flex;
    flex: 0 0 300px;
    flex-direction: column;
    gap: 16px;
    min-height: 0;
    padding: 16px;
    background: var(--color-bg-2);
    border-radius: 8px;
  }

  .block-title {
    display: flex;
    gap: 6px;
    align-items: center;
    margin-bottom: 12px;
    color: var(--color-text-1);
    font-weight: 600;
    font-size: 15px;

    &-tip {
      color: var(--color-text-3);
      font-size: 14px;
      cursor: help;
    }
  }

  .scene-spin {
    display: block;
    min-height: 72px;
  }

  .scene-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 8px;
  }

  .scene-card {
    display: flex;
    flex-direction: column;
    gap: 6px;
    align-items: center;
    justify-content: center;
    min-height: 72px;
    padding: 8px 4px;
    color: var(--color-text-2);
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s;

    &:hover {
      border-color: rgb(var(--arcoblue-4));
    }

    &.active {
      color: rgb(var(--arcoblue-6));
      background: rgb(var(--arcoblue-1));
      border-color: rgb(var(--arcoblue-6));
    }

    &-icon {
      color: rgb(var(--arcoblue-6));
      font-size: 22px;
    }

    &-name {
      max-width: 100%;
      overflow: hidden;
      font-size: 12px;
      white-space: nowrap;
      text-overflow: ellipsis;
    }
  }

  .scene-empty,
  .session-empty {
    padding: 24px 0;
    color: var(--color-text-3);
    font-size: 13px;
    text-align: center;
  }

  .session-block {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
  }

  .session-toolbar {
    display: flex;
    gap: 8px;
    margin-bottom: 12px;
  }

  .session-spin {
    display: block;
    flex: 1;
    min-height: 0;
  }

  .session-list {
    height: 100%;
    overflow-y: auto;
  }

  .session-item {
    display: flex;
    gap: 10px;
    padding: 12px;
    border-bottom: 1px solid var(--color-border-1);
    border-left: 3px solid transparent;
    border-radius: 4px;
    cursor: pointer;
    transition: background 0.2s;

    &:hover {
      background: var(--color-fill-1);
    }

    &.active {
      background: rgb(var(--arcoblue-1));
      border-left-color: rgb(var(--arcoblue-6));
    }

    &-icon {
      flex-shrink: 0;
      margin-top: 2px;
      color: rgb(var(--arcoblue-6));
      font-size: 18px;
    }

    &-body {
      display: flex;
      flex: 1;
      flex-direction: column;
      gap: 4px;
      min-width: 0;
    }

    &-row {
      display: flex;
      gap: 8px;
      align-items: center;
      justify-content: space-between;
    }

    &-title {
      overflow: hidden;
      color: var(--color-text-1);
      font-weight: 500;
      white-space: nowrap;
      text-overflow: ellipsis;
    }

    &-time,
    &-count {
      flex-shrink: 0;
      color: var(--color-text-3);
      font-size: 12px;
    }

    &-summary {
      overflow: hidden;
      color: var(--color-text-3);
      font-size: 12px;
      white-space: nowrap;
      text-overflow: ellipsis;
    }
  }

  .new-chat-card {
    display: flex;
    flex-direction: column;
    gap: 6px;
    align-items: center;
    margin-top: 12px;
    padding: 16px;
    background: linear-gradient(
      180deg,
      rgb(var(--arcoblue-1)) 0%,
      var(--color-bg-2) 100%
    );
    border: 1px solid var(--color-border-1);
    border-radius: 8px;

    &-bars {
      display: flex;
      gap: 4px;
      align-items: flex-end;
      height: 36px;

      span {
        width: 8px;
        background: rgb(var(--arcoblue-3));
        border-radius: 2px 2px 0 0;
      }
    }

    &-title {
      color: var(--color-text-1);
      font-weight: 500;
    }

    &-desc {
      margin-bottom: 4px;
      color: var(--color-text-3);
      font-size: 12px;
    }
  }

  .chat-main {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-width: 0;
    background: var(--color-bg-2);
    border-radius: 8px;
  }

  .chat-header {
    display: flex;
    gap: 12px;
    align-items: center;
    padding: 14px 20px;
    border-bottom: 1px solid var(--color-border-1);

    &-icon {
      display: flex;
      flex-shrink: 0;
      align-items: center;
      justify-content: center;
      width: 40px;
      height: 40px;
      color: rgb(var(--arcoblue-6));
      font-size: 20px;
      background: rgb(var(--arcoblue-1));
      border-radius: 8px;
    }

    &-info {
      flex: 1;
      min-width: 0;
    }

    &-title {
      overflow: hidden;
      color: var(--color-text-1);
      font-weight: 600;
      font-size: 16px;
      white-space: nowrap;
      text-overflow: ellipsis;
    }

    &-desc {
      overflow: hidden;
      color: var(--color-text-3);
      font-size: 12px;
      white-space: nowrap;
      text-overflow: ellipsis;
    }
  }

  .chat-messages {
    flex: 1;
    min-height: 0;
    padding: 20px;
    overflow-y: auto;
    background: var(--color-fill-1);
  }

  .messages-spin {
    display: block;
    min-height: 100%;
  }

  .load-more {
    margin-bottom: 12px;
    text-align: center;
  }

  .chat-welcome {
    display: flex;
    flex-direction: column;
    gap: 10px;
    align-items: center;
    padding: 80px 20px;
    text-align: center;

    &-avatar {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 64px;
      height: 64px;
      color: #fff;
      font-size: 32px;
      background: linear-gradient(
        135deg,
        rgb(var(--arcoblue-5)),
        rgb(var(--arcoblue-7))
      );
      border-radius: 50%;
    }

    &-title {
      color: var(--color-text-1);
      font-weight: 600;
      font-size: 18px;
    }

    &-desc {
      max-width: 480px;
      color: var(--color-text-3);
    }
  }

  .message-row {
    display: flex;
    gap: 12px;
    margin-bottom: 20px;

    &.is-user {
      flex-direction: row-reverse;

      .message-main {
        align-items: flex-end;
      }

      .message-bubble {
        color: var(--color-text-1);
        white-space: pre-wrap;
        background: rgb(var(--arcoblue-1));
        border-color: rgb(var(--arcoblue-2));
      }
    }
  }

  .message-avatar {
    flex-shrink: 0;
  }

  .user-avatar {
    background: rgb(var(--arcoblue-6));
  }

  .bot-avatar {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    color: #fff;
    font-size: 20px;
    background: linear-gradient(
      135deg,
      rgb(var(--arcoblue-5)),
      rgb(var(--arcoblue-7))
    );
    border-radius: 50%;
  }

  .message-main {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0;
    max-width: min(78%, 860px);
  }

  .message-bubble {
    padding: 10px 14px;
    line-height: 1.75;
    word-break: break-word;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-1);
    border-radius: 8px;

    &.failed {
      border-color: rgb(var(--red-3));
    }

    &.pending {
      display: flex;
      gap: 8px;
      align-items: center;
      color: var(--color-text-2);
    }
  }

  .tool-tag {
    margin-bottom: 6px;
  }

  .confirm-box {
    margin-top: 12px;
    padding: 12px;
    background: rgb(var(--orange-1));
    border: 1px solid rgb(var(--orange-2));
    border-radius: 6px;

    &-title {
      display: flex;
      gap: 6px;
      align-items: center;
      margin-bottom: 10px;
      color: var(--color-text-1);
      font-weight: 500;

      .arco-icon {
        color: rgb(var(--orange-6));
        font-size: 16px;
      }
    }
  }

  .message-meta {
    display: flex;
    gap: 8px;
    align-items: center;
    color: var(--color-text-4);
    font-size: 12px;
  }

  .meta-error {
    color: rgb(var(--red-6));
  }

  .chat-input {
    padding: 12px 20px 10px;
    border-top: 1px solid var(--color-border-1);

    &-box {
      display: flex;
      gap: 12px;
      align-items: flex-end;
      padding: 8px 8px 8px 4px;
      border: 1px solid var(--color-border-2);
      border-radius: 8px;
      transition: border-color 0.2s;

      &:focus-within {
        border-color: rgb(var(--arcoblue-6));
      }

      &.disabled {
        background: var(--color-fill-1);
      }
    }

    &-tip {
      margin-top: 6px;
      color: var(--color-text-4);
      font-size: 12px;
    }
  }

  .chat-textarea {
    flex: 1;
    background: transparent;
    border: none;

    :deep(.arco-textarea) {
      padding: 4px 8px;
    }

    &:focus-within,
    &:hover {
      background: transparent;
      border: none;
      box-shadow: none;
    }
  }
</style>
