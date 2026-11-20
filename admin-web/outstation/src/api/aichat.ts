import axios from 'axios';

export interface ChatScenario {
  /** 场景名称，唯一标识 */
  name: string;
  description: string;
  icon: string;
  model_name: string;
  /** 0.0-2.0 */
  temperature: number;
  available_tools: string[];
  max_recent_messages: number;
}

export type ChatSessionStatus = 'ACTIVE' | 'CLOSED';

export interface ChatSession {
  id: number;
  session_code: string;
  user_id: number;
  op_code: string;
  title: string;
  status: ChatSessionStatus;
  /** 场景名称 */
  scene: string;
  context_summary?: string;
  message_count?: number;
  /** 毫秒时间戳 */
  created_at: number;
  /** 毫秒时间戳 */
  updated_at: number;
}

export type ChatRole = 'user' | 'assistant' | 'tool';

export interface ChatHistoryMessage {
  id: string;
  sessionCode: string;
  role: ChatRole;
  content: string;
  /** ISO 8601 */
  createdAt: string;
  isNeedConfirm: boolean;
}

export type ChatContentType = 'text' | 'json' | 'markdown';

export type ChatMessageStatus =
  | 'processing'
  | 'completed'
  | 'failed'
  | 'cancelled';

export interface ChatSentMessage {
  id: number;
  session_id: number;
  sequence: number;
  is_need_confirm: boolean;
  role: ChatRole;
  content: string;
  content_type: ChatContentType;
  request_id: string;
  status: ChatMessageStatus;
  model_name: string;
  input_tokens: number;
  output_tokens: number;
  total_tokens: number;
  cost_usd: number;
  latency_ms: number;
  tool_calls: unknown[] | null;
  tool_results: unknown[] | null;
  /** 毫秒时间戳 */
  created_at: number;
}

/** 场景列表 GET /chat/scenarios */
export function queryChatScenarios() {
  return axios.get<{ scenarios: ChatScenario[] }>('/chat/scenarios');
}

/** 当前用户会话列表 GET /chat/sessions */
export function queryChatSessions() {
  return axios.get<{ sessions: ChatSession[] }>('/chat/sessions');
}

export interface CreateChatSessionParams {
  /** 缺省为场景名称 */
  title?: string;
  scene: string;
}

/** 创建会话 POST /chat/sessions */
export function createChatSession(data: CreateChatSessionParams) {
  return axios.post<ChatSession>('/chat/sessions', data);
}

export interface ChatMessagesParams {
  session_code: string;
  /** 默认 50 */
  limit?: number;
  /** 默认 0 */
  offset?: number;
  role?: 'user' | 'assistant' | 'tool' | 'all';
  /** 默认 desc */
  order?: 'asc' | 'desc';
}

/** 会话历史消息 GET /chat/sessions/messages */
export function queryChatMessages(params: ChatMessagesParams) {
  return axios.get<{ messages: ChatHistoryMessage[]; total: number }>(
    '/chat/sessions/messages',
    { params }
  );
}

export interface SendChatMessageParams {
  session_code: string;
  content: string;
  content_type?: ChatContentType;
  /** 幂等标识，UUID v4 */
  request_id: string;
  /** JSON 字符串 */
  metadata?: string;
}

/** 发送消息（同步返回 AI 回复）POST /chat/sessions/messages/send */
export function sendChatMessage(data: SendChatMessageParams) {
  return axios.post<ChatSentMessage>('/chat/sessions/messages/send', data, {
    timeout: 120 * 1000,
  });
}
