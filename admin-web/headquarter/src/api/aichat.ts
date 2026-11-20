import axios from 'axios';

export interface ChatScenario {
  /** 场景名称，唯一标识 */
  name: string;
  /** 场景描述 */
  description: string;
  /** 场景图标 */
  icon: string;
  /** 使用的 AI 模型名称 */
  model_name: string;
  /** 模型温度参数（0.0-2.0） */
  temperature: number;
  /** 可用工具列表 */
  available_tools: string[];
  /** 最大保留最近消息数 */
  max_recent_messages: number;
}

/** 获取场景列表 GET /admin/chat/scenarios（走 28003） */
export function queryChatScenarios() {
  return axios.get<{ scenarios: ChatScenario[] }>('/chat/scenarios');
}

export type ChatSessionStatus = 'ACTIVE' | 'CLOSED';

export interface ChatSession {
  /** 会话数据库 ID */
  id: number;
  /** 会话编码（唯一标识） */
  session_code: string;
  /** 所属用户 ID */
  user_id: number;
  /** 运营商编码 */
  op_code: string;
  /** 会话标题 */
  title: string;
  /** 会话状态 */
  status: ChatSessionStatus;
  /** 场景名称 */
  scene: string;
  /** 对话上下文摘要 */
  context_summary?: string;
  /** 会话中的消息数量 */
  message_count?: number;
  /** 创建时间，毫秒时间戳 */
  created_at: number;
  /** 更新时间，毫秒时间戳 */
  updated_at: number;
}

/** 获取会话列表 GET /admin/chat/sessions（走 28003） */
export function queryChatSessions() {
  return axios.get<{ sessions: ChatSession[] }>('/chat/sessions');
}

export interface CreateChatSessionParams {
  /** 会话标题，不传默认为场景名称 */
  title?: string;
  /** 场景名称 */
  scene: string;
}

/** 创建会话 POST /admin/chat/sessions（走 28003） */
export function createChatSession(data: CreateChatSessionParams) {
  return axios.post<ChatSession>('/chat/sessions', data);
}

export type ChatRole = 'user' | 'assistant' | 'tool';

export interface ChatHistoryMessage {
  /** 消息 ID */
  id: string;
  /** 所属会话编码 */
  sessionCode: string;
  role: ChatRole;
  /** 消息内容（assistant 为 answer 字段） */
  content: string;
  /** 创建时间，ISO 8601 */
  createdAt: string;
  /** 是否需要用户确认（通常表示回复中包含 SQL） */
  isNeedConfirm: boolean;
}

export interface ChatMessagesParams {
  session_code: string;
  /** 默认 50 */
  limit?: number;
  /** 默认 0 */
  offset?: number;
  role?: ChatRole | 'all';
  /** 默认 desc */
  order?: 'asc' | 'desc';
}

export interface ChatMessagesRes {
  messages: ChatHistoryMessage[];
  total: number;
}

/** 获取会话消息 GET /admin/chat/sessions/messages（走 28003） */
export function queryChatMessages(params: ChatMessagesParams) {
  return axios.get<ChatMessagesRes>('/chat/sessions/messages', { params });
}

export type ChatContentType = 'text' | 'json' | 'markdown';

export type ChatMessageStatus =
  | 'processing'
  | 'completed'
  | 'failed'
  | 'cancelled';

export interface SendChatMessageParams {
  session_code: string;
  content: string;
  /** 默认 text */
  content_type?: ChatContentType;
  /** 请求唯一标识，用于幂等性控制（UUID v4） */
  request_id: string;
  /** 附加元数据（JSON 字符串） */
  metadata?: string;
}

export interface SendChatMessageRes {
  /** 消息数据库 ID */
  id: number;
  /** 所属会话 ID */
  session_id: number;
  /** 该会话中的消息序号 */
  sequence: number;
  /** 是否需要用户确认（回复中包含可执行的 SQL） */
  is_need_confirm: boolean;
  role: ChatRole;
  /** AI 的最终回答 */
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
  /** 创建时间，毫秒时间戳 */
  created_at: number;
}

/** 发送消息 POST /admin/chat/sessions/messages/send（走 28003），AI 处理通常需要 1-5 秒 */
export function sendChatMessage(data: SendChatMessageParams) {
  return axios.post<SendChatMessageRes>('/chat/sessions/messages', data, {
    timeout: 120 * 1000,
  });
}

export type KnowledgeDocumentType =
  | 'TABLE_SCHEMA'
  | 'METRIC_DEFINITION'
  | 'SQL_EXAMPLE'
  | 'BUSINESS_RULE'
  | 'KNOWLEDGE';

/** 创建后返回 PROCESSING/COMPLETED/FAILED，列表筛选为 DRAFT/PUBLISHED/ARCHIVED */
export type KnowledgeDocumentStatus =
  | 'PROCESSING'
  | 'COMPLETED'
  | 'FAILED'
  | 'DRAFT'
  | 'PUBLISHED'
  | 'ARCHIVED';

export interface CreateKnowledgeDocumentParams {
  /** 文档标题，不超过 255 字符 */
  title: string;
  /** 文档内容，将被自动切分 */
  content: string;
  document_type: KnowledgeDocumentType;
  /** 关联的源表名 */
  source_table?: string;
  description?: string;
  tags?: string[];
  /** 自定义元数据（JSON 字符串） */
  metadata?: string;
}

export interface CreateKnowledgeDocumentRes {
  document_id: number;
  document_code: string;
  title: string;
  /** 自动分块后的块数 */
  chunk_count: number;
  status: KnowledgeDocumentStatus;
  embed_model: string;
  embed_dim: number;
  /** 创建时间，毫秒时间戳 */
  created_at: number;
}

/** 创建知识库文档 POST /admin/knowledge/documents（走 28003） */
export function createKnowledgeDocument(data: CreateKnowledgeDocumentParams) {
  return axios.post<CreateKnowledgeDocumentRes>('/knowledge/documents', data, {
    timeout: 120 * 1000,
  });
}

export interface KnowledgeDocument {
  id: number;
  document_code: string;
  title: string;
  content: string;
  document_type: KnowledgeDocumentType;
  source_table?: string;
  description?: string;
  content_hash?: string;
  version: number;
  status: KnowledgeDocumentStatus;
  chunk_count: number;
  embed_model?: string;
  embed_dim?: number;
  /** 创建时间，毫秒时间戳 */
  created_at: number;
  /** 更新时间，毫秒时间戳 */
  updated_at: number;
}

export interface KnowledgeDocumentsParams {
  document_type?: KnowledgeDocumentType;
  source_table?: string;
  status?: KnowledgeDocumentStatus;
  /** 在标题或内容中搜索 */
  keyword?: string;
  /** 默认 20，最大 100 */
  limit?: number;
  /** 默认 0 */
  offset?: number;
}

export interface KnowledgeDocumentsRes {
  documents: KnowledgeDocument[];
  total: number;
  limit: number;
  offset: number;
}

/** 查询知识库文档列表 GET /admin/knowledge/documents（走 28003） */
export function queryKnowledgeDocuments(params: KnowledgeDocumentsParams) {
  return axios.get<KnowledgeDocumentsRes>('/knowledge/documents', { params });
}

export interface DeleteKnowledgeDocumentRes {
  document_id: number;
  deleted: boolean;
  message: string;
}

/** 删除知识库文档 DELETE /admin/knowledge/documents?id=（走 28003），同时清理向量分块 */
export function deleteKnowledgeDocument(id: number | string) {
  return axios.delete<DeleteKnowledgeDocumentRes>('/knowledge/documents', {
    params: { id },
  });
}
